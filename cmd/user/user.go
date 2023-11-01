package user

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"regexp"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type SignUpRequest struct {
	Email    string
	Password string
}

type UserResponse struct {
	Email    string  `json:"email"`
	Bio      *string `json:"bio,omitempty"`
	FullName *string `json:"fullName,omitempty"`
	ApiKey   string  `json:"apiKey,omitempty"`
	Role     string  `json:"role"`
}

type UpdateProfileRequest struct {
	Email    string  `json:"email"`
	Bio      *string `json:"bio,omitempty"`
	FullName *string `json:"fullName,omitempty"`
}

type UpdatePasswordRequest struct {
	Password string `json:"password"`
}

type AddRoleRequest struct {
	Role string `json:"role"`
}

type User struct {
	gorm.Model
	Email        string  `gorm:"type:varchar(255);unique;not null"`
	Bio          *string `gorm:"type:varchar(100);"`
	FullName     *string `gorm:"type:varchar(255)"`
	Role         int8    `gorm:"type:smallint"`
	PasswordHash string  `gorm:"type:varchar(100);not null"`
	ApiKey       string  `gorm:"type:varchar(100);not null"`
}

type UserRole int8

const (
	Member     UserRole = 1
	Supervisor UserRole = 2
)

var roleText = map[UserRole]string{
	Member:     "Member",
	Supervisor: "SUPERVISOR",
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func RegisterUser(c *fiber.Ctx, db *gorm.DB) error {

	signUpRequest := new(SignUpRequest)
	if err := c.BodyParser(signUpRequest); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid input or missing data",
		})
	}

	if !validatePassword(signUpRequest.Password) {
		errorResponse := ErrorResponse{
			Code:    fiber.StatusUnprocessableEntity,
			Type:    "UNPROCESSABLE_ENTITY",
			Message: "The password must be at least 10 characters, contains numeric characters, minimum 1 uppercase letter [A-Z] and minimum 1 special character",
		}
		return c.Status(fiber.StatusUnprocessableEntity).JSON(errorResponse)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(signUpRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	apiKey, err := generateAPIKey(50)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	user := new(User)
	user.Email = signUpRequest.Email
	user.PasswordHash = string(passwordHash)
	user.Role = int8(Member)
	user.ApiKey = apiKey

	result := db.Where("email = ?", signUpRequest.Email).First(&user)
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	if result.RowsAffected > 0 {
		errorResponse := ErrorResponse{
			Code:    fiber.StatusBadRequest,
			Type:    "BAD_REQUEST",
			Message: "E-mail already exists",
		}

		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	if result := db.Create(&user); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	userResponse := new(UserResponse)
	userResponse.Email = user.Email
	userResponse.Role = roleText[UserRole(user.Role)]
	userResponse.ApiKey = user.ApiKey

	return c.Status(fiber.StatusCreated).JSON(userResponse)

}

func LoginUser(c *fiber.Ctx, db *gorm.DB) error {

	log.Println("Register user begin")
	signInRequest := new(SignUpRequest)
	if err := c.BodyParser(signInRequest); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid input or missing data",
		})
	}

	if !validatePassword(signInRequest.Password) {
		return invalidLoginResponse(c)
	}

	user := new(User)
	user.Email = signInRequest.Email

	result := db.Where("email = ?", signInRequest.Email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return invalidLoginResponse(c)
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(signInRequest.Password)); err != nil {
		return invalidLoginResponse(c)
	}

	userResponse := new(UserResponse)
	userResponse.Email = user.Email
	userResponse.Role = roleText[UserRole(user.Role)]
	userResponse.ApiKey = user.ApiKey

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

func GetProfile(c *fiber.Ctx, db *gorm.DB) error {

	accessToken := c.Get("x-access-token")
	if accessToken == "" {
		return missingHeaderResponse(c)
	}

	user := new(User)

	result := db.Where("api_key = ?", accessToken).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return missingHeaderResponse(c)
		}
	}

	if user.ApiKey != accessToken {
		return missingHeaderResponse(c)
	}

	userResponse := new(UserResponse)
	userResponse.Email = user.Email
	userResponse.Bio = user.Bio
	userResponse.Role = roleText[UserRole(user.Role)]
	userResponse.ApiKey = user.ApiKey

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

