package api

// DeviceRequestResponse is the response for POST /api/auth/device-request
type DeviceRequestResponse struct {
	DeviceCode  string `json:"device_code"`
	UserCode    string `json:"user_code"`
	ApprovalURL string `json:"approval_url"`
	ExpiresIn   int    `json:"expires_in"`
	Interval    int    `json:"interval"`
}

// DeviceApproveRequest is the request body for POST /api/auth/device-approve
type DeviceApproveRequest struct {
	DeviceCode string `json:"device_code"`
	Action     string `json:"action"`
}

// DeviceApproveStatusResponse is the response for GET /api/auth/device-approve
type DeviceApproveStatusResponse struct {
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
	Status     string `json:"status"`
	ExpiresAt  int64  `json:"expires_at"`
}

// DeviceTokenResponse is the response for GET /api/auth/device-token
type DeviceTokenResponse struct {
	Status string `json:"status"`
	Token  string `json:"token,omitempty"`
}
