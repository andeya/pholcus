package x2j

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestX2j(t *testing.T) {
	fi, fierr := os.Stat("x2j_test.xml")
	if fierr != nil {
		fmt.Println("fierr:",fierr.Error())
		return
	}
	fh, fherr := os.Open("x2j_test.xml")
	if fherr != nil {
		fmt.Println("fherr:",fherr.Error())
		return
	}
	defer fh.Close()
	buf := make([]byte,fi.Size())
	_, nerr  :=  fh.Read(buf)
	if nerr != nil {
		fmt.Println("nerr:",nerr.Error())
		return
	}
	doc := string(buf)
	fmt.Println("\nXML doc:\n",doc)

	root, berr := DocToTree(doc)
	if berr != nil {
		fmt.Println("berr:",berr.Error())
		fmt.Println("root:",root)
		return
	}
	fmt.Println("\nDocToTree():\n",root.WriteTree())

	m := make(map[string]interface{})
	m[root.key] = root.treeToMap(false)
	fmt.Println("\ntreeToMap, recast==false:\n",WriteMap(m))

	j,jerr := json.MarshalIndent(m,"","  ")
	if jerr != nil {
		fmt.Println("jerr:",jerr.Error())
	}
	fmt.Println("\njson.MarshalIndent, recast==false:\n",string(j))

	// test DocToMap() with recast
	mm, mmerr := DocToMap(doc,true)
	if mmerr != nil {
		println("mmerr:",mmerr.Error())
		return
	}
	println("\nDocToMap(), recast==true:\n",WriteMap(mm))

	// test DocToJsonIndent() with recast
	s,serr := DocToJsonIndent(doc,true)
	if serr != nil {
		fmt.Println("serr:",serr.Error())
	}
	fmt.Println("\nDocToJsonIndent, recast==true:\n",s)
}

func TestGetValue(t *testing.T) {
	// test MapValue()
	doc := `<entry><vars><foo>bar</foo><foo2><hello>world</hello></foo2></vars></entry>`
	fmt.Println("\nRead doc:",doc)
	fmt.Println("Looking for value: entry.vars")
	mm,mmerr := DocToMap(doc)
	if mmerr != nil {
		fmt.Println("merr:",mmerr.Error())
	}
	v,verr := MapValue(mm,"entry.vars",nil)
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}
	fmt.Println("Looking for value: entry.vars.foo2.hello")
	v,verr = MapValue(mm,"entry.vars.foo2.hello",nil)
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		fmt.Println(v.(string))
	}
	fmt.Println("Looking with error in path: entry.var")
	v,verr = MapValue(mm,"entry.var",nil)
	fmt.Println("verr:",verr.Error())

	// test DocValue()
	fmt.Println("DocValue() for tag path entry.vars")
	v,verr = DocValue(doc,"entry.vars")
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	}
	j,_ := json.MarshalIndent(v,"","  ")
	fmt.Println(string(j))
}


func TestGetValueWithAttr(t *testing.T) {
	doc := `<entry><vars>
		<foo item="1">bar</foo>
		<foo item="2">
			<hello item="3">world</hello>
			<hello item="4">universe</hello>
		</foo></vars></entry>`
	fmt.Println("\nRead doc:",doc)
	fmt.Println("Looking for value: entry.vars")
	mm,mmerr := DocToMap(doc)
	if mmerr != nil {
		fmt.Println("merr:",mmerr.Error())
	}
	v,verr := MapValue(mm,"entry.vars",nil)
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}

	fmt.Println("\nMapValue(): Looking for value: entry.vars.foo item=2")
	a,aerr := NewAttributeMap("item:2")
	if aerr != nil {
		fmt.Println("aerr:",aerr.Error())
	}
	v,verr = MapValue(mm,"entry.vars.foo",a)
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}

	fmt.Println("\nMapValue(): Looking for hello item:4")
	a,_ = NewAttributeMap("item:4")
	v,verr = MapValue(mm,"hello",a)
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}

	fmt.Println("\nDocValue(): Looking for entry.vars.foo.hello item:4")
	v,verr = DocValue(doc,"entry.vars.foo.hello","item:4")
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}

	fmt.Println("\nDocValue(): Looking for empty nil")
	v,verr = DocValue(doc,"")
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}

	// test 'recast' switch
	fmt.Println("\ntesting recast switch...")
	mm,mmerr = DocToMap(doc,true)
	if mmerr != nil {
		fmt.Println("merr:",mmerr.Error())
	}
	fmt.Println("MapValue(): Looking for value: entry.vars.foo item=2")
	a,aerr = NewAttributeMap("item:2")
	if aerr != nil {
		fmt.Println("aerr:",aerr.Error())
	}
	v,verr = MapValue(mm,"entry.vars.foo",a,true)
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}
}

