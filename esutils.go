package bills

func GetTopHit(hits []interface{}) (topHit map[string]interface{}) {

	var topScore float64
	var score float64
	topScore = 0
	for _, item := range hits {
		score = item.(map[string]interface{})["_score"].(float64)
		if score > topScore {
			topScore = score
			topHit = item.(map[string]interface{})
		}

	}
	return topHit
}

func GetInnerHits(res map[string]interface{}) (innerHits []interface{}, err error) {
	// TODO: check if res is a map with the right keys
	innerHits = res["hits"].(map[string]interface{})["hits"].([]interface{})
	return innerHits, nil
}

func GetInnerResults(res map[string]interface{}) (innerResults []map[string]interface{}, err error) {
	var innerHits, _ = GetInnerHits(res)
	//TODO check for error
	for index, hit := range innerHits {
		//log.Debug().Msgf("hit: %v", hit)
		innerResults = append(innerResults, hit.(map[string]interface{})["inner_results"].([]map[string]interface{})[index])
	}
	return innerResults, nil
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
