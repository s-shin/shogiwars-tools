package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type HistoryPage struct {
	UserName string
	GameType GameType
	Page     int
}

const NumRecordItemsPerPage = 10

func (p *HistoryPage) BuildURL() string {
	return fmt.Sprintf("https://shogiwars.heroz.jp/users/history/%s/web_app?gtype=%s&start=%d",
		p.UserName, p.GameType.ParamString(), p.Page*NumRecordItemsPerPage)
}

var reTime = regexp.MustCompile(`\d{4}/\d{2}/\d{2} \d{2}:\d{2}`)
var reRecordID = regexp.MustCompile(`.*//.+/games/([^?]+)`)

func (page *HistoryPage) FetchRecordItems() ([]*RecordItem, error) {
	url := page.BuildURL()
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// log.Println(doc.Text())

	items := make([]*RecordItem, 0)
	doc.Find(".contents").Each(func(i int, s *goquery.Selection) {
		item := &RecordItem{GameType: page.GameType}
		func() {
			text := s.Find(".game_date").Text()
			if text == "" {
				return
			}
			loc := reTime.FindStringIndex(text)
			if loc == nil {
				return
			}
			t, err := time.ParseInLocation("2006/01/02 15:04", text[loc[0]:loc[1]], time.Local)
			if err != nil {
				return
			}
			item.Date = t
		}()
		func() {
			var opponent string
			s.Find(".players > div > a").Each(func(i int, s *goquery.Selection) {
				fields := strings.Fields(s.Text())
				if len(fields) == 0 {
					return
				}
				p := Player{fields[0], fields[1]}
				if p.UserName != page.UserName {
					opponent = p.UserName
				}
				if i == 0 {
					item.Players.Black = p
				} else {
					item.Players.White = p
				}
			})
			if opponent == "" {
				return
			}
			if s.HasClass("winner") {
				item.Winner = page.UserName
			} else {
				item.Winner = opponent
			}
		}()
		{
			s.Find(".hashtag_badge").Each(func(i int, s *goquery.Selection) {
				item.Tags = append(item.Tags, strings.TrimSpace(s.Text()))
			})
		}
		func() {
			recordURL, ok := s.Find(".game_replay > a").First().Attr("href")
			if !ok {
				return
			}
			m := reRecordID.FindStringSubmatch(recordURL)
			if m == nil {
				return
			}
			item.RecordID = RecordID(m[1])
		}()
		items = append(items, item)
	})
	return items, nil
}
