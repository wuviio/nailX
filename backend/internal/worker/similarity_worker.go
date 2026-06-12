package worker

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/nailx/backend/internal/repository"
	"github.com/nailx/backend/internal/service"
)

// SimilarityCheckMessage は SQS メッセージのペイロード
type SimilarityCheckMessage struct {
	DesignIPID string    `json:"design_ip_id"`
	Vector     []float32 `json:"vector"`
}

// SimilarityWorker は SQS キューから設計類似度チェックジョブを処理する
type SimilarityWorker struct {
	sqsClient  *sqs.Client
	queueURL   string
	designRepo repository.DesignRepository
	designSvc  service.DesignService
	logger     *slog.Logger
}

func NewSimilarityWorker(
	sqsClient *sqs.Client,
	queueURL string,
	designRepo repository.DesignRepository,
	designSvc service.DesignService,
	logger *slog.Logger,
) *SimilarityWorker {
	return &SimilarityWorker{
		sqsClient:  sqsClient,
		queueURL:   queueURL,
		designRepo: designRepo,
		designSvc:  designSvc,
		logger:     logger,
	}
}

// Run はワーカーループを開始する（ctx のキャンセルで停止）
func (w *SimilarityWorker) Run(ctx context.Context) {
	w.logger.Info("similarity worker started")
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("similarity worker stopped")
			return
		default:
			w.pollOnce(ctx)
		}
	}
}

func (w *SimilarityWorker) pollOnce(ctx context.Context) {
	out, err := w.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &w.queueURL,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20, // long polling
	})
	if err != nil {
		if ctx.Err() != nil {
			return
		}
		w.logger.Error("sqs receive error", slog.String("err", err.Error()))
		return
	}

	for _, msg := range out.Messages {
		w.handle(ctx, msg)
	}
}

func (w *SimilarityWorker) handle(ctx context.Context, msg types.Message) {
	var payload SimilarityCheckMessage
	if err := json.Unmarshal([]byte(*msg.Body), &payload); err != nil {
		w.logger.Error("failed to unmarshal message", slog.String("err", err.Error()))
		w.deleteMessage(ctx, msg)
		return
	}

	const (
		rejectThreshold = 0.90
		forkThreshold   = 0.75
	)

	similar, err := w.designRepo.FindSimilar(ctx, payload.Vector, forkThreshold, rejectThreshold)
	if err != nil {
		w.logger.Error("similarity search failed", slog.String("design_id", payload.DesignIPID), slog.String("err", err.Error()))
		return // 削除しない（リトライさせる）
	}

	var newStatus string
	if len(similar) > 0 {
		// 最もスコアが高い類似デザインとの関係を判定
		topScore := similar[0].Score
		if topScore >= rejectThreshold {
			newStatus = "rejected"
		} else {
			newStatus = "fork_pending" // フォーク確認待ち
		}
	} else {
		newStatus = "active" // 独立デザインとして承認
	}

	if err := w.designRepo.UpdateStatus(ctx, payload.DesignIPID, newStatus); err != nil {
		w.logger.Error("failed to update design status",
			slog.String("design_id", payload.DesignIPID),
			slog.String("status", newStatus),
			slog.String("err", err.Error()),
		)
		return
	}

	w.logger.Info("similarity check completed",
		slog.String("design_id", payload.DesignIPID),
		slog.String("result", newStatus),
		slog.Int("similar_count", len(similar)),
	)

	w.deleteMessage(ctx, msg)
}

func (w *SimilarityWorker) deleteMessage(ctx context.Context, msg types.Message) {
	if _, err := w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &w.queueURL,
		ReceiptHandle: msg.ReceiptHandle,
	}); err != nil {
		w.logger.Error("failed to delete sqs message", slog.String("err", err.Error()))
	}
}
