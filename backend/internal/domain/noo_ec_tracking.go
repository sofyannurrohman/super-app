package domain

import (
	"time"

	"github.com/google/uuid"
)

type SalesmanNooEcSummary struct {
	EmployeeID          uuid.UUID `json:"employee_id"`
	EmployeeIDNumber    string    `json:"employee_id_number"`
	EmployeeName        string    `json:"employee_name"`
	JobTitle            string    `json:"job_title"`
	BranchName          string    `json:"branch_name"`
	OperatingBranchName string    `json:"operating_branch_name,omitempty"`
	SalesArea           string    `json:"sales_area"`
	TotalAssignedStores int       `json:"total_assigned_stores"`
	TotalVisits         int       `json:"total_visits"`
	EffectiveCalls      int       `json:"effective_calls"`
	EcRatio             float64   `json:"ec_ratio"`
	NooCount            int       `json:"noo_count"`
	TotalOrderCount     int       `json:"total_order_count"`
	TotalOrderAmount    float64   `json:"total_order_amount"`
}

type StoreNooEcDetail struct {
	StoreID           uuid.UUID  `json:"store_id"`
	StoreName         string     `json:"store_name"`
	OwnerName         string     `json:"owner_name"`
	PhoneNumber       string     `json:"phone_number"`
	Address           string     `json:"address"`
	City              string     `json:"city"`
	District          string     `json:"district"`
	CreatedAt         time.Time  `json:"created_at"`
	IsNooInPeriod     bool       `json:"is_noo_in_period"`
	SalesmanID        *uuid.UUID `json:"salesman_id,omitempty"`
	SalesmanIDNumber  string     `json:"salesman_id_number"`
	SalesmanName      string     `json:"salesman_name"`
	SalesmanArea      string     `json:"salesman_area"`
	LastVisitTime     *time.Time `json:"last_visit_time"`
	IsVisitedInPeriod bool       `json:"is_visited_in_period"`
	IsEffectiveCall   bool       `json:"is_effective_call"`
	OrderCount        int        `json:"order_count"`
	OrderAmount       float64    `json:"order_amount"`
	LatestOrderNumber string     `json:"latest_order_number"`
	Latitude          float64    `json:"latitude"`
	Longitude         float64    `json:"longitude"`
}

type NooEcTrackingResponse struct {
	StartDate        string                 `json:"start_date"`
	EndDate          string                 `json:"end_date"`
	TotalCalls       int                    `json:"total_calls"`
	TotalEc          int                    `json:"total_ec"`
	EcRatio          float64                `json:"ec_ratio"`
	TotalNoo         int                    `json:"total_noo"`
	TotalOrderAmount float64                `json:"total_order_amount"`
	SalesmanMatrix   []SalesmanNooEcSummary `json:"salesman_matrix"`
	StoreLogs        []StoreNooEcDetail     `json:"store_logs"`
}
