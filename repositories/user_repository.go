package repositories

import (
	"database/sql"
	"errors"
	"imooc-product/common"
	"imooc-product/datamodels"
	"strconv"
)

type IUserRepository interface {
	Conn() error
	Select(userName string) (user *datamodels.User, err error)
	Insert(user *datamodels.User) (userId int64, err error)
	SelectByID(userId int64) (user *datamodels.User, err error)
}

func NewUserMangerRepository(table string, db *sql.DB) IUserRepository {
	return &UserMangerRepository{
		table:     table,
		mysqlConn: db,
	}
}

type UserMangerRepository struct {
	table     string
	mysqlConn *sql.DB
}

func (u *UserMangerRepository) SelectByID(userId int64) (user *datamodels.User, err error) {
	if err = u.Conn(); err != nil {
		return &datamodels.User{}, err
	}
	sql := "select * form " + u.table + "where ID=" + strconv.FormatInt(userId, 10)
	row, errRow := u.mysqlConn.Query(sql)
	defer row.Close()
	if errRow != nil {
		return
	}
	result := common.GetResultRow(row)
	if len(result) == 0 {
		return &datamodels.User{}, errors.New("用户不存在")
	}
	user = &datamodels.User{}
	common.DataToStructByTagSql(result, user)
	return
}

//数据库连接
func (u *UserMangerRepository) Conn() (err error) {
	if u.mysqlConn == nil {
		mysql, errMysql := common.NewMysqlConn()
		if errMysql != nil {
			return errMysql
		}
		u.mysqlConn = mysql
	}
	if u.table == "" {
		u.table = "user"
	}
	return
}

//查询操作
func (u *UserMangerRepository) Select(userName string) (user *datamodels.User, err error) {
	if userName == "" {
		return &datamodels.User{}, errors.New("条件不能为空！")
	}
	if err = u.Conn(); err != nil {
		return &datamodels.User{}, err
	}

	sql := "Select * from " + u.table + " where userName=?"
	rows, errRows := u.mysqlConn.Query(sql, userName)
	defer rows.Close()
	if errRows != nil {
		return &datamodels.User{}, errRows
	}

	result := common.GetResultRow(rows)
	if len(result) == 0 {
		return &datamodels.User{}, errors.New("用户不存在！")
	}

	user = &datamodels.User{}
	common.DataToStructByTagSql(result, user)
	return
}

//插入操作
func (u *UserMangerRepository) Insert(user *datamodels.User) (userId int64, err error) {
	if err = u.Conn(); err != nil {
		return
	}
	sql := "INSERT " + u.table + " SET nickName=?,userName=?,passWord=?"
	stmt, errStmt := u.mysqlConn.Prepare(sql)
	if errStmt != nil {
		return userId, errStmt
	}
	result, errResult := stmt.Exec(user.NickName, user.UserName, user.PassWord)
	if errResult != nil {
		return userId, errResult
	}
	return result.LastInsertId()
}
