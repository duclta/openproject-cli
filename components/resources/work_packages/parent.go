package work_packages

import (
	"github.com/opf/openproject-cli/components/parser"
	"github.com/opf/openproject-cli/components/paths"
	"github.com/opf/openproject-cli/components/requests"
	"github.com/opf/openproject-cli/dtos"
)

func parentWorkPackage(id uint64) (*dtos.LinkDto, error) {
	response, err := requests.Get(paths.WorkPackage(id), nil)
	if err != nil {
		return nil, err
	}

	workPackage := parser.Parse[dtos.WorkPackageDto](response)
	if workPackage.Links != nil && workPackage.Links.Self != nil {
		return workPackage.Links.Self, nil
	}

	return &dtos.LinkDto{Href: paths.WorkPackage(id)}, nil
}
