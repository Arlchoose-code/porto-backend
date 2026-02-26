package structs

type SettingUpsertRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}
