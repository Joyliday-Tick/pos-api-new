package controllers

import (
	"fmt"
	"net/http"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SearchTaxInvoiceList godoc
// @Summary Get pos tax invoice list with paggination
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search branch"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/tax-invoice/list [get]
func SearchTaxInvoiceList(c *gin.Context) {

	location := c.Query("location")
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

	params := models.SearchTaxInvoiceReport{
		Location:  location,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Skip:      skip,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.TaxInvoiceReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve tax invoice list: %v", err))

		return
	}

	utils.Success(c, "Tax invoice list retrieved successfully", result)
}

// ExportTaxInvoice godoc
// @Summary Export pos tax invoice list
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search branch"
// @Param exportFormat query string false "Export form (pdf, excel)"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/tax-invoice/export [get]
func ExportTaxInvoiceList(c *gin.Context) {
	location := c.Query("location")
	exportFormat := c.Query("exportFormat")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	params := models.SearchExportTaxInvoiceReport{
		Location:     location,
		StartDate:    startDate,
		EndDate:      endDate,
		ExportFormat: exportFormat,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.ExportTaxInvoiceReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve tax invoice list: %v", err))
		return
	}
	if len(result) == 0 {
		utils.Error(c, http.StatusNotFound, "No tax invoice data found.")
		return
	}

	if exportFormat == "excel" {
		err = services.ExportTaxInvoiceToExcel(c, result, params)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export tax invoice as Excel: %v", err))
		}
		return
	} else {
		err = services.ExportToPDF(c, result, params)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export tax invoice as PDF: %v", err))
		}
		return
	}
}

// SearchStampHouseList godoc
// @Summary Get stamp house list with paggination
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search branch"
// @Param machine query string false "Search stamp machine code"
// @Param memberTel query string false "Search member tel"
// @Param startDate query string false "Search date from start date"
// @Param endDate query string false "Search date from end date"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/stamp-house/list [get]
func StampHouseListReport(c *gin.Context) {

	location := c.Query("location")
	machine := c.Query("machine")
	memberTel := c.Query("memberTel")
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

	params := models.SearchStampHouseReport{
		Location:  location,
		Machine:   machine,
		MemberTel: memberTel,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Skip:      skip,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.StampHouseListReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve stamp house list: %v", err))
		return
	}

	utils.Success(c, "Stamp house list retrieved successfully", result)
}

// ExportStampHouse godoc
// @Summary Export stamp point list
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search branch"
// @Param machine query string false "Search stamp machine code"
// @Param memberTel query string false "Search member tel"
// @Param startDate query string false "Search date from start date"
// @Param endDate query string false "Search date from end date"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/stamp-house/export [get]
func ExportStampHouseList(c *gin.Context) {
	location := c.Query("location")
	machine := c.Query("machine")
	memberTel := c.Query("memberTel")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	params := models.SearchExportStampHouseReport{
		Location:  location,
		Machine:   machine,
		MemberTel: memberTel,
		StartDate: startDate,
		EndDate:   endDate,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.ExportStampHouseListReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve stamp house list: %v", err))
		return
	}
	if len(result) == 0 {
		utils.Error(c, http.StatusNotFound, "No stamp house data found.")
		return
	}

	err = services.ExportStampHouseToExcel(c, result, params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export stamp house list as Excel: %v", err))
		return
	}

}

// AccumSpending godoc
// @Summary Get pos transaction accumulated spending by member tel
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/spending/list [get]
func AccumSpendingList(c *gin.Context) {

	memberTel := c.Query("memberTel")
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

	params := models.SearchSpendingParams{
		MemberTel: memberTel,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Skip:      skip,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.SpendingSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve spending list")
		return
	}

	utils.Success(c, "Spending list retrieved successfully", result)
}

// ExportSpendingList godoc
// @Summary Export spending list
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Param startDate query string false "Search date from start date"
// @Param endDate query string false "Search date from end date"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/spending/export [get]
func ExportSpendingList(c *gin.Context) {
	memberTel := c.Query("memberTel")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	params := models.ExportSpendingParams{
		MemberTel: memberTel,
		StartDate: startDate,
		EndDate:   endDate,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.ExportSpendingList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve spending list: %v", err))
		return
	}
	if len(result) == 0 {
		utils.Error(c, http.StatusNotFound, "No spending data found.")
		return
	}

	err = services.ExportSpendingToExcel(c, result)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export spending list as Excel: %v", err))
		return
	}

}

// AdjustPoint godoc
// @Get adjust point list
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Param location query string false "Search location"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/adjust-point/list [get]
func AdjustPointList(c *gin.Context) {

	memberTel := c.Query("memberTel")
	location := c.Query("location")
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

	params := models.SearchAdjustPointReport{
		MemberTel: memberTel,
		Location:  location,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Skip:      skip,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.AdjustPointReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve adjust point list")
		return
	}

	utils.Success(c, "Adjust point retrieved successfully", result)
}

