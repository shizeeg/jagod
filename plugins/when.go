/*****************************************************************************
Copyright (c) 2017, sh!zeeg <shizeeque@gmail.com>
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this
  list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice,
  this list of conditions and the following disclaimer in the documentation
  and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*******************************************************************************/

package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"
)

const (
	hour  = time.Hour
	day   = hour * 24
	week  = day * 7
	month = day * 31
	year  = month * 12
)

// ApproxSign is true when date is approximate
type ApproxSign bool

// Approximation flag stringer
// appx := Approx(true)
// fmt.Printf("%v%d days", approx, 30)
// prints: ~30 days
func (a ApproxSign) String() string {
	if a == true {
		// U+2248 ALMOST EQUAL TO ( ≈ )
		return fmt.Sprintf("%c", 0x2248)
	}
	return ""
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Fprintf(os.Stdout, "%s YYYY-MM-DD\n", os.Args[0])
		os.Exit(1)
	}

	date, err := time.Parse("2006-01-02", os.Args[1])
	// if used as a jagod plugin search in other args too,
	// stop after the 1st match
	// TODO(shizeeg) add localization support
	for _, d := range os.Args[1:] {
		date, err = time.Parse("2006-01-02", d)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stdout, "date must be in YYYY-MM-DD format! => %v\n", err)
		os.Exit(2)
	}

	dateMaps := map[string]int{
		"year":  date.Year() - time.Now().Year(),
		"month": int(date.Month() - time.Now().Month()),
		"day":   date.Day() - time.Now().Day(),
		"hour":  0,
	}

	date = date.AddDate(-dateMaps["year"], 0, 0)
	dur := time.Until(date)

	var approx ApproxSign
	var ago string
	if dur < 0 {
		dur = -dur
		ago = " ago"
	}
	var hours, days, months, years int
	for dur > time.Hour {
		dur -= hour
		hours++
		if hours >= 24 {
			days++
			hours = 0
		}
		if days >= 30 {
			months++
			days = 0
		}
		if months >= 12 {
			years++
			months = 0
		}
	}
	dateMaps["month"] = months
	dateMaps["day"] = days
	dateMaps["hour"] = hours

	daycount := dateMaps["year"]*365 + int(math.Floor(float64(time.Until(date)/day)))
	if daycount < 0 {
		daycount = -daycount
	}
	if hour > 0 {
		approx = true
		daycount++
	}
	if days > 0 {
		fmt.Printf("%s (%v%s%s) \n",
			formatTime(dateMaps),
			approx, pluralize("day", int(math.Abs(float64(daycount)))),
			ago,
		)
	} else {
		fmt.Printf("%s%s\n", formatTime(dateMaps), ago)
	}
}
func pluralize(noun string, amount int) string {
	if amount < 0 {
		amount = -amount
	}
	if amount != 1 {
		noun += "s"
	}
	return fmt.Sprintf("%d %s", amount, noun)
}

// formatTime returns y years, m months, d days string
func formatTime(maps map[string]int) string {
	var out []string
	// make sure order is correct
	for _, str := range []string{"year", "month", "day", "hour"} {
		if maps[str] != 0 {
			out = append(out, pluralize(str, maps[str]))
		}
	}
	return strings.Join(out, ", ")
}
