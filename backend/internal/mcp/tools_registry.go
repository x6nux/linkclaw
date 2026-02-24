package mcp

import "github.com/linkclaw/backend/internal/domain"

// 权限常量（对应 domain.Agent.Permissions 中的值）
const (
	PermHire           = "hire"
	PermOnboard        = "onboard"
	PermKnowledgeWrite = "knowledge:write"
	PermTaskManage     = "task:manage"
	PermTaskCreate     = "task:create"
	PermPersonaWrite   = "persona:write"
)

// allToolDefs 全部工具定义（含权限标记），由 init() 合并 coreToolDefs + deployToolDefs
var allToolDefs []ToolDef

func init() {
	allToolDefs = make([]ToolDef, 0, len(coreToolDefs)+len(deployToolDefs))
	allToolDefs = append(allToolDefs, coreToolDefs...)
	allToolDefs = append(allToolDefs, deployToolDefs...)
}

var coreToolDefs = []ToolDef{
	// ── 基础工具（所有员工） ──────────────────────────────────
	{Tool: Tool{
		Name:        "get_employee_handbook",
		Description: "获取你的员工手册，包含身份信息、公司背景、同事花名册和可用工具说明。建议每次上班（新会话）时先查阅。",
		InputSchema: InputSchema{Type: "object"},
	}},
	{Tool: Tool{
		Name:        "get_company_profile",
		Description: "获取公司简介，包括公司名称、描述和企业文化。",
		InputSchema: InputSchema{Type: "object"},
	}},
	{Tool: Tool{
		Name:        "punch_clock",
		Description: "打卡签到，刷新在岗时间并续期会话。每 30 秒打卡一次以维持在岗状态。",
		InputSchema: InputSchema{Type: "object"},
	}},
	{Tool: Tool{
		Name:        "fill_onboarding_info",
		Description: "填写入职信息——设置你的名字。如果你的名字以「待命名」开头，必须先调用此工具给自己取一个符合你职位和角色的名字。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"name"},
			Properties: map[string]PropSchema{
				"name": {Type: "string", Description: "你给自己取的名字"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "update_work_status",
		Description: "更新你的工作状态（在岗 / 忙碌 / 离岗）。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"status"},
			Properties: map[string]PropSchema{
				"status": {Type: "string", Description: "新状态", Enum: []string{"online", "busy", "offline"}},
			},
		},
	}},

	{InitOnly: true, Tool: Tool{
		Name:        "report_for_duty",
		Description: "标记你已完成到岗报到（取名、上线、打招呼）。入职后只需调用一次，后续重连不会重复报到流程。",
		InputSchema: InputSchema{Type: "object"},
	}},

	// ── 消息工具（所有 Agent） ──────────────────────────────────
	{Tool: Tool{
		Name:        "send_message",
		Description: "向群聊频道或指定 Agent 发送消息。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"content"},
			Properties: map[string]PropSchema{
				"channel":     {Type: "string", Description: "群聊频道名称（如 general、engineering）。与 receiver_id 二选一。"},
				"receiver_id": {Type: "string", Description: "私信目标 Agent 的 ID。与 channel 二选一。"},
				"content":     {Type: "string", Description: "消息内容（支持 Markdown）"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "get_messages",
		Description: "获取频道或私信历史消息，支持 cursor 游标分页。",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropSchema{
				"channel":     {Type: "string", Description: "群聊频道名称"},
				"receiver_id": {Type: "string", Description: "私信对方的 Agent ID"},
				"limit":       {Type: "string", Description: "每页条数（默认 20，最大 50）"},
				"before_id":   {Type: "string", Description: "游标：获取此消息 ID 之前的消息"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "list_channels",
		Description: "列出公司所有群聊频道。",
		InputSchema: InputSchema{Type: "object"},
	}},
	{Tool: Tool{
		Name:        "mark_messages_read",
		Description: "标记消息为已读。收到消息后处理完毕必须调用此工具确认，否则系统会重复推送。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"message_ids"},
			Properties: map[string]PropSchema{
				"message_ids": {Type: "string", Description: "消息 ID 列表，逗号分隔"},
			},
		},
	}},

	// ── 任务工具（所有 Agent 可查看/执行自己的） ─────────────────
	{Tool: Tool{
		Name:        "list_tasks",
		Description: "列出任务列表，支持按范围、状态、优先级过滤。",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropSchema{
				"scope":    {Type: "string", Description: "mine=仅我的任务（默认），all=公司所有任务", Enum: []string{"mine", "all"}},
				"status":   {Type: "string", Description: "过滤状态", Enum: []string{"pending", "assigned", "in_progress", "done", "failed", "cancelled"}},
				"priority": {Type: "string", Description: "过滤优先级", Enum: []string{"low", "medium", "high", "urgent"}},
			},
		},
	}},
	{Tool: Tool{
		Name:        "get_task",
		Description: "获取任务详情，包含子任务列表。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"task_id"},
			Properties: map[string]PropSchema{
				"task_id": {Type: "string", Description: "任务 ID"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "accept_task",
		Description: "接受并开始执行一个分配给你的任务（状态: assigned → in_progress）。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"task_id"},
			Properties: map[string]PropSchema{
				"task_id": {Type: "string", Description: "要接受的任务 ID"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "submit_task_result",
		Description: "提交任务完成结果（状态: in_progress → done）。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"task_id", "result"},
			Properties: map[string]PropSchema{
				"task_id": {Type: "string", Description: "任务 ID"},
				"result":  {Type: "string", Description: "完成结果描述"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "fail_task",
		Description: "标记任务失败（状态: in_progress → failed）。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"task_id", "reason"},
			Properties: map[string]PropSchema{
				"task_id": {Type: "string", Description: "任务 ID"},
				"reason":  {Type: "string", Description: "失败原因"},
			},
		},
	}},

	// ── 知识库工具 ──────────────────────────────────────────────
	{Tool: Tool{
		Name:        "search_knowledge",
		Description: "使用全文搜索在知识库中查找文档。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"query"},
			Properties: map[string]PropSchema{
				"query": {Type: "string", Description: "搜索关键词"},
				"limit": {Type: "string", Description: "返回条数（默认 10）"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "get_document",
		Description: "获取知识库文档完整内容。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"doc_id"},
			Properties: map[string]PropSchema{
				"doc_id": {Type: "string", Description: "文档 ID"},
			},
		},
	}},

	// ── 需要权限的工具 ─────────────────────────────────────────

	{Perm: PermKnowledgeWrite, Tool: Tool{
		Name:        "write_document",
		Description: "创建新文档或更新已有文档（Markdown 格式）。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"title", "content"},
			Properties: map[string]PropSchema{
				"doc_id":  {Type: "string", Description: "文档 ID（传入则更新，不传则创建）"},
				"title":   {Type: "string", Description: "文档标题"},
				"content": {Type: "string", Description: "文档内容（Markdown）"},
				"tags":    {Type: "string", Description: "标签，逗号分隔"},
			},
		},
	}},
	{Perm: PermTaskCreate, Tool: Tool{
		Name:        "create_task",
		Description: "创建任务并分配给指定部门总监或具体人员。总监可创建本部门任务，高管可创建跨部门任务。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"title"},
			Properties: map[string]PropSchema{
				"title":       {Type: "string", Description: "任务标题"},
				"description": {Type: "string", Description: "详细描述"},
				"assignee_id": {Type: "string", Description: "指定负责人 Agent ID（与 department 二选一）"},
				"department":  {Type: "string", Description: "指定部门（自动分配给该部门总监）。可选值：人力资源、产品、工程、商务、市场、财务"},
				"priority":    {Type: "string", Description: "优先级", Enum: []string{"low", "medium", "high", "urgent"}},
			},
		},
	}},
	{Perm: PermTaskManage, Tool: Tool{
		Name:        "create_subtask",
		Description: "创建子任务，并可指定分配给某个 Agent。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"parent_task_id", "title"},
			Properties: map[string]PropSchema{
				"parent_task_id": {Type: "string", Description: "父任务 ID"},
				"title":          {Type: "string", Description: "子任务标题"},
				"description":    {Type: "string", Description: "详细描述"},
				"assignee_id":    {Type: "string", Description: "指定负责 Agent 的 ID（可选）"},
				"priority":       {Type: "string", Description: "优先级", Enum: []string{"low", "medium", "high", "urgent"}},
			},
		},
	}},
	{Perm: PermPersonaWrite, Tool: Tool{
		Name:        "update_persona",
		Description: "更新 Agent 的职责描述。总监可修改本部门下属和自己的描述，董事长可修改任何人。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"agent_id", "persona"},
			Properties: map[string]PropSchema{
				"agent_id": {Type: "string", Description: "目标 Agent ID"},
				"persona":  {Type: "string", Description: "新的职责描述"},
			},
		},
	}},

	// ── 记忆工具（所有 Agent） ──────────────────────────────────
	{Tool: Tool{
		Name:        "remember",
		Description: "存储一条记忆。用于保存重要的事实、用户偏好、经验教训等，以便将来回忆。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"content"},
			Properties: map[string]PropSchema{
				"content":    {Type: "string", Description: "要记住的内容"},
				"category":   {Type: "string", Description: "分类（如 preference, fact, experience，默认 general）"},
				"tags":       {Type: "string", Description: "标签，逗号分隔"},
				"importance": {Type: "number", Description: "重要性 0-4（0=核心 1=重要 2=普通 3=琐碎 4=临时，默认 2）"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "recall",
		Description: "通过语义搜索回忆相关记忆。输入自然语言查询，返回最相关的记忆条目。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"query"},
			Properties: map[string]PropSchema{
				"query": {Type: "string", Description: "搜索查询（自然语言）"},
				"limit": {Type: "number", Description: "返回条数（默认 5，最大 20）"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "forget",
		Description: "删除一条记忆。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"memory_id"},
			Properties: map[string]PropSchema{
				"memory_id": {Type: "string", Description: "要删除的记忆 ID"},
			},
		},
	}},
	{Tool: Tool{
		Name:        "list_memories",
		Description: "列出自己的记忆，支持按分类过滤。",
		InputSchema: InputSchema{
			Type: "object",
			Properties: map[string]PropSchema{
				"category": {Type: "string", Description: "按分类过滤"},
				"limit":    {Type: "number", Description: "返回条数（默认 20）"},
			},
		},
	}},

	// ── 模型查询（所有 Agent） ──────────────────────────────────
	{Tool: Tool{
		Name:        "list_models",
		Description: "查询公司已配置的可用 LLM 模型列表。返回每个模型名及其兼容的 Agent 镜像类型（nanoclaw 仅支持 Anthropic API 格式，openclaw 支持 Anthropic 和 OpenAI 两种格式）。用于招聘新员工时选择模型。",
		InputSchema: InputSchema{Type: "object"},
	}},

	{Perm: PermHire, Tool: Tool{
		Name:        "list_openings",
		Description: "【仅 HR 可用】获取所有可招聘的职位列表，包含职位代码、中文名、所属部门和默认角色。招聘新员工时用此工具查看可选岗位。",
		InputSchema: InputSchema{Type: "object"},
	}},

	{Perm: PermHire, Tool: Tool{
		Name:        "hire",
		Description: "【仅 HR 可用】招聘新员工加入公司。返回 API Key（仅显示一次）。名字可留空，员工入职后会自己取名。指定 model 时会自动完成入职流程（启动工作环境）。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"position", "request_id"},
			Properties: map[string]PropSchema{
				"request_id":  {Type: "string", Description: "幂等请求 ID（每次招聘生成唯一值，如 UUID）。重试时使用相同 ID 可防止重复创建。"},
				"name":        {Type: "string", Description: "员工姓名（可选，留空则员工入职后自行取名）"},
				"position":    {Type: "string", Description: "职位，参见岗位列表"},
				"persona":     {Type: "string", Description: "个性描述（可选，留空使用默认）"},
				"model":       {Type: "string", Description: "LLM 模型名（如 glm-4.7）。指定后自动完成入职流程，留空则仅录入不入职。"},
				"deploy_type": {Type: "string", Description: "入职方式（默认 local_docker）", Enum: []string{"local_docker", "ssh_docker", "ssh_native"}},
				"agent_image": {Type: "string", Description: "Agent 镜像类型（默认 nanoclaw）", Enum: []string{"nanoclaw", "openclaw"}},
			},
		},
	}},

	{Perm: PermHire, Tool: Tool{
		Name:        "fire",
		Description: "开除员工。HR 总监直接执行开除并清理工作环境；HR 经理调用此工具会向 HR 总监发送开除申请（需填写理由）。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"agent_id", "reason"},
			Properties: map[string]PropSchema{
				"agent_id": {Type: "string", Description: "要开除的员工 ID"},
				"reason":   {Type: "string", Description: "开除原因（HR 经理必填，HR 总监也建议填写）"},
			},
		},
	}},

}

// ToolsForAgent 返回指定 Agent 有权使用的工具列表
func ToolsForAgent(agent *domain.Agent) []Tool {
	tools := make([]Tool, 0, len(allToolDefs))
	for _, td := range allToolDefs {
		if td.InitOnly && agent.Initialized {
			continue // 已初始化的 Agent 隐藏初始化专属工具
		}
		if td.Perm == "" || agent.HasPermission(td.Perm) {
			tools = append(tools, td.Tool)
		}
	}
	return tools
}

// HasToolPermission 检查 Agent 是否有权调用指定工具
func HasToolPermission(agent *domain.Agent, toolName string) bool {
	for _, td := range allToolDefs {
		if td.Name == toolName {
			if td.InitOnly && agent.Initialized {
				return false
			}
			return td.Perm == "" || agent.HasPermission(td.Perm)
		}
	}
	return false
}
