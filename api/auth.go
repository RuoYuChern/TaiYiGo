package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/infra"
)

func getToken(usename string, opendid string) (string, error) {
	cure := time.Now()
	c := TaoClaims{
		usename,
		opendid,
		jwt.TimePrecision.String(),
		jwt.RegisteredClaims{
			Issuer:    "TaiYiGo",
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

func doUserLogin(c *gin.Context) {
	user := dto.UserPwdReq{}
	rsp := dto.CommonResponse{Code: http.StatusBadRequest, Msg: "Can not find args"}
	if err := c.BindJSON(&user); err != nil {
		common.Logger.Infoln("Can not find args")
		c.JSON(http.StatusOK, rsp)
		return
	}

	val, err := infra.GetByKey(infra.USER_TABLE, user.Name)
	if err != nil {
		rsp.Code = http.StatusNotFound
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, rsp)
		return
	}

	signal := common.GetMd5(val, user.Noice)
	if strings.Compare(signal, user.Pwd) != 0 {
		rsp.Code = http.StatusForbidden
		rsp.Msg = "forbidden"
		common.Logger.Infof("user:[%s],pwd:[%s] !=[%s]", user.Name, user.Pwd, signal)
		c.JSON(http.StatusOK, rsp)
		return
	}
	//产生token
	jwt, err := getToken(user.Name, "1234")
	if err != nil {
		common.Logger.Warnf("jwt failed:%s", err.Error())
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, rsp)
		return
	}
	c.Header("Authorization", fmt.Sprintf("Bearer %s", jwt))
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	c.JSON(http.StatusOK, &rsp)
}

func doAddUser(c *gin.Context) {
	user := dto.UserPwdReq{}
	rsp := dto.CommonResponse{Code: http.StatusBadRequest, Msg: "Can not find args"}
	if err := c.BindJSON(&user); err != nil {
		common.Logger.Infoln("Can not find args")
		c.JSON(http.StatusOK, rsp)
		return
	}

	err := infra.SetKeyValue(infra.USER_TABLE, user.Name, user.Pwd)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, rsp)
		return
	}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	c.JSON(http.StatusOK, rsp)
}
