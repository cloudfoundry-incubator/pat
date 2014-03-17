package workloads

type TargetResponse struct {
	LoginEndpoint string `json:"authorization_endpoint"`
}

type LoginResponse struct {
	Token string `json:"access_token"`
}

type Metadata struct {
	Guid string `json:"guid"`
}

type Resource struct {
	Metadata Metadata `json:"metadata"`
}

type SpaceResponse struct {
	Resources []Resource `json:"resources"`
}
