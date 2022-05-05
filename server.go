package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct{
	Ip string
	Port int
	// 在线用户列表
	OnlineMap map[string]*User
	mapLock sync.RWMutex

	//消息广播的hannel
	Message chan string
}

//创建一个server的接口
func NewServer(ip string, port int) *Server {
	server:=&Server{
		Ip:ip,
		Port:port,
		OnlineMap:make(map[string]*User),
		Message:make(chan string),
	}
	return server
}

//监听Message广播消息channel的goroutine,一旦有消息就发送给全部在线用户
func (this *Server) ListenMessage(){
	for{
		msg:=<-this.Message

		//将msg发送给全部在线user
		this.mapLock.Lock()
		for _,cli:=range this.OnlineMap{
			cli.C<-msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息的方法
func (this *Server) Broadcast(user *User,msg string) {
	sendMsg:="["+user.Addr+"] "+user.Name+":"+msg
	this.Message <- sendMsg
}

//当前连接的业务
func (this *Server) Handler(conn net.Conn){
	// fmt.Println("连接建立成功")

	user:=NewUser(conn,this)

	user.Online()

	//监听用户是否活跃的channel
	isLive:=make(chan bool)

	//接收客户端发送的数据
	go func(){
		buf:=make([]byte,4096)
		for{
			n,err:=conn.Read(buf)
			if n==0{//合法关闭
				this.Broadcast(user,"下线")
				return
			}
			if err!=nil && err!=io.EOF{
				fmt.Println("Conn Read err:",err)
				return
			}

			//提取用户的消息（去除\n)
			msg:=string(buf[:n-1])

			//用户对msg的处理
			user.DoMessage(msg)

			//用户的任意消息，代表当前用户是活跃的
			isLive<-true
		}
	}()

	//当前handler阻塞
	for{
		select{
		case<-isLive:
			//当前用户是活跃的，应重置定时器
			//不做任何事情，为激活select，更新下面定时器
		case<-time.After(time.Second*300):
			//已经超时
			//将当前User强制关闭

			user.SendMsg("你被踢了")

			//销毁资源
			close(user.C)

			//关闭连接
			conn.Close()

			//推出当前Handler
			return //runtime.Goexit()
		}
	}
}

//启动服务器的接口
func (this *Server) Start(){
	//socket listen
	listener, err := net.Listen("tcp",fmt.Sprintf("%s:%d",this.Ip,this.Port))
	if err != nil {
		fmt.Println("net.Listen err:",err)
		return
	}
	//close listen socket
	defer listener.Close()

	//启动监听Message的goroutine
	go this.ListenMessage()

	for{
		//accept
		conn,err:=listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:",err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}
}