package rest

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/saveio/edge/common/config"
	berr "github.com/saveio/edge/http/base/error"
	"github.com/saveio/themis/common/log"
)

type handler func(map[string]interface{}) map[string]interface{}
type Action struct {
	sync.RWMutex
	name    string
	handler handler
}
type restServer struct {
	router   *Router
	listener net.Listener
	server   *http.Server
	postMap  map[string]Action //post method map
	getMap   map[string]Action //get method map
}

const (
	GET_BLK_TXS_BY_HEIGHT = "/api/v1/block/transactions/height/:height"
	GET_BLK_BY_HEIGHT     = "/api/v1/block/details/height/:height"
	GET_BLK_BY_HASH       = "/api/v1/block/details/hash/:hash"
	GET_BLK_HEIGHT        = "/api/v1/block/height"
	GET_BLK_HASH          = "/api/v1/block/hash/:height"
	GET_TX                = "/api/v1/transaction/:hash"
	GET_TXS_HEIGHT_LIMIT  = "/api/v1/transactions/:addr/:type"
	GET_STORAGE           = "/api/v1/storage/:hash/:key"
	GET_BALANCE           = "/api/v1/balance/:addr"
	GET_CONTRACT_STATE    = "/api/v1/contract/:hash"
	GET_SMTCOCE_EVT_TXS   = "/api/v1/smartcode/event/transactions/:height"
	GET_SMTCOCE_EVTS      = "/api/v1/smartcode/event/txhash/:hash"
	GET_BLK_HGT_BY_TXHASH = "/api/v1/block/height/txhash/:hash"
	GET_MERKLE_PROOF      = "/api/v1/merkleproof/:hash"
	GET_GAS_PRICE         = "/api/v1/gasprice"
	GET_ALLOWANCE         = "/api/v1/allowance/:asset/:from/:to"
	GET_UNBOUNDONG        = "/api/v1/unboundong/:addr"
	GET_GRANTONG          = "/api/v1/grantong/:addr"
	GET_MEMPOOL_TXCOUNT   = "/api/v1/mempool/txcount"
	GET_MEMPOOL_TXSTATE   = "/api/v1/mempool/txstate/:hash"
	GET_VERSION           = "/api/v1/version"
	GET_NETWORKID         = "/api/v1/networkid"

	GET_CURRENT_ACCOUNT            = "/api/v1/account"
	NEW_ACCOUNT                    = "/api/v1/account"
	LOGOUT_ACCOUNT                 = "/api/v1/account/logout"
	IMPORT_ACCOUNT_WITH_PRIVATEKEY = "/api/v1/account/import/privatekey"
	IMPORT_ACCOUNT_WITH_WALLETFILE = "/api/v1/account/import/walletfile"
	EXPORT_WALLETFILE              = "/api/v1/account/export/walletfile"
	EXPORT_WIFPRIVATEKEY           = "/api/v1/account/export/privatekey"

	ASSET_TRANSFER_DIRECT = "/api/v1/asset/transfer/direct"

	SET_CONFIG = "/api/v1/config"

	DSP_NODE_REGISTER         = "/api/v1/dsp/node/register"
	DSP_NODE_UNREGISTER       = "/api/v1/dsp/node/unregister"
	DSP_NODE_QUERY            = "/api/v1/dsp/node/query/:addr"
	DSP_NODE_UPDATE           = "/api/v1/dsp/node/update"
	DSP_NODE_WITHDRAW         = "/api/v1/dsp/node/withdraw"
	DSP_SET_USER_SPACE        = "/api/v1/dsp/client/userspace/set"
	DSP_CLIENT_GET_USER_SPACE = "/api/v1/dsp/client/userspace/:addr"
	DSP_USERSPACE_RECORDS     = "/api/v1/dsp/client/userspacerecords/:addr/:offset/:limit"

	DSP_GET_UPLOAD_FILELIST   = "/api/v1/dsp/file/uploadlist/:type/:offset/:limit"
	DSP_GET_DOWNLOAD_FILELIST = "/api/v1/dsp/file/downloadlist/:type/:offset/:limit"
	DSP_GET_FILEINFO          = "/api/v1/dsp/file/:hash"
	DSP_GET_FILE_TRANSFERLIST = "/api/v1/dsp/file/transferlist/:type/:offset/:limit"
	DSP_FILE_UPLOAD           = "/api/v1/dsp/file/upload"
	DSP_FILE_UPLOAD_FEE       = "/api/v1/dsp/file/uploadfee/:file"
	DSP_FILE_DELETE           = "/api/v1/dsp/file/delete"
	DSP_FILE_DOWNLOAD         = "/api/v1/dsp/file/download"
	DSP_FILE_DOWNLOAD_INFO    = "/api/v1/dsp/file/downloadinfo/:url"
	DSP_FILE_ENCRYPT          = "/api/v1/dsp/file/encrypt"
	DSP_FILE_DECRYPT          = "/api/v1/dsp/file/decrypt"
	DSP_FILE_SHARE_INCOME     = "/api/v1/dsp/file/share/income/:begin/:end/:offset/:limit"
	DSP_FILE_SHARE_REVENUE    = "/api/v1/dsp/file/share/revenue"
	DSP_GET_FILE_WHITELIST    = "/api/v1/dsp/file/whitelist/:hash"
	DSP_UPDATE_FILE_WHITELIST = "/api/v1/dsp/file/updatewhitelist"

	GET_CHANNEL_INIT_PROGRESS = "/api/v1/channel/init/progress"
	GET_ALL_CHANNEL           = "/api/v1/channel"
	OPEN_CHANNEL              = "/api/v1/channel/open/:partneraddr"
	DEPOSIT_CHANNEL           = "/api/v1/channel/deposit"
	WITHDRAW_CHANNEL          = "/api/v1/channel/withdraw"
	TRANSFER_BY_CHANNEL       = "/api/v1/channel/transfer/:toaddr/:amount/:paymentid"
	QUERY_CHANNEL_DEPOSIT     = "/api/v1/channel/query/deposit/:partneraddr"
	QUERY_CHANNEL             = "/api/v1/channel/query/detail/:partneraddr"
	QUERY_CHANNEL_BY_ID       = "/api/v1/channel/query/:id"

	DNS_REGISTER         = "/api/v1/dns/register"
	DNS_BIND             = "/api/v1/dns/bind"
	DNS_QUERYLINK        = "/api/v1/dns/query/link"
	DNS_REGISTER_DNS     = "/api/v1/dns/registerdns/:ip/:port/:deposit"
	DNS_UNREGISTER_DNS   = "/api/v1/dns/unregisterdns"
	DNS_QUIT             = "/api/v1/dns/quit"
	DNS_ADD_DEPOSIT      = "/api/v1/dns/addpos/:amount"
	DNS_REDUCE_DEPOSIT   = "/api/v1/dns/reducepos/:amount"
	DNS_QUERY_REG_INFOS  = "/api/v1/dns/reginfos"
	DNS_QUERY_HOST_INFOS = "/api/v1/dns/hostinfos"
	DNS_QUERY_REG_INFO   = "/api/v1/dns/reginfo/:pubkey"
	DNS_QUERY_HOST_INFO  = "/api/v1/dns/hostinfo/:addr"
)

