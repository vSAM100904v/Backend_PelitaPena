package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func CreateTrackingLaporan(c *fiber.Ctx) error {
	userToken, ok := c.Locals("user").(*jwt.Token)
    if !ok || userToken == nil {
        return c.Status(fiber.StatusInternalServerError).JSON(Response{
            Success: 0,
            Message: "Internal server error: Unable to retrieve user token",
            Data:    nil,
        })
    }
	claims, ok := userToken.Claims.(jwt.MapClaims)
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(Response{
            Success: 0,
            Message: "Internal server error: Invalid token claims",
            Data:    nil,
        })
    }

    // Ambil userID dari claims
    userIDFloat, ok := claims["user_id"].(float64)
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(Response{
            Success: 0,
            Message: "Internal server error: Invalid user ID in token",
            Data:    nil,
        })
    }
    userID := uint(userIDFloat)
	
	var trackingLaporan models.TrackingLaporan
	if err := c.BodyParser(&trackingLaporan); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	noRegistrasi := c.FormValue("no_registrasi")
	if noRegistrasi == "" {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "No Registrasi is required",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	var existingLaporan models.Laporan
	if err := database.GetGormDBInstance().Where("no_registrasi = ?", noRegistrasi).First(&existingLaporan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response := helper.ResponseWithOutData{
				Code:    http.StatusBadRequest,
				Status:  "error",
				Message: "No Registrasi not found in Laporan table",
			}
			return c.Status(http.StatusBadRequest).JSON(response)
		}
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Database error",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve multipart form",
		})
	}
	files := form.File["document"]
	imageURLs, err := helper.UploadMultipleFileToCloudinary(files)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to upload documents",
		})
	}

	trackingLaporan.Document = datatypes.JSONMap{"urls": imageURLs}
	trackingLaporan.NoRegistrasi = noRegistrasi
	trackingLaporan.Keterangan = c.FormValue("keterangan")
    now := time.Now()
    trackingLaporan.CreatedAt = now
    trackingLaporan.UpdatedAt = now


	if err := database.GetGormDBInstance().Create(&trackingLaporan).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Failed to create tracking laporan",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	var user models.User
	 db := database.GetGormDBInstance()
    if err := db.Where("id = ?", existingLaporan.UserID).First(&user).Error; err == nil && user.NotificationToken != "" {
        notificationData := models.FCMNotificationData{
            Type:      "tracking_update",
            ReportID:  noRegistrasi,
            Status:    "new_tracking",
			UpdatedBy: userID,
            UpdatedAt: now.Format(time.RFC3339),
            Notes:     trackingLaporan.Keterangan,
            DeepLink:  "laporanku://tracking/" + noRegistrasi,
        }

        docMessage := "Ada update baru nih!"
        if len(imageURLs) > 0 {
            docMessage = "Ada dokumen baru (PDF/Image) yang diunggah untuk laporanmu! 📎"
        }

        notification, err := NewNotificationFromFCMData(
            existingLaporan.UserID,
            "Update Baru pada Laporanmu!",
            "Halo! Tracking laporan dengan No. "+noRegistrasi+" telah ditambahkan. "+docMessage+" Cek sekarang yuk!",
            notificationData,
            now,
        )
        if err != nil {
            log.Printf("Error creating notification: %v", err)
        } else {
            if err := db.Create(&notification).Error; err != nil {
                log.Printf("Failed to store notification: %v", err)
            }
            if err := SendFCMNotification(user.NotificationToken, notificationData, *notification); err != nil {
                log.Printf("Failed to send FCM notification: %v", err)
            }
        }
    }
	response := helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Tracking laporan created successfully",
		Data: fiber.Map{
			"id":            trackingLaporan.ID,
			"no_registrasi": trackingLaporan.NoRegistrasi,
			"keterangan":    trackingLaporan.Keterangan,
			"document":      trackingLaporan.Document,
			"created_at":    trackingLaporan.CreatedAt,
			"updated_at":    trackingLaporan.UpdatedAt,
		},
	}

	return c.Status(http.StatusCreated).JSON(response)
}



