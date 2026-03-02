package collector

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/app/pipeline/collector/data"
	bytesSize "github.com/andeya/pholcus/common/bytes"
	"github.com/andeya/pholcus/common/closer"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
	// "github.com/andeya/pholcus/runtime/cache"
)

// outputFile writes a file cell to disk.
func (c *Collector) outputFile(file data.FileCell) {
	defer func() {
		data.PutFileCell(file)
		c.wait.Done()
	}()

	// Path format: file/"RuleName"/"time"/"Name"
	p, n := filepath.Split(filepath.Clean(file["Name"].(string)))
	// dir := filepath.Join(config.Conf().FileDir, util.FileNameReplace(c.namespace())+"__"+cache.StartTime.Format("2006-01-02 15:04:05"), p)
	dir := filepath.Join(config.Conf().FileDir, util.FileNameReplace(c.namespace()), p)

	fileName := filepath.Join(dir, util.FileNameReplace(n))

	d, err := os.Stat(dir)
	if err != nil || !d.IsDir() {
		if r := result.RetVoid(os.MkdirAll(dir, 0777)); r.IsErr() {
			logs.Log().Error(
				" *     Fail  [File download: %v | KEYIN: %v | Batch: %v]   %v [ERROR]  %v\n",
				c.Spider.GetName(), c.Spider.GetKeyin(), atomic.LoadUint64(&c.fileBatch), fileName, r.UnwrapErr(),
			)
			return
		}
	}

	// Create file with 0777 if not exists, truncate if exists
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		logs.Log().Error(
			" *     Fail  [File download: %v | KEYIN: %v | Batch: %v]   %v [ERROR]  %v\n",
			c.Spider.GetName(), c.Spider.GetKeyin(), atomic.LoadUint64(&c.fileBatch), fileName, err,
		)
		return
	}
	defer closer.LogClose(f, logs.Log().Error)

	size, err := io.Copy(f, bytes.NewReader(file["Bytes"].([]byte)))
	if err != nil {
		logs.Log().Error(
			" *     Fail  [File download: %v | KEYIN: %v | Batch: %v]   %v (%s) [ERROR]  %v\n",
			c.Spider.GetName(), c.Spider.GetKeyin(), atomic.LoadUint64(&c.fileBatch), fileName, bytesSize.Format(uint64(size)), err,
		)
		return
	}

	c.addFileSum(1)

	logs.Log().Informational(" * ")
	logs.Log().App(
		" *     [File download: %v | KEYIN: %v | Batch: %v]   %v (%s)\n",
		c.Spider.GetName(), c.Spider.GetKeyin(), atomic.LoadUint64(&c.fileBatch), fileName, bytesSize.Format(uint64(size)),
	)
	logs.Log().Informational(" * ")
}
