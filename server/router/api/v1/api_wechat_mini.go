package api

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lazyfury/go-web-start/server/config"
	"github.com/lazyfury/go-web-start/server/middleware"
	"github.com/lazyfury/go-web-start/server/model"
	"github.com/lazyfury/go-web-template/response"
	"github.com/lazyfury/go-web-template/tools"
	"github.com/lazyfury/go-web-template/tools/crypto"
	"github.com/lazyfury/go-web-template/tools/types"
)

var code2SessionKeyURL = "https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code"

type (
	code2SessionKey struct {
		OpenID     string `json:"openid"`
		Unionid    string `json:"unionid"`
		SessionKey string `json:"session_key"`
		ErrCode    int    `json:"errcode"`
		ErrMsg     string `json:"errmsg"`
	}
	wechatLoginParams struct {
		EncryptedData string
		Iv            string
	}
)

var mini = &config.Global.WechatMini

func wehcatMini(g *gin.RouterGroup) {
	mini := g.Group("/wechat-mini")
	mini.POST("/login", wechatMiniLogin)

	mini.POST("/easy-login", easyLogin)

	mini.POST("/sendMsg", sendMsg, middleware.Auth)
}
func sendMsg(c *gin.Context) {
	auth := c.MustGet("user").(*model.User)
	db := model.DB

	var user = model.User{}
	if notfoundUser := db.Model(&user).Find(map[string]interface{}{
		"id": auth.ID,
	}).Error != nil; notfoundUser {
		response.Error("没找到用户")
	}

	var miniUser = model.WechatMiniUser{UID: user.ID}
	if notfoundUser := db.Model(&miniUser).Find(&miniUser).Error != nil; notfoundUser {
		response.Error("没找到用户")
	}

	err := mini.SendSubscribeMessage(miniUser.OpenID, "LEe5SuSVcBC2wei1XW9QwouVZ79T5p3DK-8QfA3ecxM", "https://wechat.com", map[string]interface{}{
		"thing1": map[string]string{
			"value": user.Name,
		},
		"thing2": map[string]string{
			"value": "测试数据当前",
		},
		"time3": map[string]string{
			"value": "2019年10月1日 15:01",
		},
		"thing4": map[string]string{
			"value": "thing4",
		},
		"phone_number5": map[string]string{
			"value": "+86-0766-66888866",
		},
	})
	if err != nil {
		response.Error(err)
	}
	c.JSON(http.StatusOK, response.JSONSuccess("", nil))
}

func easyLogin(c *gin.Context) {
	jsCode := c.Query("js_code")
	if jsCode == "" {
		response.Error("请输入js_code")
	}

	// 请求微信服务器
	url := fmt.Sprintf(code2SessionKeyURL, mini.Appid, mini.Appsecret, jsCode)
	// fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		response.Error("获取失败")
	}

	// 解码
	var m code2SessionKey

	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		response.Error("获取微信返回内容失败")
	}
	// 获取session失败
	if m.ErrCode != 0 {
		response.Error(response.JSONError("", m))
	}

	db := model.DB
	wechatUser := &model.WechatMiniUser{OpenID: m.OpenID}
	// wechatUser.BaseControll.Model = wechatUser
	// 登陆
	// if hasOne := wechatUser.HasOne(wechatUser); hasOne {
	// 	user := &model.User{BaseControll: model.BaseControll{ID: wechatUser.UID}}
	// 	if err := db.Where(user).Find(user).Error; err == nil {
	// 		getJWT(c, user)
	// 		return
	// 	}
	// }

	// 注册
	user := &model.User{Name: tools.RandStringBytes(6), Password: crypto.SHA256String(tools.RandStringBytes(16))}
	req := c.Request
	ua := req.UserAgent()
	ip := c.ClientIP()
	user.IP = ip
	user.Ua = ua
	user.LoginTime = types.LocalTime{Time: time.Now()}

	if err := db.Table(user.TableName()).Create(&user).Error; err != nil {
		response.Error("创建用户失败")
	}

	wechatUser.SessionKey = m.SessionKey
	wechatUser.UID = user.ID
	wechatUser.Unionid = m.Unionid
	if err := db.Table(wechatUser.TableName()).Create(&wechatUser).Error; err != nil {
		response.Error("创建微信小程序用户失败")
	}

	getJWT(c, user)

}

func getJWT(c *gin.Context, user *model.User) {
	str, _ := middleware.CreateToken(*user)
	c.JSON(http.StatusOK, response.JSONSuccess("", str))
}

func wechatMiniLogin(c *gin.Context) {
	jsCode := c.Query("js_code")
	if jsCode == "" {
		response.Error("请输入js_code")
	}

	var param wechatLoginParams

	if err := c.Bind(&param); err != nil {
		response.Error("参数错误")
	}

	if param.EncryptedData == "" || param.Iv == "" {
		response.Error("请传入用户信息")
	}

	// 请求微信服务器
	url := fmt.Sprintf(code2SessionKeyURL, mini.Appid, mini.Appsecret, jsCode)
	// fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		response.Error(err)
	}

	// 解码
	var m code2SessionKey

	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		response.Error(response.JSONError("获取微信返回内容失败", err))
	}

	// 获取session失败
	if m.ErrCode != 0 {
		response.Error(response.JSONError(m.ErrMsg, m))
	}

	if m.SessionKey == "" {
		response.Error("获取session_key失败")
	}

	// baseDecode
	key, _ := base64.StdEncoding.DecodeString(m.SessionKey)
	encryptedData, _ := base64.StdEncoding.DecodeString(param.EncryptedData)
	iv, _ := base64.StdEncoding.DecodeString(param.Iv)
	// 解密数据
	result := wechatMiniDecoder(string(encryptedData), key, string(iv))

	// 解码用户信息
	var info map[string]interface{}

	if err := json.Unmarshal(result, &info); err != nil {
		response.Error("解码用户信息失败")
	}

	c.JSON(http.StatusOK, response.JSONSuccess("", info))
}

func wechatMiniDecoder(str string, key []byte, iv string) []byte {
	c, _ := aes.NewCipher(key)
	strNew := []byte(str)

	cbcDecoder := cipher.NewCBCDecrypter(c, []byte(iv))
	plaintextCopy := make([]byte, len(strNew))

	cbcDecoder.CryptBlocks(plaintextCopy, strNew)
	return plaintextCopy
}
