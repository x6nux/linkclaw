package mcp

import (
	"context"
	"encoding/json"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/llm"
	"github.com/linkclaw/backend/internal/repository"
	"github.com/linkclaw/backend/internal/service"
)

// Handler 路由 JSON-RPC 方法到具体处理器
type Handler struct {
	agentSvc     *service.AgentService
	taskSvc      *service.TaskService
	messageSvc   *service.MessageService
	knowledgeSvc *service.KnowledgeService
	memorySvc    *service.MemoryService
	companyRepo  repository.CompanyRepo
	deploySvc    *service.DeploymentService
	llmRepo      *llm.Repository
	promptSvc    *service.PromptService
	obsSvc       *service.ObservabilityService
	obsRepo      repository.ObservabilityRepo
	orgSvc       *service.OrganizationService
	contextSvc   *service.ContextService
}

func NewHandler(
	agentSvc *service.AgentService,
	taskSvc *service.TaskService,
	messageSvc *service.MessageService,
	knowledgeSvc *service.KnowledgeService,
	memorySvc *service.MemoryService,
	companyRepo repository.CompanyRepo,
	deploySvc *service.DeploymentService,
	llmRepo *llm.Repository,
	promptSvc *service.PromptService,
	obsSvc *service.ObservabilityService,
	obsRepo repository.ObservabilityRepo,
	orgSvc *service.OrganizationService,
	contextSvc *service.ContextService,
) *Handler {
	return &Handler{
		agentSvc:     agentSvc,
		taskSvc:      taskSvc,
		messageSvc:   messageSvc,
		knowledgeSvc: knowledgeSvc,
		memorySvc:    memorySvc,
		companyRepo:  companyRepo,
		deploySvc:    deploySvc,
		llmRepo:      llmRepo,
		promptSvc:    promptSvc,
		obsSvc:       obsSvc,
		obsRepo:      obsRepo,
		orgSvc:       orgSvc,
		contextSvc:   contextSvc,
	}
}

func (h *Handler) Handle(ctx context.Context, sess *Session, req Request) Response {
	switch req.Method {
	case "initialize":
		return h.handleInitialize(ctx, sess, req)
	case "tools/list":
		return h.handleToolsList(sess, req)
	case "tools/call":
		return h.handleToolsCall(ctx, sess, req)
	case "ping":
		return OKResp(req.ID, map[string]string{"status": "pong"})
	default:
		return ErrorResp(req.ID, ErrMethodNotFound, "method not found: "+req.Method)
	}
}

func (h *Handler) handleInitialize(ctx context.Context, sess *Session, req Request) Response {
	var params InitializeParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return ErrorResp(req.ID, ErrInvalidParams, "invalid params")
	}
	sess.ProtocolVersion = params.ProtocolVersion
	sess.ClientInfo = params.ClientInfo
	sess.Initialized = true

	return OKResp(req.ID, InitializeResult{
		ProtocolVersion: protocolVersion,
		ServerInfo:      ServerInfo{Name: "LinkClaw", Version: "0.1.0"},
		Capabilities:    Capabilities{Tools: map[string]any{"listChanged": false}},
	})
}

func (h *Handler) handleToolsList(sess *Session, req Request) Response {
	tools := ToolsForAgent(sess.Agent)
	return OKResp(req.ID, ToolsListResult{Tools: tools})
}

func (h *Handler) handleToolsCall(ctx context.Context, sess *Session, req Request) Response {
	if !sess.Initialized {
		return ErrorResp(req.ID, ErrInvalidRequest, "session not initialized")
	}
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return ErrorResp(req.ID, ErrInvalidParams, "invalid params")
	}

	// 权限检查
	if !HasToolPermission(sess.Agent, params.Name) {
		return ErrorResp(req.ID, ErrPermission, "权限不足：你没有权限使用工具 "+params.Name)
	}

	result := h.dispatchTool(ctx, sess, params.Name, params.Arguments)
	return OKResp(req.ID, result)
}

