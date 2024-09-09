package conf

type SecurityYAML struct {
	PublicCert    string `yaml:"publicCert"`
	PrivateKey    string `yaml:"privateKey"`
	CACert        string `yaml:"caCert"`
	CACertDir     string `yaml:"caCertDirectory"`
	RSyncPassword string `yaml:"rsyncPassword"`
}
