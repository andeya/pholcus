## pholcus
Pholcus（幽灵蛛）是一款Go语言编写的爬虫软件框架（含GUI界面），优雅的爬虫规则、可控的高并发、任意的批量任务、多种输出方式、大量Demo，并且考虑了支持分布式布局。（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）


**架构模块**

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/project.png)



**GUI界面**
随时改进中，该截图仅供参考！

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/guishow.jpg)


**安装Pholcus（幽灵蛛）**
```
go get github.com/henrylee2cn/pholcus
```



**编译Pholcus（幽灵蛛）**
```
go install -ldflags="-H windowsgui"
```
或者
```
go build -ldflags="-H windowsgui"
```



**Pholcus（幽灵蛛）加ICON**

![image](https://github.com/henrylee2cn/pholcus/blob/master/doc/addicon.jpg)



**Pholcus（幽灵蛛）添加规则**

目前添加规则的方法：在“henrylee2cn/pholcus/spiders/”中添加一个采集规则（go文件），然后在“henrylee2cn/pholcus/pholcus/gui/menu.go”中登记该规则条目。
未来计划支持动态加载外部规则的模式。。。
