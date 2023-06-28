package api

import (
	"context"
	"errors"
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

func getToken(usename string, opendid string) (string, error) {
	cure := time.Now()
	c := TaoClaims{
		usename,
		opendid,
		jwt.TimePrecision.String(),
		jwt.RegisteredClaims{
			Issuer:    "XianXian",
			Subject:   "View",
			Audience:  []string{"Tao"},
			ID:        opendid,
			ExpiresAt: jwt.NewNumericDate(cure.Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(cure),
			NotBefore: jwt.NewNumericDate(cure),
		},
	}

	toekn := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	ss, err := toekn.SignedString([]byte(common.Conf.Http.Jwt))
	if err != nil {
		common.Logger.Infof("Get token error:%s", err.Error())
		return "", err
	}
	return ss, nil
}

func verifyToken(tokenString string, c *gin.Context) error {
	if tokenString == common.Conf.Http.Token {
		return nil
	}
	token, err := jwt.ParseWithClaims(tokenString, &TaoClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(common.Conf.Http.Jwt), nil
	})
	if err != nil {
		common.Logger.Infof("Parse token error:%s", err.Error())
		return err
	}

	claims, ok := token.Claims.(*TaoClaims)
	if !ok {
		common.Logger.Infof("Parse token error: type error")
		return errors.New("type error")
	}
	c.Set("openId", claims.OpenId)
	c.Set("Username", claims.Username)
	return nil

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
	router.POST(fmt.Sprintf("%s/load-cn-history", common.Conf.Http.Prefix), jwtAuthMiddleware, loadCnSharesHistory)
	router.POST(fmt.Sprintf("%s/start-cn-stf", common.Conf.Http.Prefix), jwtAuthMiddleware, startCnSTFFlow)
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
