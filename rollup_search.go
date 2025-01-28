// Copyright 2012-present Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/olivere/elastic/v7/uritemplates"
)

// Search for documents in Elasticsearch.
type RollupSearchService struct {
	client *Client

	pretty     *bool       // pretty format the returned JSON response
	human      *bool       // return human readable values for statistics
	errorTrace *bool       // include the stack trace of returned errors
	filterPath []string    // list of filters used to reduce the response
	headers    http.Header // custom request-level HTTP headers

	searchSource                    *SearchSource // q
	source                          interface{}
	searchType                      string // search_type
	index                           []string
	typ                             []string
	routing                         string // routing
	preference                      string // preference
	requestCache                    *bool  // request_cache
	ignoreUnavailable               *bool  // ignore_unavailable
	ignoreThrottled                 *bool  // ignore_throttled
	allowNoIndices                  *bool  // allow_no_indices
	expandWildcards                 string // expand_wildcards
	lenient                         *bool  // lenient
	maxResponseSize                 int64
	allowPartialRollupSearchResults *bool // allow_partial_search_results
	typedKeys                       *bool // typed_keys
	seqNoPrimaryTerm                *bool // seq_no_primary_term
	batchedReduceSize               *int  // batched_reduce_size
	maxConcurrentShardRequests      *int  // max_concurrent_shard_requests
	preFilterShardSize              *int  // pre_filter_shard_size
	restTotalHitsAsInt              *bool // rest_total_hits_as_int

	ccsMinimizeRoundtrips *bool // ccs_minimize_roundtrips

}

// NewRollupSearchService creates a new service for searching in Elasticsearch.
func NewRollupSearchService(client *Client) *RollupSearchService {
	builder := &RollupSearchService{
		client:       client,
		searchSource: NewSearchSource(),
	}
	return builder
}

// Pretty tells Elasticsearch whether to return a formatted JSON response.
func (s *RollupSearchService) Pretty(pretty bool) *RollupSearchService {
	s.pretty = &pretty
	return s
}

// Human specifies whether human readable values should be returned in
// the JSON response, e.g. "7.5mb".
func (s *RollupSearchService) Human(human bool) *RollupSearchService {
	s.human = &human
	return s
}

// ErrorTrace specifies whether to include the stack trace of returned errors.
func (s *RollupSearchService) ErrorTrace(errorTrace bool) *RollupSearchService {
	s.errorTrace = &errorTrace
	return s
}

// FilterPath specifies a list of filters used to reduce the response.
func (s *RollupSearchService) FilterPath(filterPath ...string) *RollupSearchService {
	s.filterPath = filterPath
	return s
}

// Header adds a header to the request.
func (s *RollupSearchService) Header(name string, value string) *RollupSearchService {
	if s.headers == nil {
		s.headers = http.Header{}
	}
	s.headers.Add(name, value)
	return s
}

// Headers specifies the headers of the request.
func (s *RollupSearchService) Headers(headers http.Header) *RollupSearchService {
	s.headers = headers
	return s
}

// SearchSource sets the search source builder to use with this service.
func (s *RollupSearchService) SearchSource(searchSource *SearchSource) *RollupSearchService {
	s.searchSource = searchSource
	if s.searchSource == nil {
		s.searchSource = NewSearchSource()
	}
	return s
}

// Source allows the user to set the request body manually without using
// any of the structs and interfaces in Elastic.
func (s *RollupSearchService) Source(source interface{}) *RollupSearchService {
	s.source = source
	return s
}

// Index sets the names of the indices to use for search.
func (s *RollupSearchService) Index(index ...string) *RollupSearchService {
	s.index = append(s.index, index...)
	return s
}

// Type adds search restrictions for a list of types.
//
// Deprecated: Types are in the process of being removed. Instead of using a type, prefer to
// filter on a field on the document.
func (s *RollupSearchService) Type(typ ...string) *RollupSearchService {
	s.typ = append(s.typ, typ...)
	return s
}

// Timeout sets the timeout to use, e.g. "1s" or "1000ms".
func (s *RollupSearchService) Timeout(timeout string) *RollupSearchService {
	s.searchSource = s.searchSource.Timeout(timeout)
	return s
}

// Profile sets the Profile API flag on the search source.
// When enabled, a search executed by this service will return query
// profiling data.
func (s *RollupSearchService) Profile(profile bool) *RollupSearchService {
	s.searchSource = s.searchSource.Profile(profile)
	return s
}

