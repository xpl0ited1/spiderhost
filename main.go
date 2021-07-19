package main

import (
	"context"
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

//Global Constants

const letterBytes = "abcdefghijklmnopqrstuvwxyz234567"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)
const onion = ".onion"

//Global Variables

var othersTLD = [...]string{
	".cl",
	".com",
	".org",
	".co",
	".uk",
	".me",
	".site",
	".crypto",
	".site",
	".cafe",
}
var src = rand.NewSource(time.Now().UnixNano())

var opt = "o"

//TOR can be changed to wathever ip and port
var TOR = "127.0.0.1:9050"

var validHostsFilePath = "hosts.txt"

//Functions
var ch = make(chan time.Time, 1)

//main function
func main() {
	ct := context.Background()
	fmt.Println("Running")
	x := 1
	for true {

		fmt.Println("Cycle " + strconv.Itoa(x) + " started")

		test(x, ct)

		x++
	}
}

func test(x int, ct context.Context) {

	var wg sync.WaitGroup
	for i := 1; i <= 20; i++ {
		wg.Add(1)
		go worker(i, &wg, x, ct)
	}
	wg.Wait()
	fmt.Println("Cycle " + strconv.Itoa(x) + " done")
}

func worker(id int, wg *sync.WaitGroup, i int, ct context.Context) {

	ctx, cancel := context.WithTimeout(ct, 5*time.Second)
	defer wg.Done()
	defer cancel()

	select {
	case <-time.After(3 * time.Second):
		host := generateOne()
		fmt.Println("[Cycle #" + strconv.Itoa(i) + " Worker #" + strconv.Itoa(id) + "]Testing " + host)
		isAlive, err := pingHost(host)

		if err != nil {
			//log.Println(err)
		}

		if isAlive {
			fmt.Println("[Cycle #" + strconv.Itoa(i) + " Worker #" + strconv.Itoa(id) + "] " + host + " is a valid host")
			err := saveValidHost(host)
			if err != nil {
				log.Println(err)
				return
			}
			takeScreenshot(host)
		} else {
			fmt.Println("[Cycle #" + strconv.Itoa(i) + " Worker #" + strconv.Itoa(id) + "] " + host + " is not a valid host")
		}

		return
	case <-ctx.Done():
		fmt.Println("done") // prints "context deadline exceeded"
	}

	//time.Sleep(time.Second*5)
}

func RandStringBytesMaskImprSrc(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func generateOne() string {
	min := 16
	max := 56
	tld := ""

	rand.Seed(time.Now().UnixNano())

	if opt == "o" {
		tld = onion
	} else {
		tld = othersTLD[rand.Intn(len(othersTLD)-1+1)]
	}

	host := RandStringBytesMaskImprSrc(rand.Intn(max-min+1)+min) + tld
	//For debugging
	//host = "3g2upl4pq6kufc4m.onion"

	return host
}

func pingHost(host string) (bool, error) {
	os.Setenv("HTTP_PROXY", "socks5://"+TOR)

	//example url propub3r6espa33w.onion
	//host = "3g2upl4pq6kufc4m.onion"

	_, err := http.Get("http://" + host)

	if err != nil {
		return false, err
		log.Println(err)
	}
	return true, nil
}

func getHostContent(host string) (string, error) {
	os.Setenv("HTTP_PROXY", "socks5://"+TOR)

	//example url propub3r6espa33w.onion
	//host = "3g2upl4pq6kufc4m.onion"

	resp, err := http.Get("http://" + host)

	if err != nil {
		return "", err
		log.Println(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println(fmt.Errorf("Status error: %v", resp.StatusCode))
		return "", err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(fmt.Errorf("Read body: %v", err))
		return "", err
	}
	//fmt.Println(string(data))
	var html = string(data)
	return html, nil
}

func saveValidHost(host string) error {
	err := createFile(validHostsFilePath)
	if err != nil {
		return err
		log.Println(err)
	}

	err = writeFile(validHostsFilePath, host)
	if err != nil {
		return err
		log.Println(err)
	}
	return nil
}

func createFile(path string) error {
	// check if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

func writeFile(path string, data string) error {
	// Open file using READ & WRITE permission.
	var file, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write some text line-by-line to file.
	_, err = file.WriteString(data + "\n")
	if err != nil {
		return err
	}

	// Save file changes.
	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}

func saveScreenshot(host string, data []byte) error {
	// Open file using READ & WRITE permission.
	var file, err = os.OpenFile("./out/"+host+".png", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write some text line-by-line to file.
	_, err = file.Write(data)
	if err != nil {
		return err
	}

	// Save file changes.
	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}

//Debugging functions
func testGenerate() {
	min := 10
	max := 30
	tld := ""

	rand.Seed(time.Now().UnixNano())

	if opt == "o" {
		tld = onion
	} else {
		tld = othersTLD[rand.Intn(len(othersTLD)-1+1)]
	}

	host := RandStringBytesMaskImprSrc(rand.Intn(max-min+1)+min) + tld

	fmt.Println(host)
	testPing(host)
}

func testPing(host string) {

	os.Setenv("HTTP_PROXY", "socks4://127.0.0.1:9050")

	//example url propub3r6espa33w.onion
	host = "3g2upl4pq6kufc4m.onion"
	takeScreenshot(host)

}

func takeScreenshot(host string) {

	const (
		// These paths will be different on your system.
		seleniumPath     = "selenium-server-standalone-3.141.59.jar"
		geckoDriverPath  = "geckodriver"
		chromeDriverPath = "chromedriver"
		port             = 8080
	)
	opts := []selenium.ServiceOption{
		//selenium.StartFrameBuffer(),           // Start an X frame buffer for the browser to run in.
		selenium.ChromeDriver(chromeDriverPath), // Specify the path to GeckoDriver in order to use Firefox.
		selenium.Output(os.Stderr),              // Output debug information to STDERR.
	}
	selenium.SetDebug(true)
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		log.Println(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{}

	caps.AddChrome(chrome.Capabilities{
		Path: "/usr/bin/google-chrome",
		Args: []string{"--headless", "--proxy-server=socks5://127.0.0.1:9050"},
	})
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	// Navigate to the simple playground interface.
	if err := wd.Get("https://" + host); err != nil {
		log.Println(err)
	}

	data, err := wd.Screenshot()
	if err != nil {
		log.Println(err)
	}

	saveScreenshot(host, data)

}
