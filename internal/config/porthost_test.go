package config

import "testing"

func TestDefaultPublicPortHost(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"code-os.example.com":  "port{port}.example.com",
		"code-os.mlvcdn.com":   "port{port}.mlvcdn.com",
		"dash.foo.co.uk":       "port{port}.foo.co.uk",
		"example.com":          "port{port}.example.com",
		"CODE-OS.Example.COM.": "port{port}.example.com",
		"localhost":            "",
		"":                     "",
	}
	for input, want := range cases {
		if got := DefaultPublicPortHost(input); got != want {
			t.Fatalf("DefaultPublicPortHost(%q) = %q, want %q", input, got, want)
		}
	}
}
