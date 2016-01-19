// https://groups.google.com/forum/#!topic/golang-nuts/cok6xasvI3w
// retrieve 'src' values from 'image' tags

package main

import (
	"fmt"
	"github.com/henrylee2cn/tconfig/xml/x2j"
)

var doc = `
<doc>
	<image src="something.png"></image>
	<something>
		<image src="something else.jpg"></image>
		<title>something else</title>
	</something>
	<more_stuff>
		<some_images>
			<image src="first.gif"></image>
			<image src="second.gif"></image>
		</some_images>
	</more_stuff>
</doc>`

func main() {
	// get all image tag values - []interface{}
	images, err := x2j.ValuesForTag(doc, "image")
	if err != nil {
		fmt.Println("error parsing doc:", err.Error())
		return
	}

	sources := make([]string, 0)
	for _, v := range images {
		// ValuesForKey requires a map[string]interface{} value
		// as a starting point ... thinks its a doc
		m := make(map[string]interface{}, 1)
		m["dummy"] = v
		ss := x2j.ValuesForKey(m, "-src")
		for _, s := range ss {
			sources = append(sources, s.(string))
		}
	}

	for _, src := range sources {
		fmt.Println(src)
	}
}
