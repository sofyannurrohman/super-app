package usecase

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"example.com/hris/backend/internal/domain"
	"example.com/hris/backend/internal/repository"
	"example.com/hris/backend/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	Register(req RegisterRequest) error
	Login(req LoginRequest, secret string) (*LoginResponse, error)
	ForgotPassword(email string) error
	RefreshToken(refreshToken, secret string) (*LoginResponse, error)
}

type authUseCase struct {
	userRepo          repository.UserRepository
	companyRepo       repository.CompanyRepository
	employeeShiftRepo repository.EmployeeShiftRepository
	emailSender       utils.EmailSender
	fcmUseCase        FcmUsecase
}

func NewAuthUseCase(userRepo repository.UserRepository, companyRepo repository.CompanyRepository, employeeShiftRepo repository.EmployeeShiftRepository, emailSender utils.EmailSender, fcmUseCase FcmUsecase) AuthUseCase {
	return &authUseCase{userRepo, companyRepo, employeeShiftRepo, emailSender, fcmUseCase}
}

type RegisterRequest struct {
	Name          string            `json:"name" binding:"required"`
	Email         string            `json:"email" binding:"required,email"`
	EmployeeID    string            `json:"employee_id" binding:"required"`
	Password      string            `json:"password" binding:"required,min=6"`
	JobPositionID string            `json:"job_position_id" binding:"required"`
	BranchID      string            `json:"branch_id" binding:"required"`
	ShiftID       string            `json:"shift_id"`
	FaceEmbedding domain.FloatArray `json:"face_embedding"`
	Selfie        string            `json:"selfie"` // Base64
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Client   string `json:"client"`    // "mobile" or "web"
	FCMToken string `json:"fcm_token"` // Current device's FCM token (optional)
}

