package dsp

import "errors"

type DspErr struct {
	Code  int64
	Error error
}

const (
	SUCCESS                = 0
	INTERNAL_ERROR         = 40001
	INVALID_PARAMS         = 40002
	NO_DSP                 = 40003
	NO_DB                  = 40004
	CONTRACT_ERROR         = 40005
	INSUFFICIENT_BALANCE   = 40006
	NO_ACCOUNT             = 40007
	ACCOUNT_EXIST          = 40008
	NO_DNS                 = 40009
	INVALID_WALLET_ADDRESS = 40010
	NO_CHANNEL             = 40011

	CHAIN_INTERNAL_ERROR              = 50000
	CHAIN_GET_HEIGHT_FAILED           = 50001
	CHAIN_GET_BLK_BY_HEIGHT_FAILED    = 50002
	CHAIN_WAIT_TX_COMFIRMED_TIMEOUT   = 50003
	CHAIN_UNKNOWN_BLOCK               = 50004
	CHAIN_UNKNOWN_TX                  = 50005
	CHAIN_UNKNOWN_SMARTCONTRACT       = 50006
	CHAIN_UNKNOWN_SMARTCONTRACT_EVENT = 50007
	CHAIN_UNKNOWN_ASSET               = 50008
	CHAIN_TRANSFER_ERROR              = 50009

	ACCOUNT_NOT_LOGIN      = 50012
	WALLET_FILE_NOT_EXIST  = 50013
	ACCOUNTDATA_NOT_EXIST  = 50014
	ACCOUNT_PASSWORD_WRONG = 50015
	CREATE_ACCOUNT_FAILED  = 50016
	ACCOUNT_EXPORT_FAILED  = 50017

	FS_GET_SETTING_FAILED           = 54001
	FS_GET_USER_SPACE_FAILED        = 54002
	FS_GET_FILE_LIST_FAILED         = 54003
	FS_UPDATE_USERSPACE_FAILED      = 54004
	FS_CANT_REVOKE_OF_EXISTS_FILE   = 54005
	FS_NO_USER_SPACE_TO_REVOKE      = 54006
	FS_USER_SPACE_SECOND_TOO_SMALL  = 54007
	FS_USER_SPACE_PERMISSION_DENIED = 54008
	FS_UPLOAD_FILEPATH_ERROR        = 54009
	FS_UPLOAD_INTERVAL_TOO_SMALL    = 54010
	FS_UPLOAD_GET_FILESIZE_FAILED   = 54011
	FS_UPLOAD_CALC_FEE_FAILED       = 54012
	FS_DELETE_CALC_FEE_FAILED       = 54013
	FS_USER_SPACE_SECOND_INVALID    = 54014

	DSP_INIT_FAILED            = 55000
	DSP_START_FAILED           = 55001
	DSP_STOP_FAILED            = 55002
	DSP_UPLOAD_FILE_FAILED     = 55010
	DSP_USER_SPACE_EXPIRED     = 55011
	DSP_USER_SPACE_NOT_ENOUGH  = 55012
	DSP_UPLOAD_URL_EXIST       = 55013
	DSP_DELETE_FILE_FAILED     = 55014
	DSP_CALC_UPLOAD_FEE_FAILED = 55015
	DSP_GET_FILE_LINK_FAILED   = 55016
	DSP_ENCRYPTED_FILE_FAILED  = 55017
	DSP_DECRYPTED_FILE_FAILED  = 55018
	DSP_WHITELIST_OP_FAILED    = 55019
	DSP_GET_WHITELIST_FAILED   = 55020
	DSP_UPDATE_CONFIG_FAILED   = 55021
	DSP_UPLOAD_FILE_EXIST      = 55022
	DSP_PAUSE_UPLOAD_FAIELD    = 55023
	DSP_RESUME_UPLOAD_FAIELD   = 55024
	DSP_RETRY_UPLOAD_FAIELD    = 55025
	DSP_PAUSE_DOWNLOAD_FAIELD  = 55026
	DSP_RESUME_DOWNLOAD_FAIELD = 55027
	DSP_RETRY_DOWNLOAD_FAIELD  = 55028
	DSP_CANCEL_TASK_FAILED     = 55029

	DSP_NODE_REGISTER_FAILED            = 55030
	DSP_NODE_UNREGISTER_FAILED          = 55031
	DSP_NODE_UPDATE_FAILED              = 55032
	DSP_NODE_WITHDRAW_FAILED            = 55033
	DSP_NODE_QUERY_FAILED               = 55034
	DSP_URL_REGISTER_FAILED             = 55040
	DSP_URL_BIND_FAILED                 = 55041
	DSP_URL_DELETE_FAILED               = 55042
	DSP_DNS_REGISTER_FAILED             = 55050
	DSP_DNS_UNREGISTER_FAILED           = 55051
	DSP_DNS_UPDATE_FAILED               = 55052
	DSP_DNS_WITHDRAW_FAILED             = 55053
	DSP_DNS_QUIT_FAILED                 = 55054
	DSP_DNS_ADDPOS_FAILED               = 55055
	DSP_DNS_REDUCEPOS_FAILED            = 55056
	DSP_DNS_GET_NODE_BY_ADDR            = 55057
	DSP_DNS_QUERY_INFOS_FAILED          = 55058
	DSP_DNS_QUERY_INFO_FAILED           = 55059
	DSP_DNS_QUERY_ALLINFOS_FAILED       = 55060
	DSP_DNS_GET_EXTERNALIP_FAILED       = 55061
	DSP_USER_SPACE_PERIOD_NOT_ENOUGH    = 55062
	DSP_CUSTOM_EXPIRED_NOT_ENOUGH       = 55063
	DSP_NO_PRIVILEGE_TO_DOWNLOAD        = 55064
	DSP_DNS_UPDATE_PLUGIN_INFO_FAILED   = 55065
	DSP_DNS_QUERY_PLUGIN_INFO_FAILED    = 55066
	DSP_DNS_QUERY_ALLPLUGININFOS_FAILED = 55067

	DSP_FILE_INFO_NOT_FOUND      = 55100
	DSP_FILE_NOT_EXISTS          = 55101
	DSP_FILE_DECRYPTED_WRONG_PWD = 55102

	DSP_CHANNEL_INTERNAL_ERROR           = 56000
	DSP_CHANNEL_OPEN_FAILED              = 56001
	DSP_CHANNEL_CLOSE_FAILED             = 56002
	DSP_CHANNEL_QUERY_AVA_BALANCE_FAILED = 56003
	DSP_CHANNEL_DEPOSIT_FAILED           = 56004
	DSP_CHANNEL_WITHDRAW_FAILED          = 56005
	DSP_CHANNEL_WITHDRAW_OVERFLOW        = 56006
	DSP_CHANNEL_GET_ALL_FAILED           = 56007
	DSP_CHANNEL_MEDIATRANSFER_FAILED     = 56008
	DSP_CHANNEL_CO_SETTLE_FAILED         = 56009
	DSP_CHANNEL_INIT_NOT_FINISH          = 56010
	DSP_CHANNEL_EXIST                    = 56011
	DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST   = 56012
	DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH   = 56013
	DSP_CHANNEL_WITHDRAW_WRONG_AMOUNT    = 56014
	DSP_CHANNEL_SYNCING                  = 56015
	DSP_CHANNEL_OPEN_TO_NO_DNS           = 56016
	DSP_CHANNEL_DNS_OFFLINE              = 56017
	DSP_CHANNEL_GET_HOSTADDR_ERROR       = 56018
	DSP_CHANNEL_NOT_EXIST                = 56019

	DSP_TASK_NOT_EXIST = 58000

	DB_FIND_SHARE_RECORDS_FAILED     = 59000
	DB_SUM_SHARE_PROFIT_FAILED       = 59001
	DB_FIND_USER_SPACE_RECORD_FAILED = 59002
	DB_ADD_USER_SPACE_RECORD_FAILED  = 59003
	DB_GET_FILEINFO_FAILED           = 59004

	NET_RECONNECT_PEER_FAILED = 59100
	NET_PROXY_DISCONNECTED    = 59101
)

