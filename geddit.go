package main

import (
	"flag"
	"fmt"
	"github.com/tidwall/gjson" //For json parsing
	"io"
	"io/ioutil"
	"log"
	"math/rand" //For random file names
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var url string

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var noclean bool

var videofile string

var audiofile string

var outfile string

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

	fmt.Println("Initializing...")

	initflags()

	correcturl()

	if url == "NONE" {
		log.Fatal(red, "No post URL specified. Quitting.", reset)

	}

	fmt.Println(strings.TrimSuffix(url, "+"))

	req, err := http.NewRequest("GET", url, nil)

	checkerror(err)

	req.Header.Set("User-Agent", "Mozilla/5.0") //We need a semi-credible user agent or else reddit will just block the request

	resp, err := new(http.Client).Do(req)

	checkerror(err)

	defer resp.Body.Close()

	fmt.Println("Fetching url of the source video file...")

	body, err := ioutil.ReadAll(resp.Body)

	checkerror(err)

	video := gjson.Get(string(body), "0.data.children.0.data.secure_media.reddit_video.fallback_url")

	checkerror(err)

	fmt.Println(video.String())

	fmt.Println("Downloading source video file...")

	DLfile(video.String(), "video")

	fmt.Println("Downloading audio...")

	re := regexp.MustCompile(`(?s)\_(.*)\.`)

	m := re.ReplaceAllString(video.String(), "_audio.")

	fmt.Println(m)

	DLfile(m, "audio")

	fmt.Println("Merging audio and video...")

	args := []string{"-i", videofile, "-i", audiofile, "-c:v", "copy", "-c:a", "aac", outfile}

	cmd := exec.Command("ffmpeg", args...)

	fmt.Print(cmd.String(), "\n")

	err = cmd.Run()
	checkerror(err)

	if !noclean {
		fmt.Println("\nCleaning...")

		err = os.Remove(audiofile)
		checkerror(err)

		err = os.Remove(videofile)
		checkerror(err)

	}

	fmt.Println(green, "Video successfully downloaded as " , outfile, "!", reset)

}

func DLfile(url string, saveas string) {

	/* This function will download an mp4 file and save the file names in a variable to merge them later */

	rand.Seed(time.Now().UnixNano())

	req, err := http.NewRequest("GET", url, nil)

	checkerror(err)

	req.Header.Set("User-Agent", "Mozilla/5.0") 

	fil, err := new(http.Client).Do(req)

	checkerror(err)

	RandomName := make([]rune, 10)

	for i := range RandomName {

		RandomName[i] = letterRunes[rand.Intn(len(letterRunes))]

	}

	file, err := os.Create(fmt.Sprint(string(RandomName), ".mp4"))

	checkerror(err)

	_, err = io.Copy(file, fil.Body)
	checkerror(err)

	if saveas == "video" {
		videofile = string(RandomName) + ".mp4"
	} else {
		audiofile = string(RandomName) + ".mp4"
	}

}
