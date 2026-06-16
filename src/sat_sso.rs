use std::thread;
use std::time;
use headless_chrome::protocol::cdp::Network::Cookie;

const SAT_DTE_SELECTOR: &str = "#frmMenu\\:j_idt41 > ul > li:nth-child(5) > ul > li:nth-child(5) > ul > li:nth-child(6) > a";

pub struct SatSession {
    pub access_token: String,
    pub cookies: Vec<Cookie> 
}

impl SatSession {
    pub fn token_header(&self) -> String {
        self.cookies.iter().map(|cookie| format!("{}={}", cookie.name, cookie.value)).collect::<Vec<String>>().join(";")
    }
}

pub fn get_session() -> Result<SatSession, String> {
    let launch_options = headless_chrome::LaunchOptions {
        headless: false,
        ..Default::default()
    };

    let browser = match headless_chrome::Browser::new(launch_options) {
        Ok(browser) => browser,
        Err(e) => {
            println!("{}", e);
            return Err("Launchable browser not found".into());
        }
    };
    
    let tab = match browser.new_tab() {
        Ok(tab) => tab,
        Err(e) => {
            println!("{}", e);
            return Err("Not possible to open a new tab".into());
        }
    };
    
    tab.navigate_to("https://farm3.sat.gob.gt/menu/login.jsf").expect("A");
    tab.wait_until_navigated().expect("B");
    println!("looking for {}", SAT_DTE_SELECTOR);

    let menu_button = tab.wait_for_element_with_custom_timeout(SAT_DTE_SELECTOR, time::Duration::from_mins(1)).expect("C"); 

    thread::sleep(time::Duration::from_secs(5)); 
    menu_button.call_js_fn(
        r"function(args) {
                console.log(args);
                let element = document.querySelector(args);
                element.click();
        }", 
        vec![SAT_DTE_SELECTOR.into()],
        false
    ).expect("Clicking the menu element");

    thread::sleep(time::Duration::from_secs(1));
    let cookies = match tab.get_cookies() {
        Ok(val) => val,
        Err(_) => return Err("Cookies not found".into())
    };

    let cookies: Vec<Cookie> = cookies
        .into_iter() 
        .filter(|cookie| {
            if cookie.domain == String::from("felcons.c.sat.gob.gt") || cookie.domain == String::from(".sat.gob.gt") {
                return true;
            }
            false
        })
        .collect();

    let access_token = match cookies.iter().find(|cookie| cookie.name == "ACCESS_TOKEN") {
        Some(token) => token.value.clone(),
        None => return Err("Session token not found".into())
    };
    
    Ok(SatSession {
       access_token,
       cookies
    })
}

