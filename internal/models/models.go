package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string         `gorm:"type:text;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// Admin user for the back-office
type AdminUser struct {
	BaseModel
	Email    string `gorm:"uniqueIndex;not null" json:"email"`
	Password string `gorm:"not null" json:"-"`
	Name     string `gorm:"not null" json:"name"`
	Role     string `gorm:"default:admin" json:"role"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
}

// End-user / customer
type User struct {
	BaseModel
	Email    string  `gorm:"uniqueIndex;not null" json:"email"`
	Password string  `gorm:"not null" json:"-"`
	Name     string  `gorm:"not null" json:"name"`
	Phone    string  `json:"phone"`
	Address  string  `json:"address"`
	Orders   []Order `gorm:"foreignKey:UserID" json:"orders,omitempty"`
}

type Category struct {
	BaseModel
	Name        string    `gorm:"not null" json:"name"`
	Slug        string    `gorm:"uniqueIndex;not null" json:"slug"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	Products    []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

type Product struct {
	BaseModel
	Name          string   `gorm:"not null" json:"name"`
	Slug          string   `gorm:"uniqueIndex;not null" json:"slug"`
	Description   string   `gorm:"type:text" json:"description"`
	Content       string   `gorm:"type:text" json:"content"`
	OriginalPrice float64  `gorm:"not null" json:"original_price"`
	SalePrice     float64  `json:"sale_price"`
	SKU           string   `gorm:"uniqueIndex" json:"sku"`
	Stock         int      `gorm:"default:0" json:"stock"`
	CategoryID    string   `gorm:"index" json:"category_id"`
	Category      Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images        []Image  `gorm:"foreignKey:ProductID" json:"images,omitempty"`
	IsActive      bool     `gorm:"default:true" json:"is_active"`
	IsFeatured    bool     `gorm:"default:false" json:"is_featured"`
}

func (p Product) SalePercent() int {
	if p.OriginalPrice <= 0 || p.SalePrice <= 0 || p.SalePrice >= p.OriginalPrice {
		return 0
	}
	return int(((p.OriginalPrice - p.SalePrice) / p.OriginalPrice) * 100)
}

type Image struct {
	BaseModel
	ProductID string `gorm:"index" json:"product_id"`
	URL       string `gorm:"not null" json:"url"`
	AltText   string `json:"alt_text"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	IsPrimary bool   `gorm:"default:false" json:"is_primary"`
}

type Order struct {
	BaseModel
	UserID      string      `gorm:"index;not null" json:"user_id"`
	User        User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Status      string      `gorm:"default:pending" json:"status"` // pending, confirmed, shipping, delivered, cancelled
	TotalAmount float64     `gorm:"not null" json:"total_amount"`
	Name        string      `json:"name"`
	Phone       string      `json:"phone"`
	Address     string      `gorm:"type:text" json:"address"`
	Note        string      `gorm:"type:text" json:"note"`
	Items       []OrderItem `gorm:"foreignKey:OrderID" json:"items,omitempty"`
}

type OrderItem struct {
	BaseModel
	OrderID   string  `gorm:"index;not null" json:"order_id"`
	ProductID string  `gorm:"index;not null" json:"product_id"`
	Product   Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity  int     `gorm:"not null" json:"quantity"`
	Price     float64 `gorm:"not null" json:"price"`
}

type Banner struct {
	BaseModel
	Title     string `gorm:"not null" json:"title"`
	Subtitle  string `json:"subtitle"`
	Image     string `gorm:"not null" json:"image"`
	Link      string `json:"link"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`
}

type CompanyInfo struct {
	BaseModel
	Name        string `gorm:"not null" json:"name"`
	Tagline     string `json:"tagline"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
	LogoURL     string `json:"logo_url"`
	FacebookURL string `json:"facebook_url"`
	ZaloURL     string `json:"zalo_url"`
	Copyright   string `json:"copyright"`
}

type AboutPage struct {
	BaseModel
	Title   string `gorm:"not null" json:"title"`
	Content string `gorm:"type:text" json:"content"`
	Image   string `json:"image"`
}

type SEOBanner struct {
	BaseModel
	Page        string `gorm:"not null;index" json:"page"` // home, products, about, contact
	Title       string `json:"title"`
	Description string `gorm:"type:text" json:"description"`
	Keywords    string `json:"keywords"`
	OGImage     string `json:"og_image"`
}

// Cart item stored in session for anonymous users, or DB for logged-in
type CartItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Image     string  `json:"image"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
}
