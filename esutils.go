package bills

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
/*
func GetSimilarSections(similars map[string]interface{}) (SimilarSections, error) {
	hits, _ := GetInnerHits(similars)
	innerResults, _ := GetInnerResults(similars)
	for index, hit := range hits {
		log.Info().Msgf(" %v, %v", index, hit)
		innerResultSections = similars
	}
	return similarSections, nil
}
*/

/* Python version
def getSimilarSections(res):
  similarSections = []
  try:
    hits = getInnerHits(res)
    innerResults = getInnerResults(res)
    for index, hit in enumerate(hits):
      innerResultSections = getInnerHits(innerResults[index].get('sections'))
      billSource = hit.get('_source')
      title = ''
      dublinCore = ''
      dublinCores = billSource.get('dc', [])
      if dublinCores:
        dublinCore = dublinCores[0]

      titleMatch = re.search(r'<dc:title>(.*)?<', str(dublinCore))
      if titleMatch:
        title = titleMatch[1].strip()
      num = innerResultSections[0].get('_source', {}).get('section_number', '')
      if num:
        num = num + " "
      header = innerResultSections[0].get('_source', {}).get('section_header', '')
      match = {
        "bill_number_version": billSource.get('id', ''),
        "score": innerResultSections[0].get('_score', ''),
        "billnumber": billSource.get('billnumber', ''),
        "congress": billSource.get('_source', {}).get('congress', ''),
        "session": billSource.get('session', ''),
        "legisnum": billSource.get('legisnum', ''),
        "title": title,
        "section_num": num,
        "section_header": header,
        "date": billSource.get('date'),
      }
      similarSections.append(match)
    return similarSections
  except Exception as err:
    print(err)
    return []
*/
