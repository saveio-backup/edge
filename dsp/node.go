package dsp

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/saveio/dsp-go-sdk/utils"
	"github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/smartcontract/service/native/dns"
	fs "github.com/saveio/themis/smartcontract/service/native/savefs"
)

type DspFileUrlPatformType int

//Dsp api
func (this *Endpoint) RegisterNode(addr string, volume, serviceTime uint64) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.RegisterNode(addr, volume, serviceTime)
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", &DspErr{Code: DSP_NODE_REGISTER_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) UnregisterNode() (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.UnregisterNode()
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", &DspErr{Code: DSP_NODE_UNREGISTER_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) NodeQuery(walletAddr string) (*fs.FsNodeInfo, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	info, err := dsp.QueryNode(walletAddr)
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
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.UpdateNode(addr, volume, serviceTime)
	if err != nil {
		log.Errorf("update node err:%s", err)
		return "", &DspErr{Code: DSP_NODE_UPDATE_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

func (this *Endpoint) NodeWithdrawProfit() (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.NodeWithdrawProfit()
	if err != nil {
		log.Errorf("register node err:%s", err)
		return "", &DspErr{Code: DSP_NODE_WITHDRAW_FAILED, Error: err}
	}
	log.Infof("tx: %s", tx)
	return tx, nil
}

//Handle for DNS
func (this *Endpoint) RegisterUrl(url, link string) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	log.Debugf("register url %v link %v", url, link)
	tx, err := dsp.RegisterFileUrl(url, link)
	if err != nil {
		return "", &DspErr{Code: DSP_URL_REGISTER_FAILED, Error: err}
	}
	return tx, nil
}

func (this *Endpoint) BindUrl(url, link string) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.BindFileUrl(url, link)
	if err != nil {
		return "", &DspErr{Code: DSP_URL_BIND_FAILED, Error: err}
	}
	return tx, nil
}

func (this *Endpoint) QueryLink(url string) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	link := dsp.GetLinkFromUrl(url)
	return link, nil
}

func (this *Endpoint) UpdatePluginVersion(urlType uint64, url, fileHash, version, img, title, changeLog string, platformType DspFileUrlPatformType) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	urlVersion := utils.URLVERSION{
		Type:      urlType,
		Url:       url,
		Version:   version,
		FileHash:  fileHash,
		Img:       img,
		Title:     title,
		ChangeLog: changeLog,
		Platform:  int(platformType),
	}
	fileLink := dsp.GetLinkFromUrl(url)
	if len(fileLink) == 0 {
		return "", &DspErr{Code: DSP_GET_FILE_LINK_FAILED, Error: ErrMaps[DSP_GET_FILE_LINK_FAILED]}
	}
	_, err := dsp.GetLinkValues(fileLink)
	if err != nil {
		return "", &DspErr{Code: DSP_GET_FILE_LINK_FAILED, Error: ErrMaps[DSP_GET_FILE_LINK_FAILED]}
	}

	tx, err := dsp.UpdatePluginVersion(urlType, url, fileLink, urlVersion)
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_UPDATE_PLUGIN_INFO_FAILED, Error: err}
	}
	return tx, nil
}

func (this *Endpoint) QueryPluginVersion(url, fileHash string, platformType int) (*utils.URLVERSION, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}

	var pluginVerRet utils.URLVERSION
	var versionStr string
	pluginVer := dsp.GetPluginVersionFromUrl(url)
	if len(pluginVer) == 0 {
		return nil, &DspErr{Code: DSP_DNS_QUERY_PLUGIN_INFO_FAILED, Error: errors.New("url version info not exist")}
	}
	pluginVerArr := strings.Split(pluginVer, utils.PLUGIN_URLVERSION_SPLIT)

	for _, uvItem := range pluginVerArr {
		if len(uvItem) == 0 {
			continue
		}
		var uv utils.URLVERSION
		err := json.Unmarshal([]byte(uvItem), &uv)
		if err != nil {
			return nil, &DspErr{Code: DSP_DNS_QUERY_INFO_FAILED, Error: err}
		}
		if len(fileHash) > 0 && fileHash == uv.FileHash {
			versionStr = uv.Version
			pluginVerRet = uv
		}
		if len(fileHash) == 0 {
			if platformType == 0 || uv.Platform == 0 || platformType == uv.Platform {
				if uv.Version >= versionStr {
					versionStr = uv.Version
					pluginVerRet = uv
				}
			}
		}
	}

	if len(versionStr) == 0 {
		return nil, &DspErr{Code: DSP_DNS_QUERY_PLUGIN_INFO_FAILED, Error: ErrMaps[DSP_DNS_QUERY_PLUGIN_INFO_FAILED]}
	}
	return &pluginVerRet, nil
}