func (h *Handler) dispatchTool(ctx context.Context, sess *Session, name string, args json.RawMessage) ToolCallResult {
	switch name {
	// 基础
	case "get_employee_handbook":
		return h.toolGetIdentity(ctx, sess, args)
	case "get_company_profile":
		return h.toolGetCompanyInfo(ctx, sess, args)
	case "punch_clock":
		return h.toolPing(ctx, sess, args)
	case "fill_onboarding_info":
		return h.toolSetMyName(ctx, sess, args)
	case "update_work_status":
		return h.toolUpdateStatus(ctx, sess, args)
	case "report_for_duty":
		return h.toolMarkInitialized(ctx, sess, args)
	// 消息
	case "send_message":
		return h.toolSendMessage(ctx, sess, args)
	case "get_messages":
		return h.toolGetMessages(ctx, sess, args)
	case "list_channels":
		return h.toolListChannels(ctx, sess, args)
	case "mark_messages_read":
		return h.toolMarkMessagesRead(ctx, sess, args)
	// 任务
	case "list_tasks":
		return h.toolListTasks(ctx, sess, args)
	case "get_task":
		return h.toolGetTask(ctx, sess, args)
	case "get_task_detail":
		return h.toolGetTaskDetail(ctx, sess, args)
	case "add_task_comment":
		return h.toolAddTaskComment(ctx, sess, args)
	case "add_task_dependency":
		return h.toolAddTaskDependency(ctx, sess, args)
	case "watch_task":
		return h.toolWatchTask(ctx, sess, args)
	case "accept_task":
		return h.toolAcceptTask(ctx, sess, args)
	case "submit_task_result":
		return h.toolSubmitTaskResult(ctx, sess, args)
	case "fail_task":
		return h.toolFailTask(ctx, sess, args)
	case "create_task":
		return h.toolCreateTask(ctx, sess, args)
	case "create_subtask":
		return h.toolCreateSubtask(ctx, sess, args)
	case "update_persona":
		return h.toolUpdatePersona(ctx, sess, args)
	// 记忆
	case "remember":
		return h.toolRemember(ctx, sess, args)
	case "recall":
		return h.toolRecall(ctx, sess, args)
	case "forget":
		return h.toolForget(ctx, sess, args)
	case "list_memories":
		return h.toolListMemories(ctx, sess, args)
	// 知识库
	case "search_knowledge":
		return h.toolSearchKnowledge(ctx, sess, args)
	case "get_document":
		return h.toolGetDocument(ctx, sess, args)
	case "write_document":
		return h.toolWriteDocument(ctx, sess, args)
	// 管理
	case "list_openings":
		return h.toolListPositions(ctx, sess, args)
	case "hire":
		return h.toolCreateAgent(ctx, sess, args)
	case "fire":
		return h.toolDeleteAgent(ctx, sess, args)
	case "list_models":
		return h.toolListModels(ctx, sess, args)
	// 部署
	case "ssh_exec":
		return h.toolSSHExec(ctx, sess, args)
	case "ssh_upload":
		return h.toolSSHUpload(ctx, sess, args)
	case "docker_run":
		return h.toolDockerRun(ctx, sess, args)
	case "docker_exec_cmd":
		return h.toolDockerExecCmd(ctx, sess, args)
	case "docker_cp":
		return h.toolDockerCp(ctx, sess, args)
	case "docker_rm":
		return h.toolDockerRm(ctx, sess, args)
	case "get_plugin_info":
		return h.toolGetPluginInfo(ctx, sess, args)
	// 可观测性
	case "get_my_trace_history":
		return h.toolGetMyTraceHistory(ctx, sess, args)
	case "get_cost_status":
		return h.toolGetCostStatus(ctx, sess, args)
	case "list_observability_alerts":
		return h.toolListObsAlerts(ctx, sess, args)
	case "replay_trace":
		return h.toolReplayTrace(ctx, sess, args)
	// 组织架构
	case "get_org_chart":
		return h.toolGetOrgChart(ctx, sess, args)
	case "list_my_approvals":
		return h.toolListMyApprovals(ctx, sess, args)
	case "submit_approval":
		return h.toolSubmitApproval(ctx, sess, args)
	case "view_department":
		return h.toolViewDepartment(ctx, sess, args)
	// Context search
	case "search_context":
		return h.toolSearchContext(ctx, sess, args)
	// Agent tools (B4)
	case "agent_grep":
		return h.toolAgentGrep(ctx, sess, args)
	case "agent_read_chunk":
		return h.toolAgentReadChunk(ctx, sess, args)
	case "agent_list_symbols":
		return h.toolAgentListSymbols(ctx, sess, args)
	default:
		return ErrorResult("未知工具：" + name)
	}
}

// B4 Agent Tools

