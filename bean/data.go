package bean

type OnlineReq struct {
	ReqContent `xml:"xml"`
}

type OnlineRsp struct {
	MsgContent `xml:"xml"`
}

// SendMsgByAppReq 接收接口内消息发送至企微
type SendMsgByAppReq struct {
	ToUser  string `json:"touser"`
	AgentId int64  `json:"agentid"`
	Content string `json:"content"`
	App     string `json:"app"` //可选 chatZone、brocaAssistant
}

// SendMsgReq 发送给企微用户结构体，json内容不可改
type SendMsgReq struct {
	ToUser               string `json:"touser"`
	ToParty              string `json:"toparty"`
	ToTag                string `json:"totag"`
	MsgType              string `json:"msgtype"`
	AgentId              int64  `json:"agentid"`
	Text                 Text   `json:"text"`
	Safe                 int    `json:"safe"`
	EnableIdTrands       int    `json:"enable_id_trands"`
	EnableDuplicateCheck int    `json:"enable_duplicate_check"`
}

type Text struct {
	Content string `json:"content"`
}

// Callback 企微消息的路径参数
type Callback struct {
	MsgSignature string `json:"msg_signature"`
	Timestamp    string `json:"timestamp"`
	Nonce        string `json:"nonce"`
	EchoStr      string `json:"echostr"`
	CorpId       string `json:"corpid"`
}

// ReqContent 企微消息的Body参数
type ReqContent struct {
	ToUsername string `xml:"ToUserName"`
	AgentID    string `xml:"AgentID"`
	Encrypt    string `xml:"Encrypt"`
}

// MsgContent 企微消息的明文内容
type MsgContent struct {
	ToUsername   string `xml:"ToUserName"`
	FromUserName string `xml:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"` // 文本为text，事件为event
	Content      string `xml:"Content"`
	MsgId        int64  `xml:"MsgId"`
	AgentID      int64  `xml:"AgentID"`
}

type RspContent struct {
	ToUsername   string `xml:"ToUserName"`
	FromUsername string `xml:"FromUsername"`
	CreateTime   int64  `xml:"CreateTime"`
	MsgType      string `xml:"MsgType"` // 文本为text，事件为event
	Content      string `xml:"Content"`
}

// ModelReq 算法请求参数
type ModelReq struct {
	BizId     int64  `json:"biz_id"`
	SessionId string `json:"session_id"`
	Query     string `json:"query"`
}

// ModelRsp 算法响应参数
type ModelRsp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

type Data struct {
	Query      string   `json:"query"`
	AnswerList []Answer `json:"answer_list"`
}

type Answer struct {
	Text string `json:"text"`
}

// ChatRsp 企微接口响应
type ChatRsp struct {
	Errcode     int    `json:"errcode"`
	Errmsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
}
