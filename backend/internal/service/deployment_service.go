package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type DeploymentService struct {
	deployRepo  repository.DeploymentRepo
	agentRepo   repository.AgentRepo
	companyRepo repository.CompanyRepo
}

func NewDeploymentService(deployRepo repository.DeploymentRepo, agentRepo repository.AgentRepo, companyRepo repository.CompanyRepo) *DeploymentService {
	return &DeploymentService{deployRepo: deployRepo, agentRepo: agentRepo, companyRepo: companyRepo}
}

type DeployInput struct {
	AgentID    string
	DeployType domain.DeployType
	AgentImage domain.AgentImage
	APIKey     string // 原始 API Key，local_docker 启动时注入
	Model      string // LLM 模型名（部署时由用户选择）
	// SSH 连接信息（ssh_docker / ssh_native 时需要）
	SSHHost     string
	SSHPort     int
	SSHUser     string
	SSHPassword string
	SSHKey      string
}

// Deploy 创建部署记录；local_docker 会直接启动容器，其他类型等 HR Agent 编排
func (s *DeploymentService) Deploy(ctx context.Context, in DeployInput) (*domain.AgentDeployment, error) {
	image, ok := domain.AgentImageMap[in.AgentImage]
	if !ok {
		return nil, fmt.Errorf("unknown agent image: %s", in.AgentImage)
	}

	// 从公司设置覆盖默认镜像名
	agent, _ := s.agentRepo.GetByID(ctx, in.AgentID)
	if agent != nil {
		if company, _ := s.companyRepo.GetByID(ctx, agent.CompanyID); company != nil {
			if in.AgentImage == domain.AgentImageNanoclaw && company.NanoclawImage != "" {
				image = company.NanoclawImage
			}
			if in.AgentImage == domain.AgentImageOpenclaw && company.OpenclawPluginURL != "" {
				image = company.OpenclawPluginURL
			}
		}
	}

	containerName := fmt.Sprintf("agent-%s", in.AgentID[:8])
	d := &domain.AgentDeployment{
		ID:            uuid.New().String(),
		AgentID:       in.AgentID,
		DeployType:    in.DeployType,
		AgentImage:    in.AgentImage,
		ContainerName: containerName,
		SSHHost:       in.SSHHost,
		SSHPort:       in.SSHPort,
		SSHUser:       in.SSHUser,
		SSHPassword:   in.SSHPassword,
		SSHKey:        in.SSHKey,
		Status:        domain.DeployStatusPending,
	}

	if err := s.deployRepo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("save deployment: %w", err)
	}

	// local_docker：直接启动容器
	if in.DeployType == domain.DeployTypeLocalDocker {
		if err := s.runLocalDocker(d, image, in.APIKey, in.Model); err != nil {
			d.Status = domain.DeployStatusFailed
			d.ErrorMsg = err.Error()
			s.deployRepo.UpdateStatus(ctx, d.ID, d.Status, d.ErrorMsg) //nolint:errcheck
			return d, nil // 返回记录但带错误信息
		}
		d.Status = domain.DeployStatusRunning
		s.deployRepo.UpdateStatus(ctx, d.ID, d.Status, "") //nolint:errcheck
	}

	return d, nil
}

// getCompanyByAgent 通过 agent 查找所属公司
func (s *DeploymentService) getCompanyByAgent(agentID string) *domain.Company {
	agent, _ := s.agentRepo.GetByID(context.Background(), agentID)
	if agent == nil {
		return nil
	}
	company, _ := s.companyRepo.GetByID(context.Background(), agent.CompanyID)
	return company
}

