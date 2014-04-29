package workloads

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/cloudfoundry-community/pat/logs"
)

type httpclient interface {
	Get(token string, url string, data interface{}, responseBody interface{}) (reply Reply)
	Put(token string, url string, data interface{}, responseBody interface{}) (reply Reply)
	MultipartPut(token string, m *multipart.Writer, url string, data *bytes.Buffer, responseBody interface{}) (reply Reply)
	Post(token string, url string, data interface{}, responseBody interface{}) (reply Reply)
	PostToUaa(url string, data url.Values, responseBody interface{}) (reply Reply)
}

type Reply struct {
	Code     int
	Message  string
	Location string
}

const TRACE_REST_CALLS = true

func (client rest) Post(token string, url string, data interface{}, body interface{}) Reply {
	return client.req(token, "POST", url, "", "", "", jsonToString(data), body)
}

func (client rest) Put(token string, url string, data interface{}, body interface{}) Reply {
	return client.req(token, "PUT", url, "", "", "", jsonToString(data), body)
}

func (client rest) MultipartPut(token string, m *multipart.Writer, url string, data *bytes.Buffer, body interface{}) Reply {
	return client.req(token, "PUT", url, m.FormDataContentType(), "", "", data, body)
}

func (client rest) Get(token string, url string, data interface{}, body interface{}) Reply {
	return client.req(token, "GET", url, "", "", "", jsonToString(data), body)
}

func (client rest) PostToUaa(url string, data url.Values, reply interface{}) Reply {
	return client.req("", "POST", url, "application/x-www-form-urlencoded", "cf", "", strings.NewReader(data.Encode()), reply)
}

func (context *rest) GetSuccessfully(token string, url string, data url.Values, responseBody interface{}, fn func(reply Reply) error) error {
	reply := context.client.Get(token, url, data, responseBody)
	return checkSuccessfulReply(reply, func() error {
		return fn(reply)
	})
}

func (context *rest) PutSuccessfully(token string, url string, data interface{}, responseBody interface{}, fn func(reply Reply) error) error {
	reply := context.client.Put(token, url, data, responseBody)
	return checkSuccessfulReply(reply, func() error {
		return fn(reply)
	})
}

func (context *rest) MultipartPutSuccessfully(token string, m *multipart.Writer, url string, data *bytes.Buffer, responseBody interface{}, fn func(reply Reply) error) error {
	reply := context.client.MultipartPut(token, m, url, data, responseBody)
	return checkSuccessfulReply(reply, func() error {
		return fn(reply)
	})
}

func (context *rest) PostSuccessfully(token string, url string, data interface{}, responseBody interface{}, fn func(reply Reply) error) error {
	reply := context.client.Post(token, url, data, responseBody)
	return checkSuccessfulReply(reply, func() error {
		return fn(reply)
	})
}

func (context *rest) PostToUaaSuccessfully(url string, data url.Values, responseBody interface{}, fn func(reply Reply) error) error {
	reply := context.client.PostToUaa(url, data, responseBody)
	return checkSuccessfulReply(reply, func() error {
		return fn(reply)
	})
}

func jsonToString(data interface{}) io.Reader {
	j, _ := json.Marshal(data)
	return strings.NewReader(string(j))
}

func (client rest) req(token string, method string, url string, contentType string, authUser string, authPassword string, data io.Reader, reply interface{}) Reply {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return Reply{0, err.Error(), ""}
	}

	if authUser != "" {
		req.SetBasicAuth(authUser, authPassword)
	} else {
		req.Header.Set("Authorization", "bearer "+token)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return Reply{0, err.Error(), ""}
	}

	var logger = logs.NewLogger("workloads.rest")

	resp_body, _ := ioutil.ReadAll(resp.Body)

	if TRACE_REST_CALLS {
		body := make(map[string]interface{})
		json.Unmarshal(resp_body, &body)
		logger.Debug1f(">> %s", body)
	}

	json.Unmarshal(resp_body, &reply)
	logger.Debug1f("%s %s %s", method, url, resp.Status)
	return Reply{resp.StatusCode, resp.Status, resp.Header.Get("Location")}
}
