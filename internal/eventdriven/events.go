package eventdriven

type ProjectEvent struct {
	UUID             string      `json:"uuid"`
	Name             string      `json:"name"`
	IsTemplate       bool        `json:"is_template"`
	UserEmail        string      `json:"user_email"`
	DateFormat       string      `json:"date_format"`
	TemplateTypeUUID interface{} `json:"template_type_uuid"`
	Timezone         string      `json:"timezone"`
	OrganizationID   int         `json:"organization_id"`
	ExtraFields      struct {
	} `json:"extra_fields"`
	Authorizations []struct {
		UserEmail string `json:"user_email"`
		Role      int    `json:"role"`
	} `json:"authorizations"`
	Description      string `json:"description"`
	OrganizationUUID string `json:"organization_uuid"`
	BrainOn          bool   `json:"brain_on"`
}

type PermissionEvent struct {
	Action  string `json:"action"`
	Project string `json:"project"`
	User    string `json:"user"`
	Role    int    `json:"role"`
}
