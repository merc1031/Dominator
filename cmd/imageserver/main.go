package main

import (
	"flag"
	"fmt"
	"github.com/Symantec/Dominator/imageserver/httpd"
	imageserverRpcd "github.com/Symantec/Dominator/imageserver/rpcd"
	"github.com/Symantec/Dominator/imageserver/scanner"
	"github.com/Symantec/Dominator/lib/constants"
	"github.com/Symantec/Dominator/objectserver/filesystem"
	objectserverRpcd "github.com/Symantec/Dominator/objectserver/rpcd"
	"os"
	"path"
)

var (
	debug   = flag.Bool("debug", false, "If true, show debugging output")
	portNum = flag.Uint("portNum", constants.ImageServerPortNumber,
		"Port number to allocate and listen on for HTTP/RPC")
	dataDir = flag.String("stateDir", "/var/lib/imageserver",
		"Name of image server data directory.")
)

func main() {
	flag.Parse()
	if os.Geteuid() == 0 {
		fmt.Println("Do not run the Image Server as root")
		os.Exit(1)
	}
	fi, err := os.Lstat(*dataDir)
	if err != nil {
		fmt.Printf("Cannot stat: %s\t%s\n", *dataDir, err)
		os.Exit(1)
	}
	if !fi.IsDir() {
		fmt.Printf("%s is not a directory\n", *dataDir)
		os.Exit(1)
	}
	objSrv, err := filesystem.NewObjectServer(path.Join(*dataDir, "objects"))
	if err != nil {
		fmt.Printf("Cannot create ObjectServer\t%s\n", err)
		os.Exit(1)
	}
	imdb, err := scanner.LoadImageDataBase(*dataDir, objSrv)
	if err != nil {
		fmt.Printf("Cannot load image database\t%s\n", err)
		os.Exit(1)
	}
	imageserverRpcd.Setup(imdb)
	objectserverRpcd.Setup(objSrv)
	err = httpd.StartServer(*portNum, imdb, false)
	if err != nil {
		fmt.Printf("Unable to create http server\t%s\n", err)
		os.Exit(1)
	}
}
