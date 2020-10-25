# Slack Status
Change your presence, status, do not disturb and profile picture settings in multiple Slack workspaces using only 
one command.

## How it Works
In order to integrate with Slack, each integration needs a dedicated Slack App, meaning you are required to create
a dedicated Slack App for slack-status to work in your workspace.
Each Slack App receives a _clientId_ and _clientSecret_ - details with which we can initiate the OAuth2 flow.
Through this flow, an app can get certain permissions, as well as certain permissions to act on a user's behalf. 
With slack-status we are particularly interested in getting our app to acquire permissions to act on a user's behalf,
so that we can change the user's status via Slack's API.
When we go through the OAuth2 flow, slack-status specifies by default that it only needs access to a couple of user 
scopes ([dnd:write, users.profile:write & users:write](https://api.slack.com/scopes)).
These scopes are tied to a user, so we get a **dedicated user authorization token** with which we can issue 
API calls to change the user's status.
The **user authorization token is stored strictly in a configuration file locally** (by default in `~/.slack-status`).

Free Slack workspaces are 
[limited to 10 apps](https://slack.com/intl/en-si/help/articles/115002422943-Message-file-and-app-limits-on-the-free-version-of-Slack),
in order to reduce the number of apps in a workspace it is possible to share Slack App credentials (clientId and 
clientSecret) so that multiple people can use slack-status in your workspace.
The token used to change a user's status, the **user authorization token**, is **stored only locally** and
the Slack App that is created for slack-status only requires user scopes, getting access to the Slack App's 
clientId and clientSecret will not enable other user's to change your status.

The clientId and clientSecret are merely necessary to trigger the OAuth2 flow with which we can get the user's user 
authorization token.
We do not need any other scopes which could cause security issues.

## Install
Dependencies:
* Go

Build from source:
```shell
$ go install github.com/greenstatic/slack-status
```
### Installing in a Slack Workspace
In case you are worried about all these tokens, please read the [How it Works](#how-it-works) section to understand 
what is going on.

You will need to do this for each Slack workspace you wish to specify control using _slack-status_.

#### I do not yet have a Slack App (or access to an App's clientId & clientSecret)
1. Visit [https://api.slack.com/apps](https://api.slack.com/apps)
2. Click on **Create New App**. 
   Name the app something like _slack-status_, select your desired workspace and create your app.
3. Scroll to the top and select _Install App to Workspace_.
4. Once you are taken to the main menu of the app, on the left side of the menu under **Features**, select 
   **OAuth & Permissions**. Under **Redirect URLs** add `http://localhost:3030`.

#### I have a Slack App now
Now that you have access to your Slack's App `clientId` and `clientSecret`, run:
```shell
$ slack-status init <workspace name> <clientId> <clientSecret>
```

This should trigger the OAuth2 flow. 
You should see a URL in the console, visit it your browser (that is logged into your slack workspace).
You can review the scopes we are asking for, just click confirm.

Once you click confirm you should see a simple page saying that your User Authorization Token has been successfully required.
You may close the browser tab now.

That's it.
In your config file (default: ~/.slack-status) there should be a new entry for the workspace you just initialized.
You may add a list of `group` names to the configuration workspace entry if you like.

Try to set your status now:
```shell
$ slack-status set -w <workspace name> -m "i'm using slack-status!" --emoji :tada: -d 5m
```

## Usage
```
Set your status in Slack.

To setup the required authentication token(s) read the "init" subcommand help.

To set your status see the "set" subcommand.

Source code: https://github.com/greenstatic/slack-status

Usage:
  slack-status [flags]
  slack-status [command]

Available Commands:
  help        Help about any command
  init        Initialize slack-status in a workspace
  set         Set a custom slack status
  version     Show version

Flags:
  -c, --config string   Config file (default "~/.slack-status")
  -h, --help            help for slack-status

Use "slack-status [command] --help" for more information about a command.
```

### Examples
Set your custom status to _busy working on super secret project_:
```shell
slack-status set -m "busy working on super secret project"
```

Set your custom status to _programming_ for 45 min and turn on do not disturb for your *my_job* Slack workspace:
```shell
slack-status set -m "programming" -d 45m --dnd --workspace my_job
```

Set your custom status to _lectures_ with custom books emoji for 2h 45m and turn on do not disturb for your _work_ group Slack workspaces:
```shell
slack-status set -m "lectures" -d 2h45m --dnd --emoji ":books:" -g work
```

Clears your status and sets it back to _Active_: 
```shell
slack-status set
```

Set your custom status to _driving_ and custom car emoji with new profile picture in _~/driving.jpeg_:
```shell
slack-status set -p ~/driving.jpeg -m "driving" --emoji :car:
```

### Configuration
By default `slack-status` will create a config in `~/.slack-status` but you can override this behaviour by setting the 
path to the config explicitly using the `--config` flag.
 ```yaml
# Example slack-status config
workspaces:
  - name: my_job
    user: "<USER ID>"
    accessToken: "<USER ACCESS TOKEN>"
    groups:
      - "work"
  - name: home
    user: "<USER ID>"
    accessToken: "<USER ACCESS TOKEN>"
    groups:
      - "personal"
```

In general, you do not need to change anything, but you are encouraged to set your `group` fields to your liking.
