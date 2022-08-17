package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"log"
	"os"
)

type Anime struct {
	Link        string   `json:"link,omitempty"`
	Nombre      string   `json:"nombre,omitempty"`
	LinkImagen  string   `json:"link_imagen,omitempty"`
	Descripcion string   `json:"descripcion,omitempty"`
	Generos     []string `json:"generos,omitempty"`
	Estado      string   `json:"estado,omitempty"`
}

func main() {
	var animes []Anime

	c := colly.NewCollector(

		colly.MaxDepth(2000000),
		colly.Async(),
	)
	err := c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 100})
	if err != nil {
		return
	}

	c.OnError(func(r *colly.Response, e error) {
		log.Println("error:", e, r.Request.URL, string(r.Body))
	})
	// Visitar la siguiente pagina
	c.OnHTML("a[rel=\"next\"]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		fmt.Println("Visiting: ", link)
		// Visitar la nueva pagina en un hilo nuevo
		err := e.Request.Visit(link)
		if err != nil {
			return
		}
	})
	// Visitar los animes
	c.OnHTML("div.heromain > div.row > div.col-md-4 > a", func(element *colly.HTMLElement) {
		episodeLink := element.Attr("href")
		err := element.Request.Visit(episodeLink)
		if err != nil {
			return
		}
	})
	// Agregar el anime
	c.OnHTML("div.heroarea > div.heromain > div.acontain", func(element *colly.HTMLElement) {
		anime := Anime{
			Link:        element.Request.URL.String(),
			Nombre:      element.ChildText("h1.mobh1"),
			LinkImagen:  element.ChildAttr("div.chapterpic>img", "src"),
			Descripcion: element.ChildText("div.chapterdetls2>p"),
			Generos: element.DOM.Find("div.chapterdetls2 > table > tbody > tr > td > a").Map(func(n int, s *goquery.Selection) string {
				return s.Text()
			}),
			Estado: element.ChildText("div.chapterdetls2 > table > tbody > tr:nth-child(2) > td:nth-child(2)"),
		}
		animes = append(animes, anime)
	})
	// Visitar la pagina de animes
	err = c.Visit("https://monoschinos2.com/animes")
	if err != nil {
		return
	}
	c.Wait()
	content, err := json.MarshalIndent(animes, "", "   ")
	if err != nil {
		fmt.Println(err.Error())
	}
	err = os.WriteFile("animes.json", content, 0644)
	if err != nil {
		fmt.Println(err.Error())
	}
}
