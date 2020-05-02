package goline

import (
	"../LineThrift"
	"../thrift"
	"net/http"
	"context"
	"io/ioutil"
	//"fmt"
)

type LINE struct {
    IsLogin bool
    AuthToken string
    Certificate string
    Revision int64
    MID string
    Headers http.Header
    ChannelHeaders http.Header
    Client *LineThrift.TalkServiceClient
    Poll *LineThrift.TalkServiceClient
    Friends []string
    Limit bool
    AppName string
    TimeLineHeader map[string]string
    Number int
    SquareObsToken string
}

func NewLogin() *LINE {
	p := new(LINE)
	p.IsLogin = false
	p.AuthToken = ""
	p.Certificate = ""
	p.Revision = 0
    p.MID = ""
    p.Limit = false
    p.AppName = ""
    p.TimeLineHeader = map[string]string{}
    return p
}

func (p *LINE) Login(args ...string){
	if len(args) == 0 {
		p.AppName = "IOSIPAD\t9.12.0\tiOS\t12.0.1"
		p.LoginWithQrCode(true)
	} else if len(args) == 1 {
		p.AppName = "IOS\t9.12.0\tiOS\t12.0.1" 
		p.LoginWithAuthToken(args[0])
	} else if len(args) == 2 {
		p.AppName = args[1]
		p.LoginWithAuthToken(args[0])
	} else if len(args) == 3 {
		p.AppName = args[2]
		p.LoginWithCredential(args[0], args[1])
	}
}

func (p *LINE) LoginOtherDevice(token string, appName string) error {
	p.AppName = "IOS\t9.12.0\tiOS\t12.0.1" 
	p.AuthToken = token
	talk := p.TalkService()
	p.AppName = appName
	tauth := p.AuthService()
	qrCode, err := tauth.GetAuthQrcode(context.TODO(), true, SYSTEM_NAME, true)
	if err != nil{ return err }
	talk.VerifyQrcode(context.TODO(), qrCode.Verifier, "")
	client := &http.Client{}
	req, _ := http.NewRequest("GET", LINE_HOST_DOMAIN+LINE_CERTIFICATE_PATH, nil)
	req.Header.Set("User-Agent",USER_AGENT)
	req.Header.Set("X-Line-Application",p.AppName)
	req.Header.Set("X-Line-Carrier",CARRIER)
	req.Header.Set("X-Line-Access",qrCode.Verifier)
	p.AuthToken = qrCode.Verifier
	res, _ := client.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	x, _ := getResult([]byte(body))
	iR := x.Result
	_verifier := iR.Verifier
	loginZ := p.LoginZService()
	loginReq := loginRequestQR(1, _verifier)
	resultz, err := loginZ.LoginZ(context.TODO(),loginReq)
	if err != nil { p.AppName = "IOS\t9.12.0\tiOS\t12.0.1"; p.LoginWithAuthToken(token); return err }
	p.IsLogin = true
	p.LoginWithAuthToken(resultz.AuthToken)
	return nil
}

func GetTalkService(token string, appName string) *LineThrift.TalkServiceClient {
	httpClient, _ := thrift.NewTHttpClient(LINE_HOST_DOMAIN+LINE_API_QUERY_PATH_FIR)
	buffer := thrift.NewTBufferedTransportFactory(4096)
	trans := httpClient.(*thrift.THttpClient)
	trans.SetHeader("User-Agent",USER_AGENT)
	trans.SetHeader("X-Line-Application",appName)
	trans.SetHeader("X-Line-Carrier",CARRIER)
	trans.SetHeader("X-Line-Access",token)
	trans.SetHeader("Connection","Keep-Alive")
	buftrans, _ := buffer.GetTransport(trans)
	compactProtocol := thrift.NewTCompactProtocolFactory()
	return LineThrift.NewTalkServiceClientFactory(buftrans, compactProtocol)
}

