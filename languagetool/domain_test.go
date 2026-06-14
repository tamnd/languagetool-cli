package languagetool

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "languagetool" {
		t.Errorf("scheme = %q, want languagetool", info.Scheme)
	}
	found := false
	for _, a := range info.Aliases {
		if a == "lt" {
			found = true
		}
	}
	if !found {
		t.Errorf("aliases = %v, want to contain lt", info.Aliases)
	}
	if info.Identity.Binary != "languagetool" {
		t.Errorf("binary = %q, want languagetool", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	d := Domain{}

	typ, id, err := d.Classify("en-US")
	if err != nil || typ != "lang" || id != "en-US" {
		t.Errorf("Classify(lang code) = %q/%q/%v, want lang/en-US/nil", typ, id, err)
	}

	typ, id, err = d.Classify("fr")
	if err != nil || typ != "lang" || id != "fr" {
		t.Errorf("Classify(short lang) = %q/%q/%v, want lang/fr/nil", typ, id, err)
	}

	typ, id, err = d.Classify("This are wrong")
	if err != nil || typ != "text" {
		t.Errorf("Classify(text) = %q/%q/%v, want text/.../nil", typ, id, err)
	}

	_, _, err = d.Classify("")
	if err == nil {
		t.Error("Classify('') = nil error, want error")
	}
}

func TestLocate(t *testing.T) {
	d := Domain{}

	url, err := d.Locate("text", "hello world")
	if err != nil || url != "https://languagetool.org/" {
		t.Errorf("Locate(text) = %q/%v, want https://languagetool.org/", url, err)
	}

	url, err = d.Locate("lang", "en-US")
	if err != nil || url != "https://languagetool.org/" {
		t.Errorf("Locate(lang) = %q/%v, want https://languagetool.org/", url, err)
	}

	_, err = d.Locate("unknown", "foo")
	if err == nil {
		t.Error("Locate(unknown) = nil error, want error")
	}
}

func TestDomainRegistered(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := h.Domain("languagetool"); !ok {
		t.Fatal("languagetool domain not registered")
	}
	if _, ok := h.Domain("lt"); !ok {
		t.Fatal("lt alias not registered")
	}
}

func TestIsLangCode(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"en", true},
		{"en-US", true},
		{"fr", true},
		{"de", true},
		{"zh-CN", true},
		{"", false},
		{"This are wrong", false},
		{"hello world", false},
		{"x", false},            // too short
		{"toolong-here", false}, // too long overall
	}
	for _, tc := range cases {
		got := isLangCode(tc.input)
		if got != tc.want {
			t.Errorf("isLangCode(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}
