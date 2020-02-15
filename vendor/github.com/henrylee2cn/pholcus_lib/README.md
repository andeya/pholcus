# pholcus_lib

[Pholcus](https://github.com/henrylee2cn/pholcus) 用户公共维护的spider爬虫规则库。

## 维护规范

- 欢迎每位用户都来分享自己的爬虫规则
- 每个规则放在单一个独的子目录
- 新增规则最好提供README.md
- 新增规则时，须在根目录 `pholcus_lib.go` 文件的import组中添加类似`_ "github.com/henrylee2cn/pholcus_lib/jingdong"`的包引用声明
- 新增规则时，须在根目录README.md（本文档）的 `爬虫规则列表` 中按子目录名`a-z`的顺序插入一条相应的规则记录
- 维护旧规则时，应在规则文件或相应README.md中增加修改说明：如修改原因、修改时间、签名、联系方式等
- 凡爬虫规则的贡献者均可在其源码文件或相应README.md中留下在的签名、联系方式


## 爬虫规则列表

|子目录|规则描述|
|---|---|
|alibaba|阿里巴巴产品搜索|
|avatar|我要个性网-头像昵称搜索下载|
|baidunews|百度RSS新闻|
|baidusearch|百度搜索|
|car_home|汽车之家|
|chinanews|中国新闻网-滚动新闻|
|filetest|文件下载测试|
|ganji_gongsi|经典示例-赶集网企业名录|
|googlesearch|谷歌搜索|
|hollandandbarrett|Hollandand&Barrett商品数据|
|IJGUC|IJGUC期刊|
|jdsearch|京东搜索|
|jingdong|京东搜索(修复版)|
|jiban|羁绊动漫|
|kaola|考拉海淘|
|lewa|乐蛙登录测试|
|miyabaobei|蜜芽宝贝|
|people|人民网新闻抓取|
|shunfenghaitao|顺丰海淘|
|taobao|淘宝数据|
|taobaosearch|淘宝天猫搜索|
|wangyi|网易新闻|
|weibo_fans|微博粉丝列表|
|wukongwenda|悟空问答|
|zolpc|中关村笔记本|
|zolphone|中关村手机|
|zolslab|中关村平板|
|zhihu_bianji|知乎编辑推荐|
|zhihu_daily|知乎每日推荐|
