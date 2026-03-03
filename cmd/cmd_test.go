package cmd

import (
	"flag"
	"testing"
)

func TestFlag(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet("cmd_test", flag.ContinueOnError)
	Flag()
	if spiderflag == nil {
		t.Error("spiderflag not set")
	}
}
