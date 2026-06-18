package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const CONSULTAR_FEL_JS = `
document
	.evaluate("//span[text()='Emitir Documento Tributario Electrónico (DTE)']/..", document, null, XPathResult.FIRST_ORDERED_NODE_TYPE)
	.singleNodeValue
	.click()
`

type DatosInicialesResponse struct {
	Response struct {
		ValidationToken string `json:"validacion"`
		Token           string `json:"tokenJWT"`
	} `json:"respuesta"`
}

type SatSSOFEL struct{}

// GetSession opens a visible browser window, navigates to the SAT login page,
// waits up to one minute for the user to log in, then extracts session cookies
// and the ACCESS_TOKEN needed for API requests.
func (s *SatSSOFEL) FetchSession(forceWindow bool) (*SatSession, error) {
	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Navigate("https://farm3.sat.gob.gt/menu/login.jsf")); err != nil {
		return nil, err
	}

	if err := chromedp.Run(ctx, chromedp.WaitVisible("#btnContraerMenu", chromedp.ByQuery)); err != nil {
		return nil, err
	}

	log.Println("Waiting for page to load before attempting to click the 'Emitir Documento Tributario Electronico (DTE)' menu item...")
	time.Sleep(15 * time.Second)
	if err := chromedp.Run(ctx, chromedp.Evaluate(CONSULTAR_FEL_JS, nil)); err != nil {
		return nil, fmt.Errorf("failed to click menu element: %w", err)
	}

	log.Println("Waiting for page to load before attempting to retrieve keys...")
	time.Sleep(15 * time.Second)

	var keyAttrValue string
	var keyAttrOk bool
	if err := chromedp.Run(ctx, chromedp.AttributeValue("#frmMenu\\:idhUrlSelected", "value", &keyAttrValue, &keyAttrOk)); err != nil {
		return nil, fmt.Errorf("Failed to retrieve input keys %w", err)
	}

	url, err := url.Parse(keyAttrValue)
	if err != nil {
		return nil, fmt.Errorf("Error parsing the url %w", err)
	}

	key := url.Query().Get("Clave")
	key = strings.TrimSuffix(key, "Nit=")

	type diRequestBody struct {
		User      string `json:"usuario"`
		Key       string `json:"clave"`
		APIVA     bool
		APISR     bool
		LastMonth bool
	}

	datosInicialesRequestBody := diRequestBody{
		User:      "",
		Key:       key,
		LastMonth: false,
		APIVA:     false,
		APISR:     false,
	}

	jsonbyts, err := json.Marshal(datosInicialesRequestBody)
	if err != nil {
		return nil, fmt.Errorf("Error encoding DI request body %w", err)
	}

	diResponse, err := http.Post("https://felav02.c.sat.gob.gt/fel-rest/rest/privado/DatosIniciales", "application/json", bytes.NewReader(jsonbyts))
	if err != nil {
		return nil, fmt.Errorf("Error requesting DI data %w", err)
	}

	diResponseBytes, err := io.ReadAll(diResponse.Body)
	if err != nil {
		return nil, err
	}

	diBody := DatosInicialesResponse{}
	err = json.Unmarshal(diResponseBytes, &diBody)
	if err != nil {
		return nil, err
	}

	jwts := []JWToken{}
	jwts = append(jwts, JWToken{
		Name:  "feltoken",
		Value: diBody.Response.Token,
	})

	jwts = append(jwts, JWToken{
		Name:  "felvaltoken",
		Value: diBody.Response.ValidationToken,
	})

	return &SatSession{
		AccessToken: diBody.Response.Token,
		jwts:        jwts,
	}, nil
}
