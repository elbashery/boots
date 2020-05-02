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
	"net"
    "encoding/json"
    "os/exec"
    "strconv"
    "github.com/slayer/autorestart"
    "math/rand"
    "math/big"
)

var me = goline.NewLogin()
var owner = []string{}
var data = &Data {
    Owners: []string{"u317784607b57c5d953852660d99ca3df","ue30f597601dbead84804c31fa8f10175"},
    Bots: []string{},
    Anti: []string{},
    Set: map[string]*Group{},
    Banlist: map[string]bool{},
    SystemMode: "normal",
    Login: map[string]*Bot{},
}
var botStart = time.Now()
var Running = map[string]net.Conn{}
var Dealed = big.NewInt(0)

type Data struct {
    Owners []string `json:"owners"`
    Bots []string `json:"bots"`
    Anti []string `json:"anti"`
    Set map[string]*Group `json:"settings"`
    Banlist map[string]bool `json:"ban"`
    SystemMode string `json:"systemMode"`
    Login map[string]*Bot `json:"login"`
}

type Bot struct {
	Token string `json:"token"`
	Running bool `json:"running"`
	Type int `json:"type"`
}

type Group struct {
	Name string `json:"name"`
	PreventJoin bool `json:"preventJoin"`
	Recieve time.Time `json:"recieve"`
	Invprotect bool `json:"invprotect"`
	Qrprotect bool `json:"qrprotect"`
	Protect bool `json:"protect"`
	Manager map[string]bool `json:"manager"`
}

func restart() {
    procAttr := new(os.ProcAttr)
    procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
    os.StartProcess(os.Args[0], []string{os.Args[1]}, procAttr)
}

func fmtDuration(d time.Duration) string {
    d = d.Round(time.Second)
    h := d / time.Hour
    d -= h * time.Hour
    m := d / time.Minute
    d -= m * time.Minute
    s := d / time.Second
    return fmt.Sprintf("%02d hour %02d min %02d sec.", h, m, s)
}

func save(){
	file, _ := json.MarshalIndent(data, "", "  ")
	_ = ioutil.WriteFile("data.json", file, 0644)
}

func broadcast(data string){
	for x := range Running{
		Running[x].Write([]byte(data))
	}
}

func handleConnection(conn net.Conn) {
	bs := make([]byte, 1024)
	len, err := conn.Read(bs)
	if err != nil { fmt.Printf("Connection error: %s\n", err) }
	recv := string(bs[:len])
	mid := ""
	if strings.HasPrefix(recv, "asis_"){
		mid = recv[5:]
		fmt.Printf("system running - %s\n", mid)
		if _, ok := Running[mid]; ok{ Running[mid].Write([]byte("exit")); Running[mid].Close() }
		Running[mid] = conn
        if helper.InArray(data.Anti, mid){ 
            data.Anti = helper.Remove(data.Anti, mid)
        }
        if !(helper.InArray(data.Owners, me.MID)){ data.Owners = append(data.Owners, me.MID); save() }
        if !(helper.InArray(data.Bots, mid)){ 
            data.Bots = append(data.Bots, mid)
            save() 
		    byteData, _ := json.Marshal(data)
		    broadcast("load_"+ string(byteData))
            if !(helper.InArray(me.Friends, mid)) {me.FindAndAddContactsByMid(mid)}
        } else {
		    byteData, _ := json.Marshal(data)
		    conn.Write([]byte("load_"+ string(byteData)))
        }
	}else if strings.HasPrefix(recv, "anti_"){
		mid = recv[5:]
		fmt.Printf("system running anti js - %s\n", mid)
		if _, ok := Running[mid]; ok{ Running[mid].Write([]byte("exit")); Running[mid].Close() }
		Running[mid] = conn
        if helper.InArray(data.Bots, mid){ 
            data.Bots = helper.Remove(data.Bots, mid)
        }
        if !(helper.InArray(data.Owners, me.MID)){ data.Owners = append(data.Owners, me.MID); save() }
        if !(helper.InArray(data.Anti, mid)){ 
            data.Anti = append(data.Anti, mid)
            save() 
		    byteData, _ := json.Marshal(data)
		    broadcast("load_"+ string(byteData))
            if !(helper.InArray(me.Friends, mid)) {me.FindAndAddContactsByMid(mid)}
        } else {
		    byteData, _ := json.Marshal(data)
		    conn.Write([]byte("load_"+ string(byteData)))
        }
	}else{ conn.Close(); return }
	for{
		len, err := conn.Read(bs)
		recv := string(bs[:len])
		if err != nil { break }
		fmt.Printf("recv: %s\n", recv)
		if recv == "test"{
			conn.Write([]byte("pong"))
		}else if strings.HasPrefix(recv, "ban_"){
			data.Banlist[recv[4:37]] = true
			broadcast(recv[:37])
		}else if strings.HasPrefix(recv, "join_"){
			broadcast(recv)
		}else if strings.HasPrefix(recv, "inv_"){
			joinData := strings.Split(recv, "_")
            times, ok := big.NewInt(0).SetString(joinData[2][:13], 10)
			if !ok { fmt.Println(ok); return }
			if big.NewInt(0).Sub(times, Dealed).Cmp(big.NewInt(40)) == 1 {
			    Dealed = times
			    conn.Write([]byte("kick_"+joinData[1]))
			    fmt.Println("antijs assign")
			}
		}else if strings.HasPrefix(recv, "group_"){
			stringData := strings.Split(recv, "_")
			gset := &Group{}
			json.Unmarshal([]byte(stringData[2]), &gset)
			data.Set[stringData[1]] = gset
		}
	}
	conn.Close()
	delete(Running, mid)
	return
}

