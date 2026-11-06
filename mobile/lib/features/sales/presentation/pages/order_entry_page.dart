import 'dart:convert';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import 'package:drift/drift.dart' hide Column;
import 'package:shared_preferences/shared_preferences.dart';
import './visit_checkout_page.dart';
import 'package:hris_app/features/sync/data/repositories/sync_repository.dart';
import 'package:hris_app/core/utils/constants.dart';

class OrderEntryPage extends StatefulWidget {
  final StoreModel store;
  final String selfiePath;
  final Uint8List? selfieBytes;
  final String companyId;
  final String companyName;
  final String jobPositionTitle;
  final bool? isDistributor;

  const OrderEntryPage({
    super.key,
    required this.store,
    required this.selfiePath,
    this.selfieBytes,
    required this.companyId,
    required this.companyName,
    required this.jobPositionTitle,
    this.isDistributor,
  });

  @override
  State<OrderEntryPage> createState() => _OrderEntryPageState();
}

class _OrderEntryPageState extends State<OrderEntryPage> {
  final ApiClient apiClient = di.sl<ApiClient>();
  final AppDatabase db = di.sl<AppDatabase>();
  final NumberFormat _currency =
      NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);

  List<dynamic> _allProducts = [];
  List<dynamic> _filteredProducts = [];
  List<String> _categories = ['Semua'];
  String _selectedCategory = 'Semua';

  final Map<String, int> _cart = {};
  final Map<String, TextEditingController> _bulkQtyControllers = {};
  final Map<String, TextEditingController> _looseQtyControllers = {};
  final Map<String, TextEditingController> _bulkUnitNameControllers = {};
  final Map<String, TextEditingController> _baseUnitNameControllers = {};
  final Map<String, TextEditingController> _multiplierControllers = {};
  final Map<String, double> _discountPercent = {};
  final Map<String, TextEditingController> _discountControllers = {};

  // Fitur Harga Fleksibel Khusus Sales Booster & Sales Distributor
  bool _isDistributorProfile = false;
  bool get isSalesBooster =>
      widget.jobPositionTitle.toLowerCase().contains('booster');
  bool get isSalesDistributor =>
      widget.isDistributor == true ||
      _isDistributorProfile ||
      widget.jobPositionTitle.toLowerCase().contains('distributor') ||
      widget.store.customerType.toUpperCase() == 'DISTRIBUTOR';
  bool get hasFlexiblePricing => isSalesBooster || isSalesDistributor;
  final Map<String, String> _selectedPriceTier = {}; // 'NORMAL', 'WBP', 'DBP', 'CUSTOM'
  final Map<String, double> _customPrice = {};

  List<String> _unitOptions = [];

  bool _isLoading = false;
  String _searchQuery = '';
  final TextEditingController _searchCtrl = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadProfile();
    _unitOptions = [
      'PCS',
      'KRT',
      'KARDUS',
      'BAL',
      'KRAT',
      'PACK',
      'LUSIN',
      'BOX',
      'BTL',
      'SCH',
      'CRT'
    ];
    _bulkQtyControllers.clear();
    _looseQtyControllers.clear();
    _bulkUnitNameControllers.clear();
    _baseUnitNameControllers.clear();
    _multiplierControllers.clear();
    _fetchProducts();
  }

  @override
  void dispose() {
    _searchCtrl.dispose();
    for (var ctrl in _bulkQtyControllers.values) {
      ctrl.dispose();
    }
    for (var ctrl in _looseQtyControllers.values) {
      ctrl.dispose();
    }
    for (var ctrl in _bulkUnitNameControllers.values) {
      ctrl.dispose();
    }
    for (var ctrl in _baseUnitNameControllers.values) {
      ctrl.dispose();
    }
    for (var ctrl in _multiplierControllers.values) {
      ctrl.dispose();
    }
    for (var ctrl in _discountControllers.values) {
      ctrl.dispose();
    }
    super.dispose();
  }

  Future<void> _loadProfile() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final cached = prefs.getString('cached_employee_profile');
      if (cached != null) {
        final profile = jsonDecode(cached) as Map<String, dynamic>;
        final isDist = profile['is_distributor'] == true ||
            (profile['distributor_store_id'] != null &&
                profile['distributor_store_id'].toString().isNotEmpty) ||
            (profile['job_position']?['title'] ?? '')
                .toString()
                .toLowerCase()
                .contains('distributor');
        if (mounted && isDist) {
          setState(() {
            _isDistributorProfile = true;
          });
        }
      }
    } catch (_) {}
  }

  Future<Map<String, int>> _computeLocalStockMap() async {
    final stocks = await db.select(db.salesStock).get();
    final pendingTrxs = await (db.select(db.localTransactions)
          ..where((t) => t.syncStatus.equals('pending')))
        .get();
    final pendingLocalIds = pendingTrxs.map((t) => t.localId).toList();
    final Map<String, int> pendingReduction = {};

    if (pendingLocalIds.isNotEmpty) {
      final pendingItems = await (db.select(db.localTransactionItems)
            ..where((t) => t.transactionLocalId.isIn(pendingLocalIds)))
          .get();
      for (var item in pendingItems) {
        pendingReduction[item.productId] =
            (pendingReduction[item.productId] ?? 0) + item.quantity;
      }
    }

    return {
      for (var s in stocks)
        s.productId: (s.quantity - (pendingReduction[s.productId] ?? 0))
            .clamp(0, 999999)
    };
  }

  void _applyProductsToState(List<Map<String, dynamic>> productsData, Map<String, int> stockMap) {
    if (!mounted || productsData.isEmpty) return;
    final cats =
        productsData.map((p) => p['category'].toString()).toSet().toList();
    cats.sort();

    final hasAncillary = productsData.any((p) => _isAncillaryOrGift(
          (p['category'] ?? '').toString().toUpperCase(),
          (p['name'] ?? '').toString().toUpperCase(),
        ));

    final regularCats = cats.where((c) => !_isAncillaryOrGift(
          c.toUpperCase(),
          '',
        )).toList();

    setState(() {
      _allProducts = productsData
          .map((p) => {
                ...p,
                'stock': stockMap[p['id']] ?? 0,
              })
          .toList();
      _categories = [
        'Semua',
        if (hasAncillary) 'Barang Sampingan',
        ...regularCats,
      ];
      _applyFilters();
      _isLoading = false;
    });
  }

  Future<void> _fetchProducts() async {
    setState(() => _isLoading = true);
    try {
      // 1. Render dari local SQLite terlebih dahulu agar instan (< 50ms) saat offline maupun online
      final localProducts = await db.select(db.products).get();
      if (localProducts.isNotEmpty) {
        final stockMap = await _computeLocalStockMap();
        final localData = localProducts
            .where((p) => !_isProcurementProduct(p.category, p.name))
            .map((p) => {
                  'id': p.id,
                  'name': p.name,
                  'selling_price': p.sellingPrice,
                  'dbp': p.dbp,
                  'wbp': p.wbp,
                  'rbp': p.rbp,
                  'cost_price': p.costPrice,
                  'sku': p.sku,
                  'unit': p.unit,
                  'pcs_per_unit': p.pcsPerUnit,
                  'category': _normalizeCategory(p.category),
                  'image_url': p.imageUrl,
                })
            .toList();
        _applyProductsToState(localData, stockMap);
      }

      // 2. Refresh dari server di background jika ada sinyal
      try {
        final res = await di.sl<ApiClient>().client.get('factory/products', queryParameters: {
          'company_id': 'all',
          'scope': 'sales',
          'type': 'finished_goods',
        });
        if (res.statusCode == 200) {
          List? data;
          if (res.data is Map && res.data.containsKey('data')) {
            data = res.data['data'] as List?;
          } else if (res.data is List) {
            data = res.data as List?;
          }
          if (data != null && data.isNotEmpty) {
            final serverProducts = data
                .where((p) => !_isProcurementProduct(
                      (p['category'] ?? p['Category'] ?? '').toString(),
                      (p['name'] ?? p['Name'] ?? '').toString(),
                    ))
                .map((p) => {
                      'id': p['id'] ?? p['ID'] ?? '',
                      'name': p['name'] ?? p['Name'] ?? 'Unknown',
                      'selling_price': (p['selling_price'] ??
                              p['sellingPrice'] ??
                              p['SellingPrice'] ??
                              0 as num)
                          .toDouble(),
                      'dbp': ((p['dbp'] ?? p['DBP'] ?? 0) as num).toDouble(),
                      'wbp': ((p['wbp'] ?? p['WBP'] ?? 0) as num).toDouble(),
                      'rbp': ((p['rbp'] ?? p['RBP'] ?? 0) as num).toDouble(),
                      'cost_price': ((p['cost_price'] ?? p['costPrice'] ?? p['CostPrice'] ?? 0) as num).toDouble(),
                      'sku': p['sku'] ?? p['SKU'],
                      'unit': p['unit'] ?? p['Unit'],
                      'pcs_per_unit': p['pcs_per_unit'] ??
                          p['pcsPerUnit'] ??
                          p['PcsPerUnit'] ??
                          1,
                      'category':
                          _normalizeCategory(p['category'] ?? p['Category']),
                      'image_url': p['image_url'] ?? p['imageUrl'],
                    })
                .toList();

            // Simpan juga ke SQLite untuk pemakaian offline berikutnya
            await db.batch((batch) {
              for (final p in serverProducts) {
                batch.insert(
                  db.products,
                  ProductsCompanion.insert(
                    id: p['id'].toString(),
                    companyId: '',
                    name: p['name'].toString(),
                    sellingPrice: (p['selling_price'] as num).toDouble(),
                    dbp: Value(((p['dbp'] ?? 0) as num).toDouble()),
                    wbp: Value(((p['wbp'] ?? 0) as num).toDouble()),
                    rbp: Value(((p['rbp'] ?? 0) as num).toDouble()),
                    costPrice: Value(((p['cost_price'] ?? 0) as num).toDouble()),
                    sku: Value(p['sku']?.toString()),
                    unit: Value(p['unit']?.toString()),
                    pcsPerUnit: Value((p['pcs_per_unit'] as num?)?.toInt() ?? 1),
                    category: Value(p['category']?.toString()),
                    imageUrl: Value(p['image_url']?.toString()),
                  ),
                  mode: InsertMode.insertOrReplace,
                );
              }
            });

            try {
              await di.sl<SyncRepository>().pullSalesStock();
            } catch (_) {}

            final stockMap = await _computeLocalStockMap();
            _applyProductsToState(serverProducts, stockMap);
          }
        }
      } catch (e) {
        debugPrint('DEBUG: Fetch products from server offline/failed: $e');
      }
    } catch (e) {
      if (mounted && _allProducts.isEmpty) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Gagal memuat produk: $e')));
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  bool _isAncillaryOrGift(String cat, String n) {
    return cat.contains('SAMPINGAN') ||
        cat.contains('HADIAH') ||
        cat.contains('MERCHANDISE') ||
        cat.contains('PROMOSI') ||
        cat.contains('MANGKOK') ||
        cat.contains('PIRING') ||
        cat.contains('GELAS') ||
        n.contains('MANGKOK') ||
        n.contains('PIRING') ||
        n.contains('GELAS') ||
        n.contains('SENDOK') ||
        n.contains('GARPU');
  }

  bool _isProcurementProduct(dynamic category, dynamic name) {
    final cat = (category ?? '').toString().trim().toUpperCase();
    final n = (name ?? '').toString().trim().toUpperCase();

    // Kategori pengadaan harus menang meskipun nama cocok dengan barang hadiah.
    if (cat.contains('BAHAN') ||
        cat.contains('RAW') ||
        cat.contains('BUMBU') ||
        cat.contains('KEMASAN') ||
        cat.contains('PACKAGING') ||
        cat.contains('WADAH') ||
        n.startsWith('STIKER ') ||
        n.startsWith('STICKER ')) {
      return true;
    }

    // Hadiah, Merchandise, Barang Sampingan are commercial / sellable goods
    if (_isAncillaryOrGift(cat, n)) {
      return false;
    }

    // 1. Raw materials & ingredients
    if (cat.contains('BAHAN') ||
        cat.contains('RAW') ||
        cat.contains('BUMBU') ||
        cat.contains('KEDELAI') ||
        cat.contains('GULA') ||
        cat.contains('GARAM') ||
        n.contains('KEDELAI') ||
        n.contains('GULA') ||
        n.contains('GARAM') ||
        n.contains('TEPUNG')) {
      return true;
    }

    // 2. Packaging / Containers
    if (cat.contains('KEMASAN') ||
        cat.contains('BOTOL') ||
        cat.contains('WADAH') ||
        cat.contains('KARTON') ||
        cat.contains('DUS') ||
        cat.contains('LABEL') ||
        cat.contains('TUTUP') ||
        cat.contains('PACKAGING') ||
        cat.contains('JERIGEN') ||
        n.contains('BOTOL KOSONG') ||
        n.contains('KARTON') ||
        n.contains('TUTUP BOTOL') ||
        n.contains('LABEL')) {
      return true;
    }

    // 3. Trading / Pelengkap
    if (cat.contains('TRADING') ||
        cat.contains('PRODUK_PELENGKAP')) {
      return true;
    }

    return false;
  }

  String _normalizeCategory(dynamic cat) {
    if (cat == null) return 'Uncategorized';
    String s = cat.toString().trim();
    if (s.isEmpty) return 'Uncategorized';
    // Return title cased string to prevent duplicates like "Barang sampingan" vs "Barang Sampingan"
    return s
        .split(' ')
        .map((w) => w.isNotEmpty
            ? '${w[0].toUpperCase()}${w.substring(1).toLowerCase()}'
            : '')
        .join(' ');
  }

  void _applyFilters() {
    setState(() {
      _filteredProducts = _allProducts.where((p) {
        final matchesSearch = p['name']
                .toString()
                .toLowerCase()
                .contains(_searchQuery.toLowerCase()) ||
            p['sku']
                .toString()
                .toLowerCase()
                .contains(_searchQuery.toLowerCase());

        bool matchesCategory;
        if (_selectedCategory == 'Semua') {
          matchesCategory = true;
        } else if (_selectedCategory == 'Barang Sampingan') {
          // Chip virtual: tampilkan semua produk ancillary/gift terlepas dari kategori DB-nya
          // (Hadiah, Merchandise, Mangkok, Piring, Gelas, dll. semua muncul di sini)
          matchesCategory = _isAncillaryOrGift(
            (p['category'] ?? '').toString().toUpperCase(),
            (p['name'] ?? '').toString().toUpperCase(),
          );
        } else {
          matchesCategory = p['category'] == _selectedCategory;
        }

        return matchesSearch && matchesCategory;
      }).toList();

      // Prioritize items carried by sales (stock > 0) at the top
      _filteredProducts.sort((a, b) {
        final stockA = a['stock'] as int? ?? 0;
        final stockB = b['stock'] as int? ?? 0;
        if (stockA > 0 && stockB == 0) return -1;
        if (stockA == 0 && stockB > 0) return 1;
        return a['name'].toString().compareTo(b['name'].toString());
      });
    });
  }

  Future<void> _loadLastOrder() async {
    setState(() => _isLoading = true);
    try {
      final lastTxn = await (db.select(db.localTransactions)
            ..where((t) => t.storeId.equals(widget.store.id))
            ..where((t) => t.companyId.equals(widget.companyId))
            ..orderBy([
              (t) =>
                  OrderingTerm(expression: t.createdAt, mode: OrderingMode.desc)
            ])
            ..limit(1))
          .getSingleOrNull();

      if (lastTxn != null) {
        final items = await (db.select(db.localTransactionItems)
              ..where((t) => t.transactionLocalId.equals(lastTxn.localId)))
            .get();

        if (mounted) {
          setState(() {
            _cart.clear();
            for (final item in items) {
              final id = item.productId;
              final qty = item.quantity;
              final mult = item.piecesPerUnit > 1 ? item.piecesPerUnit : 1;
              _cart[id] = qty;

              if (_bulkQtyControllers.containsKey(id)) {
                final bulkVal = mult > 1 ? (qty ~/ mult) : 0;
                _bulkQtyControllers[id]!.text = bulkVal.toString();
              }
              if (_looseQtyControllers.containsKey(id)) {
                final looseVal = mult > 1 ? (qty % mult) : qty;
                _looseQtyControllers[id]!.text = looseVal.toString();
              }
              if (_bulkUnitNameControllers.containsKey(id)) {
                _bulkUnitNameControllers[id]!.text = item.unit;
              }
              if (_multiplierControllers.containsKey(id)) {
                _multiplierControllers[id]!.text = mult.toString();
              }
            }
          });
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
            content: Text('✓ Berhasil menyalin pesanan terakhir!'),
            backgroundColor: Colors.green,
          ));
        }
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
              content: Text('Belum ada riwayat pesanan untuk toko ini.')));
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Gagal memuat riwayat: $e')));
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _openScanner() {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => Container(
        height: MediaQuery.of(context).size.height * 0.7,
        decoration: const BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.only(
              topLeft: Radius.circular(32), topRight: Radius.circular(32)),
        ),
        child: Column(
          children: [
            const SizedBox(height: 12),
            Container(
                width: 40,
                height: 4,
                decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(2))),
            const SizedBox(height: 20),
            Text('Scan Barcode Produk',
                style: GoogleFonts.outfit(
                    fontSize: 18, fontWeight: FontWeight.w800)),
            const SizedBox(height: 20),
            Expanded(
              child: MobileScanner(
                onDetect: (capture) {
                  final List<Barcode> barcodes = capture.barcodes;
                  if (barcodes.isNotEmpty) {
                    final code = barcodes.first.rawValue;
                    if (code != null) {
                      Navigator.pop(context);
                      setState(() {
                        _searchCtrl.text = code;
                        _searchQuery = code;
                        _applyFilters();
                      });
                    }
                  }
                },
              ),
            ),
            const SizedBox(height: 40),
          ],
        ),
      ),
    );
  }

  double _getBasePriceForProduct(Map<String, dynamic> product) {
    final id = product['id']?.toString() ?? '';
    if (!hasFlexiblePricing) {
      return ((product['selling_price'] ?? 0) as num).toDouble();
    }

    if (isSalesDistributor) {
      // Khusus distributor: hanya gunakan custom price jika sudah diisi, jangan bocorkan harga pabrik
      if (_customPrice.containsKey(id) && _customPrice[id]! > 0) {
        return _customPrice[id]!;
      }
      return 0.0;
    }

    final tier = _selectedPriceTier[id] ?? 'NORMAL';
    if (tier == 'DBP') {
      final dbp = ((product['dbp'] ?? 0) as num).toDouble();
      if (dbp > 0) return dbp;
    } else if (tier == 'WBP') {
      final wbp = ((product['wbp'] ?? 0) as num).toDouble();
      if (wbp > 0) return wbp;
    } else if (tier == 'CUSTOM') {
      if (_customPrice.containsKey(id) && _customPrice[id]! > 0) {
        return _customPrice[id]!;
      }
    }
    return ((product['selling_price'] ?? 0) as num).toDouble();
  }

  String _getPriceTierLabel(Map<String, dynamic> product) {
    final id = product['id']?.toString() ?? '';
    if (isSalesDistributor) {
      return 'Harga Jual';
    }
    final tier = _selectedPriceTier[id] ?? 'NORMAL';
    switch (tier) {
      case 'DBP':
        return 'Distributor (DBP)';
      case 'WBP':
        return 'Grosir (WBP)';
      case 'CUSTOM':
        return 'Custom Booster';
      default:
        return 'Normal';
    }
  }

  void _showCustomPriceDialog(Map<String, dynamic> p) {
    final id = p['id']?.toString() ?? '';
    final isDist = isSalesDistributor;
    final currentCustom = _customPrice[id] ?? (isDist ? 0.0 : _getBasePriceForProduct(p));
    final normalPrice = ((p['selling_price'] ?? 0) as num).toDouble();
    final dbpPrice = ((p['dbp'] ?? 0) as num).toDouble();
    final wbpPrice = ((p['wbp'] ?? 0) as num).toDouble();
    final costPrice = ((p['cost_price'] ?? 0) as num).toDouble();

    final primaryColor = isDist ? const Color(0xFF4338CA) : Colors.purple;
    final bgShade = isDist ? Colors.indigo.shade50 : Colors.purple.shade50;

    // Batas bawah harga: DBP atau Cost Price (HPP) khusus booster
    final double floorPrice = dbpPrice > 0 ? dbpPrice : (costPrice > 0 ? costPrice : 0.0);

    final TextEditingController priceCtrl = TextEditingController(
      text: currentCustom > 0 ? currentCustom.toStringAsFixed(0) : '',
    );
    String? errorText;

    showDialog(
      context: context,
      builder: (ctx) {
        return StatefulBuilder(
          builder: (context, setDialogState) {
            return AlertDialog(
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(20)),
              title: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: bgShade,
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Icon(
                        isDist ? Icons.storefront_rounded : Icons.edit_note_rounded,
                        color: primaryColor,
                        size: 22),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          isDist ? 'Atur Harga Jual' : 'Harga Fleksibel Booster',
                          style: GoogleFonts.outfit(
                            fontSize: 16,
                            fontWeight: FontWeight.w800,
                            color: const Color(0xFF1E293B),
                          ),
                        ),
                        Text(
                          p['name'] ?? 'Produk',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: GoogleFonts.outfit(
                            fontSize: 12,
                            color: Colors.grey.shade600,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (!isDist) ...[
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: Colors.grey.shade200),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Referensi Harga Master:',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                              color: Colors.blueGrey,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text('Normal (HET):', style: GoogleFonts.outfit(fontSize: 11)),
                              Text(_currency.format(normalPrice), style: GoogleFonts.outfit(fontSize: 11, fontWeight: FontWeight.w700)),
                            ],
                          ),
                          if (wbpPrice > 0)
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text('Grosir (WBP):', style: GoogleFonts.outfit(fontSize: 11)),
                                Text(_currency.format(wbpPrice), style: GoogleFonts.outfit(fontSize: 11, fontWeight: FontWeight.w700, color: Colors.blue.shade700)),
                              ],
                            ),
                          if (dbpPrice > 0)
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text('Distributor (DBP):', style: GoogleFonts.outfit(fontSize: 11)),
                                Text(_currency.format(dbpPrice), style: GoogleFonts.outfit(fontSize: 11, fontWeight: FontWeight.w700, color: Colors.green.shade700)),
                              ],
                            ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 16),
                  ],
                  Text(
                    isDist
                        ? 'Tentukan Harga Jual Satuan (Rp):'
                        : 'Masukkan Harga Satuan Baru (Rp):',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: const Color(0xFF1E293B),
                    ),
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: priceCtrl,
                    keyboardType: TextInputType.number,
                    autofocus: true,
                    style: GoogleFonts.outfit(
                      fontSize: 16,
                      fontWeight: FontWeight.w800,
                      color: const Color(0xFF1E293B),
                    ),
                    decoration: InputDecoration(
                      prefixText: 'Rp ',
                      errorText: errorText,
                      filled: true,
                      fillColor: const Color(0xFFF1F5F9),
                      contentPadding: const EdgeInsets.symmetric(
                          horizontal: 16, vertical: 12),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: BorderSide.none,
                      ),
                    ),
                    onChanged: (val) {
                      final entered = double.tryParse(val.replaceAll(RegExp(r'[^0-9]'), '')) ?? 0;
                      if (entered <= 0) {
                        setDialogState(() {
                          errorText = 'Nominal harga harus lebih dari 0';
                        });
                      } else if (!isDist && floorPrice > 0 && entered < floorPrice) {
                        setDialogState(() {
                          errorText = 'Min. ${_currency.format(floorPrice)} (Batas Modal)';
                        });
                      } else {
                        setDialogState(() {
                          errorText = null;
                        });
                      }
                    },
                  ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(ctx),
                  child: Text(
                    'Batal',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w600,
                      color: Colors.grey.shade600,
                    ),
                  ),
                ),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: primaryColor,
                    shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10)),
                  ),
                  onPressed: () {
                    final entered = double.tryParse(priceCtrl.text.replaceAll(RegExp(r'[^0-9]'), '')) ?? 0;
                    if (entered <= 0) {
                      setDialogState(() {
                        errorText = 'Nominal harga harus lebih dari 0';
                      });
                      return;
                    }
                    if (!isDist && floorPrice > 0 && entered < floorPrice) {
                      setDialogState(() {
                        errorText = 'Min. ${_currency.format(floorPrice)} (Batas Modal)';
                      });
                      return;
                    }
                    setState(() {
                      _customPrice[id] = entered;
                      _selectedPriceTier[id] = 'CUSTOM';
                    });
                    Navigator.pop(ctx);
                  },
                  child: Text(
                    'Terapkan',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w700,
                      color: Colors.white,
                    ),
                  ),
                ),
              ],
            );
          },
        );
      },
    );
  }

  Widget _buildPriceTierRow(Map<String, dynamic> p) {
    if (!hasFlexiblePricing) return const SizedBox.shrink();

    final id = p['id']?.toString() ?? '';
    final qty = _cart[id] ?? 0;
    if (qty <= 0) return const SizedBox.shrink();

    final customPrice = _customPrice[id] ?? 0.0;

    // KHUSUS SALES DISTRIBUTOR (LUAR PULAU): Sembunyikan harga asli pabrik (Normal, WBP, DBP)
    if (isSalesDistributor) {
      return Container(
        margin: const EdgeInsets.only(bottom: 10),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: Colors.indigo.shade50.withOpacity(0.5),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.indigo.shade100),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Row(
              children: [
                Icon(Icons.storefront_rounded,
                    size: 16, color: Colors.indigo.shade700),
                const SizedBox(width: 6),
                Text(
                  customPrice > 0
                      ? 'Harga Jual: ${_currency.format(customPrice)}'
                      : 'Harga Jual Belum Diatur',
                  style: GoogleFonts.outfit(
                    fontSize: 11.5,
                    fontWeight: FontWeight.w800,
                    color: customPrice > 0
                        ? Colors.indigo.shade900
                        : Colors.red.shade700,
                  ),
                ),
              ],
            ),
            _buildTierChip(
              label: customPrice > 0 ? 'Ubah Harga ✎' : 'Atur Harga ✎',
              isSelected: true,
              isDistributor: true,
              onTap: () => _showCustomPriceDialog(p),
            ),
          ],
        ),
      );
    }

    final selectedTier = _selectedPriceTier[id] ?? 'NORMAL';
    final normalPrice = ((p['selling_price'] ?? 0) as num).toDouble();
    final wbpPrice = ((p['wbp'] ?? 0) as num).toDouble();
    final dbpPrice = ((p['dbp'] ?? 0) as num).toDouble();

    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.purple.shade50.withOpacity(0.5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.purple.shade100),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                Icons.bolt_rounded,
                size: 16,
                color: Colors.purple.shade700,
              ),
              const SizedBox(width: 4),
              Text(
                'Pilihan Harga Booster:',
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  color: Colors.purple.shade900,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          SingleChildScrollView(
            scrollDirection: Axis.horizontal,
            child: Row(
              children: [
                _buildTierChip(
                  label: 'Normal: ${_currency.format(normalPrice)}',
                  isSelected: selectedTier == 'NORMAL',
                  isDistributor: false,
                  onTap: () {
                    setState(() {
                      _selectedPriceTier[id] = 'NORMAL';
                    });
                  },
                ),
                const SizedBox(width: 6),
                if (wbpPrice > 0) ...[
                  _buildTierChip(
                    label: 'Grosir (WBP): ${_currency.format(wbpPrice)}',
                    isSelected: selectedTier == 'WBP',
                    isDistributor: false,
                    onTap: () {
                      setState(() {
                        _selectedPriceTier[id] = 'WBP';
                      });
                    },
                  ),
                  const SizedBox(width: 6),
                ],
                if (dbpPrice > 0) ...[
                  _buildTierChip(
                    label: 'Distributor (DBP): ${_currency.format(dbpPrice)}',
                    isSelected: selectedTier == 'DBP',
                    isDistributor: false,
                    onTap: () {
                      setState(() {
                        _selectedPriceTier[id] = 'DBP';
                      });
                    },
                  ),
                  const SizedBox(width: 6),
                ],
                _buildTierChip(
                  label: selectedTier == 'CUSTOM' && customPrice > 0
                      ? 'Custom: ${_currency.format(customPrice)} ✎'
                      : 'Custom... ✎',
                  isSelected: selectedTier == 'CUSTOM',
                  isDistributor: false,
                  onTap: () {
                    _showCustomPriceDialog(p);
                  },
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTierChip({
    required String label,
    required bool isSelected,
    required VoidCallback onTap,
    bool isDistributor = false,
  }) {
    final activeBg = isDistributor ? const Color(0xFF4338CA) : Colors.purple.shade700;
    final activeBorder = isDistributor ? const Color(0xFF3730A3) : Colors.purple.shade800;
    final inactiveBorder = isDistributor ? Colors.indigo.shade200 : Colors.purple.shade200;
    final inactiveText = isDistributor ? Colors.indigo.shade900 : Colors.purple.shade900;

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
        decoration: BoxDecoration(
          color: isSelected ? activeBg : Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: isSelected ? activeBorder : inactiveBorder,
          ),
        ),
        child: Text(
          label,
          style: GoogleFonts.outfit(
            fontSize: 10.5,
            fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
            color: isSelected ? Colors.white : inactiveText,
          ),
        ),
      ),
    );
  }

  double get _totalAmount {
    double total = 0;
    _cart.forEach((id, qty) {
      final product = _allProducts.firstWhere((p) => p['id'] == id,
          orElse: () => <String, dynamic>{});
      if (product.isNotEmpty) {
        final basePrice = _getBasePriceForProduct(product);
        final disc = _discountPercent[id] ?? 0.0;
        final effectivePrice = basePrice * (1 - disc / 100);
        total += effectivePrice * qty;
      }
    });
    return total;
  }

  double get _totalDiscount {
    double totalSaved = 0;
    _cart.forEach((id, qty) {
      final product = _allProducts.firstWhere((p) => p['id'] == id,
          orElse: () => <String, dynamic>{});
      if (product.isNotEmpty) {
        final basePrice = _getBasePriceForProduct(product);
        final disc = _discountPercent[id] ?? 0.0;
        if (disc > 0) {
          totalSaved += (basePrice * disc / 100) * qty;
        }
      }
    });
    return totalSaved;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded,
              color: Color(0xFF334155)),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Isi Order Pelanggan',
                style: GoogleFonts.outfit(
                    fontSize: 14,
                    fontWeight: FontWeight.w700,
                    color: const Color(0xFF1E293B))),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh_rounded, size: 16, color: Color(0xFF10B981)),
            onPressed: () {
              _fetchProducts();
            },
          ),
          TextButton.icon(
            onPressed: _loadLastOrder,
            icon: const Icon(Icons.history_rounded,
                size: 14, color: Color(0xFF10B981)),
            label: Text('Repeat Order',
                style: GoogleFonts.outfit(
                    fontSize: 12,
                    fontWeight: FontWeight.w700,
                    color: const Color(0xFF10B981))),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: Column(
        children: [
          _buildSearchBar(),
          _buildCategoryFilter(),
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : CustomScrollView(
                    slivers: [
                      SliverToBoxAdapter(child: _buildProductListBody()),
                      const SliverToBoxAdapter(child: SizedBox(height: 16)),
                    ],
                  ),
          ),
          _buildBottomSummary(),
        ],
      ),
    );
  }

  Widget _buildSearchBar() {
    return Container(
      padding: const EdgeInsets.fromLTRB(16, 6, 16, 6),
      color: Colors.white,
      child: Row(
        children: [
          Expanded(
            child: TextField(
              controller: _searchCtrl,
              onChanged: (val) {
                _searchQuery = val;
                _applyFilters();
              },
              decoration: InputDecoration(
                hintText: 'Cari nama produk atau SKU...',
                hintStyle: GoogleFonts.outfit(color: Colors.grey, fontSize: 14),
                prefixIcon:
                    const Icon(Icons.search_rounded, color: Color(0xFF10B981)),
                filled: true,
                fillColor: const Color(0xFFF1F5F9),
                border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(16),
                    borderSide: BorderSide.none),
                contentPadding: const EdgeInsets.symmetric(vertical: 0),
              ),
            ),
          ),
          const SizedBox(width: 12),
          GestureDetector(
            onTap: _openScanner,
            child: Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                  color: const Color(0xFF10B981).withOpacity(0.1),
                  borderRadius: BorderRadius.circular(16)),
              child: const Icon(Icons.qr_code_scanner_rounded,
                  color: Color(0xFF10B981)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCategoryFilter() {
    return Container(
      height: 50,
      color: Colors.white,
      child: ListView.builder(
        scrollDirection: Axis.horizontal,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        itemCount: _categories.length,
        itemBuilder: (context, index) {
          final cat = _categories[index];
          final isSelected = _selectedCategory == cat;
          return Padding(
            padding: const EdgeInsets.only(right: 8, bottom: 8),
            child: ChoiceChip(
              label: Text(cat,
                  style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight:
                          isSelected ? FontWeight.w800 : FontWeight.w500,
                      color: isSelected ? Colors.white : Colors.blueGrey)),
              selected: isSelected,
              onSelected: (selected) {
                setState(() {
                  _selectedCategory = cat;
                  _applyFilters();
                });
              },
              selectedColor: const Color(0xFF10B981),
              backgroundColor: const Color(0xFFF1F5F9),
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12)),
              side: BorderSide.none,
              showCheckmark: false,
            ),
          );
        },
      ),
    );
  }

  Widget _buildProductImage(Map<String, dynamic> p) {
    final rawUrl = p['image_url'] as String?;
    String imageUrl = '';
    if (rawUrl != null && rawUrl.isNotEmpty) {
      if (rawUrl.startsWith('http') || rawUrl.startsWith('assets/')) {
        imageUrl = rawUrl;
      } else {
        imageUrl = '${AppConstants.baseUrl.replaceAll('/api/v1/', '')}$rawUrl';
      }
    }

    if (imageUrl.isNotEmpty) {
      if (imageUrl.startsWith('http')) {
        return CachedNetworkImage(imageUrl: imageUrl, fit: BoxFit.contain, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,);
      } else if (imageUrl.startsWith('assets/')) {
        return Image.asset(imageUrl, fit: BoxFit.contain);
      }
    }


    return _buildPlaceholder();
  }

  Widget _buildPlaceholder() {
    return Container(
      color: const Color(0xFFF8FAFC),
      alignment: Alignment.center,
      child: const Icon(Icons.image_not_supported_outlined,
          size: 20, color: Colors.blueGrey),
    );
  }

  Widget _buildProductListBody() {
    if (_filteredProducts.isEmpty) {
      return Padding(
        padding: const EdgeInsets.symmetric(vertical: 48),
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.search_off_rounded,
                  size: 64, color: Color(0xFFE2E8F0)),
              const SizedBox(height: 16),
              Text('Produk tidak ditemukan',
                  style: GoogleFonts.outfit(color: const Color(0xFF94A3B8))),
            ],
          ),
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      itemCount: _filteredProducts.length,
      itemBuilder: (context, index) {
        final p = _filteredProducts[index];
        final id = p['id'];
        final qty = _cart[id] ?? 0;

        return Container(
          margin: const EdgeInsets.only(bottom: 10),
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              boxShadow: [
                BoxShadow(
                    color: Colors.black.withOpacity(0.04),
                    blurRadius: 12,
                    offset: const Offset(0, 4))
              ]),
          child: Column(
            children: [
              Row(
                children: [
                  Container(
                    width: 60,
                    height: 60,
                    decoration: BoxDecoration(
                      color: const Color(0xFFF8FAFC),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(color: const Color(0xFFE2E8F0)),
                    ),
                    padding: const EdgeInsets.all(4),
                    clipBehavior: Clip.antiAlias,
                    child: _buildProductImage(p),
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Builder(builder: (context) {
                          final isSampingan = _isAncillaryOrGift(
                            (p['category'] ?? '').toString().toUpperCase(),
                            (p['name'] ?? '').toString().toUpperCase(),
                          );
                          return Row(
                            children: [
                              Expanded(
                                child: Text(p['name'] ?? 'Unknown Product',
                                    style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w900,
                                        fontSize: 14,
                                        color: const Color(0xFF1E293B))),
                              ),
                              if (isSampingan)
                                Container(
                                  margin: const EdgeInsets.only(left: 6),
                                  padding: const EdgeInsets.symmetric(
                                      horizontal: 6, vertical: 2),
                                  decoration: BoxDecoration(
                                    color: Colors.amber.shade50,
                                    borderRadius: BorderRadius.circular(6),
                                    border: Border.all(
                                        color: Colors.amber.shade300,
                                        width: 0.8),
                                  ),
                                  child: Text(
                                    'Sampingan',
                                    style: GoogleFonts.outfit(
                                      fontSize: 9,
                                      fontWeight: FontWeight.w800,
                                      color: Colors.amber.shade900,
                                    ),
                                  ),
                                ),
                            ],
                          );
                        }),
                        Text('SKU: ${p['sku'] ?? '-'} • ${p['category']}',
                            style: GoogleFonts.outfit(
                                fontSize: 9,
                                color: Colors.blueGrey,
                                fontWeight: FontWeight.w500)),
                      ],
                    ),
                  ),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      // Tampilkan harga asli dan harga diskon jika ada diskon
                      Builder(builder: (context) {
                        if (isSalesDistributor) {
                          final customPrice = _customPrice[id] ?? 0.0;
                          return Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              if (customPrice > 0) ...[
                                Text(_currency.format(customPrice),
                                    style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w900,
                                        color: const Color(0xFF4338CA),
                                        fontSize: 16)),
                                Container(
                                  margin: const EdgeInsets.only(top: 2),
                                  padding: const EdgeInsets.symmetric(
                                      horizontal: 6, vertical: 1.5),
                                  decoration: BoxDecoration(
                                    color: Colors.indigo.shade50,
                                    borderRadius: BorderRadius.circular(6),
                                    border: Border.all(
                                        color: Colors.indigo.shade200, width: 0.5),
                                  ),
                                  child: Text(
                                    'Harga Jual',
                                    style: GoogleFonts.outfit(
                                      fontSize: 9.5,
                                      fontWeight: FontWeight.w800,
                                      color: Colors.indigo.shade800,
                                    ),
                                  ),
                                ),
                              ] else ...[
                                GestureDetector(
                                  onTap: () => _showCustomPriceDialog(p),
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(
                                        horizontal: 8, vertical: 4),
                                    decoration: BoxDecoration(
                                      color: Colors.amber.shade50,
                                      borderRadius: BorderRadius.circular(8),
                                      border: Border.all(color: Colors.amber.shade300),
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(Icons.edit_note_rounded,
                                            size: 14, color: Colors.amber.shade900),
                                        const SizedBox(width: 4),
                                        Text(
                                          'Atur Harga',
                                          style: GoogleFonts.outfit(
                                            fontSize: 12,
                                            fontWeight: FontWeight.w800,
                                            color: Colors.amber.shade900,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ],
                            ],
                          );
                        }

                        final disc = _discountPercent[id] ?? 0.0;
                        final basePrice = _getBasePriceForProduct(p);
                        final discountedPrice =
                            basePrice * (1 - disc / 100);
                        final tierLabel = _getPriceTierLabel(p);
                        final hasCustomTier = hasFlexiblePricing &&
                            (_selectedPriceTier[id] ?? 'NORMAL') != 'NORMAL';
                        final badgeBg = Colors.purple.shade50;
                        final badgeBorder = Colors.purple.shade200;
                        final badgeText = Colors.purple.shade800;
                        final priceColor = hasCustomTier
                            ? Colors.purple.shade700
                            : const Color(0xFF10B981);

                        return Column(
                          crossAxisAlignment: CrossAxisAlignment.end,
                          children: [
                            if (disc > 0) ...[
                              Text(_currency.format(basePrice),
                                  style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w600,
                                      color: Colors.blueGrey,
                                      fontSize: 12,
                                      decoration: TextDecoration.lineThrough)),
                              Text(_currency.format(discountedPrice),
                                  style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w900,
                                      color: const Color(0xFFEF4444),
                                      fontSize: 15)),
                            ] else
                              Text(_currency.format(basePrice),
                                  style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w900,
                                      color: priceColor,
                                      fontSize: 16)),
                            if (hasCustomTier)
                              Container(
                                margin: const EdgeInsets.only(top: 2),
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 6, vertical: 1.5),
                                decoration: BoxDecoration(
                                  color: badgeBg,
                                  borderRadius: BorderRadius.circular(6),
                                  border: Border.all(
                                      color: badgeBorder, width: 0.5),
                                ),
                                child: Text(
                                  tierLabel,
                                  style: GoogleFonts.outfit(
                                    fontSize: 9.5,
                                    fontWeight: FontWeight.w800,
                                    color: badgeText,
                                  ),
                                ),
                              ),
                          ],
                        );
                      }),
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 8, vertical: 2),
                        decoration: BoxDecoration(
                            color: (p['stock'] as int) > 0
                                ? Colors.green.withOpacity(0.1)
                                : Colors.orange.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(6)),
                        child: Text(
                            (p['stock'] as int) > 0
                                ? 'Stok: ${p['stock']}'
                                : 'Katalog',
                            style: GoogleFonts.outfit(
                                fontSize: 10,
                                fontWeight: FontWeight.w800,
                                color: (p['stock'] as int) > 0
                                    ? Colors.green
                                    : Colors.orange.shade800)),
                      ),
                    ],
                  ),
                ],
              ),
              if (hasFlexiblePricing && qty > 0) ...[
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 8),
                  child: Divider(height: 1, color: Color(0xFFF1F5F9)),
                ),
                _buildPriceTierRow(p),
              ],
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 12),
                child: Divider(height: 1, color: Color(0xFFF1F5F9)),
              ),
              // Baris diskon per produk
              _buildDiscountRow(productId: id),
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 12),
                child: Divider(height: 1, color: Color(0xFFF1F5F9)),
              ),
              _buildDualInputRow(
                productId: id,
                qty: qty,
                p: p,
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildDualInputRow({
    required String productId,
    required int qty,
    required Map<String, dynamic> p,
  }) {
    final defaultMultiplier = p['pcs_per_unit'] ?? 1;
    final multiplierCtrl = _getMultiplierCtrl(productId, defaultMultiplier);
    final multiplier = int.tryParse(multiplierCtrl.text) ?? 1;

    // Also initialize other controllers for the checkout payload
    _getBulkUnitNameCtrl(productId, p['unit']);
    _getBaseUnitNameCtrl(productId);

    final bulkUnit = p['unit'] ?? 'KRT';

    final looseCtrlText = _looseQtyControllers[productId]?.text ?? '';
    final bulkCtrlText = _bulkQtyControllers[productId]?.text ?? '';

    bool showTambahButton = qty == 0;
    if (showTambahButton) {
      if (_looseQtyControllers.containsKey(productId)) {
        final text = _looseQtyControllers[productId]!.text.trim();
        if (text != '0') showTambahButton = false;
      }
      if (_bulkQtyControllers.containsKey(productId)) {
        final text = _bulkQtyControllers[productId]!.text.trim();
        if (text != '0') showTambahButton = false;
      }
    }

    if (showTambahButton) {
      return Align(
        alignment: Alignment.centerRight,
        child: SizedBox(
          height: 36,
          child: ElevatedButton.icon(
            onPressed: () {
              setState(() {
                final defQty = multiplier > 1 ? multiplier : 1;
                _cart[productId] = defQty;
                _getBulkQtyCtrl(productId, defQty, multiplier).text =
                    multiplier > 1 ? '1' : '0';
                _getLooseQtyCtrl(productId, defQty, multiplier).text =
                    multiplier > 1 ? '0' : '1';
              });
            },
            icon: const Icon(Icons.add_rounded, size: 16, color: Colors.white),
            label: Text('TAMBAH',
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w900,
                    fontSize: 12,
                    color: Colors.white,
                    letterSpacing: 0.5)),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF10B981),
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12)),
              elevation: 0,
              padding: const EdgeInsets.symmetric(horizontal: 16),
            ),
          ),
        ),
      );
    }

    final looseText = _looseQtyControllers[productId]
            ?.text
            .replaceAll(' ', '')
            .toLowerCase() ??
        '';
    final bool hasSmartInput =
        looseText.contains('x') || looseText.contains('*');
    int smartA = 0, smartB = 0;
    if (hasSmartInput) {
      final separator = looseText.contains('x') ? 'x' : '*';
      final parts = looseText.split(separator);
      if (parts.length == 2) {
        smartA = int.tryParse(parts[0]) ?? 0;
        smartB = int.tryParse(parts[1]) ?? 0;
      }
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            if (multiplier > 1) ...[
              Expanded(
                flex: 5,
                child: Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                  decoration: BoxDecoration(
                    color: const Color(0xFF10B981).withOpacity(0.04),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                        color: const Color(0xFF10B981).withOpacity(0.1)),
                  ),
                  child: Row(
                    children: [
                      _buildMiniBtn(
                          Icons.remove,
                          () =>
                              _updateDualQty(productId, -1, true, multiplier)),
                      Expanded(
                        child: TextField(
                          controller:
                              _getBulkQtyCtrl(productId, qty, multiplier),
                          keyboardType: TextInputType.number,
                          textAlign: TextAlign.center,
                          onChanged: (val) => _onQtyTyped(productId),
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w900,
                              fontSize: 14,
                              color: const Color(0xFF1E293B)),
                          decoration: const InputDecoration(
                              isDense: true,
                              contentPadding: EdgeInsets.zero,
                              border: InputBorder.none),
                        ),
                      ),
                      Text(bulkUnit,
                          style: GoogleFonts.outfit(
                              fontSize: 10,
                              fontWeight: FontWeight.w800,
                              color: const Color(0xFF10B981))),
                      const SizedBox(width: 4),
                      _buildMiniBtn(Icons.add,
                          () => _updateDualQty(productId, 1, true, multiplier),
                          isPlus: true),
                    ],
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Text('@$multiplier',
                  style: GoogleFonts.outfit(
                      fontSize: 10,
                      fontWeight: FontWeight.w900,
                      color: Colors.orange.shade800)),
              const SizedBox(width: 8),
            ],
            Expanded(
              flex: 4,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                decoration: BoxDecoration(
                  color: const Color(0xFFF8FAFC),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFFE2E8F0)),
                ),
                child: Row(
                  children: [
                    _buildMiniBtn(Icons.remove,
                        () => _updateDualQty(productId, -1, false, multiplier)),
                    Expanded(
                      child: TextField(
                        controller:
                            _getLooseQtyCtrl(productId, qty, multiplier),
                        keyboardType:
                            TextInputType.text, // Allow 'x' to be typed
                        textAlign: TextAlign.center,
                        onChanged: (val) => _onQtyTyped(productId),
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w900,
                            fontSize: 14,
                            color: const Color(0xFF1E293B)),
                        decoration: const InputDecoration(
                            isDense: true,
                            contentPadding: EdgeInsets.zero,
                            border: InputBorder.none),
                      ),
                    ),
                    Text('PCS',
                        style: GoogleFonts.outfit(
                            fontSize: 10,
                            fontWeight: FontWeight.w800,
                            color: Colors.blueGrey)),
                    const SizedBox(width: 4),
                    _buildMiniBtn(Icons.add,
                        () => _updateDualQty(productId, 1, false, multiplier),
                        isPlus: true),
                  ],
                ),
              ),
            ),
          ],
        ),
        if (hasSmartInput)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(
                '💡 $smartA Karton (isi $smartB) = ${smartA * smartB} Pcs',
                style: GoogleFonts.outfit(
                    fontSize: 11,
                    color: Colors.green,
                    fontWeight: FontWeight.bold)),
          ),
      ],
    );
  }

  void _updateDualQty(
      String productId, int delta, bool isBulk, int multiplier) {
    final currentTotal = _cart[productId] ?? 0;

    if (multiplier > 1) {
      int bulk = currentTotal ~/ multiplier;
      int loose = currentTotal % multiplier;

      if (isBulk) {
        bulk = (bulk + delta).clamp(0, 9999);
      } else {
        loose = (loose + delta).clamp(0, 9999);
      }

      final newTotal = (bulk * multiplier) + loose;
      setState(() {
        _cart[productId] = newTotal;
        _bulkQtyControllers[productId]?.text = bulk.toString();
        _looseQtyControllers[productId]?.text = loose.toString();
      });
    } else {
      // multiplier == 1: everything is loose PCS
      final newTotal = (currentTotal + delta).clamp(0, 9999);
      setState(() {
        _cart[productId] = newTotal;
        _looseQtyControllers[productId]?.text = newTotal.toString();
      });
    }
  }

  TextEditingController _getDiscountCtrl(String productId) {
    if (!_discountControllers.containsKey(productId)) {
      _discountControllers[productId] = TextEditingController(text: '0');
    }
    return _discountControllers[productId]!;
  }

  Widget _buildDiscountRow({required String productId}) {
    final currentDisc = _discountPercent[productId] ?? 0.0;

    return Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Container(
          padding: const EdgeInsets.all(6),
          decoration: BoxDecoration(
            color: currentDisc > 0
                ? const Color(0xFFEF4444).withOpacity(0.1)
                : const Color(0xFFF1F5F9),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(
            Icons.local_offer_rounded,
            size: 14,
            color:
                currentDisc > 0 ? const Color(0xFFEF4444) : Colors.blueGrey,
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'DISKON PER PRODUK (%)',
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  fontWeight: FontWeight.w900,
                  color:
                      currentDisc > 0 ? const Color(0xFFEF4444) : Colors.blueGrey,
                  letterSpacing: 1.0,
                ),
              ),
              if (currentDisc > 0) ...[
                const SizedBox(height: 2),
                Text(
                  '${currentDisc.toStringAsFixed(currentDisc % 1 == 0 ? 0 : 1)}% off',
                  style: GoogleFonts.outfit(
                    fontSize: 12,
                    fontWeight: FontWeight.w900,
                    color: const Color(0xFFEF4444),
                  ),
                ),
              ],
            ],
          ),
        ),
        const SizedBox(width: 8),
        SizedBox(
          width: 90,
          height: 36,
          child: Container(
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(
                  color: currentDisc > 0
                      ? const Color(0xFFEF4444).withOpacity(0.5)
                      : const Color(0xFFE2E8F0)),
            ),
            child: Row(
              children: [
                const SizedBox(width: 8),
                Expanded(
                  child: TextField(
                    controller: _getDiscountCtrl(productId),
                    keyboardType:
                        const TextInputType.numberWithOptions(decimal: true),
                    style: GoogleFonts.outfit(
                      fontSize: 14,
                      fontWeight: FontWeight.w800,
                      color: const Color(0xFF1E293B),
                    ),
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: EdgeInsets.zero,
                      border: InputBorder.none,
                      hintText: '0',
                      hintStyle: GoogleFonts.outfit(
                          color: Colors.blueGrey.withOpacity(0.5)),
                    ),
                    onChanged: (val) {
                      final parsed = double.tryParse(val) ?? 0.0;
                      final clamped = parsed.clamp(0.0, 100.0);
                      setState(() {
                        _discountPercent[productId] = clamped;
                      });
                    },
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8),
                  decoration: const BoxDecoration(
                    border: Border(left: BorderSide(color: Color(0xFFE2E8F0))),
                  ),
                  child: Text(
                    '%',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      color: Colors.blueGrey,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildMiniBtn(IconData icon, VoidCallback onTap,
      {bool isPlus = false}) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        width: 32,
        height: 32,
        decoration: BoxDecoration(
          color: isPlus
              ? const Color(0xFF10B981).withOpacity(0.06)
              : const Color(0xFFF0FDF4),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Icon(icon,
            size: 14,
            color: isPlus ? const Color(0xFF10B981) : const Color(0xFF64748B)),
      ),
    );
  }

  Widget _buildBottomSummary() {
    final totalSaved = _totalDiscount;
    return Container(
      padding: const EdgeInsets.fromLTRB(24, 20, 24, 40),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: const BorderRadius.only(
            topLeft: Radius.circular(32), topRight: Radius.circular(32)),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.06),
              blurRadius: 20,
              offset: const Offset(0, -5))
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            children: [
              Expanded(
                  child: Text(widget.companyName,
                      style: GoogleFonts.outfit(
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                          color: const Color(0xFF10B981)))),
              Text(widget.store.name,
                  style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: Colors.blueGrey)),
            ],
          ),
          const SizedBox(height: 6),
          const Divider(height: 2, color: Color(0xFFF1F5F9)),
          const SizedBox(height: 6),
          if (totalSaved > 0) ...[
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Total Sebelum Diskon:',
                    style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: Colors.blueGrey,
                        fontWeight: FontWeight.w600)),
                Text(_currency.format(_totalAmount + totalSaved),
                    style: GoogleFonts.outfit(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: Colors.blueGrey,
                        decoration: TextDecoration.lineThrough)),
              ],
            ),
            const SizedBox(height: 4),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Total Diskon:',
                    style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: const Color(0xFFEF4444),
                        fontWeight: FontWeight.w700)),
                Text('- ${_currency.format(totalSaved)}',
                    style: GoogleFonts.outfit(
                        fontSize: 14,
                        fontWeight: FontWeight.w700,
                        color: const Color(0xFFEF4444))),
              ],
            ),
            const SizedBox(height: 8),
            const Divider(height: 1, color: Color(0xFFF1F5F9)),
            const SizedBox(height: 8),
          ],
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    totalSaved > 0
                        ? 'Total Setelah Diskon:'
                        : 'Total Pesanan Digital:',
                    style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: Colors.blueGrey,
                        fontWeight: FontWeight.w600),
                  ),
                  Text(_currency.format(_totalAmount),
                      style: GoogleFonts.outfit(
                          fontSize: 22,
                          fontWeight: FontWeight.w900,
                          color: const Color(0xFF1E293B))),
                ],
              ),
              Row(
                children: [
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                        color: Colors.green.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(8)),
                    child: Text(
                        '${_cart.values.fold(0, (sum, q) => sum + q)} Item',
                        style: GoogleFonts.outfit(
                            color: Colors.green,
                            fontWeight: FontWeight.w800,
                            fontSize: 12)),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            height: 64,
            child: ElevatedButton(
              onPressed:
                  _cart.values.any((q) => q > 0) ? _proceedToCheckout : null,
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF1E293B),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20)),
                elevation: 0,
              ),
              child: Text('LANJUT KE CHECKOUT',
                  style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900,
                      fontSize: 16,
                      color: Colors.white,
                      letterSpacing: 1)),
            ),
          ),
        ],
      ),
    );
  }

  TextEditingController _getBulkQtyCtrl(
      String productId, int totalPcs, int multiplier) {
    if (!_bulkQtyControllers.containsKey(productId)) {
      final bulkQty = multiplier > 1 ? (totalPcs ~/ multiplier) : 0;
      _bulkQtyControllers[productId] =
          TextEditingController(text: bulkQty > 0 ? '$bulkQty' : '0');
    }
    return _bulkQtyControllers[productId]!;
  }

  TextEditingController _getLooseQtyCtrl(
      String productId, int totalPcs, int multiplier) {
    if (!_looseQtyControllers.containsKey(productId)) {
      final looseQty = multiplier > 1 ? (totalPcs % multiplier) : totalPcs;
      _looseQtyControllers[productId] =
          TextEditingController(text: looseQty > 0 ? '$looseQty' : '0');
    }
    return _looseQtyControllers[productId]!;
  }

  TextEditingController _getBulkUnitNameCtrl(
      String productId, String? defaultUnit) {
    if (!_bulkUnitNameControllers.containsKey(productId)) {
      _bulkUnitNameControllers[productId] =
          TextEditingController(text: defaultUnit ?? 'KRT');
    }
    return _bulkUnitNameControllers[productId]!;
  }

  TextEditingController _getBaseUnitNameCtrl(String productId) {
    if (!_baseUnitNameControllers.containsKey(productId)) {
      _baseUnitNameControllers[productId] = TextEditingController(text: 'PCS');
    }
    return _baseUnitNameControllers[productId]!;
  }

  TextEditingController _getMultiplierCtrl(
      String productId, int defaultMultiplier) {
    if (!_multiplierControllers.containsKey(productId)) {
      _multiplierControllers[productId] =
          TextEditingController(text: '$defaultMultiplier');
    }
    return _multiplierControllers[productId]!;
  }

  void _onQtyTyped(String productId) {
    final multiplier =
        int.tryParse(_multiplierControllers[productId]?.text ?? '1') ?? 1;
    final bulkStr = _bulkQtyControllers[productId]?.text ?? '0';
    String looseStr = _looseQtyControllers[productId]?.text ?? '0';

    int parsedLoose = 0;
    looseStr = looseStr.replaceAll(' ', '').toLowerCase();

    if (looseStr.contains('x')) {
      final parts = looseStr.split('x');
      if (parts.length == 2) {
        final a = int.tryParse(parts[0]) ?? 0;
        final b = int.tryParse(parts[1]) ?? 0;
        parsedLoose = a * b;
      }
    } else if (looseStr.contains('*')) {
      final parts = looseStr.split('*');
      if (parts.length == 2) {
        final a = int.tryParse(parts[0]) ?? 0;
        final b = int.tryParse(parts[1]) ?? 0;
        parsedLoose = a * b;
      }
    } else {
      parsedLoose = int.tryParse(looseStr) ?? 0;
    }

    final bulk = int.tryParse(bulkStr) ?? 0;

    setState(() {
      _cart[productId] =
          multiplier > 1 ? ((bulk * multiplier) + parsedLoose) : parsedLoose;
    });
  }

  void _proceedToCheckout() {
    if (isSalesDistributor) {
      for (final entry in _cart.entries) {
        if (entry.value > 0) {
          final cp = _customPrice[entry.key] ?? 0.0;
          if (cp <= 0) {
            final p = _allProducts.firstWhere(
                (prod) => prod['id'] == entry.key,
                orElse: () => <String, dynamic>{});
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(
                  'Harap atur harga jual untuk ${p['name'] ?? 'produk'} terlebih dahulu',
                  style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
                ),
                backgroundColor: Colors.red.shade700,
              ),
            );
            if (p.isNotEmpty) _showCustomPriceDialog(p);
            return;
          }
        }
      }
    }

    final List<Map<String, dynamic>> items = [];
    _cart.forEach((id, qty) {
      if (qty > 0) {
        final p = _allProducts.firstWhere((prod) => prod['id'] == id);
        final bulkUnit = _bulkUnitNameControllers[id]?.text ?? 'KRT';
        final baseUnit = _baseUnitNameControllers[id]?.text ?? 'PCS';
        final multiplier =
            int.tryParse(_multiplierControllers[id]?.text ?? '1') ?? 1;
        final bulkQty = qty ~/ (multiplier > 0 ? multiplier : 1);
        final looseQty = qty % (multiplier > 0 ? multiplier : 1);
        String displayUnit = baseUnit;
        int orderedQty = qty;
        if (multiplier > 1 && bulkQty > 0) {
          displayUnit = bulkUnit;
          orderedQty = bulkQty;
        }
        final disc = _discountPercent[id] ?? 0.0;
        final basePrice = _getBasePriceForProduct(p);
        final effectivePrice = basePrice * (1 - disc / 100);
        final tier = _selectedPriceTier[id] ?? (isSalesDistributor ? 'CUSTOM' : 'NORMAL');
        items.add({
          'product_id': id,
          'quantity': qty,
          'ordered_quantity': orderedQty,
          'unit': displayUnit,
          'pieces_per_unit': multiplier,
          'price': basePrice,
          'effective_price': effectivePrice,
          'discount_percent': disc,
          'price_tier': tier,
          'product_name': p['name'],
          'breakdown': multiplier > 1
              ? '$bulkQty $bulkUnit + $looseQty $baseUnit'
              : '$qty $baseUnit',
        });
      }
    });

    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => VisitCheckoutPage(
          store: widget.store,
          selfiePath: widget.selfiePath,
          selfieBytes: widget.selfieBytes,
          companyId: widget.companyId,
          companyName: widget.companyName,
          items: items,
          totalAmount: _totalAmount,
          receiptPath: null,
          jobPositionTitle: widget.jobPositionTitle,
          isDistributor: isSalesDistributor,
          notes: null,
        ),
      ),
    );
  }
}
