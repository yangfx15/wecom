package mw

import (
	"fmt"
	"log"
	"os"
	"time"
	"wecom/config"
	"wecom/zaplog"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const DBName = "wecom"

var Db *gorm.DB

type Dialog struct {
	ID         int        `gorm:"type:bigint(20) NOT NULL AUTO_INCREMENT"`
	UserName   string     `gorm:"type:varchar(32) DEFAULT ''"`
	App        string     `gorm:"type:varchar(32) DEFAULT ''"`
	Question   string     `gorm:"type:text DEFAULT ''"`
	Answer     string     `gorm:"type:text DEFAULT ''"`
	MessageId  int64      `gorm:"type:bigint(20) DEFAULT 0"`
	Status     string     `gorm:"type:varchar(32) DEFAULT ''"`
	CreateTime *time.Time `gorm:"type:timestamp DEFAULT CURRENT_TIMESTAMP"`
}

func InitDB(x config.DBConfig) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		x.Username, x.Password, x.Domain, x.Port, x.Database)
	var err error
	// 日志配置
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别为info
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  true,        // 彩色打印
		},
	)
	cfg := &gorm.Config{}
	if x.DebugMode {
		cfg.Logger = newLogger
	}
	Db, err = gorm.Open(mysql.Open(dsn), cfg)
	if err != nil {
		panic("failed to connect database")
	}
	//defer db.Close()

	//自动检查 Product 结构是否变化，变化则进行迁移
	err = Db.AutoMigrate(&Dialog{})
	if err != nil {
		zaplog.Fatalf("AutoMigrate failed, %v", err)
	}
}
