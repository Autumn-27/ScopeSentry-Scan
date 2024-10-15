// wayback-------------------------------------
// @file      : types.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/10/14 20:25
// -------------------------------------------

package source

type Result struct {
	URL    string
	Source string
}

type IndexResponse struct {
	ID     string `json:"id"`
	APIURL string `json:"cdx-api"`
}

type AlienvaultResponse struct {
	URLList []url `json:"url_list"`
	HasNext bool  `json:"has_next"`
}

type url struct {
	URL string `json:"url"`
}
