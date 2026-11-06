import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';

class ProjectionBottomSheet extends StatefulWidget {
  final String storeId;
  final String storeName;
  final SalesApiService salesApi;
  final List<Map<String, dynamic>> initialProjections;
  final Function(List<Map<String, dynamic>>) onSave;

  const ProjectionBottomSheet({
    super.key,
    required this.storeId,
    required this.storeName,
    required this.salesApi,
    required this.initialProjections,
    required this.onSave,
  });

  @override
  State<ProjectionBottomSheet> createState() => _ProjectionBottomSheetState();
}

class _ProjectionBottomSheetState extends State<ProjectionBottomSheet> {
  bool _isLoading = true;
  List<dynamic> _products = [];
  List<dynamic> _filteredProducts = [];
  final Map<String, int> _quantities = {};
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    for (var proj in widget.initialProjections) {
      _quantities[proj['product_id']] = proj['quantity'];
    }
    _fetchProducts();
  }

  Future<void> _fetchProducts() async {
    try {
      final products = await widget.salesApi.getProducts();
      if (mounted) {
        setState(() {
          _products = products;
          _filteredProducts = products;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  void _onSearch(String query) {
    setState(() {
      _searchQuery = query.toLowerCase();
      _filteredProducts = _products.where((p) {
        final name = (p['name'] ?? '').toLowerCase();
        return name.contains(_searchQuery);
      }).toList();
    });
  }

  void _updateQuantity(String productId, int delta) {
    setState(() {
      final current = _quantities[productId] ?? 0;
      final next = current + delta;
      if (next <= 0) {
        _quantities.remove(productId);
      } else {
        _quantities[productId] = next;
      }
    });
  }

  void _handleSave() {
    final List<Map<String, dynamic>> result = [];
    _quantities.forEach((id, qty) {
      result.add({
        'product_id': id,
        'quantity': qty,
      });
    });
    widget.onSave(result);
  }

  @override
  Widget build(BuildContext context) {
    final mq = MediaQuery.of(context);
    return Container(
      height: mq.size.height * 0.8,
      padding: EdgeInsets.only(bottom: mq.viewInsets.bottom),
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Proyeksi Omzet',
                        style: GoogleFonts.outfit(
                            fontSize: 18,
                            fontWeight: FontWeight.w800,
                            color: const Color(0xFF1E293B))),
                    Text(widget.storeName,
                        style: GoogleFonts.outfit(
                            fontSize: 13, color: const Color(0xFF64748B))),
                  ],
                ),
                IconButton(
                  icon: const Icon(Icons.close_rounded),
                  onPressed: () => Navigator.pop(context),
                )
              ],
            ),
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: TextField(
              onChanged: _onSearch,
              decoration: InputDecoration(
                hintText: 'Cari produk...',
                prefixIcon: const Icon(Icons.search_rounded),
                filled: true,
                fillColor: const Color(0xFFF1F5F9),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(12),
                  borderSide: BorderSide.none,
                ),
                contentPadding: const EdgeInsets.symmetric(vertical: 0),
              ),
            ),
          ),
          const SizedBox(height: 8),
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _filteredProducts.length,
                    itemBuilder: (ctx, idx) {
                      final p = _filteredProducts[idx];
                      final pid = p['id'].toString();
                      final qty = _quantities[pid] ?? 0;
                      return Container(
                        margin: const EdgeInsets.only(bottom: 12),
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          border: Border.all(color: const Color(0xFFE2E8F0)),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Row(
                          children: [
                            Expanded(
                              child: Text(
                                p['name'] ?? '',
                                style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w600,
                                    fontSize: 14),
                              ),
                            ),
                            Row(
                              children: [
                                IconButton(
                                  icon: const Icon(Icons.remove_circle_outline),
                                  color: qty > 0 ? Colors.red : Colors.grey,
                                  onPressed: qty > 0
                                      ? () => _updateQuantity(pid, -1)
                                      : null,
                                ),
                                SizedBox(
                                  width: 30,
                                  child: Text(
                                    qty.toString(),
                                    textAlign: TextAlign.center,
                                    style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w800,
                                        fontSize: 16),
                                  ),
                                ),
                                IconButton(
                                  icon: const Icon(Icons.add_circle_outline),
                                  color: Colors.green,
                                  onPressed: () => _updateQuantity(pid, 1),
                                ),
                              ],
                            )
                          ],
                        ),
                      );
                    },
                  ),
          ),
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.05),
                  blurRadius: 10,
                  offset: const Offset(0, -4),
                )
              ],
            ),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _handleSave,
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF10B981),
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12)),
                ),
                child: Text('Simpan Proyeksi',
                    style: GoogleFonts.outfit(
                        fontSize: 16, fontWeight: FontWeight.w700)),
              ),
            ),
          )
        ],
      ),
    );
  }
}
