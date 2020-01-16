package dsp

import (
	"testing"

	"github.com/saveio/dsp-go-sdk/dsp"
	chainCom "github.com/saveio/themis/common"
	fs "github.com/saveio/themis/smartcontract/service/native/savefs"
)

func TestWhitelistSlice(t *testing.T) {
	whitelist := []string{"AcApvuXZTFVmyM8SA8UxG23oqFyYvf35mP"}
	whitelistObj := fs.WhiteList{
		Num:  uint64(len(whitelist)),
		List: make([]fs.Rule, 0, uint64(len(whitelist))),
	}
	t.Logf("whitelist :%v, len: %d %d", whitelist, len(whitelistObj.List), cap(whitelistObj.List))
	for i, whitelistAddr := range whitelist {
		addr, err := chainCom.AddressFromBase58(whitelistAddr)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("index :%d", i)
		whitelistObj.List = append(whitelistObj.List, fs.Rule{
			Addr:         addr,
			BaseHeight:   uint64(0),
			ExpireHeight: 100,
		})
	}
}

func TestCapSlice(t *testing.T) {
	whitelist := []string{"AcApvuXZTFVmyM8SA8UxG23oqFyYvf35mP"}
	List := make([]string, 0, uint64(len(whitelist)))
	List[0] = whitelist[0]
}

func TestGetUploadFiles(t *testing.T) {
	endPoint := &Endpoint{}
	endPoint.Dsp = &dsp.Dsp{}
}
