package queue

import (
	"fmt"
	"strconv"
	"strings"

	taskq "fission-basic/kit/task"

	"github.com/samber/lo"
)

type RallyData struct {
	WaID       string
	RallyCode  string
	NickName   string
	Channel    string
	Language   string
	SendTime   int64
	Generation string
}

func (o *RallyData) GenerationInt64() int64 {
	g, _ := strconv.ParseInt(o.Generation, 10, 64)
	return g
}

type rally struct {
	*taskq.Queue
}

func (o *rally) SendBack(d *RallyData) error {
	fmt.Println("SendBack=", o.wrapKey(d))
	return o.Queue.SendBack([]string{o.wrapKey(d)}, false)
}

func (o *rally) SendBackForce(d *RallyData) error {
	fmt.Println("SendBack=", o.wrapKey(d))
	return o.Queue.SendBack([]string{o.wrapKey(d)}, true)
}

func (o *rally) SendBacks(d []*RallyData) error {
	keys := lo.Map(d, func(rally *RallyData, _ int) string {
		return o.wrapKey(rally)
	})

	return o.Queue.SendBack(keys, false)
}

func (o *rally) wrapKey(d *RallyData) string {
	// 注意nickName放到最后，因为他可能包含特殊字符'|'
	return fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s",
		d.WaID, d.RallyCode, d.Channel, d.Language, d.SendTime, d.Generation, d.NickName)
}

func (o *rally) UnWrap(key string) (*RallyData, error) {
	strs := strings.Split(key, "|")
	if len(strs) < 7 {
		return nil, fmt.Errorf("invalid key: %s", key)
	}

	sendTime, _ := strconv.ParseInt(strs[4], 10, 64)

	return &RallyData{
		WaID:       strs[0],
		RallyCode:  strs[1],
		NickName:   strings.Join(strs[6:], "|"),
		Channel:    strs[2],
		Language:   strs[3],
		Generation: strs[5],
		SendTime:   sendTime,
	}, nil
}
