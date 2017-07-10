package main

import (
	"fmt"
	"github.com/QcloudApi/qcloud_sign_golang"
	"github.com/jinzhu/configor"
	"github.com/tidwall/gjson"
	"github.com/valyala/fasthttp"
	"strconv"
	"sync"
)

var Config = struct {
	Qcloud struct {
		       SecretId  string
		       SecretKey string
	       }
}{}

func stat(url string) {
	configor.Load(&Config, "./config.yml")

	// 替换实际的 SecretId 和 SecretKey
	secretId := Config.Qcloud.SecretId
	secretKey := Config.Qcloud.SecretKey

	// 配置
	config := map[string]interface{}{"secretId": secretId, "secretKey": secretKey, "debug": false}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.SetRequestURI(url + "1")
	fasthttp.Do(req, resp)
	json := gjson.ParseBytes(resp.Body())
	if json.Get("code").Int() == 0 {
		SendParams := map[string]interface{}{"Region": "sh", "Action": "SendMessage", "queueName": "Videostat", "msgBody": json.Get("data").String()}
		SendData, err := QcloudApi.SendRequest("cmq-queue-sh", SendParams, config)
		if err != nil {
			fmt.Print("Error.", err)
			return
		}
		fmt.Println(SendData)
	}

}

func main() {
	taskChan := make(chan int)
	TCount := 18
	var wg sync.WaitGroup //创建一个sync.WaitGroup
	go func() {
		for i := 0; i < 20000000; i++ {
			taskChan <- i
		}
		close(taskChan)
	}()
	wg.Add(TCount)
	for i := 0; i < TCount; i++ {
		i := i
		go func() {
			defer func() {
				wg.Done()
			}()
			for task := range taskChan {
				func() {
					defer func() {
						err := recover()
						if err != nil {
							fmt.Printf("任务失败：工作者i=%v, task=%v, err=%v\r\n", i, task, err)
						}
					}()
					url := "http://api.bilibili.com/archive_stat/stat?&aid=" + strconv.Itoa(i)
					stat(url)
				}()
			}
		}()
	}
	wg.Wait()

}
