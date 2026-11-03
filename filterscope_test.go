package feel

import (
	"reflect"
	"testing"

	"github.com/pblumer/feel/value"
)

// A filter over context elements reads a variable bound outside it when the element
// has no key of that name, and the element's key when it has (DMN 10.3.2.5: the
// element's entries are the innermost scope). Before, a declared outer variable
// shadowed the element's key, and one not declared was null.
func TestFilterReadsOuterVariablesBehindTheElementsKeys(t *testing.T) {
	vars := map[string]value.Value{
		"prefix": value.Str("tm-"),
		"a":      value.Str("outer"),
		"limit":  value.NumberFromInt64(2),
	}
	cases := map[string]string{
		// an outer variable that is no key of the elements
		`[{n: "tm-a"}, {n: "x"}][not(starts with(n, prefix))]`: "[{n: x}]",
		`[1, 2, 3][item > limit]`:                              "[3]",
		// the element's key shadows the outer variable of the same name
		`[{a: 1}, {a: 5}][a > 2]`: "[{a: 5}]",
		// an element without the key reads the outer variable
		`[{a: 1}, {b: 2}][a = "outer"]`: "[{b: 2}]",
		// a binding made inside the filter is inner to the element
		`[{a: 1, xs: [7]}][some a in xs satisfies a = 7]`: "[{a: 1, xs: [7]}]",
	}
	for src, want := range cases {
		if got := evalStr(t, src, vars); got.String() != want {
			t.Errorf("%q = %s, want %s", src, got, want)
		}
	}
}

// CompileStringOpen names what an expression reads only inside a filter, so a
// caller that learns its variables from "unknown variable" errors can bind them.
func TestCompileStringOpenNamesWhatOnlyAFilterReads(t *testing.T) {
	src := `count(gruppen[not(matches(mailNickname, "^" + app.prefix + "tm-"))]) + count(xs)`
	_, refs, open, err := CompileStringOpen(src, NewEnv("gruppen", "xs"))
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"gruppen", "xs"}; !reflect.DeepEqual(refs, want) {
		t.Errorf("refs = %v, want %v", refs, want)
	}
	if want := []string{"app", "mailNickname"}; !reflect.DeepEqual(open, want) {
		t.Errorf("open = %v, want %v", open, want)
	}

	// Declared, the open names are refs, and the element's key still wins.
	env := NewEnv("gruppen", "xs", "app", "mailNickname")
	ce, refs, open, err := CompileStringOpen(src, env)
	if err != nil {
		t.Fatal(err)
	}
	if len(open) != 0 || len(refs) != 4 {
		t.Fatalf("declared: refs %v open %v", refs, open)
	}
	g := func(nick string) value.Value {
		c := value.NewContext()
		c.Put("mailNickname", value.Str(nick))
		return c
	}
	app := value.NewContext()
	app.Put("prefix", value.Str("p-"))
	v, err := ce(env.NewScope(map[string]value.Value{
		"gruppen": value.NewList(g("p-tm-a"), g("p-raum"), g("x")),
		"xs":      value.NewList(),
		"app":     app,
		// an instance variable of the element key's name does not shadow the key
		"mailNickname": value.Str("p-tm-shadow"),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if v.String() != "2" {
		t.Errorf("deviations = %s, want 2", v)
	}
}