// ExportAdjustPoint godoc
// @Summary Export adjust point list
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param memberTel query string false "Search member tel"
// @Param location query string false "Search branch"
// @Param exportFormat query string false "Export form (pdf, excel)"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/adjust-point/export [get]
func ExportAdjustPointList(c *gin.Context) {
	location := c.Query("location")
	memberTel := c.Query("memberTel")
	exportFormat := c.Query("exportFormat")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	params := models.ExportAdjustPointReport{
		MemberTel:    memberTel,
		Location:     location,
		StartDate:    startDate,
		EndDate:      endDate,
		ExportFormat: exportFormat,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.ExportAdjustPointReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve adjust point list: %v", err))
		return
	}
	if len(result) == 0 {
		utils.Error(c, http.StatusNotFound, "No adjust point data found.")
		return
	}

	if exportFormat == "excel" {
		err = services.AdjExportToExcel(c, result, params)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export adjust point as Excel: %v", err))
		}
		return
	} else {
		err = services.AdjExportToPDF(c, result, params)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export adjust point as PDF: %v", err))
		}
		return
	}
}

// SearchPosVoidList godoc
// @Summary Get pos void list with paggination
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search branch"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/void/list [get]
func SearchPosVoidList(c *gin.Context) {

	location := c.Query("location")
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

	params := models.SearchTaxInvoiceReport{
		Location:  location,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Skip:      skip,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.PosVoidSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve pos void list: %v", err))

		return
	}

	utils.Success(c, "Pos void list retrieved successfully", result)
}

// SearchJoylicoinList godoc
// @Summary Get claim joylicoin report
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.SearchClaimReport true "Search parameters"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/claim/list [post]
func SearchClaimList(c *gin.Context) {

	var req models.SearchClaimReport

	// Read JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Default page
	if req.Page < 1 {
		req.Page = 1
	}

	// Default skip
	if req.Skip < 1 {
		req.Skip = 10
	}

	fmt.Printf("Request: %+v\n", req)

	result, err := services.ReturnBonusSearchList(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve claim list: %v", err))
		return
	}

	utils.Success(c, "Claim list retrieved successfully", result)
}

// ExportClaimList godoc
// @Summary Export claim list
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ExportClaimReport true "Search parameters"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/claim/export [post]
func ExportClaimList(c *gin.Context) {
	var req models.ExportClaimReport

	// Read JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := services.ExportReturnBonusSearchList(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve claim list: %v", err))
		return
	}
	if len(result) == 0 {
		utils.Error(c, http.StatusNotFound, "No claim data found.")
		return
	}

	err = services.ExportClaimToExcel(c, result)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export claim list as Excel: %v", err))
		return
	}

}

// SearchClaimDetail godoc
// @Summary Get claim detail
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.SearchClaimDetail true "Search parameters"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/claim/detail [post]
func SearchClaimDetail(c *gin.Context) {

	var req models.SearchClaimDetail

	// Read JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	fmt.Printf("Request: %+v\n", req)

	result, err := services.MeterRecordClaimDetail(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve claim detail: %v", err))
		return
	}

	utils.Success(c, "Claim detail retrieved successfully", result)
}

// SearchClearCardList godoc
// @Summary Get clear card list with paggination
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search branch"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/clear-card/list [get]
func SearchClearCardList(c *gin.Context) {

	location := c.Query("location")
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

	params := models.SearchTaxInvoiceReport{
		Location:  location,
		StartDate: startDate,
		EndDate:   endDate,
		Page:      page,
		Skip:      skip,
	}

	fmt.Printf("%+v\n", params)

	result, err := services.PosClearCardSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve clear card list: %v", err))

		return
	}

	utils.Success(c, "Clear card list retrieved successfully", result)
}

// ExportClearCardList godoc
// @Summary Export clear card list
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ExportClearCardReport true "Search parameters"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/clear-card/export [post]
func ExportClearCardList(c *gin.Context) {
	var req models.ExportClearCardReport

	// Read JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := services.ExportPosClearCardSearchList(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve clear card list: %v", err))
		return
	}
	if len(result) == 0 {
		utils.Error(c, http.StatusNotFound, "No clear card data found.")
		return
	}

	err = services.ExportClearCardToExcel(c, result)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export clear card list as Excel: %v", err))
		return
	}

}

// PosDailySalesSummary godoc
// @Get POS daily sales summary report
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search location"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/pos-daily/summary [get]
func PosDailySalesSummary(c *gin.Context) {

	location := c.Query("location")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	params := models.SearchPosSaleReport{
		Location:  location,
		StartDate: startDate,
		EndDate:   endDate,
	}

	result, err := services.GroupPosDailySalesSummaryReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve POS daily sales summary report")
		return
	}

	utils.Success(c, "POS daily sales summary report retrieved successfully", result)
}

