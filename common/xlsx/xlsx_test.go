package xlsx

import (
	"bytes"
	"testing"
	"time"
)

func TestNewFile(t *testing.T) {
	f := NewFile()
	if f == nil {
		t.Fatal("NewFile returned nil")
	}
	if f.Sheet == nil || f.Sheets == nil {
		t.Error("NewFile should initialize Sheet and Sheets")
	}
}

func TestAddSheet(t *testing.T) {
	f := NewFile()
	r := f.AddSheet("Sheet1")
	if r.IsErr() {
		t.Fatal(r.UnwrapErr())
	}
	s := r.Unwrap()
	if s.Name != "Sheet1" {
		t.Errorf("sheet name = %s", s.Name)
	}
	if len(f.Sheets) != 1 {
		t.Errorf("len(Sheets) = %d", len(f.Sheets))
	}

	r2 := f.AddSheet("Sheet1")
	if !r2.IsErr() {
		t.Error("duplicate sheet name should fail")
	}
}

func TestAddRowAddCell(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Test").Unwrap()
	row := s.AddRow()
	cell := row.AddCell()
	cell.SetString("hello")
	if cell.String() != "hello" {
		t.Errorf("cell value = %s", cell.String())
	}
}

func TestWriteAndOpenBinary(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Data").Unwrap()
	row := s.AddRow()
	row.AddCell().SetString("A1")
	row.AddCell().SetInt(42)
	row.AddCell().SetBool(true)

	var buf bytes.Buffer
	r := f.Write(&buf)
	if r.IsErr() {
		t.Fatal(r.UnwrapErr())
	}

	data := buf.Bytes()
	opened := OpenBinary(data)
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	file := opened.Unwrap()
	if len(file.Sheets) != 1 {
		t.Fatalf("sheets count = %d", len(file.Sheets))
	}
	sheet := file.Sheets[0]
	if len(sheet.Rows) != 1 {
		t.Fatalf("rows count = %d", len(sheet.Rows))
	}
	r0 := sheet.Rows[0]
	if len(r0.Cells) != 3 {
		t.Fatalf("cells count = %d", len(r0.Cells))
	}
	if r0.Cells[0].String() != "A1" {
		t.Errorf("A1 = %s", r0.Cells[0].String())
	}
	if r0.Cells[1].String() != "42" {
		t.Errorf("B1 = %s", r0.Cells[1].String())
	}
	if r0.Cells[2].String() != "1" {
		t.Errorf("C1 = %s", r0.Cells[2].String())
	}
}

func TestToSlice(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("S1").Unwrap()
	row := s.AddRow()
	row.AddCell().SetString("x")
	row.AddCell().SetFloat(3.14)

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	slice := opened.Unwrap().ToSlice()
	if slice.IsErr() {
		t.Fatal(slice.UnwrapErr())
	}
	data := slice.Unwrap()
	if len(data) != 1 || len(data[0]) != 1 || len(data[0][0]) != 2 {
		t.Errorf("ToSlice shape: %v", data)
	}
	if data[0][0][0] != "x" || data[0][0][1] != "3.14" {
		t.Errorf("ToSlice values: %v", data[0][0])
	}
}

