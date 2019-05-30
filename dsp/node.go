package dsp

import (
	"github.com/saveio/themis/common/log"
	fs "github.com/saveio/themis/smartcontract/service/native/onifs"
)

//Dsp api
func (this *Endpoint) RegisterNode(addr string, volume, serviceTime uint64) (string, error) {
	tx, err := this.Dsp.RegisterNode(addr, volume, serviceTime)
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", err
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) UnregisterNode() (string, error) {
	tx, err := this.Dsp.UnregisterNode()
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", err
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) NodeQuery(walletAddr string) (*fs.FsNodeInfo, error) {
	info, err := this.Dsp.QueryNode(walletAddr)
	if err != nil {
		log.Errorf("query node err %s", err)
		return nil, err
	}
	log.Infof("node info pledge %d", info.Pledge)
	log.Infof("node info profit %d", info.Profit)
	log.Infof("node info volume %d", info.Volume)
	log.Infof("node info restvol %d", info.RestVol)
	log.Infof("node info service time %d", info.ServiceTime)
	log.Infof("node info wallet address %s", info.WalletAddr.ToBase58())
	log.Infof("node info node address %s", info.NodeAddr)

	return info, nil
}

func (this *Endpoint) NodeUpdate(addr string, volume, serviceTime uint64) (string, error) {
	tx, err := this.Dsp.UpdateNode(addr, volume, serviceTime)
	if err != nil {
		log.Errorf("update node err:%s", err)
		return "", err
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) NodeWithdrawProfit() (string, error) {
	tx, err := this.Dsp.NodeWithdrawProfit()
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", err
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}