var ErrMaps = map[int64]error{
	INTERNAL_ERROR:         errors.New("internal error"),
	INVALID_PARAMS:         errors.New("invalid params"),
	NO_DSP:                 errors.New("no dsp"),
	NO_DB:                  errors.New("no db"),
	CONTRACT_ERROR:         errors.New("contract error"),
	INSUFFICIENT_BALANCE:   errors.New("insufficient balance"),
	NO_ACCOUNT:             errors.New("no account"),
	ACCOUNT_EXIST:          errors.New("account exist"),
	NO_DNS:                 errors.New("no dns"),
	INVALID_WALLET_ADDRESS: errors.New("no invalid wallet addr"),
	NO_CHANNEL:             errors.New("no channel service"),

	CHAIN_INTERNAL_ERROR:              errors.New("chain internal error"),
	CHAIN_GET_HEIGHT_FAILED:           errors.New("chain get height failed"),
	CHAIN_GET_BLK_BY_HEIGHT_FAILED:    errors.New("chain get blk by height failed"),
	CHAIN_WAIT_TX_COMFIRMED_TIMEOUT:   errors.New("chain wait tx comfirmed timeout"),
	CHAIN_UNKNOWN_BLOCK:               errors.New("chain unknown block"),
	CHAIN_UNKNOWN_TX:                  errors.New("chain unknown tx"),
	CHAIN_UNKNOWN_SMARTCONTRACT:       errors.New("chain unknown smartcontract"),
	CHAIN_UNKNOWN_SMARTCONTRACT_EVENT: errors.New("chain unknown smartcontract event"),
	CHAIN_UNKNOWN_ASSET:               errors.New("chain unknown asset"),
	CHAIN_TRANSFER_ERROR:              errors.New("chain transfer error"),

	ACCOUNT_NOT_LOGIN:      errors.New("account not login"),
	WALLET_FILE_NOT_EXIST:  errors.New("wallet file not exist"),
	ACCOUNTDATA_NOT_EXIST:  errors.New("accountdata not exist"),
	ACCOUNT_PASSWORD_WRONG: errors.New("account password wrong"),
	CREATE_ACCOUNT_FAILED:  errors.New("create account failed"),
	ACCOUNT_EXPORT_FAILED:  errors.New("account export failed"),

	FS_GET_SETTING_FAILED:           errors.New("fs get setting failed"),
	FS_GET_USER_SPACE_FAILED:        errors.New("fs get user space failed"),
	FS_GET_FILE_LIST_FAILED:         errors.New("fs get file list failed"),
	FS_UPDATE_USERSPACE_FAILED:      errors.New("fs update userspace failed"),
	FS_CANT_REVOKE_OF_EXISTS_FILE:   errors.New("fs cant revoke of exists file"),
	FS_NO_USER_SPACE_TO_REVOKE:      errors.New("fs no user space to revoke"),
	FS_USER_SPACE_SECOND_TOO_SMALL:  errors.New("fs user space second too small"),
	FS_USER_SPACE_PERMISSION_DENIED: errors.New("fs user space permission denied"),
	FS_UPLOAD_FILEPATH_ERROR:        errors.New("fs upload filepath error"),
	FS_UPLOAD_INTERVAL_TOO_SMALL:    errors.New("fs upload interval too small"),
	FS_UPLOAD_GET_FILESIZE_FAILED:   errors.New("fs upload get file size failed"),
	FS_UPLOAD_CALC_FEE_FAILED:       errors.New("fs upload calculate fee failed"),
	FS_DELETE_CALC_FEE_FAILED:       errors.New("fs delete calculate fee failed"),

	DSP_INIT_FAILED:                      errors.New("dsp init failed"),
	DSP_START_FAILED:                     errors.New("dsp start failed"),
	DSP_STOP_FAILED:                      errors.New("dsp stop failed"),
	DSP_UPLOAD_FILE_FAILED:               errors.New("dsp upload file failed"),
	DSP_USER_SPACE_EXPIRED:               errors.New("dsp user space expired"),
	DSP_USER_SPACE_NOT_ENOUGH:            errors.New("dsp user space not enough"),
	DSP_UPLOAD_URL_EXIST:                 errors.New("dsp upload url exist"),
	DSP_DELETE_FILE_FAILED:               errors.New("dsp delete file failed"),
	DSP_CALC_UPLOAD_FEE_FAILED:           errors.New("dsp calc upload fee failed"),
	DSP_GET_FILE_LINK_FAILED:             errors.New("dsp get file link failed"),
	DSP_ENCRYPTED_FILE_FAILED:            errors.New("dsp encrypted file failed"),
	DSP_DECRYPTED_FILE_FAILED:            errors.New("dsp decrypted file failed"),
	DSP_WHITELIST_OP_FAILED:              errors.New("dsp whitelist op failed"),
	DSP_GET_WHITELIST_FAILED:             errors.New("dsp get whitelist failed"),
	DSP_UPDATE_CONFIG_FAILED:             errors.New("dsp update config failed"),
	DSP_UPLOAD_FILE_EXIST:                errors.New("dsp upload file exist"),
	DSP_PAUSE_UPLOAD_FAIELD:              errors.New("dsp pause upload faield"),
	DSP_NODE_REGISTER_FAILED:             errors.New("dsp node register failed"),
	DSP_NODE_UNREGISTER_FAILED:           errors.New("dsp node unregister failed"),
	DSP_NODE_UPDATE_FAILED:               errors.New("dsp node update failed"),
	DSP_NODE_WITHDRAW_FAILED:             errors.New("dsp node withdraw failed"),
	DSP_NODE_QUERY_FAILED:                errors.New("dsp node query failed"),
	DSP_URL_REGISTER_FAILED:              errors.New("dsp url register failed"),
	DSP_URL_BIND_FAILED:                  errors.New("dsp url bind failed"),
	DSP_URL_DELETE_FAILED:                errors.New("dns url delete failed"),
	DSP_DNS_REGISTER_FAILED:              errors.New("dsp dns register failed"),
	DSP_DNS_UNREGISTER_FAILED:            errors.New("dsp dns unregister failed"),
	DSP_DNS_UPDATE_FAILED:                errors.New("dsp dns update failed"),
	DSP_DNS_WITHDRAW_FAILED:              errors.New("dsp dns withdraw failed"),
	DSP_DNS_QUIT_FAILED:                  errors.New("dsp dns quit failed"),
	DSP_DNS_ADDPOS_FAILED:                errors.New("dsp dns addpos failed"),
	DSP_DNS_REDUCEPOS_FAILED:             errors.New("dsp dns reducepos failed"),
	DSP_DNS_GET_NODE_BY_ADDR:             errors.New("dsp dns get node by addr"),
	DSP_DNS_QUERY_INFOS_FAILED:           errors.New("dsp dns query infos failed"),
	DSP_DNS_QUERY_INFO_FAILED:            errors.New("dsp dns query info failed"),
	DSP_DNS_QUERY_ALLINFOS_FAILED:        errors.New("dsp dns query allinfos failed"),
	DSP_DNS_UPDATE_PLUGIN_INFO_FAILED:    errors.New("dsp dns update plugin info failed"),
	DSP_DNS_QUERY_PLUGIN_INFO_FAILED:     errors.New("dsp dns query plugin info failed"),
	DSP_DNS_QUERY_ALLPLUGININFOS_FAILED:  errors.New("dsp dns query allplugininfos failed"),
	DSP_DNS_GET_EXTERNALIP_FAILED:        errors.New("dsp dns get externalip failed"),
	DSP_FILE_INFO_NOT_FOUND:              errors.New("dsp file info not found"),
	DSP_FILE_NOT_EXISTS:                  errors.New("dsp file not exists"),
	DSP_FILE_DECRYPTED_WRONG_PWD:         errors.New("dsp file decrypt password wrong"),
	DSP_CHANNEL_INTERNAL_ERROR:           errors.New("dsp channel internal error"),
	DSP_CHANNEL_OPEN_FAILED:              errors.New("dsp channel open failed"),
	DSP_CHANNEL_CLOSE_FAILED:             errors.New("dsp channel close failed"),
	DSP_CHANNEL_QUERY_AVA_BALANCE_FAILED: errors.New("dsp channel query ava balance failed"),
	DSP_CHANNEL_DEPOSIT_FAILED:           errors.New("dsp channel deposit failed"),
	DSP_CHANNEL_WITHDRAW_FAILED:          errors.New("dsp channel withdraw failed"),
	DSP_CHANNEL_WITHDRAW_OVERFLOW:        errors.New("dsp channel withdraw overflow"),
	DSP_CHANNEL_GET_ALL_FAILED:           errors.New("dsp channel get all failed"),
	DSP_CHANNEL_MEDIATRANSFER_FAILED:     errors.New("dsp channel mediatransfer failed"),
	DSP_CHANNEL_CO_SETTLE_FAILED:         errors.New("dsp channel co settle failed"),
	DSP_CHANNEL_INIT_NOT_FINISH:          errors.New("dsp channel init not finish"),
	DSP_USER_SPACE_PERIOD_NOT_ENOUGH:     errors.New("dsp user space period not enough"),
	DSP_CUSTOM_EXPIRED_NOT_ENOUGH:        errors.New("dsp custom expired height not enough"),
	DSP_CHANNEL_EXIST:                    errors.New("channel has exists or not settled"),
	DSP_CHANNEL_DOWNLOAD_DNS_NOT_EXIST:   errors.New("dsp channel of current dns does not exist"),
	DSP_CHANNEL_BALANCE_DNS_NOT_ENOUGH:   errors.New("dsp channel balance of current dns does not enough"),
	DSP_CHANNEL_WITHDRAW_WRONG_AMOUNT:    errors.New("dsp channel withdraw wrong amount"),
	DSP_CHANNEL_SYNCING:                  errors.New("dsp channel syncing"),
	DSP_CHANNEL_DNS_OFFLINE:              errors.New("dsp channel dns offline"),
	DSP_CHANNEL_NOT_EXIST:                errors.New("dsp channel not exist"),

	DSP_TASK_NOT_EXIST: errors.New("dsp task not exist"),

	DB_FIND_SHARE_RECORDS_FAILED:     errors.New("db find share records failed"),
	DB_SUM_SHARE_PROFIT_FAILED:       errors.New("db sum share profit failed"),
	DB_FIND_USER_SPACE_RECORD_FAILED: errors.New("db find user space record failed"),
	DB_ADD_USER_SPACE_RECORD_FAILED:  errors.New("db add user space record failed"),
	DB_GET_FILEINFO_FAILED:           errors.New("db get fileinfo failed"),

	NET_RECONNECT_PEER_FAILED:  errors.New("net reconnect peer failed"),
	NET_PROXY_DISCONNECTED:     errors.New("proxy has disconnted"),
	DSP_CHANNEL_OPEN_TO_NO_DNS: errors.New("dsp channel open to nodns"),
}
