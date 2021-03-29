package bills

type TitlesJson struct {
	As           string `json:"as"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	IsForPortion bool   `json:"is_for_portion"`
}

type BillMeta struct {
	BillCongressTypeNumber   string
	Titles                   []string
	TitlesWholeBill          []string
	Cosponsors               []CosponsorItem
	RelatedBills             []RelatedBillItem
	RelatedBillsByBillnumber RelatedBillMap
}

type BillMetaDoc map[string]BillMeta

type SummaryItem struct {
	As   string `json:"as"`
	Date string `json:"date"`
	Text string `json:"text"`
}

type CosponsorItem struct {
	BioGuideId        string `json:"bioguide_id"`
	Type              string `json:"type"`
	District          string `json:"district"`
	Name              string `json:"name"`
	OriginalCosponsor string `json:"original_cosponsor"`
	SponsoredAt       string `json:"sponsored_at"`
	State             string `json:"state"`
	Title             string `json:"title"`
	WithdrawnAt       string `json:"withdrawn_at"`
}

type RelatedBillItem struct {
	BillId                 string `json:"bill_id"`
	IdentifiedBy           string `json:"identified_by"`
	Reason                 string `json:"reason"`
	Type                   string `json:"type"`
	BillCongressTypeNumber string
	Sponsor                CosponsorItem
	Cosponsors             []CosponsorItem
	Titles                 []string
	TitlesWholeBill        []string
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

type DataJson struct {
	Actions          []interface{}     `json:"actions"`
	Amendments       []interface{}     `json:"amendments"`
	BillId           string            `json:"bill_id"`
	BillType         string            `json:"bill_type"`
	ByRequest        bool              `json:"by_request"`
	CommitteeReports []interface{}     `json:"committee_reports"`
	Committees       []interface{}     `json:"committees"`
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

type WordSample struct {
	BillCongressTypeNumber string
	WordList               []string
}
