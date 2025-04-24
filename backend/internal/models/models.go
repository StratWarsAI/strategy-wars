// internal/models/models.go
package models

import (
	"time"
)

// Token represents a Pump.fun token
type Token struct {
	ID                     int64     `json:"-"`
	MintAddress            string    `json:"mint"`
	CreatorAddress         string    `json:"creator"`
	Name                   string    `json:"name"`
	Symbol                 string    `json:"symbol"`
	ImageUrl               string    `json:"image_url"`
	TwitterUrl             string    `json:"twitter_url"`
	WebsiteUrl             string    `json:"website_url"`
	TelegramUrl            string    `json:"telegram_url"`
	MetadataUrl            string    `json:"metadata_url"`
	CreatedTimestamp       int64     `json:"created_timestamp"`
	MarketCap              float64   `json:"market_cap"`
	UsdMarketCap           float64   `json:"usd_market_cap"`
	Completed              bool      `json:"completed"`
	KingOfTheHillTimeStamp int64     `json:"king_of_the_hill_timestamp"`
	CreatedAt              time.Time `json:"-"`
}

// Trade represents a Pump.fun trade
type Trade struct {
	ID          int64   `json:"-"`
	TokenID     int64   `json:"-"`
	MintAddress string  `json:"mint"`
	Signature   string  `json:"signature"`
	SolAmount   float64 `json:"sol_amount"`
	TokenAmount float64 `json:"token_amount"`
	IsBuy       bool    `json:"is_buy"`
	UserAddress string  `json:"user"`
	Timestamp   int64   `json:"timestamp"`
}
