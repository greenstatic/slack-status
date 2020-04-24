# Slack Status
Change your presence, status and do not disturb settings in multiple Slack workspaces.

This is a complete re-write of [dstokes/slack-status](https://github.com/dstokes/slack-status).

## Install
Dependencies:
* Go

1. Build from source yourself using go:
    ```shell
    go get github.com/greenstatic/slack-status
    ```
1. Create your config (see example bellow) in `~/slack-status`

## Usage
```
Set your status in Slack.

Running with no arguments will cause your status to be cleared.
When enabling do not disturb (dnd) you must specify a duration.

Source code: https://github.com/greenstatic/slack-status

Usage:
  slack-status [flags]

Flags:
      --away               Set your status as away
  -c, --config string      Config file (default "~/slack-status")
      --dnd                Set status as do not disturb
  -d, --duration string    Set status duration, units can be: [m,h]. Leave blank for for no expiration
  -e, --emoji string       Emoji to set when setting your status (default ":male-technologist:")
  -g, --group string       Limit setting of status to a group
  -h, --help               help for slack-status
  -m, --message string     Status message
  -v, --version            Show version number
  -w, --workspace string   Limit setting of status to a workspace
```

### Examples
Set your custom status to _busy working on super secret project_:
```shell
slack-status -m "busy working on super secret project"
```

Set your custom status to _programming_ for 45 min and turn on do not disturb for your *my_job* Slack workspace:
```shell
slack-status -m "programming" -d 45m --dnd --workspace my_job
```

Set your custom status to _lectures_ with custom books emoji for 2h 45m and turn on do not disturb for your _work_ group Slack workspaces:
```shell
slack-status -m "lectures" -d 2h45m --dnd --emoji ":books:" -g work
```

Clears your status and sets it back to _Active_: 
```shell
slack-status
```

### Configuration
By default `slack-status` will open `~/slack-status` but you can override this behaviour by setting the path to the config explicitly using the `--config flag.
 ```yaml
# Example slack-status config
workspaces:
  - name: my_job
    accessToken: "<ACCESS TOKEN>"
    groups:
      - "work"
  - name: home
    accessToken: "<ACCESS TOKEN>"
    groups:
      - "personal"
```

#### Acquire Access Token for Workspace
The following instructions outline how to acquire an access token with which _slack-status_ can set your status for a particular Slack workspace.

1. Visit [https://api.slack.com/apps](https://api.slack.com/apps)
2. Click on *Create New App*. 
    Name the app something like _slack-status_, select your desired workspace and create your app.
3. Under the secion _Add features and functionality_ of your app select _Permissions_.
    Scroll down to the _Scopes_ section and add the following *User Token Scopes*:
    * dnd:write
    * users.profile:write
    * users:write
4. Once your have all the required scopes set, scroll to the top and select _Install App to Workspace_.
5. On the redirected page click _Allow_.
6. You will get a _OAuth Access Token_, it should start with `xoxp-`.
    This is the access token that you will need for slack-status workspace access token.

You will need to do this for each Slack workspace you wish to specify control using _slack-status_.