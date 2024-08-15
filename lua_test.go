package main

import (
	"testing"
)

func TestLuaState(t *testing.T) {
	L, err := initLuaState()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	err = L.DoString("return 1 + 1")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
