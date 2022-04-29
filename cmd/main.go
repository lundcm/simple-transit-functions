package main

import (
	"context"
	"log"
	"os"
	functions "simple-transit-functions"
	"simple-transit-functions/gtfs_download"
	"simple-transit-functions/sources_sync"
)

const (
	sourcesSync  = "sources_sync"
	gtfsDownload = "gtfs_download"
)

func main() {
	var commands = []string{
		sourcesSync,
		gtfsDownload,
	}
	if len(os.Args) < len(commands) {
		log.Fatalf("expected subcommands: \n%v\n", commands)
	}

	switch os.Args[1] {
	case sourcesSync:
		sources_sync.SourcesSync(context.Background(), nil)
	case gtfsDownload:
		if err := gtfs_download.GTFSDownload(context.Background(), functions.FirestoreEvent{}); err != nil {
			log.Fatal(err)
		}
	}
}

//func gtfsDownloadFirestoreEvent() (event functions.FirestoreEvent) {
//	event.OldValue = functions.FirestoreValue{
//		CreateTime: time.Time{},
//		Fields:     nil,
//		Name:       "",
//		UpdateTime: time.Time{},
//	}
//}
