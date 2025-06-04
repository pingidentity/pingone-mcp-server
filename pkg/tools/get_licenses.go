package tools

import (
  "context"
  "encoding/json"
  "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// getLicensesSchema defines the input JSON schema for get_licenses (no inputs)
var getLicensesSchema = json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`)

// FetchLicensesFunc fetches pages of licenses
type FetchLicensesFunc func(ctx context.Context, organizationID string) management.EntityArrayPagedIterator

// GetLicensesTool lists all licenses for the configured organization
type GetLicensesTool struct {
  fetcher FetchLicensesFunc
  organization string
}

// NewGetLicensesTool constructs a GetLicensesTool using a fetcher function
func NewGetLicensesTool(fetcher FetchLicensesFunc, organization string) Tool {
   return &GetLicensesTool{fetcher: fetcher, organization: organization}
}

// Name returns the tool name
// Name returns the tool name
func (t *GetLicensesTool) Name() string { return "get_licenses" }

// Description returns a human-readable description
// Description returns a human-readable description
func (t *GetLicensesTool) Description() string { return "Retrieve all licenses for the organization" }

// InputSchema returns the JSON schema for tool arguments
// InputSchema returns the JSON schema for tool arguments
func (t *GetLicensesTool) InputSchema() json.RawMessage { return getLicensesSchema }

// Run executes the tool
// Run executes the tool logic
func (t *GetLicensesTool) Run(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
  iterator := t.fetcher(ctx, t.organization)
  var out []map[string]interface{}
  var runErr error
  iterator(func(cursor management.PagedCursor, err error) bool {
      if err != nil {
          runErr = err
          return false
      }
      arr := cursor.EntityArray
      if arr != nil && arr.Embedded != nil {
          for _, lic := range arr.Embedded.Licenses {
              data, _ := json.Marshal(lic)
              var m map[string]interface{}
              json.Unmarshal(data, &m)
              out = append(out, m)
          }
      }
      return true
  })
  if runErr != nil {
      return nil, runErr
  }
  return map[string]interface{}{"licenses": out}, nil
}