package workloads

import (
	"archive/zip"
	"bytes"
	"errors"
	"mime/multipart"
	"net/url"
	"time"

	"github.com/julz/pat/config"
	"github.com/nu7hatch/gouuid"
)

type context struct {
	username      string
	password      string
	target        string
	loginEndpoint string
	apiEndpoint   string
	space_name    string
	space_guid    string
	token         string
	client        httpclient
}

func NewRestWorkloadContext() *context {
	ctx := &context{}
	ctx.client = ctx
	return ctx
}

func NewContext(client httpclient) *context {
	ctx := &context{}
	ctx.client = client
	return ctx
}

func (c *context) DescribeParameters(config config.Config) {
	config.StringVar(&c.target, "rest:target", "", "the target for the REST api")
	config.StringVar(&c.username, "rest:username", "", "username for REST api")
	config.StringVar(&c.password, "rest:password", "", "password for REST api")
	config.StringVar(&c.space_name, "rest:space", "dev", "space to target for REST api")
}

func (context *context) Target() error {
	body := &TargetResponse{}
	return context.GetSuccessfully(context.target+"/v2/info", nil, body, func(reply Reply) error {
		context.loginEndpoint = body.LoginEndpoint
		context.apiEndpoint = context.target
		return nil
	})
}

func (context *context) Login() error {
	body := &LoginResponse{}
	return checkTargetted(context, func() error {
		return context.PostToUaaSuccessfully(context.loginEndpoint+"/oauth/token", context.oauthInputs(), body, func(reply Reply) error {
			context.token = body.Token
			return context.targetSpace()
		})
	})
}

func (context *context) targetSpace() error {
	replyBody := &SpaceResponse{}
	return context.GetSuccessfully(context.apiEndpoint+"/v2/spaces?q=name:"+context.space_name, nil, replyBody, func(reply Reply) error {
		return checkSpaceExists(replyBody, func() error {
			context.space_guid = replyBody.Resources[0].Metadata.Guid
			return nil
		})
	})
}

func (context *context) Push() error {
	return checkLoggedIn(context, func() error {
		return context.createAppSuccessfully(func(appUri string) error {
			return context.uploadAppBitsSuccessfully(appUri, func() error {
				return context.start(appUri, func() error {
					return context.trackAppStart(appUri)
				})
			})
		})
	})
}

func (context *context) uploadAppBitsSuccessfully(appUri string, then func() error) error {
	return withGeneratedAppBits(func(b *bytes.Buffer, m *multipart.Writer) error {
		return context.MultipartPutSuccessfully(m, context.apiEndpoint+appUri+"/bits", b, nil, func(reply Reply) error {
			return then()
		})
	})
}

func (context *context) start(appUri string, then func() error) error {
	input := make(map[string]interface{})
	input["state"] = "STARTED"
	return context.PutSuccessfully(context.apiEndpoint+appUri, input, nil, func(reply Reply) error {
		return then()
	})
}

func (context *context) trackAppStart(appUri string) error {
	for {
		decoded := make(map[string]interface{})
		reply := context.client.Get(context.apiEndpoint+appUri+"/instances", nil, &decoded)

		if reply.Code < 400 || decoded["error_code"] != "CF-NotStaged" {
			if decoded["error_code"] != nil {
				return errors.New("App Failed to Stage")
			}
			break
		}

		time.Sleep(2 * time.Second)
	}

	return nil
}

func withGeneratedAppBits(fn func(b *bytes.Buffer, m *multipart.Writer) error) error {
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

	return fn(&b, multi)
}

func (context *context) createAppSuccessfully(thenWithLocation func(appUri string) error) error {
	uuid, _ := uuid.NewV4()
	createApp := struct {
		Name      string `json:"name"`
		SpaceGuid string `json:"space_guid"`
	}{uuid.String(), context.space_guid}

	return context.PostSuccessfully(context.apiEndpoint+"/v2/apps", createApp, nil, func(reply Reply) error {
		return thenWithLocation(reply.Location)
	})
}

func checkSpaceExists(s *SpaceResponse, then func() error) error {
	if !s.SpaceExists() {
		return errors.New("No space found with the given name")
	}

	return then()
}

func checkLoggedIn(c *context, then func() error) error {
	if c.token == "" {
		return errors.New("Error: not logged in")
	}

	return then()
}

func checkTargetted(context *context, then func() error) error {
	if context.loginEndpoint == "" {
		return errors.New("Not targetted")
	}

	return then()
}

func checkSuccessfulReply(reply Reply, then func() error) error {
	if err := reply.checkError(); err != nil {
		return err
	}

	return then()
}

func (r Reply) checkError() error {
	if r.Code > 399 {
		return errors.New(r.Message)
	}

	return nil
}

func (s SpaceResponse) SpaceExists() bool {
	return len(s.Resources) > 0
}

func (context *context) oauthInputs() url.Values {
	values := make(url.Values)
	values.Add("grant_type", "password")
	values.Add("username", context.username)
	values.Add("password", context.password)
	values.Add("scope", "")

	return values
}
