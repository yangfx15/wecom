package app

import (
	"fmt"
	"net/http"
	"wecom/bean"
	"wecom/service"
	"wecom/zaplog"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/", dnsHandler)

	r.GET("/cmd", dataGetHandler)
	r.POST("/send", dataSendHandler)

	r.GET("/data", dataGetHandler)
	r.POST("/data", dataPostHandler)

	r.GET("/chat_zone", dataGetHandler)
	r.POST("/chat_zone", chatZoneHandler)

	return r
}

func dnsHandler(c *gin.Context) {
	c.String(http.StatusOK, "FPbOLeHL0y5sVMDM")
}

func dataSendHandler(c *gin.Context) {
	var req bean.SendMsgByAppReq
	err := c.BindJSON(&req)
	if err != nil {
		zaplog.Error(fmt.Sprintf("%s BindJSON failed, %v", "dataSendHandler", err))
		return
	}

	if err := service.SendMsg(bean.SendMsgReq{
		ToUser:  req.ToUser,
		MsgType: "text",
		AgentId: req.AgentId,
		Text:    bean.Text{Content: req.Content},
	}, req.App); err != nil {
		c.JSON(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	c.JSON(http.StatusOK, "OK")
}

// 业务数据回调接口
func dataGetHandler(c *gin.Context) {
	cb := bean.Callback{
		MsgSignature: c.Query("msg_signature"),
		Timestamp:    c.Query("timestamp"),
		Nonce:        c.Query("nonce"),
		EchoStr:      c.Query("echostr"),
	}

	c.JSON(http.StatusOK, service.DataGet(cb))
}

// 业务数据接收响应接口
func dataPostHandler(c *gin.Context) {
	cb := bean.Callback{
		MsgSignature: c.Query("msg_signature"),
		Timestamp:    c.Query("timestamp"),
		Nonce:        c.Query("nonce"),
	}

	var req bean.OnlineReq
	err := c.BindXML(&req)
	if err != nil {
		zaplog.Error(fmt.Sprintf("BindXML failed, %v", err))
		return
	}
	go service.DataPost(req, cb, service.BrocaAssistant)

	c.JSON(http.StatusOK, "success")
}

func chatZoneHandler(c *gin.Context) {
	cb := bean.Callback{
		MsgSignature: c.Query("msg_signature"),
		Timestamp:    c.Query("timestamp"),
		Nonce:        c.Query("nonce"),
	}

	var req bean.OnlineReq
	err := c.BindXML(&req)
	if err != nil {
		zaplog.Error(fmt.Sprintf("BindXML failed, %v", err))
		return
	}
	go service.DataPost(req, cb, service.ChatZone)

	c.JSON(http.StatusOK, "success")
}
