// getmetrics.go - transform Eclipse Metrics (v3) XML report into CSV files for each metric

package main

import (
	"flag"
	"fmt"
	"github.com/henrylee2cn/tconfig/xml/x2j"
	"os"
	"time"
)

func main() {
	var file string
	flag.StringVar(&file, "file", "", "file to process")
	flag.Parse()

	fh, fherr := os.Open(file)
	if fherr != nil {
		fmt.Println("fherr:", fherr.Error())
		return
	}
	defer fh.Close()
	fs, _ := fh.Stat()
	fmt.Println(time.Now().String(), "... File Opened:", file)

	b := make([]byte, fs.Size())
	n, frerr := fh.Read(b)
	if frerr != nil {
		fmt.Println("frerr:", frerr.Error())
		return
	}
	if int64(n) != fs.Size() {
		fmt.Println("n:", n, "fs.Size():", fs.Size())
		return
	}
	fmt.Println(time.Now().String(), "... File Read - size:", fs.Size())

	m := make(map[string]interface{}, 0)
	merr := x2j.Unmarshal(b, &m)
	if merr != nil {
		fmt.Println("merr:", merr.Error())
		return
	}
	fmt.Println(time.Now().String(), "... XML Unmarshaled - len:", len(m))

	metricVals := x2j.ValuesFromKeyPath(m, "Metrics.Metric", true)
	fmt.Println(time.Now().String(), "... ValuesFromKeyPath - len:", len(metricVals))

	for _, v := range metricVals {
		aMetricVal := v.(map[string]interface{})
		id := aMetricVal["-id"].(string)
		desc := aMetricVal["-description"].(string)
		mf, mferr := os.OpenFile(id+".csv", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if mferr != nil {
			fmt.Println("mferr:", mferr.Error())
			return
		}
		fmt.Print(time.Now().String(), " id: ", id, " desc: ", desc)
		mf.WriteString(id + "," + desc + "\n")
		for key, val := range aMetricVal {
			switch key {
			case "Values":
				// extract the list of "Value" from map
				values := val.(map[string]interface{})["Value"].([]interface{})
				fmt.Println(" len(Values):", len(values))
				var gotKeys bool
				for _, vval := range values {
					valueEntry := vval.(map[string]interface{})
					// extract keys as column header on first pass
					if !gotKeys {
						// print out the keys
						var gotFirstKey bool
						for kk, _ := range valueEntry {
							if gotFirstKey {
								mf.WriteString(",")
							} else {
								gotFirstKey = true
							}
							// strip prepended hyphen
							mf.WriteString(kk[1:])
						}
						mf.WriteString("\n")
						gotKeys = true
					}
					// print out values
					var gotFirstVal bool
					for _, vv := range valueEntry {
						if gotFirstVal {
							mf.WriteString(",")
						} else {
							gotFirstVal = true
						}
						mf.WriteString(vv.(string))
					}
					mf.WriteString("\n")
				}
			case "Value":
				vv := val.(map[string]interface{})
				fmt.Println(" len(Value):", len(vv))
				mf.WriteString("value\n" + vv["-value"].(string) + "\n")
			}
		}
		mf.Close()
	}
}
