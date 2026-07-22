package mitch

import "time"

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

func FormatUserGameSession(s *UserGameSession) Any {
	return Any{
		"id":          s.ID,
		"game_id":     s.GameID,
		"user_id":     s.UserID,
		"seconds_run": s.SecondsRun,
		"last_run_at": s.LastRunAt,
		"crashed":     s.Crashed,
	}
}

// FormatUserGameSummary aggregates all of a user's sessions for a game,
// mirroring the itch.io summary payload.
func FormatUserGameSummary(store *Store, userID int64, gameID int64) Any {
	var secondsRun int64
	var lastRunAt time.Time
	for _, s := range store.ListUserGameSessionsByUserAndGame(userID, gameID) {
		secondsRun += s.SecondsRun
		if s.LastRunAt.After(lastRunAt) {
			lastRunAt = s.LastRunAt
		}
	}
	res := Any{
		"seconds_run": secondsRun,
	}
	if !lastRunAt.IsZero() {
		res["last_run_at"] = lastRunAt
	}
	return res
}

func FormatGame(game *Game) Any {
	res := Any{
		"id":             game.ID,
		"user_id":        game.UserID,
		"title":          game.Title,
		"min_price":      game.MinPrice,
		"type":           game.Type,
		"classification": game.Classification,
	}
	return res
}

func FormatUpload(upload *Upload) Any {
	res := Any{
		"id":       upload.ID,
		"game_id":  upload.GameID,
		"type":     upload.Type,
		"storage":  upload.Storage,
		"size":     upload.Size,
		"filename": upload.Filename,
		"url":      upload.URL,
	}
	platforms := Any{}
	if upload.PlatformLinux {
		platforms["linux"] = "all"
	}
	if upload.PlatformWindows {
		platforms["windows"] = "all"
	}
	if upload.PlatformMac {
		platforms["osx"] = "all"
	}
	res["platforms"] = platforms

	build := upload.Store.FindBuild(upload.Head)
	if build != nil {
		res["build"] = FormatBuild(build)
		res["channel_name"] = upload.ChannelName
	}

	return res
}

func FormatUploads(uploads []*Upload) []Any {
	var res []Any
	for _, u := range uploads {
		res = append(res, FormatUpload(u))
	}
	return res
}

func FormatBuild(build *Build) Any {
	res := Any{
		"id":              build.ID,
		"parent_build_id": build.ParentBuildID,
		"upload_id":       build.UploadID,
		"version":         build.Version,
	}
	return res
}

func FormatBuilds(builds []*Build) []Any {
	var res []Any
	for _, b := range builds {
		res = append(res, FormatBuild(b))
	}
	return res
}

func FormatBuildFile(bf *BuildFile) Any {
	res := Any{
		"size":     bf.Size,
		"type":     bf.Type,
		"sub_type": bf.SubType,
	}
	return res
}
