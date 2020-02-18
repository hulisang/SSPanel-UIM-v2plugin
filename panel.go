package v2ray_sspanel_v3_mod_Uim_plugin

import (
	"fmt"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/Manager"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/client"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/config"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/model"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/speedtest"
	"github.com/rico93/v2ray-sspanel-v3-mod_Uim-plugin/webapi"
	"github.com/robfig/cron"
	"google.golang.org/grpc"
	"reflect"
	"runtime"
)

type Panel struct {
	db              *webapi.Webapi
	manager         *Manager.Manager
	speedtestClient speedtest.Client
}

func NewPanel(gRPCConn *grpc.ClientConn, db *webapi.Webapi, cfg *config.Config) (*Panel, error) {
	opts := speedtest.NewOpts()
	speedtestClient := speedtest.NewClient(opts)
	var newpanel = Panel{
		speedtestClient: speedtestClient,
		db:              db,
		manager: &Manager.Manager{
			HandlerServiceClient:  client.NewHandlerServiceClient(gRPCConn, "MAIN_INBOUND"),
			StatsServiceClient:    client.NewStatsServiceClient(gRPCConn),
			UserRuleServiceClient: client.NewUserRuleServerClient(gRPCConn),
			NodeID:                cfg.NodeID,
			CheckRate:             cfg.CheckRate,
			SpeedTestCheckRate:    cfg.SpeedTestCheckRate,
			CurrentNodeInfo:       &model.NodeInfo{},
			NextNodeInfo:          &model.NodeInfo{},
			Users:                 map[string]model.UserModel{},
			UserToBeMoved:         map[string]model.UserModel{},
			UserToBeAdd:           map[string]model.UserModel{},
			Id2PrefixedIdmap:      map[uint]string{},
			Id2DisServer:          map[uint]string{},
		},
	}
	return &newpanel, nil
}

func (p *Panel) Start() {
	doFunc := func() {
		if err := p.do(); err != nil {
			newError("panel#do").Base(err).AtError().WriteToLog()
		}
		// Explicitly triggering GC to remove garbage
		runtime.GC()
	}
	doFunc()

	speedTestFunc := func() {
		result, err := speedtest.GetSpeedtest(p.speedtestClient)
		if err != nil {
			newError("panel#speedtest").Base(err).AtError().WriteToLog()
		}
		newError(result).AtInfo().WriteToLog()
		if p.db.UploadSpeedTest(p.manager.NodeID, result) {
			newError("succesfully upload speedtest result").AtInfo().WriteToLog()
		} else {
			newError("failed to upload speedtest result").AtInfo().WriteToLog()
		}
		// Explicitly triggering GC to remove garbage
		runtime.GC()
	}
	c := cron.New()
	err := c.AddFunc(fmt.Sprintf("@every %ds", p.manager.CheckRate), doFunc)
	if err != nil {
		fatal(err)
	}
	if p.manager.SpeedTestCheckRate > 0 {
		newErrorf("@every %dh", p.manager.SpeedTestCheckRate).AtInfo().WriteToLog()
		err = c.AddFunc(fmt.Sprintf("@every %dh", p.manager.SpeedTestCheckRate), speedTestFunc)
		if err != nil {
			newError("Can't add speed test into cron").AtWarning().WriteToLog()
		}
	}
	c.Start()
	c.Run()
}

func (p *Panel) do() error {
	p.updateManager()
	p.updateThroughout()
	return nil
}
func (p *Panel) initial() {
	newError("initial system").AtWarning().WriteToLog()
	p.manager.RemoveInbound()
	p.manager.CopyUsers()
	p.manager.UpdataUsers()
	p.manager.RemoveAllUserOutBound()
	p.manager.CurrentNodeInfo = &model.NodeInfo{}
	p.manager.NextNodeInfo = &model.NodeInfo{}
	p.manager.UserToBeAdd = map[string]model.UserModel{}
	p.manager.UserToBeMoved = map[string]model.UserModel{}
	p.manager.Users = map[string]model.UserModel{}
	p.manager.Id2PrefixedIdmap = map[uint]string{}
	p.manager.Id2DisServer = map[uint]string{}

}

