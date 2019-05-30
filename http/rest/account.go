package rest

import (
	"encoding/hex"
	"io/ioutil"
	"os"

	edgeCmd "github.com/saveio/edge/cmd"
	"github.com/saveio/edge/common"
	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/dsp"
	berr "github.com/saveio/edge/http/base/error"
	sdk "github.com/saveio/themis-go-sdk"
	"github.com/saveio/themis-go-sdk/wallet"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common/log"
	"github.com/saveio/themis/core/types"
	"github.com/saveio/themis/crypto/keypair"
	s "github.com/saveio/themis/crypto/signature"
)

func GetCurrentAccount(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	wal, err := wallet.OpenWallet(config.WalletDatFilePath())
	if err != nil {
		return ResponsePack(berr.OPEN_WALLET_ERROR)
	}
	accData, err := wal.GetDefaultAccountData()
	if err != nil {
		return ResponsePack(berr.DEFAULT_ACCOUNT_NOT_FOUND)
	}
	acc, err := wal.GetDefaultAccount([]byte(config.Parameters.BaseConfig.WalletPwd))
	if err != nil {
		return ResponsePack(berr.WRONG_PASSWORD)
	}
	acc2 := struct {
		PrivateKey string
		PublicKey  string
		Address    string
		SigScheme  s.SignatureScheme
		Label      string
	}{
		PrivateKey: hex.EncodeToString(keypair.SerializePrivateKey(acc.PrivateKey)),
		PublicKey:  hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Address:    acc.Address.ToBase58(),
		SigScheme:  acc.SigScheme,
		Label:      accData.Label,
	}
	resp["Result"] = acc2
	return resp
}

