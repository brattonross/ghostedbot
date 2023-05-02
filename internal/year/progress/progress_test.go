package progress_test

import (
	"testing"

	"github.com/brattonross/ghostedbot/internal/year/progress"
)

func TestPercentage(t *testing.T) {
	tt := []struct {
		name     string
		input    int64
		expected float64
	}{
		{
			name:     "zero percent",
			input:    1672531200,
			expected: 0,
		},
		{
			name:     "one hundred percent",
			input:    1704023999,
			expected: 100,
		},
		{
			name:     "fifty percent",
			input:    1688277599,
			expected: 50,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := progress.Percentage(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func TestToBar(t *testing.T) {
	tt := []struct {
		name     string
		input    float64
		expected string
	}{
		{
			name:     "zero percent",
			input:    0,
			expected: "░░░░░░░░░░",
		},
		{
			name:     "one hundred percent",
			input:    100,
			expected: "██████████",
		},
		{
			name:     "fifty percent",
			input:    50,
			expected: "█████░░░░░",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			actual := progress.ToBar(tc.input)
			if actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
