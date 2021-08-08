package inago

import "fmt"

// Config is a parameter store of Inago Bot
type Config struct {
	lot          float64
	volumeTriger float64
	scope        int64 // second
	settleTerm   int64
	reverse      bool    // buyTriger検知でshort/sellTriger検知でlong
	settleRange  float64 // 決済するさやの幅
	priceRatio   float64 // 価格乖離のボーナスレート

	// config2
	avgVolumePeriod      int     //平均volatilityの算出足数
	againstAvgVolumeRate float64 // 平均volatilityの何倍でエントリーか
	minimumVolume        float64 //エントリー時の最低Volume
	guardOverBb3         float64 // bolingerBand(3)よりどれくらい超えないとエントリーしないか
	minimumCandleLength  float64 // ろうそく足の長さがvolatilityの何倍を超えないとエントリーしないか
}

func NewConfig(scope int64, volumeTriger float64, settleTerm int64, settleRange, priceRatio float64, reverse bool) Config {
	return Config{
		scope:        scope,
		volumeTriger: volumeTriger,
		settleTerm:   settleTerm,
		reverse:      reverse,
		settleRange:  settleRange,
		priceRatio:   priceRatio,
	}
}

func NewConfig2(
	lot float64,
	avgVolumePeriod int,
	againstAvgVolumeRate float64,
	minimumVolume float64,
	guardOverBb3 float64,
	minimumCandle float64,
) Config {
	return Config{
		lot:                  lot,
		avgVolumePeriod:      avgVolumePeriod,
		againstAvgVolumeRate: againstAvgVolumeRate,
		guardOverBb3:         guardOverBb3,
		minimumCandleLength:  minimumCandle,
	}
}

func (c *Config) Serialize() string {
	return fmt.Sprintf("%.0f-%d-%d-%.3f-%.0f",
		c.volumeTriger,
		c.scope,
		c.settleTerm,
		c.settleRange,
		c.priceRatio)
}

func (c *Config) Serialize2() string {
	return fmt.Sprintf("%d-%.1f-%.0f-%.3f-%.1f",
		c.avgVolumePeriod,
		c.againstAvgVolumeRate,
		c.minimumVolume,
		c.guardOverBb3,
		c.minimumCandleLength,
	)
}
