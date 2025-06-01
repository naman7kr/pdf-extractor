package models

type Article struct {
	Title  string
	Author string
}

// Define the YAML structure
type ArticlesConfig struct {
	Articles []Article `yaml:"articles"`
}
