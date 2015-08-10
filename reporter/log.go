package reporter

import (
	"fmt"
	"os"
	"time"
)

// 模仿标准包log打印报告
func Print(v ...interface{}) {
	fmt.Print(time.Now().Format("2006/01/02 15:04:05") + " " + fmt.Sprint(v...))
}

func Printf(format string, v ...interface{}) {
	fmt.Printf(time.Now().Format("2006/01/02 15:04:05")+" "+format+"\n", v...)
}

func Println(v ...interface{}) {
	fmt.Println(time.Now().Format("2006/01/02 15:04:05") + " " + fmt.Sprint(v...))
}

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	Print(v...)
	os.Exit(1)
}

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	Printf(format, v...)
	os.Exit(1)
}

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	Println(v...)
	os.Exit(1)
}
