package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"shoop-golang/database"
	"shoop-golang/internal/models"
	"shoop-golang/tests/testutil"
)

func TestAdminLogin_Success(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/admin/login", nil, url.Values{
		"email":    {"admin@test.com"},
		"password": {"admin123"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/dashboard" {
		t.Errorf("expected Location /admin/dashboard, got %s", loc)
	}
}

func TestAdminLogin_InvalidPassword(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/admin/login", nil, url.Values{
		"email":    {"admin@test.com"},
		"password": {"wrongpassword"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminLogin_NonexistentUser(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/admin/login", nil, url.Values{
		"email":    {"nonexistent@test.com"},
		"password": {"any"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminLogout(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.GetWithCookies(ts, "/admin/logout", cookies)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/login" {
		t.Errorf("expected Location /admin/login, got %s", loc)
	}
}

func TestAdminDashboard_Unauthenticated(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.GetWithCookies(ts, "/admin/dashboard", nil)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/login" {
		t.Errorf("expected Location /admin/login, got %s", loc)
	}
}

func TestAdminDashboard_Authenticated(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.GetWithCookies(ts, "/admin/dashboard", cookies)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminCategories_List(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.GetWithCookies(ts, "/admin/categories", cookies)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminCategories_Create(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/admin/categories", cookies, url.Values{
		"name":        {"New Category"},
		"description": {"A test category"},
		"sort_order":  {"0"},
		"is_active":   {"on"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/categories" {
		t.Errorf("expected Location /admin/categories, got %s", loc)
	}

	var count int64
	database.DB.Model(&models.Category{}).Where("name = ?", "New Category").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 category created, got %d", count)
	}
}

func TestAdminCategories_Update(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)
	cat := testutil.CreateTestCategory(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/admin/categories/"+cat.ID, cookies, url.Values{
		"name":        {"Updated Category"},
		"description": {"Updated desc"},
		"sort_order":  {"1"},
		"is_active":   {"on"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}

	var updated models.Category
	database.DB.First(&updated, "id = ?", cat.ID)
	if updated.Name != "Updated Category" {
		t.Errorf("expected name Updated Category, got %s", updated.Name)
	}
}

func TestAdminCategories_Delete(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)
	cat := testutil.CreateTestCategory(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/admin/categories/"+cat.ID+"/delete", cookies, url.Values{})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}

	var count int64
	database.DB.Model(&models.Category{}).Where("id = ?", cat.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected category deleted, count=%d", count)
	}
}

func TestAdminProducts_List(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.GetWithCookies(ts, "/admin/products", cookies)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminProducts_Create(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)
	cat := testutil.CreateTestCategory(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/admin/products", cookies, url.Values{
		"name":           {"New Product"},
		"description":    {"A test product"},
		"content":        {""},
		"original_price": {"100000"},
		"sale_price":     {"80000"},
		"sku":            {"NEW-001"},
		"stock":          {"10"},
		"category_id":    {cat.ID},
		"is_active":      {"on"},
		"is_featured":    {"on"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/admin/products" {
		t.Errorf("expected Location /admin/products, got %s", loc)
	}

	var count int64
	database.DB.Model(&models.Product{}).Where("name = ?", "New Product").Count(&count)
	if count != 1 {
		t.Errorf("expected 1 product created, got %d", count)
	}
}

func TestAdminProducts_Update(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)
	cat := testutil.CreateTestCategory(t)
	prod := testutil.CreateTestProduct(t, cat.ID)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/admin/products/"+prod.ID, cookies, url.Values{
		"name":           {"Updated Product"},
		"description":    {"Updated desc"},
		"content":        {""},
		"original_price": {"120000"},
		"sale_price":     {"90000"},
		"sku":            {"UPD-001"},
		"stock":          {"5"},
		"category_id":    {cat.ID},
		"is_active":      {"on"},
		"is_featured":    {""},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}

	var updated models.Product
	database.DB.First(&updated, "id = ?", prod.ID)
	if updated.Name != "Updated Product" {
		t.Errorf("expected name Updated Product, got %s", updated.Name)
	}
}

func TestAdminProducts_Delete(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)
	cat := testutil.CreateTestCategory(t)
	prod := testutil.CreateTestProduct(t, cat.ID)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/admin/products/"+prod.ID+"/delete", cookies, url.Values{})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}

	var count int64
	database.DB.Model(&models.Product{}).Where("id = ?", prod.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected product deleted, count=%d", count)
	}
}

func TestAdminOrders_List(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.GetWithCookies(ts, "/admin/orders", cookies)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminOrders_Detail(t *testing.T) {
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
	resp, err := testutil.GetWithCookies(ts, "/admin/orders/"+order.ID, cookies)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminOrders_UpdateStatus(t *testing.T) {
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
	resp, err := testutil.PostForm(ts, "/admin/orders/"+order.ID+"/status", cookies, url.Values{
		"status": {"confirmed"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}

	var updated models.Order
	database.DB.First(&updated, "id = ?", order.ID)
	if updated.Status != "confirmed" {
		t.Errorf("expected status confirmed, got %s", updated.Status)
	}
}

func TestAdminUsers_List(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.GetWithCookies(ts, "/admin/users", cookies)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminUsers_Detail(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)
	user := testutil.CreateTestUser(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.GetWithCookies(ts, "/admin/users/"+user.ID, cookies)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAdminBanners_CRUD(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)

	// Create
	resp, err := testutil.PostForm(ts, "/admin/banners", cookies, url.Values{
		"title":      {"New Banner"},
		"subtitle":   {"Subtitle"},
		"link":       {"/products"},
		"sort_order": {"0"},
		"is_active":  {"on"},
	})
	if err != nil {
		t.Fatalf("post create failed: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("create: expected 302, got %d", resp.StatusCode)
	}

	var banner models.Banner
	database.DB.Where("title = ?", "New Banner").First(&banner)
	if banner.ID == "" {
		t.Fatal("banner not created")
	}

	// List
	resp, err = testutil.GetWithCookies(ts, "/admin/banners", cookies)
	if err != nil {
		t.Fatalf("get list failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("list: expected 200, got %d", resp.StatusCode)
	}

	// Update
	resp, err = testutil.PostForm(ts, "/admin/banners/"+banner.ID, cookies, url.Values{
		"title":      {"Updated Banner"},
		"subtitle":   {"Updated sub"},
		"link":       {"/about"},
		"sort_order": {"1"},
		"is_active":  {"on"},
	})
	if err != nil {
		t.Fatalf("post update failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("update: expected 302, got %d", resp.StatusCode)
	}

	var updated models.Banner
	database.DB.First(&updated, "id = ?", banner.ID)
	if updated.Title != "Updated Banner" {
		t.Errorf("expected title Updated Banner, got %s", updated.Title)
	}

	// Delete
	resp, err = testutil.PostForm(ts, "/admin/banners/"+banner.ID+"/delete", cookies, url.Values{})
	if err != nil {
		t.Fatalf("post delete failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("delete: expected 302, got %d", resp.StatusCode)
	}

	var count int64
	database.DB.Model(&models.Banner{}).Where("id = ?", banner.ID).Count(&count)
	if count != 0 {
		t.Errorf("expected banner deleted, count=%d", count)
	}
}

func TestAdminCompany_Update(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/admin/company", cookies, url.Values{
		"name":         {"Test Company Inc"},
		"tagline":      {"New tagline"},
		"email":        {"company@test.com"},
		"phone":       {"0909123456"},
		"address":     {"123 Test St"},
		"facebook_url": {"https://facebook.com/test"},
		"zalo_url":    {"https://zalo.me/test"},
		"copyright":   {"© 2026 Test"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}

	var info models.CompanyInfo
	database.DB.First(&info)
	if info.Name != "Test Company Inc" {
		t.Errorf("expected name Test Company Inc, got %s", info.Name)
	}
	if info.Email != "company@test.com" {
		t.Errorf("expected email company@test.com, got %s", info.Email)
	}
}

func TestAdminAbout_Update(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()
	testutil.CreateTestAdmin(t)

	e := testutil.NewAdminEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.AdminLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/admin/about", cookies, url.Values{
		"title":   {"Updated About Title"},
		"content": {"<p>Updated about content</p>"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}

	var about models.AboutPage
	database.DB.First(&about)
	if about.Title != "Updated About Title" {
		t.Errorf("expected title Updated About Title, got %s", about.Title)
	}
	if about.Content != "<p>Updated about content</p>" {
		t.Errorf("expected content updated, got %s", about.Content)
	}
}
