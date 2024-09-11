package conf

type ExcludeProtectYAML struct {
	ProtectCommand    []string `yaml:"protectCommand"`
	UnProtectCommand  []string `yaml:"unProtectCommand"`
	StdOutBuffStdInOn bool     `yaml:"stdOutBuffStdInOn"`
}
