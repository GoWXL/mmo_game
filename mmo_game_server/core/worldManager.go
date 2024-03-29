package core

import (
	"sync"
)

const (
	AOI_MIN_X  int = 85
	AOI_MAX_X  int = 410
	AOI_CNTS_X int = 10
	AOI_MIN_Y  int = 75
	AOI_MAX_Y  int = 400
	AOI_CNTS_Y int = 20
)

//当前的场景的世界管理模块
type WorldManager struct {
	//当前全部在线的player集合
	Players map[int32]*Player
	//保护Player集合的锁
	PLock sync.RWMutex
	//AOIManager当前地图的管理器
	AoiMgr *AOIManager
}

//对外观提供一个全局世界管理模块指针
var WorldMgrObj *WorldManager

func init() {
	//创建一个全局的世界管理对象
	WorldMgrObj = NewWorldManager()
}

//初始换方法
func NewWorldManager() *WorldManager {
	wm := &WorldManager{
		AoiMgr:  NewAOIManager(AOI_MIN_X, AOI_MAX_X, AOI_CNTS_X, AOI_MIN_Y, AOI_MAX_Y, AOI_CNTS_Y),
		Players: make(map[int32]*Player),
	}
	return wm
}

//添加一个玩家
func (wm *WorldManager) AddPlayer(player *Player) {
	//加入世界管理其中
	wm.PLock.Lock()
	wm.Players[player.Pid] = player
	wm.PLock.Unlock()
	//加入到世界地图中
	wm.AoiMgr.AddToGridByPos(int(player.Pid), player.X, player.Z)
}

//删除一个玩家
func (wm *WorldManager) RemovePlayerByPid(pid int32) {
	//从世界管理删除
	wm.PLock.Lock()
	//先通过pid 从世界管理器得到player对象
	player := wm.Players[pid]
	//从世界地图中删除
	wm.AoiMgr.RemoteFromGridbyPos(int(pid), player.X, player.Z)

	delete(wm.Players, pid)
	wm.PLock.Unlock()
}

//通过一个玩家ID得到一个player对象
func (wm *WorldManager) GetPlayerByPid(pid int32) *Player {
	wm.PLock.RLock()
	p := wm.Players[pid]
	wm.PLock.RUnlock()
	return p
}

//获取全部在线玩家的集合
func (wm *WorldManager) GetAllPlayers() []*Player {
	wm.PLock.RLock()
	defer wm.PLock.RUnlock()
	players := make([]*Player, 0)
	for _, player := range wm.Players {
		players = append(players, player)
	}
	return players
}
//获取一个格子中所有玩家
func (wm *WorldManager)GetPlayerByGid(gid int)[]*Player{
	pids:=wm.AoiMgr.GetPidsByGid(gid)
	Players:=make([]*Player,0,len(pids))
	for _,pid:=range pids{
		Players=append(Players,wm.Players[int32(pid)])
	}
	return  Players
}
