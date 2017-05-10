package redis

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"modle/session"
	"modle/utils"

	"github.com/garyburd/redigo/redis"
)

var (
	MaxPoolSize = 100
	DBNum       = 0
	Session     = "session:"
	redispder   = &Provider{}
)

func init() {
	session.Register("redis", redispder)
}

// SessionStore redis session store
type SessionStore struct {
	sid    string
	lock   sync.RWMutex
	values map[interface{}]interface{}
}

// Set value in redis session.
// it is temp value in map.
func (rs *SessionStore) Set(key, value interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values[key] = value
	return nil
}

// Get value from redis session
func (rs *SessionStore) Get(key interface{}) interface{} {
	rs.lock.RLock()
	defer rs.lock.RUnlock()
	if v, ok := rs.values[key]; ok {
		return v
	}
	return nil
}

// Delete value in redis session
func (rs *SessionStore) Delete(key interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	delete(rs.values, key)
	return nil
}

// Flush clear all values in redis session
func (rs *SessionStore) Flush() error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values = make(map[interface{}]interface{})
	return nil
}

// SessionID get session id of this redis session store
func (rs *SessionStore) SessionID() string {
	return rs.sid
}

// SessionUpdate save redis session values to database.
// must call this method to save values to database.
func (rs *SessionStore) SessionUpdate(provider interface{}) {
	rp, ok := provider.(*Provider)
	if !ok {
		return
	}
	c := rp.poollist.Get()
	defer c.Close()

	b, err := utils.EncodeGob(rs.values)
	if err != nil {
		return
	}

	sess_sid := Session + rs.sid
	c.Send("MULTI")
	c.Send("SET", sess_sid, string(b))
	c.Send("ZADD", Session, time.Now().Unix(), sess_sid)
	c.Do("EXEC")
}

// Provider redis session provider
type Provider struct {
	maxlifetime int64
	savePath    string
	poolsize    int
	password    string
	dbNum       int
	poollist    *redis.Pool
}

// SessionInit init redis session
// savepath like redis server addr,pool size,password,dbnum
// e.g. 127.0.0.1:6379,100,astaxie,0
func (rp *Provider) SessionInit(maxlifetime int64, savePath string) error {
	rp.maxlifetime = maxlifetime
	configs := strings.Split(savePath, ",")
	if len(configs) > 0 {
		rp.savePath = configs[0]
	}
	if len(configs) > 1 {
		poolsize, err := strconv.Atoi(configs[1])
		if err != nil || poolsize <= 0 {
			rp.poolsize = MaxPoolSize
		} else {
			rp.poolsize = poolsize
		}
	} else {
		rp.poolsize = MaxPoolSize
	}
	if len(configs) > 2 {
		rp.password = configs[2]
	}
	if len(configs) > 3 {
		dbnum, err := strconv.Atoi(configs[3])
		if err != nil || dbnum < 0 {
			rp.dbNum = DBNum
		} else {
			rp.dbNum = dbnum
		}
	} else {
		rp.dbNum = DBNum
	}
	rp.poollist = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", rp.savePath)
		if err != nil {
			return nil, err
		}
		if rp.password != "" {
			if _, err := c.Do("AUTH", rp.password); err != nil {
				c.Close()
				return nil, err
			}
		}
		_, err = c.Do("SELECT", rp.dbNum)
		if err != nil {
			c.Close()
			return nil, err
		}
		return c, err
	}, rp.poolsize)

	return rp.poollist.Get().Err()
}

// SessionCreate redis session by sid
func (rp *Provider) SessionCreate(sid string) (session.Store, error) {
	c := rp.poollist.Get()
	defer c.Close()

	sess_sid := Session + sid
	c.Send("MULTI")
	c.Send("SET", sess_sid, "")
	c.Send("ZADD", Session, time.Now().Unix(), sess_sid)
	if _, err := c.Do("EXEC"); err != nil {
		return nil, err
	}

	kv := make(map[interface{}]interface{})
	rs := &SessionStore{sid: sid, values: kv}
	return rs, nil
}

// SessionGet get redis session by sid
func (rp *Provider) SessionGet(sid string) (session.Store, error) {
	c := rp.poollist.Get()
	defer c.Close()

	sess_sid := Session + sid
	kvs, err := redis.String(c.Do("GET", sess_sid))
	if err != nil {
		return nil, err
	}
	var kv map[interface{}]interface{}
	if len(kvs) == 0 {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = utils.DecodeGob([]byte(kvs))
		if err != nil {
			return nil, err
		}
	}

	rs := &SessionStore{sid: sid, values: kv}
	return rs, nil
}

// SessionExist check redis session exist
func (rp *Provider) SessionExist(sid string) bool {
	c := rp.poollist.Get()
	defer c.Close()

	sess_sid := Session + sid
	if existed, err := redis.Int(c.Do("EXISTS", sess_sid)); err != nil || existed == 0 {
		return false
	}
	return true
}

// SessionDestroy delete redis session by sid
func (rp *Provider) SessionDestroy(sid string) error {
	c := rp.poollist.Get()
	defer c.Close()

	sess_sid := Session + sid
	c.Send("MULTI")
	c.Send("ZREM", Session, sess_sid)
	c.Send("DEL", sess_sid)
	if _, err := c.Do("EXEC"); err != nil {
		return err
	}

	return nil
}

// SessionGC delete expired values in redis session
func (rp *Provider) SessionGC() []string {
	c := rp.poollist.Get()
	defer c.Close()

	tokenKeys := make([]string, 0)

	sids, err := redis.Strings(c.Do("ZRANGEBYSCORE", Session, "-inf", time.Now().Unix()-rp.maxlifetime))
	if err == nil {
		tokenKeys = sids
	}

	return tokenKeys
}

// SessionAll count values in redis session
func (rp *Provider) SessionAll() int64 {
	c := rp.poollist.Get()
	defer c.Close()

	total, err := redis.Int64(c.Do("ZCOUNT", Session, "-inf", "+inf"))
	if err != nil {
		return 0
	}
	return total
}
