package core

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"math/rand"
	"mmo_game_server/pb"
	"sync"
	"time"
	"zinx/ziface"
)

type Player struct {
	Pid  int32              //玩家ID
	Conn ziface.IConnection //当前玩家的链接(与对应客户端通信)
	X    float32            //平面的x轴坐标
	Y    float32            //高度
	Z    float32            //平面的y轴坐标
	V    float32            //玩家脸朝向的方向
}

// playerID 生成器
var PidGen int32 = 1  //用于生产玩家ID计数器
var IdLock sync.Mutex //保护PidGen生成器的互斥锁
//初始化玩家的方法
func NewPlayer(conn ziface.IConnection) *Player {
	//分配一个玩家ID
	IdLock.Lock()
	id := PidGen
	PidGen++
	IdLock.Unlock()

	//创建一个玩家对象
	p := &Player{
		Pid:  id,
		Conn: conn,
		X:    float32(160 + rand.Intn(10)), //随机生成玩家上线所在的x轴坐标
		Y:    0,
		Z:    float32(140 + rand.Intn(10)), //随机在140坐标点附近 y轴坐标上线
		V:    0,                            //角度为0
	}

	return p
}

//玩家可以和对端客户端发送消息的方法
func (p *Player) SendMsg(msgID uint32, proto_struct proto.Message) error {
	//要将proto结构体 转换成 二进制的数据
	binary_proto_data, err := proto.Marshal(proto_struct)
	if err != nil {
		fmt.Println("marshal proto struct error ", err)
		return err
	}

	//再调用zinx原生的connecton.Send（msgID, 二进制数据）
	if err := p.Conn.Send(msgID, binary_proto_data); err != nil {
		fmt.Println("Player send error!", err)
		return err
	}

	return nil
}

/*
 服务器给客户端发送玩家初始ID
 */
func (p *Player) ReturnPid() {
	//定义个msg:ID 1  proto数据结构
	proto_msg := &pb.SyncPid{
		Pid: p.Pid,
	}

	//将这个消息 发送给客户端
	p.SendMsg(1, proto_msg)
}

//服务器给客户端发送一个玩家的初始化位置信息
func (p *Player) ReturnPlayerPosition() {
	//组建MsgID:200消息
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2, //2 -坐标信息
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//将这个消息 发送给客户端
	p.SendMsg(200, proto_msg)
}

//将聊天数据广播给全部的在线玩家
func (p *Player) SendTalkMsgToAll(content string) {
	//定义一个广播的消息数据类型
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  1,
		Data: &pb.BroadCast_Content{
			Content: content,
		},
	}

	//获取全部的在线玩家
	players := WorldMgrObj.GetAllPlayers()
	//向全部玩家发送proto msg数据
	for _, player := range players {
		player.SendMsg(200, proto_msg)
	}
}

//得到当前玩家周边都有哪些玩家
func (p *Player) GetSurroundingPlayers() []*Player {
	pids := WorldMgrObj.AoiMgr.GetSurroundPIDsByPos(p.X, p.Z)
	fmt.Println("Surrounding players = ", pids)

	players := make([]*Player, 0, len(pids))
	for _, pid := range pids {
		players = append(players, WorldMgrObj.GetPlayerByPid(int32(pid)))
	}

	return players
}

//将自己的消息同步给周边的玩家
func (p *Player) SyncSurrounding() {
	//获取当前玩家的周边九宫格的玩家有哪些?
	players := p.GetSurroundingPlayers()

	//构建一个广播消息200， 循环全部players 分别给player对应的客户端发送200消息（让其他玩家看见当前玩家）
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  2,
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}

	//将当前玩家id和位置消息发送给周边玩家（发送多次）
	for _, player := range players {
		player.SendMsg(200, proto_msg)
	}

	//将其他玩家告诉当前玩家  （让当前玩家能够看见周边玩家的坐标）
	//构建一个202消息  players的信息 告知当前玩家 p.send(202, ... )
	//得到全部周边玩家的player集合message Player
	players_proto_msg := make([]*pb.Player, 0, len(players))
	for _, player := range players {
		//制作一个message Player 消息
		p := &pb.Player{
			Pid: player.Pid,
			P: &pb.Position{
				X: player.X,
				Y: player.Y,
				Z: player.Z,
				V: player.V,
			},
		}

		players_proto_msg = append(players_proto_msg, p)
	}
	//创建一个 Message SyncPlayers
	syncPlayers_proto_msg := &pb.SyncPlayers{
		Ps: players_proto_msg[:],
	}
	//将当前的周边的全部的玩家信息 发送给当前的客户端
	p.SendMsg(202, syncPlayers_proto_msg)
}

