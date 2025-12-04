**[🇷🇺 Русский](README_RU.md) | [🇬🇧 English](README.md)**

## ESModManager-Go

A CLI utility for managing mods for the game **Everlasting Summer**, written in Go. It allows you to enable/disable mods, automatically detect mod information, launch the game with selected mods.

---

## Features

* **Extract information** from Ren'Py `.rpy` files
* **Fetch mod names via Steam API**
* **Enable/disable mods** by folder or codename

---

## Installation

```bash
go build
```
(Meson.build coming soon)

---

## Usage

### List mods

```bash
./esmodmanager list
```

### Enable a mod

```bash
./esmodmanager enable <folder|codename>
```

### Disable a mod

```bash
./esmodmanager disable <folder|codename>
```

### Enable/disable all mods

```bash
./esmodmanager enable ALL
./esmodmanager disable ALL
```

### Launch the game

```bash
./esmodmanager launch
```

---

## Configuration

The configuration file is created automatically on first launch: `mods.yaml`.

Example:

```yaml
game_exe: /usr/bin/steam
args:
  - -applaunch
  - "331470"
workshop_root: /path/to/SteamLibrary/.../331470
disabled_dir: /home/user/.elmod_disabled
mods:
  - name: Example Mod
    codename: example
    folder: 1234567890
    enabled: true
```

---

## How It Works

1. The program scans the Steam Workshop directory.
2. For each mod it attempts to detect:

   * codename from `.rpy` files
   * name
3. If no information is found locally, it queries the Steam API.
4. Disabled mods are temporarily moved to a separate directory.
5. After the game closes, everything is restored.

---

## Limitations

* Currently adapted only for Linux (possibly temporary)

---

## License

MIT
