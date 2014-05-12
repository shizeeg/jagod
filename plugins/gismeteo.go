package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"encoding/json"
	"encoding/xml"

	"code.google.com/p/go.net/html"
)

var (
	CityCode       int
	CityName       string
	Lang           string
	TLD            = "ru"
	pressure       = "мм.рт.с"
	wind           = "м/c"
	prefix         = "Погода в"
	months         = monthsru
	wdays          = wdaysru
	clouds         = cloudsru
	tods           = todsru
	cdirs          = cdirsru
	precipitations = precipitationsru
)

func init() {
	flag.IntVar(&CityCode, "code", 4364, "City code")
	// flag.StringVar(&CityName, "city", "", "City name")
	flag.StringVar(&Lang, "lang", "ru", "Language to output in")
	flag.Parse()
}

func main() {
	if Lang != "ru" {
		TLD = "com"
		pressure = "mmHg"
		wind = "m/s"
		prefix = "Weather in"
		months = monthsen
		wdays = wdaysen
		clouds = cloudsen
		tods = todsen
		cdirs = cdirsen
		precipitations = precipitationsen
	}
	var weather MMWeather
	if len(flag.Args()) > 0 && len(flag.Arg(0)) > 1 {
		if code, err := strconv.Atoi(flag.Arg(0)); err != nil {
			CityName = flag.Arg(0) // we got string
			city := City{}
			if err := city.GetCity(CityName); err != nil {
				fmt.Println(err)
			} else {
				CityCode = city.Code
				CityName = city.Name
			}
		} else {
			CityCode = code // we got number
		}
	}
	code, err := GetCode(CityCode)
	if err != nil {
		fmt.Println(err)
	}
	weather.Read(code)
	if CityName == "" {
		CityName = fmt.Sprint(weather.Report.Town.Sname)
	}
	fmt.Printf(prefix+": %s (%d) %d°N, %d°E\n",
		CityName,
		weather.Report.Town.Index,
		weather.Report.Town.Latitude,
		weather.Report.Town.Longitude,
	)
	for _, f := range weather.Report.Town.Forecasts {
		fmt.Printf("%s\n", f.String())
	}
}

type City struct {
	Code int
	Name string
}

func (c *City) GetCity(city string) error {
	city = url.QueryEscape(city)
	uri := fmt.Sprintf(search, TLD, city)
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	var data map[string]string
	for {
		if err := dec.Decode(&data); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		for k, v := range data {
			k = strings.Trim(k, "'")
			if code, err := strconv.Atoi(k); err == nil {
				c.Code = code
				c.Name = stripHTML(v)
				break
			}
		}
	}
	return nil
}

