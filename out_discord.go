package main

import (
	"C"
	"os"
	"fmt"
	"time"
	"unsafe"

	curl "github.com/andelf/go-curl"
	"github.com/fluent/fluent-bit-go/output"
)

/* Global Vars */
var user string
var url string
var avatar string
var POST_DATA string

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	return output.FLBPluginRegister(def, "discord", "Discord Output")
}

//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	url = output.FLBPluginConfigKey(plugin, "url")
	user = output.FLBPluginConfigKey(plugin, "username")
	avatar = output.FLBPluginConfigKey(plugin, "avatar_url")

	if user == "" {
		user = "Fluent Bit"
	}

	if url == "" {
		fmt.Print("[out_discord] URL is a required field!")
		os.Exit(1)
	}

	fmt.Printf("[out_discord] username=%s, avatar=%s, url=%s\n", user, avatar, url)

	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var count int
	var ret int
	var ts interface{}
	var record map[interface{}]interface{}

	// Create Fluent Bit decoder
	dec := output.NewDecoder(data, int(length))

	// Iterate Records
	count = 0
	for {
		// Extract Record
		ret, ts, record = output.GetRecord(dec)
		if ret != 0 {
			break
		}

		var timestamp time.Time
		switch t := ts.(type) {
		case output.FLBTime:
			timestamp = ts.(output.FLBTime).Time
		case uint64:
			timestamp = time.Unix(int64(t), 0)
		default:
			fmt.Println("time provided invalid, defaulting to now.")
			timestamp = time.Now()
		}

		// Print record keys and values -- timestamp.String() for date
		bt := "```"
		POST_DATA = fmt.Sprintf("{\"username\": \"%s\", \"avatar_url\": \"%s\", \"content\": \"**Alert from %s**\\n%sini\\n[time]  %s\\n", user, avatar, C.GoString(tag), bt, timestamp.String())

		for k, v := range record {

			// Print strings as strings
			switch v.(type) {
			case bool, int, int64, uint64, float32, float64:
			        POST_DATA = fmt.Sprintf("%s[%s]  %v\\n", POST_DATA, k, v)
			default:
			        POST_DATA = fmt.Sprintf("%s[%s]  %s\\n", POST_DATA, k, v)
			}
		}

		POST_DATA = fmt.Sprintf("%s\\n%s\"}", POST_DATA, bt)
		}

	// Send to Discord
	var sent = false

	easy := curl.EasyInit()
	defer easy.Cleanup()
	if easy != nil {

		easy.Setopt(curl.OPT_HTTPHEADER, []string{"Content-type: application/json"})
		//easy.Setopt(curl.OPT_VERBOSE, true)
		easy.Setopt(curl.OPT_URL, url)
		easy.Setopt(curl.OPT_SSL_VERIFYPEER, false)
		easy.Setopt(curl.OPT_POST, true)
		easy.Setopt(curl.OPT_READFUNCTION,
			func(ptr []byte, userdata interface{}) int {
				// WARNING: never use append()
				if !sent {
					sent = true
					ret := copy(ptr, POST_DATA)
					return ret
				}
				return 0 // sent ok
			})
		easy.Setopt(curl.OPT_POSTFIELDSIZE, len(POST_DATA))

		easy.Perform()

		if err := easy.Perform(); err != nil {
			println("ERROR: ", err.Error())
		}

		POST_DATA = ""

		count++
	}

	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
