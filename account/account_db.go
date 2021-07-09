package account

import (
	"fmt"
	"log"
	"sync"

	"github.com/JinlongWukong/DevLab/utils"
)

var AccountDB = AccountMap{Map: make(map[string]*Account)}

type AccountMap struct {
	Map  map[string]*Account `json:"account"`
	lock sync.RWMutex        `json:"-"`
}

type accountMapItem struct {
	Key   string
	Value *Account
}

func (m *AccountMap) Add(accountRequest AccountRequest) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	if _, exists := m.Map[accountRequest.Name]; exists {
		return fmt.Errorf("account already existed")
	} else {
		if ac, err := newAccount(accountRequest); err != nil {
			m.Map[accountRequest.Name] = ac
		} else {
			return err
		}
	}

	return nil
}

func (m *AccountMap) Modify(accountRequest AccountRequest) error {

	m.lock.Lock()
	defer m.lock.Unlock()

	if account, exists := m.Map[accountRequest.Name]; exists {
		switch accountRequest.Role {
		case RoleAdmin:
		case RoleGuest:
		default:
			return fmt.Errorf("account role not valid")
		}
		account.Role = accountRequest.Role
		account.Contract = accountRequest.Contract
	} else {
		if ac, err := newAccount(accountRequest); err != nil {
			m.Map[accountRequest.Name] = ac
		} else {
			return err
		}
	}

	return nil
}

// initialize admin when system bootup
// if not exists, create admin account with one-time random password stdout print
// if existed, update one-time password with stdout print
func (m *AccountMap) InitializeAdmin() {

	m.lock.Lock()
	defer m.lock.Unlock()

	if admin, exists := m.Map["admin"]; exists {
		admin.OneTimePass = utils.RandomString(10)
	} else {
		m.Map["admin"] = &Account{
			Name:        "admin",
			Role:        "admin",
			OneTimePass: utils.RandomString(10),
		}
	}

	log.Printf("Admin one-time random password: %v", m.Map["admin"].OneTimePass)
}

func (m *AccountMap) Get(key string) (account *Account, exists bool) {

	m.lock.RLock()
	defer m.lock.RUnlock()

	account, exists = m.Map[key]
	return
}

func (m *AccountMap) Del(key string) {

	m.lock.Lock()
	defer m.lock.Unlock()

	delete(m.Map, key)
}

// Iter iterates over the items in a concurrent map
// Each item is sent over a channel, so that
// we can iterate over the map using the builtin range keyword
func (m *AccountMap) Iter() <-chan accountMapItem {
	c := make(chan accountMapItem)

	f := func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		for k, v := range m.Map {
			c <- accountMapItem{k, v}
		}
		close(c)
	}
	go f()

	return c
}
