package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
	"strings"
	"encoding/json"
	"bytes"
	"os"
	"bufio"
	"io"
)

type MessageSt struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

type recevicer struct {
	name     string
	location string
}

var httpclient = http.Client{
	CheckRedirect: nil,
	Timeout:       60 * time.Second,
}

func getWeather(url string) string {
	fmt.Println("start")
	tip := ""
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.111 Safari/537.36")
	resp, err := httpclient.Do(request)
	defer resp.Body.Close()
	if err == nil {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			re := regexp.MustCompile(`<metaname="description"content="(.*)">`)
			str := strings.Replace(string(body), " ", "", -1)
			find := re.FindStringSubmatch(str)
			if len(find) > 1 {
				return strings.Replace(find[1], "墨迹天气建议您", "建议", -1)
			}
		}
	}
	return tip
}

func sendMessage(rer recevicer) {
	url := "http://localhost:8888/send"
	message := MessageSt{rer.name, getWeather("https://tianqi.moji.com/weather/china/" + rer.location)}
	b, err := json.Marshal(message)
	resp, err := http.Post(url, "application/json;charset=utf-8", bytes.NewBuffer([]byte(b)))
	if err == nil {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			fmt.Printf("receive message:%s\n", body)
		}
	} else {
		return
	}
}

func sendMessages(rers []recevicer) {
	for _, v := range rers {
		sendMessage(v)
	}
}

func main() {
	var file = "./receivers.conf"
	if len(os.Args) == 2 {
		file = os.Args[1]
	}
	f, err := os.Open(file)
	if err == nil && f != nil {
		var receivers []recevicer
		defer f.Close()
		rd := bufio.NewReader(f)
		for {
			line, err := rd.ReadString('\n')
			if err != nil || io.EOF == err {
				break
			}
			if len(line) > 0 {
				if line[:1] != "#" {
					s := strings.Split(line, `|-|`)
					if len(s) == 2 {
						receivers = append(receivers, recevicer{s[0], s[1]})
					}
				}
			}
		}
		sendMessages(receivers)
	} else {
		fmt.Printf("未找到文件:%s\n", file)
	}
}