func UpdateTrackingLaporan(c *fiber.Ctx) error {
    log.Println("===> Mulai UpdateTrackingLaporan")
    userToken, ok := c.Locals("user").(*jwt.Token)
    if !ok || userToken == nil {
        log.Println("❌ Gagal mendapatkan user token dari context")
        return c.Status(fiber.StatusInternalServerError).JSON(Response{
            Success: 0,
            Message: "Internal server error: Unable to retrieve user token",
            Data:    nil,
        })
    }
	claims, ok := userToken.Claims.(jwt.MapClaims)
    if !ok {
        log.Println("❌ Token claims tidak valid")
        return c.Status(fiber.StatusInternalServerError).JSON(Response{
            Success: 0,
            Message: "Internal server error: Invalid token claims",
            Data:    nil,
        })
    }

    // Ambil userID dari claims
    userIDFloat, ok := claims["user_id"].(float64)
    if !ok {
        log.Println("❌ User ID tidak valid dalam token claims")
        return c.Status(fiber.StatusInternalServerError).JSON(Response{
            Success: 0,
            Message: "Internal server error: Invalid user ID in token",
            Data:    nil,
        })
    }
    userID := uint(userIDFloat)
	
	trackingLaporanID := c.Params("id")
    if trackingLaporanID == "" {
        log.Println("❌ Parameter ID kosong")
        return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
            Code:    http.StatusBadRequest,
            Status:  "error",
            Message: "ID is required",
        })
    }

    db := database.GetGormDBInstance()
    var trackingLaporan models.TrackingLaporan
    if err := db.First(&trackingLaporan, trackingLaporanID).Error; err != nil {
        log.Printf("❌ Gagal menemukan tracking laporan: %v\n", err)
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return c.Status(http.StatusNotFound).JSON(helper.ResponseWithOutData{
                Code:    http.StatusNotFound,
                Status:  "error",
                Message: "Tracking Laporan not found",
            })
        }
        return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
            Code:    http.StatusInternalServerError,
            Status:  "error",
            Message: "Database error",
        })
    }

    var updatedData models.TrackingLaporan
    if err := c.BodyParser(&updatedData); err != nil {
        log.Printf("❌ Gagal parsing body request: %v\n", err)
        return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
            Code:    http.StatusBadRequest,
            Status:  "error",
            Message: "Invalid request body",
        })
    }
    log.Println("✅ Body request berhasil di-parse")
    form, err := c.MultipartForm()
    var imageURLs []string
    if err == nil && form != nil {
        files := form.File["document"]
        log.Printf("📦 Jumlah file yang diterima: %d\n", len(files))
        if len(files) > 0 {
            imageURLs, err = helper.UploadMultipleFileToCloudinary(files)
            if err != nil {
                log.Printf("❌ Gagal upload file ke Cloudinary: %v\n", err)
                return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
                    "error": "Failed to upload images",
                })
            }
            log.Printf("✅ Berhasil upload file. URLs: %v\n", imageURLs)
            trackingLaporan.Document = datatypes.JSONMap{"urls": imageURLs}
        }
    }

    if updatedData.NoRegistrasi != "" {
        trackingLaporan.NoRegistrasi = updatedData.NoRegistrasi
    }
    if updatedData.Keterangan != "" {
        trackingLaporan.Keterangan = updatedData.Keterangan
    }
    if updatedData.Document != nil && len(imageURLs) == 0 {
        trackingLaporan.Document = updatedData.Document
    }

    now := time.Now()
    trackingLaporan.UpdatedAt = now

    if err := db.Save(&trackingLaporan).Error; err != nil {
        log.Printf("❌ Gagal menyimpan perubahan ke database: %v\n", err)
        return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
            Code:    http.StatusInternalServerError,
            Status:  "error",
            Message: "Failed to update tracking laporan",
        })
    }

    // Notifikasi untuk user
    var laporan models.Laporan
    if err := db.Where("no_registrasi = ?", trackingLaporan.NoRegistrasi).First(&laporan).Error; err == nil {
        var user models.User
        if err := db.Where("id = ?", laporan.UserID).First(&user).Error; err == nil && user.NotificationToken != "" {
            notificationData := models.FCMNotificationData{
                Type:      "tracking_update",
                ReportID:  trackingLaporan.NoRegistrasi,
                Status:    "updated_tracking",
                UpdatedAt: now.Format(time.RFC3339),
                Notes:     trackingLaporan.Keterangan,
				UpdatedBy: userID,
                DeepLink:  "laporanku://tracking/" + trackingLaporan.NoRegistrasi,
            }

            docMessage := "Ada perubahan terbaru pada tracking laporanmu!"
            if len(imageURLs) > 0 {
                docMessage = "Dokumen baru (PDF/Image) telah diperbarui untuk laporanmu!"
            }

            notification, err := NewNotificationFromFCMData(
                laporan.UserID,
                "Tracking Laporanmu Diperbarui!",
                "Yay! Tracking untuk laporan No. "+trackingLaporan.NoRegistrasi+" telah diperbarui. "+docMessage+" Yuk cek detailnya!",
                notificationData,
                now,
            )
            if err != nil {
                log.Printf("Error creating notification: %v", err)
            } else {
                if err := db.Create(&notification).Error; err != nil {
                    log.Printf("Failed to store notification: %v", err)
                }
                if err := SendFCMNotification(user.NotificationToken, notificationData, *notification); err != nil {
                    log.Printf("Failed to send FCM notification: %v", err)
                }
            }
        }
    }

    response := helper.ResponseWithData{
        Code:    http.StatusOK,
        Status:  "success",
        Message: "Tracking laporan updated successfully",
        Data: fiber.Map{
            "id":            trackingLaporan.ID,
            "no_registrasi": trackingLaporan.NoRegistrasi,
            "keterangan":    trackingLaporan.Keterangan,
            "document":      trackingLaporan.Document,
            "created_at":    trackingLaporan.CreatedAt,
            "updated_at":    trackingLaporan.UpdatedAt,
        },
    }
    return c.Status(http.StatusOK).JSON(response)
}



