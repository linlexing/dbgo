package main

import ()

const (
	CHECK_LEVEL_DISABLE int64 = iota //禁用
	CHECK_LEVEL_ACCEPT               //出错时可以保存
	CHECK_LEVEL_FORCE                //出错时可以强制保存
	CHECK_LEVEL_REFUSED              //出错时不能保存，也不能强制保存
)

type Check struct {
	ID           int64
	DisplayLabel TranslateString
	Level        int64
	Fields       []string
	RunAtServer  bool
	Script       string
	SqlWhere     string
	Grade        string
}

/*type ClientCheck struct {
	ID           int64
	DisplayLabel string
	Level        int64
	Fields       []string
	RunAtServer  bool
	Script       string
}
type CheckResult struct {
	PKs         []string
	CheckID     int64
	Grade       string
	RefreshTime time.Time
	RefreshBy   string
}
*/
