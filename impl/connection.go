package impl

import (
	"github.com/gorilla/websocket"
	"sync"
	"github.com/pkg/errors"
)

type Connection struct {
	wsConn *websocket.Conn
	inChan chan []byte
	outChan chan []byte
	closeChan chan []byte
	isClose bool
	mutex sync.Mutex
}
// 流程是readloop从websocket读取数据到inchan，封装的readmessage从inchan读取数据，相当于到了一层缓冲，可以并发处理了？
// 之后业务处理拿到数据之后通过writemessage写入到outchan，再之后writeloop从outchan拿到数据写入到底层websocket中

func InitConnection(wsConn *websocket.Conn)(conn *Connection, err error){
	conn = &Connection{
		wsConn:wsConn,
		inChan:make(chan []byte, 1000),
		outChan:make(chan []byte, 1000),
		closeChan:make(chan []byte, 1),
	}

	go conn.readLoop()
	go conn.rriteLoop()

	return
}

// 实现一个线程安全的readmessage
// 读取inchan中的数据
// 封装的readmessage和websocket的readmessage不一样
func (conn *Connection) ReadMessage() (data []byte, err error){
	select {
	case data = <-conn.inChan:
	case <- conn.closeChan:
		err = errors.New("Connection is closed")
	}
	return
}
// 写入data到outchan中
func (conn *Connection) WriteMessage(data []byte) (err error){
	select {
	case conn.outChan <- data:
	case <- conn.closeChan:
		err = errors.New("Connection is closed")
	}
	return
}

func (conn *Connection) Close() {
	// 线程安全的close，可重入的close
	conn.wsConn.Close()
	// 上面的close是线程安全的，但是这个不是，channel不能多次close因此需要设置isclose
	conn.mutex.Lock()
	if(!conn.isClose){
		close(conn.closeChan)
		conn.isClose = true
	}
}

func (conn *Connection) readLoop(){
	var (
		data []byte
		err error
	)
	for{
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}
		// 这里会阻塞，等待inchan有空闲的位置，但是connclose之后，这里仍然会阻塞着，消耗资源，因此需要让他知道，通知他。
		select{
		case conn.inChan <- data:
		case <-conn.closeChan:
			goto ERR
		}
	}
	ERR:{
		conn.Close()
	}
}

func (conn *Connection) rriteLoop(){
	var(
		data []byte
		err error
	)
	for{
		data = <- conn.outChan
		if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			goto ERR
		}
	}
	ERR:{
		conn.Close()
	}

}