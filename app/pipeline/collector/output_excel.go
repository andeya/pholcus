package collector

import (
	"fmt"
	"os"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/common/xlsx"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/logs"
	"github.com/andeya/pholcus/runtime/cache"
)

// --- Excel Output ---

func init() {
	DataOutput["excel"] = func(col *Collector) (r result.VoidResult) {
		defer r.Catch()
		var (
			file   *xlsx.File
			row    *xlsx.Row
			cell   *xlsx.Cell
			sheets = make(map[string]*xlsx.Sheet)
		)

		file = xlsx.NewFile()

		for _, datacell := range col.dataBuf {
			var subNamespace = util.FileNameReplace(col.subNamespace(datacell))
			if _, ok := sheets[subNamespace]; !ok {
				r := file.AddSheet(subNamespace)
				if r.IsErr() {
					logs.Log().Error("%v", r.UnwrapErr())
					continue
				}
				sheet := r.Unwrap()
				sheets[subNamespace] = sheet
				row = sheets[subNamespace].AddRow()
				for _, title := range col.MustGetRule(datacell["RuleName"].(string)).ItemFields {
					row.AddCell().Value = title
				}
				if col.Spider.OutDefaultField() {
					row.AddCell().Value = "Url"
					row.AddCell().Value = "ParentUrl"
					row.AddCell().Value = "DownloadTime"
				}
			}

			row = sheets[subNamespace].AddRow()
			for _, title := range col.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				cell = row.AddCell()
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					cell.Value = v
				} else {
					cell.Value = util.JSONString(vd[title])
				}
			}
			if col.Spider.OutDefaultField() {
				row.AddCell().Value = datacell["Url"].(string)
				row.AddCell().Value = datacell["ParentUrl"].(string)
				row.AddCell().Value = datacell["DownloadTime"].(string)
			}
		}
		folder := config.Conf().TextDir + "/" + cache.StartTime.Format("2006-01-02 150405")
		filename := fmt.Sprintf("%v/%v__%v-%v.xlsx", folder, util.FileNameReplace(col.namespace()), col.sum[0], col.sum[1])

		f2, err := os.Stat(folder)
		if err != nil || !f2.IsDir() {
			result.RetVoid(os.MkdirAll(folder, 0777)).Unwrap()
		}

		return file.Save(filename)
	}
}
