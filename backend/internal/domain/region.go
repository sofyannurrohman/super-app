package domain

import "time"

type IndoProvince struct {
	ID        int64      `gorm:"primaryKey;column:id" json:"id"`
	Code      string     `gorm:"column:code;index" json:"code"`
	Name      string     `gorm:"column:name;index" json:"name"`
	Meta      string     `gorm:"column:meta" json:"meta"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (IndoProvince) TableName() string { return "indonesia_provinces" }

type IndoCity struct {
	ID           int64      `gorm:"primaryKey;column:id" json:"id"`
	Code         string     `gorm:"column:code;index" json:"code"`
	ProvinceCode string     `gorm:"column:province_code;index" json:"province_code"`
	Name         string     `gorm:"column:name;index" json:"name"`
	Meta         string     `gorm:"column:meta" json:"meta"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (IndoCity) TableName() string { return "indonesia_cities" }

type IndoDistrict struct {
	ID        int64      `gorm:"primaryKey;column:id" json:"id"`
	Code      string     `gorm:"column:code;index" json:"code"`
	CityCode  string     `gorm:"column:city_code;index" json:"city_code"`
	Name      string     `gorm:"column:name;index" json:"name"`
	Meta      string     `gorm:"column:meta" json:"meta"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (IndoDistrict) TableName() string { return "indonesia_districts" }

type IndoVillage struct {
	ID           int64      `gorm:"primaryKey;column:id" json:"id"`
	Code         string     `gorm:"column:code;index" json:"code"`
	DistrictCode string     `gorm:"column:district_code;index" json:"district_code"`
	Name         string     `gorm:"column:name;index" json:"name"`
	Meta         string     `gorm:"column:meta" json:"meta"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (IndoVillage) TableName() string { return "indonesia_villages" }
