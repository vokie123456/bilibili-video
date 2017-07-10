package main

import (
	"fmt"
	"github.com/QcloudApi/qcloud_sign_golang"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/jinzhu/configor"
	"github.com/tidwall/gjson"
	"os"
	"strconv"
	"strings"
)

var Config = struct {
	DB     struct {
		       Host     string
		       User     string `default:"root"`
		       Password string
		       Port     string `default:"3306"`
		       Database string
	       }
	Qcloud struct {
		       SecretId  string
		       SecretKey string
	       }
}{}

var engine *xorm.Engine

type Video struct {
	Typeid       int64
	Author       string
	Created      int64
	Length       int64
	Video_review int64
	Aid          int64
}

var users []Video

func main() {
	for {
		receiveMessage()
	}
}

func insert(users []Video) {
	loop:

	configor.Load(&Config, "./config.yml")
	var err error
	engine, err = xorm.NewEngine("mysql", Config.DB.User + ":" + Config.DB.Password + "@tcp(" + Config.DB.Host + ":" + Config.DB.Port + ")/" + Config.DB.Database)
	engine.Sync2(new(Video))
	affected, err := engine.Insert(&users)
	if err != nil {
		goto loop
	}
	fmt.Println(affected, err)
}

func receiveMessage() {
	for i := 0; i < 1000; i++ {
		configor.Load(&Config, "./config.yml")

		// 替换实际的 SecretId 和 SecretKey
		secretId := Config.Qcloud.SecretId
		secretKey := Config.Qcloud.SecretKey

		// 配置
		config := map[string]interface{}{"secretId": secretId, "secretKey": secretKey, "debug": false}

		// 请求参数

		ReceiveParams := map[string]interface{}{"Region": "sh", "Action": "ReceiveMessage", "queueName": "videoupload", "pollingWaitSeconds": 30}

		ReceiveData, err := QcloudApi.SendRequest("cmq-queue-sh", ReceiveParams, config)
		if err != nil {
			fmt.Print("Error.", err)
			return
		}

		fmt.Println(ReceiveData)
		message := gjson.Parse(ReceiveData).Get("message").String()
		if message != "" {
			insert(users)
			os.Exit(1)
		}
		msgBody := gjson.Parse(ReceiveData).Get("msgBody").String()
		video := gjson.Parse(msgBody)
		video.ForEach(func(key, value gjson.Result) bool {
			minute, _ := strconv.Atoi(strings.Split(value.Get("length").String(), ":")[0])
			second, _ := strconv.Atoi(strings.Split(value.Get("length").String(), ":")[1])
			length := int64(minute * 60) + int64(second)
			user := Video{Typeid: value.Get("typeid").Int(), Author: value.Get("author").String(), Created: value.Get("created").Int(), Length: length, Video_review: value.Get("video_review").Int(), Aid: value.Get("aid").Int()}
			users = append(users, user)
			return true // keep iterating
		})
		receiptHandle := gjson.Parse(ReceiveData).Get("receiptHandle").String()
		deleteMessage(receiptHandle)
	}
	insert(users)
	users = make([]Video, 0)

}

func deleteMessage(receiptHandle string) {
	configor.Load(&Config, "./config.yml")

	// 替换实际的 SecretId 和 SecretKey
	secretId := Config.Qcloud.SecretId
	secretKey := Config.Qcloud.SecretKey

	// 配置
	config := map[string]interface{}{"secretId": secretId, "secretKey": secretKey, "debug": false}

	DeleteParams := map[string]interface{}{"Region": "sh", "Action": "DeleteMessage", "queueName": "videoupload", "receiptHandle": receiptHandle}

	DeleteData, err := QcloudApi.SendRequest("cmq-queue-sh", DeleteParams, config)
	if err != nil {
		fmt.Print("Error.", err)
		return
	}
	fmt.Println(DeleteData)
}
