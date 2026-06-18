package dte

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"satgt/session"
	"time"
)

const DTE_BASE_URL = "https://felcons.c.sat.gob.gt/dte-agencia-virtual/api/consulta-dte"

// FelDocument represents a single FEL (Factura Electrónica en Línea) document
// returned by the SAT DTE API.
type FelDocument struct {
	IssueDate               string  `json:"fechaEmision"`
	UUID                    string  `json:"numeroUuid"`
	DocumentType            string  `json:"tipo"`
	Series                  string  `json:"serie"`
	DocumentNumber          string  `json:"numeroDocumento"`
	IssuerTaxID             string  `json:"nitEmisor"`
	IssuerName              string  `json:"nombreEmisor"`
	EstablishmentCode       string  `json:"codigoEstablecimiento"`
	EstablishmentName       string  `json:"nombreEstablecimiento"`
	RecipientSpecialType    *string `json:"receptorTipoEspecial"`
	RecipientTaxID          string  `json:"nitReceptor"`
	RecipientName           string  `json:"nombreReceptor"`
	CertifierTaxID          string  `json:"nitCertificador"`
	CertifierName           string  `json:"nombreCertificador"`
	Currency                string  `json:"moneda"`
	TotalAmount             string  `json:"granTotal"`
	VoidedStatus            string  `json:"anulado"`
	RejectedStatus          *string `json:"rechazado"`
	VoidedDate              *string `json:"fechaAnulacion"`
	VATAmount               string  `json:"totalIva"`
	PetroleumTax            string  `json:"impuestoPetroleo"`
	LodgingTourismTax       string  `json:"impuestoTurismoHospedaje"`
	TravelTourismTax        string  `json:"impuestoTurismoPasaje"`
	PressStampTax           string  `json:"impuestoTrimbrePrensa"`
	FirefightersTax         string  `json:"impuestoBomberos"`
	MunicipalTax            string  `json:"impuestoTasaMunicipal"`
	AlcoholicBeverageTax    string  `json:"impuestoBebidaAlcoholica"`
	TobaccoTax              string  `json:"impuestoTabaco"`
	CementTax               string  `json:"impuestoCemento"`
	NonAlcoholicBeverageTax string  `json:"impuestoBebidaNoAlcoholica"`
	PortFeeTax              string  `json:"impuestoTarifaPortuaria"`
	IssuerClassification    *string `json:"clasificacionEmisor"`
	IsExport                string  `json:"exportacion"`
	TemporaryIssueLocation  *string `json:"emisionUbicacionTemporal"`
}

// SatDTEParameters holds the query parameters for the SAT DTE API request.
type SatDTEParameters struct {
	OperationType  string
	Vendor         string
	DTEType        string
	AuthNumber     string
	TaxedVAT       string
	DTEStatus      string
	Series         string
	Number         string
	Currency       string
	RangeInitial   string
	RangeFinal     string
	Tax            string
	CertificatorID string
	Result         string
	DateFrom       time.Time
	DateTo         time.Time
}

// SatResponse is the generic envelope returned by all SAT API endpoints.
type SatResponse[T any] struct {
	Code        string  `json:"codigo"`
	Description *string `json:"descripcion"`
	Detail      T       `json:"detalle"`
}

// SatDTEResponse holds the paginated list of FEL documents.
type SatDTEResponse struct {
	Total     uint32        `json:"total"`
	PageTotal uint32        `json:"totalPagina"`
	Data      []FelDocument `json:"data"`
}

// toQueryString encodes the parameters as a URL query string.
// Dates are formatted as DD-MM-YYYY to match the SAT API expectation.
func (p SatDTEParameters) toQueryString() string {
	q := url.Values{}
	q.Set("tipoOperacion", p.OperationType)
	q.Set("establecimiento", p.Vendor)
	q.Set("tipoDte", p.DTEType)
	q.Set("noAutorizacion", p.AuthNumber)
	q.Set("nitIdReceptor", p.TaxedVAT)
	q.Set("estadoDte", p.DTEStatus)
	q.Set("serie", p.Series)
	q.Set("numero", p.Number)
	q.Set("moneda", p.Currency)
	q.Set("montoTotalRangoIni", p.RangeInitial)
	q.Set("montoTotalRangoFinal", p.RangeFinal)
	q.Set("impuesto", p.Tax)
	q.Set("nitCertificador", p.CertificatorID)
	q.Set("resultado", p.Result)
	q.Set("fechaEmisionIni", p.DateFrom.Format("02-01-2006"))
	q.Set("fechaEmisionFinal", p.DateTo.Format("02-01-2006"))
	return q.Encode()
}

// returns basic parameters configured to query DTE documents
func FetchDTEParameters(month int, year int, opType string) SatDTEParameters {
	fromDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	toDate := fromDate.AddDate(0, 1, -1)

	return SatDTEParameters{
		OperationType: opType,
		DateFrom:      fromDate,
		DateTo:        toDate,
	}
}

func FetchDTEData(params SatDTEParameters, session session.SatSession) (*SatResponse[SatDTEResponse], error) {
	client := http.DefaultClient
	reqURL := fmt.Sprintf("%s?%s", DTE_BASE_URL, params.toQueryString())

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", session.AccessToken)
	req.Header.Set("Cookie", session.Cookies())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SatResponse[SatDTEResponse]
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func FetchDTEXMLFiles(documents []FelDocument, params SatDTEParameters, session session.SatSession) ([]byte, error) {
	client := http.DefaultClient
	reqURL := fmt.Sprintf("%s/zip-xml?%s", DTE_BASE_URL, params.toQueryString())

	body, err := json.Marshal(documents)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", session.AccessToken)
	req.Header.Set("Cookie", session.Cookies())
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, err
	}

	file, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return file, nil
}
