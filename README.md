# API

## Paths

The root url is:

```
http://rss-notification.herokuapp.com/
```

There are currently three available requests in the API.

- (Get) List all registered RSS URLs
```
http://rss-notification.herokuapp.com/api/rss
```

- (Get) List your servers subscriptions
```
http://rss-notification.herokuapp.com/api/rss/{apiKey}
```

- (Post) Add a new RSS subscription to your server
```
http://rss-notification.herokuapp.com/api/{apiKey}
```

## How to get an API key

To get a new API key you must be the server owner of a discord server with the bot running on it. Then you'll have two relevant commands at your disposal.

- Get a new API key to use for the webAPI.
```
!newkeyrss
```
(NOTE: this will replace the old one!)

- Get the current API key. 
```
!getkeyrss
```
