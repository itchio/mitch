package mitch

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Store struct {
	Users   map[int64]*User
	APIKeys map[int64]*APIKey

	idSeed     int64
	writeMutex sync.Mutex
}

func newStore() *Store {
	return &Store{
		Users:   make(map[int64]*User),
		APIKeys: make(map[int64]*APIKey),
		idSeed:  10,
	}
}

type User struct {
	Store *Store

	ID             int64
	Username       string
	DisplayName    string
	Gamer          bool
	Developer      bool
	PressUser      bool
	AllowTelemetry bool
}

type APIKey struct {
	Store *Store

	ID        int64
	UserID    int64
	Key       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Game struct {
	ID int64
}

type Upload struct {
	ID     int64
	GameID int64
}

type Build struct {
	ID       int64
	UploadID int64
}

func (s *Store) FindAPIKeysByKey(key string) *APIKey {
	for _, k := range s.APIKeys {
		if k.Key == key {
			return k
		}
	}
	return nil
}

func (s *Store) ListAPIKeysByUser(userID int64) []*APIKey {
	var res []*APIKey
	for _, k := range s.APIKeys {
		if k.UserID == userID {
			res = append(res, k)
		}
	}
	return res
}

func (s *Store) FindUser(id int64) *User {
	return s.Users[id]
}

func (s *Store) MakeUser(displayName string) *User {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	user := &User{
		Store:       s,
		ID:          s.serial(),
		Username:    s.slugify(displayName),
		DisplayName: displayName,
		Gamer:       true,
	}
	s.Users[user.ID] = user
	return user
}

func (u *User) MakeAPIKey() *APIKey {
	s := u.Store
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	apiKey := &APIKey{
		Store:  s,
		ID:     s.serial(),
		UserID: u.ID,
		Key:    fmt.Sprintf("%s-api-key", u.Username),
	}
	s.APIKeys[apiKey.ID] = apiKey
	return apiKey
}

func (s *Store) serial() int64 {
	s.idSeed += 100
	return s.idSeed
}

var (
	invalidUsernameChars = regexp.MustCompile("^[A-Za-z0-9_]")
)

func (s *Store) slugify(input string) string {
	var res = input
	res = strings.ToLower(res)
	res = invalidUsernameChars.ReplaceAllString(res, "_")
	return res
}

func FormatUser(user *User) Any {
	res := Any{
		"id":           user.ID,
		"gamer":        user.Gamer,
		"developer":    user.Developer,
		"press_user":   user.PressUser,
		"display_name": user.DisplayName,
		"username":     user.Username,
		"url":          "http://example.org",
		"cover_url":    "http://example.org",
	}
	if user.AllowTelemetry {
		res["allow_telemetry"] = true
	}
	return res
}
