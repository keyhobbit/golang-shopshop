package seeders

import (
	"log"

	"shoop-golang/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	seedAdminUser(db)
	seedCompanyInfo(db)
	seedAboutPage(db)
	seedCategories(db)
	seedProducts(db)
	seedBanners(db)
	log.Println("Seeding completed")
}

func hashPassword(pw string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h)
}

func seedAdminUser(db *gorm.DB) {
	var count int64
	db.Model(&models.AdminUser{}).Count(&count)
	if count > 0 {
		return
	}
	db.Create(&models.AdminUser{
		Email:    "admin@occ.io.vn",
		Password: hashPassword("admin123"),
		Name:     "Super Admin",
		Role:     "admin",
		IsActive: true,
	})
	log.Println("Seeded admin user: admin@occ.io.vn / admin123")
}

func seedCompanyInfo(db *gorm.DB) {
	var count int64
	db.Model(&models.CompanyInfo{}).Count(&count)
	if count > 0 {
		return
	}
	db.Create(&models.CompanyInfo{
		Name:      "OCC.IO.VN",
		Tagline:   "Phong Thủy - Hài Hòa Năng Lượng",
		Email:     "contact@occ.io.vn",
		Phone:     "0909 123 456",
		Address:   "123 Nguyễn Huệ, Quận 1, TP.HCM",
		Copyright: "© 2026 OCC.IO.VN. All rights reserved.",
	})
}

func seedAboutPage(db *gorm.DB) {
	var count int64
	db.Model(&models.AboutPage{}).Count(&count)
	if count > 0 {
		return
	}
	db.Create(&models.AboutPage{
		Title:   "Về Chúng Tôi",
		Content: "<p>OCC.IO.VN chuyên cung cấp các sản phẩm phong thủy chất lượng cao, mang đến sự hài hòa và may mắn cho không gian sống của bạn.</p><p>Với hơn 10 năm kinh nghiệm, chúng tôi tự hào là địa chỉ uy tín hàng đầu trong lĩnh vực phong thủy.</p>",
	})
}

func seedCategories(db *gorm.DB) {
	var count int64
	db.Model(&models.Category{}).Count(&count)
	if count > 0 {
		return
	}
	categories := []models.Category{
		{Name: "Tượng Phong Thủy", Slug: "tuong-phong-thuy", Description: "Các loại tượng phong thủy mang lại may mắn", SortOrder: 1, IsActive: true},
		{Name: "Vòng Tay Phong Thủy", Slug: "vong-tay-phong-thuy", Description: "Vòng tay đá phong thủy hợp mệnh", SortOrder: 2, IsActive: true},
		{Name: "Đá Phong Thủy", Slug: "da-phong-thuy", Description: "Đá quý phong thủy tự nhiên", SortOrder: 3, IsActive: true},
		{Name: "Cây Phong Thủy", Slug: "cay-phong-thuy", Description: "Cây cảnh phong thủy cho không gian sống", SortOrder: 4, IsActive: true},
		{Name: "Tranh Phong Thủy", Slug: "tranh-phong-thuy", Description: "Tranh phong thủy trang trí nội thất", SortOrder: 5, IsActive: true},
	}
	db.Create(&categories)
}

