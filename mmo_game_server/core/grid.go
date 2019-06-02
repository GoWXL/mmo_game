package core

import (
	"fmt"
	"sync"
)

//AOI兴趣点 格子模块
type Grid struct {
	//格子ID
	GID int
	//格子左边边界坐标
	MinX int
	//格子上有边边界坐标
	MaxX int
	//格子上边边界坐标
	MinY int
	//格子下边边界坐标
	MaxY int
	//当前格子内玩家 物品 成员ID集合
	playerIDs map[int]interface{}
	//保护当前内容的map锁
	pIDlock sync.RWMutex
}

//初始化NEW方法
func NewGrid(gID, minX, maxX, minY, maxY int) *Grid {
	return &Grid{
		GID:       gID,
		MinX:      minX,
		MaxX:      maxX,
		MinY:      minY,
		MaxY:      maxY,
		playerIDs: make(map[int]interface{}),
	}
}

//从格子中添加一个玩家
func (g *Grid) Add(playerID int, player interface{}) {
	g.pIDlock.Lock()
	defer g.pIDlock.Unlock()
	//添加一个玩家
	g.playerIDs[playerID] = player
}

//从格子中删除一个玩家
func (g *Grid) Remove(playerID int) {
	g.pIDlock.Lock()
	defer g.pIDlock.Unlock()
	//删除一个玩家
	delete(g.playerIDs, playerID)
}

//从格子中获取所有玩家ID
func (g *Grid) GetPlayerIDs() (playerIDs []int) {
	g.pIDlock.RLock()
	defer g.pIDlock.RUnlock()
	//获取所有玩家ID
	for playerID, _ := range g.playerIDs {
		playerIDs = append(playerIDs, playerID)
	}
	return
}

//调试打印格子信息方法
func (g *Grid) String() string {
	return fmt.Sprintf("Grid id : %d, minX:%d, maxX:%d , minY:%d, maxY:%d, playerIDs:%v\n",
		g.GID, g.MinX, g.MaxX, g.MinY, g.MaxY, g.playerIDs)
}
