package common

type HostAddr struct {
	Protocol string
	Address  string
	Port     string
}

type UserspaceTransferType uint

const (
	TransferTypeNone UserspaceTransferType = iota
	TransferTypeIn
	TransferTypeOut
)
