package command

import (
	"fmt"
	"sync"

	"github.com/mmnalaka/medis/internal/resp"
)

type Handler struct {
	store map[string][]byte
	mu    sync.Mutex
}

func NewHandler() *Handler {
	return &Handler{
		store: make(map[string][]byte),
	}
}

// Handle commands
func (h *Handler) Handle(cmd *Command) resp.RESPData {
	switch cmd.Name {
	case "PING":
		return h.handlePing()
	case "SET":
		return h.handleSet(cmd)
	case "GET":
		return h.handleGet(cmd)
	default:
		return h.handleUnknown(cmd)
	}
}

// Handler for PING command
func (h *Handler) handlePing() resp.RESPData {
	return &resp.SimpleString{Data: "PONG"}
}

// Handler for unknown commands
func (h *Handler) handleUnknown(cmd *Command) resp.RESPData {
	return &resp.Error{Data: fmt.Sprintf("ERR unknown command %s", cmd.Name)}
}

// Handler for SET command
func (h *Handler) handleSet(cmd *Command) resp.RESPData {
	if len(cmd.Args) != 2 {
		return &resp.Error{Data: "ERR wrong number of arguments for SET"}
	}

	h.mu.Lock()
	h.store[string(cmd.Args[0])] = cmd.Args[1]
	defer h.mu.Unlock()
	return &resp.SimpleString{Data: "OK"}
}

// Handler for GET command
func (h *Handler) handleGet(cmd *Command) resp.RESPData {
	if len(cmd.Args) != 1 {
		return &resp.Error{Data: "ERR wrong number of arguments for GET"}
	}

	h.mu.Lock()
	value, exists := h.store[string(cmd.Args[0])]
	h.mu.Unlock()

	if !exists {
		// Null if the key does not exist
		return &resp.BulkString{Data: nil}
	}

	return &resp.BulkString{Data: value}
}
