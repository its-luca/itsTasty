package domain

import (
	"errors"
	"fmt"
)

var ErrNotOnSameLocation = errors.New("condensed dishes not on same location")
var ErrDishAlreadyMerged = errors.New("dish already part of merged dish")
var ErrDishNotPartOfMergedDish = errors.New("dish not part of merged dish")
var ErrMergedDishNeedsAtLeastTwoDishes = errors.New("merged dish needs at least two dishes")

type MergedDish struct {
	Name string
	//condensedDishNames are the (unique) names of the dishes condensed in this MergedDish
	condensedDishNames map[string]interface{}

	//ServedAt is the location where the MergedDish is served
	ServedAt string
}

func (m *MergedDish) GetCondensedDishNames() []string {
	r := make([]string, 0, len(m.condensedDishNames))
	for k := range m.condensedDishNames {
		r = append(r, k)
	}
	return r
}

// AddDish
// may return ErrNotOnSameLocation, ErrDishAlreadyMerged
func (m *MergedDish) AddDish(d *Dish) error {
	if d.ServedAt != m.ServedAt {
		return ErrNotOnSameLocation
	}
	if _, ok := m.condensedDishNames[d.Name]; ok {
		return ErrDishAlreadyMerged
	}

	m.condensedDishNames[d.Name] = nil

	return nil
}

// RemoveDish
// may return ErrDishNotPartOfMergedDish, ErrMergedDishNeedsAtLeastTwoDishes
func (m *MergedDish) RemoveDish(d *Dish) error {
	if d.ServedAt != m.ServedAt {
		return fmt.Errorf("%w : %v", ErrDishNotPartOfMergedDish, ErrNotOnSameLocation)
	}
	if _, ok := m.condensedDishNames[d.Name]; !ok {
		return ErrDishNotPartOfMergedDish
	}

	if len(m.condensedDishNames) == 2 {
		return ErrMergedDishNeedsAtLeastTwoDishes
	}

	delete(m.condensedDishNames, d.Name)

	return nil
}

func (m *MergedDish) DeepCopy() *MergedDish {
	rCDN := make(map[string]interface{}, len(m.condensedDishNames))
	for k := range m.condensedDishNames {
		rCDN[k] = nil
	}
	r := &MergedDish{
		Name:               m.Name,
		condensedDishNames: rCDN,
		ServedAt:           m.ServedAt,
	}
	return r
}

// NewMergedDish
// Returns ErrNotOnSameLocation if provided dishes are not on the same location
func NewMergedDish(name string, dish1 *Dish, dish2 *Dish, additionalDishes []*Dish) (*MergedDish, error) {

	//gather all dish names in slice and ensure that they are all served at the same location
	condensedDishNames := make(map[string]interface{}, 2+len(additionalDishes))
	if dish1.ServedAt != dish2.ServedAt {
		return nil, ErrNotOnSameLocation
	}
	condensedDishNames[dish1.Name] = nil
	condensedDishNames[dish2.Name] = nil

	mergedDish := &MergedDish{
		Name:               name,
		condensedDishNames: condensedDishNames,
		ServedAt:           dish1.ServedAt,
	}
	for _, v := range additionalDishes {
		if err := mergedDish.AddDish(v); err != nil {
			return nil, fmt.Errorf("cannot use dish %v : %w", *v, err)
		}
	}

	return &MergedDish{
		Name:               name,
		condensedDishNames: condensedDishNames,
		ServedAt:           dish1.ServedAt,
	}, nil
}

func NewMergedDishFomDB(name, servedAt string, condensedDishNames map[string]interface{}) *MergedDish {
	return &MergedDish{
		Name:               name,
		ServedAt:           servedAt,
		condensedDishNames: condensedDishNames,
	}
}
