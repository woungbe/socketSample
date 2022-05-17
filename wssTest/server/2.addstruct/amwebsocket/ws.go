package amwebsocket

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/phf/go-queue/queue"
)

var AMWSocketPTR *AMWSocket

var IsDebugmode bool
var HandlePath string
var Ssluse string
var Sslcrt string
var Sslkey string
var Port string

type AMWSocket struct {
	MaxClientCount        int64 //최대 사용자 클라이언트 수
	nTotalConnectionCount int64 //현재 접속된 전체 카운트

	ClientSessionsQ   *queue.Queue
	ClientSessionsMap map[int64]*AMWSession  //클라이언트 맵
	ClientLoginMap    map[string]*AMWSession //로그인된 사용자 클라이언트 리스트

	ChanEnter      chan *websocket.Conn
	ChanLeave      chan *AMWSession
	ChanAddSession chan int64 //섹션 추가
}

func GetAMWSvr() *AMWSocket {
	if AMWSocketPTR == nil {
		AMWSocketPTR = new(AMWSocket)
	}
	return AMWSocketPTR
}

func (ty *AMWSocket) Init() error {
	ty.MaxClientCount = 10

	ty.ClientSessionsMap = make(map[int64]*AMWSession) //클라이언트 맵
	ty.ClientLoginMap = make(map[string]*AMWSession)   //로그인된 사용자 클라이언트 리스트

	ty.ChanEnter = make(chan *websocket.Conn)
	ty.ChanLeave = make(chan *AMWSession)
	ty.ChanAddSession = make(chan int64)

	ty.ClientSessionsQ = queue.New()

	for i := int64(0); i < ty.MaxClientCount; i++ {
		gSession := NewAdminSeesionLog()
		gSession.Index = i
		ty.ClientSessionsQ.PushBack(gSession)
	}
	log.Println("Start AMWSocket", "Create Sessions ", strconv.FormatInt(ty.MaxClientCount, 10))

	// 있으면 여기서 눌러서 적용시켜주세요
	// IsDebugmode = IsDebugmode
	IsDebugmode = true
	HandlePath = "CheckStatusServer"
	Ssluse = "N" // Ssluse = adminconfig.GetConfigData().Ssluse
	Sslcrt = ""
	Sslkey = ""
	Port = "9183"

	go ty.sessionCtrlRUN()

	return nil
}

func (ty *AMWSocket) SetDebugMode(debugmode bool) {
	IsDebugmode = debugmode
}

func (ty *AMWSocket) SetHandlePath(link string) {
	HandlePath = link
}

func (ty *AMWSocket) SetSSL(use, crt, key string) {
	Ssluse = use
	Sslcrt = crt
	Sslkey = key
}

func (ty *AMWSocket) SetPort(port string) {
	Port = port
}

//StartWebSocketServer 웹소켓 서버 시작
func (ty *AMWSocket) StartWebSocketServer() {
	http.HandleFunc("/"+HandlePath, func(w http.ResponseWriter, r *http.Request) {
		ty.sessionWs(w, r)
	})

	log.Println("Admin WebSocket Server Starting....")

	svrport := ":" + Port
	var addr = flag.String("Admin addr", svrport, "Admin http service address")
	if Ssluse == "Y" {
		err := http.ListenAndServeTLS(*addr, Sslcrt, Sslkey, nil)
		if err != nil {
			log.Println("Error", "ListenAndServe: ", err)
			os.Exit(1)
			return
		}
	} else {
		err := http.ListenAndServe(*addr, nil)
		if err != nil {
			log.Println("Error", "ListenAndServe: ", err)
			os.Exit(1)
			return
		}
	}
}

// serveWs handles websocket requests from the peer.
func (ty *AMWSocket) sessionWs(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic  sessionWs", "Error ", err)
			}
		}
	}()

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if c != nil {
			c.Close()
		}
		log.Println("ERROR", "upgrade", err.Error())
		return
	}

	//-- 현재 연결가능한 섹션이 존재하는가부터 체크
	if ty.ClientSessionsQ.Len() <= 0 {
		//-- 더이상 접속 못함
		log.Println("Crit Error", "Admin Client Session", "Not empty Session!!!!!!!!!!!!!!")
		c.Close()
		return
	}

	//블럭 IP 처리
	/*
		strIP := r.RemoteAddr
		if strings.Contains(strIP, ":") == true {
			sar := strings.Split(strIP, ":")
			strIP = sar[0]
		}
		if strIP != "127.0.0.1" {
			b, _ := xFutureDBCommands.CheckSVRIP(strIP)
			if b == false {
				xFlog.Warn("Warn", "미등록 서버 IP", strIP)
				c.Close()
				return
			}
		}
	*/

	ty.ChanEnter <- c
}