// Collapse adds field collapsing.
func (s *RollupSearchService) Collapse(collapse *CollapseBuilder) *RollupSearchService {
	s.searchSource = s.searchSource.Collapse(collapse)
	return s
}

// PointInTime specifies an optional PointInTime to be used in the context
// of this search.
func (s *RollupSearchService) PointInTime(pointInTime *PointInTime) *RollupSearchService {
	s.searchSource = s.searchSource.PointInTime(pointInTime)
	return s
}

// RuntimeMappings specifies optional runtime mappings.
func (s *RollupSearchService) RuntimeMappings(runtimeMappings RuntimeMappings) *RollupSearchService {
	s.searchSource = s.searchSource.RuntimeMappings(runtimeMappings)
	return s
}

// TimeoutInMillis sets the timeout in milliseconds.
func (s *RollupSearchService) TimeoutInMillis(timeoutInMillis int) *RollupSearchService {
	s.searchSource = s.searchSource.TimeoutInMillis(timeoutInMillis)
	return s
}

// TerminateAfter specifies the maximum number of documents to collect for
// each shard, upon reaching which the query execution will terminate early.
func (s *RollupSearchService) TerminateAfter(terminateAfter int) *RollupSearchService {
	s.searchSource = s.searchSource.TerminateAfter(terminateAfter)
	return s
}

// SearchType sets the search operation type. Valid values are:
// "dfs_query_then_fetch" and "query_then_fetch".
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-request-search-type.html
// for details.
func (s *RollupSearchService) SearchType(searchType string) *RollupSearchService {
	s.searchType = searchType
	return s
}

// Routing is a list of specific routing values to control the shards
// the search will be executed on.
func (s *RollupSearchService) Routing(routings ...string) *RollupSearchService {
	s.routing = strings.Join(routings, ",")
	return s
}

// Preference sets the preference to execute the search. Defaults to
// randomize across shards ("random"). Can be set to "_local" to prefer
// local shards, "_primary" to execute on primary shards only,
// or a custom value which guarantees that the same order will be used
// across different requests.
func (s *RollupSearchService) Preference(preference string) *RollupSearchService {
	s.preference = preference
	return s
}

// RequestCache indicates whether the cache should be used for this
// request or not, defaults to index level setting.
func (s *RollupSearchService) RequestCache(requestCache bool) *RollupSearchService {
	s.requestCache = &requestCache
	return s
}

// Query sets the query to perform, e.g. MatchAllQuery.
func (s *RollupSearchService) Query(query Query) *RollupSearchService {
	s.searchSource = s.searchSource.Query(query)
	return s
}

// PostFilter will be executed after the query has been executed and
// only affects the search hits, not the aggregations.
// This filter is always executed as the last filtering mechanism.
func (s *RollupSearchService) PostFilter(postFilter Query) *RollupSearchService {
	s.searchSource = s.searchSource.PostFilter(postFilter)
	return s
}

// FetchSource indicates whether the response should contain the stored
// _source for every hit.
func (s *RollupSearchService) FetchSource(fetchSource bool) *RollupSearchService {
	s.searchSource = s.searchSource.FetchSource(fetchSource)
	return s
}

// FetchSourceContext indicates how the _source should be fetched.
func (s *RollupSearchService) FetchSourceContext(fetchSourceContext *FetchSourceContext) *RollupSearchService {
	s.searchSource = s.searchSource.FetchSourceContext(fetchSourceContext)
	return s
}

// Highlight adds highlighting to the search.
func (s *RollupSearchService) Highlight(highlight *Highlight) *RollupSearchService {
	s.searchSource = s.searchSource.Highlight(highlight)
	return s
}

// GlobalSuggestText defines the global text to use with all suggesters.
// This avoids repetition.
func (s *RollupSearchService) GlobalSuggestText(globalText string) *RollupSearchService {
	s.searchSource = s.searchSource.GlobalSuggestText(globalText)
	return s
}

// Suggester adds a suggester to the search.
func (s *RollupSearchService) Suggester(suggester Suggester) *RollupSearchService {
	s.searchSource = s.searchSource.Suggester(suggester)
	return s
}

// Aggregation adds an aggreation to perform as part of the search.
func (s *RollupSearchService) Aggregation(name string, aggregation Aggregation) *RollupSearchService {
	s.searchSource = s.searchSource.Aggregation(name, aggregation)
	return s
}

