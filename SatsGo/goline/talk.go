package goline

import (
	"../LineThrift"
	"context"
	"strings"
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
	"mime/multipart"
	"strconv"
    "encoding/json"
	"bytes"
	//"github.com/asmcos/requests"
	"io"
	"unicode/utf8"
	"github.com/levigross/grequests"
)

//var err error

type mention struct {
	S string `json:"S"`
	E string `json:"E"`
	M string `json:"M"`
}

// im initiating a new TalkService every function call
// because the bytes.Buffer is not thread-safe,
// and i don't know how to do that yet.

/* User Functions */

func (p *LINE) UpdateProfilePictureFromMsg(msgid string) (err error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://obs-sg.line-apps.com/talk/m/download.nhn?oid=" + msgid, nil)
    req.Header.Set("User-Agent",USER_AGENT)
    req.Header.Set("X-Line-Application",p.AppName)
    req.Header.Set("X-Line-Carrier",CARRIER)
    req.Header.Set("X-Line-Access",p.AuthToken)
	res, _ := client.Do(req)
    defer res.Body.Close()
    file, err := os.Create("temp.bin")
    io.Copy(file, res.Body)
    defer file.Close()
    p.test2File("temp.bin")
    return nil
}
/*
func (p *LINE) testFile(Filename string) error {
	req := requests.Requests()
	req.Header.Set("User-Agent",USER_AGENT)
    req.Header.Set("X-Line-Application",p.AppName)
    req.Header.Set("X-Line-Carrier",CARRIER)
    req.Header.Set("X-Line-Access",p.AuthToken)
    data := requests.Datas{
          "params": "{\"name\": \"linego.bin\", \"ver\": \"1.0\", \"oid\": \""+MID+"\", \"type\": \"image\"}",
    }
    resp, err := req.Post("https://obs-sg.line-apps.com/talk/p/upload.nhn", data, Filename)
    if err != nil{
        panic(err)
    }
    fmt.Println(resp.Text())
    return nil
}
*/

func (p *LINE) test2File(Filename string) error {
	file, _ := grequests.FileUploadFromDisk(Filename)
	file[0].FieldName = "file"
	data := map[string]string{
		"name": "linear-1576050907-8.bin",
		"ver": "1.0",
		"oid": p.MID,
		"type": "image",
	}
	datas,_ := json.Marshal(data)
	ro := &grequests.RequestOptions{
		Headers: map[string]string{
			"User-Agent": USER_AGENT,
			"X-Line-Application": p.AppName,
			"X-Line-Carrier": CARRIER,
			"X-Line-Access": p.AuthToken,
		},
		Data: map[string]string{"params": string(datas)},
		Files: file,
	}
	resp, err := grequests.Post("https://obs-sg.line-apps.com/talk/p/upload.nhn", ro)
	if err != nil{
        panic(err)
    }
    fmt.Println(resp.StatusCode)
    return nil
}

func (p *LINE) postFile(filename, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer file.Close()
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", filename)
	if err != nil {
		fmt.Println("error writing to buffer")
		return err
	}
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return err
	}
	bodyWriter.Close()
	bodyWriter.WriteField("params", "{\"name\": \"linego.bin\", \"ver\": \"1.0\", \"oid\": \""+p.MID+"\", \"type\": \"image\"}")
	fmt.Println(bodyBuf)
    client := &http.Client{}
    req, _ := http.NewRequest("POST", "https://obs-sg.line-apps.com/talk/p/upload.nhn", bodyBuf)
    req.Header.Set("User-Agent",USER_AGENT)
    req.Header.Set("X-Line-Application",p.AppName)
    req.Header.Set("X-Line-Carrier",CARRIER)
    req.Header.Set("X-Line-Access",p.AuthToken)
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	return nil
}

func (p *LINE) GetProfile() (r *LineThrift.Profile, err error) {
	res, err := p.Client.GetProfile(context.TODO())
	return res, err
}

func (p *LINE) GetSettings() (r *LineThrift.Settings, err error) {
	res, err := p.Client.GetSettings(context.TODO())
	return res, err
}

func (p *LINE) GetUserTicket() (r *LineThrift.Ticket, err error) {
	res, err := p.Client.GetUserTicket(context.TODO())
	return res, err
}

func (p *LINE) UpdateProfile(profile *LineThrift.Profile) (err error) {
	p.Client.UpdateProfile(context.TODO(), int32(0), profile)
	return err
}

func (p *LINE) UpdateSettings(settings *LineThrift.Settings) (err error) {
	p.Client.UpdateSettings(context.TODO(), int32(0), settings)
	return err
}

func (p *LINE) UpdateProfileAttribute(attr LineThrift.ProfileAttribute, value string) (err error) {
	p.Client.UpdateProfileAttribute(context.TODO(), int32(0), attr, value)
	return err
}

/* Fetch Functions */

func (p *LINE) FetchOperations(rev int64, count int32) (r []*LineThrift.Operation, err error) {
	//res, err := p.Poll.FetchOps(context.TODO(), rev, count, 0, 0)
	res, err := p.Poll.FetchOperations(context.TODO(), rev, count)
	return res, err
}

