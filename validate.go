package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"imooc-product/common"
	"imooc-product/datamodels"
	"imooc-product/encrypt"
	rabbitmq2 "imooc-product/rabbitmq"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

var hostArray = []string{"127.0.0.1", "127.0.0.1"}
var localHost = "127.0.0.1"

//数量控制接口服务器内网ip，或者SLB内网Ip
var GetOneIp = "127.0.0.1"
var GetOnePort = "8084"
var port = "8081"
var hashConsistent *common.Consistent

//rabbitmq
var rabbitMqValidate *rabbitmq2.RabbitMQ

//用来存放控制信息
type AccessControl struct {
	//用来存放用户想要存放的信息
	sourceArray map[int]interface{}
	*sync.RWMutex
}

var accessControl = &AccessControl{
	sourceArray: make(map[int]interface{}),
}

//获取指定的数据
func (m *AccessControl) GetNewRecord(uid int) interface{} {
	m.RWMutex.Lock()
	defer m.RWMutex.Unlock()
	data := m.sourceArray[uid]
	return data
}

//设置记录
func (m *AccessControl) SetNewRecord(uid int) {
	m.RWMutex.Lock()
	m.sourceArray[uid] = "hello immoc"
	m.RWMutex.Unlock()
}

func (m *AccessControl) GetDistributedRight(req *http.Request) bool {
	//获取用户id
	uid, err := req.Cookie("uid")
	if err != nil {
		return false
	}
	//采用一致性算法，根据用户id判断获取具体机器
	hostRequest, err := hashConsistent.Get(uid.Value)
	if err != nil {
		return false
	}
	//判断是否为本机
	if hostRequest == localHost {
		//执行本机读取和校验\
		return m.GetDataFromMap(uid.Value)
	} else {
		//不是本机充当代理访问数据返回结果
		return GetDataFromOtherMap(hostRequest, req)
	}
}

//获取本机map，并且处理业务逻辑，返回结构类型为bool类型
func (m *AccessControl) GetDataFromMap(uid string) (isOk bool) {
	uidInt, err := strconv.Atoi(uid)
	if err != nil {
		return false
	}
	data := m.GetNewRecord(uidInt)
	//执行逻辑判断
	if data != nil {
		return true
	}
	return
}

//获取其他节点处理结果
func GetDataFromOtherMap(host string, request *http.Request) bool {
	hostUrl := "http://" + host + ":" + port + "/checkRight"
	response, body, err := GetCurl(hostUrl, request)
	if err != nil {
		return false
	}
	if response.StatusCode == 200 {
		if string(body) == "true" {
			return true
		} else {
			return false
		}
	}
	return false
}

//模拟请求
func GetCurl(hostUrl string, request *http.Request) (response *http.Response, body []byte, err error) {
	//获取uid
	uidPre, err := request.Cookie("uid")
	if err != nil {
		return
	}
	//获取签名
	uidSign, err := request.Cookie("sign")
	if err != nil {
		return
	}
	//模拟接口访问
	client := &http.Client{}
	req, err := http.NewRequest("GET", hostUrl, nil)
	if err != nil {
		return
	}
	//手动指定，排查多余cookies
	cookieUid := &http.Cookie{Name: "uid", Value: uidPre.Value, Path: "/"}
	cookieSign := &http.Cookie{Name: "sign", Value: uidSign.Value, Path: "/"}
	req.AddCookie(cookieUid)
	req.AddCookie(cookieSign)

	//获取返回结果
	response, err = client.Do(req)
	defer response.Body.Close()
	if err != nil {
		return
	}
	body, err = ioutil.ReadAll(response.Body)
	return
}
func CheckRight(w http.ResponseWriter, r *http.Request) {
	right := accessControl.GetDistributedRight(r)
	if !right {
		w.Write([]byte("false"))
		return
	}
	w.Write([]byte("true"))
	return
}

//执行正常业务逻辑
func Check(w http.ResponseWriter, r *http.Request) {
	//执行正常业务逻辑
	fmt.Println("执行check")
	queryForm, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil || len(queryForm["productID"]) <= 0 {
		w.Write([]byte("false"))
		return
	}
	productString := queryForm["productID"][0]
	fmt.Println(productString)
	//获取用户cookie
	usercookie, err := r.Cookie("uid")
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	//1. 分布式权限验证
	right := accessControl.GetDistributedRight(r)
	if right == false {
		w.Write([]byte("false"))
	}
	//2.获取数量控制权限，防止秒杀出现超卖
	hostUrl := "http://" + GetOneIp + ":" + GetOnePort + "/getOne"
	GetCurl(hostUrl, r)
	responseValidate, validateBody, err := GetCurl(hostUrl, r)
	if err != nil {
		w.Write([]byte("false"))
		return
	}
	if responseValidate.StatusCode == 200 {
		if string(validateBody) == "true" {
			//整合下单
			//获取商品ID
			producrtID, err := strconv.ParseInt(productString, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			//获取用户ID
			userID, err := strconv.ParseInt(usercookie.Value, 10, 64)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			//3.创建消息体
			message := datamodels.NewMessage(userID, producrtID)
			//类型转换
			byteMesaage, err := json.Marshal(message)
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			//4.生成消息
			err = rabbitMqValidate.PublishSimple(string(byteMesaage))
			if err != nil {
				w.Write([]byte("false"))
				return
			}
			w.Write([]byte("true"))
			return
		}

	}

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

//身份校验
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
	//SLB 设置
	//采用hash一致性算法
	hashConsistent = common.NewConsistent()
	//从用hash一致性算法添加节点
	for _, v := range hostArray {
		hashConsistent.Add(v)
	}
	localIp, err := common.GetIntranceIp()
	if err != nil {
		fmt.Println(err)
	}
	localHost = localIp
	fmt.Println(localHost)
	rabbitMqValidate = rabbitmq2.NewRabbitMQSimple("imoocProduct")
	defer rabbitMqValidate.Destory()
	//1.过滤器
	fileter := common.NewFilter()
	//注册拦截器
	fileter.RegisterFilterUri("/check", Auth)
	fileter.RegisterFilterUri("/checkRight", Auth)
	//2.启动服务
	http.HandleFunc("/check", fileter.Handle(Check))
	http.HandleFunc("/checkRight", fileter.Handle(CheckRight))
	//启动服务
	http.ListenAndServe(":8083", nil)
}
