package conf

import (
	"bufio"
	"fmt"
	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v3"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type countable interface {
	Count() int64
}

type cDummy struct{}

func (c *cDummy) Count() int64 {
	return 0
}

type readWrapper struct {
	io.Reader
	readBytes int64
}

func (w *readWrapper) Count() int64 {
	return w.readBytes
}

func (w *readWrapper) Read(p []byte) (n int, err error) {
	n, err = w.Reader.Read(p)
	w.readBytes += int64(n)
	return
}

type writeWrapper struct {
	io.WriteSeeker
	writtenBytes int64
}

func (w *writeWrapper) Seek(offset int64, whence int) (int64, error) {
	return w.Seek(offset, whence)
}

func (w *writeWrapper) Count() int64 {
	return w.writtenBytes
}

func (w *writeWrapper) Write(p []byte) (n int, err error) {
	n, err = w.WriteSeeker.Write(p)
	w.writtenBytes += int64(n)
	return
}

func Generate(target io.WriteSeeker, existing io.Reader, noEdit bool) int64 {
	var counter countable = &cDummy{}
	if existing != nil {
		rw := &readWrapper{existing, 0}
		counter = rw
		existing = rw
	}
	var cnf *ConfigYAML
	if existing == nil {
		//Conf Defaults
		cnf = &ConfigYAML{
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
				StatusCommand: []string{"systemctl", "status"},
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
				CACertDir:     "ca",
				RSyncPassword: "RsYnC8--",
				NoSystemCerts: true,
			},
			ExcludeProtection: ExcludeProtectYAML{
				ProtectCommand:    []string{"tar", "-zcvpSPf", "-", "--numeric-owner", "--acls", "--selinux", "--xattrs", "--xattrs-include=*", "--atime-preserve", "/var/log/rsync.log", "/var/run/rsyncd.pid", "/var/run/rsync.lock", "/etc/rsyncd.conf", "/etc/rsyncd.secrets", "/etc/melon-backup"},
				UnProtectCommand:  []string{"tar", "-zxvpSPf", "-", "--numeric-owner", "--same-owner", "--acls", "--selinux", "--xattrs", "--xattrs-include=*", "--atime-preserve", "-C", "/"},
				StdOutBuffStdInOn: true,
			},
			TriggerReboot: true,
			RebootCommand: []string{"systemctl", "reboot"},
			RSyncCommand:  []string{"rsync", "-vcrlHAXogtUSxz", "--mkpath", "--open-noatime", "--super", "--delete-during", "--force", "--numeric-ids", "--timeout=300", "--port=873", "--inplace", "--exclude", "/var/log/rsync.log", "--exclude", "/var/run/rsyncd.pid", "--exclude", "/var/run/rsync.lock", "--exclude", "/dev", "--exclude", "/sys", "--exclude", "/proc", "--exclude", "/etc/rsyncd.conf", "--exclude", "/etc/rsyncd.secrets", "--exclude", "/etc/melon-backup", "--stats", "/", "rbackupuser@127.0.0.1::files/"},
			TarCommand:    []string{"tar", "-zcvpSPf", "-", "--numeric-owner", "--acls", "--selinux", "--xattrs", "--xattrs-include=*", "--atime-preserve", "--one-file-system", "--exclude=/var/log/rsync.log", "--exclude=/var/run/rsyncd.pid", "--exclude=/var/run/rsync.lock", "--exclude=/dev", "--exclude=/sys", "--exclude=/proc", "--exclude=/etc/rsyncd.conf", "--exclude=/etc/rsyncd.secrets", "--exclude=/etc/melon-backup", "/"},
			UnTarCommand:  []string{"tar", "-zxvpSUPf", "-", "--recursive-unlink", "--numeric-owner", "--same-owner", "--acls", "--selinux", "--xattrs", "--xattrs-include=*", "--atime-preserve", "--exclude=/var/log/rsync.log", "--exclude=/var/run/rsyncd.pid", "--exclude=/var/run/rsync.lock", "--exclude=/dev", "--exclude=/sys", "--exclude=/proc", "--exclude=/etc/rsyncd.conf", "--exclude=/etc/rsyncd.secrets", "--exclude=/etc/melon-backup", "-C", "/"},
			TarBufferSize: 8192,
			RSyncService:  "rsync.service",
		}
	} else {
		cDec := yaml.NewDecoder(existing)
		err := cDec.Decode(&cnf)
		if err != nil {
			log.Error(err)
			return counter.Count()
		}
		_, err = target.Seek(0, io.SeekStart)
		if err != nil {
			log.Warn(err)
		}
	}
	if !noEdit {
		stdin := bufio.NewScanner(os.Stdin)
		fmt.Println("Editor:")

		cnf.Mode = readChoice(stdin, "mode:", []string{"unknown", "backup", "restore", "store", "unstore"}, cnf.Mode)
		mode := Mode(cnf.Mode)
		if mode != Backup && mode != Restore {
			cnf.StoreFile = readString(stdin, "storeFile:", cnf.StoreFile, true)
		} else {
			cnf.StoreFile = ""
		}

		fmt.Println("services:")
		cnf.Services.List = readStringSlice(stdin, "service list:", cnf.Services.List)
		cnf.Services.Stop = readBool(stdin, "stop services:", cnf.Services.Stop)
		cnf.Services.Restore = readBool(stdin, "restore services:", cnf.Services.Restore)
		cnf.Services.StartNew = readBool(stdin, "start new services:", cnf.Services.StartNew)
		cnf.Services.ReloadCommand = readStringSlice(stdin, "reload services information command:", cnf.Services.ReloadCommand)
		cnf.Services.StopCommand = readStringSlice(stdin, "stop service base command:", cnf.Services.StopCommand)
		cnf.Services.StartCommand = readStringSlice(stdin, "start service base command:", cnf.Services.StartCommand)
		cnf.Services.StatusCommand = readStringSlice(stdin, "service status base command:", cnf.Services.StatusCommand)
		if mode != Store && mode != UnStore {
			cnf.Services.ManageRSync = readBool(stdin, "manage the rsync service:", cnf.Services.ManageRSync)
		}

		fmt.Println("net:")
		if mode == Unknown {
			cnf.Net.TargetAddr = readString(stdin, "target address:", cnf.Net.TargetAddr, true)
			cnf.Net.TargetPort = readUInt16(stdin, "target port:", cnf.Net.TargetPort)
			cnf.Net.TargetExpectedName = readString(stdin, "target expected name:", cnf.Net.TargetExpectedName, true)
			cnf.Net.ListeningAddr = readString(stdin, "listening address:", cnf.Net.ListeningAddr, true)
			cnf.Net.ListeningPort = readUInt16(stdin, "listening port:", cnf.Net.ListeningPort)
			cnf.Net.RemoteAllowedNames = readStringSlice(stdin, "allowed remote names:", cnf.Net.RemoteAllowedNames)
		} else if cnf.Net.ListeningAddr == "" {
			cnf.Net.TargetAddr = readString(stdin, "target address:", cnf.Net.TargetAddr, true)
			if cnf.Net.TargetAddr == "" {
				cnf.Net.ListeningAddr = readString(stdin, "listening address:", cnf.Net.ListeningAddr, false)
				cnf.Net.ListeningPort = readUInt16(stdin, "listening port:", cnf.Net.ListeningPort)
				cnf.Net.RemoteAllowedNames = readStringSlice(stdin, "allowed remote names:", cnf.Net.RemoteAllowedNames)
			} else {
				cnf.Net.TargetPort = readUInt16(stdin, "target port:", cnf.Net.TargetPort)
				cnf.Net.TargetExpectedName = readString(stdin, "target expected name:", cnf.Net.TargetExpectedName, false)
			}
		} else {
			cnf.Net.ListeningAddr = readString(stdin, "listening address:", cnf.Net.ListeningAddr, true)
			if cnf.Net.ListeningAddr == "" {
				cnf.Net.TargetAddr = readString(stdin, "target address:", cnf.Net.TargetAddr, false)
				cnf.Net.TargetPort = readUInt16(stdin, "target port:", cnf.Net.TargetPort)
				cnf.Net.TargetExpectedName = readString(stdin, "target expected name:", cnf.Net.TargetExpectedName, false)
			} else {
				cnf.Net.ListeningPort = readUInt16(stdin, "listening port:", cnf.Net.ListeningPort)
				cnf.Net.RemoteAllowedNames = readStringSlice(stdin, "allowed remote names:", cnf.Net.RemoteAllowedNames)
			}
		}
		if mode != Store && mode != UnStore {
			cnf.Net.ProxyLocalAddr = readString(stdin, "rsync proxy local address:", cnf.Net.ProxyLocalAddr, true)
			if cnf.Net.ProxyLocalAddr != "" {
				cnf.Net.ProxyLocalPort = readUInt16(stdin, "rsync proxy local port:", cnf.Net.ProxyLocalPort)
				cnf.Net.ProxyBufferSize = readUInt32(stdin, "rsync proxy data transfer buffer size:", cnf.Net.ProxyBufferSize)
			}
		}
		cnf.Net.KeepAliveTime = readDuration(stdin, "connection keep alive time:", cnf.Net.KeepAliveTime)

		fmt.Println("security:")
		cnf.Security.PublicCert = readString(stdin, "public certificate path:", cnf.Security.PublicCert, false)
		cnf.Security.PrivateKey = readString(stdin, "private key path:", cnf.Security.PrivateKey, false)
		cnf.Security.CACert = readString(stdin, "accepted CA certificate path:", cnf.Security.CACert, true)
		cnf.Security.CACertDir = readString(stdin, "accepted CA certificate directory path:", cnf.Security.CACertDir, true)
		if mode != Store && mode != UnStore {
			cnf.Security.RSyncPassword = readString(stdin, "rsync password path:", cnf.Security.RSyncPassword, true)
		}
		cnf.Security.NoSystemCerts = readBool(stdin, "do not use system CA certificates:", cnf.Security.NoSystemCerts)

		fmt.Println("exclude protection:")
		cnf.ExcludeProtection.ProtectCommand = readStringSlice(stdin, "protect command:", cnf.ExcludeProtection.ProtectCommand)
		cnf.ExcludeProtection.UnProtectCommand = readStringSlice(stdin, "un-protect command:", cnf.ExcludeProtection.UnProtectCommand)
		cnf.ExcludeProtection.StdOutBuffStdInOn = readBool(stdin, "buffer protected data from stdout of protect command into memory and send to un-protect command stdin:", cnf.ExcludeProtection.StdOutBuffStdInOn)

		cnf.TriggerReboot = readBool(stdin, "trigger reboot on completion:", cnf.TriggerReboot)
		cnf.RebootCommand = readStringSlice(stdin, "reboot command:", cnf.RebootCommand)
		if mode != Store && mode != UnStore {
			cnf.RSyncCommand = readStringSlice(stdin, "rsync command:", cnf.RSyncCommand)
		}
		cnf.TarCommand = readStringSlice(stdin, "tar command:", cnf.TarCommand)
		cnf.UnTarCommand = readStringSlice(stdin, "untar command:", cnf.UnTarCommand)
		cnf.TarBufferSize = readUInt32(stdin, "tar data transfer buffer size:", cnf.TarBufferSize)
		if mode != Store && mode != UnStore {
			cnf.RSyncService = readString(stdin, "rsync service name:", cnf.RSyncService, true)
		}
	}
	ww := &writeWrapper{target, 0}
	counter = ww
	target = ww
	cEnc := yaml.NewEncoder(target)
	err := cEnc.Encode(&cnf)
	if err != nil {
		log.Error(err)
	}
	return counter.Count()
}

