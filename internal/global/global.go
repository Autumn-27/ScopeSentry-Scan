// config-------------------------------------
// @file      : global.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/9/8 12:33
// -------------------------------------------

package global

import (
	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"regexp"
)

var (
	// AbsolutePath 全局变量
	AbsolutePath string
	ConfigPath   string
	ConfigDir    string
	// AppConfig Global variable to hold the loaded configuration
	AppConfig             Config
	DisallowedURLFilters  []*regexp.Regexp
	VERSION               string
	FirstRun              bool
	DictPath              string
	ExtDir                string
	SensitiveRules        []types.SensitiveRule
	Projects              []types.Project
	WebFingers            []types.WebFinger
	NotificationApi       []types.NotificationApi
	NotificationConfig    types.NotificationConfig
	PocDir                string
	SubdomainTakerFingers []types.SubdomainTakerFinger
	TakeoverFinger        = []byte(`[
    {
        "name": "github",
        "cname": [
            "github.io",
            "github.map.fastly.net"
        ],
        "response": [
            "There isn't a GitHub Pages site here.",
            "For root URLs (like http://example.com/) you must provide an index.html file"
        ]
    },
    {
        "name": "Agile CRM",
        "cname": [
            "agilecrm.com"
        ],
        "response":["Sorry, this page is no longer available."]
    },
    {
        "name":"Airee.ru",
        "cname":["airee.ru"],
        "response":["\u041e\u0448\u0438\u0431\u043a\u0430 402. \u0421\u0435\u0440\u0432\u0438\u0441 \u0410\u0439\u0440\u0438.\u0440\u0444 \u043d\u0435 \u043e\u043f\u043b\u0430\u0447\u0435\u043d"]
    },
    {
        "name":"Anima",
        "cname":["animaapp.io"],
        "response":["The page you were looking for does not exist."]
    },
    {
        "name": "heroku",
        "cname": [
            "herokudns.com",
            "herokussl.com",
            "herokuapp.com"
        ],
        "response": [
            "There's nothing here, yet.",
            "herokucdn.com/error-pages/no-such-app.html",
            "<title>No such app</title>"
        ]
    },
    {
        "name": "unbounce",
        "cname": [
            "unbouncepages.com"
        ],
        "response": [
            "Sorry, the page you were looking for doesn’t exist.",
            "The requested URL was not found on this server"
        ]
    },
    {
        "name": "tumblr",
        "cname": [
            "tumblr.com"
        ],
        "response": [
            "There's nothing here.",
            "Whatever you were looking for doesn't currently exist at this address."
        ]
    },
    {
        "name": "shopify",
        "cname": [
            "myshopify.com"
        ],
        "response": [
            "Sorry, this shop is currently unavailable.",
            "Only one step left!"
        ]
    },
    {
        "name": "instapage",
        "cname": [
            "pageserve.co",
            "secure.pageserve.co",
            "https://instapage.com/"
        ],
        "response": [
            "Looks Like You're Lost",
            "The page you're looking for is no longer available."
        ]
    },
    {
        "name": "desk",
        "cname": [
            "desk.com"
        ],
        "response": [
            "Please try again or try Desk.com free for 14 days.",
            "Sorry, We Couldn't Find That Page"
        ]
    },
    {
        "name": "campaignmonitor",
        "cname": [
            "createsend.com",
            "name.createsend.com"
        ],
        "response": [
            "Double check the URL",
            "<strong>Trying to access your account?</strong>"
        ]
    },
    {
        "name": "cargocollective",
        "cname": [
            "cargocollective.com"
        ],
        "response": [
            "404 Not Found"
        ]
    },
    {
        "name": "statuspage",
        "cname": [
            "statuspage.io"
        ],
        "response": [
            "Better Status Communication",
            "You are being <a href=\"https://www.statuspage.io\">redirected"
        ]
    },
    {
        "name": "amazonaws",
        "cname": [
            "amazonaws.com"
        ],
        "response": [
            "NoSuchBucket",
            "The specified bucket does not exist"
        ]
    },
    {
        "name": "bitbucket",
        "cname": [
            "bitbucket.org",
            "bitbucket.io"
        ],
        "response": [
            "The page you have requested does not exist",
            "Repository not found"
        ]
    },
    {
        "name":"Gemfury",
        "cname":["furyns.com"],
        "response":["404: This page could not be found."]
    },
    {
        "name": "smartling",
        "cname": [
            "smartling.com"
        ],
        "response": [
            "Domain is not configured"
        ]
    },
    {
        "name": "acquia",
        "cname": [
            "acquia.com"
        ],
        "response": [
            "If you are an Acquia Cloud customer and expect to see your site at this address",
            "The site you are looking for could not be found."
        ]
    },
    {
        "name": "fastly",
        "cname": [
            "fastly.net"
        ],
        "response": [
            "Please check that this domain has been added to a service",
            "Fastly error: unknown domain"
        ]
    },
    {
        "name": "pantheon",
        "cname": [
            "pantheonsite.io"
        ],
        "response": [
            "The gods are wise",
            "The gods are wise, but do not know of the site which you seek."
        ]
    },
    {
        "name": "zendesk",
        "cname": [
            "zendesk.com"
        ],
        "response": [
            "Help Center Closed"
        ]
    },
    {
        "name": "uservoice",
        "cname": [
            "uservoice.com"
        ],
        "response": [
            "This UserVoice subdomain is currently available!"
        ]
    },
    {
        "name": "ghost",
        "cname": [
            "ghost.io"
        ],
        "response": [
            "The thing you were looking for is no longer here",
            "The thing you were looking for is no longer here, or never was"
        ]
    },
    {
        "name": "pingdom",
        "cname": [
            "stats.pingdom.com"
        ],
        "response": [
            "pingdom"
        ]
    },
    {
        "name": "tilda",
        "cname": [
            "tilda.ws"
        ],
        "response": [
            "Domain has been assigned"
        ]
    },
    {
        "name": "wordpress",
        "cname": [
            "wordpress.com"
        ],
        "response": [
            "Do you want to register"
        ]
    },
    {
        "name": "teamwork",
        "cname": [
            "teamwork.com"
        ],
        "response": [
            "Oops - We didn't find your site."
        ]
    },
    {
        "name": "helpjuice",
        "cname": [
            "helpjuice.com"
        ],
        "response": [
            "We could not find what you're looking for."
        ]
    },
    {
        "name": "helpscout",
        "cname": [
            "helpscoutdocs.com"
        ],
        "response": [
            "No settings were found for this company:"
        ]
    },
    {
        "name": "cargo",
        "cname": [
            "cargocollective.com"
        ],
        "response": [
            "If you're moving your domain away from Cargo you must make this configuration through your registrar's DNS control panel."
        ]
    },
    {
        "name": "feedpress",
        "cname": [
            "redirect.feedpress.me"
        ],
        "response": [
            "The feed has not been found."
        ]
    },
    {
        "name": "surge",
        "cname": [
            "surge.sh",
            "na-west1.surge.sh"
        ],
        "response": [
            "project not found"
        ]
    },
    {
        "name": "surveygizmo",
        "cname": [
            "privatedomain.sgizmo.com",
            "privatedomain.surveygizmo.eu",
            "privatedomain.sgizmoca.com"
        ],
        "response": [
            "data-html-name"
        ]
    },
    {
        "name": "mashery",
        "cname": [
            "mashery.com"
        ],
        "response": [
            "Unrecognized domain <strong>"
        ]
    },
    {
        "name": "intercom",
        "cname": [
            "custom.intercom.help"
        ],
        "response": [
            "This page is reserved for artistic dogs.",
            "<h1 class=\"headline\">Uh oh. That page doesn’t exist.</h1>"
        ]
    },
    {
        "name":"HatenaBlog",
        "cname":["hatenablog.com"],
        "response":["404 Blog is not found"]
    },
    {
        "name":"LaunchRock",
        "cname":["launchrock.com"],
        "response":["HTTP_STATUS=500"]
    },
    {
        "name": "Helprace",
        "cname": ["helprace.com"],
        "response": ["HTTP_STATUS=301"]
    },
    {
        "name": "Ngrok",
        "cname": ["ngrok.io"],
        "response": ["Tunnel .*.ngrok.io not found"]
    },
    {
        "name": "SmartJobBoard",
        "cname": ["52.16.160.97"],
        "response": ["This job board website is either expired or its domain name is invalid."]
    },
    {
        "name": "Strikingly",
        "cname": ["s.strikinglydns.com"],
        "response": ["PAGE NOT FOUND."]
    },
    {
        "name": "SurveySparrow",
        "cname": ["surveysparrow.com"],
        "response": ["Account not found."]
    },
    {
        "name": "Uberflip",
        "cname": ["read.uberflip.com"],
        "response": ["The URL you've accessed does not provide a hub."]
    },
    {
        "name": "Uptimerobot",
        "cname": ["stats.uptimerobot.com"],
        "response": ["page not found"]
    },
    {
        "name": "Vercel",
        "cname": [".vercel.com"],
        "response": ["DEPLOYMENT_NOT_FOUND."]
    },
    {
        "name": "Worksites",
        "cname": ["worksites.net","69.164.223.206"],
        "response": ["Hello! Sorry, but the website you&rsquo;re looking for doesn&rsquo;t exist."]
    },
    {
        "name": "JetBrains",
        "cname": ["youtrack.cloud"],
        "response": ["is not a registered InCloud YouTrack"]
    },
    {
        "name": "webflow",
        "cname": [
            "proxy.webflow.io"
        ],
        "response": [
            "<p class=\"description\">The page you are looking for doesn't exist or has been moved.</p>"
        ]
    },
    {
        "name": "kajabi",
        "cname": [
            "endpoint.mykajabi.com"
        ],
        "response": [
            "<h1>The page you were looking for doesn't exist.</h1>"
        ]
    },
    {
        "name": "thinkific",
        "cname": [
            "thinkific.com"
        ],
        "response": [
            "You may have mistyped the address or the page may have moved."
        ]
    },
    {
        "name": "tave",
        "cname": [
            "clientaccess.tave.com"
        ],
        "response": [
            "<h1>Error 404: Page Not Found</h1>"
        ]
    },
    {
        "name": "wishpond",
        "cname": [
            "wishpond.com"
        ],
        "response": [
            "https://www.wishpond.com/404?campaign=true"
        ]
    },
    {
        "name": "aftership",
        "cname": [
            "aftership.com"
        ],
        "response": [
            "Oops.</h2><p class=\"text-muted text-tight\">The page you're looking for doesn't exist."
        ]
    },
    {
        "name": "aha",
        "cname": [
            "ideas.aha.io"
        ],
        "response": [
            "There is no portal here ... sending you back to Aha!"
        ]
    },
    {
        "name": "brightcove",
        "cname": [
            "brightcovegallery.com",
            "gallery.video",
            "bcvp0rtal.com"
        ],
        "response": [
            "<p class=\"bc-gallery-error-code\">Error Code: 404</p>"
        ]
    },
    {
        "name": "bigcartel",
        "cname": [
            "bigcartel.com"
        ],
        "response": [
            "<h1>Oops! We couldn&#8217;t find that page.</h1>"
        ]
    },
    {
        "name": "activecompaign",
        "cname": [
            "activehosted.com"
        ],
        "response": [
            "alt=\"LIGHTTPD - fly light.\""
        ]
    },
    {
        "name": "compaignmonitor",
        "cname": [
            "createsend.com"
        ],
        "response": [
            "Double check the URL or <a href=\"mailto:help@createsend.com"
        ]
    },
    {
        "name": "simplebooklet",
        "cname": [
            "simplebooklet.com"
        ],
        "response": [
            "We can't find this <a href=\"https://simplebooklet.com"
        ]
    },
    {
        "name": "getresponse",
        "cname": [
            ".gr8.com"
        ],
        "response": [
            "With GetResponse Landing Pages, lead generation has never been easier"
        ]
    },
    {
        "name": "vend",
        "cname": [
            "vendecommerce.com"
        ],
        "response": [
            "Looks like you've traveled too far into cyberspace."
        ]
    },
    {
        "name": "jetbrains",
        "cname": [
            "myjetbrains.com"
        ],
        "response": [
            "is not a registered InCloud YouTrack.",
            "is not a registered InCloud YouTrack."
        ]
    },
    {
        "name": "azure",
        "cname": [
            "azurewebsites.net",
            "cloudapp.net",
            "cloudapp.azure.com",
            "trafficmanager.net",
            "blob.core.windows.net",
            "azure-api.net",
            "azurehdinsight.net",
            "azureedge.net",
            "azurecontainer.io",
            "database.windows.net",
            "azuredatalakestore.net",
            "search.windows.net",
            "azurecr.io",
            "redis.cache.windows.net",
            "servicebus.windows.net",
            "visualstudio.com"
        ],
        "response": [
            "404 Web Site not found"
        ]
    },
    {
        "name": "readme",
        "cname": [
            "readme.io"
        ],
        "response": [
            "Project doesnt exist... yet!"
        ]
    }
]`)
)
