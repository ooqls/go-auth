package audits

type Audit struct {
	Version int64  `json:"version"`
	UserID  string `json:"user_id"`

	ResourceORN string `json:"resource_orn"`
	Action      string `json:"action"`
}
