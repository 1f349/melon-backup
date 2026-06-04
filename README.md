# Melon Backup

Allows both fully online and partially online\* backups (And restores) via the use of RSync / Tar secured over a TLS channel.

\* This does not support service by service backups as yet.

The 'daemon' is actually not a daemon in the traditional sense but either accepts or connects an action.
A compatible combination of modes is needed for both sides.

Combination pairs are reversible:
* backup - restore
* backup - store
* unstore - restore
* unstore - store

rsync mode is only available in the backup, restore pair mode.

## Usage

```
Usage: melon-backup <flags> <subcommand> <subcommand args>

Subcommands:
	commands         list all command names
	daemon           Run the daemon
	flags            describe all known top-level flags
	generate         Generate a config file
	edit             Edit a config file
	help             describe subcommands and their syntax
```

## Config Information

Command components are all string arrays of the command arguments.

* mode - backup, restore, store, unstore
* storeFile - file used for store modes
* services - services configuration
* net - net configuration
* security - security configuration
* excludeProtection - exclude protection config
* triggerReboot - execute reboot command on completion
* rebootCommand - reboot command
* rsyncCommand - rsync command
* tarCommand - tar command
* unTarCommand - un-tar command
* tarBufferSize - tar data transfer buffer size
* rsyncService - name of the rsync service

Services configuration:

* list - the services to manage;
order is important with service dependencies coming in first;
this is only executed on the backup/restore side;
starting services uses the list from the opposite side to work out which services still exist to restart and which are new
* stop - stop the services on this side once a connection starts
* restore - starts previously stopped services that still exist
* startNew - starts newly registered services
* reloadCommand - command to reload the service information
* stopCommand - command prefix to stop a service
* startCommand - command prefix to start a service
* statusCommand - command prefix to check a service state
* manageRSync - manage the rsync service

Net configuration:

* targetAddress - remote address to connect to
* targetPort - remote port to connect to
* targetExpectedName - expected certificate name of the target
* listeningAddress - local address to listen on
* listeningPort - local port to listen on
* remoteAllowedNames - a list of accepted presented certificate names
* rsyncLocalAddr - rsync proxy local listening address
* rsyncLocalPort - rsync proxy local listening port
* proxyBufferSize - rsync proxy data transfer buffer size
* keepAliveTime - timeout for connections

Security configuration:

* publicCert - path to the public certificate in PEM form
* privateKey - path to the private key in PEM form
* caCert - path to a recognised CA certificate in PEM form
* caCertDirectory - path to a directory containing recognised CA certificates in PEM form
* rsyncPassword - the password used by rsync
* noSystemCerts - whether system CA certificates are recognised

Exclude Protection configuration:

* protectCommand - command to protect data by storing it in a memory buffer or a file
* unProtectCommand - command to restore protected data by extracting it from a memory buffer or a file
* stdOutBuffStdInOn - use STD OUT and STD IN from the protection commands and an in memory buffer

(C) 1f349 2026 - GPL v3 License