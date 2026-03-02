<div align="center">
  <img src="https://github.com/andeya/pholcus/raw/master/doc/icon.png" width="120" alt="Pholcus Logo"/>
  <h1>Pholcus（幽灵蛛）</h1>
  <p><strong>纯 Go 语言编写的分布式高并发爬虫框架</strong></p>

[![GitHub release](https://img.shields.io/github/release/andeya/pholcus.svg?style=flat-square)](https://github.com/andeya/pholcus/releases)
[![GitHub stars](https://img.shields.io/github/stars/andeya/pholcus.svg?style=flat-square&label=Stars)](https://github.com/andeya/pholcus/stargazers)
[![Go Reference](https://pkg.go.dev/badge/github.com/andeya/pholcus.svg)](https://pkg.go.dev/github.com/andeya/pholcus)
[![Go Report Card](https://goreportcard.com/badge/github.com/andeya/pholcus?style=flat-square)](https://goreportcard.com/report/andeya/pholcus)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://github.com/andeya/pholcus/blob/master/LICENSE)
[![GitHub issues](https://img.shields.io/github/issues/andeya/pholcus.svg?style=flat-square)](https://github.com/andeya/pholcus/issues?q=is%3Aopen+is%3Aissue)
[![GitHub closed issues](https://img.shields.io/github/issues-closed-raw/andeya/pholcus.svg?style=flat-square)](https://github.com/andeya/pholcus/issues?q=is%3Aissue+is%3Aclosed)

<p>
  <a href="#快速开始">快速开始</a> •
  <a href="#核心特性">核心特性</a> •
  <a href="#架构设计">架构设计</a> •
  <a href="#操作界面">操作界面</a> •
  <a href="#规则编写">规则编写</a> •
  <a href="#常见问题">FAQ</a>
</p>

</div>

---

## 免责声明

> **本软件仅用于学术研究，使用者需遵守其所在地的相关法律法规，请勿用于非法用途！**
>
> 如在中国大陆频频爆出爬虫开发者涉诉与违规的 [新闻](https://github.com/HiddenStrawberry/Crawler_Illegal_Cases_In_China)。
>
> **郑重声明：因违法违规使用造成的一切后果，使用者自行承担！**

---

## 核心特性

<table>
<tr>
<td width="50%">

**运行模式**

- 单机模式 — 开箱即用
- 服务端模式 — 分发任务
- 客户端模式 — 接收并执行任务

</td>
<td width="50%">

**操作界面**

- Web UI — 跨平台，浏览器操作
- GUI — Windows 原生界面
- Cmd — 命令行批量调度

</td>
</tr>
<tr>
<td>

**数据输出**

- MySQL / MongoDB
- Kafka / Beanstalkd
- CSV / Excel
- 原文件下载

</td>
<td>

**爬虫规则**

- 静态规则（Go）— 高性能，深度定制
- 动态规则（JS/XML）— 热加载，无需编译
- 30+ 内置示例规则

</td>
</tr>
</table>

**更多亮点：**

- 高并发下载器 [surfer](app/downloader/surfer)，支持 GET / POST / HEAD，兼容 HTTP / HTTPS
- 智能 Cookie 管理：固定 UserAgent 自动保存 cookie，或随机 UserAgent 禁用 cookie
- 模拟登录、自定义 Header、POST 表单提交
- 代理 IP 池，可按频率自动更换
- 随机停歇机制，模拟人工行为
- 采集量与并发协程数可控
- 请求自动去重 + 失败请求自动重试
- 成功记录持久化，支持断点续爬
- 分布式通信全双工 Socket 框架

---

## 架构设计

<details>
<summary><b>模块结构</b></summary>
<br/>
<img src="https://github.com/andeya/pholcus/raw/master/doc/module.png" alt="模块结构" width="700"/>
</details>

<details>
<summary><b>项目架构</b></summary>
<br/>
<img src="https://github.com/andeya/pholcus/raw/master/doc/project.png" alt="项目架构" width="700"/>
</details>

<details>
<summary><b>分布式架构</b></summary>
<br/>
<img src="https://github.com/andeya/pholcus/raw/master/doc/distribute.png" alt="分布式架构" width="700"/>
</details>

### 目录结构

```
pholcus/
├── app/                    核心逻辑
│   ├── crawler/            爬虫引擎 & 并发池
│   ├── downloader/         下载器（surfer）
│   ├── pipeline/           数据管道 & 多种输出后端
│   ├── scheduler/          请求调度器
│   ├── spider/             爬虫规则引擎
│   ├── distribute/         分布式 Master/Slave 通信
│   └── aid/                辅助模块（历史记录、代理 IP）
├── config/                 配置管理
├── exec/                   启动入口 & 平台适配
├── cmd/                    命令行模式
├── gui/                    GUI 模式（Windows）
├── web/                    Web UI 模式
├── common/                 公共工具库（DB 驱动、编码、队列等）
├── logs/                   日志模块
├── runtime/                运行时缓存 & 状态
└── sample/                 示例程序 & 30+ 爬虫规则
```

---

## 快速开始

### 环境要求

- Go 1.18+（推荐 1.22+）

### 获取源码

```bash
git clone https://github.com/andeya/pholcus.git
cd pholcus
```

### 编写入口

创建 `main.go`（或参考 `sample/main.go`）：

```go
package main

import (
    "github.com/andeya/pholcus/exec"
    _ "github.com/andeya/pholcus/sample/static_rules"  // 内置规则库
    // _ "yourproject/rules"                            // 自定义规则库
)

func main() {
    // 启动界面：web / gui / cmd
    // 可通过 -a_ui 运行参数覆盖
    exec.DefaultRun("web")
}
```

### 编译运行

```bash
# 编译（非 Windows 平台自动排除 GUI 包）
go build -o pholcus ./sample/

# 查看所有可选参数
./pholcus -h
```

Windows 下隐藏 cmd 窗口的编译方式：

```bash
go build -ldflags="-H=windowsgui -linkmode=internal" -o pholcus.exe ./sample/
```

### 命令行参数一览

```bash
./pholcus -h
```

![命令行帮助](https://github.com/andeya/pholcus/raw/master/doc/help.jpg)

---

## 操作界面

### Web UI

启动后访问 `http://localhost:2015`，在浏览器中即可完成蜘蛛选择、参数配置、任务启停等全部操作。

![Web 界面](https://github.com/andeya/pholcus/raw/master/doc/webshow_1.png)

### GUI（仅 Windows）

原生桌面客户端，功能与 Web 版一致。

![GUI 界面](https://github.com/andeya/pholcus/raw/master/doc/guishow_0.jpg)

### Cmd 命令行

适用于服务器部署或 cron 定时任务场景。

```bash
pholcus -_ui=cmd -a_mode=0 -c_spider=3,8 -a_outtype=csv -a_thread=20 \
    -a_batchcap=5000 -a_pause=300 -a_proxyminute=0 \
    -a_keyins="<pholcus><golang>" -a_limit=10 -a_success=true -a_failure=true
```

---

## 规则编写

Pholcus 支持 **静态规则（Go）** 和 **动态规则（JS/XML）** 两种方式。

### 静态规则（Go）

随软件一同编译，性能最优，适合重量级采集项目。在 `sample/static_rules/` 下新建 Go 文件即可：

```go
package rules

import (
    "net/http"
    "github.com/andeya/pholcus/app/downloader/request"
    "github.com/andeya/pholcus/app/spider"
)

func init() {
    mySpider.Register()
}

var mySpider = &spider.Spider{
    Name:         "示例爬虫",
    Description:  "示例爬虫 [Auto Page] [http://example.com]",
    EnableCookie: true,
    RuleTree: &spider.RuleTree{
        Root: func(ctx *spider.Context) {
            ctx.AddQueue(&request.Request{
                URL:  "http://example.com",
                Rule: "首页",
            })
        },
        Trunk: map[string]*spider.Rule{
            "首页": {
                ParseFunc: func(ctx *spider.Context) {
                    ctx.Output(map[int]interface{}{
                        0: ctx.GetText(),
                    })
                },
            },
        },
    },
}
```

> 更多示例见 [`sample/static_rules/`](sample/static_rules/)，涵盖百度、京东、淘宝、知乎等 30+ 网站。

### 动态规则（JS/XML）

无需编译即可热加载，适合轻量级采集。将 `.pholcus.xml` 文件放入 `dyn_rules/` 目录：

```xml
<Spider>
    <Name>百度搜索</Name>
    <Description>百度搜索 [Auto Page] [http://www.baidu.com]</Description>
    <Pausetime>300</Pausetime>
    <EnableLimit>false</EnableLimit>
    <EnableCookie>true</EnableCookie>
    <EnableKeyin>true</EnableKeyin>
    <NotDefaultField>false</NotDefaultField>
    <Namespace><Script></Script></Namespace>
    <SubNamespace><Script></Script></SubNamespace>
    <Root>
        <Script param="ctx">
        ctx.JsAddQueue({
            URL: "http://www.baidu.com/s?wd=" + ctx.GetKeyin(),
            Rule: "搜索结果"
        });
        </Script>
    </Root>
    <Rule name="搜索结果">
        <ParseFunc>
            <Script param="ctx">
            ctx.Output({
                "标题": ctx.GetDom().Find("title").Text(),
                "内容": ctx.GetText()
            });
            </Script>
        </ParseFunc>
    </Rule>
</Spider>
```

> 同时兼容 `.pholcus.html` 旧格式。`<Script>` 标签内自动包裹 CDATA，无需手动转义特殊字符。

---

## 配置说明

### 运行时目录

```
├── pholcus                    可执行文件
├── dyn_rules/                 动态规则目录（可在 config.ini 中配置）
│   └── xxx.pholcus.xml        动态规则文件
└── pholcus_pkg/               运行时文件目录
    ├── config.ini             配置文件
    ├── proxy.lib              代理 IP 列表
    ├── phantomjs              PhantomJS 程序
    ├── text_out/              文本输出目录
    ├── file_out/              文件输出目录
    ├── logs/                  日志目录
    ├── history/               历史记录目录
    └── cache/                 临时缓存目录
```

### 代理 IP

在 `pholcus_pkg/proxy.lib` 文件中逐行写入代理地址：

```
http://183.141.168.95:3128
https://60.13.146.92:8088
http://59.59.4.22:8090
```

通过界面选择"代理 IP 更换频率"或命令行参数 `-a_proxyminute` 启用。

> **注意：** macOS 下使用代理 IP 功能需要 root 权限，否则无法通过 `ping` 检测可用代理。

---

## 内置爬虫规则

| 分类     | 规则名称                                                  |
| -------- | --------------------------------------------------------- |
| 搜索引擎 | 百度搜索、百度新闻、谷歌搜索、京东搜索、淘宝搜索          |
| 电商平台 | 京东、淘宝、考拉海购、蜜芽宝贝、顺丰海淘、Holland&Barrett |
| 新闻资讯 | 中国新闻网、网易新闻、人民网                              |
| 社交问答 | 知乎日报、知乎编辑推荐、悟空问答、微博粉丝                |
| 房产汽车 | 房天下二手房、汽车之家                                    |
| 数码科技 | ZOL 手机、ZOL 电脑、ZOL 平板、乐蛙                        |
| 分类信息 | 赶集公司、全国区号                                        |
| 社交工具 | QQ 头像                                                   |
| 学术期刊 | IJGUC                                                     |
| 其他     | 阿里巴巴、技版、文件下载测试                              |

---

## 常见问题

<details>
<summary><b>请求队列中重复的 URL 会自动去重吗？</b></summary>

默认自动去重。如需允许重复请求，设置 `Request.Reloadable = true`。

</details>

<details>
<summary><b>框架能否判断页面内容是否更新？</b></summary>

框架不内置页面变更检测，但可在规则中自定义实现。

</details>

<details>
<summary><b>请求成功的判定标准是什么？</b></summary>

以服务器是否返回响应流为准，而非 HTTP 状态码。即 404 页面也算"请求成功"。

</details>

<details>
<summary><b>请求失败后如何重试？</b></summary>

每个 URL 尝试下载指定次数后，若仍失败则进入 defer 队列。当前任务正常结束后自动重试。再次失败则保存至失败历史记录。下次执行同一规则时，可选择继承历史失败记录进行自动重试。

</details>

---

## 参与贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支：`git checkout -b feature/your-feature`
3. 提交更改：`git commit -m 'Add your feature'`
4. 推送分支：`git push origin feature/your-feature`
5. 提交 Pull Request

---

## 开源协议

本项目基于 [Apache License 2.0](LICENSE) 开源。

---

<div align="center">
  <sub>Created by <a href="https://github.com/andeya">andeya</a> — 如果觉得有帮助，请给个 Star 支持！</sub>
</div>
