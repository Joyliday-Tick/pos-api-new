package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	// "new-pos-api/middlewares"
	// "new-pos-api/models"
	"new-pos-api/middlewares"
	"new-pos-api/models"
	"new-pos-api/services"
	"new-pos-api/utils"

	// "strconv"

	"github.com/gin-gonic/gin"
	// "github.com/google/uuid"
)

// @Summary Create a new station
// @Tags Station
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PosStationDto true "Station Data"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/station [post]
func CreateStation(c *gin.Context) {
	var req models.PosStationDto

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	existBranchCode, err := services.FindExistStationCode(req.PosCode)
	if err == nil && existBranchCode != nil && existBranchCode.PosCode != "" {
		utils.Error(c, http.StatusBadRequest, "Station with this code already exists")
		return
	}

	if len(req.MacAddress) > 0 {
		findMacAddress, err := services.FindStationByMacAddressList(req.MacAddress)
		if err != nil {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve mac address station: %v", err))
			return
		}

		if findMacAddress != nil && findMacAddress.MacAddress != nil && len(findMacAddress.MacAddress) > 0 && findMacAddress.ID > 0 {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf(
				"Station with this MAC address already exists at id: %v, mac address: %s",
				findMacAddress.ID,
				strings.Join(utils.ToStringSlice(findMacAddress.MacAddress), ", "),
			))
			return
		}
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}
	fmt.Println("userId", userId)

	station, err := services.CreateStation(req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to create station: %v", err))
		return
	}

	utils.Success(c, "Station created successfully", station)
}

// UpdateStation godoc
// @Summary Update station by id
// @Tags Station
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body models.PosStationDto true "Station Data"
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/station/{id} [put]
func UpdateStation(c *gin.Context) {

	idStr := c.Param("id")
	stationID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid station id")
		return
	}

	var req models.PosStationDto
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// ตรวจสอบว่ามี station อยู่จริงหรือไม่
	findStation, err := services.FindStationById(stationID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findStation == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Station id: %s not found.", idStr))
		return
	}

	// ตรวจสอบว่ามี posCode ซ้ำกับ record อื่นหรือไม่
	existStation, err := services.FindExistStationCodeNotCurrent(req.PosCode, findStation.ID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if existStation != nil && existStation.PosCode != "" {
		utils.Error(c, http.StatusBadRequest, "Station with this code already exists in another record")
		return
	}

	if len(req.MacAddress) > 0 {
		// ตรวจสอบว่า mac address ซ้ำกับ record อื่นหรือไม่
		existStationMacAddress, err := services.FindStationByMacAddressListNotCurrent(req.MacAddress, findStation.ID)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, err.Error())
			return
		}

		if existStationMacAddress != nil && len(existStationMacAddress.MacAddress) > 0 && existStationMacAddress.ID != stationID {
			utils.Error(c, http.StatusBadRequest, fmt.Sprintf(
				"Station with this MAC address already exists at pos-id: %v, pos-code: %s, mac address: %s",
				existStationMacAddress.ID,
				existStationMacAddress.PosCode,
				strings.Join(utils.ToStringSlice(existStationMacAddress.MacAddress), ", "),
			))
			return
		}
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	// update station
	updateStation, err := services.UpdateStation(stationID, req, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to update station: %v", err))
		return
	}

	// // ถ้าเจอ record ที่มี MAC ซ้ำ → reset ของเดิม
	// if existStationMacAddress != nil {
	// 	if _, err = services.ResetMacAddressById(existStationMacAddress.ID, userId); err != nil {
	// 		utils.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to reset mac address for station id: %d, err: %v", existStationMacAddress.ID, err))
	// 		return
	// 	}
	// }

	utils.Success(c, "Station updated successfully", updateStation)
}

