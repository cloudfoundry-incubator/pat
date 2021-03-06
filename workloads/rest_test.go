package workloads_test

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/url"

	"github.com/cloudfoundry-incubator/pat/config"
	"github.com/cloudfoundry-incubator/pat/context"
	. "github.com/cloudfoundry-incubator/pat/workloads"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type workloads interface {
	Target(ctx context.Context) error
	Login(ctx context.Context) error
	Push(ctx context.Context) error
}

var restArgs = struct {
	username string
	password string
	target   string
	space    string
}{}

var _ = Describe("Rest", func() {
	Describe("Workloads", func() {
		var (
			client            *dummyClient
			rest              workloads
			args              []string
			replies           map[string]interface{}
			replyWithLocation map[string]string
			restContext       context.Context
		)

		BeforeEach(func() {
			restContext = context.New()
			restContext.PutInt("iterationIndex", 0)
			replies = make(map[string]interface{})
			replyWithLocation = make(map[string]string)
			client = &dummyClient{replies, replyWithLocation, make(map[call]interface{})}
			rest = NewRestWorkloadWithClient(client)
			config := config.NewConfig()
			initArgumentFlags(config)
			config.Parse(args)
			PopulateRestContext(restArgs.target, restArgs.username, restArgs.password, restArgs.space, restContext)
			args = []string{"-rest:target", "APISERVER"}

			replies["APISERVER/v2/info"] = TargetResponse{"THELOGINSERVER/PATH"}
		})

		Describe("Pushing an app", func() {
			Context("When the user has not logged in", func() {
				It("Returns an error", func() {
					err := rest.Push(restContext)
					Ω(err).Should(HaveOccurred())
				})
			})

			Context("After logging in", func() {
				BeforeEach(func() {
					replies["THELOGINSERVER/PATH/oauth/token"] = LoginResponse{"blah blah"}

					spaceReply := SpaceResponse{[]Resource{Resource{Metadata{"blah blah"}}}}
					replies["APISERVER/v2/spaces?q=name:dev"] = spaceReply

					replyWithLocation["APISERVER/v2/apps"] = "/THE-APP-URI"
					replies["APISERVER/THE-APP-URI"] = ""
					replies["APISERVER/THE-APP-URI/bits"] = ""

					err := rest.Target(restContext)
					Ω(err).ShouldNot(HaveOccurred())
					err = rest.Login(restContext)
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("Doesn't return an error", func() {
					err := rest.Push(restContext)
					Ω(err).ShouldNot(HaveOccurred())
				})

				It("POSTs a (random) name and the chosen space's guid", func() {
					rest.Push(restContext)
					data := client.ShouldHaveBeenCalledWith("POST", "APISERVER/v2/apps")
					m := mapOf(data)
					Ω(m).Should(HaveKey("name"))
					Ω(m).Should(HaveKey("space_guid"))
					Ω(m["space_guid"]).Should(Equal("blah blah"))
				})

				It("Uploads app bits", func() {
					rest.Push(restContext)
					data := client.ShouldHaveBeenCalledWith("PUT(multipart)", "APISERVER/THE-APP-URI/bits")
					Ω(data).ShouldNot(BeNil())
				})

				It("Starts the app", func() {
					rest.Push(restContext)
					data := mapOf(client.ShouldHaveBeenCalledWith("PUT", "APISERVER/THE-APP-URI"))
					Ω(data["state"]).Should(Equal("STARTED"))
				})

				Context("When the app starts immediately", func() {
					It("Doesn't return any error", func() {
						replies["APISERVER/THE-APP-URI/instances"] = "foo" // return a 200
						err := rest.Push(restContext)
						Ω(err).ShouldNot(HaveOccurred())
					})
				})

				Context("When the app status eventually returns CF-NotStaged", func() {
					PIt("Returns an error", func() {
					})
				})
			})
		})

		Describe("Logging in", func() {

			Context("When the API has been targetted", func() {
				JustBeforeEach(func() {
					rest.Target(restContext)
				})

				It("Can log in to the authorization endpoint", func() {
					rest.Login(restContext)
					client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
				})

				Context("When a username and password are configured", func() {
					BeforeEach(func() {
						args = []string{"-rest:target", "APISERVER", "-rest:space", "thespace", "-rest:username", "foo", "-rest:password", "bar"}
					})

					JustBeforeEach(func() {
						rest.Login(restContext)
					})

					It("sets grant_type password", func() {
						data := client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["grant_type"]).Should(Equal([]string{"password"}))
					})

					It("POSTs the username and password", func() {
						data := client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["username"]).Should(Equal([]string{"foo"}))
						Ω(data.(url.Values)["password"]).Should(Equal([]string{"bar"}))
					})

					It("sets empty scope", func() {
						data := client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["scope"]).Should(Equal([]string{""}))
					})

					Context("And the login is successful", func() {
						BeforeEach(func() {
							replies["THELOGINSERVER/PATH/oauth/token"] = struct {
								AccessToken string `json:"access_token"`
							}{"blah blah"}

							spaceReply := SpaceResponse{[]Resource{Resource{Metadata{"blah blah"}}}}
							replies["APISERVER/v2/spaces?q=name:thespace"] = spaceReply
						})

						It("Does not return an error", func() {
							err := rest.Login(restContext)

							Ω(err).ShouldNot(HaveOccurred())
						})

						Context("But when the space does not exist", func() {
							BeforeEach(func() {
								replies["APISERVER/v2/spaces?q=name:thespace"] = nil
							})

							It("Returns an error", func() {
								err := rest.Login(restContext)
								Ω(err).Should(HaveOccurred())
							})
						})
					})

					Context("And the login is not successful", func() {
						BeforeEach(func() {
							replies["THELOGINSERVER/path/oauth/token"] = nil
						})

						It("Does not return an error", func() {
							err := rest.Login(restContext)
							Ω(err).Should(HaveOccurred())
						})
					})
				})

				Context("When multiple usernames and passwords are configured", func() {
					BeforeEach(func() {
						args = []string{"-rest:target", "APISERVER", "-rest:space", "thespace", "-rest:username", "user1,user2,user3", "-rest:password", "pass1,pass2"}
					})

					JustBeforeEach(func() {
						rest.Login(restContext)
					})

					It("sets grant_type password", func() {
						data := client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["grant_type"]).Should(Equal([]string{"password"}))
					})

					It("POSTs the username and password", func() {
						data := client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["username"]).Should(Equal([]string{"user1"}))
						Ω(data.(url.Values)["password"]).Should(Equal([]string{"pass1"}))
					})

					It("uses different username and password with different iterationIndex", func() {
						restContext.PutInt("iterationIndex", 0)
						rest.Login(restContext)
						data := client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["username"]).Should(Equal([]string{"user1"}))
						Ω(data.(url.Values)["password"]).Should(Equal([]string{"pass1"}))

						restContext.PutInt("iterationIndex", 2)
						rest.Login(restContext)
						data = client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["username"]).Should(Equal([]string{"user3"}))
						Ω(data.(url.Values)["password"]).Should(Equal([]string{"pass1"}))
					})

					It("recycles the list of username and password when there are more workers than username", func() {
						restContext.PutInt("iterationIndex", 6)
						rest.Login(restContext)
						data := client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["username"]).Should(Equal([]string{"user1"}))
						Ω(data.(url.Values)["password"]).Should(Equal([]string{"pass1"}))
					})
				})

				Context("When multiple usernames and a single password are configured", func() {
					BeforeEach(func() {
						args = []string{"-rest:target", "APISERVER", "-rest:space", "thespace", "-rest:username", "user1,user2,user3", "-rest:password", "pass1"}
					})

					JustBeforeEach(func() {
						rest.Login(restContext)
					})

					It("sets grant_type password", func() {
						data := client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["grant_type"]).Should(Equal([]string{"password"}))
					})

					It("re-uses the only avaiable password", func() {
						restContext.PutInt("iterationIndex", 0)
						rest.Login(restContext)
						data := client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["username"]).Should(Equal([]string{"user1"}))
						Ω(data.(url.Values)["password"]).Should(Equal([]string{"pass1"}))

						restContext.PutInt("iterationIndex", 1)
						rest.Login(restContext)
						data = client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["username"]).Should(Equal([]string{"user2"}))
						Ω(data.(url.Values)["password"]).Should(Equal([]string{"pass1"}))

						restContext.PutInt("iterationIndex", 2)
						rest.Login(restContext)
						data = client.ShouldHaveBeenCalledWith("POST(uaa)", "THELOGINSERVER/PATH/oauth/token")
						Ω(data.(url.Values)["username"]).Should(Equal([]string{"user3"}))
						Ω(data.(url.Values)["password"]).Should(Equal([]string{"pass1"}))
					})
				})

			})

			Describe("When the API hasn't been targetted yet", func() {
				It("Will return an error", func() {
					err := rest.Login(restContext)
					Ω(err).To(HaveOccurred())
				})
			})
		})
	})

	Describe("HTTP calls", func() {
		Context("populate an interface with http body responses", func() {
			var (
				replies           map[string]interface{}
				replyWithLocation map[string]string
				client            *dummyClient
			)

			BeforeEach(func() {
				replies = make(map[string]interface{})
				replyWithLocation = make(map[string]string)
				client = &dummyClient{replies, replyWithLocation, make(map[call]interface{})}
			})

			It("GET /v2/info aquires the LoginEndpoint", func() {
				var targetResponse TargetResponse
				client.Get("", "", nil, &targetResponse)
				Ω(targetResponse.LoginEndpoint).Should(Equal("10.244.0.34.xip.io"))
			})

			It("GET /v2/spaces aquires the SpaceResponse", func() {
				var spaceResponse SpaceResponse
				client.Get("", "", nil, &spaceResponse)
				Ω(spaceResponse.Resources[0].Metadata.Guid).Should(Equal("123456789"))
			})

			It("POST to UAA aquires the authentication token", func() {
				var loginResponse LoginResponse
				client.Post("", "", nil, &loginResponse)
				Ω(loginResponse.Token).Should(Equal("token"))
			})
		})
	})
})