// MinScore sets the minimum score below which docs will be filtered out.
func (s *RollupSearchService) MinScore(minScore float64) *RollupSearchService {
	s.searchSource = s.searchSource.MinScore(minScore)
	return s
}

// From index to start the search from. Defaults to 0.
func (s *RollupSearchService) From(from int) *RollupSearchService {
	s.searchSource = s.searchSource.From(from)
	return s
}

// Size is the number of search hits to return. Defaults to 10.
func (s *RollupSearchService) Size(size int) *RollupSearchService {
	s.searchSource = s.searchSource.Size(size)
	return s
}

// Explain indicates whether each search hit should be returned with
// an explanation of the hit (ranking).
func (s *RollupSearchService) Explain(explain bool) *RollupSearchService {
	s.searchSource = s.searchSource.Explain(explain)
	return s
}

// Version indicates whether each search hit should be returned with
// a version associated to it.
func (s *RollupSearchService) Version(version bool) *RollupSearchService {
	s.searchSource = s.searchSource.Version(version)
	return s
}

// Sort adds a sort order.
func (s *RollupSearchService) Sort(field string, ascending bool) *RollupSearchService {
	s.searchSource = s.searchSource.Sort(field, ascending)
	return s
}

// SortWithInfo adds a sort order.
func (s *RollupSearchService) SortWithInfo(info SortInfo) *RollupSearchService {
	s.searchSource = s.searchSource.SortWithInfo(info)
	return s
}

// SortBy adds a sort order.
func (s *RollupSearchService) SortBy(sorter ...Sorter) *RollupSearchService {
	s.searchSource = s.searchSource.SortBy(sorter...)
	return s
}

// DocvalueField adds a single field to load from the field data cache
// and return as part of the search.
func (s *RollupSearchService) DocvalueField(docvalueField string) *RollupSearchService {
	s.searchSource = s.searchSource.DocvalueField(docvalueField)
	return s
}

// DocvalueFieldWithFormat adds a single field to load from the field data cache
// and return as part of the search.
func (s *RollupSearchService) DocvalueFieldWithFormat(docvalueField DocvalueField) *RollupSearchService {
	s.searchSource = s.searchSource.DocvalueFieldWithFormat(docvalueField)
	return s
}

// DocvalueFields adds one or more fields to load from the field data cache
// and return as part of the search.
func (s *RollupSearchService) DocvalueFields(docvalueFields ...string) *RollupSearchService {
	s.searchSource = s.searchSource.DocvalueFields(docvalueFields...)
	return s
}

// DocvalueFieldsWithFormat adds one or more fields to load from the field data cache
// and return as part of the search.
func (s *RollupSearchService) DocvalueFieldsWithFormat(docvalueFields ...DocvalueField) *RollupSearchService {
	s.searchSource = s.searchSource.DocvalueFieldsWithFormat(docvalueFields...)
	return s
}

// NoStoredFields indicates that no stored fields should be loaded, resulting in only
// id and type to be returned per field.
func (s *RollupSearchService) NoStoredFields() *RollupSearchService {
	s.searchSource = s.searchSource.NoStoredFields()
	return s
}

// StoredField adds a single field to load and return (note, must be stored) as
// part of the search request. If none are specified, the source of the
// document will be returned.
func (s *RollupSearchService) StoredField(fieldName string) *RollupSearchService {
	s.searchSource = s.searchSource.StoredField(fieldName)
	return s
}

// StoredFields	sets the fields to load and return as part of the search request.
// If none are specified, the source of the document will be returned.
func (s *RollupSearchService) StoredFields(fields ...string) *RollupSearchService {
	s.searchSource = s.searchSource.StoredFields(fields...)
	return s
}

// TrackScores is applied when sorting and controls if scores will be
// tracked as well. Defaults to false.
func (s *RollupSearchService) TrackScores(trackScores bool) *RollupSearchService {
	s.searchSource = s.searchSource.TrackScores(trackScores)
	return s
}

// TrackTotalHits controls if the total hit count for the query should be tracked.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.1/search-request-track-total-hits.html
// for details.
func (s *RollupSearchService) TrackTotalHits(trackTotalHits interface{}) *RollupSearchService {
	s.searchSource = s.searchSource.TrackTotalHits(trackTotalHits)
	return s
}

