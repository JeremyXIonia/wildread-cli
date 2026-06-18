package parser

import "testing"

func TestFormatFromExt(t *testing.T) {
	cases := map[string]string{
		"a.epub":     "epub",
		"a.txt":      "txt",
		"a.md":       "md",
		"a.markdown": "md",
		"a.zip":      "",
		"a":          "",
	}
	for in, want := range cases {
		if got := FormatFromExt(in); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}
