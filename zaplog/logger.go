package zaplog

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
	"wecom/config"

	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logTmFmtWithMS = "2006-01-02 15:04:05.000"

var lg *zap.Logger
var zaplog *zap.Logger

func Debug(msg string, fields ...zap.Field) {
	zaplog.Debug(msg, fields...)
}

func Debugf(msg string, a ...interface{}) {
	zaplog.Debug(fmt.Sprintf(msg, a...))
}

func Info(msg string, fields ...zap.Field) {
	zaplog.Info(msg, fields...)
}

func Infof(msg string, a ...interface{}) {
	zaplog.Info(fmt.Sprintf(msg, a...))
}

func Error(msg string, fields ...zap.Field) {
	zaplog.Error(msg, fields...)
}

func Errorf(msg string, a ...interface{}) {
	zaplog.Error(fmt.Sprintf(msg, a...))
}

func Warn(msg string, fields ...zap.Field) {
	zaplog.Warn(msg, fields...)
}

func Warnf(msg string, a ...interface{}) {
	zaplog.Warn(fmt.Sprintf(msg, a...))
}

func Panic(msg string, fields ...zap.Field) {
	zaplog.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	zaplog.Fatal(msg, fields...)
}

func Fatalf(msg string, a ...interface{}) {
	zaplog.Fatal(fmt.Sprintf(msg, a...))
}

func InitLogger(cfg *config.LogConfig) (err error) {
	writeSyncer := getLogWriter(cfg.Filename, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)
	encoder := getEncoder()
	var l = new(zapcore.Level)
	err = l.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return
	}
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout)),
		l)

	lg = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(lg) // 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可

	zaplog = zap.L()
	return
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("[%s]", t.Format(logTmFmtWithMS)))
	}
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fmt.Sprintf("[%s]", l.CapitalString()))
	}
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		//enc.AppendString(fmt.Sprintf("[%s]", caller.TrimmedPath()))
		pc, file, line, ok := runtime.Caller(6)
		if ok {
			caller.PC = pc
			caller.File = file
			caller.Line = line
		}
		// 如果用fmt.Sprintf("[%s]", caller.TrimmedPath())，则没有链接指向日志打印路径
		enc.AppendString(fmt.Sprintf(strings.Join([]string{caller.TrimmedPath()}, "")))
	}
	//return zapcore.NewJSONEncoder(encoderConfig)
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackup,
		MaxAge:     maxAge,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		lg.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					lg.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					_ = c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
