package main

import (
	"github.com/andeya/pholcus/exec"
	_ "github.com/andeya/pholcus/sample/static_rules"
	// _ "github.com/andeya/pholcus/sample/static_rules_pte" // you can also add your own rule library
)

func main() {
	// set default runtime UI and start
	// before running, set -a_ui to "web", "gui" or "cmd" to specify the UI
	// "gui" is Windows only
	exec.DefaultRun("web")
}
