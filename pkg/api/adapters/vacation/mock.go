package vacation

import (
	"context"
	"itsTasty/pkg/api/domain"
)

type MockVacationClient struct {
	vacations map[domain.DayPrecisionTime]map[string]interface{}
}

func NewMockVacationClient(dateToPeopleOnVacations map[domain.DayPrecisionTime]map[string]interface{}) *MockVacationClient {
	return &MockVacationClient{vacations: dateToPeopleOnVacations}
}

func NewEmptyVacationClient() *MockVacationClient {
	return &MockVacationClient{vacations: make(map[domain.DayPrecisionTime]map[string]interface{})}
}

func (m MockVacationClient) Vacations(_ context.Context, day domain.DayPrecisionTime) (domain.UsersOnVacation, error) {
	data := map[domain.DayPrecisionTime]map[string]interface{}{day: m.vacations[day]}
	return domain.NewUsersOnVacation(data), nil
}
