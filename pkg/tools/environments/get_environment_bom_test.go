package environments

import (
   "context"
   "testing"
   "github.com/patrickcping/pingone-go-sdk-v2/management"
)

// TestGetEnvironmentBomTool_Success verifies BOM is extracted correctly
func TestGetEnvironmentBomTool_Success(t *testing.T) {
   envID := "env1"
   // Build a fake environment with BOM containing one product
   prodID := "p1"
   prod := management.NewBillOfMaterialsProductsInner(management.ENUMPRODUCTTYPE_ONE_BASE)
   prod.Id = &prodID
   env := management.NewEnvironmentWithDefaults()
   env.Id = &envID
   env.BillOfMaterials = management.NewBillOfMaterials([]management.BillOfMaterialsProductsInner{*prod})
   // Populate required fields
   env.Region = management.EnumRegionCodeAsEnvironmentRegion(management.ENUMREGIONCODE_NA.Ptr())
   env.Type = management.ENUMENVIRONMENTTYPE_PRODUCTION
   env.License = *management.NewEnvironmentLicense("")
   fake := &fakeGetEnvironmentClient{expectedID: envID, returnedEnv: *env}
   tool := NewGetEnvironmentBomTool(fake)

   out, err := tool.Run(context.Background(), map[string]interface{}{"id": envID})
   if err != nil {
       t.Fatalf("unexpected error: %v", err)
   }
   bomIface, ok := out["bill_of_materials"].(map[string]interface{})
   if !ok {
       t.Fatalf("expected map[string]interface{}, got %T", out["bill_of_materials"])
   }
   prods, ok := bomIface["products"].([]interface{})
   if !ok {
       t.Fatalf("expected products to be []interface{}, got %T", bomIface["products"])
   }
   if len(prods) != 1 {
       t.Errorf("expected 1 product, got %d", len(prods))
   }
   first := prods[0].(map[string]interface{})
   if first["id"] != prodID {
       t.Errorf("expected product id %s, got %v", prodID, first["id"])
   }
}

// TestGetEnvironmentBomTool_MissingArgs verifies missing id error
func TestGetEnvironmentBomTool_MissingArgs(t *testing.T) {
   tool := NewGetEnvironmentBomTool(nil)
  if _, err := tool.Run(context.Background(), map[string]interface{}{}); err == nil {
       t.Fatal("expected error for missing id, got nil")
   }
}