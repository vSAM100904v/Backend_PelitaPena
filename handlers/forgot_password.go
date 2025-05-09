package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/models"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

func ForgotPassword(c *fiber.Ctx) error {
	var req ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: "Invalid request body", Data: nil})
	}

	db := database.GetGormDBInstance()

	var user models.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(Response{Success: 0, Message: "Email not found", Data: nil})
	}

	token, err := generateResetToken()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to generate reset token", Data: nil})
	}

	reset := models.PasswordReset{
		Email:     req.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
		CreatedAt: time.Now(),
	}

	if err := db.Create(&reset).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to save reset token", Data: nil})
	}

	if err := sendResetEmail(req.Email, token); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to send reset email", Data: nil})
	}

	return c.Status(http.StatusOK).JSON(Response{Success: 1, Message: "Reset email sent successfully", Data: nil})
}

func generateResetToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func sendResetEmail(email, token string) error {
	from := os.Getenv("EMAIL_SENDER")
	password := os.Getenv("EMAIL_PASSWORD")
	to := []string{email}
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	msg := []byte(fmt.Sprintf("Subject: Password Reset\n\nClick the link to reset your password: http://localhost:3000/reset-password?token=%s", token))

	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		return err
	}
	return nil
}
func ResetPassword(c *fiber.Ctx) error {
	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: "Invalid request body", Data: nil})
	}

	if req.NewPassword != req.ConfirmPassword {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: "New password and confirmation password do not match", Data: nil})
	}

	db := database.GetGormDBInstance()

	var reset models.PasswordReset
	if err := db.Where("token = ? AND expires_at > ?", req.Token, time.Now()).First(&reset).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(Response{Success: 0, Message: "Invalid or expired token", Data: nil})
	}

	var user models.User
	if err := db.Where("email = ?", reset.Email).First(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "User not found", Data: nil})
	}

	hashedPassword, err := hashiPassword(req.NewPassword)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to hash new password", Data: nil})
	}

	if err := db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(Response{Success: 0, Message: "Failed to update password", Data: nil})
	}

	db.Delete(&reset)

	return c.Status(http.StatusOK).JSON(Response{Success: 1, Message: "Password reset successfully", Data: nil})
}

func hashiPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
