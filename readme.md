# RSS Notification Bot & API

Created by: 

- Morten Tingstad Spjøtvoll	(mortespj, 257780)

- André Gunhildberget (andregg, 493561)

- Jan Zimmer (janzim, 493594)

#
Discord bot invite link: https://discordapp.com/oauth2/authorize?client_id=506877297182113798&permissions=8&scope=bot

# Features 
## Discord bot

- Subscribe to a new RSS feed using a keyword (e.g. bbc) or link (e.g. http://feeds.bbci.co.uk/news/rss.xml):
```
!newrss <keyword/link>
```

- If more than one RSS link was found by the bot, you can choose feeds by:
```
!addrss <number/s>
```
(Note: when choosing multiple numbers use space inbetween each number. E.g. "!addrss 3 7 19")

- Remove a RSS subscription.
```
!remrss <link/number>
```
(Note: if link was not provided you will first receive a list and can then choose what feeds to remove)

- Get a list of all feeds your server is subscribed to
```
!listrss
```

- Set a new channel where the bot should post RSS updates. Default: First channel in server.
```
!configure <channel id/name>
```

- Get a list of commands (Similiar to the ones here)
```
!commands
```

### For API

- Get a new API key to use for the webAPI. Only the server owner can do this.
```
!newkeyrss
```
(NOTE: this will replace the old one!)

- Get the current API key. Only the server owner can do this.
```
!getkeyrss
```

## Web API

### Paths

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

#### How to get an API key

See "For API" under Discord bot

# Original Project Plan

Our plan was to create a RSS subscription service with discord. The user invites a bot and can then choose what RSS feeds they want to subscribe to by typing a keyword (e.g. nrk) or the whole RSS link. Once the RSS feed updates, the user will get a notification on discord. In addition we'll have a web API where you're able to POST new RSS links and GET RSS subscriptions. We will also provide a simple GUI on Heroku where the admin is able to customize the bot and add new RSS links from there as well. This will mostly work as a site where you can configure your bot.

## What has/has not been achieved

We've followed through with most of what we've planned, but decided to omit the feature of having a GUI on heroku as we found out that there isn't much to configure which you couldn't simply do on discord so we discarded this idea.

# Learning experience and project hardships 

It was planned that this service could be used by many discord servers at once, but we found out that this kind of scalability was difficult to achieve. As the project grew, it became more time consuming adding new features or fixing bugs that existed at the core system (for example implementing multithreading). 

# Work hours


| Person        | Work hours    |
| ------------- |:-------------:|
| mortespj      |               |
| andregg       |               |
| janzim        |               |


