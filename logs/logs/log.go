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

// Usage:
//
// import "github.com/astaxie/beego/logs"
//
//	log := NewLogger(10000)
//	log.SetLogger("console", "")
//
//	> the first params stand for how many channel
//
// Use it like this:
//
//	log.Debug("debug")
//	log.Informational("info")
//	log.Notice("notice")
//	log.Warning("warning")
//	log.Error("error")
//	log.Critical("critical")
//	log.Alert("alert")
//	log.Emergency("emergency")
//
//  more docs http://beego.me/docs/module/logs.md

//  Modified By henrylee2cn

package logs

import (
	"errors"
	"fmt"
	"path"
	"runtime"
	"sync"
)

// RFC5424 log message levels.
const (
	LevelNothing = iota - 1
	LevelApp     // only for pholcus
	LevelEmergency
	LevelAlert
	LevelCritical
	LevelError
	LevelWarning
	LevelNotice
	LevelInformational
	LevelDebug
)

type loggerType func() LoggerInterface

// LoggerInterface defines the behavior of a log provider.
type LoggerInterface interface {
	Init(config map[string]interface{}) error
	WriteMsg(msg string, level int) error
	Destroy()
	Flush()
}

var adapters = make(map[string]loggerType)

// Register makes a log provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, log loggerType) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	// if _, dup := adapters[name]; dup {
	// 	panic("logs: Register called twice for provider " + name)
	// }
	adapters[name] = log
}

// BeeLogger's status
const (
	NULL = iota - 1
	WORK
	REST
	CLOSE
)

// BeeLogger is default logger in beego application.
// it can contain several providers and log message into all providers.
type BeeLogger struct {
	lock                sync.RWMutex
	level               int
	enableFuncCallDepth bool
	loggerFuncCallDepth int
	asynchronous        bool
	msg                 chan *logMsg
	steal               chan *logMsg
	stealLevel          int
	stealLevelPreset    int
	outputs             map[string]LoggerInterface
	status              int
}

type logMsg struct {
	level int
	msg   string
}

// NewLogger returns a new BeeLogger.
// channellen means the number of messages in chan.
// if the buffering chan is full, logger adapters write to file or other way.
func NewLogger(channellen int64, stealLevel ...int) *BeeLogger {
	bl := new(BeeLogger)
	bl.level = LevelDebug
	bl.loggerFuncCallDepth = 2
	bl.msg = make(chan *logMsg, channellen)
	bl.outputs = make(map[string]LoggerInterface)
	bl.status = WORK
	bl.steal = make(chan *logMsg, channellen)
	if len(stealLevel) > 0 {
		bl.stealLevelPreset = stealLevel[0]
	} else {
		bl.stealLevelPreset = LevelNothing
	}
	return bl
}

func (bl *BeeLogger) Async(enable bool) *BeeLogger {
	bl.asynchronous = enable
	if enable {
		go bl.startLogger()
	}
	return bl
}

// SetLogger provides a given logger adapter into BeeLogger with config string.
func (bl *BeeLogger) SetLogger(adaptername string, config map[string]interface{}) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if log, ok := adapters[adaptername]; ok {
		lg := log()
		err := lg.Init(config)
		bl.outputs[adaptername] = lg
		if err != nil {
			fmt.Println("logs.BeeLogger.SetLogger: " + err.Error())
			return err
		}
	} else {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adaptername)
	}
	return nil
}

// remove a logger adapter in BeeLogger.
func (bl *BeeLogger) DelLogger(adaptername string) error {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	if lg, ok := bl.outputs[adaptername]; ok {
		lg.Destroy()
		delete(bl.outputs, adaptername)
		return nil
	} else {
		return fmt.Errorf("logs: unknown adaptername %q (forgotten Register?)", adaptername)
	}
}

func (bl *BeeLogger) writerMsg(loglevel int, msg string) error {
	if i, s := bl.Status(); i != WORK {
		return errors.New("The current status is " + s)
	}

	lm := new(logMsg)
	lm.level = loglevel
	if bl.enableFuncCallDepth {
		_, file, line, ok := runtime.Caller(bl.loggerFuncCallDepth)
		if !ok {
			file = "???"
			line = 0
		}
		_, filename := path.Split(file)
		lm.msg = fmt.Sprintf("[%s:%d] %s", filename, line, msg)
	} else {
		lm.msg = msg
	}

	if lm.level <= bl.stealLevel {
		bl.stealOne(lm)
	}

	if bl.asynchronous {
		bl.msg <- lm
	} else {
		bl.lock.RLock()
		defer bl.lock.RUnlock()
		for name, l := range bl.outputs {
			err := l.WriteMsg(lm.msg, lm.level)
			if err != nil {
				fmt.Println("unable to WriteMsg to adapter:", name, err)
				return err
			}
		}
	}
	return nil
}

// Set log message level.
//
// If message level (such as LevelDebug) is higher than logger level (such as LevelWarning),
// log providers will not even be sent the message.
func (bl *BeeLogger) SetLevel(l int) {
	bl.level = l
}

func (bl *BeeLogger) SetStealLevel(l int) {
	bl.stealLevel = l
}

// set log funcCallDepth
func (bl *BeeLogger) SetLogFuncCallDepth(d int) {
	bl.loggerFuncCallDepth = d
}

// get log funcCallDepth for wrapper
func (bl *BeeLogger) GetLogFuncCallDepth() int {
	return bl.loggerFuncCallDepth
}

