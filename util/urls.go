package util

import (
	"fmt"
)

func URLClientInit(baseURL string) string {
	return fmt.Sprintf("%s/coreclient/init", baseURL)
}

func URLRegisterConfirm(baseURL string) string {
	return fmt.Sprintf("%s/coreclient/finish", baseURL)
}

func URLActionMessages(baseURL, actionID string) string {
	return fmt.Sprintf("%s/coreclient/action/%s", baseURL, actionID)
}

func URLActionApprove(baseURL, actionID string) string {
	return fmt.Sprintf("%s/coreclient/action/%s", baseURL, actionID)
}

func URLActionReject(baseURL, actionID string) string {
	return fmt.Sprintf("%s/coreclient/action/%s", baseURL, actionID)
}
