package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/require"
	"itsTasty/pkg/api/adapters/dishRepo"
	"itsTasty/pkg/api/domain"
	"itsTasty/pkg/api/ports/botAPI"
	"itsTasty/pkg/api/ports/userAPI"
	"itsTasty/pkg/testutils"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestBasicVoteWorkflow(t *testing.T) {

	app, ts, cleanup, err := setupTestEnv()
	defer ts.Close()
	defer func() {
		if err := cleanup(); err != nil {
			t.Fatalf("Failed to cleanup test env : %v", err)
		}
	}()
	require.NoError(t, err)
	//
	// Start of actual test
	// 1) Create two dishes -> check creation
	// 2) Create two users -> check creation
	// 3) Both users vote for one dish -> check that dish has correct rating
	// In the last step, we send the check requests multiple times as there has been a weird transient
	// failure in this api endpoint (that should now be fixed with the sqlboiler postgres backend)
	// This is intended to keep this bug from re-appearing
	//

	botApiClient, err := botAPI.NewClientWithResponses(ts.URL+"/botAPI/v1/", botAPI.WithHTTPClient(ts.Client()))
	require.NoError(t, err)

	//
	//create two dishes new dishes
	//

	type testDish struct {
		name     string
		location string
		id       int64
	}
	dish1 := testDish{
		name:     "Test Dish 1",
		location: "Test Location 1",
	}
	dish2 := testDish{
		name:     "Test Dish 2",
		location: "Test Location 2",
	}

	createDish1Resp, err := botApiClient.PostCreateOrUpdateDishWithResponse(
		context.Background(),
		botAPI.PostCreateOrUpdateDishJSONRequestBody{
			DishName: dish1.name,
			ServedAt: dish1.location},
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-API-KEY", app.conf.botAPIToken)
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, createDish1Resp.StatusCode(), http.StatusOK)
	require.True(t, createDish1Resp.JSON200.CreatedNewLocation)
	require.True(t, createDish1Resp.JSON200.CreatedNewDish)

	dish1.id = createDish1Resp.JSON200.DishID

	createDish2Resp, err := botApiClient.PostCreateOrUpdateDishWithResponse(
		context.Background(),
		botAPI.PostCreateOrUpdateDishJSONRequestBody{
			DishName: dish2.name,
			ServedAt: dish2.location},
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-API-KEY", app.conf.botAPIToken)
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, createDish2Resp.StatusCode(), http.StatusOK)
	require.True(t, createDish2Resp.JSON200.CreatedNewLocation)
	require.True(t, createDish2Resp.JSON200.CreatedNewDish)
	dish2.id = createDish2Resp.JSON200.DishID

	//check that request with wrong api key fails
	wrongApiKeyReq, err := botApiClient.PostCreateOrUpdateDishWithResponse(
		context.Background(),
		botAPI.PostCreateOrUpdateDishJSONRequestBody{
			DishName: dish2.name,
			ServedAt: dish2.location},
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-API-KEY", "wrong key")
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, wrongApiKeyReq.StatusCode())

	//
	// Check that crated dishes exist
	//

	getDishResp, err := botApiClient.GetDishesDishIDWithResponse(
		context.Background(),
		dish1.id,
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-API-KEY", app.conf.botAPIToken)
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getDishResp.StatusCode())
	require.Equal(t, dish1.location, getDishResp.JSON200.ServedAt)

	getDishResp, err = botApiClient.GetDishesDishIDWithResponse(
		context.Background(),
		dish2.id,
		func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-API-KEY", app.conf.botAPIToken)
			return nil
		})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, getDishResp.StatusCode())
	require.Equal(t, dish2.location, getDishResp.JSON200.ServedAt)

	//
	// Interact with user API two create two votings for dish 1
	//

	//create two clients that are logged in as different users
	user1, err := newUserClient("testUser1@test.mail", ts)
	require.NoError(t, err)

	user2, err := newUserClient("testUser2@test.mail", ts)

	user1Dish1Voting := userAPI.RateDishReqRatingN3
	user2Dish1Voting := userAPI.RateDishReqRatingN5
	wantDish1Rating := float32(user1Dish1Voting+user2Dish1Voting) / 2

	userApiResp, err := user1.client.PostDishesDishIDWithResponse(
		context.Background(),
		dish1.id,
		userAPI.PostDishesDishIDJSONRequestBody{
			Rating: user1Dish1Voting,
		})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, userApiResp.StatusCode())

	userApiResp, err = user2.client.PostDishesDishIDWithResponse(
		context.Background(),
		dish1.id,
		userAPI.PostDishesDishIDJSONRequestBody{
			Rating: user2Dish1Voting,
		})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, userApiResp.StatusCode())

	//
	// Use bot api to check that the rating is correct
	//Do repeated accesses as this api path had transient errors previously
	//

	for _i := 0; _i < 100; _i++ {
		getDishResp, err = botApiClient.GetDishesDishIDWithResponse(context.Background(), dish1.id,
			func(ctx context.Context, req *http.Request) error {
				req.Header.Set("X-API-KEY", app.conf.botAPIToken)
				return nil
			})
		require.NoError(t, err)
		require.Equal(t, getDishResp.StatusCode(), http.StatusOK)
		require.Equal(t, wantDish1Rating, *getDishResp.JSON200.AvgRating)
		require.Equal(t, 2, len(getDishResp.JSON200.Ratings))

	}

	for _i := 0; _i < 100; _i++ {
		createOrUpdateResp, err := botApiClient.PostCreateOrUpdateDishWithResponse(context.Background(),
			botAPI.CreateOrUpdateDishReq{
				DishName: dish1.name,
				ServedAt: dish1.location,
			},
			func(ctx context.Context, req *http.Request) error {
				req.Header.Set("X-API-KEY", app.conf.botAPIToken)
				return nil
			})
		require.NoError(t, err)
		require.Equal(t, createOrUpdateResp.StatusCode(), http.StatusOK)
		require.False(t, createOrUpdateResp.JSON200.CreatedNewLocation)
		require.False(t, createOrUpdateResp.JSON200.CreatedNewDish)
		require.Equal(t, createOrUpdateResp.JSON200.DishID, dish1.id)
	}

}