/* Message Functions */

func (p *LINE) SendMessage(to string, text string, contentMetadata map[string]string) (*LineThrift.Message, error) {
	M := &LineThrift.Message{
		From_: p.MID,
		To: to,
		Text: text,
		ContentType: 0,
		ContentMetadata: contentMetadata,
		RelatedMessageId: "0", // to be honest, i don't know what this is for, and if i don't throw something it wouldn't send the message
	}
	res, err := p.Client.SendMessage(context.TODO(), int32(0), M)
	return res, err
}

func (p *LINE) SendMention(to string, text string, mids []string) (*LineThrift.Message, error) {
	arr := []*mention{}
	mentionee := "@ZetsugLine"
	texts := strings.Split(text, "@!")
	if len(mids) == 0 || len(texts) < len(mids) { return &LineThrift.Message{}, fmt.Errorf("Invalid mids.") }
	textx := ""
	for i := 0; i < len(mids); i++ {
		textx += texts[i]
        arr = append(arr, &mention{S: strconv.Itoa(utf8.RuneCountInString(textx)), E: strconv.Itoa(utf8.RuneCountInString(textx) + 11), M:mids[i]})
        textx += mentionee
	}
	textx += texts[len(texts)-1]
	arrData,_ := json.Marshal(arr)
	mes, err := p.SendMessage(to, textx, map[string]string{"MENTION": "{\"MENTIONEES\":"+string(arrData)+"}"})
	return mes, err
}

func (p *LINE) UnsendMessage(messageId string) (err error) {
	p.Client.UnsendMessage(context.TODO(), int32(0), messageId)
	return err
}

func (p *LINE) RequestResendMessage(senderMid string, messageId string) (err error) {
	p.Client.RequestResendMessage(context.TODO(), int32(0), senderMid, messageId)
	return err
}

func (p *LINE) RespondResendMessage(receiverMid string, originalMessageId string, resendMessage *LineThrift.Message, errorCode LineThrift.ErrorCode) (err error) {
	p.Client.RespondResendMessage(context.TODO(), int32(0), receiverMid, originalMessageId, resendMessage, errorCode)
	return err
}

func (p *LINE) RemoveMessage(messageId string) (r bool, err error) {
	res, err := p.Client.RemoveMessage(context.TODO(), messageId)
	return res, err
}

func (p *LINE) RemoveAllMessages(lastMessageId string) (err error) {
	p.Client.RemoveAllMessages(context.TODO(), int32(0), lastMessageId)
	return err
}

func (p *LINE) RemoveMessageFromMyHome(ctx context.Context, messageId string) (r bool, err error) {
	res, err := p.Client.RemoveMessageFromMyHome(context.TODO(), messageId)
	return res, err
}

func (p *LINE) SendChatChecked(consumer string, lastMessageId string, sessionId int8) (err error) {
	p.Client.SendChatChecked(context.TODO(), int32(0), consumer, lastMessageId, sessionId)
	return err
}

func (p *LINE) SendEvent(message *LineThrift.Message) (r *LineThrift.Message, err error) {
	res, err := p.Client.SendEvent(context.TODO(), int32(0), message)
	return res, err
}

func (p *LINE) GetPreviousMessagesV2WithReadCount(messageBoxId string, endMessageId *LineThrift.MessageBoxV2MessageId, messagesCount int32) (r []*LineThrift.Message, err error) {
	res, err := p.Client.GetPreviousMessagesV2WithReadCount(context.TODO(), messageBoxId, endMessageId, messagesCount)
	return res, err
}

/* Contact Functions */

func (p *LINE) BlockContact(id string) (err error) {
	p.Client.BlockContact(context.TODO(), int32(0), id)
	return err
}

func (p *LINE) FindAndAddContactByMetaTag(userid string, reference string) (r *LineThrift.Contact, err error) {
	res, err := p.Client.FindAndAddContactByMetaTag(context.TODO(), int32(0), userid, reference)
	if err == nil { p.Friends, _ = p.GetAllContactIds() }
	return res, err
}

func (p *LINE) FindAndAddContactsByMid(mid string) (r map[string]*LineThrift.Contact, err error) {
	res, err := p.Client.FindAndAddContactsByMid(context.TODO(), int32(0), mid)
	if err == nil { p.Friends = append(p.Friends, mid)}
	return res, err
}

//
func (p *LINE) FindAndAddContactsByMid_thread(mid string) (r map[string]*LineThrift.Contact, err error) {
	res, err := p.TalkService().FindAndAddContactsByMid(context.TODO(), int32(0), mid)
	if err == nil { p.Friends = append(p.Friends, mid)}
	return res, err
}

func (p *LINE) FindAndAddContactsByEmail(emails []string) (r map[string]*LineThrift.Contact, err error) {
	res, err := p.Client.FindAndAddContactsByEmail(context.TODO(), int32(0), emails)
	if err == nil { p.Friends, _ = p.GetAllContactIds() }
	return res, err
}

