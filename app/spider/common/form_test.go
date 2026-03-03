package common

import (
	"strings"
	"testing"

	spider "github.com/andeya/pholcus/app/spider"
	"github.com/andeya/pholcus/common/goquery"
)

func TestNewForm(t *testing.T) {
	html := `<form action="/search" method="get">
		<input name="q" value="x" />
		<input type="submit" name="btn" value="go" />
	</form>`
	doc := goquery.NewDocumentFromReader(strings.NewReader(html))
	if doc.IsErr() {
		t.Fatalf("parse html: %v", doc.UnwrapErr())
	}
	formSel := doc.Unwrap().Find("form").First()
	sp := &spider.Spider{RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}}}
	ctx := spider.GetContext(sp, nil)
	defer spider.PutContext(ctx)

	form := NewForm(ctx, "r1", "https://example.com/page", formSel)
	if form == nil {
		t.Fatal("NewForm returned nil")
	}
	if form.Method() != "GET" {
		t.Errorf("Method() = %q, want GET", form.Method())
	}
	if form.Action() != "https://example.com/search" {
		t.Errorf("Action() = %q, want https://example.com/search", form.Action())
	}
	if form.Dom() != formSel {
		t.Error("Dom() mismatch")
	}
	form.Input("q", "newval")
	form.Inputs(map[string]string{"q": "v2"})
	if !form.Submit() {
		t.Error("Submit() = false")
	}
	if !form.Click("btn") {
		t.Error("Click(btn) = false")
	}
	if form.Click("nonexistent") {
		t.Error("Click(nonexistent) should be false")
	}
}

func TestNewForm_Post(t *testing.T) {
	html := `<form action="/submit" method="post">
		<input name="a" value="1" />
	</form>`
	doc := goquery.NewDocumentFromReader(strings.NewReader(html))
	if doc.IsErr() {
		t.Fatalf("parse html: %v", doc.UnwrapErr())
	}
	formSel := doc.Unwrap().Find("form").First()
	sp := &spider.Spider{RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}}}
	ctx := spider.GetContext(sp, nil)
	defer spider.PutContext(ctx)

	form := NewForm(ctx, "r1", "https://example.com/", formSel)
	if form == nil {
		t.Fatal("NewForm returned nil")
	}
	if form.Method() != "POST" {
		t.Errorf("Method() = %q, want POST", form.Method())
	}
	form.Submit()
}

func TestNewForm_Multipart(t *testing.T) {
	html := `<form action="/upload" method="post" enctype="multipart/form-data">
		<input name="f" value="x" />
	</form>`
	doc := goquery.NewDocumentFromReader(strings.NewReader(html))
	if doc.IsErr() {
		t.Fatalf("parse html: %v", doc.UnwrapErr())
	}
	formSel := doc.Unwrap().Find("form").First()
	sp := &spider.Spider{RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}}}
	ctx := spider.GetContext(sp, nil)
	defer spider.PutContext(ctx)

	form := NewForm(ctx, "r1", "https://example.com/", formSel)
	if form == nil {
		t.Fatal("NewForm returned nil")
	}
	form.Submit()
}

func TestNewForm_NoSchemeAndHost(t *testing.T) {
	html := `<form action="/path">
		<input name="x" value="1" />
	</form>`
	doc := goquery.NewDocumentFromReader(strings.NewReader(html))
	if doc.IsErr() {
		t.Fatalf("parse html: %v", doc.UnwrapErr())
	}
	formSel := doc.Unwrap().Find("form").First()
	sp := &spider.Spider{RuleTree: &spider.RuleTree{Trunk: map[string]*spider.Rule{}}}
	ctx := spider.GetContext(sp, nil)
	defer spider.PutContext(ctx)

	form := NewForm(ctx, "r1", "https://example.com/page", formSel)
	if form == nil {
		t.Fatal("NewForm returned nil")
	}
	if form.Action() != "https://example.com/path" {
		t.Errorf("Action() = %q", form.Action())
	}
}
