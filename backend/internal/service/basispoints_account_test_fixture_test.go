package service

func basispointsAccountForTest() *Account {
	return &Account{
		ID: 300, Platform: PlatformOpenAI, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Concurrency: 10,
		Credentials: map[string]any{"access_token": "test-token", "chatgpt_account_id": "test-account"},
		Extra:       map[string]any{openAIOAuthResponsesEndpointExtraKey: openAIOAuthResponsesEndpointBasis},
	}
}
