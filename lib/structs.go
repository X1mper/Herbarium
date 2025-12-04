package lib

type ModEntry struct {
	Name     string `yaml:"name"`
	CodeName string `yaml:"codename"`
	Folder   string `yaml:"folder"`
	Enabled  bool   `yaml:"enabled"`
}

type DB struct {
	GameExe     string     `yaml:"game_exe"`
	Args        []string   `yaml:"args,omitempty"`
	Root        string     `yaml:"workshop_root"`
	DisabledDir string     `yaml:"disabled_dir"`
	Mods        []ModEntry `yaml:"mods"`
}
