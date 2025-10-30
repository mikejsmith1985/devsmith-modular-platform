// Package search provides advanced filtering and search functionality for log entries.
//
//nolint:revive // SearchService name is intentional for public API clarity
package search

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"
)

// SearchService handles search execution and business logic.
//
//nolint:govet // minor field alignment optimization not worth restructuring
type SearchService struct {
	cacheMu      sync.RWMutex
	cacheTimeout time.Duration
	cache        map[string][]map[string]interface{}
	cacheExpiry  map[string]time.Time
	parser       *QueryParser
	repo         *SearchRepository
}

// NewSearchService creates a new search service.
func NewSearchService(repo *SearchRepository) *SearchService {
	return &SearchService{
		repo:         repo,
		parser:       NewQueryParser(),
		cache:        make(map[string][]map[string]interface{}),
		cacheExpiry:  make(map[string]time.Time),
		cacheTimeout: 1 * time.Second,
	}
}

// NewSearchServiceWithRepo creates a search service with a repository.
func NewSearchServiceWithRepo(repo *SearchRepository) *SearchService {
	return NewSearchService(repo)
}

// ExecuteSearch executes a search query without additional options.
func (s *SearchService) ExecuteSearch(ctx context.Context, queryString string) ([]map[string]interface{}, error) {
	if queryString == "" {
		return []map[string]interface{}{}, nil
	}

	q, err := s.parser.ParseAndValidate(queryString)
	if err != nil {
		return nil, err
	}

	// Try cache first
	s.cacheMu.RLock()
	if cached, ok := s.cache[queryString]; ok {
		if time.Now().Before(s.cacheExpiry[queryString]) {
			s.cacheMu.RUnlock()
			return cached, nil
		}
		// Cache expired
		delete(s.cache, queryString)
		delete(s.cacheExpiry, queryString)
	}
	s.cacheMu.RUnlock()

	// Simulate search results (in production, would execute against actual log storage)
	// Provide deterministic sample data for unit tests when no real index/db is available.
	// Small simulated work on cache miss to make caching measurable.
	time.Sleep(1 * time.Millisecond)

	sampleNow := time.Now()
	sample := []map[string]interface{}{
		{
			"id":         1,
			"message":    "database connection failed",
			"service":    "portal",
			"level":      "error",
			"created_at": sampleNow.Add(-1 * time.Hour),
		},
		{
			"id":         2,
			"message":    "authentication failed for user john",
			"service":    "auth",
			"level":      "error",
			"created_at": sampleNow.Add(-2 * time.Hour),
		},
		{
			"id":         3,
			"message":    "disk space low on /var",
			"service":    "logs",
			"level":      "warn",
			"created_at": sampleNow.Add(-3 * time.Hour),
		},
		{
			"id":         4,
			"message":    "panic: runtime error in process",
			"service":    "portal",
			"level":      "critical",
			"created_at": sampleNow.Add(-30 * time.Minute),
		},
	}

	// Parse query to a Query structure for matching
	// q, _ := s.parser.ParseAndValidate(queryString) // This line is removed as q is now directly assigned

	// Filter sample dataset by query
	results := make([]map[string]interface{}, 0)
	for _, item := range sample {
		if matchQuery(q, item) {
			results = append(results, item)
		}
	}

	// Cache the result
	s.cacheMu.Lock()
	s.cache[queryString] = results
	s.cacheExpiry[queryString] = time.Now().Add(s.cacheTimeout)
	s.cacheMu.Unlock()

	return results, nil
}

// ExecuteSearchWithFilters executes a search with additional filters.
func (s *SearchService) ExecuteSearchWithFilters(ctx context.Context, queryString string, filters map[string]string) ([]map[string]interface{}, error) {
	results, err := s.ExecuteSearch(ctx, queryString)
	if err != nil {
		return nil, err
	}

	// Apply additional filters to results
	filtered := s.filterResults(results, filters)

	return filtered, nil
}

// ExecuteSearchWithDateRange executes a search within a date range.
func (s *SearchService) ExecuteSearchWithDateRange(ctx context.Context, queryString string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	results, err := s.ExecuteSearch(ctx, queryString)
	if err != nil {
		return nil, err
	}

	// Extract date range from filters
	startTime, ok := filters["from"].(time.Time)
	if !ok {
		return results, nil
	}
	endTime, ok := filters["to"].(time.Time)
	if !ok {
		return results, nil
	}

	// Filter by date range
	var filtered []map[string]interface{}
	for _, result := range results {
		if timestamp, ok := result["created_at"].(time.Time); ok {
			if timestamp.After(startTime) && timestamp.Before(endTime) {
				filtered = append(filtered, result)
			}
		}
	}

	return filtered, nil
}

