package dsp

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/saveio/dsp-go-sdk/dsp"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/utils"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
	chainCom "github.com/saveio/themis/common"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/crypto/keypair"
	"github.com/saveio/themis/crypto/signature"
	s "github.com/saveio/themis/crypto/signature"
)

type AccountResp struct {
	PrivateKey string
	PublicKey  string
	Address    string
	SigScheme  s.SignatureScheme
	Label      string
	Wallet     string
}

func (this *Endpoint) GetDspAccount() *account.Account {
	if this == nil || this.dspAccLock == nil {
		return nil
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	return this.account
}

func (this *Endpoint) AccountExists() bool {
	return this.dspExist()
}

func (this *Endpoint) GetWalletFilePath() string {
	return config.WalletDatFilePath()
}

func (this *Endpoint) GetAccount(path, password string) (*account.Account, *DspErr) {
	wal, err := wallet.OpenWallet(path)
	if err != nil {
		return nil, &DspErr{Code: WALLET_FILE_NOT_EXIST, Error: err}
	}
	acc, err := wal.GetDefaultAccount([]byte(password))
	if err != nil {
		return nil, &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: err}
	}
	return acc, nil
}

func (this *Endpoint) GetCurrentAccount() (*AccountResp, *DspErr) {
	if this.dspExist() {
		account := this.GetDspAccount()
		if account == nil {
			return nil, &DspErr{Code: ACCOUNT_NOT_LOGIN, Error: ErrMaps[ACCOUNT_NOT_LOGIN]}
		}
		return &AccountResp{
			PublicKey: hex.EncodeToString(keypair.SerializePublicKey(account.PublicKey)),
			Address:   account.Address.ToBase58(),
			SigScheme: account.SigScheme,
			Label:     this.getDspAccountLabel(),
		}, nil
	}
	if common.FileExisted(config.WalletDatFilePath()) {
		wallet, err := account.Open(config.WalletDatFilePath())
		if err != nil {
			return nil, &DspErr{Code: WALLET_FILE_NOT_EXIST, Error: err}
		}
		data := wallet.GetDefaultAccountMetadata()
		if data == nil {
			return nil, &DspErr{Code: WALLET_FILE_NOT_EXIST, Error: err}
		}
		return &AccountResp{
			PublicKey: data.PubKey,
			Address:   data.Address,
			Label:     data.Label,
		}, &DspErr{Code: ACCOUNT_NOT_LOGIN, Error: ErrMaps[ACCOUNT_NOT_LOGIN]}
	}
	return nil, &DspErr{Code: WALLET_FILE_NOT_EXIST, Error: ErrMaps[WALLET_FILE_NOT_EXIST]}
}

func (this *Endpoint) Login(password string) (*AccountResp, *DspErr) {
	if this.dspExist() {
		if this.password != password {
			return nil, &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: ErrMaps[ACCOUNT_PASSWORD_WRONG]}
		}
		account := this.GetDspAccount()
		if account == nil {
			return nil, &DspErr{Code: ACCOUNT_NOT_LOGIN, Error: ErrMaps[ACCOUNT_NOT_LOGIN]}
		}
		return &AccountResp{
			PublicKey: hex.EncodeToString(keypair.SerializePublicKey(account.PublicKey)),
			Address:   account.Address.ToBase58(),
			SigScheme: account.SigScheme,
			Label:     this.getDspAccountLabel(),
		}, nil
	}
	service, err := Init(config.WalletDatFilePath(), password)
	if err != nil {
		if WrongWalletPasswordError(err) {
			return nil, &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: err}
		}
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: err}
	}
	if err := StartDspNode(service, true, true, true); err != nil {
		log.Errorf("dsp start err %s", err)
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: err}
	}
	account := service.GetDspAccount()
	if account == nil {
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: ErrMaps[DSP_INIT_FAILED]}
	}
	return &AccountResp{
		PublicKey: hex.EncodeToString(keypair.SerializePublicKey(account.PublicKey)),
		Address:   account.Address.ToBase58(),
		SigScheme: account.SigScheme,
		Label:     service.getDspAccountLabel(),
	}, nil
}

