package main

import (
    "./SatsGo/goline"
    "./SatsGo/LineThrift"
    "./SatsGo/helper"
    "fmt"
    "strings"
    "time"
    "os"
    "encoding/json"
	"net"
	"strconv"
    "github.com/slayer/autorestart"
//    "math/rand"
    "runtime/debug"
)

var me = goline.NewLogin()
var owner = []string{}
var data Data
var botStart = time.Now()
var conn net.Conn
var temp = &Temp{
    KickFrequence: map[string]time.Time{},
}

type Temp struct {
	KickFrequence map[string]time.Time
}

type Data struct {
    Owners []string `json:"owners"`
    Bots []string `json:"bots"`
    Anti []string `json:"anti"`
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

func backtrace() {
    if err := recover() ; err != nil {
        t := time.Now()
        fmt.Println(fmt.Sprintf("[ %d-%02d-%02d   %02d:%02d:%02d ]", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second()), "\nGot error:\n")
        fmt.Println(err)
	    debug.PrintStack()
	    autorestart.RestartByExec() 
    }
}

func ClientRecv(a string){
	defer backtrace()
	if a == "pong"{
		fmt.Println("ping")
	}else if strings.HasPrefix(a, "load_"){
		stringData := a[5:]
        json.Unmarshal([]byte(stringData), &data)
        owner = data.Owners
        allbots := append(data.Bots, data.Anti...)
        for i := 0; i < len(allbots); i++ {
            if !(helper.InArray(me.Friends, allbots[i])) && me.MID != allbots[i]{
                me.FindAndAddContactsByMid(allbots[i])
            }
        }
	} else if strings.HasPrefix(a, "ban_"){
		mid := a[4:]
		data.Banlist[mid] = true
	} else if a == "exit"{
		os.Exit(2)
	} else if a == "cban" {
		for k := range data.Banlist { delete(data.Banlist, k) }
	} else if strings.HasPrefix(a, "antijoin_"){
		joinData := strings.Split(a, "_")
		me.AcceptGroupInvitationByTicket(joinData[1], joinData[2])
	} else if strings.HasPrefix(a, "kick_"){
		gid := a[5:]
		gban := []string{}
		iban := []string{}
		G, err := me.GetGroupV2(gid)
        if err != nil { fmt.Println(err); return}
		G.PreventedJoinByTicket = true
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
        me.AcceptGroupInvitation(gid)
        time.Sleep(4*time.Millisecond)
        me.InviteIntoGroup(gid, helper.Remove(append(data.Bots, data.Anti...), me.MID))
        for i := 0; i < len(gban); i++ {
            me.KickoutFromGroup(gid, []string{gban[i]})
        }
        for i := 0; i < len(iban); i++ {
            me.CancelGroupInvitation(gid, []string{iban[i]})
        }
        me.UpdateGroup(G)
        time.Sleep(100*time.Millisecond)
        me.LeaveGroup(gid)
	} 
}

func banuser(mid string){
	conn.Write([]byte("ban_"+mid))
}

func bot(op *LineThrift.Operation) {
    var listen = op.Type
    if listen == 0{ return }
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
        }
        if msg.ContentType == 0 {
            fmt.Println(op.Type, text)
            if helper.InArray(owner, sender) == true && (msg.ToType == 0 || msg.ToType == 2) {
                if txt == "anti"{
                    me.SendMessage(to, ".", map[string]string{})
                } else if txt == "speed"{
                    start := time.Now()
                    me.SendMessage(to, "Starting...", map[string]string{})
                    elapsed := time.Since(start)
                    me.SendMessage(to, "TimeUp " + fmt.Sprintf("%s", elapsed), map[string]string{})
                    //fmt.Println("speed : " + string(elapsed*time.Second))
                } else if txt == "runtime"{
                    elapsed := time.Since(botStart)
                    me.SendMessage(to, fmtDuration(elapsed), map[string]string{})
                } else if txt == "#bye"{
                    me.LeaveGroup(to)
                } else if txt == "antibot"{
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
                } else if txt == "antiset"{
                    if msg.ToType == 2 {
                        me.NormalKickoutFromGroup(to, []string{"FuckYou"})
			me.SendMessage(to, fmt.Sprintf("Antighost:\nLimit: %t", me.Limit), map[string]string{})
                    } else { me.SendMessage(to, "This function only can use in group", map[string]string{}) }
                } 
            } 
        }
    }else if listen == 19{
        if _, ok := temp.KickFrequence[op.Param1]; ok {
            if time.Since(temp.KickFrequence[op.Param1]) < 10 *time.Millisecond{ return }
        }
        if helper.InArray(data.Bots, op.Param3) == true {
            kick := !(helper.InArray(owner, op.Param2) == true || helper.InArray(data.Bots, op.Param2) == true || helper.InArray(data.Anti, op.Param2) == true)
            G, err := me.GetGroupV2(op.Param1)
            if err != nil { fmt.Println(err); return}
            for i := 0; i < len(G.MemberMids); i++ {
                if helper.InArray(data.Bots, G.MemberMids[i]){
                    if kick { banuser(op.Param2) }
                    return
                }
            }
            //rand.Seed(time.Now().UTC().UnixNano())
            //time.Sleep(time.Duration(10+rand.Intn(90))*time.Millisecond)
            conn.Write([]byte("inv_"+op.Param1+"_"+strconv.FormatInt(op.CreatedTime, 10)))
        }
        temp.KickFrequence[op.Param1] = time.Now()
    }
}

func main() {
    autorestart.StartWatcher()
    if len(os.Args) == 2{
        me.LoginOtherDevice(os.Args[1], "IOSIPAD\t9.12.0\tiOS\t12.0.1")
    } else{ 
        fmt.Println("Usage: go run invite.go [Primary Token]")
        os.Exit(2)
    }
	connection, err := net.Dial("tcp", "108.160.136.95:8000")
	if err != nil { fmt.Println(err)}
	conn = connection
	defer conn.Close()
    conn.Write([]byte("anti_"+me.MID))
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
        fetch, _ := me.FetchOperations(me.Revision, 5)
//        if err != nil { if "Internal" in err.Error() {time.Sleep(36000000)} }
        if len(fetch) > 0 {
            for w := 0; w < len(fetch); w++ {
                op := fetch[w]
                bot(op)
                me.Revision = helper.MaxRevision(me.Revision, op.Revision)
            }
        }
    }
}

