package api

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

func TestWebHome(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWebProducts_List(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/products")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWebProducts_ListWithCategory(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/products?category=tuong-phong-thuy")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWebProducts_ListWithSearch(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/products?q=Phong")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWebProducts_Detail(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	cat := testutil.CreateTestCategory(t)
	testutil.CreateTestProduct(t, cat.ID)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/products/test-product")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWebProducts_DetailNotFound(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(ts.URL + "/products/nonexistent")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/products" {
		t.Errorf("expected Location /products, got %s", loc)
	}
}

func TestWebAbout(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/about")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWebContact(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/contact")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWebRegister_Success(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/register", nil, url.Values{
		"name":     {"New User"},
		"email":    {"newuser@test.com"},
		"password": {"password123"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

func TestWebRegister_DuplicateEmail(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/register", nil, url.Values{
		"name":     {"Another User"},
		"email":    {"user@test.com"},
		"password": {"password123"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected error field in response")
	}
}

func TestWebRegister_MissingFields(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/register", nil, url.Values{
		"name":     {""},
		"email":    {""},
		"password": {""},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected error field in response")
	}
}

func TestWebLogin_Success(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/login", nil, url.Values{
		"email":    {"user@test.com"},
		"password": {"user123"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

func TestWebLogin_InvalidPassword(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/login", nil, url.Values{
		"email":    {"user@test.com"},
		"password": {"wrongpassword"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestWebLogin_NonexistentUser(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/login", nil, url.Values{
		"email":    {"nonexistent@test.com"},
		"password": {"any"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestWebLogout(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.WebLoginCookies(t, ts)
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, _ := http.NewRequest("GET", ts.URL+"/logout", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Errorf("expected 302, got %d", resp.StatusCode)
	}
	if loc := resp.Header.Get("Location"); loc != "/" {
		t.Errorf("expected Location /, got %s", loc)
	}
}

func TestWebCart_Page(t *testing.T) {
	testutil.SetupTestDBWithSeed(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/cart")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestWebCart_AddUnauthenticated(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	cat := testutil.CreateTestCategory(t)
	prod := testutil.CreateTestProduct(t, cat.ID)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/cart/add", nil, url.Values{
		"product_id": {prod.ID},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["error"] != "login_required" {
		t.Errorf("expected error login_required, got %s", body["error"])
	}
}

func TestWebCart_AddAuthenticated(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)
	cat := testutil.CreateTestCategory(t)
	prod := testutil.CreateTestProduct(t, cat.ID)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.WebLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/cart/add", cookies, url.Values{
		"product_id": {prod.ID},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
	if _, ok := body["cartCount"]; !ok {
		t.Error("expected cartCount in response")
	}
}

func TestWebCart_Update(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)
	cat := testutil.CreateTestCategory(t)
	prod := testutil.CreateTestProduct(t, cat.ID)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// Login
	_, err := client.PostForm(ts.URL+"/login", url.Values{
		"email":    {"user@test.com"},
		"password": {"user123"},
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Add to cart
	_, err = client.PostForm(ts.URL+"/cart/add", url.Values{
		"product_id": {prod.ID},
	})
	if err != nil {
		t.Fatalf("cart add failed: %v", err)
	}

	// Update cart (increase)
	resp, err := client.PostForm(ts.URL+"/cart/update", url.Values{
		"product_id": {prod.ID},
		"action":     {"increase"},
	})
	if err != nil {
		t.Fatalf("cart update failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
}

func TestWebCart_Remove(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)
	cat := testutil.CreateTestCategory(t)
	prod := testutil.CreateTestProduct(t, cat.ID)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// Login
	_, err := client.PostForm(ts.URL+"/login", url.Values{
		"email":    {"user@test.com"},
		"password": {"user123"},
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Add to cart
	_, err = client.PostForm(ts.URL+"/cart/add", url.Values{
		"product_id": {prod.ID},
	})
	if err != nil {
		t.Fatalf("cart add failed: %v", err)
	}

	// Remove from cart
	resp, err := client.PostForm(ts.URL+"/cart/update", url.Values{
		"product_id": {prod.ID},
		"action":     {"remove"},
	})
	if err != nil {
		t.Fatalf("cart update failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %v", body["status"])
	}
}

func TestWebCheckout_Unauthenticated(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	resp, err := testutil.PostForm(ts, "/checkout", nil, url.Values{
		"name":    {"Test"},
		"phone":   {"0909111222"},
		"address": {"123 Test St"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["error"] != "login_required" {
		t.Errorf("expected error login_required, got %s", body["error"])
	}
}

func TestWebCheckout_EmptyCart(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	cookies := testutil.WebLoginCookies(t, ts)
	resp, err := testutil.PostForm(ts, "/checkout", cookies, url.Values{
		"name":    {"Test"},
		"phone":   {"0909111222"},
		"address": {"123 Test St"},
	})
	if err != nil {
		t.Fatalf("post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected error in response")
	}
}

func TestWebCheckout_Success(t *testing.T) {
	testutil.SetupTestDB(t)
	testutil.SetupSession()
	testutil.CreateTestUser(t)
	cat := testutil.CreateTestCategory(t)
	prod := testutil.CreateTestProduct(t, cat.ID)

	e := testutil.NewWebEcho()
	ts := httptest.NewServer(e)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	// Login
	_, err := client.PostForm(ts.URL+"/login", url.Values{
		"email":    {"user@test.com"},
		"password": {"user123"},
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	// Add to cart
	_, err = client.PostForm(ts.URL+"/cart/add", url.Values{
		"product_id": {prod.ID},
	})
	if err != nil {
		t.Fatalf("cart add failed: %v", err)
	}

	// Checkout
	resp, err := client.PostForm(ts.URL+"/checkout", url.Values{
		"name":    {"Checkout User"},
		"phone":   {"0909111222"},
		"address": {"456 Checkout St"},
	})
	if err != nil {
		t.Fatalf("checkout failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body["success"] != true {
		t.Errorf("expected success true, got %v", body["success"])
	}
	orderID, ok := body["order_id"].(string)
	if !ok || orderID == "" {
		t.Error("expected order_id in response")
	}

	// Verify order in DB
	var order models.Order
	if err := database.DB.Preload("Items").First(&order, "id = ?", orderID).Error; err != nil {
		t.Fatalf("order not found in DB: %v", err)
	}
	if order.Name != "Checkout User" {
		t.Errorf("expected order name Checkout User, got %s", order.Name)
	}
	if order.Address != "456 Checkout St" {
		t.Errorf("expected address 456 Checkout St, got %s", order.Address)
	}
	if len(order.Items) != 1 {
		t.Errorf("expected 1 order item, got %d", len(order.Items))
	}
}
