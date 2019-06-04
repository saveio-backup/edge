package cmd

import (
	"github.com/saveio/edge/common/config"
	"github.com/saveio/themis/cmd/utils"
	"github.com/urfave/cli"
)

func SetRpcPort(ctx *cli.Context) {
	if ctx.IsSet(utils.GetFlagName(utils.RPCPortFlag)) {
		config.Parameters.BaseConfig.JsonRpcPortOffset = ctx.Int(utils.GetFlagName(utils.RPCPortFlag))
	}
}