// SearchAfter allows a different form of pagination by using a live cursor,
// using the results of the previous page to help the retrieval of the next.
//
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-request-search-after.html
func (s *RollupSearchService) SearchAfter(sortValues ...interface{}) *RollupSearchService {
	s.searchSource = s.searchSource.SearchAfter(sortValues...)
	return s
}

// DefaultRescoreWindowSize sets the rescore window size for rescores
// that don't specify their window.
func (s *RollupSearchService) DefaultRescoreWindowSize(defaultRescoreWindowSize int) *RollupSearchService {
	s.searchSource = s.searchSource.DefaultRescoreWindowSize(defaultRescoreWindowSize)
	return s
}

// Rescorer adds a rescorer to the search.
func (s *RollupSearchService) Rescorer(rescore *Rescore) *RollupSearchService {
	s.searchSource = s.searchSource.Rescorer(rescore)
	return s
}

// IgnoreUnavailable indicates whether the specified concrete indices
// should be ignored when unavailable (missing or closed).
func (s *RollupSearchService) IgnoreUnavailable(ignoreUnavailable bool) *RollupSearchService {
	s.ignoreUnavailable = &ignoreUnavailable
	return s
}

// IgnoreThrottled indicates whether specified concrete, expanded or aliased
// indices should be ignored when throttled.
func (s *RollupSearchService) IgnoreThrottled(ignoreThrottled bool) *RollupSearchService {
	s.ignoreThrottled = &ignoreThrottled
	return s
}

// AllowNoIndices indicates whether to ignore if a wildcard indices
// expression resolves into no concrete indices. (This includes `_all` string
// or when no indices have been specified).
func (s *RollupSearchService) AllowNoIndices(allowNoIndices bool) *RollupSearchService {
	s.allowNoIndices = &allowNoIndices
	return s
}

// ExpandWildcards indicates whether to expand wildcard expression to
// concrete indices that are open, closed or both.
func (s *RollupSearchService) ExpandWildcards(expandWildcards string) *RollupSearchService {
	s.expandWildcards = expandWildcards
	return s
}

// Lenient specifies whether format-based query failures (such as providing
// text to a numeric field) should be ignored.
func (s *RollupSearchService) Lenient(lenient bool) *RollupSearchService {
	s.lenient = &lenient
	return s
}

// MaxResponseSize sets an upper limit on the response body size that we accept,
// to guard against OOM situations.
func (s *RollupSearchService) MaxResponseSize(maxResponseSize int64) *RollupSearchService {
	s.maxResponseSize = maxResponseSize
	return s
}

// AllowPartialRollupSearchResults indicates if an error should be returned if
// there is a partial search failure or timeout.
func (s *RollupSearchService) AllowPartialRollupSearchResults(enabled bool) *RollupSearchService {
	s.allowPartialRollupSearchResults = &enabled
	return s
}

// TypedKeys specifies whether aggregation and suggester names should be
// prefixed by their respective types in the response.
func (s *RollupSearchService) TypedKeys(enabled bool) *RollupSearchService {
	s.typedKeys = &enabled
	return s
}

// SeqNoPrimaryTerm is an alias for SeqNoAndPrimaryTerm.
//
// Deprecated: Use SeqNoAndPrimaryTerm.
func (s *RollupSearchService) SeqNoPrimaryTerm(enabled bool) *RollupSearchService {
	return s.SeqNoAndPrimaryTerm(enabled)
}

// SeqNoAndPrimaryTerm specifies whether to return sequence number and
// primary term of the last modification of each hit.
func (s *RollupSearchService) SeqNoAndPrimaryTerm(enabled bool) *RollupSearchService {
	s.seqNoPrimaryTerm = &enabled
	return s
}

// BatchedReduceSize specifies the number of shard results that should be reduced
// at once on the coordinating node. This value should be used as a protection
// mechanism to reduce the memory overhead per search request if the potential
// number of shards in the request can be large.
func (s *RollupSearchService) BatchedReduceSize(size int) *RollupSearchService {
	s.batchedReduceSize = &size
	return s
}

// MaxConcurrentShardRequests specifies the number of concurrent shard requests
// this search executes concurrently. This value should be used to limit the
// impact of the search on the cluster in order to limit the number of
// concurrent shard requests.
func (s *RollupSearchService) MaxConcurrentShardRequests(max int) *RollupSearchService {
	s.maxConcurrentShardRequests = &max
	return s
}

