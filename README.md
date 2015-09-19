# pholcus    [![GoDoc](https://godoc.org/github.com/tsuna/gohbase?status.png)](https://godoc.org/github.com/henrylee2cn/pholcus)

Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能。

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/icon.png)

* 稳定版： [Version 0.7.0 (Sep 20, 2015)](https://github.com/henrylee2cn/pholcus/releases).   [此处进入](https://github.com/henrylee2cn/pholcus/tree/master)

* 官方QQ群：Go大数据 42731170    [![Go大数据群](http://pub.idqqimg.com/wpa/images/group.png)](http://shang.qq.com/wpa/qunwpa?idkey=83ee3e1a4be6bdb2b08a51a044c06ae52cf10a082f7c5cf6b36c1f78e8b03589)

#### 爬虫原理

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

&nbsp;

#### 下载安装

1. 下载第三方依赖包源码，放至 GOPATH/src 目录下 [[点击下载 ZIP]](https://github.com/pholcus/dependent/archive/master.zip)

2. 下载保持更新状态的源码，命令行如下
```
go get github.com/henrylee2cn/pholcus
```

备注：Pholcus公开维护的spider规则库地址 <https://github.com/pholcus/spider_lib>

&nbsp;

#### 创建项目

```
package main

import (
    // 按界面需求选择相应版本
    "github.com/henrylee2cn/pholcus/web" // web版
    // "github.com/henrylee2cn/pholcus/cmd" // cmd版
    // "github.com/henrylee2cn/pholcus/gui" // gui版

    "github.com/henrylee2cn/pholcus/config"
    "github.com/henrylee2cn/pholcus/logs"
)

// 导入自己的规则库（须保证最后声明，即最先导入）
import (
    _ "github.com/pholcus/spider_lib" // 此为公开维护的spider规则库
    // _ "path/myrule_lib" // 同样你也可以自由添加自己的规则库
)

// 自定义相关配置，将覆盖默认值
func setConf() {
    //mongodb服务器地址
    config.MGO_OUTPUT.Host = "127.0.0.1:27017"
    // mongodb输出时的内容分类
    // key:蜘蛛规则清单
    // value:数据库名
    config.MGO_OUTPUT.DBClass = map[string]string{
        "百度RSS新闻": "1_1",
    }
    // mongodb输出时非默认数据库时以当前时间为集合名
    // h: 精确到小时 (格式 2015-08-28-09)
    // d: 精确到天 (格式 2015-08-28)
    config.MGO_OUTPUT.TableFmt = "d"

    //mysql服务器地址
    config.MYSQL_OUTPUT.Host = "127.0.0.1:3306"
    //msyql数据库
    config.MYSQL_OUTPUT.DefaultDB = "pholcus"
    //mysql用户
    config.MYSQL_OUTPUT.User = "root"
    //mysql密码
    config.MYSQL_OUTPUT.Password = ""
}

func main() {
    // 开启错误日志调试功能（打印行号及Debug信息）
    logs.Debug(true)

    defer func() {
        if err := recover(); err != nil {
            logs.Log.Emergency("%v", err)
        }
    }()

    setConf() // 不调用则为默认值

    // 开始运行
    web.Run() // web版
    // cmd.Run() // cmd版
    // gui.Run() // gui版
}

```
&nbsp;

#### Web版编译运行
```
go install (可选参数： -ip 0.0.0.0 -port 9090)
```
或者
```
go build (可选参数： -ip 0.0.0.0 -port 9090)
```
> *<font size="2">(注意：将 src/github.com/henrylee2cn/pholcus/web 文件夹拷贝至当前项目目录下，其中的go文件可删除)*
![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/webshow_1.jpg)

&nbsp;

#### GUI版编译运行

<span> 1. 编译</span>
```
go install -ldflags="-H windowsgui"
```
或者
```
go build -ldflags="-H windowsgui"
```

<span> 2. 添加ICON</span>

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/addicon.jpg)

> *<font size="2">(下图为GUI选择模式界面图例)*

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/guishow_0.jpg)

&nbsp;


#### Cmd版编译运行
```
编译命令: go install pholcus-cmd.go  或者  go build pholcus-cmd.go
查看命令参数: pholcus-cmd.exe -h
执行爬虫命令: pholcus-cmd.exe -spider=1,3 -output=csv -go=500 -docker=5000 -pase=1000,3000 -kw=pholcus,golang -page=100
```

> *<font size="2">(注：花括号“{}”中为选择参数或参数格式，多个参数值之间用逗号“,”间隔，各项参数根据采集规则的需要自行设置)*
![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/cmd.jpg)

&nbsp;

#### 第三方依赖包

```
go get github.com/pholcus/spider_lib
go get github.com/henrylee2cn/surfer
go get github.com/henrylee2cn/teleport
go get github.com/henrylee2cn/beelogs
go get github.com/henrylee2cn/mahonia
go get github.com/henrylee2cn/websocket.google
go get github.com/PuerkitoBio/goquery
go get github.com/andybalholm/cascadia
go get github.com/lxn/walk
go get github.com/lxn/win
go get github.com/tealeg/xlsx
go get github.com/go-sql-driver/mysql
go get gopkg.in/mgo.v2
<以下需翻墙下载>
go get golang.org/x/net/html
go get golang.org/x/text/encoding
go get golang.org/x/text/transform
```
> *<font size="2">（在此感谢以上开源项目的支持！）</font>*


&nbsp;

#### 开源协议

Pholcus（幽灵蛛）项目采用商业应用友好的[Apache License v2](https://github.com/henrylee2cn/pholcus/blob/master/doc/license.txt).发布
