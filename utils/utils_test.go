package utils

import "testing"

func TestGetFileRealSize(t *testing.T) {
	path := "/Users/smallyu/work/gogs/edge-deploy/cnode1"
	size, err := GetFileRealSize(path)
	if err == nil {
		t.Error("GetFileRealSize should return error")
	}
	t.Log(size)
}
