package grade

import (
	"fmt"
	"strings"
)

const (
	GRADE_ROOT Grade = "root" //最顶层
	GRADE_TAG  Grade = ""     //最低层
)

type Grade string

//判断指定的Grade能否使用，规则是本级及以上的可以使用
func (g Grade) GradeCanUse(canUseGrade interface{}) bool {
	switch v := canUseGrade.(type) {
	case Grade:
		return g == GRADE_TAG || strings.HasPrefix(string(g), string(v))
	case string:
		return g == GRADE_TAG || strings.HasPrefix(string(g), v)
	default:
		panic(fmt.Errorf("error type :%q", canUseGrade))
	}

}
