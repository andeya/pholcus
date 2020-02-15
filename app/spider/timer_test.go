package spider

import (
	"testing"
	"time"
)

func TestTimer1(t *testing.T) {
	t.Log(time.Now())
	ctx := GetContext(new(Spider), nil)
	t.Log(ctx.SetTimer("id", 3*time.Second, nil))
	t.Log(ctx.RunTimer("id"))
	t.Log(ctx.RunTimer("id"))
	t.Log(time.Now())
}

func TestTimer2(t *testing.T) {
	t.Log(time.Now())
	ctx := GetContext(new(Spider), nil)
	t.Log(ctx.SetTimer("id", 2, &Bell{13, 22, 0}))
	t.Log(ctx.RunTimer("id"))
	t.Log(time.Now())
}
