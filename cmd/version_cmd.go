package cmd

import (
	"encoding/json"

	"github.com/saveio/carrier/network"
	dspSdk "github.com/saveio/dsp-go-sdk/dsp"
	"github.com/saveio/edge/dsp"
	"github.com/saveio/max/max"
	"github.com/saveio/pylons"
	"github.com/urfave/cli"
)

var VersionCommand = cli.Command{
	Name:        "version",
	Usage:       "show version info",
	Subcommands: []cli.Command{},
	Action:      showVersion,
}

func showVersion(ctx *cli.Context) error {
	versionM := make(map[string]string)
	versionM["Edge"] = dsp.Version
	versionM["Dsp-go-sdk"] = dspSdk.Version
	versionM["Pylons"] = pylons.Version
	versionM["Max"] = max.Version
	versionM["Carrier"] = network.Version
	jsonData, err := json.Marshal(versionM)
	if err != nil {
		return err
	}
	PrintJsonData(jsonData)
	return nil
}
