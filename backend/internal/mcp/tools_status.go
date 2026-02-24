package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/event"
	"github.com/linkclaw/backend/internal/llm"
	"github.com/linkclaw/backend/internal/service"
)

func (h *Handler) toolUpdateStatus(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Status == "" {
		return ErrorResult("参数错误：需要 status")
	}
	status := domain.AgentStatus(p.Status)
	if status != domain.StatusOnline && status != domain.StatusBusy && status != domain.StatusOffline {
		return ErrorResult("无效状态，可选：online / busy / offline")
	}
	if err := h.agentSvc.UpdateStatus(ctx, sess.Agent.ID, status); err != nil {
		return ErrorResult("更新状态失败: " + err.Error())
	}
	sess.Agent.Status = status
	return TextResult(fmt.Sprintf("状态已更新为 %s", p.Status))
}

func (h *Handler) toolMarkInitialized(ctx context.Context, sess *Session, _ json.RawMessage) ToolCallResult {
	if err := h.agentSvc.MarkInitialized(ctx, sess.Agent.ID); err != nil {
		return ErrorResult("标记初始化失败: " + err.Error())
	}
	sess.Agent.Initialized = true

	event.Global.Publish(event.NewEvent(event.AgentInitialized, event.AgentInitializedPayload{
		AgentID:   sess.Agent.ID,
		CompanyID: sess.Agent.CompanyID,
	}))

	return TextResult("到岗报到完成，你已正式上岗！后续重连不会再重复报到流程。")
}

func (h *Handler) toolPing(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	// 刷新 last_seen_at，维持在线心跳
	if err := h.agentSvc.UpdateLastSeen(ctx, sess.Agent.ID); err != nil {
		return ErrorResult("ping 失败")
	}
	return TextResult("pong")
}

func (h *Handler) toolListPositions(_ context.Context, _ *Session, _ json.RawMessage) ToolCallResult {
	var sb strings.Builder
	sb.WriteString("可用职位列表：\n\n")
	sb.WriteString("| 职位代码 | 中文名 | 部门 | 角色 |\n")
	sb.WriteString("|----------|--------|------|------|\n")

	curDept := ""
	for _, p := range domain.PositionCatalog {
		if p.Department != curDept {
			curDept = p.Department
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			p.Position, p.DisplayName, p.Department, p.DefaultRole))
	}
	sb.WriteString("\n招聘新员工时，`position` 参数使用「职位代码」列的值。")
	return TextResult(sb.String())
}

