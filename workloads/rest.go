package workloads

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"mime/multipart"
	"net/url"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/pat/context"
	"github.com/cloudfoundry-incubator/pat/config"
	"github.com/nu7hatch/gouuid"
)

type rest struct {
	username   string
	password   string
	target     string
	space_name string
	client     httpclient
}

func NewRestWorkload() *rest {
	ctx := &rest{}
	ctx.client = ctx
	return ctx
}

func NewRestWorkloadWithClient(client httpclient) *rest {
	ctx := &rest{}
	ctx.client = client
	return ctx
}

func (r *rest) DescribeParameters(config config.Config) {
	config.StringVar(&r.target, "rest:target", "", "the target for the REST api")
	config.StringVar(&r.username, "rest:username", "", "username for REST api")
	config.StringVar(&r.password, "rest:password", "", "password for REST api")
	config.StringVar(&r.space_name, "rest:space", "dev", "space to target for REST api")
}

func (r *rest) Target(ctx context.Context) error {
	if _, exists := ctx.GetString("cfTarget"); r.target == "" && exists { 
		r.target, _ = ctx.GetString("cfTarget")
	}

	body := &TargetResponse{}
	return r.GetSuccessfully("", r.target+"/v2/info", nil, body, func(reply Reply) error {
		ctx.PutString("loginEndpoint", body.LoginEndpoint)
		ctx.PutString("apiEndpoint", r.target)
		return nil
	})
}

func (r *rest) Login(ctx context.Context) error {
	body := &LoginResponse{}
	workerIndex, _ := ctx.GetInt("workerIndex")
	return checkTargetted(ctx, func(loginEndpoint string, apiEndpoint string) error {		
		return r.PostToUaaSuccessfully(fmt.Sprintf("%s/oauth/token", loginEndpoint), r.oauthInputs(r.credentialsForWorker(workerIndex)), body, func(reply Reply) error {
			ctx.PutString("token", body.Token)
			return r.targetSpace(ctx)
		})
	})
}

func (r *rest) targetSpace(ctx context.Context) error {
	apiEndpoint, _ := ctx.GetString("apiEndpoint")

	replyBody := &SpaceResponse{}
	return checkLoggedIn(ctx, func(token string) error {
		return r.GetSuccessfully(token, fmt.Sprintf("%s/v2/spaces?q=name:%s", apiEndpoint, r.space_name), nil, replyBody, func(reply Reply) error {
			return checkSpaceExists(replyBody, func() error {
				ctx.PutString("space_guid", replyBody.Resources[0].Metadata.Guid)
				return nil
			})
		})
	})
}

func (r *rest) Push(ctx context.Context) error {
	return checkLoggedIn(ctx, func(token string) error {
		return r.createAppSuccessfully(ctx, func(appUri string) error {
			return r.uploadAppBitsSuccessfully(ctx, appUri, func() error {
				return r.start(ctx, appUri, func() error {
					return r.trackAppStart(ctx, appUri)
				})
			})
		})
	})
}

func (r *rest) uploadAppBitsSuccessfully(ctx context.Context, appUri string, then func() error) error {
	apiEndpoint, _ := ctx.GetString("apiEndpoint")

	return checkLoggedIn(ctx, func(token string) error {
		return withGeneratedAppBits(func(b *bytes.Buffer, m *multipart.Writer) error {
			return r.MultipartPutSuccessfully(token, m, fmt.Sprintf("%s%s/bits", apiEndpoint, appUri), b, nil, func(reply Reply) error {
				return then()
			})
		})
	})
}

func (r *rest) start(ctx context.Context, appUri string, then func() error) error {
	apiEndpoint, _ := ctx.GetString("apiEndpoint")

	input := make(map[string]interface{})
	input["state"] = "STARTED"
	return checkLoggedIn(ctx, func(token string) error {
		return r.PutSuccessfully(token, fmt.Sprintf("%s%s", apiEndpoint, appUri), input, nil, func(reply Reply) error {
			return then()
		})
	})
}

func (r *rest) trackAppStart(ctx context.Context, appUri string) error {
	return checkLoggedIn(ctx, func(token string) error {
		apiEndpoint, _ := ctx.GetString("apiEndpoint")
		for {
			decoded := make(map[string]interface{})
			reply := r.client.Get(token, fmt.Sprintf("%s%s/instances", apiEndpoint, appUri), nil, &decoded)

			if reply.Code < 400 || decoded["error_code"] != "CF-NotStaged" {
				if decoded["error_code"] != nil {
					return errors.New("App Failed to Stage")
				}
				break
			}

			time.Sleep(2 * time.Second)
		}

		return nil
	})
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

func (r *rest) createAppSuccessfully(ctx context.Context, thenWithLocation func(appUri string) error) error {
	apiEndpoint, _ := ctx.GetString("apiEndpoint")
	space_guid, _ := ctx.GetString("space_guid")

	uuid, _ := uuid.NewV4()
	createApp := struct {
		Name      string `json:"name"`
		SpaceGuid string `json:"space_guid"`
	}{uuid.String(), space_guid}

	return checkLoggedIn(ctx, func(token string) error {
		return r.PostSuccessfully(token, fmt.Sprintf("%s/v2/apps", apiEndpoint), createApp, nil, func(reply Reply) error {
			return thenWithLocation(reply.Location)
		})
	})
}

func checkSpaceExists(s *SpaceResponse, then func() error) error {
	if !s.SpaceExists() {
		return errors.New("No space found with the given name")
	}

	return then()
}

func checkLoggedIn(ctx context.Context, then func(token string) error) error {
	if _, exists := ctx.GetString("token"); !exists {
		return errors.New("Error: not logged in")
	}

	token, _ := ctx.GetString("token")
	return then(token)
}

func checkTargetted(ctx context.Context, then func(loginEndpoint string, apiEndpoint string) error) error {
	if _, exists := ctx.GetString("loginEndpoint"); !exists {
		return errors.New("Not targetted")
	}

	if _, exists := ctx.GetString("apiEndpoint"); !exists {
		return errors.New("Not targetted")
	}

	apiEndpoint, _ := ctx.GetString("apiEndpoint")
	loginEndpoint, _ := ctx.GetString("loginEndpoint")
	return then(loginEndpoint, apiEndpoint)
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

func (r *rest) credentialsForWorker(workerIndex int) (string, string) {
	var userList = strings.Split(r.username, ",")
	var passList = strings.Split(r.password, ",")
	return userList[workerIndex % len(userList)], passList[workerIndex % len(passList)]
}

func (r *rest) oauthInputs(username string, password string) url.Values {

	values := make(url.Values)
	values.Add("grant_type", "password")
	values.Add("username", username)
	values.Add("password", password)
	values.Add("scope", "")

	return values
}
