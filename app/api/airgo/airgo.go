package airgo

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/ppoonk/shy/app/api"
	"github.com/ppoonk/shy/extras/trafficlogger"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	getNodeInfo         = "/api/airgo/node/getNodeInfo"
	getUserlist         = "/api/airgo/user/getUserlist"
	reportUserTraffic   = "/api/airgo/user/reportUserTraffic"
	reportNodeStatus    = "/api/airgo/node/reportNodeStatus"
	HyTrafficStats      = "http://127.0.0.1:7654/traffic"         //默认流量统计 API
	HyTrafficStatsClear = "http://127.0.0.1:7654/traffic?clear=1" //默认流量统计 API
)

type ApiClient struct {
	Client   *resty.Client
	ETags    map[string]string
	Panel    api.Panel
	UserList map[string]string
	lock     sync.Mutex
}

func New(p api.Panel) *ApiClient {
	client := resty.New()
	client.SetRetryCount(3)
	client.SetTimeout(5 * time.Second)
	client.OnError(func(req *resty.Request, err error) {
		var v *resty.ResponseError
		if errors.As(err, &v) {
			log.Print(v.Err)
		}
	})
	client.SetBaseURL(p.APIHost)
	// Create Key for each requests
	client.SetQueryParam("key", p.Key)
	return &ApiClient{
		Client:   client,
		Panel:    p,
		UserList: make(map[string]string, 0),
	}

}

func (a *ApiClient) GetNodeInfo() (*api.NodeInfo, error) {
	res, err := a.Client.R().SetQueryParams(map[string]string{
		"id":  fmt.Sprintf("%d", a.Panel.NodeID),
		"key": a.Panel.Key,
	}).SetHeader("If-None-Match", a.ETags["userlist"]).ForceContentType("application/json").Get(getNodeInfo)
	if err != nil {
		return nil, err
	}
	if res.StatusCode() == 304 {
		return nil, errors.New("The node infomation has not changed")
	}
	// update etag
	if res.Header().Get("Etag") != "" && res.Header().Get("Etag") != a.ETags["nodeInfo"] {
		a.ETags["nodeInfo"] = res.Header().Get("Etag")
	}

	var nodeInfo NodeInfoResponse
	err = json.Unmarshal(res.Body(), &nodeInfo)
	if err != nil {
		return nil, err
	}
	return &api.NodeInfo{
		NodeID:     a.Panel.NodeID,
		Port:       nodeInfo.Port,
		SpeedLimit: nodeInfo.NodeSpeedlimit,
		Host:       nodeInfo.Host,
	}, nil

}

func (a *ApiClient) GetUserList() error {
	res, err := a.Client.R().SetQueryParams(map[string]string{
		"id":  fmt.Sprintf("%d", a.Panel.NodeID),
		"key": a.Panel.Key,
	}).SetHeader("If-None-Match", a.ETags["userlist"]).ForceContentType("application/json").Get(getUserlist)
	if err != nil {
		return err
	}
	if res.StatusCode() == 304 {
		return errors.New("The user list has not changed")
	}
	// update etag
	if res.Header().Get("Etag") != "" && res.Header().Get("Etag") != a.ETags["userlist"] {
		a.ETags["userlist"] = res.Header().Get("Etag")
	}

	var userList []UserResponse
	err = json.Unmarshal(res.Body(), &userList)

	if err != nil {
		return err
	}
	//
	a.lock.Lock()
	a.UserList = make(map[string]string, 0)
	for _, v := range userList {
		a.UserList[v.UUID] = fmt.Sprintf("%d", v.ID)

	}
	a.lock.Unlock()
	return nil

}

func (a *ApiClient) ReportUserTraffic() error {
	client := resty.New()
	res, err := client.R().ForceContentType("application/json").Get(HyTrafficStats)
	if err != nil {
		return err
	}
	data := make(map[string]*trafficlogger.TrafficStatsEntry)
	err = json.Unmarshal(res.Body(), &data)
	if err != nil {
		return err
	}

	//
	var userTrafficRequest UserTrafficRequest
	userTrafficRequest.ID = a.Panel.NodeID
	for k, v := range data {
		uID, _ := strconv.ParseInt(k, 10, 32)
		userTrafficRequest.UserTraffic = append(userTrafficRequest.UserTraffic, api.UserTraffic{
			UID:      uint32(uID),
			Upload:   int64(v.Tx),
			Download: int64(v.Rx),
		})
	}
	res, err = a.Client.R().ForceContentType("application/json").SetBody(userTrafficRequest).Post(reportUserTraffic)
	if res.StatusCode() == 200 {
		_, err = client.R().ForceContentType("application/json").Get(HyTrafficStatsClear)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *ApiClient) ReportNodeStatus() error {
	var nodeStatus NodeStatus
	nodeStatus.ID = a.Panel.NodeID

	infocpu, _ := cpu.Percent(time.Duration(time.Second), false)
	infomem, _ := mem.VirtualMemory()
	infodisk, _ := disk.Usage(".")

	nodeStatus.CPU = infocpu[0]
	nodeStatus.Mem = infomem.UsedPercent
	nodeStatus.Disk = infodisk.UsedPercent

	_, err := a.Client.R().
		ForceContentType("application/json").
		SetBody(nodeStatus).
		Post(reportUserTraffic)

	return err
}

func (a *ApiClient) Authenticate(addr net.Addr, auth string, tx uint64) (ok bool, id string) {
	id, ok = a.UserList[auth]
	if !ok || id == "" {
		return false, ""
	}
	return true, id
}

func ReadDate[T any](resp *http.Response) (T, error) {
	gzipFlag := false
	for k, v := range resp.Header {
		if strings.ToLower(k) == "content-encoding" && strings.ToLower(v[0]) == "gzip" {
			gzipFlag = true
		}
	}
	var content []byte
	var err error
	var data T

	if gzipFlag {
		gr, err := gzip.NewReader(resp.Body)
		defer gr.Close()
		if err != nil {
			return data, err
		}
		content, err = io.ReadAll(gr)
	} else {
		content, err = io.ReadAll(resp.Body)
	}
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(content, &data)
	if err != nil {
		return data, err
	}
	return data, nil

}
