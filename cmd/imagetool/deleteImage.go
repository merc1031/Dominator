package main

import (
	"errors"
	"fmt"
	"github.com/Symantec/Dominator/imageserver/client"
	"github.com/Symantec/Dominator/lib/objectclient"
	"github.com/Symantec/Dominator/lib/srpc"
	"github.com/Symantec/Dominator/proto/imageserver"
	"net/rpc"
	"os"
)

func deleteImageSubcommand(imageClient *rpc.Client, imageSClient *srpc.Client,
	objectClient *objectclient.ObjectClient, args []string) {
	if err := deleteImage(imageSClient, args[0]); err != nil {
		fmt.Fprintf(os.Stderr, "Error deleting image\t%s\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func deleteImage(imageSClient *srpc.Client, name string) error {
	var request imageserver.DeleteImageRequest
	request.ImageName = name
	var reply imageserver.DeleteImageResponse
	err := client.CallDeleteImage(imageSClient, request, &reply)
	if err != nil {
		return err
	}
	if reply.Success {
		return nil
	}
	return errors.New(reply.ErrorString)
}
