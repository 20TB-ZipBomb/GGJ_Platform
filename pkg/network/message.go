package network

import (
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/logger"
	"github.com/20TB-ZipBomb/GGJ_Platform/internal/utils/json"
)

type MessageType string

const (
	ConnectionRejected MessageType = "connection_rejected"
	CreateLobby                    = "create_lobby"
)

// Generic communication message containing a message type
type Message struct {
	Type MessageType `json:"message_type"`
}

// Creates a generic communication message containing a message type
func CreateMessageJSON(msgType MessageType) string {
	msg := &Message{
		Type: msgType,
	}

	json, err := json.MarshalJSON[Message](msg)
	if err != nil {
		logger.Errorf("Failed to marshal JSON: %s\nerr: %v", msg, err)
		return ""
	}

	return string(json)
}
