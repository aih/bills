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
