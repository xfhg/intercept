package cmd

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/theckman/yacspin"
)

var (
	basic_username string
	basic_password string
	token_auth     string
)

func gatheringData(value Rule) {

	var resp *resty.Response
	var err error

	fmt.Println("│ ")
	fmt.Println(line)
	fmt.Println("│ ")
	fmt.Println("├ API Rule #", value.ID)
	fmt.Println("│ Rule name : ", value.Name)
	fmt.Println("│ Rule description : ", value.Description)
	fmt.Println("│ Impacted Env : ", value.Environment)
	fmt.Println("│ Confidence : ", value.Confidence)
	fmt.Println("│ Tags : ", value.Tags)
	fmt.Println("│ ")

	fmt.Println("│ API ENDPOINT : ", value.Api_Endpoint)
	fmt.Println("│ ")

	client := resty.New()

	if value.Api_Insecure {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		colorYellowBold.Println("│ API INSECURE - NO TLS CHECK")
		fmt.Println("│ ")
	}

	cfg := yacspin.Config{
		Frequency:       100 * time.Millisecond,
		CharSet:         yacspin.CharSets[3],
		Suffix:          " API Gathering Data",
		SuffixAutoColon: true,
		StopCharacter:   "│ ✓",
		StopColors:      []string{"fgGreen"},
	}

	spinner, err := yacspin.New(cfg)
	if err != nil {
		LogError(err)
	}

	switch value.Api_Auth {
	case "basic":
		//fmt.Println("│ API BASIC AUTH")
		if (value.Api_Auth_Basic) != nil {
			secret, ok := os.LookupEnv("INTERCEPT_" + *value.Api_Auth_Basic)
			if ok {

				if strings.Contains(secret, ":") {
					authParts := strings.Split(secret, ":")
					if len(authParts) > 1 {
						basic_username = authParts[0]
						basic_password = authParts[1]

					}
				} else {
					LogError(errors.New("API Basic Auth Environment variable must be defined as username:password"))
				}

			} else {
				LogError(errors.New("API Auth Environment variables not set for this request"))
			}
		} else {
			LogError(errors.New("API Auth Environment variables not set for this request"))
		}
	case "token":
		//fmt.Println("│ API TOKEN AUTH")
		if (value.Api_Auth_Token) != nil {
			secret, ok := os.LookupEnv("INTERCEPT_" + *value.Api_Auth_Token)
			if ok {
				fmt.Println("│ TOKEN AUTH OK")
				token_auth = secret

			} else {

				LogError(errors.New("API Auth Environment variables not set for this request"))
			}
		} else {
			LogError(errors.New("API Auth Environment variables not set for this request"))
		}

	default:
		LogError(errors.New("API Auth not defined"))
	}

	_ = os.Remove("output_" + strconv.Itoa(value.ID))

	err = spinner.Start()
	if err != nil {
		LogError(err)
	}

	switch value.Api_Request {
	case "GET":
		if token_auth != "" {
			resp, err = client.R().
				EnableTrace().
				SetAuthToken(token_auth).
				SetHeader("Content-Type", "application/xml").
				SetBody(value.Api_Body).
				SetOutput("output_" + strconv.Itoa(value.ID)).
				Get(value.Api_Endpoint)

		} else {
			resp, err = client.R().
				EnableTrace().
				SetBasicAuth(basic_username, basic_password).
				SetHeader("Content-Type", "application/xml").
				SetBody(value.Api_Body).
				SetOutput("output_" + strconv.Itoa(value.ID)).
				Get(value.Api_Endpoint)
		}

	case "POST":
		if token_auth != "" {

			resp, err = client.R().
				EnableTrace().
				SetAuthToken(token_auth).
				SetHeader("Content-Type", "application/xml").
				SetBody(value.Api_Body).
				SetOutput("output_" + strconv.Itoa(value.ID)).
				Post(value.Api_Endpoint)

		} else {

			resp, err = client.R().
				EnableTrace().
				SetBasicAuth(basic_username, basic_password).
				SetBody(value.Api_Body).
				SetOutput("output_" + strconv.Itoa(value.ID)).
				Post(value.Api_Endpoint)

		}

	default:
		LogError(errors.New("Invalid request type"))
		return
	}

	if err != nil {
		spinner.StopFail()
		LogError(err)
	} else {
		spinner.Stop()
		fmt.Println("│ ")
		fmt.Println("│ API Response Status:", resp.Status())
		fmt.Println("│ ")

		if value.Api_Trace {

			fmt.Println("│ API Body:", value.Api_Body)

			fmt.Println("│ API Response Body:", resp.String())
			fmt.Println("│ API Response Body:", resp.Body())

			// Explore response object
			fmt.Println("│ API Response Info:")
			fmt.Println("│   Error      :", err)
			fmt.Println("│   Status Code:", resp.StatusCode())
			fmt.Println("│   Status     :", resp.Status())
			fmt.Println("│   Proto      :", resp.Proto())
			fmt.Println("│   Time       :", resp.Time())
			fmt.Println("│   Received At:", resp.ReceivedAt())
			// Explore trace info
			fmt.Println("│ API Request Trace Info:")
			ti := resp.Request.TraceInfo()
			fmt.Println("│   DNSLookup     :", ti.DNSLookup)
			fmt.Println("│   ConnTime      :", ti.ConnTime)
			fmt.Println("│   TCPConnTime   :", ti.TCPConnTime)
			fmt.Println("│   TLSHandshake  :", ti.TLSHandshake)
			fmt.Println("│   ServerTime    :", ti.ServerTime)
			fmt.Println("│   ResponseTime  :", ti.ResponseTime)
			fmt.Println("│   TotalTime     :", ti.TotalTime)
			fmt.Println("│   IsConnReused  :", ti.IsConnReused)
			fmt.Println("│   IsConnWasIdle :", ti.IsConnWasIdle)
			fmt.Println("│   ConnIdleTime  :", ti.ConnIdleTime)
			fmt.Println("│   RequestAttempt:", ti.RequestAttempt)
			fmt.Println("│   RemoteAddr    :", ti.RemoteAddr.String())
		}

	}

}

