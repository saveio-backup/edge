package utils

import (
	"github.com/urfave/cli"
)

var (
	WalletFileFlag = cli.StringFlag{
		Name:  "wallet,w",
		Usage: "Import wallet from file",
	}
	ImportOnlineWalletFlag = cli.BoolFlag{
		Name:  "online",
		Usage: "Import for online node or not",
	}
	WalletPasswordFlag = cli.StringFlag{
		Name:  "password,p",
		Usage: "Create wallet password",
	}
	WalletLabelFlag = cli.StringFlag{
		Name:  "label,l",
		Usage: "Create wallet label",
		Value: "",
	}
	WalletKeyTypeFlag = cli.StringFlag{
		Name:  "keyType,k",
		Usage: "Create wallet keyType",
		Value: "ecdsa",
	}
	WalletCurveFlag = cli.StringFlag{
		Name:  "curve,c",
		Usage: "Create wallet curve",
		Value: "P-256",
	}
	WalletSchemeFlag = cli.StringFlag{
		Name:  "scheme,s",
		Usage: "Create wallet scheme",
		Value: "SHA256withECDSA",
	}
	WalletExportTypeFlag = cli.IntFlag{
		Name:  "type,t",
		Usage: "ExportType. 0: WalletFile, 1: PrivateKey",
	}
)
