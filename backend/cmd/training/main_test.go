package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetName_returnsCharacterByIndex(t *testing.T) {
	chars := []Character{
		{FirstName: "Alice", LastName: "Smith"},
		{FirstName: "Bob", LastName: "Jones"},
	}

	assert.Equal(t, chars[0], getName(chars, 1))
	assert.Equal(t, chars[1], getName(chars, 2))
}

func TestGetName_rollsOverWhenIndexExceedsLength(t *testing.T) {
	chars := []Character{
		{FirstName: "Alice", LastName: "Smith"},
		{FirstName: "Bob", LastName: "Jones"},
	}

	assert.Equal(t, chars[0], getName(chars, 3))
	assert.Equal(t, chars[1], getName(chars, 4))
}

func TestGetName_withTrainerCharacterList(t *testing.T) {
	assert.Equal(t, characters[0], getName(characters, 1))
	assert.Equal(t, characters[len(characters)-1], getName(characters, len(characters)))
	assert.Equal(t, characters[0], getName(characters, len(characters)+1))
}

func TestProjectAvatarURL(t *testing.T) {
	url := projectAvatarURL("My Project")
	assert.Contains(t, url, "https://api.dicebear.com/9.x/shapes/svg")
	assert.Contains(t, url, "seed=My Project")
	assert.Contains(t, url, "dbeafe")
}

func TestGroupAvatarURL(t *testing.T) {
	url := groupAvatarURL("My Group")
	assert.Contains(t, url, "https://api.dicebear.com/9.x/shapes/svg")
	assert.Contains(t, url, "seed=My Group")
	assert.Contains(t, url, "fef3c7")
}

func TestGuru00Constant(t *testing.T) {
	assert.Equal(t, "King", guru00.FirstName)
	assert.Equal(t, "Arthur", guru00.LastName)
}

func TestCharactersList(t *testing.T) {
	assert.GreaterOrEqual(t, len(characters), 30)
	assert.Equal(t, "Brother", characters[0].FirstName)
	assert.Equal(t, "Zoot", characters[len(characters)-1].FirstName)
}
