// Package agent generates user agents strings for well known browsers
// and for custom browsers.
//
// When submitting patches to add user agents formats, please *always* include
// "{{.Coms}}" between the opening ( and closing ) braces, even if you're
// sure the browser would never have additional comments.
package agent

import (
	"bytes"
	"math/rand"
	"runtime"
	"strings"
	"text/template"
	"time"
)

// TemplateData structure for template data.
type TemplateData struct {
	Name string
	Ver  string
	OSN  string
	OSV  string
	Coms string
}

// OSAttributes stores OS attributes.
type OSAttributes struct {
	// OSName is the operating system name.
	OSName string
	// OSVersion is the operating system version.
	OSVersion string
	// Comments are additional comments to add to a user agent string.
	Comments []string
}

const (
	// Windows operating system.
	Windows int = iota
	// Linux based operating system.
	Linux
	// Macintosh/OS X operating system.
	Macintosh
)

// DefaultOSAttributes stores default OS attributes.
var DefaultOSAttributes = map[int]OSAttributes{
	Windows:   {"Windows NT", "10.0", []string{"Win64", "x64"}},
	Linux:     {"Linux", "x86_64", []string{}},
	Macintosh: {"Intel Mac OS X", "10_15_7", []string{}},
}

type (
	// Formats is a collection of UA format strings.
	// key is the browser version.
	// value is the browser info.
	Formats map[string]string

	// UAData stores information on a browser user agent.
	UAData struct {
		TopVersion string
		DefaultOS  int
		Formats    Formats
	}

	// UATable is a collection of UAData values.
	// key is the name of the browser.
	UATable map[string]UAData
)

// Database is the "database" of user agents.
var Database = UATable{
	"chrome": {
		"127.0.6533.73",
		Windows,
		Formats{
			"127": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{{.Ver}} Safari/537.36",
			"126": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{{.Ver}} Safari/537.36",
			"125": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{{.Ver}} Safari/537.36",
			"124": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{{.Ver}} Safari/537.36",
			"123": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{{.Ver}} Safari/537.36",
			"122": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{{.Ver}} Safari/537.36",
			"121": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{{.Ver}} Safari/537.36",
			"120": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/{{.Ver}} Safari/537.36",
		},
	},
	"firefox": {
		"127.0",
		Windows,
		Formats{
			"127": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:127.0) Gecko/20100101 Firefox/{{.Ver}}",
			"126": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:126.0) Gecko/20100101 Firefox/{{.Ver}}",
			"125": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:125.0) Gecko/20100101 Firefox/{{.Ver}}",
			"124": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:124.0) Gecko/20100101 Firefox/{{.Ver}}",
			"123": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:123.0) Gecko/20100101 Firefox/{{.Ver}}",
			"122": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:122.0) Gecko/20100101 Firefox/{{.Ver}}",
			"121": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:121.0) Gecko/20100101 Firefox/{{.Ver}}",
			"120": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}; rv:120.0) Gecko/20100101 Firefox/{{.Ver}}",
		},
	},
	"edge": {
		"127.0.2651.74",
		Windows,
		Formats{
			"127": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36 Edg/{{.Ver}}",
			"126": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36 Edg/{{.Ver}}",
			"125": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36 Edg/{{.Ver}}",
			"124": "Mozilla/5.0 ({{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/{{.Ver}}",
		},
	},
	"safari": {
		"17.5",
		Macintosh,
		Formats{
			"17": "Mozilla/5.0 (Macintosh; {{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/{{.Ver}} Safari/605.1.15",
			"16": "Mozilla/5.0 (Macintosh; {{.OSN}} {{.OSV}}{{.Coms}}) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/{{.Ver}} Safari/605.1.15",
		},
	},
	"googlebot": {
		"2.1",
		Linux,
		Formats{
			"2": "Mozilla/5.0 (compatible; Googlebot/{{.Ver}}; +http://www.google.com/bot.html{{.Coms}})",
			"1": "Googlebot/{{.Ver}} (+http://www.google.com/bot.html{{.Coms}})",
		},
	},
	"bingbot": {
		"2.0",
		Windows,
		Formats{
			"2": "Mozilla/5.0 (compatible; bingbot/{{.Ver}}; +http://www.bing.com/bingbot.htm{{.Coms}})",
		},
	},
	"yahoobot": {
		"2.0",
		Linux,
		Formats{
			"2": "Mozilla/5.0 (compatible; Yahoo! Slurp; http://help.yahoo.com/help/us/ysearch/slurp{{.Coms}})",
		},
	},
	"default": {
		"1.0",
		Linux,
		Formats{
			"1": "{{.Name}}/{{.Ver}} ({{.OSN}} {{.OSV}}{{.Coms}})",
		},
	},
}

