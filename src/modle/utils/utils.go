package utils

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"reflect"
	"time"
)

const (
	//一天的毫秒数
	MILLISECONDS_OF_DAY = 24 * MILLISECONDS_OF_HOUR
	//一小时的毫秒数
	MILLISECONDS_OF_HOUR = 60 * MILLISECONDS_OF_MINUTE
	//一分钟的毫秒数
	MILLISECONDS_OF_MINUTE = 60 * MILLISECONDS_OF_SECOND
	//一秒的毫秒数
	MILLISECONDS_OF_SECOND = 1000
)

func init() {
	rand.Seed(time.Now().Unix())
	gob.Register([]interface{}{})
	gob.Register(map[int]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[interface{}]interface{}{})
	gob.Register(map[string]string{})
	gob.Register(map[int]string{})
	gob.Register(map[int]int{})
	gob.Register(map[int]int64{})
}

//通过反射的方式，将一个interface{}类型数据转换成指定的类型的数据
func Convert(baseValue interface{}, toType reflect.Type) (ret reflect.Value, ok bool) {
	defer func() {
		if err := recover(); err != nil {
			ret = reflect.ValueOf(0)
			ok = false
		}
	}()

	ret = reflect.ValueOf(baseValue).Convert(toType)
	return ret, true
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandStr(size int) string {
	b := make([]rune, size)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// 获得指定时间的凌晨时间
func Time2Midnight(tm time.Time) time.Time {
	year, month, day := tm.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, tm.Location())
}

//获取一个时间点的 x日后的凌晨时间
func NextMidnight(tm time.Time, day int) time.Time {
	midTime := Time2Midnight(tm)
	ms := midTime.UnixNano()/1e6 + int64(day*MILLISECONDS_OF_DAY)
	return Ms2Time(ms).In(tm.Location())
}

//ms毫秒时间转成utc时间(根据传入的时间的时区在返回的时间上进行时区纠正time.Time().In(tm.Location()))
func Ms2Time(ms int64) time.Time {
	return time.Unix(ms/1e3, 0).UTC()
}

// EncodeGob encode the obj to gob
func EncodeGob(obj map[interface{}]interface{}) ([]byte, error) {
	for _, v := range obj {
		gob.Register(v)
	}
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	if err != nil {
		return []byte(""), err
	}
	return buf.Bytes(), nil
}

// DecodeGob decode data to map
func DecodeGob(encoded []byte) (map[interface{}]interface{}, error) {
	buf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buf)
	var out map[interface{}]interface{}
	err := dec.Decode(&out)
	if err != nil {
		return nil, err
	}
	return out, nil
}
