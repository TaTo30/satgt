package session

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const CONSULTAR_DTE_JS = `
document
	.evaluate("//span[text()='Consultar DTE']/..", document, null, XPathResult.FIRST_ORDERED_NODE_TYPE)
	.singleNodeValue
	.click()
`

const LOGIN_FORM_FILLING_JS = `
let usernameInput = document.getElementById("formContent:username");
let passwordInput = document.getElementById("formContent:password");
let submitButton = document.getElementById("formContent:cmdbtnIngresar");

usernameInput.value = "%s";
passwordInput.value = "%s"; 

submitButton.click();
`

type SatSSODTE struct{}

// GetSession opens a visible browser window, navigates to the SAT login page,
// waits up to one minute for the user to log in, then extracts session cookies
// and the ACCESS_TOKEN needed for API requests.
func (s *SatSSODTE) FetchSession(forceWindow bool) (*SatSession, error) {
	setHeadless := true
	userNit := os.Getenv("SATGT_USERNIT")
	userPass := os.Getenv("SATGT_USERPASSWORD")

	if forceWindow {
		setHeadless = false
	} else {
		if userNit == "" || userPass == "" {
			setHeadless = false
			log.Println("SATGT_USERNIT or SATGT_USERPASS environment variables are not set. Opening a visible browser window for manual login.")
		}
	}

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", setHeadless),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if err := chromedp.Run(ctx, chromedp.Navigate("https://farm3.sat.gob.gt/menu/login.jsf")); err != nil {
		return nil, err
	}

	if setHeadless {
		log.Println("Attempting to fill in the login form automatically...")

		time.Sleep(1 * time.Second)
		loginFormFillingScript := fmt.Sprintf(LOGIN_FORM_FILLING_JS, userNit, userPass)
		if err := chromedp.Run(ctx, chromedp.Evaluate(loginFormFillingScript, nil)); err != nil {
			return nil, fmt.Errorf("failed to fill login form: %w", err)
		}
	}

	if err := chromedp.Run(ctx, chromedp.WaitVisible("#btnContraerMenu", chromedp.ByQuery)); err != nil {
		return nil, err
	}

	log.Println("Waiting for page to load before attempting to click the 'Consultar DTE' menu item...")
	time.Sleep(15 * time.Second)
	if err := chromedp.Run(ctx, chromedp.Evaluate(CONSULTAR_DTE_JS, nil)); err != nil {
		return nil, fmt.Errorf("failed to click menu element: %w", err)
	}

	log.Println("Waiting for page to load before attempting to retrieve cookies...")
	time.Sleep(10 * time.Second)
	var rawCookies []*network.Cookie
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		rawCookies, err = network.GetCookies().Do(ctx)
		return err
	})); err != nil {
		return nil, fmt.Errorf("failed to retrieve cookies: %w", err)
	}

	var cookies []Cookie
	for _, c := range rawCookies {
		if c.Domain == "felcons.c.sat.gob.gt" || c.Domain == ".sat.gob.gt" {
			cookies = append(cookies, Cookie{
				Name:   c.Name,
				Value:  c.Value,
				Domain: c.Domain,
			})
		}
	}

	accessToken := ""
	for _, c := range cookies {
		if c.Name == "ACCESS_TOKEN" {
			accessToken = c.Value
			break
		}
	}

	if accessToken == "" {
		return nil, fmt.Errorf("session token not found in cookies")
	}

	log.Printf("Cookies retrieved, access_token found: %s\n", accessToken)

	if err := chromedp.Cancel(ctx); err != nil {
		return nil, fmt.Errorf("failed to cancel chromedp context: %w", err)
	}

	return &SatSession{
		AccessToken: accessToken,
		cookies:     cookies,
	}, nil
}
