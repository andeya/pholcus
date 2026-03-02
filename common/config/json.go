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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/andeya/gust/option"
	"github.com/andeya/gust/result"
	"github.com/andeya/pholcus/common/closer"
)

// JSONConfig is a json config parser and implements Config interface.
type JSONConfig struct {
}

// Parse returns a ConfigContainer with parsed json config map.
func (js *JSONConfig) Parse(filename string) (Configer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer closer.LogClose(file, log.Printf)
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return js.ParseData(content)
}

// ParseData returns a ConfigContainer with json string
func (js *JSONConfig) ParseData(data []byte) (Configer, error) {
	x := &JSONConfigContainer{
		data: make(map[string]interface{}),
	}
	err := json.Unmarshal(data, &x.data)
	if err != nil {
		var wrappingArray []interface{}
		err2 := json.Unmarshal(data, &wrappingArray)
		if err2 != nil {
			return nil, err
		}
		x.data["rootArray"] = wrappingArray
	}
	return x, nil
}

// JSONConfigContainer represents the json configuration.
// Keys support section::name format for get operations.
type JSONConfigContainer struct {
	data map[string]interface{}
	sync.RWMutex
}

// Bool returns the boolean value for a given key.
func (c *JSONConfigContainer) Bool(key string) result.Result[bool] {
	val := c.getData(key)
	if val != nil {
		v, err := ParseBool(val)
		return result.Ret(v, err)
	}
	return result.TryErr[bool](fmt.Errorf("not exist key: %q", key))
}

// DefaultBool returns the bool value for a given key, or defaultval on error.
func (c *JSONConfigContainer) DefaultBool(key string, defaultval bool) bool {
	return c.Bool(key).UnwrapOr(defaultval)
}

// Int returns the integer value for a given key.
func (c *JSONConfigContainer) Int(key string) result.Result[int] {
	val := c.getData(key)
	if val != nil {
		if v, ok := val.(float64); ok {
			return result.Ok(int(v))
		}
		return result.TryErr[int](errors.New("not int value"))
	}
	return result.TryErr[int](errors.New("not exist key:" + key))
}

// DefaultInt returns the integer value for a given key, or defaultval on error.
func (c *JSONConfigContainer) DefaultInt(key string, defaultval int) int {
	return c.Int(key).UnwrapOr(defaultval)
}

// Int64 returns the int64 value for a given key.
func (c *JSONConfigContainer) Int64(key string) result.Result[int64] {
	val := c.getData(key)
	if val != nil {
		if v, ok := val.(float64); ok {
			return result.Ok(int64(v))
		}
		return result.TryErr[int64](errors.New("not int64 value"))
	}
	return result.TryErr[int64](errors.New("not exist key:" + key))
}

// DefaultInt64 returns the int64 value for a given key, or defaultval on error.
func (c *JSONConfigContainer) DefaultInt64(key string, defaultval int64) int64 {
	return c.Int64(key).UnwrapOr(defaultval)
}

// Float returns the float value for a given key.
func (c *JSONConfigContainer) Float(key string) result.Result[float64] {
	val := c.getData(key)
	if val != nil {
		if v, ok := val.(float64); ok {
			return result.Ok(v)
		}
		return result.TryErr[float64](errors.New("not float64 value"))
	}
	return result.TryErr[float64](errors.New("not exist key:" + key))
}

// DefaultFloat returns the float64 value for a given key, or defaultval on error.
func (c *JSONConfigContainer) DefaultFloat(key string, defaultval float64) float64 {
	return c.Float(key).UnwrapOr(defaultval)
}

// String returns the string value for a given key.
func (c *JSONConfigContainer) String(key string) option.Option[string] {
	val := c.getData(key)
	if val != nil {
		if v, ok := val.(string); ok {
			return option.Some(v)
		}
	}
	return option.None[string]()
}

// DefaultString returns the string value for a given key, or defaultval if empty.
func (c *JSONConfigContainer) DefaultString(key string, defaultval string) string {
	return c.String(key).UnwrapOr(defaultval)
}

// Strings returns the []string value for a given key.
func (c *JSONConfigContainer) Strings(key string) []string {
	stringOpt := c.String(key)
	if stringOpt.IsNone() {
		return nil
	}
	return strings.Split(stringOpt.Unwrap(), ";")
}

// DefaultStrings returns the []string value for a given key, or defaultval if nil.
func (c *JSONConfigContainer) DefaultStrings(key string, defaultval []string) []string {
	if v := c.Strings(key); v != nil {
		return v
	}
	return defaultval
}

// GetSection returns map for the given section
func (c *JSONConfigContainer) GetSection(section string) result.Result[map[string]string] {
	if v, ok := c.data[section]; ok {
		return result.Ok(v.(map[string]string))
	}
	return result.TryErr[map[string]string](errors.New("section does not exist: " + section))
}

// SaveConfigFile writes the configuration to the given file.
func (c *JSONConfigContainer) SaveConfigFile(filename string) (err error) {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer closer.LogClose(f, log.Printf)
	b, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

// Set writes a new value for key.
func (c *JSONConfigContainer) Set(key, val string) result.VoidResult {
	c.Lock()
	defer c.Unlock()
	c.data[key] = val
	return result.OkVoid()
}

// DIY returns the raw value by a given key.
func (c *JSONConfigContainer) DIY(key string) result.Result[interface{}] {
	val := c.getData(key)
	if val != nil {
		return result.Ok[interface{}](val)
	}
	return result.TryErr[interface{}](errors.New("key does not exist"))
}

// getData returns the value for section.key or key.
func (c *JSONConfigContainer) getData(key string) interface{} {
	if len(key) == 0 {
		return nil
	}

	c.RLock()
	defer c.RUnlock()

	sectionKeys := strings.Split(key, "::")
	if len(sectionKeys) >= 2 {
		curValue, ok := c.data[sectionKeys[0]]
		if !ok {
			return nil
		}
		for _, key := range sectionKeys[1:] {
			if v, ok := curValue.(map[string]interface{}); ok {
				if curValue, ok = v[key]; !ok {
					return nil
				}
			}
		}
		return curValue
	}
	if v, ok := c.data[key]; ok {
		return v
	}
	return nil
}

func init() {
	Register("json", &JSONConfig{})
}