func (p *LINE) TalkService() *LineThrift.TalkServiceClient {
	//fmt.Println("#### TalkService Initiated. ####")
	//httpClient, _ := thrift.NewFastTHttpClient(LINE_HOST_DOMAIN+LINE_API_QUERY_PATH_FIR)
	httpClient, _ := thrift.NewTHttpClient(LINE_HOST_DOMAIN+LINE_API_QUERY_PATH_FIR)
	buffer := thrift.NewTBufferedTransportFactory(4096)
	trans := httpClient.(*thrift.THttpClient)
	trans.SetHeader("User-Agent",USER_AGENT)
	trans.SetHeader("X-Line-Application",p.AppName)
	trans.SetHeader("X-Line-Carrier",CARRIER)
	trans.SetHeader("X-Line-Access",p.AuthToken)
	trans.SetHeader("Connection","Keep-Alive")
	buftrans, _ := buffer.GetTransport(trans)
	compactProtocol := thrift.NewTCompactProtocolFactory()
	return LineThrift.NewTalkServiceClientFactory(buftrans, compactProtocol)
}

func (p *LINE) PollService() *LineThrift.TalkServiceClient {
	//fmt.Println("#### PollService Initiated. ####")
	//httpClient, _ := thrift.NewFastTHttpClient(LINE_HOST_DOMAIN+"/P4")
	httpClient, _ := thrift.NewTHttpClient(LINE_HOST_DOMAIN+"/P4")
	buffer := thrift.NewTBufferedTransportFactory(4096)
	trans := httpClient.(*thrift.THttpClient)
	trans.SetHeader("User-Agent",USER_AGENT)
	trans.SetHeader("X-Line-Application",p.AppName)
	trans.SetHeader("X-Line-Carrier",CARRIER)
	trans.SetHeader("X-Line-Access",p.AuthToken)
	trans.SetHeader("Connection","Keep-Alive")
	buftrans, _ := buffer.GetTransport(trans)
	compactProtocol := thrift.NewTCompactProtocolFactory()
	return LineThrift.NewTalkServiceClientFactory(buftrans, compactProtocol)
}

func (p *LINE) ChannelService() *LineThrift.ChannelServiceClient {
	//fmt.Println("#### ChannelService Initiated. ####")
	//httpClient, _ := thrift.NewFastTHttpClient(LINE_HOST_DOMAIN+LINE_CHAN_QUERY_PATH)
	httpClient, _ := thrift.NewTHttpClient(LINE_HOST_DOMAIN+LINE_CHAN_QUERY_PATH)
	buffer := thrift.NewTBufferedTransportFactory(4096)
	trans := httpClient.(*thrift.THttpClient)
	trans.SetHeader("User-Agent",USER_AGENT)
	trans.SetHeader("X-Line-Application",p.AppName)
	trans.SetHeader("X-Line-Carrier",CARRIER)
	trans.SetHeader("X-Line-Access",p.AuthToken)
	trans.SetHeader("Connection","Keep-Alive")
	buftrans, _ := buffer.GetTransport(trans)
	compactProtocol := thrift.NewTCompactProtocolFactory()
	return LineThrift.NewChannelServiceClientFactory(buftrans, compactProtocol)
}

func (p *LINE) CallService() *LineThrift.CallServiceClient {
	//fmt.Println("#### CallService Initiated. ####")
	//httpClient, _ := thrift.NewFastTHttpClient(LINE_HOST_DOMAIN+LINE_CALL_QUERY_PATH)
	httpClient, _ := thrift.NewTHttpClient(LINE_HOST_DOMAIN+LINE_CALL_QUERY_PATH)
	buffer := thrift.NewTBufferedTransportFactory(4096)
	trans := httpClient.(*thrift.THttpClient)
	trans.SetHeader("User-Agent",USER_AGENT)
	trans.SetHeader("X-Line-Application",p.AppName)
	trans.SetHeader("X-Line-Carrier",CARRIER)
	trans.SetHeader("X-Line-Access",p.AuthToken)
	trans.SetHeader("Connection","Keep-Alive")
	buftrans, _ := buffer.GetTransport(trans)
	compactProtocol := thrift.NewTCompactProtocolFactory()
	return LineThrift.NewCallServiceClientFactory(buftrans, compactProtocol)
}

