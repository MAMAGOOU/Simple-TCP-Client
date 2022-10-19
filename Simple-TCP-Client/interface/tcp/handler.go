package tcp

import (
	"context"
	"net"
)

type Handler interface {
	// Handle 传递超时时间，环境, conn 连接
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
