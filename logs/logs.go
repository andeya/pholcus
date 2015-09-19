package logs

import (
	"io"

	"github.com/henrylee2cn/beelogs"
)

type Logs interface {
	// 设置实时log信息显示终端
	SetOutput(show io.Writer) Logs
	// 设置日志截获水平，不设置不截获
	SetStealLevel() Logs
	// 设置是否异步输出
	Async(enable bool) Logs
	// 暂停输出日志
	Rest()
	// 恢复暂停状态，继续输出日志
	GoOn()
	// 按先后顺序实时截获日志，每次返回1条，normal标记日志是否被关闭
	StealOne() (level int, msg string, normal bool)
	// 正常关闭日志输出
	Close()
	// 返回运行状态，如0,"RUN"
	Status() (int, string)
	DelLogger(adaptername string) error
	SetLogger(adaptername string, config map[string]interface{}) error

	// 以下打印方法除正常log输出外，若为客户端或服务端模式还将进行socket信息发送
	Debug(format string, v ...interface{})
	Informational(format string, v ...interface{})
	Notice(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Error(format string, v ...interface{})
	Critical(format string, v ...interface{})
	Alert(format string, v ...interface{})
	Emergency(format string, v ...interface{})
}

const (
	// 默认日志缓存
	MaxLogCache = 10000
	// 默认日志文件
	FileName = "pholcus.log"
)

var (
	LevelEmergency     = beelogs.LevelEmergency
	LevelAlert         = beelogs.LevelAlert
	LevelCritical      = beelogs.LevelCritical
	LevelError         = beelogs.LevelError
	LevelWarning       = beelogs.LevelWarning
	LevelNotice        = beelogs.LevelNotice
	LevelInformational = beelogs.LevelInformational
	LevelDebug         = beelogs.LevelDebug
)

var Log = NewLogs()

type mylog struct {
	*beelogs.BeeLogger
}

func NewLogs(enableFuncCallDepth ...bool) Logs {
	ml := &mylog{
		BeeLogger: beelogs.NewLogger(MaxLogCache),
	}

	// 是否打印行号
	if len(enableFuncCallDepth) > 0 {
		ml.BeeLogger.EnableFuncCallDepth(enableFuncCallDepth[0])
	}

	ml.BeeLogger.SetLevel(LevelDebug)

	ml.BeeLogger.SetLogger("console", map[string]interface{}{
		"level": LevelInformational,
	})

	ml.BeeLogger.SetLogger("file", map[string]interface{}{
		"filename": FileName,
	})

	return ml
}

func (self *mylog) SetOutput(show io.Writer) Logs {
	self.BeeLogger.SetLogger("console", map[string]interface{}{
		"writer": show,
		"level":  LevelInformational,
	})
	return self
}

func (self *mylog) SetStealLevel() Logs {
	self.BeeLogger.SetStealLevel(LevelNotice)
	return self
}

func (self *mylog) Async(enable bool) Logs {
	self.BeeLogger.Async(enable)
	return self
}

func Debug(open bool) {
	if open {
		LevelEmergency = beelogs.LevelDebug
		LevelAlert = beelogs.LevelDebug
		LevelCritical = beelogs.LevelDebug
		LevelError = beelogs.LevelDebug
		LevelWarning = beelogs.LevelDebug
		LevelNotice = beelogs.LevelDebug
		LevelInformational = beelogs.LevelDebug
		LevelDebug = beelogs.LevelDebug
	} else {
		LevelEmergency = beelogs.LevelEmergency
		LevelAlert = beelogs.LevelAlert
		LevelCritical = beelogs.LevelCritical
		LevelError = beelogs.LevelError
		LevelWarning = beelogs.LevelWarning
		LevelNotice = beelogs.LevelNotice
		LevelInformational = beelogs.LevelInformational
		LevelDebug = beelogs.LevelDebug
	}
	Log = NewLogs(open)
}
