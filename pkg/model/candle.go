package model

import (
	"fmt"
	"time"
)

// Candle is a single OHLCV bar for a given instrument and timeframe.
type Candle struct {
	Instrument string
	Timeframe  Timeframe
	Timestamp  time.Time
	Open       float64
	High       float64
	Low        float64
	Close      float64
	Volume     float64
}

// NewCandle constructs a Candle and validates its fields.
func NewCandle(instrument string, tf Timeframe, ts time.Time, open, high, low, close_, volume float64) (Candle, error) {
	c := Candle{
		Instrument: instrument,
		Timeframe:  tf,
		Timestamp:  ts,
		Open:       open,
		High:       high,
		Low:        low,
		Close:      close_,
		Volume:     volume,
	}
	if err := c.Validate(); err != nil {
		return Candle{}, err
	}
	return c, nil
}

// Validate checks that the candle fields are self-consistent and non-negative.
func (c Candle) Validate() error {
	if c.Instrument == "" {
		return fmt.Errorf("candle: instrument must not be empty")
	}
	if c.Timestamp.IsZero() {
		return fmt.Errorf("candle: timestamp must not be zero")
	}
	if c.Open <= 0 || c.High <= 0 || c.Low <= 0 || c.Close <= 0 {
		return fmt.Errorf("candle: OHLC values must be positive (got O=%.4f H=%.4f L=%.4f C=%.4f)",
			c.Open, c.High, c.Low, c.Close)
	}
	if c.Volume < 0 {
		return fmt.Errorf("candle: volume must not be negative (got %.4f)", c.Volume)
	}
	if c.High < c.Low {
		return fmt.Errorf("candle: high (%.4f) must be >= low (%.4f)", c.High, c.Low)
	}
	if c.Open > c.High || c.Open < c.Low {
		return fmt.Errorf("candle: open (%.4f) must be within [low=%.4f, high=%.4f]", c.Open, c.Low, c.High)
	}
	if c.Close > c.High || c.Close < c.Low {
		return fmt.Errorf("candle: close (%.4f) must be within [low=%.4f, high=%.4f]", c.Close, c.Low, c.High)
	}
	return nil
}
