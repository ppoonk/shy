package api

type Panel struct {
	PanelType string `mapstructure:"panelType"`
	APIHost   string `mapstructure:"apiHost"`
	NodeID    uint32 `mapstructure:"nodeID"`
	Key       string `mapstructure:"key"`
	Duration  string `mapstructure:"duration"`
}

type NodeInfo struct {
	NodeID     uint32
	Port       uint32
	SpeedLimit uint64 // Bps
	Host       string
}

type UserInfo struct {
	UID         uint32
	UUID        string
	SpeedLimit  uint64 // Bps
	DeviceLimit int
}

type UserTraffic struct {
	UID      uint32
	Upload   int64
	Download int64
}
