package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/nailx/backend/internal/domain"
	"github.com/redis/go-redis/v9"
)

// SSEHub は入札リアルタイム通知のハブ。
// EKS マルチPod 環境では Redis Pub/Sub を使ってブロードキャストする。
// 各 Pod が auction:{request_id} チャンネルを Subscribe し、
// 入札書き込み時に Publish することで全 Pod のクライアントへ届く。
type SSEHub struct {
	rdb    *redis.Client
	logger *slog.Logger
	mu     sync.RWMutex
	// ローカル接続マップ: requestID → クライアントチャンネル群
	clients map[string]map[chan *domain.BidEvent]struct{}
}

func NewSSEHub(rdb *redis.Client, logger *slog.Logger) *SSEHub {
	return &SSEHub{
		rdb:     rdb,
		logger:  logger,
		clients: make(map[string]map[chan *domain.BidEvent]struct{}),
	}
}

// Start はアプリ起動時に呼ぶ。Redis の psubscribe で全 auction:* チャンネルを受信する。
func (h *SSEHub) Start(ctx context.Context) {
	pubsub := h.rdb.PSubscribe(ctx, "auction:*")
	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				// チャンネル名から requestID を取り出す
				requestID := msg.Channel[len("auction:"):]
				var event domain.BidEvent
				if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
					h.logger.Error("sse hub: unmarshal error", slog.String("err", err.Error()))
					continue
				}
				h.broadcast(requestID, &event)
			}
		}
	}()
}

// Publish は入札作成時に Redis へイベントを発行する
func (h *SSEHub) Publish(ctx context.Context, requestID string, event *domain.BidEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return h.rdb.Publish(ctx, "auction:"+requestID, payload).Err()
}

func (h *SSEHub) broadcast(requestID string, event *domain.BidEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients[requestID] {
		select {
		case ch <- event:
		default:
			// クライアントのバッファが詰まっている場合は捨てる（切断検知は StreamBids 側）
		}
	}
}

func (h *SSEHub) subscribe(requestID string) chan *domain.BidEvent {
	ch := make(chan *domain.BidEvent, 16)
	h.mu.Lock()
	if _, ok := h.clients[requestID]; !ok {
		h.clients[requestID] = make(map[chan *domain.BidEvent]struct{})
	}
	h.clients[requestID][ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *SSEHub) unsubscribe(requestID string, ch chan *domain.BidEvent) {
	h.mu.Lock()
	delete(h.clients[requestID], ch)
	if len(h.clients[requestID]) == 0 {
		delete(h.clients, requestID)
	}
	h.mu.Unlock()
	close(ch)
}

// StreamBids は SSE エンドポイントのハンドラー
// GET /api/v1/auctions/requests/:request_id/bids/stream
func (h *SSEHub) StreamBids(c echo.Context) error {
	requestID := c.Param("request_id")
	if requestID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing request_id")
	}

	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Flush()

	ch := h.subscribe(requestID)
	defer h.unsubscribe(requestID, ch)

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			// クライアント切断 or サーバーシャットダウン → goroutine をクリーンアップ
			return nil
		case event, ok := <-ch:
			if !ok {
				return nil
			}
			payload, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Response(), "event: new_bid\ndata: %s\n\n", payload)
			c.Response().Flush()
		}
	}
}
