package command

import (
	"bufio"
	"fmt"
	"strconv"

	"github.com/mmnalaka/medis/internal/resp"
)

func ReadCommand(reader *bufio.Reader) (resp.RESPData, error) {
	// Skip any leading whitespace or newlines
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}

		// If not whitespace, unread and break
		if b != '\r' && b != '\n' && b != ' ' {
			if err := reader.UnreadByte(); err != nil {
				return nil, err
			}
			break
		}
	}

	// Read the first byte to determine the type
	firstByte, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	// Put back the byte for full reading
	if err := reader.UnreadByte(); err != nil {
		return nil, err
	}

	// Read and store the data
	var data []byte
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, err
		}
		data = append(data, b)
		if isCompleteMessage(data) {
			break
		}
	}

	// Based on the first byte, create the appropriate response object
	var result resp.RESPData
	switch firstByte {
	case resp.SimpleStringPrefix:
		result = &resp.SimpleString{}
	case resp.IntegerPrefix:
		result = &resp.Integer{}
	case resp.ErrorPrefix:
		result = &resp.Error{}
	case resp.BulkStringPrefix:
		result = &resp.BulkString{}
	case resp.ArrayPrefix:
		result = &resp.Array{}
	default:
		return nil, fmt.Errorf("unknown command type: %c (%d)", firstByte, firstByte)
	}

	// Decode the response
	if err := result.Decode(data); err != nil {
		return nil, err
	}

	return result, nil
}

func isCompleteMessage(data []byte) bool {
	if len(data) < 2 {
		return false
	}

	switch data[0] {
	case resp.SimpleStringPrefix, resp.ErrorPrefix, resp.IntegerPrefix:
		return data[len(data)-2] == '\r' && data[len(data)-1] == '\n'

	case resp.BulkStringPrefix:
		return isCompleteBulkString(data)

	case resp.ArrayPrefix:
		return isCompleteArray(data)

	default:
		return false
	}
}

func isCompleteBulkString(data []byte) bool {
	// Find the first CRLF
	firstCRLF := findCRLF(data[1:])
	if firstCRLF == -1 {
		return false
	}

	// Parse length
	lengthStr := string(data[1 : 1+firstCRLF])
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return false
	}

	// Null bulk string
	if length == -1 {
		return len(data) >= firstCRLF+3 && // +3 for \r\n after length
			data[len(data)-2] == '\r' && data[len(data)-1] == '\n'
	}

	// Check if we have the complete string
	expectedLen := 1 + firstCRLF + 2 + length + 2 // $length\r\ndata\r\n
	return len(data) >= expectedLen
}

func isCompleteArray(data []byte) bool {
	// Find the first CRLF
	firstCRLF := findCRLF(data[1:])
	if firstCRLF == -1 {
		return false
	}

	// Parse array length
	lengthStr := string(data[1 : 1+firstCRLF])
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return false
	}

	// Null array
	if length == -1 {
		return len(data) >= firstCRLF+3 && // +3 for \r\n after length
			data[len(data)-2] == '\r' && data[len(data)-1] == '\n'
	}

	// Parse each element
	pos := 1 + firstCRLF + 2 // Skip *length\r\n
	elementsFound := 0

	for pos < len(data) {
		if elementsFound == length {
			return true
		}

		if pos >= len(data) {
			return false
		}

		switch data[pos] {
		case resp.BulkStringPrefix:
			end := findBulkStringEnd(data[pos:])
			if end == -1 {
				return false
			}
			pos += end
			elementsFound++
		default:
			return false // Only bulk strings expected in redis-cli commands
		}
	}

	return elementsFound == length
}

func findCRLF(data []byte) int {
	for i := 0; i < len(data)-1; i++ {
		if data[i] == '\r' && data[i+1] == '\n' {
			return i
		}
	}
	return -1
}

func findBulkStringEnd(data []byte) int {
	firstCRLF := findCRLF(data[1:])
	if firstCRLF == -1 {
		return -1
	}

	lengthStr := string(data[1 : 1+firstCRLF])
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return -1
	}

	if length == -1 {
		return firstCRLF + 3 // $-1\r\n
	}

	totalLen := 1 + firstCRLF + 2 + length + 2 // $length\r\ndata\r\n
	if totalLen > len(data) {
		return -1
	}

	return totalLen
}
