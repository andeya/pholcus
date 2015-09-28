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
                <label>Keywords ( Add " | " between multiple words )</label>\
                <textarea name="keywords" class="form-control" rows="2" placeholder="Enter ...">'+info.keywords+'</textarea>\
              </div>\
              <div class="form-group">\
                <label>Maximum Number of Pages</label>\
                <input name="maxPage" type="number" class="form-control" min="0" value="'+info.maxPage+'">\
              </div><div class="inline">' + threadNumHtml(info.threadNum) + dockerCapHtml(info.dockerCap) + '</div>' +
        '<div class="inline">' + sleepTimeHtml(info.sleepTime) + '</div>' + outputsHtml(info.outputs) + '</div>\
            <div class="box-footer">\
                ' + pauseHtml(info.mode,info.status) + '<button type="submit" id="btn-run" class="btn btn-primary" data-type="run">Run</button>\
            </div>\
          </form>' + logBoxHtml(info.mode) + '</div>';
}

var spidersHtml = function(spiders) {
    var html = '';

    for (var i in spiders.memu) {
        html += '<tr>\
            <td>\
                <div class="checkbox">\
                  <label for="spider-' + i + '">\
                    <input name="spiders" id="spider-' + i + '" type="checkbox" value="' + spiders.memu[i].name + '"' +
            function() {
                if (spiders.curr[spiders.memu[i].name]) {
                    return "checked";
                }
                return
            }() + '>\
                  </label>\
                </div>\
            </td>\
            <td><label for="spider-' + i + '">' + i + '</label></td>\
            <td><label for="spider-' + i + '">' + spiders.memu[i].name + '</label></td>\
            <td><label for="spider-' + i + '">' + spiders.memu[i].description + '</label></td>\
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

var outputsHtml = function(outputs) {
    var html = '<div class="form-group"> \
            <label>Mode Output</label>\
            <select class="form-control" name="output">';
    for (var i in outputs.memu) {
        var isSelect = "";
        if (outputs.curr == outputs.memu[i]) {
            isSelect = " selected";
        };
        html += '<option value="' + outputs.memu[i] + '"' + isSelect + '>' + outputs.memu[i] + '</option>';
    }
    return html + '</select></div>';
}

var pauseHtml = function(mode,status) {
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
