package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/tests/testutil"
)

// clientWithJar returns an http.Client with cookiejar that does not auto-follow redirects.
func clientWithJar(jar *cookiejar.Jar) *http.Client {
	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func TestE2E_CustomerRegistrationAndPurchaseFlow(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()

	cat := testutil.CreateTestCategory(t)
	prod1 := testutil.CreateTestProduct(t, cat.ID)
	prod2 := models.Product{
		Name:          "Another Product",
		Slug:          "another-product",
		Description:   "Another test product",
		OriginalPrice: 50000,
		SalePrice:     45000,
		SKU:           "TEST-002",
		Stock:         5,
		CategoryID:    cat.ID,
		IsActive:      true,
	}
	database.DB.Create(&prod2)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := clientWithJar(jar)

	// 1. Visit homepage
	resp, err := client.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("get homepage: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 1: expected 200, got %d", resp.StatusCode)
	}

	// 2. Browse products
	resp, err = client.Get(ts.URL + "/products")
	if err != nil {
		t.Fatalf("get products: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 2: expected 200, got %d", resp.StatusCode)
	}

	// 3. Try to add to cart without login → 401
	resp, err = client.PostForm(ts.URL+"/cart/add", url.Values{"product_id": {prod1.ID}})
	if err != nil {
		t.Fatalf("cart add unauthenticated: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("step 3: expected 401, got %d", resp.StatusCode)
	}
	var errBody map[string]string
	json.NewDecoder(resp.Body).Decode(&errBody)
	if errBody["error"] != "login_required" {
		t.Errorf("step 3: expected error login_required, got %s", errBody["error"])
	}

	// 4. Register new account
	resp, err = client.PostForm(ts.URL+"/register", url.Values{
		"name":     {"E2E Customer"},
		"email":    {"e2e@test.com"},
		"password": {"password123"},
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 4: expected 200, got %d", resp.StatusCode)
	}

	// 5. Add product to cart → cartCount=1
	resp, err = client.PostForm(ts.URL+"/cart/add", url.Values{"product_id": {prod1.ID}})
	if err != nil {
		t.Fatalf("cart add: %v", err)
	}
	var addBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&addBody)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 5: expected 200, got %d", resp.StatusCode)
	}
	if n, ok := addBody["cartCount"].(float64); !ok || int(n) != 1 {
		t.Errorf("step 5: expected cartCount=1, got %v", addBody["cartCount"])
	}

	// 6. Add same product again → cartCount=2
	resp, err = client.PostForm(ts.URL+"/cart/add", url.Values{"product_id": {prod1.ID}})
	if err != nil {
		t.Fatalf("cart add same: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&addBody)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 6: expected 200, got %d", resp.StatusCode)
	}
	if n, ok := addBody["cartCount"].(float64); !ok || int(n) != 2 {
		t.Errorf("step 6: expected cartCount=2, got %v", addBody["cartCount"])
	}

	// 7. Add different product → cartCount=3
	resp, err = client.PostForm(ts.URL+"/cart/add", url.Values{"product_id": {prod2.ID}})
	if err != nil {
		t.Fatalf("cart add different: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&addBody)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 7: expected 200, got %d", resp.StatusCode)
	}
	if n, ok := addBody["cartCount"].(float64); !ok || int(n) != 3 {
		t.Errorf("step 7: expected cartCount=3, got %v", addBody["cartCount"])
	}

	// 8. View cart
	resp, err = client.Get(ts.URL + "/cart")
	if err != nil {
		t.Fatalf("get cart: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 8: expected 200, got %d", resp.StatusCode)
	}

	// 9. Update cart: decrease first product → cartCount=2
	resp, err = client.PostForm(ts.URL+"/cart/update", url.Values{
		"product_id": {prod1.ID},
		"action":     {"decrease"},
	})
	if err != nil {
		t.Fatalf("cart update: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&addBody)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 9: expected 200, got %d", resp.StatusCode)
	}
	if n, ok := addBody["cartCount"].(float64); !ok || int(n) != 2 {
		t.Errorf("step 9: expected cartCount=2, got %v", addBody["cartCount"])
	}

	// 10. Checkout
	resp, err = client.PostForm(ts.URL+"/checkout", url.Values{
		"name":    {"E2E Customer"},
		"phone":   {"0909111222"},
		"address": {"123 E2E Street"},
	})
	if err != nil {
		t.Fatalf("checkout: %v", err)
	}
	var checkoutBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&checkoutBody)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 10: expected 200, got %d", resp.StatusCode)
	}
	orderID, ok := checkoutBody["order_id"].(string)
	if !ok || orderID == "" {
		t.Fatalf("step 10: expected order_id in response")
	}

	// 11. Verify order in DB with correct total and 2 items
	var order models.Order
	if err := database.DB.Preload("Items").First(&order, "id = ?", orderID).Error; err != nil {
		t.Fatalf("step 11: order not found: %v", err)
	}
	if len(order.Items) != 2 {
		t.Errorf("step 11: expected 2 items, got %d", len(order.Items))
	}
	expectedTotal := 80000.0 + 45000.0 // prod1 sale 80000, prod2 sale 45000
	if order.TotalAmount != expectedTotal {
		t.Errorf("step 11: expected total %.0f, got %.0f", expectedTotal, order.TotalAmount)
	}

	// 12. Verify cart is empty after checkout
	resp, err = client.Get(ts.URL + "/cart")
	if err != nil {
		t.Fatalf("get cart after checkout: %v", err)
	}
	// Cart page renders - we verify by adding again and checking count
	resp.Body.Close()
	resp, err = client.PostForm(ts.URL+"/cart/add", url.Values{"product_id": {prod1.ID}})
	if err != nil {
		t.Fatalf("cart add after checkout: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&addBody)
	resp.Body.Close()
	if n, ok := addBody["cartCount"].(float64); !ok || int(n) != 1 {
		t.Errorf("step 12: expected cartCount=1 (cart was cleared), got %v", addBody["cartCount"])
	}
}

func TestE2E_AdminCategoryProductManagement(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := clientWithJar(jar)

	// 1. Login as admin
	resp, err := client.PostForm(ts.URL+"/admin/login", url.Values{
		"email":    {"admin@test.com"},
		"password": {"admin123"},
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 1: expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/dashboard" {
		t.Errorf("step 1: expected Location /admin/dashboard, got %s", loc)
	}

	// 2. View dashboard
	resp, err = client.Get(ts.URL + "/admin/dashboard")
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 2: expected 200, got %d", resp.StatusCode)
	}

	// 3. Create category
	resp, err = client.PostForm(ts.URL+"/admin/categories", url.Values{
		"name":        {"E2E Category"},
		"description": {"E2E desc"},
		"sort_order": {"0"},
		"is_active":  {"on"},
	})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 3: expected 302, got %d", resp.StatusCode)
	}

	// 4. Verify category in DB
	var cat models.Category
	if err := database.DB.Where("name = ?", "E2E Category").First(&cat).Error; err != nil {
		t.Fatalf("step 4: category not found: %v", err)
	}

	// 5. Create product
	resp, err = client.PostForm(ts.URL+"/admin/products", url.Values{
		"name":           {"E2E Product"},
		"description":    {"E2E product desc"},
		"content":        {""},
		"original_price": {"100000"},
		"sale_price":     {"80000"},
		"sku":            {"E2E-001"},
		"stock":          {"10"},
		"category_id":    {cat.ID},
		"is_active":      {"on"},
		"is_featured":    {"on"},
	})
	if err != nil {
		t.Fatalf("create product: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 5: expected 302, got %d", resp.StatusCode)
	}

	// 6. Verify product in DB
	var prod models.Product
	if err := database.DB.Where("name = ?", "E2E Product").First(&prod).Error; err != nil {
		t.Fatalf("step 6: product not found: %v", err)
	}
	if prod.CategoryID != cat.ID {
		t.Errorf("step 6: expected category_id %s, got %s", cat.ID, prod.CategoryID)
	}

	// 7. Update product price
	resp, err = client.PostForm(ts.URL+"/admin/products/"+prod.ID, url.Values{
		"name":           {"E2E Product"},
		"description":    {"E2E product desc"},
		"content":        {""},
		"original_price": {"120000"},
		"sale_price":     {"95000"},
		"sku":            {"E2E-001"},
		"stock":          {"10"},
		"category_id":    {cat.ID},
		"is_active":      {"on"},
		"is_featured":    {"on"},
	})
	if err != nil {
		t.Fatalf("update product: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 7: expected 302, got %d", resp.StatusCode)
	}

	// 8. Verify updated price in DB
	database.DB.First(&prod, "id = ?", prod.ID)
	if prod.SalePrice != 95000 {
		t.Errorf("step 8: expected sale_price 95000, got %.0f", prod.SalePrice)
	}

	// 9. Delete product (soft delete)
	resp, err = client.PostForm(ts.URL+"/admin/products/"+prod.ID+"/delete", url.Values{})
	if err != nil {
		t.Fatalf("delete product: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 9: expected 302, got %d", resp.StatusCode)
	}

	// 10. Verify product soft-deleted
	var count int64
	database.DB.Model(&models.Product{}).Where("id = ?", prod.ID).Count(&count)
	if count != 0 {
		t.Errorf("step 10: expected product soft-deleted (count=0), got %d", count)
	}
	database.DB.Unscoped().Model(&models.Product{}).Where("id = ?", prod.ID).Count(&count)
	if count != 1 {
		t.Errorf("step 10: expected product exists with DeletedAt, got count %d", count)
	}

	// 11. Delete category (soft delete)
	resp, err = client.PostForm(ts.URL+"/admin/categories/"+cat.ID+"/delete", url.Values{})
	if err != nil {
		t.Fatalf("delete category: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 11: expected 302, got %d", resp.StatusCode)
	}

	// 12. Verify category soft-deleted
	database.DB.Model(&models.Category{}).Where("id = ?", cat.ID).Count(&count)
	if count != 0 {
		t.Errorf("step 12: expected category soft-deleted, got %d", count)
	}
}