func NewAccount(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("NewAccount cmd %v", cmd)
	resp := ResponsePack(berr.SUCCESS)
	if DspService != nil && DspService.Dsp != nil && DspService.Dsp.Account != nil {
		return ResponsePack(berr.ACCOUNT_HAS_EXISTS)
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	label, ok := cmd["Label"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	optionType, ok := cmd["KeyType"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	optionCurve, ok := cmd["Curve"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	optionScheme, ok := cmd["Scheme"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	wallet, err := account.Open(config.WalletDatFilePath())
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}

	keyType := edgeCmd.GetKeyTypeCode(optionType)
	curve := edgeCmd.GetCurveCode(optionCurve)
	scheme := edgeCmd.GetSchemeCode(optionScheme)
	acc, err := wallet.NewAccount(label, keyType, curve, scheme, []byte(password))
	if err != nil {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	acc2 := struct {
		PrivateKey string
		PublicKey  string
		Address    string
		SigScheme  s.SignatureScheme
		Label      string
	}{
		PrivateKey: hex.EncodeToString(keypair.SerializePrivateKey(acc.PrivateKey)),
		PublicKey:  hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Address:    acc.Address.ToBase58(),
		SigScheme:  acc.SigScheme,
		Label:      label,
	}

	resp["Result"] = acc2
	DspService, err = dsp.Init(config.WalletDatFilePath(), password)
	log.Debugf("ini DspService at new account:%v\n", DspService)
	if err != nil {
		log.Errorf("dsp init err %s", err)
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	err = dsp.StartDspNode(DspService, true, true, true)
	if err != nil {
		log.Errorf("dsp start err %s", err)
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	config.Parameters.BaseConfig.WalletPwd = password
	config.Save()
	return resp
}

func ImportWithPrivateKey(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	if DspService != nil && DspService.Dsp.Account != nil {
		return ResponsePack(berr.ACCOUNT_HAS_EXISTS)
	}
	privKeyStr, ok := cmd["PrivateKey"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	privKeyBuf, err := hex.DecodeString(privKeyStr)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	privateKey, err := keypair.DeserializePrivateKey(privKeyBuf)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	publicKey := privateKey.Public()
	addr := types.AddressFromPubKey(publicKey)
	acc := &account.Account{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    addr,
	}

	config.Parameters.BaseConfig.WalletPwd = password
	config.Save()

	acc2 := struct {
		PrivateKey string
		PublicKey  string
		Address    string
		SigScheme  s.SignatureScheme
	}{
		PrivateKey: hex.EncodeToString(keypair.SerializePrivateKey(acc.PrivateKey)),
		PublicKey:  hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Address:    acc.Address.ToBase58(),
		SigScheme:  acc.SigScheme,
	}
	// TODO: save acc2 to wallet.dat
	resp["Result"] = acc2

	err = startDspService(acc)
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	return resp
}

func ImportWithWalletData(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	if DspService != nil && DspService.Dsp.Account != nil {
		return ResponsePack(berr.ACCOUNT_HAS_EXISTS)
	}
	walletStr, ok := cmd["Wallet"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	password, ok := cmd["Password"].(string)
	if !ok {
		return ResponsePack(berr.INVALID_PARAMS)
	}
	wal, err := wallet.OpenWithWalletData([]byte(walletStr))
	if err != nil || wal == nil {
		log.Errorf("wal %v, err %s, walletStr %s", wal, err, walletStr)
		return ResponsePack(berr.INVALID_WALLET_DATA)
	}
	accData, err := wal.GetDefaultAccountData()
	if err != nil {
		return ResponsePack(berr.DEFAULT_ACCOUNT_NOT_FOUND)
	}
	acc, err := wal.GetDefaultAccount([]byte(password))
	if err != nil {
		return ResponsePack(berr.WRONG_PASSWORD)
	}
	acc2 := struct {
		PrivateKey string
		PublicKey  string
		Address    string
		SigScheme  s.SignatureScheme
		Label      string
	}{
		PrivateKey: hex.EncodeToString(keypair.SerializePrivateKey(acc.PrivateKey)),
		PublicKey:  hex.EncodeToString(keypair.SerializePublicKey(acc.PublicKey)),
		Address:    acc.Address.ToBase58(),
		SigScheme:  acc.SigScheme,
		Label:      accData.Label,
	}
	config.Parameters.BaseConfig.WalletPwd = password
	config.Save()
	err = ioutil.WriteFile(config.WalletDatFilePath(), []byte(walletStr), 0666)
	if err != nil {
		return ResponsePack(berr.WALLET_SAVE_FAILED)
	}
	DspService, err = dsp.Init(config.WalletDatFilePath(), password)
	if err != nil {
		return ResponsePack(berr.DSP_INIT_ERROR)
	}
	err = dsp.StartDspNode(DspService, true, true, true)
	if err != nil {
		return ResponsePack(berr.DSP_START_ERROR)
	}
	resp["Result"] = acc2
	return resp
}

func ExportWalletFile(cmd map[string]interface{}) map[string]interface{} {
	resp := ResponsePack(berr.SUCCESS)
	data, err := ioutil.ReadFile(config.WalletDatFilePath())
	if err != nil {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	type walletfileResp struct {
		Wallet string
	}
	resp["Result"] = &walletfileResp{
		Wallet: string(data),
	}
	return resp
}

// Logout. logout current account
func Logout(cmd map[string]interface{}) map[string]interface{} {
	log.Debugf("Logout")
	resp := ResponsePack(berr.SUCCESS)
	isExists := common.FileExisted(config.WalletDatFilePath())
	if !isExists && (DspService.Dsp.CurrentAccount() != nil && DspService.Dsp.Account != nil) {
		return ResponsePack(berr.INTERNAL_ERROR)
	}
	if isExists {
		err := os.Remove(config.WalletDatFilePath())
		if err != nil {
			return ResponsePack(berr.LOGOUT_DELETE_WALLET_ERROR)
		}
	}
	// TODO: justify whether account exists
	err := DspService.Stop()
	DspService.Dsp.Chain.Native.SetDefaultAccount(nil)
	DspService.Dsp.Chain.Native.SetDefaultAccount(nil)
	log.Debugf("after logout DspService:%v, %v", DspService, DspService.Dsp.Account)
	if err != nil {
		return ResponsePack(berr.DSP_STOP_ERROR)
	}
	return resp
}

func startDspService(acc *account.Account) error {
	DspService = &dsp.Endpoint{}
	DspService.Account = acc
	chain := sdk.NewChain()
	chain.NewRestClient().SetAddress(config.Parameters.BaseConfig.ChainRestAddr)
	DspService.Dsp.Chain = chain
	DspService.Dsp.Chain.SetDefaultAccount(acc)
	return dsp.StartDspNode(DspService, true, true, true)
}
