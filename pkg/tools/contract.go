package tools

// Request is the standard tool request schema.
type Request struct {
	ToolCallID string                 `json:"tool_call_id"`
	TaskID     string                 `json:"task_id"`
	AgentID    string                 `json:"agent_id"`
	Name       string                 `json:"name"`
	Args       map[string]interface{} `json:"args"`
}

// Response is the standard tool response schema.
type Response struct {
	ToolCallID string `json:"tool_call_id"`
	ForLLM     string `json:"for_llm"`
	ForUser    string `json:"for_user"`
	IsError    bool   `json:"is_error"`
	AuditRef   string `json:"audit_ref,omitempty"`
	Code       string `json:"code,omitempty"`
	Retriable  bool   `json:"retriable,omitempty"`
}

// ErrorCodes for tool responses.
const (
	CodeSafetyBlocked  = "SAFETY_BLOCKED"
	CodeTimeout        = "TIMEOUT"
	CodeRateLimited    = "RATE_LIMITED"
	CodePermissionDenied = "PERMISSION_DENIED"
)

// ErrorResponse creates an error response.
func ErrorResponse(toolCallID, forLLM, forUser, code string, retriable bool) Response {
	return Response{
		ToolCallID: toolCallID,
		ForLLM:     forLLM,
		ForUser:    forUser,
		IsError:    true,
		Code:       code,
		Retriable:  retriable,
	}
}

// SuccessResponse creates a success response.
func SuccessResponse(toolCallID, forLLM, forUser, auditRef string) Response {
	return Response{
		ToolCallID: toolCallID,
		ForLLM:     forLLM,
		ForUser:    forUser,
		IsError:    false,
		AuditRef:   auditRef,
	}
}