func TestE2E_AdminOrderManagement(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)
	user := testutil.CreateTestUser(t)
	cat := testutil.CreateTestCategory(t)
	prod := testutil.CreateTestProduct(t, cat.ID)
	order := testutil.CreateTestOrder(t, user.ID, prod.ID)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)

	// 3. View orders list
	resp, err := testutil.GetWithCookies(ts, "/admin/orders", cookies)
	if err != nil {
		t.Fatalf("orders list: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 3: expected 200, got %d", resp.StatusCode)
	}

	// 4. View order detail
	resp, err = testutil.GetWithCookies(ts, "/admin/orders/"+order.ID, cookies)
	if err != nil {
		t.Fatalf("order detail: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 4: expected 200, got %d", resp.StatusCode)
	}

	// 5. Update status to confirmed
	resp, err = testutil.PostForm(ts, "/admin/orders/"+order.ID+"/status", cookies, url.Values{"status": {"confirmed"}})
	if err != nil {
		t.Fatalf("update status: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 5: expected 302, got %d", resp.StatusCode)
	}
	var o models.Order
	database.DB.First(&o, "id = ?", order.ID)
	if o.Status != "confirmed" {
		t.Errorf("step 5: expected status confirmed, got %s", o.Status)
	}

	// 6-7. Update to shipping
	resp, err = testutil.PostForm(ts, "/admin/orders/"+order.ID+"/status", cookies, url.Values{"status": {"shipping"}})
	if err != nil {
		t.Fatalf("update status shipping: %v", err)
	}
	resp.Body.Close()
	database.DB.First(&o, "id = ?", order.ID)
	if o.Status != "shipping" {
		t.Errorf("step 7: expected status shipping, got %s", o.Status)
	}

	// 8. Update to delivered
	resp, err = testutil.PostForm(ts, "/admin/orders/"+order.ID+"/status", cookies, url.Values{"status": {"delivered"}})
	if err != nil {
		t.Fatalf("update status delivered: %v", err)
	}
	resp.Body.Close()
	database.DB.First(&o, "id = ?", order.ID)
	if o.Status != "delivered" {
		t.Errorf("step 8: expected status delivered, got %s", o.Status)
	}

	// 9. Try invalid status → verify DB unchanged
	resp, err = testutil.PostForm(ts, "/admin/orders/"+order.ID+"/status", cookies, url.Values{"status": {"invalid_status"}})
	if err != nil {
		t.Fatalf("update invalid status: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 9: expected 302 (redirect), got %d", resp.StatusCode)
	}
	database.DB.First(&o, "id = ?", order.ID)
	if o.Status != "delivered" {
		t.Errorf("step 9: expected status unchanged (delivered), got %s", o.Status)
	}
}

