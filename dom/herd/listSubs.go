package herd

import (
	"bufio"
	"fmt"
	"net/http"
	net_url "net/url"

	"github.com/Cloud-Foundations/Dominator/lib/json"
	"github.com/Cloud-Foundations/Dominator/lib/url"
)

func (herd *Herd) listReachableSubsHandler(w http.ResponseWriter,
	req *http.Request) {
	writer := bufio.NewWriter(w)
	defer writer.Flush()
	parsedQuery := url.ParseQuery(req.URL)
	selector, err := herd.getReachableSelector(parsedQuery)
	if err != nil {
		fmt.Fprintln(writer, err)
		return
	}
	subs := herd.getSelectedSubs(selector)
	switch parsedQuery.OutputType() {
	case url.OutputTypeText:
	case url.OutputTypeHtml:
		for _, sub := range subs {
			fmt.Fprintln(writer, sub.mdb.Hostname)
		}
	case url.OutputTypeJson:
		subNames := make([]string, 0, len(subs))
		for _, sub := range subs {
			subNames = append(subNames, sub.mdb.Hostname)
		}
		json.WriteWithIndent(writer, "  ", subNames)
		fmt.Fprintln(writer)
	}
}

func (herd *Herd) listSubsHandler(w http.ResponseWriter, req *http.Request) {
	writer := bufio.NewWriter(w)
	defer writer.Flush()
	parsedQuery := url.ParseQuery(req.URL)
	var statusToMatch string
	if uesc, e := net_url.QueryUnescape(parsedQuery.Table["status"]); e == nil {
		statusToMatch = uesc
	}
	herd.RLock()
	subNames := make([]string, 0, len(herd.subsByIndex))
	for _, sub := range herd.subsByIndex {
		if statusToMatch != "" && sub.status.String() != statusToMatch {
			continue
		}
		subNames = append(subNames, sub.mdb.Hostname)
	}
	herd.RUnlock()
	switch parsedQuery.OutputType() {
	case url.OutputTypeText:
	case url.OutputTypeHtml:
		for _, name := range subNames {
			fmt.Fprintln(writer, name)
		}
	case url.OutputTypeJson:
		json.WriteWithIndent(writer, "  ", subNames)
		fmt.Fprintln(writer)
	}
}
