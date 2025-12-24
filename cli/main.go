package main

import (
	"context"
	"fmt"
	"os"

	"herbarium/lib"

	"github.com/urfave/cli/v3"
)

func main() {
	lib.InitLocales()

	cmd := &cli.Command{
		Name:  "herbarium",
		Usage: lib.T_("Manager for Everlasting Summer mods"),
		Commands: []*cli.Command{
			{
				Name:    "list",
				Aliases: []string{"ls"},
				Usage:   lib.T_("List known mods"),
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg, err := lib.EnsureConfig()
					if err != nil {
						return err
					}

					db, err := lib.EnsureModsDB()
					if err != nil {
						return err
					}

					lib.ScanAndUpdate(cfg, db)

					if err := lib.SaveModsDB(db); err != nil {
						return err
					}

					lib.PrintMods(db)
					return nil
				},
			},

			{
				Name:      "disable",
				Aliases:   []string{"d"},
				Usage:     lib.T_("Disable mod by numeric folder, codename, or ALL"),
				ArgsUsage: "<id>",
				Action: func(ctx context.Context, c *cli.Command) error {
					return lib.ToggleEnabled(false, c.Args().First())
				},
			},

			{
				Name:      "enable",
				Aliases:   []string{"e"},
				Usage:     lib.T_("Enable mod by numeric folder, codename, or ALL"),
				ArgsUsage: "<id>",
				Action: func(ctx context.Context, c *cli.Command) error {
					return lib.ToggleEnabled(true, c.Args().First())
				},
			},

			{
				Name:    "launch",
				Aliases: []string{"start", "l"},
				Usage:   lib.T_("Launch game with current mod setup"),
				Action: func(ctx context.Context, c *cli.Command) error {
					cfg, err := lib.EnsureConfig()
					if err != nil {
						return err
					}

					if cfg.GameExe == "" {
						return fmt.Errorf(lib.T_("game_exe is empty in config"))
					}

					db, err := lib.EnsureModsDB()
					if err != nil {
						return err
					}

					lib.ScanAndUpdate(cfg, db)

					if err := lib.SaveModsDB(db); err != nil {
						return err
					}

					lib.LaunchWithMods(cfg, db)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Println(lib.T_("Error:"), err)
		os.Exit(1)
	}
}