func (p *LINE) FindAndAddContactsByUserid(userid string) (r map[string]*LineThrift.Contact, err error) {
	res, err := p.Client.FindAndAddContactsByUserid(context.TODO(), int32(0), userid)
	if err == nil { p.Friends, _ = p.GetAllContactIds() }
	return res, err
}

func (p *LINE) GetAllContactIds() (r []string, err error) {
	res, err := p.Client.GetAllContactIds(context.TODO())
	return res, err
}

func (p *LINE) GetBlockedContactIds() (r []string, err error) {
	res, err := p.Client.GetBlockedContactIds(context.TODO())
	return res, err
}

func (p *LINE) GetContact(id string) (r *LineThrift.Contact, err error) {
	res, err := p.Client.GetContact(context.TODO(), id)
	return res, err
}

func (p *LINE) GetContacts(ids []string) (r []*LineThrift.Contact, err error) {
	res, err := p.Client.GetContacts(context.TODO(), ids)
	return res, err
}

func (p *LINE) GetFavoriteMids() (r []string, err error) {
	res, err := p.Client.GetFavoriteMids(context.TODO())
	return res, err
}

func (p *LINE) GetHiddenContactMids() (r []string, err error) {
	res, err := p.Client.GetHiddenContactMids(context.TODO())
	return res, err
}

/* Group Functions */

func (p *LINE) CancelGroupInvitation(groupId string, contactIds []string) (err error) {
	go p.TalkService().CancelGroupInvitation(context.TODO(), int32(0), groupId, contactIds)
	return err
}

func (p *LINE) KickoutFromGroup(groupId string, contactIds []string) (err error) {
	go p.TalkService().KickoutFromGroup(context.TODO(), int32(0), groupId, contactIds) 
	return err
}

func (p *LINE) NormalKickoutFromGroup(groupId string, contactIds []string) (err error) {
	res := p.Client.KickoutFromGroup(context.TODO(), int32(0), groupId, contactIds)
	if strings.Contains(res.Error(), "request blocked") {
		p.Limit = true
	}
	return res
}

func (p *LINE) InviteIntoGroup(groupId string, contactIds []string) (err error) {
	go p.TalkService().InviteIntoGroup(context.TODO(), int32(0), groupId, contactIds)
	return err
}

func (p *LINE) NormalInviteIntoGroup(groupId string, contactIds []string) (err error) {
	res := p.Client.InviteIntoGroup(context.TODO(), int32(0), groupId, contactIds)
	if strings.Contains(res.Error(), "request blocked") {
		p.Limit = true
	}
	return res
}

func (p *LINE) AcceptGroupInvitation(groupId string) (err error) {
	go p.TalkService().AcceptGroupInvitation(context.TODO(), int32(0), groupId)
	return err
}

func (p *LINE) AcceptGroupInvitationByTicket(groupId string, ticketId string) (err error) {
	go p.TalkService().AcceptGroupInvitationByTicket(context.TODO(), int32(0), groupId, ticketId)
	return err
}

func (p *LINE) LeaveGroup(groupId string) (err error) {
	go p.TalkService().LeaveGroup(context.TODO(), int32(0), groupId)
	return err
}

// 
func (p *LINE) FindGroupByTicket(ticketId string) (r *LineThrift.Group, err error) {
        res, err := p.TalkService().FindGroupByTicket(context.TODO(),  ticketId)
        return res, err
}

func (p *LINE) GetGroup(groupId string) (r *LineThrift.Group, err error) {
	res, err := p.Client.GetGroup(context.TODO(), groupId)
	return res, err
}

func (p *LINE) GetGroupWithoutMembers(groupId string) (r *LineThrift.Group, err error) {
	res, err := p.Client.GetGroupWithoutMembers(context.TODO(), groupId)
	return res, err
}

func (p *LINE) GetCompactGroup(groupId string) (r *LineThrift.Group, err error) {
	res, err := p.Client.GetCompactGroup(context.TODO(), groupId)
	return res, err
}

func (p *LINE) GetGroupV2(groupId string) (r *LineThrift.Group, err error) {
	res, err := p.Client.GetGroupsV2(context.TODO(), []string{groupId})
	if err != nil { return &LineThrift.Group{}, err}
	return res[0], err
}

func (p *LINE) ReissueGroupTicket(groupId string) (r string, err error) {
	res, err := p.Client.ReissueGroupTicket(context.TODO(), groupId)
	if err != nil { fmt.Println(err.Error())}
	return res, err
}

func (p *LINE) UpdateGroup(group *LineThrift.Group) (err error) {
	go p.TalkService().UpdateGroup(context.TODO(), int32(0), group)
	return err
}

// 
func (p *LINE) VerifyQrcode(verifier string, pinCode string) (r string, err error) {
        res, err := p.TalkService().VerifyQrcode(context.TODO(), verifier, pinCode)
        return res, err
}

/* Others */

func (p *LINE) Noop() error {
	err := p.Client.Noop(context.TODO())
	return err
}

func (p *LINE) Test(groupId string) (r []string, err error) {
	res, err := p.CallService().GetGroupMemberMidsForAppPlatform(context.TODO(), groupId)
	return res, err
}
