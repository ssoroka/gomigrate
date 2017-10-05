package migrator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestCamelCase(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
	}{
		{"test", "Test"},
		{"test_case", "TestCase"},
		{"a_really_long_test_case", "AReallyLongTestCase"},
		{"a_tesT_casE_with_MixEd_caps", "ATestCaseWithMixedCaps"},
		{"_leading_underscore", "LeadingUnderscore"},
		{"_-_%$@leading_garbage", "LeadingGarbage"},
		{"trailing_underscore_", "TrailingUnderscore"},
		{"unexpected(characters", "UnexpectedCharacters"},
		{"duplicated_-__characters", "DuplicatedCharacters"},
		{"some numbers 43", "SomeNumbers43"},
		{"99 red balloons", "99RedBalloons"},
	}

	for _, test := range tests {
		result := camelCase(test.Input)
		if result != test.Expected {
			t.Errorf("Expected camelCase(%s) to be %s, but it was %s", test.Input, test.Expected, result)
		}
	}
}

func TestASTWalk(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "../default_templates/new_migrator_install.tmpl", nil, parser.ParseComments)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	v := visitor{}
	ast.Walk(&v, f)

	if !v.RunPos.IsValid() {
		t.Error("expected the run-position to be valid (non-zero), but it was", v.RunPos)
	}
}
