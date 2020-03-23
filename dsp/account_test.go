package dsp

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/saveio/dsp-go-sdk/dsp"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/themis/common/log"
)

var walletText = `{"name":"MyWallet","version":"1.1","scrypt":{"p":8,"n":16384,"r":8,"dkLen":64},"accounts":[{"address":"AHjjdbVLhfTyiNFEq2X8mFnnirZY1yK8Rq","enc-alg":"aes-256-gcm","key":"gxQmlkEeNYtztEH3oTdjOkDsjpH13cbhHLgdMfjBkuwv0qhKSciQi//FESPmYxKN","algorithm":"ECDSA","salt":"A/7nwFwsLK5gLmlXW/V5sg==","parameters":{"curve":"P-256"},"label":"pwd","publicKey":"0392f8ff7ace886c5bcd76193692c32af16db6df292ee3d893f71645a354a796eb","signatureScheme":"SHA256withECDSA","isDefault":true,"lock":false}]}`
var password = "pwd"
var privateKey = "KxSyfWPmtaRd1r7Davc32CFDSFkWrVkm53CG2WB5u8cNKqaAbyrX"

func TestMain(m *testing.M) {
	os.RemoveAll("./Chain-12345")
	os.RemoveAll("./wallet.dat")
	// os.RemoveAll("./Log")
	defer func() {
		os.RemoveAll("Chain-12345")
		os.RemoveAll("./wallet.dat")
		// os.RemoveAll("./Log")
	}()
	log.InitLog(0, os.Stdout, "./Log/")
	m.Run()
}

func TestLogoutWhenPylonsStartup(t *testing.T) {
	resp, err := (&Endpoint{}).ImportWithPrivateKey(privateKey, password, password)
	if err != nil {
		t.Fatal(err)
	}
	log.Infof("account resp +++++++ %v", resp.Address)
	for {
		result, err := DspService.GetModuleState()
		if err != nil {
			t.Fatal(err)
		}
		moduleState, ok := result.([]*dsp.ModuleStateResp)
		if !ok {
			continue
		}
		if moduleState[0].State >= 3 {
			break
		}
		<-time.After(time.Duration(1) * time.Second)
	}
	log.Infof("start success")
	if err := DspService.Logout(); err != nil {
		t.Fatal(err)
	}
}

func TestFrequencyLogout(t *testing.T) {

	for i := 0; i < 3; i++ {
		resp, err := (&Endpoint{}).ImportWithPrivateKey(privateKey, password, password)
		if err != nil {
			t.Fatal(err)
		}
		log.Warnf("+++++ login success %v", resp.Address)
		for {
			result, err := DspService.GetFilterBlockProgress()
			if err != nil {
				continue
			}
			if result.Progress == 1.0 {
				break
			}
			<-time.After(time.Duration(1) * time.Second)
		}
		dsp := DspService.getDsp()
		for {
			<-time.After(time.Duration(1) * time.Second)
			if dsp.HasChannelInstance() && dsp.ChannelFirstSyncing() {
				continue
			}
			break
		}

		log.Warnf("start logout")
		if err := DspService.Logout(); err != nil {
			t.Fatal(err)
		}
		log.Warnf("+++++ logout success")
	}

}

func TestLoginWithProxy(t *testing.T) {
	for i := 0; i < 30; i++ {
		config.UserLocalCfg()
		if err := ioutil.WriteFile("./wallet.dat", []byte(walletText), 0666); err != nil {
			t.Fatal(err)
		}
		resp, err := (&Endpoint{}).Login(password)
		if err != nil {
			t.Fatal(err)
		}
		log.Warnf("+++++ login success %v", resp.Address)
		waitConnect := 0
		for {
			result, err := DspService.GetNetworkState()
			if err != nil {
				continue
			}
			if len(result.DspProxy.HostAddr) > 0 && result.DspProxy.State == networkStateReachable {
				log.Infof("proxy is connected break")
				break
			}
			<-time.After(time.Duration(1) * time.Second)
			waitConnect++
			if waitConnect > 30 {
				t.Fatal("wait for connect proxy timeout")
			}
		}
		dsp := DspService.getDsp()
		for {
			<-time.After(time.Duration(1) * time.Second)
			if dsp.HasChannelInstance() && dsp.ChannelFirstSyncing() {
				continue
			}
			break
		}
		log.Warnf("start logout")
		if err := DspService.Logout(); err != nil {
			t.Fatal(err)
		}
		log.Warnf("+++++ logout success")
	}
}
