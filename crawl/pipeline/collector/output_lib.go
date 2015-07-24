//数据输出
package collector

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/tealeg/xlsx"
	"gopkg.in/mgo.v2"
	// "log"
	// "gopkg.in/mgo.v2/bson"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/henrylee2cn/pholcus/common/util"
	"github.com/henrylee2cn/pholcus/config"
	. "github.com/henrylee2cn/pholcus/reporter"
	"github.com/henrylee2cn/pholcus/runtime/cache"
	"os"
	"strings"
	// "time"
	"github.com/henrylee2cn/teleport"
	"sync"
)

var Output = make(map[string]func(self *Collector, dataIndex int))
var OutputLib []string

func init() {
	defer func() {
		// 获取输出方式列表
		for out, _ := range Output {
			OutputLib = append(OutputLib, out)
		}
		util.StringsSort(OutputLib)
	}()

	/************************ excel 输出 ***************************/
	Output["excel"] = func(self *Collector, dataIndex int) {
		defer func() {
			if err := recover(); err != nil {
				Log.Println(err)
			}
		}()

		var file *xlsx.File
		var sheet *xlsx.Sheet
		var row *xlsx.Row
		var cell *xlsx.Cell
		var err error

		folder1 := "result/data"
		folder2 := folder1 + "/" + self.startTime
		filename := folder2 + "/" + self.Spider.GetName() + "_" + self.Spider.GetKeyword() + " " + fmt.Sprintf("%v", self.sum[0]) + "-" + fmt.Sprintf("%v", self.sum[1]) + ".xlsx"

		// 创建文件
		file = xlsx.NewFile()

		// 添加分类数据工作表
		for Name, Rule := range self.GetRules() {
			// 跳过不输出的数据
			if len(Rule.GetOutFeild()) == 0 {
				continue
			}
			// 添加工作表
			sheet = file.AddSheet(Name)
			// 写入表头
			row = sheet.AddRow()
			for _, title := range Rule.GetOutFeild() {
				cell = row.AddCell()
				cell.Value = title
			}
			cell = row.AddCell()
			cell.Value = "当前链接"
			cell = row.AddCell()
			cell.Value = "上级链接"
			cell = row.AddCell()
			cell.Value = "下载时间"

			num := 0 //小计
			for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
				if datacell["RuleName"].(string) == Name {
					row = sheet.AddRow()
					for _, title := range Rule.GetOutFeild() {
						cell = row.AddCell()
						vd := datacell["Data"].(map[string]interface{})
						if v, ok := vd[title].(string); ok || vd[title] == nil {
							cell.Value = v
						} else {
							cell.Value = util.JsonString(vd[title])
						}
					}
					cell = row.AddCell()
					cell.Value = datacell["Url"].(string)
					cell = row.AddCell()
					cell.Value = datacell["ParentUrl"].(string)
					cell = row.AddCell()
					cell.Value = datacell["DownloadTime"].(string)
					num++
				}
			}

			// Log.Printf("[任务：%v | 关键词：%v | 小类：%v] 输出 %v 条数据！！！\n", self.Spider.GetName(), self.Spider.GetKeyword(), Name, num)

		}

		// 创建/打开目录
		f2, err := os.Stat(folder2)
		if err != nil || !f2.IsDir() {
			if err := os.MkdirAll(folder2, 0777); err != nil {
				Log.Printf("Error: %v\n", err)
			}
		}

		// 保存文件
		err = file.Save(filename)

		if err != nil {
			Log.Println(err)
		}

	}

	/************************ CSV 输出 ***************************/
	Output["csv"] = func(self *Collector, dataIndex int) {
		defer func() {
			if err := recover(); err != nil {
				Log.Println(err)
			}
		}()

		folder1 := "result/data"
		folder2 := folder1 + "/" + self.startTime
		filenameBase := folder2 + "/" + self.Spider.GetName() + "_" + self.Spider.GetKeyword() + " " + fmt.Sprintf("%v", self.sum[0]) + "-" + fmt.Sprintf("%v", self.sum[1])

		// 创建/打开目录
		f2, err := os.Stat(folder2)
		if err != nil || !f2.IsDir() {
			if err := os.MkdirAll(folder2, 0777); err != nil {
				Log.Printf("Error: %v\n", err)
			}
		}

		// 按数据分类创建文件
		for Name, Rule := range self.GetRules() {
			// 跳过不输出的数据
			if len(Rule.GetOutFeild()) == 0 {
				continue
			}

			file, err := os.Create(filenameBase + " (" + Name + ").csv")

			if err != nil {
				Log.Println(err)
				continue
			}

			file.WriteString("\xEF\xBB\xBF") // 写入UTF-8 BOM
			w := csv.NewWriter(file)
			th := Rule.GetOutFeild()
			th = append(th, []string{"当前链接", "上级链接", "下载时间"}...)
			w.Write(th)

			num := 0 //小计
			for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
				if datacell["RuleName"].(string) == Name {
					row := []string{}
					for _, title := range Rule.GetOutFeild() {
						vd := datacell["Data"].(map[string]interface{})
						if v, ok := vd[title].(string); ok || vd[title] == nil {
							row = append(row, v)
						} else {
							row = append(row, util.JsonString(vd[title]))
						}
					}

					row = append(row, datacell["Url"].(string))
					row = append(row, datacell["ParentUrl"].(string))
					row = append(row, datacell["DownloadTime"].(string))
					w.Write(row)

					num++
				}
			}
			// 发送缓存数据流
			w.Flush()
			// 关闭文件
			file.Close()
			// 输出报告
			// Log.Printf("[任务：%v | 关键词：%v | 小类：%v] 输出 %v 条数据！！！\n", self.Spider.GetName(), self.Spider.GetKeyword(), Name, num)
		}
	}

	/************************ MongoDB 输出 ***************************/

	Output["mgo"] = func(self *Collector, dataIndex int) {
		session, err := mgo.Dial(config.DB_URL) //连接数据库
		if err != nil {
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)

		db := session.DB(config.DB_NAME)         //数据库名称
		collection := db.C(config.DB_COLLECTION) //如果该集合已经存在的话，则直接返回

		for i, count := 0, len(self.DockerQueue.Dockers[dataIndex]); i < count; i++ {
			err = collection.Insert((interface{})(self.DockerQueue.Dockers[dataIndex][i]))
			if err != nil {
				panic(err)
			}
		}
	}

	/************************ HBase 输出 ***************************/
	var master = cache.Task.Master
	var port = ":" + fmt.Sprintf("%v", cache.Task.Port)
	var hbaseSocket = teleport.New().SetPackHeader("tentinet")
	var once sync.Once

	Output["hbase"] = func(self *Collector, dataIndex int) {
		once.Do(func() { hbaseSocket.Client(master, port) })
		for i, count := 0, len(self.DockerQueue.Dockers[dataIndex]); i < count; i++ {
			hbaseSocket.Request(self.DockerQueue.Dockers[dataIndex][i], "log")
		}
	}

	/************************ Mysql 输出 ***************************/

	Output["mysql"] = func(self *Collector, dataIndex int) {
		db, err := sql.Open("mysql", config.MYSQL_USER+":"+config.MYSQL_PW+"@tcp("+config.MYSQL_HOST+")/"+config.MYSQL_DB+"?charset=utf8")
		if err != nil {
			fmt.Println(err)
		}
		defer db.Close()

		var newMysql myTable
		newMysql.setTableName("`" + self.Spider.GetName() + "-" + self.Spider.GetKeyword() + `-` + self.startTime + `-(` + fmt.Sprintf("%v", self.sum[0]) + `-` + fmt.Sprintf("%v", self.sum[1]) + ")`")

		for Name, Rule := range self.GetRules() {
			//跳过不输出的数据
			if len(Rule.GetOutFeild()) == 0 {
				continue
			}

			var tempTitle []string
			for _, title := range Rule.GetOutFeild() {
				tempTitle = append(tempTitle, title)
			}
			tempTitle = append(tempTitle, "当前连接", "上级链接", "下载时间")
			newMysql.setColumnName(tempTitle)
			newMysql.createColumn()
			stmt, err := db.Prepare(newMysql.sqlString)
			util.CheckErr(err)
			_, err = stmt.Exec()
			util.CheckErr(err)

			num := 0 //小计

			for _, datacell := range self.DockerQueue.Dockers[dataIndex] {
				if datacell["RuleName"].(string) == Name {
					var tempRow []string
					for _, title := range Rule.GetOutFeild() {
						vd := datacell["Data"].(map[string]interface{})

						if v, ok := vd[title].(string); ok || vd[title] == nil {
							tempRow = append(tempRow, v)
						} else {
							j, _ := json.Marshal(vd[title])
							tempRow = append(tempRow, string(j))
						}
					}
					tempRow = append(tempRow, datacell["Url"].(string))
					tempRow = append(tempRow, datacell["ParentUrl"].(string))
					tempRow = append(tempRow, datacell["DownloadTime"].(string))

					// 输出
					newMysql.setValue(tempRow)
					newMysql.insertColumn()
					stmt, err = db.Prepare(newMysql.sqlString)
					util.CheckErr(err)
					_, err = stmt.Exec()
					util.CheckErr(err)

					num++
				}
			}
		}
	}
}

