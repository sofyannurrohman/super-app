package ai

import "testing"

func TestParsePlanValidatesToolArguments(t *testing.T) {
	plan, err := ParsePlan(`{"intent":"customer_search","needs_clarification":false,"tool":"search_customer","arguments":{"query":"Toko Maju"}}`)
	if err != nil || plan.Tool != "search_customer" {
		t.Fatalf("unexpected plan: %+v, err=%v", plan, err)
	}
}

func TestParsePlanSearchEmployee(t *testing.T) {
	plan, err := ParsePlan(`{"intent":"employee_search","needs_clarification":false,"tool":"search_employee","arguments":{"keyword":"Budi"}}`)
	if err != nil || plan.Tool != "search_employee" {
		t.Fatalf("unexpected plan: %+v, err=%v", plan, err)
	}
}

func TestParsePlanWithMarkdownAndExtraText(t *testing.T) {
	raw := "Berikut adalah perencanaannya:\n```json\n{\"intent\":\"sales_summary\",\"needs_clarification\":false,\"tool\":\"get_sales_summary\",\"arguments\":{\"start_date\":\"2025-07-01\"}}\n```\nSemoga membantu!"
	plan, err := ParsePlan(raw)
	if err != nil || plan.Tool != "get_sales_summary" {
		t.Fatalf("unexpected plan: %+v, err=%v", plan, err)
	}
}

func TestParsePlanRejectsUnknownTool(t *testing.T) {
	_, err := ParsePlan(`{"intent":"unsafe","needs_clarification":false,"tool":"run_sql","arguments":{"query":"select 1"}}`)
	if err == nil {
		t.Fatal("expected unknown tool error")
	}
}

func TestParsePlanInventoryStock(t *testing.T) {
	plan, err := ParsePlan(`{"intent":"inventory_stock","needs_clarification":false,"tool":"get_inventory_stock","arguments":{"warehouse_name":"Solo","product_query":"Kecap"}}`)
	if err != nil || plan.Tool != "get_inventory_stock" {
		t.Fatalf("unexpected plan: %+v, err=%v", plan, err)
	}
	if plan.Arguments["warehouse_name"] != "Solo" {
		t.Fatalf("expected warehouse_name Solo, got %v", plan.Arguments["warehouse_name"])
	}
}

func TestBuildPlannerPromptWithHistory(t *testing.T) {
	history := []ChatMessage{
		{Role: "user", Content: "Cek stok produk di gudang"},
		{Role: "assistant", Content: "Produk apa yang ingin Anda cek stoknya dan di gudang mana?"},
	}
	prompt := BuildPlannerPrompt(history, "perusahaan contoh Cabang Solo")
	if prompt == "" {
		t.Fatal("prompt should not be empty")
	}
}

func TestParsePlanNoneForProductKnowledge(t *testing.T) {
	plan, err := ParsePlan(`{"intent":"product_knowledge","needs_clarification":false,"tool":"none","arguments":{}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Tool != "none" {
		t.Fatalf("expected tool 'none', got %q", plan.Tool)
	}
}

func TestParsePlanEmptyToolTreatedAsNone(t *testing.T) {
	plan, err := ParsePlan(`{"intent":"greeting","needs_clarification":false,"tool":"","arguments":{}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan.Tool != "none" {
		t.Fatalf("expected tool 'none', got %q", plan.Tool)
	}
}

func TestParsePlanEmployeeAttendance(t *testing.T) {
	plan, err := ParsePlan(`{"intent":"attendance_check","needs_clarification":false,"tool":"get_employee_attendance","arguments":{"date":"2026-10-05"}}`)
	if err != nil || plan.Tool != "get_employee_attendance" {
		t.Fatalf("unexpected plan: %+v, err=%v", plan, err)
	}
	if plan.Arguments["date"] != "2026-10-05" {
		t.Fatalf("expected date 2026-10-05, got %v", plan.Arguments["date"])
	}
}

func TestParsePlanGetProductsCatalog(t *testing.T) {
	plan, err := ParsePlan(`{"intent":"product_catalog","needs_clarification":false,"tool":"get_products_catalog","arguments":{"category":"Finished Goods","query":"Kecap"}}`)
	if err != nil || plan.Tool != "get_products_catalog" {
		t.Fatalf("unexpected plan: %+v, err=%v", plan, err)
	}
	if plan.Arguments["category"] != "Finished Goods" {
		t.Fatalf("expected category Finished Goods, got %v", plan.Arguments["category"])
	}
	if plan.Arguments["query"] != "Kecap" {
		t.Fatalf("expected query Kecap, got %v", plan.Arguments["query"])
	}
}

func TestFastMatchPlanEmployeeSearch(t *testing.T) {
	cases := []struct {
		input           string
		expectedKeyword string
	}{
		{"siapa itu sofyan", "sofyan"},
		{"siapa sofyan", "sofyan"},
		{"sofyan itu siapa", "sofyan"},
		{"siapa karyawan Sofyan nur rohman", "Sofyan nur rohman"},
		{"kenal dengan Budi?", "Budi"},
		{"cari karyawan Joko Widodo", "Joko Widodo"},
		{"profil Siti Aminah", "Siti Aminah"},
		{"kontak Pak Hendra", "Hendra"},
	}

	for _, tc := range cases {
		plan, matched := FastMatchPlan(tc.input)
		if !matched {
			t.Fatalf("expected fast match for query %q, but got matched=false", tc.input)
		}
		if plan.Tool != "search_employee" {
			t.Fatalf("query %q: expected tool 'search_employee', got %q", tc.input, plan.Tool)
		}
		kw, _ := plan.Arguments["keyword"].(string)
		if kw != tc.expectedKeyword {
			t.Fatalf("query %q: expected keyword %q, got %q", tc.input, tc.expectedKeyword, kw)
		}
	}
}

func TestFastMatchPlanOtherOperationalTools(t *testing.T) {
	tests := []struct {
		input        string
		expectedTool string
	}{
		{"Siapa saja yang hadir hari ini?", "get_employee_attendance"},
		{"Daftar presensi hari ini", "get_employee_attendance"},
		{"Produk perusahaan contoh & keunggulannya", "get_products_catalog"},
		{"Ada antrean approval apa saja?", "get_pending_approvals"},
		{"Daftar cabang perusahaan contoh ada di mana saja?", "get_departments_and_branches"},
		{"Cek kas operasional pabrik", "get_kas_pabrik_summary"},
		{"Status pengiriman surat jalan hari ini", "get_delivery_orders"},
		{"Daftar kendaraan mobil kanvas", "get_vehicles_status"},
	}

	for _, tc := range tests {
		plan, matched := FastMatchPlan(tc.input)
		if !matched {
			t.Fatalf("expected fast match for query %q, but got matched=false", tc.input)
		}
		if plan.Tool != tc.expectedTool {
			t.Fatalf("query %q: expected tool %q, got %q", tc.input, tc.expectedTool, plan.Tool)
		}
	}
}

func TestFastMatchPlanFallbackToLLM(t *testing.T) {
	// Pertanyaan non-operasional atau pertanyaan dengan kalimat bebas yang tidak cocok pattern
	generalQueries := []string{
		"Halo apa kabar?",
		"Jelaskan rumus matematika luas lingkaran",
		"Bagaimana cara negosiasi dengan warung makan yang keras kepala?",
	}

	for _, q := range generalQueries {
		_, matched := FastMatchPlan(q)
		if matched {
			t.Fatalf("query %q should not be fast matched, but matched=true", q)
		}
	}
}