// runLocalDocker 在本地 Docker 启动 Agent 容器
func (s *DeploymentService) runLocalDocker(d *domain.AgentDeployment, image, apiKey, model string) error {
	network := os.Getenv("DOCKER_NETWORK")
	if network == "" {
		network = "deploy_default"
	}

	// 先清理同名旧容器（忽略错误）
	exec.Command("docker", "rm", "-f", d.ContainerName).Run() //nolint:errcheck

	args := []string{
		"run", "-d",
		"--name", d.ContainerName,
		"--restart", "unless-stopped",
		"--network", network,
	}

	switch d.AgentImage {
	case domain.AgentImageNanoclaw:
		// nanoclaw 主进程：Docker-in-Docker 需要宿主机路径做 volume mount
		// nanoclaw 的 toHostPath() 用 NANOCLAW_HOST_ROOT 把容器内 /app/... 转为宿主机路径
		dataDir := os.Getenv("AGENT_DATA_DIR")
		if dataDir == "" {
			dataDir = "/tmp/linkclaw-agents"
		}
		agentDataDir := fmt.Sprintf("%s/%s", dataDir, d.AgentID[:8])

		args = append(args,
			"-v", "/var/run/docker.sock:/var/run/docker.sock",
			"-v", agentDataDir+"/data:/app/data",
			"-v", agentDataDir+"/groups:/app/groups",
			"-v", agentDataDir+"/store:/app/store",
			"-e", "NANOCLAW_HOST_ROOT="+agentDataDir,
			"-e", "NANOCLAW_CONTAINER_PREFIX="+d.ContainerName,
			"-e", "LINKCLAW_API_KEY="+apiKey,
			"-e", "LINKCLAW_BASE_URL=http://backend:8080",
			"-e", "DOCKER_NETWORK="+network,
			"-e", "ANTHROPIC_API_KEY="+apiKey,
			"-e", "ANTHROPIC_BASE_URL=http://backend:8080",
			"-e", "ANTHROPIC_MODEL="+model,
			"-e", "ANTHROPIC_SMALL_FAST_MODEL="+model,
			"-e", "IDLE_TIMEOUT=60000", // 1 分钟空闲后关闭子容器（默认 30 分钟对 linkclaw 过长）
		)
	default:
		// openclaw 等：直接连接 LLM 代理和 SSE
		wsURL := "ws://backend:8080/api/v1/agents/me/ws"
		mcpURL := "http://backend:8080/mcp/sse"
		if company := s.getCompanyByAgent(d.AgentID); company != nil {
			if company.AgentWSUrl != "" {
				wsURL = company.AgentWSUrl
			}
			// 内部通信使用 PrivateURL
			if company.MCPPrivateURL != "" {
				mcpURL = company.MCPPrivateURL
			} else if company.MCPPublicURL != "" {
				mcpURL = company.MCPPublicURL
			}
		}
		args = append(args,
			"-e", "AGENT_API_KEY="+apiKey,
			"-e", "ANTHROPIC_API_KEY="+apiKey,
			"-e", "ANTHROPIC_BASE_URL=http://backend:8080",
			"-e", "ANTHROPIC_MODEL="+model,
			"-e", "ANTHROPIC_SMALL_FAST_MODEL="+model,
			"-e", "LINKCLAW_WS_URL="+wsURL,
			"-e", "LINKCLAW_MCP_URL="+mcpURL,
		)
	}

	args = append(args, image)

	out, err := exec.Command("docker", args...).CombinedOutput()
	output := strings.TrimSpace(string(out))
	if err != nil {
		log.Printf("docker run %s 失败: %v — %s", d.ContainerName, err, output)
		return fmt.Errorf("docker run: %s", output)
	}
	log.Printf("已启动容器 %s (%s)", d.ContainerName, output[:12])
	return nil
}

func (s *DeploymentService) GetByAgentID(ctx context.Context, agentID string) (*domain.AgentDeployment, error) {
	return s.deployRepo.GetByAgentID(ctx, agentID)
}

// UpdateStatus 更新部署状态（供 MCP 工具或内部调用）
func (s *DeploymentService) UpdateStatus(ctx context.Context, deployID string, status domain.DeployStatus, errMsg string) error {
	return s.deployRepo.UpdateStatus(ctx, deployID, status, errMsg)
}

// Stop 标记部署为停止状态（实际停止操作由 HR Agent 通过 MCP 工具执行）
func (s *DeploymentService) Stop(ctx context.Context, agentID string) error {
	d, err := s.deployRepo.GetByAgentID(ctx, agentID)
	if err != nil || d == nil {
		return fmt.Errorf("deployment not found")
	}
	return s.deployRepo.UpdateStatus(ctx, d.ID, domain.DeployStatusStopped, "")
}

// Rebuild 重建部署：重置 API Key → 删旧部署 → 创建新部署
func (s *DeploymentService) Rebuild(ctx context.Context, agentID string) (*domain.AgentDeployment, string, error) {
	old, err := s.deployRepo.GetByAgentID(ctx, agentID)
	if err != nil || old == nil {
		return nil, "", fmt.Errorf("no existing deployment for agent %s", agentID)
	}

	agent, err := s.agentRepo.GetByID(ctx, agentID)
	if err != nil || agent == nil {
		return nil, "", fmt.Errorf("agent not found: %s", agentID)
	}

	// 生成新 API Key
	newKey, err := generateAPIKey()
	if err != nil {
		return nil, "", fmt.Errorf("generate api key: %w", err)
	}
	hash := sha256.Sum256([]byte(newKey))
	hashStr := hex.EncodeToString(hash[:])
	if err := s.agentRepo.UpdateAPIKey(ctx, agentID, hashStr, newKey[:10]); err != nil {
		return nil, "", fmt.Errorf("update api key: %w", err)
	}

	// 删除旧部署记录
	if err := s.deployRepo.Delete(ctx, old.ID); err != nil {
		return nil, "", fmt.Errorf("delete old deployment: %w", err)
	}

	// 用新 key 创建新部署
	d, err := s.Deploy(ctx, DeployInput{
		AgentID:     agentID,
		DeployType:  old.DeployType,
		AgentImage:  old.AgentImage,
		APIKey:      newKey,
		Model:       agent.Model,
		SSHHost:     old.SSHHost,
		SSHPort:     old.SSHPort,
		SSHUser:     old.SSHUser,
		SSHPassword: old.SSHPassword,
		SSHKey:      old.SSHKey,
	})
	if err != nil {
		return nil, "", err
	}
	return d, newKey, nil
}
