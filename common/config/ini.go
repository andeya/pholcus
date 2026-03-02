// Copyright 2014 beego Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andeya/gust/option"
	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/closer"
)

const utf8BOM = "\xEF\xBB\xBF"

var (
	defaultSection = "default"   // section for ini items not under a named section
	bNumComment    = []byte{'#'} // hash comment prefix
	bSemComment    = []byte{';'} // semicolon comment prefix
	bEmpty         = []byte{}
	bEqual         = []byte{'='} // key-value separator
	bDQuote        = []byte{'"'} // double-quote for values
	sectionStart   = []byte{'['} // section header start
	sectionEnd     = []byte{']'} // section header end
	lineBreak      = "\n"
)

// IniConfig implements Config to parse ini file.
type IniConfig struct {
}

// Parse creates a new Config and parses the file configuration from the named file.
func (ini *IniConfig) Parse(name string) (Configer, error) {
	r := ini.parseFile(name)
	if r.IsErr() {
		return nil, r.UnwrapErr()
	}
	return r.Unwrap(), nil
}

func (ini *IniConfig) parseFile(name string) (r result.Result[*IniConfigContainer]) {
	defer r.Catch()
	file := result.Ret(os.Open(name)).Unwrap()

	cfg := &IniConfigContainer{
		file.Name(),
		make(map[string]map[string]string),
		make(map[string]string),
		make(map[string]string),
		sync.RWMutex{},
	}
	cfg.Lock()
	defer cfg.Unlock()
	defer closer.LogClose(file, log.Printf)

	var comment bytes.Buffer
	buf := bufio.NewReader(file)
	skipBOM(buf)
	section := defaultSection
	for {
		line, _, err := buf.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		if bytes.Equal(line, bEmpty) {
			continue
		}
		line = bytes.TrimSpace(line)

		if parseCommentLine(line, &comment) {
			continue
		}
		if newSection, ok := parseSectionHeader(line); ok {
			section = newSection
			applySectionComment(cfg, section, &comment)
			ensureSection(cfg, section)
			continue
		}
		ensureSection(cfg, section)
		result.RetVoid(parseKeyValueLine(line, name, ini, cfg, section, &comment)).Unwrap()
	}
	return result.Ok(cfg)
}

// skipBOM removes UTF-8 BOM from the reader if present.
func skipBOM(buf *bufio.Reader) {
	head, err := buf.Peek(3)
	if err == nil && strings.HasPrefix(string(head), utf8BOM) {
		for i := 1; i <= 3; i++ {
			buf.ReadByte()
		}
	}
}

// parseCommentLine processes comment lines (# or ;). Returns true if line was a comment.
func parseCommentLine(line []byte, comment *bytes.Buffer) bool {
	var bComment []byte
	switch {
	case bytes.HasPrefix(line, bNumComment):
		bComment = bNumComment
	case bytes.HasPrefix(line, bSemComment):
		bComment = bSemComment
	}
	if bComment == nil {
		return false
	}
	line = bytes.TrimLeft(line, string(bComment))
	if comment.Len() > 0 {
		comment.WriteByte('\n')
	}
	comment.Write(line)
	return true
}

// parseSectionHeader extracts section name from [section] lines. Returns (name, true) if valid.
func parseSectionHeader(line []byte) (string, bool) {
	if !bytes.HasPrefix(line, sectionStart) || !bytes.HasSuffix(line, sectionEnd) {
		return "", false
	}
	return strings.ToLower(string(line[1 : len(line)-1])), true
}

func applySectionComment(cfg *IniConfigContainer, section string, comment *bytes.Buffer) {
	if comment.Len() > 0 {
		cfg.sectionComment[section] = comment.String()
		comment.Reset()
	}
}

func ensureSection(cfg *IniConfigContainer, section string) {
	if _, ok := cfg.data[section]; !ok {
		cfg.data[section] = make(map[string]string)
	}
}

// parseKeyValueLine parses key=value or include directive. Merges included files into cfg.
func parseKeyValueLine(line []byte, filename string, ini *IniConfig, cfg *IniConfigContainer, section string, comment *bytes.Buffer) error {
	keyValue := bytes.SplitN(line, bEqual, 2)
	key := strings.ToLower(string(bytes.TrimSpace(keyValue[0])))

	if len(keyValue) == 1 && strings.HasPrefix(key, "include") {
		includefiles := strings.Fields(key)
		if includefiles[0] == "include" && len(includefiles) == 2 {
			otherfile := strings.Trim(includefiles[1], "\"")
			if !path.IsAbs(otherfile) {
				otherfile = path.Join(path.Dir(filename), otherfile)
			}
			incResult := ini.parseFile(otherfile)
			if incResult.IsErr() {
				return incResult.UnwrapErr()
			}
			inc := incResult.Unwrap()
			for sec, dt := range inc.data {
				if _, ok := cfg.data[sec]; !ok {
					cfg.data[sec] = make(map[string]string)
				}
				for k, v := range dt {
					cfg.data[sec][k] = v
				}
			}
			for sec, comm := range inc.sectionComment {
				cfg.sectionComment[sec] = comm
			}
			for k, comm := range inc.keyComment {
				cfg.keyComment[k] = comm
			}
			return nil
		}
	}

	if len(keyValue) != 2 {
		return errors.New("read the content error: \"" + string(line) + "\", should key = val")
	}
	val := bytes.TrimSpace(keyValue[1])
	if bytes.HasPrefix(val, bDQuote) {
		val = bytes.Trim(val, `"`)
	}
	cfg.data[section][key] = string(val)
	if comment.Len() > 0 {
		cfg.keyComment[section+"."+key] = comment.String()
		comment.Reset()
	}
	return nil
}

