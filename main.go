package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var (
	downloadDir = path.Join("Pictures", "bigpicture")
	dirHelp     = "Download dir. Defaults to ~" + downloadDir
	dir         = flag.String("dir", "", dirHelp)
)

func downloadImage(wg *sync.WaitGroup, href string) {
	defer wg.Done()

	name := path.Base(href)
	fp := path.Join(downloadDir, name)

	_, err := os.Stat(fp)
	if err != nil {
		img, err := http.Get("http:" + href)
		if err != nil {
			log.Printf("Error while downloading %s. Reason: %s\n", href, err.Error())
			return
		}

		file, err := os.Create(fp)
		if err != nil {
			log.Printf("Error while creating path %s. Reason: %s\n", fp, err.Error())
			return
		}

		io.Copy(file, img.Body)
		fmt.Printf("+")

		defer file.Close()
		defer img.Body.Close()

	} else {
		fmt.Printf(".")
	}
}

func loadPage(wg *sync.WaitGroup, href string) {
	defer wg.Done()

	doc, err := goquery.NewDocument("http://www.bostonglobe.com" + href)
	if err != nil {
		log.Fatal(err)
	}

	images := doc.Find(".photo img")
	var iwg sync.WaitGroup

	iwg.Add(images.Length())
	images.Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		go downloadImage(&iwg, src)
	})
	iwg.Wait()
}

func init() {
	flag.StringVar(dir, "d", "", dirHelp)
}

func createDir(path string) {
	if info, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.Mkdir(path, 0777); err != nil {
			log.Fatal(err)
			return
		}
	} else {
		if !info.IsDir() {
			log.Fatalf("%s exists but it's not a dir. Stopping.", path)
		}
	}
}

func main() {
	flag.Parse()

	if *dir != "" {
		downloadDir = *dir
	} else {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		downloadDir = path.Join(usr.HomeDir, downloadDir)
	}

	createDir(downloadDir)

	doc, err := goquery.NewDocument("http://www.bostonglobe.com/news/bigpicture")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	sel := doc.Find(".pictureInfo-headline")
	wg.Add(sel.Length())

	sel.Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		go loadPage(&wg, href)
	})

	wg.Wait()
}
