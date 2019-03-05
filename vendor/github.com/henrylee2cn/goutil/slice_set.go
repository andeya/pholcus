package goutil

// SetToStrings sets a element to the string set.
func SetToStrings(set []string, a string) []string {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromStrings removes a element from the string set.
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

// SetToInts sets a element to the int set.
func SetToInts(set []int, a int) []int {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromInts removes a element from the int set.
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

// SetToInt32s sets a element to the int32 set.
func SetToInt32s(set []int32, a int32) []int32 {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromInt32s removes a element from the int32 set.
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

// SetToInt64s sets a element to the int64 set.
func SetToInt64s(set []int64, a int64) []int64 {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromInt64s removes a element from the int64 set.
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

// SetToInterfaces sets a element to the interface{} set.
func SetToInterfaces(set []interface{}, a interface{}) []interface{} {
	for _, s := range set {
		if s == a {
			return set
		}
	}
	return append(set, a)
}

// RemoveFromInterfaces removes a element from the interface{} set.
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
