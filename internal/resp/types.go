package resp

import (
	"fmt"
	"strconv"
)

// RESP (Redis Serialization Protocol) type prefixes
const (
	SimpleStringPrefix = '+'
	ErrorPrefix        = '-'
	IntegerPrefix      = ':'
	BulkStringPrefix   = '$'
	ArrayPrefix        = '*'
)

// Validate the data before decoding
func validateData(data []byte, prefix byte, dataType string) error {
	// Check for prefix
	if len(data) == 0 || data[0] != prefix {
		return fmt.Errorf("invalid %s format", dataType)
	}
	// Check for CRLF ending
	if data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
		return fmt.Errorf("invalid data format: missing CRLF")
	}
	return nil
}

type RESPData interface {
	Encode() []byte
	Decode([]byte) error
}

// SimpleString is a RESP type
type SimpleString struct {
	Data string
}

// Encode the SimpleString to RESP format
// example: OK => +OK\r\n
func (s *SimpleString) Encode() []byte {
	return []byte(fmt.Sprintf("+%s\r\n", s.Data))
}

// Decode the SimpleString from RESP format
// example: +OK\r\n => OK
func (s *SimpleString) Decode(data []byte) error {
	if err := validateData(data, SimpleStringPrefix, "simple string"); err != nil {
		return err
	}
	// Remove prefix and CRLF
	s.Data = string(data[1 : len(data)-2])
	return nil
}

// Error is a RESP type
type Error struct {
	Data string
}

// Encode the Error to RESP format
// example: Error => -Error\r\n
func (e *Error) Encode() []byte {
	return []byte(fmt.Sprintf("-%s\r\n", e.Data))
}

// Decode the Error from RESP format
// example: -Error\r\n => Error
func (e *Error) Decode(data []byte) error {
	if err := validateData(data, ErrorPrefix, "error"); err != nil {
		return err
	}
	// Remove prefix and CRLF
	e.Data = string(data[1 : len(data)-2])
	return nil
}

// Integer is a RESP type
type Integer struct {
	Data int64
}

// Encode the Integer to RESP format
// example: 1 => :1\r\n
func (i *Integer) Encode() []byte {
	return []byte(fmt.Sprintf(":%d\r\n", i.Data))
}

// Decode the Integer from RESP format
// example: :1\r\n => 1
func (i *Integer) Decode(data []byte) error {
	if err := validateData(data, IntegerPrefix, "integer"); err != nil {
		return err
	}

	// Get the number part without prefix and CRLF
	numStr := string(data[1 : len(data)-2])
	if len(numStr) == 0 {
		return fmt.Errorf("invalid integer: empty number")
	}

	// Handle negative numbers
	start := 0
	sign := int64(1)
	if numStr[0] == '-' {
		sign = -1
		start = 1
		if len(numStr) == 1 {
			return fmt.Errorf("invalid integer: only minus sign")
		}
	}

	// Convert to number
	i.Data = 0
	for _, b := range numStr[start:] {
		// Check if character is a digit
		if b < '0' || b > '9' {
			return fmt.Errorf("invalid integer: non-digit character found")
		}

		// Check for overflow
		if i.Data > (1<<63-1)/10 {
			return fmt.Errorf("integer overflow")
		}

		i.Data = i.Data*10 + int64(b-'0')
	}

	i.Data *= sign
	return nil
}

// BulkString is a RESP type
type BulkString struct {
	Data []byte
}

// Encode the BulkString to RESP format
// example: Hello => $5\r\nHello\r\n (5 is the length of the string)
func (b *BulkString) Encode() []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(b.Data), b.Data))
}

