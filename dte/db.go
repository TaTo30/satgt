package dte

import (
	"log"
	"strconv"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func getTaxRecord(tax_name string, app core.App) *core.Record {
	record, err := app.FindFirstRecordByFilter("dte_tax", "name = {:name}", dbx.Params{"name": tax_name})

	if err != nil {
		collection, _ := app.FindCollectionByNameOrId("dte_tax")
		new_tax_record := core.NewRecord(collection)
		new_tax_record.Set("name", tax_name)
		new_tax_record.Set("rate", 0.0)

		if err1 := app.Save(new_tax_record); err1 != nil {
			log.Println(err1.Error())
		}

		return new_tax_record
	}

	return record
}

func addVendor(app core.App, dtevalue *DTEXML) (*core.Record, error) {
	// upsert the vendor found in the document. The record of the vendor is the one returned
	records, _ := app.FindRecordsByFilter("dte_vendor", "vat = {:vat}", "", 1, 0, dbx.Params{"vat": dtevalue.Vendor.VendorVAT})
	vendorRecord := &core.Record{}
	siteRecord := &core.Record{}

	if len(records) == 0 {
		vendorcollections, _ := app.FindCollectionByNameOrId("dte_vendor")
		vendorRecord = core.NewRecord(vendorcollections)
	} else {
		vendorRecord = records[0]
	}

	vendorRecord.Set("vat", dtevalue.Vendor.VendorVAT)
	vendorRecord.Set("name", dtevalue.Vendor.LegalName)
	vendorRecord.Set("afiliation_type", dtevalue.Vendor.AfiliationType)

	err := app.Save(vendorRecord)
	if err != nil {
		return nil, err
	}

	// upsert the vendor site found in the document
	records, _ = app.FindRecordsByFilter("dte_vendor_site", "vendor_id = {:vendor_id} && code = {:code}", "", 1, 0, dbx.Params{"vendor_id": vendorRecord.Id, "code": dtevalue.Vendor.VendorCode})
	if len(records) == 0 {
		collection, _ := app.FindCollectionByNameOrId("dte_vendor_site")
		siteRecord = core.NewRecord(collection)
	} else {
		siteRecord = records[0]
	}

	siteRecord.Set("address", dtevalue.Vendor.Address)
	siteRecord.Set("locality", dtevalue.Vendor.Locality)
	siteRecord.Set("state", dtevalue.Vendor.State)
	siteRecord.Set("country", dtevalue.Vendor.Country)
	siteRecord.Set("postal_code", dtevalue.Vendor.PostalCode)
	siteRecord.Set("name", dtevalue.Vendor.Name)
	siteRecord.Set("vendor_id", vendorRecord.Id)
	siteRecord.Set("code", dtevalue.Vendor.VendorCode)

	err = app.Save(siteRecord)
	if err != nil {
		return nil, err
	}

	vendorRecord.Set("sites", append(vendorRecord.Get("sites").([]string), siteRecord.Id))
	app.Save(vendorRecord)

	return siteRecord, nil
}

func addRecipient(app core.App, dtevalue *DTEXML) (*core.Record, error) {
	record, err := app.FindFirstRecordByData("dte_recipient", "vat", dtevalue.Recipient.RecipientVAT)
	if err != nil {
		collection, _ := app.FindCollectionByNameOrId("dte_recipient")
		record = core.NewRecord(collection)
	}

	record.Set("vat", dtevalue.Recipient.RecipientVAT)
	record.Set("name", dtevalue.Recipient.Name)
	record.Set("email", dtevalue.Recipient.Email)
	record.Set("address", dtevalue.Recipient.Address)
	record.Set("country", dtevalue.Recipient.Country)
	record.Set("state", dtevalue.Recipient.State)
	record.Set("locality", dtevalue.Recipient.Locality)
	record.Set("postal_code", dtevalue.Recipient.PostalCode)

	if err := app.Save(record); err != nil {
		return nil, err
	}

	return record, nil
}

func addItems(document_id any, app core.App, dtevalue *DTEXML) ([]string, error) {
	item_ids := []string{}
	for _, item := range dtevalue.Items {
		records, _ := app.FindRecordsByFilter("dte_item", "document_id = {:document_id} && uid = {:uid}", "", 1, 0, dbx.Params{"document_id": document_id, "uid": item.Uid})
		itemRecord := &core.Record{}

		if len(records) == 0 {
			collection, _ := app.FindCollectionByNameOrId("dte_item")
			itemRecord = core.NewRecord(collection)
		} else {
			itemRecord = records[0]
		}

		itemRecord.Set("quantity", item.Quantity)
		itemRecord.Set("measure_unit", strings.ToUpper(item.MeasureUnit))
		itemRecord.Set("price_unit", item.UnitPrice)
		itemRecord.Set("price", item.Price)
		itemRecord.Set("price_discount", item.Discount)
		itemRecord.Set("description", item.Description)
		itemRecord.Set("total", item.Total)
		itemRecord.Set("type", item.Type)
		itemRecord.Set("document_id", document_id)
		itemRecord.Set("uid", item.Uid)

		if err := app.Save(itemRecord); err != nil {
			return nil, err
		}

		for _, tax := range item.Taxes {
			dteTax := getTaxRecord(tax.Name, app)
			dteItemTax, err := app.FindFirstRecordByFilter("dte_item_tax", "tax_id = {:tax_id} && item_id = {:item_id}", dbx.Params{"tax_id": dteTax.Id, "item_id": itemRecord.Id})
			if err != nil {
				collection, _ := app.FindCollectionByNameOrId("dte_item_tax")
				dteItemTax = core.NewRecord(collection)
			}

			dteItemTax.Set("tax_id", dteTax.Id)
			dteItemTax.Set("item_id", itemRecord.Id)
			dteItemTax.Set("taxable_amount", tax.TaxableAmount)
			dteItemTax.Set("taxed_amount", tax.TaxedAmount)

			if err := app.Save(dteItemTax); err != nil {
				return nil, err
			}
		}

		item_ids = append(item_ids, itemRecord.Id)
	}

	return item_ids, nil
}

func AddDTE(app core.App, dtevalue *DTEXML) error {
	vendorRecord, err := addVendor(app, dtevalue)
	if err != nil {
		return err
	}

	recipientRecord, err := addRecipient(app, dtevalue)
	if err != nil {
		return err
	}

	records, _ := app.FindRecordsByFilter("dte_document", "uuid = {:uuid}", "", 1, 0, dbx.Params{"uuid": dtevalue.Certification.Uuid})
	dteDocumentRecord := &core.Record{}

	if len(records) == 0 {
		collection, _ := app.FindCollectionByNameOrId("dte_document")
		dteDocumentRecord = core.NewRecord(collection)
	} else {
		dteDocumentRecord = records[0]
	}

	totalf32, _ := strconv.ParseFloat(dtevalue.Total.Total, 32)
	totalTaxf32 := 0.0

	for _, v := range dtevalue.Total.Taxes {
		taxtotal, _ := strconv.ParseFloat(v.Total, 32)
		totalTaxf32 = totalTaxf32 + taxtotal
	}

	dteDocumentRecord.Set("total", totalf32)
	dteDocumentRecord.Set("total_taxes", totalTaxf32)
	dteDocumentRecord.Set("currency", dtevalue.Document.Currency)
	dteDocumentRecord.Set("emitted", dtevalue.Document.Timestamp)
	dteDocumentRecord.Set("type", dtevalue.Document.Type)
	dteDocumentRecord.Set("vendor_site_id", vendorRecord.Id)
	dteDocumentRecord.Set("recipient_id", recipientRecord.Id)
	dteDocumentRecord.Set("number", dtevalue.Certification.Number)
	dteDocumentRecord.Set("serie", dtevalue.Certification.Serie)
	dteDocumentRecord.Set("uuid", dtevalue.Certification.Uuid)

	err = app.Save(dteDocumentRecord)
	if err != nil {
		return err
	}

	document_id := dteDocumentRecord.Id
	item_ids, err := addItems(document_id, app, dtevalue)
	if err != nil {
		return err
	}

	dteDocumentRecord.Set("items", append(dteDocumentRecord.Get("items").([]string), item_ids...))
	err = app.Save(dteDocumentRecord)
	if err != nil {
		return err
	}

	for _, tax := range dtevalue.Total.Taxes {
		dteTax := getTaxRecord(tax.Name, app)
		dteDocumentTax, err := app.FindFirstRecordByFilter("dte_document_tax", "tax_id = {:tax_id} && document_id = {:document_id}", dbx.Params{"tax_id": dteTax.Id, "document_id": dteDocumentRecord.Id})
		if err != nil {
			collection, _ := app.FindCollectionByNameOrId("dte_document_tax")
			dteDocumentTax = core.NewRecord(collection)
		}

		dteDocumentTax.Set("tax_id", dteTax.Id)
		dteDocumentTax.Set("document_id", dteDocumentRecord.Id)
		dteDocumentTax.Set("taxed_amount", tax.Total)

		if err := app.Save(dteDocumentTax); err != nil {
			return err
		}
	}

	return nil
}
