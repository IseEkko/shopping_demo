package common

import (
	"net/http"
	"strings"
)

//声明一个新的数据类型
type FileterHandle func(rw http.ResponseWriter, req *http.Request) error

//拦截器结构体
type Filter struct {
	//用于存储需要拦截的uri
	fileterMap map[string]FileterHandle
}

//Filter初始化函数
func NewFilter() *Filter {
	return &Filter{fileterMap: make(map[string]FileterHandle)}
}

//注册拦截器
func (f *Filter) RegisterFilterUri(uri string, handler FileterHandle) {
	f.fileterMap[uri] = handler
}

//根据uri获取对应的handle
func (f *Filter) GetFilterHandle(uri string) FileterHandle {
	return f.fileterMap[uri]
}

//声明新的函数类型
type WebHandle func(rw http.ResponseWriter, req *http.Request)

//执行拦截器，返回函数类型
func (f *Filter) Handle(webhandle WebHandle) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		for path, handle := range f.fileterMap {
			if strings.Contains(r.RequestURI, path) {
				//执行业务拦截逻辑
				err := handle(rw, r)
				if err != nil {
					if err != nil {
						rw.Write([]byte(err.Error()))
						return
					}
					break
				}
			}
			//执行正常注册函数
			webhandle(rw, r)
		}
	}
}
