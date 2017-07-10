package main

import (
	"fmt"
	"github.com/QcloudApi/qcloud_sign_golang"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/jinzhu/configor"
	"github.com/tidwall/gjson"
	"os"
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

type Stat struct {
	View     int64
	Danmuku  int64
	Reply    int64
	Favorite int64
	Coin     int64
	Share    int64
	His_rank int64
	Aid      int64
}

func main() {
	for {
		stat_receiveMessage()
	}
}

func stat_insert(users []Stat) {
	configor.Load(&Config, "./config.yml")
	var err error
	engine, err = xorm.NewEngine("mysql", Config.DB.User + ":" + Config.DB.Password + "@tcp(" + Config.DB.Host + ":" + Config.DB.Port + ")/" + Config.DB.Database)
	engine.Sync2(new(Stat))
	affected, err := engine.Insert(&users)
	fmt.Println(affected, err)
}

func stat_receiveMessage() {
	var users []Stat
	for i := 0; i < 1000; i++ {
		configor.Load(&Config, "./config.yml")

		// 替换实际的 SecretId 和 SecretKey
		secretId := Config.Qcloud.SecretId
		secretKey := Config.Qcloud.SecretKey

		// 配置
		config := map[string]interface{}{"secretId": secretId, "secretKey": secretKey, "debug": false}

		// 请求参数

		ReceiveParams := map[string]interface{}{"Region": "sh", "Action": "ReceiveMessage", "queueName": "videostat", "pollingWaitSeconds": 30}

		ReceiveData, err := QcloudApi.SendRequest("cmq-queue-sh", ReceiveParams, config)
		if err != nil {
			fmt.Print("Error.", err)
			return
		}

		fmt.Println(ReceiveData)
		message := gjson.Parse(ReceiveData).Get("message").String()
		if message != "" {
			stat_insert(users)
			os.Exit(1)
		}
		msgBody := gjson.Parse(ReceiveData).Get("msgBody").String()
		video := gjson.Parse(msgBody)
		user := Stat{View: video.Get("view").Int(), Danmuku: video.Get("danmuku").Int(), Reply: video.Get("reply").Int(), Favorite: video.Get("favorite").Int(), Coin: video.Get("coin").Int(), Share: video.Get("share").Int(), His_rank: video.Get("his_rank").Int(), Aid: video.Get("aid").Int()}
		fmt.Println(len(users))
		users = append(users, user)

		receiptHandle := gjson.Parse(ReceiveData).Get("receiptHandle").String()
		stat_deleteMessage(receiptHandle)
	}
	//stat_insert(users)
	users = make([]Stat, 0)
}

func stat_deleteMessage(receiptHandle string) {
	configor.Load(&Config, "./config.yml")

	// 替换实际的 SecretId 和 SecretKey
	secretId := Config.Qcloud.SecretId
	secretKey := Config.Qcloud.SecretKey

	// 配置
	config := map[string]interface{}{"secretId": secretId, "secretKey": secretKey, "debug": false}

	DeleteParams := map[string]interface{}{"Region": "sh", "Action": "DeleteMessage", "queueName": "videostat", "receiptHandle": receiptHandle}

	DeleteData, err := QcloudApi.SendRequest("cmq-queue-sh", DeleteParams, config)
	if err != nil {
		fmt.Print("Error.", err)
		return
	}
	fmt.Println(DeleteData)
}
