package main

import (
	"fmt"
	"mmo_game_server/apis"
	"mmo_game_server/core"
	"zinx/ziface"
	"zinx/znet"
)
//当前客户端建立链接之后触发Hook函数
func OneConnectionAdd(conn ziface.IConnection) {
	fmt.Println("conn add")
	//创建一个玩家 将链接和玩家模块绑定
	p := core.NewPlayer(conn)
	//给客户端发送一个msgID：1
	p.ReturnPid()
	//个客户端发送一个msgid：200
	p.ReturnPlayerPosition()

	//上线成功了
	//将玩家对象添加到世界管理器中
	core.WorldMgrObj.AddPlayer(p)
	//给conn添加一个属性 pid属性
	conn.SetProperty("pid",p.Pid)
	//同步周边玩家，告知他们当前玩家已经上线，广播当前的玩家的位置信息
	p.SyncSurrounding()
}
func OnConnectionLost(conn ziface.IConnection){
	//客户端已经关闭
	//得知下线的是哪位玩家
	pid,_:=conn.GetProperty("pid")
	player:=core.WorldMgrObj.GetPlayerByPid(pid.(int32))
	//玩家的下线业务(发送消息)
	player.OffLine()
}
func main(){
	s:=znet.NewServer("mmo game server")
	//注册一些 链接创建/销毁的 Hook钩子函数
     s.AddOnConnStart(OneConnectionAdd)
	s.AddOnConnStop(OnConnectionLost)
	//针对msgID建立路由业务
	s.AddRouter(2,&apis.WorldChat{})
	s.AddRouter(3,&apis.Move{})
	//注册一些路由业务
	s.Server()
}
