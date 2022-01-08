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

var filetype string

var outfile string

var done bool = false

var noaudio bool

var novideo bool

var red = "\033[31m"

var reset = "\033[0m"

var green = "\033[32m"

/*var Yellow = "\033[33m"

var Blue = "\033[34m"

var Purple = "\033[35m"

var Cyan = "\033[36m"

var White = "\033[37m"*/

func initg() {
	flag.StringVar(&filetype, "x", "video", "What you want to download [gif, video]")
	flag.StringVar(&outfile, "o", "placeholder", "Output file")
	flag.StringVar(&url, "u", "NONE", "Url of the post")
	flag.BoolVar(&noclean, "noclean", false, "Don't remove separate audio and video files after merging them")
	flag.BoolVar(&noaudio, "noaudio", false, "Don't download audio, just video")
	flag.BoolVar(&novideo, "novideo", false, "Don't download video, just audio")
	flag.Parse()

	if noaudio {
		noclean = true //You dont need to clean if no tempfiles are created to be merged right ?
	} else if novideo {
		noclean = true
	}

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
		bar = fmt.Sprint(bar, "#")
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

	/* This function will attempt to turn the url into a .json so you don't have to ;)*/

	if strings.HasSuffix(url, ".json") {
		return
	}
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
		url = fmt.Sprint(url, ".json")
	}

}

func Generaterandomstring() string {

	rand.Seed(time.Now().UnixNano())

	RandomName := make([]rune, 10)

	for i := range RandomName {

		RandomName[i] = letterRunes[rand.Intn(len(letterRunes))]

	}

	return string(RandomName)

}

func DLfile(url string, saveas string, size int64) {

	/* This function will download an mp4 file and save the file names in a variable to merge them later */

	var wg sync.WaitGroup

	var extension string

	filename := Generaterandomstring()

	switch saveas {
	case "video":
		extension = ".mp4"
	case "audio":
		extension = ".mp4"
	case "gif":
		if outfile == "placeholder" {
			outfile = Generaterandomstring() + ".gif" //If you want to use the -o flag you should also specify the extension, so if you use it file extensions won't be automatically handled (or else we would end up wiht file.extension.extension)

		}
		filename = outfile //Since we only have to download the gif itself and no audio, we don't need temporary files, so we can "directly" name the downloaded file as the output file specified in args.
	}

	done = false

	file, err := os.Create(fmt.Sprint(filename, extension))
	checkerror(err)

	wg.Add(1)
	go func() {
		defer wg.Done()
		DownloadProgress(size, filename+extension)
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

	switch saveas {
	case "video":
		videofile = string(filename) + ".mp4"

	case "audio":
		audiofile = string(filename) + ".mp4"
	}

	wg.Wait()

}

func Mergeaudioandvideo() {
	fmt.Println("\n[+] Merging audio and video...")

	args := []string{"-i", videofile, "-i", audiofile, "-c:v", "copy", "-c:a", "aac", outfile}

	cmd := exec.Command("ffmpeg", args...)

	err := cmd.Run()
	checkerror(err)
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

func main() {

	fmt.Println("[+] Initializing...")

	initg()

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

	switch filetype {
	case "video":
		if outfile == "placeholder" {
			outfile = Generaterandomstring() + ".mp4" //If you want to use the -o flag you should also specify the extension, so if you use it file extensions won't be automatically handled (or else we would end up wiht file.extension.extension)

		}

		video := gjson.Get(string(body), "0.data.children.0.data.secure_media.reddit_video.fallback_url")

		if !novideo {

			fmt.Println("\n[+] Downloading source video file...")

			DLfile(video.String(), "video", Getsize(video.String()))

		}

		if !noaudio {

			fmt.Println("\n[+] Downloading audio...")

			re := regexp.MustCompile(`(?s)\_(.*)\.`)

			m := re.ReplaceAllString(video.String(), "_audio.")

			DLfile(m, "audio", Getsize(m))

		}

		if !noaudio && !novideo {

			Mergeaudioandvideo()

		} else {
			if noaudio {

				err = os.Rename(videofile, outfile)
				checkerror(err)

			}
			if novideo {

				err = os.Rename(audiofile, outfile)
				checkerror(err)

			}

		}

		if !noclean {
			fmt.Println("\n[+]Cleaning...")

			err = os.Remove(videofile)
			checkerror(err)

			err = os.Remove(audiofile)
			checkerror(err)

		}
		fmt.Println(green, "\r--Video successfully downloaded as ", outfile, "!--", reset)

	case "gif":

		video := gjson.Get(string(body), "0.data.children.0.data.url_overridden_by_dest")

		fmt.Println("\n[+] Downloading GIF...")
		DLfile(video.String(), "gif", Getsize(video.String()))

		fmt.Println(green, "\r--GIF successfully downloaded as ", outfile, "!--", reset)

	default:
		log.Fatal(red, `Invalid file type specified, please choose "video" or "gif"`, reset)

	}

}
