package handlers

import (
	"backend-pedika-fiber/database"
	"backend-pedika-fiber/helper"
	"backend-pedika-fiber/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func CreatePelaku(c *fiber.Ctx) error {
	var pelaku models.Pelaku
	if err := c.BodyParser(&pelaku); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	pelaku.NoRegistrasi = c.FormValue("no_registrasi")
	pelaku.NIKPelaku = c.FormValue("nik_pelaku")
	pelaku.Nama = c.FormValue("nama_pelaku")
	usia, err := strconv.Atoi(c.FormValue("usia_pelaku"))
	if err == nil {
		pelaku.Usia = usia
	}
	pelaku.AlamatPelaku = c.FormValue("alamat_pelaku")
	pelaku.AlamatDetail = c.FormValue("alamat_detail")
	pelaku.JenisKelamin = c.FormValue("jenis_kelamin")
	pelaku.Agama = c.FormValue("agama")
	pelaku.NoTelepon = c.FormValue("no_telepon")
	pelaku.Pendidikan = c.FormValue("pendidikan")
	pelaku.Pekerjaan = c.FormValue("pekerjaan")
	pelaku.StatusPerkawinan = c.FormValue("status_perkawinan")
	pelaku.Kebangsaan = c.FormValue("kebangsaan")
	pelaku.HubunganDenganKorban = c.FormValue("hubungan_dengan_korban")
	pelaku.KeteranganLainnya = c.FormValue("keterangan_lainnya")

	file, err := c.FormFile("dokumentasi_pelaku")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Failed to open image file",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		defer src.Close()

		imageURL, err := helper.UploadFileToCloudinary(src, file.Filename)
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal Mengupload Gambar",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}

		pelaku.DokumentasiPelaku = imageURL
	}
	pelaku.CreatedAt = time.Now()
	pelaku.UpdatedAt = time.Now()
	if err := database.DB.Create(&pelaku).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "gagal Menambahkan Data pelaku",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}
	response := helper.ResponseWithData{
		Code:    http.StatusCreated,
		Status:  "success",
		Message: "Berhasil Menambah Data Pelaku",
		Data:    pelaku,
	}
	return c.Status(http.StatusCreated).JSON(response)
}

func UpdatePelaku(c *fiber.Ctx) error {
	id := c.Params("id")
	var pelaku models.Pelaku
	if err := database.DB.First(&pelaku, id).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Pelaku Tidak DItemukan",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}
	if err := c.BodyParser(&pelaku); err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusBadRequest,
			Status:  "error",
			Message: "Invalid request body",
		}
		return c.Status(http.StatusBadRequest).JSON(response)
	}
	if value := c.FormValue("no_registrasi"); value != "" {
		pelaku.NoRegistrasi = value
	}
	if value := c.FormValue("nik_pelaku"); value != "" {
		pelaku.NIKPelaku = value
	}
	if value := c.FormValue("nama_pelaku"); value != "" {
		pelaku.Nama = value
	}
	if value := c.FormValue("usia_pelaku"); value != "" {
		if usia, err := strconv.Atoi(value); err == nil {
			pelaku.Usia = usia
		}
	}
	if value := c.FormValue("alamat_pelaku"); value != "" {
		pelaku.AlamatPelaku = value
	}
	if value := c.FormValue("alamat_detail"); value != "" {
		pelaku.AlamatDetail = value
	}
	if value := c.FormValue("jenis_kelamin"); value != "" {
		pelaku.JenisKelamin = value
	}
	if value := c.FormValue("agama"); value != "" {
		pelaku.Agama = value
	}
	if value := c.FormValue("no_telepon"); value != "" {
		pelaku.NoTelepon = value
	}
	if value := c.FormValue("pendidikan"); value != "" {
		pelaku.Pendidikan = value
	}
	if value := c.FormValue("pekerjaan"); value != "" {
		pelaku.Pekerjaan = value
	}
	if value := c.FormValue("status_perkawinan"); value != "" {
		pelaku.StatusPerkawinan = value
	}
	if value := c.FormValue("kebangsaan"); value != "" {
		pelaku.Kebangsaan = value
	}
	if value := c.FormValue("hubungan_dengan_korban"); value != "" {
		pelaku.HubunganDenganKorban = value
	}
	if value := c.FormValue("keterangan_lainnya"); value != "" {
		pelaku.KeteranganLainnya = value
	}

	file, err := c.FormFile("dokumentasi_pelaku")
	if err == nil {
		src, err := file.Open()
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Failed to open image file",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}
		defer src.Close()

		imageURL, err := helper.UploadFileToCloudinary(src, file.Filename)
		if err != nil {
			response := helper.ResponseWithOutData{
				Code:    http.StatusInternalServerError,
				Status:  "error",
				Message: "Gagal Mengupload Gambat",
			}
			return c.Status(http.StatusInternalServerError).JSON(response)
		}

		pelaku.DokumentasiPelaku = imageURL
	}

	pelaku.UpdatedAt = time.Now()
	if err := database.DB.Save(&pelaku).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal Mengupdate Data Pelaku",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Berhasil Mengupdated Data Pelaku",
		Data:    pelaku,
	}
	return c.Status(http.StatusOK).JSON(response)
}

func DeletePelaku(c *fiber.Ctx) error {
	id := c.Params("id")
	var pelaku models.Pelaku
	if err := database.DB.First(&pelaku, id).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusNotFound,
			Status:  "error",
			Message: "Pelaku Tidak Ditemukan",
		}
		return c.Status(http.StatusNotFound).JSON(response)
	}

	if err := database.DB.Delete(&pelaku).Error; err != nil {
		response := helper.ResponseWithOutData{
			Code:    http.StatusInternalServerError,
			Status:  "error",
			Message: "Gagal Menghapus Data Pelaku",
		}
		return c.Status(http.StatusInternalServerError).JSON(response)
	}

	response := helper.ResponseWithOutData{
		Code:    http.StatusOK,
		Status:  "success",
		Message: "Berhasil Menghapus Data Pelaku",
	}
	return c.Status(http.StatusOK).JSON(response)
}