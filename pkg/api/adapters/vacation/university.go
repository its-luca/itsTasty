package vacation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Status describes the state of the vacation request in the booking system
type Status struct {
	Status string
}

type PersonVacationSpan struct {
	Name   string
	Start  time.Time
	End    time.Time
	Status Status
}

type GetVacationResp = map[string][]PersonVacationSpan

var (
	//Confirmed means that vacation was booked successfully
	Confirmed = Status{"confirmed"}
	//Planned means that the person intends to take vacation but has not booked it yet
	Planned = Status{"planned"}
	//Pending means that the vacation has been booked but not approved
	Pending = Status{"pending"}
)

type UniversityVacationClient struct {
	url string
}

// GetVacation returns the vacations in the given timespan. Start must be at the current day or in the future
func (u *UniversityVacationClient) GetVacation(ctx context.Context, start, end time.Time) (GetVacationResp, error) {
	type getVacationArgs struct {
		StartTime time.Time
		EndTime   time.Time
	}

	args := getVacationArgs{
		StartTime: start,
		EndTime:   end,
	}

	reqBody := &bytes.Buffer{}
	if err := json.NewEncoder(reqBody).Encode(args); err != nil {
		return nil, fmt.Errorf("failed to encode args to json : %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request object : %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed : %w", err)
	}

	rawResp, err := io.ReadAll(resp.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Failed to close resp body : %v", err)
		}
	}(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response : %w", err)
	}
	vacationData := make(GetVacationResp, 0)
	if err := json.NewDecoder(bytes.NewReader(rawResp)).Decode(&vacationData); err != nil {
		return nil, fmt.Errorf("failed to decode response \"%s\" : %w", rawResp, err)
	}

	return vacationData, nil

}
