package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/linkclaw/backend/internal/domain"
)

type deploymentRepo struct{ db *gorm.DB }

func NewDeploymentRepo(db *gorm.DB) DeploymentRepo {
	return &deploymentRepo{db: db}
}

func (r *deploymentRepo) Create(ctx context.Context, d *domain.AgentDeployment) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).
		Exec(`INSERT INTO agent_deployments
			(id, agent_id, deploy_type, agent_image, container_name,
			 ssh_host, ssh_port, ssh_user, ssh_password, ssh_key, status)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
			d.ID, d.AgentID, string(d.DeployType), string(d.AgentImage),
			d.ContainerName, d.SSHHost, d.SSHPort, d.SSHUser,
			d.SSHPassword, d.SSHKey, string(d.Status)).Error
}

func (r *deploymentRepo) GetByAgentID(ctx context.Context, agentID string) (*domain.AgentDeployment, error) {
	var d domain.AgentDeployment
	result := r.db.WithContext(ctx).
		Raw(`SELECT * FROM agent_deployments WHERE agent_id = $1 ORDER BY created_at DESC LIMIT 1`, agentID).
		Scan(&d)
	if result.Error != nil {
		return nil, result.Error
	}
	if d.ID == "" {
		return nil, nil
	}
	return &d, nil
}

func (r *deploymentRepo) UpdateStatus(ctx context.Context, id string, status domain.DeployStatus, errMsg string) error {
	result := r.db.WithContext(ctx).
		Exec(`UPDATE agent_deployments SET status=$1, error_msg=$2, updated_at=NOW() WHERE id=$3`,
			string(status), errMsg, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("deployment not found")
	}
	return nil
}

func (r *deploymentRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Exec(`DELETE FROM agent_deployments WHERE id=$1`, id).Error
}
