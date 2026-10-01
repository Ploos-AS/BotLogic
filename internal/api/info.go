package api

const (
	Service = "botlogic"
	Version = "1.0.0"
	APIVersion = "v1"
)

type VersionInfo struct {
	Service string `json:"service"`
	Version string `json:"version"`
	API string `json:"api"`
}

func VersionResponse() VersionInfo { return VersionInfo{Service:Service,Version:Version,API:APIVersion} }
