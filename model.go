package main

type apiOrganisation struct {
	APIURL                 string `json:"apiUrl"`
	ID                     string `json:"id"`
	IndustryClassification struct {
		APIURL    string `json:"apiUrl"`
		ID        string `json:"id"`
		PrefLabel string `json:"prefLabel"`
	} `json:"industryClassification"`
	Labels      []string `json:"labels"`
	LeiCode     string   `json:"leiCode"`
	Memberships []struct {
		Person struct {
			APIURL    string   `json:"apiUrl"`
			ID        string   `json:"id"`
			PrefLabel string   `json:"prefLabel"`
			Types     []string `json:"types"`
		} `json:"person"`
		Title string `json:"title"`
	} `json:"memberships"`
	PrefLabel    string `json:"prefLabel"`
	Profile      string `json:"profile"`
	Subsidiaries []struct {
		APIURL    string   `json:"apiUrl"`
		ID        string   `json:"id"`
		PrefLabel string   `json:"prefLabel"`
		Types     []string `json:"types"`
	} `json:"subsidiaries"`
	Types []string `json:"types"`
}
