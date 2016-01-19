// https://groups.google.com/forum/#!searchin/golang-nuts/idnet$20netid/golang-nuts/guM3ZHHqSF0/K1pBpMqQSSwJ
// http://play.golang.org/p/BFFDxphKYK

package main

import (
	"fmt"
	"github.com/henrylee2cn/tconfig/xml/x2j"
)

// demo how to compensate for irregular tag labels in data
// "netid" vs. "idnet"
var doc1 = `
<?xml version="1.0" encoding="UTF-8"?>
<data>
    <netid>
        <disable>no</disable>
        <text1>default:text</text1>
        <word1>default:word</word1>
    </netid>
</data>
`
var doc2 = `
<?xml version="1.0" encoding="UTF-8"?>
<data>
    <idnet>
        <disable>yes</disable>
        <text1>default:text</text1>
        <word1>default:word</word1>
    </idnet>
</data>
`

func main() {
	var docs = []string{doc1, doc2}

	for n, doc := range docs {
		fmt.Println("\nTestValuesFromTagPath2(), iteration:", n, "\n", doc)

		m, _ := x2j.DocToMap(doc)
		fmt.Println("map:", x2j.WriteMap(m))

		v, _ := x2j.ValuesFromTagPath(doc, "data.*")
		fmt.Println("\npath == data.*: len(v):", len(v))
		for key, val := range v {
			fmt.Println(key, ":", val)
		}
		mm := v[0]
		for key, val := range mm.(map[string]interface{}) {
			fmt.Println(key, ":", val)
		}

		v, _ = x2j.ValuesFromTagPath(doc, "data.*.*")
		fmt.Println("\npath == data.*.*: len(v):", len(v))
		for key, val := range v {
			fmt.Println(key, ":", val)
		}
	}
}
