package main

import (
	"fmt"
	"os"

	"github.com/weijun-sh/reconciliationSystem-bridge/cmd/utils"
	"github.com/weijun-sh/reconciliationSystem-bridge/log"
	"github.com/weijun-sh/reconciliationSystem-bridge/params"
	//"github.com/weijun-sh/reconciliationSystem-bridge/tokens"

	"github.com/weijun-sh/reconciliationSystem-bridge/worker"
	"github.com/urfave/cli/v2"
)

var (
	clientIdentifier = "reco"
	// Git SHA1 commit hash of the release (set via linker flags)
	gitCommit = ""
	gitDate   = ""
	// The app that holds all commands and flags.
	app = utils.NewApp(clientIdentifier, gitCommit, gitDate, "the reconciliation command line interface")
)

func initApp() {
	// Initialize the CLI app and start action
	app.Action = reco
	app.HideVersion = true // we have a command to print the version
	app.Copyright = "Copyright 2017-2020 The anyswap Authors"
	app.Commands = []*cli.Command{
		utils.LicenseCommand,
		utils.VersionCommand,
	}
	app.Flags = []cli.Flag{
		utils.ConfigFileFlag,
		utils.PriceFileFlag,
		utils.LogFileFlag,
		utils.VerbosityFlag,
		utils.JSONFormatFlag,
		utils.ColorFormatFlag,
	}
}

func main() {
	initApp()
	if err := app.Run(os.Args); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func reco(ctx *cli.Context) error {
	utils.SetLogger(ctx)
	if ctx.NArg() > 0 {
		return fmt.Errorf("invalid command: %q", ctx.Args().Get(0))
	}

	configFile := utils.GetConfigFilePath(ctx)
	params.LoadConfig(configFile)

	priceFile := utils.GetPriceFilePath(ctx)
	params.LoadPriceConfig(priceFile)

	worker.StartWork()

	utils.TopWaitGroup.Wait()
	//log.Info("reconciliation exit normally")
	return nil
}
