package edict

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestDetailString(t *testing.T) {
	// I don't really care to test every combination, so I chose one
	// arbitrarily to at least make sure String() works.
	if VsC.String() != "vs-c" {
		t.Error("Something is wrong with the part of speech map: VsC != vs-c")
	}
}

func TestDetailFor(t *testing.T) {
	for id, str := range DetailString {
		if DetailFor[str] != id {
			t.Errorf("incorrect detail mapping\n   got: %s\n  want:%s", DetailFor[str], id)
		}
	}
}

func TestParseIdentifier(t *testing.T) {
	testData := []struct {
		input      string
		class      identifierClass
		identifier string
	}{
		{"42", none, ""},
		{"See foo", xref, "foo"},
		{"See あ・い", xref, "あ・い"},
		{"n", detail, "n"},
		{"esp. ", text, "esp. "},
	}

	for _, test := range testData {
		class, identifier := parseIdentifier(test.input)

		if class != test.class {
			fmt.Errorf("class returned by parseIdentifier:\n   got: %s\n  want: %s", class, test.class)
		}

		if identifier != test.identifier {
			fmt.Errorf("identifier returned by parseIdentifier:\n   got: %s\n  want: %s", identifier, test.identifier)
		}
	}
}

func TestParseGloss(t *testing.T) {
	testData := []struct {
		input   string
		def     string
		details []Detail
		xrefs   []string
	}{
		{
			input:   "(n) foo",
			def:     "foo",
			details: []Detail{N},
			xrefs:   nil,
		},
		{
			input:   "(n,adj-no) foo",
			def:     "foo",
			details: []Detail{N, AdjNo},
			xrefs:   nil,
		},
		{
			input:   "(See foobar) foo",
			def:     "foo",
			details: nil,
			xrefs:   []string{"foobar"},
		},
		{
			input:   "(n) (See foobar) foo",
			def:     "foo",
			details: []Detail{N},
			xrefs:   []string{"foobar"},
		},
		{
			input:   "foo",
			def:     "foo",
			details: nil,
			xrefs:   nil,
		},
		{
			input:   "(1) (abbr) (uK) (See foobar) foo",
			def:     "foo",
			details: []Detail{Abbr, UK},
			xrefs:   []string{"foobar"},
		},
	}

	for _, test := range testData {
		def, details, xrefs, err := parseGloss(test.input)
		if err != nil {
			t.Errorf("Error parsing '%s': %s", test.input, err)
			continue
		}

		if def != test.def {
			t.Errorf("Parsing %s: %s != %s", test.input, def, test.def)
		}

		if !reflect.DeepEqual(details, test.details) {
			t.Errorf("Parsing %s: details: %v != %v", test.input, details, test.details)
		}

		if !reflect.DeepEqual(xrefs, test.xrefs) {
			t.Errorf("Parsing %s: xrefs: %v != %v", test.input, details, test.details)
		}
	}
}

func TestParseKey(t *testing.T) {
	testData := []struct {
		input  string
		kanji  []string
		kana   []string
		errors bool
	}{
		{
			input:  "A;B;C [x;y;z]",
			kanji:  []string{"A", "B", "C"},
			kana:   []string{"x", "y", "z"},
			errors: false,
		},
		{
			input:  "A [x]",
			kanji:  []string{"A"},
			kana:   []string{"x"},
			errors: false,
		},
		{
			input:  "A",
			kanji:  []string{"A"},
			kana:   []string{},
			errors: false,
		},
		{
			input:  "A;B",
			kanji:  []string{"A", "B"},
			kana:   []string{},
			errors: false,
		},
		{
			input:  "A;B  [C;D]",
			kanji:  []string{"A", "B"},
			kana:   []string{"C", "D"},
			errors: false,
		},
		{
			input:  "A;B [C",
			kanji:  []string{"A", "B"},
			kana:   []string{},
			errors: true,
		},
	}

	for _, test := range testData {
		kanji, kana, err := parseKey(test.input)

		if err != nil && !test.errors {
			t.Errorf("%s: unexpected error: %s", test.input, err)
			continue
		} else if err == nil && test.errors {
			t.Errorf("%s: got success but expected error", test.input)
		}

		if !reflect.DeepEqual(kanji, test.kanji) {
			t.Errorf("%s: bad kanji:\n  got %v\n  want %v", test.input, kanji, test.kanji)
		}
		if !reflect.DeepEqual(kana, test.kana) {
			t.Errorf("%s: bad kana:\n  got %v\n  want %v", test.input, kana, test.kana)
		}
	}
}

