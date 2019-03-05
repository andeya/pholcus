# goutil [![report card](https://goreportcard.com/badge/github.com/henrylee2cn/goutil?style=flat-square)](http://goreportcard.com/report/henrylee2cn/goutil) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/henrylee2cn/goutil)

Common and useful utils for the Go project development.

## 1. Inclusion criteria

- Only rely on the Go standard package
- Functions or lightweight packages
- Non-business related general tools

## 2. Contents

- [BitSet](#bitset) A bit set
- [Calendar](#calendar) Chinese Lunar Calendar, Solar Calendar and cron time rules
- [Cmder](#cmder) Cmder exec cmd and catch the result
- [CoarseTime](#coarsetime) Current time truncated to the nearest 100ms
- [Errors](#errors) Improved errors package.
- [Graceful](#graceful) Shutdown or reboot current process gracefully.
- [GoPool](#gopool) Goroutines' pool
- [HTTPBody](#httpbody) HTTP body builder
- [ResPool](#respool) Resources' pool
- [Workshop](#workshop) Non-blocking asynchronous multiplex resource pool
- [Password](#password) Check password
- [Various](#various) Various small functions


## 3. UtilsAPI

### BitSet

A bit set.

- import it

    ```go
    "github.com/henrylee2cn/goutil/bitset"
    ```

- New creates a bit set object.

    ```go
    func New(init ...byte) *BitSet
    ```

- NewFromHex creates a bit set object from hex string.

    ```go
    func NewFromHex(s string) (*BitSet, error)
    ```

- Not returns ^b.

    ```go
    func (b *BitSet) Not() *BitSet
    ```

- And returns all the "AND" bit sets.
<br>Notes:
<br>If the bitSets are empty, returns b.

    ```go
    func (b *BitSet) And(bitSets ...*BitSet) *BitSet
    ```

- Or returns all the "OR" bit sets.
<br>Notes:
<br>If the bitSets are empty, returns b.

    ```go
    func (b *BitSet) Or(bitSet ...*BitSet) *BitSet
    ```

- Xor returns all the "XOR" bit sets.
<br>Notes:
<br>If the bitSets are empty, returns b.

    ```go
    func (b *BitSet) Xor(bitSet ...*BitSet) *BitSet
    ```


- AndNot returns all the "&^" bit sets.
<br>Notes:
<br>If the bitSets are empty, returns b.

    ```go
    func (b *BitSet) AndNot(bitSet ...*BitSet) *BitSet
    ```

- Set sets the bit bool value on the specified offset,
and returns the value of before setting.
<br>Notes:
<br>0 means the 1st bit, -1 means the bottom 1th bit, -2 means the bottom 2th bit and so on;
<br>If offset>=len(b.set), automatically grow the bit set;
<br>If the bit offset is out of the left range, returns error.

    ```go
    func (b *BitSet) Set(offset int, value bool) (bool, error)
    ```

- Get gets the bit bool value on the specified offset.
<br>Notes:
<br>0 means the 1st bit, -1 means the bottom 1th bit, -2 means the bottom 2th bit and so on;
<br>If offset>=len(b.set), returns false.

    ```go
    func (b *BitSet) Get(offset int) bool
    ```
- Range calls f sequentially for each bit present in the bit set.
<br>If f returns false, range stops the iteration.

    ```go
    func (b *BitSet) Range(f func(offset int, truth bool) bool) 
    ```

- Count counts the amount of bit set to 1 within the specified range of the bit set.
<br>Notes:
<br>0 means the 1st bit, -1 means the bottom 1th bit, -2 means the bottom 2th bit and so on.

    ```go
    func (b *BitSet) Count(start, end int) int
    ```

- Clear clears the bit set.

    ```go
    func (b *BitSet) Clear()
    ```

- Size returns the bits size.

    ```go
    func (b *BitSet) Size() int
    ```

- Bytes returns the bit set copy bytes.

    ```go
    func (b *BitSet) Bytes() []byte
    ```

- Binary returns the bit set by binary type.
<br>Notes:
<br>Paramter sep is the separator between chars.

    ```go
    func (b *BitSet) Binary(sep string) string
    ```

- String returns the bit set by hex type.

    ```go
    func (b *BitSet) String() string
    ```

- Sub returns the bit subset within the specified range of the bit set.
<br>Notes:
<br>0 means the 1st bit, -1 means the bottom 1th bit, -2 means the bottom 2th bit and so on.

    ```go
    func (b *BitSet) Sub(start, end int) *BitSet
    ```

### Calendar

Chinese Lunar Calendar, Solar Calendar and cron time rules.

- import it

    ```go
    "github.com/henrylee2cn/goutil/calendar"
    ```

[Calendar details](calendar/README.md)

### Cmder

Exec cmd and catch the result.


- import it

    ```go
    "github.com/henrylee2cn/goutil/cmder"
    ```

- Run exec cmd and catch the result.
<br>Waits for the given command to finish with a timeout.
<br>If the command times out, it attempts to kill the process.

    ```go
    func Run(cmdLine string, timeout ...time.Duration) *Result
    ```

### CoarseTime

The current time truncated to the nearest second.

- import it

    ```go
    "github.com/henrylee2cn/goutil/coarsetime"
    ```

- FloorTimeNow returns the current time from the range (now-100ms,now].
This is a faster alternative to time.Now().

    ```go
    func FloorTimeNow() time.Time
    ```

- CeilingTimeNow returns the current time from the range [now,now+100ms).
This is a faster alternative to time.Now().

    ```go
    func CeilingTimeNow() time.Time
    ```

### Errors

Errors is improved errors package.

- import it

    ```go
    "github.com/henrylee2cn/goutil/errors"
    ```

- New returns an error that formats as the given text.

    ```go
    func New(text string) error
    ```

- Errorf formats according to a format specifier and returns the string as a value that satisfies error.

    ```go
    func Errorf(format string, a ...interface{}) error
    ```

- Merge merges multi errors.

    ```go
    func Merge(errs ...error) error
    ```

- Append appends multiple errors to the error.

    ```go
    func Append(err error, errs ...error) error
    ```

### Graceful

Shutdown or reboot current process gracefully.

- import it

    ```go
    "github.com/henrylee2cn/goutil/graceful"
    ```

- GraceSignal open graceful shutdown or reboot signal.

    ```go
    func GraceSignal()
    ```

- SetShutdown sets the function which is called after the process shutdown,
and the time-out period for the process shutdown.
If 0<=timeout<5s, automatically use 'MinShutdownTimeout'(5s).
If timeout<0, indefinite period.
'firstSweepFunc' is first executed.
'beforeExitingFunc' is executed before process exiting.

    ```go
    func SetShutdown(timeout time.Duration, firstSweepFunc, beforeExitingFunc func() error)
    ```

- Shutdown closes all the frame process gracefully.
Parameter timeout is used to reset time-out period for the process shutdown.

    ```go
    func Shutdown(timeout ...time.Duration)
    ```

- Reboot all the frame process gracefully.
<br>Notes:
<br>Windows system are not supported!

    ```go
    func Reboot(timeout ...time.Duration)
    ```

- AddInherited adds the files and envs to be inherited by the new process.
<br>Notes:
<br>Only for reboot!
<br>Windows system are not supported!

    ```go
    func AddInherited(procFiles []*os.File, envs []*Env)
    ```

- Logger logger interface

    ```go
    Logger interface {
        Infof(format string, v ...interface{})
        Errorf(format string, v ...interface{})
    }
    ```

- SetLog resets logger

    ```go
    func SetLog(logger Logger)
    ```

### GoPool

GoPool is a Goroutines pool. It can control concurrent numbers, reuse goroutines.

- import it

    ```go
    "github.com/henrylee2cn/goutil/pool"
    ```

- GoPool executes concurrently incoming function via a pool of goroutines in
FILO order, i.e. the most recently stopped goroutine will execute the next
incoming function.
Such a scheme keeps CPU caches hot (in theory).

    ```go
    type GoPool struct {
        // Has unexported fields.
    }
    ```
        
- NewGoPool creates a new *GoPool.
If maxGoroutinesAmount<=0, will use default value.
If maxGoroutineIdleDuration<=0, will use default value.

    ```go
    func NewGoPool(maxGoroutinesAmount int, maxGoroutineIdleDuration time.Duration) *GoPool
    ```

- Go executes the function via a goroutine.
If returns non-nil, the function cannot be executed because exceeded maxGoroutinesAmount limit.

    ```go
    func (gp *GoPool) Go(fn func()) error
    ```

- TryGo tries to execute the function via goroutine.
If there are no concurrent resources, execute it synchronously.

    ```go
    func (gp *GoPool) TryGo(fn func())
    ```

- Stop starts GoPool.
If calling 'Go' after calling 'Stop', will no longer reuse goroutine.

    ```go
    func (gp *GoPool) Stop()
    ```

### HTTPBody

HTTP body builder.

- import it

    ```go
    "github.com/henrylee2cn/goutil/httpbody"
    ```

- NewFormBody returns form request content type and body reader.
<br> NOTE:
<br>  @values format: \<fieldName,[value]\>
<br>  @files format: \<fieldName,[fileName]\>

    ```go
    func NewFormBody(values, files url.Values) (contentType string, bodyReader io.Reader, err error)
    ```

- NewFormBody2 returns form request content type and body reader.
<br> NOTE:
<br>  @values format: \<fieldName,[value]\>
<br>  @files format: \<fieldName,[File]\>

    ```go
    func NewFormBody2(values url.Values, files Files) (contentType string, bodyReader io.Reader)
    ```

- NewFile creates a file for HTTP form.

    ```go
    func NewFile(name string, bodyReader io.Reader) File
    ```

- NewJSONBody returns JSON request content type and body reader.

    ```go
    NewJSONBody(v interface{}) (contentType string, bodyReader io.Reader, err error)
    ```

- NewXMLBody returns XML request content type and body reader.

    ```go
    NewXMLBody(v interface{}) (contentType string, bodyReader io.Reader, err error)
    ```

### ResPool

ResPool is a high availability/high concurrent resource pool, which automatically manages the number of resources.
So it is similar to database/sql's db pool.

- import it

    ```go
    "github.com/henrylee2cn/goutil/pool"
    ```

- ResPool is a pool of zero or more underlying avatar(resource).
It's safe for concurrent use by multiple goroutines.
ResPool creates and frees resource automatically;
it also maintains a free pool of idle avatar(resource).

    ```go
    type ResPool interface {
        // Name returns the name.
        Name() string
        // Get returns a object in Resource type.
        Get() (Resource, error)
        // GetContext returns a object in Resource type.
        // Support context cancellation.
        GetContext(context.Context) (Resource, error)
        // Put gives a resource back to the ResPool.
        // If error is not nil, close the avatar.
        Put(Resource, error)
        // Callback callbacks your handle function, returns the error of getting resource or handling.
        // Support recover panic.
        Callback(func(Resource) error) error
        // Callback callbacks your handle function, returns the error of getting resource or handling.
        // Support recover panic and context cancellation.
        CallbackContext(context.Context, func(Resource) error) error
        // SetMaxLifetime sets the maximum amount of time a resource may be reused.
        //
        // Expired resource may be closed lazily before reuse.
        //
        // If d <= 0, resource are reused forever.
        SetMaxLifetime(d time.Duration)
        // SetMaxIdle sets the maximum number of resources in the idle
        // resource pool.
        //
        // If SetMaxIdle is greater than 0 but less than the new MaxIdle
        // then the new MaxIdle will be reduced to match the SetMaxIdle limit
        //
        // If n <= 0, no idle resources are retained.
        SetMaxIdle(n int)
        // SetMaxOpen sets the maximum number of open resources.
        //
        // If MaxIdle is greater than 0 and the new MaxOpen is less than
        // MaxIdle, then MaxIdle will be reduced to match the new
        // MaxOpen limit
        //
        // If n <= 0, then there is no limit on the number of open resources.
        // The default is 0 (unlimited).
        SetMaxOpen(n int)
        // Close closes the ResPool, releasing any open resources.
        //
        // It is rare to close a ResPool, as the ResPool handle is meant to be
        // long-lived and shared between many goroutines.
        Close() error
        // Stats returns resource statistics.
        Stats() ResPoolStats
    }
    ```

- NewResPool creates ResPool.

    ```go
    func NewResPool(name string, newfunc func(context.Context) (Resource, error)) ResPool
    ```

- Resource is a resource that can be stored in the ResPool.

    ```go
    type Resource interface {
        // SetAvatar stores the contact with resPool
        // Do not call it yourself, it is only called by (*ResPool).get, and will only be called once
        SetAvatar(*Avatar)
        // GetAvatar gets the contact with resPool
        // Do not call it yourself, it is only called by (*ResPool).Put
        GetAvatar() *Avatar
        // Close closes the original source
        // No need to call it yourself, it is only called by (*Avatar).close
        Close() error
    }
    ```

- Avatar links a Resource with a mutex, to be held during all calls into the Avatar.

    ```go
    type Avatar struct {
        // Has unexported fields.
    }
    ```

- Free releases self to the ResPool.
If error is not nil, close it.

    ```go
    func (avatar *Avatar) Free(err error)
    ```

- ResPool returns ResPool to which it belongs.

    ```go
    func (avatar *Avatar) ResPool() ResPool
    ```

- ResPools stores ResPool.

    ```go
    type ResPools struct {
        // Has unexported fields.
    }
    ```

- NewResPools creates a new ResPools.

    ```go
    func NewResPools() *ResPools
    ```

- Clean delects and close all the ResPools.

    ```go
    func (c *ResPools) Clean()
    ```

- Del delects ResPool by name, and close the ResPool.

    ```go
    func (c *ResPools) Del(name string)
    ```

- Get gets ResPool by name.

    ```go
    func (c *ResPools) Get(name string) (ResPool, bool)
    ```

- GetAll gets all the ResPools.

    ```go
    func (c *ResPools) GetAll() []ResPool
    ```

- Set stores ResPool.
If the same name exists, will close and cover it.

    ```go
    func (c *ResPools) Set(pool ResPool)
    ```

### Workshop

Non-blocking asynchronous multiplex resource pool.

Conditions of Use:
- Limited resources
- Resources can be multiplexed non-blockingly and asynchronously
- Typical application scenarios, such as connection pool for asynchronous communication

Performance:

- The longer the business is, the more obvious the performance improvement.
<br>If the service is executed for 1ms each time, the performance is improved by about 4 times;
<br>If the business is executed for 10ms each time, the performance is improved by about 28 times
- The average time spent on each operation will not change significantly,
<br>but the overall throughput is greatly improved

- import it

    ```go
    "github.com/henrylee2cn/goutil/pool"
    ```

- Type definition

    ```go
    type (
        // Worker woker interface
        // Note: Worker can not be implemented using empty structures(struct{})!
        Worker interface {
            Health() bool
            Close() error
        }
        // Workshop working workshop
        Workshop struct {
            // Has unexported fields.
        }
    )
    ```
    
- NewWorkshop creates a new workshop(non-blocking asynchronous multiplex resource pool).
<br>If maxQuota<=0, will use default value.
<br>If maxIdleDuration<=0, will use default value.
<br>Notes:
<br>Worker can not be implemented using empty structures(struct{})!

    ```go
    func NewWorkshop(maxQuota int, maxIdleDuration time.Duration, newWorkerFunc func() (Worker, error)) *Workshop
    ```

- Close wait for all the work to be completed and close the workshop.

    ```go
    func (w *Workshop) Close()
    ```

- Callback assigns a healthy worker to execute the function.

    ```go
    func (w *Workshop) Callback(fn func(Worker) error) error
    ```

- Fire marks the worker to reduce a job.
<br>If the worker does not belong to the workshop, close the worker.

    ```go
    func (w *Workshop) Fire(worker Worker)
    ```

- Hire hires a healthy worker and marks the worker to add a job.

    ```go
    func (w *Workshop) Hire() (Worker, error)
    ```

- Stats returns the current workshop stats.

    ```go
    func (w *Workshop) Stats() *WorkshopStats
    ```

### Password

Password check password.

- import it

    ```go
    "github.com/henrylee2cn/goutil/password"
    ```

- CheckPassword checks if the password matches the format requirements.

    ```go
    func CheckPassword(pw string, flag Flag, minLen int, maxLen ...int) bool
    ```

### Various

Various small functions.

- import it

    ```go
    "github.com/henrylee2cn/goutil"
    ```

- BytesToString convert []byte type to string type.

    ```go
    func BytesToString(b []byte) string
    ```

- StringToBytes convert string type to []byte type.
<br>NOTE: panic if modify the member value of the []byte.

    ```go
    func StringToBytes(s string) []byte
    ```

- SpaceInOne combines multiple consecutive space characters into one.

    ```go
    func SpaceInOne(s string) string
    ```

- NewRandom creates a new padded Encoding defined by the given alphabet string.

    ```go
    func NewRandom(alphabet string) *Random
    ```

- RandomBytes returns securely generated random bytes. It will panic if the system's secure random number generator fails to function correctly.

    ```go
    func RandomBytes(n int) []byte
    ```

- URLRandomString returns a URL-safe, base64 encoded securely generated
<br>random string. It will panic if the system's secure random number generator
<br>fails to function correctly.
<br>The length n must be an integer multiple of 4, otherwise the last character will be padded with `=`.

    ```go
    func URLRandomString(n int) string
    ```

- CamelString converts the accepted string to a camel string (xx_yy to XxYy)

    ```go
    func CamelString(s string) string
    ```

- SnakeString converts the accepted string to a snake string (XxYy to xx_yy)

    ```go
    func SnakeString(s string) string
    ```

- ObjectName gets the type name of the object

    ```go
    func ObjectName(obj interface{}) string
    ```

- GetCallLine gets caller line information.

    ```go
    func GetCallLine(calldepth int) string 
    ```

- JsQueryEscape escapes the string in javascript standard so it can be safely placed inside a URL query.

    ```go
    func JsQueryEscape(s string) string
    ```

- JsQueryUnescape does the inverse transformation of JsQueryEscape, converting %AB into the byte 0xAB and '+' into ' ' (space).
<br>It returns an error if any % is not followed by two hexadecimal digits.

    ```go
    func JsQueryUnescape(s string) (string, error)
    ```

- Map is a concurrent map with loads, stores, and deletes.
<br>It is safe for multiple goroutines to call a Map's methods concurrently.

    ```go
    type Map interface {
        // Load returns the value stored in the map for a key, or nil if no
        // value is present.
        // The ok result indicates whether value was found in the map.
        Load(key interface{}) (value interface{}, ok bool)
        // Store sets the value for a key.
        Store(key, value interface{})
        // LoadOrStore returns the existing value for the key if present.
        // Otherwise, it stores and returns the given value.
        // The loaded result is true if the value was loaded, false if stored.
        LoadOrStore(key, value interface{}) (actual interface{}, loaded bool)
        // Range calls f sequentially for each key and value present in the map.
        // If f returns false, range stops the iteration.
        Range(f func(key, value interface{}) bool)
        // Random returns a pair kv randomly.
        // If exist=false, no kv data is exist.
        Random() (key, value interface{}, exist bool)
        // Delete deletes the value for a key.
        Delete(key interface{})
        // Clear clears all current data in the map.
        Clear()
        // Len returns the length of the map.
        Len() int
    }
    ```

- RwMap creates a new concurrent safe map with sync.RWMutex.
<br>The normal Map is high-performance mapping under low concurrency conditions.

    ```go
    func RwMap(capacity ...int) Map
    ```

- AtomicMap creates a concurrent map with amortized-constant-time loads, stores, and deletes.
<br>It is safe for multiple goroutines to call a atomicMap's methods concurrently.
From go v1.9 sync.Map.

    ```go
    func AtomicMap() Map
    ```

- SelfPath gets compiled executable file absolute path.

    ```go
    func SelfPath() string
    ```

- SelfDir gets compiled executable file directory.

    ```go
    func SelfDir() string
    ```

- RelPath gets relative path.

    ```go
    func RelPath(targpath string) string
    ```

- SelfChdir switch the working path to my own path.

    ```go
    func SelfChdir()
    ```

- FileExists reports whether the named file or directory exists.

    ```go
    func FileExists(name string) bool
    ```

- SearchFile Search a file in paths.
<br>This is often used in search config file in `/etc` `~/`

    ```go
    func SearchFile(filename string, paths ...string) (fullpath string, err error)
    ```

- GrepFile like command grep -E.
<br>For example: GrepFile(`^hello`, "hello.txt").
`\n` is striped while read

    ```go
    func GrepFile(patten string, filename string) (lines []string, err error)
    ```

- WalkDirs traverses the directory, return to the relative path.
<br>You can specify the suffix.

    ```go
    func WalkDirs(targpath string, suffixes ...string) (dirlist []string)
    ```

- IsExportedOrBuiltinType is this type exported or a builtin?

    ```go
    func IsExportedOrBuiltinType(t reflect.Type) bool
    ```

- IsExportedName is this an exported - upper case - name?

    ```go
    func IsExportedName(name string) bool
    ```

- PanicTrace trace panic stack info.

    ```go
    func PanicTrace(kb int) []byte
    ```

- ExtranetIP get external IP addr.
<br>NOTE: Query IP information from the service API: http://pv.sohu.com/cityjson?ie=utf-8

    ```go
    func ExtranetIP() (ip string, err error)
    ```

- IntranetIP get internal IP addr.

    ```go
    func IntranetIP() (string, error)
    ```

- Md5 returns the MD5 checksum string of the data.
    
    ```go
    func Md5(b []byte) string
    ```

- AESEncrypt encrypts a piece of data.
<br>The cipherkey argument should be the AES key, either 16, 24, or 32 bytes
to select AES-128, AES-192, or AES-256.

    ```go
    func AESEncrypt(cipherkey, src []byte) []byte
    ```

- AESDecrypt decrypts a piece of data.
<br>The cipherkey argument should be the AES key, either 16, 24, or 32 bytes
to select AES-128, AES-192, or AES-256.

    ```go
    func AESDecrypt(cipherkey, ciphertext []byte) ([]byte, error)
    ```

- WritePidFile writes the current PID to the specified file.

    ```go
    func WritePidFile(pidFile ...string)
    ```

- SetToStrings sets a element to the string set.

    ```go
    func SetToStrings(set []string, a string) []string
    ```

- RemoveFromStrings removes a element from the string set.

    ```go
    func RemoveFromStrings(set []string, a string) []string
    ```

- RemoveAllFromStrings removes all the a element from the string set.

    ```go
    func RemoveAllFromStrings(set []string, a string) []string
    ```

- SetToInts sets a element to the int set.

    ```go
    func SetToInts(set []int, a int) []int
    ```

- RemoveFromInts removes a element from the int set.

    ```go
    func RemoveFromInts(set []int, a int) []int
    ```

- RemoveAllFromInts removes all the a element from the int set.

    ```go
    func RemoveAllFromInts(set []int, a int) []int
    ```

- SetToInt32s sets a element to the int32 set.

    ```go
    func SetToInt32s(set []int32, a int32) []int32
    ```

- RemoveFromInt32s removes a element from the int32 set.

    ```go
    func RemoveFromInt32s(set []int32, a int32) []int32
    ```

- RemoveAllFromInt32s removes all the a element from the int32 set.

    ```go
    func RemoveAllFromInt32s(set []int32, a int32) []int32
    ```

- SetToInt64s sets a element to the int64 set.

    ```go
    func SetToInt64s(set []int64, a int64) []int64
    ```

- RemoveFromInt64s removes a element from the int64 set.

    ```go
    func RemoveFromInt64s(set []int64, a int64) []int64
    ```

- RemoveAllFromInt64s removes all the a element from the int64 set.

    ```go
    func RemoveAllFromInt64s(set []int64, a int64) []int64
    ```

- SetToInterfaces sets a element to the interface{} set.

    ```go
    func SetToInterfaces(set []interface{}, a interface{}) []interface{}
    ```

- RemoveFromInterfaces removes a element from the interface{} set.

    ```go
    func RemoveFromInterfaces(set []interface{}, a interface{}) []interface{}
    ```

- RemoveAllFromInterfaces removes all the a element from the interface{} set.

    ```go
    func RemoveAllFromInterfaces(set []interface{}, a interface{}) []interface{}
    ```

- GetFirstGopath gets the first $GOPATH value.

    ```go
    func GetFirstGopath(allowAutomaticGuessing bool) (goPath string, err error)
    ```

- TarGz compresses and archives tar.gz file.

    ```go
    func TarGz(src, dst string, includePrefix bool, logOutput func(string, ...interface{}), ignoreBaseName ...string) (err error)
    ```

- TarGzTo compresses and archives tar.gz to dst writer.

    ```go
    TarGzTo(src string, dstWriter io.Writer, includePrefix bool, logOutput func(string, ...interface{}), ignoreBaseName ...string) (err error)
    ```
