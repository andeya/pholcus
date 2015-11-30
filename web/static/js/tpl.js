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
                <label>自定义输入 ( 多任务间以 " | " 隔开 )</label>\
                <textarea name="keywords" class="form-control" rows="2" placeholder="Enter ...">' + info.keywords + '</textarea>\
              </div>\
            <div class="inline">\
              <div class="form-group">\
                <label>最大采集页数</label>\
                <input name="maxPage" type="number" class="form-control" min="0" value="' + info.maxPage + '">\
              </div>' +
        threadNumHtml(info.threadNum) +
        sleepTimeHtml(info.sleepTime) +
        dockerCapHtml(info.dockerCap) +
        outputsHtml(info.outputs) +
        inheritDeduplicationHtml(info.inheritDeduplication) +
        // deduplicationTargetHtml(info.deduplicationTarget) +
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
var threadNumHtml = function(threadNum) {
    return '<div class="form-group">\
                <label>并发协程</label>\
                <input name="threadNum" type="number" class="form-control" min="' + threadNum.min + '" max="' + threadNum.max + '" value="' + threadNum['default'] + '">\
              </div>';
}

var dockerCapHtml = function(dockerCap) {
    return '<div class="form-group">\
                <label>分批输出大小</label>\
                <input name="dockerCap" type="number" class="form-control" min="' + dockerCap['min'] + '" max="' + dockerCap['max'] + '" value="' + dockerCap['default'] + '">\
              </div>';
}

var sleepTimeHtml = function(sleepTime) {
    var html1 = '<div class="form-group">\
                <label>间隔基准</label>\
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
                <label>随机延迟</label>\
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

var outputsHtml = function(outputs) {
    var html = '<div class="form-group"> \
            <label>输出方式</label>\
            <select class="form-control" name="output">';
    for (var i in outputs.menu) {
        var isSelect = "";
        if (outputs.curr == outputs.menu[i]) {
            isSelect = " selected";
        };
        html += '<option value="' + outputs.menu[i] + '"' + isSelect + '>' + outputs.menu[i] + '</option>';
    }
    return html + '</select></div>';
}

// var deduplicationTargetHtml = function(deduplicationTarget) {
//     var html = '<div class="form-group"> \
//             <label>去重样本位置</label>\
//             <select class="form-control" name="deduplicationTarget">';
//     for (var i in deduplicationTarget.menu) {
//         var isSelect = "";
//         if (deduplicationTarget.curr == deduplicationTarget.menu[i]) {
//             isSelect = " selected";
//         };
//         html += '<option value="' + deduplicationTarget.menu[i] + '"' + isSelect + '>' + deduplicationTarget.menu[i] + '</option>';
//     }
//     return html + '</select></div>';
// }

var inheritDeduplicationHtml = function(inheritDeduplication) {
    var html = '<div class="form-group"> \
            <label>继承历史去重</label>\
            <select class="form-control" name="inheritDeduplication">';

    var True = "";
    var False = "";
    if (inheritDeduplication == true) {
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
        case _stop:
            return '<button type="button" id="btn-pause" class="btn btn-warning" onclick="pauseRecover()" disabled="disabled">Pause</button>\
            <button type="submit" id="btn-run" class="btn btn-primary" data-type="run">Run</button>';
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
