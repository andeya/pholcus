package mahonia

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestCharset(t *testing.T) {
	cs := GetCharset("UTF-8")
	if cs == nil {
		t.Fatal("GetCharset UTF-8 returned nil")
	}
	if cs.Name != "UTF-8" {
		t.Errorf("GetCharset UTF-8 name = %s", cs.Name)
	}

	cs = GetCharset("gbk")
	if cs == nil {
		t.Fatal("GetCharset gbk returned nil")
	}
	if cs.Name != "GBK" {
		t.Errorf("GetCharset gbk name = %s", cs.Name)
	}

	if GetCharset("nonexistent") != nil {
		t.Error("GetCharset nonexistent should return nil")
	}
}

func TestSimplifyName(t *testing.T) {
	if simplifyName("UTF-8") != "utf8" {
		t.Errorf("simplifyName UTF-8 = %s", simplifyName("UTF-8"))
	}
}

func TestNewDecoderEncoder(t *testing.T) {
	dec := NewDecoder("UTF-8")
	if dec == nil {
		t.Fatal("NewDecoder UTF-8 returned nil")
	}
	enc := NewEncoder("UTF-8")
	if enc == nil {
		t.Fatal("NewEncoder UTF-8 returned nil")
	}

	if NewDecoder("nonexistent") != nil {
		t.Error("NewDecoder nonexistent should return nil")
	}
	if NewEncoder("nonexistent") != nil {
		t.Error("NewEncoder nonexistent should return nil")
	}
}