// UserAgents holds all generated user agent strings.
var UserAgents = map[string][]string{}

func init() {
	for browser, userAgentData := range Database {
		if browser == "default" {
			continue
		}
		os := userAgentData.DefaultOS
		osAttribs := DefaultOSAttributes[os]
		for version, _ := range userAgentData.Formats {
			ua := createFromDetails(
				browser,
				version,
				osAttribs.OSName,
				osAttribs.OSVersion,
				osAttribs.Comments)
			UserAgents["all"] = append(UserAgents["all"], ua)

			if browser != "googlebot" && browser != "bingbot" && browser != "yahoobot" {
				UserAgents["common"] = append(UserAgents["common"], ua)
			}
		}
	}
	l := len(UserAgents["common"])
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Intn(l)
	UserAgents["all"][0], UserAgents["all"][idx] = UserAgents["all"][idx], UserAgents["all"][0]
	UserAgents["common"][0], UserAgents["common"][idx] = UserAgents["common"][idx], UserAgents["common"][0]
}

// Create generates and returns a complete user agent string.
func CreateReal() string {
	return createFromDetails("Surfer", "1.0", osName(), osVersion(), []string{runtime.Version()})
}

// CreateDefault returns a user agent string using default values.
func CreateDefault(browser string) string {
	bn := strings.ToLower(browser)
	data := Database[bn]
	os := data.DefaultOS
	osAttribs := DefaultOSAttributes[os]

	return createFromDetails(
		browser,
		data.TopVersion,
		osAttribs.OSName,
		osAttribs.OSVersion,
		osAttribs.Comments)
}

// CreateVersion generates and returns a complete user agent string for a specific browser version.
func CreateVersion(browser, version string) string {
	bn := strings.ToLower(browser)
	data := Database[bn]
	os := data.DefaultOS
	osAttribs := DefaultOSAttributes[os]

	return createFromDetails(
		browser,
		version,
		osAttribs.OSName,
		osAttribs.OSVersion,
		osAttribs.Comments)
}

// TopVersion returns the most recent version for the given browser name.
func TopVersion(bname string) string {
	bname = strings.ToLower(bname)
	data, ok := Database[bname]
	if ok {
		return data.TopVersion
	}
	return Database["default"].TopVersion
}

// Format returns the format string for the given browser name and version.
//
// When a format can't be found for a version, the first format string for the browser
// is returned. When a format can't be found for the browser the default format is
// returned.
func Format(bname, bver string) string {
	bname = strings.ToLower(bname)
	majVer := strings.Split(bver, ".")[0]
	data, ok := Database[bname]
	if ok {
		format, ok := data.Formats[majVer]
		if ok {
			return format
		} else {
			top := TopVersion(bname)
			majVer = strings.Split(top, ".")[0]
			return data.Formats[majVer]
		}
	}

	return Database["default"].Formats["1"]
}

// createFromDetails generates and returns a complete user agent string.
func createFromDetails(bname, bver, osname, osver string, c []string) string {
	if bver == "" {
		bver = TopVersion(bname)
	}
	comments := strings.Join(c, "; ")
	if comments != "" {
		comments = "; " + comments
	}

	data := TemplateData{bname, bver, osname, osver, comments}
	buff := &bytes.Buffer{}
	t := template.New("formatter")
	t.Parse(Format(bname, bver))
	t.Execute(buff, data)

	return buff.String()
}