func bot(op *LineThrift.Operation) {
    //fmt.Println(op)
    var listen = op.Type
    if listen == 0{ return }
    if listen == 26{
        msg := op.Message
        text := msg.Text
        sender := msg.From_
        receiver := msg.To
        txt := strings.ToLower(text)
        var to = sender
        if msg.ToType == 0{
            to = sender
        } else {
            to = receiver
        }
        if msg.ToType == 1{
            to = receiver
        }
        if msg.ToType == 2{
            to = receiver
        }
        if msg.ContentType == 0 {
            if helper.InArray(owner, sender) == true {
                if txt == "ping"{
                    start := time.Now()
                    //me.SendMessage(to, "Starting...", map[string]string{})
                    elapsed := time.Since(start)
                    me.SendMessage(to, "TimeUp " + fmt.Sprintf("%s", elapsed), map[string]string{})
                } else if txt == "/mid"{
                    me.SendMessage(to, "mid you - " + sender, map[string]string{})
                } else if txt == "/addcontbot"{
                	allbots := append(data.Bots, data.Anti...)
                	for i := 0; i < len(allbots); i++ {
                		if !(helper.InArray(me.Friends, allbots[i])) && me.MID != allbots[i]{
                			me.FindAndAddContactsByMid(allbots[i])
                		}
                	}
                	me.SendMessage(to, "Done", map[string]string{})
                } else if txt == "/save"{
                    save()
                    me.SendMessage(to, "Data saved.", map[string]string{})
                } else if txt == "/cban"{
                    for k := range data.Banlist { delete(data.Banlist, k) }
                    broadcast("cbn")
                    me.SendMessage(to, "Remove success", map[string]string{})
                } else if strings.HasPrefix(txt, "/rename:"){
                    name := text[8:]
                    profile, err := me.GetProfile()
                    if err != nil { fmt.Println(err)}
                    profile.DisplayName = name
                    me.UpdateProfile(profile)
                    me.SendMessage(to, "Update profile to "+name, map[string]string{})
                } else if strings.HasPrefix(txt, "/unban:"){
                    mid := text[7:]
                    delete(data.Banlist, mid)
                    byteData, _ := json.Marshal(data)
		            broadcast("load_"+ string(byteData))
                    me.SendMessage(to, "Unbanned user done.", map[string]string{})
                } else if strings.HasPrefix(txt, "/ban:"){
                    mid := text[5:]
		            broadcast("ban_"+mid)
                    me.SendMessage(to, "Banned user done.", map[string]string{})
                } else if txt == "/speed"{
                    start := time.Now()
                    //me.SendMessage(to, "Starting...", map[string]string{})
                    elapsed := time.Since(start)
                    me.SendMessage(to, "TimeUp " + fmt.Sprintf("%s", elapsed), map[string]string{})
                    //fmt.Println("speed : " + string(elapsed*time.Second))
                } else if txt == "/runtime"{
                    elapsed := time.Since(botStart)
                    me.SendMessage(to, fmtDuration(elapsed), map[string]string{})
                } else if txt == "/bye"{
                    me.LeaveGroup(to)
                } else if txt == "/restart"{
                    me.SendMessage(to, "Brb, going to pee.", map[string]string{})
                    restart()
                }  else if txt == "in"{
                    me.InviteIntoGroup(to, append(data.Bots, data.Anti...))
                } else if txt == "setrun"{
                    res := "Running bot:\n"
                    no := 0
                    for x := range Running{
                        no += 1
                        mc, err := me.GetContact(x); if err != nil { fmt.Println(err)}
                        res += fmt.Sprintf("\n%s. %s", strconv.Itoa(no), mc.DisplayName)
                    }
                    me.SendMessage(to, res, map[string]string{})
                } else if strings.HasPrefix(txt, "run "){
                    token := msg.Text[4:]
                    data.Login[token[:33]] = &Bot{Token:token, Type:1, Running:true}
                    err := exec.Command("bash", "-c", "go run sb.go "+token).Start()
                    if err != nil { fmt.Println(token, err) }
                    me.SendMessage(to, "Prefix starting "+token[:33], map[string]string{})
                    save()
                } else if strings.HasPrefix(txt, "anti:"){
                    token := msg.Text[5:]
                    data.Login[token[:33]] = &Bot{Token:token, Type:2, Running:true}
                    err := exec.Command("bash", "-c", "./invite "+token).Start()
                    if err != nil { fmt.Println(token, err) }
                    me.SendMessage(to, "Prefix starting "+token[:33], map[string]string{})
                } else if strings.HasPrefix(txt, "delete:"){
                    mid := txt[7:]
                    if _, ok := data.Login[mid]; ok{ delete(data.Login, mid) }
                    if _, ok := Running[mid]; ok{ Running[mid].Write([]byte("exit")); Running[mid].Close()}
                    save()
                    me.SendMessage(to, "Running deleted "+mid, map[string]string{})
                } else if txt == "help"{
                    me.SendMessage(to, "╔────────────\n╠⌬╭「CMD helper」\n╠⌬│  ping\n╠⌬│  /mid\n╠⌬│  /save\n╠⌬│  /cban\n╠⌬│  /addcontbot\n╠⌬│  unban: mid\n╠⌬│  /sp\n╠⌬│  /runtime\n╠⌬│  runall\n╠⌬│  /stopall\n╠⌬│  /bl\n╠⌬│  /setrun\n╠⌬│  /runtime\n╠⌬│  in\n╠⌬│  botadd @\n╠⌬│  botdel @\n╠⌬│  owneradd @\n╠⌬│  /reset\n╠⌬│  /getdata\n╠⌬│  mode normal\n╠⌬│  mode war\n╠⌬│  mode purge\n╠⌬╰•   /anti join\n╠⌬╭「CMD Bots」\n╠⌬│  upname: name\n╠⌬│  sp\n╠⌬│  runtime\n╠⌬│  bot\n╠⌬│  come\n╠⌬│  go\n╠⌬│  group\n╠⌬│  njoin: qr\n╠⌬│  respon\n╠⌬│  kick\n╠⌬│  gadd @\n╠⌬│  gdel @\n╠⌬│  gm\n╠⌬│  left\n╠⌬│  sets\n╠⌬│  invprotect on  off\n╠⌬│  qrprotect on  off\n╠⌬│  protect on  off\n╠⌬│  pro on  off\n╠⌬│  ban\n╠⌬│  unban\n╠⌬╰•   backup\n╠⌬╭「CMD Ghost」\n╠⌬│  anti\n╠⌬│  speed\n╠⌬│  runtime\n╠⌬│  antibot\n╠⌬│  antiset\n╠⌬╰•  #bye\n╚────────────", map[string]string{})
                } else if txt == "/stopall"{
                    for x := range Running{
                        Running[x].Write([]byte("exit"))
                        Running[x].Close()
                    }
                    me.SendMessage(to, "Stoped allbot.", map[string]string{})
                } else if txt == "runall"{
                    for x := range Running{
                        Running[x].Write([]byte("exit"))
                        Running[x].Close()
                    }
                    for x := range data.Login {
                        if data.Login[x].Type == 1 {
                            exec.Command("bash", "-c", "./sb "+data.Login[x].Token).Start()
                            time.Sleep(100*time.Millisecond)
                        }else if data.Login[x].Type == 2 {
                            exec.Command("bash", "-c", "go run invite.go "+data.Login[x].Token).Start()
                        }
                    }
                   me.SendMessage(to, "Starting allbot.", map[string]string{})
                } else if txt == "/bl"{
                    res := "Blacklist\n"
                    no := 0
                    for x := range data.Banlist{
                        mc, err := me.GetContact(x); if err != nil { continue }
                        no += 1
                        res += fmt.Sprintf("\n%s. %s", strconv.Itoa(no), mc.DisplayName)
                    }
                    me.SendMessage(to, res, map[string]string{})
                } else if strings.HasPrefix(txt, "botadd ") && !(msg.ContentMetadata["MENTION"] == ""){
                    value := helper.GetMidFromMentionees(msg.ContentMetadata["MENTION"])
                    for i := 0; i < len(value); i++ {
                        if !(helper.InArray(data.Bots, value[i])){ data.Bots = append(data.Bots, value[i]) }
                    } 
                    byteData, _ := json.Marshal(data)
		            broadcast("load_"+ string(byteData))
                    save()
                    me.SendMessage(to, "Added user bot.", map[string]string{})
                } else if strings.HasPrefix(txt, "owneradd ") && !(msg.ContentMetadata["MENTION"] == ""){
                    value := helper.GetMidFromMentionees(msg.ContentMetadata["MENTION"])
                    for i := 0; i < len(value); i++ {
                        if !(helper.InArray(data.Owners, value[i])){ data.Owners = append(data.Owners, value[i]) }
                    } 
                    byteData, _ := json.Marshal(data)
		            broadcast("load_"+ string(byteData))
                    save()
                    me.SendMessage(to, "Added user owner.", map[string]string{})
                } else if strings.HasPrefix(txt, "botdel ") && !(msg.ContentMetadata["MENTION"] == ""){
                    value := helper.GetMidFromMentionees(msg.ContentMetadata["MENTION"])
                    for i := 0; i < len(value); i++ {
                        if helper.InArray(data.Bots, value[i]){ data.Bots = helper.Remove(data.Bots, value[i]) }
                    } 
                    byteData, _ := json.Marshal(data)
		            broadcast("load_"+ string(byteData))
                    save()
                    me.SendMessage(to, "Deleted user bot.", map[string]string{})
                } else if txt == "/reset"{
                    for x := range Running{
                        Running[x].Write([]byte("exit"))
                        Running[x].Close()
                    }
                    for x := range data.Login {
                        delete(data.Login, x)
                    }
                    data.Bots = data.Bots[:0]
                    data.Anti = data.Anti[:0]
                    save()
                    me.SendMessage(to, "Reset tempbin.", map[string]string{})
                } else if txt == "/getdata"{
                    rand.Seed(time.Now().UTC().UnixNano())
                    if cone,ok := Running[data.Bots[rand.Intn(len(data.Bots))]]; ok {
                        cone.Write([]byte("updateall"))
                        me.SendMessage(to, "Reading data settings.", map[string]string{})
                    }else {
                        me.SendMessage(to, "Please reupdate again.", map[string]string{})
                    }
                } else if txt == "mode normal"{
                    broadcast("mode normal")
                    data.SystemMode = "normal"
                    me.SendMessage(to, "Type normal mode.", map[string]string{})
                } else if txt == "mode war"{
                    broadcast("mode war")
                    data.SystemMode = "war"
                    me.SendMessage(to, "Type war mode.", map[string]string{})
                } else if txt == "mode purge"{
                    broadcast("mode purge")
                    data.SystemMode = "purge"
                    me.SendMessage(to, "Type purge mode.", map[string]string{})
                } else if txt == "/anti join"{
                    G, _ := me.GetGroupWithoutMembers(to)
                    if G.PreventedJoinByTicket == true {
                        G.PreventedJoinByTicket = false
                        me.UpdateGroup(G)
                    }
                    ticket ,_ := me.ReissueGroupTicket(to)
                    broadcast("antijoin_"+to+"_"+ticket)
                }
            }
        }
    }else if listen == 13{
        param3 := strings.Split(op.Param3, "\x1e")
        if helper.InArray(owner, op.Param2) == true || helper.InArray(data.Bots, op.Param2) == true{
            if helper.InArray(param3, me.MID) {
                me.AcceptGroupInvitation(op.Param1)
            }
        }
    }
}

