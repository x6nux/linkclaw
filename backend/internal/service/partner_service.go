package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"

	"github.com/linkclaw/backend/internal/domain"
	"github.com/linkclaw/backend/internal/repository"
)

type PartnerService struct {
	partnerKeyRepo repository.PartnerAPIKeyRepo
	companyRepo    repository.CompanyRepo
}

func NewPartnerService(partnerKeyRepo repository.PartnerAPIKeyRepo, companyRepo repository.CompanyRepo) *PartnerService {
	return &PartnerService{
		partnerKeyRepo: partnerKeyRepo,
		companyRepo:    companyRepo,
	}
}

// GenerateKey 为公司配对生成新的 API Key
// 返回原始 key（仅显示一次）和密钥记录
func (s *PartnerService) GenerateKey(ctx context.Context, companyID, partnerSlug, name string) (rawKey string, err error) {
	// 验证配对公司存在
	partner, err := s.companyRepo.GetBySlug(ctx, partnerSlug)
	if err != nil {
		return "", fmt.Errorf("get partner company: %w", err)
	}
	if partner == nil {
		return "", fmt.Errorf("partner company %q not found", partnerSlug)
	}

	// 生成随机密钥（32 字节 = 256 位）
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("generate random key: %w", err)
	}
	rawKey = hex.EncodeToString(keyBytes)

	// 计算 hash
	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	// 生成前缀（用于 UI 显示）
	keyPrefix := "pk_" + rawKey[:8]

	k := &domain.PartnerApiKey{
		ID:          uuid.New().String(),
		CompanyID:   companyID,
		PartnerSlug: partnerSlug,
		PartnerID:   &partner.ID,
		Name:        &name,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		IsActive:    true,
	}

	if err := s.partnerKeyRepo.Create(ctx, k); err != nil {
		return "", fmt.Errorf("create partner key: %w", err)
	}

	return rawKey, nil
}

// ValidateKey 验证配对密钥
func (s *PartnerService) ValidateKey(ctx context.Context, companyID, partnerSlug, key string) error {
	// 计算输入 key 的 hash
	hash := sha256.Sum256([]byte(key))
	keyHash := hex.EncodeToString(hash[:])

	k, err := s.partnerKeyRepo.GetByKeyHash(ctx, companyID, keyHash)
	if err != nil {
		return fmt.Errorf("get partner key: %w", err)
	}
	if k == nil {
		return fmt.Errorf("invalid partner key")
	}

	// 验证配对关系
	if k.PartnerSlug != partnerSlug {
		return fmt.Errorf("partner key mismatch")
	}

	// 更新最后使用时间
	if err := s.partnerKeyRepo.UpdateLastUsed(ctx, k.ID); err != nil {
		// 非致命错误，只记录日志
	}

	return nil
}

// RevokeKey 撤销配对密钥
func (s *PartnerService) RevokeKey(ctx context.Context, companyID, partnerSlug string) error {
	k, err := s.partnerKeyRepo.GetByCompanyAndPartner(ctx, companyID, partnerSlug)
	if err != nil {
		return fmt.Errorf("get partner key: %w", err)
	}
	if k == nil {
		return nil // 已经不存在
	}

	if err := s.partnerKeyRepo.Deactivate(ctx, k.ID); err != nil {
		return fmt.Errorf("deactivate partner key: %w", err)
	}

	return nil
}

// GetKey 获取配对密钥信息（不包含完整 hash）
func (s *PartnerService) GetKey(ctx context.Context, companyID, partnerSlug string) (*domain.PartnerApiKey, error) {
	k, err := s.partnerKeyRepo.GetByCompanyAndPartner(ctx, companyID, partnerSlug)
	if err != nil {
		return nil, fmt.Errorf("get partner key: %w", err)
	}
	return k, nil
}

// ListKeys 列出公司所有的配对密钥
func (s *PartnerService) ListKeys(ctx context.Context, companyID string) ([]*domain.PartnerApiKey, error) {
	// 需要通过 raw SQL 查询，因为接口中没有定义此方法
	// 暂时返回空数组，后续可扩展
	_ = companyID
	return []*domain.PartnerApiKey{}, nil
}
