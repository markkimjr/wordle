package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	espnURL = "http://www.espn.com/mma/fighters?search="
	ufcURL  = "https://www.ufc.com/athlete/"
)

type Scraper interface {
	ScrapeInfo()
	SaveInfo()
}

type Ufc struct {
}

type UfcSaveInfo struct {
	Name             string  `bson:"name"`
	Nickname         string  `bson:"nickname"`
	Status           string  `bson:"status"`
	Country          string  `bson:"country"`
	Style            string  `bson:"style"`
	Record           string  `bson:"record"`
	Division         string  `bson:"division"`
	Age              int     `bson:"age"`
	KnockoutWins     int     `bson:"knockoutWins"`
	SubmissionWins   int     `bson:"submissionWins"`
	Height           float32 `bson:"height"`
	Weight           float32 `bson:"weight"`
	Reach            float32 `bson:"reach"`
	StrikingAccuracy float32 `bson:"strikingAccuracy"`
	TakedownAccuracy float32 `bson:"takedownAccuracy"`
}

func CheckErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func (u Ufc) SaveInfo() {

}

func (u Ufc) ScrapeInfo() []*UfcSaveInfo {
	var allFighterNames []string
	ch := make(chan []string)
	for i := 'a'; i <= 'z'; i++ {
		go u.scrapeEspnFighterNames(string(i), ch)
		fighterNames := <-ch
		if fighterNames != nil {
			allFighterNames = append(allFighterNames, fighterNames...)
		}
	}

	var fighterSaveInfo []*UfcSaveInfo
	ch2 := make(chan *UfcSaveInfo)
	for _, fighterName := range allFighterNames {
		go u.scrapeUfcFighterInfo(fighterName, ch2)
		fighterInfo := <-ch2
		if fighterInfo.Name != "" {
			fighterSaveInfo = append(fighterSaveInfo, fighterInfo)
		}
	}

	return fighterSaveInfo
}

// scrapeEspnFighterNames gets all MMA fighter names from ESPN
func (u Ufc) scrapeEspnFighterNames(i string, ch chan<- []string) {
	var fighterNames []string
	fighterMap := make(map[string]bool) // check for duplicate fighter names
	res, err := http.Get(fmt.Sprintf("%v%v", espnURL, i))
	CheckErr(err)
	defer res.Body.Close()
	if res.StatusCode == 200 {
		doc, err := goquery.NewDocumentFromReader(res.Body)
		CheckErr(err)
		doc.Find(".evenrow").Each(func(i int, s *goquery.Selection) {
			fighterName := s.Find("a").Text()
			if _, ok := fighterMap[fighterName]; !ok {
				fighterNames = append(fighterNames, fighterName)
				fighterMap[fighterName] = true
			}
		})
		doc.Find(".oddrow").Each(func(i int, s *goquery.Selection) {
			fighterName := s.Find("a").Text()
			if _, ok := fighterMap[fighterName]; !ok {
				fighterNames = append(fighterNames, fighterName)
				fighterMap[fighterName] = true
			}
		})
	}

	ch <- fighterNames
}

// scrapeUfcFighterInfo requests UFC player profiles; if request is successful, creates UfcFighterSaveInfo struct
func (u Ufc) scrapeUfcFighterInfo(fighterName string, ch chan<- *UfcSaveInfo) {
	split := strings.Split(fighterName, ",")
	rearrangedFighterName := fmt.Sprintf("%v-%v", strings.ToLower(split[1]), strings.ToLower(split[0]))

	url := fmt.Sprintf("%v%v", ufcURL, rearrangedFighterName)
	res, err := http.Get(url)
	CheckErr(err)
	fighter := UfcSaveInfo{}
	if res.StatusCode == 200 {
		fighter = u.parseUfcFighterHtml(res.Body)
	}

	ch <- &fighter
}

// parseUfcFighterHtml parses UFC HTML to get fighter record, active status, division, age, gender, nickname, wins by submission, wins by knockout, striking accuracy, takedown accuracy
func (u Ufc) parseUfcFighterHtml(resBody io.ReadCloser) UfcSaveInfo {
	doc, err := goquery.NewDocumentFromReader(resBody)
	CheckErr(err)
	name := doc.Find(".hero-profile__name").Text()
	nickname := doc.Find(".hero-profile__nickname").Text()
	division := doc.Find(".hero-profile__division-title").Text()
	record := doc.Find(".hero-profile__division-body").Text()
	age := doc.Find(".field field--name-age field--type-integer field--label-hidden field__item").Text()
	ageInt, err := strconv.Atoi(age)
	CheckErr(err)
	var statusList []string
	var fighterStats []int
	var fighterPctgs []float32
	var etcInfo []string
	doc.Find(".hero-profile__tag").Each(func(i int, s *goquery.Selection) {
		status := s.Find("p").Text()
		statusList = append(statusList, status)
	})
	doc.Find(".hero-profile__stat-numb").Each(func(i int, s *goquery.Selection) {
		stat := s.Find("p").Text()
		statInt, err := strconv.Atoi(stat)
		CheckErr(err)
		fighterStats = append(fighterStats, statInt)
	})
	doc.Find(".e-chart-circle__percent").Each(func(i int, s *goquery.Selection) {
		pctg := s.Find("text").Text()
		pctgParsed := string(pctg[:len(pctg)-1])
		pctgInt, err := strconv.Atoi(pctgParsed)
		CheckErr(err)
		pctgFloat := float32(pctgInt) * 0.01
		fighterPctgs = append(fighterPctgs, pctgFloat)
	})
	doc.Find(".c-bio__text").Each(func(i int, s *goquery.Selection) {
		etc := s.Find("div").Text()
		etcInfo = append(etcInfo, etc)
	})
	hometown := etcInfo[1]
	hometownSplit := strings.Split(hometown, ",")
	country := hometownSplit[1]
	fightingStyle := etcInfo[3]
	height := etcInfo[5]
	heightInt, err := strconv.Atoi(height)
	CheckErr(err)
	heightFloat := float32(heightInt)
	weight := etcInfo[6]
	weightInt, err := strconv.Atoi(weight)
	CheckErr(err)
	weightFloat := float32(weightInt)
	reach := etcInfo[8]
	reachInt, err := strconv.Atoi(reach)
	CheckErr(err)
	reachFloat := float32(reachInt)

	fighter := UfcSaveInfo{
		Name:             name,
		Nickname:         nickname,
		Status:           statusList[1],
		Country:          country,
		Style:            fightingStyle,
		Record:           record,
		Division:         division,
		Age:              ageInt,
		KnockoutWins:     fighterStats[0],
		SubmissionWins:   fighterStats[1],
		Height:           heightFloat,
		Weight:           weightFloat,
		Reach:            reachFloat,
		StrikingAccuracy: fighterPctgs[0],
		TakedownAccuracy: fighterPctgs[1],
	}
	return fighter
}

func main() {
	u := Ufc{}
	u.ScrapeInfo()
}
