package storagegroup

import (
	"github.com/epicchainlabs/neofs-sdk-go/object"
)

// SearchQuery returns search query to filter
// objects with storage group content.
func SearchQuery() object.SearchFilters {
	fs := object.SearchFilters{}
	fs.AddTypeFilter(object.MatchStringEqual, object.TypeStorageGroup)

	return fs
}
