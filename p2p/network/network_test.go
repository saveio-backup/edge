package network

import (
	"fmt"
	"testing"
)

func TestWalletValid(t *testing.T) {
	walletAddr := ""
	fmt.Printf("%s valid %t\n", walletAddr, isValidWalletAddr(walletAddr))
	walletAddr = "tcp://127.0.0.1:10338"
	fmt.Printf("%s valid %t\n", walletAddr, isValidWalletAddr(walletAddr))
	walletAddr = "127.0.0.1:10338"
	fmt.Printf("%s valid %t\n", walletAddr, isValidWalletAddr(walletAddr))
	walletAddr = "0a0d2ab32271f41f454f56263e21226fad41d4634c8a9002b1c2df82b27fe2f9"
	fmt.Printf("%s valid %t\n", walletAddr, isValidWalletAddr(walletAddr))
	walletAddr = "0a0d2ab32271f41f454f56263e21226fad41d4634c8a9002b1c2df82b2fe2f9"
	fmt.Printf("%s valid %t\n", walletAddr, isValidWalletAddr(walletAddr))
	walletAddr = "AHjjdbVLhfTyiNFEq2X8mFnnirZY1yK8Rq"
	fmt.Printf("%s valid %t\n", walletAddr, isValidWalletAddr(walletAddr))
}
