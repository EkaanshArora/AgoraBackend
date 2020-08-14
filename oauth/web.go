package oauth

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
)

// WebOAuthHandler is a REST route that is called when the oauth provider redirects to here and provides the code
func (o *Router) WebOAuthHandler(w http.ResponseWriter, r *http.Request) {
	redirect, token, err := Handler(w, r, o.DB, "web")
	if err != nil {
		log.Panic(err)
		fmt.Fprint(w, err)
		return
	}

	newURL, err := url.Parse(*redirect)
	if err != nil {
		log.Panic(err)
		fmt.Fprint(w, err)
		return
	}

	newURL.Path = path.Join(newURL.Path, *token)

	http.Redirect(w, r, newURL.String(), http.StatusSeeOther)
}
