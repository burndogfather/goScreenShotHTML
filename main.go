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

func main() {
	
	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("Hello World"))
	})
	
	http.ListenAndServe(":5000", nil)
	
	
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