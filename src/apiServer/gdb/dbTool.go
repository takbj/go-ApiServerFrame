package gdb

import (
	"errors"
	"misc/mylog"

	"github.com/astaxie/beego/orm"
)

const (
	ConstAccountTableName  string = "account"
	ConstUserTableName     string = "user"
)

//根据指定表的主键修改属性字段
func UpdateTableAttrByPk(table string, pk int64, attrs orm.Params) (err error) {
	defer func() {
		mylog.PrintPanicStack()
		if err != nil {
			mylog.Error("UpdateTableAttrByPk error: table,pk,attrs=", table, pk, attrs, ",excute err:", err)
		}
	}()
	o := orm.NewOrm()
	qs := o.QueryTable(table).Filter("id", pk)
	if !qs.Exist() {
		return errors.New("can not found pk info")
	}
	_, err = qs.Update(attrs)
	if err != nil {
		return err
	}
	return nil
}
