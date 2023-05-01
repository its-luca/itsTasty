package vacation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"itsTasty/pkg/api/domain"
	"log"
	"net/http"
	"net/url"
	"path"
	"time"
)

// status describes the state of the vacation request in the booking system
type status struct {
	Status string
}

type personVacationSpan struct {
	Name   string
	Start  time.Time
	End    time.Time
	Status status
}

type getVacationArgs struct {
	StartTime time.Time
	EndTime   time.Time
}

// getVacationResp maps each user to their vacation spans
type getVacationResp map[string][]personVacationSpan

var (
	//Confirmed means that vacation was booked successfully
	confirmed = status{"confirmed"}
	//Planned means that the person intends to take vacation but has not booked it yet
	planned = status{"planned"}
	//Pending means that the vacation has been booked but not approved
	pending = status{"pending"}
)

type UniversityVacationClient struct {
	baseURL *url.URL
	apiKey  string
}

func NewUniversityVacationClient(baseURL, apiKey string) (*UniversityVacationClient, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse baseURL : %v", err)
	}
	return &UniversityVacationClient{baseURL: parsedURL, apiKey: apiKey}, nil
}

// Vacations returns the vacations for the given da which may not be in the past
func (u *UniversityVacationClient) Vacations(ctx context.Context, day domain.DayPrecisionTime) (domain.UsersOnVacation, error) {

	args := getVacationArgs{
		StartTime: day.Time,
		EndTime:   day.Time,
	}

	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(args); err != nil {
		return domain.UsersOnVacation{}, fmt.Errorf("failed to encode args to json : %w", err)
	}

	endpointURL := *u.baseURL
	endpointURL.Path = path.Join(u.baseURL.Path, "getVacations")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointURL.String(), reqBody)
	if err != nil {
		return domain.UsersOnVacation{}, fmt.Errorf("failed to create request object : %w", err)
	}
	req.Header.Set("Authorization", u.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return domain.UsersOnVacation{}, fmt.Errorf("request failed : %w", err)
	}

	rawResp, err := io.ReadAll(resp.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close resp body : %v", err)
		}
	}(resp.Body)
	if err != nil {
		return domain.UsersOnVacation{}, fmt.Errorf("failed to read response : %w", err)
	}
	vacationData := make(getVacationResp)
	if err := json.NewDecoder(bytes.NewReader(rawResp)).Decode(&vacationData); err != nil {
		return domain.UsersOnVacation{}, fmt.Errorf("failed to decode response \"%s\" : %w", rawResp, err)
	}

	result := make(map[domain.DayPrecisionTime]map[string]interface{})
	if len(vacationData) == 0 {
		return domain.NewUsersOnVacation(result), nil
	}

	//convert api response to internal data structure
	//api: users -> vacation spans
	//internal: date -> users with vacation
	for user, userVacations := range vacationData {
		for _, vacation := range userVacations {
			if vacation.Status != confirmed {
				continue
			}
			current := domain.NewDayPrecisionTime(vacation.Start)
			endDate := domain.NewDayPrecisionTime(vacation.End)
			for !current.After(endDate.Time) {
				if _, ok := result[current]; !ok {
					result[current] = make(map[string]interface{})
				}
				result[current][user] = nil
				current = current.NextDay()
			}
		}
	}

	return domain.NewUsersOnVacation(result), nil

}
