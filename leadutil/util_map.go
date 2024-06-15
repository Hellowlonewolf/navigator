/**
 * @author zhagnxiaoping
 * @date  2024/6/15 12:01
 */
package leadutil

import (
	"encoding/json"
	"errors"
)

// Description:

type Map struct {
	data map[string]interface{}
}

func NewMap() *Map {
	return &Map{data: map[string]interface{}{}}
}

func (m *Map) SetValue(key string, value interface{}) {
	m.data[key] = value
}

func (m *Map) SetString(key string, value string) {
	m.SetValue(key, value)
}
func (m *Map) SetStringSlice(key string, value []string) {
	m.SetValue(key, value)
}
func (m *Map) SetInt(key string, value int) {
	m.SetValue(key, value)
}
func (m *Map) SetIntSlice(key string, value []int) {
	m.SetValue(key, value)
}
func (m *Map) SetInt64(key string, value int64) {
	m.SetValue(key, value)
}
func (m *Map) SetInt64Slice(key string, value []int64) {
	m.SetValue(key, value)
}
func (m *Map) SetBool(key string, value bool) {
	m.SetValue(key, value)
}
func (m *Map) SetBoolSlice(key string, value []bool) {
	m.SetValue(key, value)
}
func (m *Map) SetByte(key string, value byte) {
	m.SetValue(key, value)
}

func (m *Map) SetByteSlice(key string, value []byte) {
	m.SetValue(key, value)
}

func (m *Map) SetFloat32(key string, value float32) {
	m.SetValue(key, value)
}
func (m *Map) SetFloat32Slice(key string, value []float32) {
	m.SetValue(key, value)
}
func (m *Map) SetFloat64(key string, value float64) {
	m.SetValue(key, value)
}
func (m *Map) SetFloat64Slice(key string, value []float64) {
	m.SetValue(key, value)
}

func (m *Map) Update(val map[string]interface{}) {
	for key, value := range val {
		m.SetValue(key, value)
	}
}

func (m *Map) UpdateWithJsonByte(val []byte) {
	mapObj := map[string]interface{}{}

	err := json.Unmarshal(val, &mapObj)
	if err != nil {
		panic(err)
	}

	for key, value := range mapObj {
		m.SetValue(key, value)
	}
}

func (m *Map) GetInt(key string, defaultVal ...int) int {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.(int); ok {
		return val
	}

	panic(errors.New("can't get val:" + key))
}
func (m *Map) GetInt64(key string, defaultVal ...int64) int64 {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.(int64); ok {
		return val
	}

	panic(errors.New("can't get val:" + key))
}
func (m *Map) GetString(key string, defaultVal ...string) string {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.(string); ok {
		return val
	}

	panic(errors.New("can't get val:" + key))
}
func (m *Map) GetFloat32(key string, defaultVal ...float32) float32 {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.(float32); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}
func (m *Map) GetFloat64(key string, defaultVal ...float64) float64 {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.(float64); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}
func (m *Map) GetByte(key string, defaultVal ...byte) byte {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.(byte); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}
func (m *Map) GetByteSlice(key string, defaultVal ...[]byte) []byte {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.([]byte); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}

func (m *Map) GetBoolSlice(key string, defaultVal ...[]bool) []bool {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.([]bool); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}

func (m *Map) GetStringSlice(key string, defaultVal ...[]string) []string {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.([]string); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}
func (m *Map) GetInt64Slice(key string, defaultVal ...[]int64) []int64 {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.([]int64); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}

func (m *Map) GetIntSlice(key string, defaultVal ...[]int) []int {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.([]int); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}

func (m *Map) GetFloat64Slice(key string, defaultVal ...[]float64) []float64 {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.([]float64); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}
func (m *Map) GetFloat32Slice(key string, defaultVal ...[]float32) []float32 {
	v := m.GetValue(key, defaultVal)
	if val, ok := v.([]float32); ok {
		return val
	}
	panic(errors.New("can't get val:" + key))
}
func (m *Map) GetValue(key string, defaultVal ...interface{}) interface{} {
	if val, ok := m.data[key]; ok {
		return val
	}
	if len(defaultVal) == 1 {
		return defaultVal[0]
	}
	panic(errors.New("can't find key:" + key))
}

// 清空数据
func (m *Map) Clean() {
	for key, _ := range m.data {
		delete(m.data, key)
	}
}
