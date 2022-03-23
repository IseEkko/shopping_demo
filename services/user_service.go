package services

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"imooc-product/datamodels"
	"imooc-product/repositories"
)

type IuserService interface {
	IsPwdSuccess(userName string, pwd string) (user *datamodels.User, isOK bool)
	AddUser(user *datamodels.User) (userId int64, err error)
}

type UserService struct {
	UserRepository repositories.IUserRepository
}

func (u *UserService) IsPwdSuccess(userName string, pwd string) (user *datamodels.User, isOK bool) {
	var err error
	user, err = u.UserRepository.Select(userName)
	if err != nil {
		return
	}
	isOK, _ = ValidatePassword(pwd, user.PassWord)
	if !isOK {
		return &datamodels.User{}, false
	}
	return
}

//插入用户
func (u *UserService) AddUser(user *datamodels.User) (userId int64, err error) {
	pwdbyte, errpwd := GenertePassword(user.PassWord)
	if errpwd != nil {
		return userId, errpwd
	}

	user.PassWord = string(pwdbyte)
	return u.UserRepository.Insert(user)
}

//返回加密后的密码
func GenertePassword(userPassword string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
}

//检查密码
func ValidatePassword(userPassWord string, hashed string) (isOk bool, err error) {
	if err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(userPassWord)); err != nil {
		return false, errors.New("密码比对错误")
	}
	return true, nil
}

func NewUservice(userRepository repositories.IUserRepository) IuserService {
	return &UserService{UserRepository: userRepository}
}
