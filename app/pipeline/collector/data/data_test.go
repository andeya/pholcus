package data

import (
	"testing"
)

func TestGetDataCell(t *testing.T) {
	d := map[string]interface{}{"key": "value"}
	cell := GetDataCell("rule1", d, "http://example.com", "http://parent.com", "2024-01-01")

	if cell[FieldRuleName] != "rule1" {
		t.Errorf("RuleName = %v, want %q", cell[FieldRuleName], "rule1")
	}
	if cell[FieldURL] != "http://example.com" {
		t.Errorf("Url = %v, want %q", cell[FieldURL], "http://example.com")
	}
	if cell[FieldParentURL] != "http://parent.com" {
		t.Errorf("ParentUrl = %v, want %q", cell[FieldParentURL], "http://parent.com")
	}
	if cell[FieldDownloadTime] != "2024-01-01" {
		t.Errorf("DownloadTime = %v, want %q", cell[FieldDownloadTime], "2024-01-01")
	}
	data := cell["Data"].(map[string]interface{})
	if data["key"] != "value" {
		t.Errorf("Data[key] = %v, want %q", data["key"], "value")
	}
}

func TestGetFileCell(t *testing.T) {
	body := []byte("hello world")
	cell := GetFileCell("rule2", "test.txt", body)

	if cell[FieldRuleName] != "rule2" {
		t.Errorf("RuleName = %v, want %q", cell[FieldRuleName], "rule2")
	}
	if cell["Name"] != "test.txt" {
		t.Errorf("Name = %v, want %q", cell["Name"], "test.txt")
	}
	if string(cell["Bytes"].([]byte)) != "hello world" {
		t.Errorf("Bytes = %v, want %q", cell["Bytes"], "hello world")
	}
}

func TestPutDataCell(t *testing.T) {
	cell := GetDataCell("r", nil, "", "", "")
	PutDataCell(cell)
	if cell[FieldRuleName] != nil {
		t.Error("RuleName should be nil after Put")
	}
	if cell["Data"] != nil {
		t.Error("Data should be nil after Put")
	}
}

func TestPutFileCell(t *testing.T) {
	cell := GetFileCell("r", "f", []byte{1})
	PutFileCell(cell)
	if cell[FieldRuleName] != nil {
		t.Error("RuleName should be nil after Put")
	}
	if cell["Name"] != nil {
		t.Error("Name should be nil after Put")
	}
	if cell["Bytes"] != nil {
		t.Error("Bytes should be nil after Put")
	}
}

func TestPoolReuseDataCell(t *testing.T) {
	c1 := GetDataCell("a", nil, "", "", "")
	PutDataCell(c1)
	c2 := GetDataCell("b", nil, "", "", "")
	if c2[FieldRuleName] != "b" {
		t.Errorf("reused cell RuleName = %v, want %q", c2[FieldRuleName], "b")
	}
}
