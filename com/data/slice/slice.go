package slice

import (
	"errors"
	"reflect"
	"strings"
)

// 追加 string
func AppendS(strs []string, str string) []string {
	for _, s := range strs {
		if s == str {
			return strs
		}
	}
	return append(strs, str)
}

// 比较 string
func CompareS(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

// 比较 UTF8 string
func CompareUTF8S(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		for j := len(s2) - 1; j >= 0; j-- {
			if s1[i] == s2[j] {
				s2 = append(s2[:j], s2[j+1:]...)
				break
			}
		}
	}
	if len(s2) > 0 {
		return false
	}
	return true
}

// 是否包含某 string
func IsContainsS(sl []string, str string) bool {
	str = strings.ToLower(str)
	for _, s := range sl {
		if strings.ToLower(s) == str {
			return true
		}
	}
	return false
}

// 是否包含某 int64
func IsContainsI64(sl []int64, i int64) bool {
	for _, s := range sl {
		if s == i {
			return true
		}
	}
	return false
}

func Remove(slice_ptr interface{}, index int) error {
	if slice_ptr == nil {
		return errors.New("slice ptr is nil!")
	}
	slicePtrValue := reflect.ValueOf(slice_ptr)
	if slicePtrValue.Type().Kind() != reflect.Ptr {
		return errors.New("should be slice ptr!")
	}
	sliceValue := slicePtrValue.Elem()
	if sliceValue.Type().Kind() != reflect.Slice {
		return errors.New("should be slice ptr!")
	}
	if index < 0 || index >= sliceValue.Len() {
		return errors.New("index out of range!")
	}
	sliceValue.Set(reflect.AppendSlice(sliceValue.Slice(0, index), sliceValue.Slice(index+1, sliceValue.Len())))
	return nil
}
