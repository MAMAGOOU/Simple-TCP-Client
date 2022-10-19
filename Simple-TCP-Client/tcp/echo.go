package tcp

import (
	"bufio"
	"context"
	"godis/lib/logger"
	"godis/lib/sync/atomic"
	"godis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

// EchoClient 客户端
type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

// Close 做关闭的统一接口
func (e *EchoClient) Close() error {
	e.Waiting.WaitWithTimeout(10 * time.Second)
	_ = e.Conn.Close()
	return nil
}

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

func (handler *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if handler.closing.Get() {
		_ = conn.Close()
	}
	// 将新连接包装成client，
	client := &EchoClient{
		Conn: conn,
	}
	//将包装好的client塞到map
	//将value设置位空结构体，不占用任何空间
	handler.activeConn.Store(client, struct{}{})
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			//EOF为操作系统的结束符
			if err == io.EOF {
				logger.Info("Connection close")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		//告诉客户端我们在做一个回发的业务，业务没有完成之前不要关掉，但如果过了10s还没有做完那么可以关闭
		client.Waiting.Add(1)
		b := []byte(msg)
		//就是一个见到那回发操作，有什么错误也不用管
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

func (handler *EchoHandler) Close() error {
	logger.Info("handler shutting down")
	// 正在关闭状态，如果收到连接直接关闭不做服务
	handler.closing.Set(true)
	handler.activeConn.Range(func(key, value interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Conn.Close()
		//
		return true
	})
	return nil
}
