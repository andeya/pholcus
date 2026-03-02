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
	"errors"
	"strconv"
	"strings"

	"github.com/andeya/gust/option"
	"github.com/andeya/gust/result"
)

type fakeConfigContainer struct {
	data map[string]string
}

func (c *fakeConfigContainer) getData(key string) string {
	return c.data[strings.ToLower(key)]
}

func (c *fakeConfigContainer) Set(key, val string) result.VoidResult {
	c.data[strings.ToLower(key)] = val
	return result.OkVoid()
}

func (c *fakeConfigContainer) String(key string) option.Option[string] {
	v, ok := c.data[strings.ToLower(key)]
	return option.BoolOpt(v, ok)
}

func (c *fakeConfigContainer) DefaultString(key string, defaultval string) string {
	return c.String(key).UnwrapOr(defaultval)
}

func (c *fakeConfigContainer) Strings(key string) []string {
	v := c.String(key)
	if v.IsNone() {
		return nil
	}
	return strings.Split(v.Unwrap(), ";")
}

func (c *fakeConfigContainer) DefaultStrings(key string, defaultval []string) []string {
	v := c.Strings(key)
	if v == nil {
		return defaultval
	}
	return v
}

func (c *fakeConfigContainer) Int(key string) result.Result[int] {
	v, err := strconv.Atoi(c.getData(key))
	return result.Ret(v, err)
}

func (c *fakeConfigContainer) DefaultInt(key string, defaultval int) int {
	return c.Int(key).UnwrapOr(defaultval)
}

func (c *fakeConfigContainer) Int64(key string) result.Result[int64] {
	v, err := strconv.ParseInt(c.getData(key), 10, 64)
	return result.Ret(v, err)
}

func (c *fakeConfigContainer) DefaultInt64(key string, defaultval int64) int64 {
	return c.Int64(key).UnwrapOr(defaultval)
}

func (c *fakeConfigContainer) Bool(key string) result.Result[bool] {
	v, err := ParseBool(c.getData(key))
	return result.Ret(v, err)
}

func (c *fakeConfigContainer) DefaultBool(key string, defaultval bool) bool {
	return c.Bool(key).UnwrapOr(defaultval)
}

func (c *fakeConfigContainer) Float(key string) result.Result[float64] {
	v, err := strconv.ParseFloat(c.getData(key), 64)
	return result.Ret(v, err)
}

func (c *fakeConfigContainer) DefaultFloat(key string, defaultval float64) float64 {
	return c.Float(key).UnwrapOr(defaultval)
}

func (c *fakeConfigContainer) DIY(key string) result.Result[interface{}] {
	if v, ok := c.data[strings.ToLower(key)]; ok {
		return result.Ok[interface{}](v)
	}
	return result.TryErr[interface{}](errors.New("key not found"))
}

func (c *fakeConfigContainer) GetSection(section string) result.Result[map[string]string] {
	return result.TryErr[map[string]string](errors.New("not implemented in fakeConfigContainer"))
}

func (c *fakeConfigContainer) SaveConfigFile(filename string) error {
	return errors.New("not implement in the fakeConfigContainer")
}

var _ Configer = new(fakeConfigContainer)

// NewFakeConfig returns a fake Configer for testing.
func NewFakeConfig() Configer {
	return &fakeConfigContainer{
		data: make(map[string]string),
	}
}