func TestFixKey(t *testing.T) {
	testData := []struct {
		in  string
		out string
	}{
		{"foo(bar) (baz) (quux)", "foo"},
		{"foo(bar)", "foo"},
		{"foo", "foo"},
	}

	for _, test := range testData {
		got := fixKey(test.in)

		if got != test.out {
			t.Errorf("fixing key %s:\n  got %s\n want: %s\n", test.in, got, test.out)
		}
	}

}

func TestParseLine(t *testing.T) {
	testData := []struct {
		input  string
		expect Entry
	}{
		{
			input: "刖 [げつ] /(n) (arch) (obsc) (See 剕) cutting off the leg at the knee (form of punishment in ancient China)/EntL2542160/",
			expect: Entry{
				Kanji:       []string{"刖"},
				Kana:        []string{"げつ"},
				Information: []Detail{N, Arch, Obsc},
				Gloss: []Gloss{{
					"cutting off the leg at the knee (form of punishment in ancient China)",
					[]Detail{},
					[]string{"剕"}},
				},
				Sequence:           "EntL2542160",
				RecordingAvailable: false,
			},
		},
		{
			input: "ジョン;Jon [じょん] /(n) (1) (abbr) (uK) (See jrockway) my name/(2) (uk) apparently a common name for dogs/EntL0000000/",
			expect: Entry{
				Kanji:       []string{"ジョン", "Jon"},
				Kana:        []string{"じょん"},
				Information: []Detail{N},
				Gloss: []Gloss{
					{"my name", []Detail{Abbr, UK}, []string{"jrockway"}},
					{"apparently a common name for dogs", []Detail{Uk}, nil},
				},
				Sequence:           "EntL0000000",
				RecordingAvailable: false,
			},
		},
	}

	for line, test := range testData {
		got, err := parseLine(test.input)
		if err != nil {
			t.Errorf("parse error %s \non %s (line %d)", err, test.input, line)
			continue
		}

		if !reflect.DeepEqual(got, test.expect) {
			t.Errorf("unexpected entry\n   got: %v\n  want: %v", got, test.expect)
			continue
		}
	}
}

func TestParse(t *testing.T) {
	input := []string{ // These are the first few entries from edict2.
		"刖 [げつ] /(n) (arch) (obsc) (See 剕) cutting off the leg at the knee (form of punishment in ancient China)/EntL2542160/",
		"剕 [あしきり] /(n) (arch) (See 五刑) cutting off the leg at the knee (form of punishment in ancient China)/EntL2542150/",
		"劓 [はなきり] /(n) (arch) (See 五刑) cutting off the nose (form of punishment in ancient China)/EntL2542140/",
		"匜;半挿 [はそう;はぞう] /(n) (1) (esp. ) wide-mouthed ceramic vessel having a small hole in its spherical base (into which bamboo was probably inserted to pour liquids)/(2) (See 半挿・はんぞう・1) teapot-like object made typically of lacquerware and used to pour hot and cold liquids/EntL2791750/",
		"咖哩(ateji) [カレー(P);カリー] /(n) (1) (uk) curry/(2) (abbr) (uk) (See カレーライス) rice and curry/(P)/EntL1039140X/",
		"嗉嚢;そ嚢 [そのう] /(n) bird's crop/bird's craw/EntL2542030/",
		"嘈囃;そう囃 [そうざつ] /(n,vs) (obsc) (嘈囃 is sometimes read むねやけ) (See 胸焼け) heartburn/sour stomach/EntL2542040/",
	}

	reader := strings.NewReader(strings.Join(input, "\n"))
	got, err := Parse(reader)

	if err != nil {
		t.Fatal(err)
	}

	if len(got) != len(input) {
		t.Errorf("unexpected output size %d: expected %d", len(got), len(input))
	}
}

func BenchmarkEdictParse(b *testing.B) {
	fh, err := os.Open("edict2")
	if err != nil {
		b.Fatal(err)
	}

	entries, err := Parse(fh)
	fmt.Printf("entries: %d\n", len(entries))

	if err != nil {
		b.Fatal(err)
	}
}
