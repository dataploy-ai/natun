/*
Copyright (c) 2022 RaptorML authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type PrimitiveType int

const (
	PrimitiveTypeUnknown PrimitiveType = iota
	PrimitiveTypeString
	PrimitiveTypeInteger
	PrimitiveTypeFloat
	PrimitiveTypeBoolean
	PrimitiveTypeTimestamp

	PrimitiveTypeStringList
	PrimitiveTypeIntegerList
	PrimitiveTypeFloatList
	PrimitiveTypeBooleanList
	PrimitiveTypeTimestampList
)

func StringToPrimitiveType(s string) PrimitiveType {
	switch strings.ToLower(s) {
	case "string", "text":
		return PrimitiveTypeString
	case "integer", "int", "int32":
		return PrimitiveTypeInteger
	case "float", "double", "float32", "float64":
		return PrimitiveTypeFloat
	case "time", "datetime", "timestamp", "time.time", "datetime.datetime":
		return PrimitiveTypeTimestamp
	case "bool", "boolean":
		return PrimitiveTypeBoolean
	case "[]string", "[]text":
		return PrimitiveTypeStringList
	case "[]integer", "[]int", "[]int64", "[]int32":
		return PrimitiveTypeIntegerList
	case "[]float", "[]double", "[]float32", "[]float64":
		return PrimitiveTypeFloatList
	case "[]bool", "[]boolean":
		return PrimitiveTypeBooleanList
	case "[]time", "[]datetime", "[]timestamp", "[]time.time":
		return PrimitiveTypeTimestampList
	default:
		return PrimitiveTypeUnknown
	}
}

func (pt PrimitiveType) Scalar() bool {
	switch pt {
	case PrimitiveTypeStringList, PrimitiveTypeIntegerList, PrimitiveTypeFloatList, PrimitiveTypeBooleanList, PrimitiveTypeTimestampList:
		return false
	default:
		return true
	}
}
func (pt PrimitiveType) Singular() PrimitiveType {
	switch pt {
	case PrimitiveTypeStringList:
		return PrimitiveTypeString
	case PrimitiveTypeIntegerList:
		return PrimitiveTypeInteger
	case PrimitiveTypeFloatList:
		return PrimitiveTypeFloat
	case PrimitiveTypeBooleanList:
		return PrimitiveTypeBoolean
	case PrimitiveTypeTimestampList:
		return PrimitiveTypeTimestamp
	default:
		return pt
	}
}
func (pt PrimitiveType) Plural() PrimitiveType {
	switch pt {
	case PrimitiveTypeString:
		return PrimitiveTypeStringList
	case PrimitiveTypeInteger:
		return PrimitiveTypeIntegerList
	case PrimitiveTypeFloat:
		return PrimitiveTypeFloatList
	case PrimitiveTypeBoolean:
		return PrimitiveTypeBooleanList
	case PrimitiveTypeTimestamp:
		return PrimitiveTypeTimestampList
	default:
		return pt
	}
}
func (pt PrimitiveType) String() string {
	switch pt {
	case PrimitiveTypeString:
		return "string"
	case PrimitiveTypeInteger:
		return "int"
	case PrimitiveTypeFloat:
		return "float"
	case PrimitiveTypeBoolean:
		return "bool"
	case PrimitiveTypeTimestamp:
		return "timestamp"
	case PrimitiveTypeStringList:
		return "[]string"
	case PrimitiveTypeIntegerList:
		return "[]int"
	case PrimitiveTypeFloatList:
		return "[]]float"
	case PrimitiveTypeBooleanList:
		return "[]bool"
	case PrimitiveTypeTimestampList:
		return "[]timestamp"
	default:
		return "(unknown)"
	}
}
func (pt PrimitiveType) Interface() any {
	if !pt.Scalar() {
		return reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(pt.Singular().Interface())), 0, 0).Interface()
	}
	switch pt {
	case PrimitiveTypeString:
		return ""
	case PrimitiveTypeInteger:
		return 0
	case PrimitiveTypeFloat:
		return float64(0)
	case PrimitiveTypeBoolean:
		return false
	case PrimitiveTypeTimestamp:
		return time.Time{}
	default:
		return pt
	}
}

func ScalarString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return strconv.FormatInt(v.UnixMicro(), 10)
	default:
		panic("unreachable")
	}
}

func ScalarFromString(val string, scalar PrimitiveType) (any, error) {
	if !scalar.Scalar() {
		return nil, fmt.Errorf("%s is not a scalar type", scalar)
	}
	switch scalar {
	case PrimitiveTypeString:
		return val, nil
	case PrimitiveTypeInteger:
		return strconv.Atoi(val)
	case PrimitiveTypeFloat:
		return strconv.ParseFloat(val, 64)
	case PrimitiveTypeBoolean:
		return strconv.ParseBool(val)
	case PrimitiveTypeTimestamp:
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, err
		}
		return time.UnixMicro(n), nil
	default:
		panic("unreachable")
	}
}

// TypeDetect detects the PrimitiveType of the value.
func TypeDetect(t any) PrimitiveType {
	reflectType := reflect.TypeOf(t)
	if reflectType == reflect.TypeOf([]any{}) {
		for _, v := range t.([]any) {
			if reflect.TypeOf(v) != reflect.TypeOf(t.([]any)[0]) {
				return PrimitiveTypeUnknown
			}
		}
		return TypeDetect(t.([]any)[0]).Plural()
	}
	return StringToPrimitiveType(reflectType.String())
}

func NormalizeAny(t any) (any, error) {
	switch v := t.(type) {
	case []any:
		if len(v) == 0 {
			return nil, nil
		}

		ret := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(v[0])), len(v), len(v))
		for i, v2 := range v {
			ret.Index(i).Set(reflect.ValueOf(v2))
		}
		t = ret.Interface()
	}
	return t, nil
}
