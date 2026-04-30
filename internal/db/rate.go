package db

import (
	"github.com/natholdallas/natools4go/orms"
	"gorm.io/gorm"
)

type Rate struct {
	orms.IDModel[uint]
	BaseCode string  `gorm:"column:base_code;size:50;comment:Base currency" json:"base_code"` // Base currency
	Code     string  `gorm:"column:code;size:50;comment:Target currency" json:"code"`         // Target currency
	Rate     float64 `gorm:"column:rate;comment:Exchange rate" json:"rate"`                   // Exchange rate
} //	@name	Rate

func ListRateByBaseCode(tx *gorm.DB, baseCode string) []Rate {
	rates := []Rate{}
	tx.Model(&Rate{}).Where("base_code = ?", baseCode).Find(&rates)
	return rates
}

func FindRateByBaseCodeAndCode(tx *gorm.DB, baseCode, code string) (Rate, error) {
	rate := Rate{}
	err := tx.Model(&Rate{}).Where("base_code = ? and code = ?", baseCode, code).First(&rate).Error
	return rate, err
}
