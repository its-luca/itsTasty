package domain

import (
	"errors"
)

var ErrNotOnSameLocation = errors.New("condensed dishes not on same location")

type MergedDish struct {
	Name string
	//condensedDishNames are the (unique) names of the dishes condensed in this MergedDish
	condensedDishNames []string

	//ServedAt is the location where the MergedDish is served
	ServedAt string
}

func (m *MergedDish) GetCondensedDishNames() []string {
	return m.condensedDishNames
}

func NewMergedDish(name string, dish1 *Dish, dish2 *Dish, additionalDishes []*Dish) (*MergedDish, error) {

	//gather all dish names in slice and ensure that they are all served at the same location
	condensedDishNames := make([]string, 0, 2+len(additionalDishes))
	if dish1.ServedAt != dish2.ServedAt {
		return nil, ErrNotOnSameLocation
	}
	condensedDishNames = append(condensedDishNames, dish1.Name)
	condensedDishNames = append(condensedDishNames, dish2.Name)
	for _, v := range additionalDishes {
		if dish1.ServedAt != v.ServedAt {
			return nil, ErrNotOnSameLocation
		}
		condensedDishNames = append(condensedDishNames, v.Name)
	}

	return &MergedDish{
		Name:               name,
		condensedDishNames: condensedDishNames,
		ServedAt:           dish1.ServedAt,
	}, nil
}

func NewMergedDishFomDB(name, servedAt string, condensedDishNames []string) *MergedDish {
	return &MergedDish{
		Name:               name,
		ServedAt:           servedAt,
		condensedDishNames: condensedDishNames,
	}
}
