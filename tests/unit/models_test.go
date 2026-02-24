package unit

import (
	"testing"

	"shoop-golang/internal/models"
	"shoop-golang/tests/testutil"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestBaseModel_BeforeCreate(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("auto_generates_uuid_when_empty", func(t *testing.T) {
		cat := models.Category{
			Name:     "Test",
			Slug:     "test",
			IsActive: true,
		}
		if cat.ID != "" {
			t.Fatalf("expected empty ID before create, got %q", cat.ID)
		}
		if err := db.Create(&cat).Error; err != nil {
			t.Fatalf("create failed: %v", err)
		}
		if cat.ID == "" {
			t.Fatal("expected UUID to be auto-generated")
		}
		if len(cat.ID) != 36 {
			t.Errorf("expected UUID format (36 chars), got len %d", len(cat.ID))
		}
	})

	t.Run("preserves_pre_set_id", func(t *testing.T) {
		customID := "custom-id-12345"
		cat := models.Category{
			BaseModel: models.BaseModel{ID: customID},
			Name:      "Test",
			Slug:      "test-preserve",
			IsActive:  true,
		}
		if err := db.Create(&cat).Error; err != nil {
			t.Fatalf("create failed: %v", err)
		}
		if cat.ID != customID {
			t.Errorf("expected ID %q to be preserved, got %q", customID, cat.ID)
		}
	})
}

func TestProduct_SalePercent(t *testing.T) {
	testutil.SetupTestDB(t)

	tests := []struct {
		name     string
		original float64
		sale     float64
		want     int
	}{
		{"original_100k_sale_80k", 100000, 80000, 20},
		{"original_100k_sale_0", 100000, 0, 0},
		{"original_0_sale_80k", 0, 80000, 0},
		{"original_100k_sale_100k", 100000, 100000, 0},
		{"original_100k_sale_120k", 100000, 120000, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := models.Product{OriginalPrice: tt.original, SalePrice: tt.sale}
			got := p.SalePercent()
			if got != tt.want {
				t.Errorf("SalePercent() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCategory_CRUD(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("create", func(t *testing.T) {
		cat := models.Category{
			Name:     "Electronics",
			Slug:     "electronics",
			IsActive: true,
		}
		if err := db.Create(&cat).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if cat.ID == "" {
			t.Fatal("expected ID after create")
		}
	})

	t.Run("read", func(t *testing.T) {
		cat := models.Category{Name: "Books", Slug: "books", IsActive: true}
		if err := db.Create(&cat).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		var read models.Category
		if err := db.First(&read, "id = ?", cat.ID).Error; err != nil {
			t.Fatalf("read: %v", err)
		}
		if read.Name != cat.Name || read.Slug != cat.Slug {
			t.Errorf("read mismatch: got %+v", read)
		}
	})

	t.Run("update", func(t *testing.T) {
		cat := models.Category{Name: "Toys", Slug: "toys", IsActive: true}
		if err := db.Create(&cat).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := db.Model(&cat).Update("Name", "Updated Toys").Error; err != nil {
			t.Fatalf("update: %v", err)
		}
		var read models.Category
		db.First(&read, "id = ?", cat.ID)
		if read.Name != "Updated Toys" {
			t.Errorf("update failed: got Name %q", read.Name)
		}
	})

	t.Run("delete", func(t *testing.T) {
		cat := models.Category{Name: "DeleteMe", Slug: "delete-me", IsActive: true}
		if err := db.Create(&cat).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := db.Delete(&cat).Error; err != nil {
			t.Fatalf("delete: %v", err)
		}
		var count int64
		db.Model(&models.Category{}).Where("id = ?", cat.ID).Count(&count)
		if count != 0 {
			t.Error("delete: record still exists")
		}
	})
}

func TestProduct_CRUD(t *testing.T) {
	db := testutil.SetupTestDB(t)

	cat := models.Category{Name: "Test Cat", Slug: "test-cat", IsActive: true}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	t.Run("create", func(t *testing.T) {
		p := models.Product{
			Name:          "Widget",
			Slug:          "widget",
			OriginalPrice: 100000,
			SalePrice:     80000,
			SKU:           "WID-001",
			CategoryID:    cat.ID,
			IsActive:      true,
		}
		if err := db.Create(&p).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if p.ID == "" {
			t.Fatal("expected ID after create")
		}
	})

	t.Run("read_with_preload", func(t *testing.T) {
		p := models.Product{
			Name:          "Preload Product",
			Slug:          "preload-product",
			OriginalPrice: 50000,
			SalePrice:     40000,
			SKU:           "PRE-001",
			CategoryID:    cat.ID,
			IsActive:      true,
		}
		if err := db.Create(&p).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		var read models.Product
		if err := db.Preload("Category").First(&read, "id = ?", p.ID).Error; err != nil {
			t.Fatalf("read: %v", err)
		}
		if read.Category.ID != cat.ID || read.Category.Name != cat.Name {
			t.Errorf("Preload Category failed: got %+v", read.Category)
		}
	})

	t.Run("update", func(t *testing.T) {
		p := models.Product{
			Name:          "Update Me",
			Slug:          "update-me",
			OriginalPrice: 100,
			SalePrice:     80,
			SKU:           "UPD-001",
			CategoryID:    cat.ID,
			IsActive:      true,
		}
		if err := db.Create(&p).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := db.Model(&p).Update("Name", "Updated Name").Error; err != nil {
			t.Fatalf("update: %v", err)
		}
		var read models.Product
		db.First(&read, "id = ?", p.ID)
		if read.Name != "Updated Name" {
			t.Errorf("update failed: got Name %q", read.Name)
		}
	})

	t.Run("delete", func(t *testing.T) {
		p := models.Product{
			Name:          "Delete Product",
			Slug:          "delete-product",
			OriginalPrice: 100,
			SalePrice:     80,
			SKU:           "DEL-001",
			CategoryID:    cat.ID,
			IsActive:      true,
		}
		if err := db.Create(&p).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := db.Delete(&p).Error; err != nil {
			t.Fatalf("delete: %v", err)
		}
		var count int64
		db.Model(&models.Product{}).Where("id = ?", p.ID).Count(&count)
		if count != 0 {
			t.Error("delete: record still exists")
		}
	})
}

func TestUser_CRUD(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("create_and_read", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
		u := models.User{
			Email:    "user@example.com",
			Password: string(hash),
			Name:     "Test User",
		}
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		var read models.User
		if err := db.First(&read, "id = ?", u.ID).Error; err != nil {
			t.Fatalf("read: %v", err)
		}
		if read.Password == "" {
			t.Error("expected password to be stored (not empty)")
		}
		if read.Name != u.Name {
			t.Errorf("name mismatch: got %q", read.Name)
		}
	})
}

func TestOrder_WithItems(t *testing.T) {
	db := testutil.SetupTestDB(t)

	user := models.User{
		Email:    "order@test.com",
		Password: "hashed",
		Name:     "Order User",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	cat := models.Category{Name: "Cat", Slug: "cat", IsActive: true}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	product := models.Product{
		Name:          "Order Product",
		Slug:          "order-product",
		OriginalPrice: 50000,
		SalePrice:     40000,
		SKU:           "ORD-001",
		CategoryID:    cat.ID,
		IsActive:      true,
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}

	order := models.Order{
		UserID:      user.ID,
		Status:      "pending",
		TotalAmount: 80000,
		Name:        "Customer",
		Phone:       "0909111222",
		Address:     "123 St",
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	items := []models.OrderItem{
		{OrderID: order.ID, ProductID: product.ID, Quantity: 2, Price: 40000},
	}
	if err := db.Create(&items).Error; err != nil {
		t.Fatalf("create order items: %v", err)
	}

	var read models.Order
	if err := db.Preload("Items").Preload("Items.Product").First(&read, "id = ?", order.ID).Error; err != nil {
		t.Fatalf("read with preload: %v", err)
	}
	if len(read.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(read.Items))
	}
	if read.Items[0].Quantity != 2 || read.Items[0].Price != 40000 {
		t.Errorf("item mismatch: got %+v", read.Items[0])
	}
	if read.Items[0].Product.ID != product.ID {
		t.Errorf("Preload Product failed: got product ID %q", read.Items[0].Product.ID)
	}
}

func TestBanner_CRUD(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("create", func(t *testing.T) {
		b := models.Banner{
			Title:    "Summer Sale",
			Subtitle: "50% off",
			Image:    "/banner.jpg",
			Link:     "/sale",
			IsActive: true,
		}
		if err := db.Create(&b).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if b.ID == "" {
			t.Fatal("expected ID after create")
		}
	})

	t.Run("read", func(t *testing.T) {
		b := models.Banner{Title: "Read Banner", Image: "/read.jpg", IsActive: true}
		if err := db.Create(&b).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		var read models.Banner
		if err := db.First(&read, "id = ?", b.ID).Error; err != nil {
			t.Fatalf("read: %v", err)
		}
		if read.Title != b.Title {
			t.Errorf("read mismatch: got Title %q", read.Title)
		}
	})

	t.Run("update", func(t *testing.T) {
		b := models.Banner{Title: "Update Banner", Image: "/upd.jpg", IsActive: true}
		if err := db.Create(&b).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := db.Model(&b).Update("Title", "Updated Banner").Error; err != nil {
			t.Fatalf("update: %v", err)
		}
		var read models.Banner
		db.First(&read, "id = ?", b.ID)
		if read.Title != "Updated Banner" {
			t.Errorf("update failed: got Title %q", read.Title)
		}
	})

	t.Run("delete", func(t *testing.T) {
		b := models.Banner{Title: "Delete Banner", Image: "/del.jpg", IsActive: true}
		if err := db.Create(&b).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := db.Delete(&b).Error; err != nil {
			t.Fatalf("delete: %v", err)
		}
		var count int64
		db.Model(&models.Banner{}).Where("id = ?", b.ID).Count(&count)
		if count != 0 {
			t.Error("delete: record still exists")
		}
	})
}

func TestCompanyInfo_CRUD(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("create", func(t *testing.T) {
		c := models.CompanyInfo{
			Name:    "Acme Corp",
			Tagline: "Best products",
			Email:   "info@acme.com",
		}
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if c.ID == "" {
			t.Fatal("expected ID after create")
		}
	})

	t.Run("read", func(t *testing.T) {
		c := models.CompanyInfo{Name: "Read Company", Email: "read@co.com"}
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		var read models.CompanyInfo
		if err := db.First(&read, "id = ?", c.ID).Error; err != nil {
			t.Fatalf("read: %v", err)
		}
		if read.Name != c.Name {
			t.Errorf("read mismatch: got Name %q", read.Name)
		}
	})

	t.Run("update", func(t *testing.T) {
		c := models.CompanyInfo{Name: "Update Company", Email: "upd@co.com"}
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := db.Model(&c).Update("Name", "Updated Company").Error; err != nil {
			t.Fatalf("update: %v", err)
		}
		var read models.CompanyInfo
		db.First(&read, "id = ?", c.ID)
		if read.Name != "Updated Company" {
			t.Errorf("update failed: got Name %q", read.Name)
		}
	})
}

func TestAboutPage_CRUD(t *testing.T) {
	db := testutil.SetupTestDB(t)

	t.Run("create", func(t *testing.T) {
		a := models.AboutPage{
			Title:   "About Us",
			Content: "We are great.",
			Image:   "/about.jpg",
		}
		if err := db.Create(&a).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if a.ID == "" {
			t.Fatal("expected ID after create")
		}
	})

	t.Run("read", func(t *testing.T) {
		a := models.AboutPage{Title: "Read About", Content: "Content"}
		if err := db.Create(&a).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		var read models.AboutPage
		if err := db.First(&read, "id = ?", a.ID).Error; err != nil {
			t.Fatalf("read: %v", err)
		}
		if read.Title != a.Title {
			t.Errorf("read mismatch: got Title %q", read.Title)
		}
	})

	t.Run("update", func(t *testing.T) {
		a := models.AboutPage{Title: "Update About", Content: "Old"}
		if err := db.Create(&a).Error; err != nil {
			t.Fatalf("create: %v", err)
		}
		if err := db.Model(&a).Update("Content", "New content").Error; err != nil {
			t.Fatalf("update: %v", err)
		}
		var read models.AboutPage
		db.First(&read, "id = ?", a.ID)
		if read.Content != "New content" {
			t.Errorf("update failed: got Content %q", read.Content)
		}
	})
}

func TestSoftDelete(t *testing.T) {
	db := testutil.SetupTestDB(t)

	cat := models.Category{Name: "Soft Cat", Slug: "soft-cat", IsActive: true}
	if err := db.Create(&cat).Error; err != nil {
		t.Fatalf("create category: %v", err)
	}

	p := models.Product{
		Name:          "Soft Delete Product",
		Slug:          "soft-delete-product",
		OriginalPrice: 100,
		SalePrice:     80,
		SKU:           "SOFT-001",
		CategoryID:    cat.ID,
		IsActive:      true,
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := db.Delete(&p).Error; err != nil {
		t.Fatalf("delete: %v", err)
	}

	var normal models.Product
	err := db.First(&normal, "id = ?", p.ID).Error
	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound for normal query after soft delete, got %v", err)
	}

	var unscoped models.Product
	if err := db.Unscoped().First(&unscoped, "id = ?", p.ID).Error; err != nil {
		t.Fatalf("Unscoped query failed: %v", err)
	}
	if unscoped.ID != p.ID {
		t.Errorf("Unscoped should find soft-deleted record: got ID %q", unscoped.ID)
	}
}
