package main

import (
	"net/http"
	"github.com/gorilla/websocket"
	"go-websocket/impl"
	"time"
)

var (
	// 定义一个upgrader
	upgrader = websocket.Upgrader{
		// 允许跨域，从zhibo.com跨域到websocket.com。
		CheckOrigin:func(r *http.Request) bool {
			return true;
		},
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request){
	var (
		wsConn *websocket.Conn
		err error
		conn *impl.Connection
		//msgType int
		data []byte
	)
	// 原来的http连接规范 w.Write([]byte("hello"))
	// 使用标准库做协议转换
	// 完成http握手应答,加上了Upgrader.websocket,返回websocket长连接conn
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		return
	}
	if conn, err = impl.InitConnection(wsConn); err != nil {
		goto ERR
	}

	// 心跳消息
	go func(){
		var(
			err error
		)
		for{
			if err = conn.WriteMessage([]byte("heartbeat")); err != nil {
				return
			}
			time.Sleep(1*time.Second)
		}
	}()

	// 正常流程
	for{
		if data, err = conn.ReadMessage(); err != nil{
			goto ERR
		}
		if err := conn.WriteMessage(data); err != nil{
			goto ERR
		}
	}

	ERR:
		conn.Close()
	//// websocket.conn做数据收发
	//for {
	//	// Text, Binary一般使用text因为json比较多
	//	if _, data, err = conn.ReadMessage(); err != nil {
	//		goto ERR
	//	}
	//	if err = conn.WriteMessage(websocket.TextMessage, data); err != nil{
	//		goto ERR
	//	}
	//}
	//ERR:
	//	println(err.Error())
	//	conn.Close()
}

func main() {
	// http://localhost:7777/ws
	http.HandleFunc("/ws", wsHandler)

	http.ListenAndServe("0.0.0.0:7777", nil)
}
