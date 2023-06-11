package factory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

type Factory struct {
	model         interface{}
	columns       []string
	times         int
	fillDataSlice []interface{}
}

func (t *Factory) NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) UseFactory(model interface{}) *Factory {

	f.model = model

	type_model := reflect.TypeOf(model)

	num := type_model.NumField()

	col_data := make([]string, num)

	for i := 0; i < num; i++ {

		col_data[i] = type_model.Field(i).Name
	}
	f.columns = col_data
	return f
}

func (u *Factory) Count(number int) *Factory {

	u.times = number
	return u
}

func (u *Factory) Create(child interface{}) {

	//反射获取child的Definition方法
	method := reflect.ValueOf(child).MethodByName("Definition")
	//调用方法
	for i := 0; i < u.times; i++ {

		mapData := method.Call(nil)[0].Interface().(map[string]interface{})

		data := make([]map[string]interface{}, 0)
		for k, v := range mapData {
			for _, v2 := range u.columns {
				if k == v2 {
					//向channel中写入数据
					data = append(data, map[string]interface{}{k: v})
				}
			}
		}

		u.fillDataSlice = append(u.fillDataSlice, data)
	}
	//打印（此时可以执行数据库插入操作）
	PrettyPrint(u.fillDataSlice)
}

func PrettyPrint(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Println(v)
		return
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "  ")
	if err != nil {
		fmt.Println(v)
		return
	}

	fmt.Println(out.String())
}
