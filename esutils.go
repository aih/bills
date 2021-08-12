package bills

import (
	"strings"
)

func GetTopHit(hits Hits_ES) (topHit Hit_ES) {

	var topScore float32
	var score float32
	topScore = 0
	for _, item := range hits {
		score = item.Score
		if score > topScore {
			topScore = score
			topHit = item
		}

	}
	return topHit
}

func GetHitsES(results SearchResult_ES) (innerHits Hits_ES, err error) {
	return results.Hits.Hits, nil
}

func GetInnerHits(results SearchResult_ES) (innerHits []InnerHits, err error) {
	var hitsES, _ = GetHitsES(results)
	for _, hit := range hitsES {
		innerHits = append(innerHits, hit.InnerHits)
	}
	return innerHits, nil
}

// similars is the result of the MLT query
func GetSimilarSections(results SearchResult_ES) (similarSections SimilarSections, err error) {
	hits, _ := GetHitsES(results)
	innerHits, _ := GetInnerHits(results)
	for index, hit := range hits {
		var topInnerResultSectionHit InnerHit
		var similarSection SimilarSection
		innerResultSectionHits := innerHits[index].Sections.Hits.Hits
		if len(innerResultSectionHits) > 0 {
			// The first section matched is the best section (and usu. the only real match in the bill)
			topInnerResultSectionHit = innerResultSectionHits[0]
		}
		billSource := hit.Source
		similarSection.BillNumber = billSource.BillNumber
		similarSection.BillNumberVersion = billSource.ID
		similarSection.Congress = billSource.Congress
		similarSection.Session = billSource.Session
		similarSection.Legisnum = billSource.Legisnum
		similarSection.Score = topInnerResultSectionHit.Score
		similarSection.SectionNum = topInnerResultSectionHit.Source.SectionNumber + " "
		similarSection.SectionHeader = topInnerResultSectionHit.Source.SectionHeader
		similarSection.Date = billSource.Date
		dublinCores := billSource.DC
		dublinCore := ""
		if len(dublinCores) > 0 {
			dublinCore = dublinCores[0]
			result := DcTitle_Regexp.FindAllStringSubmatch(dublinCore, -1)
			title := ""
			if len(result) > 0 && len(result[0]) > 1 {
				title = strings.Trim(result[0][1], " ")
			}
			similarSection.Title = title
			similarSections = append(similarSections, similarSection)
		}
	}
	return similarSections, nil
}
