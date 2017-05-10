package config

var (
	C_RedisCfg RedisCfg
)

type RedisCfg struct {
	Address     string `json:"address"`
	MaxPoolSize int    `json:"maxPoolSize"`
	Password    string `json:"password"`
	Dbnum       int    `json:"dbnum"`
}

func init() {
	registerCfg("redis", "config/rediscfg.json", &C_RedisCfg)
}