func (ty *AMWSocket) addNewSession(nCount int64) {
	defer func() {
		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
			}
		}
	}()
	var i int64
	for i = 0; i < nCount; i++ {
		newObj := NewAdminSeesionLog()
		newObj.Index = i + ty.MaxClientCount

		ty.ClientSessionsQ.PushBack(newObj)
	}
	ty.MaxClientCount = ty.MaxClientCount + nCount

	tot := ty.MaxClientCount
	cur := ty.nTotalConnectionCount
	totq := tot - cur

	strP := fmt.Sprintf("\n -------------- Admin Session Info ----------------\n ")
	strP2 := fmt.Sprintf("Max Clint= %d \n Current User Session Count = %d \n %d Free Sessions", tot, cur, totq)
	strP3 := fmt.Sprintf("%s%s\n-------------------------------------------------", strP, strP2)
	log.Println(strP3)

}

//sessionCtrlRUN 섹션 런
func (ty *AMWSocket) sessionCtrlRUN() {
	//ticker := time.NewTicker(1 * time.Second)
	defer func() {
		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
			}
		}
	}()
	for {
		select {
		case client := <-ty.ChanEnter:
			ty.addSession(client)
		case client := <-ty.ChanLeave:
			ty.closeClientSession(client)
		case nCount := <-ty.ChanAddSession:
			ty.addNewSession(nCount)
		}
	}
}

//CloseClientSession 클라언트 연결종료시
func (ty *AMWSocket) closeClientSession(se *AMWSession) {
	defer func() {
		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
			}
		}
	}()

	ty.nTotalConnectionCount--
	if ty.nTotalConnectionCount <= 0 {
		ty.nTotalConnectionCount = 0
	}

	delete(ty.ClientSessionsMap, se.Index)
	se.Conn.Close()
	se.Conn = nil
	ty.ClientSessionsQ.PushBack(se)

	if IsDebugmode == true {
		log.Println("Admin Web Socket UnConnect Clint ", "Index  ", se.Index)
		log.Println("Admin Web Socket count ", "total", ty.nTotalConnectionCount)
	}
}

//AddSession 클라이언트 추가
func (ty *AMWSocket) addSession(Conn *websocket.Conn) {
	var client *AMWSession

	defer func() {
		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "WebSocket Add Session fail ", err)
			}
		}
	}()

	ty.nTotalConnectionCount++

	if ty.ClientSessionsQ.Len() <= 0 {
		//-- 더이상 접속 못함
		log.Println("Crit Error", "Admin Client Session", "Not empty Session!!!!!!!!!!!!!!")
		Conn.Close()
		return
	}

	tempObj := ty.ClientSessionsQ.PopFront()
	if tempObj == nil {
		Conn.Close()
		log.Println("Crit Error", "Admin Client Session", "Not empty Session!!!!!!!!!!!!!!")
	} else {
		client = tempObj.(*AMWSession)
		client.Clear()
		client.Conn = Conn
		ty.ClientSessionsMap[client.Index] = client
		if IsDebugmode == true {
			log.Println("Admin Web Socket Connect Clint ", "Index  ", client.Index)
		}
	}

	if client != nil {
		go client.ReadMessage()
		go client.WriteMessage()

	}
}

//GetSessionCnts 현재 접속자 정보
func (ty *AMWSocket) GetSessionCnts() (maxCnt int64, currCnt int64) {
	return ty.MaxClientCount, ty.nTotalConnectionCount
}

//Broadcast 브로캐스트
func (ty *AMWSocket) Broadcast(msg []byte, bLoing bool) {
	defer func() {
		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
			}
		}
	}()

	for _, value := range ty.ClientSessionsMap {
		if value != nil && value.Conn != nil {
			if bLoing {
				if value.BLogin {
					value.SendMessageByte(msg)
				}
			} else {
				value.SendMessageByte(msg)
			}
		}
	}
}

//SendUserSession 특정 유저에게 메시지 전송
func (ty *AMWSocket) SendUserSession(strUserID string, msg []byte) bool {

	defer func() {
		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
			}
		}
	}()
	obj := ty.ClientLoginMap[strUserID]
	if obj != nil {
		obj.SendMessageByte(msg)
		return true
	}
	return false
}

//AddLoginUser 로그인 사용자 추가
func (ty *AMWSocket) addLoginUser(strUserID string, se *AMWSession) {
	defer func() {
		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic", "Error ", err)
			}
		}
	}()
	ty.ClientLoginMap[strUserID] = se
}

//RemoveLoginUser 로그인 사용자 삭제
func (ty *AMWSocket) removeLoginUser(strUserID string, se *AMWSession) {
	defer func() {
		if IsDebugmode == false {
			if err := recover(); err != nil {
				log.Println("Crit Error !!!!! Panic ", "Error ", err)
			}
		}
	}()
	delete(ty.ClientLoginMap, strUserID)
}
