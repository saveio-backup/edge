package dsp

import (
	"os"
	"testing"
	"time"

	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/crypto/keypair"
	s "github.com/saveio/themis/crypto/signature"
)

//map info, to get some information easily
type keyTypeInfo struct {
	name string
	code keypair.KeyType
}

var keyTypeMap = map[string]keyTypeInfo{
	"":  {"ecdsa", keypair.PK_ECDSA},
	"1": {"ecdsa", keypair.PK_ECDSA},
	"2": {"sm2", keypair.PK_SM2},
	"3": {"ed25519", keypair.PK_EDDSA},

	"ecdsa":   {"ecdsa", keypair.PK_ECDSA},
	"sm2":     {"sm2", keypair.PK_SM2},
	"ed25519": {"ed25519", keypair.PK_EDDSA},
}

type curveInfo struct {
	name string
	code byte
}

var curveMap = map[string]curveInfo{
	"":  {"P-256", keypair.P256},
	"1": {"P-224", keypair.P224},
	"2": {"P-256", keypair.P256},
	"3": {"P-384", keypair.P384},
	"4": {"P-521", keypair.P521},

	"P-224": {"P-224", keypair.P224},
	"P-256": {"P-256", keypair.P256},
	"P-384": {"P-384", keypair.P384},
	"P-521": {"P-521", keypair.P521},

	"224": {"P-224", keypair.P224},
	"256": {"P-256", keypair.P256},
	"384": {"P-384", keypair.P384},
	"521": {"P-521", keypair.P521},

	"SM2P256V1": {"SM2P256V1", keypair.SM2P256V1},
	"ED25519":   {"ED25519", keypair.ED25519},
}

type schemeInfo struct {
	name string
	code s.SignatureScheme
}

var schemeMap = map[string]schemeInfo{
	"":  {"SHA256withECDSA", s.SHA256withECDSA},
	"1": {"SHA224withECDSA", s.SHA224withECDSA},
	"2": {"SHA256withECDSA", s.SHA256withECDSA},
	"3": {"SHA384withECDSA", s.SHA384withECDSA},
	"4": {"SHA512withECDSA", s.SHA512withECDSA},
	"5": {"SHA3-224withECDSA", s.SHA3_224withECDSA},
	"6": {"SHA3-256withECDSA", s.SHA3_256withECDSA},
	"7": {"SHA3-384withECDSA", s.SHA3_384withECDSA},
	"8": {"SHA3-512withECDSA", s.SHA3_512withECDSA},
	"9": {"RIPEMD160withECDSA", s.RIPEMD160withECDSA},

	"SHA224withECDSA":    {"SHA224withECDSA", s.SHA224withECDSA},
	"SHA256withECDSA":    {"SHA256withECDSA", s.SHA256withECDSA},
	"SHA384withECDSA":    {"SHA384withECDSA", s.SHA384withECDSA},
	"SHA512withECDSA":    {"SHA512withECDSA", s.SHA512withECDSA},
	"SHA3-224withECDSA":  {"SHA3-224withECDSA", s.SHA3_224withECDSA},
	"SHA3-256withECDSA":  {"SHA3-256withECDSA", s.SHA3_256withECDSA},
	"SHA3-384withECDSA":  {"SHA3-384withECDSA", s.SHA3_384withECDSA},
	"SHA3-512withECDSA":  {"SHA3-512withECDSA", s.SHA3_512withECDSA},
	"RIPEMD160withECDSA": {"RIPEMD160withECDSA", s.RIPEMD160withECDSA},

	"SM3withSM2":      {"SM3withSM2", s.SM3withSM2},
	"SHA512withEdDSA": {"SHA512withEdDSA", s.SHA512withEDDSA},
}

func TestOpenChannel(t *testing.T) {
	os.RemoveAll("./Chain-0")
	os.RemoveAll("./wallet.dat")
	defer func() {
		log.Debugf("remove data")
		os.RemoveAll("Chain-0")
		os.RemoveAll("./wallet.dat")
	}()
	resp, err := (&Endpoint{}).NewAccount("1", keyTypeMap["ecdsa"].code, curveMap["P-256"].code, schemeMap["SHA256withECDSA"].code, []byte("123"), false)
	if err != nil {
		t.Fatal(err)
	}
	log.Infof("account resp +++++++ %v", resp.Address)
	for {
		prog, err := DspService.GetFilterBlockProgress()
		if err != nil {
			t.Fatal(err)
		}
		syncing, err := DspService.IsChannelProcessBlocks()
		log.Infof("progress resp %v syncing %t, err %v", prog, syncing, err)
		if prog.Progress == 1 && !syncing && err == nil {
			break
		}
		<-time.After(time.Duration(5) * time.Second)
	}

	id, err := DspService.OpenPaymentChannel("AN8rkVqfo3dNAzJY3Dzh3iikzjhEu3vwCJ", 10)
	if err != nil {
		t.Fatal(err)
	}
	log.Infof("open channel id %v", id)
}
