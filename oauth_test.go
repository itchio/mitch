package mitch

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func postForm(t *testing.T, s Server, path string, form url.Values) (int, map[string]any) {
	t.Helper()
	res, err := http.PostForm("http://"+s.Address().String()+path, form)
	require.NoError(t, err)
	defer res.Body.Close()
	var body map[string]any
	require.NoError(t, json.NewDecoder(res.Body).Decode(&body))
	return res.StatusCode, body
}

func Test_DeviceGrant(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s, err := NewServer(ctx)
	require.NoError(t, err)
	store := s.Store()
	user := store.MakeUser("Device User")

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	status, body := postForm(t, s, "/oauth/device", url.Values{
		"client_id":             {"client-1"},
		"scope":                 {"itch"},
		"code_challenge":        {pkceChallenge(verifier)},
		"code_challenge_method": {"S256"},
	})
	require.Equal(t, 200, status)
	deviceCode := body["device_code"].(string)
	userCode := body["user_code"].(string)
	assert.Contains(t, body["verification_uri_complete"], "?code="+userCode)

	poll := url.Values{"client_id": {"client-1"}, "device_code": {deviceCode}}
	status, body = postForm(t, s, "/oauth/device/poll", poll)
	require.Equal(t, 200, status)
	assert.Equal(t, "pending", body["status"])

	status, body = postForm(t, s, "/oauth/device/poll", url.Values{"client_id": {"other"}, "device_code": {deviceCode}})
	assert.Equal(t, 400, status)
	assert.Equal(t, []any{"invalid_grant"}, body["errors"])

	store.FindDeviceAuthByUserCode(userCode).Approve(user)
	status, body = postForm(t, s, "/oauth/device/poll", poll)
	require.Equal(t, 200, status)
	assert.Equal(t, "approved", body["status"])
	code := body["code"].(string)

	exchange := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {"client-1"},
		"redirect_uri":  {DeviceRedirectURI},
		"code_verifier": {verifier},
		"device_info":   {`{"platform":"test"}`},
	}
	bad := url.Values{}
	for k, v := range exchange {
		bad[k] = v
	}
	bad.Set("code_verifier", "wrong")
	status, _ = postForm(t, s, "/oauth/token", bad)
	assert.Equal(t, 400, status)

	status, body = postForm(t, s, "/oauth/token", exchange)
	require.Equal(t, 200, status)
	key := body["key"].(map[string]any)["key"].(string)
	apiKey := store.FindAPIKeysByKey(key)
	require.NotNil(t, apiKey)
	assert.Equal(t, user.ID, apiKey.UserID)
	assert.Equal(t, `{"platform":"test"}`, apiKey.DeviceInfo)

	// single use
	status, _ = postForm(t, s, "/oauth/token", exchange)
	assert.Equal(t, 400, status)

	status, body = postForm(t, s, "/oauth/device", url.Values{
		"client_id":             {"client-1"},
		"code_challenge":        {"x"},
		"code_challenge_method": {"S256"},
	})
	require.Equal(t, 200, status)
	da := store.FindDeviceAuth(body["device_code"].(string))
	da.Deny()
	status, body = postForm(t, s, "/oauth/device/poll", url.Values{"client_id": {"client-1"}, "device_code": {da.DeviceCode}})
	require.Equal(t, 200, status)
	assert.Equal(t, "denied", body["status"])
}
