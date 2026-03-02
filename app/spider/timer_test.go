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
	bell := time.Now().Add(2 * time.Second)
	t.Log(ctx.SetTimer("id", 1, &Bell{bell.Hour(), bell.Minute(), bell.Second()}))
	t.Log(ctx.RunTimer("id"))
	t.Log(time.Now())
}
