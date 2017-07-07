package main

import (
	"github.com/valyala/fasthttp"
	"github.com/tidwall/gjson"
	"fmt"
	"github.com/jinzhu/configor"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"strings"
	"strconv"
	"time"
)

var Config = struct {
	DB struct {
		   Host     string
		   User     string `default:"root"`
		   Password string
		   Port     string   `default:"3306"`
		   Database string
	   }
}{}

var engine *xorm.Engine

type Video struct {
	Typeid       int64
	Comment      int64
	Play         int64
	Author       string
	Created      int64
	Length       int64
	Video_review int64
	Favorites    int64
	Aid          int64
}

func writedb(url string) {
	var users []Video
	loop:
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.SetRequestURI(url)
	fasthttp.Do(req, resp)
	player := gjson.GetBytes(resp.Body(), "data.vlist")
	if len(player.Array()) != 1 {
		player.ForEach(func(key, value gjson.Result) bool {
			minute, _ := strconv.Atoi(strings.Split(value.Get("length").String(), ":")[0])
			second, _ := strconv.Atoi(strings.Split(value.Get("length").String(), ":")[1])
			length := int64(minute * 60) + int64(second)
			user := Video{Typeid:value.Get("typeid").Int(), Comment: value.Get("comment").Int(), Play:value.Get("play").Int(), Author:value.Get("author").String(), Created:value.Get("created").Int(), Length:length, Video_review:value.Get("video_review").Int(), Favorites:value.Get("favorites").Int(), Aid:value.Get("aid").Int()}
			users = append(users, user)
			return true // keep iterating
		})
		configor.Load(&Config, "./config.yml")
		var err error
		engine, err = xorm.NewEngine("mysql", Config.DB.User + ":" + Config.DB.Password + "@tcp(" + Config.DB.Host + ":" + Config.DB.Port + ")/" + Config.DB.Database)
		engine.Sync2(new(Video))
		affected, err := engine.Insert(&users)
		if err != nil {
			goto loop
		}
		fmt.Println(affected, err)
	} else {
		fmt.Println("empty")
	}
}

func video(url string) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.SetRequestURI(url + "1")
	fasthttp.Do(req, resp)
	pages := gjson.ParseBytes(resp.Body()).Get("data").Get("pages").Int()
	for i := 1; i < int(pages) + 1; i++ {
		href := url + strconv.Itoa(i)
		writedb(href)
		time.Sleep(2 * time.Second)
	}
}

func main() {
	for i:=0;i<10;i++{
	fmt.Printf("number:%d",i)
	video("http://space.bilibili.com/ajax/member/getSubmitVideos?mid="+strconv.Itoa(i)+"&pagesize=200&tid=0&page=")
	}
}