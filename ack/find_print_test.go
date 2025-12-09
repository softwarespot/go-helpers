package ack

import (
	"strings"
	"testing"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

type matchedTest struct {
	lines []string
}

func newMatchesOutputTest() *matchedTest {
	return &matchedTest{
		lines: nil,
	}
}

func (mt *matchedTest) Write(p []byte) (n int, err error) {
	mt.lines = append(mt.lines, string(p))
	return len(p), nil
}

func Test_FindPrint(t *testing.T) {
	singleLine := "test test"
	multiLines := `test
test
example example
test
example example
example example
test
test
test
test test
test
test
test
test
example example
example example
example example
`

	tests := []struct {
		name string
		text string
		term string
		opts FindOptions
		want []string
	}{
		{
			name: "multiple matcheses no context",
			text: multiLines,
			term: "test",
			opts: FindOptions{
				UseCaseSensitive: false,
				UseRegExp:        false,
				MaxCount:         0,
				BeforeContext:    0,
				AfterContext:     0,
				BufferSize:       0,
			},
			want: []string{
				"test.txt\n",
				"1:test\n",
				"2:test\n",
				"4:test\n",
				"7:test\n",
				"8:test\n",
				"9:test\n",
				"10:test test\n",
				"11:test\n",
				"12:test\n",
				"13:test\n",
				"14:test\n",
			},
		},
		{
			name: "multiple matches with after context",
			text: multiLines,
			term: "test",
			opts: FindOptions{
				UseCaseSensitive: false,
				UseRegExp:        false,
				MaxCount:         0,
				BeforeContext:    0,
				AfterContext:     1,
				BufferSize:       0,
			},
			want: []string{
				"test.txt\n",
				"1:test\n",
				"2:test\n",
				"3-example example\n",
				"4:test\n",
				"5-example example\n",
				"--\n",
				"7:test\n",
				"8:test\n",
				"9:test\n",
				"10:test test\n",
				"11:test\n",
				"12:test\n",
				"13:test\n",
				"14:test\n",
				"15-example example\n",
			},
		},
		{
			name: "multiple matches with before context and after context",
			text: multiLines,
			term: "test",
			opts: FindOptions{
				UseCaseSensitive: false,
				UseRegExp:        false,
				MaxCount:         0,
				BeforeContext:    2,
				AfterContext:     2,
				BufferSize:       0,
			},
			want: []string{
				"test.txt\n",
				"1:test\n",
				"2:test\n",
				"3-example example\n",
				"4:test\n",
				"5-example example\n",
				"6-example example\n",
				"7:test\n",
				"8:test\n",
				"9:test\n",
				"10:test test\n",
				"11:test\n",
				"12:test\n",
				"13:test\n",
				"14:test\n",
				"15-example example\n",
				"16-example example\n",
			},
		},
		{
			name: "multiple matches with max count and after context",
			text: multiLines,
			term: "test",
			opts: FindOptions{
				UseCaseSensitive: false,
				UseRegExp:        false,
				MaxCount:         2,
				BeforeContext:    0,
				AfterContext:     10,
				BufferSize:       0,
			},
			want: []string{
				"test.txt\n",
				"1:test\n",
				"2:test\n",
				"3-example example\n",
			},
		},
		{
			name: "single match with max count and after context",
			text: singleLine,
			term: "test",
			opts: FindOptions{
				UseCaseSensitive: false,
				UseRegExp:        false,
				MaxCount:         2,
				BeforeContext:    0,
				AfterContext:     10,
				BufferSize:       0,
			},
			want: []string{
				"test.txt\n",
				"1:test test\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.text)
			w := newMatchesOutputTest()
			err := FindPrint(r, w, "test.txt", tt.term, tt.opts, PrintOptions{
				LocationsWithMatches:    false,
				LocationsWithoutMatches: false,
				CountsOnly:              false,
				IsPiped:                 false,
				NoColor:                 true,
			})
			testhelpers.AssertNoError(t, err)
			testhelpers.AssertEqual(t, w.lines, tt.want)
		})
	}
}
