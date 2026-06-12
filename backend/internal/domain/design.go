package domain

import (
	"encoding/json"
	"time"
)

type DesignStatus string

const (
	DesignStatusPending  DesignStatus = "pending"
	DesignStatusActive   DesignStatus = "active"
	DesignStatusRejected DesignStatus = "rejected"
	DesignStatusArchived DesignStatus = "archived"
)

type SimilarityResult string

const (
	SimilarityResultIndependent SimilarityResult = "independent"
	SimilarityResultFork        SimilarityResult = "fork"
	SimilarityResultRejected    SimilarityResult = "rejected"
)

// 類似度閾値（Admin設定で上書き可能）
const (
	DefaultSimilarityForkThreshold   = 0.75
	DefaultSimilarityRejectThreshold = 0.90
)

type DesignIP struct {
	ID               string           `json:"id"`
	CreatorID        string           `json:"creator_id"`
	ParentIPID       *string          `json:"parent_ip_id,omitempty"`
	Title            string           `json:"title"`
	Description      *string          `json:"description,omitempty"`
	PreviewImageURL  string           `json:"preview_image_url"`
	DesignData       json.RawMessage  `json:"design_data"`
	SimilarityHash   *string          `json:"-"`
	ForkDepth        int              `json:"fork_depth"`
	Status           DesignStatus     `json:"status"`
	IsPublic         bool             `json:"is_public"`
	GenderTag        string           `json:"gender_tag"`
	StyleTags        []string         `json:"style_tags"`
	UsageCount       int              `json:"usage_count"`
	TotalRoyaltyYen  int              `json:"total_royalty_yen"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

type DesignRoyaltyNode struct {
	ID           string  `json:"id"`
	DesignIPID   string  `json:"design_ip_id"`
	UserID       string  `json:"user_id"`
	SharePercent float64 `json:"share_percent"`
	DepthLevel   int     `json:"depth_level"`
}

type Material struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Category      string          `json:"category"`
	TextureType   *string         `json:"texture_type,omitempty"`
	ThumbnailURL  string          `json:"thumbnail_url"`
	Model3DURL    *string         `json:"model_3d_url,omitempty"`
	TextureParams json.RawMessage `json:"texture_params"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
}

// RoyaltyDistributionItem はロイヤリティ分配の1件分
type RoyaltyDistributionItem struct {
	DesignIPID string
	UserID     string
	AmountYen  int
	Percent    float64
	DepthLevel int
}

// CalculateRoyalty はロイヤリティ分配を計算する（1円未満切り捨て）
// 深さ3段打ち切り: 直接70% / 親21% / 祖父9%
func CalculateRoyalty(totalIPFeeYen int, nodes []DesignRoyaltyNode) []RoyaltyDistributionItem {
	result := make([]RoyaltyDistributionItem, 0, len(nodes))
	for _, node := range nodes {
		// floor: 1円未満切り捨て
		amount := int(float64(totalIPFeeYen) * node.SharePercent / 100.0)
		result = append(result, RoyaltyDistributionItem{
			DesignIPID: node.DesignIPID,
			UserID:     node.UserID,
			AmountYen:  amount,
			Percent:    node.SharePercent,
			DepthLevel: node.DepthLevel,
		})
	}
	return result
}
