package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	"github.com/gin-gonic/gin"
)

// SearchMeterList godoc
// @Summary Get Meter Record List with mobile number and start date
// @Tags Meter Record
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/meter-record/list [get]
func SearchMeterList(c *gin.Context) {
	memberTel := c.Query("memberTel")

	params := models.SearchMeterRecordParams{
		MemberTel: memberTel,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.MeterRecordSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve meter record list: %v", err))
		return
	}

	utils.Success(c, "Meter record list retrieved successfully", result)
}

// SearchMeterJubuJibiSumEcoin godoc
// @Summary Get Meter Record Jubu Jibi
// @Tags Meter Record
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Param location query string false "Search location"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/meter-record/jubu-jibi [get]
func MeterRecordSumEcoinJubuJibi(c *gin.Context) {
	memberTel := c.Query("memberTel")
	location := c.Query("location")

	params := models.SearchPrizeCounterParams{
		MemberTel: memberTel,
		Location:  location,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.SumMeterRecordJubuJibi(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve meter record sum ecoin: %v", err))
		return
	}

	utils.Success(c, "Meter record sum ecoin retrieved successfully", result)
}

// SearchMeterJubuJibiList godoc
// @Summary Get Meter Record Jubu Jibi List
// @Tags Meter Record
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Param location query string false "Search location"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/meter-record/jubu-jibi/list [get]
func MeterRecordJubuJibiList(c *gin.Context) {
	memberTel := c.Query("memberTel")
	location := c.Query("location")

	params := models.SearchPrizeCounterParams{
		MemberTel: memberTel,
		Location:  location,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.MeterRecordJubuJibiList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve meter record jubu jibi list: %v", err))
		return
	}

	utils.Success(c, "Meter record jubu jibi retrieved successfully", result)
}

// SearchMeterList godoc
// @Summary Get Meter Record List with mobile number and start date
// @Tags Meter Record
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param cardNo query string false "Search card number"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/meter-record/check-card/list [get]
func CheckCardlist(c *gin.Context) {
	cardNo := c.Query("cardNo")

	params := models.SearchCheckCard{
		CardNo: cardNo,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.CheckCardList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve check card list: %v", err))
		return
	}

	utils.Success(c, "Check card list retrieved successfully", result)
}

// ExportCheckCard godoc
// @Summary Export check card list
// @Tags Meter Record
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param cardNo query string false "Search card number"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/meter-record/check-card/export [get]
func ExportCheckCardList(c *gin.Context) {
	cardNo := c.Query("cardNo")

	params := models.SearchCheckCard{
		CardNo: cardNo,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.CheckCardList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve check card list: %v", err))
		return
	}
	if len(result) == 0 {
		utils.Error(c, http.StatusNotFound, "No check card data found.")
		return
	}

	err = services.ExportCheckCardToExcel(c, result)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export check card list as Excel: %v", err))
		return
	}

}

// SearchMeterAllSpendList godoc
// @Summary Get Meter Record all spend List
// @Tags Meter Record
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/meter-record/all-spend/list [get]
func MeterRecordAllSpendList(c *gin.Context) {
	memberTel := c.Query("memberTel")

	params := models.SearchMeterParams{
		MemberTel: memberTel,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.MeterRecordAllSpend(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve meter record all apend list: %v", err))
		return
	}

	utils.Success(c, "Meter record all spend retrieved successfully", result)
}

// UpdateMeterAfterRefundRequest godoc
// @Summary Update Meter Record after refund request
// @Tags Meter Record
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/meter-record/prize-status/update [post]
func UpdateMeterAfterRefundRequest(c *gin.Context) {
	memberTel := c.Query("memberTel")

	params := models.SearchMeterParams{
		MemberTel: memberTel,
	}

	fmt.Printf("%+v\n", params)

	err := services.UpdateMeterRecordJubuJibi(memberTel)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to update meter record prize status: %v", err))
		return
	}

	utils.Success(c, "Meter record update successfully", "")
}
