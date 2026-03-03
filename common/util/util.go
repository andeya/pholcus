// Package util provides common utility functions such as MD5, random numbers, path handling, etc.
package util

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"io"
	r "math/rand"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/net/html/charset"

	"github.com/andeya/gust/option"
	"github.com/andeya/gust/result"
)

const (
	// USE_KEYIN is the initial value for enabling Keyin in Spider.
	USE_KEYIN = "\r\t\n"
)

var (
	re             = regexp.MustCompile(">[ \t\n\v\f\r]+<")
	jsonpKeyRegexp = regexp.MustCompile(`([^\s\:\{\,\d"]+|[a-z][a-z\d]*)\s*\:`)
	isNumRegexp    = regexp.MustCompile(`^\d+$`)
)

// JSONPToJSON modify jsonp string to json string
// Example: forbar({a:"1",b:2}) to {"a":"1","b":2}
func JSONPToJSON(json string) string {
	start := strings.Index(json, "{")
	end := strings.LastIndex(json, "}")
	start1 := strings.Index(json, "[")
	if start1 >= 0 && (start == -1 || start > start1) {
		start = start1
		end = strings.LastIndex(json, "]")
	}
	if end > start && end != -1 && start != -1 {
		json = json[start : end+1]
	}
	json = strings.ReplaceAll(json, "\\'", "")
	return jsonpKeyRegexp.ReplaceAllString(json, "\"$1\":")
}

// Mkdir creates the directory for the given path.
func Mkdir(filePath string) result.VoidResult {
	p, _ := path.Split(filePath)
	if p == "" {
		return result.OkVoid()
	}
	d, err := os.Stat(p)
	if err != nil || !d.IsDir() {
		if err = os.MkdirAll(p, 0777); err != nil {
			return result.FmtErrVoid("failed to create path [%v]: %v", filePath, err)
		}
	}
	return result.OkVoid()
}

// The GetWDPath gets the work directory path.
func GetWDPath() string {
	wd := os.Getenv("GOPATH")
	if wd == "" {
		panic("GOPATH is not set in env.")
	}
	return wd
}

// The IsDirExists judges path is directory or not.
func IsDirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return fi.IsDir()
}

// The IsFileExists judges path is file or not.
func IsFileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return !fi.IsDir()
}

// walkPath resolves targpath to an absolute path.
func walkPath(targpath string) result.Result[string] {
	if filepath.IsAbs(targpath) {
		return result.Ok(targpath)
	}
	return result.Ret(filepath.Abs(targpath))
}

// WalkFiles walks files under targpath, optionally filtered by suffixes.
func WalkFiles(targpath string, suffixes ...string) (ret result.Result[[]string]) {
	defer ret.Catch()
	targpath = walkPath(targpath).Unwrap()
	var filelist []string
	result.RetVoid(filepath.Walk(targpath, func(retpath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		if len(suffixes) == 0 {
			filelist = append(filelist, retpath)
			return nil
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(retpath, suffix) {
				filelist = append(filelist, retpath)
			}
		}
		return nil
	})).Unwrap()
	return result.Ok(filelist)
}

// WalkDir walks directories under targpath, optionally filtered by suffixes.
func WalkDir(targpath string, suffixes ...string) (ret result.Result[[]string]) {
	defer ret.Catch()
	targpath = walkPath(targpath).Unwrap()
	var dirlist []string
	result.RetVoid(filepath.Walk(targpath, func(retpath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !f.IsDir() {
			return nil
		}
		if len(suffixes) == 0 {
			dirlist = append(dirlist, retpath)
			return nil
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(retpath, suffix) {
				dirlist = append(dirlist, retpath)
			}
		}
		return nil
	})).Unwrap()
	return result.Ok(dirlist)
}

