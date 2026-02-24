package mcp

import "encoding/json"

// JSON-RPC 2.0 基础结构

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// 标准 JSON-RPC 错误码
const (
	ErrParseError     = -32700
	ErrInvalidRequest = -32600
	ErrMethodNotFound = -32601
	ErrInvalidParams  = -32602
	ErrInternal       = -32603
	ErrUnauthorized   = -32001
	ErrPermission     = -32002
	ErrNotFound       = -32003
)

func ErrorResp(id interface{}, code int, msg string) Response {
	return Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg}}
}

func OKResp(id interface{}, result interface{}) Response {
	return Response{JSONRPC: "2.0", ID: id, Result: result}
}

// MCP 协议结构（2024-11-05）

type InitializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	ClientInfo      ClientInfo     `json:"clientInfo"`
	Capabilities    map[string]any `json:"capabilities"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	ServerInfo      ServerInfo     `json:"serverInfo"`
	Capabilities    Capabilities   `json:"capabilities"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Capabilities struct {
	Tools map[string]any `json:"tools,omitempty"`
}

// Tool 定义（序列化给客户端）
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// ToolDef 内部工具定义（带权限标记）
// Perm 为空表示所有 Agent 可用，非空则需要 agent.HasPermission(Perm)
type ToolDef struct {
	Tool
	Perm        string `json:"-"`
	InitOnly    bool   `json:"-"` // 仅未初始化的 Agent 可见
}

type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]PropSchema  `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

type PropSchema struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

func TextResult(text string) ToolCallResult {
	return ToolCallResult{Content: []ContentBlock{{Type: "text", Text: text}}}
}

func ErrorResult(msg string) ToolCallResult {
	return ToolCallResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: msg}}}
}
