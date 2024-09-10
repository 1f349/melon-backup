package processing

import (
	"github.com/1f349/melon-backup/comm"
	"github.com/1f349/melon-backup/conf"
	"github.com/charmbracelet/log"
	"strconv"
)

func Start(cnf conf.ConfigYAML, debug bool) int {
	var err error
	var commLstn *comm.Listener = nil
	var commClient *comm.Client = nil
	if cnf.Net.ListeningAddr != "" {
		log.Info("Starting Listener on: " + cnf.Net.ListeningAddr + ":" + strconv.Itoa(int(cnf.Net.ListeningPort)))
		commLstn, err = comm.NewListener(cnf, debug)
		if err != nil {
			if debug {
				log.Error(err)
			}
		} else {
			defer commLstn.Close()
			log.Info("Listener started!")
		}
	}
	if cnf.Net.TargetAddr != "" {
		log.Info("Starting Connection to Target at: " + cnf.Net.TargetAddr + ":" + strconv.Itoa(int(cnf.Net.TargetPort)))
		commClient, err = comm.NewClient(cnf, debug)
		if err != nil {
			if debug {
				log.Error(err)
			}
			log.Error("Unable to connect to the target!")
			return 2
		} else {
			defer commClient.Close()
			log.Info("Target Connection started!")
		}
	} else if commLstn != nil {
		log.Info("Waiting for Target Connection...")
		commClient, err = commLstn.Accept()
		if err != nil {
			if debug {
				log.Error(err)
			}
			log.Error("Unable to connect to a target!")
			return 3
		} else {
			defer commClient.Close()
			log.Info("Target Connection started!")
		}
	} else {
		log.Error("Configuration for target address missing!")
		return 1
	}
	sL := StopServices(cnf, debug)
	defer StartServices(cnf, sL, getServiceSliceFromSenderData(commClient.SenderData), debug)
	remoteMode := conf.ModeFromInt(commClient.SenderData.Mode)
	switch cnf.GetMode() {
	case conf.Backup:
		if remoteMode == conf.Restore {
			commClient.ActivateWithPacketProcessing()
			tsk := NewRsyncSender(cnf, commClient, debug)
			tsk.StartAndWait(debug)
		} else if remoteMode == conf.Store {
			conn := commClient.ActivateForPureConnection()
			if conn == nil {
				log.Error("Pure Connection Error!")
				return 6
			}
			tsk := NewTarTask(conn, cnf, debug)
			tsk.WaitOnCompletion(debug)
		} else {
			log.Error("Remote Mode Incompatible!")
			return 5
		}
	case conf.UnStore:
		conn := commClient.ActivateForPureConnection()
		if conn == nil {
			log.Error("Pure Connection Error!")
			return 6
		}
		if remoteMode == conf.Store || remoteMode == conf.Restore {
			tsk := NewUnFileTask(conn, cnf, debug)
			tsk.WaitOnCompletion()
		} else {
			log.Error("Remote Mode Incompatible!")
			return 5
		}
	case conf.Restore:
		if remoteMode == conf.Backup {
			commClient.ActivateWithPacketProcessing()
			tsk := NewRsyncIngester(cnf, commClient, debug)
			tsk.Wait(debug)
		} else if remoteMode == conf.UnStore {
			conn := commClient.ActivateForPureConnection()
			if conn == nil {
				log.Error("Pure Connection Error!")
				return 6
			}
			tsk := NewUnTarTask(conn, cnf, debug)
			tsk.WaitOnCompletion(debug)
		} else {
			log.Error("Remote Mode Incompatible!")
			return 5
		}
	case conf.Store:
		conn := commClient.ActivateForPureConnection()
		if conn == nil {
			log.Error("Pure Connection Error!")
			return 6
		}
		if remoteMode == conf.UnStore || remoteMode == conf.Backup {
			tsk := NewFileTask(conn, cnf, debug)
			tsk.WaitOnCompletion()
		} else {
			log.Error("Remote Mode Incompatible!")
			return 5
		}
	default:
		log.Error("Unknown Mode!")
		return 4
	}
	return 0
}

func getServiceSliceFromSenderData(p *comm.SenderPacket) []string {
	if p == nil || p.Services == nil {
		return nil
	}
	return p.Services.List
}
