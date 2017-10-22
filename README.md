# go-ms-graph

> Client library for [Microsoft Graph API][graph-api].

#### *Alpha Software Warning*
This is not production ready! Do not depend on this for production code! This
was written to scratch an itch I had to access the outlook API. I'm releasing it
to save people time working through setting up the auth.


### Usage Example

#### Authentication
Setup a Microsoft App according [to these directions][app-direction]. You need a client secret and a client

```go
package main

import(
    "github.com/mcordell/go-ms-graph/auth"
    "fmt"
)

func main() {
    authConfig := auth.DefaultConfig
    // client id from your setup
    authConfig.ClientID = ClientID
    // client secret from your setup
    authConfig.ClientSecret = x.ClientSecret

    // Preform one time login
    authCode := auth.LoginRequest(authConfig)
    t, err := auth.GetTokens(authConfig, authCode, "user.read mail.read calendars.readwrite")
    if err != nil {
        panic(err)
    }

    // t contains refresh/access tokens
    fmt.Fprintf("Token: %s", t.AccessToken)
}
```

[graph-api]: https://developer.microsoft.com/en-us/graph
[app-direction]: https://developer.microsoft.com/en-us/graph/docs/concepts/auth_v2_user#1-register-your-app
