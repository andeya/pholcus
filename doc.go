// Pholcus（幽灵蛛）是一款纯Go语言编写的支持分布式的高并发、重量级爬虫软件，定位于互联网数据采集，为具备一定Go或JS编程基础的人提供一个只需关注规则定制的功能强大的爬虫工具。
// 它支持单机、服务端、客户端三种运行模式，拥有Web、GUI、命令行三种操作界面；规则简单灵活、批量任务并发、输出方式丰富（mysql/mongodb/kafka/csv/excel等）、有大量Demo共享；另外它还支持横纵向两种抓取模式，支持模拟登录和任务暂停、取消等一系列高级功能。
// （官方QQ群：Go大数据 42731170）。
package main

/**
Version 1.2 升级日志

概述：占用资源更少，运行更稳定

一、输出功能升级

添加kafka数据库输出
基本重新了mysql输出模块，提升输出稳定性与输出效率
增加输出文件目录的配置项
大量优化结果收集模块，提升I/O性能与状态控制性
移除文件输出目录的日期后缀
调整文件名哈希临界长度为>8
移除数据输出通道容量的配置项DATA_CHAN_CAP，由分批输出用户设置项直接决定


二、下载功能升级

增强自动转码功能
当响应头未指定编码类型时，从请求头读取
都未指定编码类型或编码类型为utf8时，不做转码，节约内存
增加支持自动解压缩deflate和zlib编码的响应流
升级surfer下载器，修复POST提交时下载内核中Content-Type被覆盖的bug，修复Request.GetHeader()==nil时panic的bug
修复输出图片等文件时，下载补全的bug
Context.text字段类型由string改为[]byte
将HTTP状态码大于等于400的请求自动标记为下载失败


三、采集规则模块升级

更新*Request.GetTemp(key string, defaultValue interface{}) interface{}，defaultValue不再作为结果接收容器，当键值对不存在时，返回值为参数defaultValue。
Spider.Register()方法改为接受Spider类型（之前为*Spider），从而可以使用 "var XXXSpider=Spider{}.Register()" 的方式进行规则声明
优化任务停止条件，Spider.Root退出之前，任务不可终止
修复动态规则解析bug
同名采集规则的名称自动添加加"(2)"形式的序号后缀
优化crawler采集引擎的随机停顿逻辑
添加 Context.Log() 日志打印接口


四、其他优化

修复某些情况下在非win系统中log日志引发的panic
修复web版启动时偶然性打不开页面的bug
web版实时日志在超过2000条时自定清除前1000条
优化scheduler调度器
调整分布式模块字面量命名
修复CUP占用高的问题，采集过程的最低使用率从 20% 降低到 1%
加快任务的主动终止，基本已将延时控制在秒级
通过数据输出速率来抑制采集下载速率，从而降低不必要的内存占用
将依赖包全部移入vendor中，方便下载且利用程序稳定
*/
