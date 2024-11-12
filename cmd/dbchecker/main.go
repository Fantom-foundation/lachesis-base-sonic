package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/Fantom-foundation/lachesis-base/abft"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/urfave/cli/v2"
)

var (
	DbPathFlag = cli.StringFlag{
		Name:     "db",
		Usage:    "sqlite3 event db path",
		Required: true,
	}
	EpochMinFlag = cli.UintFlag{
		Name:  "epoch.min",
		Usage: "Lower bound (inclusive) for epochs to be checked",
	}
	EpochMaxFlag = cli.UintFlag{
		Name:  "epoch.max",
		Usage: "Upper bound (inclusive) for epochs to be checked",
	}
)

func main() {
	app := &cli.App{
		Name:        "Event DB Checker",
		Description: "Consensus regression testing tool",
		Copyright:   "(c) 2024 Fantom Foundation",
		Flags:       []cli.Flag{&DbPathFlag, &EpochMinFlag, &EpochMaxFlag},
		Action:      run,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx *cli.Context) error {
	conn, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", ctx.String(DbPathFlag.Name)))
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.Ping(); err != nil {
		return err
	}

	epochMin, epochMax, err := abft.GetEpochRange(conn)
	if err != nil {
		return err
	}
	if ctx.IsSet(EpochMinFlag.Name) {
		epochMin = max(epochMin, idx.Epoch(ctx.Uint(EpochMinFlag.Name)))
	}
	if ctx.IsSet(EpochMaxFlag.Name) {
		epochMax = min(epochMax, idx.Epoch(ctx.Uint(EpochMaxFlag.Name)))
	}
	if epochMin > epochMax {
		return fmt.Errorf("invalid range of epochs requested: [%d, %d]", epochMin, epochMax)
	}

	for epoch := epochMin; epoch <= epochMax; epoch++ {
		if err := abft.CheckEpochAgainstDB(conn, epoch); err != nil {
			return err
		}
	}
	return nil
}