// WalkRelFiles walks files under targpath and returns relative paths, optionally filtered by suffixes.
func WalkRelFiles(targpath string, suffixes ...string) (ret result.Result[[]string]) {
	defer ret.Catch()
	targpath = walkPath(targpath).Unwrap()
	var filelist []string
	result.RetVoid(filepath.Walk(targpath, func(retpath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		relpath := RelPath(retpath).Unwrap()
		if len(suffixes) == 0 {
			filelist = append(filelist, relpath)
			return nil
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(relpath, suffix) {
				filelist = append(filelist, relpath)
			}
		}
		return nil
	})).Unwrap()
	return result.Ok(filelist)
}

// WalkRelDir walks directories under targpath and returns relative paths, optionally filtered by suffixes.
func WalkRelDir(targpath string, suffixes ...string) (ret result.Result[[]string]) {
	defer ret.Catch()
	targpath = walkPath(targpath).Unwrap()
	var dirlist []string
	result.RetVoid(filepath.Walk(targpath, func(retpath string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !f.IsDir() {
			return nil
		}
		relpath := RelPath(retpath).Unwrap()
		if len(suffixes) == 0 {
			dirlist = append(dirlist, relpath)
			return nil
		}
		for _, suffix := range suffixes {
			if strings.HasSuffix(relpath, suffix) {
				dirlist = append(dirlist, relpath)
			}
		}
		return nil
	})).Unwrap()
	return result.Ok(dirlist)
}

