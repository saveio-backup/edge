package plot

import (
	"fmt"
	"testing"
)

func TestGetMinStartNonce(t *testing.T) {
	sn, err := GetMinStartNonce("3988", "/Users/zhijie/Desktop/onchain/save-test/plot_test")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("start nonce %v\n", sn)
}
