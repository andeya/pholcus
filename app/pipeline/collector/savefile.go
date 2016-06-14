package collector

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/henrylee2cn/pholcus/app/pipeline/collector/data"
	"github.com/henrylee2cn/pholcus/common/bytes"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	"github.com/henrylee2cn/pholcus/runtime/cache"
)

//文件输出管理
func (self *Collector) SaveFile() {
	for !(self.CtrlLen() == 0 && len(self.FileChan) == 0) {
		select {
		case file := <-self.FileChan:
			self.outCount[2]++

			// 路径： file/"RuleName"/"time"/"Name"
			p, n := filepath.Split(filepath.Clean(file["Name"].(string)))
			dir := filepath.Join(config.FILE_DIR, util.FileNameReplace(self.namespace())+"__"+cache.StartTime.Format("2006年01月02日 15时04分05秒"), p)

			// 创建/打开目录
			d, err := os.Stat(dir)
			if err != nil || !d.IsDir() {
				if err := os.MkdirAll(dir, 0777); err != nil {
					logs.Log.Error("Error: %v\n", err)
				}
			}

			// 输出统计
			self.addFileSum(1)

			// 创建文件
			fileName := filepath.Join(dir, util.FileNameReplace(n))
			f, _ := os.Create(fileName)
			size, _ := io.Copy(f, file["Body"].(io.ReadCloser))
			f.Close()
			file["Body"].(io.ReadCloser).Close()

			// 打印报告
			logs.Log.Informational(" * ")
			logs.Log.App(" *     [任务：%v | KEYIN：%v]   成功下载文件： %v (%s)\n",
				self.Spider.GetName(), self.Spider.GetKeyin(), fileName, bytes.Format(uint64(size)))
			logs.Log.Informational(" * ")

			self.outCount[3]++

			// 复用FileCell
			data.PutFileCell(file)
		default:
			runtime.Gosched()
		}
	}
}
