package postgres

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// encodeCursor は created_at + id をカーソル文字列にエンコードする
func encodeCursor(t time.Time, id string) string {
	raw := fmt.Sprintf("%d:%s", t.UnixMilli(), id)
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

// decodeCursor はカーソル文字列を created_at (time.Time) と id にデコードする
func decodeCursor(cursor string) (time.Time, string, error) {
	b, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", err
	}
	parts := strings.SplitN(string(b), ":", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor format")
	}
	ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, "", err
	}
	return time.UnixMilli(ms), parts[1], nil
}

// itoa は int を文字列に変換する（fmt.Sprintf の代替）
func itoa(n int) string {
	return strconv.Itoa(n)
}

// vectorToString は []float32 を pgvector が受け付ける文字列表現に変換する
// 例: "[0.1,0.2,0.3]"
func vectorToString(v []float32) string {
	if len(v) == 0 {
		return "[]"
	}
	sb := strings.Builder{}
	sb.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 32))
	}
	sb.WriteByte(']')
	return sb.String()
}