//格子切换视野问题
func (p *Player) OnExchangeAoiGrid(oldGrid, newGrid int) {
	//获取旧的九宫格成员
	oldGrids := WorldMgrObj.AoiMgr.GetSurroundGridsByGid(oldGrid)
	//将旧的九宫格成员添加到哈希表中
	oldGridMap := make(map[int]bool, len(oldGrids))
	for _, grid := range oldGrids {
		oldGridMap[grid.GID] = true
	}
	//获取新的九宫格成员
	newGrids := WorldMgrObj.AoiMgr.GetSurroundGridsByGid(newGrid)
	//将新的九宫格成员加入到哈希表中
	newGridMap := make(map[int]bool, len(newGrids))
	for _, grid := range newGrids {
		newGridMap[grid.GID] = true
	}
	//处理视野消失
	offline_msg := &pb.SyncPid{
		Pid: p.Pid,
	}
	//找到在旧的九宫格出现新的九宫格没有出现的玩家
	leavingGrids := make([]*Grid, 0)
	for _, grid := range oldGrids {
		if _, ok := newGridMap[grid.GID]; !ok {
			leavingGrids = append(leavingGrids, grid)
		}
	}
	//获取leavingGrids所有玩家
	for _, grid := range leavingGrids {
		players := WorldMgrObj.GetPlayerByGid(grid.GID)
		for _, player := range players {
			player.SendMsg(201, offline_msg)
			//将消失玩家在自己客户端消失
			another_offline_msg := &pb.SyncPid{
				Pid: player.Pid,
			}
			p.SendMsg(201, another_offline_msg)
		}
	}
	//处理玩家视野出现
  online_msg:=&pb.BroadCast{
  	Pid:p.Pid,
  	Tp:2,
  	Data:&pb.BroadCast_P{
  		P:&pb.Position{
  			X:p.X,
  			Y:p.Y,
  			Z:p.Z,
  			V:p.V,
		},
	},
  }
  enteringGrids:=make([]*Grid,0)
  for _,grid:=range newGrids{
  	if _,ok:=oldGridMap[grid.GID];!ok{
  		enteringGrids=append(enteringGrids,grid)
	}
  }
  //获取所有enteringGrids的玩家
  for _,grid:=range enteringGrids{
  	players:=WorldMgrObj.GetPlayerByGid(grid.GID)
  	for _,player:=range players{
  		player.SendMsg(200,online_msg)
  		//让新玩家出现在自己视野中
		another_online_msg := &pb.BroadCast{
			Pid:player.Pid,
			Tp:2,
			Data:&pb.BroadCast_P{
				&pb.Position{
					X:player.X,
					Y:player.Y,
					Z:player.Z,
					V:player.V,
				},
			},
		}
		time.Sleep(5*time.Second)
		p.SendMsg(200,another_online_msg)
	}
  }
}

//更新广播当前玩家的最新位置
func (p *Player) UpdatePosition(x, y, z, v float32) {
	//计算当前玩家是否已经跨越格子
	//旧的格子
	oldGrid:=WorldMgrObj.AoiMgr.GetGidByPos(p.X,p.Z)
	//新的格子
	newGrid:=WorldMgrObj.AoiMgr.GetGidByPos(x,z)
	if oldGrid!=newGrid{
		//触发格子切换
		//把pid从旧的aoi格子中删除
		WorldMgrObj.AoiMgr.RemovePidFromGrid(int(p.Pid),oldGrid)
		//将新的pid添加到新的aoi格子中
		WorldMgrObj.AoiMgr.AddPidToGrid(int(p.Pid),newGrid)
		//视野消失业务
		p.OnExchangeAoiGrid(oldGrid,newGrid)
	}
	//需要将最新的坐标 更新给当前玩家
	p.X = x
	p.Y = y
	p.Z = z
	p.V = v

	//组建广播proto协议 MSGID:200, Tp-4
	proto_msg := &pb.BroadCast{
		Pid: p.Pid,
		Tp:  4, //更新坐标
		Data: &pb.BroadCast_P{
			P: &pb.Position{
				X: p.X,
				Y: p.Y,
				Z: p.Z,
				V: p.V,
			},
		},
	}
	//获取当前玩家周边的AOI九宫格之内的玩家 player
	players := p.GetSurroundingPlayers()
	//依次调用Player对象 Send方法将200消息发过去
	for _, player := range players {
		player.SendMsg(200, proto_msg) //每个玩家都会给格子的 client客户端发送200消息
	}
}
func (p *Player) OffLine() {
	//	得到当前玩家的周边的玩家有哪些  players
	players := p.GetSurroundingPlayers()

	//制作一个消息MSgID:201
	proto_msg := &pb.SyncPid{
		Pid: p.Pid,
	}
	//	给周边的玩家广播一个消息 MsgID:201
	for _, player := range players {
		player.SendMsg(201, proto_msg) //客户端就会将当前的玩家从视野中删除
	}
	//	将该下线的玩家 从世界管理器移除
	WorldMgrObj.RemovePlayerByPid(p.Pid)
	//	将该下线玩家从 地图AOIManager中移除
	WorldMgrObj.AoiMgr.RemoteFromGridbyPos(int(p.Pid), p.X, p.Z)
}