func TestConvertStringUTF8(t *testing.T) {
	dec := NewDecoder("UTF-8")
	enc := NewEncoder("UTF-8")

	s := "Hello 世界"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringGBK(t *testing.T) {
	dec := NewDecoder("GBK")
	enc := NewEncoder("GBK")

	s := "中文测试"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringOK(t *testing.T) {
	enc := NewEncoder("UTF-8")
	result, ok := enc.ConvertStringOK("abc")
	if !ok || result != "abc" {
		t.Errorf("ConvertStringOK: got %q ok=%v", result, ok)
	}

	dec := NewDecoder("UTF-8")
	result, ok = dec.ConvertStringOK("abc")
	if !ok || result != "abc" {
		t.Errorf("Decoder ConvertStringOK: got %q ok=%v", result, ok)
	}
}

func TestReader(t *testing.T) {
	dec := NewDecoder("UTF-8")
	r := dec.NewReader(strings.NewReader("Hello"))
	buf := make([]byte, 32)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got != "Hello" {
		t.Errorf("Reader Read: got %q", got)
	}
}

func TestReaderReadRune(t *testing.T) {
	dec := NewDecoder("UTF-8")
	r := dec.NewReader(strings.NewReader("A"))
	c, size, err := r.ReadRune()
	if err != nil {
		t.Fatal(err)
	}
	if c != 'A' || size != 1 {
		t.Errorf("ReadRune: got %c size=%d", c, size)
	}
}

func TestWriter(t *testing.T) {
	enc := NewEncoder("UTF-8")
	var buf bytes.Buffer
	w := enc.NewWriter(&buf)
	n, err := w.Write([]byte("Hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Errorf("Writer Write: n=%d", n)
	}
	if buf.String() != "Hello" {
		t.Errorf("Writer: got %q", buf.String())
	}
}

func TestWriterWriteRune(t *testing.T) {
	enc := NewEncoder("UTF-8")
	var buf bytes.Buffer
	w := enc.NewWriter(&buf)
	_, err := w.WriteRune('A')
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "A" {
		t.Errorf("WriteRune: got %q", buf.String())
	}
}

func TestStatusConstants(t *testing.T) {
	if SUCCESS != 0 {
		t.Errorf("SUCCESS = %d", SUCCESS)
	}
}

func TestDecoderConvertStringEmpty(t *testing.T) {
	dec := NewDecoder("UTF-8")
	if dec.ConvertString("") != "" {
		t.Error("empty string decode")
	}
}

func TestEncoderConvertStringEmpty(t *testing.T) {
	enc := NewEncoder("UTF-8")
	if enc.ConvertString("") != "" {
		t.Error("empty string encode")
	}
}

func TestReaderEOF(t *testing.T) {
	dec := NewDecoder("UTF-8")
	r := dec.NewReader(strings.NewReader(""))
	buf := make([]byte, 10)
	n, err := r.Read(buf)
	if n != 0 || err != io.EOF {
		t.Errorf("Read at EOF: n=%d err=%v", n, err)
	}
}

func TestReaderLargeRead(t *testing.T) {
	s := strings.Repeat("x", 5000)
	dec := NewDecoder("UTF-8")
	r := dec.NewReader(strings.NewReader(s))
	buf := make([]byte, 10000)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf[:n]) != s {
		t.Errorf("large read: got %d bytes", n)
	}
}

func TestConvertStringLongString(t *testing.T) {
	s := strings.Repeat("中", 1000)
	enc := NewEncoder("GBK")
	dec := NewDecoder("GBK")
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("long roundtrip: len=%d", len(decoded))
	}
}

func TestConvertStringBig5(t *testing.T) {
	dec := NewDecoder("Big5")
	enc := NewEncoder("Big5")
	s := "繁體中文"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("Big5 roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringGB2312(t *testing.T) {
	dec := NewDecoder("GB2312")
	enc := NewEncoder("GB2312")
	s := "简体"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("GB2312 roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringGB18030(t *testing.T) {
	dec := NewDecoder("GB18030")
	enc := NewEncoder("GB18030")
	s := "中文"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("GB18030 roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringISO8859_1(t *testing.T) {
	dec := NewDecoder("ISO-8859-1")
	enc := NewEncoder("ISO-8859-1")
	s := "café"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("ISO-8859-1 roundtrip: got %q want %q", decoded, s)
	}
}

func TestWriterPartialRune(t *testing.T) {
	enc := NewEncoder("UTF-8")
	var buf bytes.Buffer
	w := enc.NewWriter(&buf)
	w.Write([]byte{0xC0})
	w.Write([]byte("a"))
	if buf.String() != "\ufffda" {
		t.Errorf("partial rune: got %q", buf.String())
	}
}

func TestWriterWriteRuneMultiByte(t *testing.T) {
	enc := NewEncoder("UTF-8")
	var buf bytes.Buffer
	w := enc.NewWriter(&buf)
	w.WriteRune('世')
	if buf.String() != "世" {
		t.Errorf("WriteRune: got %q", buf.String())
	}
}

func TestReaderSmallBuffer(t *testing.T) {
	dec := NewDecoder("UTF-8")
	r := dec.NewReader(strings.NewReader("世界"))
	buf := make([]byte, 2)
	n, _ := r.Read(buf)
	if n > 0 && string(buf[:n]) != "世" {
		t.Errorf("small buffer: got %q", string(buf[:n]))
	}
}

func TestConvertStringOKInvalidUTF8(t *testing.T) {
	dec := NewDecoder("UTF-8")
	_, ok := dec.ConvertStringOK("a\xfe\xffb")
	if ok {
		t.Error("invalid UTF-8 should return ok=false")
	}
}

func TestSimplifyNameWithDigits(t *testing.T) {
	if simplifyName("ISO-8859-1") != "iso88591" {
		t.Errorf("simplifyName ISO-8859-1 = %s", simplifyName("ISO-8859-1"))
	}
}

func TestConvertStringUTF16(t *testing.T) {
	dec := NewDecoder("UTF-16")
	enc := NewEncoder("UTF-16")
	s := "Hello"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("UTF-16 roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringUTF16BE(t *testing.T) {
	dec := NewDecoder("UTF-16BE")
	enc := NewEncoder("UTF-16BE")
	s := "AB"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("UTF-16BE roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringUTF16LE(t *testing.T) {
	dec := NewDecoder("UTF-16LE")
	enc := NewEncoder("UTF-16LE")
	s := "AB"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("UTF-16LE roundtrip: got %q want %q", decoded, s)
	}
}

func TestReaderPartialUTF8AtEOF(t *testing.T) {
	dec := NewDecoder("UTF-8")
	r := dec.NewReader(strings.NewReader("\xE4"))
	buf := make([]byte, 32)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if n > 0 && !strings.Contains(string(buf[:n]), "\ufffd") {
		t.Errorf("expected replacement for partial UTF-8: got %q", string(buf[:n]))
	}
}

func TestEncoderConvertStringOKInvalidRune(t *testing.T) {
	enc := NewEncoder("GBK")
	_, ok := enc.ConvertStringOK("a\uFFFDb")
	if ok {
		t.Error("invalid rune in string should return ok=false")
	}
}

func TestDecoderConvertStringOKNoRoom(t *testing.T) {
	dec := NewDecoder("UTF-8")
	_, ok := dec.ConvertStringOK("\x80")
	if ok {
		t.Error("invalid UTF-8 should return ok=false")
	}
}

func TestConvertStringShiftJIS(t *testing.T) {
	dec := NewDecoder("Shift_JIS")
	enc := NewEncoder("Shift_JIS")
	s := "日本語"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("Shift_JIS roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringEUCJP(t *testing.T) {
	dec := NewDecoder("EUC-JP")
	enc := NewEncoder("EUC-JP")
	s := "日"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("EUC-JP roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringEUCKR(t *testing.T) {
	dec := NewDecoder("EUC-KR")
	enc := NewEncoder("EUC-KR")
	s := "한글"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("EUC-KR roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringASCII(t *testing.T) {
	dec := NewDecoder("ASCII")
	enc := NewEncoder("ASCII")
	s := "Hello"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("ASCII roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringLatin1(t *testing.T) {
	dec := NewDecoder("ISO-8859-1")
	enc := NewEncoder("Latin1")
	s := "café"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("Latin1 roundtrip: got %q want %q", decoded, s)
	}
}

func TestEncoderConvertStringNoRoom(t *testing.T) {
	enc := NewEncoder("UTF-16")
	s := strings.Repeat("x", 5000)
	result := enc.ConvertString(s)
	if len(result) < 10000 {
		t.Errorf("UTF-16 encode: got %d bytes", len(result))
	}
}

func TestConvertStringISO2022JP(t *testing.T) {
	dec := NewDecoder("ISO-2022-JP")
	enc := NewEncoder("ISO-2022-JP")
	s := "A"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("ISO-2022-JP roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringCP51932(t *testing.T) {
	dec := NewDecoder("CP51932")
	enc := NewEncoder("CP51932")
	s := "日"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("CP51932 roundtrip: got %q want %q", decoded, s)
	}
}

func TestConvertStringTCVN3(t *testing.T) {
	dec := NewDecoder("TCVN3")
	enc := NewEncoder("TCVN3")
	s := "a"
	encoded := enc.ConvertString(s)
	decoded := dec.ConvertString(encoded)
	if decoded != s {
		t.Errorf("TCVN3 roundtrip: got %q want %q", decoded, s)
	}
}

func TestReaderReadRuneEOF(t *testing.T) {
	dec := NewDecoder("UTF-8")
	r := dec.NewReader(strings.NewReader(""))
	_, _, err := r.ReadRune()
	if err != io.EOF {
		t.Errorf("ReadRune at EOF: got %v", err)
	}
}

func TestWriterBufferedWrite(t *testing.T) {
	enc := NewEncoder("UTF-8")
	var buf bytes.Buffer
	w := enc.NewWriter(&buf)
	w.Write([]byte("a"))
	w.Write([]byte("b"))
	if buf.String() != "ab" {
		t.Errorf("buffered write: got %q", buf.String())
	}
}
