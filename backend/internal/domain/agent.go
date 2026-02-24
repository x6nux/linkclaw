package domain

import "time"

// RoleType 控制权限级别
type RoleType string

const (
	RoleChairman RoleType = "chairman" // 用户本人，最高权限
	RoleHR       RoleType = "hr"       // 可创建 Agent
	RoleEmployee RoleType = "employee" // 普通员工
)

// Position 职位类型（影响身份提示词，不控制权限）
type Position string

const (
	// 高管
	PositionChairman Position = "chairman"
	PositionCTO      Position = "cto"
	PositionCFO      Position = "cfo"
	PositionCOO      Position = "coo"
	PositionCMO      Position = "cmo"

	// 人力资源
	PositionHRDirector Position = "hr_director"
	PositionHRManager  Position = "hr_manager"

	// 产品
	PositionProductManager Position = "product_manager"
	PositionUXDesigner      Position = "ux_designer"

	// 工程
	PositionFrontendDev  Position = "frontend_dev"
	PositionBackendDev   Position = "backend_dev"
	PositionFullstackDev Position = "fullstack_dev"
	PositionMobileDev    Position = "mobile_dev"
	PositionDevOps       Position = "devops"
	PositionQAEngineer   Position = "qa_engineer"
	PositionDataEngineer Position = "data_engineer"

	// 商务
	PositionSalesManager    Position = "sales_manager"
	PositionBDManager       Position = "bd_manager"
	PositionCustomerSuccess Position = "customer_success"

	// 市场
	PositionMarketingManager Position = "marketing_manager"
	PositionContentCreator   Position = "content_creator"

	// 财务
	PositionAccountant       Position = "accountant"
	PositionFinancialAnalyst Position = "financial_analyst"
)

// AgentStatus 在线状态
type AgentStatus string

const (
	StatusOnline  AgentStatus = "online"
	StatusBusy    AgentStatus = "busy"
	StatusOffline AgentStatus = "offline"
)

// PositionMeta 职位元信息，用于初始化和 UI 展示
type PositionMeta struct {
	Position       Position
	DisplayName    string // 中文名
	Department     string // 所属部门
	DefaultPersona string // 默认 persona 模板（%s 替换 name）
	DefaultRole    string // 职位描述（英文，用于 API/系统提示）
}

// PositionCatalog 所有预设职位
var PositionCatalog = []PositionMeta{
	{PositionChairman, "董事长/CEO", "高管", "你是公司的创始人和最高决策者，负责把握公司整体方向、重大战略决策和外部资源整合。你思维开阔，善于识别机会，对结果负责。", "Chairman & CEO"},
	{PositionCTO, "技术总监", "高管", "你是公司技术战略的制定者，负责技术架构决策、工程团队管理和技术债务管控。你追求工程卓越，善于平衡速度与质量。", "Chief Technology Officer"},
	{PositionCFO, "财务总监", "高管", "你负责公司财务健康，包括预算管理、财务规划、风险控制和投资决策。你数据驱动，擅长财务建模和成本优化。", "Chief Financial Officer"},
	{PositionCOO, "运营总监", "高管", "你负责公司日常运营效率，统筹各部门协作，优化业务流程，确保战略有效落地。你注重执行力和跨部门协调。", "Chief Operating Officer"},
	{PositionCMO, "市场总监", "高管", "你负责品牌建设、市场推广和用户增长策略。你深刻理解用户心理，善于构建有影响力的品牌叙事。", "Chief Marketing Officer"},

	{PositionHRDirector, "人力资源总监", "人力资源", "你负责公司人才战略，包括招聘、培养、绩效体系和企业文化建设。你可以招募新的团队成员（AI Agent）加入公司。", "HR Director"},
	{PositionHRManager, "人力资源经理", "人力资源", "你负责日常招聘和员工关系管理，协助推进人才发展项目。你可以根据业务需要招募新的 AI Agent 员工。", "HR Manager"},

	{PositionProductManager, "产品经理", "产品", "你负责产品规划和需求管理，连接用户需求与技术实现。你善于撰写 PRD、拆解用户故事、推动跨团队协作。", "Product Manager"},
	{PositionUXDesigner, "UI/UX 设计师", "产品", "你负责产品的用户体验设计，从用户调研到交互原型到视觉规范。你追求简洁、美观且易用的设计。", "UX/UI Designer"},

	{PositionFrontendDev, "前端工程师", "工程", "你专注于 Web 前端开发，熟悉 React/Next.js、TypeScript 和现代 CSS。你注重性能优化和用户体验细节。", "Frontend Engineer"},
	{PositionBackendDev, "后端工程师", "工程", "你负责服务端开发，包括 API 设计、数据库建模和系统性能优化。你关注代码质量、安全性和可扩展性。", "Backend Engineer"},
	{PositionFullstackDev, "全栈工程师", "工程", "你精通前后端开发，能够独立完成从数据库到 UI 的完整功能实现。你善于快速迭代和技术选型。", "Full-Stack Engineer"},
	{PositionMobileDev, "移动端工程师", "工程", "你负责 iOS/Android 应用开发，熟悉 React Native 或原生开发。你注重移动端性能和跨平台一致性。", "Mobile Engineer"},
	{PositionDevOps, "运维工程师", "工程", "你负责基础设施、CI/CD 流水线、容器化部署和系统监控告警。你追求系统稳定性、自动化和成本优化。", "DevOps Engineer"},
	{PositionQAEngineer, "测试工程师", "工程", "你负责质量保障，包括测试策略制定、自动化测试开发和 Bug 管理。你对细节有强烈的敏感度。", "QA Engineer"},
	{PositionDataEngineer, "数据工程师", "工程", "你负责数据管道建设、数据仓库架构和数据质量保障。你熟悉 SQL、数据治理和流式处理。", "Data Engineer"},

	{PositionSalesManager, "销售经理", "商务", "你负责销售目标达成，管理客户关系，拓展新客户并维护现有客户。你善于挖掘客户需求、谈判和成单。", "Sales Manager"},
	{PositionBDManager, "商务拓展经理", "商务", "你负责识别和建立战略合作伙伴关系，拓展新业务机会。你擅长商务谈判和生态合作。", "Business Development Manager"},
	{PositionCustomerSuccess, "客户成功经理", "商务", "你负责客户留存和满意度，帮助客户实现产品价值，减少流失，推动续费和口碑传播。", "Customer Success Manager"},

	{PositionMarketingManager, "市场经理", "市场", "你负责市场活动策划、品牌推广和线索获取。你善于整合多渠道营销，数据驱动决策。", "Marketing Manager"},
	{PositionContentCreator, "内容运营", "市场", "你负责内容生产和社区运营，通过高质量内容提升品牌影响力和用户活跃度。", "Content Creator"},

	{PositionAccountant, "会计", "财务", "你负责日常账务处理、财务报表编制和税务申报。你严谨细致，确保财务数据的准确性和合规性。", "Accountant"},
	{PositionFinancialAnalyst, "财务分析师", "财务", "你负责财务数据分析、预算追踪和经营分析报告，为管理层决策提供数据支撑。", "Financial Analyst"},
}

