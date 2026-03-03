package collector

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/util"
	"github.com/andeya/pholcus/config"
	"github.com/andeya/pholcus/runtime/cache"
)

// --- CSV Output ---

func init() {
	DataOutput["csv"] = func(col *Collector) (r result.VoidResult) {
		defer r.Catch()
		var (
			namespace = util.FileNameReplace(col.namespace())
			sheets    = make(map[string]*csv.Writer)
		)
		for _, datacell := range col.dataBuf {
			var subNamespace = util.FileNameReplace(col.subNamespace(datacell))
			if _, ok := sheets[subNamespace]; !ok {
				folder := config.Conf().TextDir + "/" + cache.StartTime.Format("2006-01-02 150405") + "/" + joinNamespaces(namespace, subNamespace)
				filename := fmt.Sprintf("%v/%v-%v.csv", folder, col.sum[0], col.sum[1])

				f, err := os.Stat(folder)
				if err != nil || !f.IsDir() {
					result.RetVoid(os.MkdirAll(folder, 0777)).Unwrap()
				}

				file, err := os.Create(filename)
				result.RetVoid(err).Unwrap()
				defer func(ns string, f *os.File) {
					if w := sheets[ns]; w != nil {
						w.Flush()
					}
					f.Close()
				}(subNamespace, file)

				file.WriteString("\xEF\xBB\xBF") // UTF-8 BOM

				sheets[subNamespace] = csv.NewWriter(file)
				th := col.MustGetRule(datacell["RuleName"].(string)).ItemFields
				if col.Spider.OutDefaultField() {
					th = append(th, "Url", "ParentUrl", "DownloadTime")
				}
				sheets[subNamespace].Write(th)
			}

			row := []string{}
			for _, title := range col.MustGetRule(datacell["RuleName"].(string)).ItemFields {
				vd := datacell["Data"].(map[string]interface{})
				if v, ok := vd[title].(string); ok || vd[title] == nil {
					row = append(row, v)
				} else {
					row = append(row, util.JSONString(vd[title]))
				}
			}
			if col.Spider.OutDefaultField() {
				row = append(row, datacell["Url"].(string))
				row = append(row, datacell["ParentUrl"].(string))
				row = append(row, datacell["DownloadTime"].(string))
			}
			sheets[subNamespace].Write(row)
		}
		return result.OkVoid()
	}
}
