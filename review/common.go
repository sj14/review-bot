package review

func MissingReviewers(reviewedBy []string, approvers map[string]string) []string {
	var missing []string
	for userID, userName := range approvers {
		approved := false
		for _, approverID := range reviewedBy {
			if userID == approverID {
				approved = true
				break
			}
		}
		if !approved {
			missing = append(missing, userName)
		}
	}

	return missing
}
