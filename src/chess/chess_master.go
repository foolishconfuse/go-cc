package chess

import (
	"log"
	"os"
	"strconv"
	"time"
	"fmt"
)

var currentBoard chessBoard
var _debugFlag = true

type chessMaster struct {
	chessBoard chessBoard
	depth int8
	evaluator *chessBoardEvaluator
	generator *chessMovementGenerator
}

func newChessMaster(depth int8, pvPath string) *chessMaster {
	cm := &chessMaster{}
	cm.depth = depth
	cm.initChessBoard()
	cm.evaluator = newChessBoardEvaluator(pvPath)
	cm.generator = newChessMovementGenerator()
	return cm
}

func (cm *chessMaster) initChessBoard() {
	initBoard := [][]byte {
		{ 2, 1, 2, 2, 2, 4, 2, 5, 2, 6, 2, 5, 2, 4, 2, 2, 2, 1 },
		{ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 },
		{ 0, 0, 2, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 3, 0, 0 },
		{ 2, 7, 0, 0, 2, 7, 0, 0, 2, 7, 0, 0, 2, 7, 0, 0, 2, 7 },
		{ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 },
		{ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 },
		{ 1, 7, 0, 0, 1, 7, 0, 0, 1, 7, 0, 0, 1, 7, 0, 0, 1, 7 },
		{ 0, 0, 1, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 3, 0, 0 },
		{ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 },
		{ 1, 1, 1, 2, 1, 4, 1, 5, 1, 6, 1, 5, 1, 4, 1, 2, 1, 1 },
	}
	cm.chessBoard = [][]*chess{}
	for row := 0; row < _BOARD_ROW; row++ {
		cols := []*chess{}
		for col := 0; col < _BOARD_COL * 2; col+=2 {
			cols = append(cols, &chess{ _type: chessType(initBoard[row][col + 1]), color: chessColor(initBoard[row][col]) })
		}
		cm.chessBoard = append(cm.chessBoard, cols)
	}
}

func (cm *chessMaster) loadChessBoard(value string) {
	if len(value) % 2 != 0 {
		log.Fatalln("error when LoadChessBoard...")
		os.Exit(1)
	}
	cm.chessBoard = [][]*chess{}
	for row := 0; row < _BOARD_ROW; row++ {
		cols := []*chess{}
		for col := 0; col < _BOARD_COL; col++ {
			cols = append(cols, &chess{ _type: _CHESS_NULL, color: _COLOR_NULL })
		}
		cm.chessBoard = append(cm.chessBoard, cols)
	}
	idx := 0
	for i := 0; i < len(value); i+=2 {
		row := idx / _BOARD_COL
		col := idx % _BOARD_COL
		t, _ := strconv.Atoi(string(value[i]))
		c, _ := strconv.Atoi(string(value[i + 1]))
		cm.chessBoard[row][col]._type = chessType(t)
		cm.chessBoard[row][col].color = chessColor(c)
		idx++
	}
}

func (cm *chessMaster) dump() {
	cm.chessBoard.dump()
}

func (cm *chessMaster) convertMoves(moves []move, parentNode *chessBoardNode, depth int8, nodeType nodeType) []*chessBoardNode {
	nodes := make([]*chessBoardNode, 100)
	nodes = nodes[:0]
	for _, v := range moves {
		node := getChessBoardNode()
		if node.children != nil {
			node.children = []*chessBoardNode {}
		}
		node.move = v
		node.parent = parentNode
		node.depth = depth
		node.setNodeType(nodeType)
		node.discard = false
		node.setValueCount = 0
		nodes = append(nodes, node)
	}
	if parentNode != nil {
		parentNode.children = nodes
	}
	return nodes
}

func (cm *chessMaster) isAllWaitForEvalNode(nodes *chessBoardNodeList) bool {
	for e := nodes.front(); e != nil; e = e.next {
		if e.parent != nil {
			return false
		}
	}
	return true
}

