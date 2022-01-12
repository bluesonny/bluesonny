package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"go-websocket-chat/handler"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter()
	//开启协程启动connection服务管理中心
	h := handler.NewUser()
	go h.Run() //等待hr构造完整数据 然后给发送通道
	//创建ws服务
	router.HandleFunc("/ws", h.Mys)
	//启动http服务
	log.Printf("启动http服务")
	if err := http.ListenAndServe("127.0.0.1:8090", router); err != nil {
		fmt.Println("err:", err)
	}
}
