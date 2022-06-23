package main
import (
	"net/http"
	"encoding/json"
	"context"
	"fmt"
	"log"
	"time"
	"os"
	"os/signal"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

//메인함수
func main() {
	//서비스로 실행하기 위한 OS시그널처리
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs)
	go func() {
		s := <-sigs
		log.Printf("받은 시그널: %s\n", s)
		AppCleanup()
		os.Exit(1)
	}()
	
	//8000번 포트로 http 서버열기
	//nginx연결됨 (https://git.coco.sqs.kr/proxy-8000)
	err := http.ListenAndServe(":8000", http.HandlerFunc(requestHandler))
	if err != nil {
		//http 서버 실행실패시 에러처리
		log.Println("Failed to ListenAndServe : ", err)
	}
}

//err.Error() 를 통해 error메시지를 string으로 변환
type errorString struct {
	s string
}
func (e *errorString) Error() string {
	return e.s
}

//요청이 들어오면 실행되는 함수
func requestHandler(res http.ResponseWriter, req *http.Request) {
	
	//반환될 Response 사전정의
	res.WriteHeader(http.StatusCreated)
	resdata := make(map[string]string)
	
	//FORM > POST 데이터 가져오기
	req.ParseForm()
	postdata := req.PostForm
	
	//POST 데이터에서 url이라는 값을 찾아서 String을 벗기기(?)
	if ( postdata["url"] != nil && postdata["element"] != nil){ 
		
		//Map풀기
		url := postdata["url"][0]
		element := postdata["element"][0]
		
		//Chromedp설정
		taskCtx, cancel := chromedp.NewContext(
			context.Background(),
			chromedp.WithLogf(log.Printf),
		)
		defer cancel()
		
		//최대 대기시간은 15초
		taskCtx, cancel = context.WithTimeout(taskCtx, 15*time.Second)
		defer cancel()
		
		//사이트 캡쳐해서 버퍼생성
		var pdfBuffer []byte
		if err := chromedp.Run(taskCtx, pdfGrabber(url, element, &pdfBuffer)); err != nil {
			//실패시 fail출력
			res.Header().Set("Content-Type", "application/json")
			resdata["status"] = "fail"
			resdata["errormsg"] = err.Error()
			output, err := json.Marshal(resdata)
			if err != nil {
				log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			}
			res.Write(output)
			return 
		}
		
		//성공시 PDF형태로출력
		res.Header().Set("Content-Type", "application/pdf")
		res.Write(pdfBuffer)
		return 
		
	}else{
		//실패시 fail출력
		res.Header().Set("Content-Type", "application/json")
		resdata["status"] = "fail"
		resdata["errormsg"] = "Parameter ERROR!"
		output, err := json.Marshal(resdata)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		res.Write(output)
		return 
	}
}

//PDF생성함수
func pdfGrabber(url string, sel string, res *[]byte) chromedp.Tasks {
	//실행시간 측정시작
	//start := time.Now()
	return chromedp.Tasks{
		emulation.SetUserAgentOverride("WebScraper 1.0"), //USER AGENT설정
		chromedp.Navigate(url),
		chromedp.WaitVisible(sel, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			//실행시간 측정종료
			//fmt.Printf("\nDuration: %f secs\n", time.Since(start).Seconds())
			return nil
		}),
	}
}
