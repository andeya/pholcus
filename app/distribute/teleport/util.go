package teleport

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"strconv"
)

// MakeHash converts a string to a hash value.
func MakeHash(s string) string {
	const IEEE = 0xedb88320
	var IEEETable = crc32.MakeTable(IEEE)
	hash := fmt.Sprintf("%x", crc32.Checksum([]byte(s), IEEETable))
	return hash
}

// HashString encodes a string to a 64-bit hash value.
func HashString(encode string) uint64 {
	hash := fnv.New64()
	hash.Write([]byte(encode))
	return hash.Sum64()
}

// MakeUnique generates a unique fingerprint for an object (method 1).
func MakeUnique(obj interface{}) string {
	baseString, _ := json.Marshal(obj)
	return strconv.FormatUint(HashString(string(baseString)), 10)
}

// MakeMd5 generates an MD5 fingerprint for an object (method 2).
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
