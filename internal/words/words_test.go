package words_test

import (
	"testing"

	"github.com/brattonross/ghostedbot/internal/words"
)

func TestLeftPad(t *testing.T) {
	tt := []struct {
		name   string
		str    string
		length int
		char   string
		want   string
	}{
		{
			name:   "no char",
			str:    "test",
			length: 10,
			char:   "",
			want:   "      test",
		},
		{
			name:   "same length",
			str:    "test",
			length: 4,
			char:   "",
			want:   "test",
		},
		{
			name:   "custom char",
			str:    "test",
			length: 10,
			char:   "x",
			want:   "xxxxxxtest",
		},
		{
			name:   "length less than char",
			str:    "t",
			length: 2,
			char:   "xxxx",
			want:   "xt",
		},
		{
			name:   "long char",
			str:    "t",
			length: 7,
			char:   "ccccc",
			want:   "cccccct",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := words.LeftPad(tc.str, tc.length, tc.char)
			if got != tc.want {
				t.Errorf("got %s; want %s", got, tc.want)
			}
		})
	}
}
