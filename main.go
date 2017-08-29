package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"html/template"
	//"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"wfp/alipaydemo/ini"
)

var conf = ini.NewConf("config.ini")
var (
	alipay_Gateway_Web = conf.String("alipay_config", "Alipay_Gateway_Web", "")
	app_ID             = conf.String("alipay_config", "APP_ID", "")
	charset            = conf.String("alipay_config", "Charset", "")
	private_key        []byte
	productCode        = conf.String("alipay_config", "Product_Code", "")
	sign_Type          = conf.String("alipay_config", "Sign_Type", "")
	version            = conf.String("alipay_config", "Version", "")
	web_Pay_Method     = conf.String("alipay_config", "Web_Pay_Method", "")
	timestamp          string
	sign               string
	biz_content        string
)
var tpl *template.Template

func init() {
	conf.Parse()
	tpl = template.Must(template.ParseGlob("html/*"))
}

type BizContent struct {
	Out_trade_no string `json:"out_trade_no"`
	Product_code string `json:"product_code"`
	Subject      string `json:"subject"`
	Total_amount string `json:"total_amount"`
}

func AlipayHandler_Web(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tpl.ExecuteTemplate(w, "paydemo.html", nil)
	}
	if r.Method == http.MethodPost {
		var bizContent BizContent
		sSubject := r.FormValue("subject")
		sFee := r.FormValue("fee")
		bizContent.Subject = sSubject
		bizContent.Total_amount = sFee
		bizContent.Product_code = *productCode
		bizContent.Out_trade_no = "201326810518"
		result, _ := json.Marshal(bizContent)
		biz_content = string(result)
		timestamp = time.Now().Format("2006-01-02 15:04:05")
		var p = url.Values{}
		p.Add("app_id", *app_ID)
		p.Add("biz_content", biz_content)
		p.Add("method", *web_Pay_Method)
		p.Add("charset", *charset)
		p.Add("sign_type", *sign_Type)
		p.Add("timestamp", timestamp)
		p.Add("version", *version)
		var keys = make([]string, 0, 0)
		for key, _ := range p {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var pList = make([]string, 0, 0)
		for _, key := range keys {
			var value = strings.TrimSpace(p.Get(key))
			if len(value) > 0 {
				pList = append(pList, key+"="+value)
			}
		}
		var src = strings.Join(pList, "&")
		Sign([]byte(src))
		p.Add("sign", sign)
		biz_content = strings.Replace(biz_content, "\"", "&quot;", -1)

		w.Write([]byte(`
			<html>
			<head>
			<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
			</head>
			<body>
			<form id="alipaysubmit" name="alipaysubmit" action="` + *alipay_Gateway_Web + "?" + "charset=" + *charset + `" method="get" style='display:none;'>
				<input type="hidden" name="app_id" value="` + *app_ID + `">
				<input type="hidden" name="biz_content" value="` + biz_content + `">
				<input type="hidden" name="method" value="` + *web_Pay_Method + `">
				<input type="hidden" name="charset" value="` + *charset + `">
				<input type="hidden" name="sign_type" value="` + *sign_Type + `">
				<input type="hidden" name="timestamp" value="` + timestamp + `">
				<input type="hidden" name="version" value="` + *version + `">
				<input type="hidden" name="sign" value="` + sign + `">
			</form>
			<script>document.forms['alipaysubmit'].submit();</script>
			</body>
			</html>`))

	}
}

func Sign(data []byte) {
	private_key, _ = ioutil.ReadFile("rsa_private_key.pem")
	block, _ := pem.Decode(private_key)
	priv, _ := x509.ParsePKCS1PrivateKey(block.Bytes)
	h := sha256.New()
	h.Write(data)
	hashed := h.Sum(nil)
	miwen, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hashed)
	sign = base64.StdEncoding.EncodeToString(miwen)

}

func main() {

	http.HandleFunc("/alipay_web", AlipayHandler_Web)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}

}
