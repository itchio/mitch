package mitch

import (
	"fmt"
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
	u.SetZipContentsCustom(func(ac *ArchiveContext) {
		ac.Entry("hello.txt").String("Just a test file")
	})
}

func (u *Upload) SetZipContentsCustom(f func(ac *ArchiveContext)) {
	ac := &ArchiveContext{
		Entries: make(map[string]*ArchiveEntry),
		Name:    fmt.Sprintf("upload-%d.zip", u.ID),
	}
	f(ac)
	u.SetHostedContents(ac.Name, ac.CompressZip())
}

func (u *Upload) SetHostedContents(filename string, contents []byte) {
	u.Storage = "hosted"
	u.Filename = filename
	f := u.Store.UploadCDNFile(u.CDNPath(), u.Filename, contents)
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
