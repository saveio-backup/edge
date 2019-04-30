package dsp

import (
	"fmt"
	"os"
	"testing"

	"github.com/saveio/dsp-go-sdk/store"
	"github.com/saveio/dsp-go-sdk/task"
)

func TestClientDB(t *testing.T) {
	ep := &Endpoint{}
	db, err := store.NewLevelDBStore("./edge_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove("./edge_test")
	ep.db = db
	task1 := &task.ProgressInfo{
		TaskKey:  "1",
		FileName: "1",
	}
	task2 := &task.ProgressInfo{
		TaskKey:  "2",
		FileName: "2",
	}
	err = ep.AddProgress(task1)
	if err != nil {
		t.Fatal(err)
	}
	err = ep.AddProgress(task2)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := ep.GetAllProgressKeys()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("keys %v, len:%d\n", keys, len(keys))
	info, err := ep.GetProgress("1")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("info %v\n", info)
}
