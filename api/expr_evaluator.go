package api

import "fmt"

// 定义操作符类型常量
const (
	TYPE_OPERAND = 1 // 操作数
	TYPE_OR      = 2 // OR 操作符
	TYPE_AND     = 3 // AND 操作符
	TYPE_LGROUP  = 4 // 左括号
	TYPE_RGROUP  = 5 // 右括号
)

// 表达式元素结构
type ExprElement struct {
	Type int  // 元素类型
	Val  bool // 值(对于操作符和括号，默认为 false)
}

// 表达式求值器结构
type ExprEvaluator struct {
	elements      []ExprElement // 表达式元素列表
	operandStack  []bool        // 操作数栈
	operatorStack []int         // 操作符栈
}

// 创建新的表达式求值器
func NewExprEvaluator(elements []ExprElement) *ExprEvaluator {
	return &ExprEvaluator{
		elements: elements,
	}
}

// 计算表达式结果
func (e *ExprEvaluator) Evaluate() (bool, error) {
	for _, elem := range e.elements {
		switch elem.Type {
		case TYPE_OPERAND:
			e.operandStack = append(e.operandStack, elem.Val)
		case TYPE_OR, TYPE_AND:
			for len(e.operatorStack) > 0 && e.hasHigherPrecedence(e.operatorStack[len(e.operatorStack)-1], elem.Type) {
				if err := e.executeOperation(); err != nil {
					return false, err
				}
			}
			e.operatorStack = append(e.operatorStack, elem.Type)
		case TYPE_LGROUP:
			e.operatorStack = append(e.operatorStack, elem.Type)
		case TYPE_RGROUP:
			for len(e.operatorStack) > 0 && e.operatorStack[len(e.operatorStack)-1] != TYPE_LGROUP {
				if err := e.executeOperation(); err != nil {
					return false, err
				}
			}
			if len(e.operatorStack) == 0 {
				return false, fmt.Errorf("mismatched parentheses")
			}
			// 弹出左括号
			e.operatorStack = e.operatorStack[:len(e.operatorStack)-1]
		}
	}

	// 处理剩余的操作符
	for len(e.operatorStack) > 0 {
		if err := e.executeOperation(); err != nil {
			return false, err
		}
	}

	if len(e.operandStack) != 1 {
		return false, fmt.Errorf("invalid expression")
	}

	return e.operandStack[0], nil
}

// 判断操作符 op1 是否具有比 op2 更高的优先级
func (e *ExprEvaluator) hasHigherPrecedence(operator1, operator2 int) bool {
	// 左括号不参与优先级比较
	if operator1 == TYPE_LGROUP {
		return false
	}
	// 数值越大，优先级越高
	return operator1 >= operator2
}

// 执行操作，专注于处理二元操作
func (e *ExprEvaluator) executeOperation() error {
	// 二元操作，操作符的长度大于0，操作数的数量大于等于2
	if len(e.operatorStack) == 0 || len(e.operandStack) < 2 {
		return fmt.Errorf("invalid expression")
	}
	operator := e.operatorStack[len(e.operatorStack)-1]
	e.operatorStack = e.operatorStack[:len(e.operatorStack)-1]

	// 获取操作数
	right := e.operandStack[len(e.operandStack)-1]
	left := e.operandStack[len(e.operandStack)-2]
	e.operandStack = e.operandStack[:len(e.operandStack)-2]

	var result bool
	switch operator {
	case TYPE_AND:
		result = left && right
	case TYPE_OR:
		result = left || right
	default:
		return fmt.Errorf("unknown operator")
	}
	e.operandStack = append(e.operandStack, result)
	return nil
}
