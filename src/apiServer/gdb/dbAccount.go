package gdb

import (
	// "misc/timefix"
	"fmt"
	"reflect"
	"time"
	// "net"

	"modle/utils"

	"github.com/astaxie/beego/orm"
)


type DBAccount struct {
	AccId int64 `json:"id" orm:"pk;auto;column(id)"` //帐号ID

	AccName string `json:"AccName" orm:"column(AccName)"`
	Password string `json:"Password" orm:"column(Password)"`

	RegisterIP string `json:"RegisterIP" orm:"column(RegisterIP)"`
	RegisterTime int64 `json:"RegisterTime" orm:"column(RegisterTime)"`

	LastToken    string `json:"-" orm:"column(LastToken)"` //最后一次登录token
}

func NewAccount(accName, accPwd, ipStr string) (acc *DBAccount, err error) {
	acc = &DBAccount{
		AccName:       accName,
		Password:      accPwd,
		RegisterIP:    ipStr,
		RegisterTime:  time.Now().Unix(),
	}

	o := orm.NewOrm()
	if acc.AccId, err = o.Insert(acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func GetAccout(accName, accPwd string) (acc *DBAccount, err error) {
	acc = &DBAccount{}
	qs := orm.NewOrm().QueryTable(ConstAccountTableName).Filter("AccName", accName).Filter("Password", accPwd)
	if err = qs.One(acc); err != nil {
		return nil, err
	}

	return acc, nil
}

func (acc *DBAccount) UpdateInfo(key string, newValue interface{}, needCheck bool) (err error) {
	userValue := reflect.ValueOf(acc).Elem()

	fieldType, exist := userValue.Type().FieldByName(key)
	if !exist {
		return fmt.Errorf("unknow field:%v\n", key)
	}

	field := userValue.FieldByName(key)
	if !field.IsValid() {
		return fmt.Errorf("unknow field:%v", key)
	}

	if needCheck {
		canUpdate := fieldType.Tag.Get("canUpdate")
		if canUpdate != "1" {
			return fmt.Errorf("this field:%v can't set!", key)
		}
	}

	valueInterface, ok := utils.Convert(newValue, field.Type())
	if !ok {
		return fmt.Errorf("cant set value %v(type=%v) to field:%v, expect:%v !", newValue, reflect.ValueOf(newValue).Type(), key, field.Type())
	}
	field.Set(valueInterface)

	UpdateTableAttrByPk(ConstAccountTableName, acc.AccId, orm.Params{key: newValue})

	return nil
}

//自定义表名
func (acc *DBAccount) TableName() string {
	return ConstAccountTableName
}

// 多字段索引
func (acc *DBAccount) TableIndex() [][]string {
	return [][]string{
		[]string{"Channel", "Openid"},
	}
}

// 多字段唯一键
func (acc *DBAccount) TableUnique() [][]string {
	return [][]string{
	// []string{"id"},
	}
}

// 设置引擎为 INNODB/MYISAM
func (acc *DBAccount) TableEngine() string {
	return "INNODB"
}