func TestE2E_AdminBannerManagement(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)

	// 2. Create banner (POST without image - handler allows empty image)
	resp, err := testutil.PostForm(ts, "/admin/banners", cookies, url.Values{
		"title":      {"E2E Banner"},
		"subtitle":   {"E2E Subtitle"},
		"link":       {"/products"},
		"sort_order": {"0"},
		"is_active":  {"on"},
	})
	if err != nil {
		t.Fatalf("create banner: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 2: expected 302, got %d", resp.StatusCode)
	}

	var banner models.Banner
	if err := database.DB.Where("title = ?", "E2E Banner").First(&banner).Error; err != nil {
		t.Fatalf("step 3: banner not found: %v", err)
	}

	// 4. Update banner title
	resp, err = testutil.PostForm(ts, "/admin/banners/"+banner.ID, cookies, url.Values{
		"title":      {"Updated E2E Banner"},
		"subtitle":   {"E2E Subtitle"},
		"link":       {"/products"},
		"sort_order": {"0"},
		"is_active":  {"on"},
	})
	if err != nil {
		t.Fatalf("update banner: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 4: expected 302, got %d", resp.StatusCode)
	}

	// 5. Verify updated
	database.DB.First(&banner, "id = ?", banner.ID)
	if banner.Title != "Updated E2E Banner" {
		t.Errorf("step 5: expected title Updated E2E Banner, got %s", banner.Title)
	}

	// 6. Delete banner
	resp, err = testutil.PostForm(ts, "/admin/banners/"+banner.ID+"/delete", cookies, url.Values{})
	if err != nil {
		t.Fatalf("delete banner: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 6: expected 302, got %d", resp.StatusCode)
	}

	// 7. Verify soft-deleted
	var count int64
	database.DB.Model(&models.Banner{}).Where("id = ?", banner.ID).Count(&count)
	if count != 0 {
		t.Errorf("step 7: expected banner soft-deleted, got %d", count)
	}
}