// GetStationBylocation godoc
// @Summary Get station by location
// @Tags Station
// @Security BasicAuth
// @Security BearerAuth
// @Param   location  path  string  true  "Location"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/station/location/{location} [get]
func GetStationByLocation(c *gin.Context) {
	location := c.Param("location")

	findLocation, err := services.FindStationByLocation(location)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve station: %v", err))
		return
	}

	utils.Success(c, "Station retrieved successfully", findLocation)
}

// GetStationByMacAddress godoc
// @Summary Get station by mac address
// @Tags Station
// @Security BasicAuth
// @Security BearerAuth
// @Param   macAddress  path  string  true  "MacAddress"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/station/mac-address/{macAddress} [get]
func GetStationByMacAddress(c *gin.Context) {
	macAddress := c.Param("macAddress")

	findMacAddress, err := services.FindStationByMacAddress(macAddress)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve station: %v", err))
		return
	}

	utils.Success(c, "Station retrieved successfully", findMacAddress)
}

// GetStationById godoc
// @Summary Get station by id
// @Tags Station
// @Security BasicAuth
// @Security BearerAuth
// @Param   id  path  string  true  "Id"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/station/{id} [get]
func GetStationById(c *gin.Context) {
	idStr := c.Param("id")
	dInt, err := strconv.Atoi(idStr)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findStation, err := services.FindStationById(dInt)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve station: %v", err))
		return
	}

	utils.Success(c, "Station retrieved successfully", findStation)
}

// SearchStationList godoc
// @Summary Get station list with paggination
// @Tags Station
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param search query string false "Search text code or description"
// @Param macAddress query string false "Search by mac address"
// @Param location query string false "Search by location"
// @Param page query int false "Page number"
// @Param skip query int false "Items per page"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/station/list [get]
func SearchStationList(c *gin.Context) {

	search := c.Query("search")
	location := c.Query("location")
	macAddress := c.Query("macAddress")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	skip, err := strconv.Atoi(c.DefaultQuery("skip", "10"))
	if err != nil || skip < 1 {
		skip = 10
	}

	params := models.SearchStationParams{
		Search:     search,
		Location:   location,
		MacAddress: macAddress,
		Page:       page,
		Skip:       skip,
	}
	result, err := services.StationSearchList(params)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve station: %v", err))
		return
	}

	utils.Success(c, "Station retrieved successfully", result)
}

// DeleteStationById godoc
// @Summary Delete station by id
// @Tags Station
// @Security BasicAuth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param   id  path  string  true  "Id"
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/station/{id} [delete]
func DeleteStationById(c *gin.Context) {
	idStr := c.Param("id")
	dInt, err := strconv.Atoi(idStr)
	if err != nil {
		// Handle error เช่น หาก id ไม่ใช่ตัวเลข
		fmt.Println("Invalid id")
		return
	}

	findStation, err := services.FindStationById(dInt)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	if findStation == nil {
		utils.Error(c, http.StatusNotFound, fmt.Sprintf("Station id: %s not found.", idStr))
		return
	}

	userId, err := middlewares.GetUserIdFromClaims(c)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Error extracting userId: %v", err))
		return
	}

	deleteStation, err := services.DeleteStationById(dInt, userId)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to delete station: %v", err))
		return
	}

	utils.Success(c, "Station deleted successfully", deleteStation)
}

// GetStationByMacAddressList godoc
// @Summary Get station by mac address list
// @Tags Station
// @Security BasicAuth
// @Security BearerAuth
// @Param body body models.MacAddressList true "Mac Address Array"
// @Accept json
// @Produce json
// @Success 200 {object} utils.StandardSuccessResponse
// @Failure 500 {object} utils.StandardErrorResponse
// @Router /api/station/mac-address-list [post]
func GetStationByMacAddressList(c *gin.Context) {
	var req models.MacAddressList

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	findMacAddress, err := services.FindStationByMacAddressList(req.MacAddressList)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, fmt.Sprintf("Failed to retrieve station: %v", err))
		return
	}

	utils.Success(c, "Station retrieved successfully", findMacAddress)
}
