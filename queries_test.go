package mitch

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Select(t *testing.T) {
	assert := assert.New(t)

	s := newStore()
	var user *User
	for i := 0; i < 4; i++ {
		u := s.MakeUser(fmt.Sprintf("user %d", i))
		if user == nil {
			user = u
		}
		for j := 0; j < 4; j++ {
			u.MakeGame(fmt.Sprintf("game %d-%d", i, j))
		}
	}

	var games []*Game
	s.Select(&games, SortBy("ID", "asc").ForMap(s.Games), Eq{"UserID": user.ID})
	assert.EqualValues(4, len(games))
	for _, g := range games {
		assert.EqualValues(user.ID, g.UserID)
	}

	lastGameID := int64(-1)
	for _, g := range games {
		assert.True(lastGameID < g.ID)
		lastGameID = g.ID
	}

	largestID := int64(0)
	for _, g := range games {
		if g.ID > largestID {
			largestID = g.ID
		}
	}

	var game Game
	success := s.SelectOne(&game, SortBy("ID", "desc").ForMap(s.Games), Eq{"UserID": user.ID})
	assert.True(success)
	assert.EqualValues(game.ID, largestID)

	success = s.SelectOne(&game, NoSort().ForMap(s.Games), Eq{"UserID": -1})
	assert.False(success)
}
