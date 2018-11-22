package mitch

func (g *Game) CanBeViewedBy(user *User) bool {
	if g.CanBeEditedBy(user) {
		return true
	}
	if g.Published {
		return true
	}
	return false
}

func (g *Game) CanBeEditedBy(user *User) bool {
	// TODO: game admins
	if g.UserID == user.ID {
		return true
	}
	return false
}

func (u *Upload) CanBeDownloadedBy(user *User) bool {
	// TODO: download keys, min prices, etc.
	g := u.Store.FindGame(u.GameID)
	return g.CanBeViewedBy(user)
}
