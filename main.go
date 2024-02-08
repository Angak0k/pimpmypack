package main

import (
	"log"

	_ "github.com/Angak0k/pimpmypack/docs"
	"github.com/Angak0k/pimpmypack/pkg/accounts"
	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/inventories"
	"github.com/Angak0k/pimpmypack/pkg/packs"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func init() {

	// init env
	err := config.EnvInit(".env")
	if err != nil {
		log.Fatalf("Error loading .env file or environment variable : %v", err)
	}
	println("Environment variables loaded")

	// init DB
	err = database.DatabaseInit()
	if err != nil {
		log.Fatalf("Error connecting database : %v", err)
	}
	println("Database connected")

	// init DB migration
	err = database.DatabaseMigrate()
	if err != nil {
		log.Fatalf("Error migrating database : %v", err)
	}
	println("Database migrated")

}

// @title PimpMyPack API
// @description API server to manage Backpack Inventory and Packing Lists
// @version 1.0
// @host pmp-dev.alki.earth
// @Schemes https
// @BasePath /api
func main() {

	if config.Stage == "DEV" || config.Stage == "LOCAL" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	public := router.Group("/api")
	public.POST("/register", accounts.Register)
	public.POST("/login", accounts.Login)
	public.GET("/confirmemail", accounts.ConfirmEmail)
	public.POST("/forgotpassword", accounts.ForgotPassword)
	public.GET("/sharedlist/:sharing_code", packs.SharedList)

	protected := router.Group("/api/v1")
	protected.Use(security.JwtAuthProcessor())
	protected.GET("/myaccount", accounts.GetMyAccount)
	protected.PUT("/myaccount", accounts.PutMyAccount)
	protected.PUT("/mypassword", accounts.PutMyPassword)
	protected.GET("/myinventory", inventories.GetMyInventory)
	protected.GET("/mypacks", packs.GetMyPacks)
	protected.GET("/mypack/:id", packs.GetMyPackByID)
	protected.POST("/mypack", packs.PostMyPack)
	protected.PUT("/mypack/:id", packs.PutMyPackByID)
	protected.DELETE("/mypack/:id", packs.DeleteMyPackByID)
	protected.GET("/mypack/:id/packcontents", packs.GetMyPackContentsByPackID)
	protected.POST("/mypack/:id/packcontent", packs.PostMyPackContent)
	protected.PUT("/mypack/:id/packcontent/:item_id", packs.PutMyPackContentByID)
	protected.DELETE("/mypack/:id/packcontent/:item_id", packs.DeleteMyPackContentByID)
	protected.GET("/myinventory/:id", inventories.GetMyInventoryByID)
	protected.POST("/myinventory", inventories.PostMyInventory)
	protected.PUT("/myinventory/:id", inventories.PutMyInventoryByID)
	protected.DELETE("/myinventory/:id", inventories.DeleteMyInventoryByID)
	protected.POST("/importfromlighterpack", packs.ImportFromLighterPack)

	private := router.Group("/api/admin")
	private.Use(security.JwtAuthAdminProcessor())
	private.GET("/accounts", accounts.GetAccounts)
	private.GET("/accounts/:id", accounts.GetAccountByID)
	private.POST("/accounts", accounts.PostAccount)
	private.PUT("/accounts/:id", accounts.PutAccountByID)
	private.DELETE("/accounts/:id", accounts.DeleteAccountByID)
	private.GET("/inventories", inventories.GetInventories)
	private.GET("/inventories/:id", inventories.GetInventoryByID)
	private.POST("/inventories", inventories.PostInventory)
	private.PUT("/inventories/:id", inventories.PutInventoryByID)
	private.DELETE("/inventories/:id", inventories.DeleteInventoryByID)
	private.GET("/packs", packs.GetPacks)
	private.GET("/packs/:id", packs.GetPackByID)
	private.POST("/packs", packs.PostPack)
	private.PUT("/packs/:id", packs.PutPackByID)
	private.DELETE("/packs/:id", packs.DeletePackByID)
	private.GET("/packcontents", packs.GetPackContents)
	private.GET("/packcontents/:id", packs.GetPackContentByID)
	private.POST("/packcontents", packs.PostPackContent)
	private.PUT("/packcontents/:id", packs.PutPackContentByID)
	private.DELETE("/packcontents/:id", packs.DeletePackContentByID)
	private.GET("/packs/:id/packcontents", packs.GetPackContentsByPackID)

	if config.Stage == "DEV" || config.Stage == "LOCAL" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	if config.Stage == "LOCAL" {
		err := router.Run("localhost:8080")
		if err != nil {
			panic(err)
		}
	} else {
		err := router.Run(":8080")
		if err != nil {
			panic(err)
		}
	}
}
