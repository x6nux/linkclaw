package service

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type KnowledgeService struct {
	knowledgeRepo repository.KnowledgeRepo
}

func NewKnowledgeService(knowledgeRepo repository.KnowledgeRepo) *KnowledgeService {
	return &KnowledgeService{knowledgeRepo: knowledgeRepo}
}

func (s *KnowledgeService) Search(ctx context.Context, companyID, query string, limit int) ([]*domain.KnowledgeDoc, error) {
	return s.knowledgeRepo.Search(ctx, companyID, query, limit)
}

func (s *KnowledgeService) GetByID(ctx context.Context, id string) (*domain.KnowledgeDoc, error) {
	return s.knowledgeRepo.GetByID(ctx, id)
}

func (s *KnowledgeService) List(ctx context.Context, companyID string, limit, offset int) ([]*domain.KnowledgeDoc, int, error) {
	return s.knowledgeRepo.List(ctx, companyID, limit, offset)
}

type WriteDocInput struct {
	DocID     string // 有则更新，无则创建
	CompanyID string
	AuthorID  string
	Title     string
	Content   string
	Tags      string // 逗号分隔
}

func (s *KnowledgeService) Write(ctx context.Context, in WriteDocInput) (*domain.KnowledgeDoc, error) {
	tags := []string{}
	for _, t := range strings.Split(in.Tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	if in.DocID != "" {
		doc, err := s.knowledgeRepo.GetByID(ctx, in.DocID)
		if err != nil {
			return nil, err
		}
		if doc != nil {
			doc.Title = in.Title
			doc.Content = in.Content
			doc.Tags = tags
			if err = s.knowledgeRepo.Update(ctx, doc); err != nil {
				return nil, err
			}
			return doc, nil
		}
	}

	var authorID *string
	if in.AuthorID != "" {
		authorID = &in.AuthorID
	}
	doc := &domain.KnowledgeDoc{
		ID:        uuid.New().String(),
		CompanyID: in.CompanyID,
		Title:     in.Title,
		Content:   in.Content,
		Tags:      tags,
		AuthorID:  authorID,
	}
	if err := s.knowledgeRepo.Create(ctx, doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (s *KnowledgeService) Delete(ctx context.Context, id string) error {
	return s.knowledgeRepo.Delete(ctx, id)
}
