package lib

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func getDisabledDir(db *DB) string {
	homeDir, _ := os.UserHomeDir()
	dir := db.DisabledDir
	if dir == "" {
		dir = filepath.Join(homeDir, ".elmod_disabled")
	}
	return dir
}

func restoreMoved(pairs [][2]string) error {
	for i := len(pairs) - 1; i >= 0; i-- {
		src := pairs[i][1]
		dst := pairs[i][0]

		if _, err := os.Stat(dst); err == nil {
			fmt.Println("warning: destination exists, skipping restore:", dst)
			continue
		}

		if err := os.Rename(src, dst); err != nil {
			if err := copyDir(src, dst); err != nil {
				return err
			}

			if err := os.RemoveAll(src); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func moveDisabledMods(db *DB) (moved [][2]string, err error) {
	root := db.Root
	disdir := getDisabledDir(db)

	if err := os.MkdirAll(disdir, 0755); err != nil {
		return nil, err
	}

	copy := false

	for _, m := range db.Mods {
		if !m.Enabled {
			src := filepath.Join(root, m.Folder)
			dst := filepath.Join(disdir, m.Folder)

			if _, err := os.Stat(src); os.IsNotExist(err) {
				continue
			}

			if err := os.Rename(src, dst); err != nil {
				if copy == false {
					fmt.Printf("\033[31mERROR:\033[0m Your game folder and the disabled mods folder are on different drives.\n")
					fmt.Printf("\033[31mMoving via os.Rename failed.\033[0m\n")
					fmt.Printf("\033[31mMoving large folders frequently across drives can cause extra wear on SSD/HDD.\033[0m\n")
					fmt.Printf("\033[31mConsider changing disabled_dir in the config or confirm you want to continue.\033[0m\n")
					fmt.Print("Continue? [Y/n]: ")

					reader := bufio.NewReader(os.Stdin)
					input, _ := reader.ReadString('\n')
					input = strings.TrimSpace(input)
					if input != "" && strings.ToLower(input) != "y" {
						fmt.Println("Operation cancelled by user.")
						return nil, fmt.Errorf("operation cancelled by user")
					} else {
						copy = true
					}
				}

				if err := copyDir(src, dst); err != nil {
					return moved, err
				}

				if err := os.RemoveAll(src); err != nil {
					return moved, err
				}
			}
			moved = append(moved, [2]string{src, dst})
		}
	}
	return moved, nil
}

func isProcessRunning(name string) (bool, error) {
	out, err := exec.Command("ps", "ax").Output()
	if err != nil {
		return false, err
	}

	lines := strings.SplitSeq(string(out), "\n")
	for l := range lines {
		if strings.Contains(l, name) {
			return true, nil
		}
	}
	return false, nil
}

func launchGame(exe string, args []string) error {
	cmd := exec.Command(exe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	fmt.Println("Launching via Steam:", exe, args)

	if err := cmd.Start(); err != nil {
		return err
	}

	target := "Everlasting Sum"

	for {
		running, _ := isProcessRunning(target)
		if running {
			break
		}
		time.Sleep(time.Second)
	}
	for {
		running, _ := isProcessRunning(target)
		if !running {
			break
		}
		time.Sleep(time.Second)
	}
	fmt.Println("Target process exited.")
	return nil
}

func LaunchWithMods(db *DB) {
	moved, err := moveDisabledMods(db)
	if err != nil {
		fmt.Println("error disabling mods:", err)
		os.Exit(1)
	}

	restore := func() {
		if len(moved) > 0 {
			fmt.Println("Restoring", len(moved), "folders...")
			if err := restoreMoved(moved); err != nil {
				fmt.Println("restore error:", err)
			}
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Interrupted — restoring...")
		restore()
		os.Exit(1)
	}()

	if err := launchGame(db.GameExe, db.Args); err != nil {
		fmt.Println("game launch error:", err)
	}

	restore()
	fmt.Println("Game exited — mods restored.")
}
