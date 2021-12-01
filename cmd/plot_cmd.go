package cmd

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	"github.com/saveio/edge/common/config"
	eUtils "github.com/saveio/edge/utils"
	"github.com/saveio/edge/utils/plot"
	"github.com/saveio/themis/account"
	"github.com/saveio/themis/common/log"
	"github.com/urfave/cli"
)

var PlotCommand = cli.Command{
	Name:  "plot",
	Usage: "Plot file",
	Subcommands: []cli.Command{
		{
			Action:    generatePlotFile,
			Name:      "generate",
			Usage:     "Generate plotfile",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.PlotSystemFlag,
				flags.PlotNumericIDFlag,
				flags.PlotStartNonceFlag,
				flags.PlotNoncesFlag,
				flags.PlotNumFlag,
				flags.PlotSizeFlag,
				flags.PlotPathFlag,
			},
			Description: "Generate plotfile",
		},
		{
			Action:    listPlotFile,
			Name:      "list",
			Usage:     "list plotfile",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.PlotPathFlag,
			},
			Description: "list plotfile",
		},
		{
			Action:    addPlotFile,
			Name:      "mine",
			Usage:     "add plot file to mine",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.PlotTaskIdFlag,
				flags.PlotPathFlag,
				flags.CreateSectorFlag,
			},
			Description: "add plotfile to mine",
		},
		{
			Action:      getAllProvedPlotFile,
			Name:        "list-all-proved",
			Usage:       "List all proved plots",
			ArgsUsage:   "[arguments...]",
			Flags:       []cli.Flag{},
			Description: "List all proved plots",
		},

		{
			Action:      getAllPlotTasks,
			Name:        "list-tasks",
			Usage:       "List all plot tasks",
			ArgsUsage:   "[arguments...]",
			Flags:       []cli.Flag{},
			Description: "List all plot tasks",
		},
	},
	Description: `./dsp plot --help command to view help information.`,
}

// plot command
func generatePlotFile(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.PlotSystemFlag)) {
		PrintErrorMsg("Missing argument --system.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	system := ctx.String(flags.GetFlagName(flags.PlotSystemFlag))
	numericId := ctx.String(flags.GetFlagName(flags.PlotNumericIDFlag))
	if len(numericId) == 0 {
		numericId, _ = numericdFromDefaultWallet(ctx)
		if len(numericId) == 0 {
			PrintErrorMsg("Missing argument. --numericId or wallet not found")
			cli.ShowSubcommandHelp(ctx)
			return nil
		}
	}
	path := ctx.String(flags.GetFlagName(flags.PlotPathFlag))
	start := ctx.Uint64(flags.GetFlagName(flags.PlotStartNonceFlag))
	nonces := ctx.Uint64(flags.GetFlagName(flags.PlotNoncesFlag))
	size := ctx.Uint64(flags.GetFlagName(flags.PlotSizeFlag))
	if len(path) == 0 {
		path = config.PlotPath()
	}
	if nonces == 0 {
		nonces = size / plot.DEFAULT_PLOT_SIZEKB
		if nonces == 0 || nonces%8 != 0 {
			PrintErrorMsg("Invalid argument. size should be an integer multiple of 2048")
			cli.ShowSubcommandHelp(ctx)
			return nil
		}
		var err error
		start, err = plot.GetMinStartNonce(numericId, path)
		if err != nil {
			PrintErrorMsg(err.Error())
			return nil
		}
	}

	if !ctx.IsSet(flags.GetFlagName(flags.PlotStartNonceFlag)) &&
		!ctx.IsSet(flags.GetFlagName(flags.PlotNoncesFlag)) &&
		size == 0 {
		PrintErrorMsg("Missing argument. --nonce --startNonce")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	if nonces == 0 || nonces%8 != 0 {
		PrintErrorMsg("Invalid argument. nonces should be an integer multiple of 8")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	num := ctx.Uint64(flags.GetFlagName(flags.PlotNumFlag))
	for i := uint64(0); i < num; i++ {
		log.Infof("system %v, numericId %v, path %v, start %v, nonces %v", system, numericId, path, start, nonces)
		ret, err := utils.GeneratePlotFile(system, numericId, path, start, nonces)
		if err != nil {
			PrintErrorMsg(err.Error())
			return nil
		}
		PrintJsonData(ret)
		start += nonces
	}

	return nil
}

func listPlotFile(ctx *cli.Context) error {

	path := ctx.String(flags.GetFlagName(flags.PlotPathFlag))
	if len(path) == 0 {
		path = config.PlotPath()
	}
	hexPath := hex.EncodeToString([]byte(path))
	ret, err := utils.GetAllPlotFile(string(hexPath))
	if err != nil {
		PrintErrorMsg(err.Error())
		return nil
	}
	PrintJsonData(ret)
	return nil
}

func numericdFromDefaultWallet(ctx *cli.Context) (string, error) {
	optionFile := checkWalletFileName(ctx)
	wallet, err := account.Open(optionFile)
	if err != nil {
		return "", fmt.Errorf("open wallet:%s error:%s", optionFile, err)
	}
	accNum := wallet.GetAccountNum()
	if accNum == 0 {
		return "", fmt.Errorf("no account for wallet")
	}
	accMeta := wallet.GetAccountMetadataByIndex(1)
	return fmt.Sprintf("%v", eUtils.WalletAddressToId([]byte(accMeta.Address))), nil
}

func addPlotFile(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.PlotPathFlag)) {
		PrintErrorMsg("Missing argument --path.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	path := ctx.String(flags.GetFlagName(flags.PlotPathFlag))
	taskId := ctx.String(flags.GetFlagName(flags.PlotTaskIdFlag))
	createSector := ctx.Bool(flags.GetFlagName(flags.CreateSectorFlag))

	info, err := os.Stat(path)
	if err != nil {
		PrintErrorMsg(err.Error())
		return nil
	}
	if info.IsDir() {
		ret, err := utils.AddPlotFiles(path, createSector)
		if err != nil {
			PrintErrorMsg(err.Error())
			return nil
		}
		PrintJsonData(ret)
	} else {
		ret, err := utils.AddPlotFile(taskId, path, createSector)
		if err != nil {
			PrintErrorMsg(err.Error())
			return nil
		}
		PrintJsonData(ret)
	}
	return nil
}

func getAllProvedPlotFile(ctx *cli.Context) error {

	ret, err := utils.GetAllProvedPlotFile()
	log.Infof("getAllProvedPlotFile cmd %v, err %v", ret, err)
	if err != nil {
		PrintErrorMsg(err.Error())
		return nil
	}
	PrintJsonData(ret)
	return nil
}

func getAllPlotTasks(ctx *cli.Context) error {

	ret, err := utils.GetAllPlotTasks()
	log.Infof("GetAllPlotTasks cmd %v, err %v", ret, err)
	if err != nil {
		PrintErrorMsg(err.Error())
		return nil
	}
	PrintJsonData(ret)
	return nil
}
