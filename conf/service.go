package conf

type ServiceYAML struct {
	List          []string `yaml:"list"`
	Stop          bool     `yaml:"stop"`
	Restore       bool     `yaml:"restore"`
	StartNew      bool     `yaml:"startNew"`
	ReloadCommand []string `yaml:"reloadCommand"`
	StopCommand   []string `yaml:"stopCommand"`
	StartCommand  []string `yaml:"startCommand"`
	ManageRSync   bool     `yaml:"manageRSync"`
}
