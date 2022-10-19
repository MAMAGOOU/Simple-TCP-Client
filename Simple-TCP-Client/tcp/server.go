package tcp

import (
	"context"
	"godis/interface/tcp"
	"godis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Config struct {
	Address string
}

// ListenAndServeWithSignal 监听服务使用信号
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {

	closeChan := make(chan struct{})
	// 给closeChan发送信号
	sigChan := make(chan os.Signal)
	// 转发给sigChan
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}

	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info("start listen")
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe 监听服务，一个接口一个协程
func ListenAndServe(
	listener net.Listener,
	handler tcp.Handler,
	closeChan <-chan struct{}) {
	go func() {
		<-closeChan
		logger.Info("shutting down")
		// 系统监听
		_ = listener.Close()
		// 业务引擎
		_ = handler.Close()
	}()
	// 方法在退出时关闭handler and listener
	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()
	ctx := context.Background()
	var waitDone sync.WaitGroup
	for true {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accepted link")
		// 服务加一，服务玩减一
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	//死循环出来时，不要马上退出，等待全部完成
	waitDone.Wait()
}
