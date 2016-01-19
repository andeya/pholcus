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

// package yaml for config provider

//
// Usage:
// import(
//   _ "github.com/henrylee2cn/config/yaml"
//   "github.com/henrylee2cn/config"
// )
//
//  cnf, err := config.NewConfig("yaml", "config.yaml")

package yaml

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/henrylee2cn/pholcus/common/config"
	"gopkg.in/yaml.v2"
)

// YAMLConfig is a yaml config parser and implements Config interface.
type YAMLConfig struct{}

// Parse returns a ConfigContainer with parsed yaml config map.
func (yaml *YAMLConfig) Parse(filename string) (y config.ConfigContainer, err error) {
	cnf, err := ReadYmlReader(filename)
	if err != nil {
		return
	}
	y = &YAMLConfigContainer{
		data: cnf,
	}
	return
}

func (yaml *YAMLConfig) ParseData(data []byte) (config.ConfigContainer, error) {
	// Save memory data to temporary file
	tmpName := path.Join(os.TempDir(), "beego", fmt.Sprintf("%d", time.Now().Nanosecond()))
	os.MkdirAll(path.Dir(tmpName), os.ModePerm)
	if err := ioutil.WriteFile(tmpName, data, 0655); err != nil {
		return nil, err
	}
	return yaml.Parse(tmpName)
}

// Read yaml file to map.
// if json like, use json package, unless yaml package.
func ReadYmlReader(path string) (cnf map[string]interface{}, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil || len(buf) < 3 {
		return
	}

	if string(buf[0:1]) == "{" {
		log.Println("Look like a Json, try json umarshal")
		err = json.Unmarshal(buf, &cnf)
		if err == nil {
			log.Println("It is Json Map")
			return
		}
	}

	err = yaml.Unmarshal(buf, &cnf)
	if err != nil {
		log.Println("yaml ERR>", string(buf), err)
		return
	}

	if cnf == nil {
		log.Println("yaml output nil? Pls report bug\n" + string(buf))
		return
	}

	return
}

// A Config represents the yaml configuration.
type YAMLConfigContainer struct {
	data map[string]interface{}
	sync.Mutex
}

// Bool returns the boolean value for a given key.
func (c *YAMLConfigContainer) Bool(key string) (bool, error) {
	if v, ok := c.data[key].(bool); ok {
		return v, nil
	}
	return false, errors.New("not bool value")
}

// DefaultBool return the bool value if has no error
// otherwise return the defaultval
func (c *YAMLConfigContainer) DefaultBool(key string, defaultval bool) bool {
	if v, err := c.Bool(key); err != nil {
		return defaultval
	} else {
		return v
	}
}

// Int returns the integer value for a given key.
func (c *YAMLConfigContainer) Int(key string) (int, error) {
	if v, ok := c.data[key].(int64); ok {
		return int(v), nil
	}
	return 0, errors.New("not int value")
}

// DefaultInt returns the integer value for a given key.
// if err != nil return defaltval
func (c *YAMLConfigContainer) DefaultInt(key string, defaultval int) int {
	if v, err := c.Int(key); err != nil {
		return defaultval
	} else {
		return v
	}
}

// Int64 returns the int64 value for a given key.
func (c *YAMLConfigContainer) Int64(key string) (int64, error) {
	if v, ok := c.data[key].(int64); ok {
		return v, nil
	}
	return 0, errors.New("not bool value")
}

// DefaultInt64 returns the int64 value for a given key.
// if err != nil return defaltval
func (c *YAMLConfigContainer) DefaultInt64(key string, defaultval int64) int64 {
	if v, err := c.Int64(key); err != nil {
		return defaultval
	} else {
		return v
	}
}

// Float returns the float value for a given key.
func (c *YAMLConfigContainer) Float(key string) (float64, error) {
	if v, ok := c.data[key].(float64); ok {
		return v, nil
	}
	return 0.0, errors.New("not float64 value")
}

// DefaultFloat returns the float64 value for a given key.
// if err != nil return defaltval
func (c *YAMLConfigContainer) DefaultFloat(key string, defaultval float64) float64 {
	if v, err := c.Float(key); err != nil {
		return defaultval
	} else {
		return v
	}
}

// String returns the string value for a given key.
func (c *YAMLConfigContainer) String(key string) string {
	if v, ok := c.data[key].(string); ok {
		return v
	}
	return ""
}

