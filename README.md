**[🇷🇺 Русский](README_RU.md) | [🇬🇧 English](README.md)**

# Herbarium

A mod manager for the game **Everlasting Summer**, written in Go, including a CLI utility and a modern GUI on GTK4 / LibAdwaita.
Allows enabling and disabling mods, automatically detecting mod information, and launching the game with only the desired mods.

![logo](data/icons/hicolor/scalable/apps/ru.ximper.Herbarium.svg)
![mainwindow](data/images/1-mainwindow.png)

---

## Features

* **Extract information** from Ren'Py `.rpy` files
* **Fetch mod names via Steam API**
* **Enable/disable mods** by folder or codename

---

## Installation

```bash
meson setup _build --prefix=/usr
meson install -C _build/
```

---

## CLI Usage

### Show mod list
```bash
herbarium-cli list
```

### Enable a mod
```bash
herbarium-cli enable <folder|codename>
```

### Disable a mod
```bash
herbarium-cli disable <folder|codename>
```

### Enable/disable all mods
```bash
herbarium-cli enable ALL
herbarium-cli disable ALL
```

### Launch the game
```bash
herbarium-cli launch
```

---

## Configuration

On first launch, the program automatically creates two files in `$HOME/.config/ru.ximper.Herbarium`:

`config.yaml` — program configuration

Contains game launch and path settings.

```yaml
game_exe: /usr/bin/steam
args:
  - -applaunch
  - "331470"
workshop_root: /home/user/.steam/steam/steamapps/workshop/content/331470
disabled_dir: /home/user/.elmod_disabled
```

`mods_db.yaml` — mods database
Stores detected mods and their state.

```yaml
mods:
  - name: Example Mod
    codename: example
    folder: "1234567890"
    enabled: true
    discovered_at: 2024-12-17T13:42:11Z
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
