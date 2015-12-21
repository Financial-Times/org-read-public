package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jawher/mow.cli"
	"github.com/jmcvetta/neoism"
	"log"
	"net/http"
	"os"
)

func main() {

	app := cli.App("org-read-public", "A RESTful API for the public organisations read endpoint")
	neoURL := app.StringOpt("neo-url", "http://localhost:7474/db/data", "neo4j endpoint URL")
	port := app.IntOpt("port", 8080, "Port to listen on")

	app.Action = func() {
		runServer(*neoURL, *port)
	}

	app.Run(os.Args)
}

var db *neoism.Database

func runServer(neoURL string, port int) {

	var err error
	db, err = neoism.Connect(neoURL)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("connected to %s\n", neoURL)

	m := mux.NewRouter()
	http.Handle("/", m)

	m.HandleFunc("/organisations/{uuid}", getHandler).Methods("GET")

	log.Printf("listening on %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Printf("web stuff failed: %v\n", err)
	}

}

func getHandler(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uuid := vars["uuid"]
	if uuid == "" { //TODO: check this is a uuid.
		http.Error(w, "uuid invalid", http.StatusBadRequest)
		return
	}

	org, found, err := queryOrg(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	enc := json.NewEncoder(w)
	if err = enc.Encode(org); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

type resultSubsidiary struct {
	UUID      string   `json:"uuid"`
	PrefLabel string   `json:"prefLabel"`
	LeiCode   string   `json:"leiCode"`
	Labs      []string `json:"labs"`
}

type resultIndustry struct {
	UUID      string `json:"uuid"`
	PrefLabel string `json:"prefLabel"`
	IcbCode   string `json:"icbCode"`
}

type resultOrg struct {
	UUID        string   `json:"uuid"`
	PrefLabel   string   `json:"prefLabel"`
	HiddenLabel string   `json:"hiddenLabel"`
	LegalName   string   `json:"legalName"`
	ShortName   string   `json:"shortName"`
	LeiCode     string   `json:"leiCode"`
	Labs        []string `json:"labs"`
}

func queryOrg(uuid string) (org apiOrganisation, found bool, err error) {

	statement := `
		MATCH (org:Organisation {uuid: {uuid}})
		OPTIONAL MATCH (org)-[:SUB_ORG_OF]->(par:Organisation)
		OPTIONAL MATCH (org)-[:IN_INDUSTRY]->(ind:Industry)
		OPTIONAL MATCH (sub:Organisation)-[:SUB_ORG_OF]->(org)
		RETURN
		{uuid:org.uuid, prefLabel:org.prefLabel, leiCode:org.leiIdentifier, labs:labels(org), hiddenLabel:org.hiddenLabel,
				legalName:org.legalName, shortName:org.shortName} as organisation,
		{uuid:par.uuid, prefLabel:par.prefLabel, leiCode:par.leiIdentifier, labs:labels(par)} as parent,
		{uuid:ind.uuid, prefLabel:ind.prefLabel, icbCode:ind.icbCode} as industry,
		collect({uuid:sub.uuid,prefLabel:sub.prefLabel, leiCode:sub.leiIdentifier, labs:labels(sub)}) as subs
	`

	var result []struct {
		Industry resultIndustry     `json:"industry"`
		MainOrg  resultOrg          `json:"organisation"`
		Parent   resultOrg          `json:"parent"`
		Subs     []resultSubsidiary `json:"subs"`
	}

	err = db.Cypher(&neoism.CypherQuery{
		Statement: statement,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &result,
	})

	if err != nil {
		return
	}

	if len(result) != 1 {
		if len(result) != 0 {
			err = fmt.Errorf("Invalid result count %d", len(result))
		}
		return
	}

	r := result[0]

	org = apiOrganisation{
		APIURL:                 fmt.Sprintf("http://test.api.ft.com/organisations/%s", r.MainOrg.UUID),
		ID:                     fmt.Sprintf("http://api.ft.com/things/%s", r.MainOrg.UUID),
		PrefLabel:              r.MainOrg.PrefLabel,
		LeiCode:                r.MainOrg.LeiCode,
		IndustryClassification: mapIndustry(r.Industry),
		Labels:                 nil, // added later
		Memberships:            nil, // added later
		Profile:                "",  //TODO
		Subsidiaries:           mapSubsidiaries(r.Subs),
		Types:                  mapTypes(r.MainOrg.Labs),
	}

	//TODO: more labels. think about sorting?
	org.Labels = []string{
		r.MainOrg.PrefLabel,
		r.MainOrg.HiddenLabel,
		r.MainOrg.LegalName,
		r.MainOrg.ShortName,
	}

	org.Memberships, err = queryMemberships(uuid)
	if err != nil {
		return
	}

	found = true
	return
}

func queryMemberships(uuid string) (memberships []apiMembership, err error) {
	statement := `
		MATCH (o:Organisation{uuid: {uuid}})<-[emp:HAS_ORGANISATION]-(m:Membership)-[mem:HAS_MEMBER]->(p:Person)
		OPTIONAL MATCH (cont:Content)-[:MENTIONS]->(p), (cont)-[:MENTIONS]->(o)
		RETURN p.prefLabel as name, p.uuid as uuid, labels(p) as labs, m.prefLabel as title, count(cont) as count 
		order by count desc, name limit 1000
	`

	var result []struct {
		Name   string   `json:"name"`
		UUID   string   `json:"uuid"`
		Labels []string `json:"labs"`
		Title  string   `json:"title"`
		Count  int      `json:"count"`
	}

	err = db.Cypher(&neoism.CypherQuery{
		Statement: statement,
		Parameters: map[string]interface{}{
			"uuid": uuid,
		},
		Result: &result,
	})

	if err != nil {
		return
	}

	for _, res := range result {
		membership := apiMembership{}
		membership.Person.ID = fmt.Sprintf("http://api.ft.com/things/%s", res.UUID)
		membership.Person.APIURL = fmt.Sprintf("http://test.api.ft.com/people/%s", res.UUID)
		membership.Person.PrefLabel = res.Name
		membership.Person.Types = mapTypes(res.Labels)
		membership.Title = res.Title
		memberships = append(memberships, membership)
	}

	/*
		j, err := json.MarshalIndent(result, "  ", "  ")
		if err != nil {
			return
		}
		fmt.Println(string(j))
	*/

	return
}

func mapSubsidiaries(subs []resultSubsidiary) (apiSubs []apiSubsidiary) {
	if len(subs) != 0 {
		for _, s := range subs {
			apiSubs = append(apiSubs, apiSubsidiary{
				ID:        fmt.Sprintf("http://api.ft.com/things/%s", s.UUID),
				APIURL:    fmt.Sprintf("http://test.api.ft.com/organisations/%s", s.UUID),
				PrefLabel: s.PrefLabel,
				Types:     mapTypes(s.Labs),
			})
		}
	}

	//TODO: think about sorting? What order do we want the subs returned in?

	return
}

func mapIndustry(ind resultIndustry) *apiIndustryClassification {
	if ind.UUID != "" {
		return &apiIndustryClassification{
			ID:        fmt.Sprintf("http://api.ft.com/things/%s", ind.UUID),
			APIURL:    fmt.Sprintf("http://test.api.ft.com/things/%s", ind.UUID),
			PrefLabel: ind.PrefLabel,
		}
	}
	return nil
}

func mapTypes(in []string) (out []string) {
	for _, t := range in {
		mappedType := typesMap[t]
		if mappedType == "" {
			//log.Printf("can't map type %s, skipping", t)
		} else {
			out = append(out, mappedType)
		}
	}
	return
}

var typesMap = map[string]string{
	"Company":       "http://www.ft.com/ontology/company/Company",
	"Organisation":  "http://www.ft.com/ontology/organisation/Organisation",
	"PublicCompany": "http://www.ft.com/ontology/company/PublicCompany",
	"Person":        "http://www.ft.com/ontology/person/Person",
}
