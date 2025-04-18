package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"

	jsoniter "github.com/json-iterator/go"
)

var (
	tool = flag.String("tool", "", "tool")
	code = flag.String("code", "", "code")

	parall = flag.Int64("p", 0, "parall")

	input  = flag.String("i", "", "input")
	output = flag.String("o", "", "output")

	start = flag.Int("s", 0, "start")
)

func main() {
	flag.Parse()

	if *tool == "create" {
		createGroup()
	} else if *tool == "join" {
		if *code == "" {
			panic("missing code")
		}
		joinGroup(*code)
	} else if *tool == "data" {
		if *input == "" || *output == "" {
			panic("missing input or output")
		}
		callData()
	} else {
		panic("unknown tool")
	}
}

func callData() {
	file, err := os.OpenFile(*input, os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("打开文件出错: %v", err)
	}
	defer file.Close()

	wfile, err := os.OpenFile(*output, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("打开文件出错: %v", err)
		return
	}

	i := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		i++
		line := scanner.Text()
		// Do something with the line
		args := strings.Split(line, "参数=")
		if len(args) != 2 {
			fmt.Println("first line=", i)
			// fmt.Println(line)
			continue
		}

		args2 := strings.Split(args[1], ", 回调响应")
		if len(args2) != 2 {
			fmt.Println("sencond line=", i)
			// fmt.Println(args[1])
			continue
		}

		if !strings.Contains(args2[0], "🎁") {
			fmt.Println("not contains=", i)
			continue
		}

		var m map[string]interface{}
		err = jsoniter.Unmarshal([]byte(args2[0]), &m)
		if err != nil {
			fmt.Println("json error=", err, ", line=", i)
			continue
		}

		fmt.Fprint(wfile, m["body"])
		fmt.Fprint(wfile, "\n")
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

}

func createGroup() {
	body := `{
    "messaging_product": "whatsapp",
    "contacts": [
        {
            "wa_id": "85266831314",
            "profile": {
                "name": "牛牛"
            }
        }
    ],
    "messages": [
        {
            "from": "639692369842",
            "id": "wamid.HBgLODUyNTc0ODE5MjAVAgASGBQzQTIyRjgxQUQ3NEE5RDdDNDgzMwG=",
            "timestamp": "1735470750",
            "type": "text",
            "text": {
               "body": "I'm joining the MLBB GOLDEN MONTH bonus sharing event to win 🎁 amazing rewards including $1,000 cash, OPPO phone, 100,000 MLBB Diamonds, and an exclusive skin!\nUse My Code: a020100000"
            },
            "cost": {
                "currency": "USD",
                "price": 0,
                "foreign_price": 0,
                "cdr_type": 1,
                "message_id": "wamid.HBgLODUyNTc0ODE5MjAVAgASGBQzQTIyRjgxQUQ3NEE5RDdDNDgzMwA=",
                "direction": 2
            }
        }
    ],
    "metadata": {
        "display_phone_number": "639692369842",
        "phone_number_id": "296085720248808"
    },
    "app_id": "1533",
    "business_phone": "639692369842",
    "merchant_phone": "639692369842",
    "channel": 2
}`

	uri := `http://localhost:9001/events/mlbb25031gateway/activity/userAttendInfo`

	if *parall <= 0 {
		panic("parall <= 0")
	}

	tasks := make(chan func(context.Context), 2*(*parall))
	ctx := context.TODO()

	go func() {
		for i := 0; i < int(*parall); i++ {
			for req := range tasks {
				req(ctx)
			}
		}
	}()

	for i := *start; i < (*start)+200*int(*parall); i++ {
		tasks <- func(ctx context.Context) {
			req, err := http.NewRequest("POST", uri, strings.NewReader(body))
			if err != nil {
				fmt.Printf("err=%v, body=%s\n", err, body)
				panic(err)
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("err=%v, req=%+v\n", err, req)
				panic(err)
			}

			if resp.StatusCode != 200 {
				fmt.Println(resp.Status)
				// panic(resp.StatusCode)
			}
		}
	}

	close(tasks)

	return
}

func joinGroup(helpCode string) {
	uri := `http://localhost:9000/events/mlbb25031gateway/activity/userAttendInfo`
	body := `{
    "messaging_product": "whatsapp",
    "contacts": [
        {
            "wa_id": "{{.WaID}}",
            "profile": {
                "name": "{{.Name}}"
            }
        }
    ],
    "messages": [
        {
            "from": "639692369842",
            "id": "{{.MessageID}}", 
            "timestamp": "1735470750",
            "type": "text",
            "text": {
               "body": "I'm accepting {{1}}'s invitation to join the MLBB GOLDEN MONTH bonus sharing event! 🎁 Win $1,000 Cash, OPPO Phone, 100,000 MLBB Diamonds, an exclusive skin, and more!\nMy Event Code: a0201{{.HelpCode}}"
            },
            "cost": {
                "currency": "USD",
                "price": 0,
                "foreign_price": 0,
                "cdr_type": 1,
                "message_id": "{{.MessageID}}",
                "direction": 2
            }
        }
    ],
    "metadata": {
        "display_phone_number": "639692369842",
        "phone_number_id": "296085720248808"
    },
    "app_id": "1533",
    "business_phone": "639692369842",
    "merchant_phone": "639692369842",
    "channel": 2
}`

	if *parall <= 0 {
		panic("parall <= 0")
	}

	tasks := make(chan func(context.Context), 2*(*parall))
	ctx := context.TODO()
	closeChan := make(chan struct{})

	go func() {
		defer func() {
			closeChan <- struct{}{}
		}()
		for i := 0; i < int(*parall); i++ {
			for req := range tasks {
				req(ctx)
			}
		}
	}()

	for i := *start; i < *start+200*int(*parall); i++ {
		tasks <- func(ctx context.Context) {
			// 解析模板
			tmpl, err := template.New("example").Parse(body)
			if err != nil {
				log.Fatalf("解析模板出错: %v", err)
			}

			data := map[string]string{
				"WaID":      fmt.Sprintf("8526683132%d", i),
				"Name":      fmt.Sprintf("Name-%d", i),
				"MessageID": fmt.Sprintf("wamid.HBgLODUyNTc0ODE5MjAVAgASGBQzQTIyRjgxQUQ3NEE5RDdDNDgzMwG%d", i),
				"HelpCode":  helpCode,
			}

			var buf bytes.Buffer
			err = tmpl.Execute(&buf, data)
			if err != nil {
				log.Fatalf("执行模板出错: %v", err)
				return
			}

			req, err := http.NewRequest("POST", uri, strings.NewReader(buf.String()))
			if err != nil {
				fmt.Printf("err=%v, body=%s\n", err, body)
				panic(err)
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("err=%v, req=%+v\n", err, req)
				panic(err)
			}

			if resp.StatusCode != 200 {
				panic(resp.StatusCode)
			}
		}
	}

	close(tasks)
	<-closeChan

	return
}
