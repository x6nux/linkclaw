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
	AgentID       string       `json:"agent_id"       gorm:"column:agent_id"`
	DeployType    DeployType   `json:"deploy_type"    gorm:"column:deploy_type"`
	AgentImage    AgentImage   `json:"agent_image"    gorm:"column:agent_image"`
	ContainerName string       `json:"container_name" gorm:"column:container_name"`
	SSHHost       string       `json:"ssh_host"       gorm:"column:ssh_host"`
	SSHPort       int          `json:"ssh_port"       gorm:"column:ssh_port"`
	SSHUser       string       `json:"ssh_user"       gorm:"column:ssh_user"`
	SSHPassword   string       `json:"-"             gorm:"column:ssh_password"` // 不输出到 JSON
	SSHKey        string       `json:"-"             gorm:"column:ssh_key"`
	Status        DeployStatus `json:"status"        gorm:"column:status"`
	ErrorMsg      string       `json:"error_msg,omitempty" gorm:"column:error_msg"`
	CreatedAt     time.Time    `json:"created_at"     gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time    `json:"updated_at"     gorm:"column:updated_at;autoUpdateTime"`
}
