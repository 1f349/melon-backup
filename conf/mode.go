package conf

type Mode string

const Backup = Mode("backup")
const Restore = Mode("restore")
const Store = Mode("store")
const UnStore = Mode("unstore")
const Unknown = Mode("unknown")
