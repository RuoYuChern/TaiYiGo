package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"taiyigo.com/common"
)

type TaoClaims struct {
	Username string `json:"username"`
	OpenId   string `json:"openid"`
	Randomid string `json:"randomid"`
	jwt.RegisteredClaims
}

func gateFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()
		latency := time.Since(t)
		status := c.Writer.Status()
		common.Logger.Infof("perf stat: url=[%s], time used=[%d], status=[%d]", c.FullPath(), latency.Milliseconds(), status)
	}
}

func jwtAuthMiddleware(c *gin.Context) {
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		c.String(http.StatusForbidden, "Authorization empty")
		c.Abort()
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if (len(parts) != 2) || (parts[0] != "Bearer") {
		c.String(http.StatusForbidden, "Authorization error")
		c.Abort()
		return
	}

	err := verifyToken(parts[1], c)
	if err != nil {
		c.String(http.StatusForbidden, "Authorization faild")
		c.Abort()
		return
	}
	c.Next()
}

func hello(c *gin.Context) {
	c.String(http.StatusOK, "Hello")
}

func auth(c *gin.Context) {
	c.String(http.StatusOK, "Auth ok")
}

type restServer struct {
	common.TaiyiApp
	srv *http.Server
}

func (api *restServer) GetName() string {
	return "rest-api"
}

func (api *restServer) Start(ctx *context.Context) error {
	router := gin.New()
	router.MaxMultipartMemory = 8 << 20
	router.Use(gateFilter())
	router.Use(gin.Recovery())
	router.GET(fmt.Sprintf("%s/hello", common.Conf.Http.Prefix), hello)
	router.GET(fmt.Sprintf("%s/auth", common.Conf.Http.Prefix), jwtAuthMiddleware, auth)
	router.GET(fmt.Sprintf("%s/hq/get-stf", common.Conf.Http.Prefix), getStfRecord)
	router.GET(fmt.Sprintf("%s/hq/get-trend", common.Conf.Http.Prefix), getSymbolTrend)
	router.GET(fmt.Sprintf("%s/hq/get-pair-trend", common.Conf.Http.Prefix), getSymbolPairTrend)
	router.GET(fmt.Sprintf("%s/hq/get-dash", common.Conf.Http.Prefix), getDashboard)
	router.GET(fmt.Sprintf("%s/hq/get-up-down", common.Conf.Http.Prefix), getUpDown)
	router.GET(fmt.Sprintf("%s/hq/get-hot", common.Conf.Http.Prefix), getHot)
	router.GET(fmt.Sprintf("%s/hq/get-symbol-n", common.Conf.Http.Prefix), getSymbolLastN)
	router.GET(fmt.Sprintf("%s/hq/get-cn-rt-price", common.Conf.Http.Prefix), getCnRtPrice)
	router.GET(fmt.Sprintf("%s/hq/get-forward", common.Conf.Http.Prefix), getForward)
	router.POST(fmt.Sprintf("%s/hq/do-quant", common.Conf.Http.Prefix), jwtAuthMiddleware, postQuantPredit)
	router.GET(fmt.Sprintf("%s/trade/get-trading-stat", common.Conf.Http.Prefix), tradingStat)

	router.POST(fmt.Sprintf("%s/auth/do-login", common.Conf.Http.Prefix), doUserLogin)
	router.POST(fmt.Sprintf("%s/auth/do-add-user", common.Conf.Http.Prefix), jwtAuthMiddleware, doAddUser)

	router.POST(fmt.Sprintf("%s/trade/do-trading", common.Conf.Http.Prefix), jwtAuthMiddleware, doTrading)
	router.POST(fmt.Sprintf("%s/trade/modify-trading", common.Conf.Http.Prefix), jwtAuthMiddleware, modifyTrading)

	router.POST(fmt.Sprintf("%s/load-cn-history", common.Conf.Http.Prefix), jwtAuthMiddleware, loadCnSharesHistory)
	router.POST(fmt.Sprintf("%s/load-cn-basic", common.Conf.Http.Prefix), jwtAuthMiddleware, loadCnBasic)
	router.POST(fmt.Sprintf("%s/justify-kv", common.Conf.Http.Prefix), jwtAuthMiddleware, justifyKeyValue)
	router.POST(fmt.Sprintf("%s/admin-cmd", common.Conf.Http.Prefix), jwtAuthMiddleware, adminCommond)
	api.srv = &http.Server{
		Addr:    fmt.Sprintf(":%d", common.Conf.Http.Port),
		Handler: router,
	}
	go func() {
		/**进行连接**/
		common.Logger.Infof("server listen on:%d", common.Conf.Http.Port)
		if err := api.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			common.Logger.Errorf("listen error:%s", err.Error())
		}
	}()
	return nil
}

func (api *restServer) Stop(ctx *context.Context) {
	if err := api.srv.Shutdown(*ctx); err != nil {
		common.Logger.Warn("Server shutdown:", err)
	}
}

func StartApi() {
	api := &restServer{}
	err := common.GetApp().AddService(api)
	if err != nil {
		panic(err)
	}
}
