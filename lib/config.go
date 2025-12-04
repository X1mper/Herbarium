package lib

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gopkg.in/yaml.v3"
)

var (
	modsLineRe = regexp.MustCompile(
		`(?m)^\s*\$?\s*mods\s*\[\s*["']([^"'\]]+)["']\s*\]\s*=\s*u?["']([\s\S]*?)["']`,
	)
	braceRe = regexp.MustCompile(`\{[^}]*\}`)
)

func EnsureDB(cfgPath string) (*DB, error) {
	homeDir, _ := os.UserHomeDir()

	db, err := loadDB(cfgPath)
	if err == nil {
		return db, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	db = &DB{
		GameExe:     "/usr/bin/steam",
		Args:        []string{"-applaunch", "331470"},
		Root:        filepath.Join(homeDir, ".steam/steam/steamapps/workshop/content/331470"),
		DisabledDir: filepath.Join(homeDir, ".elmod_disabled"),
		Mods:        []ModEntry{},
	}
	SaveDB(cfgPath, db)
	fmt.Println("Created new config at", cfgPath)
	return db, nil
}

func ToggleEnabled(cfgPath string, enable bool, id string) error {
	if id == "" {
		return fmt.Errorf("provide folder id, codename, or ALL")
	}

	db, err := EnsureDB(cfgPath)
	if err != nil {
		return err
	}

	ScanAndUpdate(db)

	if id == "ALL" {
		setAllEnabled(db, enable)
	} else {
		if err := setEnabled(db, id, enable); err != nil {
			return err
		}
	}

	SaveDB(cfgPath, db)

	action := map[bool]string{true: "Enabled", false: "Disabled"}[enable]
	fmt.Printf("%s %s\n", action, id)
	return nil
}

func loadDB(path string) (*DB, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var db DB
	if err := yaml.Unmarshal(b, &db); err != nil {
		return nil, err
	}
	return &db, nil
}

func SaveDB(path string, db *DB) error {
	b, err := yaml.Marshal(db)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0644)
}

func setEnabled(db *DB, id string, enabled bool) error {
	idx := -1
	for i, m := range db.Mods {
		if m.Folder == id || m.CodeName == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("mod not found: %s", id)
	}
	db.Mods[idx].Enabled = enabled
	return nil
}

func setAllEnabled(db *DB, enabled bool) {
	for i := range db.Mods {
		db.Mods[i].Enabled = enabled
	}
}
