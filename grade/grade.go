package grade

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	GRADE_ROOT Grade = "root" //最顶层
	GRADE_TAG  Grade = ""     //最低层
)

type Grade string

//判断指定的Grade能否使用，规则是本级及以上的可以使用
func (g Grade) CanUse(canUseGrade interface{}) bool {
	switch v := canUseGrade.(type) {
	case Grade:
		return g == GRADE_TAG || strings.HasPrefix(string(g), string(v))
	case string:
		return g == GRADE_TAG || strings.HasPrefix(string(g), v)
	default:
		panic(fmt.Errorf("error type :%q", canUseGrade))
	}

}
func (g Grade) Child(str string) Grade {
	return Grade(string(g) + "/" + str)
}
func (g Grade) String() string {
	return string(g)
}

type GradeVersion []int64

func (g GradeVersion) String() string {
	rev := make([]string, len(g))
	for i, v := range g {
		rev[i] = strconv.FormatInt(v, 10)
	}
	//  1.212.23.32
	return strings.Join(rev, ".")
}

//the g > d,then return true
func (g GradeVersion) GE(d GradeVersion) bool {
	for i, v := range g {
		if len(d) > i {
			if v > d[i] {
				return true
			} else if v < d[i] {
				return false
			}
		} else {
			return true
		}
	}
	return len(g) == len(d)
}
func ParseGradeVersion(value string) (GradeVersion, error) {
	if value == "" {
		return GradeVersion{}, nil
	}
	strs := strings.Split(value, ".")
	rev := make(GradeVersion, len(strs))
	for i, v := range strs {
		var err error
		rev[i], err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return rev, nil
}
