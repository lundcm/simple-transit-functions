package sources_sync

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"net/http"
	"os"
	"simple-transit-functions"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"github.com/jszwec/csvutil"
)

// GOOGLE_CLOUD_PROJECT is automatically set by the Cloud Functions runtime.
var projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")

// client is a Firestore client, reused between function invocations.
var client *firestore.Client

func init() {
	// Use the application default credentials.
	conf := &firebase.Config{ProjectID: projectID}

	// Use context.Background() because the app/client should persist across invocations.
	ctx := context.Background()

	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v", err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}
}

func readCSVFromUrl(url string) ([]simple_transit_functions.GTFSSource, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	var sources []simple_transit_functions.GTFSSource

	dec, err := csvutil.NewDecoder(reader)
	if err != nil {
		return nil, err
	}

	for {
		var s simple_transit_functions.GTFSSource
		if err = dec.Decode(&s); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		sources = append(sources, s)
	}

	return sources, nil
}

func downloadSources(ctx context.Context) error {
	url := "https://storage.googleapis.com/storage/v1/b/mdb-csv/o/sources.csv?alt=media"
	sources, err := readCSVFromUrl(url)
	if err != nil {
		return err
	}

	for _, source := range sources {
		if _, err = client.Collection("transit").Doc(source.MdbSourceId).Set(ctx, source); err != nil {
			return err
		}

		log.Println(source)
	}

	return nil
}

func SourcesSync(ctx context.Context, _ interface{}) {
	if projectID == "" {
		log.Fatalln("GOOGLE_CLOUD_PROJECT environment variable must be set.")
	}
	err := downloadSources(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
