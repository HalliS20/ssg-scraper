package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

var (
	client  *resty.Client
	links   []string
	baseUrl string
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: program <filename> <extra endings>")
		return
	}

	fmt.Println("Welcome to the scraping")
	baseUrl = os.Args[1]
	getHttp("/")

	if len(os.Args) > 2 {
		for _, value := range os.Args[2:] {
			getHttp(value)
		}
	}
}

func Contains[T comparable](arr []T, item T) bool {
	arrLen := len(arr)
	for i := 0; i < arrLen; i++ {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func getHttp(ending string) {
	if Contains(links, ending) {
		return
	}
	links = append(links, ending)
	client = resty.New()
	site, err := client.R().Get("https://" + baseUrl + ending)
	if err != nil {
		fmt.Println("get request error: ", err)
		return
	}
	respStr := site.String()

	findAnchors(respStr)
	stringToHtml(respStr, ending)
}

func findAnchors(respStr string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(respStr))
	if err != nil {
		fmt.Println("error parsing document: ", err)
	}

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			getHttp(href)
		} else {
			fmt.Println("No links found")
		}
	})
}

func dirPaths(path string) string {
	parts := strings.Split(path, "/")

	return strings.Join(parts[:len(parts)-1], "/")
}

func stringToHtml(htmlStr string, path string) {
	// neededDirs := dirPaths(baseUrl + path)
	err := os.MkdirAll(baseUrl+path, 0755)
	if err != nil {
		fmt.Println("error making directories: ")
		return
	}
	newFilePath := baseUrl + path + "/index.html"
	if path == "/" {
		newFilePath = baseUrl + "/index.html"
	}

	file, err := os.Create(newFilePath)
	if err != nil {
		fmt.Println("os create error: ", err)
		return
	}

	defer file.Close()

	_, err = file.WriteString(htmlStr)
	if err != nil {
		fmt.Println("error writing to file: ", err)
		return
	}

	fmt.Println("HTML content saved to ", newFilePath)
}
