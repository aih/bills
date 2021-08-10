package bills

import "encoding/json"

type TitlesJson struct {
	As           string `json:"as"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	IsForPortion bool   `json:"is_for_portion"`
}

type BillMeta struct {
	Actions                  []ActionItem      `json:"actions"`
	Congress                 string            `json:"congress"`
	BillType                 string            `json:"bill_type"`
	Number                   string            `json:"number"`
	BillCongressTypeNumber   string            `json:"bill_congress_type_number"`
	History                  interface{}       `json:"history"`
	OfficialTitle            string            `json:"official_title"`
	PopularTitle             string            `json:"popular_title"`
	ShortTitle               string            `json:"short_title"`
	Titles                   []string          `json:"titles"`
	TitlesWholeBill          []string          `json:"titles_whole_bill"`
	Cosponsors               []CosponsorItem   `json:"cosponsors"`
	Committees               []CommitteeItem   `json:"committees"`
	RelatedBills             []RelatedBillItem `json:"related_bills"`
	RelatedBillsByBillnumber RelatedBillMap    `json:"related_dict"`
}

type BillMetaDoc map[string]BillMeta

type ActionItem struct {
	ActedAt    string   `json:"acted_at"`
	ActionCode string   `json:"action_code"`
	References []string `json:"references"`
	Text       string   `json:"text"`
	Type       string   `json:"type"`
}

/*
type HistoryItem struct {
	Active bool `json:"active"`
	ActiveAt string `json:"active_at"`
 	AwaitingSignature bool `json:"awaiting_signature"`
 	Enacted bool `json:"enacted"`
 	HousePassageResult string `json:"house_passage_result"`
 	HousePassageResultAt string `json:"house_passage_result_at"`
 	Vetoed bool `json:"vetoed"`
 	Active bool `json:"active"`
 	ActiveAt string `json:"active_at"`
 	AwaitingSignature bool `json:"awaiting_signature"`
 	Enacted bool `json:"enacted"``
 	HousePassageResult string `json:"house_passage_result"`
 	HousePassageResultAt string `json:"house_passage_result_at"`
 	Vetoed bool `json:"vetoed"`
}
*/

type SummaryItem struct {
	As   string `json:"as"`
	Date string `json:"date"`
	Text string `json:"text"`
}

type CosponsorItem struct {
	BioguideId string `json:"bioguide_id"`
	ThomasId   string `json:"thomas_id"`
	// Type              string `json:"type"`
	District          string `json:"district"`
	Name              string `json:"name"`
	OriginalCosponsor bool   `json:"original_cosponsor"`
	SponsoredAt       string `json:"sponsored_at"`
	State             string `json:"state"`
	Title             string `json:"title"`
	// WithdrawnAt       string `json:"withdrawn_at"`
}

type CommitteeItem struct {
	Activity       []string `json:"activity"`
	Committee      string   `json:"committee"`
	CommitteeId    string   `json:"committee_id"`
	Subcommittee   string   `json:"subcommittee"`
	SubcommitteeId string   `json:"subcomittee_id"`
}

type RelatedBillItem struct {
	BillId                 string `json:"bill_id"`
	IdentifiedBy           string `json:"identified_by"`
	Reason                 string `json:"reason"`
	Type                   string `json:"type"`
	BillCongressTypeNumber string `json:"bill_congress_type_number"`
	//Sponsor                CosponsorItem   `json:"sponsor"`
	//Cosponsors             []CosponsorItem `json:"cosponsors"`
	Titles          []string `json:"titles"`
	TitlesWholeBill []string `json:"titles_whole_bill"`
}

type RelatedBillMap map[string]RelatedBillItem

type SimilarBillItem struct {
	Date                          string  `json:"date"`
	Score                         float64 `json:"score"`
	Title                         string  `json:"title"`
	Session                       string  `json:"session"`
	Congress                      string  `json:"congress"`
	Legisnum                      string  `json:"legisnum"`
	Billnumber                    string  `json:"billnumber"`
	SectionNum                    string  `json:"section_num"`
	SectionIndex                  string  `json:"sectionIndex"`
	SectionHeader                 string  `json:"section_header"`
	BillCongressTypeNumberVersion string  `json:"bill_number_version"`
	TargetSectionHeader           string  `json:"target_section_header"`
	TargetSectionNumber           string  `json:"target_section_number"`
}
type SimilarSection struct {
	BillNumberVersion string `json:"bill_number_version"`
	Score             string `json:"score"`
	BillNumber        string `json:"bill_number"`
	Congress          string `json:"congress"`
	Session           string `json:"session"`
	Legisnum          string `json:"legisnum"`
	Title             string `json:"title"`
	Section           string `json:"section"`
	SectionHeader     string `json:"section_header"`
	Date              string `json:"date"`
}

