package mitch

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
)

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

func (u *User) MakeGame(title string) *Game {
	s := u.Store
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	game := &Game{
		Store:  s,
		ID:     s.serial(),
		Type:   "default",
		UserID: u.ID,
		Title:  title,
	}
	s.Games[game.ID] = game
	return game
}

func (g *Game) Publish() {
	g.Published = true
}

func (g *Game) MakeUpload(title string) *Upload {
	s := g.Store
	s.writeMutex.Lock()
	defer s.writeMutex.Unlock()

	upload := &Upload{
		Store:  s,
		ID:     s.serial(),
		GameID: g.ID,
		Type:   "default",
	}
	s.Uploads[upload.ID] = upload
	return upload
}

func (u *Upload) SetAllPlatforms() {
	u.PlatformWindows = true
	u.PlatformLinux = true
	u.PlatformMac = true
}

func (u *Upload) SetZipContents() {
	u.Storage = "hosted"
	u.Filename = fmt.Sprintf("upload-%d.zip", u.ID)

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	w, err := zw.Create("hello.txt")
	must(err)
	_, err = io.WriteString(w, "Just a test file.")
	must(err)
	must(zw.Close())

	f := u.Store.UploadCDNFile(u.CDNPath(), u.Filename, buf.Bytes())
	u.Size = f.Size
}

func (u *Upload) CDNPath() string {
	return fmt.Sprintf("/uploads/%d", u.ID)
}

func (s *Store) UploadCDNFile(path string, filename string, contents []byte) *CDNFile {
	f := &CDNFile{
		Path:     path,
		Filename: filename,
		Size:     int64(len(contents)),
		Contents: contents,
	}
	s.CDNFiles[path] = f
	return f
}
