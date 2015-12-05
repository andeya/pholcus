package collector

import (
	"io"
	"os"
	"path"
	"runtime"

	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
)

//文件输出管理
func (self *Collector) SaveFile() {
	for !(self.CtrlLen() == 0 && len(self.FileChan) == 0) {
		select {
		case file := <-self.FileChan:
			self.outCount[2]++

			// 统计输出文件数
			self.setFileSum(1)

			// 路径： file/"RuleName"/"time"/"Name"
			p, n := path.Split(file["Name"].(string))
			dir := config.COMM_PATH.FILE + `/` + util.FileNameReplace(self.namespace()) + "__" + self.startTime.Format("2006年01月02日 15时04分05秒") + `/` + p

			// 创建/打开目录
			d, err := os.Stat(dir)
			if err != nil || !d.IsDir() {
				if err := os.MkdirAll(dir, 0777); err != nil {
					logs.Log.Error("Error: %v\n", err)
				}
			}

			// 创建文件
			fileName := dir + util.FileNameReplace(n)
			f, _ := os.Create(fileName)
			io.Copy(f, file["Body"].(io.ReadCloser))
			f.Close()
			file["Body"].(io.ReadCloser).Close()

			// 打印报告
			logs.Log.Informational(" * ")
			logs.Log.Notice(" *     [任务：%v | 关键词：%v]   成功下载文件： %v \n", self.Spider.GetName(), self.Spider.GetKeyword(), fileName)
			logs.Log.Informational(" * ")

			self.outCount[3]++
		default:
			runtime.Gosched()
		}
	}
}
