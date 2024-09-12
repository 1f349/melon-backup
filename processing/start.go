package processing

import (
	"github.com/1f349/melon-backup/comm"
	"github.com/1f349/melon-backup/conf"
	"github.com/1f349/melon-backup/utils"
	"github.com/charmbracelet/log"
	"strconv"
)

func Start(cnf conf.ConfigYAML) int {
	var err error
	var commLstn *comm.Listener = nil
	var commClient *comm.Client = nil

	if cnf.Net.ListeningAddr != "" {
		log.Info("Starting Listener on: " + cnf.Net.ListeningAddr + ":" + strconv.Itoa(int(cnf.Net.ListeningPort)))
		commLstn, err = comm.NewListener(cnf)
		if err != nil {
			if conf.Debug {
				log.Error(err)
			}
		} else {
			defer commLstn.Close()
			log.Info("Listener started!")
		}
	}

	if cnf.Net.TargetAddr != "" {
		log.Info("Starting Connection to Target at: " + cnf.Net.TargetAddr + ":" + strconv.Itoa(int(cnf.Net.TargetPort)))
		commClient, err = comm.NewClient(cnf)
		if err != nil {
			if conf.Debug {
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
			if conf.Debug {
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

	remoteMode := conf.ModeFromInt(commClient.SenderData.Mode)
	log.Info("Local Mode: " + cnf.GetMode())
	log.Info("Remote Mode: " + remoteMode)

	sL := StopServices(cnf)
	if cnf.TriggerReboot && commClient.SenderData.RequestReboot {
		defer func() {
			ReloadServices(cnf)
			startReboot(cnf)
		}()
	} else {
		defer StartServices(cnf, sL, getServiceSliceFromSenderData(commClient.SenderData))
	}

	if cnf.GetMode() == conf.Restore {
		var protBuffer *utils.BufferDummyClose
		if cnf.ExcludeProtection.StdOutBuffStdInOn {
			protBuffer = &utils.BufferDummyClose{}
		}
		if len(cnf.ExcludeProtection.ProtectCommand) > 0 {
			if protBuffer == nil {
				tsk := NewCommandTask(cnf, utils.CreateCmd(cnf.ExcludeProtection.ProtectCommand), "Protect")
				if tsk != nil {
					tsk.StartAndWait()
				}
			} else {
				tsk := NewCommandToConnTask(protBuffer, false, "Protect", utils.CreateCmd(cnf.ExcludeProtection.ProtectCommand), cnf)
				if tsk != nil {
					tsk.WaitOnCompletion()
				}
			}
			if len(cnf.ExcludeProtection.UnProtectCommand) > 0 {
				if protBuffer == nil {
					defer func() {
						tsk := NewCommandTask(cnf, utils.CreateCmd(cnf.ExcludeProtection.UnProtectCommand), "UnProtect")
						if tsk != nil {
							tsk.StartAndWait()
						}
					}()
				} else {
					defer func() {
						tsk := NewConnToCommandTask(protBuffer, false, "UnProtect", utils.CreateCmd(cnf.ExcludeProtection.UnProtectCommand), cnf)
						if tsk != nil {
							tsk.WaitOnCompletion()
						}
					}()
				}
			}
		}
	}

	switch cnf.GetMode() {
	case conf.Backup:
		if remoteMode == conf.Restore && len(cnf.RSyncCommand) > 0 {
			commClient.ActivateWithPacketProcessing()
			tsk := NewRsyncSender(cnf, commClient)
			if tsk != nil {
				tsk.StartAndWait()
			} else {
				return 7
			}
		} else if remoteMode == conf.Store || (len(cnf.RSyncCommand) < 1 && remoteMode == conf.Restore) {
			conn := commClient.ActivateForPureConnection()
			if conn == nil {
				log.Error("Pure Connection Error!")
				return 6
			}
			tsk := NewCommandToConnTask(conn, true, "Tar", utils.CreateCmd(cnf.TarCommand), cnf)
			if tsk != nil {
				tsk.WaitOnCompletion()
			} else {
				return 7
			}
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
			tsk := NewUnFileTask(conn, cnf)
			if tsk != nil {
				tsk.WaitOnCompletion()
			} else {
				return 7
			}
		} else {
			log.Error("Remote Mode Incompatible!")
			return 5
		}
	case conf.Restore:
		if remoteMode == conf.Backup && len(cnf.RSyncCommand) > 0 {
			commClient.ActivateWithPacketProcessing()
			tsk := NewRsyncIngester(cnf, commClient)
			if tsk != nil {
				tsk.Wait()
			} else {
				return 7
			}
		} else if remoteMode == conf.UnStore || (len(cnf.RSyncCommand) < 1 && remoteMode == conf.Backup) {
			conn := commClient.ActivateForPureConnection()
			if conn == nil {
				log.Error("Pure Connection Error!")
				return 6
			}
			tsk := NewConnToCommandTask(conn, true, "UnTar", utils.CreateCmd(cnf.UnTarCommand), cnf)
			if tsk != nil {
				tsk.WaitOnCompletion()
			}
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
			tsk := NewFileTask(conn, cnf)
			if tsk != nil {
				tsk.WaitOnCompletion()
			} else {
				return 7
			}
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

func startReboot(cnf conf.ConfigYAML) {
	if cnf.GetMode() == conf.Store || cnf.GetMode() == conf.UnStore {
		return
	}
	cmd := utils.CreateCmd(cnf.RebootCommand)
	if cmd != nil {
		err := cmd.Run()
		if err != nil && conf.Debug {
			log.Error(err)
		}
	}
}
