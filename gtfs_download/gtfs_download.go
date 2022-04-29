package gtfs_download

import (
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/api/iterator"
	"log"
	"os"
	"simple-transit-functions"
	"time"
)

var projectId = os.Getenv("GOOGLE_CLOUD_PROJECT")

var client *bigquery.Client
var cStorage *storage.Client
var psClient *pubsub.Client

func init() {
	ctx := context.Background()

	var err error
	client, err = bigquery.NewClient(ctx, projectId)
	if err != nil {
		log.Fatalf("bigquery.NewClient: %v", err)
	}
	cStorage, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("storage.NewClient: %v", err)
	}
	psClient, err = pubsub.NewClient(ctx, projectId)
	if err != nil {
		log.Fatal(err)
	}

	defer func(client *bigquery.Client) {
		err := client.Close()
		if err != nil {
			log.Fatalf("client.Close: %v", err)
		}
	}(client)
	defer func(cStorage *storage.Client) {
		err := cStorage.Close()
		if err != nil {
			log.Fatalf("storage.Close: %v", err)
		}
	}(cStorage)
	defer func(psClient *pubsub.Client) {
		err := psClient.Close()
		if err != nil {
			log.Fatalf("pubsub.Close: %v", err)
		}
	}(psClient)
}

func query(ctx context.Context) *bigquery.RowIterator {
	return client.Dataset("source").Table("mobility-database-catalog").Read(ctx)
}

func sendMessage(ctx context.Context, msg MobilityDatabaseCatalog) (err error) {
	topic := psClient.Topic("transit-sources-download")
	msgStr, err := json.Marshal(msg)
	if err != nil {
		return
	}
	if _, err = topic.Publish(ctx, &pubsub.Message{Data: msgStr}).Get(ctx); err != nil {
		return fmt.Errorf("could not publish message: %v", err)
	}
	return
}

func GTFSDownload(ctx context.Context, e simple_transit_functions.FirestoreEvent) (err error) {
	if projectId == "" {
		log.Fatalln("GOOGLE_CLOUD_PROJECT environment variable must be set.")
	}

	rows := query(ctx)
	for {
		var row MobilityDatabaseCatalog
		err = rows.Next(&row)
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error iterating through results: %v", err)
		}
		if err = sendMessage(ctx, row); err != nil {
			return
		}
		return
	}
	return
}

type MobilityDatabaseCatalog struct {
	MdbSourceId                         int                 `bigquery:"mdb_source_id"`
	DataType                            string              `bigquery:"data_type"`
	LocationCountryCode                 bigquery.NullString `bigquery:"location_country_code,nullable"`
	LocationSubdivisionName             bigquery.NullString `bigquery:"location_subdivision_name,nullable"`
	LocationMunicipality                bigquery.NullString `bigquery:"location_municipality"`
	Provider                            string              `bigquery:"provider"`
	Name                                bigquery.NullString `bigquery:"name"`
	UrlsDirectDownload                  bigquery.NullString `bigquery:"urls_direct_download"`
	UrlsLatest                          bigquery.NullString `bigquery:"urls_latest"`
	UrlsLicense                         bigquery.NullString `bigquery:"urls_license"`
	LocationBoundingBoxMinimumLatitude  bigquery.NullString `bigquery:"location_bounding_box_minimum_latitude"`
	LocationBoundingBoxMaximumLatitude  bigquery.NullString `bigquery:"location_bounding_box_maximum_latitude"`
	LocationBoundingBoxMinimumLongitude bigquery.NullString `bigquery:"location_bounding_box_minimum_longitude"`
	LocationBoundingBoxMaximumLongitude bigquery.NullString `bigquery:"location_bounding_box_maximum_longitude"`
	LocationBoundingBoxExtractedOn      time.Time           `bigquery:"location_bounding_box_extracted_on"`
}