type SimilarSections []SimilarSection

type DataJson struct {
	Actions          []ActionItem      `json:"actions"`
	Amendments       []interface{}     `json:"amendments"`
	BillId           string            `json:"bill_id"`
	BillType         string            `json:"bill_type"`
	ByRequest        bool              `json:"by_request"`
	CommitteeReports []interface{}     `json:"committee_reports"`
	Committees       []CommitteeItem   `json:"committees"`
	Congress         string            `json:"congress"`
	Cosponsors       []CosponsorItem   `json:"cosponsors"`
	EnactedAs        string            `json:"enacted_as"`
	History          interface{}       `json:"history"`
	IntroducedAt     string            `json:"introduced_at"`
	Number           string            `json:"number"`
	OfficialTitle    string            `json:"official_title"`
	PopularTitle     string            `json:"popular_title"`
	RelatedBills     []RelatedBillItem `json:"related_bills"`
	ShortTitle       string            `json:"short_title"`
	Sponsor          string            `json:"sponsor"`
	Status           string            `json:"status"`
	StatusAt         string            `json:"status_at"`
	Subjects         []interface{}     `json:"subjects"`
	SubjectsTopTerm  string            `json:"subjects_top_term"`
	Summary          SummaryItem       `json:"summary"`
	Titles           []TitlesJson      `json:"titles"`
	UpdatedAt        string            `json:"updated_at"`
	Url              string            `json:"url"`
}

// SearchResult represents the result of the search operation
type SearchResult_ES struct {
	Took     uint64 `json:"took"`
	TimedOut bool   `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Failed     int `json:"failed"`
		Skipped    int `json:"skipped"`
	} `json:"_shards"`
	Hits ResultHits `json:"hits"`
}

// ResultHits represents the result of the search hits
type ResultHits struct {
	Total    int     `json:"total"`
	MaxScore float32 `json:"max_score"`
	Relation string  `json:"relation"`
	Value    int     `json:"value"`
	Hits     []struct {
		ID        string          `json:"_id"`
		Index     string          `json:"_index"`
		Type      string          `json:"_type"`
		Score     float32         `json:"_score"`
		Source    json.RawMessage `json:"_source"`
		InnerHits struct {
			Hits struct {
				Hits     []InnerHit
				MaxScore float32 `json:"max_score"`
				Total    struct {
					Relation string `json:"relation"`
					Value    int    `json:"value"`
				} `json:"total"`
			} `json:"hits"`
		} `json:"inner_hits"`
	} `json:"hits"`
}

type InnerHit struct {
	ID     string  `json:"_id"`
	Index  string  `json:"_index"`
	Type   string  `json:"_type"`
	Score  float32 `json:"_score"`
	Source struct {
		BillNumber  string        `json:"bill_number"`
		BillVersion string        `json:"bill_version"`
		Congress    string        `json:"congress"`
		Date        string        `json:"date"`
		Legisnum    string        `json:"legisnum"`
		Session     string        `json:"session"`
		DCTitle     string        `json:"dc_title"`
		Type        string        `json:"type"`
		Sections    []SectionItem `json:"sections"`
	} `json:"_source"`
}

type SectionItem struct {
	SectionNumber string `json:"section_number"`
	SectionHeader string `json:"section_header"`
	SectionText   string `json:"section_text"`
	SectionXML    string `json:"section_xml"`
}

type ResultInnerHits []struct {
	Index  string          `json:"_index"`
	Type   string          `json:"_type"`
	ID     string          `json:"_id"`
	Score  float32         `json:"_score"`
	Source json.RawMessage `json:"_source"`
	//Highlight map[string][]string `json:"highlight,omitempty"`
	Sections InnerHitSections `json:"sections"`
}

//"_id":"ccN083gBppu2L0JvvoHQ","_index":"billsections","_nested":{"field":"sections","offset":1},"_score":43.656067,"_source":{"section_header":"Repeal of estate and gift taxes","section_number":"2.","section_text":"2.Re

type InnerHitSections struct {
	Hits struct {
		Hits []struct {
			SectionHit struct {
				ID     string `json:"_id"`
				Index  string `json:"_index"`
				Nested struct {
					Field  string `json:"field"`
					Offset int    `json:"offset"`
				} `json:"_nested"`
				Score  string `json:"_score"`
				Source struct {
					SectionNumber string `json:"section_number"`
					SectionHeader string `json:"section_header"`
					SectionText   string `json:"section_text"`
				} `json:"_source"`
			}
		} `json:"hits"`
	} `json:"hits"`
}

type WordSample struct {
	BillCongressTypeNumber string   `json:"bill_congress_type_number"`
	WordList               []string `json:"word_list"`
}