//init restful server
func InitRestServer() ApiServer {
	rt := &restServer{}

	rt.router = NewRouter()
	rt.registryMethod()
	rt.initGetHandler()
	rt.initPostHandler()
	return rt
}

//start server
func (this *restServer) Start() error {
	retPort := int(config.Parameters.BaseConfig.PortBase + uint32(config.Parameters.BaseConfig.HttpRestPortOffset))

	if retPort == 0 {
		log.Fatal("Not configure HttpRestPort port ")
		return nil
	}

	tlsFlag := false
	if tlsFlag || retPort%1000 == TLS_PORT {
		var err error
		this.listener, err = this.initTlsListen()
		if err != nil {
			log.Error("Https Cert: ", err.Error())
			return err
		}
	} else {
		var err error
		this.listener, err = net.Listen("tcp", ":"+strconv.Itoa(retPort))
		if err != nil {
			log.Fatal("net.Listen: ", err.Error())
			return err
		}
	}
	this.server = &http.Server{Handler: this.router}
	err := this.server.Serve(this.listener)

	if err != nil {
		log.Fatal("ListenAndServe: ", err.Error())
		return err
	}

	return nil
}

//resigtry handler method
func (this *restServer) registryMethod() {
	getMethodMap := map[string]Action{
		GET_BLK_TXS_BY_HEIGHT: {name: "getblocktxsbyheight", handler: GetBlockTxsByHeight},
		GET_BLK_BY_HEIGHT:     {name: "getblockbyheight", handler: GetBlockByHeight},
		GET_BLK_BY_HASH:       {name: "getblockbyhash", handler: GetBlockByHash},
		GET_BLK_HEIGHT:        {name: "getblockheight", handler: GetBlockHeight},
		GET_BLK_HASH:          {name: "getblockhash", handler: GetBlockHash},
		GET_TX:                {name: "gettransaction", handler: GetTransactionByHash},
		GET_TXS_HEIGHT_LIMIT:  {name: "gettxsbyheightlimit", handler: GetTxByHeightAndLimit},
		GET_STORAGE:           {name: "getstorage", handler: GetStorage},
		GET_BALANCE:           {name: "getbalance", handler: GetBalance},
		GET_CONTRACT_STATE:    {name: "getcontract", handler: GetContractState},
		GET_SMTCOCE_EVT_TXS:   {name: "getsmartcodeeventbyheight", handler: GetSmartCodeEventTxsByHeight},
		GET_SMTCOCE_EVTS:      {name: "getsmartcodeeventbyhash", handler: GetSmartCodeEventByTxHash},
		GET_BLK_HGT_BY_TXHASH: {name: "getblockheightbytxhash", handler: GetBlockHeightByTxHash},
		GET_MERKLE_PROOF:      {name: "getmerkleproof", handler: GetMerkleProof},
		GET_GAS_PRICE:         {name: "getgasprice", handler: GetGasPrice},
		GET_ALLOWANCE:         {name: "getallowance", handler: GetAllowance},
		GET_MEMPOOL_TXCOUNT:   {name: "getmempooltxcount", handler: GetMemPoolTxCount},
		GET_MEMPOOL_TXSTATE:   {name: "getmempooltxstate", handler: GetMemPoolTxState},
		GET_VERSION:           {name: "getversion", handler: GetNodeVersion},
		GET_NETWORKID:         {name: "getnetworkid", handler: GetNetworkId},

		GET_CURRENT_ACCOUNT:  {name: "getcurrentaccount", handler: GetCurrentAccount},
		EXPORT_WALLETFILE:    {name: "exportwalletfile", handler: ExportWalletFile},
		EXPORT_WIFPRIVATEKEY: {name: "exportwifprivatekey", handler: ExportWIFPrivateKey},

		DSP_NODE_UNREGISTER:       {name: "unregisternode", handler: UnregisterNode},
		DSP_NODE_QUERY:            {name: "querynode", handler: NodeQuery},
		DSP_NODE_WITHDRAW:         {name: "withdrawnode", handler: NodeWithdrawProfit},
		DSP_CLIENT_GET_USER_SPACE: {name: "getuserspacesss", handler: GetUserSpace},
		DSP_USERSPACE_RECORDS:     {name: "getuserspacerecords", handler: GetUserSpaceRecords},
		DSP_GET_UPLOAD_FILELIST:   {name: "getuploadfilelist", handler: GetUploadFiles},
		DSP_GET_DOWNLOAD_FILELIST: {name: "getdownloadfilelist", handler: GetDownloadFiles},
		DSP_GET_FILE_TRANSFERLIST: {name: "gettransferlist", handler: GetTransferList},
		DSP_FILE_UPLOAD_FEE:       {name: "uploadfilefee", handler: CalculateUploadFee},
		DSP_FILE_DOWNLOAD_INFO:    {name: "getdownloadinfo", handler: GetDownloadFileInfo},
		DSP_FILE_SHARE_INCOME:     {name: "getfileshareincome", handler: GetFileShareIncome},
		DSP_FILE_SHARE_REVENUE:    {name: "getfilesharerevenue", handler: GetFileShareRevenue},
		DSP_GET_FILE_WHITELIST:    {name: "getwhitelist", handler: GetFileWhiteList},

		GET_CHANNEL_INIT_PROGRESS: {name: "channelinitprogress", handler: GetChannelInitProgress},
		GET_ALL_CHANNEL:           {name: "getallchannels", handler: GetAllChannels},
		OPEN_CHANNEL:              {name: "openchannel", handler: OpenChannel},
		TRANSFER_BY_CHANNEL:       {name: "transferbychannel", handler: TransferByChannel},
		QUERY_CHANNEL_DEPOSIT:     {name: "querydeposit", handler: QueryChannelDeposit},
		QUERY_CHANNEL:             {name: "querychannel", handler: QueryChannel},
		QUERY_CHANNEL_BY_ID:       {name: "querychannelbyid", handler: QueryChannelByID},

		DNS_REGISTER_DNS:     {name: "registerdns", handler: RegisterDns},
		DNS_UNREGISTER_DNS:   {name: "unregisterdns", handler: UnRegisterDns},
		DNS_QUIT:             {name: "quitdns", handler: QuitDns},
		DNS_ADD_DEPOSIT:      {name: "addpos", handler: AddPos},
		DNS_REDUCE_DEPOSIT:   {name: "reducepos", handler: ReducePos},
		DNS_QUERY_REG_INFOS:  {name: "queryreginfos", handler: QueryRegInfos},
		DNS_QUERY_HOST_INFOS: {name: "queryhostinfos", handler: QueryHostInfos},
		DNS_QUERY_REG_INFO:   {name: "queryreginfo", handler: QueryRegInfo},
		DNS_QUERY_HOST_INFO:  {name: "queryhostinfo", handler: QueryHostInfo},
	}
	this.getMap = getMethodMap

	postMethodMap := map[string]Action{
		ASSET_TRANSFER_DIRECT: {name: "assettransferdirect", handler: AssetTransferDirect},

		NEW_ACCOUNT:                    {name: "newaccount", handler: NewAccount},
		LOGOUT_ACCOUNT:                 {name: "logout", handler: Logout},
		IMPORT_ACCOUNT_WITH_PRIVATEKEY: {name: "importaccountwithprivatekey", handler: ImportWithPrivateKey},
		IMPORT_ACCOUNT_WITH_WALLETFILE: {name: "importaccountwithwalletfile", handler: ImportWithWalletData},

		DSP_NODE_REGISTER:         {name: "registernode", handler: RegisterNode},
		DSP_NODE_UPDATE:           {name: "updatenode", handler: NodeUpdate},
		DSP_SET_USER_SPACE:        {name: "setuserspace", handler: SetUserSpace},
		DSP_FILE_UPLOAD:           {name: "uploadfile", handler: UploadFile},
		DSP_FILE_DELETE:           {name: "deletefile", handler: DeleteFile},
		DSP_FILE_DOWNLOAD:         {name: "downloadfile", handler: DownloadFile},
		DSP_FILE_ENCRYPT:          {name: "encryptfile", handler: EncryptFile},
		DSP_FILE_DECRYPT:          {name: "decryptfile", handler: DecryptFile},
		DSP_UPDATE_FILE_WHITELIST: {name: "updatewhitelist", handler: WhiteListOperate},

		DEPOSIT_CHANNEL:  {name: "depositchannel", handler: DepositChannel},
		WITHDRAW_CHANNEL: {name: "withdrawchannel", handler: WithdrawChannel},

		DNS_REGISTER:  {name: "registerurl", handler: RegisterUrl},
		DNS_BIND:      {name: "bindurl", handler: BindUrl},
		DNS_QUERYLINK: {name: "querylink", handler: QueryLink},

		SET_CONFIG: {name: "setconfig", handler: SetConfig},
	}
	this.postMap = postMethodMap
}
func (this *restServer) getPath(url string) string {
	//path for themis-go-sdk
	if strings.Contains(url, strings.TrimRight(GET_BLK_TXS_BY_HEIGHT, ":height")) {
		return GET_BLK_TXS_BY_HEIGHT
	} else if strings.Contains(url, strings.TrimRight(GET_BLK_BY_HEIGHT, ":height")) {
		return GET_BLK_BY_HEIGHT
	} else if strings.Contains(url, strings.TrimRight(GET_BLK_BY_HASH, ":hash")) {
		return GET_BLK_BY_HASH
	} else if strings.Contains(url, strings.TrimRight(GET_BLK_HASH, ":height")) {
		return GET_BLK_HASH
	} else if strings.Contains(url, strings.TrimRight(GET_TX, ":hash")) {
		return GET_TX
	} else if strings.Contains(url, strings.TrimRight(GET_TXS_HEIGHT_LIMIT, ":addr/:type")) {
		return GET_TXS_HEIGHT_LIMIT
	} else if strings.Contains(url, strings.TrimRight(GET_STORAGE, ":hash/:key")) {
		return GET_STORAGE
	} else if strings.Contains(url, strings.TrimRight(GET_BALANCE, ":addr")) {
		return GET_BALANCE
	} else if strings.Contains(url, strings.TrimRight(GET_CONTRACT_STATE, ":hash")) {
		return GET_CONTRACT_STATE
	} else if strings.Contains(url, strings.TrimRight(GET_SMTCOCE_EVT_TXS, ":height")) {
		return GET_SMTCOCE_EVT_TXS
	} else if strings.Contains(url, strings.TrimRight(GET_SMTCOCE_EVTS, ":hash")) {
		return GET_SMTCOCE_EVTS
	} else if strings.Contains(url, strings.TrimRight(GET_BLK_HGT_BY_TXHASH, ":hash")) {
		return GET_BLK_HGT_BY_TXHASH
	} else if strings.Contains(url, strings.TrimRight(GET_MERKLE_PROOF, ":hash")) {
		return GET_MERKLE_PROOF
	} else if strings.Contains(url, strings.TrimRight(GET_ALLOWANCE, ":asset/:from/:to")) {
		return GET_ALLOWANCE
	} else if strings.Contains(url, strings.TrimRight(GET_UNBOUNDONG, ":addr")) {
		return GET_UNBOUNDONG
	} else if strings.Contains(url, strings.TrimRight(GET_GRANTONG, ":addr")) {
		return GET_GRANTONG
	} else if strings.Contains(url, strings.TrimRight(GET_MEMPOOL_TXSTATE, ":hash")) {
		return GET_MEMPOOL_TXSTATE
	}

	// //path for asset
	// if strings.Contains(url, "/api/v1/asset/transfer/direct") {
	// 	return ASSET_TRANSFER_DIRECT
	// }

	//path for Dsp
	if strings.Contains(url, strings.TrimSuffix(DSP_NODE_QUERY, ":addr")) {
		return DSP_NODE_QUERY
	} else if strings.Contains(url, strings.TrimSuffix(DSP_CLIENT_GET_USER_SPACE, ":addr")) {
		return DSP_CLIENT_GET_USER_SPACE
	} else if strings.Contains(url, strings.TrimSuffix(DSP_GET_UPLOAD_FILELIST, ":type/:offset/:limit")) {
		return DSP_GET_UPLOAD_FILELIST
	} else if strings.Contains(url, strings.TrimSuffix(DSP_GET_DOWNLOAD_FILELIST, ":type/:offset/:limit")) {
		return DSP_GET_DOWNLOAD_FILELIST
	} else if strings.Contains(url, strings.TrimSuffix(DSP_GET_FILE_TRANSFERLIST, ":type/:offset/:limit")) {
		return DSP_GET_FILE_TRANSFERLIST
	} else if strings.Contains(url, strings.TrimSuffix(DSP_FILE_UPLOAD_FEE, ":file")) {
		return DSP_FILE_UPLOAD_FEE
	} else if strings.Contains(url, strings.TrimSuffix(DSP_FILE_DOWNLOAD_INFO, ":url")) {
		return DSP_FILE_DOWNLOAD_INFO
	} else if strings.Contains(url, strings.TrimSuffix(DSP_FILE_SHARE_INCOME, ":begin/:end/:offset/:limit")) {
		return DSP_FILE_SHARE_INCOME
	} else if strings.Contains(url, DSP_FILE_SHARE_REVENUE) {
		return DSP_FILE_SHARE_REVENUE
	} else if strings.Contains(url, strings.TrimSuffix(DSP_GET_FILE_WHITELIST, ":hash")) {
		return DSP_GET_FILE_WHITELIST
	} else if strings.Contains(url, strings.TrimSuffix(DSP_USERSPACE_RECORDS, ":addr/:offset/:limit")) {
		return DSP_USERSPACE_RECORDS
	}

	//path for channel
	if strings.Contains(url, strings.TrimRight(OPEN_CHANNEL, ":partneraddr")) {
		return OPEN_CHANNEL
	} else if strings.Contains(url, strings.TrimRight(QUERY_CHANNEL, ":partneraddr")) {
		return QUERY_CHANNEL
		// } else if strings.Contains(url, strings.TrimRight(DEPOSIT_CHANNEL, ":partneraddr/:amount")) {
		// 	return DEPOSIT_CHANNEL
	} else if strings.Contains(url, strings.TrimRight(TRANSFER_BY_CHANNEL, ":toaddr/:amount/:paymentid")) {
		return TRANSFER_BY_CHANNEL
	} else if strings.Contains(url, strings.TrimRight(QUERY_CHANNEL_DEPOSIT, ":partneraddr")) {
		return QUERY_CHANNEL_DEPOSIT
	} else if strings.Contains(url, strings.TrimRight(QUERY_CHANNEL_BY_ID, ":id")) {
		return QUERY_CHANNEL_BY_ID
	} else if strings.Contains(url, GET_CHANNEL_INIT_PROGRESS) {
		return GET_CHANNEL_INIT_PROGRESS
	}

	//path for DNS
	if strings.Contains(url, strings.TrimRight(DNS_REGISTER_DNS, ":ip/:port/:deposit")) {
		return DNS_REGISTER_DNS
	} else if strings.Contains(url, strings.TrimRight(DNS_ADD_DEPOSIT, ":amount")) {
		return DNS_ADD_DEPOSIT
	} else if strings.Contains(url, strings.TrimRight(DNS_REDUCE_DEPOSIT, ":amount")) {
		return DNS_REDUCE_DEPOSIT
	} else if strings.Contains(url, strings.TrimRight(DNS_QUERY_REG_INFO, ":pubkey")) {
		return DNS_QUERY_REG_INFO
	} else if strings.Contains(url, strings.TrimRight(DNS_QUERY_HOST_INFO, ":addr")) {
		return DNS_QUERY_HOST_INFO
	}

	return url
}

