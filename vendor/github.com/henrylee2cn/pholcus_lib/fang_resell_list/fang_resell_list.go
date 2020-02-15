package pholcus_lib

// 基础包
import (
	"github.com/henrylee2cn/pholcus/app/downloader/request" //必需
	"github.com/henrylee2cn/pholcus/common/goquery"         //DOM解析
	//"github.com/henrylee2cn/pholcus/logs"               //信息输出
	. "github.com/henrylee2cn/pholcus/app/spider" //必需
	// . "github.com/henrylee2cn/pholcus/app/spider/common"          //选用
	//"github.com/henrylee2cn/pholcus/logs/logs"
	// 字符串处理包
	// "regexp"
	//"strconv"
	//"strings"
	// 其他包
	// "fmt"
	// "math"
	// "time"
	//"strings"
	//"strings"
	"strings"
	"github.com/henrylee2cn/pholcus/logs"
	"strconv"
)

func init() {
	fangList.Register()
}

var fangList = &Spider{
	Name:         "resell house of fang.com",
	Description:  "fang.com http://esf.zz.fang.com/house/i31/",
	EnableCookie: true,
	RuleTree: &RuleTree{
		Root: func(ctx *Context) {
			var i = 1;
			//for i = 1; i < 101; i++ {
			ctx.AddQueue(&request.Request{
				Url:  "http://esf.zz.fang.com/house/i3" + strconv.Itoa(i) + "/",
				Rule: "fang_collection",
				Temp: map[string]interface{}{"p": 1},
			})
			//}
		},

		Trunk: map[string]*Rule{
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
					"locationType", //楼层所在高低
					"remoteId",     //搜房id
					"business",
				},
				ParseFunc: func(ctx *Context) {
					//获取当页搜房的所有数据
					ctx.GetDom().Find(".houseList dl").Each(
						func(i int, s *goquery.Selection) {
							var communityName, totalFloor, rooms, halls, locationType, remoteId, buildTime, address, direction, area, price, unitPrice, business string;
							communityName = s.Find(".info p.mt10 a span").Text();

							address = s.Find(".info p.mt10 span.iconAdress").Text();
							business = "";

							sp := strings.Split(address,"-");
							if(len(sp) == 2){
								address = sp[1];
								business = sp[0];
							}
							//获取年代中的一吨
							roomLineTmp := s.Find("dd.info p.mt12").Text();
							roomLine := strings.Fields(roomLineTmp);

							if (len(roomLine) == 4 ) {
								//替换掉厅
								roomsTmp := roomLine[0];
								roomsTmp = strings.Replace(roomsTmp, "厅", "", 1);
								roomsS := strings.Split(roomsTmp, "室");
								if (len(roomsS) == 2) {
									rooms = roomsS[0];
									halls = roomsS[1];
								}
								//楼类型和层高获取
								buildingTmp := roomLine[1];
								buildingTmpSec := strings.Split(buildingTmp, "(共");
								if (len(buildingTmpSec) == 2) {
									locationType = strings.Replace(buildingTmpSec[0], "|", "", 1);
									totalFloor = strings.Replace(buildingTmpSec[1], "层)", "", 1);
								}

								buildTime = strings.Replace(roomLine[3], "|建筑年代：", "", 1);
								direction = strings.Replace(roomLine[2], "|", "", 1);
								direction = strings.Replace(direction, "向", "", 1);
							}

							area = s.Find("dd.info div.area").Children().Eq(0).Text();
							price = s.Find("dd.info div.moreInfo").Children().Eq(0).Text();
							unitPrice = s.Find("dd.info div.moreInfo").Children().Eq(1).Text();
							remoteTmp, exists := s.Find("dd.info p.title a").Attr("href");
							if (exists) {
								remoteAttr := strings.Split(remoteTmp,"_");
								remoteId = strings.Replace(remoteAttr[1],".htm","",1);
							}

							logs.Log.Critical("当前房源id: %v", remoteId)
							//解析传入的片段
							// 结果存入Response中转
							ctx.Output(map[int]interface{}{
								0:  strings.Trim(communityName, " "),
								1:  strings.Trim(totalFloor, " "),
								2:  strings.Trim(rooms, " "),
								3:  strings.Trim(halls, " "),
								4:  strings.Trim(buildTime, " "),
								5:  strings.Trim(address, " "),
								6:  strings.Trim(direction, " "),
								7:  strings.Trim(strings.Replace(area,"㎡","",1), " "),
								8:  strings.Trim(strings.Replace(price,"万","",1), " "),
								9:  strings.Trim(strings.Replace(unitPrice,"元/㎡","",1), " "),
								10: strings.Trim(locationType, " "),
								11: strings.Trim(remoteId, " "),
								12: strings.Trim(business, " "),
							})
						})
					ctx.Parse("getContent")
				},
			},
		},
	},
}