// PreFilterShardSize specifies a threshold that enforces a pre-filter roundtrip
// to prefilter search shards based on query rewriting if the number of shards
// the search request expands to exceeds the threshold. This filter roundtrip
// can limit the number of shards significantly if for instance a shard can
// not match any documents based on it's rewrite method i.e. if date filters are
// mandatory to match but the shard bounds and the query are disjoint.
func (s *RollupSearchService) PreFilterShardSize(threshold int) *RollupSearchService {
	s.preFilterShardSize = &threshold
	return s
}

// RestTotalHitsAsInt indicates whether hits.total should be rendered as an
// integer or an object in the rest search response.
func (s *RollupSearchService) RestTotalHitsAsInt(enabled bool) *RollupSearchService {
	s.restTotalHitsAsInt = &enabled
	return s
}

// CCSMinimizeRoundtrips indicates whether network round-trips should be minimized
// as part of cross-cluster search requests execution.
func (s *RollupSearchService) CCSMinimizeRoundtrips(enabled bool) *RollupSearchService {
	s.ccsMinimizeRoundtrips = &enabled
	return s
}

// buildURL builds the URL for the operation.
func (s *RollupSearchService) buildURL() (string, url.Values, error) {
	var err error
	var path string

	if len(s.index) > 0 && len(s.typ) > 0 {
		path, err = uritemplates.Expand("/{index}/{type}/_rollup_search", map[string]string{
			"index": strings.Join(s.index, ","),
			"type":  strings.Join(s.typ, ","),
		})
	} else if len(s.index) > 0 {
		path, err = uritemplates.Expand("/{index}/_rollup_search", map[string]string{
			"index": strings.Join(s.index, ","),
		})
	} else if len(s.typ) > 0 {
		path, err = uritemplates.Expand("/_all/{type}/_rollup_search", map[string]string{
			"type": strings.Join(s.typ, ","),
		})
	} else {
		path = "/_rollup_search"
	}
	if err != nil {
		return "", url.Values{}, err
	}

	// Add query string parameters
	params := url.Values{}
	if v := s.pretty; v != nil {
		params.Set("pretty", fmt.Sprint(*v))
	}
	if v := s.human; v != nil {
		params.Set("human", fmt.Sprint(*v))
	}
	if v := s.errorTrace; v != nil {
		params.Set("error_trace", fmt.Sprint(*v))
	}
	if len(s.filterPath) > 0 {
		params.Set("filter_path", strings.Join(s.filterPath, ","))
	}
	if s.searchType != "" {
		params.Set("search_type", s.searchType)
	}
	if s.routing != "" {
		params.Set("routing", s.routing)
	}
	if s.preference != "" {
		params.Set("preference", s.preference)
	}
	if v := s.requestCache; v != nil {
		params.Set("request_cache", fmt.Sprint(*v))
	}
	if v := s.allowNoIndices; v != nil {
		params.Set("allow_no_indices", fmt.Sprint(*v))
	}
	if s.expandWildcards != "" {
		params.Set("expand_wildcards", s.expandWildcards)
	}
	if v := s.lenient; v != nil {
		params.Set("lenient", fmt.Sprint(*v))
	}
	if v := s.ignoreUnavailable; v != nil {
		params.Set("ignore_unavailable", fmt.Sprint(*v))
	}
	if v := s.ignoreThrottled; v != nil {
		params.Set("ignore_throttled", fmt.Sprint(*v))
	}
	if s.seqNoPrimaryTerm != nil {
		params.Set("seq_no_primary_term", fmt.Sprint(*s.seqNoPrimaryTerm))
	}
	if v := s.allowPartialRollupSearchResults; v != nil {
		params.Set("allow_partial_search_results", fmt.Sprint(*v))
	}
	if v := s.typedKeys; v != nil {
		params.Set("typed_keys", fmt.Sprint(*v))
	}
	if v := s.batchedReduceSize; v != nil {
		params.Set("batched_reduce_size", fmt.Sprint(*v))
	}
	if v := s.maxConcurrentShardRequests; v != nil {
		params.Set("max_concurrent_shard_requests", fmt.Sprint(*v))
	}
	if v := s.preFilterShardSize; v != nil {
		params.Set("pre_filter_shard_size", fmt.Sprint(*v))
	}
	if v := s.restTotalHitsAsInt; v != nil {
		params.Set("rest_total_hits_as_int", fmt.Sprint(*v))
	}
	if v := s.ccsMinimizeRoundtrips; v != nil {
		params.Set("ccs_minimize_roundtrips", fmt.Sprint(*v))
	}
	return path, params, nil
}

