## pholcus

Pholcus（幽灵蛛）是一款纯Go语言编写的重量级爬虫软件，清新的GUI界面，优雅的爬虫规则、可控的高并发、任意的批量任务、多种输出方式、大量Demo，支持横纵向两种抓取模式，支持模拟登录和任务取消等，并且考虑了支持分布式布局。 （官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）


#### 框架模块

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/project.png)


#### 框架特点
 1. Pholcus（幽灵蛛）以高效率，高灵活性和人性化设计为开发的指导思想；

 2. 继承Go语言“少即是多”的风格，GUI界面尽量少得呈现技术层面的参数配置，而在程序内部做智能化参数调控；

 3. 对采集规则进行了精心设计，结构化规则、高度封装、通用方法集、自由灵活的发挥空间，让你轻松添加规则；

 4. 每个pholcus程序既可以是服务器也可以是客户端，通过socket传递request来实现任务分发，其中hpolcus模块充当管理核心的角色，负责分发给其他节点和本地队列请求以及实时log，(比如，让Pholcus软件同时在10台电脑运行，你就拥有了10个节点，自然形成分布式)。目前该功能的架构已经初步完成，接口即将实现，敬请关注。

#### GUI界面
随时改进中，该截图仅供参考！

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/guishow.jpg)


#### 初次安装
```
go get github.com/henrylee2cn/pholcus
```



#### 编译运行
```
go install -ldflags="-H windowsgui"
```
或者
```
go build -ldflags="-H windowsgui"
```



#### 添加ICON

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/addicon.jpg)



#### 添加规则

 - 目前添加规则的方法：在“henrylee2cn/pholcus/spiders/”中添加一个采集规则（go文件），然后在“henrylee2cn/pholcus/pholcus/gui/menu.go”中登记该规则条目；

 - 未来计划支持动态加载外部规则的模式……


#### 第三方依赖包


```
go get "github.com/lxn/walk"
go get "github.com/lxn/win"
go get "github.com/PuerkitoBio/goquery"
go get "github.com/bitly/go-simplejson"
go get "github.com/tealeg/xlsx"
go get "gopkg.in/mgo.v2"
go get "code.google.com/p/mahonia"
go get "golang.org/x/net/html/charset"
```
