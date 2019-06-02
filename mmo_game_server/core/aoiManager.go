package core

import "fmt"

//AOI格子管理模块
type AOIManager struct {
	//区域的左边边界
	MinX int
	//区域的右边边界
	MaxX int
	//x轴方向格子的数量
	CntsX int
	//区域的上边边界
	MinY int
	//区域的下边边界
	MaxY int
	//y轴方向格子的数量
	CntsY int
	//整体地图中所有的格子对象key：value
	grids map[int]*Grid
}

//得到每个格子在X轴方向的宽度
func (m *AOIManager) GridWidth() int {
	return (m.MaxX - m.MinX) / m.CntsX
}

//得到每个格子在Y轴方向的高度
func (m *AOIManager) GridHeight() int {
	return (m.MaxY - m.MinY) / m.CntsY
}

//初始化地图
func NewAOIManager(minX, maxX, cntsX, minY, maxY, cntsY int) *AOIManager {
	aoiMgr := &AOIManager{
		MinX:  minX,
		MaxX:  maxX,
		MinY:  minY,
		MaxY:  maxY,
		CntsX: cntsX,
		CntsY: cntsY,
		grids: make(map[int]*Grid),
	}

	//初始化隶属于当前地图的全部格子
	for y := 0; y < cntsY; y++ {
		for x := 0; x < cntsY; x++ {
			//得到当前格子ID
			gid := cntsX*y + x
			//给AOIManger添加格子
			aoiMgr.grids[gid] = NewGrid(gid,
				aoiMgr.MinX+x*aoiMgr.GridWidth(),
				aoiMgr.MinX+(x+1)*aoiMgr.GridWidth(),
				aoiMgr.MinY+y*aoiMgr.GridHeight(),
				aoiMgr.MinY+(y+1)*aoiMgr.GridHeight() )
		}
	}
	return aoiMgr
}

//打印当前的地图信息
func (m *AOIManager) String() string {
	s := fmt.Sprintf("AOIManager : \n MinX:%d,MaxX:%d,cntsX:%d, minY:%d, maxY:%d,cntsY:%d, Grids inManager:\n",
		m.MinX, m.MaxX, m.CntsX, m.MinY, m.MaxY, m.CntsY)

	//打印全部的格子
	for _, grid := range m.grids {
		s += fmt.Sprintln(grid)
	}

	return s
}

//添加一个PlayerID到一个AOI格子中
func (m *AOIManager) AddPidToGrid(pid, gid int) {
	m.grids[gid].Add(pid, nil)
}

//移除一个PlayerID 从一个AOI区域中
func (m *AOIManager) RemovePidFromGrid(pid, gid int) {
	m.grids[gid].Remove(pid)
}

//通过格子ID获取当前格子的全部PlayerID
func (m *AOIManager) GetPidsByGid(gid int) (playerIDs []int) {
	playerIDs = m.grids[gid].GetPlayerIDs()
	return
}

//通过一个格子ID得到当前格子的周边九宫格的格子ID集合
func (m *AOIManager) GetSurroundGridsByGid(gid int) (grids []*Grid) {
	//判断gid是否在AOI中
	if _, OK := m.grids[gid]; !OK {
		return
	}
	//将当前GID放入到九宫格切片中去
	grids = append(grids, m.grids[gid])
	//通过格子ID得到x轴编号
	idx := gid % m.CntsX
	//判断gid左边是否有格子 右边是否有格子
	if idx > 0 {
		grids = append(grids, m.grids[gid-1])
	}
	if idx < m.CntsX-1 {
		grids = append(grids, m.grids[gid+1])
	}
	//得到一个X轴的格子集合 遍历这个集合判断上面是否有ID 下面是否有ID
	gidsX := make([]int, 0, len(grids))
	for _, v := range grids {
		gidsX = append(gidsX, v.GID)
	}
	for _, gid := range gidsX {
		//通过GID得到y轴编号
		idy := gid / m.CntsX
		if idy > 0 {
			grids = append(grids, m.grids[gid-m.CntsX])
		}
		if idy < m.CntsY-1 {
			grids = append(grids, m.grids[gid+m.CntsX])
		}

	}
	return
}

//通过x，y坐标得到对应的格子ID
func (m *AOIManager) GetGidByPos(x, y float32) int {
	if x < 0 || int(x) >= m.MaxX {
		return -1
	}
	if y < 0 || int(y) >= m.MaxY {
		return -1
	}
	//根据坐标 得到当前玩家所在格子ID
	idx := (int(x) - m.MinX) / m.GridWidth()
	idy := (int(y) - m.MinY) / m.GridHeight()

	//gid  = idy*cntsX + idx
	gid := idy*m.CntsX + idx

	return gid
}

//根据一个坐标 得到 周边九宫格之内的全部的 玩家ID集合
func (m *AOIManager) GetSurroundPIDsByPos(x, y float32) (playerIDs []int) {

	//通过x，y得到一个格子对应的ID
	gid := m.GetGidByPos(x, y)

	//通过格子ID 得到周边九宫格 集合
	grids := m.GetSurroundGridsByGid(gid)

	fmt.Println("gid = ", gid)

	//将分别将九宫格内的全部的玩家 放在 playerIDs
	for _, grid := range grids {
		playerIDs = append(playerIDs, grid.GetPlayerIDs()...)
	}

	return
}

//通过坐标 将pid 加入到一个格子中
func (m *AOIManager) AddToGridByPos(pID int, x, y float32) {
	gID := m.GetGidByPos(x, y)
	//取出当前的格子
	grid := m.grids[gID]
	//给格子添加玩家
	grid.Add(pID, nil)
}

//通过坐标 把一个player从一个格子中删除
func (m *AOIManager) RemoteFromGridbyPos(pID int, x, y float32) {
	gID := m.GetGidByPos(x, y)

	grid := m.grids[gID]

	grid.Remove(pID)
}
