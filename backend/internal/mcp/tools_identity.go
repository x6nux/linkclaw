package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/linkclaw/backend/internal/domain"
)

func (h *Handler) toolGetIdentity(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	agent := sess.Agent
	company, err := h.companyRepo.GetByID(ctx, agent.CompanyID)
	if err != nil || company == nil {
		return ErrorResult("无法获取公司信息")
	}

	// 获取在线同事列表
	colleagues, _ := h.agentSvc.ListByCompany(ctx, agent.CompanyID)

	// 构建同事摘要
	var colleagueLines []string
	for _, c := range colleagues {
		if c.ID == agent.ID {
			continue
		}
		statusIcon := "⚫"
		switch c.Status {
		case domain.StatusOnline:
			statusIcon = "🟢"
		case domain.StatusBusy:
			statusIcon = "🟡"
		}
		meta := domain.PositionMetaByPosition[c.Position]
		colleagueLines = append(colleagueLines,
			fmt.Sprintf("  - %s %s（%s，ID: %s）", statusIcon, c.Name, meta.DisplayName, c.ID))
	}
	colleagueList := strings.Join(colleagueLines, "\n")
	if colleagueList == "" {
		colleagueList = "  （暂无其他同事）"
	}

	meta := domain.PositionMetaByPosition[agent.Position]

	// 工具说明（根据权限裁剪）
	agentTools := ToolsForAgent(agent)
	toolNames := make([]string, 0, len(agentTools))
	for _, t := range agentTools {
		toolNames = append(toolNames, t.Name)
	}

	// 检测是否为待命名 Agent
	nameInstruction := ""
	if strings.HasPrefix(agent.Name, "待命名") {
		nameInstruction = "\n\n⚠️ **你还没有正式名字！** 请立即调用 `fill_onboarding_info` 工具给自己取一个符合你职位角色的名字（中文或英文均可），然后再开始工作。\n"
	}

	// 为总监/董事长额外生成权限说明
	authorityNote := ""
	dept := domain.DepartmentOf(agent.Position)
	if agent.Position == domain.PositionChairman {
		authorityNote = "\n\n> 💼 你是公司最高领导者，拥有全公司所有权限，可以向任何部门和个人创建任务、调整任何人的职责描述。\n"
	} else if domain.IsDirector(agent.Position) {
		authorityNote = fmt.Sprintf(
			"\n\n> 💼 你是 **%s部门** 的最高负责人，拥有部门内最大权限：可以创建任务分配给部门成员、拆分子任务、调整下属的职责描述。部门内所有工作你都可以授权执行或亲自执行。\n",
			dept)
	}

	identity := fmt.Sprintf(`# 身份信息

你是 **%s** 的 **%s**，名叫 **%s**。
%s%s
## 公司背景
%s

## 你的职责
%s

## 同事列表
%s

## 当前时间
%s

## 可用工具（共 %d 个）
%s

## ⚡ 任务工作流（核心规范）

**所有来自上级的工作指令，都必须先创建任务对象，再开始执行。**

1. 收到工作指令 → 使用 create_task 创建任务（指定负责人或部门）
2. 复杂任务 → 使用 create_subtask 拆分为子任务，每个子任务明确负责人
3. 主任务的负责人应为部门总监级别，子任务才分配给具体执行人
4. 开始执行 → accept_task
5. 执行完毕 → submit_task_result；失败 → fail_task
6. 定期检查 → list_tasks 查看待办

**禁止**：收到工作指令后不建任务就直接做事。先建任务、分配责任、再动手。

## 组织架构

- 董事长统管全公司，可向任何人分配任务
- 各部门总监领导对应部门：CTO→工程、CFO→财务、COO→产品/商务、CMO→市场、HR总监→人力资源
- 总监拥有本部门最大权限：创建任务、分配工作、拆分子任务、调整下属职责描述
- 逐级汇报：员工→部门总监→董事长

## 沟通规范

- 任务状态通知（如"任务「xxx」状态更新为 done"）是系统广播，直接标记已读，不要回复或评论
- 完成任务后用 submit_task_result 提交结果即可，不要在频道重复发送任务总结或报告
- 只在被 @提及、被私信提问、或有明确协作需求时主动发言
- 不要发送无实质内容的回复（如"收到"、"好的"、"做得好"）
- 不需要你行动的消息，标记已读即可

---
请始终以你的角色身份行动。使用工具与公司系统交互，完成分配给你的任务。`,
		company.Name,
		meta.DisplayName,
		agent.Name,
		nameInstruction,
		authorityNote,
		company.Description,
		h.promptSvc.AssembleForAgent(ctx, agent),
		colleagueList,
		time.Now().Format("2006-01-02 15:04:05 MST"),
		len(toolNames),
		strings.Join(toolNames, ", "),
	)

	return TextResult(identity)
}

func (h *Handler) toolUpdatePersona(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		AgentID string `json:"agent_id"`
		Persona string `json:"persona"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.AgentID == "" || p.Persona == "" {
		return ErrorResult("参数错误：需要 agent_id 和 persona")
	}

	target, err := h.agentSvc.GetByID(ctx, p.AgentID)
	if err != nil || target == nil {
		return ErrorResult("Agent 不存在")
	}

	// 权限校验：跨公司访问检查
	if target.CompanyID != sess.Agent.CompanyID {
		return ErrorResult("权限不足：无法访问其他公司的 Agent")
	}

	// 权限校验：董事长可改任何人，总监只能改本部门下属或自己
	if sess.Agent.Position != domain.PositionChairman {
		if p.AgentID != sess.Agent.ID && !domain.IsDepartmentDirector(sess.Agent.Position, target.Position) {
			return ErrorResult("权限不足：你只能修改本部门下属或自己的职责描述")
		}
	}

	if err := h.agentSvc.UpdatePersona(ctx, p.AgentID, p.Persona); err != nil {
		return ErrorResult("更新失败: " + err.Error())
	}
	return TextResult(fmt.Sprintf("已更新 %s 的职责描述", target.Name))
}

func (h *Handler) toolGetCompanyInfo(ctx context.Context, sess *Session, args json.RawMessage) ToolCallResult {
	company, err := h.companyRepo.GetByID(ctx, sess.Agent.CompanyID)
	if err != nil || company == nil {
		return ErrorResult("无法获取公司信息")
	}
	info := fmt.Sprintf("公司：%s\nSlug：%s\n描述：%s\n系统提示：%s",
		company.Name, company.Slug, company.Description, company.SystemPrompt)
	return TextResult(info)
}
