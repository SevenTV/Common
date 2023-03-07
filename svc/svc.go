package svc

// AppIdentity is used to infer data from the app currently running
type AppIdentity struct {
	Name string
	Web  string
	CDN  string
}
