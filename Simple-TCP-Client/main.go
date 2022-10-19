package main

import (
	"fmt"
	"godis/config"
	"godis/lib/logger"
	"godis/tcp"
	"os"
)

const configFile string = "redis.conf"

// 如果没有配置文件，默认的配置
var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

// 判断配置文件是否存在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}
func main() {
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "godis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})
	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}
	// 启动服务
	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
		},
		tcp.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}
