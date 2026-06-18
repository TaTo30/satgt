package dte

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
)

type DTEXML struct {
	XMLName xml.Name `xml:"GTDocumento"`
	Vendor  struct {
		VendorVAT      string `xml:"NITEmisor,attr"`
		VendorCode     string `xml:"CodigoEstablecimiento,attr"`
		Name           string `xml:"NombreComercial,attr"`
		LegalName      string `xml:"NombreEmisor,attr"`
		Email          string `xml:"CorreoEmisor,attr"`
		AfiliationType string `xml:"AfiliacionIVA,attr"`
		Address        string `xml:"DireccionEmisor>Direccion"`
		PostalCode     string `xml:"DireccionEmisor>CodigoPostal"`
		Locality       string `xml:"DireccionEmisor>Municipio"`
		State          string `xml:"DireccionEmisor>Departamento"`
		Country        string `xml:"DireccionEmisor>Pais"`
	} `xml:"SAT>DTE>DatosEmision>Emisor"`
	Recipient struct {
		RecipientVAT string `xml:"IDReceptor,attr"`
		Email        string `xml:"CorreoReceptor,attr"`
		Name         string `xml:"NombreReceptor,attr"`
		Address      string `xml:"DireccionReceptor>Direccion"`
		PostalCode   string `xml:"DireccionReceptor>CodigoPostal"`
		Locality     string `xml:"DireccionReceptor>Municipio"`
		State        string `xml:"DireccionReceptor>Departamento"`
		Country      string `xml:"DireccionReceptor>Pais"`
	} `xml:"SAT>DTE>DatosEmision>Receptor"`
	Document struct {
		Currency  string `xml:"CodigoMoneda,attr"`
		Timestamp string `xml:"FechaHoraEmision,attr"`
		Type      string `xml:"Tipo,attr"`
	} `xml:"SAT>DTE>DatosEmision>DatosGenerales"`
	Certification struct {
		Serie  string `xml:"Serie,attr"`
		Number string `xml:"Numero,attr"`
		Uuid   string `xml:",chardata"`
	} `xml:"SAT>DTE>Certificacion>NumeroAutorizacion"`
	Total struct {
		Total string `xml:"GranTotal"`
		Taxes []struct {
			Name  string `xml:"NombreCorto,attr"`
			Total string `xml:"TotalMontoImpuesto,attr"`
		} `xml:"TotalImpuestos>TotalImpuesto"`
	} `xml:"SAT>DTE>DatosEmision>Totales"`
	Items []struct {
		Type        string `xml:"BienOServicio,attr"`
		Uid         string `xml:"NumeroLinea,attr"`
		Quantity    string `xml:"Cantidad"`
		MeasureUnit string `xml:"UnidadMedida"`
		Description string `xml:"Descripcion"`
		UnitPrice   string `xml:"PrecioUnitario"`
		Price       string `xml:"Precio"`
		Discount    string `xml:"Descuento"`
		Total       string `xml:"Total"`
		Taxes       []struct {
			Name          string `xml:"NombreCorto"`
			Code          string `xml:"CodigoUnidadGravable"`
			TaxableAmount string `xml:"MontoGravable"`
			TaxedAmount   string `xml:"MontoImpuesto"`
		} `xml:"Impuestos>Impuesto"`
	} `xml:"SAT>DTE>DatosEmision>Items>Item"`
}

type DTEFile struct {
	Bytes    []byte
	Filename string
}

func ExtractXMLFilesFromZip(zipFile []byte) ([]DTEFile, error) {
	reader := bytes.NewReader(zipFile)
	zipreader, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return nil, err
	}

	files := []DTEFile{}
	for _, file := range zipreader.File {
		fileReader, _ := file.Open()
		writerBuffer := new(bytes.Buffer)
		io.CopyN(writerBuffer, fileReader, int64(file.UncompressedSize64))

		defer fileReader.Close()

		newdtefile := DTEFile{Bytes: writerBuffer.Bytes(), Filename: file.Name}
		files = append(files, newdtefile)
	}

	return files, nil
}

func ParseXMLFile(xmlfile []byte) (*DTEXML, error) {
	parsedFile := DTEXML{}
	if err := xml.Unmarshal(xmlfile, &parsedFile); err != nil {
		return nil, err
	}

	return &parsedFile, nil
}
