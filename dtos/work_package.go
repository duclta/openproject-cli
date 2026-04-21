package dtos

import (
	"github.com/opf/openproject-cli/models"
)

type WorkPackageLinksDto struct {
	Self              *LinkDto   `json:"self,omitempty"`
	AddAttachment     *LinkDto   `json:"addAttachment,omitempty"`
	Status            *LinkDto   `json:"status,omitempty"`
	Project           *LinkDto   `json:"project,omitempty"`
	Assignee          *LinkDto   `json:"assignee,omitempty"`
	Type              *LinkDto   `json:"type,omitempty"`
	Priority          *LinkDto   `json:"priority,omitempty"`
	Parent            *LinkDto   `json:"parent,omitempty"`
	Version           *LinkDto   `json:"version,omitempty"`
	CustomActions     []*LinkDto `json:"customActions,omitempty"`
	PrepareAttachment *LinkDto   `json:"prepareAttachment,omitempty"`
}

type WorkPackageDto struct {
	Id          int64                `json:"id,omitempty"`
	Subject     string               `json:"subject,omitempty"`
	Links       *WorkPackageLinksDto `json:"_links,omitempty"`
	Description *LongTextDto         `json:"description,omitempty"`
	StartDate   *string              `json:"startDate"`
	DueDate     *string              `json:"dueDate"`
	Embedded    *embeddedDto         `json:"_embedded,omitempty"`
	LockVersion int                  `json:"lockVersion"`
}

type embeddedDto struct {
	CustomActions []*CustomActionDto `json:"customActions"`
}

type workPackageElements struct {
	Elements []*WorkPackageDto `json:"elements"`
}

type WorkPackageCollectionDto struct {
	Embedded workPackageElements `json:"_embedded"`
	Type     string              `json:"_type"`
	Total    int64               `json:"total"`
	Count    int64               `json:"count"`
	PageSize int64               `json:"pageSize"`
	Offset   int64               `json:"offset"`
}

type CreateWorkPackageDto struct {
	Subject string `json:"subject"`
}

/////////////// MODEL CONVERSION ///////////////

func (dto *WorkPackageDto) Convert() *models.WorkPackage {
	wp := &models.WorkPackage{
		Id:          uint64(dto.Id),
		Subject:     dto.Subject,
		Description: dto.Description.Raw,
		LockVersion: dto.LockVersion,
	}

	if dto.StartDate != nil {
		wp.StartDate = *dto.StartDate
	}
	if dto.DueDate != nil {
		wp.DueDate = *dto.DueDate
	}

	if dto.Links != nil {
		if dto.Links.Type != nil {
			wp.Type = dto.Links.Type.Title
		}
		if dto.Links.Assignee != nil {
			wp.Assignee = dto.Links.Assignee.Title
		}
		if dto.Links.Status != nil {
			wp.Status = dto.Links.Status.Title
		}
		if dto.Links.Priority != nil {
			wp.Priority = dto.Links.Priority.Title
		}
		if dto.Links.Parent != nil {
			wp.Parent = dto.Links.Parent.Title
		}
		if dto.Links.Version != nil {
			wp.Version = dto.Links.Version.Title
		}
	}

	return wp
}

func (dto *WorkPackageCollectionDto) Convert() *models.WorkPackageCollection {
	var workPackages = make([]*models.WorkPackage, len(dto.Embedded.Elements))

	for idx, p := range dto.Embedded.Elements {
		workPackages[idx] = p.Convert()
	}

	return &models.WorkPackageCollection{
		Total:    dto.Total,
		Count:    dto.Count,
		PageSize: dto.PageSize,
		Offset:   dto.Offset,
		Items:    workPackages,
	}
}
