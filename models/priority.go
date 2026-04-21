package models

type Priority struct {
	Id        uint64
	Name      string
	Position  int
	Color     string
	IsDefault bool
	IsActive  bool
}
