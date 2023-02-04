package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/akyoto/cache"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getClient2() {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// authenticate and get configuration
	config, err := google.JWTConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Println(err)
		return
	}

	ctx := context.Background()

	// create client with config and context
	client := config.Client(ctx)

	// create new service using client
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("srv: %v", srv)
}

var (
	srv     *sheets.Service
	zaCache *cache.Cache
)

func getData() string {
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
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
		return "No data found"
	}
	var output string
	for _, row := range resp.Values {
		// Print columns A and E, which correspond to indices 0 and 4.
		//fmt.Printf("%s, %s\n", row[0], row[4])
		fmt.Printf("%s, %+v\n", row[0], row)
		output += fmt.Sprintf("%s, %+v\n", row[0], row)
	}

	return output
}

func fooHandler(w http.ResponseWriter, _ *http.Request) {

	// Read from the cache
	data, found := zaCache.Get("data")
	if !found {
		data = getData()
		//zaCache.Set("data", data, 1*time.Minute)
	}
	w.Write([]byte(data.(string)))
}

func main() {

	zaCache = cache.New(5 * time.Minute)

	ctx := context.Background()

	//getClient2()

	b, err := os.ReadFile("credentials.json")
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
	//client := getClient(config)
	// create client with config and context
	client := config.Client(ctx)

	srv, err = sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	http.HandleFunc("/foo", fooHandler)

	http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

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
