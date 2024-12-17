package mocks

import (
	"time"
	"snippetbox.mike9708.net/internal/models"
)


var mockSnippet = &models.Snippet{
	ID: 1,
	Title: "An old siletn pound",
	Content: "An old siletn pound...",
	Created: time.Now(),
	Expires: time.Now(),
}

type SnippetModel struct{}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	return 2, nil
}

func (m *SnippetModel) Get(id int) (*models.Snippet, error) {
	switch id {
		case 1:
			return mockSnippet, nil
		default:
			return nil, models.ErrNoRecord
			
	}
}


func (m *SnippetModel) Latest() ([]*models.Snippet, error) {
	return []*models.Snippet{mockSnippet}, nil
}


