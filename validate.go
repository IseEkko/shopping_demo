package main

import (
	"errors"
	"fmt"
	"imooc-product/common"
	"imooc-product/encrypt"
	"net/http"
)

func Check(w http.ResponseWriter, r *http.Request) {
	//执行正常业务逻辑
	fmt.Println("执行check")
}

//统一验证拦截器，每一个接口都需要提前验证
func Auth(w http.ResponseWriter, r *http.Request) error {
	//添加基于cookie的权限验证
	fmt.Println("执行验证！")
	err := CheckUserInfo(r)
	if err != nil {
		return errors.New("校验错误")
	}
	return nil
}

func CheckUserInfo(r *http.Request) error {
	//获取Uid：cookie
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		return errors.New("用户Uid cookie 获取失败")
	}
	//获取用户加密串
	signCookie, err := r.Cookie("sign")
	if err != nil {
		return errors.New("用户加密串 Cookie 获取失败！")
	}
	//对信息进行加密
	fmt.Println(signCookie.Value)
	signByte, err := encrypt.DePwdCode(signCookie.Value)
	fmt.Println(signByte)
	if err != nil {
		return errors.New("加密串已被篡改")
	}
	if CheckInfo(uidCookie.Value, string(signByte)) {
		return nil
	}
	return errors.New("身份校验失败")
}

//自定义逻辑判断
func CheckInfo(checkStr string, signStr string) bool {
	if checkStr == signStr {
		return true
	}
	return false
}
func main() {
	//1.过滤器
	fileter := common.NewFilter()
	//注册拦截器
	fileter.RegisterFilterUri("/check", Auth)
	//2.启动服务
	http.HandleFunc("/check", fileter.Handle(Check))
	//启动服务
	http.ListenAndServe(":8083", nil)
}