func (this *Endpoint) NewAccount(label string, typeCode keypair.KeyType, curveCode byte, sigScheme s.SignatureScheme, pwd []byte, createOnly bool) (*AccountResp, *DspErr) {
	wallet, err := account.Open(config.WalletDatFilePath())
	if err != nil {
		return nil, &DspErr{Code: WALLET_FILE_NOT_EXIST, Error: err}
	}
	if wallet.GetAccountNum() > 0 {
		existMeta := wallet.GetAccountMetadataByLabel(label)
		if existMeta != nil {
			log.Debugf("exist same label for addr %v, label %s", existMeta.Address, label)
			if err := wallet.SetLabel(existMeta.Address, existMeta.Address); err != nil {
				return nil, &DspErr{Code: CREATE_ACCOUNT_FAILED, Error: err}
			}
		}
	}
	acc, err := wallet.NewAccount(label, typeCode, curveCode, sigScheme, []byte(pwd))
	if err != nil {
		log.Debugf("err %v", err)
		return nil, &DspErr{Code: CREATE_ACCOUNT_FAILED, Error: err}
	}
	key, err := keypair.Key2WIF(acc.PrivateKey)
	if err != nil {
		return nil, &DspErr{Code: CREATE_ACCOUNT_FAILED, Error: err}
	}
	acc2 := &AccountResp{
		PrivateKey: string(key),
		Label:      label,
	}
	if createOnly {
		config.Save()
		data, err := ioutil.ReadFile(config.WalletDatFilePath())
		if err != nil {
			return nil, &DspErr{Code: ACCOUNT_EXPORT_FAILED, Error: err}
		}
		acc2.Wallet = string(data)
		os.Remove(config.WalletDatFilePath())
		return acc2, nil
	}
	acc2 = &AccountResp{
		PublicKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Address:   acc.Address.ToBase58(),
		SigScheme: acc.SigScheme,
		Label:     label,
	}
	service, err := Init(config.WalletDatFilePath(), string(pwd))
	log.Debugf("ini DspService at new account:%v\n", service)
	if err != nil {
		log.Errorf("dsp init err %s", err)
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: err}
	}
	if err := StartDspNode(service, true, true, true); err != nil {
		log.Errorf("dsp start err %s", err)
	}
	config.Save()
	return acc2, nil
}