func TestCellTypes(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Cells").Unwrap()
	row := s.AddRow()
	row.AddCell().SetInt64(12345)
	row.AddCell().SetDate(time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
	row.AddCell().SetValue("str")
	row.AddCell().SetValue(100)

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestWriteSlice(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Slice").Unwrap()
	row := s.AddRow()
	slice := []string{"a", "b", "c"}
	n := row.WriteSlice(&slice, -1)
	if n != 3 {
		t.Errorf("WriteSlice: got %d", n)
	}
}

func TestWriteStruct(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Struct").Unwrap()
	row := s.AddRow()
	type rec struct {
		Name  string
		Count int
		Rate  float64
	}
	r := rec{Name: "x", Count: 10, Rate: 1.5}
	n := row.WriteStruct(&r, -1)
	if n != 3 {
		t.Errorf("WriteStruct: got %d", n)
	}
}

func TestMultipleSheets(t *testing.T) {
	f := NewFile()
	f.AddSheet("S1").Unwrap()
	f.AddSheet("S2").Unwrap()
	row := f.Sheets[1].AddRow()
	row.AddCell().SetString("B2")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	file := opened.Unwrap()
	if len(file.Sheets) != 2 {
		t.Errorf("sheets = %d", len(file.Sheets))
	}
}

func TestSetColWidth(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Cols").Unwrap()
	s.SetColWidth(0, 2, 15.0)
	row := s.AddRow()
	row.AddCell().SetString("x")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestCellStyle(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Style").Unwrap()
	row := s.AddRow()
	cell := row.AddCell()
	cell.SetString("styled")
	cell.GetStyle().Font.Bold = true

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestCellFormula(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Formula").Unwrap()
	row := s.AddRow()
	row.AddCell().SetInt(1)
	row.AddCell().SetInt(2)
	cell := row.AddCell()
	cell.SetFormula("A1+B1")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestRowHeight(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Height").Unwrap()
	row := s.AddRow()
	row.SetHeightCM(1.5)
	row.AddCell().SetString("x")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestCellMerge(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Merge").Unwrap()
	row := s.AddRow()
	cell := row.AddCell()
	cell.SetString("merged")
	cell.Merge(1, 1)

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestSheetCell(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Cell").Unwrap()
	s.AddRow().AddCell().SetString("A1")
	cell := s.Cell(0, 0)
	if cell.String() != "A1" {
		t.Errorf("Cell(0,0) = %s", cell.String())
	}
	empty := s.Cell(5, 5)
	if empty.String() != "" {
		t.Errorf("empty cell = %s", empty.String())
	}
}

func TestOpenBinaryInvalid(t *testing.T) {
	r := OpenBinary([]byte("not xlsx"))
	if !r.IsErr() {
		t.Error("invalid xlsx should fail")
	}
}

func TestFileToSliceInvalidPath(t *testing.T) {
	r := FileToSlice("/nonexistent/path.xlsx")
	if !r.IsErr() {
		t.Error("FileToSlice invalid path should fail")
	}
}

func TestFileToSliceValid(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Data").Unwrap()
	s.AddRow().AddCell().SetString("v1")
	tmp := t.TempDir() + "/slice.xlsx"
	if f.Save(tmp).IsErr() {
		t.Fatal("Save failed")
	}
	r := FileToSlice(tmp)
	if r.IsErr() {
		t.Fatal(r.UnwrapErr())
	}
	data := r.Unwrap()
	if len(data) != 1 || data[0][0][0] != "v1" {
		t.Errorf("FileToSlice: %v", data)
	}
}

func TestWriteSliceInvalid(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("X").Unwrap()
	row := s.AddRow()
	var notSlice string = "x"
	n := row.WriteSlice(&notSlice, 1)
	if n != -1 {
		t.Errorf("WriteSlice invalid: got %d", n)
	}
}

func TestWriteStructInvalid(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("X").Unwrap()
	row := s.AddRow()
	var notStruct string = "x"
	n := row.WriteStruct(&notStruct, 1)
	if n != -1 {
		t.Errorf("WriteStruct invalid: got %d", n)
	}
}

func TestSetColWidthInvalid(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("X").Unwrap()
	err := s.SetColWidth(2, 0, 10.0)
	if err == nil {
		t.Error("SetColWidth invalid range should fail")
	}
}

func TestCellBool(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("B").Unwrap()
	row := s.AddRow()
	row.AddCell().SetBool(true)
	row.AddCell().SetBool(false)

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	sheet := opened.Unwrap().Sheets[0]
	c0 := sheet.Rows[0].Cells[0]
	c1 := sheet.Rows[0].Cells[1]
	if !c0.Bool() || c1.Bool() {
		t.Errorf("Bool: %v %v", c0.Bool(), c1.Bool())
	}
}

func TestCellFloat(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("F").Unwrap()
	row := s.AddRow()
	row.AddCell().SetFloat(3.14159)

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	cell := opened.Unwrap().Sheets[0].Rows[0].Cells[0]
	fv, err := cell.Float()
	if err != nil || fv < 3.14 || fv > 3.15 {
		t.Errorf("Float: %v %v", fv, err)
	}
}

func TestCellInt(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("I").Unwrap()
	row := s.AddRow()
	row.AddCell().SetInt(999)

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	cell := opened.Unwrap().Sheets[0].Rows[0].Cells[0]
	iv, err := cell.Int()
	if err != nil || iv != 999 {
		t.Errorf("Int: %v %v", iv, err)
	}
}

func TestMultipleRowsWithGaps(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Gaps").Unwrap()
	s.AddRow().AddCell().SetString("R1")
	s.AddRow().AddCell().SetString("R2")
	row3 := s.AddRow()
	row3.AddCell().SetString("A")
	row3.AddCell().SetString("B")
	row3.AddCell().SetString("C")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	file := opened.Unwrap()
	slice := file.ToSlice()
	if slice.IsErr() {
		t.Fatal(slice.UnwrapErr())
	}
	data := slice.Unwrap()
	if len(data[0]) != 3 {
		t.Errorf("rows = %d", len(data[0]))
	}
	if data[0][2][2] != "C" {
		t.Errorf("R3C3 = %s", data[0][2][2])
	}
}

func TestCellFormulaRoundtrip(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("F").Unwrap()
	row := s.AddRow()
	row.AddCell().SetInt(10)
	row.AddCell().SetInt(20)
	cell := row.AddCell()
	cell.SetFormula("A1+B1")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestColAccess(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Col").Unwrap()
	row := s.AddRow()
	row.AddCell().SetString("a")
	row.AddCell().SetString("b")
	col := s.Col(1)
	if col == nil {
		t.Fatal("Col(1) nil")
	}
}

func TestCellSetValueTypes(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("V").Unwrap()
	row := s.AddRow()
	row.AddCell().SetValue(int32(1))
	row.AddCell().SetValue(float32(2.5))
	row.AddCell().SetValue([]byte("bytes"))

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestSheetFormat(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Fmt").Unwrap()
	s.SheetFormat.DefaultColWidth = 12.0
	s.SheetFormat.DefaultRowHeight = 18.0
	s.AddRow().AddCell().SetString("x")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestCellFormattedValue(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Fmt").Unwrap()
	row := s.AddRow()
	row.AddCell().SetInt(100)
	row.AddCell().SetFloat(2.5)

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	c0 := opened.Unwrap().Sheets[0].Rows[0].Cells[0]
	c1 := opened.Unwrap().Sheets[0].Rows[0].Cells[1]
	if c0.FormattedValue() != "100" {
		t.Errorf("FormattedValue int = %s", c0.FormattedValue())
	}
	if c1.FormattedValue() == "" {
		t.Errorf("FormattedValue float empty")
	}
}

func TestSaveToTempFile(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("Save").Unwrap()
	s.AddRow().AddCell().SetString("test")

	tmp := t.TempDir() + "/test.xlsx"
	r := f.Save(tmp)
	if r.IsErr() {
		t.Fatal(r.UnwrapErr())
	}
	opened := OpenFile(tmp)
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	file := opened.Unwrap()
	if file.Sheets[0].Rows[0].Cells[0].String() != "test" {
		t.Errorf("saved file content wrong")
	}
}

func TestWriteSliceWithLimit(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("L").Unwrap()
	row := s.AddRow()
	slice := []string{"a", "b", "c", "d"}
	n := row.WriteSlice(&slice, 2)
	if n != 2 {
		t.Errorf("WriteSlice limit: got %d", n)
	}
}

func TestWriteStructWithLimit(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("L").Unwrap()
	row := s.AddRow()
	type R struct {
		A, B, C string
	}
	r := R{"a", "b", "c"}
	n := row.WriteStruct(&r, 2)
	if n != 2 {
		t.Errorf("WriteStruct limit: got %d", n)
	}
}

func TestDateCellFormattedValue(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("D").Unwrap()
	row := s.AddRow()
	cell := row.AddCell()
	cell.SetDate(time.Date(2025, 3, 15, 0, 0, 0, 0, time.UTC))

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	c := opened.Unwrap().Sheets[0].Rows[0].Cells[0]
	fv := c.FormattedValue()
	if fv == "" || len(fv) < 8 {
		t.Errorf("FormattedValue date = %s", fv)
	}
}

func TestCellType(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("T").Unwrap()
	row := s.AddRow()
	row.AddCell().SetString("s")
	row.AddCell().SetInt(1)
	row.AddCell().SetBool(true)
	if s.Cell(0, 0).Type() != CellTypeString {
		t.Error("CellType string")
	}
	if s.Cell(0, 1).Type() != CellTypeNumeric {
		t.Error("CellType numeric")
	}
	if s.Cell(0, 2).Type() != CellTypeBool {
		t.Error("CellType bool")
	}
}

func TestCellInt64(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("I64").Unwrap()
	row := s.AddRow()
	row.AddCell().SetInt64(9876543210)

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	c := opened.Unwrap().Sheets[0].Rows[0].Cells[0]
	v, err := c.Int64()
	if err != nil || v != 9876543210 {
		t.Errorf("Int64: %v %v", v, err)
	}
}

func TestCellSetDateTime(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("DT").Unwrap()
	row := s.AddRow()
	row.AddCell().SetDateTime(time.Date(2025, 1, 15, 14, 30, 0, 0, time.UTC))

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestColSetType(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("CT").Unwrap()
	s.AddRow().AddCell().SetString("x")
	col := s.Col(0)
	col.SetType(CellTypeNumeric)
	_ = col.GetStyle()
	col.SetStyle(NewStyle())
}

func TestCellSafeFormattedValue(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("S").Unwrap()
	row := s.AddRow()
	row.AddCell().SetFloatWithFormat(1.5, "#,##0.00")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	c := opened.Unwrap().Sheets[0].Rows[0].Cells[0]
	_, err := c.SafeFormattedValue()
	if err != nil {
		t.Errorf("SafeFormattedValue: %v", err)
	}
}

func TestXLSXReaderError(t *testing.T) {
	e := &XLSXReaderError{Err: "test error"}
	if e.Error() != "test error" {
		t.Errorf("XLSXReaderError.Error = %s", e.Error())
	}
}

func TestFileDate1904(t *testing.T) {
	f := NewFile()
	f.Date1904 = true
	s := f.AddSheet("D").Unwrap()
	row := s.AddRow()
	row.AddCell().SetDate(time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC))

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestWriteSliceInterfaceElement(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("I").Unwrap()
	row := s.AddRow()
	slice := []interface{}{"a", 1, true}
	n := row.WriteSlice(&slice, -1)
	if n != 3 {
		t.Errorf("WriteSlice interface: got %d", n)
	}
}

func TestWriteStructDefaultField(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("D").Unwrap()
	row := s.AddRow()
	type R struct {
		A int
		B complex64
		C string
	}
	r := R{A: 1, C: "x"}
	n := row.WriteStruct(&r, -1)
	if n != 2 {
		t.Errorf("WriteStruct with default: got %d", n)
	}
}

func TestCellPercentFormat(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("P").Unwrap()
	row := s.AddRow()
	row.AddCell().SetFloatWithFormat(0.5, "0%")
	row.AddCell().SetFloatWithFormat(0.1234, "0.00%")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	c0 := opened.Unwrap().Sheets[0].Rows[0].Cells[0]
	c1 := opened.Unwrap().Sheets[0].Rows[0].Cells[1]
	if c0.FormattedValue() == "" {
		t.Error("percent format empty")
	}
	if c1.FormattedValue() == "" {
		t.Error("percent format empty")
	}
}

func TestCellScientificFormat(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("E").Unwrap()
	row := s.AddRow()
	row.AddCell().SetFloatWithFormat(1234.5, "0.00e+00")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	c := opened.Unwrap().Sheets[0].Rows[0].Cells[0]
	if c.FormattedValue() == "" {
		t.Error("scientific format empty")
	}
}

func TestCellNegativeFormat(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("N").Unwrap()
	row := s.AddRow()
	row.AddCell().SetFloatWithFormat(-100, "#,##0 ;(#,##0)")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	c := opened.Unwrap().Sheets[0].Rows[0].Cells[0]
	if c.FormattedValue() == "" {
		t.Error("negative format empty")
	}
}

func TestHiddenSheet(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("H").Unwrap()
	s.Hidden = true
	s.AddRow().AddCell().SetString("x")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestSelectedSheet(t *testing.T) {
	f := NewFile()
	f.AddSheet("S1").Unwrap()
	s2 := f.AddSheet("S2").Unwrap()
	s2.Selected = true
	s2.AddRow().AddCell().SetString("x")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestOpenReaderAt(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("R").Unwrap()
	s.AddRow().AddCell().SetString("data")

	var buf bytes.Buffer
	f.Write(&buf)
	data := buf.Bytes()
	r := bytes.NewReader(data)
	opened := OpenReaderAt(r, int64(len(data)))
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	file := opened.Unwrap()
	if file.Sheets[0].Rows[0].Cells[0].String() != "data" {
		t.Error("OpenReaderAt content wrong")
	}
}

func TestCellFormulaGetter(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("F").Unwrap()
	row := s.AddRow()
	cell := row.AddCell()
	cell.SetFormula("=SUM(A1:A10)")
	if cell.Formula() != "=SUM(A1:A10)" {
		t.Errorf("Formula = %s", cell.Formula())
	}
}

func TestCellGetNumberFormat(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("N").Unwrap()
	row := s.AddRow()
	row.AddCell().SetFloatWithFormat(1.0, "0.00")
	c := s.Cell(0, 0)
	if c.GetNumberFormat() != "0.00" {
		t.Errorf("GetNumberFormat = %s", c.GetNumberFormat())
	}
}

func TestColHidden(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("H").Unwrap()
	s.SetColWidth(0, 0, 10)
	s.Cols[0].Hidden = true
	s.AddRow().AddCell().SetString("x")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}

func TestRowHidden(t *testing.T) {
	f := NewFile()
	s := f.AddSheet("H").Unwrap()
	row := s.AddRow()
	row.Hidden = true
	row.AddCell().SetString("x")

	var buf bytes.Buffer
	f.Write(&buf)
	opened := OpenBinary(buf.Bytes())
	if opened.IsErr() {
		t.Fatal(opened.UnwrapErr())
	}
	_ = opened.Unwrap()
}
