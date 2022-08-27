package main

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"itsTasty/pkg/oidcAuth"
	"time"
)

type authSessionStorageAdapter struct {
	session *scs.SessionManager
}

func NewAuthSessionStorageManager(session *scs.SessionManager) *authSessionStorageAdapter {
	gob.Register(time.Time{})
	return &authSessionStorageAdapter{session: session}
}

func (asa *authSessionStorageAdapter) StoreString(ctx context.Context, key, value string) error {
	asa.session.Put(ctx, key, value)
	return nil
}

func (asa *authSessionStorageAdapter) GetString(ctx context.Context, key string) (string, error) {
	return asa.session.GetString(ctx, key), nil
}

func (asa *authSessionStorageAdapter) StoreTime(ctx context.Context, key string, value time.Time) error {
	asa.session.Put(ctx, key, value)
	return nil
}

func (asa *authSessionStorageAdapter) GetTime(ctx context.Context, key string) (time.Time, error) {
	return asa.session.GetTime(ctx, key), nil
}

func (asa *authSessionStorageAdapter) StoreProfile(ctx context.Context, key string, profile oidcAuth.UserProfile) error {
	asJson, err := json.Marshal(&profile)
	if err != nil {
		return fmt.Errorf("json.Marhsal : %v", err)
	}
	return asa.StoreString(ctx, key, string(asJson))
}

func (asa *authSessionStorageAdapter) GetProfile(ctx context.Context, key string) (oidcAuth.UserProfile, error) {
	p := oidcAuth.UserProfile{}
	raw, err := asa.GetString(ctx, key)
	if err != nil {
		return oidcAuth.UserProfile{}, err
	}
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return oidcAuth.UserProfile{}, fmt.Errorf("json.Unmarshal : %v", err)
	}
	return p, nil
}

func (asa *authSessionStorageAdapter) ClearEntry(ctx context.Context, key string) error {
	asa.session.Remove(ctx, key)
	return nil
}

func (asa *authSessionStorageAdapter) Destroy(ctx context.Context) error {
	return asa.session.Destroy(ctx)
}
