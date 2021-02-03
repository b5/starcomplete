package starcomplete

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.starlark.net/starlark"
)

type TestCase struct {
	program string
	pos     Position
	expect  []Completion
}

func TestGoodCompletions(t *testing.T) {
	filename := ""
	predeclared := starlark.StringDict{}
	modules := []ModuleInfo{
		{Name: "net/http", Documentation: "make network requests", DefaultImportSymbol: "http"},
	}
	cases := []TestCase{
		{program: `load("","")`, pos: NewPosition(1, 5), expect: []Completion{
			modules[0].Completion(NewPosition(1, 5)),
		}},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("GoodCompletions_%d", i), func(t *testing.T) {
			got, err := Completions(filename, c.program, c.pos, predeclared, modules)
			if err != nil {
				t.Error(err)
			}

			if diff := cmp.Diff(c.expect, got); diff != "" {
				t.Errorf("result mistmatch (-want +got):\n%s", diff)
			}

		})
	}
}
