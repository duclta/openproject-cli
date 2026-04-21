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

func availableVersions(projectId uint64) (VersionDtos, error) {
	response, err := requests.Get(paths.ProjectVersions(projectId), nil)
	if err != nil {
		return nil, err
	}

	return parser.Parse[dtos.VersionCollectionDto](response).Embedded.Elements, nil
}

func findVersion(input string, availableVersions VersionDtos) *dtos.VersionDto {
	versionAsId, versionId := common.ParseId(input)

	var found []*dtos.VersionDto
	for _, v := range availableVersions {
		if versionAsId && versionId == uint64(v.Id) ||
			!versionAsId && strings.ToLower(input) == strings.ToLower(v.Name) {
			found = append(found, v)
		}
	}

	if len(found) == 1 {
		return found[0]
	}

	return nil
}

type VersionDtos []*dtos.VersionDto

func (list VersionDtos) Convert() []*models.Version {
	return common.Reduce(list,
		func(acc []*models.Version, item *dtos.VersionDto) []*models.Version {
			return append(acc, item.Convert())
		}, []*models.Version{})
}