func TestE2E_AdminCompanyAndAboutManagement(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	adminEcho := testutil.NewAdminEcho()
	webEcho := testutil.NewWebEcho()
	tsAdmin := httptest.NewServer(adminEcho)
	tsWeb := httptest.NewServer(webEcho)
	defer tsAdmin.Close()
	defer tsWeb.Close()

	cookies := testutil.AdminLoginCookies(t, tsAdmin)

	// 2. Update company info
	resp, err := testutil.PostForm(tsAdmin, "/admin/company", cookies, url.Values{
		"name":         {"E2E Company Inc"},
		"tagline":      {"E2E Tagline"},
		"email":        {"e2e@company.com"},
		"phone":        {"0909123456"},
		"address":      {"456 E2E Ave"},
		"facebook_url": {"https://facebook.com/e2e"},
		"zalo_url":     {"https://zalo.me/e2e"},
		"copyright":    {"© 2026 E2E"},
	})
	if err != nil {
		t.Fatalf("update company: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 2: expected 302, got %d", resp.StatusCode)
	}

	// 3. Verify company info in DB
	var info models.CompanyInfo
	database.DB.First(&info)
	if info.Name != "E2E Company Inc" {
		t.Errorf("step 3: expected name E2E Company Inc, got %s", info.Name)
	}
	if info.Email != "e2e@company.com" {
		t.Errorf("step 3: expected email e2e@company.com, got %s", info.Email)
	}

	// 4. Update about page
	resp, err = testutil.PostForm(tsAdmin, "/admin/about", cookies, url.Values{
		"title":   {"E2E About Title"},
		"content": {"<p>E2E about content</p>"},
	})
	if err != nil {
		t.Fatalf("update about: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 4: expected 302, got %d", resp.StatusCode)
	}

	// 5. Verify about page in DB
	var about models.AboutPage
	database.DB.First(&about)
	if about.Title != "E2E About Title" {
		t.Errorf("step 5: expected title E2E About Title, got %s", about.Title)
	}
	if about.Content != "<p>E2E about content</p>" {
		t.Errorf("step 5: expected content updated, got %s", about.Content)
	}

	// 6. Visit web about page
	resp, err = http.Get(tsWeb.URL + "/about")
	if err != nil {
		t.Fatalf("get about: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 6: expected 200, got %d", resp.StatusCode)
	}
}

