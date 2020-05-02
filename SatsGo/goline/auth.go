package goline

import (
	"../LineThrift"
	"net/http"
	"fmt"
	"io/ioutil"
	//"time"
	"encoding/json"
	"context"
	"crypto/rsa"
    "crypto/sha256"
    "crypto/rand"
//    "encoding/base64"
	"strconv"
	"os"
	"math/big"
)

type InsideResults struct {
	Verifier string `json:"verifier"`
	AuthPhase string `json:"authPhase"`
}
type Results struct {
	Result InsideResults `json:"result"`
	Timestamp string `json:"timestamp"`
}

func getResult(body []byte) (*Results, error) {
	var s = new(Results)
	err := json.Unmarshal(body, &s)
	return s, err
}

func Encrypt(plaintext string, publickey *rsa.PublicKey) (string, error) {
    ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publickey, []byte(plaintext), []byte(""))
    hex := fmt.Sprintf("%x", ciphertext)
    return hex, err
}

func (s *LINE) LoadService(){
    s.Client = s.TalkService()
	s.Poll = s.PollService()
	if s.IsLogin != true {
		panic("[Error]Not yet logged in.")
	}
	//a:=time.Now()
	//fmt.Println(time.Since(a))
	getprof, err := s.Client.GetProfile(context.TODO())
	if err != nil {
		panic(err)
	}
	s.MID = getprof.Mid
	fmt.Println("Loginbot:")
	fmt.Println("DisplayName: ", getprof.DisplayName)
	fmt.Println("MID: ", getprof.Mid)
	fmt.Println("AuthToken: ", s.AuthToken)
	//s.SetTimelineHeaders()
	//s.AcquireEncryptedAccessToken()
	rev, err := s.PollService().GetLastOpRevision(context.TODO())
	if err != nil { panic(err) }
	s.Revision = rev
	s.Friends, _ = s.TalkService().GetAllContactIds(context.TODO())
	fmt.Println("Made in by sats.")
}

func loginRequestQR(identity LineThrift.IdentityProvider, verifier string) *LineThrift.LoginRequest{
	lreq := &LineThrift.LoginRequest{
		Type: 1,
		KeepLoggedIn: true,
		IdentityProvider: identity,
		AccessLocation: IP_ADDR,
		SystemName: SYSTEM_NAME,
		Verifier: verifier,
		E2eeVersion: 0,
	}
	return lreq
}

// not yet working, still learning...
func (s *LINE) LoginWithCredential(email string, password string) {
    fmt.Println(email, password)
    tauth := s.AuthService()
    rsaKey, _ := tauth.GetRSAKeyInfo(context.TODO(), 1)
    fmt.Println(rsaKey.SessionKey)
    rsaSessionKey := rsaKey.SessionKey
    message := (string(len(rsaSessionKey)) + rsaSessionKey + string(len(email)) + email + string(len(password)) + password)
    Nvalue, _ := new(big.Int).SetString(rsaKey.Nvalue, 16)
    Evalue, _:= strconv.ParseInt(rsaKey.Evalue,16,64)
    E , _ := strconv.Atoi(strconv.FormatInt(Evalue, 10))
    pubKey := &rsa.PublicKey{
	N: Nvalue,
	E: E,
    }
    crypto, _ := Encrypt(message, pubKey)
    fmt.Println(crypto)
    _, err := os.OpenFile(email + ".crt", os.O_RDONLY|os.O_CREATE, 0666)
    if os.IsNotExist(err) {
        ioutil.WriteFile(email + ".crt", []byte("Hello"), 0755)
    }
    f, err := ioutil.ReadFile(email + ".crt")
    certificate := string(f)
    auth := s.LoginZService()
    login := &LineThrift.LoginRequest{ Type: 0, IdentityProvider: 1, Identifier: rsaKey.Keynm, Password: crypto, KeepLoggedIn: true, AccessLocation: IP_ADDR, SystemName: SYSTEM_NAME, Certificate: certificate, E2eeVersion: 0}
    fmt.Println(login)
    result, err := auth.LoginZ(context.TODO(), login)
    if err != nil { panic(err) }
    if result.Type == 3{
        fmt.Println(result.PinCode)
        s.AuthToken = result.Verifier
        client := &http.Client{}
        req, _ := http.NewRequest("GET", LINE_HOST_DOMAIN+LINE_CERTIFICATE_PATH, nil)
        req.Header.Set("User-Agent",USER_AGENT)
	req.Header.Set("X-Line-Application",s.AppName)
	req.Header.Set("X-Line-Carrier",CARRIER)
	req.Header.Set("X-Line-Access",result.Verifier)
	res, _ := client.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	x, _ := getResult([]byte(body))
	result, err := auth.LoginZ(context.TODO(), &LineThrift.LoginRequest{Type: 1, KeepLoggedIn: true, Verifier: x.Result.Verifier, E2eeVersion: 0})
	if err != nil {
	    panic(err)
	}
	f, _ := os.Open(email + ".crt")
	f.WriteString(result.Certificate)
        s.IsLogin = true
	s.LoginWithAuthToken(result.AuthToken)
    } else if result.Type == 1{
        s.IsLogin = true
	s.LoginWithAuthToken(result.AuthToken)
    }
// 	fmt.Println(message)

}


func (s *LINE) LoginWithQrCode(keepLoggedIn bool){
	tauth := s.AuthService()
	qrCode, err := tauth.GetAuthQrcode(context.TODO(), keepLoggedIn, SYSTEM_NAME, true)
	if err != nil{
		panic(err)
	}

	// by jay
	fo, err := os.Create("url_login.txt")
	if err == nil {
	    ss := qrCode.Verifier
	    buf := make([]byte, 1024)
	    buf = []byte(ss)
	    _, err := fo.Write(buf[0:len(ss)])
	    if err == nil {
		fo.Close()
	    }
	}

	fmt.Println("line://au/q/"+qrCode.Verifier)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", LINE_HOST_DOMAIN+LINE_CERTIFICATE_PATH, nil)
	req.Header.Set("User-Agent",USER_AGENT)
	req.Header.Set("X-Line-Application",s.AppName)
	req.Header.Set("X-Line-Carrier",CARRIER)
	req.Header.Set("X-Line-Access",qrCode.Verifier)
	s.AuthToken = qrCode.Verifier
	res, _ := client.Do(req)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	x, _ := getResult([]byte(body))
	iR := x.Result
	_verifier := iR.Verifier 
	loginZ := s.LoginZService()
	loginReq := loginRequestQR(1, _verifier)
	resultz, err := loginZ.LoginZ(context.TODO(),loginReq)
	if err != nil {
		panic(err)
	}
	s.IsLogin = true
	s.LoginWithAuthToken(resultz.AuthToken)
}

func (s *LINE) LoginWithAuthToken(authToken string){ 
	s.AuthToken = authToken
	s.IsLogin = true
	headers := make(http.Header)
	headers.Add("X-Line-Application", APP_TYPE)
	headers.Add("X-Line-Access", s.AuthToken)
	headers.Add("X-Line-Carrier", CARRIER)
	headers.Add("User-Agent", USER_AGENT)
	s.Headers = headers
	s.LoadService()
}