func (this *Endpoint) ImportWithPrivateKey(wif, label, password string) (*AccountResp, *DspErr) {
	privateKey, err := keypair.WIF2Key([]byte(wif))
	if err != nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	publicKey := privateKey.Public()
	addr := types.AddressFromPubKey(publicKey)
	k, err := keypair.EncryptPrivateKey(privateKey, addr.ToBase58(), []byte(password))
	if err != nil {
		return nil, &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: err}
	}
	if common.FileExisted(config.WalletDatFilePath()) {
		if err := os.RemoveAll(config.WalletDatFilePath()); err != nil {
			return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
	}
	wallet, err := account.Open(config.WalletDatFilePath())
	if err != nil {
		return nil, &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: err}
	}
	var accMeta account.AccountMetadata
	accMeta.Address = k.Address
	accMeta.KeyType = k.Alg
	accMeta.EncAlg = k.EncAlg
	accMeta.Hash = k.Hash
	accMeta.Key = k.Key
	accMeta.Curve = k.Param["curve"]
	accMeta.Salt = k.Salt
	accMeta.Label = label
	accMeta.PubKey = hex.EncodeToString(keypair.SerializePublicKey(privateKey.Public()))
	accMeta.SigSch = signature.SHA256withECDSA.Name()
	err = wallet.ImportAccount(&accMeta)
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	acc := &account.Account{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    addr,
	}
	acc2 := &AccountResp{
		PrivateKey: wif,
		PublicKey:  hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Address:    acc.Address.ToBase58(),
		SigScheme:  signature.SHA256withECDSA,
		Label:      label,
	}
	config.Save()
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	if this.dspExist() {
		account := this.GetDspAccount()
		if account == nil {
			return nil, &DspErr{Code: ACCOUNT_NOT_LOGIN, Error: ErrMaps[ACCOUNT_NOT_LOGIN]}
		}
		if account.Address.ToBase58() != acc.Address.ToBase58() {
			return nil, &DspErr{Code: ACCOUNT_EXIST, Error: ErrMaps[ACCOUNT_EXIST]}
		}
		return acc2, nil
	}
	service, err := Init(config.WalletDatFilePath(), password)
	if err != nil {
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: err}
	}
	if err := StartDspNode(service, true, true, true); err != nil {
		log.Errorf("dsp start err %s", err)
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: err}
	}
	return acc2, nil
}
func (this *Endpoint) ImportWithWalletData(walletStr, password, walletPath string) (*AccountResp, *DspErr) {
	wal, err := wallet.OpenWithWalletData([]byte(walletStr))
	if err != nil || wal == nil {
		return nil, &DspErr{Code: INVALID_PARAMS, Error: err}
	}
	accData, err := wal.GetDefaultAccountData()
	if err != nil {
		return nil, &DspErr{Code: ACCOUNTDATA_NOT_EXIST, Error: err}
	}
	acc, err := wal.GetDefaultAccount([]byte(password))
	if err != nil {
		return nil, &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: err}
	}
	acc2 := &AccountResp{
		PublicKey: hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Address:   acc.Address.ToBase58(),
		SigScheme: acc.SigScheme,
		Label:     accData.Label,
	}
	if len(walletPath) > 0 {
		if common.IsAbsPath(walletPath) {
			log.Debugf("is abs path %v", walletPath)
			config.Parameters.BaseConfig.WalletDir = walletPath
		} else {
			base, err := filepath.Abs(".")
			if err != nil {
				return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
			}
			fullPath := filepath.Join(base, walletPath)
			relPath, err := filepath.Rel(config.Parameters.BaseConfig.BaseDir, fullPath)
			log.Debugf("not abs path %v, base %v, rel %v", walletPath, base, relPath)
			if err != nil {
				return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
			}
			config.Parameters.BaseConfig.WalletDir = relPath
		}
	}
	config.Save()
	if len(walletPath) == 0 {
		err = ioutil.WriteFile(config.WalletDatFilePath(), []byte(walletStr), 0666)
		if err != nil {
			return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
	}
	if this.dspExist() {
		account := this.GetDspAccount()
		if account == nil {
			return nil, &DspErr{Code: ACCOUNT_NOT_LOGIN, Error: ErrMaps[ACCOUNT_NOT_LOGIN]}
		}
		if account.Address.ToBase58() != acc.Address.ToBase58() {
			return nil, &DspErr{Code: ACCOUNT_EXIST, Error: ErrMaps[ACCOUNT_EXIST]}
		}
		return acc2, nil
	}
	service, err := Init(config.WalletDatFilePath(), password)
	if err != nil {
		if WrongWalletPasswordError(err) {
			return nil, &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: err}
		}
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: err}
	}

	if err := StartDspNode(service, true, true, true); err != nil {
		log.Errorf("dsp start err %s", err)
		return nil, &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: err}
	}
	return acc2, nil
}

type WalletfileResp struct {
	Wallet string
}
type WIFKeyResp struct {
	PrivateKey string
}

func (this *Endpoint) ExportWIFPrivateKey() (*WIFKeyResp, *DspErr) {
	acc, derr := this.GetAccount(config.WalletDatFilePath(), this.getDspPassword())
	if derr != nil {
		return nil, derr
	}
	key, err := keypair.Key2WIF(acc.PrivateKey)
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	return &WIFKeyResp{PrivateKey: string(key)}, nil
}

func (this *Endpoint) ExportWalletFile() (*WalletfileResp, *DspErr) {
	data, err := ioutil.ReadFile(config.WalletDatFilePath())
	if err != nil {
		return nil, &DspErr{Code: ACCOUNT_EXPORT_FAILED, Error: err}
	}
	return &WalletfileResp{Wallet: string(data)}, nil
}

