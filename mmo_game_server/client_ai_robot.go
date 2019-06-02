package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"math/rand"
	"mmo_game_server/pb"
	"net"
	"os"
	"os/signal"
	"time"
)

type Message struct {
	Len   uint32
	MsgID uint32
	Data  []byte
}
type TcpClient struct {
	conn   net.Conn
	Pid    int32
	X      float32
	Y      float32
	Z      float32
	V      float32
	online chan bool
}

func newtcpclient(ip string, port int) *TcpClient {
	addstr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.Dial("tcp", addstr)
	if err == nil {
		client := &TcpClient{
			conn: conn,
			Pid:  0,
			X:    0,
			Y:    0,
			Z:    0,
			V:    0,
			online:make(chan bool),
		}
		return client
	} else {
		panic(err)
	}
}

//封包业务
func (this *TcpClient) Pack(msgId uint32, data []byte) ([]byte, error) {
	outbuf := bytes.NewBuffer([]byte{})
	if err := binary.Write(outbuf, binary.LittleEndian, uint32(len(data))); err != nil {
		return nil, err
	}
	if err := binary.Write(outbuf, binary.LittleEndian, msgId); err != nil {
		return nil, err
	}
	if err := binary.Write(outbuf, binary.LittleEndian, data); err != nil {
		return nil, err
	}
	return outbuf.Bytes(), nil
}

//拆包业务
func (this *TcpClient) Unpack(headData []byte) (*Message, error) {
	headBufReader := bytes.NewReader(headData)
	head := &Message{}
	if err := binary.Read(headBufReader, binary.LittleEndian, &head.Len); err != nil {
		return nil, err
	}
	if err := binary.Read(headBufReader, binary.LittleEndian, &head.MsgID); err != nil {
		return nil, err
	}
	return head, nil
}
//当前客户端发包
func (this *TcpClient) SendMsg(msgID uint32, data proto.Message) {
	binary_data, err := proto.Marshal(data)
	if err != nil {
		fmt.Println("marshal error", err)
		return
	}
	//打包LTV
	sendData, err := this.Pack(msgID, binary_data)
	if err == nil {
		this.conn.Write(sendData)
	} else {
		fmt.Println(err)
	}
}

//简单的ai动作
func (this *TcpClient)AIRobotAction(){
	//聊天 移动
	tp:=rand.Intn(2)
	if tp==0{
		//自动发送聊天信息
		content:=fmt.Sprintf("hello 我是 player :%d, 你是谁！？",this.Pid)
		msg:=&pb.Talk{
			Content:content,
		}
		this.SendMsg(2,msg)
	}else {
		//自动移动
		x:=this.X
		z:=this.Z
		randpos:=rand.Intn(2)
		if randpos==0{
			x-=float32(rand.Intn(10))
			z-=float32(rand.Intn(10))
		}else{
			x+=float32(rand.Intn(10))
			z+=float32(rand.Intn(10))
		}
		//纠正坐标
		if x>410{
			x=410
		}else if x<85{
			x=85
		}
		if z>400{
			z=400
		}else if z<75{
			z=75
		}
		randv:=rand.Intn(2)
		v :=this.V
		if randv==0{
			v=25
		}else{
			v=350
		}
		msg:=&pb.Position{
			X:x,
			Y:this.Y,
			Z:z,
			V:v,
		}
		this.SendMsg(3,msg)
	}

}
func (this *TcpClient) DoMsg(msg *Message) {
	fmt.Println("msgId =", msg.MsgID, ", msgLen = ", msg.Len, ", msg.Data=", msg.Data)
	if msg.MsgID == 1 {
		//服务器绘制给客户端 分配ID
		//解析proto协议
		syncpid := &pb.SyncPid{}
		proto.Unmarshal(msg.Data, syncpid)
		this.Pid = syncpid.Pid
	} else if msg.MsgID == 200 {
		//服务器回执给客户端广播数据
		//解析proto数据
		bdata := &pb.BroadCast{}
		proto.Unmarshal(msg.Data, bdata)
		if bdata.Tp == 2 && bdata.Pid == this.Pid {
			//更新本人坐标
			this.X = bdata.GetP().X
			this.Y = bdata.GetP().Y
			this.Z = bdata.GetP().Z
			this.V = bdata.GetP().V
			fmt.Printf("Player ID: %d online.. at(%f,%f,%f,%f)\n", this.Pid, this.X, this.Y, this.Z, this.V)
			this.online <- true
		} else if bdata.Tp == 1 {
			fmt.Println(fmt.Sprintf("世界聊天: 玩家:%d 说的话是 %s", bdata.Pid, bdata.GetContent()))

		}

	}

}


//永久的进行客户端读写业务
func (this *TcpClient) Start() {
	go func() {
		for {
			fmt.Println("deal server msg read and write")
			headData := make([]byte, 8)
			if _, err := io.ReadFull(this.conn, headData); err != nil {
				fmt.Println(err)
				return
			}
			messageHead, err := this.Unpack(headData)
			if err != nil {
				return
			}
			if messageHead.Len > 0 {
				messageHead.Data = make([]byte, messageHead.Len)
				if _, err := io.ReadFull(this.conn, messageHead.Data); err != nil {
					return
				}
			}
			this.DoMsg(messageHead)
		}
	}()
	select {
	case <-this.online:
		go func() {
			for  {
				this.AIRobotAction()
				time.Sleep(2*time.Second)
			}
		}()
	}
}
func main() {
	for i:=0;i<20;i++{
		client := newtcpclient("192.168.229.132", 8999)
		client.Start()
		time.Sleep(2*time.Second)
	}
	//阻塞
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Kill, os.Interrupt)
	sig := <-c
	fmt.Sprintf("rece signal", sig)
	return
}
