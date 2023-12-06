package luna

import "testing"

func TestValidLuna(t *testing.T) {
	tests := []struct {
		name     string
		number   string
		expected bool
	}{
		{
			name:     "valid number",
			number:   "4532015112830366",
			expected: true,
		},
		{
			name:     "invalid number",
			number:   "1234567890123456",
			expected: false,
		},
		{
			name:     "invalid with non-digit",
			number:   "45320151128303a6",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Valid(tt.number)
			if result != tt.expected {
				t.Errorf("Valid() for number %v, expected %v, got %v", tt.number, tt.expected, result)
			}
		})
	}
}