/************************ Only For Mysql 输出 ***************************/

//sql转换结构体
type myTable struct {
	myTablename  string
	myColumnname []string
	value        []string
	sqlString    string
}

//设置msyql表名
func (self *myTable) setTableName(tableName string) {
	self.myTablename = tableName
}

//设置字段名，字段为切片
func (self *myTable) setColumnName(columnName []string) {
	self.myColumnname = columnName
}

//设置字段值
func (self *myTable) setValue(value []string) {
	self.value = value
}

//创建表字段
func (self *myTable) createColumn() {
	if self.myTablename != "" {
		self.sqlString = `create table ` + self.myTablename + `(`
		self.sqlString += ` id int(8) not null primary key auto_increment`
		if self.myColumnname != nil {
			for _, value := range self.myColumnname {
				self.sqlString += `,` + value + ` varchar(255) not null`
			}
		}
		self.sqlString += `);`
	}
}

//插入字段
//insert into table1(field1,field2) values(value1,value2)
func (self *myTable) insertColumn() {
	if self.myTablename != "" {
		self.sqlString = `insert into ` + self.myTablename + `(`
		if self.myColumnname != nil {
			for _, v1 := range self.myColumnname {
				self.sqlString += "`" + v1 + "`" + `,`
			}
			tem_string := self.sqlString[:len(self.sqlString)-1]
			self.sqlString = string(tem_string)
			self.sqlString += `)values(`
		}
		if self.value != nil {
			for _, v2 := range self.value {
				v2 = strings.Replace(v2, `"`, `\"`, -1)
				self.sqlString += `"` + v2 + `"` + `,`
			}
			tem_string := self.sqlString[:len(self.sqlString)-1]
			self.sqlString = string(tem_string)
			self.sqlString += `);`
		}
	}
}
