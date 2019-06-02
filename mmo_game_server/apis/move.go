package apis

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"mmo_game_server/core"
	"mmo_game_server/pb"
	"zinx/ziface"
	"zinx/znet"
)

//业务更新坐标 路由业务
type Move struct {
	znet.BaseRouter
}

func (m *Move) Handle(request ziface.IRequest) {
	//解析客户端发送过来的proto协议
	proto_msg := &pb.Position{}
	proto.Unmarshal(request.GetMsg().GetMsgData(), proto_msg)

	//通过链接属性得到当前玩家ID
	pid, _ := request.GetConnection().GetProperty("pid")
	fmt.Println("player id = ", pid.(int32), " move --> ", proto_msg.X, ", ", proto_msg.Z, ", ", proto_msg.V)
	//通过pid得到当前的玩家对象
	player := core.WorldMgrObj.GetPlayerByPid(pid.(int32))
	//玩家对象方法 将当前的新坐标位置发送给周边全部玩家
	player.UpdatePosition(proto_msg.X, proto_msg.Y, proto_msg.Z, proto_msg.V)
}
