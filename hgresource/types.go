package main

type Source struct {
	Uri string `json:"uri"`
	PrivateKey string `json:"private_key"`
	IncludePaths []string `json:"paths"`
	ExcludePaths []string `json:"ignore_paths"`
	Branch string `json:"branch"`
}

type Version struct {
	Ref string `json:"ref"`
}