func TestMergedDishCRUDOperations(t *testing.T) {
	app, ts, cleanup, err := setupTestEnv()
	defer ts.Close()
	defer func() {
		if err := cleanup(); err != nil {
			t.Fatalf("Failed to cleanup test env : %v", err)
		}
	}()
	require.NoError(t, err)

	//
	// Start of actual test
	// 1) Create four dishes. 3 at Location A, one at location B
	// 2) Create merged dish out of two of the dishes -> check creation
	// 3) Add third dish to merged dish
	// 4) Remove dish from merged dish
	// 5) Delete the merged dish
	//

	botApiClient, err := botAPI.NewClientWithResponses(ts.URL+"/botAPI/v1/", botAPI.WithHTTPClient(ts.Client()))
	require.NoError(t, err)

	//
	//create test dishes
	//

	type testDish struct {
		name     string
		location string
		id       int64
	}
	dish1L1 := testDish{
		name:     "Test Dish 1",
		location: "Test Location 1",
	}
	dish2L1 := testDish{
		name:     "Test Dish 2",
		location: "Test Location 1",
	}
	dish3L1 := testDish{
		name:     "Test Dish 3",
		location: "Test Location 1",
	}
	dish1L2 := testDish{
		name:     "Test Dish 4",
		location: "Test Location 2",
	}

	testDishes := []*testDish{&dish1L1, &dish2L1, &dish3L1, &dish1L2}

	for i, v := range testDishes {
		resp, err := botApiClient.PostCreateOrUpdateDishWithResponse(
			context.Background(),
			botAPI.PostCreateOrUpdateDishJSONRequestBody{
				DishName: v.name,
				ServedAt: v.location},
			func(ctx context.Context, req *http.Request) error {
				req.Header.Set("X-API-KEY", app.conf.botAPIToken)
				return nil
			})
		require.NoError(t, err)
		require.Equal(t, resp.StatusCode(), http.StatusOK)
		//for the first and the last test dish, a new location should be created
		require.Equal(t, i == 0 || i == 3, resp.JSON200.CreatedNewLocation)
		require.True(t, resp.JSON200.CreatedNewDish)

		v.id = resp.JSON200.DishID
	}

	//
	// Create merged dish via user api
	//
	user1, err := newUserClient("testUser1@test.mail", ts)
	require.NoError(t, err)

	wantMergedDishName := "Merged Dish"
	wantMergedDishIDsV1 := []int64{dish1L1.id, dish2L1.id}

	jsonResp, err := user1.client.PostMergedDishesWithResponse(context.Background(),
		userAPI.CreateMergedDishReq{
			MergedDishes: wantMergedDishIDsV1,
			Name:         wantMergedDishName,
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, jsonResp.StatusCode())
	mergedDishID := jsonResp.JSON200.MergedDishID

	//try to create merged dish with dishes from different locations. SHOULD FAIL
	resp, err := user1.client.PostMergedDishes(context.Background(),
		userAPI.CreateMergedDishReq{
			Name:         "Malformed merged dish with dishes from different locations",
			MergedDishes: []int64{dish1L1.id, dish1L2.id},
		},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	//Add dish3L1 to merged dish
	resp, err = user1.client.PatchMergedDishesMergedDishID(context.Background(), mergedDishID,
		userAPI.MergedDishUpdateReq{
			AddDishIDs:    &[]int64{dish3L1.id},
			RemoveDishIDs: nil,
		})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	//Check that dish was actually added
	mergedDishResp, err := user1.client.GetMergedDishesMergedDishIDWithResponse(context.Background(), mergedDishID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, mergedDishResp.StatusCode())
	require.ElementsMatch(t, []int64{dish1L1.id, dish2L1.id, dish3L1.id}, mergedDishResp.JSON200.ContainedDishIDs)

	//Remove dish1L1 from merged dish
	resp, err = user1.client.PatchMergedDishesMergedDishID(context.Background(), mergedDishID,
		userAPI.MergedDishUpdateReq{
			AddDishIDs:    nil,
			RemoveDishIDs: &[]int64{dish1L1.id},
		})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	//Check that dish was actually removed
	mergedDishResp, err = user1.client.GetMergedDishesMergedDishIDWithResponse(context.Background(), mergedDishID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, mergedDishResp.StatusCode())
	require.ElementsMatch(t, []int64{dish2L1.id, dish3L1.id}, mergedDishResp.JSON200.ContainedDishIDs)

	//Delete the merged dish
	resp, err = user1.client.DeleteMergedDishesMergedDishID(context.Background(), mergedDishID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	//Check that the merged dish was actually removed
	resp, err = user1.client.GetMergedDishesMergedDishID(context.Background(), mergedDishID)
	require.NoError(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

type testUser struct {
	Email  string
	client *userAPI.ClientWithResponses
}

// newUserClient is a helper function that creates a client that is logged in as the given user
func newUserClient(userEmail string, server *httptest.Server) (*testUser, error) {
	certpool := x509.NewCertPool()
	certpool.AddCert(server.Certificate())

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar : %v", err)
	}

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certpool,
			},
		},
		Jar: jar,
	}

	apiClient, err := userAPI.NewClientWithResponses(server.URL+"/userAPI/v1", userAPI.WithHTTPClient(c))
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate userAPI client : %v", err)
	}

	loginURL, err := url.Parse(server.URL + "/authAPI/login")
	if err != nil {
		return nil, fmt.Errorf("failed to parse login url : %v", err)
	}
	q := loginURL.Query()
	q.Add("userEmail", userEmail)
	q.Add("redirectTo", server.URL)
	loginURL.RawQuery = q.Encode()
	resp, err := c.Get(loginURL.String())
	if err != nil {
		return nil, fmt.Errorf("failed to perform login request : %v", err)
	}
	//this is a bit hacky. Our request gets redirected to the frontend during login.
	//However, in the test environment we do not serve the frontend. Thus we expect a 404 here

	if resp.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("login failed : %v", resp.StatusCode)
	}

	//To counteract the uncertainty from the hack 404 return code, we use the GetUsersMe endpoint
	//to check that we are actually logged in
	apiResp, err := apiClient.GetUsersMeWithResponse(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to check if login was successful : %v", err)
	}
	if apiResp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("login was not successful : %v", err)
	}
	if apiResp.JSON200.Email != userEmail {
		return nil, fmt.Errorf("endpoint GetUsersMe returned user %v instead of expected user %v",
			apiResp.JSON200.Email, userEmail)
	}

	return &testUser{
		Email:  userEmail,
		client: apiClient,
	}, nil
}

