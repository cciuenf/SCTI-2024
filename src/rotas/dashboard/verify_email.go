package dashboard

import (
  DB "SCTI/database"
  "net/http"
  "net/url"
  "fmt"
  "os"

  gomail "gopkg.in/mail.v2"
)

func SentError(w http.ResponseWriter, err error) {
  w.Header().Set("Content-Type", "text/html")
  w.Write([]byte(`
      <div class="failure">
        Falha ao enviar o email de verificação: ` + err.Error() + `
    </div>
  `))
}

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
  cookie, err := r.Cookie("accessToken")
  if err != nil {
    // fmt.Println("Error Getting cookie:", err)
    http.Redirect(w, r, "/login", http.StatusSeeOther)
    return
  }

  if cookie.Value == "-1" {
    // fmt.Println("Invalid accessToken")
    http.Redirect(w, r, "/login", http.StatusSeeOther)
  }

  email := DB.GetEmail(cookie.Value)
  code, err := DB.GetCode(cookie.Value)
  if err != nil {
    SentError(w, err)
    return
  }

  from := os.Getenv("GMAIL_SENDER")
  pass := os.Getenv("GMAIL_PASS")

  encodedEmail := url.QueryEscape(email)
  verificationLink := fmt.Sprintf("http://localhost:8080/verify?code=%s&email=%s", code, encodedEmail)
  notMeLink := fmt.Sprintf("http://localhost:8080/delete?code=%s&email=%s", code, encodedEmail)

  htmlBody := `
    <!DOCTYPE html>
    <html>
    <head>
        <style>
            .button {
                display: inline-block;
                padding: 10px 20px;
                font-size: 16px;
                cursor: pointer;
                text-align: center;
                text-decoration: none;
                outline: none;
                color: #ffffff;
                background-color: #4CAF50;
                border: none;
                border-radius: 15px;
                box-shadow: 0 9px #999;
            }
            .button:hover {background-color: #3e8e41}
            .button:active {
                background-color: #3e8e41;
                box-shadow: 0 5px #666;
                transform: translateY(4px);
            }
        </style>
    </head>
    <body>
        <p>Clique no botão abaixo para verificar seu email:</p>
        <a href="` + verificationLink + `" class="button">Verificar Email</a>
        <a href="` + notMeLink + `" class="button">Não fui eu</a>
    </body>
    </html>
  `

  plainBody := "Clique aqui para verificar seu email:\n" + verificationLink


  msg := gomail.NewMessage()
  msg.SetHeader("From", from)
  msg.SetHeader("To", email)
  msg.SetHeader("Subject", "Verificação de email SCTI")
  msg.SetBody("text/plain", plainBody)
  msg.AddAlternative("text/html", htmlBody)

  dialer := gomail.NewDialer("smtp.gmail.com", 587, from, pass)

  if err := dialer.DialAndSend(msg); err != nil {
    SentError(w, err)
    return
  }

  w.Header().Set("Content-Type", "text/html")
  w.Write([]byte(`
      <div class="success">
          Email de verificação enviado com sucesso!!
      </div>
  `))
}
