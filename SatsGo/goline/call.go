package goline

import (
	"../LineThrift"
	"context"
	// "fmt"
)

func (p *LINE) AcquireCallRoute(to string) (r []string, err error) {
	res, err := p.Client.AcquireCallRoute(context.TODO(), to)
	return res, err
}

func (p *LINE) AcquireGroupCallRoute(chatMid string, mediaType LineThrift.GroupCallMediaType) (r *LineThrift.GroupCallRoute, err error) {
	res, err := p.CallService().AcquireGroupCallRoute(context.TODO(), chatMid, mediaType)
	return res, err
}

func (p *LINE) GetGroupCall(chatMid string) (r *LineThrift.GroupCall, err error) {
	res, err := p.CallService().GetGroupCall(context.TODO(), chatMid)
	return res, err
}

func (p *LINE) InviteIntoGroupCall(chatMid string, memberMids []string, mediaType LineThrift.GroupCallMediaType) (err error) {
	err = p.CallService().InviteIntoGroupCall(context.TODO(), chatMid, memberMids, mediaType)
	return err
}