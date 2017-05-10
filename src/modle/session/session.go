package session

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	closeChannel      chan bool
	onTimeOutCallBack CallBackFunc
	provides          = make(map[string]Provider)
)

type CallBackFunc func(key string, a_sessionData interface{})

/////////////////////////////////
// Store contains all data for one session process with specific id.
type Store interface {
	Set(key, value interface{}) error   //set session value
	Get(key interface{}) interface{}    //get session value
	Delete(key interface{}) error       //delete session value
	SessionID() string                  //back current sessionID(token)
	SessionUpdate(provider interface{}) //update the resource & save data to provider
	Flush() error                       //delete all data
}

// Provider contains global session methods and saved SessionStores.
// it can operate a SessionStore by its id.
type Provider interface {
	SessionInit(gclifetime int64, config string) error
	SessionGet(sid string) (Store, error)
	SessionCreate(sid string) (Store, error)
	SessionExist(sid string) bool
	SessionDestroy(sid string) error
	SessionAll() int64 //get all active session
	SessionGC() []string
}

// Register makes a session provide available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provide
}

type ManagerConfig struct {
	Maxlifetime     int64  `json:"maxLifetime"`
	Gclifetime      int64  `json:"gclifetime"`
	ProviderConfig  string `json:"providerConfig"`
	SessionIDLength int64  `json:"sessionIDLength"`
}

// Manager contains Provider and its configuration.
type Manager struct {
	provider          Provider
	config            *ManagerConfig
	onTimeOutCallBack CallBackFunc
	lock              sync.Mutex
}

// NewManager Create new Manager with provider name and json config string.
// provider name:
// 1. memory
// 2. redis
// 3. mysql
// json config:
func NewManager(provideName string, cf *ManagerConfig, a_onTimeOutCallBack CallBackFunc) (*Manager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}

	if cf.Maxlifetime == 0 {
		cf.Maxlifetime = cf.Gclifetime
	}

	err := provider.SessionInit(cf.Maxlifetime, cf.ProviderConfig)
	if err != nil {
		return nil, err
	}

	if cf.SessionIDLength == 0 {
		cf.SessionIDLength = 16
	}

	return &Manager{
		provider:          provider,
		config:            cf,
		onTimeOutCallBack: a_onTimeOutCallBack,
	}, nil
}

func (manager *Manager) SessionCreate() (session Store, err error) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	count := 0
	var sid string
	var errs error
	for {
		// Generate a new session
		sid, errs = manager.sessionID()
		if errs != nil {
			return nil, errs
		}
		count++
		if count >= 1000 {
			return nil, fmt.Errorf("create session sid repeat error")
		}

		if !manager.provider.SessionExist(sid) {
			break
		}
	}

	session, err = manager.provider.SessionCreate(sid)
	if err != nil {
		return nil, err
	}
	return

}

//get session
func (manager *Manager) GetSessionStore(sid string) (session Store, err error) {
	if sid == "" {
		return nil, fmt.Errorf("Session is not Exist %v", sid)
	}
	return manager.provider.SessionGet(sid)
}

// SessionDestroy Destroy session .
func (manager *Manager) SessionDestroy(session Store) {
	manager.lock.Lock()
	defer manager.lock.Unlock()
	manager.provider.SessionDestroy(session.SessionID())
}

// GC Start session gc process.
// it can do gc in times after gc lifetime.
func (manager *Manager) GC() {
	time.AfterFunc(time.Duration(manager.config.Gclifetime)*time.Second, func() { manager.GC() })
	tokenKeys := manager.provider.SessionGC()
	if len(tokenKeys) > 0 && manager.onTimeOutCallBack != nil {
		for _, token := range tokenKeys {
			if token != "" {
				manager.onTimeOutCallBack(token, nil)
			}
		}
	}
}

func (manager *Manager) sessionID() (string, error) {
	b := make([]byte, manager.config.SessionIDLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", fmt.Errorf("Could not successfully read from the system CSPRNG.")
	}
	return hex.EncodeToString(b), nil
}

// SessionUpdate update session .
func (manager *Manager) SessionUpdate(session Store) {
	session.SessionUpdate(manager.provider)
}
