package main

import (
	"fmt"
	"github.com/linlexing/dbgo/oftenfun"
	"github.com/robertkrimen/otto/ast"
	"github.com/robertkrimen/otto/parser"
	"github.com/robertkrimen/otto/token"
	"strconv"
	"strings"
)

const (
	CHECK_SQL_OUTLINE_TABLENAME = "_check_outline_"
)

type SqlParser struct {
	StringIdentifiers []string
	AdditionParse     func(exp ast.Expression) (string, bool, error)
	Fields            []string
	HasSubQuery       bool
}

func (b *SqlParser) parseOperate(tk token.Token, left, right ast.Expression) (string, bool, error) {
	tkstr := ""
	revfmt := "(%s %s %s)"
	switch tk {
	case token.LOGICAL_AND:
		tkstr = "and"
		revfmt = "%s %s %s"
	case token.LOGICAL_OR:
		tkstr = "or"
	case token.PLUS:
		tkstr = "||"
		revfmt = "%s %s %s"
	default:
		return "", false, fmt.Errorf("the token %s invalid", tk)
	}
	var l, r string
	var err error
	var isstrl, isstrr bool
	l, isstrl, err = b.ParseExpression(left)
	if err != nil {
		return "", false, err
	}
	r, isstrr, err = b.ParseExpression(right)
	if err != nil {
		return "", false, err
	}
	if tk == token.PLUS && !isstrl && !isstrr {
		tkstr = "+"
	}
	return fmt.Sprintf(revfmt, l, tkstr, r), isstrl || isstrr, nil
}

func (b *SqlParser) parseComparisonOperate(tk token.Token, left, right ast.Expression) (string, error) {
	tkstr := ""
	switch tk {
	case token.LESS, token.LESS_OR_EQUAL, token.GREATER, token.GREATER_OR_EQUAL:
		tkstr = tk.String()
	case token.EQUAL:
		if _, ok := right.(*ast.NullLiteral); ok {
			tkstr = "IS"
		} else {
			tkstr = "="
		}
	case token.NOT_EQUAL:
		if _, ok := right.(*ast.NullLiteral); ok {
			tkstr = "IS NOT"
		} else {
			tkstr = "<>"
		}
	default:
		return "", fmt.Errorf("the token [%s] invalid", tk)
	}
	var l, r string
	var err error
	l, _, err = b.ParseExpression(left)
	if err != nil {
		return "", err
	}
	r, _, err = b.ParseExpression(right)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %s", l, tkstr, r), nil

}
func (b *SqlParser) processFunction(funcName string, params ...string) (string, error) {
	processParams := func(strSql string, params ...string) string {
		for i, v := range params {
			strSql = strings.Replace(strSql, fmt.Sprintf("$%d", i+1), fmt.Sprintf("%s.%s", CHECK_SQL_OUTLINE_TABLENAME, v), -1)
		}
		return strSql
	}
	switch funcName {
	case "query":
		if len(params) == 0 {
			return "", fmt.Errorf("must least one parameter")
		}
		b.HasSubQuery = true
		return fmt.Sprintf("(%s)", processParams(params[0], params[1:]...)), nil
	case "exists":
		if len(params) == 0 {
			return "", fmt.Errorf("must least one parameter")
		}
		b.HasSubQuery = true
		return fmt.Sprintf("exists(%s)", processParams(params[0], params[1:]...)), nil
	case "notexists":
		if len(params) == 0 {
			return "", fmt.Errorf("must least one parameter")
		}
		b.HasSubQuery = true
		return fmt.Sprintf("not exists(%s)", processParams(params[0], params[1:]...)), nil
	case "regexp_like":
		if len(params) != 2 {
			return "", fmt.Errorf("must have two parameters")
		}
		return fmt.Sprintf("%s~%s", params[0], params[1]), nil

	default:
		return fmt.Sprintf("%s(%s)", funcName, strings.Join(params, ",")), nil
	}

}
func (b *SqlParser) ParseExpression(exp ast.Expression) (string, bool, error) {
	switch t := exp.(type) {
	case *ast.BinaryExpression:
		if t.Comparison {
			str, err := b.parseComparisonOperate(t.Operator, t.Left, t.Right)
			return str, false, err
		} else {
			return b.parseOperate(t.Operator, t.Left, t.Right)
		}
	case *ast.Identifier:
		b.Fields = append(b.Fields, t.Name)
		return t.Name, oftenfun.In(t.Name, b.StringIdentifiers...), nil
	case *ast.NullLiteral:
		return "NULL", false, nil
	case *ast.StringLiteral:
		str := strconv.Quote(t.Value)
		str = strings.Replace(str[1:len(str)-1], "'", "''", -1)
		return "E'" + str + "'", true, nil
	case *ast.NumberLiteral:
		return fmt.Sprintf("%v", t.Value), false, nil
	case *ast.BooleanLiteral:
		if t.Value {
			return "TRUE", false, nil
		} else {
			return "FALSE", false, nil
		}
	case *ast.UnaryExpression:
		if t.Operator == token.NOT {
			str, _, err := b.ParseExpression(t.Operand)
			if err != nil {
				return "", false, nil
			}
			return "NOT(" + str + ")", false, nil
		} else {
			return "", false, fmt.Errorf("unary express only [not] is valid,the %s invalid", t.Operator)
		}
	case *ast.CallExpression:
		//只进行单纯的函数调用，成员函数调用不处理
		if _, ok := t.Callee.(*ast.Identifier); ok {
			call, isstr, err := b.ParseExpression(t.Callee)
			paramstr := make([]string, len(t.ArgumentList))
			if err != nil {
				return "", isstr, err
			}
			b.Fields = b.Fields[:len(b.Fields)-1]
			for i, v := range t.ArgumentList {
				paramstr[i], _, err = b.ParseExpression(v)
				if err != nil {
					return "", isstr, err
				}
			}
			str, err := b.processFunction(call, paramstr...)
			return str, isstr, err
		} else {
			if b.AdditionParse != nil {
				return b.AdditionParse(t)
			}
			return "", false, fmt.Errorf("invalid express:%T", exp)
		}

	default:
		if b.AdditionParse != nil {
			return b.AdditionParse(t)
		}
		return "", false, fmt.Errorf("invalid express:%T", exp)
	}
}
func ParseToSql(src string, strIden ...string) (string, error) {
	p := &SqlParser{strIden, nil, nil, false}
	f, err := parser.ParseFunction("", src)
	if err != nil {
		return "", err
	}
	str, _, err := p.ParseExpression(f.Body.(*ast.BlockStatement).List[0].(*ast.ExpressionStatement).Expression)
	return str, err

}
