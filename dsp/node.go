package dsp

import (
	"encoding/hex"

	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/smartcontract/service/native/dns"
	fs "github.com/saveio/themis/smartcontract/service/native/savefs"
)

//Dsp api
func (this *Endpoint) RegisterNode(addr string, volume, serviceTime uint64) (string, *DspErr) {
	tx, err := this.Dsp.RegisterNode(addr, volume, serviceTime)
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", &DspErr{Code: DSP_NODE_REGISTER_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) UnregisterNode() (string, *DspErr) {
	tx, err := this.Dsp.UnregisterNode()
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", &DspErr{Code: DSP_NODE_UNREGISTER_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) NodeQuery(walletAddr string) (*fs.FsNodeInfo, *DspErr) {
	info, err := this.Dsp.QueryNode(walletAddr)
	if err != nil {
		return nil, &DspErr{Code: DSP_NODE_QUERY_FAILED, Error: err}
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

func (this *Endpoint) NodeUpdate(addr string, volume, serviceTime uint64) (string, *DspErr) {
	tx, err := this.Dsp.UpdateNode(addr, volume, serviceTime)
	if err != nil {
		log.Errorf("update node err:%s", err)
		return "", &DspErr{Code: DSP_NODE_UPDATE_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) NodeWithdrawProfit() (string, *DspErr) {
	tx, err := this.Dsp.NodeWithdrawProfit()
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", &DspErr{Code: DSP_NODE_WITHDRAW_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

//Handle for DNS
func (this *Endpoint) RegisterUrl(url, link string) (string, *DspErr) {
	log.Debugf("register url %v link %v", url, link)
	tx, err := this.Dsp.RegisterFileUrl(url, link)
	if err != nil {
		return "", &DspErr{Code: DSP_URL_REGISTER_FAILED, Error: err}
	}
	return tx, nil
}

func (this *Endpoint) BindUrl(url, link string) (string, *DspErr) {
	tx, err := this.Dsp.BindFileUrl(url, link)
	if err != nil {
		return "", &DspErr{Code: DSP_URL_BIND_FAILED, Error: err}
	}
	return tx, nil
}

func (this *Endpoint) QueryLink(url string) (string, *DspErr) {
	link := this.Dsp.GetLinkFromUrl(url)
	return link, nil
}

func (this *Endpoint) RegisterDns(ip, port string, amount uint64) (string, *DspErr) {
	tx, err := this.Dsp.Chain.Native.Dns.DNSNodeReg([]byte(ip), []byte(port), amount)
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_REGISTER_FAILED, Error: err}
	}
	return hex.EncodeToString(common.ToArrayReverse(tx[:])), nil
}

func (this *Endpoint) UnRegisterDns() (string, *DspErr) {
	tx, err := this.Dsp.Chain.Native.Dns.UnregisterDNSNode()
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_UNREGISTER_FAILED, Error: err}
	}
	return hex.EncodeToString(common.ToArrayReverse(tx[:])), nil
}

func (this *Endpoint) QuitDns() (string, *DspErr) {
	tx, err := this.Dsp.Chain.Native.Dns.QuitNode()
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_QUIT_FAILED, Error: err}
	}
	return hex.EncodeToString(common.ToArrayReverse(tx[:])), nil
}

func (this *Endpoint) AddPos(amount uint64) (string, *DspErr) {
	tx, err := this.Dsp.Chain.Native.Dns.AddInitPos(amount)
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_ADDPOS_FAILED, Error: err}
	}

	return hex.EncodeToString(common.ToArrayReverse(tx[:])), nil
}

func (this *Endpoint) ReducePos(amount uint64) (string, *DspErr) {
	tx, err := this.Dsp.Chain.Native.Dns.ReduceInitPos(amount)
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_REDUCEPOS_FAILED, Error: err}
	}

	return hex.EncodeToString(common.ToArrayReverse(tx[:])), nil
}

func (this *Endpoint) QueryRegInfos() (*dns.PeerPoolMap, *DspErr) {
	ma, err := this.Dsp.Chain.Native.Dns.GetPeerPoolMap()
	if err != nil {
		return nil, &DspErr{Code: DSP_DNS_QUERY_INFOS_FAILED, Error: err}
	}
	return ma, nil
}

func (this *Endpoint) QueryRegInfo(pubkey string) (interface{}, *DspErr) {
	if pubkey == "self" {
		pubkey = ""
	}
	item, err := this.Dsp.Chain.Native.Dns.GetPeerPoolItem(pubkey)
	if err != nil {
		return nil, &DspErr{Code: DSP_DNS_QUERY_INFO_FAILED, Error: err}
	}
	return item, nil
}

func (this *Endpoint) QueryHostInfos() (interface{}, *DspErr) {
	all, err := this.Dsp.Chain.Native.Dns.GetAllDnsNodes()
	if err != nil {
		return nil, &DspErr{Code: DSP_DNS_QUERY_ALLINFOS_FAILED, Error: err}
	}
	return all, nil
}

func (this *Endpoint) QueryHostInfo(addr string) (interface{}, *DspErr) {
	var address common.Address
	if addr != "self" {
		tmpaddr, err := common.AddressFromBase58(addr)
		if err != nil {
			return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
		}
		address = tmpaddr
	}
	info, err := this.Dsp.Chain.Native.Dns.GetDnsNodeByAddr(address)
	if err != nil {
		return nil, &DspErr{Code: DSP_DNS_GET_NODE_BY_ADDR, Error: err}
	}
	return info, nil
}