func UpdateProfile(c *fiber.Ctx, db *gorm.DB) error {

	accessToken := c.Get("x-access-token")
	if accessToken == "" {
		return missingHeaderResponse(c)
	}

	updateProfileRequest := new(UpdateProfileRequest)
	if err := c.BodyParser(updateProfileRequest); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid input or missing data",
		})
	}

	user := new(User)

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid input or missing data",
		})
	}

	result := db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return missingHeaderResponse(c)
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid input or missing data",
			})
		}
	}

	user.Email = updateProfileRequest.Email
	user.Bio = updateProfileRequest.Bio
	user.FullName = updateProfileRequest.FullName

	result = db.Save(user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	userResponse := new(UserResponse)
	userResponse.Email = user.Email
	userResponse.Bio = user.Bio
	userResponse.Role = roleText[UserRole(user.Role)]

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

func UpdatePassword(c *fiber.Ctx, db *gorm.DB) error {

	accessToken := c.Get("x-access-token")
	if accessToken == "" {
		return missingHeaderResponse(c)
	}

	updatePasswordRequest := new(UpdatePasswordRequest)
	if err := c.BodyParser(updatePasswordRequest); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid input or missing data",
		})
	}

	user := new(User)

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid input or missing data",
		})
	}

	result := db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return missingHeaderResponse(c)
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid input or missing data",
			})
		}
	}

	if user.ApiKey != accessToken {
		return missingHeaderResponse(c)
	}

	if !validatePassword(updatePasswordRequest.Password) {
		errorResponse := ErrorResponse{
			Code:    fiber.StatusUnprocessableEntity,
			Type:    "UNPROCESSABLE_ENTITY",
			Message: "The password must be at least 10 characters, contains numeric characters, minimum 1 uppercase letter [A-Z] and minimum 1 special character",
		}

		return c.Status(fiber.StatusUnprocessableEntity).JSON(errorResponse)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(updatePasswordRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
	}

	user.PasswordHash = string(passwordHash)
	result = db.Save(user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	userResponse := new(UserResponse)
	userResponse.Email = user.Email
	userResponse.Bio = user.Bio
	userResponse.Role = roleText[UserRole(user.Role)]

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

func AddRole(c *fiber.Ctx, db *gorm.DB) error {

	accessToken := c.Get("x-access-token")
	if accessToken == "" {
		return missingHeaderResponse(c)
	}

	addRoleRequest := new(AddRoleRequest)
	if err := c.BodyParser(addRoleRequest); err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid input or missing data",
		})
	}

	user := new(User)

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid input or missing data",
		})
	}

	result := db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return missingHeaderResponse(c)
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid input or missing data",
			})
		}
	}

	var role UserRole
	for ur, v := range roleText {
		if v == addRoleRequest.Role {
			role = ur
			user.Role = int8(ur)
		}
	}
	if role == 0 {
		errorResponse := ErrorResponse{
			Code:    fiber.StatusBadRequest,
			Type:    "BAD_REQUEST",
			Message: "Invalid role",
		}
		return c.Status(fiber.StatusBadRequest).JSON(errorResponse)
	}

	result = db.Save(user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	userResponse := new(UserResponse)
	userResponse.Email = user.Email
	userResponse.Bio = user.Bio
	userResponse.Role = roleText[UserRole(user.Role)]

	return c.Status(fiber.StatusOK).JSON(userResponse)
}

func DeleteProfile(c *fiber.Ctx, db *gorm.DB) error {

	accessToken := c.Get("x-access-token")
	if accessToken == "" {
		return missingHeaderResponse(c)
	}

	user := new(User)

	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Bad Request",
			"message": "Invalid input or missing data",
		})
	}

	result := db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return missingHeaderResponse(c)
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Bad Request",
				"message": "Invalid input or missing data",
			})
		}
	}

	result = db.Delete(user)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(result.Error.Error())
	}

	userResponse := new(UserResponse)
	userResponse.Email = user.Email
	userResponse.Bio = user.Bio
	userResponse.Role = roleText[UserRole(user.Role)]

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}

func missingHeaderResponse(c *fiber.Ctx) error {
	missingHeaderResponse := ErrorResponse{
		Code:    fiber.StatusUnauthorized,
		Type:    fiber.ErrUnauthorized.Message,
		Message: "Missing x-access-token header variable",
	}
	return c.Status(fiber.StatusUnauthorized).JSON(missingHeaderResponse)
}

func invalidLoginResponse(c *fiber.Ctx) error {
	errorResponse := ErrorResponse{
		Code:    fiber.StatusUnprocessableEntity,
		Type:    "UNPROCESSABLE_ENTITY",
		Message: "Invalid e-mail or password",
	}
	return c.Status(fiber.StatusUnprocessableEntity).JSON(errorResponse)
}

func generateAPIKey(length int) (string, error) {
	// Since 1 byte = 2 hexadecimal characters, divide the length by 2
	byteLength := length / 2

	// Create a byte slice of the required length
	bytes := make([]byte, byteLength)

	// Read random bytes into the slice
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Encode the byte slice as a hexadecimal string
	apiKey := hex.EncodeToString(bytes)
	return apiKey, nil
}

func validatePassword(password string) bool {
	if len(password) < 10 {
		return false
	}

	var (
		upperCaseRegex   = regexp.MustCompile(`[A-Z]`)
		digitRegex       = regexp.MustCompile(`\d`)
		specialCharRegex = regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\[\]\{\};:'",<>\./\?\\|` + "`" + `]`)
	)

	if !upperCaseRegex.MatchString(password) {
		return false
	}
	if !digitRegex.MatchString(password) {
		return false
	}
	if !specialCharRegex.MatchString(password) {
		return false
	}

	return true
}
