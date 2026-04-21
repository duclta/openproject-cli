package dtos

import (
	"github.com/opf/openproject-cli/components/common"
	"github.com/opf/openproject-cli/models"
)

type PriorityDto struct {
	Id       int64  `json:"id"`
	Name     string `json:"name"`
	Position int    `json:"position"`
	Color    string `json:"color"`
	IsDefault bool  `json:"isDefault"`
	IsActive bool   `json:"isActive"`
}

type priorityElements struct {
	Elements []*PriorityDto `json:"elements"`
}

type PriorityCollectionDto struct {
	Embedded *priorityElements `json:"_embedded"`
	Type     string            `json:"_type"`
}

/////////////// MODEL CONVERSION ///////////////

func (dto *PriorityDto) Convert() *models.Priority {
	return &models.Priority{
		Id:        uint64(dto.Id),
		Name:      dto.Name,
		Position:  dto.Position,
		Color:     dto.Color,
		IsDefault: dto.IsDefault,
		IsActive:  dto.IsActive,
	}
}

func (dto *PriorityCollectionDto) Convert() []*models.Priority {
	return common.Reduce(
		dto.Embedded.Elements,
		func(state []*models.Priority, p *PriorityDto) []*models.Priority {
			return append(state, p.Convert())
		},
		[]*models.Priority{},
	)
}