// ExportPosDailySalesSummary godoc
// @Summary Export POS daily sales summary
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search location"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/pos-daily/summary/export [get]
func ExportPosDailySalesSummary(c *gin.Context) {

	location := c.Query("location")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	params := models.SearchPosSaleReport{
		Location:  location,
		StartDate: startDate,
		EndDate:   endDate,
	}

	branch, err := services.FindExistBranchCode(location)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve branch information")
		return
	}

	branchName := ""

	if branch != nil {
		branchName = branch.BranchName
	}

	result, err := services.GroupPosDailySalesSummaryReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve POS daily sales summary report")
		return
	}
	if len(result.Data) == 0 {
		utils.Error(c, http.StatusNotFound, "No POS daily sales summary data found.")
		return
	}

	err = services.ExportPOSDailyToPDF(c, result, params, branchName)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export POS daily sales summary as PDF: %v", err))
		return
	}

}

// PosMonthlySalesSummary godoc
// @Get POS monthly sales summary report
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search location"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/pos-monthly/summary [get]
func PosMonthlySalesSummary(c *gin.Context) {

	location := c.Query("location")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	params := models.SearchPosSaleReport{
		Location:  location,
		StartDate: startDate,
		EndDate:   endDate,
	}

	result, err := services.GroupPosMonthlySalesSummaryReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve POS monthly sales summary report")
		return
	}

	utils.Success(c, "POS monthly sales summary report retrieved successfully", result)
}

// PosSaleReportByPayment godoc
// @Get POS monthly sales summary report
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search location"
// @Param paymentId query string false "Search payment id"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/pos-sale/payment [get]
func PosSaleReportByPayment(c *gin.Context) {

	location := c.Query("location")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	paymentId := c.Query("paymentId")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	skip, err := strconv.Atoi(c.DefaultQuery("skip", "10"))
	if err != nil || skip < 1 {
		skip = 10
	}

	params := models.SearchPosSalePaymentReport{
		Location:      location,
		StartDate:     startDate,
		EndDate:       endDate,
		PaymentTypeId: paymentId,
		Page:          page,
		Skip:          skip,
	}

	result, err := services.PosSalePaymentReport(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve POS sale payment report")
		return
	}

	utils.Success(c, "POS sale payment report retrieved successfully", result)
}

// ExportPosSaleReportByPayment godoc
// @Summary Export POS sale report by payment
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ExportPosSalePaymentReport true "Search parameters"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/pos-sale/payment/export [post]
func ExportPosSaleReportByPayment(c *gin.Context) {
	var req models.ExportPosSalePaymentReport

	// Read JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := services.ExportPosSalePaymentReport(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve POS sale payment report: %v", err))
		return
	}
	if len(result) == 0 {
		utils.Error(c, http.StatusNotFound, "No POS sale payment data found.")
		return
	}

	err = services.ExportPosSalePaymentToExcel(c, result)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export POS sale payment report as Excel: %v", err))
		return
	}

}

// ClaimPrizeReport godoc
// @Get Cliam prize report
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param location query string false "Search location"
// @Param memberTel query string false "Search member tel"
// @Param startDate query string false "Search bill date from start date"
// @Param endDate query string false "Search bill date from end date"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/claim-prize/list [get]
func ClaimPrizeList(c *gin.Context) {

	location := c.Query("location")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	memberTel := c.Query("memberTel")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	skip, err := strconv.Atoi(c.DefaultQuery("skip", "10"))
	if err != nil || skip < 1 {
		skip = 10
	}

	params := models.SearchClaimPrizeParams{
		Location:  location,
		StartDate: startDate,
		EndDate:   endDate,
		MemberTel: memberTel,
		Page:      page,
		Skip:      skip,
	}

	result, err := services.ClaimPrizeSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to retrieve claim prize report")
		return
	}

	utils.Success(c, "Claim prize report retrieved successfully", result)
}

// ExportClaimPrize godoc
// @Summary Export cliam prize
// @Tags Report
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.ExportClaimPrizeParams true "Search parameters"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/report/claim-prize/export [post]
func ExportClaimPrize(c *gin.Context) {
	var req models.ExportClaimPrizeParams

	// Read JSON body
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := services.ExportClaimPrizeSearch(req)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve claim prize report: %v", err))
		return
	}
	if len(result) == 0 {
		utils.Error(c, http.StatusNotFound, "No Claim prize data found.")
		return
	}

	err = services.ExportClaimPrizeSearchToExcel(c, result)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to export claim prize report as Excel: %v", err))
		return
	}

}