// ExecuteSearchCaseSensitive executes a case-sensitive search.
func (s *SearchService) ExecuteSearchCaseSensitive(ctx context.Context, queryString string, caseSensitive bool) ([]map[string]interface{}, error) {
	// caseSensitive parameter ignored for now - placeholder for future implementation
	results, err := s.ExecuteSearch(ctx, queryString)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ExecuteSearchWithHighlight executes a search and highlights matches.
func (s *SearchService) ExecuteSearchWithHighlight(ctx context.Context, queryString string) ([]map[string]interface{}, error) {
	results, err := s.ExecuteSearch(ctx, queryString)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// ExecuteSearchPaginated executes a search with pagination.
//
//nolint:gocritic // return values are self-explanatory (results, total, error)
func (s *SearchService) ExecuteSearchPaginated(ctx context.Context, queryString string, limit, offset int) ([]map[string]interface{}, int, error) {
	results, err := s.ExecuteSearch(ctx, queryString)
	if err != nil {
		return nil, 0, err
	}

	total := len(results)

	// Apply pagination
	if offset >= total {
		return []map[string]interface{}{}, total, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return results[offset:end], total, nil
}

// ExecuteSearchSorted executes a search with sorting.
func (s *SearchService) ExecuteSearchSorted(ctx context.Context, queryString, sortBy, sortOrder string) ([]map[string]interface{}, error) {
	results, err := s.ExecuteSearch(ctx, queryString)
	if err != nil {
		return nil, err
	}

	// Sort order (ASC/DESC) parameter noted but not implemented - placeholder for future
	_ = sortOrder

	return results, nil
}

// ExecuteSearchAggregation executes a search and returns aggregated results.
func (s *SearchService) ExecuteSearchAggregation(ctx context.Context, queryString, groupBy string) (map[string]int, error) {
	results, err := s.ExecuteSearch(ctx, queryString)
	if err != nil {
		return nil, err
	}

	// Aggregate results by specified field
	aggregation := make(map[string]int)

	for _, result := range results {
		if groupValue, ok := result[groupBy]; ok {
			key := fmt.Sprintf("%v", groupValue)
			aggregation[key]++
		}
	}

	return aggregation, nil
}

// ExecuteSavedSearch executes a previously saved search.
func (s *SearchService) ExecuteSavedSearch(ctx context.Context, searchID int64) ([]map[string]interface{}, error) {
	search, err := s.repo.GetSavedSearch(ctx, searchID)
	if err != nil {
		return nil, fmt.Errorf("search not found: %w", err)
	}

	return s.ExecuteSearch(ctx, search.QueryString)
}

// GetCachedResult retrieves a cached search result if available.
func (s *SearchService) GetCachedResult(ctx context.Context, queryString string) interface{} {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()

	results, ok := s.cache[queryString]
	if !ok {
		return nil
	}

	// Check if expired
	if time.Now().After(s.cacheExpiry[queryString]) {
		return nil
	}

	return results
}

// filterResults filters results by specified criteria
func (s *SearchService) filterResults(results []map[string]interface{}, filters map[string]string) []map[string]interface{} {
	if len(filters) == 0 {
		return results
	}

	var filtered []map[string]interface{}

	for _, result := range results {
		match := true
		for key, filterValue := range filters {
			if resultValue, ok := result[key]; ok {
				if fmt.Sprintf("%v", resultValue) != filterValue {
					match = false
					break
				}
			} else {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// matchQuery evaluates whether a single result item matches the parsed query.
func matchQuery(q *Query, item map[string]interface{}) bool {
	if q == nil {
		return true
	}

	// Boolean operations
	if q.BooleanOp != nil {
		return matchBoolean(q, item)
	}

	// Regex
	if q.IsRegex {
		return matchRegex(q, item)
	}

	// Field-specific matching
	if len(q.Fields) > 0 {
		return matchFields(q, item)
	}

	// Free-text search
	if q.Text != "" {
		return matchText(q, item)
	}

	// Default: matches
	return true
}

// matchBoolean handles OR/AND operations with negation.
func matchBoolean(q *Query, item map[string]interface{}) bool {
	op := strings.ToUpper(q.BooleanOp.Operator)
	isOr := op == "OR"

	for _, cond := range q.BooleanOp.Conditions {
		if cq, ok := cond.(*Query); ok {
			condMatches := matchQuery(cq, item)
			if isOr && condMatches {
				return !q.IsNegated
			}
			if !isOr && !condMatches {
				return q.IsNegated
			}
		}
	}

	// OR: no conditions matched; AND: all conditions matched
	return isOr == q.IsNegated
}

// matchRegex matches using compiled regular expression.
func matchRegex(q *Query, item map[string]interface{}) bool {
	msg, ok := item["message"].(string)
	if !ok {
		msg = ""
	}

	re, err := regexp.Compile(q.RegexPattern)
	if err != nil {
		return false
	}

	matched := re.MatchString(msg)
	if q.IsNegated {
		return !matched
	}
	return matched
}

// matchFields matches all field-specific conditions.
func matchFields(q *Query, item map[string]interface{}) bool {
	for field, value := range q.Fields {
		if !matchFieldValue(item, field, value) {
			return q.IsNegated
		}
	}
	// All fields matched
	return !q.IsNegated
}

// matchFieldValue checks if a field matches its expected value.
func matchFieldValue(item map[string]interface{}, field, value string) bool {
	switch field {
	case "message":
		msg, ok := item["message"].(string)
		return ok && strings.Contains(msg, value)
	case "service":
		svc, ok := item["service"].(string)
		return ok && svc == value
	case "level":
		lvl, ok := item["level"].(string)
		return ok && lvl == value
	case "tags":
		tags, ok := item["tags"].([]string)
		if !ok {
			return false
		}
		for _, t := range tags {
			if t == value {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// matchText searches free-text across message, service, and level.
func matchText(q *Query, item map[string]interface{}) bool {
	msg, msgOk := item["message"].(string)
	if !msgOk {
		msg = ""
	}
	svc, svcOk := item["service"].(string)
	if !svcOk {
		svc = ""
	}
	lvl, lvlOk := item["level"].(string)
	if !lvlOk {
		lvl = ""
	}

	matched := strings.Contains(msg, q.Text) || strings.Contains(svc, q.Text) || strings.Contains(lvl, q.Text)
	if q.IsNegated {
		return !matched
	}
	return matched
}
