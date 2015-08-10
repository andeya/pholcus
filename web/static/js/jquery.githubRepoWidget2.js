/**
 * Original: https://github.com/JoelSutherland/GitHub-jQuery-Repo-Widget
 * Modify by tsl0922@gmail.com 
 */
$(function() {

    var i = 0;

    $('.github-widget').each(function() {

        if (i == 0) $('head').append('<style type="text/css">.github-box{font-family:helvetica,arial,sans-serif;font-size:13px;line-height:18px;background:#fafafa;color:#666;border-radius:3px}.github-box a{color:#4183c4;border:0;text-decoration:none}.github-box .github-box-title{position:relative;border-radius:3px 3px 0 0;background:#fcfcfc;background:-moz-linear-gradient(#fcfcfc,#ebebeb);background:-webkit-linear-gradient(#fcfcfc,#ebebeb);}.github-box .github-box-title h3{font-family:helvetica,arial,sans-serif;font-weight:normal;font-size:16px;color:gray;margin:0;padding:10px 10px 10px 80px;background:url(http://www.oschina.net/img/github_logo.gif) center left no-repeat}.github-box .github-box-title h3 .repo{font-weight:bold}.github-box .github-box-title .github-stats{position:absolute;top:8px;right:10px;background:white;border:1px solid #ddd;border-radius:3px;font-size:11px;font-weight:bold;line-height:21px;height:21px;padding-left:2px;}.github-box .github-box-title .github-stats a{display:inline-block;height:21px;color:#666;padding:0 5px 0 5px;}.github-box .github-box-title .github-stats .watchers{border-right:1px solid #ddd;background-position:3px -2px;}.github-box .github-box-title .github-stats .forks{background-position:0 -52px;padding-left:5px}.github-box .github-box-content{padding:10px;font-weight:300}.github-box .github-box-content p{margin:0}.github-box .github-box-content .link{font-weight:bold}.github-box .github-box-download{position:relative;border-top:1px solid #ddd;background:white;border-radius:0 0 3px 3px;padding:10px;height:24px}.github-box .github-box-download .updated{margin:0;font-size:11px;color:#666;line-height:24px;font-weight:300}.github-box .github-box-download .updated strong{font-weight:bold;color:#000}.github-box .github-box-download .download{position:absolute;display:block;top:10px;right:10px;height:24px;line-height:24px;font-size:12px;color:#666;font-weight:bold;text-shadow:0 1px 0 rgba(255,255,255,0.9);padding:0 10px;border:1px solid #ddd;border-bottom-color:#bbb;border-radius:3px;background:#f5f5f5;background:-moz-linear-gradient(#f5f5f5,#e5e5e5);background:-webkit-linear-gradient(#f5f5f5,#e5e5e5);}.github-box .github-box-download .download:hover{color:#527894;border-color:#cfe3ed;border-bottom-color:#9fc7db;background:#f1f7fa;background:-moz-linear-gradient(#f1f7fa,#dbeaf1);background:-webkit-linear-gradient(#f1f7fa,#dbeaf1);</style>');
        i++;

        var $container = $(this);
        var repo_name = $container.data('repo');
        var html_encode = function(str) {
            if (!str || str.length == 0) return "";
            return str.replace(/</g, "&lt;").replace(/>/g, "&gt;");
        }

        $.ajax({
            url: 'https://api.github.com/repos/' + repo_name,
            dataType: 'jsonp',

            success: function(results) {
                var repo = results.data;

                var url_regex = /((http|https):\/\/)*[\w-]+(\.[\w-]+)+([\w.,@?^=%&amp;:\/~+#-]*[\w@?^=%&amp;\/~+#-])?/
                if (repo.homepage && (m = repo.homepage.match(url_regex))) {
                    if (m[0] && !m[1]) repo.homepage = 'http://' + m[0];
                } else {
                    repo.homepage = '';
                }

                var $widget = $(' \
					<div class="github-box repo">  \
					    <div class="github-box-title"> \
					        <div class="github-stats"> \
					        Star<a class="watchers" title="Star" href="' + repo.url.replace('api.', '').replace('repos/', '') + '/stargazers" target="_blank">' + repo.stargazers_count + '</a> \
					        Fork<a class="forks" title="Forks" href="' + repo.url.replace('api.', '').replace('repos/', '') + '/network" target="_blank">' + repo.forks + '</a> \
					        </div> \
					    </div> \
					</div> \
				');

                $widget.appendTo($container);
            }
        })
    });

});