package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/akyoto/cache"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

var (
	srv         *sheets.Service
	zaCache     *cache.Cache
	lastUpdated time.Time

	//go:embed var
	res embed.FS
)

func getData() (string, error) {
	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	//spreadsheetId := "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"
	//spreadsheetId := "15wo2Ppki4ybpHGla2PwwkBs_K4UbyhixVzuYdNa4AtY"
	//readRange := "Sheet1!A2:E"

	// https://docs.google.com/spreadsheets/d/1hht9Vne2h3icOGX0hGscNklclJEZY7gYY-18n3KnGJo/edit#gid=0
	spreadsheetId := "1hht9Vne2h3icOGX0hGscNklclJEZY7gYY-18n3KnGJo"
	readRange := "status!A1:E"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		msg := fmt.Sprintf("Unable to retrieve data from sheet: %v", err)
		//log.Println(msg)
		return "", errors.New(msg)
	}

	if len(resp.Values) == 0 {
		//fmt.Println("No data found.")
		return "No data found", nil
	}
	var output string
	for _, row := range resp.Values {
		// Print columns A and E, which correspond to indices 0 and 4.
		//fmt.Printf("%s, %s\n", row[0], row[4])
		fmt.Printf("%s, %+v\n", row[0], row)
		output += fmt.Sprintf("%s, %+v\n", row[0], row)
	}

	return output, nil
}

func bailout(w http.ResponseWriter, msg string, status int) {
	log.Println(msg)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(msg))
}

func fooHandler(w http.ResponseWriter, _ *http.Request) {

	var err error
	var loc *time.Location
	// Read from the cache
	data, found := zaCache.Get("data")
	if !found {
		if data, err = getData(); err != nil {
			bailout(w, err.Error(), http.StatusInternalServerError)
			return
		}
		zaCache.Set("data", data, 1*time.Minute)
		if loc, err = time.LoadLocation("America/New_York"); err != nil {
			msg := fmt.Sprintf("error loading location: %s\n", err)
			bailout(w, msg, http.StatusInternalServerError)
			return
		}
		lastUpdated = time.Now().In(loc)
	}
	w.Write([]byte(data.(string)))
	w.Write([]byte("\n"))
	w.Write([]byte(fmt.Sprintf("\nLast updated: %s\n", lastUpdated.Format(time.RFC3339))))
}

func main() {

	zaCache = cache.New(20 * time.Minute)

	ctx := context.Background()

	b, err := res.ReadFile("var/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	// authenticate and get configuration
	config, err := google.JWTConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	//config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	//config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	// create client with config and context
	client := config.Client(ctx)

	srv, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	http.HandleFunc("/foo", fooHandler)

	// http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	// })

	log.Printf("running app on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

	return

	// https://docs.google.com/spreadsheets/d/<SPREADSHEETID>/edit#gid=<SHEETID>

	// sheetId := <SHEETID>
	// spreadsheetId := <SPREADSHEETID>

	// // Convert sheet ID to sheet name.
	// response1, err := srv.Spreadsheets.Get(spreadsheetId).Fields("sheets(properties(sheetId,title))").Do()
	// if err != nil || response1.HTTPStatusCode != 200 {
	// 	log.Error(err)
	// 	return
	// }

	// sheetName := ""
	// for _, v := range response1.Sheets {
	// 	prop := v.Properties
	// 	if prop.SheetId == int64(sheetId) {
	// 		sheetName = prop.Title
	// 		break
	// 	}
	// }

	// valueInputOption := "USER_ENTERED"
	// insertDataOption := "INSERT_ROWS"

	// //Append value to the sheet.
	// row := &sheets.ValueRange{
	// 	Values: [][]any{
	// 		{"C", "o", "5"},
	// 	},
	// }

	// resp2, err := srv.Spreadsheets.Values.Append(spreadsheetId, "status", row).ValueInputOption(valueInputOption).InsertDataOption(insertDataOption).Context(ctx).Do()
	// if err != nil || resp2.HTTPStatusCode != 200 {
	// 	fmt.Printf("err = %+v\n", err)
	// 	return
	// }
	// fmt.Printf("resp = %+v\n", resp2)
}