//get request params
func (this *restServer) getParams(r *http.Request, url string, req map[string]interface{}) map[string]interface{} {
	//params for themis go sdk
	switch url {
	case GET_BLK_TXS_BY_HEIGHT:
		req["Height"] = getParam(r, "height")
	case GET_BLK_BY_HEIGHT:
		req["Raw"], req["Height"] = r.FormValue("raw"), getParam(r, "height")
	case GET_BLK_BY_HASH:
		req["Raw"], req["Hash"] = r.FormValue("raw"), getParam(r, "hash")
	case GET_BLK_HEIGHT:
	case GET_BLK_HASH:
		req["Height"] = getParam(r, "height")
	case GET_TX:
		req["Hash"], req["Raw"] = getParam(r, "hash"), r.FormValue("raw")
	case GET_TXS_HEIGHT_LIMIT:
		req["Addr"], req["Type"], req["Asset"], req["Height"], req["Limit"] = getParam(r, "addr"), getParam(r, "type"), r.FormValue("asset"), r.FormValue("height"), r.FormValue("limit")
	case GET_STORAGE:
		req["Hash"], req["Key"] = getParam(r, "hash"), getParam(r, "key")
	case GET_BALANCE:
		req["Addr"] = getParam(r, "addr")
	case GET_CONTRACT_STATE:
		req["Hash"], req["Raw"] = getParam(r, "hash"), r.FormValue("raw")
	case GET_SMTCOCE_EVT_TXS:
		req["Height"] = getParam(r, "height")
	case GET_SMTCOCE_EVTS:
		req["Hash"] = getParam(r, "hash")
	case GET_BLK_HGT_BY_TXHASH:
		req["Hash"] = getParam(r, "hash")
	case GET_MERKLE_PROOF:
		req["Hash"] = getParam(r, "hash")
	case GET_ALLOWANCE:
		req["Asset"] = getParam(r, "asset")
		req["From"], req["To"] = getParam(r, "from"), getParam(r, "to")
	case GET_UNBOUNDONG:
		req["Addr"] = getParam(r, "addr")
	case GET_GRANTONG:
		req["Addr"] = getParam(r, "addr")
	case GET_MEMPOOL_TXSTATE:
		req["Hash"] = getParam(r, "hash")
	default:
	}

	// params for asset
	// switch url {
	// case ASSET_TRANSFER_DIRECT:
	// 	req["To"], req["Asset"], req["Amount"] = getParam(r, "to"), getParam(r, "asset"), getParam(r, "amount")
	// default:
	// }

	//params for Dsp
	switch url {
	case DSP_NODE_QUERY:
		req["Addr"] = getParam(r, "addr")
	case DSP_CLIENT_GET_USER_SPACE:
		req["Addr"] = getParam(r, "addr")
	case DSP_GET_UPLOAD_FILELIST:
		req["Type"], req["Offset"], req["Limit"] = getParam(r, "type"), getParam(r, "offset"), getParam(r, "limit")
	case DSP_GET_DOWNLOAD_FILELIST:
		req["Type"], req["Offset"], req["Limit"] = getParam(r, "type"), getParam(r, "offset"), getParam(r, "limit")
	case DSP_GET_FILE_TRANSFERLIST:
		req["Type"], req["Offset"], req["Limit"] = getParam(r, "type"), getParam(r, "offset"), getParam(r, "limit")
	case DSP_FILE_UPLOAD_FEE:
		req["Path"], req["Duration"], req["Interval"], req["Times"], req["CopyNum"], req["WhiteList"] = getParam(r, "file"), r.FormValue("duration"), r.FormValue("interval"), r.FormValue("times"), r.FormValue("copynum"), r.FormValue("whitelistcount")
	case DSP_FILE_DOWNLOAD_INFO:
		req["Url"] = getParam(r, "url")
	case DSP_FILE_SHARE_INCOME:
		req["Begin"] = getParam(r, "begin")
		req["End"] = getParam(r, "end")
		req["Offset"] = getParam(r, "offset")
		req["Limit"] = getParam(r, "limit")
	case DSP_GET_FILE_WHITELIST:
		req["FileHash"] = getParam(r, "hash")
	case DSP_USERSPACE_RECORDS:
		req["Addr"], req["Offset"], req["Limit"] = getParam(r, "addr"), getParam(r, "offset"), getParam(r, "limit")
	default:
	}

	//params for channel
	switch url {
	case OPEN_CHANNEL:
		req["Partner"] = getParam(r, "partneraddr")
	// case DEPOSIT_CHANNEL:
	// 	req["Partner"] = getParam(r, "partneraddr")
	// 	req["Amount"] = getParam(r, "amount")
	case TRANSFER_BY_CHANNEL:
		req["Amount"] = getParam(r, "amount")
		req["To"], req["PaymentId"] = getParam(r, "toaddr"), getParam(r, "paymentid")
	case QUERY_CHANNEL_DEPOSIT:
		req["Partner"] = getParam(r, "partneraddr")
	case QUERY_CHANNEL:
		req["Partner"] = getParam(r, "partneraddr")
	case QUERY_CHANNEL_BY_ID:
		req["Id"] = getParam(r, "id")
	default:
	}

	//params for Dns
	switch url {
	case DNS_REGISTER_DNS:
		req["Ip"] = getParam(r, "ip")
		req["Port"] = getParam(r, "port")
		req["Deposit"] = getParam(r, "deposit")
	case DNS_ADD_DEPOSIT:
		req["Amount"] = getParam(r, "amount")
	case DNS_REDUCE_DEPOSIT:
		req["Amount"] = getParam(r, "amount")
	case DNS_QUERY_REG_INFO:
		req["Pubkey"] = getParam(r, "pubkey")
	case DNS_QUERY_HOST_INFO:
		req["Addr"] = getParam(r, "addr")
	default:
	}

	return req
}

