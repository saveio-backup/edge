package dsp

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	dsp "github.com/saveio/dsp-go-sdk"
	dspCom "github.com/saveio/dsp-go-sdk/common"
	dspCfg "github.com/saveio/dsp-go-sdk/config"
	"github.com/saveio/dsp-go-sdk/store"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp/storage"
	chanCom "github.com/saveio/pylons/common"
	saveSdk "github.com/saveio/themis-go-sdk"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common/log"
	fs "github.com/saveio/themis/smartcontract/service/native/onifs"
)

type Endpoint struct {
	Account  *account.Account
	Chain    *saveSdk.Chain
	Dsp      *dsp.Dsp
	progress sync.Map
	db       *store.LevelDBStore
	sqliteDB *storage.SQLiteStorage
	closhCh  chan struct{}
}

func Init(walletDir, pwd string) (*Endpoint, error) {
	this := &Endpoint{
		closhCh: make(chan struct{}, 1),
	}
	log.Debugf("walletDir: %s, %d", walletDir, len(walletDir))
	chain := saveSdk.NewChain()
	chain.NewRestClient().SetAddress(config.Parameters.BaseConfig.ChainRestAddr)
	this.Chain = chain
	if len(walletDir) == 0 {
		return this, nil
	}
	_, err := os.Open(walletDir)
	if err != nil {
		return this, nil
	}
	wallet, err := wallet.OpenWallet(walletDir)
	if err != nil {
		log.Error("Client open wallet Error, msg:", err)
		return nil, err
	}
	this.Account, err = wallet.GetDefaultAccount([]byte(pwd))
	if err != nil {
		log.Error("Client get default account Error, msg:", err)
		return nil, err
	}
	this.Chain.SetDefaultAccount(this.Account)

	db, err := store.NewLevelDBStore(config.ClientLevelDBPath())
	if err != nil {
		log.Errorf("NewLevelDBStore err %s, config.ClientLevelDBPath():%v", err, config.ClientLevelDBPath())
		return nil, err
	}
	this.db = db
	sqliteDB, err := storage.NewSQLiteStorage(config.ClientSqliteDBPath())
	if err != nil {
		log.Errorf("sqlite err %s", err)
		return nil, err
	}
	this.sqliteDB = sqliteDB
	log.Debug("Endpoint init successed")
	return this, nil
}

func StartDspNode(endpoint *Endpoint, startListen, startShare, startChannel bool) error {
	// [TODO]: use NAT to find out IP
	channelListenAddr := fmt.Sprintf("127.0.0.1:%d", int(config.Parameters.BaseConfig.PortBase+uint32(config.Parameters.BaseConfig.ChannelPortOffset)))
	dspConfig := &dspCfg.DspConfig{
		DBPath:               config.DspDBPath(),
		FsRepoRoot:           config.FsRepoRootPath(),
		FsFileRoot:           config.FsFileRootPath(),
		FsType:               dspCfg.FSType(config.Parameters.FsConfig.FsType),
		ChainRpcAddr:         config.Parameters.BaseConfig.ChainRpcAddr,
		ChannelClientType:    config.Parameters.BaseConfig.ChannelClientType,
		ChannelListenAddr:    channelListenAddr,
		ChannelProtocol:      config.Parameters.BaseConfig.ChannelProtocol,
		ChannelRevealTimeout: config.Parameters.BaseConfig.ChannelRevealTimeout,
		ChannelDBPath:        config.ChannelDBPath(),
		DnsNodeMaxNum:        config.Parameters.BaseConfig.DnsNodeMaxNum,
		SeedInterval:         config.Parameters.BaseConfig.SeedInterval,
		DnsChannelDeposit:    config.Parameters.BaseConfig.DnsChannelDeposit,
	}
	log.Debugf("dspConfig.dbpath %v", dspConfig.DBPath)
	log.Debugf("dspConfig.FsRepoRoot %v", dspConfig.FsRepoRoot)
	log.Debugf("dspConfig.ChannelDBPath %v", dspConfig.ChannelDBPath)
	log.Debugf("WalletDatFilePath %v", config.WalletDatFilePath())

	//Skip init fs if Dsp doesn't start listen
	if !startListen {
		dspConfig.FsRepoRoot = ""
	}
	if !startChannel {
		dspConfig.ChannelListenAddr = ""
	}

	log.Debugf("dspConfig.ChannelListenAddr :%s ,acc %v\n", dspConfig.ChannelListenAddr, endpoint.Account)
	dspSrv := dsp.NewDsp(dspConfig, endpoint.Account)
	if dspSrv == nil {
		return errors.New("dsp server init failed")
	}
	log.Debug("Create Dsp success")

	if startListen {
		dspListenPort := int(config.Parameters.BaseConfig.PortBase + uint32(config.Parameters.BaseConfig.DspPortOffset))
		if dspListenPort == 0 {
			log.Fatal("Not configure HttpRestPort port ")
			return nil
		}

		//[TODO] dsp-go-sdk may only need port and construct address itself
		dspListenAddr := config.Parameters.BaseConfig.DspProtocol + "://127.0.0.1:" + strconv.Itoa(dspListenPort)
		log.Debugf("start dsp at %s", dspListenAddr)
		err := dspSrv.Start(dspListenAddr)
		if err != nil {
			return err
		}
		dspSrv.RegNodeEndpoint(dspSrv.CurrentAccount().Address, dspConfig.ChannelProtocol+"://"+dspConfig.ChannelListenAddr)
		if startShare {
			//[TODO] price needed to be discuss
			dspSrv.SetUnitPriceForAllFile(dspCom.ASSET_USDT, common.DSP_DOWNLOAD_UNIT_PRICE)
			go dspSrv.StartShareServices()
		}
	}

	// start channel only
	if !startListen && startChannel {
		err := dspSrv.StartChannelService()
		if err != nil {
			return err
		}
	}
	log.Info("Dsp start successed.")
	endpoint.Dsp = dspSrv
	if startListen {
		go endpoint.RegisterProgressCh()
		go endpoint.RegisterShareNotificationCh()
	}
	// testWalletAddr := "AGGTaoJ8Ygim7zVi5ZZqrXy8EQqgNQJxYx"
	// if dspSrv.WalletAddress() != testWalletAddr {
	// 	log.Warn("Start test+++++++++++=")
	// 	// TEST
	// 	amount := uint64(1)
	// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 	paymentId := r.Int31()
	// 	log.Debugf("paying to , id %v", paymentId)
	// 	dspSrv.Channel.SetHostAddr(testWalletAddr, "tcp://127.0.0.1:14001")
	// 	err := dspSrv.Channel.MediaTransfer(paymentId, amount, testWalletAddr)
	// 	if err != nil {
	// 		log.Debugf("payingmentid %d, failed err %s", paymentId, err)
	// 	} else {
	// 		log.Debugf("payment id %d,  price:%d success", paymentId, amount)
	// 	}
	// 	os.Exit(1)
	// }
	return nil
}