func (p *Panel) updateManager() {
	newNodeinfo, err := p.db.GetNodeInfo(p.manager.NodeID)
	if err != nil {
		newError(err).AtWarning().WriteToLog()
		p.initial()
		return
	}
	if newNodeinfo.Ret != 1 {
		newError(newNodeinfo.Data).AtWarning().WriteToLog()
		p.initial()
		return
	}
	newErrorf("old node info %s ", p.manager.NextNodeInfo.Server_raw).AtInfo().WriteToLog()
	newErrorf("new node info %s", newNodeinfo.Data.Server_raw).AtInfo().WriteToLog()
	if p.manager.NextNodeInfo.Server_raw != newNodeinfo.Data.Server_raw {
		p.manager.NextNodeInfo = newNodeinfo.Data
		if err = p.manager.UpdateServer(); err != nil {
			newError(err).AtWarning().WriteToLog()
		}
		p.manager.UserChanged = true
	}
	users, err := p.db.GetALLUsers(p.manager.NextNodeInfo)
	if err != nil {
		newError(err).AtDebug().WriteToLog()
	}
	newError("now begin to check users").AtInfo().WriteToLog()
	current_user := p.manager.GetUsers()
	// remove user by prefixed_id
	for key, _ := range current_user {
		_, ok := users.Data[key]
		if !ok {
			p.manager.Remove(key)
			newErrorf("need to remove client: %s.", key).AtInfo().WriteToLog()
		}
	}
	// add users
	for key, value := range users.Data {
		current, ok := current_user[key]
		if !ok {
			p.manager.Add(value)
			newErrorf("need to add user email %s", key).AtInfo().WriteToLog()
		} else {
			if !reflect.DeepEqual(value, current) {
				p.manager.Remove(key)
				p.manager.Add(value)
				newErrorf("need to add user email %s due to method or password changed", key).AtInfo().WriteToLog()
			}
		}

	}

	if p.manager.UserChanged {
		p.manager.UserChanged = false
		newErrorf("Before Update, Current Users %d need to be add %d need to be romved %d", len(p.manager.Users),
			len(p.manager.UserToBeAdd), len(p.manager.UserToBeMoved)).AtWarning().WriteToLog()
		p.manager.UpdataUsers()
		newErrorf("After Update, Current Users %d need to be add %d need to be romved %d", len(p.manager.Users),
			len(p.manager.UserToBeAdd), len(p.manager.UserToBeMoved)).AtWarning().WriteToLog()
		p.manager.CurrentNodeInfo = p.manager.NextNodeInfo
	} else {
		newError("check ports finished. No need to update ").AtInfo().WriteToLog()
	}
	if newNodeinfo.Data.Sort == 12 {
		newError("Start to check relay rules ").AtInfo().WriteToLog()
		p.updateOutbounds()
	}
}
func (p *Panel) updateOutbounds() {
	data, err := p.db.GetDisNodeInfo(p.manager.NodeID)
	if err != nil {
		newError(err).AtWarning().WriteToLog()
		p.initial()
		return
	}
	if data.Ret != 1 {
		newError(data.Data).AtWarning().WriteToLog()
		p.initial()
		return
	}
	if len(data.Data) > 0 {
		newErrorf("Recieve %d User Rules", len(data.Data)).AtInfo().WriteToLog()
		globalSettingindex := -1
		for index, value := range data.Data {
			if value.UserId == 0 {
				globalSettingindex = index
				break
			}
		}
		if globalSettingindex != -1 {
			nextserver := data.Data[globalSettingindex]
			newErrorf("Got A Global Rule %s ", nextserver.Server_raw).AtInfo().WriteToLog()
			remove_count := 0
			add_count := 0
			for _, user := range p.manager.Users {
				currentserver, find := p.manager.Id2DisServer[user.UserID]
				nextserver.UserId = user.UserID
				if find {
					if currentserver != nextserver.Server_raw {
						p.manager.RemoveOutBound(currentserver+fmt.Sprintf("%d", user.UserID), user.UserID)
						err := p.manager.AddOuntBound(nextserver)
						if err != nil {
							newError("ADDOUTBOUND ").Base(err).AtInfo().WriteToLog()
						} else {

							remove_count += 1
							add_count += 1
						}
					}
				} else {
					err := p.manager.AddOuntBound(nextserver)
					if err != nil {
						newError("ADDOUTBOUND ").Base(err).AtInfo().WriteToLog()
					} else {
						add_count += 1
					}
				}
			}
			p.manager.Id2DisServer = map[uint]string{}
			for _, user := range p.manager.Users {
				p.manager.Id2DisServer[user.UserID] = nextserver.Server_raw
			}
			newErrorf("Add %d and REMOVE %d  Rules, Current Rules %d", add_count, remove_count, len(p.manager.Id2DisServer)).AtInfo().WriteToLog()

		} else {
			remove_count := 0
			add_count := 0
			for _, value := range data.Data {
				_, find := p.manager.Id2DisServer[value.UserId]
				if !find {
					err := p.manager.AddOuntBound(value)
					if err != nil {
						newError("ADDOUTBOUND ").Base(err).AtInfo().WriteToLog()
					} else {
						add_count += 1
					}
				}
			}
			for id, currentserver := range p.manager.Id2DisServer {
				flag := false
				currenttag := currentserver + fmt.Sprintf("%d", id)
				for _, nextserver := range data.Data {
					if id == nextserver.UserId && currenttag == nextserver.Server_raw+fmt.Sprintf("%d", nextserver.UserId) {
						flag = true
						break
					} else if id == nextserver.UserId && currenttag != nextserver.Server_raw+fmt.Sprintf("%d", nextserver.UserId) {
						p.manager.RemoveOutBound(currenttag, id)
						err := p.manager.AddOuntBound(nextserver)
						if err != nil {
							newError("ADDOUTBOUND ").Base(err).AtInfo().WriteToLog()
						} else {
							remove_count += 1
							add_count += 1
						}
						flag = true
						break
					}
					if !flag {
						p.manager.RemoveOutBound(currenttag, id)
						remove_count += 1
					}
				}

			}

			p.manager.Id2DisServer = map[uint]string{}
			for _, nextserver := range data.Data {
				p.manager.Id2DisServer[nextserver.UserId] = nextserver.Server_raw
			}
			newErrorf("Add %d and REMOVE %d  Rules, Current Rules %d", add_count, remove_count, len(p.manager.Id2DisServer)).AtInfo().WriteToLog()
		}
	} else {
		newErrorf("There is No User Rules, Need To Remove %d RULEs", len(p.manager.Id2DisServer)).AtInfo().WriteToLog()
		if len(p.manager.Id2DisServer) > 0 {
			remove_count := 0
			add_count := 0
			for id, currentserver := range p.manager.Id2DisServer {
				currenttag := currentserver + fmt.Sprintf("%d", id)
				p.manager.RemoveOutBound(currenttag, id)
				remove_count += 1
			}
			p.manager.Id2DisServer = map[uint]string{}
			newErrorf("Add %d and REMOVE %d  Rules, Current Rules %d", add_count, remove_count, len(p.manager.Id2DisServer)).AtInfo().WriteToLog()

		}
	}

}

