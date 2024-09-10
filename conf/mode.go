package conf

type Mode string

const Backup = Mode("backup")
const Restore = Mode("restore")
const Store = Mode("store")
const UnStore = Mode("unstore")
const Unknown = Mode("unknown")

func (m Mode) ToInt() int {
	switch m {
	case Backup:
		return 1
	case Restore:
		return 2
	case Store:
		return 3
	case UnStore:
		return 4
	default:
		return 0
	}
}

func ModeFromInt(i int) Mode {
	switch i {
	case 1:
		return Backup
	case 2:
		return Restore
	case 3:
		return Store
	case 4:
		return UnStore
	default:
		return Unknown
	}
}