// ParseData parse ini the data
func (ini *IniConfig) ParseData(data []byte) (Configer, error) {
	tmpName := path.Join(os.TempDir(), "beego", fmt.Sprintf("%d", time.Now().Nanosecond()))
	os.MkdirAll(path.Dir(tmpName), os.ModePerm)
	if err := os.WriteFile(tmpName, data, 0655); err != nil {
		return nil, err
	}
	return ini.Parse(tmpName)
}

// IniConfigContainer represents the ini configuration.
// Keys support section::name format for get/set operations.
type IniConfigContainer struct {
	filename       string
	data           map[string]map[string]string // section => key:val
	sectionComment map[string]string            // section => comment
	keyComment     map[string]string            // section.key => comment
	sync.RWMutex
}

func (c *IniConfigContainer) MainKeys() []string {
	l := len(c.data[defaultSection])
	a := make([]string, l)
	i := 0
	for k := range c.data[defaultSection] {
		a[i] = k
		i++
	}
	return a
}

func (c *IniConfigContainer) Sections() []string {
	l := len(c.data)
	a := make([]string, l)
	i := 0
	for k := range c.data {
		if k == defaultSection {
			continue
		}
		a[i] = k
		i++
	}
	return a
}

func (c *IniConfigContainer) SectionKeys(section string) []string {
	l := len(c.data[section])
	a := make([]string, l)
	i := 0
	for k := range c.data[section] {
		a[i] = k
		i++
	}
	return a
}

// Bool returns the boolean value for a given key.
func (c *IniConfigContainer) Bool(key string) result.Result[bool] {
	v, err := ParseBool(c.getdata(key))
	return result.Ret(v, err)
}

// DefaultBool returns the boolean value for a given key, or defaultval on error.
func (c *IniConfigContainer) DefaultBool(key string, defaultval bool) bool {
	return c.Bool(key).UnwrapOr(defaultval)
}

// Int returns the integer value for a given key.
func (c *IniConfigContainer) Int(key string) result.Result[int] {
	v, err := strconv.Atoi(c.getdata(key))
	return result.Ret(v, err)
}

// DefaultInt returns the integer value for a given key, or defaultval on error.
func (c *IniConfigContainer) DefaultInt(key string, defaultval int) int {
	return c.Int(key).UnwrapOr(defaultval)
}

// Int64 returns the int64 value for a given key.
func (c *IniConfigContainer) Int64(key string) result.Result[int64] {
	v, err := strconv.ParseInt(c.getdata(key), 10, 64)
	return result.Ret(v, err)
}

// DefaultInt64 returns the int64 value for a given key, or defaultval on error.
func (c *IniConfigContainer) DefaultInt64(key string, defaultval int64) int64 {
	return c.Int64(key).UnwrapOr(defaultval)
}

// Float returns the float value for a given key.
func (c *IniConfigContainer) Float(key string) result.Result[float64] {
	v, err := strconv.ParseFloat(c.getdata(key), 64)
	return result.Ret(v, err)
}

// DefaultFloat returns the float64 value for a given key, or defaultval on error.
func (c *IniConfigContainer) DefaultFloat(key string, defaultval float64) float64 {
	return c.Float(key).UnwrapOr(defaultval)
}

// String returns the string value for a given key.
func (c *IniConfigContainer) String(key string) option.Option[string] {
	v, ok := c.getdataWithExists(key)
	return option.BoolOpt(v, ok)
}

// DefaultString returns the string value for a given key, or defaultval if empty.
func (c *IniConfigContainer) DefaultString(key string, defaultval string) string {
	return c.String(key).UnwrapOr(defaultval)
}

// Strings returns the []string value for a given key.
// Return nil if config value does not exist or is empty.
func (c *IniConfigContainer) Strings(key string) []string {
	v := c.String(key)
	if v.IsNone() {
		return nil
	}
	return strings.Split(v.Unwrap(), ";")
}

