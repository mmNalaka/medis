package resp

import "fmt"

// RESP type prefixes
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
	// Remove prefix and CRLF
	i.Data = 0
	for _, b := range data[1 : len(data)-2] {
		i.Data = i.Data*10 + int64(b-'0')
	}
	return nil
}
