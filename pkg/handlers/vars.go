package handlers

import (
	"github.com/gamaops/gamago/pkg/id"
)

var sessionIDGenerator *id.IDGenerator
var challengeGenerator *id.IDGenerator
var nonceGenerator *id.IDGenerator

func SetupIDGenerators() error {
	var err error
	sessionIDGenerator, err = id.NewIDGenerator(10)
	if err != nil {
		return err
	}
	challengeGenerator, err = id.NewIDGenerator(12)
	if err != nil {
		return err
	}
	nonceGenerator, err = id.NewIDGenerator(12)
	if err != nil {
		return err
	}
	return nil
}