func readBool(scanner *bufio.Scanner, prompt string, original bool) bool {
	var or string
	if original {
		or = "y"
	} else {
		or = "n"
	}
	choice := readChoice(scanner, prompt, []string{"y", "n"}, or)
	if choice == "y" {
		return true
	}
	return false
}

func readString(scanner *bufio.Scanner, prompt string, original string, allowEmpty bool) string {
	fmt.Println(prompt)
	fmt.Println("Current:", original)
	if allowEmpty {
		fmt.Println("Empty entry DOES NOT keep current")
	}
	if !scanner.Scan() || (!allowEmpty && scanner.Text() == "") {
		return original
	}
	return scanner.Text()
}

func readUInt16(scanner *bufio.Scanner, prompt string, original uint16) uint16 {
	newVal := readString(scanner, prompt, fmt.Sprintf("%d", original), true)
	if newVal == "" {
		return 0
	}
	toRet, err := strconv.ParseUint(newVal, 10, 16)
	if err != nil || toRet > math.MaxUint16 {
		fmt.Println("Invalid, keeping original")
		return original
	}
	return uint16(toRet)
}

func readUInt32(scanner *bufio.Scanner, prompt string, original uint32) uint32 {
	newVal := readString(scanner, prompt, fmt.Sprintf("%d", original), true)
	if newVal == "" {
		return 0
	}
	toRet, err := strconv.ParseUint(newVal, 10, 32)
	if err != nil || toRet > math.MaxUint32 {
		fmt.Println("Invalid, keeping original")
		return original
	}
	return uint32(toRet)
}

