package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyPrefixBase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single word", "WarmDesk", "WAR"},
		{"two words", "My Project", "MPY"},
		{"three words", "Very Big Project", "VBP"},
		{"four words, only first 3 used", "A B C D", "ABC"},
		{"short name pads with X", "AB", "ABX"},
		{"single char pads with XX", "A", "AXX"},
		{"empty string pads with XXX", "", "XXX"},
		{"lowercase input uppercased", "hello world", "HWE"},
		{"symbols split words", "hello-world foo", "HWF"},
		{"digits are included", "123 project", "1P2"},
		{"multiple symbols ignored", "a   b", "ABX"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, keyPrefixBase(tt.input))
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "Hello World", "hello-world"},
		{"multiple spaces", "Hello   World", "hello-world"},
		{"leading/trailing dashes", "  --Hello--  ", "hello"},
		{"special chars", "Hello! World?", "hello-world"},
		{"mixed case", "UPPERCASE lower", "uppercase-lower"},
		{"digits", "Version 2", "version-2"},
		{"all special", "@#$%", "project"},
		{"empty string", "", "project"},
		{"unicode treated as dash", "café", "caf"},
		{"single word", "Project", "project"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, slugify(tt.input))
		})
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "10"},
		{42, "42"},
		{100, "100"},
		{999, "999"},
		{1000, "1000"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, itoa(tt.input))
		})
	}
}
