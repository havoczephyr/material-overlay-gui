package api

import (
	"encoding/json"
	"strconv"
)

// SMWResponse is the top-level response from action=ask.
type SMWResponse struct {
	Query SMWQuery `json:"query"`
}

type SMWQuery struct {
	Results SMWResults `json:"results"`
}

// SMWResults handles the dual-type: object when results exist, empty array when not.
type SMWResults map[string]SMWResultEntry

func (r *SMWResults) UnmarshalJSON(data []byte) error {
	var arr []json.RawMessage
	if err := json.Unmarshal(data, &arr); err == nil && len(arr) == 0 {
		*r = make(SMWResults)
		return nil
	}

	var m map[string]SMWResultEntry
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	*r = m
	return nil
}

type SMWResultEntry struct {
	Printouts map[string]json.RawMessage `json:"printouts"`
	Fulltext  string                     `json:"fulltext"`
	Fullurl   string                     `json:"fullurl"`
	Namespace int                        `json:"namespace"`
	Exists    string                     `json:"exists"`
}

// SMWPageValue represents a _wpg type printout value (page reference).
type SMWPageValue struct {
	Fulltext     string `json:"fulltext"`
	Fullurl      string `json:"fullurl"`
	Namespace    int    `json:"namespace"`
	Exists       string `json:"exists"`
	Displaytitle string `json:"displaytitle"`
}

// SearchResponse is the top-level response from action=query&list=search.
type SearchResponse struct {
	Query SearchQuery `json:"query"`
}

type SearchQuery struct {
	SearchInfo SearchInfo     `json:"searchinfo"`
	Search     []SearchResult `json:"search"`
}

type SearchInfo struct {
	TotalHits int `json:"totalhits"`
}

type SearchResult struct {
	NS        int    `json:"ns"`
	Title     string `json:"title"`
	PageID    int    `json:"pageid"`
	Snippet   string `json:"snippet"`
	Timestamp string `json:"timestamp"`
}

// ExtractTextValues unmarshals a printout as a text-type ([]string).
// Falls back to numeric arrays (e.g., [0] from "Genesys cost").
func ExtractTextValues(raw json.RawMessage) []string {
	var texts []string
	if err := json.Unmarshal(raw, &texts); err == nil {
		return texts
	}

	var nums []float64
	if err := json.Unmarshal(raw, &nums); err == nil {
		out := make([]string, len(nums))
		for i, n := range nums {
			if n == float64(int64(n)) {
				out[i] = strconv.FormatInt(int64(n), 10)
			} else {
				out[i] = strconv.FormatFloat(n, 'f', -1, 64)
			}
		}
		return out
	}

	return nil
}

// ExtractPageValues unmarshals a printout as a page-type ([]SMWPageValue).
func ExtractPageValues(raw json.RawMessage) []SMWPageValue {
	var pages []SMWPageValue
	if err := json.Unmarshal(raw, &pages); err == nil {
		return pages
	}
	return nil
}