func DeleteTrackingLaporan(c *fiber.Ctx) error {
   userToken, ok := c.Locals("user").(*jwt.Token)
    if !ok || userToken == nil {
        return c.Status(fiber.StatusInternalServerError).JSON(Response{
            Success: 0,
            Message: "Internal server error: Unable to retrieve user token",
            Data:    nil,
        })
    }
	claims, ok := userToken.Claims.(jwt.MapClaims)
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(Response{
            Success: 0,
            Message: "Internal server error: Invalid token claims",
            Data:    nil,
        })
    }

    // Ambil userID dari claims
    userIDFloat, ok := claims["user_id"].(float64)
    if !ok {
        return c.Status(fiber.StatusInternalServerError).JSON(Response{
            Success: 0,
            Message: "Internal server error: Invalid user ID in token",
            Data:    nil,
        })
    }
    userID := uint(userIDFloat)
	
	trackingLaporanID := c.Params("id")
    if trackingLaporanID == "" {
        return c.Status(http.StatusBadRequest).JSON(helper.ResponseWithOutData{
            Code:    http.StatusBadRequest,
            Status:  "error",
            Message: "ID is required",
        })
    }

    db := database.GetGormDBInstance()
    var trackingLaporan models.TrackingLaporan
    if err := db.First(&trackingLaporan, trackingLaporanID).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return c.Status(http.StatusNotFound).JSON(helper.ResponseWithOutData{
                Code:    http.StatusNotFound,
                Status:  "error",
                Message: "Tracking Laporan not found",
            })
        }
        return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
            Code:    http.StatusInternalServerError,
            Status:  "error",
            Message: "Database error",
        })
    }

    noRegistrasi := trackingLaporan.NoRegistrasi
    now := time.Now()

    // Notifikasi sebelum delete
    var laporan models.Laporan
    if err := db.Where("no_registrasi = ?", noRegistrasi).First(&laporan).Error; err == nil {
        var user models.User
        if err := db.Where("id = ?", laporan.UserID).First(&user).Error; err == nil && user.NotificationToken != "" {
            notificationData := models.FCMNotificationData{
                Type:      "tracking_update",
                ReportID:  noRegistrasi,
                Status:    "deleted_tracking",
                UpdatedAt: now.Format(time.RFC3339),
                Notes:     "Tracking Laporan telah dihapus",
				UpdatedBy: userID,
			}

            notification, err := NewNotificationFromFCMData(
                laporan.UserID,
                "Tracking Laporan telah dihapus!",
                "Halo! Tracking untuk laporan No. "+noRegistrasi+" telah dihapus dari sistem. Ada pertanyaan? Hubungi kami ya!",
                notificationData,
                now,
            )
            if err != nil {
                log.Printf("Error creating notification: %v", err)
            } else {
                if err := db.Create(&notification).Error; err != nil {
                    log.Printf("Failed to store notification: %v", err)
                }
                if err := SendFCMNotification(user.NotificationToken, notificationData, *notification); err != nil {
                    log.Printf("Failed to send FCM notification: %v", err)
                }
            }
        }
    }

    if err := db.Delete(&trackingLaporan).Error; err != nil {
        return c.Status(http.StatusInternalServerError).JSON(helper.ResponseWithOutData{
            Code:    http.StatusInternalServerError,
            Status:  "error",
            Message: "Failed to delete tracking laporan",
        })
    }

    return c.Status(http.StatusOK).JSON(helper.ResponseWithOutData{
        Code:    http.StatusOK,
        Status:  "success",
        Message: "Tracking laporan deleted successfully",
    })
}