func (p *LINE) SquareService() *LineThrift.SquareServiceClient {
	//fmt.Println("#### SquareService Initiated. ####")
	//httpClient, _ := thrift.NewFastTHttpClient(LINE_HOST_DOMAIN+LINE_SQUARE_QUERY_PATH)
	httpClient, _ := thrift.NewTHttpClient(LINE_HOST_DOMAIN+LINE_SQUARE_QUERY_PATH)
	buffer := thrift.NewTBufferedTransportFactory(4096)
	trans := httpClient.(*thrift.THttpClient)
	trans.SetHeader("User-Agent",USER_AGENT)
	trans.SetHeader("X-Line-Application",p.AppName)
	trans.SetHeader("X-Line-Carrier",CARRIER)
	trans.SetHeader("X-Line-Access",p.AuthToken)
	trans.SetHeader("Connection","Keep-Alive")
	buftrans, _ := buffer.GetTransport(trans)
	compactProtocol := thrift.NewTCompactProtocolFactory()
	return LineThrift.NewSquareServiceClientFactory(buftrans, compactProtocol)
}

func (p *LINE) LiffService() *LineThrift.LiffServiceClient {
	//fmt.Println("#### LiffService Initiated. ####")
	httpClient, _ := thrift.NewTHttpClient(LINE_HOST_DOMAIN+LINE_LIFF_QUERY_PATH)
	buffer := thrift.NewTBufferedTransportFactory(4096)
	trans := httpClient.(*thrift.THttpClient)
	trans.SetHeader("User-Agent",USER_AGENT)
	trans.SetHeader("X-Line-Application",p.AppName)
	trans.SetHeader("X-Line-Carrier",CARRIER)
	trans.SetHeader("X-Line-Access",p.AuthToken)
	trans.SetHeader("Connection","Keep-Alive")
	buftrans, _ := buffer.GetTransport(trans)
	compactProtocol := thrift.NewTCompactProtocolFactory()
	return LineThrift.NewLiffServiceClientFactory(buftrans, compactProtocol)
}

func (p *LINE) AuthService() *LineThrift.TalkServiceClient {
	//httpClient, _ := thrift.NewFastTHttpClient(LINE_HOST_DOMAIN+LINE_LOGIN_QUERY_PATH)
	httpClient, _ := thrift.NewTHttpClient(LINE_HOST_DOMAIN+LINE_LOGIN_QUERY_PATH)
	buffer := thrift.NewTBufferedTransportFactory(4096)
	trans := httpClient.(*thrift.THttpClient)
	trans.SetHeader("User-Agent",USER_AGENT)
	trans.SetHeader("X-Line-Application",p.AppName)
	trans.SetHeader("X-Line-Carrier",CARRIER)
	trans.SetHeader("Connection","Keep-Alive")
	buftrans, _ := buffer.GetTransport(trans)
	compactProtocol := thrift.NewTCompactProtocolFactory()
	return LineThrift.NewTalkServiceClientFactory(buftrans, compactProtocol)
}

func (p *LINE) LoginZService() *LineThrift.AuthServiceClient {
	httpClient, _ := thrift.NewTHttpClient(LINE_HOST_DOMAIN+LINE_AUTH_QUERY_PATH)
	buffer := thrift.NewTBufferedTransportFactory(4096)
	trans := httpClient.(*thrift.THttpClient)
	trans.SetHeader("User-Agent",USER_AGENT)
	trans.SetHeader("X-Line-Application",p.AppName)
	trans.SetHeader("X-Line-Carrier",CARRIER)
	trans.SetHeader("X-Line-Access",p.AuthToken)
	trans.SetHeader("Connection","Keep-Alive")
	buftrans, _ := buffer.GetTransport(trans)
	compactProtocol := thrift.NewTCompactProtocolFactory()
	return LineThrift.NewAuthServiceClientFactory(buftrans, compactProtocol)
}

