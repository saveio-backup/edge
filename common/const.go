package common

const (
	DSP_DOWNLOAD_UNIT_PRICE         = 1    // dsp download price for byte
	DSP_URL_RAMDOM_NAME_LEN         = 8    // dsp custom name length
	MAX_PROGRESS_INFO_NUM           = 100  // max progress info num for get all progress api
	POLL_TX_COMFIRMED_TIMEOUT       = 10   // tx timeout
	MAX_SYNC_HEIGHT_OFFSET          = 50   // max sync height offset
	MAX_REG_CHANNEL_TIMES           = 5    // max register channel times
	MAX_REG_CHANNEL_BACKOFF         = 10   // max register channel backoff
	MAX_HEALTH_CHECK_INTERVAL       = 5    // health check interval
	MAX_STATE_CHANGE_CHECK_INTERVAL = 5    // state change  check interval
	EVENT_CHANGE_CHECK_INTERVAL     = 3    // event change check interval
	BLOCK_TIME                      = 1    // block time
	BLOCK_CONFIRM                   = 1    // block confirm
	BLOCK_DELAY                     = 3    // block delay
	MAX_CACHE_SIZE                  = 1000 // max cache size
)

// default config
const (
	DEFAULT_MAX_UNPAID_PAYMENT    = 5               // max unpaid payment
	DEFAULT_MAX_LOG_SIZE          = 5 * 1024 * 1024 // max log size
	DEFAULT_MAX_DNS_NODE_NUM      = 100             // max dns node num
	DEFAULT_MAX_UPLOAD_TASK_NUM   = 10000           // max upload task num
	DEFAULT_MAX_DOWNLOAD_TASK_NUM = 10000           // max share task num
	DEFAULT_MAX_SHARE_TASK_NUM    = 10000           // max download task num
	DEFAULT_SEED_INTERVAL         = 600             // max seed service interval
	DEFAULT_TRACKER_PROTOCOL      = "tcp"           // default tracker protocol
	DEFAULT_TRACKER_PORT_OFFSET   = 337             // tracker port offset
	DEFAULT_WS_PORT_OFFSET        = 339             // tracker port offset
	DEFAULT_PLOT_PATH             = "./plots"       // default plot path
)

// network common
const (
	MAX_WAIT_FOR_CONNECTED_TIMEOUT = 15               // wait for connected timeout
	COMPRESS_DATA_SIZE             = 1048576          // > 1MB data need to be compressed
	START_PROXY_TIMEOUT            = 20               // timeout for start proxy
	START_P2P_TIMEOUT              = 25               // timeout for start p2p
	BACKOFF_INIT_DELAY             = 2                // backoff initial delay
	BACKOFF_MAX_ATTEMPTS           = 50               // backoff max attempts
	KEEPALIVE_TIMEOUT              = 15               // keepalive timeout
	EVENT_ACTOR_TIMEOUT            = 15               // event actor timeout
	NETWORK_DIAL_TIMEOUT           = 5                // network dial timeout
	MAX_MSG_RETRY                  = 1                // max msg retry count
	ACK_MSG_CHECK_INTERVAL         = 20               // ack msg check interval
	MAX_ACK_MSG_TIMEOUT            = 60               // max timeout for ack msg
	MAX_RECEIVED_MSG_CACHE         = 500              // max received msg cache
	SESSION_SPEED_RECORD_INTERVAL  = 3                // speed record interval
	MAX_SESSION_RECORD_SPEED_LEN   = 3                // max record speed array length
	MAX_WRITE_BUFFER_SIZE          = 24 * 1024 * 1024 // max write buffer size 24MB
	MAX_MSG_RETRY_FAILED           = 3                // max msg retry failed count
	MAX_PEER_RECONNECT_TIMEOUT     = 60               // peer reconnect timeout
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
