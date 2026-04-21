package models

type WorkPackage struct {
	Id          uint64
	Subject     string
	Type        string
	Assignee    string
	Status      string
	Priority    string
	Parent      string
	Version     string
	Description string
	StartDate   string
	DueDate     string
	LockVersion int
}

type WorkPackageCollection struct {
	Total    int64
	Count    int64
	PageSize int64
	Offset   int64
	Items    []*WorkPackage
}
