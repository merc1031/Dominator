package httpd

import (
	"bufio"
	"fmt"
	"net/http"
	"strings"

	"github.com/Cloud-Foundations/Dominator/lib/filesystem"
	"github.com/Cloud-Foundations/Dominator/lib/html"
)

func (s state) listComputedInodesHandler(w http.ResponseWriter,
	req *http.Request) {
	writer := bufio.NewWriter(w)
	defer writer.Flush()
	imageName := req.URL.RawQuery
	fmt.Fprintf(writer, "<title>image %s computed inodes</title>\n", imageName)
	fmt.Fprintln(writer, `<style>
                          table, th, td {
                          border-collapse: collapse;
                          }
                          </style>`)
	fmt.Fprintln(writer, "<body>")
	fmt.Fprintln(writer, "<h3>")
	if image := s.imageDataBase.GetImage(imageName); image == nil {
		fmt.Fprintf(writer, "Image: %s UNKNOWN!\n", imageName)
	} else {
		fmt.Fprintf(writer, "Computed files for image: %s\n", imageName)
		fmt.Fprintln(writer, "</h3>")
		fmt.Fprintln(writer, `<table border="1" style="width:100%">`)
		tw, _ := html.NewTableWriter(writer, true, "Filename", "Data Source")
		computedFiles, _ := s.imageDataBase.GetImageComputedFiles(imageName)
		listComputedInodes(tw, computedFiles)
		tw.Close()
	}
	fmt.Fprintln(writer, "</body>")
}

func listComputedInodes(tw *html.TableWriter,
	computedFiles []filesystem.ComputedFile) {
	for _, computedFile := range computedFiles {
		var source string
		if strings.HasPrefix(computedFile.Source, "localhost:") {
			source = computedFile.Source
		} else {
			source = fmt.Sprintf("<a href=\"http://%s\">%s</a>",
				computedFile.Source, computedFile.Source)
		}
		tw.WriteRow("", "", computedFile.Filename, source)
	}
}