// setupTestEnv prepares a new test environment. The caller must call the returned cleanup function
// to terminate/free the resources allocated by this function. The caller must close ts once they are done with it
func setupTestEnv() (app *application, ts *httptest.Server, cleanupFN func() error, err error) {
	//
	//instantiate db backend
	//

	db, err := testutils.GlobalDockerPool.GetPostgresIntegrationTestDB()
	if err != nil {
		err = fmt.Errorf("getPostgresIntegrationTestDB failed : %v", err)
		return
	}

	dockerCleanupFN := func() error {
		if err := testutils.GlobalDockerPool.Cleanup(); err != nil {
			return fmt.Errorf("failed to cleanup docker pool : %v", err)
		}
		return nil
	}
	defer func() {
		if err != nil {
			if cleanupErr := dockerCleanupFN(); cleanupErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to cleanup docker resources : %v", err))
			}
		}
	}()

	migrationSource := &migrate.FileMigrationSource{Dir: "../../migrations/postgres"}
	repo, err := dishRepo.NewPostgresRepo(db, migrationSource)
	if err != nil {
		err = fmt.Errorf("NewPostgresRepo failed : %v", err)
		return

	}

	dbFactory := func() (domain.DishRepo, error) {
		return repo, nil
	}

	repoCleanupFN := func() error {
		if err := repo.DropRepo(context.Background()); err != nil {
			return fmt.Errorf("failed to drop repo : %v", err)
		}
		return nil
	}
	defer func() {
		if err != nil {
			if cleanupErr := repoCleanupFN(); cleanupErr != nil {
				err = errors.Join(err, fmt.Errorf("failed to cleanup db repo : %v", err))
			}
		}
	}()
	//
	// Prepare config
	//

	config := config{
		urlAfterLogin:   "https://localhost/welcome",
		urlAfterLogout:  "https://localhost/login",
		botAPIToken:     "testBotApiToken",
		sessionSecret:   "testSessionSecret",
		devMode:         true,
		devCORS:         "https://localhost",
		sessionLifetime: 10 * time.Minute,
	}

	app, err = newApplication(&config, dbFactory)
	if err != nil {
		err = fmt.Errorf("failed to istantiate app : %v", err)
		return
	}

	//
	// Setup go http test server
	//

	ts = httptest.NewTLSServer(app.router)

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create cookiejar : %v", err)
	}
	ts.Client().Jar = jar

	cleanupFN = func() error {
		repoErr := repoCleanupFN()
		dockerErr := dockerCleanupFN()
		return errors.Join(repoErr, dockerErr)
	}
	return app, ts, cleanupFN, nil
}
