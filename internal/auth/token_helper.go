package auth

import (
	"encoding/json"
	"fmt"
)

func UnmarshalTokens(tokensInterface interface{}) (*TokenResponse, error) {
	tokenData, err := json.Marshal(tokensInterface)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token data: %v", err)
	}

	var tokens TokenResponse
	if err := json.Unmarshal(tokenData, &tokens); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token data: %v", err)
	}

	return &tokens, nil
}
