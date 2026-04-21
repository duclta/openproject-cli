package work_packages

import (
	"bytes"
	"encoding/json"

	"github.com/opf/openproject-cli/components/parser"
	"github.com/opf/openproject-cli/components/paths"
	"github.com/opf/openproject-cli/components/requests"
	"github.com/opf/openproject-cli/dtos"
	"github.com/opf/openproject-cli/models"
)

func Comment(id uint64, text string) (*models.Activity, error) {
	activity := dtos.ActivityDto{
		Comment: &dtos.LongTextDto{
			Format: "markdown",
			Raw:    text,
		},
	}

	data, err := json.Marshal(activity)
	if err != nil {
		return nil, err
	}

	requestData := requests.RequestData{ContentType: "application/json", Body: bytes.NewReader(data)}
	response, err := requests.Post(paths.WorkPackageActivities(id), &requestData)
	if err != nil {
		return nil, err
	}

	resultingActivity := parser.Parse[dtos.ActivityDto](response)
	converted, err := resultingActivity.Convert()
	if err != nil {
		return nil, err
	}

	return converted, nil
}
