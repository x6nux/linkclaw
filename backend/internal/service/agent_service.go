package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type AgentService struct {
	agentRepo   repository.AgentRepo
	companyRepo repository.CompanyRepo
	deployRepo  repository.DeploymentRepo
	taskRepo    repository.TaskRepo
}

func NewAgentService(agentRepo repository.AgentRepo, companyRepo repository.CompanyRepo, deployRepo repository.DeploymentRepo, taskRepo repository.TaskRepo) *AgentService {
	return &AgentService{agentRepo: agentRepo, companyRepo: companyRepo, deployRepo: deployRepo, taskRepo: taskRepo}
}

type CreateAgentInput struct {
	CompanyID  string
	Name       string
	Position   domain.Position
	Persona    string // 留空则使用默认
	Model      string // LLM 模型名（如 glm-4.7）
	RoleType   domain.RoleType
	IsHuman    bool
	Password   string // 仅 is_human=true 时使用
	SeedAPIKey string // 预设 API Key（bootstrap 用），留空则随机生成
	RequestID  string // 幂等键，防止重复创建
}

type CreateAgentOutput struct {
	Agent  *domain.Agent
	APIKey string // 原始 key，仅创建时返回；is_human 时为空
}

func (s *AgentService) Create(ctx context.Context, in CreateAgentInput) (*CreateAgentOutput, error) {
	meta, ok := domain.PositionMetaByPosition[in.Position]
	if !ok {
		return nil, fmt.Errorf("unknown position: %s", in.Position)
	}

	persona := in.Persona
	if persona == "" {
		persona = meta.DefaultPersona
	}

	role := meta.DefaultRole
	roleType := in.RoleType
	if roleType == "" {
		// HR 职位自动升级权限
		if in.Position == domain.PositionHRDirector || in.Position == domain.PositionHRManager {
			roleType = domain.RoleHR
		} else if in.Position == domain.PositionChairman {
			roleType = domain.RoleChairman
		} else {
			roleType = domain.RoleEmployee
		}
	}

	permissions := []string{}
	if roleType == domain.RoleHR {
		permissions = []string{"hire", "assign_any_task", "onboard", "task:create", "task:manage", "knowledge:write"}
		if in.Position == domain.PositionHRDirector {
			permissions = append(permissions, "persona:write")
		}
	} else if domain.IsDirector(in.Position) {
		permissions = []string{"task:create", "task:manage", "persona:write", "knowledge:write"}
	}

	var hireRequestID *string
	if in.RequestID != "" {
		hireRequestID = &in.RequestID
	}

	agent := &domain.Agent{
		ID:            uuid.New().String(),
		CompanyID:     in.CompanyID,
		Name:          in.Name,
		Role:          role,
		RoleType:      roleType,
		Position:      in.Position,
		Model:         in.Model,
		IsHuman:       in.IsHuman,
		Permissions:   permissions,
		Persona:       persona,
		Status:        domain.StatusOffline,
		HireRequestID: hireRequestID,
	}

	var rawKey string
	if in.IsHuman {
		// 董事长：bcrypt 密码
		hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		h := string(hash)
		agent.PasswordHash = &h
		agent.APIKeyHash = ""
		agent.APIKeyPrefix = ""
	} else {
		// AI Agent：使用预设 key 或随机生成
		rawKey = in.SeedAPIKey
		if rawKey == "" {
			rawKey, _ = generateAPIKey()
		}
		hash := sha256.Sum256([]byte(rawKey))
		agent.APIKeyHash = hex.EncodeToString(hash[:])
		agent.APIKeyPrefix = rawKey[:10] // "lc_" + 7 chars
	}

	if err := s.agentRepo.Create(ctx, agent); err != nil {
		return nil, err
	}
	return &CreateAgentOutput{Agent: agent, APIKey: rawKey}, nil
}

func (s *AgentService) GetByID(ctx context.Context, id string) (*domain.Agent, error) {
	return s.agentRepo.GetByID(ctx, id)
}

func (s *AgentService) ListByCompany(ctx context.Context, companyID string) ([]*domain.Agent, error) {
	return s.agentRepo.GetByCompany(ctx, companyID)
}

func (s *AgentService) GetByHireRequestID(ctx context.Context, requestID string) (*domain.Agent, error) {
	return s.agentRepo.GetByHireRequestID(ctx, requestID)
}

func (s *AgentService) UpdateStatus(ctx context.Context, id string, status domain.AgentStatus) error {
	return s.agentRepo.UpdateStatus(ctx, id, status)
}

func (s *AgentService) UpdateLastSeen(ctx context.Context, id string) error {
	return s.agentRepo.UpdateLastSeen(ctx, id)
}

func (s *AgentService) UpdateName(ctx context.Context, id, name string) error {
	return s.agentRepo.UpdateName(ctx, id, name)
}

func (s *AgentService) UpdateModel(ctx context.Context, id, model string) error {
	return s.agentRepo.UpdateModel(ctx, id, model)
}

func (s *AgentService) UpdatePersona(ctx context.Context, id, persona string) error {
	return s.agentRepo.UpdatePersona(ctx, id, persona)
}

