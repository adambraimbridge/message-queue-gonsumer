package consumer

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"regexp"
	"strings"
)

type message struct {
	Value     string `json:"value"` //base64 encoded
	Partition int    `json:"partition"`
	Offset    int    `json:"offset"`
}

func parseResponse(data []byte) ([]Message, error) {
	var resp []message
	err := json.Unmarshal(data, &resp)
	if err != nil {
		log.Printf("ERROR - parsing json message %q failed with error %v", data, err.Error())
		return nil, err
	}
	msgs := make([]Message, 0)
	for _, m := range resp {
		log.Printf("DEBUG - parsing msg of partition %d and offset %d", m.Partition, m.Offset)
		msgs = append(msgs, parseMessage(m.Value))
	}
	return msgs, nil
}

func parseMessage(raw string) (m Message) {
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		log.Printf("ERROR - failure in decoding base64 value: %s", err.Error())
		return
	}
	m.Headers = parseHeaders(string(decoded[:]))
	m.Body = parseBody(string(decoded[:]))
	return
}

var re = regexp.MustCompile("[\\w-]*:[\\w\\-:/. ]*")

var kre = regexp.MustCompile("[\\w-]*:")
var vre = regexp.MustCompile(":[\\w-:/. ]*")

func parseHeaders(msg string) map[string]string {
	//naive
	i := strings.Index(msg, "{")
	headerLines := re.FindAllString(msg[:i], -1)

	headers := make(map[string]string)
	for _, line := range headerLines {
		key, value := parseHeader(line)
		headers[key] = value
	}
	return headers
}

func parseHeader(header string) (string, string) {
	key := kre.FindString(header)
	value := vre.FindString(header)
	return key[:len(key)-1], strings.TrimSpace(value[1:])
}
func parseBody(msg string) string {
	//naive
	f := strings.Index(msg, "{")
	l := strings.LastIndex(msg, "}")
	return msg[f : l+1]
}
