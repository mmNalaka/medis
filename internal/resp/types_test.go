package resp

import (
	"reflect"
	"testing"
)

func TestSimpleString_Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic OK string",
			input:    "OK",
			expected: "+OK\r\n",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "+\r\n",
		},
		{
			name:     "string with spaces",
			input:    "Hello World",
			expected: "+Hello World\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SimpleString{Data: tt.input}
			result := string(s.Encode())
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSimpleString_Decode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "valid simple string",
			input:       "+OK\r\n",
			expected:    "OK",
			shouldError: false,
		},
		{
			name:        "missing prefix",
			input:       "OK\r\n",
			shouldError: true,
		},
		{
			name:        "missing CRLF",
			input:       "+OK",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SimpleString{}
			err := s.Decode([]byte(tt.input))

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldError && s.Data != tt.expected {
				t.Errorf("got %q, want %q", s.Data, tt.expected)
			}
		})
	}
}

func TestError_Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic error",
			input:    "Error message",
			expected: "-Error message\r\n",
		},
		{
			name:     "empty error",
			input:    "",
			expected: "-\r\n",
		},
		{
			name:     "error with special chars",
			input:    "Error: Key not found!",
			expected: "-Error: Key not found!\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Error{Data: tt.input}
			result := string(e.Encode())
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestError_Decode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "valid error",
			input:       "-Error message\r\n",
			expected:    "Error message",
			shouldError: false,
		},
		{
			name:        "missing prefix",
			input:       "Error message\r\n",
			shouldError: true,
		},
		{
			name:        "missing CRLF",
			input:       "-Error message",
			shouldError: true,
		},
		{
			name:        "empty error",
			input:       "-\r\n",
			expected:    "",
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Error{}
			err := e.Decode([]byte(tt.input))

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldError && e.Data != tt.expected {
				t.Errorf("got %q, want %q", e.Data, tt.expected)
			}
		})
	}
}

func TestInter_Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "basic integer",
			input:    12345,
			expected: ":12345\r\n",
		},
		{
			name:     "zero integer",
			input:    0,
			expected: ":0\r\n",
		},
		{
			name:     "negative integer",
			input:    -12345,
			expected: ":-12345\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Integer{Data: tt.input}
			result := string(i.Encode())
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestInteger_Decode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int64
		shouldError bool
	}{
		{
			name:        "valid integer",
			input:       ":12345\r\n",
			expected:    12345,
			shouldError: false,
		},
		{
			name:        "missing prefix",
			input:       "12345\r\n",
			shouldError: true,
		},
		{
			name:        "missing CRLF",
			input:       ":12345",
			shouldError: true,
		},
		{
			name:        "invalid integer",
			input:       ":abc\r\n",
			shouldError: true,
		},
		{
			name:        "invalid integer",
			input:       ":\r\n",
			shouldError: true,
		},
		{
			name:     "negative integer",
			input:    ":-12345\r\n",
			expected: -12345,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Integer{}
			err := i.Decode([]byte(tt.input))
			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldError && i.Data != tt.expected {
				t.Errorf("got %d, want %d", i.Data, tt.expected)
			}
		})
	}
}

func TestBulkStringEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic string",
			input:    "Hello",
			expected: "$5\r\nHello\r\n",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "$0\r\n\r\n",
		},
		{
			name:     "string with spaces",
			input:    "Hello World",
			expected: "$11\r\nHello World\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BulkString{Data: []byte(tt.input)}
			result := string(b.Encode())
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBulkStringDecode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "valid string",
			input:       "$5\r\nHello\r\n",
			expected:    "Hello",
			shouldError: false,
		},
		{
			name:        "missing prefix",
			input:       "Hello\r\n",
			shouldError: true,
		},
		{
			name:        "missing CRLF",
			input:       "$5\r\nHello",
			shouldError: true,
		},
		{
			name:        "invalid string",
			input:       "$5Hello\r\n",
			shouldError: true,
		},
		{
			name:        "null string",
			input:       "$-1\r\n",
			expected:    "",
			shouldError: false,
		},
		{
			name:        "long string",
			input:       "$1000\r\n" + string(make([]byte, 1000)) + "\r\n",
			expected:    string(make([]byte, 1000)),
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BulkString{}
			err := b.Decode([]byte(tt.input))
			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldError && string(b.Data) != tt.expected {
				t.Errorf("got %q, want %q", b.Data, tt.expected)
			}
		})
	}
}

func TestArrayEncode(t *testing.T) {
	tests := []struct {
		name     string
		input    *Array
		expected []byte
	}{
		{
			name:     "empty array",
			input:    &Array{Data: []RESPData{}},
			expected: []byte("*0\r\n"),
		},
		{
			name: "array with integers",
			input: &Array{Data: []RESPData{
				&Integer{Data: 1},
				&Integer{Data: 2},
			}},
			expected: []byte("*2\r\n:1\r\n:2\r\n"),
		},
		{
			name:     "nil array",
			input:    &Array{Data: nil},
			expected: []byte("*-1\r\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.input.Encode()
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("Encode() = %v, expected %v", actual, tt.expected)
			}
		})
	}
}

func TestArrayDecode(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expected    *Array
		shouldError bool
	}{
		{
			name:        "empty array",
			input:       []byte("*0\r\n"),
			expected:    &Array{Data: []RESPData{}},
			shouldError: false,
		},
		{
			name:  "array with integers",
			input: []byte("*2\r\n:1\r\n:2\r\n"),
			expected: &Array{Data: []RESPData{
				&Integer{Data: 1},
				&Integer{Data: 2},
			}},
			shouldError: false,
		},
		{
			name:        "nil array",
			input:       []byte("*-1\r\n"),
			expected:    &Array{Data: nil},
			shouldError: false,
		},
		{
			name:        "invalid array length",
			input:       []byte("*a\r\n"),
			expected:    &Array{Data: nil},
			shouldError: true,
		},
		{
			name:        "missing CRLF",
			input:       []byte("*2\r"),
			expected:    &Array{Data: nil},
			shouldError: true,
		},
		{
			name:  "array with bulk string",
			input: []byte("*1\r\n$5\r\nhello\r\n"),
			expected: &Array{Data: []RESPData{
				&BulkString{Data: []byte("hello")},
			}},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := &Array{}
			err := actual.Decode(tt.input)

			if tt.shouldError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.shouldError && !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("got %v, want %v", actual, tt.expected)
			}
		})
	}
}
