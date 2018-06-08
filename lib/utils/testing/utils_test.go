package testing

import (
	"testing"

	"github.com/jtopjian/yak/lib/utils"

	"github.com/stretchr/testify/assert"
)

func TestUtils_parseSimplified(t *testing.T) {
	testCases := []struct {
		testCase string
		expected map[string]interface{}
	}{
		{
			`foo=bar bar=baz`,
			map[string]interface{}{
				"foo": "bar",
				"bar": "baz",
			},
		},
		{
			`foo="hello world" bar=baz`,
			map[string]interface{}{
				"foo": "hello world",
				"bar": "baz",
			},
		},
	}

	for _, i := range testCases {
		params, err := utils.ParseSimplified(i.testCase)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, i.expected, params)
	}
}
