package common

const (
	DSP_DOWNLOAD_UNIT_PRICE         = 1   // dsp download price for byte
	DSP_URL_RAMDOM_NAME_LEN         = 8   // dsp custom name length
	MAX_PROGRESS_INFO_NUM           = 100 // max progress info num for get all progress api
	POLL_TX_COMFIRMED_TIMEOUT       = 10  // tx timeout
	MAX_SYNC_HEIGHT_OFFSET          = 50  // max sync height offset
	MAX_REG_CHANNEL_TIMES           = 5   // max register channel times
	MAX_REG_CHANNEL_BACKOFF         = 10  // max register channel backoff
	MAX_HEALTH_CHECK_INTERVAL       = 5   // health check interval
	MAX_STATE_CHANGE_CHECK_INTERVAL = 5   // state change  check interval
	EVENT_CHANGE_CHECK_INTERVAL     = 3   // event change check interval
	BLOCK_TIME                      = 1   // block time
	BLOCK_CONFIRM                   = 3   // block confirm
)

// default config
const (
	MAX_UNPAID_PAYMENT = 5 // max unpaid payment
)

// network common
const (
	MAX_WAIT_FOR_CONNECTED_TIMEOUT = 15      // wait for connected timeout
	COMPRESS_DATA_SIZE             = 1048576 // > 1MB data need to be compressed
	START_PROXY_TIMEOUT            = 20      // timeout for start proxy
	START_P2P_TIMEOUT              = 25      // timeout for start p2p
	BACKOFF_INIT_DELAY             = 2       // backoff initial delay
	BACKOFF_MAX_ATTEMPTS           = 50      // backoff max attempts
	KEEPALIVE_TIMEOUT              = 15      // keepalive timeout
	EVENT_ACTOR_TIMEOUT            = 15      // event actor timeout
	NETWORK_DIAL_TIMEOUT           = 5       // network dial timeout
)

// asset
const (
	SAVE_ASSET = "save"
)

const (
	EDGE_DB_NAME   = "client"
	DSP_DB_NAME    = "dsp"
	PYLONS_DB_NAME = "channel"
	SQLITE_DB_NAME = "edge-sqlite.db"
)

// Chain ID
const (
	DEVNET_CHAIN_ID  = "0"
	TESTNET_CHAIN_ID = "1"
	MAINNET_CHAIN_ID = "2"
)
