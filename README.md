# HipChat Output for Mozilla Heka

A Mozilla Heka Output for Atlassian's HipChat.

# Installation

You will need an authentication token from Atlassian for HipChat's API. Follow [the directions](https://www.hipchat.com/docs/api/auth) in the HipChat API documentation.  A notification token will sufffice.

Create or add to the file {heka_root}/cmake/plugin_loader.cmake
```
add_external_plugin(git https://github.com/christianvozar/heka-hipchat-output master)
```

Then build Heka per normal instructions for your platform.

Additional instructions can be found in the [Heka documentation for external plugins](http://hekad.readthedocs.org/en/latest/installing.html#build-include-externals).

# Parameters

- payload_only (bool)
    If set to true, then only the message payload string will be included,
    otherwise the entire `Message` struct will be sent in JSON format. 
    (default: true)
- auth_token (string)
    HipChat Authorization token. Notification token is appropriate.
- send_to (array of strings)
    - array of email addresses to send the message to
- room_id (string)
    ID or name of the room to send message.
- from (string)
    Name the message will appear be sent. Must be less than 15 characters long. May contain letters, numbers, -, _, and spaces.
    (default: "heka")
- notify (bool, optional)
    Whether or not this message should trigger a notification for people in the room (change the tab color, play a sound, etc).
    Each recipient's notification preferences are taken into account.
    (default: false)
- color (string, optional)
    Background color of message within HipChat.
    (default: "gray")

## Example HipChat Output Configuration File

```
[HipchatOutput]
type = "HipchatOutput"
payload_only = false
auth_token = ""
room_id = "Notifications"
notify = true
color = "gray"
```
