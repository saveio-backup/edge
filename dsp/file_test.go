package dsp

import (
	"bytes"
	"encoding/hex"
	"github.com/saveio/themis/crypto/keypair"
	"strings"
	"testing"

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

func TestDecodePrivateKey(t *testing.T) {
	privKey := "f1472f1fc52a8674d361b7e6af23ada4522526aca304b9729c5a9518b909f1b6"
	privateKey, err := hex.DecodeString(strings.TrimPrefix(privKey, "0x"))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("privateKey :%v", privateKey)
	//pKey, err := keypair.DeserializePrivateKey(privateKey)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("pKey :%v", pKey)

	pri, pub, err := keypair.GenerateKeyPairWithSeed(
		keypair.PK_ECDSA,
		bytes.NewReader(privateKey),
		keypair.P256,
	)
	t.Log(pri, pub, err)
	//
	//pKey, err := keypair.DeserializePrivateKey(pri)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("pKey :%v", pKey)
}
