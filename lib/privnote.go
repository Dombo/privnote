package lib

import (
	"bufio"
	crand "crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/miquella/ask"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"encoding/binary"

	"log"
)

var (
	autoPassLength = 9
	autoPassChars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"
)

type privnote struct {
	HasManualPass bool   `json:"has_manual_pass"`
	Policy        int    `json:"policy"`
	Expires       string `json:"expires_js"`
	Link          string `json:"note_link"`
	DontAsk       bool   `json:"dont_ask"`
}

func CreateNote(cmd *cobra.Command, expires string) {
	var noteContents string
	piped, err := os.Stdin.Stat()
	if err != nil {
		log.Panicf("failed to open pipe %v", err)
	}
	if piped.Mode()&os.ModeNamedPipe == 0 {
		if cmd.Flags().Lookup("file").Changed {
			dat, err := ioutil.ReadFile(cmd.Flags().Lookup("file").Value.String())
			if err != nil {
				log.Panicf("failed to open file beforefor encryption %v", err)
			}
			noteContents = string(dat)
		}
	} else {
		pipescanner := bufio.NewScanner(os.Stdin)

		for pipescanner.Scan() {
			noteContents += pipescanner.Text()
		}
	}

	var password, hasManualPass string
	if cmd.Flags().Lookup("password").Changed {
		password, err = promptPassword()
		if err != nil {
			log.Panicf("failed to get user input for password %v", err)
		}
		hasManualPass = "true"
	} else if viper.GetString("password") != "" {
		password = viper.GetString("password")
		hasManualPass = "true"
	} else {
		password = makePassword()
		hasManualPass = "false"
	}

	if hasManualPass == "true" {
		err = checkOpensslExists()
		if err != nil {
			log.Panicf("failed to satisfy runtime requirements: %v", err)
		}
	}

	reqUrl, err := url.Parse("https://privnote.com/legacy/")
	if err != nil {
		log.Panicf("failed to parse privnote URL %v", err)
	}

	reqForm := url.Values{}
	reqForm.Add("data", string(encrypt(noteContents, password)))
	reqForm.Add("has_manual_pass", hasManualPass)
	reqForm.Add("duration_hours", expires)
	reqForm.Add("dont_ask", viper.GetString("do-not-prompt"))
	reqForm.Add("data_type", "T")
	reqForm.Add("notify_email", viper.GetString("notify-email"))
	reqForm.Add("notify_ref", viper.GetString("notify-reference"))

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, reqUrl.String(), strings.NewReader(reqForm.Encode()))
	r.Header.Add("DNT", "1")
	r.Header.Add("Connection", "keep-alive")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(reqForm.Encode())))
	r.Header.Add("Origin", "https://privnote.com")
	r.Header.Add("Referer", "https://privnote.com/")
	r.Header.Add("User-Agent", "privnote/v0.0.0 (https://github.com/dombo/privnote)")
	r.Header.Add("X-Requested-With", "XMLHttpRequest")

	res, err := client.Do(r)
	if err != nil {
		log.Panicf("failed to hit privnote %v", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panicf("failed to read response %v", err)
	}

	err = res.Body.Close()
	if err != nil {
		log.Panicf("failed to close body after reading %v", err)
	}

	var note privnote
	err = json.Unmarshal(body, &note)
	if err != nil {
		log.Panicf("failed parse response %v", err)
	}

	var linkable string
	if hasManualPass == "true" {
		linkable = note.Link
	} else {
		linkable = fmt.Sprintf("%s#%s", note.Link, password) // Constructs a link including the generated 'password'
	}
	fmt.Println(linkable)
}

func encrypt(message string, password string) []byte {
	cmd := exec.Command("openssl", "enc", "-e", "-aes-256-cbc", "-k", password, "-a", "-md", "md5")

	stdin, err := cmd.StdinPipe()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, message)
	}()

	out, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	// TODO re-evaluate this for it's warning it gives you when you use CombinedOutput()

	return out
}

func checkOpensslExists() error {
	_, err := exec.LookPath("openssl")
	if err != nil {
		return fmt.Errorf("didn't find 'openssl' executable, install it via your package manager %v\n", err)
	} else {
		return nil
	}
}

// makePassword has been modified somewhat from the original to rely on
//	the various OS's random number generator as the browser sandbox would
/*
// Reference implementation as per common.js
   var make_password = function () {
       var length = autoPassLength;
       var chars = autoPassChars;
       var str = "";
       for (var i = 0; i < length; i++) {
           pos = Math.floor(Math.random() * chars.length);
           str += chars.charAt(pos);
       }
       return str;
   }
*/
func makePassword() string {
	var keyspace = autoPassChars
	var keylength = autoPassLength
	var password = ""
	var source cryptoSource
	rnd := rand.New(source)

	for i := 0; i < keylength; i++ {
		password += string(keyspace[rnd.Intn(len(keyspace))])
	}

	return password
}

func promptPassword() (string, error) {
	err := ask.Print("Please enter your desired password!\n")
	if err != nil {
		return "", err
	}

	return ask.HiddenAsk("Password: ")
}

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}
