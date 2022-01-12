package handler

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type User struct {
	//当前在线connection信息
	c map[*Connection]bool
	//删除connection
	u chan *Connection
	//传递数据
	b chan []byte
	//加入connection
	r chan *Connection
}

func NewUser() User {
	return User{
		c: make(map[*Connection]bool),
		u: make(chan *Connection),
		b: make(chan []byte),
		r: make(chan *Connection),
	}
}

var wu = &websocket.Upgrader{ReadBufferSize: 512, WriteBufferSize: 512, CheckOrigin: func(r *http.Request) bool { return true }}

//websocket服务
func (h *User) Mys(w http.ResponseWriter, r *http.Request) {
	log.Printf("启动Mys服务")
	//协议升级
	ws, err := wu.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	//创建连接
	log.Printf("创建连接")
	c := &Connection{ws: ws, sc: make(chan []byte, 256), data: &Data{}}
	//connection加入hub管理
	log.Printf("h.r <-c 把连接给用户管理起来")
	h.r <- c //把连接给用户管理起来
	log.Printf("启动一个写协程 go c.Writer()")
	go c.Writer()
	log.Printf("去读一下用户发了什么东西 c.Reader")
	c.Reader(*h)
	//退出登录
	defer h.Logout(c)
}

//用户中心，维护多个用户的connection

func (h *User) Run() {
	log.Printf("启动主go服务等候...")
	for {
		select {
		//用户连接，添加connection信息
		case c := <-h.r:
			log.Printf("启动主go服务等候...用户连接来了...")
			h.c[c] = true
			c.data.Ip = c.ws.RemoteAddr().String()
			c.data.Type = "handshake"
			c.data.UserList = UserList
			data_b, _ := json.Marshal(c.data)
			//发送给写入器
			log.Printf("把组装好的用户消息发送给写入器")
			c.sc <- data_b
		//删除指定用户连接
		case c := <-h.u:
			log.Printf("退出聊天室...")
			if _, ok := h.c[c]; ok {
				delete(h.c, c)
				close(c.sc)
			}
		//向聊天室在线人员发送信息
		case data := <-h.b:
			log.Printf("接收到 hb数据，向聊天室在线人员发送信息，所有人的连接信息：%+v\n", h.c)
			for c := range h.c {
				select {
				//发送数据
				case c.sc <- data:
				//发送不成功则删除connection信息
				default:
					delete(h.c, c)
					close(c.sc)
				}
			}
		}
	}
}

//退出
func (h *User) Logout(c *Connection) {
	c.data.Type = "logout"
	UserList = Del(UserList, c.data.User)
	c.data.UserList = UserList
	c.data.Content = c.data.User
	data_b, _ := json.Marshal(c.data)
	h.b <- data_b
	h.r <- c
	h.u <- c
}