func processAPIType(value Rule) {

	if cfgEnv == "" {
		cfgEnv = "先锋"
	}

	rules := loadUpRules()

	pwddir := GetWd()

	rgembed, _ := prepareEmbeddedExecutable()

	searchPatternFile := strings.Join([]string{pwddir, "/", "search_regex"}, "")

	scanPath := "output_" + strconv.Itoa(value.ID)

	if !FileExists(scanPath) {
		LogError(errors.New("API Output not found"))
	}

	exception := ContainsInt(rules.Exceptions, value.ID)

	if exception && !auditNox && !value.Enforcement {

		colorRedBold.Println("│")
		colorRedBold.Println("│ ", rules.ExceptionMessage)
		colorRedBold.Println("│")

	} else {

		codePatternScan := []string{"--pcre2", "-p", "-o", "-A0", "-B0", "-C0", "-i", "-U", "-f", searchPatternFile, scanPath}
		xcmd := exec.Command(rgembed, codePatternScan...)
		xcmd.Stdout = os.Stdout
		xcmd.Stderr = os.Stderr
		errr := xcmd.Run()

		if errr != nil {
			if xcmd.ProcessState.ExitCode() == 2 {
				LogError(errr)
			} else {

				envfound := FindMatchingString(cfgEnv, value.Environment, ",")
				if (envfound || strings.Contains(value.Environment, "all") || value.Environment == "") && value.Fatal {

					colorRedBold.Println("│")
					colorRedBold.Println("│ NON COMPLIANT : ")
					colorRedBold.Println("│ ", value.Error)
					colorRedBold.Println("│")
					fatal = true
					stats.Fatal++
				} else {

					colorRedBold.Println("│")
					colorRedBold.Println("│ NOT FOUND")
					colorRedBold.Println("│ ", value.Error)
					colorRedBold.Println("│")
					warning = true

				}
				colorRedBold.Println("│")
				colorRedBold.Println("│ API Rule : ", value.Name)
				colorRedBold.Println("│ Target Environment : ", value.Environment)
				colorRedBold.Println("│ Suggested Solution : ", value.Solution)
				colorRedBold.Println("│")
				fmt.Println("│ ")
				stats.Total++
				stats.Dirty++

			}
		} else {

			colorGreenBold.Println("│ ")
			colorGreenBold.Println("│ Compliant")
			stats.Clean++
			stats.Total++
			fmt.Println("│ ")

		}

		jsonOutputFile := strings.Join([]string{pwddir, "/", strconv.Itoa(value.ID), ".json"}, "")
		jsonoutfile, erroutjson := os.Create(jsonOutputFile)
		if erroutjson != nil {
			LogError(erroutjson)
		}
		defer jsonoutfile.Close()
		writer := bufio.NewWriter(jsonoutfile)
		defer writer.Flush()

		codePatternScanJSON := []string{"--pcre2", "--no-heading", "-o", "-p", "-i", "-U", "--json", "-f", searchPatternFile, scanPath}
		xcmdJSON := exec.Command(rgembed, codePatternScanJSON...)
		xcmdJSON.Stdout = jsonoutfile
		xcmdJSON.Stderr = os.Stderr
		errrJSON := xcmdJSON.Run()

		if errrJSON != nil {
			if xcmdJSON.ProcessState.ExitCode() == 2 {
				LogError(errrJSON)
			} else {

				ProcessOutput(strings.Join([]string{strconv.Itoa(value.ID), ".json"}, ""), strconv.Itoa(value.ID), value.Name, value.Description, value.Error, value.Solution, value.Fatal)
				GenerateSarif()
				colorRedBold.Println("│ ")
			}
		} else {
			colorGreenBold.Println("│")
			os.Remove(jsonOutputFile)

		}

	}

}
