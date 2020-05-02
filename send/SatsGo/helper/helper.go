package helper

import (
	"../LineThrift"
	"encoding/json"
	"fmt"
	"time"
	"strconv"
)

type mentionMsg struct {
	MENTIONEES []struct {
		S string `json:"S"`
		E string `json:"E"`
		M string `json:"M"`
	} `json:"MENTIONEES"`
}

func GetMentionMsg() mentionMsg {
	return mentionMsg{}
}

func MaxRevision (a, b int64) int64 {
    if a > b {
	return a
    }
    return b
}

func InArray(arr []string, str string) bool {
   for _, a := range arr {
      if a == str {
         return true
      }
   }
   return false
}

//
func InArray_int64(arr []int64, data int64) bool {
   for _, a := range arr {
      if a == data {
         return true
      }
   }
   return false
}

func InMap(dict map[string]bool, key string) bool {
    _, ok := dict[key]
    return ok
}

func zip(a []interface{}, b []string) ([]interface{}, error) {
    if len(a) != len(b) {
        return nil, fmt.Errorf("zip: arguments must be of same length")
    }
    r := make([]interface{}, len(a))
    for i, e := range a {
        r[i] = []interface{}{e, b[i]}
    }
    return r, nil
}

func IndexOf(element string, data []string) int {
   for k, v := range data {
       if element == v {
           return k
       }
   }
   return -1
}

func Remove(s []string, r string) []string {
    new := make([]string, len(s))
    copy(new, s)
    for i, v := range new {
        if v == r {
            return append(new[:i], new[i+1:]...)
        }
    }
    return s
}

func SplitList(s []string, a int) [][]string {
	t, r := len(s)/a, len(s)%a
	res := [][]string{}
	for x := 0; x < t; x++ {
		res = append(res, s[a*x:a*(x+1)])
	}
	if r != 0 {
		res = append(res, s[a*t:])
	}
	return res
}

func MatchList(a []string, b []string) []string {
	res := []string{}
	for _, v := range a {
		if InArray(b, v) && !InArray(res, v) {
			res = append(res, v)
		}
	}
	return res
}

func Shift(s []string, a int) []string {
	if a >= len(s) || a < 0 {
		fmt.Println("Invalid number of length.")
		return s
	}
	x := s[:a]
	y := s[a:]
	return append(y, x...)
}

func GetMidFromMentionees(data string) []string{
	var midmen []string
	var midbefore []string
	res := mentionMsg{}
	json.Unmarshal([]byte(data), &res)
	for _, v := range res.MENTIONEES {
		if InArray(midbefore, v.M) == false {
			midbefore = append(midbefore, v.M)
			midmen = append(midmen, v.M)
		} 
	}
	return midmen
}

func Log(optype LineThrift.OpType, logtype string, str string) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	a:=time.Now().In(loc)
	yyyy := strconv.Itoa(a.Year())
	MM := a.Month().String()
	dd := strconv.Itoa(a.Day())
	hh := a.Hour()
	mm := a.Minute()
	ss := a.Second()
	var hhconv string
	var mmconv string
	var ssconv string
	if hh < 10 {
		hhconv = "0"+strconv.Itoa(hh)
	} else {
		hhconv = strconv.Itoa(hh)
	}
	if mm < 10 {
		mmconv = "0"+strconv.Itoa(mm)
	} else {
		mmconv = strconv.Itoa(mm)
	}
	if ss < 10 {
		ssconv = "0"+strconv.Itoa(ss)
	} else {
		ssconv = strconv.Itoa(ss)
	}
	times := yyyy+"-"+MM+"-"+dd+" "+hhconv+":"+mmconv+":"+ssconv
	fmt.Println("["+times+"]["+optype.String()+"]["+logtype+"]"+str)
}
