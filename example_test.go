package feel_test

import (
	"fmt"

	"github.com/pblumer/feel"
	"github.com/pblumer/feel/value"
)

// Compile a FEEL expression once, then evaluate it against different inputs.
func Example() {
	// Declare the variables the expression may reference.
	env := feel.NewEnv("Season", "Guest Count")

	// Parse + type-check + compile into a reusable Go closure.
	expr, err := feel.CompileString(
		`if Season = "Winter" and Guest Count > 8 then "Spareribs" else "Salad"`,
		env,
	)
	if err != nil {
		panic(err)
	}

	// Evaluate: bind values by name and run the closure.
	out, err := expr(env.NewScope(map[string]value.Value{
		"Season":      value.Str("Winter"),
		"Guest Count": value.NumberFromInt64(10),
	}))
	if err != nil {
		panic(err)
	}
	fmt.Println(out)

	// Output:
	// Spareribs
}
