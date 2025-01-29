// Copyright 2012-present Oliver Eilhard. All rights reserved.
// Use of this source code is governed by a MIT-license.
// See http://olivere.mit-license.org/license.txt for details.

package elastic

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/olivere/elastic/v7/uritemplates"
)

// Search for documents in Elasticsearch.
// Search for documents in Elasticsearch.
type RollupSearchService struct {
	SearchService
}

// NewRollupSearchService creates a new service for searching in Elasticsearch.
func NewRollupSearchService(client *Client) ElasticSearchService {
	builder := &RollupSearchService{
		SearchService: SearchService{
			client:       client,
			searchSource: NewSearchSource(),
		},
	}
	return builder
}

func (s *RollupSearchService) Index(index ...string) ElasticSearchService {
	s.index = append(s.index, index...)
	return s
}

// buildURL builds the URL for the operation.
func (s *RollupSearchService) BuildURL() (string, url.Values, error) {
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
	if v := s.allowPartialSearchResults; v != nil {
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

func (s *RollupSearchService) Do(ctx context.Context) (*SearchResult, error) {
	// Check pre-conditions
	if err := s.Validate(); err != nil {
		return nil, err
	}

	// Get URL for request
	path, params, err := s.BuildURL()
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
	ret := new(SearchResult)
	if err := s.client.decoder.Decode(res.Body, ret); err != nil {
		ret.Header = res.Header
		return nil, err
	}
	ret.Header = res.Header
	return ret, nil
}
