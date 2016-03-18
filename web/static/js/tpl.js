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
                    </tr>' + spidersHtml(info.spiders) + '</tbody></table></div></div>\
            <div class="form-2 box">\
              <div class="form-group">\
                <label>自定义配置（多任务请分别多包一层“<>”）</label>\
                <textarea name="Keyins" class="form-control" rows="2" placeholder="Enter ...">' + info.Keyins + '</textarea>\
              </div>\
            <div class="inline">\
              <div class="form-group">\
                <label>采集上限（默认限制URL数）</label>\
                <input name="Limit" type="number" class="form-control" min="0" value="' + info.Limit + '">\
              </div>' +
        ThreadNumHtml(info.ThreadNum) +
        PausetimeHtml(info.Pausetime) +
        ProxyMinuteHtml(info.ProxyMinute) +
        DockerCapHtml(info.DockerCap) +
        OutTypeHtml(info.OutType) +
        SuccessInheritHtml(info.SuccessInherit) +
        FailureInheritHtml(info.FailureInherit) +
        '</div>' +
        '</div>\
            <div class="box-footer">\
                ' + btnHtml(info.mode, info.status) +
        '</div>\
          </form>' + logBoxHtml(info.mode) + '</div>';
}

var spidersHtml = function(spiders) {
    var html = '';

    for (var i in spiders.menu) {
        html += '<tr>\
            <td>\
                <div class="checkbox">\
                  <label for="spider-' + i + '">\
                    <input name="spiders" id="spider-' + i + '" type="checkbox" value="' + spiders.menu[i].name + '"' +
            function() {
                if (spiders.curr[spiders.menu[i].name]) {
                    return "checked";
                }
                return
            }() + '>\
                  </label>\
                </div>\
            </td>\
            <td><label for="spider-' + i + '">' + i + '</label></td>\
            <td><label for="spider-' + i + '">' + spiders.menu[i].name + '</label></td>\
            <td><label for="spider-' + i + '">' + spiders.menu[i].description + '</label></td>\
        <tr>'
    }

    return html;
}
var ThreadNumHtml = function(ThreadNum) {
    return '<div class="form-group">\
                <label>并发协程</label>\
                <input name="ThreadNum" type="number" class="form-control" min="' + ThreadNum.min + '" max="' + ThreadNum.max + '" value="' + ThreadNum.curr + '">\
              </div>';
}

var DockerCapHtml = function(DockerCap) {
    return '<div class="form-group">\
                <label>分批输出限制</label>\
                <input name="DockerCap" type="number" class="form-control" min="' + DockerCap.min + '" max="' + DockerCap.max + '" value="' + DockerCap.curr + '">\
              </div>';
}

var PausetimeHtml = function(Pausetime) {
    var html = '<div class="form-group">\
                <label>暂停时长参考</label>\
                <select class="form-control" name="Pausetime">';
    for (var i in Pausetime.menu) {
        var isSelect = ""
        if (Pausetime.menu[i] == Pausetime.curr[0]) {
            isSelect = " selected";
        };
        if (Pausetime.menu[i] == 0) {
            html += '<option value="' + Pausetime.menu[i] + '"' + isSelect + '>' + "无暂停" + '</option>';
        } else {
            html += '<option value="' + Pausetime.menu[i] + '"' + isSelect + '>' + Pausetime.menu[i] + ' ms</option>';
        }
    };
    html += '</select></div>';
    return html;
}

var ProxyMinuteHtml = function(ProxyMinute) {
    var html = '<div class="form-group">\
                <label>代理IP更换频率</label>\
                <select class="form-control" name="ProxyMinute">';
    for (var i in ProxyMinute.menu) {
        var isSelect = ""
        if (ProxyMinute.menu[i] == ProxyMinute.curr[0]) {
            isSelect = " selected";
        };
        if (ProxyMinute.menu[i] == 0) {
            html += '<option value="' + ProxyMinute.menu[i] + '"' + isSelect + '>' + "不使用代理" + '</option>';
        } else {
            html += '<option value="' + ProxyMinute.menu[i] + '"' + isSelect + '>' + ProxyMinute.menu[i] + ' min</option>';
        }
    };
    html += '</select></div>';
    return html;
}

var OutTypeHtml = function(OutType) {
    var html = '<div class="form-group"> \
            <label>输出方式</label>\
            <select class="form-control" name="OutType">';
    for (var i in OutType.menu) {
        var isSelect = "";
        if (OutType.curr == OutType.menu[i]) {
            isSelect = " selected";
        };
        html += '<option value="' + OutType.menu[i] + '"' + isSelect + '>' + OutType.menu[i] + '</option>';
    }
    return html + '</select></div>';
}

var SuccessInheritHtml = function(SuccessInherit) {
    var html = '<div class="form-group"> \
            <label>继承并保存成功记录</label>\
            <select class="form-control" name="SuccessInherit">';

    var True = "";
    var False = "";
    if (SuccessInherit == true) {
        True = " selected";
    } else {
        False = " selected";
    };

    html += '<option value="true"' + True + '>' + "Yes" + '</option>';
    html += '<option value="false"' + False + '>' + "No" + '</option>';
    return html + '</select></div>';
}

var FailureInheritHtml = function(FailureInherit) {
    var html = '<div class="form-group"> \
            <label>继承并保存失败记录</label>\
            <select class="form-control" name="FailureInherit">';

    var True = "";
    var False = "";
    if (FailureInherit == true) {
        True = " selected";
    } else {
        False = " selected";
    };

    html += '<option value="true"' + True + '>' + "Yes" + '</option>';
    html += '<option value="false"' + False + '>' + "No" + '</option>';
    return html + '</select></div>';
}

var btnHtml = function(mode, status) {
    if (parseInt(mode) != offline) {
        return '<button type="submit" id="btn-run" class="btn btn-primary" data-type="run">Run</button>';
    }
    switch (status) {
        case _stopped:
            return '<button type="button" id="btn-pause" class="btn btn-warning" onclick="pauseRecover()" disabled="disabled">Pause</button>\
            <button type="submit" id="btn-run" class="btn btn-primary" data-type="run">Run</button>';
        case _stop:
            return '<button type="button" id="btn-pause" class="btn btn-warning" onclick="pauseRecover()" disabled="disabled">Pause</button>\
            <button type="submit" id="btn-run" class="btn btn-danger" data-type="stop" disabled="disabled">Stopping...</button>';
        case _run:
            return '<button type="button" id="btn-pause" class="btn btn-warning" onclick="pauseRecover()" style="display:inline-block;" >Pause</button>\
            <button type="submit" id="btn-run" class="btn btn-danger" data-type="stop">Stop</button>';
        case _pause:
            return '<button type="button" id="btn-pause" class="btn btn-info" onclick="pauseRecover()" style="display:inline-block;" >Go on...</button>\
            <button type="submit" id="btn-run" class="btn btn-danger" data-type="stop">Stop</button>';
    }
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
