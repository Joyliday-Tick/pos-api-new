package routes

import (
	// "log"
	"new-pos-api/controllers"
	v2_controllers "new-pos-api/controllers/v2"

	"new-pos-api/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine) {
	// r.POST("auth/login", controllers.Login)
	api := r.Group("/api")
	api.Use(middlewares.AuthMiddleware()) // ดักทุก /api/*

	//*user*//
	api.GET("/users", controllers.GetUsers)
	api.GET("/users/:username", controllers.GetUsersByUsername)
	api.GET("/user/list", controllers.SearchUserList)
	api.POST("/user", middlewares.RequireRole(middlewares.RoleAdministrator), controllers.CreateUser)
	api.PUT("/user/:id", middlewares.RequireRole(middlewares.RoleAdministrator), controllers.UpdateUser)
	api.GET("/user/:id", controllers.GetUserById)
	api.DELETE("/user/:id", middlewares.RequireRole(middlewares.RoleAdministrator), controllers.DeleteUserById)

	api.POST("/auth/login", controllers.Login)

	//*card type *//
	api.GET("/card-type", controllers.GetCardTypeList)
	api.GET("/card-type/list", controllers.SearchCardTypeList)
	api.POST("/card-type", controllers.CreateCardType)
	api.PUT("/card-type/:id", controllers.UpdateCardType)
	api.GET("/card-type/:id", controllers.GetCardTypeById)
	api.DELETE("/card-type/:id", controllers.DeleteCardTypeById)

	//*play type *//
	api.GET("/play-type", controllers.GetPlayTypeList)
	api.GET("/play-type/list", controllers.SearchPlayTypeList)
	api.POST("/play-type", controllers.CreatePlayType)
	api.PUT("/play-type/:id", controllers.UpdatePlayType)
	api.GET("/play-type/:id", controllers.GetPlayTypeById)
	api.DELETE("/play-type/:id", controllers.DeletePlayTypeById)

	//*card entity *//
	// GET
	api.GET("card/check/info/:cardNo", controllers.CheckCardDetailByCardNo)
	api.GET("card/check/package-info/:cardNo", controllers.CheckCardPackageInfoByCardNo)
	api.GET("card/check/package/:cardNo", controllers.CheckCardPackageByCardNo)
	api.GET("card/check/card-type/:cardNo", controllers.CheckCardTypeByCardNo)
	api.GET("card/check/:cardNo", controllers.CheckCardByCardNo)
	api.GET("card/tel/:tel", controllers.CheckCardByTel)
	api.GET("card/refund/:tel", controllers.CheckCardRefundByTel)
	api.GET("card/member", controllers.CheckCardMemberByTel)
	api.GET("card/:memberTel/:cardNo", controllers.GetCardActiveByTelAndCardNo) // 2 params, specific
	api.GET("/card/:memberTel", controllers.GetCardEntityByMembertel)           // 1 param, general

	// POST
	api.POST("/card/lock-card", controllers.LockCard)
	api.POST("/card/unlock-card", controllers.UnLockCard)
	api.POST("/card/register", controllers.RegisterCardEntity)
	api.POST("/card/topup-pos", controllers.TopupCardPOS)
	api.POST("/card/clear-card", controllers.ClearCard)
	api.POST("/card/transfer", controllers.TransferCard)

	// DELETE
	api.DELETE("card/:cardNo", middlewares.RequireRole(middlewares.RoleAdministrator, middlewares.RoleRMBackoffice, middlewares.RoleManager, middlewares.RoleAssistManager), controllers.DeleteCardEntityPermanent)

	//*card play *//
	api.POST("/card-play", controllers.CreateCardPlay)
	api.POST("/card-play-verify", controllers.VerifyCardPlay)
	api.POST("/card-play-deduct", controllers.DeductCardPlay)

	//*pos menu *//
	api.GET("/pos-menu/location/:location", controllers.GetPosMenuByLocation)
	api.GET("/pos-menu/group/:groupMenuId", controllers.GetPosMenuByGroupMenuId)
	api.GET("/pos-menu/sale/list", controllers.SalePosMenuList)
	api.GET("/pos-menu/list", controllers.SearchPosMenuList)
	api.GET("/pos-menu/sale-all", controllers.GetPosMenuSaleAll)
	api.GET("/pos-menu/:id", controllers.GetPosMenuById)
	api.POST("/pos-menu", controllers.CreatePosMenu)
	api.PUT("/pos-menu/:id", controllers.UpdatePosMenu)
	api.DELETE("/pos-menu/:id", controllers.DeletePosMenuById)

	//*group menu *//
	api.GET("/group-menu", controllers.GetGroupMenuList)
	api.GET("/group-menu/list", controllers.SearchGroupMenuList)
	api.POST("/group-menu", controllers.CreateGroupMenu)
	api.PUT("/group-menu/:id", controllers.UpdateGroupMenu)
	api.GET("/group-menu/:id", controllers.GetGroupMenuById)
	api.DELETE("/group-menu/:id", controllers.DeleteGroupMenuById)

	//*branch group*//
	api.GET("/branch-group", controllers.GetBranchGroupList)
	api.GET("/branch-group/list", controllers.SearchBranchGroupList)
	api.POST("/branch-group", controllers.CreateBranchGroup)
	api.PUT("/branch-group/:id", controllers.UpdateBranchGroup)
	api.GET("/branch-group/:id", controllers.GetBranchGroupById)
	api.DELETE("/branch-group/:id", controllers.DeleteBranchGroupById)

	//*branch*//
	api.GET("/branch", controllers.GetBranchList)
	api.GET("/branch/list", controllers.SearchBranchList)
	api.POST("/branch", controllers.CreateBranch)
	api.PUT("/branch/:code", controllers.UpdateBranch)
	api.GET("/branch/:code", controllers.GetBranchByCode)
	api.DELETE("/branch/:code", controllers.DeleteBranchByCode)

	//*station*//
	api.GET("/station/mac-address/:macAddress", controllers.GetStationByMacAddress)
	api.POST("/station/mac-address-list", controllers.GetStationByMacAddressList)
	api.GET("/station/location/:location", controllers.GetStationByLocation)
	api.GET("/station/:id", controllers.GetStationById)
	api.GET("/station/list", controllers.SearchStationList)
	api.POST("/station", controllers.CreateStation)
	api.PUT("/station/:id", controllers.UpdateStation)
	api.DELETE("/station/:id", controllers.DeleteStationById)

	//*member*//
	api.POST("/member/name", controllers.UpdateMemberName)
	api.POST("/member/deduct-joylicoin/:tel", controllers.UpdateJoylicoin)
	api.POST("/member/deduct-totalpoint/:tel", controllers.UpdateTotalPoint)
	api.GET("/member/search-sync/:tel", controllers.GetMemberByTelAndUpdate)
	api.GET("/member/:tel", controllers.GetMemberByTel)
	api.POST("/member/:tel", controllers.UpdateSkill)
	api.POST("/member", controllers.CreateMember)
	api.PUT("/member", controllers.UpdateMember)

	//*master payment*//
	api.GET("/master-payment", controllers.GetMasterPaymentList)

	//*pos transaction*//
	api.POST("/pos-transaction", controllers.CreatePosTransaction)
	api.GET("/pos-transaction/gen-bill/:pos_id", controllers.GenerateBillNo)
	api.GET("/pos-transaction/list", controllers.SearchPostransactionList)
	api.GET("/pos-transaction/sale-report", controllers.PosSaleReport)

	//*pos sub transaction*//
	api.GET("/pos-sub-transaction/:billNo", controllers.GetPosSubTransactionByBillNo)

	//*pos void*//
	api.POST("/pos-void", middlewares.RequireRole(middlewares.RoleAdministrator, middlewares.RoleRMBackoffice, middlewares.RoleManager, middlewares.RoleAssistManager), controllers.CreatePosVoid)

	//*bonus setting*//
	api.GET("/bonus-setting", controllers.GetBonusSetting)

	//*metert record*//
	api.GET("/meter-record/list", controllers.SearchMeterList)
	api.GET("/meter-record/jubu-jibi", controllers.MeterRecordSumEcoinJubuJibi)
	api.GET("/meter-record/jubu-jibi/list", controllers.MeterRecordJubuJibiList)
	api.GET("/meter-record/all-spend/list", controllers.MeterRecordAllSpendList)
	api.GET("/meter-record/check-card/list", controllers.CheckCardlist)
	api.GET("/meter-record/check-card/export", controllers.ExportCheckCardList)
	api.POST("/meter-record/prize-status/update", controllers.UpdateMeterAfterRefundRequest)

	//*prize counter*//
	api.GET("/prize-counter/jubu-jibi", controllers.SearchPrizeCounterJubuJibi)

	//*stamp house*//
	api.GET("/stamp-house/list", controllers.SearchStampPointList)

	//*claim*//
	api.POST("/claim", controllers.NewClaimJoylicoin)

	//*crm*//
	api.GET("/crm/customer/:tel", controllers.GetCustomerByTel)
	api.GET("/crm/branch/:branchCode", controllers.GetBranchCrmByCode)
	api.GET("/crm/score-type", controllers.GetScoreType)

	//*user group*//
	api.GET("/user-group", controllers.GetUserGroupList)
	api.GET("/user-group/list", controllers.SearchUserGroupList)
	api.POST("/user-group", controllers.CreateUserGroup)
	api.PUT("/user-group/:id", controllers.UpdateUserGroup)
	api.GET("/user-group/:id", controllers.GetUserGroupById)
	api.DELETE("/user-group/:id", controllers.DeleteUserGroupById)

	//*user role*//
	api.GET("/user-role", controllers.GetUserRoleList)
	api.GET("/user-role/list", controllers.SearchUserRoleList)
	api.POST("/user-role", middlewares.RequireRole(middlewares.RoleAdministrator), controllers.CreateUserRole)
	api.PUT("/user-role/:id", middlewares.RequireRole(middlewares.RoleAdministrator), controllers.UpdateUserRole)
	api.GET("/user-role/:id", controllers.GetUserRoleById)
	api.DELETE("/user-role/:id", middlewares.RequireRole(middlewares.RoleAdministrator), controllers.DeleteUserRoleById)

	//*master logo*//
	api.GET("/logo", controllers.GetLogoList)

	//*report*//
	api.GET("/report/void/list", controllers.SearchPosVoidList)
	api.GET("/report/clear-card/list", controllers.SearchClearCardList)
	api.GET("/report/tax-invoice/list", controllers.SearchTaxInvoiceList)
	api.GET("/report/tax-invoice/export", controllers.ExportTaxInvoiceList)
	api.GET("/report/stamp-house/list", controllers.StampHouseListReport)
	api.GET("/report/stamp-house/export", controllers.ExportStampHouseList)
	api.GET("/report/spending/list", controllers.AccumSpendingList)
	api.GET("/report/spending/export", controllers.ExportSpendingList)
	api.GET("/report/adjust-point/list", controllers.AdjustPointList)
	api.GET("/report/adjust-point/export", controllers.ExportAdjustPointList)
	api.GET("/report/pos-daily/summary/export", controllers.ExportPosDailySalesSummary)
	api.GET("/report/pos-daily/summary", controllers.PosDailySalesSummary)
	api.GET("/report/pos-monthly/summary", controllers.PosMonthlySalesSummary)
	api.GET("/report/pos-sale/payment", controllers.PosSaleReportByPayment)
	api.GET("/report/claim-prize/list", controllers.ClaimPrizeList)
	api.POST("/report/claim/list", controllers.SearchClaimList)
	api.POST("/report/claim/export", controllers.ExportClaimList)
	api.POST("/report/claim/detail", controllers.SearchClaimDetail)
	api.POST("/report/clear-card/export", controllers.ExportClearCardList)
	api.POST("/report/claim-prize/export", controllers.ExportClaimPrize)
	api.POST("/report/pos-sale/payment/export", controllers.ExportPosSaleReportByPayment)

	//*jubu jibi*//
	api.POST("/jubu-jibi/deposit", controllers.DepositJubuJibi)
	api.POST("/jubu-jibi/redeem", controllers.RedeemJubuJibi)

	//*stamp machine*//
	api.GET("/stamp-machine", controllers.GetStampMachineList)

	//*adjust point*//
	api.POST("/adjust-point", middlewares.RequireRole(middlewares.RoleAdministrator, middlewares.RoleRMBackoffice, middlewares.RoleManager, middlewares.RoleAssistManager), controllers.CreateAdjustPoint)

	//*machine group*//
	api.GET("/machine-group/v1/:id", controllers.GetMachineGroupByIdV1)
	api.GET("/machine-group/sub/:groupId", controllers.GetMachineSubGroupByGroupId)
	api.GET("/machine-group/list", controllers.SearchMachineGroupList)

	api.GET("/machine-group/:id", controllers.GetMachineGroupById)
	api.GET("/machine-group", controllers.GetMachineGroupList)

	api.POST("/machine-group", controllers.CreateMachineGroup)
	api.PUT("/machine-group/:id", controllers.UpdateMachineGroup)
	api.DELETE("/machine-group/:id", controllers.DeleteMachineGroupById)

	//*refund request*//
	api.GET("/refund-request/:tel", controllers.GetRefundRequestByTel)
	api.POST("/refund-request", controllers.CreateRefundRequest)

	//*deposit cron*//
	api.GET("/deposit-cron/:tel", controllers.GetDepositCronByTel)

	//*refund confirm*//
	api.POST("/refund-confirm", middlewares.RequireRole(middlewares.RoleAdministrator, middlewares.RoleRMBackoffice, middlewares.RoleManager, middlewares.RoleAssistManager), controllers.CreateRefundConfirm)

	//*master action*//
	api.GET("/master-action", controllers.GetMasterActionList)

	//*claim prize*//
	api.POST("/claim-prize", controllers.CreateClaimPrize)

	//*discount cash*//
	api.GET("/discountcash/:member_tel", controllers.GetDiscountCashByMemberTel)

	//*customer tier*//
	api.GET("/customer-tier/:tel", controllers.GetCustomerTierByTel)
	api.POST("/customer-tier/import/first", controllers.ImportCustomerTierFirst)
	api.POST("/customer-tier/import", controllers.ImportCustomerTier)

	v2 := api.Group("/v2")
	{
		v2.POST("/verify-card", v2_controllers.VerifyCard)
		v2.POST("/verify-card-smc", v2_controllers.VerifyCardSMC)
	}

}
