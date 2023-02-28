package service

import (
	"bytes"
	"context"
	"encoding/json"
	en_xml "encoding/xml"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"wecom/bean"
	"wecom/mw"
	"wecom/zaplog"
)

const (
	Token          = "5IXII6XXrc4gXTrISnyMCxd4"                    //替换开发申请的
	EncodingAESKey = "Ip0vscwi23rvJX5tYHGxVI2kQUfRgjAl8L8mRKNhalR" //替换开发申请的
	Corpid         = ""                                            //替换开发申请的

	ErrCodeTokenExpired = 42001

	ChatPrefix        = "/heartBeat/chat/"
	ModelBrocaPath    = "/chat/api/v1/query"
	ModelChatZonePath = "/chat/api/v1/chatzone/query"

	// 追一企业ID
	CorpId = "ww3a8dea9a55a5e51f"

	// Broca对话助手
	BrocaAssistant = "brocaAssistant"
	// ChatZone
	ChatZone = "chatZone"

	CorpSecret  = "Nse7KnZRKGP0c8co1o9nw-GQw0OFTOxrKkl1xxsGFpU"
	CorpSecret2 = "Qm9lAUt41vYS-QSeU9AGe9TKtXk0MFltEhy99JFFp2U"

	// 企微获取Access-Token
	URLGetWeChatToken = "https://qyapi.weixin.qq.com/cgi-bin/gettoken"
	// 企微推送消息
	URLSendWeChatMsg = "https://qyapi.weixin.qq.com/cgi-bin/linkedcorp/message/send"
)

// 应用凭证
var AppToSecret = map[string]string{
	BrocaAssistant: CorpSecret,
	ChatZone:       CorpSecret2,
}

// 应用Token
var AppToToken = map[string]string{
	BrocaAssistant: "",
	ChatZone:       "",
}

// 应用标识
var AppToFlag = map[string]string{
	BrocaAssistant: "0",
	ChatZone:       "1",
}

var AppToModelPath = map[string]string{
	BrocaAssistant: ModelBrocaPath,
	ChatZone:       ModelChatZonePath,
}

func GetAccessToken(app string) (string, error) {
	var token string
	url := fmt.Sprintf("%s?corpid=%s&corpsecret=%s", URLGetWeChatToken, CorpId, AppToSecret[app])

	httpResponse, err := HTTPPostRequest(url, nil)
	if err != nil {
		zaplog.Error(fmt.Sprintf("%s HTTPPostRequest failed, %v", "GetAccessToken", err))
		return token, err
	}
	var resp bean.ChatRsp
	err = json.Unmarshal(httpResponse, &resp)
	if err != nil || resp.Errcode != 0 {
		zaplog.Error(fmt.Sprintf("%s Unmarshal failed, %v, %v", "GetAccessToken", err, resp))
		return token, err
	}

	return resp.AccessToken, nil
}

func SendMsg(req bean.SendMsgReq, app string) error {
	if AppToToken[app] == "" {
		token, err := GetAccessToken(app)
		if err != nil {
			zaplog.Error(fmt.Sprintf("GetAccessToken err:%v", err))
			return err
		}
		AppToToken[app] = token
	}

	return SendMsgToWeChat(AppToToken[app], app, req)
}

func DataGet(cb bean.Callback) int {
	zaplog.Info(fmt.Sprintf("Callback: %+v\n", cb))

	wxcpt := NewWXBizMsgCrypt(Token, EncodingAESKey, Corpid, XmlType)

	echoStr, cryptErr := wxcpt.VerifyURL(cb.MsgSignature, cb.Timestamp, cb.Nonce, cb.EchoStr)
	if cryptErr != nil {
		zaplog.Error(fmt.Sprintf("VerifyURL failed:%v", cryptErr))
	}
	zaplog.Info(fmt.Sprintf("return:%v\n", string(echoStr)))

	echoInt, _ := strconv.Atoi(string(echoStr))
	return echoInt
}