func (h *Handler) toolAgentGrep(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var params struct {
		Pattern     string   `json:"pattern"`
		FilePattern string   `json:"file_pattern,omitempty"`
		IgnoreCase  bool     `json:"ignore_case,omitempty"`
		MaxResults  int      `json:"max_results,omitempty"`
		DirectoryIDs []string `json:"directory_ids,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return ErrorResult("invalid parameters")
	}

	if params.Pattern == "" {
		return ErrorResult("pattern is required")
	}

	// 获取目录
	var dirs []*domain.ContextDirectory
	if len(params.DirectoryIDs) > 0 {
		for _, id := range params.DirectoryIDs {
			d, err := h.contextSvc.GetDirectoryByID(ctx, id)
			if err != nil || d == nil || !d.IsActive {
				continue
			}
			dirs = append(dirs, d)
		}
	} else {
		allDirs, err := h.contextSvc.ListDirectories(ctx, sess.Agent.CompanyID)
		if err != nil {
			return ErrorResult("failed to list directories")
		}
		for _, d := range allDirs {
			if d.IsActive {
				dirs = append(dirs, d)
			}
		}
	}

	if len(dirs) == 0 {
		return ErrorResult("no active directories to search")
	}

	// 使用 ContextSearchAgent 执行 grep
	agent := service.NewContextSearchAgent(h.contextSvc.GetLLMClient(), h.contextSvc.GetRepo())
	result := agent.ExecuteGrep(service.AgentToolCall{
		ID:        "grep-call",
		Name:      "grep",
		Arguments: args,
	}, dirs, 1024*1024)

	if result.IsError {
		return ErrorResult(result.Content)
	}

	return TextResult(result.Content)
}

func (h *Handler) toolAgentReadChunk(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var params struct {
		Path         string   `json:"path"`
		Offset       int      `json:"offset,omitempty"`
		Limit        int      `json:"limit,omitempty"`
		DirectoryIDs []string `json:"directory_ids,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return ErrorResult("invalid parameters")
	}

	if params.Path == "" {
		return ErrorResult("path is required")
	}

	// 获取目录
	var dirs []*domain.ContextDirectory
	if len(params.DirectoryIDs) > 0 {
		for _, id := range params.DirectoryIDs {
			d, err := h.contextSvc.GetDirectoryByID(ctx, id)
			if err != nil || d == nil || !d.IsActive {
				continue
			}
			dirs = append(dirs, d)
		}
	} else {
		allDirs, err := h.contextSvc.ListDirectories(ctx, sess.Agent.CompanyID)
		if err != nil {
			return ErrorResult("failed to list directories")
		}
		for _, d := range allDirs {
			if d.IsActive {
				dirs = append(dirs, d)
			}
		}
	}

	if len(dirs) == 0 {
		return ErrorResult("no active directories to search")
	}

	// 使用 ContextSearchAgent 执行 read_chunk
	agent := service.NewContextSearchAgent(h.contextSvc.GetLLMClient(), h.contextSvc.GetRepo())
	result := agent.ExecuteReadChunk(service.AgentToolCall{
		ID:        "read-chunk-call",
		Name:      "read_chunk",
		Arguments: args,
	}, dirs, 1024*1024)

	if result.IsError {
		return ErrorResult(result.Content)
	}

	return TextResult(result.Content)
}

func (h *Handler) toolAgentListSymbols(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var params struct {
		Path         string   `json:"path"`
		SymbolType   string   `json:"symbol_type,omitempty"`
		DirectoryIDs []string `json:"directory_ids,omitempty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return ErrorResult("invalid parameters")
	}

	if params.Path == "" {
		return ErrorResult("path is required")
	}

	// 获取目录
	var dirs []*domain.ContextDirectory
	if len(params.DirectoryIDs) > 0 {
		for _, id := range params.DirectoryIDs {
			d, err := h.contextSvc.GetDirectoryByID(ctx, id)
			if err != nil || d == nil || !d.IsActive {
				continue
			}
			dirs = append(dirs, d)
		}
	} else {
		allDirs, err := h.contextSvc.ListDirectories(ctx, sess.Agent.CompanyID)
		if err != nil {
			return ErrorResult("failed to list directories")
		}
		for _, d := range allDirs {
			if d.IsActive {
				dirs = append(dirs, d)
			}
		}
	}

	if len(dirs) == 0 {
		return ErrorResult("no active directories to search")
	}

	// 使用 ContextSearchAgent 执行 list_symbols
	agent := service.NewContextSearchAgent(h.contextSvc.GetLLMClient(), h.contextSvc.GetRepo())
	result := agent.ExecuteListSymbols(service.AgentToolCall{
		ID:        "list-symbols-call",
		Name:      "list_symbols",
		Arguments: args,
	}, dirs, 1024*1024)

	if result.IsError {
		return ErrorResult(result.Content)
	}

	return TextResult(result.Content)
}
