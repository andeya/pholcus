// websocket
var wsUri = "ws://" + location.hostname + ":" + location.port + "/ws";
var ws = null;
if ('WebSocket' in window) {
    ws = new WebSocket(wsUri);
} else if ('MozWebSocket' in window) {
    ws = new MozWebSocket(wsUri);
}

ws.onopen = function() {
    console.log("connected to " + wsUri);
};


ws.onclose = function(e) {
    console.log("connection closed (" + wsUri + " : " + e.code + "," + e.reason + ")");
}

ws.onerror = function(e) {
    for (var p in e) {
        console.log(p + "=" + e[p]);
    }
};

// 发送api
ws.onsend = function(data) {
    var dataStr = JSON.stringify(data);
    ws.send(dataStr);
    console.log("send: " + dataStr);
}

// 接收api
ws.onmessage = function(m) {
    var data = JSON.parse(m.data)
    console.log(data);
    switch (data.operate) {
        // 初始化运行参数
        case "init":
            $("#init").text(" 开  启 ").css({
                "background-color": "#337ab7",
                "border-color": "#2e6da4"
            });
            // 设置当前运行模式
            mode = data.mode;
            // 打开软件界面
            var index = layer.open({
                type: 1,
                title: data.title,
                content: Html(data),
                // area: ['300px', '195px'],
                maxmin: false,
                scrollbar: false,
                move: false,
            });

            layer.full(index);
            $(".layui-layer-close1").attr("title", "退出").click(function() {
                Close();
            });

            if (data.status == _unknow || data.status == _stop) {
                break;
            };

            // 任务开始通知
        case "run":
            if (data.status == _unknow) {
                return
            };
            $("#btn-run").text("Stop").attr("data-type", "stop");

            if (data.mode == offline) {
                $("#btn-run").text("Stop").attr("data-type", "stop").addClass("btn-danger").removeClass("btn-primary");
                $("#btn-pause").text("Pause").removeAttr("disabled").show();
            };
            break;

            // 任务结束通知
        case "stop":
            $("#btn-run").text("Run").attr("data-type", "run").removeAttr("disabled");
            if (data.mode == offline) {
                $("#btn-run").text("Run").attr("data-type", "run").addClass("btn-primary").removeClass("btn-danger");
                $("#btn-pause").hide();
            };
            break;
    }
}


// 当前运行模式
var mode = "";

function selectMode(m) {
    switch (m) {
        case offline:
            $("#js_mode").text("单机模式");
            $("#step1 .js_port").hide();
            $("#step1 .js_ip").hide();
            $("#mode").val(offline);
            break;
        case server:
            $("#js_mode").text("服务端模式");
            $("#step1 .js_ip").hide();
            $("#step1 .js_port").show();
            $("#mode").val(server);
            break;
        case client:
            $("#js_mode").text("客户端模式");
            $("#step1 .js_ip").show();
            $("#step1 .js_port").show();
            $("#mode").val(client);
            break;
        default:
            $("#js_mode").text("运行模式");
            $("#step1 .js_port").hide();
            $("#step1 .js_ip").hide();
            $("#mode").val(unset);
    }
    $("#init").removeAttr("disabled");
}


// 执行入口
function home() {
    switch (parseInt($("#mode").val())) {
        case offline:
            $("#js_mode").text("单机模式");
            break;
        case server:
            $("#js_mode").text("服务端模式");
            break;
        case client:
            $("#js_mode").text("客户端模式");
            break;
        default:
            $("#init").attr("disabled", "disabled");
            return;
    }
    Open('goon');
}

// 按模式启动Pholcus
function Open(operate) {
    $("#init").text(" 开  启 …").css({
        "background-color": "#286090",
        "border-color": "#204d74"
    });
    var formJson = {
        'operate': operate,
        'mode': document.step1.elements['mode'].value,
        'port': document.step1.elements['port'].value,
        'ip': document.step1.elements['ip'].value,
    };
    ws.onsend(formJson);
    return false;
}

// 退出
function Close() {
    ws.onsend({
        'operate': 'exit'
    });
    selectMode(unset);
    $("#init").attr("disabled", "disabled");
}

// 开始或停止运行任务
function runStop() {
    if ($("#btn-run").attr("data-type") == 'run') {
        ws.onsend(getForm());
    } else if (mode == offline) {
        $("#btn-run").text("Stopping...").attr("disabled", "disabled");
        ws.onsend({
            'operate': 'stop'
        });
    };
    return false;
};

// 获取表单值
function getForm() {
    return {
        'operate': 'run',
        'spiders': getSpiders(),
        'keywords': document.pholcus.elements['keywords'].value,
        'threadNum': document.pholcus.elements['threadNum'].value,
        'maxPage': document.pholcus.elements['maxPage'].value,
        'dockerCap': document.pholcus.elements['dockerCap'].value,
        'baseSleeptime': document.pholcus.elements['baseSleeptime'].value,
        'randomSleepPeriod': document.pholcus.elements['randomSleepPeriod'].value,
        'output': document.pholcus.elements['output'].value,
    }
}

// 返回选择的蜘蛛
function getSpiders() {
    var spiders = [];
    var spiderAll = document.getElementsByName('spiders');
    for (var i = spiderAll.length - 1; i >= 0; i--) {
        if (spiderAll[i].checked) {
            spiders[spiders.length] = spiderAll[i].value;
        }
    };
    return spiders
};

// 暂停恢复运行
function pauseRecover(self) {
    ws.onsend({
        'operate': 'pauseRecover'
    });
    if ($(self).text() == "Pause") {
        $(self).text("Go on...").addClass("btn-info").removeClass("btn-warning");
    } else {
        $(self).text("Pause").addClass("btn-warning").removeClass("btn-info");
    };
    return false;
};

// ********************************* 打印log信息 ************************************** \\
var wsLogUri = "ws://" + location.hostname + ":" + location.port + "/ws/log";
var wslog = null;


if ('WebSocket' in window) {
    wslog = new WebSocket(wsLogUri);
} else if ('MozWebSocket' in window) {
    wslog = new MozWebSocket(wsLogUri);
}

wslog.onopen = function() {
    console.log("connected to " + wsLogUri);
};


wslog.onclose = function(e) {
    console.log("connection closed (" + wsLogUri + " : " + e.code + "," + e.reason + ")");
};

// 接收api, 打印Log
wslog.onmessage = function(m) {
    var div = document.createElement("div");
    div.className = "item";
    div.innerHTML = '<p class="message">' + m.data.replace(/\s/g, '&nbsp;') + '</p>';
    document.getElementById('log-box').appendChild(div);
    document.getElementById('log-box').scrollTop = document.getElementById('log-box').scrollHeight;
};


window.onbeforeunload = function() {
    wslog.close();
    ws.close();
};
