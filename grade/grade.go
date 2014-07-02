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

type GradeVersion []int64

func (g GradeVersion) String() string {
	rev := make([]string, len(g))
	for i, v := range g {
		rev[i] = strconv.FormatInt(v, 10)
	}
	//  1.212.23.32
	return strings.Join(rev, ".")
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
