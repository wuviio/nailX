package impl

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/google/uuid"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
	workerPkg "github.com/nailx/backend/internal/worker"
)

type designService struct {
	designRepo repository.DesignRepository
	sqsClient  *sqs.Client
	sqsQueue   string
}

func NewDesignService(designRepo repository.DesignRepository, sqsClient *sqs.Client, sqsQueue string) service.DesignService {
	return &designService{designRepo: designRepo, sqsClient: sqsClient, sqsQueue: sqsQueue}
}

func (s *designService) RegisterDesign(ctx context.Context, userID string, req service.RegisterDesignRequest) (string, string, error) {
	forkDepth := 0
	if req.ParentIPID != nil {
		parent, err := s.designRepo.FindByID(ctx, *req.ParentIPID)
		if err != nil {
			return "", "", err
		}
		if parent.Status != domain.DesignStatusActive {
			return "", "", domain.ErrInvalidRequest
		}
		forkDepth = parent.ForkDepth + 1
		if forkDepth > 3 {
			return "", "", domain.ErrInvalidRequest // フォーク深さ上限
		}
	}

	designID := uuid.NewString()
	design := &domain.DesignIP{
		ID:              designID,
		CreatorID:       userID,
		ParentIPID:      req.ParentIPID,
		Title:           req.Title,
		Description:     req.Description,
		PreviewImageURL: req.PreviewImageURL,
		DesignData:      req.DesignData,
		ForkDepth:       forkDepth,
		Status:          domain.DesignStatusPending,
		IsPublic:        req.IsPublic,
		GenderTag:       req.GenderTag,
		StyleTags:       req.StyleTags,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := s.designRepo.Create(ctx, design); err != nil {
		return "", "", err
	}

	// SQS に類似度チェックジョブを投入
	// ベクトルは SageMaker embedding 結果を想定（現フェーズはゼロベクトルで登録し後でワーカーが更新）
	jobID := uuid.NewString()
	msg, _ := json.Marshal(workerPkg.SimilarityCheckMessage{
		DesignIPID: designID,
		Vector:     []float32{},
	})
	if _, err := s.sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.sqsQueue),
		MessageBody: aws.String(string(msg)),
		MessageGroupId: aws.String(designID), // FIFO キューの場合
	}); err != nil {
		// SQS 投入失敗は致命的ではない（ワーカーは別途リトライ）
		_ = err
	}

	return designID, jobID, nil
}

func (s *designService) GetSimilarityCheckStatus(ctx context.Context, jobID string) (*service.SimilarityCheckResult, error) {
	// TODO: SQS/DynamoDB でジョブステータスを管理（Phase 2）
	return &service.SimilarityCheckResult{Status: "pending"}, nil
}

func (s *designService) ListFeed(ctx context.Context, filter service.DesignFeedFilter) ([]*domain.DesignIP, string, error) {
	return s.designRepo.ListFeed(ctx, repository.DesignFeedFilter{
		GenderTag: filter.GenderTag,
		StyleTags: filter.StyleTags,
		Status:    filter.Status,
		Sort:      filter.Sort,
		Cursor:    filter.Cursor,
		Limit:     filter.Limit,
	})
}

func (s *designService) GetDesign(ctx context.Context, id string) (*domain.DesignIP, []*domain.DesignRoyaltyNode, error) {
	design, err := s.designRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	nodes, err := s.designRepo.ListRoyaltyNodes(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	return design, nodes, nil
}

func (s *designService) ListByCreator(ctx context.Context, creatorID string, cursor string, limit int) ([]*domain.DesignIP, string, error) {
	return s.designRepo.ListByCreator(ctx, creatorID, cursor, limit)
}

func (s *designService) UpdateDesign(ctx context.Context, userID string, callerRole domain.UserRole, designID string, req service.UpdateDesignRequest) (*domain.DesignIP, error) {
	design, err := s.designRepo.FindByID(ctx, designID)
	if err != nil {
		return nil, err
	}
	if design.CreatorID != userID && callerRole != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	if req.Title != nil {
		design.Title = *req.Title
	}
	if req.Description != nil {
		design.Description = req.Description
	}
	if req.IsPublic != nil {
		design.IsPublic = *req.IsPublic
		if *req.IsPublic {
			design.Status = domain.DesignStatusActive
		} else {
			design.Status = domain.DesignStatusArchived
		}
	}
	design.UpdatedAt = time.Now()
	if err := s.designRepo.Update(ctx, design); err != nil {
		return nil, err
	}
	return design, nil
}
