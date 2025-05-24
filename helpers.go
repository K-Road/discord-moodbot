package main

import (
	"unicode"
	"unicode/utf8"
)

var emojiRanges = []*unicode.RangeTable{
	{
		R32: []unicode.Range32{
			{Lo: 0x1F600, Hi: 0x1F64F, Stride: 1}, // Emoticons
			{Lo: 0x1F300, Hi: 0x1F5FF, Stride: 1}, // Misc Symbols and Pictographs
			{Lo: 0x1F680, Hi: 0x1F6FF, Stride: 1}, // Transport & Map
			{Lo: 0x1F900, Hi: 0x1F9FF, Stride: 1}, // Supplemental Symbols and Pictographs
			{Lo: 0x1FA70, Hi: 0x1FAFF, Stride: 1}, // Symbols & Pictographs Extended-A
		},
		LatinOffset: 0,
	},
}

func isProbabyEmoji(s string) bool {
	if s == "" {
		return false
	}

	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError || size != len(s) {
		return false
	}

	//check rune in common unicode range for emojis
	for _, table := range emojiRanges {
		if unicode.Is(table, r) {
			return true
		}
	}
	return false
}