func main() {
	autorestart.StartWatcher()
	jsonFile, err := os.Open("data.json")
    if err != nil { fmt.Println(err)}
    byteValue, _ := ioutil.ReadAll(jsonFile)
    json.Unmarshal(byteValue, &data)
    owner = data.Owners
    me.Login("u82083cf74207fd5f91835cf06d3eab2b:aWF0OiAxNTg1OTM4MDE4MTg2Cg==..1qCiqitkQfnuNLqVxM6bu/+9V+4=")
    //allbots := append(data.Bots, data.Anti...)
    //for i := 0; i < len(allbots); i++ {
        //if !(helper.InArray(me.Friends, allbots[i])) && me.MID != allbots[i]{
            //me.FindAndAddContactsByMid(allbots[i])
        //}
    //}
    fmt.Println("Login Helper, ProcessID: ", os.Getpid())
    listener, err := net.Listen("tcp", "108.61.126.129:8000")
    if err != nil { fmt.Println(err.Error()) }
    defer listener.Close()
    go func() {
        for {
		    conn, err := listener.Accept()
		    if err != nil { fmt.Printf("Some connection error: %s\n", err) }
		    go handleConnection(conn)
	    }
	}()
    for {
        fetch, _ := me.FetchOperations(me.Revision, 5)
        if len(fetch) > 0 {
            for i := 0; i < len(fetch); i++ {
                op := fetch[i]
                bot(op)
                me.Revision = helper.MaxRevision(me.Revision, op.Revision)
                save()
            }
        }
    }
}


