package main

import (
    "./SatsGo/goline"
    "./SatsGo/LineThrift"
    "./SatsGo/helper"
    "fmt"
    "strings"
    "time"
    "io/ioutil"
    "os"
    "encoding/json"
	"net"
	"strconv"
    "github.com/slayer/autorestart"
    "runtime/debug"
	"runtime"
)

var me = goline.NewLogin()
var owner = []string{}
var data Data
var botStart = time.Now()
var conn net.Conn
var temp = &Temp{
    KickFrequence: map[string]time.Time{},
    WarTime: map[string]time.Time{},
	ContactBan: false,
	ContactUnban: false,
	Prokick: []string{},
}

type Temp struct {
	KickFrequence map[string]time.Time
	WarTime map[string]time.Time
	ContactBan bool
	ContactUnban bool
	Prokick []string
}

type Group struct {
	Name string `json:"name"`
	PreventJoin bool `json:"preventJoin"`
	Ticket string `json:"ticket"`
	Recieve time.Time `json:"recieve"`
	Invprotect bool `json:"invprotect"`
	Qrprotect bool `json:"qrprotect"`
	Protect bool `json:"protect"`
	Manager map[string]bool `json:"manager"`
}

type Data struct {
    Owners []string `json:"owners"`
    Bots []string `json:"bots"`
    Anti []string `json:"anti"`
    Set map[string]*Group `json:"settings"`
    SystemMode string `json:"systemMode"`
    Banlist map[string]bool `json:"ban"`
}

func fmtDuration(d time.Duration) string {
    d = d.Round(time.Second)
    h := d / time.Hour
    d -= h * time.Hour
    m := d / time.Minute
    d -= m * time.Minute
    s := d / time.Second
    return fmt.Sprintf("%02d day %02d hour %02d min %02d sec.", h/24, h%24, m, s)
}
func save(){
	file, _ := json.MarshalIndent(data, "", "  ")
	_ = ioutil.WriteFile("data.json", file, 0644)
}
func ClientRecv(a string){
	if strings.HasPrefix(a, "load_"){
		stringData := a[5:]
        json.Unmarshal([]byte(stringData), &data)
        owner = data.Owners
        allbots := append(data.Bots, data.Anti...)
        for i := 0; i < len(allbots); i++ {
            if !(helper.InArray(me.Friends, allbots[i])) && me.MID != allbots[i]{
                me.FindAndAddContactsByMid(allbots[i])
            }
        }
        getProkick()
	} else if strings.HasPrefix(a, "ban_"){
		mid := a[4:37]
		data.Banlist[mid] = true
	} else if a == "exit"{
		os.Exit(2)
	} else if strings.HasPrefix(a, "join_"){
		joinData := strings.Split(a, "_")
		if _, ok := data.Set[joinData[1]] ; !ok || data.Set[joinData[1]].Ticket == ""{
			me.AcceptGroupInvitationByTicket(joinData[1], joinData[2])
		}
	} else if a == "cbn" {
		for k := range data.Banlist { delete(data.Banlist, k) }
	} else if a == "updateall" {
		for k := range data.Set {
			byteData, _ := json.Marshal(data.Set[k])
		    conn.Write([]byte("group_"+k+"_"+ string(byteData)))
		    time.Sleep(100*time.Millisecond)
		}
	} else if a == "mode purge" {
		data.SystemMode = "purge"
		getProkick()
	} else if a == "mode normal" {
		data.SystemMode = "normal"
		getProkick()
	} else if a == "mode war" {
		data.SystemMode = "war"
		getProkick()
	}
}

func banuser(mid string){
	if _, ok := data.Banlist[mid]; !ok {
		conn.Write([]byte("ban_"+mid))
	}
}

func getProkick() {
	index := helper.IndexOf(me.MID, data.Bots)
    if data.SystemMode == "purge" {
	    temp.Prokick = []string{helper.Shift(data.Bots, 1)[index]}
    }else {
		if len(data.Bots) > 3 {
			temp.Prokick = helper.Shift(data.Bots, index)[1:4]
		} else {
			temp.Prokick = data.Bots
		}
	}
}

