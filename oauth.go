package mitch

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// The device grant (RFC 8628) and the token exchange behind it, as
// api.itch.io serves them to butler. No consent page: a test approves or
// denies a DeviceAuth directly through the store.

type DeviceAuthStatus string

const (
	DeviceAuthPending  DeviceAuthStatus = "pending"
	DeviceAuthApproved DeviceAuthStatus = "approved"
	DeviceAuthDenied   DeviceAuthStatus = "denied"
	DeviceAuthExpired  DeviceAuthStatus = "expired"
)

type DeviceAuth struct {
	Store *Store

	DeviceCode    string
	UserCode      string
	ClientID      string
	CodeChallenge string
	Status        DeviceAuthStatus
	// The OAuth code handed out by the poll once approved
	Code string
}

type OAuthCode struct {
	Store *Store

	Code          string
	ClientID      string
	CodeChallenge string
	RedirectURI   string
	UserID        int64
	Used          bool
}

const DeviceRedirectURI = "urn:itchio:poll"

// Sent as expires_in and interval; seconds
const (
	DeviceAuthExpiresIn = 600
	DeviceAuthInterval  = 1
)

func (s *Store) MakeDeviceAuth(clientID string, codeChallenge string) *DeviceAuth {
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	id := uuid.New()
	da := &DeviceAuth{
		Store:         s,
		DeviceCode:    uuid.NewString(),
		UserCode:      fmt.Sprintf("%04X-%04X", id[0:2], id[2:4]),
		ClientID:      clientID,
		CodeChallenge: codeChallenge,
		Status:        DeviceAuthPending,
	}
	s.DeviceAuths[da.DeviceCode] = da
	return da
}

func (s *Store) FindDeviceAuth(deviceCode string) *DeviceAuth {
	return s.DeviceAuths[deviceCode]
}

func (s *Store) FindDeviceAuthByUserCode(userCode string) *DeviceAuth {
	for _, da := range s.DeviceAuths {
		if da.UserCode == userCode {
			return da
		}
	}
	return nil
}

func (s *Store) FindOAuthCode(code string) *OAuthCode {
	return s.OAuthCodes[code]
}

// Approve is what the consent page does: the next poll hands out a code
// that exchanges for a key belonging to user.
func (da *DeviceAuth) Approve(user *User) {
	s := da.Store
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	if da.Status != DeviceAuthPending {
		panic(fmt.Sprintf("device auth is %s, not pending", da.Status))
	}
	oc := &OAuthCode{
		Store:         s,
		Code:          uuid.NewString(),
		ClientID:      da.ClientID,
		CodeChallenge: da.CodeChallenge,
		RedirectURI:   DeviceRedirectURI,
		UserID:        user.ID,
	}
	s.OAuthCodes[oc.Code] = oc
	da.Code = oc.Code
	da.Status = DeviceAuthApproved
}

func (da *DeviceAuth) Deny() {
	da.setStatus(DeviceAuthDenied)
}

func (da *DeviceAuth) Expire() {
	da.setStatus(DeviceAuthExpired)
}

func (da *DeviceAuth) setStatus(status DeviceAuthStatus) {
	s := da.Store
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()
	da.Status = status
}

func pkceChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func FormatAPIKey(key *APIKey) Any {
	return Any{
		"id":      key.ID,
		"user_id": key.UserID,
		"key":     key.Key,
	}
}

func registerOAuthRoutes(route func(string, coolHandler)) {
	route("/oauth/device", func(r *response) {
		r.RespondTo(RespondToMap{
			"POST": func() {
				if r.Param("client_id") == "" {
					Throw(404, "unknown client")
				}
				if r.Param("code_challenge") == "" {
					Throw(400, "code_challenge required")
				}
				if r.Param("code_challenge_method") != "S256" {
					Throw(400, "unsupported code_challenge_method")
				}
				da := r.store.MakeDeviceAuth(r.Param("client_id"), r.Param("code_challenge"))
				verificationURI := r.makeURL("/user/oauth/device")
				r.WriteJSON(Any{
					"device_code":               da.DeviceCode,
					"user_code":                 da.UserCode,
					"verification_uri":          verificationURI,
					"verification_uri_complete": verificationURI + "?code=" + da.UserCode,
					"expires_in":                DeviceAuthExpiresIn,
					"interval":                  DeviceAuthInterval,
				})
			},
		})
	})

	route("/oauth/device/poll", func(r *response) {
		r.RespondTo(RespondToMap{
			"POST": func() {
				da := r.store.FindDeviceAuth(r.Param("device_code"))
				if da == nil || da.ClientID != r.Param("client_id") {
					Throw(400, "invalid_grant")
				}
				// tests flip the status via Approve/Deny/Expire while the client polls
				r.store.writeMutex.Lock()
				status, code := da.Status, da.Code
				r.store.writeMutex.Unlock()

				switch status {
				case DeviceAuthPending:
					r.WriteJSON(Any{"status": "pending", "interval": DeviceAuthInterval})
				case DeviceAuthApproved:
					r.WriteJSON(Any{"status": "approved", "code": code})
				default:
					r.WriteJSON(Any{"status": string(status)})
				}
			},
		})
	})

	route("/oauth/token", func(r *response) {
		r.RespondTo(RespondToMap{
			"POST": func() {
				if r.Param("grant_type") != "authorization_code" {
					Throw(400, "unsupported_grant_type")
				}
				oc := r.store.FindOAuthCode(r.Param("code"))
				if oc == nil || oc.Used ||
					oc.ClientID != r.Param("client_id") ||
					oc.RedirectURI != r.Param("redirect_uri") ||
					pkceChallenge(r.Param("code_verifier")) != oc.CodeChallenge {
					Throw(400, "invalid_grant")
				}
				user := r.store.FindUser(oc.UserID)
				if user == nil {
					Throw(500, "code has no user")
				}

				r.store.writeMutex.Lock()
				oc.Used = true
				r.store.writeMutex.Unlock()

				key := user.MakeAPIKey()
				key.Key = strings.ReplaceAll(uuid.NewString(), "-", "")
				key.DeviceInfo = r.Param("device_info")

				r.WriteJSON(Any{
					"key":    FormatAPIKey(key),
					"cookie": Any{},
				})
			},
		})
	})
}
