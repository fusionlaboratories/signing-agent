package defs

import (
	"fmt"
)

func URLAPIKey(baseURL, workspaceID, APIkeyID string) string {
	return fmt.Sprintf("%s/workspaces/%s/apikeys/%s", baseURL, workspaceID, APIkeyID)
}

func URLToken(baseURL, workspaceID string) string {
	return fmt.Sprintf("%s/workspaces/%s/token", baseURL, workspaceID)
}

func URLTokenRefresh(baseURL, workspaceID string) string {
	return fmt.Sprintf("%s/workspaces/%s/token/refresh", baseURL, workspaceID)
}

func URLAction(baseURL, actionID string) string {
	return fmt.Sprintf("%s/actions/%s", baseURL, actionID)
}

func URLlocalFeed(httpAddr string) string {
	return fmt.Sprintf("ws://%s%s/client/feed", httpAddr, PathPrefix)
}
