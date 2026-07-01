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
	return orms.IFind[Rate](tx, "base_code = ?", baseCode)
}

func FindRateByBaseCodeAndCode(tx *gorm.DB, baseCode, code string) (Rate, error) {
	return orms.First[Rate](tx, "base_code = ? and code = ?", baseCode, code)
}
