package utils

import "testing"

func TestGetFileRealSize(t *testing.T) {
	path := "/Users/smallyu/work/gogs/edge-deploy/node1/testaaa37"
	size, err := GetFileRealSize(path)
	if err != nil {
		t.Error(err)
	}
	t.Log(size)
	t.Log(size / 1000)
	t.Log(size / 1000 / 1000)
}
