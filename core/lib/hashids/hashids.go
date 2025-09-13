package hashids

import (
	"errors"

	"github.com/meraf00/swytch/core"
	"github.com/speps/go-hashids/v2"
)

type HashID interface {
	EncodeID(id uint) (string, error)
	DecodeID(hash string) (uint, error)
}

type hashIDService struct {
	hasher *hashids.HashID
}

func NewHashIDService(config core.EncryptionConfig) (HashID, error) {
	hd := hashids.NewData()
	hd.Salt = config.HashSalt
	hd.MinLength = 8

	hasher, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, err
	}

	return &hashIDService{hasher: hasher}, nil
}

func (s *hashIDService) EncodeID(id uint) (string, error) {
	return s.hasher.Encode([]int{int(id)})
}

func (s *hashIDService) DecodeID(hash string) (uint, error) {
	numbers, err := s.hasher.DecodeWithError(hash)
	if err != nil || len(numbers) == 0 {
		return 0, err
	}
	if numbers[0] < 0 {
		return 0, errors.New("decoded ID is negative, cannot convert to uint")
	}
	return uint(numbers[0]), nil
}
