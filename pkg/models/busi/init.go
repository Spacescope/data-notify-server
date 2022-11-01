package busi

var (
	Tables []interface{}
)

func init() {
	Tables = append(Tables,
		new(TipsetsState),
		new(Topics),
	)
}
