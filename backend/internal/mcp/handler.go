package mcp

import (
	"context"
	"encoding/json"

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
	indexingSvc  *service.IndexingService
	companyRepo  repository.CompanyRepo
	deploySvc    *service.DeploymentService
	llmRepo      *llm.Repository
	promptSvc    *service.PromptService
	obsSvc       *service.ObservabilityService
	obsRepo      repository.ObservabilityRepo
	orgSvc       *service.OrganizationService
}

func NewHandler(
	agentSvc *service.AgentService,
	taskSvc *service.TaskService,
	messageSvc *service.MessageService,
	knowledgeSvc *service.KnowledgeService,
	memorySvc *service.MemoryService,
	indexingSvc *service.IndexingService,
	companyRepo repository.CompanyRepo,
	deploySvc *service.DeploymentService,
	llmRepo *llm.Repository,
	promptSvc *service.PromptService,
	obsSvc *service.ObservabilityService,
	obsRepo repository.ObservabilityRepo,
	orgSvc *service.OrganizationService,
) *Handler {
	return &Handler{
		agentSvc:     agentSvc,
		taskSvc:      taskSvc,
		messageSvc:   messageSvc,
		knowledgeSvc: knowledgeSvc,
		memorySvc:    memorySvc,
		indexingSvc:  indexingSvc,
		companyRepo:  companyRepo,
		deploySvc:    deploySvc,
		llmRepo:      llmRepo,
		promptSvc:    promptSvc,
		obsSvc:       obsSvc,
		obsRepo:      obsRepo,
		orgSvc:       orgSvc,
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
	// 上下文索引
	case "search_code":
		return h.toolSearchCode(ctx, sess, args)
	case "index_repository":
		return h.toolIndexRepository(ctx, sess, args)
	case "get_index_status":
		return h.toolGetIndexStatus(ctx, sess, args)
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
	default:
		return ErrorResult("未知工具: " + name)
	}
}
