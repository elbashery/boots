package goline

import (
	"../LineThrift"
	"context"
	//"fmt"
    //"encoding/json"
	"github.com/levigross/grequests"
)

func (p *LINE) AllowLiff() error {
	url := "https://access.line.me/dialog/api/permissions"
	data := map[string][]string{"on": []string{"P","CM"}, "off": []string{}}	
	ro := &grequests.RequestOptions{
	    Headers: map[string]string{ "X-Line-Access": p.AuthToken, "X-Line-Application": p.AppName, "X-Line-ChannelId": LINE_LIFF_ID[:10], "Content-Type": "application/json" },
	    JSON: data,
    }
    _, err := grequests.Post(url, ro)
    return err
}

func (p *LINE) SendFlex(to string, data string) error {
	token, err := p.LiffService().IssueLiffView(context.TODO(), &LineThrift.LiffViewRequest{LiffId: LINE_LIFF_ID, Context: &LineThrift.LiffContext{Chat: &LineThrift.LiffChatContext{ChatMid: to}}})
	if liffexcption, ok := err.(*LineThrift.LiffException); ok && liffexcption.Code == 3 {
		p.AllowLiff()
	}
    datas := map[string][]string{"messages": []string{data}}
    ro := &grequests.RequestOptions{
	    Headers: map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + token.AccessToken },
	    JSON: datas,
    }
    _, err = grequests.Post("https://api.line.me/message/v3/share", ro)
    return err
}

