package utils

import (
	"strings"
	"github.com/leleo886/lopic/models"
)

func WordCount(images []models.Image) map[string]int{
	wordCount := make(map[string]int)
	for _, image := range images {
		for _, word := range image.Tags {
			word = strings.TrimSpace(word)
			if word != "" {
				wordCount[word]++
			}
		}
	}
	return wordCount
}