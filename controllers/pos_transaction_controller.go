package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary Create a new pos transaction
// @Tags POS Transaction
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PosTransactionDto true "Pos Transaction Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-transaction [post]
func CreatePosTransaction(c *gin.Context) {
	var req models.PosTransactionDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusUnauthorized, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	posTransaction, err := services.CreatePosTransaction(req, userId)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create pos transaction: %v", err))
		return
	}

	fmt.Println("transaction", posTransaction)
	utils.Success(c, "POS transaction created successfully", posTransaction)
}

// SearchPosTransactionList godoc
// @Summary Get pos transaction list with paggination
// @Tags POS Transaction
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param billNo query string false "Search Bill number"
// @Param memberTel query string false "Search member tel"
// @Param cardNo query string false "Search card number"
// @Param posId query string false "Search pos branch"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-transaction/list [get]
func SearchPostransactionList(c *gin.Context) {

	billNo := c.Query("billNo")
	memberTel := c.Query("memberTel")
	cardNo := c.Query("cardNo")
	posId := c.Query("posId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	skip, err := strconv.Atoi(c.DefaultQuery("skip", "10"))
	if err != nil || skip < 1 {
		skip = 10
	}

	params := models.SearchPosTransactionParams{
		BillNo:    billNo,
		MemberTel: memberTel,
		CardNo:    cardNo,
		POSID:     posId,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Skip:      skip,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.PosTransactionSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve pos transaction list")
		return
	}

	utils.Success(c, "Pos transaction list retrieved successfully", result)
}

// SearchPosSaleReport godoc
// @Summary Get pos sale report
// @Tags POS Transaction
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search branch"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-transaction/sale-report [get]
func PosSaleReport(c *gin.Context) {

	location := c.Query("location")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	params := models.SearchPosSaleReport{
		Location:  location,
		StartDate: startDate,
		EndDate:   endDate,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.PosSaleReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve pos sale report")
		return
	}

	utils.Success(c, "Pos sale report retrieved successfully", result)
}

// GetBillNoByPosId godoc
// @Summary Get bill no by pos id
// @Tags POS Transaction
// @Security BasicAuth
// @Security BearerAuth
// @Param   pos_id  path  string  true  "POS ID"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/pos-transaction/gen-bill/{pos_id} [get]
func GenerateBillNo(c *gin.Context) {
	posId := c.Param("pos_id")

	billNo := services.GenerateBillNo(posId)

	utils.Success(c, "Generate bill no successfully", billNo)
}