func (this *Endpoint) Logout() *DspErr {
	log.Debugf("Logout ++++")
	isExists := common.FileExisted(config.WalletDatFilePath())
	if !isExists || !this.dspExist() {
		log.Debugf("logout of no wallet dat files")
		if this != nil {
			this.cleanDspAccount()
			this.notifyAccountLogout()
			log.Debugf("notify user logout")
		}
		if isExists {
			err := os.Remove(config.WalletDatFilePath())
			if err != nil {
				return &DspErr{Code: INTERNAL_ERROR, Error: err}
			}
		}
		DspService = &Endpoint{}
		return nil
	}
	dsp := this.getDsp()
	syncing, _ := this.IsChannelProcessBlocks()
	if syncing || (dsp != nil && dsp.HasChannelInstance() && dsp.ChannelFirstSyncing()) {
		return &DspErr{Code: DSP_CHANNEL_SYNCING, Error: ErrMaps[DSP_CHANNEL_SYNCING]}
	}
	if isExists {
		err := os.Remove(config.WalletDatFilePath())
		if err != nil {
			return &DspErr{Code: INTERNAL_ERROR, Error: err}
		}
	}
	// TODO: justify whether account exists
	err := this.Stop()
	if err != nil {
		return &DspErr{Code: DSP_STOP_FAILED, Error: err}
	}
	this.cleanDspAccount()
	this.notifyAccountLogout()
	log.Debugf("notify user logout")
	DspService = &Endpoint{}
	return nil
}

func (this *Endpoint) CheckPassword(pwd string) *DspErr {
	pwdHash := utils.Sha256HexStr(this.getDspPassword())
	log.Debugf("CheckPassword: %s, %s, %s", this.getDspPassword(), pwd, pwdHash)
	if len(pwdHash) != len(pwd) {
		return &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: ErrMaps[ACCOUNT_PASSWORD_WRONG]}
	}
	if pwdHash != pwd {
		return &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: ErrMaps[ACCOUNT_PASSWORD_WRONG]}
	}
	return nil
}

func (this *Endpoint) dspExist() bool {
	if this == nil || this.dspAccLock == nil {
		return false
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	if this.dsp != nil && this.account != nil {
		return true
	}
	return false
}

func (this *Endpoint) getDsp() *dsp.Dsp {
	if this == nil || this.dspAccLock == nil {
		return nil
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	return this.dsp
}

func (this *Endpoint) setDsp(dsp *dsp.Dsp) {
	if this == nil || this.dspAccLock == nil {
		return
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	this.dsp = dsp
}

func (this *Endpoint) getDspAccountLabel() string {
	if this == nil || this.dspAccLock == nil {
		return ""
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	if this.account != nil {
		return this.accountLabel
	}
	return ""
}

func (this *Endpoint) getDspPassword() string {
	if this == nil || this.dspAccLock == nil {
		return ""
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	if this.account != nil {
		return this.password
	}
	return ""
}

func (this *Endpoint) cleanDspAccount() {
	if this == nil || this.dspAccLock == nil {
		return
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	this.accountLabel = ""
	if this.account != nil {
		this.account = nil
	}
	if this.dsp != nil {
		this.dsp.SetAccount(nil)
	}
}

func (this *Endpoint) getDspWalletAddr() chainCom.Address {
	if this == nil || this.dspAccLock == nil {
		return chainCom.ADDRESS_EMPTY
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	if this.account == nil {
		return chainCom.ADDRESS_EMPTY
	}
	return this.account.Address
}

func (this *Endpoint) getDspWalletAddress() string {
	if this == nil || this.dspAccLock == nil {
		return ""
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	if this.account == nil {
		return ""
	}
	return this.account.Address.ToBase58()
}

func (this *Endpoint) setDspAccount(acc *account.Account, accLabel, pwd string) {
	if this == nil || this.dspAccLock == nil {
		return
	}
	this.dspAccLock.Lock()
	defer this.dspAccLock.Unlock()
	this.account = acc
	this.accountLabel = accLabel
	this.password = pwd
}
