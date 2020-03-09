package dsp

import "testing"

func TestOpenChannel(t *testing.T) {
	edge, err := Init("./wallet.dat", "pwd")
	if err != nil {
		t.Fatal(err)
	}
	if err := StartDspNode(edge, true, true, true); err != nil {
		t.Fatal(err)
	}

}
