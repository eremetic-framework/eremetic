//go:generate rm -vf assets/assets.go
//go:generate go-bindata-assetfs -pkg assets -o assets/assets.go ./static/... ./templates/...

package main