func DataPost(req bean.OnlineReq, cb bean.Callback, app string) error {
	zaplog.Debug(fmt.Sprintf("DataPost req:%+v, cb:%v", req, cb))
	wxcpt := NewWXBizMsgCrypt(Token, EncodingAESKey, Corpid, XmlType)

	// 解析出明文
	echoStr, cryptErr := wxcpt.VerifyURL(cb.MsgSignature, cb.Timestamp, cb.Nonce, req.Encrypt)
	if cryptErr != nil {
		zaplog.Error(fmt.Sprintf("VerifyURL failed:%v", cryptErr))
		return fmt.Errorf("%v", cryptErr)
	}
	req.Encrypt = string(echoStr)

	var msgContent bean.MsgContent
	err := en_xml.Unmarshal(echoStr, &msgContent)
	if err != nil {
		zaplog.Error(fmt.Sprintf("Unmarshal err:%v", err))
		return err
	}
	zaplog.Info(fmt.Sprintf("MsgContent: %+v", msgContent))
	if msgContent.MsgType != "text" {
		return nil
	}

	sessionId := fmt.Sprintf("%s_%s", msgContent.FromUserName, app)

	// 调用算法接口
	data, err := GetDialogFromModel(bean.ModelReq{
		BizId:     msgContent.MsgId,
		SessionId: sessionId,
		Query:     msgContent.Content,
	}, app)
	if err != nil {
		zaplog.Error(fmt.Sprintf("GetDialogFromModel err:%v", err))
		return err
	}

	err = SendMsg(bean.SendMsgReq{
		ToUser:  msgContent.FromUserName,
		MsgType: msgContent.MsgType,
		AgentId: msgContent.AgentID,
		Text:    bean.Text{Content: fmt.Sprintf("%v", data)},
	}, app)
	if err != nil {
		zaplog.Error(fmt.Sprintf("SendMsg to user[%s] failed:%s", msgContent.FromUserName, err))
		return err
	}
	err = mw.Db.Create(&mw.Dialog{
		UserName:  msgContent.FromUserName,
		App:       app,
		Question:  msgContent.Content,
		Answer:    fmt.Sprintf("%v", data),
		MessageId: msgContent.MsgId,
		Status:    "SUCCESS",
	}).Error
	if err != nil {
		zaplog.Error(fmt.Sprintf("Insert Dialog failed, %v: %s", msgContent, err))
		return err
	}

	return nil
}

func HTTPPostRequest(url string, req interface{}) ([]byte, error) {
	client := http.Client{}
	dataByte, err := json.Marshal(req)
	if err != nil {
		return []byte{}, err
	}
	bodyReader := bytes.NewReader(dataByte)

	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bodyReader)
	if err != nil {
		return []byte{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return []byte{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("error: code is %d", resp.StatusCode)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return body, err
}

func GetDialogFromModel(req bean.ModelReq, app string) (string, error) {
	var rsp string

	ips := mw.EC.GetWithPrefix(ChatPrefix)
	if len(ips) == 0 {
		return rsp, fmt.Errorf("chat IP is empty from ETCD, prefix: %v", ChatPrefix)
	}
	rd := rand.Intn(len(ips))
	key := ips[rd].Key
	ip := strings.TrimPrefix(string(key), ChatPrefix)

	url := fmt.Sprintf("http://%s%s", ip, AppToModelPath[app])
	httpResponse, err := HTTPPostRequest(url, req)
	if err != nil {
		zaplog.Error(fmt.Sprintf("HTTPPostRequest failed, %v", err))
		return rsp, err
	}

	var resp bean.ModelRsp
	err = json.Unmarshal(httpResponse, &resp)
	if err != nil {
		zaplog.Error(fmt.Sprintf("Unmarshal failed, %v", err))
		return rsp, err
	}
	zaplog.Debug(fmt.Sprintf("HTTPPost[%v] Resp:%v", url, resp))

	code := resp.Code
	if code != 0 || len(resp.Data.AnswerList) == 0 {
		zaplog.Error(fmt.Sprintf("HTTPPost Code:%v", code))
		return rsp, fmt.Errorf("get answer from model failed, %v", resp)
	}

	rsp = resp.Data.AnswerList[0].Text
	return rsp, nil
}

func SendMsgToWeChat(token, app string, req bean.SendMsgReq) error {
	url := fmt.Sprintf("%s?access_token=%s", URLSendWeChatMsg, token)
	respBody, err := HTTPPostRequest(url, req)
	if err != nil {
		zaplog.Error(fmt.Sprintf("HTTPPostRequest failed, %v", err))
		return err
	}
	var resp bean.ChatRsp
	err = json.Unmarshal(respBody, &resp)

	if err != nil {
		zaplog.Error(fmt.Sprintf("Send message to wechat failed, %v, %v", err, resp))
		return fmt.Errorf(fmt.Sprintf("Unmarshal failed, %v, %v", err, resp))
	}
	if resp.Errcode != 0 {
		// token过期
		if resp.Errcode == ErrCodeTokenExpired {
			zaplog.Info(fmt.Sprintf("GetAccessToken again"))
			AppToToken[app], err = GetAccessToken(app)
			if err != nil {
				zaplog.Error(fmt.Sprintf("GetAccessToken failed, %v", err))
				return fmt.Errorf(fmt.Sprintf("GetAccessToken failed, %v", err))
			}
			// 可能死循环，即获取的token一直不可用，依赖于企微接口返回token的正确性
			return SendMsgToWeChat(AppToToken[app], app, req)
		}
		zaplog.Error(fmt.Sprintf("Send message to wechat failed, %v, %v", err, resp))
		return fmt.Errorf("resp failed, %v, %v", err, resp)
	}

	return nil
}
