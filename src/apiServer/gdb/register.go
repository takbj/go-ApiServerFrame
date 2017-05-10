package gdb

import (
	"github.com/astaxie/beego/orm"
)

func Register() {
	orm.RegisterModel(new(DBUser))
	orm.RegisterModel(new(DBAccount))
}
