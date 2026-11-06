package usecase

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"example.com/hris/backend/internal/domain"
	"gorm.io/gorm"
)

type NotificationUsecase interface {
	CheckAndQueueExpirations() error
}

type notificationUsecase struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewNotificationUsecase(db *gorm.DB, redisClient *redis.Client) NotificationUsecase {
	return &notificationUsecase{
		db:    db,
		redis: redisClient,
	}
}

type ExpiryAlertPayload struct {
	Type         string `json:"type"`
	Category     string `json:"category"`
	ItemName     string `json:"item_name"`
	Detail       string `json:"detail"`
	DaysLeft     int    `json:"days_left"`
	ExpiryDate   string `json:"expiry_date"`
}

func (u *notificationUsecase) CheckAndQueueExpirations() error {
	ctx := context.Background()
	now := time.Now()
	threshold := now.Add(30 * 24 * time.Hour)

	// 1. Check Vehicle Tax Expirations
	var vehicles []domain.Vehicle
	if err := u.db.Where("tax_expiry_date IS NOT NULL AND tax_expiry_date <= ?", threshold).Find(&vehicles).Error; err != nil {
		log.Printf("[Cron] Error checking vehicles tax expiry: %v", err)
	} else {
		for _, v := range vehicles {
			daysLeft := int(v.TaxExpiryDate.Sub(now).Hours() / 24)
			if daysLeft < 0 {
				daysLeft = 0
			}

			payload := ExpiryAlertPayload{
				Type:       "EXPIRY_ALERT",
				Category:   "VEHICLE_TAX",
				ItemName:   v.Name,
				Detail:     v.LicensePlate,
				DaysLeft:   daysLeft,
				ExpiryDate: v.TaxExpiryDate.Format("2006-01-02"),
			}

			if payloadBytes, err := json.Marshal(payload); err == nil {
				u.redis.LPush(ctx, "queue:email_alerts", payloadBytes)
			}
		}
	}

	// 2. Check BPOM Expirations
	var productsBpom []domain.Product
	if err := u.db.Where("bpom_expiry_date IS NOT NULL AND bpom_expiry_date <= ?", threshold).Find(&productsBpom).Error; err != nil {
		log.Printf("[Cron] Error checking BPOM expiry: %v", err)
	} else {
		for _, p := range productsBpom {
			daysLeft := int(p.BPOMExpiryDate.Sub(now).Hours() / 24)
			if daysLeft < 0 {
				daysLeft = 0
			}

			payload := ExpiryAlertPayload{
				Type:       "EXPIRY_ALERT",
				Category:   "BPOM",
				ItemName:   p.Name,
				Detail:     p.BPOMNumber,
				DaysLeft:   daysLeft,
				ExpiryDate: p.BPOMExpiryDate.Format("2006-01-02"),
			}

			if payloadBytes, err := json.Marshal(payload); err == nil {
				u.redis.LPush(ctx, "queue:email_alerts", payloadBytes)
			}
		}
	}

	// 3. Check Halal Expirations
	var productsHalal []domain.Product
	if err := u.db.Where("halal_expiry_date IS NOT NULL AND halal_expiry_date <= ?", threshold).Find(&productsHalal).Error; err != nil {
		log.Printf("[Cron] Error checking HALAL expiry: %v", err)
	} else {
		for _, p := range productsHalal {
			daysLeft := int(p.HalalExpiryDate.Sub(now).Hours() / 24)
			if daysLeft < 0 {
				daysLeft = 0
			}

			payload := ExpiryAlertPayload{
				Type:       "EXPIRY_ALERT",
				Category:   "HALAL",
				ItemName:   p.Name,
				Detail:     p.HalalNumber,
				DaysLeft:   daysLeft,
				ExpiryDate: p.HalalExpiryDate.Format("2006-01-02"),
			}

			if payloadBytes, err := json.Marshal(payload); err == nil {
				u.redis.LPush(ctx, "queue:email_alerts", payloadBytes)
			}
		}
	}

	return nil
}
