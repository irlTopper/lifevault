package utility

import "strconv"

func ReverseMap(m map[string]string) map[string]string {
	n := make(map[string]string)
	for k, v := range m {
		n[v] = k
	}
	return n
}

func InStringSlice(str string, list []string) bool {
	for _, item := range list {
		if item == str {
			return true
		}
	}

	return false
}

func InIntSlice(i int, list []int) bool {
	for _, item := range list {
		if item == i {
			return true
		}
	}

	return false
}

func InInt64Slice(i int64, list []int64) bool {
	for _, item := range list {
		if item == i {
			return true
		}
	}

	return false
}

type concatString func(string, string) string

func AppendComma(a string, b string) string {
	return a + ", " + b
}

func Reduce(operation concatString, array []int64) string {
	toReturn := ""
	for idx, value := range array {
		if idx == 0 {
			toReturn = strconv.FormatInt(value, 10)
		} else {
			toReturn = operation(toReturn, strconv.FormatInt(value, 10))
		}
	}
	return toReturn
}
