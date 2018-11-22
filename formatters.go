package mitch

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

func FormatGame(game *Game) Any {
	res := Any{
		"id":        game.ID,
		"user_id":   game.UserID,
		"title":     game.Title,
		"min_price": game.MinPrice,
		"type":      game.Type,
	}
	return res
}

func FormatUpload(upload *Upload) Any {
	res := Any{
		"id":      upload.ID,
		"game_id": upload.GameID,
		"type":    upload.Type,
		"storage": upload.Storage,
		"size":    upload.Size,
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
