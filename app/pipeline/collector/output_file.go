package collector

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/henrylee2cn/pholcus/app/pipeline/collector/data"
	bytesSize "github.com/henrylee2cn/pholcus/common/bytes"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	"github.com/henrylee2cn/pholcus/logs"
	// "github.com/henrylee2cn/pholcus/runtime/cache"
)

// 文件输出
func (self *Collector) outputFile(file data.FileCell) {
	// 复用FileCell
	defer data.PutFileCell(file)

	self.outCount[2]++

	// 路径： file/"RuleName"/"time"/"Name"
	p, n := filepath.Split(filepath.Clean(file["Name"].(string)))
	// dir := filepath.Join(config.FILE_DIR, util.FileNameReplace(self.namespace())+"__"+cache.StartTime.Format("2006年01月02日 15时04分05秒"), p)
	dir := filepath.Join(config.FILE_DIR, util.FileNameReplace(self.namespace()), p)

	// 文件名
	fileName := filepath.Join(dir, util.FileNameReplace(n))

	// 创建/打开目录
	d, err := os.Stat(dir)
	if err != nil || !d.IsDir() {
		if err := os.MkdirAll(dir, 0777); err != nil {
			self.outCount[3]++
			logs.Log.Error(
				" *     Fail  [文件下载：%v | KEYIN：%v | 批次：%v]   %v [ERROR]  %v\n",
				self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[3], fileName, err,
			)
			return
		}
	}

	// 文件不存在就以0777的权限创建文件，如果存在就在写入之前清空内容
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		self.outCount[3]++
		logs.Log.Error(
			" *     Fail  [文件下载：%v | KEYIN：%v | 批次：%v]   %v [ERROR]  %v\n",
			self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[3], fileName, err,
		)
		return
	}

	size, err := io.Copy(f, bytes.NewReader(file["Bytes"].([]byte)))
	f.Close()
	if err != nil {
		self.outCount[3]++
		logs.Log.Error(
			" *     Fail  [文件下载：%v | KEYIN：%v | 批次：%v]   %v (%s) [ERROR]  %v\n",
			self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[3], fileName, bytesSize.Format(uint64(size)), err,
		)
		return
	}

	// 输出统计
	self.outCount[3]++
	self.addFileSum(1)

	// 打印报告
	logs.Log.Informational(" * ")
	logs.Log.App(
		" *     [文件下载：%v | KEYIN：%v | 批次：%v]   %v (%s)\n",
		self.Spider.GetName(), self.Spider.GetKeyin(), self.outCount[3], fileName, bytesSize.Format(uint64(size)),
	)
	logs.Log.Informational(" * ")
}