// enable log funcCallDepth
func (bl *BeeLogger) EnableFuncCallDepth(b bool) {
	bl.enableFuncCallDepth = b
}

// start logger chan reading.
// when chan is not empty, write logs.
func (bl *BeeLogger) startLogger() {
	for bl.asynchronous || len(bl.msg) > 0 {
		select {
		case bm := <-bl.msg:
			bl.lock.RLock()
			for _, l := range bl.outputs {
				err := l.WriteMsg(bm.msg, bm.level)
				if err != nil {
					fmt.Println("ERROR, unable to WriteMsg:", err)
				}
			}
			bl.lock.RUnlock()
		}
	}
}

// Log APP level message.
func (bl *BeeLogger) App(format string, v ...interface{}) {
	if LevelApp > bl.level {
		return
	}
	msg := fmt.Sprintf("[P] "+format, v...)
	bl.writerMsg(LevelApp, msg)
}

// Log EMERGENCY level message.
func (bl *BeeLogger) Emergency(format string, v ...interface{}) {
	if LevelEmergency > bl.level {
		return
	}
	msg := fmt.Sprintf("[M] "+format, v...)
	bl.writerMsg(LevelEmergency, msg)
}

// Log ALERT level message.
func (bl *BeeLogger) Alert(format string, v ...interface{}) {
	if LevelAlert > bl.level {
		return
	}
	msg := fmt.Sprintf("[A] "+format, v...)
	bl.writerMsg(LevelAlert, msg)
}

// Log CRITICAL level message.
func (bl *BeeLogger) Critical(format string, v ...interface{}) {
	if LevelCritical > bl.level {
		return
	}
	msg := fmt.Sprintf("[C] "+format, v...)
	bl.writerMsg(LevelCritical, msg)
}

// Log ERROR level message.
func (bl *BeeLogger) Error(format string, v ...interface{}) {
	if LevelError > bl.level {
		return
	}
	msg := fmt.Sprintf("[E] "+format, v...)
	bl.writerMsg(LevelError, msg)
}

// Log WARNING level message.
func (bl *BeeLogger) Warning(format string, v ...interface{}) {
	if LevelWarning > bl.level {
		return
	}
	msg := fmt.Sprintf("[W] "+format, v...)
	bl.writerMsg(LevelWarning, msg)
}

// Log NOTICE level message.
func (bl *BeeLogger) Notice(format string, v ...interface{}) {
	if LevelNotice > bl.level {
		return
	}
	msg := fmt.Sprintf("[N] "+format, v...)
	bl.writerMsg(LevelNotice, msg)
}

// Log INFORMATIONAL level message.
func (bl *BeeLogger) Informational(format string, v ...interface{}) {
	if LevelInformational > bl.level {
		return
	}
	msg := fmt.Sprintf("[I] "+format, v...)
	bl.writerMsg(LevelInformational, msg)
}

// Log DEBUG level message.
func (bl *BeeLogger) Debug(format string, v ...interface{}) {
	if LevelDebug > bl.level {
		return
	}
	msg := fmt.Sprintf("[D] "+format, v...)
	bl.writerMsg(LevelDebug, msg)
}

// flush all chan data.
func (bl *BeeLogger) Flush() {
	for _, l := range bl.outputs {
		l.Flush()
	}
}

// close logger, flush all chan data and destroy all adapters in BeeLogger.
func (bl *BeeLogger) Close() {
	bl.lock.Lock()
	bl.status = CLOSE
	close(bl.steal)
	bl.lock.Unlock()

	bl.lock.RLock()
	defer bl.lock.RUnlock()
	for {
		if len(bl.msg) > 0 {
			bm := <-bl.msg
			for _, l := range bl.outputs {
				err := l.WriteMsg(bm.msg, bm.level)
				if err != nil {
					fmt.Println("ERROR, unable to WriteMsg (while closing logger):", err)
				}
			}
			continue
		}
		break
	}

	for _, l := range bl.outputs {
		l.Flush()
		l.Destroy()
	}
}

func (bl *BeeLogger) Rest() {
	if i, _ := bl.Status(); i != WORK {
		return
	}
	bl.SetStatus(REST)
}

func (bl *BeeLogger) GoOn() {
	if i, _ := bl.Status(); i != REST {
		return
	}
	bl.SetStatus(WORK)
}

// EnableStealOne set whether to enable steal-one.
func (bl *BeeLogger) EnableStealOne(enable bool) {
	if enable {
		bl.stealLevel = bl.stealLevelPreset
	} else {
		bl.stealLevel = LevelNothing
	}
}

// get a log message
func (bl *BeeLogger) StealOne() (level int, msg string, ok bool) {
	lm := <-bl.steal
	if lm == nil {
		return 0, "", false
	}
	return lm.level, lm.msg, true
}

func (bl *BeeLogger) stealOne(lm *logMsg) {
	bl.lock.RLock()
	defer bl.lock.RUnlock()
	if bl.status == CLOSE {
		return
	}
	bl.steal <- lm
}

func (bl *BeeLogger) Status() (int, string) {
	bl.lock.RLock()
	defer bl.lock.RUnlock()

	switch bl.status {
	case WORK:
		return WORK, "WORK"
	case REST:
		return REST, "REST"
	case CLOSE:
		return CLOSE, "CLOSE"
	}
	return NULL, "NULL"
}

func (bl *BeeLogger) SetStatus(status int) {
	bl.lock.Lock()
	defer bl.lock.Unlock()
	bl.status = status
}
