package di

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

type DIC interface {
	Inject(s interface{})
	Provide(i interface{})
	ProvideByKey(key string, i interface{})
}
type _DIC struct {
	provides map[reflect.Type]interface{}
	maps     map[string]interface{}
}

func NewDIC() DIC {
	var dic = _DIC{}
	dic.provides = make(map[reflect.Type]interface{})
	dic.maps = make(map[string]interface{})
	return dic
}

func (dic _DIC) ProvideByKey(key string, i interface{}) {
	dic.maps[key] = i
}
func (dic _DIC) Provide(i interface{}) {
	var t = reflect.TypeOf(i)
	dic.provides[t] = i
}
func (dic _DIC) Inject(s interface{}) {
	var ctrl = reflect.TypeOf(s)
	if ctrl.Kind() != reflect.Ptr {
		panic("Object for injection should be pointer, have " + ctrl.Kind().String())
	}
	var el = ctrl.Elem()

	for fieldI := 0; fieldI < el.NumField(); fieldI++ {
		var field = el.Field(fieldI)
		var diTag = field.Tag.Get("di")
		if diTag == "-" {
			continue
		}
		var inject interface{}
		var ok bool
		if diTag != "" {
			inject, ok = dic.maps[diTag]
		} else {
			inject, ok = dic.provides[field.Type]

		}
		if ok {
			reflect.ValueOf(s).Elem().Field(fieldI).Set(reflect.ValueOf(inject))
		}
	}
}

type DICHandle struct {
	dic DIC
}

func (d DICHandle) Handle(reqFn interface{}, resFn interface{}) func(ctx *gin.Context) {
	return GinHandleWithDI(d.dic, reqFn, resFn)
}
func NewDICHandle(dic DIC) DICHandle {
	return DICHandle{dic: dic}
}

//reqFn is method of struct with 1 argument *gin.Context and any return values.
//Example: `func Get(c Controller) (ctx *gin.Context){ return 123 }`
//resFn is func, which will be called with *gin.Context and `reqFn` return values
//Example: `func(ctx *gin.Context, val int){}`
func GinHandleWithDI(dic DIC, reqFn interface{}, resFn interface{}) func(ctx *gin.Context) {
	var ctrl = reflect.TypeOf(reqFn).In(0)
	if ctrl.Kind() != reflect.Struct {
		panic("Request function should be method of struct")
	}
	return func(ctx *gin.Context) {
		var ctrlNew = reflect.New(ctrl)

		dic.Inject(ctrlNew.Interface())

		var reqFnPtr = reflect.ValueOf(reqFn).Pointer()
		var args []reflect.Value
		args = append(args, reflect.ValueOf(ctx))
		for i := 0; i < ctrl.NumMethod(); i++ {
			if ctrl.Method(i).Func.Pointer() == reqFnPtr {
				res := ctrlNew.Elem().Method(i).Call(args)
				res = append([]reflect.Value{reflect.ValueOf(ctx)}, res...)
				reflect.ValueOf(resFn).Call(res)
				break
			}
		}
	}
}