func backtrace() {
    if err := recover() ; err != nil {
        t := time.Now()
        fmt.Println(fmt.Sprintf("[ %d-%02d-%02d   %02d:%02d:%02d ]", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()), "\nGot error:\n")
        fmt.Println(err)
	    debug.PrintStack()
	    autorestart.RestartByExec() 
    }
}

func bot(op *LineThrift.Operation) {
	defer backtrace()
	runtime.GOMAXPROCS(20)
    var listen = op.Type
    if listen == 0{ return }
    fmt.Println(listen)
    if listen == 26{
        msg := op.Message
        text := msg.Text
        sender := msg.From_
        receiver := msg.To
        txt := strings.ToLower(text)
        var to = sender
        if msg.ToType == 0 {
            to = sender
        } else {
            to = receiver
            if _, ok := data.Set[to] ; !ok{
                G, _ := me.GetGroupWithoutMembers(to)
                data.Set[to] = &Group{Name: G.Name, PreventJoin: G.PreventedJoinByTicket, Ticket: "", Recieve: botStart, Invprotect: false, Qrprotect: false, Protect:true, Manager:map[string]bool{}}
                data.Set[to].Ticket, _ = me.ReissueGroupTicket(to)
            }
            if time.Since(data.Set[to].Recieve) < 1 *time.Second{ return }
            data.Set[to].Recieve = time.Now()
        }
        if msg.ContentType == 0 {
            if helper.InArray(owner, sender) == true && (msg.ToType == 0 || msg.ToType == 2) {
                if strings.HasPrefix(txt, "upname:"){
                    name := text[7:]
                    profile, _ := me.GetProfile()
                    profile.DisplayName = name
                    me.UpdateProfile(profile)
                    me.SendMessage(to, "Successed rename as "+name, map[string]string{})
                } else if txt == "sp"{
                    start := time.Now()
                    //me.SendMessage(to, "Starting...", map[string]string{})
                    elapsed := time.Since(start)
                    me.SendMessage(to, "Speed" + fmt.Sprintf("%s", elapsed), map[string]string{})
                } else if txt == "runtime"{
                    elapsed := time.Since(botStart)
                    me.SendMessage(to, fmtDuration(elapsed), map[string]string{})
                } else if txt == "listbot"{
		    res := "List bot:\n"
                    for x,y := range data.Bots{
                        mc, err := me.GetContact(y); if err != nil { continue }
                        res += fmt.Sprintf("\n%s. %s", strconv.Itoa(x+1), mc.DisplayName)
                    }
		    res += "\nList ghost:\n"
                    for x,y := range data.Anti{
                        mc, err := me.GetContact(y); if err != nil { continue }
                        res += fmt.Sprintf("\n%s. %s", strconv.Itoa(x+1), mc.DisplayName)
                    }
                    me.SendMessage(to, res, map[string]string{})
                } else if txt == "come"{
                    if msg.ToType == 2 {
                        if data.Set[to].PreventJoin == true {
                            data.Set[to].PreventJoin = false
                            me.UpdateGroup(&LineThrift.Group{ID: to, Name: data.Set[to].Name, PreventedJoinByTicket: false, PictureStatus: "1234567890"})
                        }
                        if data.Set[to].Ticket == "" { data.Set[to].Ticket, _ = me.ReissueGroupTicket(to) }
                        conn.Write([]byte("join_"+to+"_"+data.Set[to].Ticket))
                    }
                } else if txt == "go" {
                    me.SendMention(to, "@!", []string{sender})
                } else if txt == "group" {
		    ret := "List groups:"
                    for x := range data.Set{
                        ret += "\n- " + data.Set[x].Name
                    }
                    me.SendMessage(to, ret, map[string]string{})
                } else if strings.HasPrefix(txt, "njoin:"){
                    for x := range data.Set{
                        if strings.HasPrefix(data.Set[x].Name, text[6:]){
                            me.SendMessage(to, "https://line.me/R/ti/g/" + data.Set[x].Ticket, map[string]string{})
                        }
                    }
                }
            } 
            if helper.InArray(owner, sender) == true || (msg.ToType == 2 && data.Set[to].Manager[sender] == true) {
                if txt == "res"{
                    me.SendMessage(to, "􀌃􀆁􏿿", map[string]string{})
                } else if txt == "mid"{
		    me.SendMessage(to, "mid : " + sender, map[string]string{})
                } else if strings.HasPrefix(txt, "kick ") && !(msg.ContentMetadata["MENTION"] == ""){
                    value := helper.GetMidFromMentionees(msg.ContentMetadata["MENTION"])
                    for i := 0; i < len(value); i++ {
                        banuser(value[i])
                        me.KickoutFromGroup(to, []string{value[i]})
                    }
                    temp.WarTime[to] = time.Now()
                } else if strings.HasPrefix(txt, "gadd ") && !(msg.ContentMetadata["MENTION"] == ""){
                    value := helper.GetMidFromMentionees(msg.ContentMetadata["MENTION"])
                    for i := 0; i < len(value); i++ {
                        data.Set[to].Manager[value[i]] = true
                    } 
                    me.SendMessage(to, "Added user manager.", map[string]string{})
                } else if strings.HasPrefix(txt, "gdel ") && !(msg.ContentMetadata["MENTION"] == ""){
                    value := helper.GetMidFromMentionees(msg.ContentMetadata["MENTION"])
                    for i := 0; i < len(value); i++ {
                        delete(data.Set[to].Manager, value[i])
                    } 
                    me.SendMessage(to, "Delete user manager.", map[string]string{})
                } else if txt == "gm"{
		    res := "List manager:\n"
                    for x := range data.Set[to].Manager {
                        mc, err := me.GetContact(x); if err != nil { continue }
                        res += fmt.Sprintf("\n- %s", mc.DisplayName)
                    }
                    me.SendMessage(to, res, map[string]string{})
                } else if txt == "left"{
                    data.Set[to].Ticket = ""
                    me.LeaveGroup(to)
                } else if txt == "sets"{
                    if msg.ToType == 2 {
                        me.NormalKickoutFromGroup(to, []string{"FuckYou"})
			me.SendMessage(to, fmt.Sprintf("Setting bot:\nLimit: %t\nMode: %s\nPortect: %t\nInvPortect: %t\nQrPortect: %t", me.Limit, data.SystemMode, data.Set[to].Protect, data.Set[to].Invprotect, data.Set[to].Qrprotect), map[string]string{})
                    } else { me.SendMessage(to, "This function only can use in group", map[string]string{}) }
                } else if txt == "cek"{
                    if msg.ToType == 2 {
                        me.NormalKickoutFromGroup(to, []string{"FuckYou"})
                me.SendMessage(to, fmt.Sprintf("Info\nStatus: %t", me.Limit), map[string]string{})
                        } else { me.SendMessage(to, "This function only can use in group", map[string]string{}) }
                } else if txt == "invprotect on"{
                    data.Set[to].Invprotect = true
                    me.SendMessage(to, txt + " invite protect enable.", map[string]string{})
                } else if txt == "invprotect off"{
                    data.Set[to].Invprotect = false
                    me.SendMessage(to, txt + " invite protect disable.", map[string]string{})
                } else if txt == "qrprotect on"{
                    data.Set[to].Qrprotect = true
                    me.SendMessage(to, txt + " qr protect enable.", map[string]string{})
                } else if txt == "qrprotect off"{
                    data.Set[to].Qrprotect = false
                    me.SendMessage(to, txt + " qr protect disable.", map[string]string{})
                } else if txt == "protect on"{
                    data.Set[to].Protect = true
                    me.SendMessage(to, txt + " protect enable.", map[string]string{})
                } else if txt == "protect off"{
                    data.Set[to].Protect = false
                    me.SendMessage(to, txt + " protect disable.", map[string]string{})
                } else if txt == "pro on"{
                    data.Set[to].Protect = true
                    data.Set[to].Qrprotect = true
                    data.Set[to].Invprotect = true
                    me.SendMessage(to, "protect enable.", map[string]string{})
                } else if txt == "pro off"{
                    data.Set[to].Protect = false
                    data.Set[to].Qrprotect = false
                    data.Set[to].Invprotect = false
                    me.SendMessage(to, "protect disable.", map[string]string{})
                } else if txt == "ban"{
                    temp.ContactBan = true
                    me.SendMessage(to, "Please send contact.", map[string]string{})
                }  else if txt == "unban"{
                    temp.ContactUnban = true
                    me.SendMessage(to, "Please send contact.", map[string]string{})
                }else if txt == "backup"{
                    byteData, _ := json.Marshal(data.Set[to])
	                conn.Write([]byte("group_"+to+"_"+ string(byteData)))
                    me.SendMessage(to, "Backup group protection.", map[string]string{})
                }
            }
        } else if msg.ContentType == 1 {
            me.UpdateProfilePictureFromMsg(msg.ID)
        } else if msg.ContentType == 13 {
            if temp.ContactBan == true {
                banuser(msg.ContentMetadata["mid"])
                me.SendMessage(to, "Banned user success.", map[string]string{})
            } else if temp.ContactBan == true {
                delete(data.Banlist, msg.ContentMetadata["mid"])
                me.SendMessage(to, "Unbanned user success.", map[string]string{})
            }
        }
    }else if listen == 13{
        param3 := strings.Split(op.Param3, "\x1e")
        if helper.InArray(param3, me.MID) {
            me.AcceptGroupInvitation(op.Param1)
            if !(helper.InArray(owner, op.Param2) == true || helper.InArray(data.Bots, op.Param2) == true || helper.InArray(data.Anti, op.Param2) == true) {
                data.Set[op.Param1].Manager[op.Param2] = true
            }
            return
        }
        if _, ok := data.Set[op.Param1].Manager[op.Param2]; helper.InArray(owner, op.Param2) == true || helper.InArray(data.Bots, op.Param2) == true || helper.InArray(data.Anti, op.Param2) == true || ok {
            return
        } else if data.SystemMode == "purge" {
            if data.Banlist[op.Param2] == true{
            	for i := 0; i < len(param3); i++ {
            		banuser(param3[i])
            		//me.CancelGroupInvitation(op.Param1, []string{param3[i]})
            	}
            }
            if data.Banlist[op.Param3] == true{
                banuser(op.Param2)
                //me.KickoutFromGroup(op.Param1, []string{op.Param2})
            }
        } else if _, ok := temp.WarTime[op.Param1]; (ok && time.Since(temp.WarTime[op.Param1]) < 1000 *time.Millisecond) || data.Set[op.Param1].Invprotect == true{
            banuser(op.Param2)
            if data.SystemMode == "war" {
            	for k := range data.Banlist {
            		me.CancelGroupInvitation(op.Param1, []string{k})
            	}
            } else {
            	for i := 0; i < len(param3); i++ {
            		banuser(param3[i])
            		me.CancelGroupInvitation(op.Param1, []string{param3[i]})
            	}
            }
            me.KickoutFromGroup(op.Param1, []string{op.Param2})
            temp.WarTime[op.Param1] = time.Now()
        }
        if _, ok := data.Set[op.Param1] ; !ok{
            G, _ := me.GetGroupWithoutMembers(op.Param1)
            data.Set[op.Param1] = &Group{Name: G.Name, PreventJoin: G.PreventedJoinByTicket, Ticket: "", Recieve: botStart, Invprotect: false, Qrprotect: false, Protect:true, Manager:map[string]bool{}}
            data.Set[op.Param1].Ticket, _ = me.ReissueGroupTicket(op.Param1)
        }
    }else if listen == 11{
        if _, ok := data.Set[op.Param1] ; !ok{
            G, _ := me.GetGroupWithoutMembers(op.Param1)
            data.Set[op.Param1] = &Group{Name: G.Name, PreventJoin: G.PreventedJoinByTicket, Ticket: "", Recieve: botStart, Invprotect: false, Qrprotect: false, Protect:true, Manager:map[string]bool{}}
            data.Set[op.Param1].Ticket, _ = me.ReissueGroupTicket(op.Param1)
        }
        if _, ok := data.Set[op.Param1].Manager[op.Param2]; !(ok || helper.InArray(owner, op.Param2) == true || helper.InArray(data.Bots, op.Param2) == true){
            if data.SystemMode == "purge" {
                banuser(op.Param2)
            } else if _, ok := temp.WarTime[op.Param1]; (ok && time.Since(temp.WarTime[op.Param1]) < 1000 *time.Millisecond) || data.Set[op.Param1].Qrprotect == true {
                me.KickoutFromGroup(op.Param1, []string{op.Param2})
                me.UpdateGroup(&LineThrift.Group{ID: op.Param1, Name: data.Set[op.Param1].Name, PreventedJoinByTicket: true, PictureStatus: "1234567890"})
                banuser(op.Param2)
            }
        }
        G, _ := me.GetGroupWithoutMembers(op.Param1)
        data.Set[op.Param1].Name = G.Name
        data.Set[op.Param1].PreventJoin = G.PreventedJoinByTicket
    }else if listen == 19 {
        if _, ok := data.Set[op.Param1] ; !ok{
            G, _ := me.GetGroupWithoutMembers(op.Param1)
            data.Set[op.Param1] = &Group{Name: G.Name, PreventJoin: G.PreventedJoinByTicket, Ticket: "", Recieve: botStart, Invprotect: false, Qrprotect: false, Protect:true, Manager:map[string]bool{}}
            data.Set[op.Param1].Ticket, _ = me.ReissueGroupTicket(op.Param1)
        } 
        _, ok := data.Set[op.Param1].Manager[op.Param2]
        kick := !(helper.InArray(owner, op.Param2) == true || helper.InArray(data.Bots, op.Param2) == true || helper.InArray(data.Anti, op.Param2) == true || ok )
        if op.Param3 == me.MID{
            data.Set[op.Param1].Ticket = ""
            if kick { banuser(op.Param2) }
            return
        } else if helper.InArray(temp.Prokick, op.Param3) == true {
            if kick {
                banuser(op.Param2)
                temp.WarTime[op.Param1] = time.Now()
                if data.SystemMode == "purge" {
                    time.Sleep(100*time.Millisecond)
                    G, err := me.GetGroupV2(op.Param1)
                    if err != nil { fmt.Println(err); return}
                    G.PreventedJoinByTicket = true
                    gban := []string{}
	                iban := []string{}
                    for i := 0; i < len(G.MemberMids); i++ {
                        if _, ok := data.Banlist[G.MemberMids[i]] ; ok{
                            gban = append(gban, G.MemberMids[i])
                        }
                    }
                    for i := 0; i < len(G.InviteeMids); i++ {
                        if _, ok := data.Banlist[G.InviteeMids[i]] ; ok{
                            iban = append(iban, G.InviteeMids[i])
                        }
                    }
                    for i := 0; i < len(gban); i++ {
                        me.KickoutFromGroup(op.Param1, []string{gban[i]})
                    }
                    for i := 0; i < len(iban); i++ {
                        me.CancelGroupInvitation(op.Param1, []string{iban[i]})
                    }
                    me.InviteIntoGroup(op.Param1, helper.Remove(append(data.Bots, data.Anti...), me.MID))
                    me.UpdateGroup(G)
                } else {
                    if _, ok := temp.KickFrequence[op.Param2]; ok {
                        if time.Since(temp.KickFrequence[op.Param2]) < 4 *time.Millisecond{ return }
                    }
                    G, _ := me.GetGroupWithoutMembers(op.Param1)
                    if me.Limit || data.Set[op.Param1].Protect == false {
                    	if data.Set[op.Param1].PreventJoin == true {
                    		if data.Set[op.Param1].Ticket == ""{
                    			data.Set[op.Param1].Ticket, _ = me.ReissueGroupTicket(op.Param1)
                    		}
                    	}
                    	data.Set[op.Param1].PreventJoin = false
                    	me.UpdateGroup(&LineThrift.Group{ID: op.Param1, Name: data.Set[op.Param1].Name, PreventedJoinByTicket: false, PictureStatus: "1234567890"})
                    	data.Set[op.Param1].Ticket, _ = me.ReissueGroupTicket(op.Param1)
                    	conn.Write([]byte("join_"+op.Param1+"_"+data.Set[op.Param1].Ticket))
                    	time.Sleep(1*time.Millisecond)
                    }else {
                    	if data.SystemMode == "normal" {
                    		me.InviteIntoGroup(op.Param1, helper.Remove(append(data.Bots, data.Anti...), me.MID))
                    		me.KickoutFromGroup(op.Param1, []string{op.Param2})
                    		temp.KickFrequence[op.Param2] = time.Now()
                    	} else {
                    		for k := range data.Banlist {
                    			me.KickoutFromGroup(op.Param1, []string{k})
                    			temp.KickFrequence[k] = time.Now()
                    		}
                    		if G.PreventedJoinByTicket == false{
                    			me.UpdateGroup(&LineThrift.Group{ID: op.Param1, Name: data.Set[op.Param1].Name, PreventedJoinByTicket: false, PictureStatus: "1234567890"})
                    			conn.Write([]byte("join_"+op.Param1+"_"+data.Set[op.Param1].Ticket))
                    		} else {
                    			me.InviteIntoGroup(op.Param1, helper.Remove(data.Bots, me.MID))
                    		}
                    	}
                    }
                }
            }
        } else if kick && data.Set[op.Param1].Protect == true && data.SystemMode == "normal"{ me.KickoutFromGroup(op.Param1, []string{op.Param2}); banuser(op.Param2) }
        if data.SystemMode != "purge" {
            if _, ok := data.Set[op.Param1].Manager[op.Param3]; ok || helper.InArray(owner, op.Param3) == true{
                mids,_:= me.GetAllContactIds()
                if !helper.InArray(mids, op.Param3){ me.FindAndAddContactsByMid(op.Param3)}
                me.InviteIntoGroup(op.Param1, []string{op.Param3})
            }
         }
    }else if listen == 17{
        if data.SystemMode == "purge" {
            return
        }
        if _, ok := data.Banlist[op.Param2]; ok{
            me.KickoutFromGroup(op.Param1, []string{op.Param2})
            if _, ok := data.Set[op.Param1] ; ok{
                if data.Set[op.Param1].PreventJoin == false{
                    me.UpdateGroup(&LineThrift.Group{ID: op.Param1, Name: data.Set[op.Param1].Name, PreventedJoinByTicket: true, PictureStatus: "1234567890"})
                }
            }else {
                G, _ := me.GetGroupWithoutMembers(op.Param1)
                if G.PreventedJoinByTicket == false{
                    G.PreventedJoinByTicket = true
                    me.UpdateGroup(G)
                }
                data.Set[op.Param1] = &Group{Name: G.Name, PreventJoin: true, Ticket: "", Recieve: botStart, Invprotect: true, Qrprotect: true, Protect:true, Manager:map[string]bool{}}
                data.Set[op.Param1].Ticket, _ = me.ReissueGroupTicket(op.Param1)
            }
        }else if _, ok := temp.WarTime[op.Param1]; ok {
            if time.Since(temp.WarTime[op.Param1]) < 1000 *time.Millisecond{ 
                if !(helper.InArray(owner, op.Param2) == true || helper.InArray(data.Bots, op.Param2) == true || helper.InArray(data.Anti, op.Param2) == true) {
                    banuser(op.Param2)
                    me.KickoutFromGroup(op.Param1, []string{op.Param2}) 
                }
            }
        }
    } else if listen == 32{
        if _, ok := data.Set[op.Param1] ; !ok{
            G, _ := me.GetGroupWithoutMembers(op.Param1)
            data.Set[op.Param1] = &Group{Name: G.Name, PreventJoin: G.PreventedJoinByTicket, Ticket: "", Recieve: botStart, Invprotect: true, Qrprotect: true, Protect:true, Manager:map[string]bool{}}
            data.Set[op.Param1].Ticket, _ = me.ReissueGroupTicket(op.Param1)
        }
        if _, ok := data.Set[op.Param1].Manager[op.Param2]; !(helper.InArray(owner, op.Param2) == true || helper.InArray(data.Bots, op.Param2) == true || ok ){
            if data.SystemMode == "purge" {
                banuser(op.Param2)
                return
            }else if _, ok := temp.WarTime[op.Param1]; (ok && time.Since(temp.WarTime[op.Param1]) < 1000 *time.Millisecond) || data.Set[op.Param1].Invprotect == true{
            	banuser(op.Param2)
            	me.KickoutFromGroup(op.Param1, []string{op.Param2})
            	G, _ := me.GetGroupWithoutMembers(op.Param1)
                if G.PreventedJoinByTicket == true{
                    G.PreventedJoinByTicket = false
                }
                me.UpdateGroup(&LineThrift.Group{ID: op.Param1, Name: data.Set[op.Param1].Name, PreventedJoinByTicket: false, PictureStatus: "1234567890"})
                conn.Write([]byte("join_"+op.Param1+"_"+data.Set[op.Param1].Ticket))
            }
            if helper.InArray(data.Anti, op.Param3) == true {
                me.KickoutFromGroup(op.Param1, []string{op.Param2})
                me.InviteIntoGroup(op.Param1, data.Anti)
            }
        }
    }
    save()
}

