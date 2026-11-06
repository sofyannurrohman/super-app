import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:geolocator/geolocator.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_state.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';

import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:uuid/uuid.dart';
import 'package:drift/drift.dart' hide Column;
import 'package:intl/intl.dart';
import 'package:image_picker/image_picker.dart';
import 'package:hris_app/features/epkwt/presentation/widgets/signature_canvas.dart';
import 'package:hris_app/core/utils/image_compress_helper.dart';
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'visit_success_page.dart';

class InstantOrderPage extends StatefulWidget {
  final StoreModel? initialStore;
  final String? initialVisitId;

  const InstantOrderPage({super.key, this.initialStore, this.initialVisitId});

  @override
  State<InstantOrderPage> createState() => _InstantOrderPageState();
}

class _InstantOrderPageState extends State<InstantOrderPage> {
  final _storeApi = StoreApiService(
    apiClient: di.sl<ApiClient>(),
    prefs: di.sl<SharedPreferences>(),
  );
  final _salesApi = SalesApiService(apiClient: di.sl<ApiClient>());
  final _db = di.sl<AppDatabase>();

  List<StoreModel> _stores = [];
  List<Map<String, dynamic>> _products = [];
  List<dynamic> _pendingVisits = [];

  StoreModel? _selectedStore;
  String? _selectedVisitId;
  List<Map<String, dynamic>> _cartItems = [];

  // --- Discount ---
  final Map<String, double> _discountPercent = {};
  final Map<String, TextEditingController> _discountControllers = {};

  // --- Flexible Pricing Khusus Booster & Distributor ---
  String _jobPositionTitle = '';
  bool _isDistributorProfile = false;
  bool get isSalesBooster =>
      _jobPositionTitle.toLowerCase().contains('booster');
  bool get isSalesDistributor =>
      _isDistributorProfile ||
      _jobPositionTitle.toLowerCase().contains('distributor') ||
      (_selectedStore?.customerType.toUpperCase() == 'DISTRIBUTOR');
  bool get hasFlexiblePricing => isSalesBooster || isSalesDistributor;
  final Map<String, String> _selectedPriceTier =
      {}; // 'NORMAL', 'WBP', 'DBP', 'CUSTOM'
  final Map<String, double> _customPrice = {};

  // --- Payment ---
  String _paymentMethod = 'CASH'; // 'CASH' or 'TRANSFER'
  String? _transferProofPath; // local path of transfer proof image
  List<Map<String, dynamic>> _bankAccounts = [];
  Map<String, dynamic>? _selectedBankAccount;
  bool _isLoadingBanks = false;
  String _branchId = '';
  String _companyId = '';

  // --- Signatures ---
  String? _salesSignatureBase64;
  String? _customerSignatureBase64;

  bool _isLoading = true;
  bool _isSubmitting = false;
  bool _isOffline = false;

  final NumberFormat _currencyFormat = NumberFormat.currency(
    locale: 'id_ID',
    symbol: 'Rp',
    decimalDigits: 0,
  );
  final ImagePicker _picker = ImagePicker();

  @override
  void dispose() {
    for (var ctrl in _discountControllers.values) {
      ctrl.dispose();
    }
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    _loadProfile();
    if (widget.initialStore != null) {
      _selectedStore = widget.initialStore;
      _selectedVisitId = widget.initialVisitId;
    }
    _fetchData();
  }

