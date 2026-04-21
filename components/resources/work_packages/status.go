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

func availableStatuses() (StatusDtos, error) {
	response, err := requests.Get(paths.Status(), nil)
	if err != nil {
		return nil, err
	}

	return parser.Parse[dtos.StatusCollectionDto](response).Embedded.Elements, nil
}

func findStatus(input string, availableStatuses StatusDtos) *dtos.StatusDto {
	statusAsId, statusId := common.ParseId(input)

	var found []*dtos.StatusDto
	for _, s := range availableStatuses {
		if statusAsId && statusId == s.Id ||
			!statusAsId && strings.ToLower(input) == strings.ToLower(s.Name) {
			found = append(found, s)
		}
	}

	if len(found) == 1 {
		return found[0]
	}

	return nil
}

type StatusDtos []*dtos.StatusDto

func (list StatusDtos) Convert() []*models.Status {
	return common.Reduce(list,
		func(acc []*models.Status, item *dtos.StatusDto) []*models.Status {
			return append(acc, item.Convert())
		}, []*models.Status{})
}
