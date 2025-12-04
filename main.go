package main

import (
	"context"
	"esmodmanager/lib"

	"fmt"

	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	const cfgPath = "mods.yaml"

	cmd := &cli.Command{
		Name:  "esmodmanager",
		Usage: "Manager for Everlasting Summer mods",
		Commands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List known mods",
				Action: func(ctx context.Context, c *cli.Command) error {
					db, err := lib.EnsureDB(cfgPath)
					if err != nil {
						return err
					}
					lib.ScanAndUpdate(db)
					lib.SaveDB(cfgPath, db)
					lib.PrintMods(db)
					return nil
				},
			},

			{
				Name:      "disable",
				Usage:     "Disable mod by numeric folder, codename, or ALL",
				ArgsUsage: "<id>",
				Action: func(ctx context.Context, c *cli.Command) error {
					return lib.ToggleEnabled(cfgPath, false, c.Args().First())
				},
			},

			{
				Name:      "enable",
				Usage:     "Enable mod by numeric folder, codename, or ALL",
				ArgsUsage: "<id>",
				Action: func(ctx context.Context, c *cli.Command) error {
					return lib.ToggleEnabled(cfgPath, true, c.Args().First())
				},
			},

			{
				Name:  "launch",
				Usage: "Launch game with current mod setup",
				Action: func(ctx context.Context, c *cli.Command) error {
					db, err := lib.EnsureDB(cfgPath)
					if err != nil {
						return err
					}

					if db.GameExe == "" {
						return fmt.Errorf("game_exe is empty in config. Set it in %s", cfgPath)
					}

					lib.ScanAndUpdate(db)
					lib.SaveDB(cfgPath, db)
					lib.LaunchWithMods(db)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println("Error:", err)
	}
}
