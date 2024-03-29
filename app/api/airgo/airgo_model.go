package airgo

import "github.com/ppoonk/shy/app/api"

type NodeInfoResponse struct {
	ID             uint32 `json:"id"`
	NodeSpeedLimit uint64 `json:"node_speed_limit"`
	TrafficRate    uint32 `json:"traffic_rate"`
	Protocol       string `json:"protocol"`
	Remarks        string `json:"remarks"`
	Address        string `json:"address"`
	Port           uint32 `json:"port"`
	Host           string `json:"host"`
}

type UserResponse struct {
	ID   uint32 `json:"id"`
	UUID string `json:"uuid"`
}

type UserTrafficRequest struct {
	ID          uint32            `json:"id"`
	UserTraffic []api.UserTraffic `json:"user_traffic"`
}
type NodeStatus struct {
	ID uint32 `json:"id"`
	NodeStatusItem
}
type NodeStatusItem struct {
	CPU    float64
	Mem    float64
	Disk   float64
	Uptime uint64
}
