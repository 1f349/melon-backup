package conf

import (
	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
	"io"
	"time"
)

func Generate(target io.Writer) {
	cEnc := yaml.NewEncoder(target)
	//Conf Defaults
	cnf := &ConfigYAML{
		Mode:      string(Unknown),
		StoreFile: "store.tar.gz",
		Services: ServiceYAML{
			List:          []string{""},
			Stop:          true,
			Restore:       true,
			StartNew:      true,
			ReloadCommand: []string{"systemctl", "daemon-reload"},
			StopCommand:   []string{"systemctl", "stop"},
			StartCommand:  []string{"systemctl", "start"},
			ManageRSync:   true,
		},
		Net: NetYAML{
			TargetAddr:         "127.0.0.1",
			TargetPort:         872,
			TargetExpectedName: "localhost",
			ListeningAddr:      "127.0.0.1",
			ListeningPort:      872,
			RemoteAllowedNames: []string{"127.0.0.1", "localhost"},
			ProxyLocalAddr:     "127.0.0.1",
			ProxyLocalPort:     873,
			ProxyBufferSize:    8192,
			KeepAliveTime:      time.Second * 5,
		},
		Security: SecurityYAML{
			PublicCert:    "me.pem",
			PrivateKey:    "me.priv",
			CACert:        "ca.pem",
			CACertDir:     "/certs",
			RSyncPassword: "RsYnC8--",
			NoSystemCerts: true,
		},
		ExcludeProtection: ExcludeProtectYAML{
			ProtectCommand:    []string{"tar", "-zcvpSPf", "-", "--numeric-owner", "--acls", "--selinux", "--xattrs", "--atime-preserve", "/var/log/rsync.log", "/var/run/rsyncd.pid", "/var/run/rsync.lock", "/etc/rsyncd.conf", "/etc/rsyncd.secrets", "/etc/melon-backup"},
			UnProtectCommand:  []string{"tar", "-zxvpSPf", "-", "--numeric-owner", "--same-owner", "--acls", "--selinux", "--xattrs", "--atime-preserve", "-C", "/"},
			StdOutBuffStdInOn: true,
		},
		TriggerReboot: true,
		RebootCommand: []string{"systemctl", "reboot"},
		RSyncCommand:  []string{"rsync", "-vcrlHAXogtUSxz", "--mkpath", "--open-noatime", "--super", "--delete-during", "--force", "--numeric-ids", "--timeout=300", "--port=873", "--inplace", "--exclude", "/var/log/rsync.log", "--exclude", "/var/run/rsyncd.pid", "--exclude", "/var/run/rsync.lock", "--exclude", "/dev", "--exclude", "/sys", "--exclude", "/proc", "--exclude", "/etc/rsyncd.conf", "--exclude", "/etc/rsyncd.secrets", "--exclude", "/etc/melon-backup", "--stats", "/", "rbackupuser@127.0.0.1::files/"},
		TarCommand:    []string{"tar", "-zcvpSPf", "-", "--numeric-owner", "--acls", "--selinux", "--xattrs", "--xattrs-include=*", "--atime-preserve", "--one-file-system", "--exclude=/var/log/rsync.log", "--exclude=/var/run/rsyncd.pid", "--exclude=/var/run/rsync.lock", "--exclude=/dev", "--exclude=/sys", "--exclude=/proc", "--exclude=/etc/rsyncd.conf", "--exclude=/etc/rsyncd.secrets", "--exclude=/etc/melon-backup", "/"},
		UnTarCommand:  []string{"tar", "-zxvpSUPf", "-", "--recursive-unlink", "--numeric-owner", "--same-owner", "--acls", "--selinux", "--xattrs", "--xattrs-include=*", "--atime-preserve", "--exclude=/var/log/rsync.log", "--exclude=/var/run/rsyncd.pid", "--exclude=/var/run/rsync.lock", "--exclude=/dev", "--exclude=/sys", "--exclude=/proc", "--exclude=/etc/rsyncd.conf", "--exclude=/etc/rsyncd.secrets", "--exclude=/etc/melon-backup", "-C", "/"},
		TarBufferSize: 8192,
	}
	err := cEnc.Encode(&cnf)
	if err != nil {
		log.Error(err)
	}
}
