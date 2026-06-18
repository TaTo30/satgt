app.routes.superuserOnly("#/satgt", (route) => {
  const internalData = store({
    month: "",
    year: "",
    type: "R",
    isLoading: false,
    responseMessage: ""
  })

  const today = new Date()
  return t.div({ className: "page page-content full-height" },
    t.div({}, 
      t.h3({ style: "padding: 0.3em;" }, "Retrieve DTE Documents"),
      t.span({ style: "font-style: italic; color: var(--accentTxtHintColor); padding: 0.5em" }, "This will open an instance of Google Chrome (Chrome is required) to retrieve credentials of SAT DTE subsystem. The window will close automatically until the process finish to scrap the needed data. No credentials are stored in any form."),
      t.div({ style: "display: flex; flex-direction: row; gap: 1em; padding: 0.5em" }, 
        t.div({ style: "display: flex; flex-direction: column "}, 
          t.label({ for: "satgt-dte-month" }, "Month"),
          app.components.select({
            options: [
              {value: 1, label: "January"},
              {value: 2, label: "February"},
              {value: 3, label: "March"},
              {value: 4, label: "April"},
              {value: 5, label: "May"},
              {value: 6, label: "June"},
              {value: 7, label: "July"},
              {value: 8, label: "August"},
              {value: 9, label: "September"},
              {value: 10, label: "October"},
              {value: 11, label: "November"},
              {value: 12, label: "December"},
            ],
            placeholder: "All months",
            id: "satgt-dte-month",
            onchange: (selected) => {
              internalData.month = selected[0].value
            }
          })
        ),
        
        t.div({ style: "display: flex; flex-direction: column" }, 
          t.label({ for: "satgt-dte-year" }, "Year"),
          app.components.select({
            options: () => {
              const minYear = 2010
              const maxYear = today.getFullYear()
              const yearOptions = []
              for(let year = maxYear; year >= minYear; year--){
                yearOptions.push({value: year, label: year})
              }
              return yearOptions
            },
            placeholder: "Current year",
            id: "satgt-dte-year",
            onchange: (selected) => {
              internalData.year = selected[0].value
            }
          })
        ),

        t.div({ style: "display: flex; flex-direction: column "}, 
          t.label({ for: "satgt-dte-type" }, "Type"),
          app.components.select({
            options: [
              {value: "R", label: "Received"},
              {value: "E", label: "Emitted"},
            ],
            value: () => internalData.type,
            required: true,
            placeholder: "Operation Type",
            id: "satgt-dte-type",
            onchange: (selected) => {
              internalData.type = selected[0].value
            }
          })
        ),

        t.button({ 
          className: "btn outline", 
          type: "button", 
          onclick: async () => {
              const dteUrl = new URL("/satgt/dte", window.location.origin) 
              const queryParams = new URLSearchParams()

              queryParams.set("type", internalData.type)
              if (internalData.year) 
                queryParams.set("year", internalData.year)
             
              if (internalData.month)
                queryParams.set("month", internalData.month)

              dteUrl.search = queryParams.toString()
              internalData.isLoading = true

              const response = await fetch(dteUrl)
              internalData.responseMessage = await response.text()
              internalData.isLoading = false
            } 
          },
          t.span({ className: "txt" },
          "Request Data"
          )
        )
      ),
      t.div({ className: "txt", style: "padding: 0.5em" }, () => internalData.isLoading ? "Processing..." : internalData.responseMessage)
    )
  )
})

app.store.headerLinks.push({
  href: "#/satgt",
  label: "SAT",
  icon: "ri-currency-fill"
})
