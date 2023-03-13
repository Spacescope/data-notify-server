package busi

import (
	"time"
)

type TipsetsState struct {
	Id            uint64    `json:"-" xorm:"bigserial pk autoincr"`
	TopicId       uint64    `json:"topic_id" xorm:"notnull unique(ttv)"`
	TopicName     string    `json:"topic_name" xorm:"varchar(64)"` //redundancy
	Tipset        uint64    `json:"tipset" xorm:"notnull unique(ttv)"`
	Version       uint32    `json:"version" xorm:"integer default 0 unique(ttv) comment('chainnotify has three event: current/revert/apply, version means different tipset with same height')"`
	State         uint16    `json:"state" xorm:"integer default 0 comment('0 - enqueue, 1 - task successful, 2 - task failed')"` // 没有timeout, timeout计算时发现
	NotFoundState uint8     `json:"not_found_state" xorm:"smallint default 0 comment('1 - tipset not found, can't find tipset is a special case of failed tasks, when state equal 2, this field takes effect')"`
	RetryTimes    uint16    `json:"retry_time" xorm:"integer default 0"`
	Description   string    `json:"description" xorm:"text"`
	CreateDate    time.Time `json:"create_date" xorm:"created comment('enqueue time')"`
	LastUpdate    time.Time `json:"last_update" xorm:"updated comment('feeback time')"`
}

type Topics struct {
	Id         uint64    `json:"-" xorm:"bigserial pk autoincr"`
	TopicName  string    `json:"topic_name" xorm:"varchar(64) notnull unique"`
	State      uint8     `json:"state" xorm:"smallint default 0 comment('0 - on, 1 - off')"`
	CreateDate time.Time `json:"create_date" xorm:"created"`
	LastUpdate time.Time `json:"last_update" xorm:"updated"`
}