type LoginResponse struct {
	AccessToken  string       `json:"token"` // Changed to "token" to match frontend
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type UserResponse struct {
	CompanyName          string   `json:"company_name,omitempty"`
	CompanyAddress       string   `json:"company_address,omitempty"`
	CompanyLogoURL       string   `json:"company_logo_url,omitempty"`
	ID                   string   `json:"id"`
	Email                string   `json:"email"`
	Role                 string   `json:"role"`
	BranchID             string   `json:"branch_id,omitempty"`
	Menus                []string `json:"menus,omitempty"`
	HasGlobalAccess      bool     `json:"has_global_access"`
	DistributorStoreID   string   `json:"distributor_store_id,omitempty"`
	DistributorStoreName string   `json:"distributor_store_name,omitempty"`
}

func (u *authUseCase) Register(req RegisterRequest) error {
	existing, _ := u.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return errors.New("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &domain.User{
		Email:        req.Email,
		PasswordHash: string(hashed),
		Role:         domain.RoleEmployee, // Default role
		IsActive:     true,
	}

	// Assign self-registered employees to the configured trial company.
	defaultCompany, err := u.companyRepo.FindByName(os.Getenv("TRIAL_COMPANY_NAME"))
	if err == nil && defaultCompany != nil {
		user.CompanyID = &defaultCompany.ID
	}

	jobPosID, err := uuid.Parse(req.JobPositionID)
	if err != nil {
		return errors.New("invalid job position ID format")
	}

	branchID, err := uuid.Parse(req.BranchID)
	if err != nil {
		return errors.New("invalid branch ID format")
	}

	employee := &domain.Employee{
		FirstName:        req.Name,
		EmployeeIDNumber: req.EmployeeID,
		JobPositionID:    &jobPosID,
		BranchID:         &branchID,
		JoinDate:         time.Now(),
		CompanyID:        user.CompanyID, // Inherit from user
	}

	// Handle Face Data if provided
	if req.Selfie != "" && len(req.FaceEmbedding) > 0 {
		uploadDir := "uploads/faces"
		_ = os.MkdirAll(uploadDir, 0755)

		imageData, err := base64.StdEncoding.DecodeString(req.Selfie)
		if err == nil {
			compressedData, err := utils.CompressImage(imageData, 800, 75)
			if err == nil {
				filename := fmt.Sprintf("reg_%s_%d.jpg", req.EmployeeID, time.Now().Unix())
				filePath := filepath.Join(uploadDir, filename)
				if err := os.WriteFile(filePath, compressedData, 0644); err == nil {
					employee.FaceEmbedding = req.FaceEmbedding
					employee.FaceReferenceURL = "/" + filePath
					employee.IsFaceRegistered = true
				}
			}
		}
	}

	// Set permanent default shift if provided
	if req.ShiftID != "" {
		shiftID, err := uuid.Parse(req.ShiftID)
		if err == nil {
			employee.DefaultShiftID = &shiftID
		}
	}

	err = u.userRepo.CreateWithEmployee(user, employee)
	if err != nil {
		return err
	}

	return nil
}

func (u *authUseCase) Login(req LoginRequest, secret string) (*LoginResponse, error) {
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	user, err := u.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("Email atau kata sandi tidak valid")
	}

	if !user.IsActive {
		return nil, errors.New("user is inactive")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("Email atau kata sandi tidak valid")
	}

	companyID := uuid.Nil
	if user.CompanyID != nil {
		companyID = *user.CompanyID
	}

	var accessToken, refreshToken string
	if strings.ToLower(req.Client) == "mobile" {
		newMobileSessionID := uuid.New().String()

		// 1. Kick out previous mobile session if exists, EXCEPT current device's FCM token
		if u.fcmUseCase != nil {
			_ = u.fcmUseCase.SendForceLogoutExcept(user.ID, req.FCMToken, newMobileSessionID)
		}

		// 2. Update mobile session ID in DB
		_ = u.userRepo.UpdateMobileSession(user.ID, &newMobileSessionID)

		// 3. Issue JWT with mobile_session_id claim
		var err error
		accessToken, refreshToken, err = utils.GenerateTokensWithSession(user.ID, companyID, string(user.Role), newMobileSessionID, secret)
		if err != nil {
			return nil, errors.New("failed to generate token")
		}
	} else {
		var err error
		accessToken, refreshToken, err = utils.GenerateTokens(user.ID, companyID, string(user.Role), secret)
		if err != nil {
			return nil, errors.New("failed to generate token")
		}
	}

	branchIDStr := ""
	if user.Role != domain.RoleSuperAdmin {
		if bid, err := u.userRepo.GetBranchIDByUserID(user.ID); err == nil {
			branchIDStr = bid
		}
	}

	var menus []string
	if user.AccessRole != nil && len(user.AccessRole.Menus) > 0 {
		for _, m := range user.AccessRole.Menus {
			if m.Path != "" {
				menus = append(menus, m.Path)
			}
		}
	}

	hasGlobalAccess := false
	if user.Role == domain.RoleSuperAdmin || user.Role == domain.RoleCustomerSupport || user.Role == domain.RoleAdminDesigner || user.Role == domain.RoleValidatorDesain || user.Role == domain.RoleValidatorToko || user.Role == domain.RoleAdminPurchasing {
		hasGlobalAccess = true
	} else if user.Role == domain.RoleAuditor {
		isBooster := false
		if branch, err := u.userRepo.GetBranchByUserID(user.ID); err == nil && branch != nil {
			isBooster = strings.Contains(strings.ToLower(branch.Name), "booster")
		}
		if !isBooster {
			hasGlobalAccess = true
		}
	}
	if !hasGlobalAccess {
		if hasPerm, err := u.userRepo.HasPermission(user.ID, "view_all_branches"); err == nil && hasPerm {
			hasGlobalAccess = true
		}
	}

	distributorStoreIDStr := ""
	distributorStoreName := ""
	if user.DistributorStoreID != nil {
		distributorStoreIDStr = user.DistributorStoreID.String()
		if user.DistributorStore != nil {
			distributorStoreName = user.DistributorStore.Name
		}
	}

	companyName, companyAddress, companyLogoURL := "", "", ""
	if user.Company != nil {
		companyName = user.Company.Name
		companyAddress = user.Company.Address
		companyLogoURL = user.Company.LogoURL
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserResponse{
			CompanyName: companyName, CompanyAddress: companyAddress, CompanyLogoURL: companyLogoURL,
			ID:                   user.ID.String(),
			Email:                user.Email,
			Role:                 string(user.Role),
			BranchID:             branchIDStr,
			Menus:                menus,
			HasGlobalAccess:      hasGlobalAccess,
			DistributorStoreID:   distributorStoreIDStr,
			DistributorStoreName: distributorStoreName,
		},
	}, nil
}

