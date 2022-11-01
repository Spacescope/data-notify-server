package core

import (
	"errors"
)

type Topic struct {
	Topic string `form:"topic" json:"topic" binding:"required" example:"messages/vm_messages..."`
}

type TipsetState struct {
	Topic         string `form:"topic" json:"topic" binding:"required"`
	Tipset        uint64 `form:"tipset" json:"tipset" desc:"tipset.Height()"`
	Version       uint16 `form:"version" json:"version"`
	State         uint16 `form:"state" json:"state" desc:"1 - task successful, 2 - task failed"`
	NotFoundState uint8  `form:"not_found_state" json:"not_found_state" desc:"1 - tipset not found, can't find tipset is a special case of failed tasks, when state equal 2, this field takes effect"`
	Description   string `form:"description" json:"description"`
}

func (r *TipsetState) Validate() error {
	switch r.State {
	case 1:
	case 2:
	default:
		return errors.New("unkown state")
	}

	switch r.NotFoundState {
	case 0:
	case 1:
	default:
		return errors.New("unkown not_found_state")
	}

	if r.State == 1 && r.NotFoundState == 1 {
		return errors.New("This field will not takes effect when state equal 1")
	}

	return nil
}