  Future<void> _loadProfile() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final cached = prefs.getString('cached_employee_profile');
      if (cached != null) {
        final profile = jsonDecode(cached) as Map<String, dynamic>;
        final title =
            (profile['job_position']?['title'] ?? profile['job_title'] ?? '')
                .toString();
        final isDist =
            profile['is_distributor'] == true ||
            (profile['distributor_store_id'] != null &&
                profile['distributor_store_id'].toString().isNotEmpty) ||
            title.toLowerCase().contains('distributor');
        if (mounted) {
          setState(() {
            if (title.isNotEmpty) _jobPositionTitle = title;
            if (isDist) _isDistributorProfile = true;
          });
        }
      }
    } catch (_) {}
  }

  // ─────────────────────────────────────────────────
  // DATA FETCHING — Network-First with Offline Fallback
  // ─────────────────────────────────────────────────

  Future<void> _fetchData() async {
    if (!mounted) return;
    setState(() {
      _isLoading = true;
      _isOffline = false;
    });

    bool serverSuccess = false;

    // 1. Coba ambil data dari server (Network-First)
    try {
      await Future.wait([
        _fetchStoresFromServer(),
        _fetchProductsFromServer(),
        _fetchPendingVisits(),
        _fetchBankAccounts(),
      ]);
      serverSuccess = true;
    } catch (e) {
      debugPrint(
        'DEBUG InstantOrder: Server fetch failed ($e), falling back to local.',
      );
      serverSuccess = false;
    }

    // 2. Jika server gagal, fallback ke local DB
    if (!serverSuccess || _stores.isEmpty || _products.isEmpty) {
      await _fetchFromLocal();
      if (mounted) {
        setState(() => _isOffline = true);
      }
    }

    if (mounted) {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _fetchStoresFromServer() async {
    String? employeeId;
    if (mounted) {
      final profileState = context.read<ProfileBloc>().state;
      if (profileState is ProfileLoaded) {
        employeeId = profileState.profile['id']?.toString();
        _jobPositionTitle =
            profileState.profile['job_position']?['title']?.toString() ?? '';
        _isDistributorProfile =
            profileState.profile['is_distributor'] == true ||
            (profileState.profile['distributor_store_id'] != null &&
                profileState.profile['distributor_store_id']
                    .toString()
                    .isNotEmpty);
      }
    }

    final stores = await _storeApi.getStores(assignedEmployeeId: employeeId);
    if (mounted) setState(() => _stores = stores);

    // Setelah dapat dari server, simpan ke local DB sebagai cache toko
    // (toko disimpan via sync biasa, tidak ada tabel khusus lokal untuk store)
    try {
      final prefs = await SharedPreferences.getInstance();
      final storesJson = jsonEncode(stores.map((s) => s.toJson()).toList());
      await prefs.setString('cached_stores_list', storesJson);
    } catch (e) {
      debugPrint('DEBUG InstantOrder: Failed to cache stores: $e');
    }
  }

  static bool _isAncillaryOrGift(String cat, String n) {
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

  static bool _isProcurementProduct(dynamic category, dynamic name) {
    final cat = (category ?? '').toString().toUpperCase();
    final n = (name ?? '').toString().toUpperCase();

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
    if (cat.contains('TRADING') || cat.contains('PRODUK_PELENGKAP')) {
      return true;
    }

    return false;
  }

  Future<void> _fetchProductsFromServer() async {
    final api = di.sl<ApiClient>();
    final res = await api.client.get(
      'factory/products',
      queryParameters: {
        'company_id': 'all',
        'scope': 'sales',
        'type': 'finished_goods',
      },
    );

    if (res.statusCode == 200) {
      List? raw;
      if (res.data is Map && res.data.containsKey('data')) {
        raw = res.data['data'] as List?;
      } else if (res.data is List) {
        raw = res.data as List?;
      }

      if (raw != null && raw.isNotEmpty) {
        final sellableRaw = raw
            .where(
              (p) => !_isProcurementProduct(
                p['category'] ?? p['Category'],
                p['name'] ?? p['Name'],
              ),
            )
            .toList();

        final mapped = sellableRaw
            .map<Map<String, dynamic>>(
              (p) => {
                'id': p['id'] ?? p['ID'] ?? '',
                'name': p['name'] ?? p['Name'] ?? 'Unknown',
                'selling_price':
                    ((p['selling_price'] ??
                                p['sellingPrice'] ??
                                p['SellingPrice'] ??
                                0)
                            as num)
                        .toDouble(),
                'dbp': ((p['dbp'] ?? p['DBP'] ?? 0) as num).toDouble(),
                'wbp': ((p['wbp'] ?? p['WBP'] ?? 0) as num).toDouble(),
                'rbp': ((p['rbp'] ?? p['RBP'] ?? 0) as num).toDouble(),
                'cost_price':
                    ((p['cost_price'] ?? p['costPrice'] ?? p['CostPrice'] ?? 0)
                            as num)
                        .toDouble(),
                'sku': p['sku'] ?? p['SKU'] ?? '',
                'unit': p['unit'] ?? p['Unit'] ?? 'PCS',
                'pcs_per_unit':
                    p['pcs_per_unit'] ??
                    p['pcsPerUnit'] ??
                    p['PcsPerUnit'] ??
                    1,
                'category': p['category'] ?? p['Category'] ?? 'Uncategorized',
                'image_url': p['image_url'] ?? p['imageUrl'],
              },
            )
            .toList();

        if (mounted) setState(() => _products = mapped);

        // Update local DB agar fallback selalu segar
        try {
          await _db.batch((batch) {
            for (var p in sellableRaw) {
              batch.insert(
                _db.products,
                ProductsCompanion.insert(
                  id: p['id'] ?? p['ID'] ?? '',
                  companyId: p['company_id'] ?? p['companyId'] ?? '',
                  name: p['name'] ?? p['Name'] ?? 'Unknown',
                  sellingPrice:
                      (p['selling_price'] ??
                              p['sellingPrice'] ??
                              p['SellingPrice'] ??
                              0 as num)
                          .toDouble(),
                  dbp: Value(((p['dbp'] ?? p['DBP'] ?? 0) as num).toDouble()),
                  wbp: Value(((p['wbp'] ?? p['WBP'] ?? 0) as num).toDouble()),
                  rbp: Value(((p['rbp'] ?? p['RBP'] ?? 0) as num).toDouble()),
                  costPrice: Value(
                    ((p['cost_price'] ?? p['costPrice'] ?? 0) as num)
                        .toDouble(),
                  ),
                  sku: Value(p['sku'] ?? p['SKU']),
                  unit: Value(p['unit'] ?? p['Unit']),
                  pcsPerUnit: Value(
                    p['pcs_per_unit'] ??
                        p['pcsPerUnit'] ??
                        p['PcsPerUnit'] ??
                        1,
                  ),
                  category: Value(p['category'] ?? p['Category']),
                  warehouseStock: const Value(0),
                  imageUrl: Value(p['image_url'] ?? p['imageUrl']),
                ),
                mode: InsertMode.insertOrReplace,
              );
            }
          });
        } catch (e) {
          debugPrint(
            'DEBUG InstantOrder: Failed to cache products to local DB: $e',
          );
        }
      }
    }
  }

  Future<void> _fetchPendingVisits() async {
    try {
      final visits = await _salesApi.getPendingVisits();
      if (mounted) setState(() => _pendingVisits = visits);
    } catch (_) {}
  }

  Future<void> _fetchBankAccounts() async {
    if (!mounted) return;
    setState(() => _isLoadingBanks = true);
    try {
      if (_branchId.isEmpty || _companyId.isEmpty) {
        final prefs = await SharedPreferences.getInstance();
        _branchId = prefs.getString('branch_id') ?? '';
        _companyId = prefs.getString('company_id') ?? '';
      }
      if (mounted) {
        final profileState = context.read<ProfileBloc>().state;
        if (profileState is ProfileLoaded) {
          final p = profileState.profile;
          _branchId =
              p['branch_id']?.toString() ??
              p['BranchID']?.toString() ??
              _branchId;
          _companyId =
              p['company_id']?.toString() ??
              p['CompanyID']?.toString() ??
              _companyId;
        }
      }
      final banks = await _salesApi.getBankAccounts(branchId: _branchId);
      if (mounted) {
        setState(() {
          _bankAccounts = banks;
          _isLoadingBanks = false;
          if (_bankAccounts.isNotEmpty) {
            _selectedBankAccount = _bankAccounts.firstWhere(
              (b) => b['is_active'] == true,
              orElse: () => _bankAccounts.first,
            );
          }
        });
      }
    } catch (_) {
      if (mounted) setState(() => _isLoadingBanks = false);
    }
  }

  Future<void> _fetchFromLocal() async {
    try {
      // Toko: ambil dari shared_preferences atau minimal kosongkan,
      // karena StoreModel tidak disimpan di Drift DB lokal.
      try {
        final prefs = await SharedPreferences.getInstance();
        final cachedStoresStr = prefs.getString('cached_stores_list');
        if (cachedStoresStr != null && cachedStoresStr.isNotEmpty) {
          final List<dynamic> decoded = jsonDecode(cachedStoresStr);
          final loadedStores = decoded
              .map((e) => StoreModel.fromJson(e))
              .toList();
          if (mounted) setState(() => _stores = loadedStores);
        }
      } catch (e) {
        debugPrint('DEBUG InstantOrder: Failed to load cached stores: $e');
      }

      // Produk: ambil dari tabel products di lokal DB
      final localProducts = await _db.select(_db.products).get();
      if (mounted) {
        setState(() {
          _products = localProducts
              .where((p) => !_isProcurementProduct(p.category, p.name))
              .map(
                (p) => {
                  'id': p.id,
                  'name': p.name,
                  'selling_price': p.sellingPrice,
                  'dbp': p.dbp,
                  'wbp': p.wbp,
                  'rbp': p.rbp,
                  'cost_price': p.costPrice,
                  'sku': p.sku ?? '',
                  'unit': p.unit ?? 'PCS',
                  'pcs_per_unit': p.pcsPerUnit,
                  'category': p.category ?? 'Uncategorized',
                  'image_url': p.imageUrl,
                },
              )
              .toList();
        });
      }
    } catch (e) {
      debugPrint('DEBUG InstantOrder: fetchFromLocal error: $e');
    }
  }

  // ─────────────────────────────────────────────────
  // STORE PICKER
  // ─────────────────────────────────────────────────

  void _showStorePicker() {
    if (_stores.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'Data toko tidak tersedia. Periksa koneksi internet Anda.',
          ),
        ),
      );
      return;
    }
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => _SearchPickerSheet<StoreModel>(
        title: 'Cari Toko',
        items: _stores,
        searchHint: 'Ketik nama toko...',
        itemBuilder: (context, store) => ListTile(
          leading: CircleAvatar(
            backgroundColor: AppColors.blueSalesLight.withValues(alpha: 0.2),
            child: const Icon(
              Icons.storefront_rounded,
              color: AppColors.blueSales,
            ),
          ),
          title: Text(
            store.name,
            style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
          ),
          subtitle: Text(
            store.address,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: GoogleFonts.outfit(fontSize: 12),
          ),
          onTap: () {
            setState(() => _selectedStore = store);
            Navigator.pop(context);
          },
        ),
        filterItem: (store, query) =>
            store.name.toLowerCase().contains(query.toLowerCase()),
      ),
    );
  }

  void _showPendingVisitsPicker() {
    if (_pendingVisits.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Tidak ada kunjungan pending saat ini.')),
      );
      return;
    }

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => _SearchPickerSheet<dynamic>(
        title: 'Pilih Kunjungan Pending',
        items: _pendingVisits,
        searchHint: 'Ketik nama toko...',
        itemBuilder: (context, visit) {
          final storeName = visit['store']?['name'] ?? 'Unknown Store';
          final storeAddress = visit['store']?['address'] ?? '';
          final checkInTimeStr = visit['check_in_time'] ?? '';

          String timeFormatted = '';
          if (checkInTimeStr.isNotEmpty) {
            try {
              final dt = DateTime.parse(checkInTimeStr).toLocal();
              timeFormatted = DateFormat('HH:mm').format(dt);
            } catch (_) {}
          }

          return ListTile(
            leading: CircleAvatar(
              backgroundColor: AppColors.warning.withValues(alpha: 0.2),
              child: const Icon(
                Icons.access_time_rounded,
                color: AppColors.warning,
              ),
            ),
            title: Text(
              storeName,
              style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
            ),
            subtitle: Text(
              'Check-in: $timeFormatted\n$storeAddress',
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
              style: GoogleFonts.outfit(fontSize: 12),
            ),
            isThreeLine: true,
            onTap: () {
              final storeId = visit['store_id'];
              StoreModel? matchedStore;
              try {
                matchedStore = _stores.firstWhere((s) => s.id == storeId);
              } catch (_) {}

              setState(() {
                _selectedVisitId = visit['id'];
                if (matchedStore != null) {
                  _selectedStore = matchedStore;
                } else {
                  _selectedStore = StoreModel(
                    id: storeId,
                    name: storeName,
                    address: storeAddress,
                    ownerName: '-',
                  );
                }
              });
              Navigator.pop(context);
            },
          );
        },
        filterItem: (visit, query) {
          final name = (visit['store']?['name'] ?? '').toString().toLowerCase();
          return name.contains(query.toLowerCase());
        },
      ),
    );
  }

  // ─────────────────────────────────────────────────
  // PRODUCT PICKER & CART
  // ─────────────────────────────────────────────────

  void _showProductPicker() {
    if (_products.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'Data produk tidak tersedia. Tarik ke bawah untuk refresh.',
          ),
        ),
      );
      return;
    }

    final rawCategories = _products
        .map((p) => (p['category'] ?? 'Uncategorized').toString().trim())
        .where((c) => c.isNotEmpty)
        .toSet()
        .toList();
    rawCategories.sort();

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => _SearchPickerSheet<Map<String, dynamic>>(
        title: 'Cari Produk',
        items: _products,
        searchHint: 'Ketik nama produk, SKU, atau kategori...',
        categories: ['Semua', ...rawCategories],
        categoryExtractor: (product) =>
            (product['category'] ?? 'Uncategorized').toString(),
        itemBuilder: (context, product) {
          final isSampingan = _isAncillaryOrGift(
            (product['category'] ?? '').toString().toUpperCase(),
            (product['name'] ?? '').toString().toUpperCase(),
          );
          final categoryName = product['category'] ?? 'Uncategorized';
          final sku = (product['sku'] ?? '').toString();

          return ListTile(
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 4,
            ),
            leading: CircleAvatar(
              backgroundColor: isSampingan
                  ? Colors.amber.withValues(alpha: 0.15)
                  : AppColors.primaryGreen.withValues(alpha: 0.1),
              child: Icon(
                isSampingan ? Icons.redeem_rounded : Icons.inventory_2_rounded,
                color: isSampingan
                    ? Colors.amber.shade800
                    : AppColors.primaryGreen,
              ),
            ),
            title: Row(
              children: [
                Expanded(
                  child: Text(
                    product['name'] ?? '-',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w700,
                      fontSize: 15,
                    ),
                  ),
                ),
                if (isSampingan)
                  Container(
                    margin: const EdgeInsets.only(left: 6),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 2,
                    ),
                    decoration: BoxDecoration(
                      color: Colors.amber.shade50,
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(
                        color: Colors.amber.shade300,
                        width: 0.8,
                      ),
                    ),
                    child: Text(
                      'Sampingan',
                      style: GoogleFonts.outfit(
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                        color: Colors.amber.shade900,
                      ),
                    ),
                  ),
              ],
            ),
            subtitle: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: 2),
                Text(
                  '${sku.isNotEmpty ? "$sku • " : ""}$categoryName',
                  style: GoogleFonts.outfit(
                    fontSize: 11,
                    color: const Color(0xFF64748B),
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  _currencyFormat.format(product['selling_price'] ?? 0),
                  style: GoogleFonts.outfit(
                    fontSize: 13,
                    color: AppColors.primaryGreen,
                    fontWeight: FontWeight.w800,
                  ),
                ),
              ],
            ),
            onTap: () {
              _addProductToCart(product);
              Navigator.pop(context);
            },
          );
        },
        filterItem: (product, query) {
          final name = (product['name'] ?? '').toString().toLowerCase();
          final sku = (product['sku'] ?? '').toString().toLowerCase();
          final cat = (product['category'] ?? '').toString().toLowerCase();
          final q = query.toLowerCase();
          return name.contains(q) || sku.contains(q) || cat.contains(q);
        },
      ),
    );
  }

  void _addProductToCart(Map<String, dynamic> product) {
    setState(() {
      final existingIndex = _cartItems.indexWhere(
        (item) => item['product']['id'] == product['id'],
      );
      if (existingIndex >= 0) {
        _cartItems[existingIndex]['quantity'] += 1;
      } else {
        _cartItems.add({'product': product, 'quantity': 1});
      }
    });
  }

  void _updateQuantity(int index, int delta) {
    setState(() {
      final currentQty = _cartItems[index]['quantity'] as int;
      final newQty = currentQty + delta;
      if (newQty > 0) {
        _cartItems[index]['quantity'] = newQty;
      } else {
        _cartItems.removeAt(index);
      }
    });
  }

  void _setQuantity(int index, int qty) {
    setState(() {
      if (qty > 0) {
        _cartItems[index]['quantity'] = qty;
      } else {
        _cartItems.removeAt(index);
      }
    });
  }

  void _showEditQuantityDialog(
    BuildContext context,
    int index,
    int currentQty,
  ) {
    final TextEditingController qtyController = TextEditingController(
      text: currentQty.toString(),
    );

    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: Text(
            'Edit Kuantitas',
            style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
          ),
          content: TextField(
            controller: qtyController,
            keyboardType: TextInputType.number,
            autofocus: true,
            decoration: const InputDecoration(
              labelText: 'Jumlah',
              border: OutlineInputBorder(),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text(
                'Batal',
                style: GoogleFonts.outfit(color: Colors.grey),
              ),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primaryGreen,
              ),
              onPressed: () {
                final int? newQty = int.tryParse(qtyController.text);
                if (newQty != null) {
                  _setQuantity(index, newQty);
                }
                Navigator.pop(context);
              },
              child: Text(
                'Simpan',
                style: GoogleFonts.outfit(color: Colors.white),
              ),
            ),
          ],
        );
      },
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
    final currentCustom =
        _customPrice[id] ?? (isDist ? 0.0 : _getBasePriceForProduct(p));
    final normalPrice = ((p['selling_price'] ?? 0) as num).toDouble();
    final dbpPrice = ((p['dbp'] ?? 0) as num).toDouble();
    final wbpPrice = ((p['wbp'] ?? 0) as num).toDouble();
    final costPrice = ((p['cost_price'] ?? 0) as num).toDouble();

    final primaryColor = isDist ? const Color(0xFF4338CA) : Colors.purple;
    final bgShade = isDist ? Colors.indigo.shade50 : Colors.purple.shade50;

    // Batas bawah harga: DBP atau Cost Price (HPP) khusus booster
    final double floorPrice = dbpPrice > 0
        ? dbpPrice
        : (costPrice > 0 ? costPrice : 0.0);

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
                borderRadius: BorderRadius.circular(20),
              ),
              title: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: bgShade,
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Icon(
                      isDist
                          ? Icons.storefront_rounded
                          : Icons.edit_note_rounded,
                      color: primaryColor,
                      size: 22,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          isDist
                              ? 'Atur Harga Jual'
                              : 'Harga Fleksibel Booster',
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
                              Text(
                                'Normal (HET):',
                                style: GoogleFonts.outfit(fontSize: 11),
                              ),
                              Text(
                                _currencyFormat.format(normalPrice),
                                style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ],
                          ),
                          if (wbpPrice > 0)
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'Grosir (WBP):',
                                  style: GoogleFonts.outfit(fontSize: 11),
                                ),
                                Text(
                                  _currencyFormat.format(wbpPrice),
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.blue.shade700,
                                  ),
                                ),
                              ],
                            ),
                          if (dbpPrice > 0)
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'Distributor (DBP):',
                                  style: GoogleFonts.outfit(fontSize: 11),
                                ),
                                Text(
                                  _currencyFormat.format(dbpPrice),
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.green.shade700,
                                  ),
                                ),
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
                        horizontal: 16,
                        vertical: 12,
                      ),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: BorderSide.none,
                      ),
                    ),
                    onChanged: (val) {
                      final entered =
                          double.tryParse(
                            val.replaceAll(RegExp(r'[^0-9]'), ''),
                          ) ??
                          0;
                      if (entered <= 0) {
                        setDialogState(() {
                          errorText = 'Nominal harga harus lebih dari 0';
                        });
                      } else if (!isDist &&
                          floorPrice > 0 &&
                          entered < floorPrice) {
                        setDialogState(() {
                          errorText =
                              'Min. ${_currencyFormat.format(floorPrice)} (Batas Modal)';
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
                      borderRadius: BorderRadius.circular(10),
                    ),
                  ),
                  onPressed: () {
                    final entered =
                        double.tryParse(
                          priceCtrl.text.replaceAll(RegExp(r'[^0-9]'), ''),
                        ) ??
                        0;
                    if (entered <= 0) {
                      setDialogState(() {
                        errorText = 'Nominal harga harus lebih dari 0';
                      });
                      return;
                    }
                    if (!isDist && floorPrice > 0 && entered < floorPrice) {
                      setDialogState(() {
                        errorText =
                            'Min. ${_currencyFormat.format(floorPrice)} (Batas Modal)';
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
    final customPrice = _customPrice[id] ?? 0.0;

    // KHUSUS SALES DISTRIBUTOR (LUAR PULAU): Sembunyikan harga asli pabrik (Normal, WBP, DBP)
    if (isSalesDistributor) {
      return Container(
        margin: const EdgeInsets.only(top: 8, bottom: 4),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
        decoration: BoxDecoration(
          color: Colors.indigo.shade50.withValues(alpha: 0.5),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.indigo.shade100),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Row(
              children: [
                Icon(
                  Icons.storefront_rounded,
                  size: 16,
                  color: Colors.indigo.shade700,
                ),
                const SizedBox(width: 6),
                Text(
                  customPrice > 0
                      ? 'Harga Jual: ${_currencyFormat.format(customPrice)}'
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
      margin: const EdgeInsets.only(top: 8, bottom: 4),
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.purple.shade50.withValues(alpha: 0.5),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.purple.shade100),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.bolt_rounded, size: 16, color: Colors.purple.shade700),
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
                  label: 'Normal: ${_currencyFormat.format(normalPrice)}',
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
                    label: 'Grosir (WBP): ${_currencyFormat.format(wbpPrice)}',
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
                    label:
                        'Distributor (DBP): ${_currencyFormat.format(dbpPrice)}',
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
                      ? 'Custom: ${_currencyFormat.format(customPrice)} ✎'
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
    final activeBg = isDistributor
        ? const Color(0xFF4338CA)
        : Colors.purple.shade700;
    final activeBorder = isDistributor
        ? const Color(0xFF3730A3)
        : Colors.purple.shade800;
    final inactiveBorder = isDistributor
        ? Colors.indigo.shade200
        : Colors.purple.shade200;
    final inactiveText = isDistributor
        ? Colors.indigo.shade900
        : Colors.purple.shade900;

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
        decoration: BoxDecoration(
          color: isSelected ? activeBg : Colors.white,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: isSelected ? activeBorder : inactiveBorder),
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
    return _cartItems.fold(0.0, (sum, item) {
      final p = item['product'] as Map<String, dynamic>;
      final basePrice = _getBasePriceForProduct(p);
      final qty = item['quantity'] as int;
      final id = p['id'];
      final disc = _discountPercent[id] ?? 0.0;
      final effectivePrice = basePrice * (1 - disc / 100);
      return sum + (effectivePrice * qty);
    });
  }

  // ─────────────────────────────────────────────────
  // TRANSFER PROOF — Pick from camera or gallery
  // ─────────────────────────────────────────────────

  Future<void> _pickTransferProof() async {
    final source = await showModalBottomSheet<ImageSource>(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) => Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'Ambil Bukti Transfer',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w900,
                fontSize: 16,
              ),
            ),
            const SizedBox(height: 20),
            ListTile(
              leading: const CircleAvatar(
                backgroundColor: Color(0xFFE0F2FE),
                child: Icon(Icons.camera_alt_rounded, color: Color(0xFF0EA5E9)),
              ),
              title: Text(
                'Kamera',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
              ),
              onTap: () => Navigator.pop(ctx, ImageSource.camera),
            ),
            ListTile(
              leading: const CircleAvatar(
                backgroundColor: Color(0xFFF0FDF4),
                child: Icon(
                  Icons.photo_library_rounded,
                  color: Color(0xFF22C55E),
                ),
              ),
              title: Text(
                'Galeri',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
              ),
              onTap: () => Navigator.pop(ctx, ImageSource.gallery),
            ),
          ],
        ),
      ),
    );

    if (source == null) return;
    final XFile? image = await _picker.pickImage(
      source: source,
      imageQuality: 75,
      maxWidth: 1080,
    );
    if (image != null && mounted) {
      setState(() => _transferProofPath = image.path);
    }
  }

  // ─────────────────────────────────────────────────
  // SUBMIT — Save Local First, then try server
  // ─────────────────────────────────────────────────

  Future<void> _submitOrder() async {
    if (_isSubmitting) return;

    // Validasi
    if (_selectedStore == null) {
      _showSnack('Pilih toko terlebih dahulu', isError: true);
      return;
    }
    if (_cartItems.isEmpty) {
      _showSnack('Tambahkan minimal 1 produk', isError: true);
      return;
    }

    if (isSalesDistributor) {
      for (final item in _cartItems) {
        final p = item['product'] as Map<String, dynamic>;
        final id = p['id']?.toString() ?? '';
        final cp = _customPrice[id] ?? 0.0;
        if (cp <= 0) {
          _showSnack(
            'Harap atur harga jual untuk ${p['name'] ?? 'produk'} terlebih dahulu',
            isError: true,
          );
          _showCustomPriceDialog(p);
          return;
        }
      }
    }
    if (_salesSignatureBase64 == null || _customerSignatureBase64 == null) {
      _showSnack('Tanda tangan sales dan pelanggan wajib diisi', isError: true);
      return;
    }

    setState(() => _isSubmitting = true);

    Position? position;
    try {
      position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
        timeLimit: const Duration(seconds: 5),
      );
    } catch (_) {
      // Ignore if permission denied or timeout
    }

    try {
      final localId = const Uuid().v4();
      final now = DateTime.now();
      final totalAmt = _totalAmount;
      String combinedSignature = '';
      if (_salesSignatureBase64 != null && _customerSignatureBase64 != null) {
        String salesPath = _salesSignatureBase64!;
        String custPath = _customerSignatureBase64!;
        try {
          final salesB64 = _salesSignatureBase64!.contains(',')
              ? _salesSignatureBase64!.split(',').last
              : _salesSignatureBase64!;
          final salesBytes = base64Decode(salesB64);
          salesPath = await ImageCompressHelper.saveSignatureToFile(
            salesBytes,
            prefix: 'sig_sales_${localId.substring(0, 6)}',
          );
        } catch (_) {}

        try {
          final custB64 = _customerSignatureBase64!.contains(',')
              ? _customerSignatureBase64!.split(',').last
              : _customerSignatureBase64!;
          final custBytes = base64Decode(custB64);
          custPath = await ImageCompressHelper.saveSignatureToFile(
            custBytes,
            prefix: 'sig_cust_${localId.substring(0, 6)}',
          );
        } catch (_) {}

        combinedSignature = '$salesPath,$custPath';
      }
      final receiptNo =
          'INV-INSTANT-${DateFormat('yyyyMMdd').format(now)}-${localId.substring(0, 4).toUpperCase()}';

      // ── STEP 1: Simpan ke local DB (selalu, bahkan saat online) ──
      await _db.transaction(() async {
        await _db
            .into(_db.localTransactions)
            .insert(
              LocalTransactionsCompanion.insert(
                localId: localId,
                companyId: '00000000-0000-0000-0000-000000000000',
                companyName: 'Instant Order',
                storeId: _selectedStore!.id,
                storeName: _selectedStore!.name,
                totalAmount: totalAmt,
                receiptNo: Value(receiptNo),
                paidAmount: Value(totalAmt),
                signaturePath: Value(combinedSignature),
                paymentMethod: Value(_paymentMethod),
                companyBankAccountId:
                    _paymentMethod == 'TRANSFER' && _selectedBankAccount != null
                    ? Value(_selectedBankAccount!['id']?.toString())
                    : const Value.absent(),
                receiptPath: _transferProofPath != null
                    ? Value(_transferProofPath)
                    : const Value.absent(),
                createdAt: now,
                syncStatus: const Value('pending'),
                isMt: const Value(false),
                latitude: position != null
                    ? Value(position.latitude)
                    : const Value.absent(),
                longitude: position != null
                    ? Value(position.longitude)
                    : const Value.absent(),
                visitId: _selectedVisitId != null
                    ? Value(_selectedVisitId)
                    : const Value.absent(),
              ),
            );

        for (final item in _cartItems) {
          final p = item['product'] as Map<String, dynamic>;
          final qty = item['quantity'] as int;
          final id = p['id'] ?? '';
          final disc = _discountPercent[id] ?? 0.0;
          final basePrice = _getBasePriceForProduct(p);

          await _db
              .into(_db.localTransactionItems)
              .insert(
                LocalTransactionItemsCompanion.insert(
                  transactionLocalId: localId,
                  productId: id,
                  productName: p['name'] ?? '',
                  quantity: qty,
                  orderedQuantity: Value(qty),
                  unit: Value(p['unit']?.toString() ?? 'PCS'),
                  piecesPerUnit: Value((p['pcs_per_unit'] ?? 1) as int),
                  price: basePrice,
                  discountPercent: Value(disc),
                ),
              );
        }
      });

      // ── STEP 2: Trigger background sync (kirim ke server jika ada internet) ──
      // SyncBloc akan mendeteksi pending record & mengirimkannya, termasuk upload foto
      di.sl<SyncBloc>().add(const SyncDataRequested(force: true));

      if (mounted) {
        final itemsJson = _cartItems.map((item) {
          final p = item['product'] as Map<String, dynamic>;
          final id = p['id'] ?? '';
          final disc = _discountPercent[id] ?? 0.0;
          final basePrice = _getBasePriceForProduct(p);
          final effectivePrice = basePrice * (1 - disc / 100);
          final tier =
              _selectedPriceTier[id] ?? (isSalesDistributor ? 'DBP' : 'NORMAL');

          return {
            'product_name': p['name'],
            'quantity': item['quantity'],
            'price': basePrice,
            'effective_price': effectivePrice,
            'discount_percent': disc,
            'price_tier': tier,
          };
        }).toList();

        final transaction = {
          'receipt_no': receiptNo,
          'store': {
            'name': _selectedStore!.name,
            'address': _selectedStore!.address,
          },
          'company': {'name': 'Instant Order'},
          'total_amount': totalAmt,
          'transaction_date': now.toUtc().toIso8601String(),
          'status': 'LUNAS',
          'payment_method': _paymentMethod,
          'company_bank_account_id': _selectedBankAccount?['id'],
          'items': itemsJson,
        };

        // Trigger Sync in background so it doesn't get stuck in pending if online
        di.sl<SyncBloc>().add(const SyncDataRequested(force: true));

        Navigator.of(context).pushReplacement(
          MaterialPageRoute(
            builder: (_) =>
                VisitSuccessPage(transaction: transaction, localId: localId),
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        _showSnack('Gagal menyimpan transaksi: $e', isError: true);
        setState(() => _isSubmitting = false);
      }
    }
  }

  void _showSnack(String msg, {bool isError = false}) {
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg, style: GoogleFonts.outfit()),
        backgroundColor: isError ? Colors.red.shade600 : AppColors.primaryGreen,
      ),
    );
  }

  // ─────────────────────────────────────────────────
  // BUILD
  // ─────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return Scaffold(
        backgroundColor: const Color(0xFFF8FAFC),
        body: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              CircularProgressIndicator(color: AppColors.blueSales),
              const SizedBox(height: 16),
              Text(
                'Memuat data dari server...',
                style: GoogleFonts.outfit(
                  color: const Color(0xFF64748B),
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
      );
    }

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'Order Operasional',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.w900,
            fontSize: 18,
            color: Colors.white,
          ),
        ),
        backgroundColor: AppColors.blueSalesDark,
        elevation: 0,
        centerTitle: true,
        iconTheme: const IconThemeData(color: Colors.white),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh_rounded, color: Colors.white),
            tooltip: 'Refresh data dari server',
            onPressed: _fetchData,
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _fetchData,
        color: AppColors.blueSales,
        child: SafeArea(
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // ── Offline Banner ──
                if (_isOffline) _buildOfflineBanner(),

                // ── Info Toko ──
                _buildStoreSectionHeader(),
                const SizedBox(height: 12),
                _buildStorePicker(),
                const SizedBox(height: 32),

                // ── Keranjang ──
                _buildCartSectionHeader(),
                const SizedBox(height: 8),
                _buildCartList(),
                if (_cartItems.isNotEmpty) ...[
                  const SizedBox(height: 16),
                  _buildTotalCard(),
                ],
                const SizedBox(height: 32),

                // ── Metode Pembayaran ──
                _buildSectionTitle('METODE PEMBAYARAN', Icons.payment_rounded),
                const SizedBox(height: 12),
                _buildPaymentMethodSelector(),
                if (_paymentMethod == 'TRANSFER') ...[
                  const SizedBox(height: 16),
                  _buildBankSelector(),
                  const SizedBox(height: 16),
                  _buildTransferProofPicker(),
                ],
                const SizedBox(height: 32),

                // ── Tanda Tangan ──
                _buildSectionTitle('PENGESAHAN', Icons.draw_rounded),
                const SizedBox(height: 16),
                _buildSignatureCard(
                  label: 'Tanda Tangan Sales',
                  icon: Icons.badge_rounded,
                  iconColor: AppColors.blueSales,
                  isSigned: _salesSignatureBase64 != null,
                  onSave: (b64) {
                    setState(() => _salesSignatureBase64 = b64);
                    _showSnack('Tanda tangan sales disimpan');
                  },
                ),
                const SizedBox(height: 16),
                _buildSignatureCard(
                  label: 'Tanda Tangan Pelanggan',
                  icon: Icons.person_rounded,
                  iconColor: Colors.orange,
                  isSigned: _customerSignatureBase64 != null,
                  onSave: (b64) {
                    setState(() => _customerSignatureBase64 = b64);
                    _showSnack('Tanda tangan pelanggan disimpan');
                  },
                ),

                const SizedBox(height: 40),
                _buildSubmitButton(),
                const SizedBox(height: 24),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // ─────────────────────────────────────────────────
  // WIDGET BUILDERS
  // ─────────────────────────────────────────────────

  Widget _buildOfflineBanner() {
    return Container(
      margin: const EdgeInsets.only(bottom: 20),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF7ED),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFFDBA74)),
      ),
      child: Row(
        children: [
          const Icon(
            Icons.wifi_off_rounded,
            color: Color(0xFFF97316),
            size: 20,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Mode Offline',
                  style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w800,
                    fontSize: 13,
                    color: const Color(0xFFC2410C),
                  ),
                ),
                Text(
                  'Menampilkan data terakhir dari lokal. Transaksi akan otomatis terkirim saat ada internet.',
                  style: GoogleFonts.outfit(
                    fontSize: 11,
                    color: const Color(0xFF9A3412),
                  ),
                ),
              ],
            ),
          ),
          TextButton(
            onPressed: _fetchData,
            style: TextButton.styleFrom(
              padding: const EdgeInsets.symmetric(horizontal: 8),
              minimumSize: Size.zero,
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
            child: Text(
              'Coba Lagi',
              style: GoogleFonts.outfit(
                fontSize: 11,
                fontWeight: FontWeight.w800,
                color: const Color(0xFFF97316),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStoreSectionHeader() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        _buildSectionTitle('INFORMASI TOKO', Icons.store_rounded),
        if (_pendingVisits.isNotEmpty)
          TextButton.icon(
            onPressed: _showPendingVisitsPicker,
            icon: const Icon(Icons.access_time_rounded, size: 16),
            label: Text(
              'Pilih yg Pending',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w700,
                fontSize: 13,
              ),
            ),
            style: TextButton.styleFrom(
              foregroundColor: AppColors.warning,
              padding: EdgeInsets.zero,
              minimumSize: const Size(0, 0),
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
          ),
      ],
    );
  }

  Widget _buildStorePicker() {
    return InkWell(
      onTap: _showStorePicker,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFFE2E8F0)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.03),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: AppColors.blueSalesLight.withValues(alpha: 0.15),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.search_rounded,
                color: AppColors.blueSales,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _selectedStore?.name ?? 'Pilih Toko Pelanggan',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w800,
                      fontSize: 16,
                      color: _selectedStore != null
                          ? const Color(0xFF1E293B)
                          : const Color(0xFF94A3B8),
                    ),
                  ),
                  if (_selectedStore != null) ...[
                    const SizedBox(height: 4),
                    Text(
                      _selectedStore!.address,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: const Color(0xFF64748B),
                      ),
                    ),
                  ],
                ],
              ),
            ),
            const Icon(Icons.chevron_right_rounded, color: Color(0xFFCBD5E1)),
          ],
        ),
      ),
    );
  }

  Widget _buildCartSectionHeader() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        _buildSectionTitle('KERANJANG PESANAN', Icons.shopping_cart_rounded),
        TextButton.icon(
          onPressed: _showProductPicker,
          icon: const Icon(Icons.add_circle_outline_rounded, size: 18),
          label: Text(
            'Tambah',
            style: GoogleFonts.outfit(fontWeight: FontWeight.w800),
          ),
          style: TextButton.styleFrom(
            foregroundColor: AppColors.primaryGreen,
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
          ),
        ),
      ],
    );
  }

  Widget _buildCartList() {
    if (_cartItems.isEmpty) {
      return Container(
        width: double.infinity,
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFFE2E8F0)),
        ),
        child: Column(
          children: [
            const Icon(Icons.inbox_rounded, size: 48, color: Color(0xFFCBD5E1)),
            const SizedBox(height: 12),
            Text(
              'Belum ada produk',
              style: GoogleFonts.outfit(
                fontSize: 14,
                fontWeight: FontWeight.w700,
                color: const Color(0xFF64748B),
              ),
            ),
            Text(
              'Ketuk "Tambah" untuk memilih produk',
              style: GoogleFonts.outfit(
                fontSize: 12,
                color: const Color(0xFF94A3B8),
              ),
            ),
          ],
        ),
      );
    }

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE2E8F0)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: ListView.separated(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        itemCount: _cartItems.length,
        separatorBuilder: (_, __) =>
            const Divider(height: 1, color: Color(0xFFF1F5F9)),
        itemBuilder: (context, index) {
          final item = _cartItems[index];
          final p = item['product'] as Map<String, dynamic>;
          final id = p['id'];
          final qty = item['quantity'] as int;
          final basePrice = _getBasePriceForProduct(p);
          final tierLabel = _getPriceTierLabel(p);
          final hasCustomTier =
              hasFlexiblePricing &&
              (_selectedPriceTier[id] ??
                      (isSalesDistributor ? 'DBP' : 'NORMAL')) !=
                  'NORMAL';
          final badgeBg = isSalesDistributor
              ? Colors.indigo.shade50
              : Colors.purple.shade50;
          final badgeBorder = isSalesDistributor
              ? Colors.indigo.shade200
              : Colors.purple.shade200;
          final badgeText = isSalesDistributor
              ? Colors.indigo.shade800
              : Colors.purple.shade800;
          final tierPriceColor = hasCustomTier
              ? (isSalesDistributor
                    ? const Color(0xFF4338CA)
                    : Colors.purple.shade700)
              : const Color(0xFF64748B);

          final disc = _discountPercent[id] ?? 0.0;
          final effectivePrice = basePrice * (1 - disc / 100);
          final subtotal = effectivePrice * qty;

          return Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            p['name'] ?? '-',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800,
                              fontSize: 14,
                              color: const Color(0xFF1E293B),
                            ),
                          ),
                          const SizedBox(height: 4),
                          if (isSalesDistributor) ...[
                            Builder(
                              builder: (_) {
                                final cp = _customPrice[id] ?? 0.0;
                                if (cp > 0) {
                                  return Text(
                                    '${_currencyFormat.format(cp)} / ${p['unit'] ?? 'Pcs'}',
                                    style: GoogleFonts.outfit(
                                      fontSize: 12,
                                      fontWeight: FontWeight.w800,
                                      color: const Color(0xFF4338CA),
                                    ),
                                  );
                                }
                                return GestureDetector(
                                  onTap: () => _showCustomPriceDialog(p),
                                  child: Container(
                                    margin: const EdgeInsets.only(top: 2),
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 3,
                                    ),
                                    decoration: BoxDecoration(
                                      color: Colors.amber.shade50,
                                      borderRadius: BorderRadius.circular(6),
                                      border: Border.all(
                                        color: Colors.amber.shade300,
                                      ),
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(
                                          Icons.edit_note_rounded,
                                          size: 14,
                                          color: Colors.amber.shade900,
                                        ),
                                        const SizedBox(width: 4),
                                        Text(
                                          'Atur Harga Jual ✎',
                                          style: GoogleFonts.outfit(
                                            fontSize: 11,
                                            fontWeight: FontWeight.w800,
                                            color: Colors.amber.shade900,
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                );
                              },
                            ),
                          ] else if (disc > 0) ...[
                            Text(
                              '${_currencyFormat.format(basePrice)} / ${p['unit'] ?? 'Pcs'}',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                color: Colors.blueGrey.shade300,
                                decoration: TextDecoration.lineThrough,
                              ),
                            ),
                            Text(
                              '${_currencyFormat.format(effectivePrice)} / ${p['unit'] ?? 'Pcs'}',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                color: const Color(0xFFEF4444),
                              ),
                            ),
                          ] else
                            Text(
                              '${_currencyFormat.format(basePrice)} / ${p['unit'] ?? 'Pcs'}',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                fontWeight: hasCustomTier
                                    ? FontWeight.w800
                                    : FontWeight.w500,
                                color: tierPriceColor,
                              ),
                            ),
                          if (hasCustomTier && !isSalesDistributor)
                            Container(
                              margin: const EdgeInsets.only(top: 2),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 6,
                                vertical: 1.5,
                              ),
                              decoration: BoxDecoration(
                                color: badgeBg,
                                borderRadius: BorderRadius.circular(6),
                                border: Border.all(
                                  color: badgeBorder,
                                  width: 0.5,
                                ),
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
                          const SizedBox(height: 4),
                          Text(
                            _currencyFormat.format(subtotal),
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w900,
                              fontSize: 13,
                              color: AppColors.primaryGreen,
                            ),
                          ),
                        ],
                      ),
                    ),
                    Container(
                      decoration: BoxDecoration(
                        color: const Color(0xFFF1F5F9),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        children: [
                          IconButton(
                            icon: const Icon(
                              Icons.remove_rounded,
                              size: 18,
                              color: Color(0xFF64748B),
                            ),
                            onPressed: () => _updateQuantity(index, -1),
                            constraints: const BoxConstraints(
                              minWidth: 36,
                              minHeight: 36,
                            ),
                            padding: EdgeInsets.zero,
                          ),
                          GestureDetector(
                            onTap: () =>
                                _showEditQuantityDialog(context, index, qty),
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 12,
                                vertical: 8,
                              ),
                              color: Colors.transparent,
                              child: Text(
                                '$qty',
                                style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w900,
                                  fontSize: 14,
                                  decoration: TextDecoration.underline,
                                  decorationStyle: TextDecorationStyle.dashed,
                                ),
                              ),
                            ),
                          ),
                          IconButton(
                            icon: const Icon(
                              Icons.add_rounded,
                              size: 18,
                              color: AppColors.primaryGreen,
                            ),
                            onPressed: () => _updateQuantity(index, 1),
                            constraints: const BoxConstraints(
                              minWidth: 36,
                              minHeight: 36,
                            ),
                            padding: EdgeInsets.zero,
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                if (hasFlexiblePricing) ...[
                  _buildPriceTierRow(p),
                  const SizedBox(height: 6),
                ],
                const SizedBox(height: 8),
                _buildDiscountRow(productId: id),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildTotalCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.primaryGreen.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
          color: AppColors.primaryGreen.withValues(alpha: 0.3),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            'Total Pembayaran',
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w800,
              fontSize: 14,
              color: const Color(0xFF1E293B),
            ),
          ),
          Text(
            _currencyFormat.format(_totalAmount),
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w900,
              fontSize: 18,
              color: AppColors.primaryGreen,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPaymentMethodSelector() {
    return Row(
      children: [
        Expanded(
          child: _buildPaymentOption(
            label: 'Tunai',
            icon: Icons.payments_rounded,
            value: 'CASH',
            activeColor: AppColors.primaryGreen,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: _buildPaymentOption(
            label: 'Transfer',
            icon: Icons.account_balance_rounded,
            value: 'TRANSFER',
            activeColor: AppColors.blueSales,
          ),
        ),
      ],
    );
  }

  Widget _buildPaymentOption({
    required String label,
    required IconData icon,
    required String value,
    required Color activeColor,
  }) {
    final isSelected = _paymentMethod == value;
    return GestureDetector(
      onTap: () => setState(() => _paymentMethod = value),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(vertical: 16, horizontal: 12),
        decoration: BoxDecoration(
          color: isSelected ? activeColor.withValues(alpha: 0.1) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: isSelected ? activeColor : const Color(0xFFE2E8F0),
            width: isSelected ? 2 : 1,
          ),
          boxShadow: [
            if (isSelected)
              BoxShadow(
                color: activeColor.withValues(alpha: 0.15),
                blurRadius: 8,
                offset: const Offset(0, 4),
              ),
          ],
        ),
        child: Row(
          children: [
            Icon(
              icon,
              color: isSelected ? activeColor : const Color(0xFF94A3B8),
              size: 22,
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Text(
                label,
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800,
                  fontSize: 14,
                  color: isSelected ? activeColor : const Color(0xFF94A3B8),
                ),
              ),
            ),
            if (isSelected)
              Icon(Icons.check_circle_rounded, color: activeColor, size: 18),
          ],
        ),
      ),
    );
  }

  Widget _buildBankSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'PILIH REKENING TUJUAN TRANSFER',
          style: GoogleFonts.outfit(
            fontSize: 11,
            fontWeight: FontWeight.w900,
            color: Colors.blueGrey,
            letterSpacing: 1.2,
          ),
        ),
        const SizedBox(height: 8),
        if (_isLoadingBanks)
          const Center(
            child: Padding(
              padding: EdgeInsets.symmetric(vertical: 12),
              child: CircularProgressIndicator(),
            ),
          )
        else if (_bankAccounts.isNotEmpty) ...[
          Container(
            margin: const EdgeInsets.only(bottom: 12),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 4),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: Colors.indigo.shade200, width: 1.5),
              boxShadow: [
                BoxShadow(
                  color: Colors.indigo.withOpacity(0.05),
                  blurRadius: 8,
                  offset: const Offset(0, 2),
                ),
              ],
            ),
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                isExpanded: true,
                value: _selectedBankAccount?['id']?.toString(),
                hint: Text(
                  'Pilih Rekening Bank',
                  style: GoogleFonts.outfit(fontSize: 13, color: Colors.grey),
                ),
                icon: const Icon(
                  Icons.keyboard_arrow_down_rounded,
                  color: Colors.indigo,
                ),
                items: _bankAccounts.map((b) {
                  final bName = b['bank_name']?.toString() ?? 'Bank';
                  final accNum = b['account_number']?.toString() ?? '';
                  final companyName =
                      b['company']?['name']?.toString() ??
                      b['account_holder']?.toString() ??
                      '';
                  final branchName = b['branch']?['name']?.toString();
                  final tag = branchName != null && branchName.isNotEmpty
                      ? ' - $branchName'
                      : '';
                  final displayName = companyName.isNotEmpty
                      ? '$accNum - $companyName$tag ($bName)'
                      : '$accNum$tag ($bName)';
                  return DropdownMenuItem<String>(
                    value: b['id']?.toString(),
                    child: Row(
                      children: [
                        const Icon(
                          Icons.account_balance_rounded,
                          size: 18,
                          color: Colors.indigo,
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(
                            displayName,
                            style: GoogleFonts.outfit(
                              fontSize: 13,
                              fontWeight: FontWeight.w600,
                              color: const Color(0xFF1E293B),
                            ),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      ],
                    ),
                  );
                }).toList(),
                onChanged: (val) {
                  if (val != null) {
                    setState(() {
                      _selectedBankAccount = _bankAccounts.firstWhere(
                        (b) => b['id']?.toString() == val,
                        orElse: () => _bankAccounts.first,
                      );
                    });
                  }
                },
              ),
            ),
          ),
        ],
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.indigo.withOpacity(0.04),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.indigo.withOpacity(0.15)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  _selectedBankAccount?['bank_logo_url'] != null &&
                          (_selectedBankAccount!['bank_logo_url'] as String)
                              .isNotEmpty
                      ? CachedNetworkImage(
                          imageUrl:
                              _selectedBankAccount!['bank_logo_url'] as String,
                          width: 50,
                          height: 25,
                          placeholder: (context, url) => const Center(
                            child: CircularProgressIndicator(strokeWidth: 2),
                          ),
                          errorWidget: (context, url, error) => const Icon(
                            Icons.account_balance,
                            color: Colors.indigo,
                            size: 28,
                          ),
                          memCacheWidth: 800,
                        )
                      : const Icon(
                          Icons.account_balance_rounded,
                          color: Colors.indigo,
                          size: 28,
                        ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _selectedBankAccount?['bank_name'] as String? ??
                              'Bank Central Asia (BCA)',
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.bold,
                            fontSize: 14,
                            color: const Color(0xFF1E293B),
                          ),
                        ),
                        if (_selectedBankAccount?['company'] != null)
                          Text(
                            'PT: ${_selectedBankAccount!['company']['name']}',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              color: Colors.indigo.shade700,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        if (_selectedBankAccount?['branch'] != null)
                          Text(
                            'Cabang: ${_selectedBankAccount!['branch']['name']}',
                            style: GoogleFonts.outfit(
                              fontSize: 11,
                              color: Colors.blueGrey.shade700,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'NOMOR REKENING',
                        style: GoogleFonts.outfit(
                          fontSize: 9,
                          color: Colors.blueGrey,
                          fontWeight: FontWeight.w800,
                          letterSpacing: 0.5,
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        _selectedBankAccount?['account_number'] as String? ??
                            '817-529-3021',
                        style: GoogleFonts.outfit(
                          fontSize: 16,
                          fontWeight: FontWeight.w900,
                          color: const Color(0xFF1E293B),
                        ),
                      ),
                    ],
                  ),
                  IconButton(
                    icon: const Icon(Icons.copy_rounded, color: Colors.indigo),
                    tooltip: 'Salin Rekening',
                    onPressed: () {
                      final accNo =
                          (_selectedBankAccount?['account_number'] as String? ??
                                  '8175293021')
                              .replaceAll('-', '')
                              .replaceAll(' ', '');
                      Clipboard.setData(ClipboardData(text: accNo));
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(
                          content: Text('Nomor rekening disalin!'),
                          duration: Duration(seconds: 1),
                        ),
                      );
                    },
                  ),
                ],
              ),
              const Divider(height: 16),
              Text(
                'NAMA PENERIMA',
                style: GoogleFonts.outfit(
                  fontSize: 9,
                  color: Colors.blueGrey,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 0.5,
                ),
              ),
              const SizedBox(height: 2),
              Text(
                _selectedBankAccount?['account_holder'] as String? ??
                    'Perusahaan Demo',
                style: GoogleFonts.outfit(
                  fontSize: 13,
                  fontWeight: FontWeight.bold,
                  color: const Color(0xFF1E293B),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildTransferProofPicker() {
    return GestureDetector(
      onTap: _pickTransferProof,
      child: Container(
        width: double.infinity,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: _transferProofPath != null
                ? AppColors.primaryGreen
                : const Color(0xFFE2E8F0),
            width: _transferProofPath != null ? 2 : 1,
          ),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.02),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: _transferProofPath != null
            ? Stack(
                children: [
                  ClipRRect(
                    borderRadius: BorderRadius.circular(15),
                    child: Image.file(
                      File(_transferProofPath!),
                      width: double.infinity,
                      height: 200,
                      fit: BoxFit.cover,
                    ),
                  ),
                  Positioned(
                    top: 8,
                    right: 8,
                    child: GestureDetector(
                      onTap: () => setState(() => _transferProofPath = null),
                      child: Container(
                        padding: const EdgeInsets.all(6),
                        decoration: const BoxDecoration(
                          color: Colors.black54,
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(
                          Icons.close_rounded,
                          color: Colors.white,
                          size: 16,
                        ),
                      ),
                    ),
                  ),
                  Positioned(
                    bottom: 8,
                    right: 8,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: AppColors.primaryGreen,
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(
                            Icons.check_rounded,
                            color: Colors.white,
                            size: 14,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            'Bukti Tersimpan',
                            style: GoogleFonts.outfit(
                              color: Colors.white,
                              fontWeight: FontWeight.w700,
                              fontSize: 11,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              )
            : Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: const Color(0xFFEFF6FF),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.add_photo_alternate_rounded,
                        color: AppColors.blueSales,
                        size: 32,
                      ),
                    ),
                    const SizedBox(height: 12),
                    Text(
                      'Lampirkan Bukti Transfer (Opsional)',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w800,
                        fontSize: 14,
                        color: AppColors.blueSales,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'Bisa diupload sekarang atau nanti di sore hari',
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: const Color(0xFF94A3B8),
                      ),
                    ),
                  ],
                ),
              ),
      ),
    );
  }

  Widget _buildSignatureCard({
    required String label,
    required IconData icon,
    required Color iconColor,
    required bool isSigned,
    required void Function(String base64) onSave,
  }) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE2E8F0)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          Row(
            children: [
              Icon(icon, color: iconColor, size: 20),
              const SizedBox(width: 8),
              Text(
                label,
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w800,
                  fontSize: 14,
                  color: const Color(0xFF1E293B),
                ),
              ),
              const Spacer(),
              if (isSigned)
                const Icon(
                  Icons.check_circle_rounded,
                  color: AppColors.primaryGreen,
                  size: 20,
                ),
            ],
          ),
          const SizedBox(height: 12),
          SignatureCanvas(onSave: onSave),
        ],
      ),
    );
  }

  Widget _buildSubmitButton() {
    return SizedBox(
      width: double.infinity,
      height: 56,
      child: ElevatedButton(
        style: ElevatedButton.styleFrom(
          backgroundColor: AppColors.primaryGreen,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(16),
          ),
          elevation: 4,
          shadowColor: AppColors.primaryGreen.withValues(alpha: 0.4),
        ),
        onPressed: _isSubmitting ? null : _submitOrder,
        child: _isSubmitting
            ? const SizedBox(
                width: 24,
                height: 24,
                child: CircularProgressIndicator(
                  color: Colors.white,
                  strokeWidth: 3,
                ),
              )
            : Text(
                'SELESAI & SIMPAN TRANSAKSI',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  fontSize: 15,
                  color: Colors.white,
                  letterSpacing: 1,
                ),
              ),
      ),
    );
  }

  Widget _buildSectionTitle(String title, IconData icon) {
    return Row(
      children: [
        Icon(icon, size: 16, color: const Color(0xFF94A3B8)),
        const SizedBox(width: 8),
        Text(
          title,
          style: GoogleFonts.outfit(
            fontSize: 11,
            fontWeight: FontWeight.w900,
            color: const Color(0xFF64748B),
            letterSpacing: 1.5,
          ),
        ),
      ],
    );
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
                ? const Color(0xFFEF4444).withValues(alpha: 0.1)
                : const Color(0xFFF1F5F9),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(
            Icons.local_offer_rounded,
            size: 14,
            color: currentDisc > 0 ? const Color(0xFFEF4444) : Colors.blueGrey,
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'DISKON (%)',
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  fontWeight: FontWeight.w900,
                  color: currentDisc > 0
                      ? const Color(0xFFEF4444)
                      : Colors.blueGrey,
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
                    ? const Color(0xFFEF4444).withValues(alpha: 0.5)
                    : const Color(0xFFE2E8F0),
              ),
            ),
            child: Row(
              children: [
                const SizedBox(width: 8),
                Expanded(
                  child: TextField(
                    controller: _getDiscountCtrl(productId),
                    keyboardType: const TextInputType.numberWithOptions(
                      decimal: true,
                    ),
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
                        color: Colors.blueGrey.withValues(alpha: 0.5),
                      ),
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
}

// ─────────────────────────────────────────────────
// REUSABLE SEARCH PICKER BOTTOM SHEET
// ─────────────────────────────────────────────────

class _SearchPickerSheet<T> extends StatefulWidget {
  final String title;
  final List<T> items;
  final String searchHint;
  final Widget Function(BuildContext, T) itemBuilder;
  final bool Function(T, String) filterItem;
  final List<String>? categories;
  final String Function(T)? categoryExtractor;

  const _SearchPickerSheet({
    required this.title,
    required this.items,
    required this.searchHint,
    required this.itemBuilder,
    required this.filterItem,
    this.categories,
    this.categoryExtractor,
  });

  @override
  State<_SearchPickerSheet<T>> createState() => _SearchPickerSheetState<T>();
}

class _SearchPickerSheetState<T> extends State<_SearchPickerSheet<T>> {
  late List<T> _filteredItems;
  String _selectedCategory = 'Semua';
  String _currentQuery = '';

  @override
  void initState() {
    super.initState();
    _filteredItems = widget.items;
  }

  void _applyFilter() {
    setState(() {
      _filteredItems = widget.items.where((item) {
        final matchesQuery =
            _currentQuery.isEmpty || widget.filterItem(item, _currentQuery);
        final matchesCategory =
            _selectedCategory == 'Semua' ||
            (widget.categoryExtractor != null &&
                widget.categoryExtractor!(item).trim().toLowerCase() ==
                    _selectedCategory.trim().toLowerCase());
        return matchesQuery && matchesCategory;
      }).toList();
    });
  }

  void _onSearch(String query) {
    _currentQuery = query;
    _applyFilter();
  }

  void _onCategorySelected(String category) {
    _selectedCategory = category;
    _applyFilter();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      height: MediaQuery.of(context).size.height * 0.85,
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      child: Column(
        children: [
          const SizedBox(height: 12),
          Container(
            width: 40,
            height: 4,
            decoration: BoxDecoration(
              color: const Color(0xFFE2E8F0),
              borderRadius: BorderRadius.circular(2),
            ),
          ),
          const SizedBox(height: 16),
          Text(
            widget.title,
            style: GoogleFonts.outfit(
              fontSize: 18,
              fontWeight: FontWeight.w900,
              color: const Color(0xFF1E293B),
            ),
          ),
          const SizedBox(height: 16),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            child: TextField(
              onChanged: _onSearch,
              autofocus: true,
              decoration: InputDecoration(
                hintText: widget.searchHint,
                hintStyle: GoogleFonts.outfit(color: const Color(0xFF94A3B8)),
                prefixIcon: const Icon(
                  Icons.search_rounded,
                  color: Color(0xFF94A3B8),
                ),
                filled: true,
                fillColor: const Color(0xFFF8FAFC),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(16),
                  borderSide: BorderSide.none,
                ),
                contentPadding: const EdgeInsets.symmetric(vertical: 12),
              ),
              style: GoogleFonts.outfit(),
            ),
          ),
          if (widget.categories != null && widget.categories!.isNotEmpty) ...[
            const SizedBox(height: 10),
            SizedBox(
              height: 36,
              child: ListView.separated(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                scrollDirection: Axis.horizontal,
                itemCount: widget.categories!.length,
                separatorBuilder: (_, __) => const SizedBox(width: 8),
                itemBuilder: (context, index) {
                  final cat = widget.categories![index];
                  final isSelected = _selectedCategory == cat;
                  return GestureDetector(
                    onTap: () => _onCategorySelected(cat),
                    child: AnimatedContainer(
                      duration: const Duration(milliseconds: 200),
                      padding: const EdgeInsets.symmetric(
                        horizontal: 14,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: isSelected
                            ? AppColors.primaryGreen
                            : const Color(0xFFF1F5F9),
                        borderRadius: BorderRadius.circular(20),
                        border: isSelected
                            ? null
                            : Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Center(
                        child: Text(
                          cat,
                          style: GoogleFonts.outfit(
                            fontSize: 12,
                            fontWeight: isSelected
                                ? FontWeight.w700
                                : FontWeight.w600,
                            color: isSelected
                                ? Colors.white
                                : const Color(0xFF64748B),
                          ),
                        ),
                      ),
                    ),
                  );
                },
              ),
            ),
          ],
          const SizedBox(height: 12),
          Expanded(
            child: _filteredItems.isEmpty
                ? Center(
                    child: Text(
                      'Tidak ada hasil ditemukan',
                      style: GoogleFonts.outfit(color: const Color(0xFF94A3B8)),
                    ),
                  )
                : ListView.separated(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 8,
                    ),
                    itemCount: _filteredItems.length,
                    separatorBuilder: (_, __) =>
                        const Divider(height: 1, color: Color(0xFFF1F5F9)),
                    itemBuilder: (context, index) =>
                        widget.itemBuilder(context, _filteredItems[index]),
                  ),
          ),
        ],
      ),
    );
  }
}
