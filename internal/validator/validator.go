package validator

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

type Validator struct {
	FieldsErr map[string] string
	NonFieldsErr []string
}

var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func(v *Validator) Matches(data string, rx *regexp.Regexp) bool {
	return rx.MatchString(data)
}

func(v *Validator) MinChars(data string, i int) bool {
	return utf8.RuneCountInString(data) >= i 
}

func (v *Validator) Valid() bool {
	return len(v.FieldsErr) == 0 && len(v.NonFieldsErr) == 0
}

func (v *Validator) AddFieldError(key string, msg string) {
	if v.FieldsErr == nil {
		v.FieldsErr = make(map[string]string)
	}
	if _,exist := v.FieldsErr[key]; !exist{
		v.FieldsErr[key] = msg
	}
}

func (v *Validator) AddNonFieldError(msg string) {
	v.NonFieldsErr = append(v.NonFieldsErr, msg )
}

func (v *Validator) CheckFields(ok bool, key string, msg string) {
	if !ok {
		v.AddFieldError(key, msg)
	}	
}

func (v *Validator) CheckLen(data string, size int) bool {
	return utf8.RuneCountInString(data) <= size 
}

func (v *Validator) CheckBlanck(data string) bool {
	return strings.TrimSpace(data) != ""
}

func  PermittedValues[T comparable](value T, permittedValues ...T) bool{
	for i:= range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}	
	return false
}