// Validate checks if the operation is valid.
func (s *RollupSearchService) Validate() error {
	return nil
}

// Do executes the search and returns a RollupSearchResult.
func (s *RollupSearchService) Do(ctx context.Context) (*RollupSearchResult, error) {
	// Check pre-conditions
	if err := s.Validate(); err != nil {
		return nil, err
	}

	// Get URL for request
	path, params, err := s.buildURL()
	if err != nil {
		return nil, err
	}

	// Perform request
	var body interface{}
	if s.source != nil {
		body = s.source
	} else {
		src, err := s.searchSource.Source()
		if err != nil {
			return nil, err
		}
		body = src
	}
	res, err := s.client.PerformRequest(ctx, PerformRequestOptions{
		Method:          "POST",
		Path:            path,
		Params:          params,
		Body:            body,
		Headers:         s.headers,
		MaxResponseSize: s.maxResponseSize,
	})
	if err != nil {
		return nil, err
	}

	// Return search results
	ret := new(RollupSearchResult)
	if err := s.client.decoder.Decode(res.Body, ret); err != nil {
		ret.Header = res.Header
		return nil, err
	}
	ret.Header = res.Header
	return ret, nil
}

// RollupSearchResult is the result of a search in Elasticsearch.
type RollupSearchResult struct {
	Header          http.Header                `json:"-"`
	TookInMillis    int64                      `json:"took,omitempty"`             // search time in milliseconds
	TerminatedEarly bool                       `json:"terminated_early,omitempty"` // request terminated early
	NumReducePhases int                        `json:"num_reduce_phases,omitempty"`
	Clusters        *RollupSearchResultCluster `json:"_clusters,omitempty"`    // 6.1.0+
	ScrollId        string                     `json:"_scroll_id,omitempty"`   // only used with Scroll and Scan operations
	Hits            *RollupSearchHits          `json:"hits,omitempty"`         // the actual search hits
	Suggest         RollupSearchSuggest        `json:"suggest,omitempty"`      // results from suggesters
	Aggregations    Aggregations               `json:"aggregations,omitempty"` // results from aggregations
	TimedOut        bool                       `json:"timed_out,omitempty"`    // true if the search timed out
	Error           *ErrorDetails              `json:"error,omitempty"`        // only used in MultiGet
	Profile         *RollupSearchProfile       `json:"profile,omitempty"`      // profiling results, if optional Profile API was active for this search
	Shards          *ShardsInfo                `json:"_shards,omitempty"`      // shard information
	Status          int                        `json:"status,omitempty"`       // used in MultiSearch
	PitId           string                     `json:"pit_id,omitempty"`       // Point In Time ID
}

// RollupSearchResultCluster holds information about a search response
// from a cluster.
type RollupSearchResultCluster struct {
	Successful int `json:"successful,omitempty"`
	Total      int `json:"total,omitempty"`
	Skipped    int `json:"skipped,omitempty"`
}

// RollupTotalHits is a convenience function to return the number of hits for
// a search result. The return value might not be accurate, unless
// track_total_hits parameter has set to true.
func (r *RollupSearchResult) RollupTotalHits() int64 {
	if r != nil && r.Hits != nil && r.Hits.RollupTotalHits != nil {
		return r.Hits.RollupTotalHits.Value
	}
	return 0
}

// Each is a utility function to iterate over all hits. It saves you from
// checking for nil values. Notice that Each will ignore errors in
// serializing JSON and hits with empty/nil _source will get an empty
// value
func (r *RollupSearchResult) Each(typ reflect.Type) []interface{} {
	if r.Hits == nil || r.Hits.Hits == nil || len(r.Hits.Hits) == 0 {
		return nil
	}
	slice := make([]interface{}, 0, len(r.Hits.Hits))
	for _, hit := range r.Hits.Hits {
		v := reflect.New(typ).Elem()
		if hit.Source == nil {
			slice = append(slice, v.Interface())
			continue
		}
		if err := json.Unmarshal(hit.Source, v.Addr().Interface()); err == nil {
			slice = append(slice, v.Interface())
		}
	}
	return slice
}

