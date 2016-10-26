package storage

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

// NewScoreStorage initializes the storage
func NewScoreStorage() *ScoreStorage {
	return &ScoreStorage{make(map[string]*model.Score), 1}
}

// ScoreStorage stores all car shares
type ScoreStorage struct {
	scores  map[string]*model.Score
	idCount int
}

// GetAll returns the score map (because we need the ID as key too)
func (s ScoreStorage) GetAll() map[string]*model.Score {
	return s.scores
}

// GetOne score
func (s ScoreStorage) GetOne(id string) (model.Score, error) {
	score, ok := s.scores[id]
	if ok {
		return *score, nil
	}
	errMessage := fmt.Sprintf("Score for id %s not found", id)
	return model.Score{}, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
}

// Insert a score
func (s *ScoreStorage) Insert(score model.Score) string {
	id := fmt.Sprintf("%d", s.idCount)
	score.ID = id
	s.scores[id] = &score
	s.idCount++
	return id
}

// Delete one :(
func (s *ScoreStorage) Delete(id string) error {
	_, exists := s.scores[id]
	if !exists {
		return fmt.Errorf("Score with id %s does not exist", id)
	}
	delete(s.scores, id)

	return nil
}

// Update a score
func (s *ScoreStorage) Update(score model.Score) error {
	_, exists := s.scores[score.ID]
	if !exists {
		return fmt.Errorf("Score with id %s does not exist", score.ID)
	}
	s.scores[score.ID] = &score

	return nil
}
