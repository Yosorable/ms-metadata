package utils

import (
	"fmt"
	"regexp"
)

func CheckFieldName(name string) (succ bool) {
	match, err := regexp.Match("^[a-z][a-z|_|0-9]*$", []byte(name))
	if err != nil {
		panic(err)
	}
	succ = match
	return
}

func PackFieldNameForSql(name string) string {
	return fmt.Sprintf("`%s`", name)
}
