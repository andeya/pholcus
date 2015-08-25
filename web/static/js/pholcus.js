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
                closeBtn: false
            });
            layer.full(index);
            // if (wslog == null) {
            //     startLog();
            // };
            break;

            // 任务开始通知
        case "run":
            if (data.status != 1) {
                return
            };
            $("#btn-run").text("Stop");
            $("#btn-run").attr("data-type", "stop");

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

var Html = function(info) {
    if (info.mode == client) {
        return logBoxHtml(client);
    };
    return '<div class="step2"><form role="form" id="js-form" name="pholcus" onsubmit="return runStop();" method="POST" enctype="multipart/form-data">\
            <div class="box form-1">\
              <!--<div class="box-header"><h3 class="box-title">All Spiders</h3></div>-->\
              <div class="box-body table-responsive no-padding" id="spider-box">\
                <table class="table table-hover">\
                  <tbody id="allSpiders">\
                    <tr>\
                      <th>#</th>\
                      <th>ID</th>\
                      <th>Name</th>\
                      <th>Description</th>\
                    </tr>' + spiderMenuHtml(info.spiderMenu) + '</tbody></table></div></div>\
            <div class="form-2 box">\
              <div class="form-group">\
                <label>Keywords ( Add " | " between multiple words )</label>\
                <textarea name="keywords" class="form-control" rows="2" placeholder="Enter ..."></textarea>\
              </div>\
              <div class="form-group">\
                <label>Maximum Number of Pages</label>\
                <input name="maxPage" type="number" class="form-control" min="0" value="0">\
              </div><div class="inline">' + threadNumHtml(info.threadNum) + dockerCapHtml(info.dockerCap) + '</div>' +
        '<div class="inline">' + sleepTimeHtml(info.sleepTime) + '</div>' + outputMenuHtml(info.outputMenu) + '</div>\
            <div class="box-footer">\
                ' + pauseHtml(info.mode) + '<button type="submit" id="btn-run" class="btn btn-primary" data-type="run">Run</button>\
            </div>\
          </form>' + logBoxHtml(info.mode) + '</div>';
}

var spiderMenuHtml = function(spiderMenu) {
    var html = '';

    for (var i in spiderMenu) {
        html += '<tr>\
            <td>\
                <div class="checkbox">\
                  <label for="spider-' + i + '">\
                    <input name="spiders" id="spider-' + i + '" type="checkbox" value="' + spiderMenu[i].name + '" >\
                  </label>\
                </div>\
            </td>\
            <td><label for="spider-' + i + '">' + i + '</label></td>\
            <td><label for="spider-' + i + '">' + spiderMenu[i].name + '</label></td>\
            <td><label for="spider-' + i + '">' + spiderMenu[i].description + '</label></td>\
        <tr>'
    }

    return html;
}
var threadNumHtml = function(threadNum) {
    return '<div class="form-group">\
                <label>Large Number of Coroutine</label>\
                <input name="threadNum" type="number" class="form-control" min="' + threadNum.min + '" max="' + threadNum.max + '" value="' + threadNum['default'] + '">\
              </div>';
}

var dockerCapHtml = function(dockerCap) {
    return '<div class="form-group">\
                <label>Size of Batch Output</label>\
                <input name="dockerCap" type="number" class="form-control" min="' + dockerCap['min'] + '" max="' + dockerCap['max'] + '" value="' + dockerCap['default'] + '">\
              </div>';
}

var sleepTimeHtml = function(sleepTime) {
    var html1 = '<div class="form-group">\
                <label>Minimum Pause Time</label>\
                <select class="form-control" name="baseSleeptime">';
    for (var i in sleepTime.base) {
        var isSelect = ""
        if (sleepTime.base[i] == sleepTime['default'][0]) {
            isSelect = " selected";
        };
        html1 += '<option value="' + sleepTime.base[i] + '"' + isSelect + '>' + sleepTime.base[i] + ' ms</option>';
    };
    html1 += '</select></div>';

    var html2 = '<div class="form-group">\
                <label>Random Pause Time</label>\
                <select class="form-control" name="randomSleepPeriod">';
    for (var i in sleepTime.random) {
        var isSelect = "";
        if (sleepTime.random[i] == sleepTime['default'][1]) {
            isSelect = " selected";
        };
        html2 += '<option value="' + sleepTime.random[i] + '"' + isSelect + '>' + sleepTime.random[i] + ' ms</option>';
    };
    html2 += '</select></div>';

    return html1 + html2;
}

var outputMenuHtml = function(outputMenu) {
    var html = '<div class="form-group"> \
            <label>Mode Output</label>\
            <select class="form-control" name="output">';
    for (var i in outputMenu) {
        var isSelect = "";
        if (i == 0) {
            isSelect = " selected";
        };
        html += '<option value="' + outputMenu[i] + '"' + isSelect + '>' + outputMenu[i] + '</option>';
    }
    return html + '</select></div>';
}

var pauseHtml = function(mode) {
    if (parseInt(mode) != offline) {
        return "";
    }
    return '<button type="button" id="btn-pause" class="btn btn-warning" onclick="pauseRecover(this)" disabled="disabled">Pause</button>';
}

var logBoxHtml = function(m) {
    if (m == client) {
        return '<div class="box log client">\
              <div class="box-body chat" id="log-box">\
              </div>\
          </div>';
    };
    return '<div class="box log">\
              <div class="box-body chat" id="log-box">\
              </div>\
          </div>';
}

$(function() {
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
    }
})

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
    }
}

// 按模式启动Pholcus
function firstStep() {
    $("#init").text(" 开  启 …").css({
        "background-color": "#286090",
        "border-color": "#204d74"
    });
    var formJson = {
        'operate': 'init',
        'mode': document.step1.elements['mode'].value,
        'port': document.step1.elements['port'].value,
        'ip': document.step1.elements['ip'].value,
    };
    ws.onsend(formJson);
    return false;
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