func (p *Panel) updateThroughout() {
	current_user := p.manager.GetUsers()
	usertraffic := []model.UserTrafficLog{}
	userIPS := []model.UserOnLineIP{}
	for _, value := range current_user {
		current_upload, err := p.manager.StatsServiceClient.GetUserUplink(value.Email)
		current_download, err := p.manager.StatsServiceClient.GetUserDownlink(value.Email)
		if err != nil {
			newError(err).AtDebug().WriteToLog()
		}
		if current_upload+current_download > 0 {

			newErrorf("USER %s has use %d", value.Email, current_upload+current_download).AtDebug().WriteToLog()
			usertraffic = append(usertraffic, model.UserTrafficLog{
				UserID:   value.UserID,
				Downlink: current_download,
				Uplink:   current_upload,
			})
			current_user_ips, err := p.manager.StatsServiceClient.GetUserIPs(value.Email)
			if current_upload+current_download > 1024 {
				if err != nil {
					newError(err).AtDebug().WriteToLog()
				}
				for index := range current_user_ips {
					userIPS = append(userIPS, model.UserOnLineIP{
						UserId: value.UserID,
						Ip:     current_user_ips[index],
					})
				}

			}
		}
	}
	if p.db.UpLoadUserTraffics(p.manager.NodeID, usertraffic) {
		newErrorf("Successfully upload %d users traffics", len(usertraffic)).AtInfo().WriteToLog()
	} else {
		newError("update trafic failed").AtDebug().WriteToLog()
	}
	if p.db.UpLoadOnlineIps(p.manager.NodeID, userIPS) {
		newErrorf("Successfully upload %d ips", len(userIPS)).AtInfo().WriteToLog()
	} else {
		newError("update trafic failed").AtDebug().WriteToLog()
	}
	if p.db.UploadSystemLoad(p.manager.NodeID) {
		newError("Uploaded systemLoad successfully").AtInfo().WriteToLog()
	} else {
		newError("Failed to uploaded systemLoad ").AtDebug().WriteToLog()
	}
}
