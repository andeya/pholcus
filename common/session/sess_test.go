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

package session

import (
	"crypto/aes"
	"encoding/json"
	"testing"
)

func Test_gob(t *testing.T) {
	a := make(map[interface{}]interface{})
	a["username"] = "astaxie"
	a[12] = 234
	a["user"] = User{"asta", "xie"}
	b, err := EncodeGob(a)
	if err != nil {
		t.Error(err)
	}
	c, err := DecodeGob(b)
	if err != nil {
		t.Error(err)
	}
	if len(c) == 0 {
		t.Error("decodeGob empty")
	}
	if c["username"] != "astaxie" {
		t.Error("decode string error")
	}
	if c[12] != 234 {
		t.Error("decode int error")
	}
	if c["user"].(User).Username != "asta" {
		t.Error("decode struct error")
	}
}

func TestEncodeGobDecodeGob(t *testing.T) {
	tests := []struct {
		name    string
		input   map[interface{}]interface{}
		wantErr bool
	}{
		{"empty map", map[interface{}]interface{}{}, false},
		{"string value", map[interface{}]interface{}{"k": "v"}, false},
		{"int key", map[interface{}]interface{}{42: "val"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := EncodeGob(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeGob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			dec, err := DecodeGob(enc)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeGob() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(dec) != len(tt.input) {
				t.Errorf("DecodeGob() len = %d, want %d", len(dec), len(tt.input))
			}
		})
	}
}

func TestDecodeGobInvalid(t *testing.T) {
	_, err := DecodeGob([]byte("invalid"))
	if err == nil {
		t.Error("DecodeGob(invalid) expected error")
	}
}

func TestRandomCreateBytes(t *testing.T) {
	tests := []struct {
		n         int
		alphabets []byte
	}{
		{10, nil},
		{0, nil},
		{5, []byte("01")},
	}
	for _, tt := range tests {
		got := RandomCreateBytes(tt.n, tt.alphabets...)
		if len(got) != tt.n {
			t.Errorf("RandomCreateBytes(%d) len = %d", tt.n, len(got))
		}
	}
}

type User struct {
	Username string
	NickName string
}

func TestGenerate(t *testing.T) {
	str := generateRandomKey(20)
	if len(str) != 20 {
		t.Fatal("generate length is not equal to 20")
	}
}

func TestCookieEncodeDecode(t *testing.T) {
	hashKey := "testhashKey"
	blockkey := generateRandomKey(16)
	block, err := aes.NewCipher(blockkey)
	if err != nil {
		t.Fatal("NewCipher:", err)
	}
	securityName := string(generateRandomKey(20))
	val := make(map[interface{}]interface{})
	val["name"] = "astaxie"
	val["gender"] = "male"
	str, err := encodeCookie(block, hashKey, securityName, val)
	if err != nil {
		t.Fatal("encodeCookie:", err)
	}
	dst := make(map[interface{}]interface{})
	dst, err = decodeCookie(block, hashKey, securityName, str, 3600)
	if err != nil {
		t.Fatal("decodeCookie", err)
	}
	if dst["name"] != "astaxie" {
		t.Fatal("dst get map error")
	}
	if dst["gender"] != "male" {
		t.Fatal("dst get map error")
	}
}

func TestParseConfig(t *testing.T) {
	s := `{"cookieName":"gosessionid","gclifetime":3600}`
	cf := new(managerConfig)
	cf.EnableSetCookie = true
	err := json.Unmarshal([]byte(s), cf)
	if err != nil {
		t.Fatal("parse json error,", err)
	}
	if cf.CookieName != "gosessionid" {
		t.Fatal("parseconfig get cookiename error")
	}
	if cf.Gclifetime != 3600 {
		t.Fatal("parseconfig get gclifetime error")
	}

	cc := `{"cookieName":"gosessionid","enableSetCookie":false,"gclifetime":3600,"ProviderConfig":"{\"cookieName\":\"gosessionid\",\"securityKey\":\"beegocookiehashkey\"}"}`
	cf2 := new(managerConfig)
	cf2.EnableSetCookie = true
	err = json.Unmarshal([]byte(cc), cf2)
	if err != nil {
		t.Fatal("parse json error,", err)
	}
	if cf2.CookieName != "gosessionid" {
		t.Fatal("parseconfig get cookiename error")
	}
	if cf2.Gclifetime != 3600 {
		t.Fatal("parseconfig get gclifetime error")
	}
	if cf2.EnableSetCookie != false {
		t.Fatal("parseconfig get enableSetCookie error")
	}
	cconfig := new(cookieConfig)
	err = json.Unmarshal([]byte(cf2.ProviderConfig), cconfig)
	if err != nil {
		t.Fatal("parse ProviderConfig err,", err)
	}
	if cconfig.CookieName != "gosessionid" {
		t.Fatal("ProviderConfig get cookieName error")
	}
	if cconfig.SecurityKey != "beegocookiehashkey" {
		t.Fatal("ProviderConfig get securityKey error")
	}
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name       string
		provider   string
		config     string
		wantErr    bool
		errContain string
	}{
		{
			name:     "unknown provider",
			provider: "unknown",
			config:   `{}`,
			wantErr:  true,
		},
		{
			name:     "invalid JSON",
			provider: "memory",
			config:   `{invalid}`,
			wantErr:  true,
		},
		{
			name:     "memory ok",
			provider: "memory",
			config:   `{"cookieName":"sid","gclifetime":3600}`,
			wantErr:  false,
		},
		{
			name:     "maxlifetime zero uses gclifetime",
			provider: "memory",
			config:   `{"cookieName":"sid","gclifetime":3600}`,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := NewManager(tt.provider, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if m == nil {
				t.Error("NewManager() returned nil manager")
			}
		})
	}
}