func main() {
    autorestart.StartWatcher()
    if len(os.Args) == 2{
        me.LoginOtherDevice(os.Args[1], "IOS 10.1.1 iPhone OS 1")
    } else{ 
        fmt.Println("Usage: go run sb.go [Primary Token]")
        os.Exit(2)
    }
	connection, err := net.Dial("tcp", "108.61.126.129:8000")
	if err != nil { fmt.Println(err)}
	conn = connection
	defer conn.Close()
	if me.MID == "" { os.Exit(2) }
    conn.Write([]byte("asis_"+me.MID))
    go func() {
        for{
            recv := ""
            for {
		        buf := make([]byte, 1024)
		        reqLen, err := conn.Read(buf)
		        recv += string(buf[:reqLen])
		        if err != nil || len(buf[:reqLen]) != 1024 { break }
	        }
	        if recv != "" {
	            ClientRecv(recv)
			    fmt.Println(recv)
	            recv = ""
	        }
        }
    }()
    for {
        fetch, err := me.FetchOperations(me.Revision, 5)
        if talkexception, ok := err.(*LineThrift.TalkException); ok {
        	if talkexception.Code == 20 {
                time.Sleep(1*time.Hour)
        	} else if talkexception.Code == 8 {
                time.Sleep(1*time.Hour)
        		me.LoginOtherDevice(os.Args[1], "IOS 10.1.1 iPhone OS 1")
        	}
        }
        if err == nil {
            if len(fetch) > 0 {
                for w := 0; w < len(fetch); w++ {
                    op := fetch[w]
                    bot(op)
                    me.Revision = helper.MaxRevision(me.Revision, op.Revision)
                    save()
                }
            }
        }
    }
}

