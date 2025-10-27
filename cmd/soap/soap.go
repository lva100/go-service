package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/lva100/go-service/config"
)

type SoapResponse struct {
	XMLName xml.Name
	Body    SoapBody
}

type SoapBody struct {
	XMLName               xml.Name
	GetPersonDataResponse SoapPersonDataResponse `xml:"getPersonDataResponse"`
}

type SoapPersonDataResponse struct {
	XMLName           xml.Name `xml:"getPersonDataResponse"`
	ExternalRequestId string   `xml:"externalRequestId"`
	Pd                SoapPD   `xml:"pd"`
}

type SoapPD struct {
	XMLName xml.Name `xml:"pd"`
	Oip     string   `xml:"oip"`
}

func main() {
	var result SoapResponse

	config.Init()

	url := "http://10.255.87.30/api/t-foms/integration/ws/24.1.2/wsdl/mpiPersonInfoServiceWs"
	client := &http.Client{}
	sRequestContent := generateRequestContent(config.GetTestEnp())
	requestContent := []byte(sRequestContent)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestContent))
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Content-Type", "text/xml; charset=utf-8")
	req.Header.Add("Accept", "text/xml")
	req.Header.Add("X-Auth-Token", config.GetToken())
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalln("Error Respose " + resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	err = xml.Unmarshal(body, &result)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(result)

}

func generateRequestContent(enp string) string {
	type QueryData struct {
		Enp string
	}

	const templ = `<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:mpip="http://ffoms.ru/types/24.1.2/mpiPersonInfoSchema" xmlns:com="http://ffoms.ru/types/24.1.2/commonTypes">
		   <soapenv:Header/>
		   <soapenv:Body>
		      <mpip:getPersonDataRequest>
		         <com:externalRequestId>ECE1D14C-AEA8-4D2A-87F6-BE7A5EBCAA66</com:externalRequestId>
		         <mpip:personDataSearchParams>
		            <mpip:personSearchInfo>
		               <mpip:pcy>
		                  <com:enp>{{.Enp}}</com:enp>
		               </mpip:pcy>
		            </mpip:personSearchInfo>
		         </mpip:personDataSearchParams>
		      </mpip:getPersonDataRequest>
		   </soapenv:Body>
		</soapenv:Envelope>`

	querydata := QueryData{Enp: enp}
	tmpl, err := template.New("getFerzlPersonData").Parse(templ)
	if err != nil {
		panic(err)
	}
	var doc bytes.Buffer
	err = tmpl.Execute(&doc, querydata)
	if err != nil {
		panic(err)
	}
	return doc.String()
}
