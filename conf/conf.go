package conf

type ConfigYAML struct {
	Mode          string       `yaml:"mode"`
	StoreFile     string       `yaml:"storeFile"`
	Services      ServiceYAML  `yaml:"services"`
	Net           NetYAML      `yaml:"net"`
	Security      SecurityYAML `yaml:"security"`
	TriggerReboot bool         `yaml:"triggerReboot"`
	RebootCommand []string     `yaml:"rebootCommand"`
	RSyncCommand  []string     `yaml:"rsyncCommand"`
	TarCommand    []string     `yaml:"tarCommand"`
	UnTarCommand  []string     `yaml:"unTarCommand"`
}

func (c ConfigYAML) GetMode() Mode {
	switch c.Mode {
	case string(Backup):
		return Backup
	case string(Restore):
		return Restore
	case string(Store):
		return Store
	case string(UnStore):
		return UnStore
	default:
		return Unknown
	}
}
