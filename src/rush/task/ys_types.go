package task

type YsProduct struct {
	AttributeList struct {
		IsWaitingRoomProduct     bool      `json:"isWaitingRoomProduct"`
		PreviewTo                string `json:"preview_to"`
	} `json:"attribute_list"`
}