// RollupSearchHits specifies the list of search hits.
type RollupSearchHits struct {
	RollupTotalHits *RollupTotalHits   `json:"total,omitempty"`     // total number of hits found
	MaxScore        *float64           `json:"max_score,omitempty"` // maximum score of all hits
	Hits            []*RollupSearchHit `json:"hits,omitempty"`      // the actual hits returned
}

// RollupNestedHit is a nested innerhit
type RollupNestedHit struct {
	Field  string           `json:"field"`
	Offset int              `json:"offset,omitempty"`
	Child  *RollupNestedHit `json:"_nested,omitempty"`
}

// RollupTotalHits specifies total number of hits and its relation
type RollupTotalHits struct {
	Value    int64  `json:"value"`    // value of the total hit count
	Relation string `json:"relation"` // how the value should be interpreted: accurate ("eq") or a lower bound ("gte")
}

// UnmarshalJSON into RollupTotalHits, accepting both the new response structure
// in ES 7.x as well as the older response structure in earlier versions.
// The latter can be enabled with RestTotalHitsAsInt(true).
func (h *RollupTotalHits) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		return nil
	}
	var v struct {
		Value    int64  `json:"value"`    // value of the total hit count
		Relation string `json:"relation"` // how the value should be interpreted: accurate ("eq") or a lower bound ("gte")
	}
	if err := json.Unmarshal(data, &v); err != nil {
		var count int64
		if err2 := json.Unmarshal(data, &count); err2 != nil {
			return err // return inner error
		}
		h.Value = count
		h.Relation = "eq"
		return nil
	}
	*h = v
	return nil
}

// RollupSearchHit is a single hit.
type RollupSearchHit struct {
	Score          *float64                             `json:"_score,omitempty"`   // computed score
	Index          string                               `json:"_index,omitempty"`   // index name
	Type           string                               `json:"_type,omitempty"`    // type meta field
	Id             string                               `json:"_id,omitempty"`      // external or internal
	Uid            string                               `json:"_uid,omitempty"`     // uid meta field (see MapperService.java for all meta fields)
	Routing        string                               `json:"_routing,omitempty"` // routing meta field
	Parent         string                               `json:"_parent,omitempty"`  // parent meta field
	Version        *int64                               `json:"_version,omitempty"` // version number, when Version is set to true in RollupSearchService
	SeqNo          *int64                               `json:"_seq_no"`
	PrimaryTerm    *int64                               `json:"_primary_term"`
	Sort           []interface{}                        `json:"sort,omitempty"`            // sort information
	Highlight      RollupSearchHitHighlight             `json:"highlight,omitempty"`       // highlighter information
	Source         json.RawMessage                      `json:"_source,omitempty"`         // stored document source
	Fields         RollupSearchHitFields                `json:"fields,omitempty"`          // returned (stored) fields
	Explanation    *RollupSearchExplanation             `json:"_explanation,omitempty"`    // explains how the score was computed
	MatchedQueries []string                             `json:"matched_queries,omitempty"` // matched queries
	InnerHits      map[string]*RollupSearchHitInnerHits `json:"inner_hits,omitempty"`      // inner hits with ES >= 1.5.0
	Nested         *RollupNestedHit                     `json:"_nested,omitempty"`         // for nested inner hits
	Shard          string                               `json:"_shard,omitempty"`          // used e.g. in Search Explain
	Node           string                               `json:"_node,omitempty"`           // used e.g. in Search Explain

	// HighlightFields
	// SortValues
	// MatchedFilters
}

// RollupSearchHitFields helps to simplify resolving slices of specific types.
type RollupSearchHitFields map[string]interface{}

// Strings returns a slice of strings for the given field, if there is any
// such field in the hit. The method ignores elements that are not of type
// string.
func (f RollupSearchHitFields) Strings(fieldName string) ([]string, bool) {
	slice, ok := f[fieldName].([]interface{})
	if !ok {
		return nil, false
	}
	results := make([]string, 0, len(slice))
	for _, item := range slice {
		if v, ok := item.(string); ok {
			results = append(results, v)
		}
	}
	return results, true
}

// Float64s returns a slice of float64's for the given field, if there is any
// such field in the hit. The method ignores elements that are not of
// type float64.
func (f RollupSearchHitFields) Float64s(fieldName string) ([]float64, bool) {
	slice, ok := f[fieldName].([]interface{})
	if !ok {
		return nil, false
	}
	results := make([]float64, 0, len(slice))
	for _, item := range slice {
		if v, ok := item.(float64); ok {
			results = append(results, v)
		}
	}
	return results, true
}

