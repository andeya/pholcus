package goutil

import "strconv"

// StringsToBools converts string slice to bool slice.
func StringsToBools(a []string) ([]bool, error) {
	r := make([]bool, len(a))
	for k, v := range a {
		i, err := strconv.ParseBool(v)
		if err != nil {
			return nil, err
		}
		r[k] = i
	}
	return r, nil
}

// StringsToFloat32s converts string slice to float32 slice.
func StringsToFloat32s(a []string) ([]float32, error) {
	r := make([]float32, len(a))
	for k, v := range a {
		i, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return nil, err
		}
		r[k] = float32(i)
	}
	return r, nil
}

// StringsToFloat64s converts string slice to float64 slice.
func StringsToFloat64s(a []string) ([]float64, error) {
	r := make([]float64, len(a))
	for k, v := range a {
		i, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}
		r[k] = i
	}
	return r, nil
}

// StringsToInts converts string slice to int slice.
func StringsToInts(a []string) ([]int, error) {
	r := make([]int, len(a))
	for k, v := range a {
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		r[k] = i
	}
	return r, nil
}

// StringsToInt64s converts string slice to int64 slice.
func StringsToInt64s(a []string) ([]int64, error) {
	r := make([]int64, len(a))
	for k, v := range a {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		r[k] = i
	}
	return r, nil
}

// StringsToInt32s converts string slice to int32 slice.
func StringsToInt32s(a []string) ([]int32, error) {
	r := make([]int32, len(a))
	for k, v := range a {
		i, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return nil, err
		}
		r[k] = int32(i)
	}
	return r, nil
}

// StringsToInt16s converts string slice to int16 slice.
func StringsToInt16s(a []string) ([]int16, error) {
	r := make([]int16, len(a))
	for k, v := range a {
		i, err := strconv.ParseInt(v, 10, 16)
		if err != nil {
			return nil, err
		}
		r[k] = int16(i)
	}
	return r, nil
}

// StringsToInt8s converts string slice to int8 slice.
func StringsToInt8s(a []string) ([]int8, error) {
	r := make([]int8, len(a))
	for k, v := range a {
		i, err := strconv.ParseInt(v, 10, 8)
		if err != nil {
			return nil, err
		}
		r[k] = int8(i)
	}
	return r, nil
}

// StringsToUint8s converts string slice to uint8 slice.
func StringsToUint8s(a []string) ([]uint8, error) {
	r := make([]uint8, len(a))
	for k, v := range a {
		i, err := strconv.ParseUint(v, 10, 8)
		if err != nil {
			return nil, err
		}
		r[k] = uint8(i)
	}
	return r, nil
}

// StringsToUint16s converts string slice to uint16 slice.
func StringsToUint16s(a []string) ([]uint16, error) {
	r := make([]uint16, len(a))
	for k, v := range a {
		i, err := strconv.ParseUint(v, 10, 16)
		if err != nil {
			return nil, err
		}
		r[k] = uint16(i)
	}
	return r, nil
}

// StringsToUint32s converts string slice to uint32 slice.
func StringsToUint32s(a []string) ([]uint32, error) {
	r := make([]uint32, len(a))
	for k, v := range a {
		i, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			return nil, err
		}
		r[k] = uint32(i)
	}
	return r, nil
}

// StringsToUint64s converts string slice to uint64 slice.
func StringsToUint64s(a []string) ([]uint64, error) {
	r := make([]uint64, len(a))
	for k, v := range a {
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		r[k] = uint64(i)
	}
	return r, nil
}

// StringsToUints converts string slice to uint slice.
func StringsToUints(a []string) ([]uint, error) {
	r := make([]uint, len(a))
	for k, v := range a {
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, err
		}
		r[k] = uint(i)
	}
	return r, nil
}

// StringsConvert converts the string slice to a new slice using fn.
// If fn returns error, exit the conversion and return the error.
func StringsConvert(a []string, fn func(string) (string, error)) ([]string, error) {
	ret := make([]string, len(a))
	for i, s := range a {
		r, err := fn(s)
		if err != nil {
			return nil, err
		}
		ret[i] = r
	}
	return ret, nil
}

// StringsConvertMap converts the string slice to a new map using fn.
// If fn returns error, exit the conversion and return the error.
func StringsConvertMap(a []string, fn func(string) (string, error)) (map[string]string, error) {
	ret := make(map[string]string, len(a))
	for _, s := range a {
		r, err := fn(s)
		if err != nil {
			return nil, err
		}
		ret[s] = r
	}
	return ret, nil
}

// IntersectStrings calculate intersection of two sets.
func IntersectStrings(set1, set2 []string) []string {
	var intersect []string
	var long, short = set1, set2
	if len(set1) < len(set2) {
		long, short = set2, set1
	}

	buf := make([]string, len(short))
	copy(buf, short)
	short = buf

	for _, m := range long {
		if len(short) == 0 {
			break
		}
		for j, n := range short {
			if m == n {
				intersect = append(intersect, n)
				short = short[:j+copy(short[j:], short[j+1:])]
				break
			}
		}
	}
	return intersect
}

