import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:drift/drift.dart' as drift;
import 'dart:convert';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/injection.dart' as di;

class ManualStockAdjustmentPage extends StatefulWidget {
  const ManualStockAdjustmentPage({super.key});

  @override
  State<ManualStockAdjustmentPage> createState() =>
      _ManualStockAdjustmentPageState();
}

class _ManualStockAdjustmentPageState extends State<ManualStockAdjustmentPage> {
  final _db = di.sl<AppDatabase>();
  bool _isLoading = true;
  List<Map<String, dynamic>> _stockItems = [];

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);
    try {
      final stocks = await _db.select(_db.salesStock).get();
      final products = await _db.select(_db.products).get();

      final List<Map<String, dynamic>> items = [];
      for (var s in stocks) {
        if (s.quantity > 0) {
          final prod = products.where((p) => p.id == s.productId).firstOrNull;
          if (prod != null) {
            items.add({
              'product_id': s.productId,
              'name': prod.name,
              'original_qty': s.quantity,
              'current_qty': s.quantity,
            });
          }
        }
      }

      setState(() {
        _stockItems = items;
        _isLoading = false;
      });
    } catch (e) {
      debugPrint('Error loading stock: $e');
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _saveChanges() async {
    setState(() => _isLoading = true);
    try {
      // Call API First
      final List<Map<String, dynamic>> adjustItems = [];
      for (var item in _stockItems) {
        if (item['current_qty'] != item['original_qty']) {
          adjustItems.add({
            'product_id': item['product_id'],
            'quantity': item['current_qty'],
          });
        }
      }

      if (adjustItems.isNotEmpty) {
        final prefs = await SharedPreferences.getInstance();
        final String? profileStr = prefs.getString('cached_employee_profile');
        if (profileStr == null) throw Exception('Profile not found');
        
        final Map<String, dynamic> profile = jsonDecode(profileStr);
        final employeeId = profile['id'] ?? profile['ID'];
        if (employeeId == null) throw Exception('Employee ID not found');

        final apiService = di.sl<SalesApiService>();
        await apiService.adjustSalesStock(employeeId, adjustItems);
      }

      await _db.transaction(() async {
        for (var item in _stockItems) {
          if (item['current_qty'] != item['original_qty']) {
            await (_db.update(_db.salesStock)
                  ..where((t) => t.productId.equals(item['product_id'])))
                .write(SalesStockCompanion(
                    quantity: drift.Value(item['current_qty'])));
          }
        }
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Penyesuaian stok berhasil disimpan.',
                style: GoogleFonts.outfit(color: Colors.white)),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal menyimpan: $e',
                style: GoogleFonts.outfit(color: Colors.white)),
            backgroundColor: Colors.red,
          ),
        );
        setState(() => _isLoading = false);
      }
    }
  }

  void _updateQuantity(int index, int delta) {
    setState(() {
      int newQty = _stockItems[index]['current_qty'] + delta;
      if (newQty < 0) newQty = 0;
      // Sesuai permintaan, fitur ini ditujukan untuk "pengurangan" stok.
      if (newQty > _stockItems[index]['original_qty']) {
        newQty = _stockItems[index]['original_qty'];
      }
      _stockItems[index]['current_qty'] = newQty;
    });
  }

  @override
  Widget build(BuildContext context) {
    final bool hasChanges =
        _stockItems.any((i) => i['current_qty'] != i['original_qty']);

    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded,
              color: Color(0xFF1E293B)),
          onPressed: () => Navigator.pop(context),
        ),
        title: Text(
          'Penyesuaian Stok',
          style: GoogleFonts.outfit(
            color: const Color(0xFF1E293B),
            fontWeight: FontWeight.w800,
            fontSize: 20,
          ),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _stockItems.isEmpty
              ? Center(
                  child: Text(
                    'Tidak ada stok bawaan saat ini.',
                    style: GoogleFonts.outfit(
                      fontSize: 16,
                      color: Colors.blueGrey,
                    ),
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(20),
                  itemCount: _stockItems.length,
                  itemBuilder: (context, index) {
                    final item = _stockItems[index];
                    final diff = item['original_qty'] - item['current_qty'];
                    final bool changed = diff > 0;

                    return Container(
                      margin: const EdgeInsets.only(bottom: 12),
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(
                            color: changed
                                ? Colors.orange.withOpacity(0.5)
                                : Colors.transparent),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.02),
                            blurRadius: 10,
                            offset: const Offset(0, 4),
                          )
                        ],
                      ),
                      child: Row(
                        children: [
                          Container(
                            padding: const EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: changed
                                  ? Colors.orange.withOpacity(0.1)
                                  : Colors.blue.withOpacity(0.1),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: Icon(
                              Icons.inventory_2_rounded,
                              color: changed
                                  ? Colors.orange.shade700
                                  : Colors.blueAccent,
                              size: 24,
                            ),
                          ),
                          const SizedBox(width: 16),
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  item['name'],
                                  style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w700,
                                    fontSize: 15,
                                    color: const Color(0xFF1E293B),
                                  ),
                                ),
                                const SizedBox(height: 4),
                                if (changed)
                                  Text(
                                    'Berkurang: $diff pcs',
                                    style: GoogleFonts.outfit(
                                      fontSize: 12,
                                      fontWeight: FontWeight.w600,
                                      color: Colors.orange.shade700,
                                    ),
                                  )
                                else
                                  Text(
                                    'Stok Awal: ${item['original_qty']}',
                                    style: GoogleFonts.outfit(
                                      fontSize: 12,
                                      color: Colors.blueGrey,
                                    ),
                                  ),
                              ],
                            ),
                          ),
                          Container(
                            decoration: BoxDecoration(
                              color: const Color(0xFFF1F5F9),
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Row(
                              children: [
                                IconButton(
                                  icon: const Icon(Icons.remove_rounded,
                                      size: 20),
                                  color: Colors.red,
                                  onPressed: item['current_qty'] > 0
                                      ? () => _updateQuantity(index, -1)
                                      : null,
                                ),
                                Container(
                                  constraints:
                                      const BoxConstraints(minWidth: 30),
                                  alignment: Alignment.center,
                                  child: Text(
                                    '${item['current_qty']}',
                                    style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w800,
                                      fontSize: 16,
                                    ),
                                  ),
                                ),
                                IconButton(
                                  icon: const Icon(Icons.add_rounded, size: 20),
                                  color: Colors.blue,
                                  onPressed: item['current_qty'] <
                                          item['original_qty']
                                      ? () => _updateQuantity(index, 1)
                                      : null,
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    );
                  },
                ),
      bottomNavigationBar: hasChanges
          ? Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: Colors.white,
                boxShadow: [
                  BoxShadow(
                      color: Colors.black.withOpacity(0.05),
                      blurRadius: 10,
                      offset: const Offset(0, -4))
                ],
              ),
              child: ElevatedButton(
                onPressed: _isLoading ? null : _saveChanges,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF2563EB),
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16),
                  ),
                ),
                child: _isLoading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                          color: Colors.white,
                          strokeWidth: 2,
                        ),
                      )
                    : Text(
                        'Simpan Penyesuaian',
                        style: GoogleFonts.outfit(
                          fontSize: 16,
                          fontWeight: FontWeight.w800,
                          color: Colors.white,
                        ),
                      ),
              ),
            )
          : null,
    );
  }
}
