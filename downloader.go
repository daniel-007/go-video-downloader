/**
 * Go Video Downloader
 *
 *    Copyright 2017 Tenta, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * For any questions, please contact developer@tenta.io
 *
 * downloader.go: Downloader interface
 */

package downloader

import (
	"fmt"
	"net/http"

	"github.com/tenta-browser/go-pcre-matcher"
	"github.com/tenta-browser/go-video-downloader/extractor"
	"github.com/tenta-browser/go-video-downloader/runtime"
)

// GenDate is date when this downloader was generated
const GenDate = "20171122-0113"

var (
	masterRegexp     string
	masterRegexpComp matcher.Regexp
)

// CheckResult holds all the info obtained by verifying if an URL contains
// a downloadable video (extractorKey for now)
type CheckResult struct {
	url          string
	extractorKey string
}

// Connector holds settings/callbacks related to establishing connections,
// downloading content, etc.
type Connector struct {
	Client    *http.Client
	UserAgent string
	Cookie    string
}

// VideoData holds the final result of the extraction
type VideoData struct {
	URL      string
	Title    string
	Filename string
	AgeLimit int
}

// Check verifies if url contains a downloadable video; returns nil if not
func Check(url string) *CheckResult {
	// if the master regexp isn't compiled, compile it
	if masterRegexpComp == nil {
		masterRegexpComp = runtime.ReMustCompile(masterRegexp, 0)
	}
	// check if something matches the url
	matcher := masterRegexpComp.Search(url)
	if matcher == nil {
		return nil
	}
	// something matched the url, now we have to find it :|
	for extractorKey := range extractor.GetFactories() {
		if matcher.GroupPresentByName("extr_" + extractorKey) {
			return &CheckResult{
				url:          url,
				extractorKey: extractorKey,
			}
		}
	}
	// master regexp and the extractors are not in sync
	panic("Url was matched but no extractor found: " + url)
}

// Extract extracts video info based on the results of a successful Check()
func Extract(checkResult *CheckResult, connector *Connector) (*VideoData, error) {
	return extractInternal(checkResult.url, checkResult.extractorKey, connector, nil)
}

func extractInternal(url string, extractorKey string, connector *Connector,
	referrer runtime.ExtractorResult) (*VideoData, error) {

	// create extractor corresponding to extractorKey
	factory := extractor.GetFactories()[extractorKey]
	if factory == nil {
		return nil, fmt.Errorf("Unsupported extractor: %s", extractorKey)
	}
	extractor := factory()

	// create fresh context and inject it into the extractor
	ctx := newContextFromConnector(connector)
	ctx.ExtractorKey = extractorKey
	extractor.SetContext(ctx)

	// run extractor
	res, err := extractor.Extract(url)
	if err != nil {
		return nil, err
	}

	switch res := res.(type) {
	case *runtime.VideoResult:
		return &VideoData{
			URL:      res.URL,
			Title:    res.Title,
			Filename: res.Filename,
			AgeLimit: res.AgeLimit,
		}, nil
	case *runtime.URLResult:
		if res.ExtractorKey == "" {
			// no extractorKey was provided by the extractor,
			// have to do a check based on the URL
			checkResult := Check(res.URL)
			if checkResult == nil {
				return nil, fmt.Errorf("Check failed for: %s", res.URL)
			}
			res.ExtractorKey = checkResult.extractorKey
		}
		return extractInternal(res.URL, res.ExtractorKey, connector, res)
	default:
		panic(fmt.Sprintf("Unhandled ExtractorType: %T", res))
	}
}

func newContextFromConnector(connector *Connector) *runtime.Context {
	headers := make(map[string]string)
	if connector.UserAgent != "" {
		headers["User-Agent"] = connector.UserAgent
	}
	if connector.Cookie != "" {
		headers["Cookie"] = connector.Cookie
	}
	var client *http.Client
	if connector.Client != nil {
		client = connector.Client
	} else {
		client = &http.Client{}
	}
	return &runtime.Context{
		Client:  client,
		Headers: headers,
	}
}
