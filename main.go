// Package gsharer implements a GUI around using Lua scripts + Go's HTTP client for uploading files to arbitrary
// destinations.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aarzilli/golua/lua"
	"github.com/adrg/xdg"

	// "github.com/h2non/filetype"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

var ConfigFilePath string
var VERBOSE int

// Just an io.ReadCloser with a name for filling in HTTP forms.
type NamedStream struct {
	io.ReadCloser
	name string
}

// These functions should push the HTTP response handler
// to the top of the provided Lua state's stack.
type ResponseHandlerPusher func(L *lua.State) error

// This should hold the data needed for workers to make a request and parse the response.
type UploadJob struct {
	request        *http.Request
	pushResHandler ResponseHandlerPusher
}

// Each worker is responsible for its own lua state and HTTP client.
type Worker struct {
	L          *lua.State
	httpClient *http.Client
	logger     *slog.Logger
}

func newWorker() (*Worker, error) {
	L, err := initLuaState()
	if err != nil {
		return nil, err
	}
	return &Worker{
		L:          L,
		httpClient: &http.Client{},
	}, nil
}

func (worker *Worker) Close() {
	worker.L.Close()
	worker.httpClient.CloseIdleConnections()
}

func main() {
	slog.SetDefault(GsharerLogger())
	ConfigFilePath, err := xdg.SearchConfigFile("gsharer/main.lua")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(2)
	}
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Value:   2,
				Usage:   "Sets the log level: 0=ERROR, 1=WARN, 2=INFO, 3+=DEBUG",
				EnvVars: []string{"GSHARER_VERBOSE"},
			},
		},
		Before: func(ctx *cli.Context) error {
			switch verbosity := ctx.Int("verbose"); {
			case verbosity == 0:
				slog.SetLogLoggerLevel(slog.LevelError)
			case verbosity == 1:
				slog.SetLogLoggerLevel(slog.LevelWarn)
			case verbosity == 2:
				slog.SetLogLoggerLevel(slog.LevelInfo)
			case verbosity >= 3:
				slog.SetLogLoggerLevel(slog.LevelDebug)
			default:
				return errors.New(fmt.Sprintf("Invalid log level set, was %v, only 0 or more accepted", verbosity))
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "file",
				Usage: "upload 1 file from stdin or all the files from the argument list",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "destination",
						Aliases: []string{"d"},
						Value:   "auto",
						Usage:   "where to upload to",
					},
					&cli.UintFlag{
						Name:    "threads",
						Aliases: []string{"t"},
						Value:   4,
						Usage:   "limits the amount of threads made when #files > #threads",
					},
					&cli.StringFlag{
						Name:  "name",
						Value: "",
						Usage: "if uploading raw stdin data, the name of the filename when uploaded (if applicable to the destination)",
					},
					&cli.BoolFlag{
						Name:    "strict",
						Aliases: []string{"s"},
						Value:   false,
						Usage:   "when set, errors if any argument is not able to be uploaded (e.g. file doesn't exist, rate limit, etc)",
					},
					&cli.BoolFlag{
						Name:    "confirm",
						Aliases: []string{"c"},
						Value:   false,
						Usage:   "before uploading anything, prompts with confirmation prompt showing each request that will be made",
					},
					&cli.BoolFlag{
						Name:    "interactive",
						Aliases: []string{"i"},
						Value:   false,
						Usage:   "Prompts for destination-specific parameters, such as userhash or whatnot",
					},
					&cli.UintFlag{
						Name:    "batch",
						Aliases: []string{"b"},
						Value:   1,
						Usage:   "If 0, auto-batches file uploads based on the destination parameters. If 1 or above, limits the batch sizes to at most the given number.",
					},
				},
				Action: func(cCtx *cli.Context) error {
					if cCtx.Bool("interactive") {
						os.Setenv("GSHARER_INTERACTIVE", "1")
					}
					nArg := cCtx.NArg()
					if nArg > 0 {
						// Upload files from args.
						args := cCtx.Args()

						// Validate files into jobs
						var files []*os.File
						set := make(map[string]struct{})
						for i := range nArg {
							fullFilepath, err := filepath.Abs(args.Get(i))
							if err != nil {
								return err
							}
							if _, alreadyAJob := set[fullFilepath]; alreadyAJob {
								continue
							}
							set[fullFilepath] = struct{}{}

							file, err := os.Open(fullFilepath)
							if err != nil {
								if cCtx.Bool("strict") {
									return err
								}
								continue
							}
							defer file.Close()
							files = append(files, file)
						}

						uploadJobs, err := batchIntoJobs(files, cCtx.String("destination"), false)
						if err != nil {
							return err
						}

						// create workers to do upload jobs
						maxThreads := min(len(uploadJobs), cCtx.Int("threads"))
						queue := make(chan *UploadJob)
						wg := &sync.WaitGroup{}
						defer wg.Wait()

						for i := range maxThreads {
							wg.Add(1)
							go func() error {
								defer wg.Done()

								worker, err := newWorker(logger.With(slog.Int("workerNumber", i)))
								if err != nil {
									return err
								}
								defer worker.Close()
								defer logger.Info(fmt.Sprintf("worker %d done", i))

								for uploadJob := range queue {
									url, err := worker.Upload(uploadJob)
									if err != nil {
										logger.Error(err.Error())
										continue
									}
									fmt.Println(url)
								}
								return nil
							}()
						}

						// push the jobs into the queue
						for _, job := range uploadJobs {
							queue <- job
						}
						close(queue)
					} else {
						// upload via stdin
						worker, err := newWorker(logger.With(slog.Int("number", 0)))
						if err != nil {
							return err
						}
						defer worker.Close()

						logger.Info("Reading from stdin...")
						name := cCtx.String("name")
						job, err := CreateJob(worker.L, cCtx.String("destination"),
							[]NamedStream{
								{
									ReadCloser: os.Stdin,
									name:       name,
								},
							})
						if err != nil {
							return err
						}

						url, err := worker.Upload(job)
						if err != nil {
							return err
						}

						fmt.Println(url)
					}
					return nil
				},
			},
			{
				Name:  "sync",
				Usage: "sync with a server",
				Action: func(cCtx *cli.Context) error {
					logger.Error("Not functional, but u can look at this progress bar!")
					bar := progressbar.Default(100)
					for i := 0; i < 100; i++ {
						bar.Add(1)
					}
					return nil
				},
			},
			{
				Name:  "config",
				Usage: "print config file location",
				Action: func(cCtx *cli.Context) error {
					fmt.Println(ConfigFilePath)
					return nil
				},
			},
			{
				Name:  "auth",
				Usage: "authenticates to a supported OAUTH provider",
				Action: func(cCtx *cli.Context) error {
					fmt.Println(ConfigFilePath)
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		slog.Error(fmt.Errorf("[gsharer], %v", err).Error())
		os.Exit(1)
	}
}

