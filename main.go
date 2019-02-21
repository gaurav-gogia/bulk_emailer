package main

import (
	"crypto/tls"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/smtp"
	"os"
	"strings"
)

const (
	name     = ""
	id       = ""
	pas      = ""
	toorg    = ""
	csvpath  = "./responses.csv"
	filename = "./email.html"
)

type mail struct {
	name    string
	from    string
	to      string
	bcc     []string
	subject string
	body    string
	pass    string
}

type smtpServer struct {
	host      string
	port      string
	tlsConfig *tls.Config
}

func (s *smtpServer) serverName() string {
	return s.host + ":" + s.port
}

func (m *mail) initmail() error {
	var err error
	m.name = name
	m.from = id
	m.subject = "Feedback Please 😃"
	m.body, err = readFile(filename)
	if err != nil {
		return err
	}
	m.to = toorg
	m.pass = pas
	m.bcc, err = readCsv(csvpath)
	if err != nil {
		return err
	}
	return nil
}

func (m *mail) buildmsg() string {
	var header string
	mime := "1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	header += fmt.Sprintf("From: %s\r\n", m.name)
	header += fmt.Sprintf("To: %s\r\n", m.to)
	if len(m.bcc) > 0 {
		header += fmt.Sprintf("Bcc: %s\r\n", strings.Join(m.bcc, ";"))
	}
	header += fmt.Sprintf("Subject: %s\r\n", m.subject)
	header += fmt.Sprintf("MIME-version: %s\r\n", mime)
	header += "\r\n" + m.body

	return header
}

func (m *mail) authserver() (*smtp.Client, error) {
	smtpServer := smtpServer{host: "smtp.gmail.com", port: "465"}
	smtpServer.tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         smtpServer.host,
	}

	auth := smtp.PlainAuth("", m.from, m.pass, smtpServer.host)

	conn, err := tls.Dial("tcp", smtpServer.serverName(), smtpServer.tlsConfig)
	if err != nil {
		return nil, err
	}

	client, err := smtp.NewClient(conn, smtpServer.host)
	if err != nil {
		return nil, err
	}

	if err := client.Auth(auth); err != nil {
		return nil, err
	}

	return client, nil
}

func (m *mail) send(msgbody string, client *smtp.Client) error {
	if err := client.Mail(m.from); err != nil {
		return err
	}

	if err := client.Rcpt(m.to); err != nil {
		return err
	}

	for i, k := range m.bcc {
		fmt.Println("Sending to: ", k)
		if err := client.Rcpt(k); err != nil {
			m.bcc = append(m.bcc[:i], m.bcc[i+1:]...)
			fmt.Println(k+" is not reachable.\nError: ", err.Error())
		}
	}

	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(msgbody))
	if err != nil {
		return err
	}

	if err := w.Close(); err != nil {
		return err
	}

	client.Quit()
	return nil
}

func readFile(fname string) (string, error) {
	f, err := ioutil.ReadFile(fname)
	if err != nil {
		return "", err
	}
	return string(f), nil
}

func readCsv(fname string) ([]string, error) {
	var emails []string
	var flag bool
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	csv := csv.NewReader(file)
	for {
		col, err := csv.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		} else {
			if !flag {
				flag = true
				continue
			} else {
				emails = append(emails, col[1])
			}
		}
	}
	return emails, nil
}

func handle(e error) {
	if e != nil {
		panic(e)
	}
}

func run() {
	var m mail
	err := m.initmail()
	handle(err)

	msgbbody := m.buildmsg()

	fmt.Println("Authenticating....")

	client, err := m.authserver()
	handle(err)

	fmt.Println("Sending....")

	err = m.send(msgbbody, client)
	handle(err)

}

func main() {
	fmt.Println("Revving engines....")

	run()

	fmt.Println("Done!")
}