func GetCode(NewCode int) (code int, err error) {
	uri := fmt.Sprintf(forecast, TLD, NewCode)
	resp, err := http.Get(uri)
	if err != nil {
		return code, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return code, err
	}
	var c uint64
	var f func(*html.Node)
	var found bool
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			for _, div := range n.Attr {
				if div.Key == "class" && div.Val == "hslice" {
					found = true
				}
				if found && div.Key == "id" {
					c, err = strconv.ParseUint(div.Val, 10, 64)
					code = int(c)
					found = false
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	return
}

const (
	gismeteo = "http://informer.gismeteo.com/xml/%d.xml"
	forecast = "http://www.gismeteo.%s/city/daily/%d/"
	search   = "http://www.gismeteo.%s/ajax/city_search/?searchQuery=x%s"
	utf8     = "абвгдеёжзийклмнопрстуфхцчшщъыьэюяАБВГДЕЁЖЗИЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯ"
	//search   = "http://m.gismeteo.ru/citysearch/by_name/?gis_search=%s"
)

type CDir int
type Cloudiness int
type Month int
type Precipitation int
type TOD int
type WD int
type Winstr string

const (
	January = 1 + iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

const (
	Night TOD = 0 + iota
	Morning
	Day
	Evening
)

const (
	Sunday WD = 1 + iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

const (
	N CDir = 0 + iota
	NE
	E
	SE
	S
	SW
	W
	NW
)

const (
	Rain Precipitation = 4 + iota
	Downpour
	Snow
	Snowfall
	Storm
	NoData
	NoPrecipitation
)

const (
	Fair Cloudiness = 0 + iota
	Cloudy
	Clouds
	Overcast
)

// Month abbreviations (english)
var monthsen = [...]string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

var monthsru = [...]string{
	"Январь",
	"Февраль",
	"Март",
	"Апрель",
	"Май",
	"Июнь",
	"Июль",
	"Август",
	"Сенябрь",
	"Октябрь",
	"Ноябрь",
	"Декабрь",
}

// Times Of Day (englisth) for TOD enum
var todsen = [...]string{
	"Night",
	"Morning",
	"Day",
	"Evening",
}

// Times Of Day (russian) for TOD enum
var todsru = [...]string{
	"Ночь",
	"Утро",
	"День",
	"Вечер",
}

// Week days (english)
var wdaysen = [...]string{
	"Sunday",
	"Monday",
	"Tuesday",
	"Wednesday",
	"Thursday",
	"Friday",
	"Saturday",
}

// Week days (russian)
var wdaysru = [...]string{
	"Воскресенье",
	"Понедельник",
	"Вторник",
	"Среда",
	"Четверг",
	"Пятница",
	"Суббота",
}

// Cardinal (compass) Directions (english)
var cdirsen = [...]string{
	"N", "NE", "E", "SE", "S", "SW", "W", "NW",
}

// Cardinal (compass) Directions (russian)
var cdirsru = [...]string{
	"С", "CВ", "В", "ЮВ", "Ю", "ЮЗ", "З", "СЗ",
}

// Precipitations (english)
var precipitationsen = [...]string{
	"Rain",
	"Downpour",
	"Light snow",
	"Snowfall",
	"Storm",
	"No data",
	"No precipitation",
}

// Precipitations (russian)
var precipitationsru = [...]string{
	"Дождь",
	"Ливень",
	"Небольшой снег",
	"Снегопад",
	"Буря",
	"Нет данных",
	"Без осадков",
}

// Cloudiness (english)
var cloudsen = [...]string{
	"Fair",
	"Partly cloudy",
	"Cloudy",
	"Overcast",
}

// Cloudiness (russian)
var cloudsru = [...]string{
	"Ясно",
	"Облачно с прояснениями",
	"Облачно",
	"Тучи",
}

func (m Month) String() string         { return months[m-1] }
func (t TOD) String() string           { return tods[t] }
func (w WD) String() string            { return wdays[w-1] }
func (c CDir) String() string          { return cdirs[c] }
func (p Precipitation) String() string { return precipitations[p-4] }
func (c Cloudiness) String() string    { return clouds[c] }
func (s Winstr) String() string {
	out, _ := url.QueryUnescape(string(s))
	return ToUTF8(out)
}

type Forecast struct {
	XMLName   xml.Name `xml:"FORECAST"`
	Day       int      `xml:"day,attr"`
	Month     Month    `xml:"month,attr"`
	Year      int      `xml:"year,attr"`
	Hour      int      `xml:"hour,attr"`
	TOD       TOD      `xml:"tod,attr"`
	Predict   int      `xml:"predict,attr"`
	Weekday   WD       `xml:"weekday,attr"`
	Phenomena struct {
		XMLName       xml.Name      `xml:"PHENOMENA"`
		Cloudiness    Cloudiness    `xml:"cloudiness,attr"`
		Precipitation Precipitation `xml:"precipitation,attr"`
		RPower        int           `xml:"rpower,attr"`
		SPower        int           `xml:"spower,attr"`
	}
	Pressure struct {
		XMLName xml.Name `xml:"PRESSURE"`
		Max     int      `xml:"max,attr"`
		Min     int      `xml:"min,attr"`
	}
	Wind struct {
		XMLName xml.Name `xml:"WIND"`
		Dir     CDir     `xml:"direction,attr"`
		Max     int      `xml:"max,attr"`
		Min     int      `xml:"min,attr"`
	}
	RelWet struct {
		XMLName xml.Name `xml:"RELWET"`
		Max     int      `xml:"max,attr"`
		Min     int      `xml:"min,attr"`
	}
	Heat struct {
		XMLName xml.Name `xml:"HEAT"`
		Max     int      `xml:"max,attr"`
		Min     int      `xml:"min,attr"`
	}
	Temp struct {
		XMLName xml.Name `xml:"TEMPERATURE"`
		Max     int      `xml:"max,attr"`
		Min     int      `xml:"min,attr"`
	}
}

// String is pretty-printer for Forecast struct
func (f *Forecast) String() string {
	mid := func(max, min int) int {
		return (min + max) / 2
	}
	space := strings.Repeat(" ", 8-strings.Count(f.TOD.String(), ""))
	out := fmt.Sprintf("%s%s\t%2d %.3s: %+.2d..%+.2d °C, %-3d %s %d%% %-2s %d-%d %s %s, %s",
		f.TOD, space, f.Day, f.Month, f.Temp.Min, f.Temp.Max,
		mid(f.Pressure.Max, f.Pressure.Min), pressure,
		mid(f.RelWet.Max, f.RelWet.Min),
		f.Wind.Dir, f.Wind.Min, f.Wind.Max, wind,
		f.Phenomena.Cloudiness, f.Phenomena.Precipitation,
	)
	out = strings.Replace(out, "+0", " +", -1)
	return strings.Replace(out, "-0", " -", -1)
}

type MMWeather struct {
	XMLName xml.Name `xml:"MMWEATHER"`
	Report  struct {
		XMLName xml.Name `xml:"REPORT"`
		Type    string   `xml:"type,attr"`
		Town    struct {
			XMLName   xml.Name   `xml:"TOWN"`
			Index     int        `xml:"index,attr"`
			Sname     Winstr     `xml:"sname,attr"`
			Latitude  int        `xml:"latitude,attr"`
			Longitude int        `xml:"longitude,attr"`
			Forecasts []Forecast `xml:"FORECAST"`
		}
	}
}

// Read fetches xml and fills MMWeather struct with data
func (w *MMWeather) Read(code int) error {
	uri := fmt.Sprintf(gismeteo, code)
	resp, err := http.Get(uri)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	dec := xml.NewDecoder(resp.Body)
	if err := dec.Decode(&w); err != nil {
		return err
	}
	return nil
}

// ToUTF8 converts cp1251 string to utf8 string
// it's incomplete and should be used only with gismeteo's bad coded city info
func ToUTF8(str string) string {
	win2utf8 := map[byte]rune{
		'\xE0': 'а', '\xE1': 'б', '\xE2': 'в', '\xE3': 'г',
		'\xE4': 'д', '\xE5': 'е', '\xB8': 'ё', '\xE6': 'ж',
		'\xE7': 'з', '\xE8': 'и', '\xE9': 'й', '\xEA': 'к',
		'\xEB': 'л', '\xEC': 'м', '\xED': 'н', '\xEE': 'о',
		'\xEF': 'п', '\xF0': 'р', '\xF1': 'с', '\xF2': 'т',
		'\xF3': 'у', '\xF4': 'ф', '\xF5': 'х', '\xF6': 'ц',
		'\xF7': 'ч', '\xF8': 'ш', '\xF9': 'щ', '\xFA': 'ъ',
		'\xFB': 'ы', '\xFC': 'ъ', '\xFD': 'э', '\xFE': 'ю',
		'\xFF': 'я',
		'\xC0': 'А', '\xC1': 'Б', '\xC2': 'В', '\xC3': 'Г',
		'\xC4': 'Д', '\xC5': 'Е', '\xA8': 'Ё', '\xC6': 'Ж',
		'\xC7': 'З', '\xC8': 'И', '\xC9': 'Й', '\xCA': 'К',
		'\xCB': 'Л', '\xCC': 'М', '\xCD': 'Н', '\xCE': 'О',
		'\xCF': 'П', '\xD0': 'Р', '\xD1': 'С', '\xD2': 'Т',
		'\xD3': 'У', '\xD4': 'Ф', '\xD5': 'Х', '\xD6': 'Ц',
		'\xD7': 'Ч', '\xD8': 'Ш', '\xD9': 'Щ', '\xDA': 'Ъ',
		'\xDB': 'Ы', '\xDC': 'Ь', '\xDD': 'Э', '\xDE': 'Ю',
		'\xDF': 'Я',
		'\x20': ' ', '-': '-', '(': '(', ')': ')', '+': '+',
	}
	var out string
	for _, ch := range []byte(str) {
		if r, ok := win2utf8[ch]; ok {
			out += fmt.Sprintf("%c", r)
		}
	}
	return out
}

func stripHTML(str string) string {
	z := html.NewTokenizer(strings.NewReader(str))
	var out []byte
	for {
		tt := z.Next()
		switch tt {
		case html.TextToken:
			out = append(out, z.Text()...)
		case html.ErrorToken:
			return fmt.Sprintf("%s", out)
		}
	}
	return fmt.Sprintf("%s", out)
}
