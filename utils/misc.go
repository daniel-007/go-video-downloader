/**
 * Go Video Downloader
 *
 *    Copyright 2017 Tenta, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * For any questions, please contact developer@tenta.io
 *
 * misc.go: Miscellaneous utilities
 */

package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// Debug mode permits logging
var Debug = false

// Log logs a message from a running extractor
func Log(msg string, args ...interface{}) {
	if Debug {
		fmt.Printf(msg+"\n", args...)
	}
}

// StrIn returns whether needle can be found in haystack
func StrIn(needle string, haystack ...string) bool {
	for _, hay := range haystack {
		if needle == hay {
			return true
		}
	}
	return false
}

// MakeSet creates a map from the specified keys;
// the values are irrelevant, this just wants to simulate a set
func MakeSet(values []string) map[string]struct{} {
	res := make(map[string]struct{})
	for _, value := range values {
		res[value] = struct{}{}
	}
	return res
}

// FixJSONFloats converts the integral float values to ints in the unmarshalled JSON structure
func FixJSONFloats(val interface{}) interface{} {
	var fixVal func(val reflect.Value) (bool, reflect.Value)
	fixVal = func(val reflect.Value) (bool, reflect.Value) {
		if val.IsValid() {
			switch val.Type().Kind() {
			case reflect.Float64:
				f := val.Float()
				i := int(f)
				if f == float64(i) {
					return true, reflect.ValueOf(i)
				}
			case reflect.Slice:
				len := val.Len()
				for i := 0; i < len; i++ {
					el := val.Index(i)
					if changed, newElVal := fixVal(el.Elem()); changed {
						el.Set(newElVal)
					}
				}
			case reflect.Map:
				for _, key := range val.MapKeys() {
					if changed, newElVal := fixVal(val.MapIndex(key).Elem()); changed {
						val.SetMapIndex(key, newElVal)
					}
				}
			}
		}
		return false, val
	}
	_, newVal := fixVal(reflect.ValueOf(val))
	if !newVal.IsValid() {
		return nil
	}
	return newVal.Interface()
}

// SanitizeURL implements utils.py/sanitize_url
func SanitizeURL(url string) string {
	if strings.HasPrefix(url, "//") {
		return "http:" + url
	}
	return url
}
