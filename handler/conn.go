package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

//用户连接结构体
type Connection struct {
	ws   *websocket.Conn
	sc   chan []byte
	data *Data
}

//用户在线名单列表
var UserList = []string{}

type Data struct {
	Ip       string   `json:"ip"`
	User     string   `json:"user"`
	From     string   `json:"from"`
	Type     string   `json:"type"`
	Content  string   `json:"content"`
	UserList []string `json:"user_list"`
}

//数据写入器
func (c *Connection) Writer() {
	//取出发送信息并写入
	log.Printf("写消息进入工作模式...")
	for message := range c.sc {
		log.Printf("消息体：" + string(message))
		c.ws.WriteMessage(websocket.TextMessage, message)
		log.Printf("发送消息体完成")
	}
	c.ws.Close()
}

//数据读取器
func (c *Connection) Reader(h User) {
	log.Printf("循环读数据")
	for {
		//接收ws信息
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			h.r <- c
			break
		}
		json.Unmarshal(message, &c.data)
		//log.Printf("用户结构%+v\n", c.data)
		log.Printf("读到消息内容%+v\n", string(message))
		//解析信息类型
		switch c.data.Type {
		//用户登录
		case "login":
			c.data.User = c.data.Content
			c.data.From = c.data.User
			//在线人数增加
			UserList = append(UserList, c.data.User)
			c.data.UserList = UserList
			data_b, _ := json.Marshal(c.data)
			//发送信息
			//log.Printf("发送给h.b的数据%v", string(data_b))
			h.b <- data_b
			log.Printf("发送给h.b的数据成功")
		case "user":
			c.data.Type = "user"
			data_b, _ := json.Marshal(c.data)
			h.b <- data_b
		//用户登出
		case "logout":
			c.data.Type = "logout"
			//在线人数减少
			UserList = Del(UserList, c.data.User)
			data_b, _ := json.Marshal(c.data)
			//删除连接
			h.b <- data_b
			//发送用户离线信息
			h.r <- c
			h.u <- c
		default:
			fmt.Print("========default================")
		}
	}
}

//删除登出的用户，维护在线用户名单
func Del(slice []string, user string) []string {
	count := len(slice)
	if count == 0 {
		return slice
	}
	if count == 1 && slice[0] == user {
		return []string{}
	}
	var n_slice = []string{}
	for i := range slice {
		if slice[i] == user && i == count {
			return slice[:count]
		} else if slice[i] == user {
			n_slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	return n_slice
}
