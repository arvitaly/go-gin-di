package di

import "testing"

import "github.com/gin-gonic/gin"

type B struct {
	C int
}
type A struct {
	B B
	D interface{} `di:"val1"`
	F F           `di:"-"`
	J string      `di:"conf1"`
}
type D struct {
	E string
}
type F struct {
	G int
}

func TestInject(t *testing.T) {

	var dic = NewDIC()
	var b = B{C: 49}
	dic.Provide(b)
	dic.ProvideByKey("val1", D{E: "val2"})
	dic.Provide(F{G: 100})
	var strVal = "127.0.0.1:7070"
	dic.ProvideByKey("conf1", strVal)
	var a = A{}
	dic.Inject(&a)
	if a.B.C != 49 {
		t.Error("Not injected")
	}
	if a.D.(D).E != "val2" {
		t.Error("D is not injected")
	}
	if a.F.G != 0 {
		t.Error("F was injected, but no need")
	}
	if a.J != strVal {
		t.Error("String injection was not injected")
	}
}

type CtrlInject1 struct {
	Val1 int
}

type Controller struct {
	Inject1 CtrlInject1
}

func (c Controller) Get(ctx *gin.Context) {

}

func TestGinHandleWithDI(t *testing.T) {
	var dic = NewDIC()
	dic.Provide(CtrlInject1{Val1: 144})
	var ctx = new(gin.Context)
	ctx.Set("val2", 119)
	var ready = false
	var dich = NewDICHandle(dic)
	var handle = dich.Handle(Controller.Get, func(ctx *gin.Context) {
		var val, _ = ctx.Get("val2")
		if val != 119 {
			t.Error("Not found val2 in context")
		}
		ready = true
	})
	handle(ctx)
	if !ready {
		t.Error("Response func was not called")
	}
}
func TestInvalidResponse(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("Need panic")
		}
	}()
	GinHandleWithDI(NewDIC(), Controller.Get, 123)

}
