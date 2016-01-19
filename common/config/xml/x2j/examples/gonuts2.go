// https://groups.google.com/forum/#!topic/golang-nuts/V83jUKluLnM
// http://play.golang.org/p/alWGk4MDBc

package main

import (
	"fmt"
	"github.com/henrylee2cn/tconfig/xml/x2j"
)

const strXml = `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
<s:Body>
<GetClaimStatusCodesResponse xmlns="http://tempuri.org/">
<GetClaimStatusCodesResult xmlns:a="http://schemas.datacontract.org/2004/07/MRA.Claim.WebService.Domain" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
<a:ClaimStatusCodeRecord>
<a:Active>true</a:Active>
<a:Code>A</a:Code>
<a:Description>Initial Claim Review/Screening</a:Description>
</a:ClaimStatusCodeRecord>
<a:ClaimStatusCodeRecord>
<a:Active>true</a:Active>
<a:Code>B</a:Code>
<a:Description>Initial Contact Made w/ Provider</a:Description>
</a:ClaimStatusCodeRecord>
</GetClaimStatusCodesResult>
</GetClaimStatusCodesResponse>
</s:Body>
</s:Envelope>
`

const strXmlA = `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
<s:Body>
<GetClaimStatusCodesResponse xmlns="http://tempuri.org/">
<GetClaimStatusCodesResult xmlns:a="http://schemas.datacontract.org/2004/07/MRA.Claim.WebService.Domain" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
<a:ClaimStatusCodeRecord>
<a:Active>true</a:Active>
<a:Code>A</a:Code>
<a:Description>Initial Claim Review/Screening</a:Description>
</a:ClaimStatusCodeRecord>
</GetClaimStatusCodesResult>
</GetClaimStatusCodesResponse>
</s:Body>
</s:Envelope>
`

const strXml2 = `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
<s:Body>
<GetClaimStatusCodesResponse xmlns="http://tempuri.org/">
<GetClaimStatusCodesResult xmlns:a="http://schemas.datacontract.org/2004/07/MRA.Claim.WebService.Domain" xmlns:i="http://www.w3.org/2001/XMLSchema-instance">
</GetClaimStatusCodesResult>
</GetClaimStatusCodesResponse>
</s:Body>
</s:Envelope>
`

func main() {
	docs := []string{strXml, strXmlA, strXml2}
	fullPath(docs)
	partPath1(strXml)
	partPath2(strXml)
	partPath3(strXml)
	partPath4(strXml)
	partPath5(strXml)
	partPath6(strXml)
}

func fullPath(docs []string) {
	for i, doc := range docs {
		fmt.Println("\ndoc:", i)
		fmt.Println(doc)
		path := "Envelope.Body.GetClaimStatusCodesResponse.GetClaimStatusCodesResult.ClaimStatusCodeRecord"
		av, err := x2j.ValuesFromTagPath(doc, path)
		if err != nil {
			fmt.Println("err:", err.Error())
			return
		}
		if av == nil {
			fmt.Println("path:", path)
			fmt.Println("No ClaimStatusCodesResult code records.")
			continue
		}
		fmt.Println("\nPath:", path)
		fmt.Println("Number of code records:", len(av))
		fmt.Println("av:", av, "\n")
		for _, v := range av {
			switch v.(type) {
			case map[string]interface{}:
				fmt.Println("map[string]interface{}:", v.(map[string]interface{}))
			case []map[string]interface{}:
				fmt.Println("[]map[string]interface{}:", v.([]map[string]interface{}))
			case []interface{}:
				fmt.Println("[]interface{}:", v.([]interface{}))
			case interface{}:
				fmt.Println("interface{}:", v.(interface{}))
			}
		}
	}
}

func partPath1(doc string) {
	path := "Envelope.Body.*.*.ClaimStatusCodeRecord"
	av, err := x2j.ValuesFromTagPath(doc, path)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	if av == nil {
		fmt.Println("path:", path)
		fmt.Println("No ClaimStatusCodesResult code records.")
		return
	}
	fmt.Println("\nPath:", path)
	fmt.Println("Number of code records:", len(av))
	fmt.Println("av:", av, "\n")
}

func partPath2(doc string) {
	path := "Envelope.Body.*.*.*"
	av, err := x2j.ValuesFromTagPath(doc, path)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	if av == nil {
		fmt.Println("path:", path)
		fmt.Println("No ClaimStatusCodesResult code records.")
		return
	}
	fmt.Println("\nPath:", path)
	fmt.Println("Number of code records:", len(av))
	fmt.Println("av:", av, "\n")
}

func partPath3(doc string) {
	path := "*.*.*.*.*"
	av, err := x2j.ValuesFromTagPath(doc, path)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	if av == nil {
		fmt.Println("path:", path)
		fmt.Println("No ClaimStatusCodesResult code records.")
		return
	}
	fmt.Println("\nPath:", path)
	fmt.Println("Number of code records:", len(av))
	fmt.Println("av:", av, "\n")
}

func partPath4(doc string) {
	path := "*.*.*.*.*.Description"
	av, err := x2j.ValuesFromTagPath(doc, path)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	if av == nil {
		fmt.Println("path:", path)
		fmt.Println("No ClaimStatusCodesResult code records.")
		return
	}
	fmt.Println("\nPath:", path)
	fmt.Println("Number of code records:", len(av))
	fmt.Println("av:", av, "\n")
}

func partPath5(doc string) {
	path := "*.*.*.*.*.*"
	av, err := x2j.ValuesFromTagPath(doc, path)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	if av == nil {
		fmt.Println("path:", path)
		fmt.Println("No ClaimStatusCodesResult code records.")
		return
	}
	fmt.Println("\nPath:", path)
	fmt.Println("Number of code records:", len(av))
	fmt.Println("av:", av, "\n")
}

func partPath6(doc string) {
	path := "*.*.*.*.*.*.*"
	av, err := x2j.ValuesFromTagPath(doc, path)
	if err != nil {
		fmt.Println("err:", err.Error())
		return
	}
	if av == nil {
		fmt.Println("path:", path)
		fmt.Println("No ClaimStatusCodesResult code records.")
		return
	}
	fmt.Println("\nPath:", path)
	fmt.Println("Number of code records:", len(av))
	fmt.Println("av:", av, "\n")
}
