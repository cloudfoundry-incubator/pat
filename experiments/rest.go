package experiments

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/julz/pat/config"
	"github.com/nu7hatch/gouuid"
)

type Context struct {
	username      string
	password      string
	target        string
	loginEndpoint string
	apiEndpoint   string
	token         string
}

func NewContext() *Context {
	return &Context{}
}

func (c *Context) DescribeParameters(config config.Config) {
	config.StringVar(&c.target, "rest:target", "", "the target for the REST api")
	config.StringVar(&c.username, "rest:username", "", "username for REST api")
	config.StringVar(&c.password, "rest:password", "", "password for REST api")
}

func (context *Context) Target() error {
	resp, err := http.Get(context.target + "/v2/info")
	info := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&info)
	fmt.Println(info)
	context.loginEndpoint = info["authorization_endpoint"].(string)
	context.apiEndpoint = context.target
	fmt.Println("login endpoint: " + context.loginEndpoint)
	return err
}

func (context *Context) Login() error {
	values := make(url.Values)
	values.Add("grant_type", "password")
	values.Add("username", context.username)
	values.Add("password", context.password)
	values.Add("scope", "")

	client := &http.Client{}
	fmt.Println("Accessing: " + context.loginEndpoint + "/oauth/token")
	req, _ := http.NewRequest("POST", context.loginEndpoint+"/oauth/token", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("cf", "")
	resp, err := client.Do(req)
	body := struct {
		Token string `json:"access_token"`
	}{}
	fmt.Println(resp)
	err = json.NewDecoder(resp.Body).Decode(&body)
	if resp.StatusCode > 400 {
		err = errors.New("Failed to log in: " + resp.Status)
	}
	if err == nil {
		fmt.Println("Got login token")
		context.token = body.Token
		return nil
	} else {
		fmt.Println("Boo, error: " + err.Error())
		return err
	}
}

func (context *Context) Push() error {

	uuid, _ := uuid.NewV4()
	createApp := struct {
		Name      string `json:"name"`
		SpaceGuid string `json:"space_guid"`
	}{uuid.String(), "dd601ecd-2986-4de7-b8a0-49ba7bcdb576"}

	encoded, _ := json.Marshal(createApp)
	fmt.Println("POST " + context.apiEndpoint + "/v2/apps")
	req, _ := http.NewRequest("POST", context.apiEndpoint+"/v2/apps", strings.NewReader(string(encoded)))
	req.Header.Set("Authorization", "bearer "+context.token)
	client := &http.Client{}
	resp, _ := client.Do(req)
	fmt.Println(resp)
	appUri := resp.Header.Get("Location")
	fmt.Println("guid ", appUri)

	var b bytes.Buffer
	multi := multipart.NewWriter(&b)
	appbits, _ := multi.CreateFormFile("application", "app.zip")
	zipper := zip.NewWriter(appbits)
	configru, _ := zipper.Create("config.ru")
	configru.Write([]byte("app = lambda do |env|\n body = 'Hello, World!'\n [200, { 'Content-Type' => 'text/plain', 'Content-Length' => body.length.to_s }, [body] ]\n end\n\n run app"))
	gemfile, _ := zipper.Create("Gemfile")
	gemfile.Write([]byte("source \"https://rubgems.org\" \n\ngem \"rack\""))
	gemfilelock, _ := zipper.Create("Gemfile.lock")
	gemfilelock.Write([]byte("GEM\n  remote: https://rubygems.org/\n  specs:\n    track (1.5.2)\n\nPLATFORMS\n  ruby\n\nDEPENDENCIES\n  rack"))
	zipper.Close()
	resources, _ := multi.CreateFormField("resources")
	resources.Write([]byte("[]"))
	multi.Close()
	req, _ = http.NewRequest("PUT", context.apiEndpoint+appUri+"/bits", &b)
	req.Header.Set("Authorization", "bearer "+context.token)
	req.Header.Set("Content-Type", multi.FormDataContentType())
	resp, _ = client.Do(req)
	fmt.Println(resp)

	req, _ = http.NewRequest("PUT", context.apiEndpoint+appUri, strings.NewReader("{ \"state\": \"STARTED\" }"))
	req.Header.Set("Authorization", "bearer "+context.token)
	resp, _ = client.Do(req)
	fmt.Println(resp)

	for {
		req, _ = http.NewRequest("GET", context.apiEndpoint+appUri+"/instances", nil)
		req.Header.Set("Authorization", "bearer "+context.token)
		resp, _ = client.Do(req)
		fmt.Println(resp)
		decoded := make(map[string]interface{})
		json.NewDecoder(resp.Body).Decode(&decoded)
		fmt.Println(decoded["error_code"])

		if resp.StatusCode < 400 || decoded["error_code"] != "CF-NotStaged" {
			if decoded["error_code"] != nil {
				return errors.New("App Failed to Stage")
			}
			break
		}

		time.Sleep(2 * time.Second)
	}

	fmt.Println("Push succesful! ")
	return nil
}
