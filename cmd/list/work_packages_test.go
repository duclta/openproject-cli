package list

import (
	"strconv"
	"testing"

	"github.com/opf/openproject-cli/components/requests"
)

func TestBuildQuery_UsesDefaultLimitAndSort(t *testing.T) {
	resetWorkPackagesCommandState()
	t.Cleanup(resetWorkPackagesCommandState)

	sortByValue, err := buildSortBy(defaultWorkPackageSort)
	if err != nil {
		t.Fatalf("build default sort value: %v", err)
	}

	query, err := buildQuery()
	if err != nil {
		t.Fatalf("build query: %v", err)
	}

	expected := requests.NewQuery(
		map[string]string{
			"offset":   strconv.Itoa(defaultWorkPackagePage),
			"pageSize": strconv.Itoa(defaultWorkPackageLimit),
			"sortBy":   sortByValue,
		},
		nil,
	)

	if !query.Equals(expected) {
		t.Fatalf("expected %+v, got %+v", expected, query)
	}
}

func TestBuildQuery_UsesCustomLimitAndSort(t *testing.T) {
	resetWorkPackagesCommandState()
	t.Cleanup(resetWorkPackagesCommandState)

	limit = -1
	page = 3
	sortBy = "status:asc"

	sortByValue, err := buildSortBy(sortBy)
	if err != nil {
		t.Fatalf("build custom sort value: %v", err)
	}

	query, err := buildQuery()
	if err != nil {
		t.Fatalf("build query: %v", err)
	}

	expected := requests.NewQuery(
		map[string]string{
			"offset":   strconv.Itoa(page),
			"pageSize": strconv.Itoa(limit),
			"sortBy":   sortByValue,
		},
		nil,
	)

	if !query.Equals(expected) {
		t.Fatalf("expected %+v, got %+v", expected, query)
	}
}

func TestBuildQuery_SkipsPaginationAndSortForTotal(t *testing.T) {
	resetWorkPackagesCommandState()
	t.Cleanup(resetWorkPackagesCommandState)

	showTotal = true

	query, err := buildQuery()
	if err != nil {
		t.Fatalf("build query: %v", err)
	}

	if !query.Equals(requests.NewEmptyQuery()) {
		t.Fatalf("expected empty query, got %+v", query)
	}
}

func TestBuildQuery_RejectsInvalidLimit(t *testing.T) {
	resetWorkPackagesCommandState()
	t.Cleanup(resetWorkPackagesCommandState)

	limit = 0

	_, err := buildQuery()
	if err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func TestBuildQuery_RejectsInvalidPage(t *testing.T) {
	resetWorkPackagesCommandState()
	t.Cleanup(resetWorkPackagesCommandState)

	page = 0

	_, err := buildQuery()
	if err == nil {
		t.Fatal("expected invalid page error")
	}
}

func TestBuildQuery_RejectsInvalidSort(t *testing.T) {
	resetWorkPackagesCommandState()
	t.Cleanup(resetWorkPackagesCommandState)

	sortBy = "id"

	_, err := buildQuery()
	if err == nil {
		t.Fatal("expected invalid sort error")
	}
}

func resetWorkPackagesCommandState() {
	assignee = ""
	projectId = 0
	showTotal = false
	statusFilter = ""
	typeFilter = ""
	includeSubProjects = false
	limit = defaultWorkPackageLimit
	page = defaultWorkPackagePage
	sortBy = defaultWorkPackageSort

	for _, filter := range activeFilters {
		*filter.ValuePointer() = filter.DefaultValue()
	}
}
