package mitch

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

func (s *Store) FindGame(id int64) *Game {
	return s.Games[id]
}

func (s *Store) FindUpload(id int64) *Upload {
	return s.Uploads[id]
}

func (s *Store) FindBuild(id int64) *Build {
	return s.Builds[id]
}

func (s *Store) ListUploadsByGame(gameID int64) []*Upload {
	return s.SelectUploads(NoSort(), Eq{"GameID": gameID})
}

func (s *Store) ListGameAdminsByGame(gameID int64) []*GameAdmin {
	return s.SelectGameAdmins(NoSort(), Eq{"GameID": gameID})
}

// selects

func (s *Store) SelectGameAdmins(vsb *ValuesSortBuilder, eq Eq) (res []*GameAdmin) {
	s.Select(&res, vsb.ForMap(s.GameAdmins), eq)
	return
}

func (s *Store) SelectUploads(vsb *ValuesSortBuilder, eq Eq) (res []*Upload) {
	s.Select(&res, vsb.ForMap(s.Uploads), eq)
	return
}
