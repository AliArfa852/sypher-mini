package tools

import (
	"context"
	"sync"

	"github.com/sypherexx/sypher-mini/pkg/bus"
)

// ReplyTarget holds channel and chat ID for replies.
type ReplyTarget struct {
	Channel string
	ChatID  string
}

// MessageTool sends a message to a channel (relay to outbound bus).
type MessageTool struct {
	msgBus       *bus.MessageBus
	safeMode     bool
	taskTargets  map[string]ReplyTarget
	targetsMu    sync.RWMutex
}

// NewMessageTool creates a message tool.
func NewMessageTool(msgBus *bus.MessageBus, safeMode bool) *MessageTool {
	return &MessageTool{
		msgBus:      msgBus,
		safeMode:    safeMode,
		taskTargets: make(map[string]ReplyTarget),
	}
}

// SetReplyTarget sets the reply target for a task (called by agent loop).
func (t *MessageTool) SetReplyTarget(taskID string, channel, chatID string) {
	t.targetsMu.Lock()
	defer t.targetsMu.Unlock()
	t.taskTargets[taskID] = ReplyTarget{Channel: channel, ChatID: chatID}
}

// GetReplyTarget returns the channel and chat ID for a task, or empty strings.
func (t *MessageTool) GetReplyTarget(taskID string) (channel, chatID string) {
	t.targetsMu.RLock()
	defer t.targetsMu.RUnlock()
	if target, ok := t.taskTargets[taskID]; ok {
		return target.Channel, target.ChatID
	}
	return "", ""
}

// ClearReplyTarget removes the reply target when task completes.
func (t *MessageTool) ClearReplyTarget(taskID string) {
	t.targetsMu.Lock()
	defer t.targetsMu.Unlock()
	delete(t.taskTargets, taskID)
}

// Execute publishes an outbound message.
func (t *MessageTool) Execute(ctx context.Context, req Request) Response {
	if t.safeMode {
		return ErrorResponse(req.ToolCallID,
			"message disabled in safe mode",
			"Message sending is disabled in safe mode.",
			CodePermissionDenied, false)
	}

	content, _ := req.Args["content"].(string)
	if content == "" {
		return ErrorResponse(req.ToolCallID,
			"Missing 'content' argument",
			"Content is required.",
			CodePermissionDenied, false)
	}

	channel := "cli"
	chatID := "default"
	t.targetsMu.RLock()
	if target, ok := t.taskTargets[req.TaskID]; ok {
		channel = target.Channel
		chatID = target.ChatID
	}
	t.targetsMu.RUnlock()

	t.msgBus.PublishOutbound(bus.OutboundMessage{
		Channel: channel,
		ChatID:  chatID,
		Content: content,
	})

	return SuccessResponse(req.ToolCallID,
		"Message sent.",
		"Message sent.",
		"")
}
