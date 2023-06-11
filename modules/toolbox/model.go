package toolbox

type SettingModel struct {
	Name  string `gorm:"primaryKey" json:"name"`
	Type  int    `json:"type"`
	Value string `json:"value"`
}

func (SettingModel) TableName() string {
	return "tbl_toolbox_setting"
}

type HealthModel struct {
	Url string
}

func (HealthModel) TableName() string {
	return "tbl_toolbox_health"
}

type SessionModel struct {
}

func (SessionModel) TableName() string {
	return "tbl_toolbox_session"
}
