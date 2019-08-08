## 搜房爬取二手房列表

### 说明

	仅爬取列表页, 字段: 
	"communityName":小区名,
	"totalFloor":总层数,
	"rooms":房间数,
	"halls":厅数量,
	"buildTime":建筑年代,
	"address":地址,
	"direction":朝向,
	"area":面积,
	"price":价格,
	"unitPrice"单价,
	"locationType"所在层数高低,

### 代码说明

	1.目前仅仅爬取了搜房二手房的列表页, 一次爬取一页
	2.如果有需要就修改37行打开多页爬取
	3.在使用中发现,如果爬取的页面数太多会导致蜘蛛崩溃, 原因未知, 待查