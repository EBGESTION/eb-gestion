package main

import (
	"crypto/rand"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed web/*
var embedded embed.FS

type User struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Password string `json:"-"`
}
type Account struct {
	Username     string `json:"username"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	PasswordHash string `json:"passwordHash"`
	Active       bool   `json:"active"`
}
type Product struct {
	ID                   int                `json:"id"`
	Code                 string             `json:"code"`
	Name                 string             `json:"name"`
	Category             string             `json:"category"`
	Unit                 string             `json:"unit"`
	Warehouse            string             `json:"warehouse"`
	Stock                float64            `json:"stock"`
	WarehouseStocks      map[string]float64 `json:"warehouseStocks,omitempty"`
	MinStock             float64            `json:"minStock"`
	Cost                 float64            `json:"cost"`
	MainSupplierID       int                `json:"mainSupplierId"`
	AlternateSupplierIDs []int              `json:"alternateSupplierIds"`
	Notes                string             `json:"notes"`
	Active               bool               `json:"active"`
}
type Supplier struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	BusinessName  string   `json:"businessName"`
	RUT           string   `json:"rut"`
	Category      string   `json:"category"`
	Contact       string   `json:"contact"`
	Phone         string   `json:"phone"`
	Email         string   `json:"email"`
	OrderDays     []string `json:"orderDays"`
	OrderDeadline string   `json:"orderDeadline"`
	DeliveryDays  []string `json:"deliveryDays"`
	PaymentMethod string   `json:"paymentMethod"`
	PaymentTerms  string   `json:"paymentTerms"`
	MinimumOrder  float64  `json:"minimumOrder"`
	Notes         string   `json:"notes"`
	LastPurchase  string   `json:"lastPurchase"`
	LastDispatch  string   `json:"lastDispatch"`
	Active        bool     `json:"active"`
}
type PurchaseItem struct {
	ID                int     `json:"id"`
	ProductID         int     `json:"productId"`
	Name              string  `json:"name"`
	Unit              string  `json:"unit"`
	Warehouse         string  `json:"warehouse"`
	Quantity          float64 `json:"quantity"`
	UnitPrice         float64 `json:"unitPrice"`
	Subtotal          float64 `json:"subtotal"`
	ReceivedQuantity  float64 `json:"receivedQuantity"`
	LastReceivedPrice float64 `json:"lastReceivedPrice"`
	Note              string  `json:"note"`
}
type PurchaseReceiptItem struct {
	ProductID int     `json:"productId"`
	Name      string  `json:"name"`
	Quantity  float64 `json:"quantity"`
	Unit      string  `json:"unit"`
	UnitPrice float64 `json:"unitPrice"`
	Warehouse string  `json:"warehouse"`
	Note      string  `json:"note"`
}
type PurchaseReceipt struct {
	ID            int                   `json:"id"`
	ReceivedAt    string                `json:"receivedAt"`
	ReceivedBy    string                `json:"receivedBy"`
	Document      string                `json:"document"`
	IssueDate     string                `json:"issueDate"`
	DueDate       string                `json:"dueDate"`
	PaymentMethod string                `json:"paymentMethod"`
	PaymentStatus string                `json:"paymentStatus"`
	PaymentDate   string                `json:"paymentDate"`
	Note          string                `json:"note"`
	Items         []PurchaseReceiptItem `json:"items"`
	Total         float64               `json:"total"`
}
type Expense struct {
	ID            int     `json:"id"`
	Source        string  `json:"source"`
	Category      string  `json:"category"`
	SupplierID    int     `json:"supplierId"`
	SupplierName  string  `json:"supplierName"`
	Document      string  `json:"document"`
	IssueDate     string  `json:"issueDate"`
	DueDate       string  `json:"dueDate"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"paymentMethod"`
	Status        string  `json:"status"`
	PaymentDate   string  `json:"paymentDate"`
	Note          string  `json:"note"`
	PurchaseID    int     `json:"purchaseId"`
	ReceiptID     int     `json:"receiptId"`
	CreatedAt     string  `json:"createdAt"`
	CreatedBy     string  `json:"createdBy"`
}
type PriceHistory struct {
	ID           int     `json:"id"`
	ProductID    int     `json:"productId"`
	ProductName  string  `json:"productName"`
	SupplierID   int     `json:"supplierId"`
	SupplierName string  `json:"supplierName"`
	PurchaseID   int     `json:"purchaseId"`
	OrderNumber  string  `json:"orderNumber"`
	UnitPrice    float64 `json:"unitPrice"`
	Quantity     float64 `json:"quantity"`
	CreatedAt    string  `json:"createdAt"`
}
type PurchaseOrder struct {
	ID                    int               `json:"id"`
	Number                string            `json:"number"`
	SupplierID            int               `json:"supplierId"`
	SupplierName          string            `json:"supplierName"`
	Status                string            `json:"status"`
	CreatedAt             string            `json:"createdAt"`
	UpdatedAt             string            `json:"updatedAt"`
	ExpectedDate          string            `json:"expectedDate"`
	OrderedAt             string            `json:"orderedAt"`
	CreatedBy             string            `json:"createdBy"`
	Note                  string            `json:"note"`
	Items                 []PurchaseItem    `json:"items"`
	Total                 float64           `json:"total"`
	ReceivedTotal         float64           `json:"receivedTotal"`
	Receipts              []PurchaseReceipt `json:"receipts"`
	ClosedAt              string            `json:"closedAt"`
	ClosedBy              string            `json:"closedBy"`
	CloseReason           string            `json:"closeReason"`
	ClosedWithDifferences bool              `json:"closedWithDifferences"`
}
type RequestItem struct {
	ID           int      `json:"id"`
	ProductID    int      `json:"productId"`
	Name         string   `json:"name"`
	Unit         string   `json:"unit"`
	Reason       string   `json:"reason"`
	Stock        float64  `json:"stock"`
	RequestedQty float64  `json:"requestedQty"`
	DeliveredQty *float64 `json:"deliveredQty"`
}
type Request struct {
	ID                   int           `json:"id"`
	Requester            string        `json:"requester"`
	Area                 string        `json:"area"`
	Status               string        `json:"status"`
	CreatedAt            string        `json:"createdAt"`
	UpdatedAt            string        `json:"updatedAt"`
	DeliveredBy          string        `json:"deliveredBy"`
	DestinationWarehouse string        `json:"destinationWarehouse"`
	Note                 string        `json:"note"`
	Items                []RequestItem `json:"items"`
}
type Movement struct {
	ID        int     `json:"id"`
	ProductID int     `json:"productId"`
	Warehouse string  `json:"warehouse"`
	Type      string  `json:"type"`
	Quantity  float64 `json:"quantity"`
	Balance   float64 `json:"balance"`
	Reason    string  `json:"reason"`
	User      string  `json:"user"`
	CreatedAt string  `json:"createdAt"`
}
type Audit struct {
	ID        int    `json:"id"`
	User      string `json:"user"`
	Action    string `json:"action"`
	CreatedAt string `json:"createdAt"`
}
type Store struct {
	Products          []Product           `json:"products"`
	Suppliers         []Supplier          `json:"suppliers"`
	Purchases         []PurchaseOrder     `json:"purchases"`
	PriceHistory      []PriceHistory      `json:"priceHistory"`
	Expenses          []Expense           `json:"expenses"`
	Requests          []Request           `json:"requests"`
	Movements         []Movement          `json:"movements"`
	Audit             []Audit             `json:"audit"`
	NotificationReads map[string][]string `json:"notificationReads"`
	Accounts          map[string]Account  `json:"accounts"`
	SuggestionSnoozed map[string]string   `json:"suggestionSnoozed"`
}

type Session struct {
	User      User
	ExpiresAt time.Time
}

type App struct {
	mu       sync.RWMutex
	path     string
	store    Store
	sessions map[string]Session
}

func passwordHash(v string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(v))) }
func defaultAccounts() map[string]Account {
	return map[string]Account{
		"stefy":   {Username: "stefy", Name: "Stefy", Role: "admin_bodega", PasswordHash: passwordHash("123456"), Active: true},
		"claudio": {Username: "claudio", Name: "Claudio", Role: "gerencia", PasswordHash: passwordHash("123456"), Active: true},
		"dani":    {Username: "dani", Name: "Dani", Role: "bodega", PasswordHash: passwordHash("123456"), Active: true},
		"jessica": {Username: "jessica", Name: "Jessica", Role: "cocina", PasswordHash: passwordHash("123456"), Active: true},
		"caja":    {Username: "caja", Name: "Caja", Role: "caja", PasswordHash: passwordHash("123456"), Active: true},
	}
}
func accountUser(a Account) User { return User{Username: a.Username, Name: a.Name, Role: a.Role} }

func now() string { return time.Now().Format("2006-01-02 15:04:05") }
func seed() Store {
	return Store{
		Accounts: defaultAccounts(), SuggestionSnoozed: map[string]string{},
		Products: []Product{
			{ID: 1, Code: "P001", Name: "Salmón Atlántico", Category: "Congelados", Unit: "Kg", Warehouse: "Congelados", Stock: 25, MinStock: 10, Cost: 11250, MainSupplierID: 1, AlternateSupplierIDs: []int{2}, Active: true},
			{ID: 2, Code: "P002", Name: "Camarón 51/60", Category: "Congelados", Unit: "Kg", Warehouse: "Congelados", Stock: 12, MinStock: 5, Cost: 8900, MainSupplierID: 1, Active: true},
			{ID: 3, Code: "P003", Name: "Arroz G1", Category: "Abarrotes", Unit: "Kg", Warehouse: "Abarrotes y Aseo", Stock: 45, MinStock: 20, Cost: 1450, MainSupplierID: 2, AlternateSupplierIDs: []int{3}, Active: true},
			{ID: 4, Code: "P004", Name: "Aceite Vegetal 1L", Category: "Abarrotes", Unit: "Botella", Warehouse: "Abarrotes y Aseo", Stock: 8, MinStock: 5, Cost: 1900, MainSupplierID: 2, Active: true},
			{ID: 5, Code: "P005", Name: "Limón", Category: "Verduras", Unit: "Kg", Warehouse: "Cocina", Stock: 15, MinStock: 7, Cost: 1200, MainSupplierID: 3, Active: true},
			{ID: 6, Code: "P006", Name: "Guantes", Category: "Aseo", Unit: "Caja", Warehouse: "Cocina", Stock: 6, MinStock: 3, Cost: 4500, MainSupplierID: 2, Active: true},
		},
		Suppliers: []Supplier{{ID: 1, Name: "Proveedor Congelados", BusinessName: "Proveedor Congelados SpA", Category: "Congelados", Contact: "Ventas", OrderDays: []string{"Lunes", "Miércoles"}, OrderDeadline: "12:00", DeliveryDays: []string{"Martes", "Jueves"}, PaymentMethod: "Transferencia", Active: true}, {ID: 2, Name: "Distribuidora Abarrotes", BusinessName: "Distribuidora Abarrotes Ltda.", Category: "Abarrotes", Contact: "Ventas", OrderDays: []string{"Lunes", "Martes", "Miércoles", "Jueves"}, OrderDeadline: "12:00", DeliveryDays: []string{"Martes", "Viernes"}, PaymentMethod: "Transferencia", Active: true}, {ID: 3, Name: "Proveedor Verduras", BusinessName: "Proveedor Verduras", Category: "Verduras", Contact: "Ventas", OrderDays: []string{"Lunes", "Miércoles", "Viernes"}, OrderDeadline: "11:00", DeliveryDays: []string{"Martes", "Jueves", "Sábado"}, PaymentMethod: "Contado", Active: true}},
		Purchases: []PurchaseOrder{}, PriceHistory: []PriceHistory{}, Expenses: []Expense{}, Requests: []Request{}, Movements: []Movement{}, Audit: []Audit{}, NotificationReads: map[string][]string{},
	}
}

func newApp(path string) *App {
	a := &App{path: path, sessions: map[string]Session{}}
	b, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(b, &a.store)
		if len(a.store.Suppliers) == 0 {
			a.store.Suppliers = seed().Suppliers
		}
		for i := range a.store.Suppliers {
			if a.store.Suppliers[i].OrderDays == nil {
				a.store.Suppliers[i].OrderDays = []string{}
			}
			if a.store.Suppliers[i].DeliveryDays == nil {
				a.store.Suppliers[i].DeliveryDays = []string{}
			}
			if strings.TrimSpace(a.store.Suppliers[i].BusinessName) == "" {
				a.store.Suppliers[i].BusinessName = a.store.Suppliers[i].Name
			}
		}
		for i := range a.store.Products {
			// Migra nombres históricos de bodegas sin perder productos ni movimientos.
			warehouse := strings.ToLower(strings.TrimSpace(a.store.Products[i].Warehouse))
			switch warehouse {
			case "abarrotes", "abarrotes y líquidos", "abarrotes y liquidos":
				a.store.Products[i].Warehouse = "Abarrotes y Aseo"
			case "barra":
				a.store.Products[i].Warehouse = "Líquidos"
			}
			if strings.TrimSpace(a.store.Products[i].Category) == "" {
				a.store.Products[i].Category = a.store.Products[i].Warehouse
			}
			if a.store.Products[i].AlternateSupplierIDs == nil {
				a.store.Products[i].AlternateSupplierIDs = []int{}
			}
			ensureWarehouseStocks(&a.store.Products[i])
		}
		if a.store.Purchases == nil {
			a.store.Purchases = []PurchaseOrder{}
		}
		if a.store.Expenses == nil {
			a.store.Expenses = []Expense{}
		}
		for i := range a.store.Purchases {
			for j := range a.store.Purchases[i].Items {
				warehouse := strings.ToLower(strings.TrimSpace(a.store.Purchases[i].Items[j].Warehouse))
				switch warehouse {
				case "abarrotes", "abarrotes y líquidos", "abarrotes y liquidos":
					a.store.Purchases[i].Items[j].Warehouse = "Abarrotes y Aseo"
				case "barra":
					a.store.Purchases[i].Items[j].Warehouse = "Líquidos"
				}
			}
			for j := range a.store.Purchases[i].Receipts {
				for k := range a.store.Purchases[i].Receipts[j].Items {
					warehouse := strings.ToLower(strings.TrimSpace(a.store.Purchases[i].Receipts[j].Items[k].Warehouse))
					switch warehouse {
					case "abarrotes", "abarrotes y líquidos", "abarrotes y liquidos":
						a.store.Purchases[i].Receipts[j].Items[k].Warehouse = "Abarrotes y Aseo"
					case "barra":
						a.store.Purchases[i].Receipts[j].Items[k].Warehouse = "Líquidos"
					}
				}
			}
			if a.store.Purchases[i].Items == nil {
				a.store.Purchases[i].Items = []PurchaseItem{}
			}
			if a.store.Purchases[i].Receipts == nil {
				a.store.Purchases[i].Receipts = []PurchaseReceipt{}
			}
			if a.store.Purchases[i].Number == "" {
				a.store.Purchases[i].Number = fmt.Sprintf("OC-%06d", a.store.Purchases[i].ID)
			}
		}
		if a.store.PriceHistory == nil {
			a.store.PriceHistory = []PriceHistory{}
		}
		if a.store.NotificationReads == nil {
			a.store.NotificationReads = map[string][]string{}
		}
		if a.store.Accounts == nil || len(a.store.Accounts) == 0 {
			a.store.Accounts = defaultAccounts()
		} else {
			// Asegura los perfiles base sin reemplazar contraseñas ya personalizadas.
			for k, v := range defaultAccounts() {
				if _, ok := a.store.Accounts[k]; !ok {
					a.store.Accounts[k] = v
				}
			}
		}
		if a.store.SuggestionSnoozed == nil {
			a.store.SuggestionSnoozed = map[string]string{}
		}
		for i := range a.store.Movements {
			if strings.TrimSpace(a.store.Movements[i].Warehouse) == "" {
				if p := a.productLocked(a.store.Movements[i].ProductID); p != nil {
					a.store.Movements[i].Warehouse = p.Warehouse
				}
			} else {
				a.store.Movements[i].Warehouse = normalizeWarehouse(a.store.Movements[i].Warehouse)
			}
		}
		_ = a.saveLocked()
	} else {
		a.store = seed()
		_ = a.saveLocked()
	}
	return a
}
func (a *App) saveLocked() error {
	b, err := json.MarshalIndent(a.store, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.path), 0755); err != nil {
		return err
	}
	if old, err := os.ReadFile(a.path); err == nil {
		if err := os.WriteFile(a.path+".bak", old, 0644); err != nil {
			return err
		}
		backupDir := filepath.Join(filepath.Dir(a.path), "respaldos")
		if err := os.MkdirAll(backupDir, 0755); err != nil {
			return err
		}
		name := "data-" + time.Now().Format("20060102-150405.000000000") + ".json"
		if err := os.WriteFile(filepath.Join(backupDir, name), old, 0644); err != nil {
			return err
		}
		// Mantiene los respaldos automáticos recientes para evitar que miles de archivos ralenticen Windows.
		entries, _ := os.ReadDir(backupDir)
		auto := []os.DirEntry{}
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), "data-") {
				auto = append(auto, e)
			}
		}
		if len(auto) > 120 {
			sort.Slice(auto, func(i, j int) bool { return auto[i].Name() < auto[j].Name() })
			for _, e := range auto[:len(auto)-120] {
				_ = os.Remove(filepath.Join(backupDir, e.Name()))
			}
		}
	}
	return os.WriteFile(a.path, b, 0644)
}
func (a *App) saveFastLocked() error {
	b, err := json.MarshalIndent(a.store, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.path), 0755); err != nil {
		return err
	}
	tmp := a.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, a.path)
}
func token() string { b := make([]byte, 24); _, _ = rand.Read(b); return hex.EncodeToString(b) }
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func readJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
func (a *App) auth(r *http.Request) (User, bool) {
	t := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if t == "" {
		return User{}, false
	}
	a.mu.RLock()
	s, ok := a.sessions[t]
	a.mu.RUnlock()
	if !ok || time.Now().After(s.ExpiresAt) {
		return User{}, false
	}
	// La sesión dura 8 horas. No se escribe el mapa en cada petición: evita
	// serializar/bloquear todas las llamadas de refresco de la interfaz.
	return s.User, true
}
func require(a *App, roles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			u, ok := a.auth(r)
			if !ok {
				writeJSON(w, 401, map[string]string{"error": "Sesión inválida"})
				return
			}
			if len(roles) > 0 {
				good := false
				for _, x := range roles {
					if u.Role == x {
						good = true
					}
				}
				if !good {
					writeJSON(w, 403, map[string]string{"error": "Sin permiso"})
					return
				}
			}
			r.Header.Set("X-User", u.Username)
			next(w, r)
		}
	}
}
func (a *App) current(r *http.Request) User {
	a.mu.RLock()
	defer a.mu.RUnlock()
	acc, ok := a.store.Accounts[r.Header.Get("X-User")]
	if !ok {
		return User{}
	}
	return accountUser(acc)
}
func (a *App) auditLocked(user, action string) {
	a.store.Audit = append(a.store.Audit, Audit{len(a.store.Audit) + 1, user, action, now()})
}
func (a *App) productLocked(id int) *Product {
	for i := range a.store.Products {
		if a.store.Products[i].ID == id {
			return &a.store.Products[i]
		}
	}
	return nil
}

var warehouseNames = []string{"Congelados", "Abarrotes y Aseo", "Líquidos", "Envases", "Cocina", "Barra y Caja", "Mantención"}

func normalizeWarehouse(v string) string {
	v = strings.TrimSpace(v)
	switch strings.ToLower(v) {
	case "abarrotes", "abarrotes y líquidos", "abarrotes y liquidos":
		return "Abarrotes y Aseo"
	case "barra", "barra/caja", "caja", "barra y caja":
		return "Barra y Caja"
	}
	return v
}

func ensureWarehouseStocks(p *Product) {
	if p.WarehouseStocks == nil {
		p.WarehouseStocks = map[string]float64{}
	}
	p.Warehouse = normalizeWarehouse(p.Warehouse)
	if _, ok := p.WarehouseStocks[p.Warehouse]; !ok {
		p.WarehouseStocks[p.Warehouse] = p.Stock
	} else {
		p.Stock = p.WarehouseStocks[p.Warehouse]
	}
}

func stockAt(p *Product, warehouse string) float64 {
	ensureWarehouseStocks(p)
	warehouse = normalizeWarehouse(warehouse)
	return p.WarehouseStocks[warehouse]
}

func setStockAt(p *Product, warehouse string, value float64) {
	ensureWarehouseStocks(p)
	warehouse = normalizeWarehouse(warehouse)
	if value < 0 {
		value = 0
	}
	p.WarehouseStocks[warehouse] = value
	if warehouse == p.Warehouse {
		p.Stock = value
	}
}

func addStockAt(p *Product, warehouse string, delta float64) float64 {
	value := stockAt(p, warehouse) + delta
	setStockAt(p, warehouse, value)
	return stockAt(p, warehouse)
}

func totalStock(p *Product) float64 {
	ensureWarehouseStocks(p)
	total := 0.0
	for _, v := range p.WarehouseStocks {
		total += v
	}
	return total
}

func requestDestination(area string) string {
	switch strings.ToLower(strings.TrimSpace(area)) {
	case "cocina":
		return "Cocina"
	case "caja", "barra", "barra y caja":
		return "Barra y Caja"
	}
	return ""
}

func (a *App) purchaseLocked(id int) *PurchaseOrder {
	for i := range a.store.Purchases {
		if a.store.Purchases[i].ID == id {
			return &a.store.Purchases[i]
		}
	}
	return nil
}
func purchaseNumber(id int) string { return fmt.Sprintf("OC-%06d", id) }
func (a *App) nextPurchaseIDLocked() int {
	maxID := 0
	for _, po := range a.store.Purchases {
		if po.ID > maxID {
			maxID = po.ID
		}
	}
	return maxID + 1
}
func (a *App) nextExpenseIDLocked() int {
	maxID := 0
	for _, e := range a.store.Expenses {
		if e.ID > maxID {
			maxID = e.ID
		}
	}
	return maxID + 1
}

func purchaseTotal(items []PurchaseItem) float64 {
	total := 0.0
	for _, it := range items {
		total += it.Quantity * it.UnitPrice
	}
	return total
}

func (a *App) productHasHistoryLocked(id int) bool {
	for _, m := range a.store.Movements {
		if m.ProductID == id {
			return true
		}
	}
	for _, q := range a.store.Requests {
		for _, it := range q.Items {
			if it.ProductID == id {
				return true
			}
		}
	}
	return false
}
func (a *App) supplierLocked(id int) *Supplier {
	for i := range a.store.Suppliers {
		if a.store.Suppliers[i].ID == id {
			return &a.store.Suppliers[i]
		}
	}
	return nil
}
func cleanSupplier(s Supplier) Supplier {
	s.Name = strings.TrimSpace(s.Name)
	s.BusinessName = strings.TrimSpace(s.BusinessName)
	s.RUT = strings.TrimSpace(s.RUT)
	s.Category = strings.TrimSpace(s.Category)
	s.Contact = strings.TrimSpace(s.Contact)
	s.Phone = strings.TrimSpace(s.Phone)
	s.Email = strings.TrimSpace(s.Email)
	s.OrderDeadline = strings.TrimSpace(s.OrderDeadline)
	s.PaymentMethod = strings.TrimSpace(s.PaymentMethod)
	s.PaymentTerms = strings.TrimSpace(s.PaymentTerms)
	s.Notes = strings.TrimSpace(s.Notes)
	s.LastPurchase = strings.TrimSpace(s.LastPurchase)
	s.LastDispatch = strings.TrimSpace(s.LastDispatch)
	if s.BusinessName == "" {
		s.BusinessName = s.Name
	}
	if s.OrderDays == nil {
		s.OrderDays = []string{}
	}
	if s.DeliveryDays == nil {
		s.DeliveryDays = []string{}
	}
	return s
}

func cleanProduct(p Product) Product {
	p.Code = strings.TrimSpace(p.Code)
	p.Name = strings.TrimSpace(p.Name)
	p.Category = strings.TrimSpace(p.Category)
	p.Unit = strings.TrimSpace(p.Unit)
	p.Warehouse = strings.TrimSpace(p.Warehouse)
	p.Notes = strings.TrimSpace(p.Notes)
	if p.AlternateSupplierIDs == nil {
		p.AlternateSupplierIDs = []int{}
	}
	return p
}

func appDataDir() string {
	if runtime.GOOS == "windows" {
		if base := os.Getenv("LOCALAPPDATA"); base != "" {
			return filepath.Join(base, "EBGestion")
		}
	}
	if base, err := os.UserConfigDir(); err == nil {
		return filepath.Join(base, "EBGestion")
	}
	return filepath.Join(".", "EBGestion-Datos")
}
func migrateLegacyData(exeDir, dataDir string) {
	_ = os.MkdirAll(dataDir, 0755)
	dest := filepath.Join(dataDir, "data.json")
	if _, err := os.Stat(dest); err == nil {
		return
	}
	legacy := filepath.Join(exeDir, "data.json")
	if b, err := os.ReadFile(legacy); err == nil {
		_ = os.WriteFile(dest, b, 0644)
	}
}

type BackupInfo struct {
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	Size      int64  `json:"size"`
}

func (a *App) backupDir() string { return filepath.Join(filepath.Dir(a.path), "respaldos") }
func (a *App) createBackupLocked(prefix string) (BackupInfo, error) {
	b, err := json.MarshalIndent(a.store, "", "  ")
	if err != nil {
		return BackupInfo{}, err
	}
	if err := os.MkdirAll(a.backupDir(), 0755); err != nil {
		return BackupInfo{}, err
	}
	name := prefix + "-" + time.Now().Format("20060102-150405") + ".json"
	path := filepath.Join(a.backupDir(), name)
	if err := os.WriteFile(path, b, 0644); err != nil {
		return BackupInfo{}, err
	}
	st, _ := os.Stat(path)
	return BackupInfo{Name: name, CreatedAt: time.Now().Format("2006-01-02 15:04:05"), Size: st.Size()}, nil
}
func main() {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	dataDir := appDataDir()
	migrateLegacyData(dir, dataDir)
	app := newApp(filepath.Join(dataDir, "data.json"))
	mux := http.NewServeMux()
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		var in struct{ Username, Password string }
		if readJSON(r, &in) != nil {
			writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
			return
		}
		username := strings.ToLower(strings.TrimSpace(in.Username))
		app.mu.RLock()
		acc, ok := app.store.Accounts[username]
		app.mu.RUnlock()
		if !ok || !acc.Active || acc.PasswordHash != passwordHash(in.Password) {
			writeJSON(w, 401, map[string]string{"error": "Usuario o contraseña incorrectos"})
			return
		}
		u := accountUser(acc)
		t := token()
		app.mu.Lock()
		app.sessions[t] = Session{User: u, ExpiresAt: time.Now().Add(8 * time.Hour)}
		app.auditLocked(u.Name, "Inicio de sesión")
		_ = app.saveFastLocked()
		app.mu.Unlock()
		writeJSON(w, 200, map[string]any{"token": t, "username": u.Username, "name": u.Name, "role": u.Role})
	})
	mux.HandleFunc("/api/logout", require(app)(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		t := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		u := app.current(r)
		app.mu.Lock()
		delete(app.sessions, t)
		app.auditLocked(u.Name, "Cierre de sesión")
		err := app.saveFastLocked()
		app.mu.Unlock()
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "No fue posible registrar el cierre de sesión"})
			return
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/me", require(app)(func(w http.ResponseWriter, r *http.Request) {
		u := app.current(r)
		writeJSON(w, 200, map[string]string{"username": u.Username, "name": u.Name, "role": u.Role})
	}))
	mux.HandleFunc("/api/system", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		app.mu.RLock()
		snoozed := app.store.SuggestionSnoozed
		app.mu.RUnlock()
		writeJSON(w, 200, map[string]any{"dataPath": app.path, "backupPath": app.backupDir(), "version": "0.9.6.1 Barra y Caja - Corrección", "lanIP": lanIP(), "suggestionSnoozed": snoozed})
	}))
	mux.HandleFunc("/api/backups", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			_ = os.MkdirAll(app.backupDir(), 0755)
			entries, _ := os.ReadDir(app.backupDir())
			rows := []BackupInfo{}
			for i := len(entries) - 1; i >= 0; i-- {
				e := entries[i]
				if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".json") {
					continue
				}
				st, err := e.Info()
				if err != nil {
					continue
				}
				rows = append(rows, BackupInfo{Name: e.Name(), CreatedAt: st.ModTime().Format("2006-01-02 15:04:05"), Size: st.Size()})
			}
			writeJSON(w, 200, rows)
		case "POST":
			u := app.current(r)
			app.mu.Lock()
			info, err := app.createBackupLocked("manual")
			if err == nil {
				app.auditLocked(u.Name, "Creó respaldo manual "+info.Name)
				err = app.saveLocked()
			}
			app.mu.Unlock()
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": "No fue posible crear el respaldo"})
				return
			}
			writeJSON(w, 201, info)
		default:
			http.NotFound(w, r)
		}
	}))
	mux.HandleFunc("/api/backups/restore", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		var in struct {
			Name string `json:"name"`
		}
		if readJSON(r, &in) != nil || filepath.Base(in.Name) != in.Name {
			writeJSON(w, 400, map[string]string{"error": "Respaldo inválido"})
			return
		}
		b, err := os.ReadFile(filepath.Join(app.backupDir(), in.Name))
		if err != nil {
			writeJSON(w, 404, map[string]string{"error": "Respaldo no encontrado"})
			return
		}
		var restored Store
		if json.Unmarshal(b, &restored) != nil || restored.Products == nil {
			writeJSON(w, 400, map[string]string{"error": "El archivo de respaldo está dañado"})
			return
		}
		u := app.current(r)
		app.mu.Lock()
		_, _ = app.createBackupLocked("antes-restaurar")
		app.store = restored
		app.auditLocked(u.Name, "Restauró respaldo "+in.Name)
		err = app.saveLocked()
		app.mu.Unlock()
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "No fue posible restaurar el respaldo"})
			return
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/suppliers", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		u := app.current(r)
		switch r.Method {
		case "GET":
			app.mu.RLock()
			defer app.mu.RUnlock()
			writeJSON(w, 200, app.store.Suppliers)
		case "POST":
			if u.Role != "admin_bodega" && u.Role != "gerencia" {
				writeJSON(w, 403, map[string]string{"error": "Sin permiso"})
				return
			}
			var in Supplier
			if readJSON(r, &in) != nil {
				writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
				return
			}
			in = cleanSupplier(in)
			if in.Name == "" {
				writeJSON(w, 400, map[string]string{"error": "Ingresa el nombre de fantasía del proveedor"})
				return
			}
			if in.MinimumOrder < 0 {
				writeJSON(w, 400, map[string]string{"error": "La compra mínima no puede ser negativa"})
				return
			}
			app.mu.Lock()
			defer app.mu.Unlock()
			for _, x := range app.store.Suppliers {
				if strings.EqualFold(x.Name, in.Name) {
					writeJSON(w, 409, map[string]string{"error": "Ya existe un proveedor con ese nombre"})
					return
				}
			}
			maxID := 0
			for _, x := range app.store.Suppliers {
				if x.ID > maxID {
					maxID = x.ID
				}
			}
			in.ID = maxID + 1
			in.Active = true
			app.store.Suppliers = append(app.store.Suppliers, in)
			app.auditLocked(u.Name, "Creó proveedor "+in.Name)
			_ = app.saveLocked()
			writeJSON(w, 201, in)
		default:
			http.NotFound(w, r)
		}
	}))
	mux.HandleFunc("/api/suppliers/", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		id, _ := strconv.Atoi(parts[2])
		u := app.current(r)
		app.mu.Lock()
		defer app.mu.Unlock()
		sup := app.supplierLocked(id)
		if sup == nil {
			writeJSON(w, 404, map[string]string{"error": "Proveedor no encontrado"})
			return
		}
		if len(parts) == 4 && parts[3] == "toggle" && r.Method == "POST" {
			sup.Active = !sup.Active
			app.auditLocked(u.Name, fmt.Sprintf("Cambió estado de proveedor %s a %t", sup.Name, sup.Active))
			_ = app.saveLocked()
			writeJSON(w, 200, sup)
			return
		}
		if r.Method == "PUT" {
			var in Supplier
			if readJSON(r, &in) != nil {
				writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
				return
			}
			in = cleanSupplier(in)
			if in.Name == "" {
				writeJSON(w, 400, map[string]string{"error": "Ingresa el nombre de fantasía del proveedor"})
				return
			}
			if in.MinimumOrder < 0 {
				writeJSON(w, 400, map[string]string{"error": "La compra mínima no puede ser negativa"})
				return
			}
			for _, x := range app.store.Suppliers {
				if x.ID != id && strings.EqualFold(x.Name, in.Name) {
					writeJSON(w, 409, map[string]string{"error": "Ya existe otro proveedor con ese nombre"})
					return
				}
			}
			in.ID = id
			in.Active = sup.Active
			*sup = in
			app.auditLocked(u.Name, "Editó proveedor "+in.Name)
			_ = app.saveLocked()
			writeJSON(w, 200, sup)
			return
		}
		http.NotFound(w, r)
	}))
	mux.HandleFunc("/api/products", require(app)(func(w http.ResponseWriter, r *http.Request) {
		u := app.current(r)
		switch r.Method {
		case "GET":
			app.mu.RLock()
			defer app.mu.RUnlock()
			rows := append([]Product(nil), app.store.Products...)
			if u.Role == "cocina" || u.Role == "caja" || u.Role == "bodega" {
				for i := range rows {
					rows[i].Cost = 0
					rows[i].MainSupplierID = 0
					rows[i].AlternateSupplierIDs = []int{}
				}
			}
			writeJSON(w, 200, rows)
		case "POST":
			if u.Role != "admin_bodega" && u.Role != "gerencia" {
				writeJSON(w, 403, map[string]string{"error": "Sin permiso"})
				return
			}
			var in Product
			if readJSON(r, &in) != nil {
				writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
				return
			}
			in = cleanProduct(in)
			if in.Name == "" || in.Code == "" || in.Unit == "" || in.Warehouse == "" || in.Category == "" {
				writeJSON(w, 400, map[string]string{"error": "Completa código, nombre, categoría, unidad y bodega"})
				return
			}
			if in.Stock < 0 || in.MinStock < 0 || in.Cost < 0 {
				writeJSON(w, 400, map[string]string{"error": "Stock, mínimo y costo no pueden ser negativos"})
				return
			}
			app.mu.Lock()
			defer app.mu.Unlock()
			for _, p := range app.store.Products {
				if strings.EqualFold(p.Name, in.Name) || strings.EqualFold(p.Code, in.Code) {
					writeJSON(w, 409, map[string]string{"error": "Ya existe un producto con ese nombre o código"})
					return
				}
			}
			maxID := 0
			for _, p := range app.store.Products {
				if p.ID > maxID {
					maxID = p.ID
				}
			}
			in.ID = maxID + 1
			in.Active = true
			in.Warehouse = normalizeWarehouse(in.Warehouse)
			in.WarehouseStocks = map[string]float64{in.Warehouse: in.Stock}
			app.store.Products = append(app.store.Products, in)
			app.auditLocked(u.Name, "Creó producto "+in.Name)
			_ = app.saveLocked()
			writeJSON(w, 201, in)
		default:
			http.NotFound(w, r)
		}
	}))
	mux.HandleFunc("/api/products/", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		id, _ := strconv.Atoi(parts[2])
		u := app.current(r)
		app.mu.Lock()
		defer app.mu.Unlock()
		p := app.productLocked(id)
		if p == nil {
			writeJSON(w, 404, map[string]string{"error": "Producto no encontrado"})
			return
		}
		if len(parts) == 4 && parts[3] == "toggle" && r.Method == "POST" {
			p.Active = !p.Active
			app.auditLocked(u.Name, fmt.Sprintf("Cambió estado de %s a %t", p.Name, p.Active))
			_ = app.saveLocked()
			writeJSON(w, 200, p)
			return
		}
		if r.Method == "PUT" {
			var in Product
			if readJSON(r, &in) != nil {
				writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
				return
			}
			in = cleanProduct(in)
			if in.Name == "" || in.Code == "" || in.Unit == "" || in.Warehouse == "" || in.Category == "" {
				writeJSON(w, 400, map[string]string{"error": "Completa código, nombre, categoría, unidad y bodega"})
				return
			}
			for _, x := range app.store.Products {
				if x.ID != id && (strings.EqualFold(x.Name, in.Name) || strings.EqualFold(x.Code, in.Code)) {
					writeJSON(w, 409, map[string]string{"error": "Ya existe otro producto con ese nombre o código"})
					return
				}
			}
			in.ID = id
			in.Active = p.Active
			if app.productHasHistoryLocked(id) {
				// Con historial, la ubicación principal y los stocks se modifican por movimientos, no editando la ficha.
				in.Warehouse = p.Warehouse
				in.Stock = p.Stock
				in.WarehouseStocks = p.WarehouseStocks
			} else {
				in.Warehouse = normalizeWarehouse(in.Warehouse)
				in.WarehouseStocks = map[string]float64{in.Warehouse: in.Stock}
			}
			*p = in
			app.auditLocked(u.Name, "Editó producto "+in.Name)
			_ = app.saveLocked()
			writeJSON(w, 200, p)
			return
		}
		http.NotFound(w, r)
	}))
	mux.HandleFunc("/api/dashboard", require(app)(func(w http.ResponseWriter, r *http.Request) {
		app.mu.RLock()
		defer app.mu.RUnlock()
		var value float64
		critical := 0
		for _, p := range app.store.Products {
			if p.Active {
				value += totalStock(&p) * p.Cost
				if p.Stock <= p.MinStock {
					critical++
				}
			}
		}
		pending := 0
		for _, q := range app.store.Requests {
			if q.Status == "pendiente" || q.Status == "preparando" {
				pending++
			}
		}
		u := app.current(r)
		result := map[string]any{"products": len(app.store.Products), "critical": critical, "pendingRequests": pending}
		if u.Role == "admin_bodega" || u.Role == "gerencia" {
			result["inventoryValue"] = value
		}
		writeJSON(w, 200, result)
	}))
	mux.HandleFunc("/api/notifications", require(app)(func(w http.ResponseWriter, r *http.Request) {
		u := app.current(r)
		type notice struct {
			ID       string `json:"id"`
			Icon     string `json:"icon"`
			Text     string `json:"text"`
			Time     string `json:"time"`
			Type     string `json:"type"`
			TargetID int    `json:"targetId"`
			Priority int    `json:"priority"`
		}
		build := func() []notice {
			notes := []notice{}
			if u.Role == "admin_bodega" || u.Role == "gerencia" {
				for _, q := range app.store.Requests {
					if q.Status == "pendiente" {
						notes = append(notes, notice{fmt.Sprintf("request:%d", q.ID), "📋", "Nueva solicitud de " + q.Requester, q.CreatedAt, "request", q.ID, 20})
					}
				}
				for _, p := range app.store.Products {
					if p.Active && p.Stock <= p.MinStock {
						notes = append(notes, notice{fmt.Sprintf("stock:%d:%.3f", p.ID, p.Stock), "⚠️", p.Name + " con stock crítico", "Ahora", "stock", p.ID, 80})
					}
				}
				dayNames := map[time.Weekday]string{time.Monday: "Lunes", time.Tuesday: "Martes", time.Wednesday: "Miércoles", time.Thursday: "Jueves", time.Friday: "Viernes", time.Saturday: "Sábado", time.Sunday: "Domingo"}
				today := dayNames[time.Now().Weekday()]
				dateKey := time.Now().Format("2006-01-02")
				for _, sup := range app.store.Suppliers {
					if !sup.Active {
						continue
					}
					for _, d := range sup.OrderDays {
						if d == today {
							msg := "Hoy corresponde pedido a " + sup.Name
							if sup.OrderDeadline != "" {
								msg += " · hasta " + sup.OrderDeadline
							}
							notes = append(notes, notice{fmt.Sprintf("supplier:%d:%s", sup.ID, dateKey), "🛒", msg, "Hoy", "supplier", sup.ID, 90})
							break
						}
					}
				}
				for _, po := range app.store.Purchases {
					if po.Status == "solicitada" || po.Status == "parcial" {
						notes = append(notes, notice{fmt.Sprintf("purchase:%d:%s", po.ID, po.Status), "📦", po.Number + " pendiente de recepción", po.ExpectedDate, "purchase", po.ID, 70})
					}
				}
				todayDate := time.Now().Truncate(24 * time.Hour)
				for _, e := range app.store.Expenses {
					if e.Status != "pendiente" || strings.TrimSpace(e.DueDate) == "" {
						continue
					}
					due, err := time.Parse("2006-01-02", e.DueDate)
					if err != nil {
						continue
					}
					days := int(due.Sub(todayDate).Hours() / 24)
					if days <= 7 {
						label := "vence en " + strconv.Itoa(days) + " días"
						if days < 0 {
							label = "vencida"
						} else if days == 0 {
							label = "vence hoy"
						}
						notes = append(notes, notice{fmt.Sprintf("expense:%d:%s", e.ID, e.DueDate), "💰", e.SupplierName + " · $" + fmt.Sprintf("%.0f", e.Amount) + " · " + label, e.DueDate, "expense", e.ID, 10})
					}
				}
			}
			sort.SliceStable(notes, func(i, j int) bool {
				if notes[i].Priority == notes[j].Priority {
					return notes[i].Time < notes[j].Time
				}
				return notes[i].Priority < notes[j].Priority
			})
			return notes
		}
		switch r.Method {
		case "GET":
			app.mu.RLock()
			defer app.mu.RUnlock()
			readSet := map[string]bool{}
			for _, id := range app.store.NotificationReads[u.Username] {
				readSet[id] = true
			}
			visible := []notice{}
			for _, n := range build() {
				if !readSet[n.ID] {
					visible = append(visible, n)
				}
			}
			writeJSON(w, 200, visible)
		case "POST":
			app.mu.Lock()
			defer app.mu.Unlock()
			var in struct {
				IDs []string `json:"ids"`
			}
			_ = readJSON(r, &in)
			if len(in.IDs) == 0 {
				for _, n := range build() {
					in.IDs = append(in.IDs, n.ID)
				}
			}
			seen := map[string]bool{}
			merged := []string{}
			for _, id := range app.store.NotificationReads[u.Username] {
				if !seen[id] {
					seen[id] = true
					merged = append(merged, id)
				}
			}
			for _, id := range in.IDs {
				if !seen[id] {
					seen[id] = true
					merged = append(merged, id)
				}
			}
			if len(merged) > 500 {
				merged = merged[len(merged)-500:]
			}
			app.store.NotificationReads[u.Username] = merged
			if err := app.saveFastLocked(); err != nil {
				writeJSON(w, 500, map[string]string{"error": "No fue posible guardar las notificaciones: " + err.Error()})
				return
			}
			writeJSON(w, 200, map[string]bool{"ok": true})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	mux.HandleFunc("/api/requests", require(app)(func(w http.ResponseWriter, r *http.Request) {
		u := app.current(r)
		switch r.Method {
		case "GET":
			app.mu.RLock()
			rows := []Request{}
			for _, q := range app.store.Requests {
				if u.Role == "cocina" || u.Role == "caja" || u.Role == "bodega" {
					if q.Requester == u.Name {
						rows = append(rows, q)
					}
				} else {
					rows = append(rows, q)
				}
			}
			for i := range rows {
				for j := range rows[i].Items {
					if p := app.productLocked(rows[i].Items[j].ProductID); p != nil {
						rows[i].Items[j].Stock = stockAt(p, p.Warehouse)
					}
				}
			}
			app.mu.RUnlock()
			writeJSON(w, 200, rows)
		case "POST":
			var in struct {
				Items []struct {
					ProductID int     `json:"productId"`
					Qty       float64 `json:"qty"`
				} `json:"items"`
				Note string `json:"note"`
			}
			if readJSON(r, &in) != nil || len(in.Items) == 0 {
				writeJSON(w, 400, map[string]string{"error": "Agrega al menos un producto"})
				return
			}
			app.mu.Lock()
			id := len(app.store.Requests) + 1
			items := []RequestItem{}
			for i, x := range in.Items {
				p := app.productLocked(x.ProductID)
				if p != nil && x.Qty > 0 {
					items = append(items, RequestItem{i + 1, p.ID, p.Name, p.Unit, "", p.Stock, x.Qty, nil})
				}
			}
			if len(items) == 0 {
				app.mu.Unlock()
				writeJSON(w, 400, map[string]string{"error": "No hay productos válidos"})
				return
			}
			area := "Bodega"
			switch u.Role {
			case "cocina":
				area = "Cocina"
			case "caja":
				area = "Caja"
			case "admin_bodega":
				area = "Administración"
			case "gerencia":
				area = "Gerencia"
			}
			q := Request{ID: id, Requester: u.Name, Area: area, Status: "pendiente", CreatedAt: now(), UpdatedAt: now(), Note: in.Note, Items: items}
			app.store.Requests = append(app.store.Requests, q)
			app.auditLocked(u.Name, fmt.Sprintf("Creó solicitud #%d", id))
			err := app.saveLocked()
			app.mu.Unlock()
			if err != nil {
				writeJSON(w, 500, map[string]string{"error": "No fue posible guardar la solicitud: " + err.Error()})
				return
			}
			writeJSON(w, 201, map[string]any{"ok": true, "id": id})
		default:
			http.NotFound(w, r)
		}
	}))
	mux.HandleFunc("/api/requests/", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) != 4 {
			http.NotFound(w, r)
			return
		}
		id, _ := strconv.Atoi(parts[2])
		action := parts[3]
		u := app.current(r)
		app.mu.Lock()
		defer app.mu.Unlock()
		var q *Request
		for i := range app.store.Requests {
			if app.store.Requests[i].ID == id {
				q = &app.store.Requests[i]
				break
			}
		}
		if q == nil {
			writeJSON(w, 404, map[string]string{"error": "Solicitud no encontrada"})
			return
		}
		if action == "prepare" && r.Method == "POST" {
			q.Status = "preparando"
			q.UpdatedAt = now()
			app.auditLocked(u.Name, fmt.Sprintf("Preparó solicitud #%d", id))
			_ = app.saveLocked()
			writeJSON(w, 200, map[string]bool{"ok": true})
			return
		}
		if action == "deliver" && r.Method == "POST" {
			if q.Status != "pendiente" && q.Status != "preparando" {
				writeJSON(w, 200, map[string]any{"ok": true, "status": q.Status, "alreadyProcessed": true})
				return
			}
			var in struct {
				DestinationWarehouse string `json:"destinationWarehouse"`
				Items                []struct {
					ItemID       int     `json:"itemId"`
					DeliveredQty float64 `json:"deliveredQty"`
					Reason       string  `json:"reason"`
				} `json:"items"`
			}
			if readJSON(r, &in) != nil {
				writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
				return
			}
			destination := requestDestination(q.Area)
			if destination == "" {
				destination = normalizeWarehouse(in.DestinationWarehouse)
			}
			anyRequestedDelivery := false
			for _, x := range in.Items {
				if x.DeliveredQty > 0 {
					anyRequestedDelivery = true
					break
				}
			}
			if anyRequestedDelivery && strings.TrimSpace(destination) == "" {
				writeJSON(w, 400, map[string]string{"error": "Selecciona la bodega destino"})
				return
			}

			// Primero valida todo; recién después modifica inventario, para evitar movimientos parciales.
			type deliveryPlan struct {
				item           *RequestItem
				product        *Product
				qty            float64
				reason, source string
			}
			plans := []deliveryPlan{}
			complete := true
			anyDelivered := false
			for _, x := range in.Items {
				var item *RequestItem
				for j := range q.Items {
					if q.Items[j].ID == x.ItemID {
						item = &q.Items[j]
						break
					}
				}
				if item == nil {
					continue
				}
				p := app.productLocked(item.ProductID)
				if p == nil {
					continue
				}
				source := p.Warehouse
				available := stockAt(p, source)
				qty := x.DeliveredQty
				if qty < 0 {
					qty = 0
				}
				if qty > available {
					qty = available
				}
				if qty < item.RequestedQty && strings.TrimSpace(x.Reason) == "" {
					writeJSON(w, 400, map[string]string{"error": "Indica el motivo cuando entregas menos"})
					return
				}
				if qty > 0 && normalizeWarehouse(destination) == normalizeWarehouse(source) {
					writeJSON(w, 400, map[string]string{"error": "La bodega destino debe ser distinta de la bodega origen (" + source + ")"})
					return
				}
				plans = append(plans, deliveryPlan{item: item, product: p, qty: qty, reason: x.Reason, source: source})
				if qty > 0 {
					anyDelivered = true
				}
				if qty < item.RequestedQty {
					complete = false
				}
			}
			for _, plan := range plans {
				qty := plan.qty
				item := plan.item
				p := plan.product
				item.DeliveredQty = &qty
				item.Reason = plan.reason
				if qty <= 0 {
					continue
				}
				sourceBalance := addStockAt(p, plan.source, -qty)
				destBalance := addStockAt(p, destination, qty)
				reason := fmt.Sprintf("Traspaso Solicitud #%d", id)
				app.store.Movements = append(app.store.Movements, Movement{ID: len(app.store.Movements) + 1, ProductID: p.ID, Warehouse: plan.source, Type: "salida", Quantity: -qty, Balance: sourceBalance, Reason: reason + " → " + destination, User: u.Name, CreatedAt: now()})
				app.store.Movements = append(app.store.Movements, Movement{ID: len(app.store.Movements) + 1, ProductID: p.ID, Warehouse: destination, Type: "entrada", Quantity: qty, Balance: destBalance, Reason: reason + " ← " + plan.source, User: u.Name, CreatedAt: now()})
			}
			if complete {
				q.Status = "entregada"
			} else if anyDelivered {
				q.Status = "parcial"
			} else {
				q.Status = "no_entregada"
			}
			q.DeliveredBy = u.Name
			q.DestinationWarehouse = destination
			q.UpdatedAt = now()
			app.auditLocked(u.Name, fmt.Sprintf("Finalizó solicitud #%d (%s) destino %s", id, q.Status, destination))
			if err := app.saveLocked(); err != nil {
				writeJSON(w, 500, map[string]string{"error": "No fue posible guardar la entrega: " + err.Error()})
				return
			}
			writeJSON(w, 200, map[string]any{"ok": true, "status": q.Status, "destinationWarehouse": destination})
			return
		}

		http.NotFound(w, r)
	}))

	mux.HandleFunc("/api/purchase-suggestions/snooze", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		var in struct {
			ProductIDs []int `json:"productIds"`
		}
		if readJSON(r, &in) != nil {
			writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
			return
		}
		app.mu.Lock()
		defer app.mu.Unlock()
		today := time.Now().Format("2006-01-02")
		for _, id := range in.ProductIDs {
			app.store.SuggestionSnoozed[strconv.Itoa(id)] = today
		}
		_ = app.saveLocked()
		writeJSON(w, 200, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/api/users", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			app.mu.RLock()
			rows := []map[string]any{}
			for _, a := range app.store.Accounts {
				rows = append(rows, map[string]any{"username": a.Username, "name": a.Name, "role": a.Role, "active": a.Active})
			}
			app.mu.RUnlock()
			sort.Slice(rows, func(i, j int) bool { return rows[i]["role"].(string) < rows[j]["role"].(string) })
			writeJSON(w, 200, rows)
		case "PUT":
			var in struct {
				Username, Name, Role, Password string
				Active                         *bool `json:"active"`
			}
			if readJSON(r, &in) != nil {
				writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
				return
			}
			key := strings.ToLower(strings.TrimSpace(in.Username))
			app.mu.Lock()
			defer app.mu.Unlock()
			a, ok := app.store.Accounts[key]
			if !ok {
				writeJSON(w, 404, map[string]string{"error": "Usuario no encontrado"})
				return
			}
			if strings.TrimSpace(in.Name) != "" {
				a.Name = strings.TrimSpace(in.Name)
			}
			if strings.TrimSpace(in.Role) != "" {
				a.Role = strings.TrimSpace(in.Role)
			}
			if strings.TrimSpace(in.Password) != "" {
				if len(in.Password) < 6 {
					writeJSON(w, 400, map[string]string{"error": "La contraseña debe tener al menos 6 caracteres"})
					return
				}
				a.PasswordHash = passwordHash(in.Password)
			}
			if in.Active != nil {
				a.Active = *in.Active
			}
			app.store.Accounts[key] = a
			app.auditLocked(r.Header.Get("X-User"), "Actualizó usuario "+key)
			if err := app.saveLocked(); err != nil {
				writeJSON(w, 500, map[string]string{"error": "No fue posible guardar el usuario: " + err.Error()})
				return
			}
			writeJSON(w, 200, map[string]bool{"ok": true})
		default:
			http.NotFound(w, r)
		}
	}))
	mux.HandleFunc("/api/purchases", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			app.mu.RLock()
			defer app.mu.RUnlock()
			rows := make([]PurchaseOrder, len(app.store.Purchases))
			for i := range app.store.Purchases {
				rows[len(rows)-1-i] = app.store.Purchases[i]
			}
			writeJSON(w, 200, rows)
		case "POST":
			var in struct {
				SupplierID   int    `json:"supplierId"`
				ExpectedDate string `json:"expectedDate"`
				Note         string `json:"note"`
				Items        []struct {
					ProductID int     `json:"productId"`
					Quantity  float64 `json:"quantity"`
					UnitPrice float64 `json:"unitPrice"`
					Note      string  `json:"note"`
				} `json:"items"`
			}
			if readJSON(r, &in) != nil || in.SupplierID <= 0 || len(in.Items) == 0 {
				writeJSON(w, 400, map[string]string{"error": "Selecciona proveedor y agrega productos"})
				return
			}
			u := app.current(r)
			app.mu.Lock()
			defer app.mu.Unlock()
			sup := app.supplierLocked(in.SupplierID)
			if sup == nil || !sup.Active {
				writeJSON(w, 400, map[string]string{"error": "Proveedor inválido o inactivo"})
				return
			}
			items := []PurchaseItem{}
			for i, row := range in.Items {
				if row.ProductID <= 0 || row.Quantity <= 0 || row.UnitPrice < 0 {
					writeJSON(w, 400, map[string]string{"error": "Completa cantidad y precio de todos los productos"})
					return
				}
				p := app.productLocked(row.ProductID)
				if p == nil || !p.Active {
					writeJSON(w, 400, map[string]string{"error": "Uno de los productos no existe o está inactivo"})
					return
				}
				linked := p.MainSupplierID == in.SupplierID
				if !linked {
					for _, sid := range p.AlternateSupplierIDs {
						if sid == in.SupplierID {
							linked = true
							break
						}
					}
				}
				if !linked {
					writeJSON(w, 400, map[string]string{"error": fmt.Sprintf("%s no está asociado a este proveedor", p.Name)})
					return
				}
				items = append(items, PurchaseItem{ID: i + 1, ProductID: p.ID, Name: p.Name, Unit: p.Unit, Warehouse: p.Warehouse, Quantity: row.Quantity, UnitPrice: row.UnitPrice, Subtotal: row.Quantity * row.UnitPrice, Note: strings.TrimSpace(row.Note)})
			}
			id := app.nextPurchaseIDLocked()
			t := now()
			po := PurchaseOrder{ID: id, Number: purchaseNumber(id), SupplierID: sup.ID, SupplierName: sup.Name, Status: "borrador", CreatedAt: t, UpdatedAt: t, ExpectedDate: in.ExpectedDate, CreatedBy: u.Name, Note: strings.TrimSpace(in.Note), Items: items, Total: purchaseTotal(items)}
			app.store.Purchases = append(app.store.Purchases, po)
			app.auditLocked(u.Name, "Creó "+po.Number+" para "+sup.Name)
			_ = app.saveLocked()
			writeJSON(w, 201, po)
		default:
			http.NotFound(w, r)
		}
	}))
	mux.HandleFunc("/api/purchases/", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/purchases/"), "/"), "/")
		if len(parts) < 1 {
			http.NotFound(w, r)
			return
		}
		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		action := ""
		if len(parts) > 1 {
			action = parts[1]
		}
		u := app.current(r)
		app.mu.Lock()
		defer app.mu.Unlock()
		po := app.purchaseLocked(id)
		if po == nil {
			writeJSON(w, 404, map[string]string{"error": "Orden no encontrada"})
			return
		}
		if action == "submit" && r.Method == "POST" {
			if po.Status != "borrador" {
				writeJSON(w, 400, map[string]string{"error": "Solo un borrador puede marcarse como solicitado"})
				return
			}
			po.Status = "solicitada"
			po.OrderedAt = now()
			po.UpdatedAt = now()
			app.auditLocked(u.Name, "Marcó "+po.Number+" como solicitada")
			_ = app.saveLocked()
			writeJSON(w, 200, po)
			return
		}
		if action == "cancel" && r.Method == "POST" {
			if po.Status == "anulada" {
				writeJSON(w, 400, map[string]string{"error": "La orden ya está anulada"})
				return
			}
			po.Status = "anulada"
			po.UpdatedAt = now()
			app.auditLocked(u.Name, "Anuló "+po.Number)
			_ = app.saveLocked()
			writeJSON(w, 200, po)
			return
		}
		if action == "receive" && r.Method == "POST" {
			if po.Status != "solicitada" && po.Status != "parcial" {
				writeJSON(w, 400, map[string]string{"error": "Solo se pueden recibir órdenes solicitadas o parciales"})
				return
			}
			var in struct {
				Document      string `json:"document"`
				IssueDate     string `json:"issueDate"`
				DueDate       string `json:"dueDate"`
				PaymentMethod string `json:"paymentMethod"`
				PaymentStatus string `json:"paymentStatus"`
				PaymentDate   string `json:"paymentDate"`
				Note          string `json:"note"`
				CloseOrder    bool   `json:"closeOrder"`
				CloseReason   string `json:"closeReason"`
				Items         []struct {
					ProductID int     `json:"productId"`
					Quantity  float64 `json:"quantity"`
					UnitPrice float64 `json:"unitPrice"`
					Note      string  `json:"note"`
				} `json:"items"`
			}
			if readJSON(r, &in) != nil || len(in.Items) == 0 {
				writeJSON(w, 400, map[string]string{"error": "Ingresa las cantidades recibidas"})
				return
			}
			if in.CloseOrder && strings.TrimSpace(in.CloseReason) == "" {
				writeJSON(w, 400, map[string]string{"error": "Indica el motivo para cerrar la orden con diferencias"})
				return
			}

			// La recepción se procesa como una transacción: primero se valida todo,
			// luego se aplican los cambios y solo al final se guarda. Si ocurre un
			// error inesperado, se restaura el estado anterior para evitar stocks parciales.
			before, snapshotErr := json.Marshal(app.store)
			if snapshotErr != nil {
				writeJSON(w, 500, map[string]string{"error": "No se pudo preparar la recepción. Intenta nuevamente."})
				return
			}
			committed := false
			defer func() {
				if rec := recover(); rec != nil {
					if !committed {
						_ = json.Unmarshal(before, &app.store)
					}
					writeJSON(w, 500, map[string]string{"error": fmt.Sprintf("La recepción no se guardó por un error interno: %v", rec)})
				}
			}()

			type validatedReceiptRow struct {
				Item      *PurchaseItem
				Product   *Product
				Quantity  float64
				UnitPrice float64
				Note      string
			}
			validated := make([]validatedReceiptRow, 0, len(in.Items))
			anyReceived := false
			for _, row := range in.Items {
				if row.Quantity < 0 {
					writeJSON(w, 400, map[string]string{"error": "La cantidad recibida no puede ser negativa"})
					return
				}
				if row.UnitPrice < 0 {
					writeJSON(w, 400, map[string]string{"error": "El precio no puede ser negativo"})
					return
				}
				var item *PurchaseItem
				for i := range po.Items {
					if po.Items[i].ProductID == row.ProductID {
						item = &po.Items[i]
						break
					}
				}
				if item == nil {
					writeJSON(w, 400, map[string]string{"error": "Producto no pertenece a la orden"})
					return
				}
				p := app.productLocked(item.ProductID)
				if p == nil {
					writeJSON(w, 400, map[string]string{"error": "Producto no encontrado: " + item.Name})
					return
				}
				if row.Quantity > 0 {
					anyReceived = true
				}
				validated = append(validated, validatedReceiptRow{Item: item, Product: p, Quantity: row.Quantity, UnitPrice: row.UnitPrice, Note: strings.TrimSpace(row.Note)})
			}
			if !anyReceived && !in.CloseOrder {
				writeJSON(w, 400, map[string]string{"error": "Debes recibir al menos un producto con cantidad mayor a cero"})
				return
			}

			receipt := PurchaseReceipt{ID: len(po.Receipts) + 1, ReceivedAt: now(), ReceivedBy: u.Name, Document: strings.TrimSpace(in.Document), IssueDate: strings.TrimSpace(in.IssueDate), DueDate: strings.TrimSpace(in.DueDate), PaymentMethod: strings.TrimSpace(in.PaymentMethod), PaymentStatus: strings.TrimSpace(in.PaymentStatus), PaymentDate: strings.TrimSpace(in.PaymentDate), Note: strings.TrimSpace(in.Note), Items: []PurchaseReceiptItem{}}
			for _, row := range validated {
				if row.Quantity == 0 {
					continue
				}
				p := row.Product
				item := row.Item
				balance := addStockAt(p, p.Warehouse, row.Quantity)
				p.Cost = row.UnitPrice
				item.ReceivedQuantity += row.Quantity
				item.LastReceivedPrice = row.UnitPrice
				receipt.Total += row.Quantity * row.UnitPrice
				receipt.Items = append(receipt.Items, PurchaseReceiptItem{ProductID: p.ID, Name: p.Name, Quantity: row.Quantity, Unit: p.Unit, UnitPrice: row.UnitPrice, Warehouse: p.Warehouse, Note: row.Note})
				app.store.Movements = append(app.store.Movements, Movement{ID: len(app.store.Movements) + 1, ProductID: p.ID, Warehouse: p.Warehouse, Type: "entrada", Quantity: row.Quantity, Balance: balance, Reason: "Recepción " + po.Number, User: u.Name, CreatedAt: now()})
				app.store.PriceHistory = append(app.store.PriceHistory, PriceHistory{ID: len(app.store.PriceHistory) + 1, ProductID: p.ID, ProductName: p.Name, SupplierID: po.SupplierID, SupplierName: po.SupplierName, PurchaseID: po.ID, OrderNumber: po.Number, UnitPrice: row.UnitPrice, Quantity: row.Quantity, CreatedAt: now()})
			}
			if anyReceived {
				po.Receipts = append(po.Receipts, receipt)
				po.ReceivedTotal += receipt.Total
			}
			po.UpdatedAt = now()
			complete := true
			for _, it := range po.Items {
				if it.ReceivedQuantity+0.000001 < it.Quantity {
					complete = false
					break
				}
			}
			if in.CloseOrder {
				// Si el usuario decide cerrar, respetamos la decisión aunque haya productos
				// sin recibir. Si todo llegó (o llegó más), queda simplemente como recibida.
				if complete {
					po.Status = "recibida"
				} else {
					po.Status = "cerrada_diferencias"
					po.ClosedWithDifferences = true
				}
				po.ClosedAt = now()
				po.ClosedBy = u.Name
				po.CloseReason = strings.TrimSpace(in.CloseReason)
			} else if complete {
				po.Status = "recibida"
			} else {
				po.Status = "parcial"
			}
			if po.Status == "recibida" || po.Status == "cerrada_diferencias" {
				if sup := app.supplierLocked(po.SupplierID); sup != nil {
					sup.LastPurchase = now()
					sup.LastDispatch = now()
				}
			}
			// Las facturas de mercadería permanecen en Compras/Proveedores; no se duplican en Gastos.
			app.auditLocked(u.Name, fmt.Sprintf("Registró recepción de %s por $%.0f", po.Number, receipt.Total))
			if err := app.saveLocked(); err != nil {
				_ = json.Unmarshal(before, &app.store)
				writeJSON(w, 500, map[string]string{"error": "No se pudo guardar la recepción. No se modificó el inventario: " + err.Error()})
				return
			}
			committed = true
			writeJSON(w, 200, po)
			return
		}
		if action != "" {
			http.NotFound(w, r)
			return
		}
		switch r.Method {
		case "PUT":
			if po.Status != "borrador" {
				writeJSON(w, 400, map[string]string{"error": "Solo se pueden editar órdenes en borrador"})
				return
			}
			var in struct {
				SupplierID   int    `json:"supplierId"`
				ExpectedDate string `json:"expectedDate"`
				Note         string `json:"note"`
				Items        []struct {
					ProductID int     `json:"productId"`
					Quantity  float64 `json:"quantity"`
					UnitPrice float64 `json:"unitPrice"`
					Note      string  `json:"note"`
				} `json:"items"`
			}
			if readJSON(r, &in) != nil || in.SupplierID <= 0 || len(in.Items) == 0 {
				writeJSON(w, 400, map[string]string{"error": "Selecciona proveedor y agrega productos"})
				return
			}
			sup := app.supplierLocked(in.SupplierID)
			if sup == nil || !sup.Active {
				writeJSON(w, 400, map[string]string{"error": "Proveedor inválido o inactivo"})
				return
			}
			items := []PurchaseItem{}
			for i, row := range in.Items {
				if row.ProductID <= 0 || row.Quantity <= 0 || row.UnitPrice < 0 {
					writeJSON(w, 400, map[string]string{"error": "Completa cantidad y precio de todos los productos"})
					return
				}
				p := app.productLocked(row.ProductID)
				if p == nil || !p.Active {
					writeJSON(w, 400, map[string]string{"error": "Producto inválido"})
					return
				}
				linked := p.MainSupplierID == in.SupplierID
				if !linked {
					for _, sid := range p.AlternateSupplierIDs {
						if sid == in.SupplierID {
							linked = true
							break
						}
					}
				}
				if !linked {
					writeJSON(w, 400, map[string]string{"error": fmt.Sprintf("%s no está asociado a este proveedor", p.Name)})
					return
				}
				items = append(items, PurchaseItem{ID: i + 1, ProductID: p.ID, Name: p.Name, Unit: p.Unit, Warehouse: p.Warehouse, Quantity: row.Quantity, UnitPrice: row.UnitPrice, Subtotal: row.Quantity * row.UnitPrice, Note: strings.TrimSpace(row.Note)})
			}
			po.SupplierID = sup.ID
			po.SupplierName = sup.Name
			po.ExpectedDate = in.ExpectedDate
			po.Note = strings.TrimSpace(in.Note)
			po.Items = items
			po.Total = purchaseTotal(items)
			po.UpdatedAt = now()
			app.auditLocked(u.Name, "Editó "+po.Number)
			_ = app.saveLocked()
			writeJSON(w, 200, po)
		case "DELETE":
			if po.Status != "borrador" {
				writeJSON(w, 400, map[string]string{"error": "Solo se pueden eliminar órdenes en borrador"})
				return
			}
			idx := -1
			for i := range app.store.Purchases {
				if app.store.Purchases[i].ID == id {
					idx = i
					break
				}
			}
			if idx >= 0 {
				app.store.Purchases = append(app.store.Purchases[:idx], app.store.Purchases[idx+1:]...)
			}
			app.auditLocked(u.Name, "Eliminó "+po.Number)
			_ = app.saveLocked()
			writeJSON(w, 200, map[string]bool{"ok": true})
		default:
			http.NotFound(w, r)
		}
	}))
	mux.HandleFunc("/api/expenses", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			app.mu.RLock()
			defer app.mu.RUnlock()
			rows := make([]Expense, len(app.store.Expenses))
			for i := range app.store.Expenses {
				rows[len(rows)-1-i] = app.store.Expenses[i]
			}
			writeJSON(w, 200, rows)
		case "POST":
			var in Expense
			if readJSON(r, &in) != nil || strings.TrimSpace(in.SupplierName) == "" || in.Amount < 0 {
				writeJSON(w, 400, map[string]string{"error": "Completa proveedor o servicio y monto"})
				return
			}
			u := app.current(r)
			app.mu.Lock()
			defer app.mu.Unlock()
			in.ID = app.nextExpenseIDLocked()
			in.Source = "manual"
			in.CreatedAt = now()
			in.CreatedBy = u.Name
			if in.Status != "pagada" {
				in.Status = "pendiente"
			}
			if in.Status == "pagada" && in.PaymentDate == "" {
				in.PaymentDate = time.Now().Format("2006-01-02")
			}
			app.store.Expenses = append(app.store.Expenses, in)
			app.auditLocked(u.Name, "Registró gasto "+in.SupplierName)
			_ = app.saveLocked()
			writeJSON(w, 201, in)
		default:
			http.NotFound(w, r)
		}
	}))
	mux.HandleFunc("/api/expenses/", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/expenses/"), "/"))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		u := app.current(r)
		app.mu.Lock()
		defer app.mu.Unlock()
		var e *Expense
		for i := range app.store.Expenses {
			if app.store.Expenses[i].ID == id {
				e = &app.store.Expenses[i]
				break
			}
		}
		if e == nil {
			writeJSON(w, 404, map[string]string{"error": "Gasto no encontrado"})
			return
		}
		if r.Method == "DELETE" {
			idx := -1
			for i := range app.store.Expenses {
				if app.store.Expenses[i].ID == id {
					idx = i
					break
				}
			}
			if idx < 0 {
				writeJSON(w, 404, map[string]string{"error": "Gasto no encontrado"})
				return
			}
			name := app.store.Expenses[idx].SupplierName
			app.store.Expenses = append(app.store.Expenses[:idx], app.store.Expenses[idx+1:]...)
			app.auditLocked(u.Name, "Eliminó gasto "+name)
			_ = app.saveLocked()
			writeJSON(w, 200, map[string]bool{"ok": true})
			return
		}
		if r.Method == "PUT" {
			var in Expense
			if readJSON(r, &in) != nil {
				writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
				return
			}
			e.Category = in.Category
			e.SupplierName = in.SupplierName
			e.Document = in.Document
			e.IssueDate = in.IssueDate
			e.DueDate = in.DueDate
			e.Amount = in.Amount
			e.PaymentMethod = in.PaymentMethod
			e.Status = in.Status
			e.PaymentDate = in.PaymentDate
			e.Note = in.Note
			if e.Status != "pagada" {
				e.Status = "pendiente"
				e.PaymentDate = ""
			} else if e.PaymentDate == "" {
				e.PaymentDate = time.Now().Format("2006-01-02")
			}
			app.auditLocked(u.Name, "Actualizó gasto "+e.SupplierName)
			_ = app.saveLocked()
			writeJSON(w, 200, e)
			return
		}
		http.NotFound(w, r)
	}))
	mux.HandleFunc("/api/movements", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		app.mu.RLock()
		defer app.mu.RUnlock()
		type MovementView struct {
			ID          int     `json:"id"`
			ProductID   int     `json:"productId"`
			ProductName string  `json:"productName"`
			Warehouse   string  `json:"warehouse"`
			Type        string  `json:"type"`
			Quantity    float64 `json:"quantity"`
			Balance     float64 `json:"balance"`
			Reason      string  `json:"reason"`
			User        string  `json:"user"`
			CreatedAt   string  `json:"createdAt"`
		}
		rows := []MovementView{}
		for i := len(app.store.Movements) - 1; i >= 0; i-- {
			m := app.store.Movements[i]
			name, warehouse := "Producto", normalizeWarehouse(m.Warehouse)
			for _, p := range app.store.Products {
				if p.ID == m.ProductID {
					name = p.Name
					if warehouse == "" {
						warehouse = p.Warehouse
					}
					break
				}
			}
			rows = append(rows, MovementView{m.ID, m.ProductID, name, warehouse, m.Type, m.Quantity, m.Balance, m.Reason, m.User, m.CreatedAt})
		}
		writeJSON(w, 200, rows)
	}))
	mux.HandleFunc("/api/inventory/physical", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		var in struct {
			Warehouse string `json:"warehouse"`
			Note      string `json:"note"`
			Items     []struct {
				ProductID int     `json:"productId"`
				NewStock  float64 `json:"newStock"`
			} `json:"items"`
		}
		if readJSON(r, &in) != nil || strings.TrimSpace(in.Warehouse) == "" || len(in.Items) == 0 {
			writeJSON(w, 400, map[string]string{"error": "Inventario físico incompleto"})
			return
		}
		u := app.current(r)
		app.mu.Lock()
		defer app.mu.Unlock()
		changed := 0
		for _, row := range in.Items {
			if row.NewStock < 0 {
				writeJSON(w, 400, map[string]string{"error": "El stock contado no puede ser negativo"})
				return
			}
			p := app.productLocked(row.ProductID)
			if p == nil {
				continue
			}
			oldStock := stockAt(p, in.Warehouse)
			diff := row.NewStock - oldStock
			if diff == 0 {
				continue
			}
			setStockAt(p, in.Warehouse, row.NewStock)
			app.store.Movements = append(app.store.Movements, Movement{ID: len(app.store.Movements) + 1, ProductID: p.ID, Warehouse: in.Warehouse, Type: "ajuste", Quantity: diff, Balance: row.NewStock, Reason: "Inventario físico · " + in.Warehouse + func() string {
				if strings.TrimSpace(in.Note) != "" {
					return " · " + strings.TrimSpace(in.Note)
				}
				return ""
			}(), User: u.Name, CreatedAt: now()})
			changed++
		}
		app.auditLocked(u.Name, fmt.Sprintf("Confirmó inventario físico de %s (%d ajustes)", in.Warehouse, changed))
		if err := app.saveLocked(); err != nil {
			writeJSON(w, 500, map[string]string{"error": "No fue posible guardar el inventario físico: " + err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "changed": changed})
	}))
	mux.HandleFunc("/api/inventory/adjust", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		var in struct {
			ProductID int     `json:"productId"`
			Warehouse string  `json:"warehouse"`
			NewStock  float64 `json:"newStock"`
			Reason    string  `json:"reason"`
			Note      string  `json:"note"`
		}
		if readJSON(r, &in) != nil || in.ProductID <= 0 || in.NewStock < 0 || strings.TrimSpace(in.Reason) == "" {
			writeJSON(w, 400, map[string]string{"error": "Completa producto, nuevo stock y motivo"})
			return
		}
		u := app.current(r)
		app.mu.Lock()
		defer app.mu.Unlock()
		p := app.productLocked(in.ProductID)
		if p == nil {
			writeJSON(w, 404, map[string]string{"error": "Producto no encontrado"})
			return
		}
		warehouse := normalizeWarehouse(in.Warehouse)
		if warehouse == "" {
			warehouse = p.Warehouse
		}
		oldStock := stockAt(p, warehouse)
		diff := in.NewStock - oldStock
		if diff == 0 {
			writeJSON(w, 400, map[string]string{"error": "El nuevo stock es igual al actual"})
			return
		}
		setStockAt(p, warehouse, in.NewStock)
		detail := in.Reason + " · " + warehouse
		if strings.TrimSpace(in.Note) != "" {
			detail += ": " + strings.TrimSpace(in.Note)
		}
		balance := stockAt(p, warehouse)
		app.store.Movements = append(app.store.Movements, Movement{ID: len(app.store.Movements) + 1, ProductID: p.ID, Warehouse: warehouse, Type: "ajuste", Quantity: diff, Balance: balance, Reason: detail, User: u.Name, CreatedAt: now()})
		app.auditLocked(u.Name, fmt.Sprintf("Ajustó stock de %s en %s: %+g (saldo %.2f)", p.Name, warehouse, diff, balance))
		if err := app.saveLocked(); err != nil {
			writeJSON(w, 500, map[string]string{"error": "No fue posible guardar el ajuste: " + err.Error()})
			return
		}
		writeJSON(w, 200, map[string]any{"ok": true, "difference": diff, "balance": balance})
	}))

	sub, _ := fs.Sub(embedded, "web")
	mux.Handle("/", http.FileServer(http.FS(sub)))
	port := 8080
	for ; port <= 8090; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}

		ip := lanIP()
		localURL := fmt.Sprintf("http://127.0.0.1:%d", port)
		fmt.Printf("\nEB Gestión 1.0 Desktop - Entre Bahías activa\nPC: %s\nTeléfono: http://%s:%d\nDatos: %s\nRespaldos: %s\n\n", localURL, ip, port, dataDir, app.backupDir())

		serveErr := make(chan error, 1)
		go func() {
			serveErr <- http.Serve(ln, mux)
		}()

		// La ventana de escritorio usa WebView2 en Windows. El servidor continúa
		// escuchando en la red local para Cocina, Gerencia y los demás equipos.
		if err := runDesktop(localURL); err != nil {
			fmt.Printf("No se pudo abrir la ventana de escritorio: %v\n", err)
		}

		_ = ln.Close()
		select {
		case err := <-serveErr:
			if err != nil && err != http.ErrServerClosed && !strings.Contains(err.Error(), "use of closed network connection") {
				fmt.Printf("El servidor terminó con error: %v\n", err)
			}
		case <-time.After(2 * time.Second):
		}
		return
	}
	fmt.Println("No fue posible iniciar EB Gestión: los puertos 8080 a 8090 están ocupados.")
}
func lanIP() string {
	c, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).IP.String()
}
