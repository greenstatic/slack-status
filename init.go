package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"time"
)

const (
	defaultOAuthScopes = "dnd:write,users.profile:write,users:write"
)

var initCmd = &cobra.Command{
	Use:   "init <workspace name> <Slack App clientId> <Slack App clientSecret>",
	Short: "Initialize slack-status in a workspace",
	Long: `Initializes slack-status for a user in the defined workspace.
For this you will be required to use an existing or create a Slack App in your workspace.

Make sure that under "OAuth & Permissions" you have set the Redirect URL the same as the one
you will use with the CLI (recommended: http://localhost:3030 - this will partially match the one
we use by default)

See https://github.com/greenstatic/slack-status for more detailed instructions.`,
	Args: cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]
		clientId := args[1]
		clientSecret := args[2]
		redirectUrl := _stringOrPanic(cmd.PersistentFlags().GetString("redirectUri"))

		configPath := GetConfigFilepath(_stringOrPanic(cmd.Root().PersistentFlags().GetString("config")))
		c := Config{}
		if err := c.Parse(configPath); err != nil {
			log.Printf("No config file exists: %s", err)
		}

		log.Printf("Initializing slack-status for workspace %s with the clientId:%s", workspaceName, clientId)
		log.Printf("Visit URL to trigger OAuth2 flow:")

		fmt.Println("")
		fmt.Println(initOAuth2V2FlowURL(clientId,
			_stringOrPanic(cmd.PersistentFlags().GetString("scopes")),
			redirectUrl))
		fmt.Println("")

		tempCode := oAuth2HttpServer(_stringOrPanic(cmd.PersistentFlags().GetString("httpBind")))
		user, accessToken, err := oAuth2TemporaryAuthorizationCodeForAuthorizationCode(clientId, clientSecret,
			tempCode, redirectUrl)

		if err != nil {
			log.Fatalf("Failed to convert temporary authorization code for authorization code, error: %s", err)
		}

		w := Workspace{
			Name:        workspaceName,
			User:        user,
			AccessToken: accessToken,
		}
		c.Workspaces = append(c.Workspaces, w)

		if err := c.Save(configPath); err != nil {
			log.Fatalf("Failed to save config file, error: %s", err)
		}

		log.Printf("Successfully saved authorization token to slack-status config %s", configPath)
	},
}

func init() {
	initCmd.PersistentFlags().String("httpBind", "127.0.0.1:3030", "HTTP Bind details, this has to match redirectUri port")
	initCmd.PersistentFlags().String("scopes", defaultOAuthScopes, "V2 OAuth2 Slack scopes")
	initCmd.PersistentFlags().String("redirectUri", "http://localhost:3030/redirect", "OAuth2 redirect_uri, has to match httpBind port")
}

func initOAuth2V2FlowURL(clientId, scopes, redirectUri string) string {
	return fmt.Sprintf("https://slack.com/oauth/v2/authorize?user_scope=%s&client_id=%s&redirect_uri=%s",
		scopes, clientId, redirectUri)
}

func oAuth2HttpServer(bindHttp string) string {
	log.Printf("Starting HTTP server on: %s", bindHttp)

	temporaryAuthorizationCode := ""
	done := make(chan bool)

	s := http.Server{Addr: bindHttp}

	go func() {
		http.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("HTTP OAuth2 redirect handler received: %s \nbody: %s", r.URL.String(), r.Body)

			if code := r.URL.Query().Get("code"); code == "" {
				log.Print("Error: HTTP OAuth2 redirect handler did not receive code parameter")
				w.WriteHeader(400)
				w.Write(nil)
			} else {
				log.Printf("HTTP OAuth2 redirect handler successfully received Temporary Authorization Code: %s", code)
				fmt.Fprintln(w, "Successfully got Temporary Authorization Code, you may close this tab.")
				temporaryAuthorizationCode = code
				done <- true
			}
		})

		s.ListenAndServe()
	}()

	<-done
	log.Printf("Successfully got Temporary Authorization Code, shutting down HTTP server")

	if err := s.Close(); err != nil {
		log.Panic(err)
	}

	return temporaryAuthorizationCode
}

func oAuth2TemporaryAuthorizationCodeForAuthorizationCode(clientId, clientSecret, temporaryAuthorizationCode,
	redirectUri string) (user string, accessToken string, err error) {

	httpClient := http.Client{Timeout: 10 * time.Second}
	resp, err := slack.GetOAuthV2Response(&httpClient, clientId, clientSecret, temporaryAuthorizationCode, redirectUri)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to convert temporary authorization code for authorization code")
	}

	if resp == nil {
		return "", "", errors.Wrap(err, "response is empty")
	}

	if resp.AuthedUser.TokenType != "user" {
		return "", "", errors.Wrap(err, "response does not contain user auth token")
	}

	return resp.AuthedUser.ID, resp.AuthedUser.AccessToken, nil
}
