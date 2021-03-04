package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func main() {

	group := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		group.Add(1)
		go func() {
			time.Sleep(time.Second / 1000)
			defer group.Done()
			params := struct {
				Op     string   `json:"op"`
				Params []string `json:"params"`
			}{}
			params.Op = "cp"
			params.Params = []string{randomString(9), "QmayKQWJgWmr46DqWseADyUanmqxh662hPNdAkdHRjiQQH"}
			bs, _ := json.Marshal(params)
			res, err := http.Post("http://127.0.0.1:10087/fs", "application/json", bytes.NewReader(bs))
			if err == nil {
				bs, _ = ioutil.ReadAll(res.Body)
				fmt.Println(string(bs))
			}
		}()
	}
	group.Wait()
}

const (
	chars = "abcdefghijklmnopqrstuvwxyz0123456789"
)

// RandomStrings returns a slice of randomly generated strings.
func RandomStrings(maxlen uint, n int) []string {
	ss := make([]string, 0)
	for i := 0; i < n; i++ {
		ss = append(ss, randomString(maxlen))
	}
	return ss
}

func randomString(l uint) string {
	s := make([]byte, l)
	for i := 0; i < int(l); i++ {
		s[i] = chars[rand.Intn(len(chars))]
	}
	return string(s)
}
