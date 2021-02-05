package consensus

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func RunHttpServer() {
	u, err := url.Parse("http://127.0.0.1:5001")
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ModifyResponse = func(response *http.Response) error {
		if response.StatusCode == 200 {
			if response.Request.URL.Path == "/api/v0/add" {

				b, err := ioutil.ReadAll(response.Body) //Read html
				if err != nil {
					return err
				}
				err = response.Body.Close()
				if err != nil {
					return err
				}

				body := ioutil.NopCloser(bytes.NewReader(b))
				response.Body = body
				fmt.Println(string(b))
				return nil
			}
			fmt.Println(response.Request.URL.Path)
		}
		return nil
	}
	err = http.ListenAndServe("127.0.0.1:25001", proxy)
	panic(err)
}
