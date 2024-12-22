package service

import (
	"testing"

	"github.com/guisithos/go-ride-names/internal/strava"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStravaClient is a mock implementation of the Strava client
type MockStravaClient struct {
	mock.Mock
}

func (m *MockStravaClient) GetActivity(id int64) (*strava.Activity, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*strava.Activity), args.Error(1)
}

func (m *MockStravaClient) UpdateActivity(id int64, name string) error {
	args := m.Called(id, name)
	return args.Error(0)
}

func (m *MockStravaClient) GetAuthenticatedAthlete() (*strava.Athlete, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*strava.Athlete), args.Error(1)
}

func (m *MockStravaClient) GetAthleteActivities(page, perPage int, before, after int64) ([]strava.Activity, error) {
	args := m.Called(page, perPage, before, after)
	return args.Get(0).([]strava.Activity), args.Error(1)
}

func TestActivityService_RenameActivity(t *testing.T) {
	tests := []struct {
		name          string
		activityID    int64
		mockActivity  *strava.Activity
		mockError     error
		updateError   error
		expectedError bool
	}{
		{
			name:       "successful rename of default activity",
			activityID: 123,
			mockActivity: &strava.Activity{
				ID:        123,
				Name:      "Morning Run",
				SportType: "Run",
			},
			mockError:     nil,
			updateError:   nil,
			expectedError: false,
		},
		{
			name:       "skip non-default activity name",
			activityID: 456,
			mockActivity: &strava.Activity{
				ID:        456,
				Name:      "Epic Trail Run",
				SportType: "Run",
			},
			mockError:     nil,
			updateError:   nil,
			expectedError: false,
		},
		{
			name:          "error getting activity",
			activityID:    789,
			mockActivity:  nil,
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(MockStravaClient)

			// Setup mock expectations
			mockClient.On("GetActivity", tt.activityID).Return(tt.mockActivity, tt.mockError)
			if tt.mockActivity != nil && defaultActivityNames[tt.mockActivity.Name] {
				mockClient.On("UpdateActivity", tt.activityID, mock.AnythingOfType("string")).Return(tt.updateError)
			}

			// Create service with mock client
			service := NewActivityService(mockClient)

			// Execute test
			err := service.RenameActivity(tt.activityID)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify all expectations were met
			mockClient.AssertExpectations(t)
		})
	}
}

func TestActivityService_ListActivities(t *testing.T) {
	tests := []struct {
		name           string
		page           int
		perPage        int
		before         int64
		after          int64
		updateNames    bool
		mockActivities []strava.Activity
		mockError      error
		expectedError  bool
	}{
		{
			name:    "successful listing without name updates",
			page:    1,
			perPage: 10,
			mockActivities: []strava.Activity{
				{ID: 1, Name: "Morning Run", SportType: "Run"},
				{ID: 2, Name: "Epic Trail Run", SportType: "Run"},
			},
			updateNames:   false,
			mockError:     nil,
			expectedError: false,
		},
		{
			name:    "successful listing with name updates",
			page:    1,
			perPage: 10,
			mockActivities: []strava.Activity{
				{ID: 1, Name: "Morning Run", SportType: "Run"},
				{ID: 2, Name: "Epic Trail Run", SportType: "Run"},
			},
			updateNames:   true,
			mockError:     nil,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(MockStravaClient)

			mockClient.On("GetAthleteActivities",
				tt.page, tt.perPage, tt.before, tt.after).Return(tt.mockActivities, tt.mockError)

			if tt.updateNames {
				for _, activity := range tt.mockActivities {
					if defaultActivityNames[activity.Name] {
						mockClient.On("UpdateActivity",
							activity.ID, mock.AnythingOfType("string")).Return(nil)
					}
				}
			}

			service := NewActivityService(mockClient)
			activities, err := service.ListActivities(tt.page, tt.perPage, tt.before, tt.after, tt.updateNames)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, activities)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.mockActivities), len(activities))
			}

			mockClient.AssertExpectations(t)
		})
	}
}