func TestE2E_MultipleUsersCartIsolation(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()

	cat := testutil.CreateTestCategory(t)
	prodX := testutil.CreateTestProduct(t, cat.ID)
	prodY := models.Product{
		Name:          "Product Y",
		Slug:          "product-y",
		Description:   "Product Y",
		OriginalPrice: 100000,
		SalePrice:     90000,
		SKU:           "TEST-Y",
		Stock:         5,
		CategoryID:    cat.ID,
		IsActive:      true,
	}
	database.DB.Create(&prodY)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	// User A: register, add product X
	jarA, _ := cookiejar.New(nil)
	clientA := &http.Client{Jar: jarA, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	_, err := clientA.PostForm(ts.URL+"/register", url.Values{
		"name":     {"User A"},
		"email":    {"usera@test.com"},
		"password": {"pass123"},
	})
	if err != nil {
		t.Fatalf("user A register: %v", err)
	}
	_, err = clientA.PostForm(ts.URL+"/cart/add", url.Values{"product_id": {prodX.ID}})
	if err != nil {
		t.Fatalf("user A add to cart: %v", err)
	}

	// User B: register (different cookie jar), add product Y
	jarB, _ := cookiejar.New(nil)
	clientB := &http.Client{Jar: jarB, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	_, err = clientB.PostForm(ts.URL+"/register", url.Values{
		"name":     {"User B"},
		"email":    {"userb@test.com"},
		"password": {"pass123"},
	})
	if err != nil {
		t.Fatalf("user B register: %v", err)
	}
	resp, err := clientB.PostForm(ts.URL+"/cart/add", url.Values{"product_id": {prodY.ID}})
	if err != nil {
		t.Fatalf("user B add to cart: %v", err)
	}
	var bodyB map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&bodyB)
	resp.Body.Close()
	if n, ok := bodyB["cartCount"].(float64); !ok || int(n) != 1 {
		t.Errorf("user B: expected cartCount=1, got %v", bodyB["cartCount"])
	}

	// User A checks cart → only product X
	resp, err = clientA.Get(ts.URL + "/cart")
	if err != nil {
		t.Fatalf("user A get cart: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("user A cart: expected 200, got %d", resp.StatusCode)
	}
	// Cart page renders - we verify by adding product Y again for user A and checking count
	// If user A had product Y, adding again would increase count. User A should only have X.
	resp, err = clientA.PostForm(ts.URL+"/cart/add", url.Values{"product_id": {prodX.ID}})
	if err != nil {
		t.Fatalf("user A add X again: %v", err)
	}
	var bodyA map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&bodyA)
	resp.Body.Close()
	if n, ok := bodyA["cartCount"].(float64); !ok || int(n) != 2 {
		t.Errorf("user A: expected cartCount=2 (only X, added twice), got %v", bodyA["cartCount"])
	}

	// User B checks cart → only product Y
	resp, err = clientB.PostForm(ts.URL+"/cart/add", url.Values{"product_id": {prodY.ID}})
	if err != nil {
		t.Fatalf("user B add Y again: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&bodyB)
	resp.Body.Close()
	if n, ok := bodyB["cartCount"].(float64); !ok || int(n) != 2 {
		t.Errorf("user B: expected cartCount=2 (only Y, added twice), got %v", bodyB["cartCount"])
	}
}

func TestE2E_ProductSearchAndFilter(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := &http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}

	// 2. Browse all products
	resp, err := client.Get(ts.URL + "/products")
	if err != nil {
		t.Fatalf("get products: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 2: expected 200, got %d", resp.StatusCode)
	}

	// 3. Filter by category (seed has tuong-phong-thuy)
	resp, err = client.Get(ts.URL + "/products?category=tuong-phong-thuy")
	if err != nil {
		t.Fatalf("get products by category: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 3: expected 200, got %d", resp.StatusCode)
	}

	// 4. Search by name
	resp, err = client.Get(ts.URL + "/products?q=Phong")
	if err != nil {
		t.Fatalf("get products by search: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 4: expected 200, got %d", resp.StatusCode)
	}

	// 5. View product detail (seed has tuong-phat-di-lac-ngoc-bich)
	resp, err = client.Get(ts.URL + "/products/tuong-phat-di-lac-ngoc-bich")
	if err != nil {
		t.Fatalf("get product detail: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 5: expected 200, got %d", resp.StatusCode)
	}
}

func TestE2E_AuthenticationEdgeCases(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)

	adminEcho := testutil.NewAdminEcho()
	webEcho := testutil.NewWebEcho()
	tsAdmin := httptest.NewServer(adminEcho)
	tsWeb := httptest.NewServer(webEcho)
	defer tsAdmin.Close()
	defer tsWeb.Close()

	client := clientWithJar(nil)
	jar, _ := cookiejar.New(nil)
	client.Jar = jar

	// 1. Try accessing admin dashboard without login → 302 to login
	resp, err := client.Get(tsAdmin.URL + "/admin/dashboard")
	if err != nil {
		t.Fatalf("dashboard unauthenticated: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 1: expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/login" {
		t.Errorf("step 1: expected Location /admin/login, got %s", loc)
	}

	// 2. Login with wrong password → 200 (stays on login page)
	resp, err = client.PostForm(tsAdmin.URL+"/admin/login", url.Values{
		"email":    {"admin@test.com"},
		"password": {"wrongpassword"},
	})
	if err != nil {
		t.Fatalf("wrong password login: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("step 2: expected 200, got %d", resp.StatusCode)
	}

	testutil.CreateTestAdmin(t)

	// 3. Login successfully → 302 to dashboard
	resp, err = client.PostForm(tsAdmin.URL+"/admin/login", url.Values{
		"email":    {"admin@test.com"},
		"password": {"admin123"},
	})
	if err != nil {
		t.Fatalf("admin login: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 3: expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/dashboard" {
		t.Errorf("step 3: expected Location /admin/dashboard, got %s", loc)
	}

	// 4. Logout → 302 to login
	resp, err = client.Get(tsAdmin.URL + "/admin/logout")
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 4: expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/login" {
		t.Errorf("step 4: expected Location /admin/login, got %s", loc)
	}

	// 5. Try accessing dashboard after logout → 302 to login
	resp, err = client.Get(tsAdmin.URL + "/admin/dashboard")
	if err != nil {
		t.Fatalf("dashboard after logout: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 5: expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/login" {
		t.Errorf("step 5: expected Location /admin/login, got %s", loc)
	}

	// 6. Web: register with duplicate email → 400
	resp, err = client.PostForm(tsWeb.URL+"/register", url.Values{
		"name":     {"Another"},
		"email":    {"user@test.com"},
		"password": {"pass123"},
	})
	if err != nil {
		t.Fatalf("duplicate register: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("step 6: expected 400, got %d", resp.StatusCode)
	}

	// 7. Web: login, logout, verify cart is cleared
	jarWeb, _ := cookiejar.New(nil)
	clientWeb := &http.Client{Jar: jarWeb, CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }}
	cat := testutil.CreateTestCategory(t)
	prod := testutil.CreateTestProduct(t, cat.ID)

	_, err = clientWeb.PostForm(tsWeb.URL+"/login", url.Values{
		"email":    {"user@test.com"},
		"password": {"user123"},
	})
	if err != nil {
		t.Fatalf("web login: %v", err)
	}
	_, err = clientWeb.PostForm(tsWeb.URL+"/cart/add", url.Values{"product_id": {prod.ID}})
	if err != nil {
		t.Fatalf("add to cart: %v", err)
	}
	resp, err = clientWeb.Get(tsWeb.URL + "/logout")
	if err != nil {
		t.Fatalf("web logout: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("step 7: expected 302 on logout, got %d", resp.StatusCode)
	}
	// After logout, session is cleared - adding to cart should require login again
	resp, err = clientWeb.PostForm(tsWeb.URL+"/cart/add", url.Values{"product_id": {prod.ID}})
	if err != nil {
		t.Fatalf("cart add after logout: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("step 7: expected 401 (cart cleared/session gone), got %d", resp.StatusCode)
	}
}
