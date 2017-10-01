package logs

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs/logs"
)

type (
	Logs interface {
		// 设置实时log信息显示终端
		SetOutput(show io.Writer) Logs
		// 暂停输出日志
		Rest()
		// 恢复暂停状态，继续输出日志
		GoOn()
		// 是否开启日志捕获副本模式
		EnableStealOne(bool)
		// 按先后顺序实时捕获日志副本，每次返回1条，normal标记日志是否被关闭
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
		App(format string, v ...interface{})
		Notice(format string, v ...interface{})
		Warning(format string, v ...interface{})
		Error(format string, v ...interface{})
		Critical(format string, v ...interface{})
		Alert(format string, v ...interface{})
		Emergency(format string, v ...interface{})
	}
	mylog struct {
		*logs.BeeLogger
	}
)

var Log = func() Logs {
	p, _ := path.Split(config.LOG)
	// 不存在目录时创建目录
	d, err := os.Stat(p)
	if err != nil || !d.IsDir() {
		if err := os.MkdirAll(p, 0777); err != nil {
			// Log.Error("Error: %v\n", err)
		}
	}

	ml := &mylog{
		BeeLogger: logs.NewLogger(config.LOG_CAP, config.LOG_FEEDBACK_LEVEL),
	}

	// 是否打印行信息
	ml.BeeLogger.EnableFuncCallDepth(config.LOG_LINEINFO)
	// 全局日志打印级别（亦是日志文件输出级别）
	ml.BeeLogger.SetLevel(config.LOG_LEVEL)
	// 是否异步输出日志
	ml.BeeLogger.Async(config.LOG_ASYNC)
	// 设置日志显示位置
	ml.BeeLogger.SetLogger("console", map[string]interface{}{
		"level": config.LOG_CONSOLE_LEVEL,
	})

	// 是否保存所有日志到本地文件
	if config.LOG_SAVE {
		err = ml.BeeLogger.SetLogger("file", map[string]interface{}{
			"filename": config.LOG,
		})
		if err != nil {
			fmt.Printf("日志文档创建失败：%v", err)
		}
	}

	return ml
}()

func (self *mylog) SetOutput(show io.Writer) Logs {
	self.BeeLogger.SetLogger("console", map[string]interface{}{
		"writer": show,
		"level":  config.LOG_CONSOLE_LEVEL,
	})
	return self
}
