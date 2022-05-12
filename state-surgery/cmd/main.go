package main

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
	"math/big"
	"os"
	surgery "state-surgery"
)

func main() {
	log.Root().SetHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(isatty.IsTerminal(os.Stderr.Fd()))))

	app := &cli.App{
		Name:  "surgery",
		Usage: "migrates data from v0 to Bedrock",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "data-dir",
				Aliases:  []string{"d"},
				Usage:    "data directory to read",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "state-root",
				Aliases: []string{"r"},
				Usage:   "state root to dump",
			},
			&cli.StringFlag{
				Name:     "genesis-file",
				Aliases:  []string{"g"},
				Usage:    "path to a genesis file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "out-dir",
				Aliases:  []string{"o"},
				Usage:    "path to output directory",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "expected-total-supply",
				Aliases:  []string{"e"},
				Usage:    "expected total ETH supply",
				Required: true,
			},
		},
		Action: action,
	}

	if err := app.Run(os.Args); err != nil {
		log.Crit("error in migration", "err", err)
	}
}

func action(cliCtx *cli.Context) error {
	dataDir := cliCtx.String("data-dir")
	stateRoot := cliCtx.String("state-root")
	outDir := cliCtx.String("out-dir")
	genesisPath := cliCtx.String("genesis-file")
	expectedSupplyStr := cliCtx.String("expected-total-supply")
	expectedSupply, ok := new(big.Int).SetString(expectedSupplyStr, 10)
	if !ok {
		return errors.New("invalid total supply")
	}

	genesis, err := surgery.ReadGenesisFromFile(genesisPath)
	if err != nil {
		return err
	}

	var stateRootHash common.Hash
	if stateRoot != "" {
		stateRootHash = common.HexToHash(stateRoot)
	}
	return surgery.Migrate(dataDir, stateRootHash, genesis, outDir, expectedSupply)
}
