package service // dnywonnt.me/alerts2incidents/internal/service

import (
	"fmt"
	"regexp"
	"time"

	"dnywonnt.me/alerts2incidents/internal/models"

	log "github.com/sirupsen/logrus"
)

// FindMatchingAlerts checks a slice of alerts against a rule's conditions to find matches.
func FindMatchingAlerts(alerts []models.Alert, rule *models.Rule) ([]models.Alert, error) {
	// Logging the start of the matching process with relevant rule and alerts info.
	log.WithFields(log.Fields{
		"numAlerts":        len(alerts),
		"ruleID":           rule.ID,
		"conditionsLength": len(rule.AlertsSummaryConditions),
	}).Debug("Starting finding matching alerts based on the rule")

	// Initializing a slice to hold the alerts that match the rule conditions.
	matchingAlerts := []models.Alert{}
	// Map to track which alerts have been used/matched to prevent re-matching.
	usedAlertIndexes := make(map[int]struct{})

	// Iterate over each summary condition in the rule.
	for i, condition := range rule.AlertsSummaryConditions {
		conditionMatchFound := false

		// Check each alert against the current condition.
		for j, alert := range alerts {
			if _, used := usedAlertIndexes[j]; used {
				continue // Skip alerts that have already been matched.
			}

			// QuoteMeta escapes special characters in `condition` for safe use in
			// regular expressions, producing `quotedCondition`.
			quotedCondition := regexp.QuoteMeta(condition)

			// Compile the summary condition into a regular expression.
			compiledRegex, err := regexp.Compile(quotedCondition)
			if err != nil {
				return nil, fmt.Errorf("error compiling regex for condition '%s': %w", condition, err)
			}

			// Check if the alert's summary matches the condition and it occurred within the specified interval.
			if compiledRegex.MatchString(alert.Summary) && time.Since(alert.CreatedAt) >= rule.AlertsActivityIntervalConditions[i] {
				matchingAlerts = append(matchingAlerts, alert)
				usedAlertIndexes[j] = struct{}{} // Mark this alert as used.
				conditionMatchFound = true
				break // Stop checking once a match is found for a condition.
			}
		}

		// If no matching alert is found for a condition, log and exit the loop.
		if !conditionMatchFound {
			log.WithFields(log.Fields{
				"conditionIndex":    i,
				"summaryCondition":  condition,
				"intervalCondition": rule.AlertsActivityIntervalConditions[i].String(),
			}).Debug("No matching alerts found for the condition")
			break
		}
	}

	// If not all conditions have matching alerts, log the event and return no results.
	if len(matchingAlerts) != len(rule.AlertsSummaryConditions) {
		log.WithFields(log.Fields{
			"matchingAlertsCount":     len(matchingAlerts),
			"requiredConditionsCount": len(rule.AlertsSummaryConditions),
		}).Debug("Insufficient matching alerts to satisfy the rule conditions")
		return nil, nil
	}

	// Log the successful finding of matching alerts.
	log.WithFields(log.Fields{
		"numMatchingAlerts": len(matchingAlerts),
		"ruleID":            rule.ID,
	}).Debug("Successfully found matching alerts")

	return matchingAlerts, nil
}
