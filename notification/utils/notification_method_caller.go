package utils

import (
	"fmt"
	"github.com/goravel/framework/contracts/notification"
	"reflect"
)

func CallToMethod(notification interface{}, methodName string, notifiable notification.Notifiable) (map[string]interface{}, error) {
	v := reflect.ValueOf(notification)
	if !v.IsValid() {
		return nil, fmt.Errorf("invalid notification value")
	}

	// 查找方法（优先指针接收者）
	method := v.MethodByName(methodName)
	if !method.IsValid() && v.CanAddr() {
		method = v.Addr().MethodByName(methodName)
	}
	if !method.IsValid() {
		return nil, fmt.Errorf("method %s not found", methodName)
	}

	// 调用方法
	results := method.Call([]reflect.Value{reflect.ValueOf(notifiable)})
	if len(results) == 0 {
		return nil, fmt.Errorf("method %s returned no values", methodName)
	}

	// 处理错误返回
	if len(results) >= 2 && !results[1].IsNil() {
		if err, ok := results[1].Interface().(error); ok {
			return nil, err
		}
		return nil, fmt.Errorf("second return of %s is not error", methodName)
	}

	// 转换第一个返回值为 map[string]interface{}
	first := results[0].Interface()
	switch data := first.(type) {
	case map[string]interface{}:
		return data, nil
	case map[string]string:
		out := make(map[string]interface{}, len(data))
		for k, v := range data {
			out[k] = v
		}
		return out, nil
	}

	// 处理结构体
	if rv := reflect.ValueOf(first); rv.Kind() == reflect.Struct {
		out := make(map[string]interface{})
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			if field.PkgPath == "" { // 仅导出字段
				out[field.Name] = rv.Field(i).Interface()
			}
		}
		return out, nil
	}

	return nil, fmt.Errorf("unsupported return type from %s", methodName)
}
