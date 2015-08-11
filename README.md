# pholcus    [![GoDoc](https://godoc.org/github.com/tsuna/gohbase?status.png)](https://godoc.org/github.com/henrylee2cn/pholcus)

Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能。

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/icon.png)

* 稳定版： [Version 0.6.0 (Aug 10, 2015)](https://github.com/henrylee2cn/pholcus/releases).   [此处进入](https://github.com/henrylee2cn/pholcus/tree/master)

* 官方QQ群：Go大数据 42731170    [![Go大数据群](http://pub.idqqimg.com/wpa/images/group.png)](http://shang.qq.com/wpa/qunwpa?idkey=83ee3e1a4be6bdb2b08a51a044c06ae52cf10a082f7c5cf6b36c1f78e8b03589)

#### 框架模块

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/project.png)


#### 框架特点
 1. Pholcus（幽灵蛛）以高效率，高灵活性和人性化设计为开发的指导思想；

 2. 支持单机、服务端、客户端三种运行模式，即支持分布式布局，适用于各种业务需要；
 
 3. 支持Web、GUI、命令行三种操作界面，适用于各种运行环境；
 
 4. 支持mysql/mongodb/csv/excel等多种输出方式，且可以轻松添加更多输出方式；
 
 5. 采用surfer高并发下载器，支持 GET/POST/HEAD 方法及 http/https 协议，同时支持固定UserAgent自动保存cookie与随机大量UserAgent禁用cookie两种模式，高度模拟浏览器行为，可实现模拟登录等功能；

 6. 服务器/客户端模式采用teleport高并发socketAPI框架，全双工长连接通信，内部数据传输格式为JSON；
 
 7. 对采集规则进行了精心设计，规则灵活简单、高度封装，用于通用方法集与大量Demo，让你轻松添加规则；
 
 8. 支持横纵向两种抓取模式，并且支持任务暂停、取消等操作。


#### 下载安装

1. 下载需要翻墙的依赖包（[点击此处](https://raw.githubusercontent.com/henrylee2cn/pholcus/master/doc/%E9%9C%80%E8%A6%81%E7%BF%BB%E5%A2%99%E7%9A%84%E4%BE%9D%E8%B5%96%E5%8C%85%E5%9C%A8%E8%BF%99%E9%87%8C-%E8%A7%A3%E5%8E%8B%E8%87%B3gopath.rar)），并将其解压至 GOPATH/src 目录；

2. 下载剩余全部源码，命令行如下
```
go get github.com/henrylee2cn/pholcus
```

 > *<font size=2>注意：go get执行完成后，提示出现多个main函数的错误是正常的，这是由于支持下面的多种编译方式所致。</font>*



#### Web编译运行
```
go install pholcus-web.go
```
或者
```
go build pholcus-web.go
```

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/webshow_1.jpg)


#### GUI编译运行
```
go install -ldflags="-H windowsgui" pholcus-gui.go
```
或者
```
go build -ldflags="-H windowsgui" pholcus-gui.go
```

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/guishow_0.jpg)



#### 命令行编译运行
```
编译命令: go install pholcus-cmd.go  或者  go build pholcus-cmd.go
查看命令参数: pholcus-cmd.exe -h
执行爬虫命令: pholcus-cmd.exe -spider=3,8 -output=csv -go=500 -docker=5000 -pase=1000,3000 -kw=pholcus,golang -page=100
(注：花括号“{}”中为选择参数或参数格式，多个参数值之间用逗号“,”间隔，各项参数根据采集规则的需要自行设置)
```

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/cmd.jpg)




#### 添加ICON

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/addicon.jpg)



#### 添加规则

 - 添加一条规则的方法：只需在“henrylee2cn/pholcus/spider/spiders/”中增加一个采集规则（go文件），框架将自动添加该规则到GUI任务列表！



#### 第三方依赖包

```
go get "github.com/henrylee2cn/surfer"
go get "github.com/henrylee2cn/teleport"
go get "github.com/PuerkitoBio/goquery"
go get "github.com/bitly/go-simplejson"
go get "github.com/henrylee2cn/mahonia"
go get "github.com/andybalholm/cascadia"
go get "github.com/lxn/walk"
go get "github.com/lxn/win"
go get "github.com/tealeg/xlsx"
go get "github.com/go-sql-driver/mysql"
go get "gopkg.in/mgo.v2"
<以下需翻墙下载>
go get "golang.org/x/net/html"
go get "golang.org/x/text/encoding"
go get "golang.org/x/text/transform"
```
> *<font size="2">（在此感谢以上开源项目的支持！）</font>*




#### 开源协议

Pholcus（幽灵蛛）项目采用商业应用友好的[Apache License v2](https://github.com/henrylee2cn/pholcus/blob/master/doc/license.txt).发布
