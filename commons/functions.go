package commons

import (
	"fmt"
	"reflect"
	"runtime"
)

func Filter[T any](source []T, filteringFunc func(T) bool) []T {
	result := make([]T, 0, len(source)/5)

	for _, v := range source {
		v := v
		if filteringFunc(v) {
			result = append(result, v)
		}
	}

	return result
}

func AnyMatches[T any](source []T, matcher func(T) bool) bool {
	for _, v := range source {
		v := v
		if matcher(v) {
			return true
		}
	}
	return false
}

func MapTo[T, V any](source []T, mapper func(int, T) V) (result []V) {
	for i, s := range source {
		s := s
		result = append(result, mapper(i, s))
	}

	return result
}

func SplitByLength(str string, width int, prefix string) []string {
	lines := make([]string, 0, len(str)/width)
	from := 0
	for i := 0; true; i++ {
		to := from
		if i == 0 {
			to = from + width
			if to > len(str) {
				to = len(str)
			}
			lines = append(lines, str[from:to])
		} else {
			to = from + (width - len(prefix))
			if to > len(str) {
				to = len(str)
			}
			lines = append(lines, fmt.Sprintf("%s%s", prefix, str[from:to]))
		}
		from = to
		if from == len(str) {
			break
		}
	}
	return lines
}

func GetFunctionDescription(val any) string {
	ptr := reflect.ValueOf(val).Pointer()
	descriptor := runtime.FuncForPC(ptr)
	file, line := descriptor.FileLine(ptr)
	return fmt.Sprintf("%s:%d %s", file, line, descriptor.Name())
}

func Values[K comparable, V any](aMap map[K]V) []V {
	result := make([]V, 0, len(aMap))
	for _, v := range aMap {
		result = append(result, v)
	}
	return result
}

func Pairs[K comparable, V any](aMap map[K]V) []Pair[K, V] {
	result := make([]Pair[K, V], 0, len(aMap))
	for k, v := range aMap {
		result = append(result, Pair[K, V]{
			First:  k,
			Second: v,
		})
	}
	return result
}