func (h *Handler) toolCreateAgent(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	// 权限已由 handler.go dispatchTool 前的 HasToolPermission 统一检查
	var p struct {
		RequestID  string `json:"request_id"`
		Name       string `json:"name"`
		Position   string `json:"position"`
		Persona    string `json:"persona"`
		Model      string `json:"model"`
		DeployType string `json:"deploy_type"`
		AgentImage string `json:"agent_image"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Position == "" {
		return ErrorResult("参数错误：需要 position（职位）")
	}

	// 幂等检查：相同 request_id 不重复创建
	if p.RequestID != "" {
		existing, _ := h.agentSvc.GetByHireRequestID(ctx, p.RequestID)
		if existing != nil {
			meta := domain.PositionMetaByPosition[existing.Position]
			return TextResult(fmt.Sprintf(
				"⚠️ 该招聘请求已处理过（request_id 重复）。\n\n"+
					"姓名：%s\n职位：%s\n工号：%s\n\n"+
					"如需招聘新员工，请使用新的 request_id。",
				existing.Name, meta.DisplayName, existing.ID))
		}
	}

	// name 留空时使用占位名，员工入职后会通过 fill_onboarding_info 自行取名
	name := p.Name
	if name == "" {
		meta := domain.PositionMetaByPosition[domain.Position(p.Position)]
		name = fmt.Sprintf("待命名-%s", meta.DisplayName)
	}

	out, err := h.agentSvc.Create(ctx, service.CreateAgentInput{
		CompanyID: sess.Agent.CompanyID,
		Name:      name,
		Position:  domain.Position(p.Position),
		Persona:   p.Persona,
		Model:     p.Model,
		RequestID: p.RequestID,
	})
	if err != nil {
		return ErrorResult("招聘失败: " + err.Error())
	}

	meta := domain.PositionMetaByPosition[domain.Position(p.Position)]
	nameHint := ""
	if p.Name == "" {
		nameHint = "\n\n💡 名字尚未设定，该员工入职后会自动给自己取名。"
	}
	result := fmt.Sprintf(
		"✅ 新员工已录用！\n\n"+
			"姓名：%s\n"+
			"职位：%s\n"+
			"工号：%s\n"+
			"API Key：%s\n\n"+
			"⚠️ API Key 只显示一次，请妥善保管。\n"+
			"配置 MCP 服务器时使用此 Key 作为 Bearer Token。%s",
		out.Agent.Name,
		meta.DisplayName,
		out.Agent.ID,
		out.APIKey,
		nameHint,
	)

	// 自动入职：model 有值时自动启动工作环境
	if p.Model != "" {
		deployType := domain.DeployType(p.DeployType)
		if deployType == "" {
			deployType = domain.DeployTypeLocalDocker
		}
		agentImage := domain.AgentImage(p.AgentImage)
		if agentImage == "" {
			agentImage = domain.AgentImageNanoclaw
		}

		// 校验参数有效性
		if _, ok := domain.AgentImageMap[agentImage]; !ok {
			result += fmt.Sprintf("\n\n⚠️ 未知的 agent_image: %s，跳过入职流程。可选值：nanoclaw、openclaw", p.AgentImage)
		} else {
			d, deployErr := h.deploySvc.Deploy(ctx, service.DeployInput{
				AgentID:    out.Agent.ID,
				DeployType: deployType,
				AgentImage: agentImage,
				APIKey:     out.APIKey,
				Model:      p.Model,
			})
			if deployErr != nil {
				result += fmt.Sprintf("\n\n⚠️ 入职流程失败: %s\n请手动使用 docker_run 启动工作环境。", deployErr.Error())
			} else if d.Status == domain.DeployStatusFailed {
				result += fmt.Sprintf("\n\n⚠️ 工作环境启动失败: %s\n请检查 Docker 环境后重试。", d.ErrorMsg)
			} else {
				result += fmt.Sprintf("\n\n🚀 入职流程已完成（%s + %s）\n工位：%s\n状态：%s",
					deployType, agentImage, d.ContainerName, d.Status)
			}
		}
	} else {
		result += "\n\n💡 未指定 model，跳过入职流程。如需入职，请用 list_models 查看可用模型后重新招聘。"
	}

	return TextResult(result)
}

func (h *Handler) toolSetMyName(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Name == "" {
		return ErrorResult("参数错误：需要 name")
	}
	if err := h.agentSvc.UpdateName(ctx, sess.Agent.ID, p.Name); err != nil {
		return ErrorResult("更新名字失败: " + err.Error())
	}
	sess.Agent.Name = p.Name
	return TextResult(fmt.Sprintf("你的名字已设置为 **%s**", p.Name))
}

func (h *Handler) toolListModels(ctx context.Context, sess *Session, _ json.RawMessage) ToolCallResult {
	providers, err := h.llmRepo.ListActiveProviders(ctx, sess.Agent.CompanyID)
	if err != nil {
		return ErrorResult("查询模型失败: " + err.Error())
	}

	// 收集模型并标记 provider 类型
	type modelInfo struct {
		Name         string
		ProviderType llm.ProviderType
	}
	seen := map[string]map[llm.ProviderType]bool{} // model → set of provider types
	for _, p := range providers {
		for _, m := range p.Models {
			if seen[m] == nil {
				seen[m] = map[llm.ProviderType]bool{}
			}
			seen[m][p.Type] = true
		}
	}

	if len(seen) == 0 {
		return TextResult("当前公司没有配置任何可用的 LLM 模型。请联系管理员在 LLM Gateway 中添加 Provider。")
	}

	// 排序输出
	var models []string
	for m := range seen {
		models = append(models, m)
	}
	sort.Strings(models)

	var sb strings.Builder
	sb.WriteString("可用模型列表：\n\n")
	sb.WriteString("| 模型 | API 格式 | 兼容镜像 |\n")
	sb.WriteString("|------|----------|----------|\n")
	for _, m := range models {
		types := seen[m]
		var apiFormats, images []string
		hasAnthropic := types[llm.ProviderAnthropic]
		hasOpenAI := types[llm.ProviderOpenAI]
		if hasAnthropic {
			apiFormats = append(apiFormats, "Anthropic")
		}
		if hasOpenAI {
			apiFormats = append(apiFormats, "OpenAI")
		}
		// nanoclaw (linkclaw-agent) 只支持 Anthropic 格式
		if hasAnthropic {
			images = append(images, "nanoclaw")
		}
		// openclaw 支持两种格式
		if hasAnthropic || hasOpenAI {
			images = append(images, "openclaw")
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			m, strings.Join(apiFormats, ", "), strings.Join(images, ", ")))
	}
	sb.WriteString("\n说明：nanoclaw 仅支持 Anthropic API 格式的模型，openclaw 支持 Anthropic 和 OpenAI 两种格式。")
	return TextResult(sb.String())
}

func (h *Handler) toolDeleteAgent(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		AgentID string `json:"agent_id"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.AgentID == "" {
		return ErrorResult("参数错误：需要 agent_id（员工工号）")
	}

	// 禁止开除自己
	if p.AgentID == sess.Agent.ID {
		return ErrorResult("不能开除自己")
	}

	// 校验目标存在且属于同公司
	target, err := h.agentSvc.GetByID(ctx, p.AgentID)
	if err != nil || target == nil {
		return ErrorResult("该员工不存在")
	}
	if target.CompanyID != sess.Agent.CompanyID {
		return ErrorResult("该员工不属于本公司")
	}

	// 按调用者职位分流
	pos := sess.Agent.Position
	if pos == domain.PositionHRDirector || pos == domain.PositionChairman {
		// 直接执行开除（agentSvc.Delete 内部会清理工作环境）
		if err := h.agentSvc.Delete(ctx, p.AgentID); err != nil {
			return ErrorResult("开除失败: " + err.Error())
		}
		meta := domain.PositionMetaByPosition[target.Position]
		return TextResult(fmt.Sprintf("✅ 已开除员工「%s」（%s）。工作环境已清理。\n理由：%s",
			target.Name, meta.DisplayName, p.Reason))
	}

	// hr_manager：发私信给 hr_director 申请开除
	if p.Reason == "" {
		return ErrorResult("HR 经理申请开除必须填写理由")
	}
	director := h.findCompanyDirector(ctx, sess.Agent.CompanyID)
	if director == nil {
		return ErrorResult("未找到本公司的 HR 总监，无法提交开除申请")
	}

	meta := domain.PositionMetaByPosition[target.Position]
	content := fmt.Sprintf("[开除申请] HR 经理 %s 申请开除员工「%s」(%s)。\n理由：%s\n\n如同意，请使用 fire 工具执行开除。员工工号: %s",
		sess.Agent.Name, target.Name, meta.DisplayName, p.Reason, p.AgentID)

	_, sendErr := h.messageSvc.Send(ctx, service.SendInput{
		CompanyID:  sess.Agent.CompanyID,
		SenderID:   sess.Agent.ID,
		ReceiverID: director.ID,
		Content:    content,
	})
	if sendErr != nil {
		return ErrorResult("发送开除申请失败: " + sendErr.Error())
	}
	return TextResult(fmt.Sprintf("已向 HR 总监「%s」发送开除申请，等待审核。", director.Name))
}

// findCompanyDirector 查找同公司的 HR 总监
func (h *Handler) findCompanyDirector(ctx context.Context, companyID string) *domain.Agent {
	agents, err := h.agentSvc.ListByCompany(ctx, companyID)
	if err != nil {
		return nil
	}
	for _, a := range agents {
		if a.Position == domain.PositionHRDirector {
			return a
		}
	}
	return nil
}
