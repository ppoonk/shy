package api

import "net"

type Api interface {
	GetNodeInfo() (*NodeInfo, error)
	GetUserList() error
	ReportUserTraffic() error
	ReportNodeStatus() error
	Authenticate(addr net.Addr, auth string, tx uint64) (ok bool, id string)
}
