package api

import (
	"net/http"
)

func (as *Server) Sso1(w http.ResponseWriter, r *http.Request) {
	// exec.Command("rundll32", "url.dll,FileProtocolHandler", "https://127.0.0.1:3333/").Start()
	// username := strings.ToLower(r.URL.Query()["name"][0])
	// u, err := models.GetUserByUsername(username)
	// if err != nil {
	// 	log.Error(err)
	// 	return
	// }

	// if u.AccountLocked {
	// 	return
	// }
	// u.LastLogin = time.Now().UTC()
	// err = models.PutUser(&u)
	// if err != nil {
	// 	log.Error(err)
	// }
	// // If we've logged in, save the session and redirect to the dashboard
	// session := ctx.Get(r, "session").(*sessions.Session)
	// session.Values["id"] = u.Id
	// session.Save(r, w)
	// cookie, err := r.Cookie("gophish")
	// fmt.Println("ccccccccccccc")
	// fmt.Println(cookie)
	// cookie = &http.Cookie{
	// 	Name:     session.Name(),
	// 	Value:    session.Values,
	// 	HttpOnly: true,
	// }
	// http.SetCookie(w, cookie)

	// http.Redirect(w, r, "https://google.com", http.StatusFound)

}