func TestStuff_1(t *testing.T) {
	doc := `<doc>
				<tag item="1">val2</tag>
				<tag item="2">val2</tag>
				<tag item="2" instance="2">val3</tag>
			</doc>`

	fmt.Println(doc)
	m,merr := DocToMap(doc)
	if merr != nil {
		fmt.Println("merr:",merr.Error())
	} else {
		fmt.Println(WriteMap(m))
	}

	fmt.Println("\nDocValue(): tag")
	v,verr := DocValue(doc,"doc.tag")
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}

	fmt.Println("\nDocValue(): item:2 instance:2")
	v,verr = DocValue(doc,"doc.tag","item:2","instance:2")
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}
}

func TestStuff_2(t *testing.T) {
	doc := `<tag item="1">val2</tag>
				<tag item="2">val2</tag>
				<tag item="2" instance="2">val3</tag>`

	fmt.Println(doc)
	m,merr := DocToMap(doc)
	if merr != nil {
		fmt.Println("merr:",merr.Error())
	} else {
		fmt.Println(WriteMap(m))
	}

	fmt.Println("\nDocValue(): tag")
	v,verr := DocValue(doc,"tag")
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}

	fmt.Println("\nDocValue(): item:2 instance:2")
	v,verr = DocValue(doc,"tag","item:2","instance:2")
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	} else {
		j, jerr := json.MarshalIndent(v,"","  ")
		if jerr != nil {
			fmt.Println("jerr:",jerr.Error())
		} else {
			fmt.Println(string(j))
		}
	}
}

func procMap(m map[string]interface{}) bool {
	fmt.Println("procMap:",WriteMap(m),"\n")
	return true
}

func procMapToJson(m map[string]interface{}) bool {
	b,_ := json.MarshalIndent(m,"","  ")
	fmt.Println("procMap:",string(b),"\n")
	return true
}

func procErr(err error) bool {
	fmt.Println("procError err:",err.Error())
	return true
}

func TestBulk(t *testing.T) {
	fmt.Println("\nBulk Message Processing Tests")
	// if err := XmlMsgsFromFile("x2m_bulk.xml",procMap,procErr); err != nil {
	if err := XmlMsgsFromFile("x2m_bulk.xml",procMapToJson,procErr); err != nil {
		fmt.Println("XmlMsgsFromFile err:",err.Error())
	}
}

func TestBulkBuffer(t *testing.T) {
	fmt.Println("\nTest arbitrary XML buffer with no ending rootTag - then read from closed buffer.")
	doc := `<doc><tag /></doc><doc><one att="1">my test doc</one></doc><doc><tag>missing end element</tag>`
	b := NewXmlBuffer(doc)
	for {
		m,merr := b.NextMap()
		if merr != nil {
			fmt.Println("merr:",merr.Error())
		}
		if m == nil {
			break
		}
		fmt.Println("m:",m)
	}
	b.Close()
	_,err := b.NextMap()
	fmt.Println("err:",err.Error())
}


func TestTagAndKey(t *testing.T) {
	var doc string
	doc = `<doc>
		<sections>
			<section>one</section>
			<section>
				<parts>
					<part>two.one</part>
					<part>two.two</part>
				</parts>
			</section>
		</sections>
		<partitions>
			<parts>
				<sections>
					<section>one</section>
					<section>two</section>
				</sections>
			</parts>
		</partitions>	
	</doc>`

	fmt.Println("\nTestTagAndKey()\n",doc)
	v,verr := ValuesForTag(doc,"parts")
	if verr != nil {
		fmt.Println("verr:",verr.Error())
	}
	fmt.Println("tag: parts :: len:",len(v),"v:",v)
	v, _ = ValuesForTag(doc,"not_a_tag")
	if v == nil {
		fmt.Println("no 'not_a_tag' tag")
	} else {
		fmt.Println("key: not_a_tag :: len:",len(v),"v:",v)
	}

	m,merr := DocToMap(doc)
	if merr != nil {
		fmt.Println("merr:",merr.Error())
	}
	v = ValuesForKey(m,"section")
	fmt.Println("key: section :: len:",len(v),"v:",v)

	v = ValuesForKey(m,"not_a_key")
	if v == nil {
		fmt.Println("no 'not_a_key' key")
	} else {
		fmt.Println("key: not_a-key :: len:",len(v),"v:",v)
	}
}

// ---------------- x2j_fast.go ----------------
/*
func Test_F_DocToMap(t *testing.T) {
	var doc string = `<doc>
		<sections>
			<section>one</section>
			<section>
				<parts>
					<part>two.one</part>
					<part>two.two</part>
				</parts>
			</section>
		</sections>
		<partitions>
			<parts>
				<sections>
					<section>one</section>
					<section>two</section>
				</sections>
			</parts>
		</partitions>	
	</doc>`
	fmt.Println("\nF_DocToMap()")
	fmt.Println(doc)
	m,err := F_DocToMap(doc)
	if err != nil {
		fmt.Println("err:",err.Error())
	}
	fmt.Println(WriteMap(m),"\n")
}
*/