func (this *Endpoint) QueryPluginsInfo() ([]utils.URLVERSION, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	pluginList, err := dsp.QueryPluginsInfo()
	if err != nil {
		return nil, &DspErr{Code: DSP_DNS_QUERY_ALLPLUGININFOS_FAILED, Error: err}
	}

	var pluginsInfo []utils.URLVERSION
	for _, plugin := range pluginList.List {
		pluginVerArr := strings.Split(string(plugin.Desc), utils.PLUGIN_URLVERSION_SPLIT)
		var pluginVerLatest utils.URLVERSION
		var versionStr string

		for _, uvItem := range pluginVerArr {
			if len(uvItem) == 0 {
				continue
			}
			var uv utils.URLVERSION
			err := json.Unmarshal([]byte(uvItem), &uv)
			if err != nil {
				return nil, &DspErr{Code: DSP_DNS_QUERY_INFO_FAILED, Error: err}
			}
			if uv.Version >= versionStr {
				versionStr = uv.Version
				pluginVerFinal = uv
			}
		}
		pluginsInfo = append(pluginsInfo, pluginVerLatest)
	}
	return pluginsInfo, nil
}

func (this *Endpoint) RegisterDns(ip, port string, amount uint64) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.DNSNodeReg([]byte(ip), []byte(port), amount)
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_REGISTER_FAILED, Error: err}
	}
	return tx, nil
}

func (this *Endpoint) UnRegisterDns() (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.UnregisterDNSNode()
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_UNREGISTER_FAILED, Error: err}
	}
	return tx, nil
}

func (this *Endpoint) QuitDns() (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.QuitNode()
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_QUIT_FAILED, Error: err}
	}
	return tx, nil
}

func (this *Endpoint) AddPos(amount uint64) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.AddInitPos(amount)
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_ADDPOS_FAILED, Error: err}
	}

	return tx, nil
}

func (this *Endpoint) ReducePos(amount uint64) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	tx, err := dsp.ReduceInitPos(amount)
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_REDUCEPOS_FAILED, Error: err}
	}

	return tx, nil
}

func (this *Endpoint) QueryRegInfos() (*dns.PeerPoolMap, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	ma, err := dsp.GetPeerPoolMap()
	if err != nil {
		return nil, &DspErr{Code: DSP_DNS_QUERY_INFOS_FAILED, Error: err}
	}
	return ma, nil
}

func (this *Endpoint) QueryRegInfo(pubkey string) (interface{}, *DspErr) {
	if pubkey == "self" {
		pubkey = ""
	}
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	item, err := dsp.GetPeerPoolItem(pubkey)
	if err != nil {
		return nil, &DspErr{Code: DSP_DNS_QUERY_INFO_FAILED, Error: err}
	}
	return item, nil
}

func (this *Endpoint) QueryHostInfos() (interface{}, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	all, err := dsp.GetAllDnsNodes()
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
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	info, err := dsp.GetDnsNodeByAddr(address)
	if err != nil {
		return nil, &DspErr{Code: DSP_DNS_GET_NODE_BY_ADDR, Error: err}
	}
	return info, nil
}

func (this *Endpoint) GetAllOnlineDNS() (map[string]string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return nil, &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	return dsp.GetAllOnlineDNS(), nil
}

func (this *Endpoint) GetExternalIP(walletAddr string) (string, *DspErr) {
	dsp := this.getDsp()
	if dsp == nil {
		return "", &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	hostAddr, err := dsp.GetExternalIP(walletAddr)
	if err != nil {
		return "", &DspErr{Code: DSP_DNS_GET_EXTERNALIP_FAILED, Error: ErrMaps[DSP_DNS_GET_EXTERNALIP_FAILED]}
	}
	return hostAddr, nil
}
