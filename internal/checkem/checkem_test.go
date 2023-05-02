package checkem_test

import (
	"testing"

	"github.com/brattonross/ghostedbot/internal/checkem"
)

func TestCheckem(t *testing.T) {
	dt := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no repeating digits",
			input:    "1234567890",
			expected: "1234567890",
		},
		{
			name:     "2 repeating digits",
			input:    "1234567899",
			expected: "1234567899 - <:EZ:1103063620209885214> <a:Clap:1103063782760124540> gratz on the dubs",
		},
		{
			name:     "3 repeating digits",
			input:    "123456777",
			expected: "123456777 - <:EZ:1103063620209885214> <a:Clap:1103063782760124540> gratz on the trips",
		},
		{
			name:     "12 repeating digits",
			input:    "12333333333333",
			expected: "12333333333333 - <:Paggi:1103063622474792980> <a:Clap:1103063782760124540> you got more than 10 repeating digits?!",
		},
	}

	for _, tc := range dt {
		t.Run(tc.name, func(t *testing.T) {
			actual := checkem.Checkem(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}
