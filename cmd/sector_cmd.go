package cmd

import (
	"encoding/json"

	"github.com/saveio/edge/cmd/flags"
	"github.com/saveio/edge/cmd/utils"
	fs "github.com/saveio/themis/smartcontract/service/native/savefs"
	"github.com/urfave/cli"
)

var SectorCommand = cli.Command{
	Name:  "sector",
	Usage: "Sector Management",
	Subcommands: []cli.Command{
		{
			Action:    createSector,
			Name:      "create",
			Usage:     "Create sector",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspSectorIdFlag,
				flags.DspSectorProveLevelFlag,
				flags.DspSectorSizeFlag,
				flags.DspSectorPlotFlag,
			},
			Description: "Create sector",
		},

		{
			Action:    deleteSector,
			Name:      "delete",
			Usage:     "Delete sector",
			ArgsUsage: "<sectorid>",
			Flags: []cli.Flag{
				flags.DspSectorIdFlag,
			},
			Description: "Delete sector",
		},
		{
			Action:    getSectorInfo,
			Name:      "getsector",
			Usage:     "Get sector info",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspSectorIdFlag,
			},
			Description: "Get sector info",
		},
		{
			Action:    getSectorInfosForNode,
			Name:      "getsectorsfornode",
			Usage:     "Get all sector info for a node",
			ArgsUsage: "[arguments...]",
			Flags: []cli.Flag{
				flags.DspWalletAddrFlag,
			},
			Description: "Get all sector info for a node",
		},
	},
	Description: `./dsp sector --help command to view help information.`,
}

//node command
func createSector(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspSectorIdFlag)) ||
		!ctx.IsSet(flags.GetFlagName(flags.DspSectorSizeFlag)) {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	sectorId := ctx.String(flags.GetFlagName(flags.DspSectorIdFlag))
	proveLevel := ctx.Uint64(flags.GetFlagName(flags.DspSectorProveLevelFlag))
	size := ctx.Uint64(flags.GetFlagName(flags.DspSectorSizeFlag))
	isPlot := ctx.Bool(flags.GetFlagName(flags.DspSectorPlotFlag))
	ret, err := utils.CreateSector(sectorId, proveLevel, size, isPlot)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func deleteSector(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspSectorIdFlag)) {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	sectorId := ctx.String(flags.GetFlagName(flags.DspSectorIdFlag))
	ret, err := utils.DeleteSector(sectorId)
	if err != nil {
		return err
	}
	PrintJsonData(ret)
	return nil
}

func getSectorInfo(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspSectorIdFlag)) {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	sectorId := ctx.String(flags.GetFlagName(flags.DspSectorIdFlag))
	ret, err := utils.GetSectorInfo(sectorId)
	if err != nil {
		return err
	}

	sectorInfo := new(fs.SectorInfo)
	err = json.Unmarshal(ret, sectorInfo)
	if err != nil {
		return err
	}

	PrintJsonObject(formatSectorInfo(sectorInfo))
	return nil

}

func getSectorInfosForNode(ctx *cli.Context) error {
	if !ctx.IsSet(flags.GetFlagName(flags.DspWalletAddrFlag)) {
		PrintErrorMsg("Missing argument.")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}

	addr := ctx.String(flags.GetFlagName(flags.DspWalletAddrFlag))
	ret, err := utils.GetSectorInfosForNode(addr)
	if err != nil {
		return err
	}

	sectorInfos := new(fs.SectorInfos)
	err = json.Unmarshal(ret, sectorInfos)
	if err != nil {
		return err
	}

	PrintJsonObject(formatSectorInfosForNode(sectorInfos))
	return nil
}

type SectorInfo struct {
	NodeAddr         string
	SectorID         uint64
	Size             uint64
	ProveLevel       uint64
	FirstProveHeight uint64
	NextProveHeight  uint64
	TotalBlockNum    uint64
	FileNum          uint64
	GroupNum         uint64
	IsPlots          bool
	FileList         []string
}

func formatSectorInfo(sectorInfo *fs.SectorInfo) *SectorInfo {
	info := &SectorInfo{
		NodeAddr:         sectorInfo.NodeAddr.ToBase58(),
		SectorID:         sectorInfo.SectorID,
		Size:             sectorInfo.Size,
		ProveLevel:       sectorInfo.ProveLevel,
		FirstProveHeight: sectorInfo.FirstProveHeight,
		NextProveHeight:  sectorInfo.NextProveHeight,
		TotalBlockNum:    sectorInfo.TotalBlockNum,
		GroupNum:         sectorInfo.GroupNum,
		IsPlots:          sectorInfo.IsPlots,
		FileList:         nil,
	}

	info.FileNum = uint64(len(sectorInfo.FileList.List))
	for _, fileHash := range sectorInfo.FileList.List {
		info.FileList = append(info.FileList, string(fileHash.Hash))
	}

	return info
}

type SectorInfos struct {
	SectorCount uint64
	SectorInfos []*SectorInfo
}

func formatSectorInfosForNode(sectorInfos *fs.SectorInfos) *SectorInfos {
	info := &SectorInfos{
		SectorCount: sectorInfos.SectorCount,
		SectorInfos: nil,
	}

	for _, sectorInfo := range sectorInfos.Sectors {
		info.SectorInfos = append(info.SectorInfos, formatSectorInfo(sectorInfo))
	}

	return info
}
