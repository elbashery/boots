package goline

import (
	"../LineThrift"
	"context"
	"strings"
)

func (p *LINE) AcquireEncryptedAccessToken() (string, error) {
	res, err := p.Client.AcquireEncryptedAccessToken(context.TODO(), 2)
	token := strings.Split(res, "\x1e")[1]
	p.SquareObsToken = token
	return token, err
}

func (p *LINE) FetchMyEvents(syncToken string, limit int32, continuationToken string) (*LineThrift.FetchMyEventsResponse, error) {
	req := &LineThrift.FetchMyEventsRequest{
		SyncToken: syncToken,
		Limit: limit,
		ContinuationToken: continuationToken,
	}
	res, err := p.SquareService().FetchMyEvents(context.TODO(), req)
	return res, err
}

func (p *LINE) GetJoinedSquares() (*LineThrift.GetJoinedSquaresResponse, error) {
	req := &LineThrift.GetJoinedSquaresRequest{
		ContinuationToken: "",
		Limit: 50,
	}
	res, err := p.SquareService().GetJoinedSquares(context.TODO(), req)
	return res, err
}

func (p *LINE) GetJoinedSquareChats() (*LineThrift.GetJoinedSquareChatsResponse, error) {
	req := &LineThrift.GetJoinedSquareChatsRequest{
		ContinuationToken: "",
		Limit: 50,
	}
	res, err := p.SquareService().GetJoinedSquareChats(context.TODO(), req)
	return res, err
}

func (p *LINE) GetSquare(squareMid string) (*LineThrift.GetSquareResponse, error) {
	req := &LineThrift.GetSquareRequest{
		Mid: squareMid,
	}
	res, err := p.SquareService().GetSquare(context.TODO(), req)
	return res, err
}

func (p *LINE) GetSquareChat(squareMid string) (*LineThrift.GetSquareChatResponse, error) {
	req := &LineThrift.GetSquareChatRequest{
		SquareChatMid: squareMid,
	}
	res, err := p.SquareService().GetSquareChat(context.TODO(), req)
	return res, err
}

func (p *LINE) SendSquareMessage(squareChatMid string, text string) (*LineThrift.SendMessageResponse, error) {
	SS := p.SquareService()
	M := &LineThrift.Message{
		To: squareChatMid,
		Text: text,
		ContentType: 0,
		ContentMetadata: nil,
		RelatedMessageId: "0", // to be honest, i don't know what this is for, and if i don't throw something it wouldn't send the message
	}
	sqM := &LineThrift.SquareMessage{
		Message: M,
	}
	req := &LineThrift.SendMessageRequest{
		SquareChatMid: squareChatMid,
		SquareMessage: sqM,
	}
	res, err := SS.SendMessage(context.TODO(), req)
	return res, err
}