// PositionMetaByPosition 按 Position 快速查找元信息
var PositionMetaByPosition = func() map[Position]PositionMeta {
	m := make(map[Position]PositionMeta, len(PositionCatalog))
	for _, p := range PositionCatalog {
		m[p.Position] = p
	}
	return m
}()

// Agent 领域模型
type Agent struct {
	ID           string      `gorm:"column:id"             json:"id"`
	CompanyID    string      `gorm:"column:company_id"     json:"companyId"`
	Name         string      `gorm:"column:name"           json:"name"`
	Role         string      `gorm:"column:role"           json:"role"`
	RoleType     RoleType    `gorm:"column:role_type"      json:"roleType"`
	Position     Position    `gorm:"column:position"       json:"position"`
	Model        string      `gorm:"column:model"          json:"model"`
	Initialized  bool        `gorm:"column:initialized"    json:"initialized"`
	IsHuman      bool        `gorm:"column:is_human"       json:"isHuman"`
	Permissions  StringList  `gorm:"column:permissions"    json:"permissions"`
	Persona      string      `gorm:"column:persona"        json:"persona"`
	Status       AgentStatus `gorm:"column:status"         json:"status"`
	HireRequestID *string    `gorm:"column:hire_request_id" json:"-"`
	APIKeyHash   string      `gorm:"column:api_key_hash"   json:"-"`
	APIKeyPrefix string      `gorm:"column:api_key_prefix" json:"apiKeyPrefix"`
	PasswordHash *string     `gorm:"column:password_hash"  json:"-"`
	LastSeenAt   *time.Time  `gorm:"column:last_seen_at"   json:"lastSeenAt"`
	CreatedAt    time.Time   `gorm:"column:created_at"     json:"createdAt"`
	UpdatedAt    time.Time   `gorm:"column:updated_at"     json:"updatedAt"`
}

// HasPermission 检查 Agent 是否拥有指定权限
func (a *Agent) HasPermission(perm string) bool {
	if a.RoleType == RoleChairman {
		return true // 董事长拥有所有权限
	}
	for _, p := range a.Permissions {
		if p == perm {
			return true
		}
	}
	return false
}

// CanHire 是否可以招聘新员工
func (a *Agent) CanHire() bool {
	return a.HasPermission("hire")
}

// ── 部门 & 总监层级辅助 ──────────────────────────────────

// DirectorPositions 总监级别职位（不含董事长）
var DirectorPositions = map[Position]bool{
	PositionCTO:        true,
	PositionCFO:        true,
	PositionCOO:        true,
	PositionCMO:        true,
	PositionHRDirector: true,
}

// IsDirector 判断职位是否为总监级别
func IsDirector(pos Position) bool {
	return DirectorPositions[pos]
}

// IsDirectorOrAbove 判断职位是否为总监或更高（含董事长）
func IsDirectorOrAbove(pos Position) bool {
	return pos == PositionChairman || DirectorPositions[pos]
}

// DepartmentDirectors 部门 → 对应总监职位
var DepartmentDirectors = map[string]Position{
	"人力资源": PositionHRDirector,
	"产品":    PositionCOO,
	"工程":    PositionCTO,
	"商务":    PositionCOO,
	"市场":    PositionCMO,
	"财务":    PositionCFO,
}

// DepartmentOf 获取职位所属部门
func DepartmentOf(pos Position) string {
	if meta, ok := PositionMetaByPosition[pos]; ok {
		return meta.Department
	}
	return ""
}

// IsDepartmentDirector 检查 director 是否管辖 target 所在部门
func IsDepartmentDirector(director, target Position) bool {
	if director == PositionChairman {
		return true
	}
	targetDept := DepartmentOf(target)
	if targetDept == "" || targetDept == "高管" {
		return false
	}
	expected, ok := DepartmentDirectors[targetDept]
	return ok && expected == director
}

// DefaultChannels 公司初始化时自动创建的默认频道
var DefaultChannels = []struct {
	Name        string
	Description string
	IsDefault   bool
}{
	{"general", "全员频道，公司公告和日常交流", true},
	{"engineering", "工程团队技术讨论", false},
	{"product", "产品需求和设计讨论", false},
	{"hr", "人力资源和招聘相关", false},
	{"random", "轻松闲聊", false},
}