// StringsDistinct creates a string set that
// removes the same elements and returns them in their original order.
func StringsDistinct(a []string) (set []string) {
	m := make(map[string]bool, len(a))
	set = make([]string, 0, len(a))
	for _, s := range a {
		if m[s] {
			continue
		}
		set = append(set, s)
		m[s] = true
	}
	return set
}

// SetToStrings sets a element to the string set.
func SetToStrings(set []string, a string) []string {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromStrings removes the first element from the string set.
func RemoveFromStrings(set []string, a string) []string {
	for i, s := range set {
		if s == a {
			return append(set[:i], set[i+1:]...)
		}
	}
	return set
}

// RemoveAllFromStrings removes all the a element from the string set.
func RemoveAllFromStrings(set []string, a string) []string {
	length := len(set)
	for {
		set = RemoveFromStrings(set, a)
		if length == len(set) {
			return set
		}
		length = len(set)
	}
}

// IntsDistinct creates a int set that
// removes the same elements and returns them in their original order.
func IntsDistinct(a []int) (set []int) {
	m := make(map[int]bool, len(a))
	set = make([]int, 0, len(a))
	for _, s := range a {
		if m[s] {
			continue
		}
		set = append(set, s)
		m[s] = true
	}
	return set
}

// SetToInts sets a element to the int set.
func SetToInts(set []int, a int) []int {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromInts removes the first element from the int set.
func RemoveFromInts(set []int, a int) []int {
	for i, s := range set {
		if s == a {
			return append(set[:i], set[i+1:]...)
		}
	}
	return set
}

// RemoveAllFromInts removes all the a element from the int set.
func RemoveAllFromInts(set []int, a int) []int {
	length := len(set)
	for {
		set = RemoveFromInts(set, a)
		if length == len(set) {
			return set
		}
		length = len(set)
	}
}

// Int32sDistinct creates a int32 set that
// removes the same element32s and returns them in their original order.
func Int32sDistinct(a []int32) (set []int32) {
	m := make(map[int32]bool, len(a))
	set = make([]int32, 0, len(a))
	for _, s := range a {
		if m[s] {
			continue
		}
		set = append(set, s)
		m[s] = true
	}
	return set
}

// SetToInt32s sets a element to the int32 set.
func SetToInt32s(set []int32, a int32) []int32 {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromInt32s removes the first element from the int32 set.
func RemoveFromInt32s(set []int32, a int32) []int32 {
	for i, s := range set {
		if s == a {
			return append(set[:i], set[i+1:]...)
		}
	}
	return set
}

// RemoveAllFromInt32s removes all the a element from the int32 set.
func RemoveAllFromInt32s(set []int32, a int32) []int32 {
	length := len(set)
	for {
		set = RemoveFromInt32s(set, a)
		if length == len(set) {
			return set
		}
		length = len(set)
	}
}

// Int64sDistinct creates a int64 set that
// removes the same element64s and returns them in their original order.
func Int64sDistinct(a []int64) (set []int64) {
	m := make(map[int64]bool, len(a))
	set = make([]int64, 0, len(a))
	for _, s := range a {
		if m[s] {
			continue
		}
		set = append(set, s)
		m[s] = true
	}
	return set
}

// SetToInt64s sets a element to the int64 set.
func SetToInt64s(set []int64, a int64) []int64 {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromInt64s removes the first element from the int64 set.
func RemoveFromInt64s(set []int64, a int64) []int64 {
	for i, s := range set {
		if s == a {
			return append(set[:i], set[i+1:]...)
		}
	}
	return set
}

// RemoveAllFromInt64s removes all the a element from the int64 set.
func RemoveAllFromInt64s(set []int64, a int64) []int64 {
	length := len(set)
	for {
		set = RemoveFromInt64s(set, a)
		if length == len(set) {
			return set
		}
		length = len(set)
	}
}

// InterfacesDistinct creates a interface{} set that
// removes the same elementerface{}s and returns them in their original order.
func InterfacesDistinct(a []interface{}) (set []interface{}) {
	m := make(map[interface{}]bool, len(a))
	set = make([]interface{}, 0, len(a))
	for _, s := range a {
		if m[s] {
			continue
		}
		set = append(set, s)
		m[s] = true
	}
	return set
}

// SetToInterfaces sets a element to the interface{} set.
func SetToInterfaces(set []interface{}, a interface{}) []interface{} {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromInterfaces removes the first element from the interface{} set.
func RemoveFromInterfaces(set []interface{}, a interface{}) []interface{} {
	for i, s := range set {
		if s == a {
			return append(set[:i], set[i+1:]...)
		}
	}
	return set
}

// RemoveAllFromInterfaces removes all the a element from the interface{} set.
func RemoveAllFromInterfaces(set []interface{}, a interface{}) []interface{} {
	length := len(set)
	for {
		set = RemoveFromInterfaces(set, a)
		if length == len(set) {
			return set
		}
		length = len(set)
	}
}
