package token_store

import (
	"time"

	"github.com/Dataman-Cloud/rolex/plugins/account"
	log "github.com/Sirupsen/logrus"
)

type tokenStore struct {
	AccountId string
	ExpireAt  time.Time
}

type Default struct {
	account.TokenStore

	Store map[string]*tokenStore
}

func NewDefaultStore() *Default {
	return &Default{
		Store: make(map[string]*tokenStore),
	}
}

func (d *Default) Set(token, accountId string, expiredAt time.Time) error {
	log.Debugf("Set ", token, " ", accountId, " ", expiredAt)
	d.Store[token] = &tokenStore{AccountId: accountId, ExpireAt: expiredAt}
	return nil
}

func (d *Default) Get(token string) (string, error) {
	log.Debugf("Get ", token)
	if tokenStore, ok := d.Store[token]; ok {
		if tokenStore.ExpireAt.After(time.Now()) {
			log.Debugf("Get ", tokenStore.AccountId)
			return tokenStore.AccountId, nil
		} else {
			return "", TokenExpired
		}
	} else {
		return "", TokenNotFound
	}
}

func (d *Default) Del(token string) error {
	delete(d.Store, token)
	return nil
}
