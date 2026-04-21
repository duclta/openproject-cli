package printer

import (
	"fmt"

	"github.com/opf/openproject-cli/models"
)

func Priorities(priorities []*models.Priority) {
	for _, p := range priorities {
		printPriority(p)
	}
}

func printPriority(priority *models.Priority) {
	id := fmt.Sprintf("#%d", priority.Id)
	activePrinter.Printf("[%s] %s\n", Red(id), Cyan(priority.Name))
}
