package main

import (
	"flag"
	"fmt"
	"time"
	"wecom/app"
	"wecom/config"
	"wecom/mw"
	"wecom/zaplog"

	"code.in.wezhuiyi.com/nlp-platform/service-tools/golang/pkg/util"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	etcd "go.etcd.io/etcd/v3/clientv3"
)

func main() {
	etcdEP1 := flag.String("etcd1", "172.18.160.34:2379", "etcd1 end point")
	etcdEP2 := flag.String("etcd2", "172.18.160.34:2479", "etcd2 end point")
	etcdEP3 := flag.String("etcd3", "172.18.160.34:2579", "etcd3 end point")

	dbIP := flag.String("db_ip", "172.18.160.34", "mysql DB IP address")
	dbPort := flag.Int("db_port", 3306, "mysql DB port")
	dbUsername := flag.String("db_username", "root", "mysql DB username")
	dbPassword := flag.String("db_password", "uWXf87plmQGz8zMM", "mysql DB password")
	debugMode := flag.Bool("debug_mode", false, "mysql DB debug mode")

	logFilename := flag.String("log", ".idea/wecom.log", "log filename")
	logLevel := flag.String("log_level", "DEBUG", "log level")
	safeMode := flag.Bool("safe_mode", true, "https mode")
	certPath := flag.String("cert_path", "D://runSpace//wecom//ca//server.crt", "server ca")
	keyPath := flag.String("key_path", "D://runSpace//wecom//ca//server.pem", "private key")
	flag.Parse()

	util.SetProjectRootEnv("WECOM")

	mw.InitETCD(etcd.Config{
		Endpoints:   []string{*etcdEP1, *etcdEP2, *etcdEP3},
		DialTimeout: 5 * time.Second,
	})

	// init logger
	config.InitConfig(*logLevel, *logFilename)

	if err := zaplog.InitLogger(config.Conf.LogConfig); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		return
	}

	gin.SetMode(config.Conf.Mode)

	mw.InitDB(config.DBConfig{
		Domain:    *dbIP,
		Port:      *dbPort,
		Username:  *dbUsername,
		Password:  *dbPassword,
		Database:  mw.DBName,
		DebugMode: *debugMode,
	})

	// 创建一个默认的路由引擎
	r := app.SetupRouter()

	// 注册zap相关中间件
	r.Use(zaplog.GinLogger(), zaplog.GinRecovery(true))

	// 是否启动HTTPs服务，默认在0.0.0.0:80启动服务
	addr := fmt.Sprintf(":%v", 80)
	var err error
	if *safeMode {
		err = r.RunTLS(addr, *certPath, *keyPath)
	} else {
		err = r.Run(addr)
	}
	if err != nil {
		zap.L().Error(fmt.Sprintf("gin run failed, %v", err))
		return
	}
}