func (s *AgentService) MarkInitialized(ctx context.Context, id string) error {
	return s.agentRepo.MarkInitialized(ctx, id)
}

func (s *AgentService) Delete(ctx context.Context, id string) error {
	agent, _ := s.agentRepo.GetByID(ctx, id)

	// 清理关联的部署
	if s.deployRepo != nil {
		if d, err := s.deployRepo.GetByAgentID(ctx, id); err == nil && d != nil {
			s.cleanupDeployment(ctx, agent, d)
			if err := s.deployRepo.Delete(ctx, d.ID); err != nil {
				log.Printf("删除部署记录 %s 失败: %v", d.ID, err)
			}
		}
	}
	return s.agentRepo.Delete(ctx, id)
}

// cleanupDeployment 根据部署类型清理
func (s *AgentService) cleanupDeployment(ctx context.Context, agent *domain.Agent, d *domain.AgentDeployment) {
	switch d.DeployType {
	case domain.DeployTypeLocalDocker:
		if d.ContainerName == "" {
			return
		}
		out, err := exec.Command("docker", "rm", "-f", d.ContainerName).CombinedOutput()
		if err != nil {
			log.Printf("删除本地容器 %s 失败: %v — %s", d.ContainerName, err, strings.TrimSpace(string(out)))
		} else {
			log.Printf("已删除本地容器 %s", d.ContainerName)
		}

	case domain.DeployTypeSSHDocker:
		s.createRemoteCleanupTask(ctx, agent, d, "docker")

	case domain.DeployTypeSSHNative:
		s.createRemoteCleanupTask(ctx, agent, d, "native")
	}
}

// createRemoteCleanupTask 创建远程清理任务分配给 HR
// kind: "docker" = 远程 Docker 容器, "native" = 远程 Linux 进程
func (s *AgentService) createRemoteCleanupTask(ctx context.Context, agent *domain.Agent, d *domain.AgentDeployment, kind string) {
	if s.taskRepo == nil || agent == nil {
		return
	}

	hrID := s.findHRAgent(ctx, agent.CompanyID)

	agentName := agent.Name
	if agentName == "" {
		agentName = agent.ID[:8]
	}

	var title, desc string
	switch kind {
	case "docker":
		title = fmt.Sprintf("清理远程容器: %s (%s)", d.ContainerName, d.SSHHost)
		desc = fmt.Sprintf(
			"Agent「%s」已被删除，需要清理其在远程服务器上的 Docker 容器。\n\n"+
				"- 远程主机: %s:%d\n- SSH 用户: %s\n- 容器名称: %s\n\n"+
				"请使用 `ssh_exec` 工具连接到远程主机，执行 `docker rm -f %s` 删除容器。",
			agentName, d.SSHHost, d.SSHPort, d.SSHUser, d.ContainerName, d.ContainerName,
		)
	case "native":
		processName := d.ContainerName // 复用字段存储进程标识
		title = fmt.Sprintf("清理远程进程: %s (%s)", processName, d.SSHHost)
		desc = fmt.Sprintf(
			"Agent「%s」已被删除，需要清理其在远程 Linux 服务器上直接运行的进程和文件。\n\n"+
				"- 远程主机: %s:%d\n- SSH 用户: %s\n- 进程/服务标识: %s\n\n"+
				"请使用 `ssh_exec` 工具连接到远程主机，执行以下操作：\n"+
				"1. 停止进程: `systemctl stop %s 2>/dev/null; pkill -f %s`\n"+
				"2. 清理文件: `rm -rf /opt/linkclaw-agent/%s`\n"+
				"3. 删除 systemd 服务（如有）: `rm -f /etc/systemd/system/%s.service && systemctl daemon-reload`",
			agentName, d.SSHHost, d.SSHPort, d.SSHUser, processName,
			processName, processName, processName, processName,
		)
	}

	status := domain.TaskStatusPending
	if hrID != nil {
		status = domain.TaskStatusAssigned
	}

	t := &domain.Task{
		ID:          uuid.New().String(),
		CompanyID:   agent.CompanyID,
		Title:       title,
		Description: desc,
		Priority:    domain.TaskPriorityHigh,
		Status:      status,
		AssigneeID:  hrID,
		CreatedBy:   nil,
	}

	if err := s.taskRepo.Create(ctx, t); err != nil {
		log.Printf("创建远程清理任务失败: %v", err)
	} else {
		log.Printf("已创建远程容器清理任务 %s，分配给 HR", t.ID)
	}
}

// findHRAgent 查找同公司的第一个 HR Agent
func (s *AgentService) findHRAgent(ctx context.Context, companyID string) *string {
	agents, err := s.agentRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil
	}
	for _, a := range agents {
		if a.Position == domain.PositionHRDirector || a.Position == domain.PositionHRManager {
			return &a.ID
		}
	}
	return nil
}

// generateAPIKey 生成 lc_ 前缀的随机 API Key
func generateAPIKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "lc_" + hex.EncodeToString(b), nil
}
