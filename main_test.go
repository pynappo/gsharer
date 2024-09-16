package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-test/deep"
)

var sampletxt = "./testdata/sample.txt"

// Litterbox is a temporary filehost so uploading small files w/ minimum duration should be pretty harmless
// TODO: maybe find a different filehost or bootstrap a local http server instead?
func formTestLitterboxRequest() (req *http.Request, length int, err error) {
	url := "https://litterbox.catbox.moe/resources/internals/api.php"
	method := "POST"

	payload := new(bytes.Buffer)
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("reqtype", "fileupload")
	_ = writer.WriteField("time", "1h")
	path, _ := filepath.Abs("./testdata/sample.txt")
	file, err := os.Open(path)
	defer file.Close()
	part, err := writer.CreateFormFile("fileToUpload", path)
	_, err = io.Copy(part, file)
	if err != nil {
		return
	}
	errWrite := writer.Close()
	if errWrite != nil {
		return
	}

	req, err = http.NewRequest(method, url, payload)

	if err != nil {
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return
}

// Litterbox is a temporary filehost so uploading small files w/ minimum duration should be pretty harmless
func TestUploadToLitterbox(t *testing.T) {
	worker, err := newWorker()
	if err != nil {
		t.Error(err)
	}
	path, _ := filepath.Abs("./testdata/sample.txt")
	file, err := os.Open(path)
	if err != nil {
		t.Error(err)
	}
	streams := []NamedStream{
		{
			name:       path,
			ReadCloser: file,
		},
	}
	job, err := CreateJob(worker.L, "litterbox", streams)
	if err != nil {
		t.Error(err)
	}
	url, err := worker.Upload(job)
	if err != nil {
		t.Error(err)
	}
	t.Log("url", url)
}

func TestLocalUpload(t *testing.T) {
	echoServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, string(body))
	}))
	worker, err := newWorker()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	err := worker.L.DoString()
	job := CreateJob(worker.L, "auto")
	worker.Upload(&UploadJob{
		request: &http.Request{},
	})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log("url", url)
}

func TestCreateJob(t *testing.T) {
	worker, err := newWorker()
	if err != nil {
		t.Error(err)
	}
	path, _ := filepath.Abs("./testdata/sample.txt")
	file, err := os.Open(path)
	if err != nil {
		t.Error(err)
	}
	streams := []NamedStream{
		{
			name:       path,
			ReadCloser: file,
		},
	}
	job, err := CreateJob(worker.L, "litterbox", streams)
	if err != nil {
		t.Error(err)
	}
	testRequest, _, err := formTestLitterboxRequest()
	if err != nil {
		t.Error(err)
	}
	if diff := deep.Equal(job.request, testRequest); diff != nil {
		t.Log(diff)
	}
}

func TestWorker(t *testing.T) {
	_, err := newWorker()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
func TestUpload(t *testing.T) {
	req, _, err := formTestLitterboxRequest()
	if err != nil {
		t.Log("Test request did not form properly", err)
		t.FailNow()
	}

	worker, err := newWorker()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	url, err := worker.Upload(&UploadJob{
		request: req,
	})
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log("url", url)
}