func seedProducts(db *gorm.DB) {
	var count int64
	db.Model(&models.Product{}).Count(&count)
	if count > 0 {
		return
	}

	var categories []models.Category
	db.Find(&categories)
	if len(categories) == 0 {
		return
	}

	products := []models.Product{
		{Name: "Tượng Phật Di Lặc Ngọc Bích", Slug: "tuong-phat-di-lac-ngoc-bich", Description: "Tượng Phật Di Lặc bằng ngọc bích tự nhiên, mang lại may mắn và tài lộc", OriginalPrice: 2500000, SalePrice: 1990000, SKU: "TPT-001", Stock: 15, CategoryID: categories[0].ID, IsActive: true, IsFeatured: true},
		{Name: "Tượng Tỳ Hưu Vàng", Slug: "tuong-ty-huu-vang", Description: "Tỳ Hưu vàng phong thủy chiêu tài lộc", OriginalPrice: 3200000, SalePrice: 2690000, SKU: "TPT-002", Stock: 10, CategoryID: categories[0].ID, IsActive: true, IsFeatured: true},
		{Name: "Vòng Tay Thạch Anh Hồng", Slug: "vong-tay-thach-anh-hong", Description: "Vòng tay thạch anh hồng tự nhiên, hợp mệnh Hỏa", OriginalPrice: 850000, SalePrice: 650000, SKU: "VT-001", Stock: 30, CategoryID: categories[1].ID, IsActive: true, IsFeatured: true},
		{Name: "Vòng Tay Mắt Hổ", Slug: "vong-tay-mat-ho", Description: "Vòng tay đá mắt hổ mang lại sức mạnh và bảo vệ", OriginalPrice: 750000, SalePrice: 590000, SKU: "VT-002", Stock: 25, CategoryID: categories[1].ID, IsActive: true, IsFeatured: false},
		{Name: "Thạch Anh Tím Tự Nhiên", Slug: "thach-anh-tim-tu-nhien", Description: "Khối thạch anh tím tự nhiên, thanh lọc năng lượng", OriginalPrice: 4500000, SalePrice: 3800000, SKU: "DA-001", Stock: 5, CategoryID: categories[2].ID, IsActive: true, IsFeatured: true},
		{Name: "Đá Fluorite Cầu Vồng", Slug: "da-fluorite-cau-vong", Description: "Đá Fluorite nhiều màu sắc, tăng cường trí tuệ", OriginalPrice: 1200000, SalePrice: 980000, SKU: "DA-002", Stock: 12, CategoryID: categories[2].ID, IsActive: true, IsFeatured: false},
		{Name: "Cây Kim Tiền Phong Thủy", Slug: "cay-kim-tien-phong-thuy", Description: "Cây kim tiền mang lại tài lộc cho gia chủ", OriginalPrice: 500000, SalePrice: 420000, SKU: "CT-001", Stock: 20, CategoryID: categories[3].ID, IsActive: true, IsFeatured: true},
		{Name: "Cây Lưỡi Hổ", Slug: "cay-luoi-ho", Description: "Cây lưỡi hổ thanh lọc không khí, hút tài lộc", OriginalPrice: 350000, SalePrice: 0, SKU: "CT-002", Stock: 18, CategoryID: categories[3].ID, IsActive: true, IsFeatured: false},
		{Name: "Tranh Mã Đáo Thành Công", Slug: "tranh-ma-dao-thanh-cong", Description: "Tranh ngựa phong thủy mang lại thành công", OriginalPrice: 1800000, SalePrice: 1500000, SKU: "TR-001", Stock: 8, CategoryID: categories[4].ID, IsActive: true, IsFeatured: true},
		{Name: "Tranh Cửu Ngư Quần Hội", Slug: "tranh-cuu-ngu-quan-hoi", Description: "Tranh 9 con cá phong thủy, biểu tượng thịnh vượng", OriginalPrice: 2200000, SalePrice: 1850000, SKU: "TR-002", Stock: 6, CategoryID: categories[4].ID, IsActive: true, IsFeatured: false},
	}
	db.Create(&products)
}

func seedBanners(db *gorm.DB) {
	var count int64
	db.Model(&models.Banner{}).Count(&count)
	if count > 0 {
		return
	}
	banners := []models.Banner{
		{Title: "Phong Thủy Cho Mọi Nhà", Subtitle: "Khám phá bộ sưu tập phong thủy độc đáo", Image: "/static/images/banners/banner1.jpg", Link: "/products", SortOrder: 1, IsActive: true},
		{Title: "Giảm Giá Đến 30%", Subtitle: "Ưu đãi đặc biệt cho sản phẩm phong thủy", Image: "/static/images/banners/banner2.jpg", Link: "/products", SortOrder: 2, IsActive: true},
		{Title: "Vòng Tay Phong Thủy", Subtitle: "Bộ sưu tập vòng tay đá quý mới nhất", Image: "/static/images/banners/banner3.jpg", Link: "/products?category=vong-tay-phong-thuy", SortOrder: 3, IsActive: true},
	}
	db.Create(&banners)
}
