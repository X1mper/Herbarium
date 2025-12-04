package lib

import (
	"bufio"
	"esmodmanager/lib/steamapi"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func ScanAndUpdate(db *DB) {
	root := db.Root
	if root == "" {
		root = "/mnt/ssd2tb/SteamLibrary/steamapps/workshop/content/331470"
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		log.Println("scan error:", err)
		return
	}

	type result struct {
		Folder string
		Entry  ModEntry
	}

	results := make(chan result, len(entries))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	existingFolders := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		folder := e.Name()
		if folder == filepath.Base(getDisabledDir(db)) || strings.HasPrefix(folder, ".") {
			continue
		}
		existingFolders[folder] = true

		skip := false
		for _, m := range db.Mods {
			if m.Folder == folder && m.Name != "" {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		wg.Add(1)
		go func(folder string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fullPath := filepath.Join(root, folder)
			log.Println("Processing folder:", folder)
			codename, pretty := extractFromFolder(fullPath)
			log.Printf("Result for %s -> codename: %s, name: %s\n", folder, codename, pretty)

			results <- result{
				Folder: folder,
				Entry: ModEntry{
					Name:     pretty,
					CodeName: codename,
					Folder:   folder,
					Enabled:  true,
				},
			}
		}(folder)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	found := map[string]ModEntry{}
	for r := range results {
		found[r.Folder] = r.Entry
	}

	newList := []ModEntry{}
	seen := map[string]bool{}

	for _, m := range db.Mods {
		if !existingFolders[m.Folder] {
			log.Println("Removing missing folder from DB:", m.Folder)
			continue
		}
		if existing, ok := found[m.Folder]; ok {
			if m.Name != "" {
				existing.Name = m.Name
			}
			existing.Enabled = m.Enabled
			newList = append(newList, existing)
			seen[m.Folder] = true
		} else {
			newList = append(newList, m)
			seen[m.Folder] = true
		}
	}

	for folder, m := range found {
		if !seen[folder] {
			newList = append(newList, m)
		}
	}

	db.Mods = newList
}

func extractFromFolder(folder string) (codename, name string) {
	log.Println("Extracting from folder:", folder)
	codename, name = extractFromScripts(folder)

	if codename == "" || name == "" {
		id := filepath.Base(folder)
		if title, err := steamapi.FetchSteamTitle(id); err == nil && title != "" {
			log.Println("Steam API success for", id)
			if codename == "" {
				codename = id
			}
			if name == "" {
				name = title
			}
		} else {
			log.Println("Steam API failed for", id, "using folder name")
		}
	}

	if codename == "" {
		codename = filepath.Base(folder)
	}
	if name == "" {
		name = filepath.Base(folder)
	}

	return codename, name
}

func parseRpyFile(path string) (codename, pretty string) {
	f, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		m := modsLineRe.FindStringSubmatch(line)
		if len(m) == 3 {
			k := m[1]
			val := strings.TrimSpace(m[2])
			clean := braceRe.ReplaceAllString(val, "")
			clean = strings.ReplaceAll(clean, `\"`, `"`)
			clean = strings.Trim(clean, `"`)
			clean = strings.Join(strings.Fields(clean), " ")
			return k, clean
		}
	}
	return "", ""
}

func extractFromScripts(folder string) (codename, name string) {
	codename, name = "", ""
	filepath.WalkDir(folder, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		low := strings.ToLower(d.Name())
		if !strings.HasSuffix(low, ".rpy") {
			return nil
		}
		c, n := parseRpyFile(p)
		if c != "" && codename == "" {
			codename = c
		}

		if n != "" && name == "" {
			name = n
		}

		if codename != "" &&
			name != "" {
			return fs.SkipDir
		}
		return nil
	})
	return
}
