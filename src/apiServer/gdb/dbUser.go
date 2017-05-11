package gdb

import (
	"fmt"
	"reflect"
	// "time"

	"modle/utils"
	"apiServer/config"

	"github.com/astaxie/beego/orm"
	"misc/mylog"
)

type DBUser struct {
	UserId int64 `canUpdate:"0" json:"UserId" orm:"pk;auto;column(id)"` //用户ID
	AccId  int64 `canUpdate:"0" json:"-" orm:"column(AccId)"`           //帐号ID

	HeadUrl string `canUpdate:"1" json:"HeadUrl" orm:"column(HeadUrl)"` //头像url
	Name    string `canUpdate:"1" json:"Name" orm:"column(Name)"`       //名称 len(Name) = [1-10]
	Sex     int8   `canUpdate:"1" json:"Sex" orm:"column(Sex)"`         //性别 1为"男"，2为女，0为未知
	Age     int    `canUpdate:"1" json:"Age" orm:"column(Age)"`         //年龄 [1-100]
	Sign    string `canUpdate:"1" json:"Sign" orm:"column(Sign)"`       //签名
}

func NewUser(accId int64) (user *DBUser, err error) {
	defer func() {
		if err != nil {
			mylog.Error( err )
		}
	}()

	cfg := config.C_UserCfg
	user = &DBUser{
		AccId:      accId,
		Age:        cfg.Age,
	}
	if user.Sex == 0 {
		user.Sex = cfg.Sex
	}

	var id int64
	id, err = orm.NewOrm().Insert(user)
	if err != nil {
		return nil, err
	}

	user.UserId = id

	return user, nil
}

func FindUserByAcc(accId int64) (*DBUser, error) {
	dbuser := &DBUser{AccId: accId}
	o := orm.NewOrm()
	if err := o.Read(dbuser, "accid"); err != nil {
		return nil, err
	}

	return dbuser, nil
}

func FindUserById(userId int64) (*DBUser, error) {
	o := orm.NewOrm()
	qs := o.QueryTable(ConstUserTableName).Filter("id", userId)
	if !qs.Exist() {
		return nil, fmt.Errorf("can't find this user")
	}

	user := &DBUser{}
	if err := qs.One(user); err != nil {
		return nil, err
	}

	return user, nil
}

func FindMultiUsersById(userIds []int64) ([]*DBUser, error) {
	o := orm.NewOrm()

	idInterface := make([]interface{}, len(userIds))
	for i, userId := range userIds {
		idInterface[i] = userId
	}
	qs := o.QueryTable(ConstUserTableName).Filter("id__in", idInterface)
	var users []*DBUser
	if _, err := qs.All(&users); err != nil {
		return nil, err
	}

	return users, nil
}

func (user *DBUser) UpdateInfo(key string, newValue interface{}, needCheck bool) (err error) {
	userValue := reflect.ValueOf(user).Elem()

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

	UpdateTableAttrByPk(ConstUserTableName, user.UserId, orm.Params{key: newValue})

	return nil
}

//自定义表名
func (user *DBUser) TableName() string {
	return ConstUserTableName
}

// 多字段索引
func (u *DBUser) TableIndex() [][]string {
	return [][]string{
		[]string{"AccId"},
		[]string{"Name"},
	}
}

// 多字段唯一键
func (u *DBUser) TableUnique() [][]string {
	return [][]string{
		[]string{"AccId"},
	}
}

// 设置引擎为 INNODB/MYISAM
func (u *DBUser) TableEngine() string {
	return "INNODB"
}
