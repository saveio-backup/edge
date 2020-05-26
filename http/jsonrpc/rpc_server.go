/*
 * Copyright (C) 2019 The themis Authors
 * This file is part of The themis library.
 *
 * The themis is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The themis is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package jsonrpc privides a function to start json rpc server
package jsonrpc

import (
	"net/http"
	"strconv"

	"fmt"

	"github.com/saveio/edge/common/config"
	"github.com/saveio/edge/http/base/rpc"
)

func StartRPCServer() error {
	http.HandleFunc("/", rpc.Handle)
	rpc.HandleFunc("getcurrentaccount", rpc.GetCurrentAccount)
	rpc.HandleFunc("newaccount", rpc.NewAccount)
	rpc.HandleFunc("importwithprivatekey", rpc.ImportWithPrivateKey)
	rpc.HandleFunc("importwithwalletdata", rpc.ImportWithWalletData)
	rpc.HandleFunc("exportwalletfile", rpc.ExportWalletFile)
	rpc.HandleFunc("exportprivatekey", rpc.ExportPrivateKey)
	rpc.HandleFunc("logout", rpc.Logout)

	rpc.HandleFunc("getnodeversion", rpc.GetNodeVersion)
	rpc.HandleFunc("getnetworkid", rpc.GetNetworkId)
	rpc.HandleFunc("getblockheight", rpc.GetBlockHeight)
	rpc.HandleFunc("getblockhash", rpc.GetBlockHash)
	rpc.HandleFunc("getblockbyhash", rpc.GetBlockByHash)
	rpc.HandleFunc("getblockheightbytxhash", rpc.GetBlockHeightByTxHash)
	rpc.HandleFunc("getblocktxsbyheight", rpc.GetBlockTxsByHeight)
	rpc.HandleFunc("getblockbyheight", rpc.GetBlockByHeight)
	rpc.HandleFunc("gettransactionbyhash", rpc.GetTransactionByHash)
	rpc.HandleFunc("getsmartcodeeventtxsbyheight", rpc.GetSmartCodeEventTxsByHeight)
	rpc.HandleFunc("getsmartcodeeventbytxhash", rpc.GetSmartCodeEventByTxHash)
	rpc.HandleFunc("getcontractstate", rpc.GetContractState)
	rpc.HandleFunc("getstorage", rpc.GetStorage)
	rpc.HandleFunc("getbalance", rpc.GetBalance)
	rpc.HandleFunc("getmerkleproof", rpc.GetMerkleProof)
	rpc.HandleFunc("getgasprice", rpc.GetGasPrice)
	rpc.HandleFunc("getallowance", rpc.GetAllowance)
	rpc.HandleFunc("getmempooltxcount", rpc.GetMemPoolTxCount)
	rpc.HandleFunc("getmempooltxstate", rpc.GetMemPoolTxState)
	rpc.HandleFunc("gettxbyheightandlimit", rpc.GetTxByHeightAndLimit)
	rpc.HandleFunc("assettransferdirect", rpc.AssetTransferDirect)
	rpc.HandleFunc("setconfig", rpc.SetConfig)

	rpc.HandleFunc("getallchannels", rpc.GetAllChannels)
	rpc.HandleFunc("openchannel", rpc.OpenChannel)
	rpc.HandleFunc("openalldnschannel", rpc.OpenToAllDNSChannel)
	rpc.HandleFunc("closechannel", rpc.CloseChannel)
	rpc.HandleFunc("closeallchannel", rpc.CloseAllChannel)
	rpc.HandleFunc("depositchannel", rpc.DepositChannel)
	rpc.HandleFunc("withdrawchannel", rpc.WithdrawChannel)
	rpc.HandleFunc("querychanneldeposit", rpc.QueryChannelDeposit)
	rpc.HandleFunc("querychannel", rpc.QueryChannel)
	rpc.HandleFunc("querychannelbyid", rpc.QueryChannelByID)
	rpc.HandleFunc("transferbychannel", rpc.TransferByChannel)
	rpc.HandleFunc("getchannelinitprogress", rpc.GetChannelInitProgress)
	rpc.HandleFunc("channelcooperativesettle", rpc.ChannelCooperativeSettle)

	rpc.HandleFunc("uploadfile", rpc.UploadFile)
	rpc.HandleFunc("deletefile", rpc.DeleteFile)
	rpc.HandleFunc("downloadfile", rpc.DownloadFile)
	rpc.HandleFunc("getuploadfiles", rpc.GetUploadFiles)
	rpc.HandleFunc("getdownloadfiles", rpc.GetDownloadFiles)
	rpc.HandleFunc("gettransferlist", rpc.GetTransferList)
	rpc.HandleFunc("calculateuploadfee", rpc.CalculateUploadFee)
	rpc.HandleFunc("getdownloadfileinfo", rpc.GetDownloadFileInfo)
	rpc.HandleFunc("encryptfile", rpc.EncryptFile)
	rpc.HandleFunc("decryptfile", rpc.DecryptFile)
	rpc.HandleFunc("getfileshareincome", rpc.GetFileShareIncome)
	rpc.HandleFunc("getfilesharerevenue", rpc.GetFileShareRevenue)
	rpc.HandleFunc("whitelistoperate", rpc.WhiteListOperate)
	rpc.HandleFunc("getfilewhitelist", rpc.GetFileWhiteList)
	rpc.HandleFunc("getuserspace", rpc.GetUserSpace)
	rpc.HandleFunc("setuserspace", rpc.SetUserSpace)
	rpc.HandleFunc("getuserspacerecords", rpc.GetUserSpaceRecords)

	rpc.HandleFunc("registernode", rpc.RegisterNode)
	rpc.HandleFunc("unregisternode", rpc.UnregisterNode)
	rpc.HandleFunc("nodequery", rpc.NodeQuery)
	rpc.HandleFunc("nodeupdate", rpc.NodeUpdate)
	rpc.HandleFunc("nodewithdrawprofit", rpc.NodeWithdrawProfit)
	rpc.HandleFunc("registerurl", rpc.RegisterUrl)
	rpc.HandleFunc("bindurl", rpc.BindUrl)
	rpc.HandleFunc("querylink", rpc.QueryLink)
	rpc.HandleFunc("registerdns", rpc.RegisterDns)
	rpc.HandleFunc("unregisterdns", rpc.UnRegisterDns)
	rpc.HandleFunc("quitdns", rpc.QuitDns)
	rpc.HandleFunc("addpos", rpc.AddPos)
	rpc.HandleFunc("reducepos", rpc.ReducePos)
	rpc.HandleFunc("queryreginfos", rpc.QueryRegInfos)
	rpc.HandleFunc("queryreginfo", rpc.QueryRegInfo)
	rpc.HandleFunc("queryhostinfos", rpc.QueryHostInfos)
	rpc.HandleFunc("queryhostinfo", rpc.QueryHostInfo)
	rpc.HandleFunc("querypublicip", rpc.QueryPublicIP)

	err := http.ListenAndServe(":"+strconv.Itoa(int(config.Parameters.BaseConfig.PortBase+uint32(config.Parameters.BaseConfig.JsonRpcPortOffset))), nil)
	if err != nil {
		return fmt.Errorf("ListenAndServe error:%s", err)
	}
	return nil
}
