use reqwest;
mod sat_sso;
mod sat_dte;

fn main() {
   // let sat_session = match sat_sso::get_session() {
   //     Ok(session) => session,
   //     Err(e) => panic!("{}", e)
   // };
   // 
   // println!("{:?}", sat_session.cookies);
   // println!("{}", sat_session.access_token);
   // println!("{}", sat_session.token_header());

    let dte_parameters = sat_dte::SatDTE::received_dte(5, 2025);
    println!("{:?}", dte_parameters);    
    let dte_query = serde_qs::to_string(&dte_parameters).unwrap();
    println!("{dte_query}");
}

