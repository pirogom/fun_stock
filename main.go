package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var adminSrvPort int = 30301

type oneBarData struct {
	Date  string `json:"Date"`
	Open  int    `json:"Open"`
	High  int    `json:"High"`
	Low   int    `json:"Low"`
	Close int    `json:"Close"`
}

/**
*	execCommandStart
**/
func execCommandStart(dir string, name string, arg ...string) error {
	cmd := exec.Command(name, arg...)

	//
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Start()
	return err
}

/**
*	openWeb
**/
func openWeb(url string) {
	r := strings.NewReplacer("&", "^&")
	stripURL := r.Replace(url)

	execCommandStart("", "rundll32.exe", "url.dll,FileProtocolHandler", stripURL)
}

/**
* 관리자 페이지 열기
**/
func openWebPage(hasDelay bool) {
	if hasDelay == true {
		time.Sleep(1000 * time.Millisecond)
	}
	tmpOpenAddr := fmt.Sprintf("http://127.0.0.1:%d/", adminSrvPort)
	openWeb(tmpOpenAddr)
}

/**
*	atoi
**/
func atoi(str string) int {
	num, err := strconv.Atoi(str)

	if err != nil {
		return 0
	}

	return num
}

/**
*	dataProc
**/
func dataProc(w http.ResponseWriter, r *http.Request, path string) {

	fo, err := os.Open(".\\data\\" + path)

	if err != nil {
		http.Error(w, "file not found", 406)
		return
	}

	defer fo.Close()

	reader := bufio.NewReader(fo)

	barArr := []oneBarData{}

	for {
		line, isPrefix, err := reader.ReadLine()

		if isPrefix || err != nil {
			break
		}

		var Date string
		var OpenPrice string
		var HighPrice string
		var LowPrice string
		var ClosePrice string

		rl := string(line)

		scanCnt, scanErr := fmt.Sscanf(rl, "%s %s %s %s %s", &Date, &OpenPrice, &HighPrice, &LowPrice, &ClosePrice)

		if scanErr != nil {
			panic(scanErr)
		}

		if scanCnt == 5 {

			oneBar := oneBarData{}

			oneBar.Date = Date
			oneBar.Open = atoi(strings.Replace(OpenPrice, ",", "", -1))
			oneBar.High = atoi(strings.Replace(HighPrice, ",", "", -1))
			oneBar.Low = atoi(strings.Replace(LowPrice, ",", "", -1))
			oneBar.Close = atoi(strings.Replace(ClosePrice, ",", "", -1))
			barArr = append(barArr, oneBar)
		}
	}
	// end of read file

	buf, jerr := json.Marshal(&barArr)
	if jerr != nil {
		http.Error(w, "json error", 406)
		return
	}

	w.Write(buf)
}

/**
*	webPageProc
**/
func webPageProc(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	defer func() {
		r.Body.Close()
	}()

	ti := strings.LastIndex(r.URL.Path, "/")

	path := strings.Replace(r.URL.Path[ti+1:], "/", "\\", -1)

	if path == "" {
		http.ServeFile(w, r, ".\\www\\index.html")
	} else if strings.Index(path, ".html") != -1 {
		http.ServeFile(w, r, ".\\www\\"+path)
	} else if strings.Index(path, ".dat") != -1 {
		dataProc(w, r, path)
	} else if path == "bg.png" {
		http.ServeFile(w, r, ".\\www\\bg.png")
	} else if strings.Index(path, ".css") != -1 || strings.Index(path, ".js") != -1 {
		http.ServeFile(w, r, ".\\www\\"+path)
	} else {
		http.Error(w, "404 not found", 404)
	}
}

/**
*	webserv
**/
func webserv() {

	http.HandleFunc("/", webPageProc)

	servAddr := fmt.Sprintf("%s:%d", "127.0.0.1", adminSrvPort)

	err := http.ListenAndServe(servAddr, nil)

	if err != nil {
		os.Exit(0)
	}
}

/**
*	main
**/
func main() {

	go openWebPage(true)

	webserv()
}