// RelPath converts targpath to a path relative to the current working directory.
func RelPath(targpath string) (ret result.Result[string]) {
	defer ret.Catch()
	basepath := result.Ret(filepath.Abs("./")).Unwrap()
	rel := result.Ret(filepath.Rel(basepath, targpath)).Unwrap()
	return result.Ok(strings.ReplaceAll(rel, `\`, `/`))
}

// The IsNum judges string is number or not.
func IsNum(a string) bool {
	return isNumRegexp.MatchString(a)
}

// XML2MapStr converts simple XML to a string map (supports UTF-8).
func XML2MapStr(xmldoc string) map[string]string {
	var t xml.Token
	var err error
	inputReader := strings.NewReader(xmldoc)
	decoder := xml.NewDecoder(inputReader)
	decoder.CharsetReader = func(s string, r io.Reader) (io.Reader, error) {
		return charset.NewReader(r, s)
	}
	m := make(map[string]string, 32)
	key := ""
	for t, err = decoder.Token(); err == nil; t, err = decoder.Token() {
		switch token := t.(type) {
		case xml.StartElement:
			key = token.Name.Local
		case xml.CharData:
			content := string(token)
			m[key] = content
		default:
		}
	}
	return m
}

// MakeHash converts a string to a CRC32 hash hex string.
func MakeHash(s string) string {
	const IEEE = 0xedb88320
	var IEEETable = crc32.MakeTable(IEEE)
	hash := fmt.Sprintf("%x", crc32.Checksum([]byte(s), IEEETable))
	return hash
}

func HashString(encode string) uint64 {
	hash := fnv.New64()
	hash.Write([]byte(encode))
	return hash.Sum64()
}

// MakeUnique creates a unique fingerprint for obj (method 1: FNV-64).
func MakeUnique(obj interface{}) string {
	b, _ := json.Marshal(obj)
	hash := fnv.New64()
	hash.Write(b)
	return strconv.FormatUint(hash.Sum64(), 10)
}

// MakeMd5 creates an MD5 fingerprint for obj (method 2).
func MakeMd5(obj interface{}, length int) string {
	if length > 32 {
		length = 32
	}
	h := md5.New()
	baseString, _ := json.Marshal(obj)
	h.Write([]byte(baseString))
	s := hex.EncodeToString(h.Sum(nil))
	return s[:length]
}

// JSONString converts obj to a JSON string.
func JSONString(obj interface{}) string {
	b, _ := json.Marshal(obj)
	s := fmt.Sprintf("%+v", Bytes2String(b))
	r := strings.ReplaceAll(s, `\u003c`, "<")
	r = strings.ReplaceAll(r, `\u003e`, ">")
	return r
}

// CheckErr checks and logs the error if non-nil.
func CheckErr(err error) {
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
}
func CheckErrPanic(err error) {
	if err != nil {
		panic(err)
	}
}

// FileNameReplace replaces invalid filename characters with similar alternatives.
func FileNameReplace(fileName string) string {
	var q = 1
	r := []rune(fileName)
	size := len(r)
	for i := 0; i < size; i++ {
		switch r[i] {
		case '"':
			if q%2 == 1 {
				r[i] = '\u201c'
			} else {
				r[i] = '\u201d'
			}
			q++
		case ':':
			r[i] = '\uff1a'
		case '*':
			r[i] = '\u00d7'
		case '<':
			r[i] = '\uff1c'
		case '>':
			r[i] = '\uff1e'
		case '?':
			r[i] = '\uff1f'
		case '/':
			r[i] = '\uff0f'
		case '|':
			r[i] = '\u2223'
		case '\\':
			r[i] = '\u2572'
		}
	}
	return strings.ReplaceAll(string(r), USE_KEYIN, ``)
}

// ExcelSheetNameReplace replaces invalid Excel sheet name characters with underscores.
func ExcelSheetNameReplace(fileName string) string {
	r := []rune(fileName)
	size := len(r)
	for i := 0; i < size; i++ {
		switch r[i] {
		case ':', '\uff1a', '*', '?', '\uff1f', '/', '\uff0f', '\\', '\u2572', ']', '[':
			r[i] = '_'
		}
	}
	return strings.ReplaceAll(string(r), USE_KEYIN, ``)
}

// Atoa extracts a string from an interface{} value, returning None if nil.
func Atoa(str interface{}) option.Option[string] {
	if str == nil {
		return option.None[string]()
	}
	return option.Some(strings.Trim(str.(string), " "))
}

// Atoi extracts an int from an interface{} value, returning None if nil.
func Atoi(str interface{}) option.Option[int] {
	if str == nil {
		return option.None[int]()
	}
	i, _ := strconv.Atoi(strings.Trim(str.(string), " "))
	return option.Some(i)
}

// Atoui extracts a uint from an interface{} value, returning None if nil.
func Atoui(str interface{}) option.Option[uint] {
	if str == nil {
		return option.None[uint]()
	}
	u, _ := strconv.Atoi(strings.Trim(str.(string), " "))
	return option.Some(uint(u))
}

// RandomCreateBytes generate random []byte by specify chars.
func RandomCreateBytes(n int, alphabets ...byte) []byte {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	var randby bool
	if num, err := rand.Read(bytes); num != n || err != nil {
		r.Seed(time.Now().UnixNano())
		randby = true
	}
	for i, b := range bytes {
		if len(alphabets) == 0 {
			if randby {
				bytes[i] = alphanum[r.Intn(len(alphanum))]
			} else {
				bytes[i] = alphanum[b%byte(len(alphanum))]
			}
		} else {
			if randby {
				bytes[i] = alphabets[r.Intn(len(alphabets))]
			} else {
				bytes[i] = alphabets[b%byte(len(alphabets))]
			}
		}
	}
	return bytes
}

// KeyinsParse splits user-provided custom keyins into unique tokens.
func KeyinsParse(keyins string) []string {
	keyins = strings.TrimSpace(keyins)
	if keyins == "" {
		return []string{}
	}
	for _, v := range re.FindAllString(keyins, -1) {
		keyins = strings.ReplaceAll(keyins, v, "><")
	}
	m := map[string]bool{}
	for _, v := range strings.Split(keyins, "><") {
		v = strings.TrimPrefix(v, "<")
		v = strings.TrimSuffix(v, ">")
		if v == "" {
			continue
		}
		m[v] = true
	}
	s := make([]string, len(m))
	i := 0
	for k := range m {
		s[i] = k
		i++
	}
	return s
}

// Bytes2String converts []byte to string via direct pointer conversion.
// Both share the same underlying memory; modifying one affects the other.
// Much faster than string([]byte{}) for large conversions.
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// String2Bytes converts string to []byte via direct pointer conversion.
// Both share the same underlying memory; modifying one affects the other.
// Do not mutate the returned slice directly (e.g. b[1]='d') or the program may panic.
func String2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}