func (cm *chessMaster) search(value string) string {
	st := time.Now()
	cm.loadChessBoard(value)
	cm.dump()
	currentBoard = cm.chessBoard
	mainQueue := newChessBoardNodeList()
	waitForEvalQueue := newChessBoardNodeList()
	moves := make([]move, 100)
	moves = moves[:0]
	moves = cm.generator.generateMoves(cm.chessBoard, _COLOR_BLACK)
	mainQueue.pushFrontSlice(cm.convertMoves(moves, nil, 1, _NODE_TYPE_MIN))
	clipCount := 0
	anotherClipCount := 0

	for mainQueue.len() > 0 {
		node := mainQueue.popFront()
		if node.isDiscard() {
			clipCount++
			if node.depth < cm.depth {
				anotherClipCount++
			}
			continue
		}
		if node.depth < cm.depth {
			waitForEvalQueue.pushFront(node)
			nodeType := _NODE_TYPE_NULL
			color := _COLOR_NULL
			if node.nodeType == _NODE_TYPE_MIN {
				nodeType = _NODE_TYPE_MAX
				color = _COLOR_RED
			} else {
				nodeType = _NODE_TYPE_MIN
				color = _COLOR_BLACK
			}
			moves = moves[:0]
			moves = cm.generator.generateMoves(node.getCurrentChessBoard(), color)
			mainQueue.pushFrontSlice(cm.convertMoves(moves, node, node.depth + 1, nodeType))
		} else {
			v := cm.evaluator.eval(node.getCurrentChessBoard())
			if v <= _MIN_VALUE || v >= _MAX_VALUE {
				log.Fatalln("value overflow...")
			}
			node.parent.setValue(v, node)
		}
	}
	for waitForEvalQueue.len() > 0 {
		if cm.isAllWaitForEvalNode(waitForEvalQueue) {
			break
		}
		node := waitForEvalQueue.popFront()
		if node.parent == nil {
			waitForEvalQueue.pushBack(node)
		} else {
			node.parent.setValue(node.getValue(), node)
		}
	}
	var score int16 = _MIN_VALUE
 	var targetNode *chessBoardNode = nil
	tempQueue := newChessBoardNodeList()
	for waitForEvalQueue.len() > 0 {
		node := waitForEvalQueue.popFront()
		tempQueue.pushBack(node)
		nodeScore := node.getValue()
		if nodeScore != _MAX_VALUE && nodeScore > score {
			score = nodeScore
			targetNode = node
		}
	}
	if targetNode == nil {
		log.Fatalln("search targetNode == nil...")
	}
	result := targetNode.getCurrentChessBoard().string()
	resultValue := targetNode.value
	if _debugFlag {
		fmt.Println("+++++++++++++++++++++++++++++++++")
		tmp := targetNode
		for tmp != nil {
			fmt.Println(tmp.getCurrentChessBoard().dumpString())
			if tmp.valueNodeForDebug == nil {
				fmt.Println("---：")
				fmt.Println(tmp.getCurrentChessBoard().dumpString())
				fmt.Println(cm.evaluator.eval(tmp.getCurrentChessBoard()))
			}
			tmp = tmp.valueNodeForDebug
		}
		fmt.Println("+++++++++++++++++++++++++++++++++")
	}
	for tempQueue.len() > 0 {
		node := tempQueue.popFront()
		if node == nil {
			log.Println("what the xx?...tempQueue.len=", tempQueue.len())
			break
		}
		if node.children != nil {
			tempQueue.pushFrontSlice(node.children)
		}
		returnChessBoardNode(node)
	}
	log.Printf("depth: %d, clip1: %d, clip2: %d, value: %d, time cost: %f, node:(%d-%d)", cm.depth, clipCount, anotherClipCount, resultValue, time.Since(st).Seconds(), _getNodeNum, _returnNodeNum)
	clearChessBoardNodeCounter()
	return result
}