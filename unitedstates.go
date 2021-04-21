package bills

// Downloads and processes bill metadata and text
// Go implementation of the unitedstates/congress functions to:
// ./run govinfo --collections=BILLS --congress=$CONGRESSNUM --extract=mods,xml,premis --bulkdata=BILLSTATUS
// ./run bills

/* From the unitedstates/congress repository documentation:
This process has two parts. First, the XML data must be fetched from Govinfo. This script pulls the bill status XML and on subsequent runs only pulls new and changed files:

./run govinfo --bulkdata=BILLSTATUS
Then run the bills task to process any new and changed files:

./run bills
It's recommended to do this two-step process no more than every 6 hours, as the data is not updated more frequently than that (and often really only once daily).
*/

// globals
var (
	GOVINFO_BASE_URL                = "https://www.govinfo.gov/"
	COLLECTION_BASE_URL             = GOVINFO_BASE_URL + "app/details/"
	BULKDATA_BASE_URL               = GOVINFO_BASE_URL + "bulkdata/"
	COLLECTION_SITEMAPINDEX_PATTERN = GOVINFO_BASE_URL + "sitemap/{collection}_sitemap_index.xml"
	BULKDATA_SITEMAPINDEX_PATTERN   = GOVINFO_BASE_URL + "sitemap/bulkdata/{collection}/sitemapindex.xml"
	FDSYS_BILLSTATUS_FILENAME       = "fdsys_billstatus.xml"
	// for xpath
	NS = map[string]string{"x": "http://www.sitemaps.org/schemas/sitemap/0.9"}
)
