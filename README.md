# pholcus    [![GoDoc](https://godoc.org/github.com/tsuna/gohbase?status.png)](https://godoc.org/github.com/henrylee2cn/pholcus)

Pholcus（幽灵蛛）是一款纯Go语言编写的高并发、分布式、重量级爬虫软件，支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/csv/excel等）、有大量Demo共享；同时她还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能。

![image](https://github.com/henrylee2cn/pholcus/raw/master/doc/icon.png)

* 稳定版： [Version 0.7.6 (Dec 17, 2015)](https://github.com/henrylee2cn/pholcus/releases)

* 官方QQ群：Go大数据 42731170    [![Go大数据群](http://pub.idqqimg.com/wpa/images/group.png)](http://shang.qq.com/wpa/qunwpa?idkey=83ee3e1a4be6bdb2b08a51a044c06ae52cf10a082f7c5cf6b36c1f78e8b03589)

#### 爬虫原理

![image](https://github.com/henrylee2cn/pholcus/raw/master/doc/project.png)


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
    "github.com/henrylee2cn/pholcus/config"
    "github.com/henrylee2cn/pholcus/exec"
    // "github.com/henrylee2cn/pholcus/logs"

    _ "github.com/pholcus/spider_lib"     // 此为公开维护的spider规则库
    // _ "spider_lib_pte" // 同样你也可以自由添加自己的规则库
)

func main() {
    // 设置运行时默认操作界面，并开始运行
    // 运行软件前，可设置 -a_ui 参数为"web"、"gui"或"cmd"，指定本次运行的操作界面
    // 其中"gui"仅支持Windows系统
    exec.DefaultRun("web")
}

// 自定义相关配置，将覆盖默认值
func init() {
    // 允许日志打印行号
    // logs.ShowLineNum()

    //mongodb链接字符串
    config.MGO.CONN_STR = "127.0.0.1:27017"
    //mongodb数据库
    config.MGO.DB = "pholcus"
    //mongodb连接池容量
    config.MGO.MAX_CONNS = 1024

    //mysql服务器地址
    config.MYSQL.CONN_STR = "root:@tcp(127.0.0.1:3306)"
    //msyql数据库
    config.MYSQL.DB = "pholcus"
    //mysql连接池容量
    config.MYSQL.MAX_CONNS = 1024

    // 历史记录文件名前缀
    config.HISTORY.FILE_NAME_PREFIX = "history"

    // 代理IP完整文件名
    config.PROXY_FULL_FILE_NAME = "proxy.pholcus"

    // Surfer-Phantom下载器配置
    config.SURFER_PHANTOM.FULL_APP_NAME = "phantomjs" //phantomjs软件相对路径与名称
}
```
&nbsp;

#### 编译运行
正常编译方法
```
go install 或者 go build
```
Windows下隐藏cmd窗口的编译方法
```
go install -ldflags="-H windowsgui" 或者 go build -ldflags="-H windowsgui"
```
查看可选参数: 
```
pholcus -h
```
![image](https://github.com/henrylee2cn/pholcus/raw/master/doc/help.jpg)

&nbsp;

> *<font size="2">(注意：当运行web操作界面时请将 src/github.com/henrylee2cn/pholcus/web 文件夹拷贝至当前项目目录下，其中的go文件可删除)，Web版操作界面截图如下：*

![image](https://github.com/henrylee2cn/pholcus/raw/master/doc/webshow_1.jpg)

&nbsp;

> *<font size="2">GUI版操作界面之模式选择界面截图如下*

![image](https://github.com/henrylee2cn/pholcus/raw/master/doc/guishow_0.jpg)

&nbsp;

> *<font size="2">Cmd版运行参数设置示例如下*

```
pholcus -a_ui=cmd -c_spider=3,8 -c_output=csv -c_thread=20 -c_docker=5000 -c_pause=300 -c_proxy=0 -c_keyword=pholcus,golang -c_maxpage=10 -c_inherit_y=true -c_inherit_n=true
```

&nbsp;

#### 第三方依赖包

```
go get github.com/pholcus/spider_lib
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

Pholcus（幽灵蛛）项目采用商业应用友好的[Apache License v2](https://github.com/henrylee2cn/pholcus/raw/master/doc/license.txt).发布
