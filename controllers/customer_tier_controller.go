package controllers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"

	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
)

// ImportCustomerTierFirst godoc
// @Summary Import customer tier first time
// @Description Import customer tier from CSV or XLSX
// @Security BasicAuth
// @Security BearerAuth
// @Tags Customer Tier
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV or XLSX file"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/customer-tier/import/first [post]
func ImportCustomerTierFirst(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open file"})
		return
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	var items []models.CustomerTier
	now := utils.TimeNowAsia()
	telMap := make(map[string]bool)

	switch ext {

	case ".csv":
		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		if err != nil {
			utils.Error(c, http.StatusBadRequest, "invalid csv")
			return
		}

		for i, row := range records {
			if i == 0 || len(row) < 2 {
				continue
			}

			tel := strings.TrimSpace(row[0])
			if telMap[tel] {
				continue
			}
			telMap[tel] = true

			items = append(items, models.CustomerTier{
				Tel:        tel,
				Tier:       strings.ToUpper(strings.TrimSpace(row[1])),
				CreateDate: *now,
			})
		}

	case ".xlsx":
		excel, err := excelize.OpenReader(f)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, "invalid xlsx")
			return
		}

		sheet := excel.GetSheetName(0)
		rows, _ := excel.GetRows(sheet)

		for i, row := range rows {
			if i == 0 || len(row) < 2 {
				continue
			}

			tel := strings.TrimSpace(row[0])
			if telMap[tel] {
				continue
			}
			telMap[tel] = true

			items = append(items, models.CustomerTier{
				Tel:        tel,
				Tier:       strings.ToUpper(strings.TrimSpace(row[1])),
				CreateDate: *now,
			})
		}

	default:
		utils.Error(c, http.StatusBadRequest, "only csv or xlsx allowed")
		return
	}

	if len(items) == 0 {
		utils.Error(c, http.StatusBadRequest, "no data")
		return
	}

	if err := services.CreateBatchCustomerTierTx(items, 500); err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to import customer tier: %v", err))
		return
	}

	utils.Success(c, "Customer tier imported successfully", gin.H{
		"message": "import success",
		"total":   len(items),
	})
}

// ImportCustomerTier godoc
// @Summary Import customer tier
// @Description Import customer tier from CSV or XLSX
// @Security BasicAuth
// @Security BearerAuth
// @Tags Customer Tier
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "CSV or XLSX file"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/customer-tier/import [post]
func ImportCustomerTier(c *gin.Context) {
	fmt.Println("ImportCustomerTier")

	file, err := c.FormFile("file")
	if err != nil {
		utils.Error(c, 400, "file is required")
		return
	}

	f, err := file.Open()
	if err != nil {
		utils.Error(c, 500, "cannot open file")
		return
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	var dtos []models.CustomerTierImportDTO

	switch ext {

	// ================= CSV =================
	case ".csv":
		records, err := csv.NewReader(f).ReadAll()
		if err != nil {
			utils.Error(c, 400, "invalid csv")
			return
		}

		telMap := map[string]bool{}

		for i, row := range records {
			if i == 0 || len(row) < 2 {
				continue
			}

			tel := strings.TrimSpace(row[0])
			if tel == "" || telMap[tel] {
				continue
			}
			telMap[tel] = true

			dtos = append(dtos, models.CustomerTierImportDTO{
				Tel:  tel,
				Tier: strings.ToUpper(strings.TrimSpace(row[1])),
			})
		}

	// ================= XLSX =================
	case ".xlsx":
		excel, err := excelize.OpenReader(f)
		if err != nil {
			utils.Error(c, 400, "invalid xlsx")
			return
		}

		sheet := excel.GetSheetName(0)
		rows, err := excel.GetRows(sheet)
		if err != nil {
			utils.Error(c, 400, "cannot read sheet")
			return
		}

		telMap := map[string]bool{}

		for i, row := range rows {
			if i == 0 || len(row) < 2 {
				continue
			}

			tel := strings.TrimSpace(row[0])
			if tel == "" || telMap[tel] {
				continue
			}
			telMap[tel] = true

			dtos = append(dtos, models.CustomerTierImportDTO{
				Tel:  tel,
				Tier: strings.ToUpper(strings.TrimSpace(row[1])),
			})
		}

	default:
		utils.Error(c, 400, "only csv or xlsx allowed")
		return
	}

	if len(dtos) == 0 {
		utils.Error(c, 400, "no data")
		return
	}

	if err := services.ImportCustomerTier(dtos); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, "import success", gin.H{
		"total": len(dtos),
	})
}

// GetCustomerTierByTel godoc
// @Summary Get customer tier by tel
// @Tags Customer Tier
// @Security BasicAuth
// @Security BearerAuth
// @Param   tel  path  string  true  "Tel"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/customer-tier/{tel} [get]
func GetCustomerTierByTel(c *gin.Context) {
	tel := c.Param("tel")

	findCustomerTier, err := services.FindExistCustomerTier(tel)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve card type: %v", err))
		return
	}

	utils.Success(c, "Customer tier retrieved successfully", findCustomerTier)
}