// RollupSearchHitInnerHits is used for inner hits.
type RollupSearchHitInnerHits struct {
	Hits *RollupSearchHits `json:"hits,omitempty"`
}

// RollupSearchExplanation explains how the score for a hit was computed.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-request-explain.html.
type RollupSearchExplanation struct {
	Value       float64                   `json:"value"`             // e.g. 1.0
	Description string                    `json:"description"`       // e.g. "boost" or "ConstantScore(*:*), product of:"
	Details     []RollupSearchExplanation `json:"details,omitempty"` // recursive details
}

// Suggest

// RollupSearchSuggest is a map of suggestions.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-suggesters.html.
type RollupSearchSuggest map[string][]RollupSearchSuggestion

// RollupSearchSuggestion is a single search suggestion.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-suggesters.html.
type RollupSearchSuggestion struct {
	Text    string                         `json:"text"`
	Offset  int                            `json:"offset"`
	Length  int                            `json:"length"`
	Options []RollupSearchSuggestionOption `json:"options"`
}

// RollupSearchSuggestionOption is an option of a RollupSearchSuggestion.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-suggesters.html.
type RollupSearchSuggestionOption struct {
	Text            string              `json:"text"`
	Index           string              `json:"_index"`
	Type            string              `json:"_type"`
	Id              string              `json:"_id"`
	Score           float64             `json:"score"`  // term and phrase suggesters uses "score" as of 6.2.4
	ScoreUnderscore float64             `json:"_score"` // completion and context suggesters uses "_score" as of 6.2.4
	Highlighted     string              `json:"highlighted"`
	CollateMatch    bool                `json:"collate_match"`
	Freq            int                 `json:"freq"` // from TermSuggestion.Option in Java API
	Source          json.RawMessage     `json:"_source"`
	Contexts        map[string][]string `json:"contexts,omitempty"`
}

// RollupSearchProfile is a list of shard profiling data collected during
// query execution in the "profile" section of a RollupSearchResult
type RollupSearchProfile struct {
	Shards []RollupSearchProfileShardResult `json:"shards"`
}

// RollupSearchProfileShardResult returns the profiling data for a single shard
// accessed during the search query or aggregation.
type RollupSearchProfileShardResult struct {
	ID           string                          `json:"id"`
	Searches     []RollupQueryProfileShardResult `json:"searches"`
	Aggregations []RollupProfileResult           `json:"aggregations"`
	Fetch        *RollupProfileResult            `json:"fetch"`
}

// RollupQueryProfileShardResult is a container class to hold the profile results
// for a single shard in the request. It comtains a list of query profiles,
// a collector tree and a total rewrite tree.
type RollupQueryProfileShardResult struct {
	Query       []RollupProfileResult `json:"query,omitempty"`
	RewriteTime int64                 `json:"rewrite_time,omitempty"`
	Collector   []interface{}         `json:"collector,omitempty"`
}

// RollupCollectorResult holds the profile timings of the collectors used in the
// search. Children's CollectorResults may be embedded inside of a parent
// RollupCollectorResult.
type RollupCollectorResult struct {
	Name      string                  `json:"name,omitempty"`
	Reason    string                  `json:"reason,omitempty"`
	Time      string                  `json:"time,omitempty"`
	TimeNanos int64                   `json:"time_in_nanos,omitempty"`
	Children  []RollupCollectorResult `json:"children,omitempty"`
}

// RollupProfileResult is the internal representation of a profiled query,
// corresponding to a single node in the query tree.
type RollupProfileResult struct {
	Type          string                 `json:"type"`
	Description   string                 `json:"description,omitempty"`
	NodeTime      string                 `json:"time,omitempty"`
	NodeTimeNanos int64                  `json:"time_in_nanos,omitempty"`
	Breakdown     map[string]int64       `json:"breakdown,omitempty"`
	Children      []RollupProfileResult  `json:"children,omitempty"`
	Debug         map[string]interface{} `json:"debug,omitempty"`
}

// Aggregations (see search_aggs.go)

// Highlighting

// RollupSearchHitHighlight is the highlight information of a search hit.
// See https://www.elastic.co/guide/en/elasticsearch/reference/7.0/search-request-highlighting.html
// for a general discussion of highlighting.
type RollupSearchHitHighlight map[string][]string