func readDuration(scanner *bufio.Scanner, prompt string, original time.Duration) time.Duration {
	newVal := readString(scanner, prompt, original.String(), false)
	toRet, err := time.ParseDuration(newVal)
	if err != nil {
		fmt.Println("Invalid, keeping original")
		return original
	}
	return toRet
}

func readChoice(scanner *bufio.Scanner, prompt string, choices []string, original string) string {
	fmt.Println(prompt)
	fmt.Println("Choices: ", strings.Join(choices, ", "))
	fmt.Println("Current:", original)
	if !scanner.Scan() || scanner.Text() == "" {
		return original
	}
	selected := strings.ToLower(scanner.Text())
	if slices.Contains(choices, selected) {
		return selected
	} else {
		idx, err := strconv.ParseUint(selected, 10, 64)
		if err != nil || idx >= uint64(len(choices)) {
			fmt.Println("Invalid, keeping original")
		} else {
			return choices[idx]
		}
	}
	return original
}

func readStringSlice(scanner *bufio.Scanner, prompt string, original []string) []string {
	fmt.Println(prompt)
	fmt.Println("Current Values:")
	choices := []string{"+", "-"}
	toRet := make([]string, 0, len(original))
	for _, val := range original {
		c := readChoice(scanner, val+" : keep, remove", choices, "+")
		if c == "+" {
			toRet = append(toRet, val)
		}
	}
	fmt.Println("Appended Values, end by entering an empty entry:")
	nVal := " "
	for nVal != "" {
		if scanner.Scan() && scanner.Text() != "" {
			toRet = append(toRet, scanner.Text())
		} else {
			nVal = ""
		}
	}
	return toRet
}
