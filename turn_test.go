package main

import (
	"testing"
)

var (
	// expectstr = "Съешь ещё этих мягких булок, да выпей чаю."
	// teststr   = "C]tim to` 'nb[ vzurb[ ,ekjr? lf dsgtq xf./"
	expected = "абвгдеёжзийклмнопрстуфхцшщьыъэюя.АБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦШЩЬЫЪЭЮЯ,"
	teststr  = "f,dult`;pbqrkvyjghcnea[wioms]'.z/F<DULT~:PBQRKVYJGHCNEA{WIOMS}\">Z?"
)

func TestConverting(t *testing.T) {
	out := Turn(teststr)
	if out != expected {
		t.Errorf("Test failed!\nExpect string: %q\ngot:           %q",
			expected, out)
	}
}
