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
	BLOCK_TIME                      = 5   // block time
)

// network common

const (
	MAX_WAIT_FOR_CONNECTED_TIMEOUT = 10 // wait for connected timeout
)
