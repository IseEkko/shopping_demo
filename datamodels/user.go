package datamodels

type User struct {
	ID       int64  `json:"id" form:"ID" sql:"ID" imooc:"ID"`
	NickName string `json:"nickName" form:"nickName" sql:"nickName" imooc:"NickName"`
	UserName string `json:"userName" form:"userName" sql:"userName" imooc:"UserName"`
	PassWord string `json:"passWord" form:"passWord" sql:"passWord" imooc:"PassWord"`
}