func (u *authUseCase) ForgotPassword(email string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	user, err := u.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return errors.New("email tidak terdaftar di sistem kami")
	}

	// Generate a temporary random password
	tempPassword := utils.GenerateRandomString(8)
	hashed, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal menggenerasi kata sandi baru")
	}

	// Update user password in database
	user.PasswordHash = string(hashed)
	err = u.userRepo.Update(user)
	if err != nil {
		return errors.New("gagal memperbarui kata sandi di sistem")
	}

	// Send email
	subject := "Pemulihan Kata Sandi - Akun HRIS"
	body := fmt.Sprintf(`
		<html>
			<body>
				<h2>Halo, %s</h2>
				<p>Kami telah menerima permintaan pemulihan kata sandi untuk akun Anda.</p>
				<p>Kata sandi sementara Anda adalah: <b>%s</b></p>
				<p>Silakan gunakan kata sandi ini untuk masuk dan segera ubah kata sandi Anda di menu Profil demi keamanan.</p>
				<br>
				<p>Terima kasih,<br>Tim HR</p>
			</body>
		</html>
	`, email, tempPassword)

	err = u.emailSender.SendEmail(email, subject, body)
	if err != nil {
		fmt.Printf("Error sending email to %s: %v\n", email, err)
		return errors.New("gagal mengirim email instruksi. silakan coba lagi nanti")
	}

	return nil
}
func (u *authUseCase) RefreshToken(refreshToken, secret string) (*LoginResponse, error) {
	claims, err := utils.ValidateRefreshToken(refreshToken, secret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return nil, errors.New("invalid user id in token")
	}

	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if !user.IsActive {
		return nil, errors.New("user is inactive")
	}

	companyID := uuid.Nil
	if user.CompanyID != nil {
		companyID = *user.CompanyID
	}

	var accessToken, newRefreshToken string
	if user.MobileSessionID != nil && *user.MobileSessionID != "" {
		accessToken, newRefreshToken, err = utils.GenerateTokensWithSession(user.ID, companyID, string(user.Role), *user.MobileSessionID, secret)
		if err != nil {
			return nil, errors.New("failed to generate token")
		}
	} else {
		accessToken, newRefreshToken, err = utils.GenerateTokens(user.ID, companyID, string(user.Role), secret)
		if err != nil {
			return nil, errors.New("failed to generate token")
		}
	}

	branchIDStr := ""
	if user.Role != domain.RoleSuperAdmin {
		if bid, err := u.userRepo.GetBranchIDByUserID(user.ID); err == nil {
			branchIDStr = bid
		}
	}

	var menus []string
	if user.AccessRole != nil && len(user.AccessRole.Menus) > 0 {
		for _, m := range user.AccessRole.Menus {
			if m.Path != "" {
				menus = append(menus, m.Path)
			}
		}
	}

	hasGlobalAccess := false
	if user.Role == domain.RoleSuperAdmin || user.Role == domain.RoleCustomerSupport || user.Role == domain.RoleAdminDesigner || user.Role == domain.RoleValidatorDesain || user.Role == domain.RoleValidatorToko || user.Role == domain.RoleAdminPurchasing {
		hasGlobalAccess = true
	} else if user.Role == domain.RoleAuditor {
		isBooster := false
		if branch, err := u.userRepo.GetBranchByUserID(user.ID); err == nil && branch != nil {
			isBooster = strings.Contains(strings.ToLower(branch.Name), "booster")
		}
		if !isBooster {
			hasGlobalAccess = true
		}
	}
	if !hasGlobalAccess {
		if hasPerm, err := u.userRepo.HasPermission(user.ID, "view_all_branches"); err == nil && hasPerm {
			hasGlobalAccess = true
		}
	}

	distributorStoreIDStr := ""
	distributorStoreName := ""
	if user.DistributorStoreID != nil {
		distributorStoreIDStr = user.DistributorStoreID.String()
		if user.DistributorStore != nil {
			distributorStoreName = user.DistributorStore.Name
		}
	}

	companyName, companyAddress, companyLogoURL := "", "", ""
	if user.Company != nil {
		companyName = user.Company.Name
		companyAddress = user.Company.Address
		companyLogoURL = user.Company.LogoURL
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User: UserResponse{
			CompanyName: companyName, CompanyAddress: companyAddress, CompanyLogoURL: companyLogoURL,
			ID:                   user.ID.String(),
			Email:                user.Email,
			Role:                 string(user.Role),
			BranchID:             branchIDStr,
			Menus:                menus,
			HasGlobalAccess:      hasGlobalAccess,
			DistributorStoreID:   distributorStoreIDStr,
			DistributorStoreName: distributorStoreName,
		},
	}, nil
}
