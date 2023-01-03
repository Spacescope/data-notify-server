package core

import (
	"errors"
)

type Topic struct {
	Topic string `form:"topic" json:"topic" binding:"required" example:"messages/vm_messages..."`
}

type Force struct {
	Force bool `form:"force" json:"force" desc:"force to walk"`
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

type Walk struct {
	MinHeight uint64 `form:"from" json:"from"`
	MaxHeight uint64 `form:"to" json:"to"`
	Topic     string `form:"topic" json:"topic" binding:"required" desc:"example: all, it means walk all topics"`
	Force     bool   `form:"force" json:"force" desc:"force to walk"`
	Lotus0    string `form:"-" json:"-"`
	Mq        string `form:"-" json:"-"`
}

func (r *Walk) Validate() error {
	if r.MinHeight > r.MaxHeight {
		return errors.New("'from' should less or equal than 'to'")
	}

	return nil
}

type Gap struct {
	Lotus0 string `form:"-" json:"-"`
	Mq     string `form:"-" json:"-"`
}

type Retry struct {
	Lotus0 string `form:"-" json:"-"`
	Mq     string `form:"-" json:"-"`
}
