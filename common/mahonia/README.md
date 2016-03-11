mahonia
=======

 character-set conversion library implemented in Go.

 Mahonia is a character-set conversion library implemented in Go.
 All data is compiled into the executable; it doesn't need any external data files.

 Copy from http://code.google.com/p/mahonia/

install
-------

  go get github.com/henrylee2cn/mahonia

example
-------

```
  package main
  import "fmt"
  import "github.com/henrylee2cn/mahonia"
  func main(){
    enc:=mahonia.NewEncoder("gbk")
    //converts a  string from UTF-8 to gbk encoding.
    fmt.Println(enc.ConvertString("hello,世界"))  
  }
```
