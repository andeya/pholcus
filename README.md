## pholcus

Pholcus（幽灵蛛）是一款纯Go语言编写的重量级爬虫软件，清新的GUI界面，优雅的爬虫规则、可控的高并发、任意的批量任务、多种输出方式、大量Demo，更重要的是它支持socket长连接、全双工并发分布式，支持横纵向两种抓取模式，支持模拟登录和任务取消等。

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/icon.png)

* 稳定版： [Version 0.5.1 (July 24, 2015)](https://github.com/henrylee2cn/pholcus/releases).   [此处进入](https://github.com/henrylee2cn/pholcus/tree/master)

* 官方QQ群：Go大数据 42731170    [![Go大数据群](http://pub.idqqimg.com/wpa/images/group.png)](http://shang.qq.com/wpa/qunwpa?idkey=83ee3e1a4be6bdb2b08a51a044c06ae52cf10a082f7c5cf6b36c1f78e8b03589)

#### 框架模块

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/project.png)


#### 框架特点
 1. Pholcus（幽灵蛛）以高效率，高灵活性和人性化设计为开发的指导思想；

 2. 继承Go语言“少即是多”的风格，GUI界面尽量少得呈现技术层面的参数配置，而在程序内部做智能化参数调控；
 
 3. 对采集规则进行了精心设计，结构化规则、高度封装、通用方法集、自由灵活的发挥空间，让你轻松添加规则；
 
 4. 每个pholcus程序既可以是服务器也可以是客户端，通过socket传递request来实现任务分发，其中node模块充当管理核心的角色，负责分发给其他节点和本地队列请求以及实时log，比如，让Pholcus软件同时在10台电脑运行，你就拥有了10个节点，自然形成分布式。
 
 5. 支持横纵向两种抓取模式，并支持任务取消操作。


#### 下载安装

1. [这里下载需要翻墙的依赖包](https://raw.githubusercontent.com/henrylee2cn/pholcus/master/doc/%E9%9C%80%E8%A6%81%E7%BF%BB%E5%A2%99%E7%9A%84%E4%BE%9D%E8%B5%96%E5%8C%85%E5%9C%A8%E8%BF%99%E9%87%8C-%E8%A7%A3%E5%8E%8B%E8%87%B3gopath.rar)
2. 下载其他全部源码，命令行如下
```
go get github.com/henrylee2cn/pholcus
```



#### GUI编译运行
```
go install -ldflags="-H windowsgui" pholcus-gui.go
```
或者
```
go build -ldflags="-H windowsgui" pholcus-gui.go
```



#### 命令行编译运行
```
编译命令: go install pholcus-cmd.go  或者  go build pholcus-cmd.go
查看命令参数: pholcus-cmd.exe -h
执行爬虫命令: pholcus-cmd.exe -spider=3,8 -output=csv -go=500 -docker=5000 -pase=1000,3000 -kw=pholcus,golang -page=100
(注：花括号“{}”中为选择参数或参数格式，多个参数值之间用逗号“,”间隔，各项参数根据采集规则的需要自行设置)
```


#### GUI界面
随时改进中，该截图仅供参考！

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/guishow_0.jpg)
![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/guishow_1.jpg)

#### 命令行界面
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
（在此对以上依赖包的开源项目表示感谢！）



#### 开源协议

Pholcus（幽灵蛛）项目采用商业应用友好的[Apache License v2](https://github.com/henrylee2cn/pholcus/blob/master/doc/license.txt).发布
