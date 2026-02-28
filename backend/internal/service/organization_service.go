package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/event"
	"github.com/linkclaw/backend/internal/repository"
)

type OrganizationService struct {
	deptRepo     repository.DepartmentRepo
	agentRepo    repository.AgentRepo
	approvalRepo repository.ApprovalRepo
}

func NewOrganizationService(
	deptRepo repository.DepartmentRepo,
	agentRepo repository.AgentRepo,
	approvalRepo repository.ApprovalRepo,
) *OrganizationService {
	return &OrganizationService{
		deptRepo:     deptRepo,
		agentRepo:    agentRepo,
		approvalRepo: approvalRepo,
	}
}

type OrgChart struct {
	Departments []OrgDept `json:"departments"`
}

type OrgDept struct {
	Department *domain.Department `json:"department"`
	Members    []*domain.Agent    `json:"members"`
}

func (s *OrganizationService) CreateDepartment(
	ctx context.Context,
	companyID, name, slug, desc string,
	directorAgentID, parentDeptID *string,
) (*domain.Department, error) {
	d := &domain.Department{
		ID:              uuid.New().String(),
		CompanyID:       companyID,
		Name:            name,
		Slug:            slug,
		Description:     desc,
		DirectorAgentID: directorAgentID,
		ParentDeptID:    parentDeptID,
		CreatedAt:       time.Now(),
	}
	if err := s.deptRepo.Create(ctx, d); err != nil {
		return nil, fmt.Errorf("create department: %w", err)
	}
	return d, nil
}

func (s *OrganizationService) UpdateDepartment(
	ctx context.Context,
	id, name, slug, desc string,
	directorAgentID, parentDeptID *string,
) error {
	d, err := s.deptRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get department: %w", err)
	}
	if d == nil {
		return fmt.Errorf("department not found")
	}

	d.Name = name
	d.Slug = slug
	d.Description = desc
	d.DirectorAgentID = directorAgentID
	d.ParentDeptID = parentDeptID

	if err := s.deptRepo.Update(ctx, d); err != nil {
		return fmt.Errorf("update department: %w", err)
	}
	return nil
}

func (s *OrganizationService) DeleteDepartment(ctx context.Context, id string, companyID string) error {
	members, err := s.agentRepo.ListByDepartment(ctx, companyID, id)
	if err != nil {
		return fmt.Errorf("list department members: %w", err)
	}
	for _, m := range members {
		if err := s.agentRepo.UpdateDepartment(ctx, m.ID, nil); err != nil {
			return fmt.Errorf("clear agent department: %w", err)
		}
	}
	return s.deptRepo.Delete(ctx, id)
}

func (s *OrganizationService) ListDepartments(ctx context.Context, companyID string) ([]*domain.Department, error) {
	return s.deptRepo.List(ctx, companyID)
}

func (s *OrganizationService) GetDepartment(ctx context.Context, id string) (*domain.Department, error) {
	return s.deptRepo.GetByID(ctx, id)
}

func (s *OrganizationService) AssignAgentToDepartment(ctx context.Context, agentID, departmentID string) error {
	if err := s.agentRepo.UpdateDepartment(ctx, agentID, &departmentID); err != nil {
		return fmt.Errorf("assign agent to department: %w", err)
	}
	return nil
}

func (s *OrganizationService) SetManager(ctx context.Context, agentID string, managerID *string) error {
	return s.agentRepo.UpdateManager(ctx, agentID, managerID)
}

func (s *OrganizationService) BuildOrgChart(ctx context.Context, companyID string) (*OrgChart, error) {
	departments, err := s.deptRepo.List(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("list departments: %w", err)
	}

	chart := &OrgChart{
		Departments: make([]OrgDept, 0, len(departments)),
	}

	for _, d := range departments {
		members, err := s.agentRepo.ListByDepartment(ctx, companyID, d.ID)
		if err != nil {
			return nil, fmt.Errorf("list department members: %w", err)
		}
		chart.Departments = append(chart.Departments, OrgDept{
			Department: d,
			Members:    members,
		})
	}

	return chart, nil
}

func (s *OrganizationService) CreateApproval(
	ctx context.Context,
	companyID, requesterID string,
	reqType domain.ApprovalRequestType,
	payload json.RawMessage,
	reason string,
) (*domain.ApprovalRequest, error) {
	approverID, err := s.resolveApproverID(ctx, companyID, requesterID)
	if err != nil {
		return nil, err
	}

	if len(payload) == 0 {
		payload = json.RawMessage("{}")
	}

	req := &domain.ApprovalRequest{
		ID:          uuid.New().String(),
		CompanyID:   companyID,
		RequesterID: requesterID,
		ApproverID:  approverID,
		RequestType: reqType,
		Status:      domain.ApprovalPending,
		Payload:     payload,
		Reason:      reason,
		CreatedAt:   time.Now(),
	}
	if err := s.approvalRepo.Create(ctx, req); err != nil {
		return nil, fmt.Errorf("create approval: %w", err)
	}
	return req, nil
}

func (s *OrganizationService) ApproveRequest(ctx context.Context, id, decisionReason string) error {
	req, err := s.approvalRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get approval request: %w", err)
	}
	if req == nil {
		return fmt.Errorf("approval request not found")
	}

	now := time.Now()
	if err := s.approvalRepo.UpdateStatus(ctx, id, domain.ApprovalApproved, decisionReason, &now); err != nil {
		return fmt.Errorf("approve request: %w", err)
	}

	if req.RequestType == domain.ApprovalHire {
		event.Global.Publish(event.NewEvent(event.ApprovalApproved, event.ApprovalApprovedPayload{
			RequestID:   req.ID,
			CompanyID:   req.CompanyID,
			RequestType: string(req.RequestType),
			RequesterID: req.RequesterID,
		}))
	}
	return nil
}

func (s *OrganizationService) RejectRequest(ctx context.Context, id, decisionReason string) error {
	req, err := s.approvalRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get approval request: %w", err)
	}
	if req == nil {
		return fmt.Errorf("approval request not found")
	}

	now := time.Now()
	return s.approvalRepo.UpdateStatus(ctx, id, domain.ApprovalRejected, decisionReason, &now)
}

func (s *OrganizationService) ListApprovals(ctx context.Context, q repository.ApprovalQuery) ([]*domain.ApprovalRequest, int, error) {
	return s.approvalRepo.List(ctx, q)
}

func (s *OrganizationService) resolveApproverID(ctx context.Context, companyID, requesterID string) (*string, error) {
	requester, err := s.agentRepo.GetByID(ctx, requesterID)
	if err != nil {
		return nil, fmt.Errorf("resolve requester: %w", err)
	}
	if requester == nil {
		return nil, fmt.Errorf("requester not found")
	}

	dept := domain.DepartmentOf(requester.Position)
	directorPos, ok := domain.DepartmentDirectors[dept]
	if !ok {
		return nil, nil
	}

	agents, err := s.agentRepo.GetByCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("resolve approver: %w", err)
	}
	for _, a := range agents {
		if a.Position == directorPos && a.ID != requesterID {
			id := a.ID
			return &id, nil
		}
	}
	return nil, nil
}
