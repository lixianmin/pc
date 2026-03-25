package engine

import "github.com/lixianmin/pc/baml_client/types"

type ChatResult = types.Union8BashOrChatResponseOrEditOrReadOrUseSkillOrWebFetchOrWebSearchOrWrite

func defaultInt(num *int64, defaultValue int) int {
	if num == nil {
		return defaultValue
	}

	return int(*num)
}
