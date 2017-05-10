package config

var (
	C_MysqlCfg MysqlCfg
)

type MysqlCfg struct {
	MysqlAddr    string `json:"MysqlAddr"`
	MaxIdleConns int    `json:"MaxIdleConns"`
	MaxOpenConns int    `json:"MaxOpenConns"`
	Debug        bool   `json:"Debug"`
}
func init(){
	registerCfg("mysql", "config/mysqlcfg.json",	&C_MysqlCfg)
}