// DefaultStrings returns the []string value for a given key, or defaultval if nil.
func (c *IniConfigContainer) DefaultStrings(key string, defaultval []string) []string {
	v := c.Strings(key)
	if v == nil {
		return defaultval
	}
	return v
}

// GetSection returns map for the given section
func (c *IniConfigContainer) GetSection(section string) result.Result[map[string]string] {
	if v, ok := c.data[section]; ok {
		return result.Ok(v)
	}
	return result.TryErr[map[string]string](errors.New("section does not exist"))
}

func (c *IniConfigContainer) GetAllSections() map[string]map[string]string {
	return c.data
}

// SaveConfigFile writes the configuration to the given file.
func (c *IniConfigContainer) SaveConfigFile(filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer closer.LogClose(f, log.Printf)

	getCommentStr := func(section, key string) string {
		comment, ok := "", false
		if len(key) == 0 {
			comment, ok = c.sectionComment[section]
		} else {
			comment, ok = c.keyComment[section+"."+key]
		}

		if ok {
			if len(comment) == 0 || len(strings.TrimSpace(comment)) == 0 {
				return string(bNumComment)
			}
			prefix := string(bNumComment)
			return prefix + strings.ReplaceAll(comment, lineBreak, lineBreak+prefix)
		}
		return ""
	}

	buf := bytes.NewBuffer(nil)
	if dt, ok := c.data[defaultSection]; ok {
		keys := []string{}
		for key := range dt {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			val := dt[key]
			if key != " " {
				if v := getCommentStr(defaultSection, key); len(v) > 0 {
					if _, err = buf.WriteString(v + lineBreak); err != nil {
						return err
					}
				}

				if _, err = buf.WriteString(key + string(bEqual) + val + lineBreak); err != nil {
					return err
				}
			}
		}

		if _, err = buf.WriteString(lineBreak); err != nil {
			return err
		}
	}
	sections := []string{}
	for section := range c.data {
		sections = append(sections, section)
	}
	sort.Strings(sections)
	for _, section := range sections {
		dt := c.data[section]
		if section != defaultSection {
			if v := getCommentStr(section, ""); len(v) > 0 {
				if _, err = buf.WriteString(v + lineBreak); err != nil {
					return err
				}
			}

			if _, err = buf.WriteString(string(sectionStart) + section + string(sectionEnd) + lineBreak); err != nil {
				return err
			}
			keys := []string{}
			for key := range dt {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				val := dt[key]
				if key != " " {
					if v := getCommentStr(section, key); len(v) > 0 {
						if _, err = buf.WriteString(v + lineBreak); err != nil {
							return err
						}
					}

					if _, err = buf.WriteString(key + string(bEqual) + val + lineBreak); err != nil {
						return err
					}
				}
			}

			if _, err = buf.WriteString(lineBreak); err != nil {
				return err
			}
		}
	}

	if _, err = buf.WriteTo(f); err != nil {
		return err
	}
	return nil
}

// Set writes a new value for key. Use "section::key" for section-specific keys.
func (c *IniConfigContainer) Set(key, value string) result.VoidResult {
	c.Lock()
	defer c.Unlock()
	if len(key) == 0 {
		return result.TryErrVoid(errors.New("key is empty"))
	}

	var (
		section, k string
		sectionKey = strings.Split(key, "::")
	)

	if len(sectionKey) >= 2 {
		section = sectionKey[0]
		k = sectionKey[1]
	} else {
		section = defaultSection
		k = sectionKey[0]
	}

	if _, ok := c.data[section]; !ok {
		c.data[section] = make(map[string]string)
	}
	c.data[section][k] = value
	return result.OkVoid()
}

// DIY returns the raw value by a given key.
func (c *IniConfigContainer) DIY(key string) result.Result[interface{}] {
	v, ok := c.getdataWithExists(key)
	if ok {
		return result.Ok[interface{}](v)
	}
	return result.TryErr[interface{}](errors.New("key not found"))
}

// getdata returns the value for section.key or key.
func (c *IniConfigContainer) getdata(key string) string {
	v, _ := c.getdataWithExists(key)
	return v
}

// getdataWithExists returns the value and whether the key exists.
func (c *IniConfigContainer) getdataWithExists(key string) (string, bool) {
	if len(key) == 0 {
		return "", false
	}
	c.RLock()
	defer c.RUnlock()

	var (
		section, k string
		sectionKey = strings.Split(strings.ToLower(key), "::")
	)
	if len(sectionKey) >= 2 {
		section = sectionKey[0]
		k = sectionKey[1]
	} else {
		section = defaultSection
		k = sectionKey[0]
	}
	if v, ok := c.data[section]; ok {
		if vv, ok := v[k]; ok {
			return vv, true
		}
	}
	return "", false
}

func init() {
	Register("ini", &IniConfig{})
}
