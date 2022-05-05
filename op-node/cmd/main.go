package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum-optimism/optimism/op-node/flags"

	opnode "github.com/ethereum-optimism/optimism/op-node"
	"github.com/ethereum-optimism/optimism/op-node/node"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli"
)

var (
	Version     = "0.0.0"
	GitCommit   = ""
	GitDate     = ""
	VersionMeta = "dev"
)

// VersionWithMeta holds the textual version string including the metadata.
var VersionWithMeta = func() string {
	v := Version
	if GitCommit != "" {
		v += "-" + GitCommit[:8]
	}
	if GitDate != "" {
		v += "-" + GitDate
	}
	if VersionMeta != "" {
		v += "-" + VersionMeta
	}
	return v
}()

func main() {
	// Set up logger with a default INFO level in case we fail to parse flags,
	// otherwise the final critical log won't show what the parsing error was.
	log.Root().SetHandler(
		log.LvlFilterHandler(
			log.LvlInfo,
			log.StreamHandler(os.Stdout, log.TerminalFormat(true)),
		),
	)

	app := cli.NewApp()
	app.Flags = flags.Flags
	app.Version = VersionWithMeta
	app.Name = "op-node"
	app.Usage = "Optimism Rollup Node"
	app.Description = "The deposit only rollup node drives the L2 execution engine based on L1 deposits."

	app.Action = RollupNodeMain
	err := app.Run(os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

func RollupNodeMain(ctx *cli.Context) error {
	log.Info("Initializing Rollup Node")
	cfg, err := opnode.NewConfig(ctx)
	if err != nil {
		log.Error("Unable to create the rollup node config", "error", err)
		return err
	}
	logCfg, err := opnode.NewLogConfig(ctx)
	if err != nil {
		log.Error("Unable to create the log config", "error", err)
		return err
	}

	n, err := node.New(context.Background(), cfg, logCfg.NewLogger(), VersionWithMeta)
	if err != nil {
		log.Error("Unable to create the rollup node", "error", err)
		return err
	}
	log.Info("Starting rollup node")

	if err := n.Start(context.Background()); err != nil {
		log.Error("Unable to start rollup node", "error", err)
		return err
	}
	defer n.Stop()

	log.Info("Rollup node started")

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, []os.Signal{
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}...)
	<-interruptChannel

	return nil

}
