package main

type apiIndustryClassification struct {
	APIURL    string `json:"apiUrl"`
	ID        string `json:"id"`
	PrefLabel string `json:"prefLabel"`
}

type apiSubsidiary struct {
	APIURL    string   `json:"apiUrl"`
	ID        string   `json:"id"`
	PrefLabel string   `json:"prefLabel"`
	Types     []string `json:"types"`
}

type apiMembership struct {
	Person struct {
		APIURL    string   `json:"apiUrl"`
		ID        string   `json:"id"`
		PrefLabel string   `json:"prefLabel"`
		Types     []string `json:"types"`
	} `json:"person"`
	Title string `json:"title"`
}

type apiOrganisation struct {
	APIURL                 string                     `json:"apiUrl"`
	ID                     string                     `json:"id"`
	IndustryClassification *apiIndustryClassification `json:"industryClassification,omitempty"`
	Labels                 []string                   `json:"labels"`
	LeiCode                string                     `json:"leiCode,omitempty"`
	Memberships            []apiMembership            `json:"memberships"`
	PrefLabel              string                     `json:"prefLabel"`
	Profile                string                     `json:"profile,omitempty"`
	Subsidiaries           []apiSubsidiary            `json:"subsidiaries",omitempty`
	Types                  []string                   `json:"types"`
}