func StartDdnsNode(endpoint *Endpoint) {

}

// Stop. stop endpoint instance
func (this *Endpoint) Stop() error {
	close(this.closhCh)
	err := this.db.Close()
	if err != nil {
		return err
	}
	err = this.sqliteDB.Close()
	if err != nil {
		return err
	}
	return this.Dsp.Stop()
}

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

//pylons api
func (this *Endpoint) OpenPaymentChannel(partnerAddr string) (chanCom.ChannelID, error) {
	return this.Dsp.Channel.OpenChannel(partnerAddr)
}

func (this *Endpoint) ClosePaymentChannel(regAddr, tokenAddr, partnerAddr string, retryTimeout float64) {
	//[TODO] call channel close function of dsp-go-sdk
	return
}

func (this *Endpoint) DepositToChannel(partnerAddr string, totalDeposit uint64) error {
	return this.Dsp.Channel.SetDeposit(partnerAddr, totalDeposit)
}

func (this *Endpoint) Transfer(paymentId int32, amount uint64, to string) error {
	return this.Dsp.Channel.DirectTransfer(paymentId, amount, to)
}

func (this *Endpoint) GetChannelListByOwnerAddress(addr string, tokenAddr string) *list.List {
	//[TODO] call dsp-go-sdk function to return channel list
	//[NOTE] addr and token Addr should NOT be needed. addr mean PaymentNetworkID
	//tokenAddr mean TokenAddress. Need comfirm the behavior when integrate dsp-go-sdk with pylons
	return list.New()
}

func (this *Endpoint) QuerySpecialChannelDeposit(partnerAddr string) (uint64, error) {
	return this.Dsp.Channel.GetTotalDepositBalance(partnerAddr)
}

func (this *Endpoint) QuerySpecialChannelAvaliable(partnerAddr string) (uint64, error) {
	return this.Dsp.Channel.GetAvaliableBalance(partnerAddr)
}

// ChannelWithdraw. withdraw amount of asset from channel
func (this *Endpoint) ChannelWithdraw(partnerAddr string, amount uint64) error {
	totalWithdraw, err := this.Dsp.Channel.GetTotalWithdraw(partnerAddr)
	if err != nil {
		return err
	}
	bal := amount + totalWithdraw
	if bal-amount != totalWithdraw {
		return errors.New("withdraw overflow")
	}
	success, err := this.Dsp.Channel.Withdraw(partnerAddr, bal)
	if err != nil {
		return err
	}
	if !success {
		return errors.New("withdraw failed")
	}
	return nil
}

// ChannelCooperativeSettle. settle channel cooperatively
func (this *Endpoint) ChannelCooperativeSettle(partnerAddr string) error {
	return this.Dsp.Channel.CooperativeSettle(partnerAddr)
}
