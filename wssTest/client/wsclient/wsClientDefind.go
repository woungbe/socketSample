package wsclient

import (
	"fmt"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 81920
)

/*
config에서 바로 가져다 쓰면 좋은점 ..
동기화 하기 편함..
이것저것 신경 안써도 됨 .
복잡하게 추가 삭제가 쉬워짐

config에서 바로 가져다 쓰면 단점
모듈별 분리하기가 쉽지 않음
재사용성이 떨어짐
라이브러리화 시키는에 에러사항이 있음
하던 사람만 사용가능 -
*/

type WSRequest struct {
	Command string
}

type ConfigData struct {
	IsDebugmode bool
	URL         string
	Port        string
	Link        string
	Ssluse      bool
	Sslcrt      string
	SslKey      string
}

// 기본값 설정하기
func (ty *ConfigData) Default() {
	ty.IsDebugmode = true
	ty.URL = "localhost"
	ty.Port = "9183"
	ty.Link = "status"
	ty.Ssluse = false
	ty.Sslcrt = ""
	ty.SslKey = ""
}

func (ty *ConfigData) SetDebugMode(aa bool) *ConfigData {
	ty.IsDebugmode = aa
	return ty
}

func (ty *ConfigData) SetURL(url string) *ConfigData {
	ty.URL = url
	return ty
}

func (ty *ConfigData) SetPort(port string) *ConfigData {
	ty.Port = port
	return ty
}

func (ty *ConfigData) SetLink(link string) *ConfigData {
	ty.Link = link
	return ty
}

func (ty *ConfigData) SetSSL(ssluse bool, sslcrt, sslkey string) *ConfigData {
	ty.Ssluse = ssluse
	ty.Sslcrt = sslcrt
	ty.SslKey = sslkey
	return ty
}

func (ty *ConfigData) GetPath() string {
	if ty.Ssluse {
		return fmt.Sprintf("wss://%s:%s/%s", ty.URL, ty.Port, ty.Link)
	} else {
		return fmt.Sprintf("ws://%s:%s/%s", ty.URL, ty.Port, ty.Link)
	}
}
