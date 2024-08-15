package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	// "log"
	"path/filepath"
	"strings"

	"github.com/aarzilli/golua/lua"
)

//go:embed lua
var luaLibraries embed.FS

func initLuaState() (*lua.State, error) {
	L := lua.NewState()
	L.OpenLibs()
	// create the gsharer global table
	err := doEmbeddedFile(L, "lua/gsharer/init.lua")
	if err != nil {
		return nil, err
	}
	// push an embedded lua module loader onto the gsharer table
	L.GetGlobal("gsharer")
	L.PushString("_embedded_loader")
	L.PushGoClosure(func(L1 *lua.State) int {
		modname := L1.CheckString(1)
		filename := filepath.Join("lua", strings.ReplaceAll(modname, ".", "/")+".lua")
		// fmt.Println("searching for " + filename)
		luaErr := loadEmbeddedFile(L1, filename)
		// if we didn't find lua/foo.lua, try lua/foo/init.lua
		if luaErr == lua.LUA_ERRFILE {
			filename = filepath.Join("lua", strings.ReplaceAll(modname, ".", "/")+"/init.lua")
			luaErr = loadEmbeddedFile(L1, filename)
		}
		if luaErr != 0 {
			// if verboseLogging {
			// 	fmt.Printf("[embedded lua loader] could not load module %s (searched for file %s)\n", modname, filename)
			// 	switch luaErr {
			// 	case lua.LUA_ERRFILE:
			// 		fmt.Printf("[embedded lua loader] error code LUA_ERRFILE")
			// 	case lua.LUA_ERRSYNTAX:
			// 		fmt.Printf("[embedded lua loader] error code LUA_ERRSYNTAX")
			// 	case lua.LUA_ERRMEM:
			// 		fmt.Printf("[embedded lua loader] error code LUA_ERRMEM")
			// 	case lua.LUA_ERRERR:
			// 		fmt.Printf("[embedded lua loader] error code LUA_ERRERR")
			// 	}
			// }
			// maybe print more detailed messages later if needed
			return 0
		}
		// tell lua that we loaded a file
		return 1
	})
	L.SetTable(-3)

	// setup gsharer global w/ utils and such
	err = doEmbeddedFile(L, "lua/gsharer/load_globals.lua")
	if err != nil {
		return nil, err
	}
	return L, nil
}

func loadEmbeddedFile(L *lua.State, embeddedFilename string) int {
	embeddedFile, err := luaLibraries.Open(embeddedFilename)
	if err != nil {
		return lua.LUA_ERRFILE
	}
	defer embeddedFile.Close()
	bytes, err := io.ReadAll(embeddedFile)
	if err != nil {
		return lua.LUA_ERRFILE
	}
	return L.LoadString(string(bytes))
}

func doEmbeddedFile(L *lua.State, embeddedFilename string) error {
	code := loadEmbeddedFile(L, embeddedFilename)
	if code != 0 {
		switch code {
		case lua.LUA_ERRFILE:
			return errors.New("Could not load embedded file " + embeddedFilename)
		case lua.LUA_ERRSYNTAX:
			return errors.New("Embedded file " + embeddedFilename + " had invalid syntax")
		case lua.LUA_ERRMEM:
			return errors.New("Memory issue when loading embedded file " + embeddedFilename)
		case lua.LUA_ERRERR:
			return errors.New("Lua error when loading embedded file " + embeddedFilename)
		}
	}
	return L.Call(0, lua.LUA_MULTRET)
}

// Recursively convert a lua table (assuming string/number keys only) into a go string map
func luaTableToStringMap(L *lua.State, tableIndex int) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if tableIndex < 0 {
		tableIndex = L.GetTop() + 1 + tableIndex
	}
	if !L.IsTable(tableIndex) {
		return nil, errors.New("value given was a " + L.Typename(-1) + ", not a table")
	}
	L.PushNil()
	keysTraversed := 0
	defer L.Pop(keysTraversed)
	for L.Next(tableIndex) != 0 {
		key := L.ToString(-2)
		switch L.Type(-1) {
		case lua.LUA_TTABLE:
			tableValue, err := luaTableToStringMap(L, -1)
			result[key] = tableValue
			if err != nil {
				return result, err
			}
		case lua.LUA_TBOOLEAN:
			result[key] = L.ToBoolean(-1)
		case lua.LUA_TNUMBER:
			fallthrough
		case lua.LUA_TSTRING:
			result[key] = L.ToString(-1)
		case lua.LUA_TFUNCTION:
			dumpErrCode := L.Dump()
			if dumpErrCode != 0 {
				return result, errors.New(fmt.Sprintf("could not dump function to bytecode, error code: %v", dumpErrCode))
			}
			bytecode := L.ToBytes(-1)
			L.Pop(1)
			result[key] = func(L *lua.State) error {
				errCode := L.Load(bytecode, "lua bytecode dumped by luaTableToStringMap with key "+key)
				if errCode != 0 {
					return errors.New("could not load bytecode from luaTableToStringMap")
				}
				return nil
			}
		default:
			return result, errors.New("unexpected value type in table with key " + key)
		}
		L.Pop(1)
		keysTraversed++
	}
	return result, nil
}

func printStack(L *lua.State) {
	top := L.GetTop()
	for i := top; i >= 1; i-- {
		fmt.Printf("%d\t(type %s)\t", i, L.Typename(int(L.Type(i))))
		switch L.Type(i) {
		case lua.LUA_TNUMBER:
			fmt.Printf("%g\n", L.ToNumber(i))
		case lua.LUA_TSTRING:
			fmt.Printf("\"%s\" \n", L.ToString(i))
		case lua.LUA_TBOOLEAN:
			fmt.Printf("%v\n", L.ToBoolean(i))
		case lua.LUA_TNIL:
			fmt.Printf("%s\n", "nil")
		default:
			fmt.Printf("%v\n", L.ToPointer(i))
		}
	}
}
