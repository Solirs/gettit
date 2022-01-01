package main

import (
	"flag"
	"fmt"
	humanize "github.com/dustin/go-humanize" //For conversion from bytes to kilobytes megabytes etc..
	"github.com/tidwall/gjson"               //For json parsing
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand" //For random file names
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var url string

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var noclean bool

var videofile string

var audiofile string

var outfile string

var done bool = false

var red = "\033[31m"

var reset = "\033[0m"

var green = "\033[32m"

/*var Yellow = "\033[33m"

var Blue = "\033[34m"

var Purple = "\033[35m"

var Cyan = "\033[36m"

var White = "\033[37m"*/

func initflags() {
	flag.StringVar(&outfile, "o", "out.mp4", "Output file")
	flag.StringVar(&url, "u", "NONE", "Url of the post")
	flag.BoolVar(&noclean, "noclean", false, "Don't remove separate audio and video files after merging them")
	flag.Parse()
}

func checkerror(err error) {
	if err != nil {
		log.Fatal(err)
	}

}

func Printprogress(path string, total float64) {

	file, err := os.Open(path)
	checkerror(err)

	fi, err := file.Stat()
	checkerror(err)
	size := fi.Size()

	if size == 0 {
		size = 1
	}

	var percent float64 = float64(size) / float64(total) * 100 //Get the percentage of size to total

	barpercent := math.Ceil(percent/10) * 10 //Ceil the percentage to 10

	var bar string

	bar = fmt.Sprint(bar, "[")

	for i := 0; i < int(barpercent)/5; i++ {
		bar = fmt.Sprint(bar, "#") //Fill the loading bar with one # for every 5%
	}

	for i := 0; i < 20-int(barpercent)/5; i++ {
		bar = fmt.Sprint(bar, " ") //And fill the rest with spaces, we want the bar to be 20 wide (excluding the brackets) so we fill the difference between the number of bars we will fill and 20
	}

	bar = fmt.Sprint(bar, "]")

	percentformat := fmt.Sprintf("%.0f", percent)

	fmt.Print(green, "\r", bar, percentformat, "% ", humanize.Bytes(uint64(size)), "/", humanize.Bytes(uint64(total)))

	fmt.Print(reset)

}

func DownloadProgress(total int64, path string) {

	stop := false

	for !stop {
		switch done {
		case true:
			stop = true

			Printprogress(path, float64(total))

		default:

			Printprogress(path, float64(total))

		}
		time.Sleep(time.Millisecond * 100)
	}
}

func correcturl() {

	/* This function will attempt to turn the url into a .json*/

	if strings.HasSuffix(url, ".json") {
		return
	}
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
		url = fmt.Sprint(url, ".json")
	}

}

func main() {

	fmt.Println("[+] Initializing...")

	initflags()

	correcturl()

	if url == "NONE" {
		log.Fatal(red, "No post URL specified. Quitting.", reset)

	}

	req, err := http.NewRequest("GET", url, nil)

	checkerror(err)

	req.Header.Set("User-Agent", "Mozilla/5.0") //We need a semi-credible user agent or else reddit will just block the request

	resp, err := new(http.Client).Do(req)

	checkerror(err)

	defer resp.Body.Close()

	fmt.Println("\n[+] Fetching url of the source video file...")

	body, err := ioutil.ReadAll(resp.Body)

	checkerror(err)

	video := gjson.Get(string(body), "0.data.children.0.data.secure_media.reddit_video.fallback_url")

	checkerror(err)

	fmt.Println("\n[+] Downloading source video file...")

	DLvideo(video.String(), "video", Getsize(video.String()))

	fmt.Println("\n[+] Downloading audio...")

	re := regexp.MustCompile(`(?s)\_(.*)\.`)

	m := re.ReplaceAllString(video.String(), "_audio.")

	DLvideo(m, "audio", Getsize(m))

	fmt.Println("\n[+] Merging audio and video...")

	args := []string{"-i", videofile, "-i", audiofile, "-c:v", "copy", "-c:a", "aac", outfile}

	cmd := exec.Command("ffmpeg", args...)

	err = cmd.Run()
	checkerror(err)

	if !noclean {
		fmt.Println("\n[+]Cleaning...")

		err = os.Remove(audiofile)
		checkerror(err)

		err = os.Remove(videofile)
		checkerror(err)

	}

	fmt.Println(green, "\r--Video successfully downloaded as ", outfile, "!--", reset)

}

func Generaterandomstring() string{

	rand.Seed(time.Now().UnixNano())

	RandomName := make([]rune, 10)

	for i := range RandomName {

		RandomName[i] = letterRunes[rand.Intn(len(letterRunes))]

	}

	return string(RandomName)


}

func DLvideo(url string, saveas string, size int64) {

	var wg sync.WaitGroup

	/* This function will download an mp4 file and save the file names in a variable to merge them later */

	done = false

	filename := Generaterandomstring()


	file, err := os.Create(fmt.Sprint(filename, ".mp4"))
	checkerror(err)

	wg.Add(1)
	go func() {
		defer wg.Done()
		DownloadProgress(size, filename +".mp4")
	}()

	req, err := http.NewRequest("GET", url, nil)

	checkerror(err)

	req.Header.Set("User-Agent", "Mozilla/5.0")

	fil, err := new(http.Client).Do(req)

	checkerror(err)

	checkerror(err)

	_, err = io.Copy(file, fil.Body)
	checkerror(err)

	done = true

	if saveas == "video" {
		videofile = string(filename) + ".mp4"
	} else {
		audiofile = string(filename) + ".mp4"
	}

	wg.Wait()

}

func Getsize(url string) int64 {

	req, err := http.NewRequest("HEAD", url, nil)

	checkerror(err)

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := new(http.Client).Do(req)
	checkerror(err)

	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	checkerror(err)

	return int64(size)

}
