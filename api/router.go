package api

import (
	"context"
	"errors"
	"gproxy/api/middleware"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wytools/rlog/rotation"
)

var httpServer *http.Server

func StartHttpServer(port int) {
	router := InitRouter()
	httpServer = &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: router,
	}

	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen: %v\n", err)
		}
	}()

	if err := middleware.LoadEvmChains(); err != nil {
		panic(err)
	}
}

func StopHttpServer() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

// InitRouter init the router
func InitRouter() *gin.Engine {
	GinLogger, err := rotation.NewSizeWithLockLogger("logs/http/gin.log", 100*1024*1024, 10)
	if err != nil {
		log.Panicf("Create GinLogger file error. %v", err)
	}

	gin.SetMode(gin.ReleaseMode)
	api := gin.New()
	api.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: GinLogger}), gin.Recovery())

	api.GET("/chains", GetChains)

	api.GET("/rpc", GetRpcByChainId)

	api.GET("/mint_nicknft", MintNFT)

	api.GET("/smr_price", SmrPrice)

	api.GET("/faucet", Faucet)

	api.GET("/proxy/account", GetProxyAccount)

	group := api.Group("/group").Use(middleware.SignIpRateLimiterWare)
	{
		group.POST("/filter", FilterGroupV2)
		group.POST("/filter/v2", FilterGroupV2)
		group.POST("/verify", VerifyGroup)
	}

	mainAcc := api.Group("/proxy").Use(middleware.SignIpRateLimiterWare).Use(middleware.VerifyEvmSign)
	{
		mainAcc.POST("/register", RegisterProxy)
	}

	solAcc := api.Group("/proxy").Use(middleware.SignIpRateLimiterWare).Use(middleware.VerifySolSign)
	{
		solAcc.POST("/register/solana", RegisterProxy)
	}

	proxy := api.Group("/proxy").Use(middleware.SignIpRateLimiterWare).Use(middleware.VerifyEd25519Sign)
	{
		proxy.POST("/mint_nicknft", MintNameNftForMM)
		proxy.POST("/send", SendTxEssence)
		proxy.POST("/send/asyn", SendTxEssenceAsyn)
	}

	return api
}