type dummyClient struct {
	replies           map[string]interface{}
	replyWithLocation map[string]string
	calls             map[call]interface{}
}

type call struct {
	method string
	path   string
}

func (d *dummyClient) ShouldHaveBeenCalledWith(method string, path string) interface{} {
	Ω(d.calls).Should(HaveKey(call{method, path}))
	return d.calls[call{method, path}]
}

func (d *dummyClient) Req(method string, host string, data interface{}, s interface{}) (reply Reply) {
	resp := `{"authorization_endpoint":"10.244.0.34.xip.io","access_token":"token","guid":"123456789","metadata":{"guid":"123456789"},"resources":[{"metadata":{"guid":"123456789"}}]}`

	json.Unmarshal([]byte(resp), &s)

	d.calls[call{method, host}] = data
	if d.replyWithLocation[host] != "" {
		return Reply{201, "Moved", d.replyWithLocation[host]}
	}
	if d.replies[host] == nil {
		return Reply{400, "Some error", ""}
	}
	b, _ := json.Marshal(d.replies[host])
	json.NewDecoder(bytes.NewReader(b)).Decode(s)
	return Reply{200, "Success", ""}
}

func (d *dummyClient) Get(token string, host string, data interface{}, s interface{}) (reply Reply) {
	return d.Req("GET", host, data, s)
}

func (d *dummyClient) MultipartPut(token string, m *multipart.Writer, host string, data *bytes.Buffer, s interface{}) (reply Reply) {
	return d.Req("PUT(multipart)", host, data, s)
}

func (d *dummyClient) Put(token string, host string, data interface{}, s interface{}) (reply Reply) {
	return d.Req("PUT", host, data, s)
}

func (d *dummyClient) Post(token string, host string, data interface{}, s interface{}) (reply Reply) {
	return d.Req("POST", host, data, s)
}

func (d *dummyClient) PostToUaa(host string, data url.Values, s interface{}) (reply Reply) {
	return d.Req("POST(uaa)", host, data, s)
}

func mapOf(data interface{}) map[string]interface{} {
	d, _ := json.Marshal(data)
	m := make(map[string]interface{})
	json.Unmarshal(d, &m)
	return m
}

func initArgumentFlags(config config.Config) {
	config.StringVar(&restArgs.target, "rest:target", "", "the target for the REST api")
	config.StringVar(&restArgs.username, "rest:username", "", "username for REST api")
	config.StringVar(&restArgs.password, "rest:password", "", "password for REST api")
	config.StringVar(&restArgs.space, "rest:space", "dev", "space to target for REST api")
}
