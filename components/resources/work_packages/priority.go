package work_packages

import (
	"strings"

	"github.com/opf/openproject-cli/components/common"
	"github.com/opf/openproject-cli/components/parser"
	"github.com/opf/openproject-cli/components/paths"
	"github.com/opf/openproject-cli/components/requests"
	"github.com/opf/openproject-cli/dtos"
	"github.com/opf/openproject-cli/models"
)

func availablePriorities() (PriorityDtos, error) {
	response, err := requests.Get(paths.Priorities(), nil)
	if err != nil {
		return nil, err
	}

	return parser.Parse[dtos.PriorityCollectionDto](response).Embedded.Elements, nil
}

func findPriority(input string, availablePriorities PriorityDtos) *dtos.PriorityDto {
	priorityAsId, priorityId := common.ParseId(input)

	var found []*dtos.PriorityDto
	for _, p := range availablePriorities {
		if priorityAsId && priorityId == uint64(p.Id) ||
			!priorityAsId && strings.ToLower(input) == strings.ToLower(p.Name) {
			found = append(found, p)
		}
	}

	if len(found) == 1 {
		return found[0]
	}

	return nil
}

type PriorityDtos []*dtos.PriorityDto

func (list PriorityDtos) Convert() []*models.Priority {
	return common.Reduce(list,
		func(acc []*models.Priority, item *dtos.PriorityDto) []*models.Priority {
			return append(acc, item.Convert())
		}, []*models.Priority{})
}
