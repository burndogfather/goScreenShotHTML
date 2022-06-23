package main
import (
	"net/http"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func requestHandler(rw http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	fmt.Fprintln(rw, "(1)", r.FormValue("hello"))
	fmt.Println("request.Form::")
	for key, value := range req.Form {
		fmt.Printf("Key:%s, Value:%s\n", key, value)
	}
	
	
	fmt.Println("Method : ", req.Method)
	fmt.Println("URL : ", req.URL)
	fmt.Println("Header : ", req.Header)
 
	b, _ := ioutil.ReadAll(req.Body)
	defer req.Body.Close()
	fmt.Println("Body : ", string(b))
 
	switch req.Method {
	case "POST":
		rw.Write([]byte("post request success !"))
	case "GET":
		rw.Write([]byte("get request success !"))
	}
}

func main() {
	
	err := http.ListenAndServe(":8000", http.HandlerFunc(requestHandler))
	if err != nil {
		fmt.Println("Failed to ListenAndServe : ", err)
	}
	
	
	taskCtx, cancel := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()
	taskCtx, cancel = context.WithTimeout(taskCtx, 15*time.Second)
	defer cancel()
	var pdfBuffer []byte
	if err := chromedp.Run(taskCtx, pdfGrabber("https://www.naver.com", "body", &pdfBuffer)); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("naver.pdf", pdfBuffer, 0644); err != nil {
		log.Fatal(err)
	}
}

func pdfGrabber(url string, sel string, res *[]byte) chromedp.Tasks {

	start := time.Now()
	return chromedp.Tasks{
		emulation.SetUserAgentOverride("WebScraper 1.0"),
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			fmt.Printf("\nDuration: %f secs\n", time.Since(start).Seconds())
			return nil
		}),
	}
}