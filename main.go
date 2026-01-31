package main

import (
	"fmt"

	_ "github.com/Angak0k/pimpmypack/docs"
	"github.com/Angak0k/pimpmypack/pkg/accounts"
	"github.com/Angak0k/pimpmypack/pkg/config"
	"github.com/Angak0k/pimpmypack/pkg/database"
	"github.com/Angak0k/pimpmypack/pkg/images"
	"github.com/Angak0k/pimpmypack/pkg/inventories"
	"github.com/Angak0k/pimpmypack/pkg/packs"
	"github.com/Angak0k/pimpmypack/pkg/security"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	envDev   = "DEV"
	envLocal = "LOCAL"
)

func initApp() error {
	// init env
	err := config.EnvInit(".env")
	if err != nil {
		return fmt.Errorf("error loading .env file or environment variable : %w", err)
	}

	// init DB
	err = database.Initialization()
	if err != nil {
		return fmt.Errorf("error connecting database : %w", err)
	}

	// init DB migration
	err = database.Migrate()
	if err != nil {
		return fmt.Errorf("error migrating database : %w", err)
	}

	return nil
}

// @title PimpMyPack API
// @description API server to manage Backpack Inventory and Packing Lists
// @version 1.0
// @host pmp-dev.alki.earth
// @Schemes https
// @BasePath /api
func main() {
	err := initApp()
	if err != nil {
		panic(err)
	}

	if config.Stage == envDev || config.Stage == envLocal {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	setupRoutes(router)

	startServer(router)
}

func setupRoutes(router *gin.Engine) {
	setupPublicRoutes(router)
	setupProtectedRoutes(router)
	setupPrivateRoutes(router)
	setupSwaggerRoutes(router)
}

func setupPublicRoutes(router *gin.Engine) {
	public := router.Group("/api")
	public.POST("/register", accounts.Register)
	public.POST("/login", accounts.Login)
	public.GET("/confirmemail", accounts.ConfirmEmail)
	public.POST("/forgotpassword", accounts.ForgotPassword)
	public.GET("/sharedlist/:sharing_code", packs.SharedList)
	public.GET("/v1/packs/:id/image", images.GetPackImage)
}

func setupProtectedRoutes(router *gin.Engine) {
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
	protected.POST("/mypack/:id/share", packs.ShareMyPack)
	protected.DELETE("/mypack/:id/share", packs.UnshareMyPack)
	protected.GET("/mypack/:id/packcontents", packs.GetMyPackContentsByPackID)
	protected.POST("/mypack/:id/packcontent", packs.PostMyPackContent)
	protected.PUT("/mypack/:id/packcontent/:item_id", packs.PutMyPackContentByID)
	protected.DELETE("/mypack/:id/packcontent/:item_id", packs.DeleteMyPackContentByID)
	protected.GET("/myinventory/:id", inventories.GetMyInventoryByID)
	protected.POST("/myinventory", inventories.PostMyInventory)
	protected.PUT("/myinventory/:id", inventories.PutMyInventoryByID)
	protected.DELETE("/myinventory/:id", inventories.DeleteMyInventoryByID)
	protected.POST("/importfromlighterpack", packs.ImportFromLighterPack)
	protected.POST("/mypack/:id/image", images.UploadPackImage)
	protected.DELETE("/mypack/:id/image", images.DeletePackImage)
}

func setupPrivateRoutes(router *gin.Engine) {
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
}

func setupSwaggerRoutes(router *gin.Engine) {
	if config.Stage == envDev || config.Stage == envLocal {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}
}

func startServer(router *gin.Engine) {
	address := ":8080"
	if config.Stage == envLocal {
		address = "localhost:8080"
	}

	if err := router.Run(address); err != nil {
		panic(err)
	}
}