//init get handler
func (this *restServer) initGetHandler() {

	for k := range this.getMap {
		this.router.Get(k, func(w http.ResponseWriter, r *http.Request) {

			var req = make(map[string]interface{})
			var resp map[string]interface{}

			fmt.Printf("get path for %v\n", r.URL.Path)
			url := this.getPath(r.URL.Path)
			if h, ok := this.getMap[url]; ok {
				req = this.getParams(r, url, req)
				resp = h.handler(req)
				resp["Action"] = h.name
			} else {
				resp = ResponsePack(berr.INVALID_METHOD)
			}
			this.response(w, resp)
		})
	}
}

//init post handler
func (this *restServer) initPostHandler() {
	for k := range this.postMap {
		this.router.Post(k, func(w http.ResponseWriter, r *http.Request) {

			body, _ := ioutil.ReadAll(r.Body)
			defer r.Body.Close()

			var req = make(map[string]interface{})
			var resp map[string]interface{}

			// url := this.getPath(r.URL.Path)
			url := r.URL.Path
			if h, ok := this.postMap[url]; ok {
				if err := json.Unmarshal(body, &req); err == nil {
					req = this.getParams(r, url, req)
					resp = h.handler(req)
					resp["Action"] = h.name
				} else {
					resp = ResponsePack(berr.ILLEGAL_DATAFORMAT)
					resp["Action"] = h.name
				}
			} else {
				resp = ResponsePack(berr.INVALID_METHOD)
			}
			this.response(w, resp)
		})
	}
	//Options
	for k := range this.postMap {
		this.router.Options(k, func(w http.ResponseWriter, r *http.Request) {
			this.write(w, []byte{})
		})
	}

}