func batchIntoJobs(files []*os.File, destination string, autobatch bool) (jobs []*UploadJob, err error) {
	L, err := initLuaState()
	if !autobatch {
		for _, file := range files {
			namedStreams := []NamedStream{
				{
					ReadCloser: file,
					name:       file.Name(),
				},
			}
			job, jobErr := CreateJob(L, destination, namedStreams)
			if jobErr != nil {
				err = jobErr
				return
			}
			jobs = append(jobs, job)
			return
		}
		return
	}
	// var buckets map[string][]*http.Request
	return
}

// Uploads the files to a single destination.
func (worker *Worker) Upload(job *UploadJob) (url string, err error) {
	bar := progressbar.DefaultBytes(1, "Uploading to "+job.request.URL.Path)
	reader := progressbar.NewReader(job.request.Body, bar)
	bar.ChangeMax64(job.request.ContentLength)
	job.request.Body = &reader
	res, err := worker.httpClient.Do(job.request)
	if err != nil {
		err = fmt.Errorf("error making request, %v", err)
		return
	}
	url, err = worker.ParseResponse(res, job.pushResHandler)
	if err != nil {
		err = fmt.Errorf("error parsing response, %v", err)
		return
	}
	return
}

func CreateJob(L *lua.State, destination string, streams []NamedStream) (job *UploadJob, err error) {
	job = &UploadJob{}
	// use lua to help form the request
	if len(streams) == 0 {
		err = errors.New("Can't form a request with nothing to upload")
		return
	}
	if ConfigFilePath == "" {
		doEmbeddedFile(L, "lua/gsharer/default_config.lua")
	} else {
		L.DoFile(ConfigFilePath) // should push a function
	}
	countArgs := 1
	L.PushString(destination)
	for _, stream := range streams {
		L.PushString(stream.name)
		countArgs++
	}
	err = L.Call(countArgs, 1)
	if err != nil {
		return
	}

	printStack(L)
	if err != nil {
		return
	}

	fmt.Println(job.pushResHandler)
	// create the actual request
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	for k, v := range reqArgs {
		err = writer.WriteField(k, v.(string))
		if err != nil {
			return
		}
	}
	for _, namedStream := range streams {
		part, fileErr := writer.CreateFormFile(reqFileFormName, namedStream.name)
		err = fileErr
		if err != nil {
			return
		}
		if _, err = io.Copy(part, namedStream); err != nil {
			return
		}
	}
	if err = writer.Close(); err != nil {
		return
	}
	job.request, err = http.NewRequest(reqMethod, reqURL, body)
	if err != nil {
		return
	}
	job.request.Header.Add("Content-Type", writer.FormDataContentType())
	return
}

func (worker *Worker) ParseResponse(response *http.Response, pushResHandler func(L *lua.State) error) (string, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	if response.StatusCode >= 300 {
		return "", errors.New(fmt.Sprintf("Response did not respond with a good code: %v", response))
	}

	defer response.Body.Close()
	if pushResHandler == nil {
		return "", errors.New("pushResHandler null")
	}
	pushResHandler(worker.L)
	worker.L.PushString(string(body))
	err = worker.L.Call(1, 1)
	if err != nil {
		err = fmt.Errorf("could not call response handler from lua, %v", err)
		return "", err
	}
	url := worker.L.ToString(-1)
	return url, nil
}
