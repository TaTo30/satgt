use chrono::{Datelike, NaiveDate};
use serde::{Serialize};

#[derive(Serialize, Default, Debug)]
pub struct SatDTE {
    #[serde(rename = "tipoOperacion")]
    operation_type: String, 

    #[serde(rename = "establecimiento")]
    vendor: String,
    
    #[serde(rename = "tipoDte")]
    dte_type: String,

    #[serde(rename = "noAutorizacion")]
    auth_number: String,

    #[serde(rename = "nitIdReceptor")]
    taxed_vat: String,

    #[serde(rename = "estadoDte")]
    dte_status: String,

    #[serde(rename = "serie")]
    serie: String,

    #[serde(rename = "numero")]
    number: String,

    #[serde(rename = "moneda")]
    currency: String,

    #[serde(rename = "montoTotalRangoIni")]
    range_initial: String,
    
    #[serde(rename = "montoTotalRangoFinal")]
    range_final: String, 

    #[serde(rename = "impuesto")]
    tax: String,

    #[serde(rename = "nitCertificador")]
    certificator_id: String,
    
    #[serde(rename = "resultado")]
    result: String,

    #[serde(serialize_with = "serialize_date")]
    #[serde(rename = "fechaEmisionIni")]
    date_from: NaiveDate,

    #[serde(rename = "fechaEmisionFinal")]
    #[serde(serialize_with = "serialize_date")]
    date_to: NaiveDate,
}

impl SatDTE {
    pub fn emitted_dte(month: u32, year: i32) -> SatDTE {
        let from_date = chrono::NaiveDate::from_ymd_opt(year, month, 1).unwrap(); 
        let month_num_days = from_date.num_days_in_month(); 
        let to_date = chrono::NaiveDate::from_ymd_opt(year, month, month_num_days as u32).unwrap();
        SatDTE {
            date_to: to_date, 
            date_from: from_date,
            operation_type: String::from("E"),
            ..Default::default()
        }
    }

    pub fn received_dte(month: u32, year: i32) -> SatDTE {
        let from_date = chrono::NaiveDate::from_ymd_opt(year, month, 1).unwrap(); 
        let month_num_days = from_date.num_days_in_month(); 
        let to_date = chrono::NaiveDate::from_ymd_opt(year, month, month_num_days as u32).unwrap();
        SatDTE {
            date_to: to_date, 
            date_from: from_date,
            operation_type: String::from("R"),
            ..Default::default()
        }
    }
}

fn serialize_date<S>(date: &NaiveDate, serializer: S) -> Result<S::Ok, S::Error> 
    where S: serde::Serializer, 
{
    let formatted = date.format("%d-%m-%Y").to_string();
    serializer.serialize_str(&formatted)
}