// DefaultString returns the string value for a given key.
// if err != nil return defaltval
func (c *YAMLConfigContainer) DefaultString(key string, defaultval string) string {
	if v := c.String(key); v == "" {
		return defaultval
	} else {
		return v
	}
}

// Strings returns the []string value for a given key.
func (c *YAMLConfigContainer) Strings(key string) []string {
	return strings.Split(c.String(key), ";")
}

// DefaultStrings returns the []string value for a given key.
// if err != nil return defaltval
func (c *YAMLConfigContainer) DefaultStrings(key string, defaultval []string) []string {
	if v := c.Strings(key); len(v) == 0 {
		return defaultval
	} else {
		return v
	}
}

// GetSection returns map for the given section
func (c *YAMLConfigContainer) GetSection(section string) (map[string]string, error) {
	if v, ok := c.data[section]; ok {
		var r = make(map[string]string)
		for kk, vv := range v.(map[interface{}]interface{}) {
			r[kk.(string)] = vv.(string)
		}
		return r, nil
	} else {
		return nil, errors.New("not exist setction")
	}
}

// SaveConfigFile save the config into file
func (c *YAMLConfigContainer) SaveConfigFile(filename string) (err error) {
	// Write configuration file by filename.
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	out, err := yaml.Marshal(c.data)
	f.Write(out)
	return err
}

// WriteValue writes some new values for root key.
// if write to one section, the key need be "section::key".
func (c *YAMLConfigContainer) SetDIY(key string, val interface{}) error {
	c.Lock()
	defer c.Unlock()
	if len(key) == 0 {
		return errors.New("key is empty")
	}

	var sectionKey []string = strings.Split(key, "::")
	var deep = len(sectionKey)

	switch {
	case deep >= 1:
		if _, ok := c.data[sectionKey[0]]; !ok {
			c.data[sectionKey[0]] = make(map[string]interface{})
		}
		if deep == 1 {
			c.data[key] = val
			break
		}
		fallthrough
	case deep >= 2:
		if _, ok := c.data[sectionKey[0]].(map[string]interface{})[sectionKey[1]]; !ok {
			c.data[sectionKey[0]].(map[string]interface{})[sectionKey[1]] = make(map[string]interface{})
		}
		if deep == 2 {
			c.data[sectionKey[0]].(map[string]interface{})[sectionKey[1]] = val
			break
		}
		fallthrough
	case deep >= 3:
		if _, ok := c.data[sectionKey[0]].(map[string]interface{})[sectionKey[1]].(map[string]interface{})[sectionKey[2]]; !ok {
			c.data[sectionKey[0]].(map[string]interface{})[sectionKey[1]].(map[string]interface{})[sectionKey[2]] = make(map[string]interface{})
		}
		if deep == 3 {
			c.data[sectionKey[0]].(map[string]interface{})[sectionKey[1]].(map[string]interface{})[sectionKey[2]] = val
			break
		}
		fallthrough
	case deep >= 4:
		if _, ok := c.data[sectionKey[0]].(map[string]interface{})[sectionKey[1]].(map[string]interface{})[sectionKey[2]].(map[string]interface{})[sectionKey[3]]; !ok {
			c.data[sectionKey[0]].(map[string]interface{})[sectionKey[1]].(map[string]interface{})[sectionKey[2]].(map[string]interface{})[sectionKey[3]] = make(map[string]interface{})
		}
		if deep == 4 {
			c.data[sectionKey[0]].(map[string]interface{})[sectionKey[1]].(map[string]interface{})[sectionKey[2]].(map[string]interface{})[sectionKey[3]] = val
			break
		}
		fallthrough
	default:
		return errors.New("索引深度不可超过 4 级！")
	}
	return nil
}

func (c *YAMLConfigContainer) Set(key, val string) error {
	c.Lock()
	defer c.Unlock()
	if len(key) == 0 {
		return errors.New("key is empty")
	}
	c.data[key] = val
	return nil
}

// DIY returns the raw value by a given key.
func (c *YAMLConfigContainer) DIY(key string) (v interface{}, err error) {
	if v, ok := c.data[key]; ok {
		return v, nil
	}
	return nil, errors.New("not exist key")
}

func init() {
	config.Register("yaml", &YAMLConfig{})
}
