package domain

import "time"

type DeployType string
type AgentImage string
type DeployStatus string

const (
	DeployTypeLocalDocker DeployType = "local_docker"
	DeployTypeSSHDocker   DeployType = "ssh_docker"
	DeployTypeSSHNative   DeployType = "ssh_native" // 直接在 Linux 上部署（无 Docker）
)

const (
	AgentImageNanoclaw AgentImage = "nanoclaw"
	AgentImageOpenclaw AgentImage = "openclaw"
)

const (
	DeployStatusPending  DeployStatus = "pending"
	DeployStatusRunning  DeployStatus = "running"
	DeployStatusStopped  DeployStatus = "stopped"
	DeployStatusFailed   DeployStatus = "failed"
)

// AgentImageMap nanoclaw/openclaw 对应的完整镜像名
var AgentImageMap = map[AgentImage]string{
	AgentImageNanoclaw: "nanoclaw:latest",
	AgentImageOpenclaw: "ghcr.io/qwibitai/openclaw:latest",
}

type AgentDeployment struct {
	ID            string       `json:"id"            gorm:"primaryKey"`
	AgentID       string       `json:"agentId"       gorm:"column:agent_id"`
	DeployType    DeployType   `json:"deployType"    gorm:"column:deploy_type"`
	AgentImage    AgentImage   `json:"agentImage"    gorm:"column:agent_image"`
	ContainerName string       `json:"containerName" gorm:"column:container_name"`
	SSHHost       string       `json:"sshHost"       gorm:"column:ssh_host"`
	SSHPort       int          `json:"sshPort"       gorm:"column:ssh_port"`
	SSHUser       string       `json:"sshUser"       gorm:"column:ssh_user"`
	SSHPassword   string       `json:"-"             gorm:"column:ssh_password"` // 不输出到 JSON
	SSHKey        string       `json:"-"             gorm:"column:ssh_key"`
	Status        DeployStatus `json:"status"        gorm:"column:status"`
	ErrorMsg      string       `json:"errorMsg,omitempty" gorm:"column:error_msg"`
	CreatedAt     time.Time    `json:"createdAt"     gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time    `json:"updatedAt"     gorm:"column:updated_at;autoUpdateTime"`
}
