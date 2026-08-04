package main

import (
	"crypto/rand"
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
type Product struct {
	ID                   int     `json:"id"`
	Code                 string  `json:"code"`
	Name                 string  `json:"name"`
	Category             string  `json:"category"`
	Unit                 string  `json:"unit"`
	Warehouse            string  `json:"warehouse"`
	Stock                float64 `json:"stock"`
	MinStock             float64 `json:"minStock"`
	Cost                 float64 `json:"cost"`
	MainSupplierID       int     `json:"mainSupplierId"`
	AlternateSupplierIDs []int   `json:"alternateSupplierIds"`
	Notes                string  `json:"notes"`
	Active               bool    `json:"active"`
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
	ID         int                   `json:"id"`
	ReceivedAt string                `json:"receivedAt"`
	ReceivedBy string                `json:"receivedBy"`
	Document   string                `json:"document"`
	Note       string                `json:"note"`
	Items      []PurchaseReceiptItem `json:"items"`
	Total      float64               `json:"total"`
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
	ID          int           `json:"id"`
	Requester   string        `json:"requester"`
	Area        string        `json:"area"`
	Status      string        `json:"status"`
	CreatedAt   string        `json:"createdAt"`
	UpdatedAt   string        `json:"updatedAt"`
	DeliveredBy string        `json:"deliveredBy"`
	Note        string        `json:"note"`
	Items       []RequestItem `json:"items"`
}
type Movement struct {
	ID        int     `json:"id"`
	ProductID int     `json:"productId"`
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
	Requests          []Request           `json:"requests"`
	Movements         []Movement          `json:"movements"`
	Audit             []Audit             `json:"audit"`
	NotificationReads map[string][]string `json:"notificationReads"`
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

var users = map[string]User{
	"stefy":   {"stefy", "Stefy", "admin_bodega", "123456"},
	"claudio": {"claudio", "Claudio", "gerencia", "123456"},
	"jessica": {"jessica", "Jessica", "cocina", "123456"},
	"caja":    {"caja", "Caja", "caja", "123456"},
}

func now() string { return time.Now().Format("2006-01-02 15:04:05") }
func seed() Store {
	return Store{
		Products: []Product{
			{ID: 1, Code: "P001", Name: "Salmón Atlántico", Category: "Congelados", Unit: "Kg", Warehouse: "Congelados", Stock: 25, MinStock: 10, Cost: 11250, MainSupplierID: 1, AlternateSupplierIDs: []int{2}, Active: true},
			{ID: 2, Code: "P002", Name: "Camarón 51/60", Category: "Congelados", Unit: "Kg", Warehouse: "Congelados", Stock: 12, MinStock: 5, Cost: 8900, MainSupplierID: 1, Active: true},
			{ID: 3, Code: "P003", Name: "Arroz G1", Category: "Abarrotes", Unit: "Kg", Warehouse: "Abarrotes", Stock: 45, MinStock: 20, Cost: 1450, MainSupplierID: 2, AlternateSupplierIDs: []int{3}, Active: true},
			{ID: 4, Code: "P004", Name: "Aceite Vegetal 1L", Category: "Abarrotes", Unit: "Botella", Warehouse: "Abarrotes", Stock: 8, MinStock: 5, Cost: 1900, MainSupplierID: 2, Active: true},
			{ID: 5, Code: "P005", Name: "Limón", Category: "Verduras", Unit: "Kg", Warehouse: "Cocina", Stock: 15, MinStock: 7, Cost: 1200, MainSupplierID: 3, Active: true},
			{ID: 6, Code: "P006", Name: "Guantes", Category: "Aseo", Unit: "Caja", Warehouse: "Cocina", Stock: 6, MinStock: 3, Cost: 4500, MainSupplierID: 2, Active: true},
		},
		Suppliers: []Supplier{{ID: 1, Name: "Proveedor Congelados", BusinessName: "Proveedor Congelados SpA", Category: "Congelados", Contact: "Ventas", OrderDays: []string{"Lunes", "Miércoles"}, OrderDeadline: "12:00", DeliveryDays: []string{"Martes", "Jueves"}, PaymentMethod: "Transferencia", Active: true}, {ID: 2, Name: "Distribuidora Abarrotes", BusinessName: "Distribuidora Abarrotes Ltda.", Category: "Abarrotes", Contact: "Ventas", OrderDays: []string{"Lunes", "Martes", "Miércoles", "Jueves"}, OrderDeadline: "12:00", DeliveryDays: []string{"Martes", "Viernes"}, PaymentMethod: "Transferencia", Active: true}, {ID: 3, Name: "Proveedor Verduras", BusinessName: "Proveedor Verduras", Category: "Verduras", Contact: "Ventas", OrderDays: []string{"Lunes", "Miércoles", "Viernes"}, OrderDeadline: "11:00", DeliveryDays: []string{"Martes", "Jueves", "Sábado"}, PaymentMethod: "Contado", Active: true}},
		Purchases: []PurchaseOrder{}, PriceHistory: []PriceHistory{}, Requests: []Request{}, Movements: []Movement{}, Audit: []Audit{}, NotificationReads: map[string][]string{},
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
			if strings.TrimSpace(a.store.Products[i].Category) == "" {
				a.store.Products[i].Category = a.store.Products[i].Warehouse
			}
			if a.store.Products[i].AlternateSupplierIDs == nil {
				a.store.Products[i].AlternateSupplierIDs = []int{}
			}
		}
		if a.store.Purchases == nil {
			a.store.Purchases = []PurchaseOrder{}
		}
		for i := range a.store.Purchases {
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
	tmp := a.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0644); err != nil {
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
	a.mu.Lock()
	defer a.mu.Unlock()
	s, ok := a.sessions[t]
	if !ok {
		return User{}, false
	}
	if time.Now().After(s.ExpiresAt) {
		delete(a.sessions, t)
		return User{}, false
	}
	// Sesión deslizante: renueva actividad por 8 horas.
	s.ExpiresAt = time.Now().Add(8 * time.Hour)
	a.sessions[t] = s
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
func (a *App) current(r *http.Request) User { u, _ := users[r.Header.Get("X-User")]; return u }
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
		u, ok := users[strings.ToLower(strings.TrimSpace(in.Username))]
		if !ok || u.Password != in.Password {
			writeJSON(w, 401, map[string]string{"error": "Usuario o contraseña incorrectos"})
			return
		}
		t := token()
		app.mu.Lock()
		app.sessions[t] = Session{User: u, ExpiresAt: time.Now().Add(8 * time.Hour)}
		app.auditLocked(u.Name, "Inicio de sesión")
		_ = app.saveLocked()
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
		err := app.saveLocked()
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
		writeJSON(w, 200, map[string]any{"dataPath": app.path, "backupPath": app.backupDir(), "version": "1.0 Estabilidad 1", "lanIP": lanIP()})
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
			if u.Role == "cocina" || u.Role == "caja" {
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
				in.Stock = p.Stock
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
				value += p.Stock * p.Cost
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
			ID   string `json:"id"`
			Icon string `json:"icon"`
			Text string `json:"text"`
			Time string `json:"time"`
		}
		build := func() []notice {
			notes := []notice{}
			if u.Role == "admin_bodega" || u.Role == "gerencia" {
				for _, q := range app.store.Requests {
					if q.Status == "pendiente" {
						notes = append(notes, notice{fmt.Sprintf("request:%d", q.ID), "📋", "Nueva solicitud de " + q.Requester, q.CreatedAt})
					}
				}
				for _, p := range app.store.Products {
					if p.Active && p.Stock <= p.MinStock {
						notes = append(notes, notice{fmt.Sprintf("stock:%d:%.3f", p.ID, p.Stock), "⚠️", p.Name + " con stock crítico", "Ahora"})
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
							notes = append(notes, notice{fmt.Sprintf("supplier:%d:%s", sup.ID, dateKey), "🛒", msg, "Hoy"})
							break
						}
					}
				}
				for _, po := range app.store.Purchases {
					if po.Status == "solicitada" || po.Status == "parcial" {
						notes = append(notes, notice{fmt.Sprintf("purchase:%d:%s", po.ID, po.Status), "📦", po.Number + " pendiente de recepción", po.ExpectedDate})
					}
				}
			}
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
			_ = app.saveLocked()
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
				if u.Role == "cocina" || u.Role == "caja" {
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
						rows[i].Items[j].Stock = p.Stock
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
			if u.Role == "cocina" {
				area = "Cocina"
			}
			if u.Role == "caja" {
				area = "Caja"
			}
			q := Request{id, u.Name, area, "pendiente", now(), now(), "", in.Note, items}
			app.store.Requests = append(app.store.Requests, q)
			app.auditLocked(u.Name, fmt.Sprintf("Creó solicitud #%d", id))
			_ = app.saveLocked()
			app.mu.Unlock()
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
			var in struct {
				Items []struct {
					ItemID       int     `json:"itemId"`
					DeliveredQty float64 `json:"deliveredQty"`
					Reason       string  `json:"reason"`
				} `json:"items"`
			}
			if readJSON(r, &in) != nil {
				writeJSON(w, 400, map[string]string{"error": "Datos inválidos"})
				return
			}
			anyDelivered := false
			complete := true
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
				qty := x.DeliveredQty
				if qty < 0 {
					qty = 0
				}
				if qty > p.Stock {
					qty = p.Stock
				}
				if qty < item.RequestedQty && strings.TrimSpace(x.Reason) == "" {
					writeJSON(w, 400, map[string]string{"error": "Indica el motivo cuando entregas menos"})
					return
				}
				item.DeliveredQty = &qty
				item.Reason = x.Reason
				if qty > 0 {
					anyDelivered = true
					p.Stock -= qty
					app.store.Movements = append(app.store.Movements, Movement{len(app.store.Movements) + 1, p.ID, "salida", -qty, p.Stock, fmt.Sprintf("Solicitud #%d", id), u.Name, now()})
				}
				if qty < item.RequestedQty {
					complete = false
				}
			}
			if complete {
				q.Status = "entregada"
			} else if anyDelivered {
				q.Status = "parcial"
			} else {
				q.Status = "no_entregada"
			}
			q.DeliveredBy = u.Name
			q.UpdatedAt = now()
			app.auditLocked(u.Name, fmt.Sprintf("Finalizó solicitud #%d (%s)", id, q.Status))
			_ = app.saveLocked()
			writeJSON(w, 200, map[string]any{"ok": true, "status": q.Status})
			return
		}
		http.NotFound(w, r)
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
				Document    string `json:"document"`
				Note        string `json:"note"`
				CloseOrder  bool   `json:"closeOrder"`
				CloseReason string `json:"closeReason"`
				Items       []struct {
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
			receipt := PurchaseReceipt{ID: len(po.Receipts) + 1, ReceivedAt: now(), ReceivedBy: u.Name, Document: strings.TrimSpace(in.Document), Note: strings.TrimSpace(in.Note), Items: []PurchaseReceiptItem{}}
			anyReceived := false
			for _, row := range in.Items {
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
				pending := item.Quantity - item.ReceivedQuantity
				if row.Quantity < 0 || row.Quantity > pending+0.000001 {
					writeJSON(w, 400, map[string]string{"error": fmt.Sprintf("Cantidad inválida para %s. Pendiente: %.2f", item.Name, pending)})
					return
				}
				if row.Quantity == 0 {
					continue
				}
				if row.UnitPrice < 0 {
					writeJSON(w, 400, map[string]string{"error": "El precio no puede ser negativo"})
					return
				}
				p := app.productLocked(item.ProductID)
				if p == nil {
					writeJSON(w, 400, map[string]string{"error": "Producto no encontrado"})
					return
				}
				anyReceived = true
				p.Stock += row.Quantity
				p.Cost = row.UnitPrice
				item.ReceivedQuantity += row.Quantity
				item.LastReceivedPrice = row.UnitPrice
				receipt.Total += row.Quantity * row.UnitPrice
				receipt.Items = append(receipt.Items, PurchaseReceiptItem{ProductID: p.ID, Name: p.Name, Quantity: row.Quantity, Unit: p.Unit, UnitPrice: row.UnitPrice, Warehouse: p.Warehouse, Note: strings.TrimSpace(row.Note)})
				app.store.Movements = append(app.store.Movements, Movement{ID: len(app.store.Movements) + 1, ProductID: p.ID, Type: "entrada", Quantity: row.Quantity, Balance: p.Stock, Reason: "Recepción " + po.Number, User: u.Name, CreatedAt: now()})
				app.store.PriceHistory = append(app.store.PriceHistory, PriceHistory{ID: len(app.store.PriceHistory) + 1, ProductID: p.ID, ProductName: p.Name, SupplierID: po.SupplierID, SupplierName: po.SupplierName, PurchaseID: po.ID, OrderNumber: po.Number, UnitPrice: row.UnitPrice, Quantity: row.Quantity, CreatedAt: now()})
			}
			if !anyReceived {
				writeJSON(w, 400, map[string]string{"error": "Debes recibir al menos un producto con cantidad mayor a cero"})
				return
			}
			po.Receipts = append(po.Receipts, receipt)
			po.ReceivedTotal += receipt.Total
			po.UpdatedAt = now()
			complete := true
			for _, it := range po.Items {
				if it.ReceivedQuantity+0.000001 < it.Quantity {
					complete = false
					break
				}
			}
			if complete {
				po.Status = "recibida"
				if sup := app.supplierLocked(po.SupplierID); sup != nil {
					sup.LastPurchase = now()
					sup.LastDispatch = now()
				}
			} else if in.CloseOrder {
				po.Status = "cerrada_diferencias"
				po.ClosedAt = now()
				po.ClosedBy = u.Name
				po.CloseReason = strings.TrimSpace(in.CloseReason)
				if po.CloseReason == "" {
					po.CloseReason = "Orden cerrada con cantidades no recibidas"
				}
				po.ClosedWithDifferences = true
			} else {
				po.Status = "parcial"
			}
			app.auditLocked(u.Name, fmt.Sprintf("Registró recepción de %s por $%.0f", po.Number, receipt.Total))
			_ = app.saveLocked()
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
			name, warehouse := "Producto", ""
			for _, p := range app.store.Products {
				if p.ID == m.ProductID {
					name, warehouse = p.Name, p.Warehouse
					break
				}
			}
			rows = append(rows, MovementView{m.ID, m.ProductID, name, warehouse, m.Type, m.Quantity, m.Balance, m.Reason, m.User, m.CreatedAt})
		}
		writeJSON(w, 200, rows)
	}))
	mux.HandleFunc("/api/inventory/adjust", require(app, "admin_bodega", "gerencia")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.NotFound(w, r)
			return
		}
		var in struct {
			ProductID int     `json:"productId"`
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
		diff := in.NewStock - p.Stock
		if diff == 0 {
			writeJSON(w, 400, map[string]string{"error": "El nuevo stock es igual al actual"})
			return
		}
		p.Stock = in.NewStock
		detail := in.Reason
		if strings.TrimSpace(in.Note) != "" {
			detail += ": " + strings.TrimSpace(in.Note)
		}
		app.store.Movements = append(app.store.Movements, Movement{len(app.store.Movements) + 1, p.ID, "ajuste", diff, p.Stock, detail, u.Name, now()})
		app.auditLocked(u.Name, fmt.Sprintf("Ajustó stock de %s: %+g (saldo %.2f)", p.Name, diff, p.Stock))
		_ = app.saveLocked()
		writeJSON(w, 200, map[string]any{"ok": true, "difference": diff, "balance": p.Stock})
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
