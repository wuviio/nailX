package worker

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/nailx/backend/internal/domain"
	"github.com/nailx/backend/internal/repository"
)

// RoyaltyDistributionMessage は SQS メッセージのペイロード
type RoyaltyDistributionMessage struct {
	PaymentID    string `json:"payment_id"`
	BookingID    string `json:"booking_id"`
	DesignIPID   string `json:"design_ip_id"`
	TotalRoyalty int    `json:"total_royalty_yen"`
}

// RoyaltyWorker は SQS キューからロイヤリティ分配ジョブを処理する
type RoyaltyWorker struct {
	sqsClient   *sqs.Client
	queueURL    string
	paymentRepo repository.PaymentRepository
	designRepo  repository.DesignRepository
	logger      *slog.Logger
}

func NewRoyaltyWorker(
	sqsClient *sqs.Client,
	queueURL string,
	paymentRepo repository.PaymentRepository,
	designRepo repository.DesignRepository,
	logger *slog.Logger,
) *RoyaltyWorker {
	return &RoyaltyWorker{
		sqsClient:   sqsClient,
		queueURL:    queueURL,
		paymentRepo: paymentRepo,
		designRepo:  designRepo,
		logger:      logger,
	}
}

func (w *RoyaltyWorker) Run(ctx context.Context) {
	w.logger.Info("royalty worker started")
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("royalty worker stopped")
			return
		default:
			w.pollOnce(ctx)
		}
	}
}

func (w *RoyaltyWorker) pollOnce(ctx context.Context) {
	out, err := w.sqsClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            &w.queueURL,
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
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

func (w *RoyaltyWorker) handle(ctx context.Context, msg types.Message) {
	var payload RoyaltyDistributionMessage
	if err := json.Unmarshal([]byte(*msg.Body), &payload); err != nil {
		w.logger.Error("failed to unmarshal royalty message", slog.String("err", err.Error()))
		w.deleteMessage(ctx, msg)
		return
	}

	// フォーク系譜を取得してロイヤリティ分配を計算
	chain, err := w.designRepo.GetRoyaltyChain(ctx, payload.DesignIPID)
	if err != nil {
		w.logger.Error("failed to get royalty chain",
			slog.String("design_id", payload.DesignIPID),
			slog.String("err", err.Error()),
		)
		return
	}

	distributions := domain.CalculateRoyalty(payload.TotalRoyalty, chain)

	if err := w.paymentRepo.CreateRoyaltyDistributions(ctx, payload.PaymentID, distributions); err != nil {
		w.logger.Error("failed to create royalty distributions",
			slog.String("payment_id", payload.PaymentID),
			slog.String("err", err.Error()),
		)
		return
	}

	w.logger.Info("royalty distribution completed",
		slog.String("payment_id", payload.PaymentID),
		slog.String("design_id", payload.DesignIPID),
		slog.Int("distributions", len(distributions)),
	)

	w.deleteMessage(ctx, msg)
}

func (w *RoyaltyWorker) deleteMessage(ctx context.Context, msg types.Message) {
	if _, err := w.sqsClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      &w.queueURL,
		ReceiptHandle: msg.ReceiptHandle,
	}); err != nil {
		w.logger.Error("failed to delete sqs message", slog.String("err", err.Error()))
	}
}
