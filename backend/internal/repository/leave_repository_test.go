package repository

import (
	"testing"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUpdateLeaveKeepsEditedForeignKeysWhenAssociationsArePreloaded(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=localhost user=test dbname=test sslmode=disable",
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, SkipDefaultTransaction: true})
	if err != nil {
		t.Fatal(err)
	}

	oldEmployeeID, newEmployeeID := uuid.New(), uuid.New()
	oldTypeID, newTypeID := uuid.New(), uuid.New()
	leave := &domain.LeaveRequest{
		ID:          uuid.New(),
		EmployeeID:  newEmployeeID,
		LeaveTypeID: newTypeID,
		Employee:    &domain.Employee{ID: oldEmployeeID},
		LeaveType:   &domain.LeaveType{ID: oldTypeID},
	}

	if err := NewLeaveRepository(db).Update(leave); err != nil {
		t.Fatal(err)
	}
	if leave.EmployeeID != newEmployeeID || leave.LeaveTypeID != newTypeID {
		t.Fatalf("preloaded associations overwrote edited IDs: employee=%s, leave type=%s", leave.EmployeeID, leave.LeaveTypeID)
	}
}
