package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"

	"github.com/antchfx/htmlquery"
)

func main() {

	var URL string

	downloadLocation := flag.String("path", ".", "Path where to download mods")
	flag.Parse()

	if len(os.Args) == 2 {
		URL = os.Args[1]
	} else {
		fmt.Print("Please provide url: ")
		fmt.Scanln(&URL)
	}

	modList := findModURLs(URL)

	if len(modList) == 0 {
		fmt.Println("No mods found or wrong URL")
		return
	}
	var wg sync.WaitGroup
	for _, mod := range modList {
		wg.Add(1)
		go downloadModZipFile(getDownloadLinkFromModSite(mod), *downloadLocation, mod, &wg)
	}
	wg.Wait()
	println("Downloaded all mods")
}

func findModURLs(multiModURL string) []string {
	ret := make([]string, 0)

	doc, err := htmlquery.LoadURL(multiModURL)
	if err != nil {
		fmt.Println("Erorr getting mods URL: ", err)
		return ret
	}

	list, err := htmlquery.QueryAll(doc, "//a[contains(@href, 'mod.php')]")

	if err != nil {
		fmt.Println("Erorr finding mod links: ", err)
		return ret
	}
	for _, n := range list {
		// fmt.Printf("%d %s(%s)\n", i, htmlquery.InnerText(n), htmlquery.SelectAttr(n, "href"))
		ret = append(ret, htmlquery.SelectAttr(n, "href"))
	}
	return ret
}

func getDownloadLinkFromModSite(modURL string) string {
	// //a[contains(@href, 'modHub/storage')]
	doc, err := htmlquery.LoadURL("https://www.farming-simulator.com/" + modURL)
	n, err := htmlquery.Query(doc, "//a[contains(@href, 'modHub/storage')]")
	if err != nil {
		fmt.Println("Erorr while getting downlaod link: ", err)
	}
	return htmlquery.SelectAttr(n, "href")
}

func downloadModZipFile(modDownloadURL, downloadLocation, refURL string, wg *sync.WaitGroup) {
	defer wg.Done()
	// Get name from base path from url
	modParsURL, _ := url.Parse(modDownloadURL)
	urlPath := modParsURL.Path
	name := path.Base(urlPath)

	downloadPath := downloadLocation + "/" + name
	// ITS BAD TO CHECK !os.IsNotExist(err)
	if _, err := os.Stat(downloadPath); !os.IsNotExist(err) {
		fmt.Println(name, "already downloaded")
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", modDownloadURL, nil)
	if err != nil {
		fmt.Println("Error preparing download req", err)
		return
	}
	req.Header.Add("Referer", "https://www.farming-simulator.com/"+refURL)

	resp, err := client.Do(req)
	if err != nil {
		print("Erroed while requesting mod", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Print("Could not donwload mod ", name, " got status: ", resp.StatusCode)
		return
	}

	fmt.Println("Downloading mod: ", name, "...")

	out, err := os.Create(downloadPath + ".downloading")
	defer out.Close()
	if err != nil {
		fmt.Println("Error creating a file", err)
		return
	}
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error when copying file", err)
	}
	os.Rename(downloadPath+".downloading", downloadPath)
	fmt.Println("Completed downloading", name)
	return
}
