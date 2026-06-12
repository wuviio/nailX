package impl

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
)

type arService struct {
	arRepo repository.ARRepository
}

func NewARService(arRepo repository.ARRepository) service.ARService {
	return &arService{arRepo: arRepo}
}

func (s *arService) CreateSession(ctx context.Context, userID string, req service.CreateARSessionRequest) (*domain.ARSession, error) {
	session := &domain.ARSession{
		ID:                   uuid.NewString(),
		UserID:               userID,
		DesignIPID:           req.DesignIPID,
		DetectedNailLengthMM: req.DetectedNailLengthMM,
		HasExistingGel:       req.HasExistingGel,
		DetectedNailShape:    req.DetectedNailShape,
		HandSnapshotURL:      req.HandSnapshotURL,
		CreatedAt:            time.Now(),
	}

	// 爪長さ・形状から施術時間・ジェル量を推定
	if req.DetectedNailLengthMM != nil {
		min, ml := estimateTreatment(*req.DetectedNailLengthMM, req.HasExistingGel)
		session.EstimatedTreatmentMin = &min
		session.EstimatedGelAmountML = &ml
	}

	if err := s.arRepo.Create(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *arService) GetSession(ctx context.Context, userID, sessionID string) (*domain.ARSession, error) {
	session, err := s.arRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil || session.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return session, nil
}

// estimateTreatment は爪長さ（mm）と既存ジェル有無から施術時間（分）とジェル使用量（ml）を推定する。
// 基準値: 爪長さ 8mm → 90 分、既存ジェルあり +20 分
func estimateTreatment(lengthMM float64, hasGel *bool) (minMin int, mlML float64) {
	baseMins := 60 + int(math.Round(lengthMM*3.75)) // 8mm → 90min
	if hasGel != nil && *hasGel {
		baseMins += 20
	}
	// ジェル量: 爪枚数10枚 × 0.2ml 基準、長さ係数を乗算
	mlML = math.Round(10*0.2*(0.5+lengthMM/16)*10) / 10
	return baseMins, mlML
}
