package dsp

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"os"

	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
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

func (this *Endpoint) AccountExists() bool {
	if this != nil && this.Dsp != nil && this.Dsp.Account != nil {
		return true
	}
	return false
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
	if this != nil && this.Account != nil {
		return &AccountResp{
			PublicKey: hex.EncodeToString(keypair.SerializePublicKey(this.Account.PublicKey)),
			Address:   this.Account.Address.ToBase58(),
			SigScheme: this.Account.SigScheme,
			Label:     this.AccountLabel,
		}, nil
	}
	if common.FileExisted(config.WalletDatFilePath()) {
		return nil, &DspErr{Code: ACCOUNT_NOT_LOGIN, Error: ErrMaps[ACCOUNT_NOT_LOGIN]}
	}
	return nil, &DspErr{Code: WALLET_FILE_NOT_EXIST, Error: ErrMaps[WALLET_FILE_NOT_EXIST]}
}

func (this *Endpoint) Login(password string) (*AccountResp, *DspErr) {
	service, err := Init(config.WalletDatFilePath(), password)
	if err != nil {
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: err}
	}
	go func() {
		err = StartDspNode(service, true, true, true)
		if err != nil {
			log.Errorf("dsp start err %s", err)
		}
	}()
	return &AccountResp{
		PublicKey: hex.EncodeToString(keypair.SerializePublicKey(service.Account.PublicKey)),
		Address:   service.Account.Address.ToBase58(),
		SigScheme: service.Account.SigScheme,
		Label:     service.AccountLabel,
	}, nil
}

func (this *Endpoint) NewAccount(label string, typeCode keypair.KeyType, curveCode byte, sigScheme s.SignatureScheme, pwd []byte, createOnly bool) (*AccountResp, *DspErr) {
	wallet, err := account.Open(config.WalletDatFilePath())
	if err != nil {
		return nil, &DspErr{Code: WALLET_FILE_NOT_EXIST, Error: err}
	}
	acc, err := wallet.NewAccount(label, typeCode, curveCode, sigScheme, []byte(pwd))
	if err != nil {
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
	go func() {
		err = StartDspNode(service, true, true, true)
		if err != nil {
			log.Errorf("dsp start err %s", err)
		}
	}()
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
	service, err := Init(config.WalletDatFilePath(), password)
	if err != nil {
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: err}
	}
	go func() {
		err = StartDspNode(service, true, true, true)
		if err != nil {
			log.Errorf("dsp start err %s", err)
		}
	}()
	return acc2, nil
}
func (this *Endpoint) ImportWithWalletData(walletStr, password string) (*AccountResp, *DspErr) {
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
	config.Save()
	err = ioutil.WriteFile(config.WalletDatFilePath(), []byte(walletStr), 0666)
	if err != nil {
		return nil, &DspErr{Code: INTERNAL_ERROR, Error: err}
	}
	service, err := Init(config.WalletDatFilePath(), password)
	if err != nil {
		return nil, &DspErr{Code: DSP_INIT_FAILED, Error: err}
	}
	go func() {
		err = StartDspNode(service, true, true, true)
		if err != nil {
			log.Errorf("dsp start err %s", err)
		}
	}()
	return acc2, nil
}

type WalletfileResp struct {
	Wallet string
}
type WIFKeyResp struct {
	PrivateKey string
}

func (this *Endpoint) ExportWIFPrivateKey(password string) (*WIFKeyResp, *DspErr) {
	acc, derr := this.GetAccount(config.WalletDatFilePath(), password)
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
	isExists := common.FileExisted(config.WalletDatFilePath())
	if !isExists || this.Dsp == nil || this.Dsp.Account == nil {
		return &DspErr{Code: NO_DSP, Error: ErrMaps[NO_DSP]}
	}
	syncing, _ := this.IsChannelProcessBlocks()
	if syncing {
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
	this.Account = nil
	this.Dsp.Account = nil
	this.AccountLabel = ""
	this.Dsp.Chain.Native.SetDefaultAccount(nil)
	this.Dsp.Chain.Native.SetDefaultAccount(nil)
	if err != nil {
		return &DspErr{Code: DSP_STOP_FAILED, Error: err}
	}
	DspService = &Endpoint{}
	return nil
}

func (this *Endpoint) CheckPassword(pwd string) *DspErr {
	pwdBuf := sha256.Sum256([]byte(this.Password))
	pwdHash := hex.EncodeToString(pwdBuf[:])
	log.Debugf("CheckPassword: %s, %s, %s", this.Password, pwd, pwdHash)
	if len(pwdHash) != len(pwd) {
		return &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: ErrMaps[ACCOUNT_PASSWORD_WRONG]}
	}
	if pwdHash != pwd {
		return &DspErr{Code: ACCOUNT_PASSWORD_WRONG, Error: ErrMaps[ACCOUNT_PASSWORD_WRONG]}
	}
	return nil
}
