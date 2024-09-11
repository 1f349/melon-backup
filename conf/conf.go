package conf

import "github.com/1f349/melon-backup/utils"

var Debug bool

type ConfigYAML struct {
	Mode              string             `yaml:"mode"`
	StoreFile         string             `yaml:"storeFile"`
	Services          ServiceYAML        `yaml:"services"`
	Net               NetYAML            `yaml:"net"`
	Security          SecurityYAML       `yaml:"security"`
	ExcludeProtection ExcludeProtectYAML `yaml:"excludeProtection"`
	TriggerReboot     bool               `yaml:"triggerReboot"`
	RebootCommand     []string           `yaml:"rebootCommand"`
	RSyncCommand      []string           `yaml:"rsyncCommand"`
	TarCommand        []string           `yaml:"tarCommand"`
	UnTarCommand      []string           `yaml:"unTarCommand"`
	TarBufferSize     uint32             `yaml:"tarBufferSize"`
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

func (c ConfigYAML) GetTarBufferSize() uint32 {
	if c.TarBufferSize < 256 {
		return 256
	}
	return c.TarBufferSize
}

func (c ConfigYAML) GetStoreFile() string {
	return getAbsPath(utils.GetCWD(), c.StoreFile)
}