// Decode the BulkString from RESP format
// example: $5\r\nHello\r\n => Hello
// example: Null bulk string: $-1\r\n => Null
func (b *BulkString) Decode(data []byte) error {
	if err := validateData(data, BulkStringPrefix, "bulk string"); err != nil {
		return err
	}

	// Find the first CRLF which appears after the length
	// Cannot just slice the string since the length could be more than 1 digit
	firstCRLFIndex := -1
	for i := 1; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			firstCRLFIndex = i
			break
		}
	}

	// When the the length is not available
	if firstCRLFIndex == -1 {
		return fmt.Errorf("invalid bulk string: missing CRLF")
	}

	// Parse length
	lengthStr := string(data[1:firstCRLFIndex])
	// Null buld string
	if lengthStr == "-1" {
		b.Data = nil
		return nil
	}
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return fmt.Errorf("invalid bulk string length: %v", err)
	}

	// Validate string length
	expectedLen := firstCRLFIndex + 2 + length + 2 // First CRLF + length + string + final CRLF
	if len(data) != expectedLen {
		return fmt.Errorf("invalid bulk string: length mismatch")
	}

	// Extract the string data
	b.Data = make([]byte, length)
	copy(b.Data, data[firstCRLFIndex+2:len(data)-2])

	return nil
}

// Array is a RESP type that can contain any other RESP type
type Array struct {
	Data []RESPData
}

// Encode the Array to RESP format
// example: [1, 2] => *2\r\n:1\r\n:2\r\n
func (a *Array) Encode() []byte {
	if a.Data == nil {
		return []byte("*-1\r\n")
	}

	result := []byte(fmt.Sprintf("*%d\r\n", len(a.Data)))
	for _, item := range a.Data {
		result = append(result, item.Encode()...)
	}

	return result
}

// Decode the Array from RESP format
// example: *2\r\n:1\r\n:2\r\n => [1, 2]
func (a *Array) Decode(data []byte) error {
	a.Data = nil // Initialize to nil before any validation

	if err := validateData(data, ArrayPrefix, "array"); err != nil {
		return err
	}

	// Find the first CRLF which appears after the length
	firstCRLFIndex := -1
	for i := 1; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			firstCRLFIndex = i
			break
		}
	}

	if firstCRLFIndex == -1 {
		return fmt.Errorf("invalid array format: missing CRLF")
	}

	// Parse the length
	lengthStr := string(data[1:firstCRLFIndex])
	if lengthStr == "-1" {
		a.Data = nil
		return nil
	}
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return fmt.Errorf("invalid array length: %v", err)
	}

	// Parse elements
	a.Data = make([]RESPData, length)
	currentIndex := firstCRLFIndex + 2 // Skip the first CRLF

	for i := 0; i < length; i++ {
		// Validate the array length
		if currentIndex >= len(data) {
			a.Data = nil
			return fmt.Errorf("invalid array: length mismatch, unexpected end of data")
		}

		// Determine the RESP data type
		var element RESPData

		switch data[currentIndex] {
		case SimpleStringPrefix:
			element = &SimpleString{}
		case ErrorPrefix:
			element = &Error{}
		case IntegerPrefix:
			element = &Integer{}
		case BulkStringPrefix:
			element = &BulkString{}
		case ArrayPrefix:
			element = &Array{}
		default:
			a.Data = nil
			return fmt.Errorf("invalid array element RESP data type: %c", data[currentIndex])
		}

		// Find end of the current element
		endIndex := findElementEnd(data[currentIndex:])
		if endIndex == -1 {
			a.Data = nil
			return fmt.Errorf("invalid array element: missing end")
		}

		// Decode element
		if err := element.Decode(data[currentIndex : currentIndex+endIndex]); err != nil {
			a.Data = nil
			return err
		}

		a.Data[i] = element

		currentIndex += endIndex // Move to the next element
	}

	return nil
}

func findElementEnd(data []byte) int {
	switch data[0] {
	case BulkStringPrefix:
		// For bulk strings, we need to find the first CRLF to get the length
		firstCRLFIndex := -1
		for i := 1; i < len(data)-1; i++ {
			if data[i] == '\r' && data[i+1] == '\n' {
				firstCRLFIndex = i
				break
			}
		}
		if firstCRLFIndex == -1 {
			return -1
		}

		// Parse length
		lengthStr := string(data[1:firstCRLFIndex])
		// Null bulk string
		if lengthStr == "-1" {
			return firstCRLFIndex + 2
		}
		// Non-null bulk string
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return -1
		}
		return firstCRLFIndex + 2 + length + 2 // first CRLF + length + second CRLF
	default:
		// For simple strings, errors, and integers
		for i := 1; i < len(data)-1; i++ {
			if data[i] == '\r' && data[i+1] == '\n' {
				return i + 2
			}
		}
	}

	return -1
}
