package utils

import (
	"testing"
)

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello_world", "helloWorld"},
		{"user_name", "userName"},
		{"", ""},
		{"single", "single"},
		{"multiple_under_scores", "multipleUnderScores"},
	}

	for _, test := range tests {
		result := ToCamelCase(test.input)
		if result != test.expected {
			t.Errorf("ToCamelCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"helloWorld", "hello_world"},
		{"userName", "user_name"},
		{"", ""},
		{"single", "single"},
		{"HTTPSConnection", "h_t_t_p_s_connection"},
	}

	for _, test := range tests {
		result := ToSnakeCase(test.input)
		if result != test.expected {
			t.Errorf("ToSnakeCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestToKebabCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"helloWorld", "hello-world"},
		{"userName", "user-name"},
		{"", ""},
		{"single", "single"},
	}

	for _, test := range tests {
		result := ToKebabCase(test.input)
		if result != test.expected {
			t.Errorf("ToKebabCase(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "olleh"},
		{"world", "dlrow"},
		{"", ""},
		{"a", "a"},
		{"12345", "54321"},
	}

	for _, test := range tests {
		result := Reverse(test.input)
		if result != test.expected {
			t.Errorf("Reverse(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"\t\n", true},
		{"hello", false},
		{" hello ", false},
	}

	for _, test := range tests {
		result := IsEmpty(test.input)
		if result != test.expected {
			t.Errorf("IsEmpty(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"hello world", 5, "he..."},
		{"hello", 10, "hello"},
		{"hello world", 8, "hello..."},
		{"", 5, ""},
	}

	for _, test := range tests {
		result := Truncate(test.input, test.length)
		if result != test.expected {
			t.Errorf("Truncate(%q, %d) = %q, expected %q", test.input, test.length, result, test.expected)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		input    string
		substrs  []string
		expected bool
	}{
		{"hello world", []string{"hello"}, true},
		{"hello world", []string{"foo", "world"}, true},
		{"hello world", []string{"foo", "bar"}, false},
		{"", []string{"hello"}, false},
	}

	for _, test := range tests {
		result := Contains(test.input, test.substrs...)
		if result != test.expected {
			t.Errorf("Contains(%q, %v) = %v, expected %v", test.input, test.substrs, result, test.expected)
		}
	}
}

func TestContainsAll(t *testing.T) {
	tests := []struct {
		input    string
		substrs  []string
		expected bool
	}{
		{"hello world", []string{"hello", "world"}, true},
		{"hello world", []string{"hello", "foo"}, false},
		{"hello world", []string{}, true},
		{"", []string{"hello"}, false},
	}

	for _, test := range tests {
		result := ContainsAll(test.input, test.substrs...)
		if result != test.expected {
			t.Errorf("ContainsAll(%q, %v) = %v, expected %v", test.input, test.substrs, result, test.expected)
		}
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		input    string
		start    int
		end      int
		maskChar rune
		expected string
	}{
		{"1234567890", 2, 6, '*', "12****7890"},
		{"hello", 1, 3, 'x', "hxxlo"},
		{"", 0, 5, '*', ""},
		{"test", -1, 2, '*', "test"},
	}

	for _, test := range tests {
		result := Mask(test.input, test.start, test.end, test.maskChar)
		if result != test.expected {
			t.Errorf("Mask(%q, %d, %d, %c) = %q, expected %q",
				test.input, test.start, test.end, test.maskChar, result, test.expected)
		}
	}
}
