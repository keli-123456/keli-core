package http

import (
	"context"
	"fmt"
	"sync"

	"github.com/xtls/xray-core/common/protocol"
)

type userStore struct {
	mu          sync.RWMutex
	requireAuth bool
	byEmail     map[string]*protocol.MemoryUser
	byUsername  map[string]*protocol.MemoryUser
	passwords   map[string]string
}

func newUserStore(accounts map[string]string, level uint32) *userStore {
	store := &userStore{
		requireAuth: accounts != nil,
		byEmail:     make(map[string]*protocol.MemoryUser),
		byUsername:  make(map[string]*protocol.MemoryUser),
		passwords:   make(map[string]string),
	}
	for username, password := range accounts {
		account := &Account{Username: username, Password: password}
		store.storeLocked(&protocol.MemoryUser{
			Level:   level,
			Email:   username,
			Account: account,
		}, account)
	}
	return store
}

func (s *userStore) RequireAuth() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.requireAuth
}

func (s *userStore) Authenticate(username, password string) (*protocol.MemoryUser, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.passwords[username] != password {
		return nil, false
	}
	user := s.byUsername[username]
	if user == nil {
		return nil, false
	}
	return cloneMemoryUser(user), true
}

func (s *userStore) AddUser(_ context.Context, user *protocol.MemoryUser) error {
	if user == nil {
		return fmt.Errorf("user is nil")
	}
	account, ok := user.Account.(*Account)
	if !ok || account == nil {
		return fmt.Errorf("unsupported http account type %T", user.Account)
	}
	if account.Username == "" {
		return fmt.Errorf("http account username is empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byEmail[user.Email]; ok {
		return fmt.Errorf("user %s already exists", user.Email)
	}
	if _, ok := s.byUsername[account.Username]; ok {
		return fmt.Errorf("user %s already exists", account.Username)
	}
	s.requireAuth = true
	s.storeLocked(user, account)
	return nil
}

func (s *userStore) RemoveUser(_ context.Context, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.byEmail[email]
	if user == nil {
		return fmt.Errorf("user %s not found", email)
	}
	account, _ := user.Account.(*Account)
	if account != nil {
		delete(s.byUsername, account.Username)
		delete(s.passwords, account.Username)
	}
	delete(s.byEmail, email)
	return nil
}

func (s *userStore) GetUser(_ context.Context, email string) *protocol.MemoryUser {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneMemoryUser(s.byEmail[email])
}

func (s *userStore) GetUsers(_ context.Context) []*protocol.MemoryUser {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]*protocol.MemoryUser, 0, len(s.byEmail))
	for _, user := range s.byEmail {
		users = append(users, cloneMemoryUser(user))
	}
	return users
}

func (s *userStore) GetUsersCount(_ context.Context) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int64(len(s.byEmail))
}

func (s *Server) AddUser(ctx context.Context, user *protocol.MemoryUser) error {
	return s.users.AddUser(ctx, user)
}

func (s *Server) RemoveUser(ctx context.Context, email string) error {
	return s.users.RemoveUser(ctx, email)
}

func (s *Server) GetUser(ctx context.Context, email string) *protocol.MemoryUser {
	return s.users.GetUser(ctx, email)
}

func (s *Server) GetUsers(ctx context.Context) []*protocol.MemoryUser {
	return s.users.GetUsers(ctx)
}

func (s *Server) GetUsersCount(ctx context.Context) int64 {
	return s.users.GetUsersCount(ctx)
}

func (s *userStore) storeLocked(user *protocol.MemoryUser, account *Account) {
	clone := cloneMemoryUser(user)
	s.byEmail[clone.Email] = clone
	s.byUsername[account.Username] = clone
	s.passwords[account.Username] = account.Password
}

func cloneMemoryUser(user *protocol.MemoryUser) *protocol.MemoryUser {
	if user == nil {
		return nil
	}
	clone := *user
	return &clone
}
