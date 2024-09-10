package conf

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/1f349/melon-backup/utils"
	"os"
	"path"
)

func loadCertificate(pth string, pool *x509.CertPool) {
	bts, err := os.ReadFile(pth)
	if err == nil {
		var ppb *pem.Block
		ppb, bts = pem.Decode(bts)
		for ppb != nil {
			crt, err := x509.ParseCertificate(ppb.Bytes)
			if err == nil {
				pool.AddCert(crt)
			}
			ppb, bts = pem.Decode(bts)
		}
	}
}

func getAbsPath(root string, pth string) string {
	if path.IsAbs(pth) {
		return pth
	}
	return path.Join(root, pth)
}

func (c SecurityYAML) GetCert() *tls.Certificate {
	cwd := utils.GetCWD()
	cert, err := tls.LoadX509KeyPair(getAbsPath(cwd, c.PublicCert), getAbsPath(cwd, c.PrivateKey))
	if err != nil {
		return nil
	}
	return &cert
}

func (c SecurityYAML) GetCertPool() *x509.CertPool {
	var pool *x509.CertPool
	var err error
	if c.NoSystemCerts {
		pool = x509.NewCertPool()
	} else {
		pool, err = x509.SystemCertPool()
		if err != nil {
			pool = x509.NewCertPool()
		}
	}
	cwd := utils.GetCWD()
	fCAPath := getAbsPath(cwd, c.CACert)
	if st, err := os.Stat(fCAPath); err == nil && !st.IsDir() {
		loadCertificate(fCAPath, pool)
	}
	dCAPath := getAbsPath(cwd, c.CACertDir)
	if st, err := os.Stat(dCAPath); err == nil && st.IsDir() {
		dfs, err := os.ReadDir(dCAPath)
		if err == nil {
			for _, df := range dfs {
				if st, err := os.Stat(path.Join(dCAPath, df.Name())); err == nil && !st.IsDir() {
					loadCertificate(path.Join(dCAPath, df.Name()), pool)
				}
			}
		}
	}
	return pool
}