func (this *restServer) write(w http.ResponseWriter, data []byte) {
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json;charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Write(data)
}

//response
func (this *restServer) response(w http.ResponseWriter, resp map[string]interface{}) {
	desc, ok := resp["Desc"].(string)
	if ok && len(desc) == 0 {
		resp["Desc"] = berr.ErrMap[resp["Error"].(int64)]
	}
	data, err := json.Marshal(resp)
	if err != nil {
		log.Fatal("HTTP Handle - json.Marshal: %v", err)
		return
	}
	this.write(w, data)
}

//stop restful server
func (this *restServer) Stop() {
	if this.server != nil {
		this.server.Shutdown(context.Background())
		log.Error("Close restful ")
	}
}

//restart server
func (this *restServer) Restart(cmd map[string]interface{}) map[string]interface{} {
	go func() {
		time.Sleep(time.Second)
		this.Stop()
		time.Sleep(time.Second)
		go this.Start()
	}()

	var resp = ResponsePack(berr.SUCCESS)
	return resp
}

//init tls
func (this *restServer) initTlsListen() (net.Listener, error) {

	certPath := config.Parameters.BaseConfig.HttpCertPath
	keyPath := config.Parameters.BaseConfig.HttpKeyPath

	// load cert
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		log.Error("load keys fail", err)
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	restPort := strconv.Itoa(int(config.Parameters.BaseConfig.PortBase + uint32(config.Parameters.BaseConfig.HttpRestPortOffset)))
	log.Info("TLS listen port is ", restPort)
	listener, err := tls.Listen("tcp", ":"+restPort, tlsConfig)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return listener, nil
}
