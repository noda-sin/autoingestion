package main

import (
	"compress/gzip"
	"flag"
	"github.com/golang/glog"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func main() {

	const ad = time.Duration(2) * time.Hour * 24 // 2 days
	defaultDay := time.Now().Add(-ad).UTC().Format("20060102")

	user := flag.String("user", "", "User for iTunes connect")
	pass := flag.String("pass", "", "User for iTunes connect")
	vnd := flag.String("vnd", "", "Vendor number for iTunes connect")
	date := flag.String("date", defaultDay, "YYYYMMDD  Date time of log")
	proxy := flag.String("proxy", "", "PROTOCOL://HOST[:PORT]  Use proxy on given port")

	flag.Parse()

	if len(*user) == 0 || len(*pass) == 0 {
		glog.Fatalln("User and pass required")
		return
	}

	if len(*vnd) == 0 {
		glog.Fatalln("Vendor number required")
		return
	}

	if len(*proxy) > 0 {
		os.Setenv("HTTP_PROXY", *proxy)
	}

	params := url.Values{
		"USERNAME":     {*user},
		"PASSWORD":     {*pass},
		"VNDNUMBER":    {*vnd},
		"TYPEOFREPORT": {"Sales"},
		"DATETYPE":     {"Daily"},
		"REPORTTYPE":   {"Summary"},
		"REPORTDATE":   {*date},
	}

	resp, err := http.PostForm("https://reportingitc.apple.com/autoingestion.tft?", params)

	if err != nil {
		glog.Fatalln(err)
		return
	}

	defer resp.Body.Close()

	gzFilename := resp.Header.Get("filename")
	gzOut, err := os.Create(gzFilename)

	if err != nil {
		glog.Fatalln(err)
		return
	}

	defer gzOut.Close()

	glog.Infoln("Downloading", gzFilename)

	_, err = io.Copy(gzOut, resp.Body)

	if err != nil {
		glog.Fatalln(err)
		return
	}

	gzFile, err := os.Open(gzFilename)

	if err != nil {
		glog.Fatalln(err)
		return
	}

	defer gzFile.Close()

	gzIn, err := gzip.NewReader(gzFile)

	if err != nil {
		glog.Fatalln(err)
		return
	}

	defer gzIn.Close()

	txtFilename := strings.Replace(gzFilename, ".gz", "", -1)
	txtOut, err := os.Create(txtFilename)

	if err != nil {
		glog.Fatalln(err)
		return
	}

	defer txtOut.Close()

	_, err = io.Copy(txtOut, gzIn)

	if err != nil {
		glog.Fatalln(err)
		return
	}
}
