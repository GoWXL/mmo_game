package apis

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"mmo_game_server/core"
	"mmo_game_server/pb"
	"zinx/ziface"
	"zinx/znet"
)

//世界聊天 路由业务
type WorldChat struct {
	znet.BaseRouter
}

func (wc *WorldChat) Handle(request ziface.IRequest) {

	//解析数据端传过来的protobuf数据
	proto_msg := &pb.Talk{}
	err := proto.Unmarshal(request.GetMsg().GetMsgData(), proto_msg)
    if err!=nil{
    	fmt.Println("Talk Message unmarshal error",err)
		return
	}
	//通过获取链接属性 得到当前玩家ID
	pid,err:=request.GetConnection().GetProperty("pid")
	if err!=nil{
		fmt.Println("get pid error ",err)
		return
	}
	//通过pid得到对应的对象
	player:=core.WorldMgrObj.GetPlayerByPid(pid.(int32))
	//当前的聊天数据广播给全部的在线玩家

	//当前玩家的windows客户端发送过来的消息
	player.SendTalkMsgToAll(proto_msg.GetContent())
}
