package main

import (
	"fmt"
	"strconv"
	"strings"
)

type Color string

const (
	Black Color = "black"
	White Color = "white"
)

func ParseCSAColor(s string) (Color, error) {
	if s == "+" {
		return Black, nil
	} else if s == "-" {
		return White, nil
	}
	return Black, fmt.Errorf("invalid color string")
}

func FormatCSAColor(c Color) (string, error) {
	if c == Black {
		return "+", nil
	}
	if c == White {
		return "-", nil
	}
	return "", fmt.Errorf("invalid color value")
}

//---

type SquareNumber int

func (n SquareNumber) IsValid() bool {
	return 1 <= n && n <= 9
}

func ParseCSASquareNumber(s string) (SquareNumber, error) {
	n, err := strconv.Atoi(string(s[0]))
	if err != nil {
		return 0, fmt.Errorf("invalid SquareNumber: %s", err)
	}
	return SquareNumber(n), nil
}

//---

type Square [2]SquareNumber

func (sq *Square) IsValid() bool {
	return sq[0].IsValid() && sq[1].IsValid()
}

func ParseCSASquare(s string) (*Square, error) {
	if len(s) != 2 {
		return nil, fmt.Errorf("invalid length: %d", len(s))
	}
	x, err := ParseCSASquareNumber(string(s[0]))
	if err != nil {
		return nil, fmt.Errorf("x: %s", err)
	}
	y, err := ParseCSASquareNumber(string(s[1]))
	if err != nil {
		return nil, fmt.Errorf("y: %s", err)
	}
	return &Square{x, y}, nil
}

func FormatCSASquare(sq *Square) (string, error) {
	if sq == nil {
		return "00", nil
	}
	return fmt.Sprint("%d%d", sq[0], sq[1]), nil
}

//---

type Piece string

const (
	FU Piece = "FU"
	KY Piece = "KY"
	KE Piece = "KE"
	GI Piece = "GI"
	KI Piece = "KI"
	KA Piece = "KA"
	HI Piece = "HI"
	OU Piece = "OU"
	TO Piece = "TO"
	NY Piece = "NY"
	NK Piece = "NK"
	NG Piece = "NG"
	UM Piece = "UM"
	RY Piece = "RY"
)

var pieces = []Piece{FU, KY, KE, GI, KI, KA, HI, OU, TO, NY, NK, NG, UM, RY}

var pieceSet map[Piece]struct{}

func init() {
	pieceSet = make(map[Piece]struct{}, len(pieces))
	for _, p := range pieces {
		pieceSet[p] = struct{}{}
	}
}

func ParseCSAPiece(s string) (Piece, error) {
	p := Piece(s)
	if _, ok := pieceSet[p]; ok {
		return p, nil
	}
	return "", fmt.Errorf("invalid CSA piece string")
}

func FormatCSAPiece(s string) (string, error) {
	return string(s), nil
}

//---

type EventType string

const (
	EMove   EventType = "MOVE"
	EResign EventType = "RESIGN"
)

type Event interface {
	EventType() EventType
}

type MoveEvent struct {
	Type      EventType `json:"type"`
	Color     Color     `json:"color"`
	SrcSquare *Square   `json:"srcSquare"`
	DstSquare *Square   `json:"dstSquare"`
	DstPiece  Piece     `json:"dstPiece"`
	Time      int       `json:"time"`
}

func FormatCSAMoveEvent(e *MoveEvent) (string, error) {
	return "TODO", nil
}

func (e *MoveEvent) EventType() EventType {
	return e.Type
}

func (e *MoveEvent) IsDrop() bool {
	return e.SrcSquare == nil
}

type ResignEvent struct {
	Type  EventType `json:"type"`
	Color Color     `json:"color"`
}

func FormatCSAResignEvent(e *ResignEvent) (string, error) {
	return "%TORYO", nil
}

func (e *ResignEvent) EventType() EventType {
	return e.Type
}

func ParseCSAEvent(s string) (Event, error) {
	if s[:8] == "GOTE_WIN" {
		return &ResignEvent{
			Type:  EResign,
			Color: Black,
		}, nil
	} else if s[:9] == "SENTE_WIN" {
		return &ResignEvent{
			Type:  EResign,
			Color: White,
		}, nil
	} else if s[:4] == "DRAW" {
		return nil, fmt.Errorf("TODO")
	}
	color, err := ParseCSAColor(string(s[0]))
	if err != nil {
		return nil, err
	}
	srcSquare, err := ParseCSASquare(s[1:3])
	if err != nil {
		return nil, err
	}
	dstSquare, err := ParseCSASquare(s[3:5])
	if err != nil {
		return nil, err
	}
	piece, err := ParseCSAPiece(s[5:7])
	if err != nil {
		return nil, err
	}
	return &MoveEvent{
		Type:      EMove,
		Color:     color,
		SrcSquare: srcSquare,
		DstSquare: dstSquare,
		DstPiece:  piece,
	}, nil
}

func FormatCSAEvent(e Event) (string, error) {
	return "TODO", nil
}

//---

type Record struct {
	Events []Event `json:"events"`
}

func FormatCSARecord(record *Record) (string, error) {
	ss := make([]string, 0, len(record.Events))
	for _, event := range record.Events {
		s, err := FormatCSAEvent(event)
		if err != nil {
			return "", err
		}
		ss = append(ss, s)
	}
	return strings.Join(ss, "\n"), nil
}
