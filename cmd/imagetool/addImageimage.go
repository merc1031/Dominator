package main

import (
	"errors"
	"fmt"
	"github.com/Symantec/Dominator/lib/image"
	objectclient "github.com/Symantec/Dominator/lib/objectserver/client"
	"github.com/Symantec/Dominator/lib/srpc"
	"net/rpc"
	"os"
)

func addImageimageSubcommand(args []string) {
	imageClient, imageSClient, objectClient := getClients()
	err := addImageimage(imageClient, imageSClient, objectClient, args[0],
		args[1], args[2], args[3])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error adding image: \"%s\"\t%s\n", args[0], err)
		os.Exit(1)
	}
	os.Exit(0)
}

func addImageimage(imageClient *rpc.Client, imageSClient *srpc.Client,
	objectClient *objectclient.ObjectClient,
	name, oldImageName, filterFilename, triggersFilename string) error {
	imageExists, err := checkImage(imageClient, name)
	if err != nil {
		return errors.New("error checking for image existance: " + err.Error())
	}
	if imageExists {
		return errors.New("image exists")
	}
	newImage := new(image.Image)
	if err := loadImageFiles(newImage, objectClient, filterFilename,
		triggersFilename); err != nil {
		return err
	}
	fs, err := getFsOfImage(imageSClient, oldImageName)
	if err != nil {
		return err
	}
	if err := spliceComputedFiles(fs); err != nil {
		return err
	}
	if fs, err = applyDeleteFilter(fs); err != nil {
		return err
	}
	fs = fs.Filter(newImage.Filter)
	newImage.FileSystem = fs
	return addImage(imageSClient, name, newImage)
}