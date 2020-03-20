package dsp

import (
	"os"
	"testing"
	"time"

	"github.com/saveio/dsp-go-sdk/dsp"
	"github.com/saveio/themis/common/log"
)

func TestLogoutWhenPylonsStartup(t *testing.T) {
	os.RemoveAll("./Chain-0")
	os.RemoveAll("./wallet.dat")
	defer func() {
		log.Debugf("remove data")
		os.RemoveAll("Chain-0")
		os.RemoveAll("./wallet.dat")
	}()
	resp, err := (&Endpoint{}).ImportWithPrivateKey("KxSyfWPmtaRd1r7Davc32CFDSFkWrVkm53CG2WB5u8cNKqaAbyrX", "pwd", "pwd")
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
	os.RemoveAll("./Chain-0")
	os.RemoveAll("./wallet.dat")
	os.RemoveAll("./Log")
	defer func() {
		log.Debugf("remove data")
		os.RemoveAll("Chain-0")
		os.RemoveAll("./wallet.dat")
		os.RemoveAll("./Log")
	}()
	log.InitLog(0, os.Stdout, "./Log/")
	for i := 0; i < 3; i++ {
		resp, err := (&Endpoint{}).ImportWithPrivateKey("KxSyfWPmtaRd1r7Davc32CFDSFkWrVkm53CG2WB5u8cNKqaAbyrX", "pwd", "pwd")
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
