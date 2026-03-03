package rules

// base packages
import (
	"github.com/andeya/pholcus/app/downloader/request" // required
	"github.com/andeya/pholcus/common/goquery"         // DOM parsing

	//"github.com/andeya/pholcus/logs"               // logging
	spider "github.com/andeya/pholcus/app/spider" // required
	// . "github.com/andeya/pholcus/app/spider/common"          // optional
	//"github.com/andeya/pholcus/logs/logs"
	// string processing packages
	// "regexp"
	//"strconv"
	//"strings"
	// other packages
	// "fmt"
	// "math"
	// "time"
	//"strings"
	//"strings"
	"strconv"
	"strings"

	"github.com/andeya/pholcus/logs"
)

func init() {
	fangList.Register()
}

var fangList = &spider.Spider{
	Name:         "resell house of fang.com",
	Description:  "fang.com http://esf.zz.fang.com/house/i31/",
	EnableCookie: true,
	RuleTree: &spider.RuleTree{
		Root: func(ctx *spider.Context) {
			var i = 1
			//for i = 1; i < 101; i++ {
			ctx.AddQueue(&request.Request{
				URL:  "http://esf.zz.fang.com/house/i3" + strconv.Itoa(i) + "/",
				Rule: "fang_collection",
				Temp: map[string]interface{}{"p": 1},
			})
			//}
		},

		Trunk: map[string]*spider.Rule{
			"fang_collection": {
				ItemFields: []string{
					"communityName",
					"totalFloor",
					"rooms",
					"halls",
					"buildTime",
					"address",
					"direction",
					"area",
					"price",
					"unitPrice",
					"locationType", // floor level (high/low)
					"remoteId",     // fang.com id
					"business",
				},
				ParseFunc: func(ctx *spider.Context) {
					// get all fang.com data on current page
					ctx.GetDom().Find(".houseList dl").Each(
						func(i int, s *goquery.Selection) {
							var communityName, totalFloor, rooms, halls, locationType, remoteId, buildTime, address, direction, area, price, unitPrice, business string
							communityName = s.Find(".info p.mt10 a span").Text()

							address = s.Find(".info p.mt10 span.iconAdress").Text()
							business = ""

							sp := strings.Split(address, "-")
							if len(sp) == 2 {
								address = sp[1]
								business = sp[0]
							}
							// get year from room line
							roomLineTmp := s.Find("dd.info p.mt12").Text()
							roomLine := strings.Fields(roomLineTmp)

							if len(roomLine) == 4 {
								// remove "厅" (hall)
								roomsTmp := roomLine[0]
								roomsTmp = strings.Replace(roomsTmp, "厅", "", 1)
								roomsS := strings.Split(roomsTmp, "室")
								if len(roomsS) == 2 {
									rooms = roomsS[0]
									halls = roomsS[1]
								}
								// get building type and floor count
								buildingTmp := roomLine[1]
								buildingTmpSec := strings.Split(buildingTmp, "(共")
								if len(buildingTmpSec) == 2 {
									locationType = strings.Replace(buildingTmpSec[0], "|", "", 1)
									totalFloor = strings.Replace(buildingTmpSec[1], "层)", "", 1)
								}

								buildTime = strings.Replace(roomLine[3], "|建筑年代：", "", 1)
								direction = strings.Replace(roomLine[2], "|", "", 1)
								direction = strings.Replace(direction, "向", "", 1)
							}

							area = s.Find("dd.info div.area").Children().Eq(0).Text()
							price = s.Find("dd.info div.moreInfo").Children().Eq(0).Text()
							unitPrice = s.Find("dd.info div.moreInfo").Children().Eq(1).Text()
							remoteTmp := s.Find("dd.info p.title a").Attr("href")
							if remoteTmp.IsSome() {
								remoteAttr := strings.Split(remoteTmp.Unwrap(), "_")
								remoteId = strings.Replace(remoteAttr[1], ".htm", "", 1)
							}

							logs.Log().Critical("当前房源id: %v", remoteId)
							// parse passed fragment
							// store results in Response
							ctx.Output(map[int]interface{}{
								0:  strings.Trim(communityName, " "),
								1:  strings.Trim(totalFloor, " "),
								2:  strings.Trim(rooms, " "),
								3:  strings.Trim(halls, " "),
								4:  strings.Trim(buildTime, " "),
								5:  strings.Trim(address, " "),
								6:  strings.Trim(direction, " "),
								7:  strings.Trim(strings.Replace(area, "㎡", "", 1), " "),
								8:  strings.Trim(strings.Replace(price, "万", "", 1), " "),
								9:  strings.Trim(strings.Replace(unitPrice, "元/㎡", "", 1), " "),
								10: strings.Trim(locationType, " "),
								11: strings.Trim(remoteId, " "),
								12: strings.Trim(business, " "),
							})
						})
				},
			},
		},
	},
}
