package goline

import (
	"context"
	"github.com/levigross/grequests"
    "net/url"
    "encoding/json"
    "fmt"
)

func (p *LINE) SetTimelineHeaders() {
	channel, _ := p.ChannelService().ApproveChannelAndIssueChannelToken(context.TODO(), TIMELINE_CHANNEL_ID)
	p.TimeLineHeader = map[string]string{
        "Content-Type": "application/json",
        "X-Line-Carrier": CARRIER,
        "X-Line-ChannelToken": channel.ChannelAccessToken,
        "X-Line-Application": p.AppName,
        "User-Agent": USER_AGENT,
        "X-Line-Mid": p.MID,
    }
    fmt.Println("ChannelToken: ", channel.ChannelAccessToken)
}

func (p *LINE) TimelineRequests(url string, method string, data interface{}) (res *grequests.Response, err error){
	if method == "GET"{
		ro := &grequests.RequestOptions{
		    Headers: p.TimeLineHeader,
	    }
	    res, err := grequests.Get(url, ro)
	    return res, err
	}else if method == "POST"{
		ro := &grequests.RequestOptions{
		    Headers: p.TimeLineHeader,
		    JSON: data,
	    }
	    res, err := grequests.Post(url, ro)
	    return res, err
	}else {
		return &grequests.Response{}, nil
	}
}

func (p *LINE) DeletePost(pid string, to string) (interface {}, error) {
    if to == "" { to = p.MID}
    var data interface{}
    params := url.Values{}
    params.Add("homeId", to)
    params.Add("postId", pid)
    url := "https://gwx.line.naver.jp/mh/v48/post/delete.json?"+ params.Encode()
    r, err := p.TimelineRequests(url, "GET", data)
    r.JSON(data)
    return data, err
}

func (p *LINE) SendPostToTalk(mid string, postId string) (interface {}, error) {
	params := url.Values{}
    var data interface{}
    params.Add("receiveMid", mid)
    params.Add("postId", postId)
    url := "https://gwx.line.naver.jp/mh/api/v39/post/sendPostToTalk.json?"+ params.Encode()
    r, err := p.TimelineRequests(url, "GET", data)
    r.JSON(data)
    return data, err
}

func (p *LINE) GetGroupPost(groupid string) (interface{}, error) {
	params := url.Values{}
    var res interface{}
    params.Add("homeId", groupid)
    //params.Add("postLimit", "100")
    params.Add("commentLimit", "50")
    params.Add("likeLimit", "500")
    params.Add("sourceType", "TALKROOM")
    url := "https://gwx.line.naver.jp/mh/api/v39/post/list.json?"+ params.Encode()
    r, err := p.TimelineRequests(url, "GET", res)
    r.JSON(res)
    return res, err
}

func (p *LINE) LikePost(postId string, likeType string) (interface{}, error) {
	params := url.Values{}
	params.Add("homeId", p.MID)
	params.Add("sourceType", "TIMELINE")
    url := "https://gwx.line.naver.jp/mh/api/v39/like/create.json?"+ params.Encode()
    datas := map[string]string{"likeType": likeType, "activityExternalId": postId, "actorId": p.MID}
    r,err := p.TimelineRequests(url, "POST", datas)
    var data interface{}
    r.JSON(data)
    return data, err
}

func (p *LINE) CommantPost(postId string, text string) (interface{}, error) {
    params := url.Values{}
	params.Add("homeId", p.MID)
	params.Add("sourceType", "TIMELINE")
    url := "https://gwx.line.naver.jp/mh/api/v39/comment/create.json?"+ params.Encode()
    datas := map[string]string{"commentText": text, "activityExternalId": postId, "actorId": p.MID}
    r,err := p.TimelineRequests(url, "POST", datas)
    var data interface{}
    r.JSON(data)
    return data, err
}

func (p *LINE) CreatePost(text string, mid string) (interface{}, error) {
    params := url.Values{}
	params.Add("homeId", mid)
	params.Add("sourceType", "TIMELINE")
    url := "https://gwx.line.naver.jp/mh/api/v39/post/create.json?"+ params.Encode()
    var data interface{}
    strdata := "{\"postInfo\": {\"readPermission\": {\"type\": \"ALL\"}}, \"sourceType\": \"TIMELINE\", \"contents\": {\"text\": \""+text+"\"}}"
    json.Unmarshal([]byte(strdata), &data)
    r,err := p.TimelineRequests(url, "POST", data)
    r.JSON(data)
    return data, err
}

