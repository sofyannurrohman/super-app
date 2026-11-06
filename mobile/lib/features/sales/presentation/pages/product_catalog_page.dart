import 'package:cached_network_image/cached_network_image.dart';
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/features/sales/data/models/product_model.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sync/data/repositories/sync_repository.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:intl/intl.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:share_plus/share_plus.dart';
import 'package:path_provider/path_provider.dart';
import 'dart:io';
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';

class ProductCatalogPage extends StatefulWidget {
  const ProductCatalogPage({super.key});

  @override
  State<ProductCatalogPage> createState() => _ProductCatalogPageState();
}

class _ProductCatalogPageState extends State<ProductCatalogPage> {
  final _apiService = SalesApiService(apiClient: di.sl<ApiClient>());
  final TextEditingController _searchController = TextEditingController();
  bool _isLoading = true;
  bool _isGridView = false; // Simple & readable list view as default
  bool _isDistributor = false;
  List<ProductModel> _allProducts = [];
  List<ProductModel> _filteredProducts = [];
  String? _error;

  @override
  void initState() {
    super.initState();
    _checkDistributorAndFetch();
  }

  Future<void> _checkDistributorAndFetch() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final cachedProfile = prefs.getString('cached_employee_profile');
      if (cachedProfile != null) {
        final profile = jsonDecode(cachedProfile) as Map<String, dynamic>;
        final isDist =
            profile['is_distributor'] == true ||
            (profile['distributor_store_id'] != null &&
                profile['distributor_store_id'].toString().isNotEmpty) ||
            (profile['job_position']?['title'] ?? '')
                .toString()
                .toLowerCase()
                .contains('distributor');
        if (isDist) {
          if (mounted) {
            setState(() {
              _isDistributor = true;
              _isLoading = false;
            });
          }
          return;
        }
      }
    } catch (_) {}
    _fetchProducts();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _fetchProducts() async {
    if (!mounted) return;
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      final db = di.sl<AppDatabase>();

      // Pull latest master data from backend (fails silently if offline)
      await di.sl<SyncRepository>().pullMasterData();

      final products = await db.select(db.products).get();
      if (mounted) {
        setState(() {
          _allProducts = products
              .map(
                (p) => ProductModel(
                  id: p.id,
                  name: p.name,
                  sellingPrice: p.sellingPrice,
                  sku: p.sku ?? '',
                  unit: p.unit ?? '',
                  category: p.category ?? '',
                  brand: 'Karsa HR',
                  description: '',
                  imageUrl: p.imageUrl != null && p.imageUrl!.isNotEmpty
                      ? (p.imageUrl!.startsWith('http') ||
                                p.imageUrl!.startsWith('assets/')
                            ? p.imageUrl!
                            : '${AppConstants.baseUrl.replaceAll('/api/v1/', '')}${p.imageUrl}')
                      : '',
                  warehouseStock: p.warehouseStock,
                ),
              )
              .toList();
          _filteredProducts = List.from(_allProducts);
        });
      }
    } catch (e) {
      if (mounted) setState(() => _error = e.toString());
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _onSearchChanged(String query) {
    setState(() {
      final q = query.trim().toLowerCase();
      if (q == '') {
        _filteredProducts = List.from(_allProducts);
      } else {
        _filteredProducts = _allProducts.where((p) {
          final name = p.name.toLowerCase();
          final sku = p.sku.toLowerCase();
          final cat = p.category.toLowerCase();
          return name.contains(q) || sku.contains(q) || cat.contains(q);
        }).toList();
      }
    });
  }

  Widget _buildProductImage(ProductModel product) {
    if (product.imageUrl.isNotEmpty) {
      if (product.imageUrl.startsWith('http')) {
        return CachedNetworkImage(
          imageUrl: product.imageUrl,
          fit: BoxFit.contain,
          placeholder: (context, url) =>
              const Center(child: CircularProgressIndicator()),
          errorWidget: (context, url, error) => const Icon(Icons.error),
          memCacheWidth: 800,
        );
      } else if (product.imageUrl.startsWith('assets/')) {
        return Image.asset(product.imageUrl, fit: BoxFit.contain);
      }
    }

    final name = product.name.toLowerCase();
    String? localAsset;

    // Smartly map available premium assets to the matching product keywords
    if (name.contains('manis')) {
      localAsset = 'assets/brand.png';
    } else if (name.contains('asin')) {
      localAsset = 'assets/brand.png';
    } else if (name.contains('sambal') || name.contains('saus')) {
      localAsset = 'assets/brand.png';
    }

    if (localAsset != null) {
      return Image.asset(localAsset, fit: BoxFit.contain);
    }

    return _buildPlaceholder();
  }

  Widget _buildPlaceholder() {
    return Container(
      color: const Color(0xFFF1F5F9),
      child: const Icon(
        Icons.image_not_supported_outlined,
        size: 28,
        color: Colors.blueGrey,
      ),
    );
  }

  Future<void> _showShareCatalogBottomSheet() async {
    List<StoreModel> stores = [];

    setState(() => _isLoading = true);
    try {
      final storeApi = StoreApiService(
        apiClient: di.sl<ApiClient>(),
        prefs: di.sl<SharedPreferences>(),
      );
      final apiStores = await storeApi.getStores();
      stores = apiStores;
    } catch (e) {
      debugPrint('Error fetching stores for sharing: $e');
    }
    setState(() => _isLoading = false);

    if (!mounted) return;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        return _ShareCatalogBottomSheet(stores: stores);
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_isDistributor) {
      return Scaffold(
        backgroundColor: const Color(0xFFF8FAFC),
        appBar: AppBar(
          title: Text(
            'KATALOG PRODUK',
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w900,
              fontSize: 16,
              letterSpacing: 1.2,
              color: const Color(0xFF1E293B),
            ),
          ),
          backgroundColor: Colors.white,
          elevation: 0,
          leading: IconButton(
            icon: const Icon(
              Icons.arrow_back_ios_new_rounded,
              color: Colors.black87,
            ),
            onPressed: () => Navigator.pop(context),
          ),
        ),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(32.0),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Container(
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    color: Colors.amber.shade50,
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    Icons.lock_person_rounded,
                    size: 64,
                    color: Colors.amber.shade800,
                  ),
                ),
                const SizedBox(height: 24),
                Text(
                  'Akses Dibatasi',
                  style: GoogleFonts.outfit(
                    fontSize: 20,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B),
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  'Menu katalog produk dan daftar harga resmi pabrik tidak tersedia untuk akun distributor.',
                  textAlign: TextAlign.center,
                  style: GoogleFonts.outfit(
                    fontSize: 14,
                    color: const Color(0xFF64748B),
                    height: 1.5,
                  ),
                ),
                const SizedBox(height: 24),
                ElevatedButton.icon(
                  onPressed: () => Navigator.pop(context),
                  icon: const Icon(
                    Icons.arrow_back_rounded,
                    color: Colors.white,
                  ),
                  label: Text(
                    'Kembali ke Dashboard',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w700,
                      color: Colors.white,
                    ),
                  ),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.primaryGreen,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 24,
                      vertical: 12,
                    ),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      body: RefreshIndicator(
        onRefresh: _fetchProducts,
        color: AppColors.primaryGreen,
        child: CustomScrollView(
          slivers: [
            _buildSliverAppBar(),
            if (_isLoading)
              const SliverFillRemaining(
                child: Center(
                  child: CircularProgressIndicator(
                    color: AppColors.primaryGreen,
                  ),
                ),
              )
            else if (_error != null)
              _buildErrorState()
            else if (_filteredProducts.isEmpty)
              _buildEmptyState()
            else if (_isGridView)
              _buildProductGrid()
            else
              _buildProductList(),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _showShareCatalogBottomSheet,
        backgroundColor: AppColors.primaryGreen,
        elevation: 8,
        icon: const Icon(Icons.share_rounded, color: Colors.white, size: 20),
        label: Text(
          'KIRIM KATALOG WA',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.w900,
            color: Colors.white,
            fontSize: 13,
            letterSpacing: 0.5,
          ),
        ),
      ),
    );
  }

  Widget _buildSliverAppBar() {
    return SliverAppBar(
      expandedHeight: 140,
      floating: false,
      pinned: true,
      backgroundColor: AppColors.primaryGreenDark, // Dark Karsa HR Green
      elevation: 0,
      leading: IconButton(
        icon: const Icon(
          Icons.arrow_back_ios_rounded,
          color: Colors.white,
          size: 20,
        ),
        onPressed: () => Navigator.pop(context),
      ),
      actions: [
        IconButton(
          icon: Icon(
            _isGridView ? Icons.view_list_rounded : Icons.grid_view_rounded,
            color: Colors.white,
            size: 22,
          ),
          tooltip: _isGridView
              ? 'Ubah ke Tampilan List'
              : 'Ubah ke Tampilan Grid',
          onPressed: () {
            setState(() {
              _isGridView = !_isGridView;
            });
          },
        ),
        const SizedBox(width: 8),
      ],
      flexibleSpace: FlexibleSpaceBar(
        background: Stack(
          fit: StackFit.expand,
          children: [
            Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(
                  colors: AppColors.primaryGradient,
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
              ),
            ),
            Positioned(
              right: -20,
              top: -20,
              child: Icon(
                Icons.inventory_2_rounded,
                size: 150,
                color: Colors.white.withOpacity(0.05),
              ),
            ),
          ],
        ),
        title: Text(
          'KATALOG PRODUK',
          style: GoogleFonts.outfit(
            fontSize: 16,
            fontWeight: FontWeight.w900,
            color: Colors.white,
            letterSpacing: 1,
          ),
        ),
        titlePadding: const EdgeInsets.only(left: 56, bottom: 64),
      ),
      bottom: PreferredSize(
        preferredSize: const Size.fromHeight(60),
        child: Container(
          padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
          child: Container(
            height: 48,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.1),
                  blurRadius: 10,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: TextField(
              controller: _searchController,
              onChanged: _onSearchChanged,
              decoration: InputDecoration(
                hintText: 'Cari nama produk atau SKU...',
                hintStyle: GoogleFonts.outfit(
                  fontSize: 14,
                  color: Colors.blueGrey.shade300,
                ),
                prefixIcon: const Icon(
                  Icons.search_rounded,
                  color: AppColors.primaryGreen,
                ),
                border: InputBorder.none,
                contentPadding: const EdgeInsets.symmetric(vertical: 14),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildProductList() {
    return SliverPadding(
      padding: const EdgeInsets.fromLTRB(
        16,
        12,
        16,
        80,
      ), // extra bottom padding for FAB
      sliver: SliverList(
        delegate: SliverChildBuilderDelegate(
          (context, index) => _buildListProductCard(_filteredProducts[index]),
          childCount: _filteredProducts.length,
        ),
      ),
    );
  }

  Widget _buildListProductCard(ProductModel product) {
    final currencyFormat = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    Color stockColor = const Color(0xFF10B981); // Available
    String stockLabel = 'Tersedia';
    if (product.warehouseStock <= 0) {
      stockColor = const Color(0xFFEF4444); // Out of stock
      stockLabel = 'Habis';
    } else if (product.warehouseStock < 10) {
      stockColor = const Color(0xFFF59E0B); // Low stock
      stockLabel = 'Stok Menipis';
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(18),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Row(
        children: [
          // Image on the Left
          Container(
            width: 80,
            height: 80,
            decoration: BoxDecoration(
              color: const Color(0xFFF8FAFC),
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: const Color(0xFFE2E8F0)),
            ),
            padding: const EdgeInsets.all(6),
            child: GestureDetector(
              onTap: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(
                    builder: (_) => ProductImageViewerPage(
                      product: product,
                      heroTag: 'list_${product.id}',
                    ),
                  ),
                );
              },
              child: Hero(
                tag: 'list_${product.id}',
                child: _buildProductImage(product),
              ),
            ),
          ),
          const SizedBox(width: 14),
          // Info in the Center & Right
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 6,
                        vertical: 3,
                      ),
                      decoration: BoxDecoration(
                        color: const Color(0xFFE2E8F0),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Text(
                        product.category.toUpperCase(),
                        style: GoogleFonts.outfit(
                          fontSize: 8,
                          fontWeight: FontWeight.w900,
                          color: const Color(0xFF475569),
                        ),
                      ),
                    ),
                    Text(
                      product.sku,
                      style: GoogleFonts.outfit(
                        fontSize: 9,
                        color: Colors.blueGrey.shade400,
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 6),
                Text(
                  product.name,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: GoogleFonts.outfit(
                    fontSize: 14,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B),
                    height: 1.25,
                  ),
                ),
                const SizedBox(height: 8),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      currencyFormat.format(product.sellingPrice),
                      style: GoogleFonts.outfit(
                        fontSize: 15,
                        fontWeight: FontWeight.w900,
                        color: AppColors.primaryGreen,
                      ),
                    ),
                    Row(
                      children: [
                        Container(
                          width: 6,
                          height: 6,
                          decoration: BoxDecoration(
                            color: stockColor,
                            shape: BoxShape.circle,
                          ),
                        ),
                        const SizedBox(width: 6),
                        Text(
                          '$stockLabel (${product.warehouseStock})',
                          style: GoogleFonts.outfit(
                            fontSize: 10,
                            fontWeight: FontWeight.w800,
                            color: stockColor,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildProductGrid() {
    return SliverPadding(
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 80),
      sliver: SliverGrid(
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 2,
          mainAxisSpacing: 12,
          crossAxisSpacing: 12,
          childAspectRatio: 0.68,
        ),
        delegate: SliverChildBuilderDelegate(
          (context, index) =>
              _buildCompactProductCard(_filteredProducts[index]),
          childCount: _filteredProducts.length,
        ),
      ),
    );
  }

  Widget _buildCompactProductCard(ProductModel product) {
    final currencyFormat = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    Color stockColor = const Color(0xFF10B981);
    String stockLabel = 'Tersedia';
    if (product.warehouseStock <= 0) {
      stockColor = const Color(0xFFEF4444);
      stockLabel = 'Habis';
    } else if (product.warehouseStock < 10) {
      stockColor = const Color(0xFFF59E0B);
      stockLabel = 'Stok Menipis';
    }

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      clipBehavior: Clip.antiAlias,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            flex: 12,
            child: Container(
              width: double.infinity,
              color: const Color(0xFFF8FAFC),
              child: Stack(
                children: [
                  Center(
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: GestureDetector(
                        onTap: () {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (_) => ProductImageViewerPage(
                                product: product,
                                heroTag: 'grid_${product.id}',
                              ),
                            ),
                          );
                        },
                        child: Hero(
                          tag: 'grid_${product.id}',
                          child: _buildProductImage(product),
                        ),
                      ),
                    ),
                  ),
                  Positioned(
                    top: 8,
                    left: 8,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.9),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        product.category,
                        style: GoogleFonts.outfit(
                          fontSize: 8,
                          fontWeight: FontWeight.w900,
                          color: const Color(0xFF64748B),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
          Expanded(
            flex: 11,
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        product.name,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                        style: GoogleFonts.outfit(
                          fontSize: 13,
                          fontWeight: FontWeight.w800,
                          color: const Color(0xFF1E293B),
                          height: 1.2,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        product.sku,
                        style: GoogleFonts.outfit(
                          fontSize: 9,
                          color: const Color(0xFF94A3B8),
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ],
                  ),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        currencyFormat.format(product.sellingPrice),
                        style: GoogleFonts.outfit(
                          fontSize: 14,
                          fontWeight: FontWeight.w900,
                          color: AppColors.primaryGreen,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Row(
                        children: [
                          Container(
                            width: 6,
                            height: 6,
                            decoration: BoxDecoration(
                              color: stockColor,
                              shape: BoxShape.circle,
                            ),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            '$stockLabel (${product.warehouseStock})',
                            style: GoogleFonts.outfit(
                              fontSize: 10,
                              fontWeight: FontWeight.w800,
                              color: stockColor,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorState() {
    return SliverFillRemaining(
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.error_outline_rounded,
              color: Color(0xFFEF4444),
              size: 48,
            ),
            const SizedBox(height: 16),
            Text(
              'Gagal memuat produk',
              style: GoogleFonts.outfit(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Text(
              _error!,
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(fontSize: 12, color: Colors.blueGrey),
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: _fetchProducts,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.primaryGreen,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              child: const Text(
                'Coba Lagi',
                style: TextStyle(color: Colors.white),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return SliverFillRemaining(
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.search_off_rounded,
              color: Colors.blueGrey.shade200,
              size: 64,
            ),
            const SizedBox(height: 16),
            Text(
              'Produk tidak ditemukan',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w800,
                color: const Color(0xFF1E293B),
                fontSize: 18,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Coba kata kunci lain atau perbarui katalog',
              style: GoogleFonts.outfit(fontSize: 13, color: Colors.blueGrey),
            ),
          ],
        ),
      ),
    );
  }
}

class _ShareCatalogBottomSheet extends StatefulWidget {
  final List<StoreModel> stores;
  const _ShareCatalogBottomSheet({required this.stores});

  @override
  State<_ShareCatalogBottomSheet> createState() =>
      _ShareCatalogBottomSheetState();
}

class _ShareCatalogBottomSheetState extends State<_ShareCatalogBottomSheet> {
  StoreModel? _selectedStore;
  final TextEditingController _phoneCtrl = TextEditingController();
  final TextEditingController _customMessageCtrl = TextEditingController();

  @override
  void initState() {
    super.initState();
    _customMessageCtrl.text =
        "Halo, silakan lihat produk yang tersedia di katalog aplikasi kami.";
  }

  @override
  void dispose() {
    _phoneCtrl.dispose();
    _customMessageCtrl.dispose();
    super.dispose();
  }

  Future<void> _onStoreSelected(StoreModel? store) async {
    setState(() {
      _selectedStore = store;
    });

    if (store != null) {
      final prefs = await SharedPreferences.getInstance();
      final savedPhone = prefs.getString('store_wa_${store.id}');
      if (savedPhone != null && savedPhone.isNotEmpty) {
        setState(() {
          _phoneCtrl.text = savedPhone;
        });
      } else {
        setState(() {
          _phoneCtrl.clear();
        });
      }
      setState(() {
        _customMessageCtrl.text =
            "Halo *${store.ownerName}* dari *${store.name}*, silakan lihat produk yang tersedia di katalog aplikasi kami.";
      });
    }
  }

  Future<void> _sendWhatsApp() async {
    final rawPhone = _phoneCtrl.text.trim();
    if (rawPhone.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Nomor WhatsApp wajib diisi'),
          backgroundColor: Colors.amber,
        ),
      );
      return;
    }

    String phone = rawPhone.replaceAll(RegExp(r'\D'), '');
    if (phone.startsWith('0')) {
      phone = '62${phone.substring(1)}';
    } else if (phone.startsWith('+')) {
      phone = phone.substring(1);
    }
    if (!phone.startsWith('62') && !phone.startsWith('6')) {
      phone = '62$phone';
    }

    if (_selectedStore != null) {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('store_wa_${_selectedStore!.id}', rawPhone);
    }

    final message = _customMessageCtrl.text.trim();

    try {
      await Share.share(message);
      if (mounted) Navigator.pop(context);
    } catch (e) {
      // Fallback jika gambar gagal dimuat, luncurkan WA hanya dengan teks
      final url = Uri.parse(
        "https://wa.me/$phone?text=${Uri.encodeComponent(message)}",
      );
      try {
        if (await canLaunchUrl(url)) {
          await launchUrl(url, mode: LaunchMode.externalApplication);
          if (mounted) Navigator.pop(context);
        } else {
          throw Exception(
            'Tidak dapat membuka WhatsApp. Pastikan aplikasi WA terinstal.',
          );
        }
      } catch (innerE) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Gagal meluncurkan WA: $innerE'),
              backgroundColor: Colors.redAccent,
            ),
          );
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      padding: EdgeInsets.only(
        left: 24,
        right: 24,
        top: 24,
        bottom: MediaQuery.of(context).viewInsets.bottom + 24,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'KIRIM KATALOG WA',
                style: GoogleFonts.outfit(
                  fontSize: 16,
                  fontWeight: FontWeight.w900,
                  color: AppColors.primaryGreenDark,
                ),
              ),
              IconButton(
                icon: const Icon(Icons.close_rounded, color: Colors.blueGrey),
                onPressed: () => Navigator.pop(context),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Text(
            'PILIH TOKO / PELANGGAN',
            style: GoogleFonts.outfit(
              fontSize: 10,
              fontWeight: FontWeight.w800,
              color: Colors.blueGrey.shade400,
              letterSpacing: 1,
            ),
          ),
          const SizedBox(height: 8),
          DropdownButtonFormField<StoreModel>(
            initialValue: _selectedStore,
            hint: Text(
              'Pilih Toko Terdaftar (Opsional)',
              style: GoogleFonts.outfit(fontSize: 13),
            ),
            items: widget.stores
                .map(
                  (s) => DropdownMenuItem(
                    value: s,
                    child: Text(
                      '${s.name} (${s.ownerName})',
                      style: GoogleFonts.outfit(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                )
                .toList(),
            onChanged: _onStoreSelected,
            decoration: InputDecoration(
              prefixIcon: const Icon(
                Icons.storefront_rounded,
                color: AppColors.primaryGreen,
                size: 20,
              ),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
              ),
              contentPadding: const EdgeInsets.symmetric(horizontal: 16),
            ),
          ),
          const SizedBox(height: 20),
          Text(
            'NOMOR WHATSAPP *',
            style: GoogleFonts.outfit(
              fontSize: 10,
              fontWeight: FontWeight.w800,
              color: Colors.blueGrey.shade400,
              letterSpacing: 1,
            ),
          ),
          const SizedBox(height: 8),
          TextFormField(
            controller: _phoneCtrl,
            keyboardType: TextInputType.phone,
            style: GoogleFonts.outfit(
              fontSize: 14,
              fontWeight: FontWeight.w600,
            ),
            decoration: InputDecoration(
              hintText: 'Contoh: 08123456789',
              prefixIcon: const Icon(
                Icons.phone_iphone_rounded,
                color: AppColors.primaryGreen,
                size: 20,
              ),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
              ),
              contentPadding: const EdgeInsets.symmetric(horizontal: 16),
            ),
          ),
          const SizedBox(height: 20),
          Text(
            'PESAN WHATSAPP',
            style: GoogleFonts.outfit(
              fontSize: 10,
              fontWeight: FontWeight.w800,
              color: Colors.blueGrey.shade400,
              letterSpacing: 1,
            ),
          ),
          const SizedBox(height: 8),
          TextFormField(
            controller: _customMessageCtrl,
            maxLines: 3,
            style: GoogleFonts.outfit(
              fontSize: 13,
              fontWeight: FontWeight.w600,
              color: const Color(0xFF334155),
            ),
            decoration: InputDecoration(
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
              ),
              contentPadding: const EdgeInsets.all(12),
            ),
          ),
          const SizedBox(height: 24),
          SizedBox(
            width: double.infinity,
            height: 52,
            child: ElevatedButton.icon(
              onPressed: _sendWhatsApp,
              icon: const Icon(
                Icons.send_rounded,
                color: Colors.white,
                size: 20,
              ),
              label: Text(
                'LUNCURKAN WHATSAPP',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  color: Colors.white,
                  letterSpacing: 0.5,
                ),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF059669),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                elevation: 0,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class ProductImageViewerPage extends StatelessWidget {
  final ProductModel product;
  final String heroTag;

  const ProductImageViewerPage({
    Key? key,
    required this.product,
    required this.heroTag,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    Widget imageWidget = _getProductImageWidget(product);

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
        title: Text(
          product.name,
          style: GoogleFonts.outfit(color: Colors.white, fontSize: 16),
        ),
      ),
      body: Center(
        child: InteractiveViewer(
          minScale: 0.5,
          maxScale: 4.0,
          child: Hero(tag: heroTag, child: imageWidget),
        ),
      ),
    );
  }

  Widget _getProductImageWidget(ProductModel product) {
    if (product.imageUrl.isNotEmpty) {
      if (product.imageUrl.startsWith('http')) {
        return CachedNetworkImage(
          imageUrl: product.imageUrl,
          fit: BoxFit.contain,
          placeholder: (context, url) =>
              const Center(child: CircularProgressIndicator()),
          errorWidget: (context, url, error) => const Icon(Icons.error),
          memCacheWidth: 800,
        );
      } else if (product.imageUrl.startsWith('assets/')) {
        return Image.asset(product.imageUrl, fit: BoxFit.contain);
      }
    }

    final name = product.name.toLowerCase();
    String? localAsset;

    if (name.contains('manis')) {
      localAsset = 'assets/brand.png';
    } else if (name.contains('asin')) {
      localAsset = 'assets/brand.png';
    } else if (name.contains('sambal') || name.contains('saus')) {
      localAsset = 'assets/brand.png';
    }

    if (localAsset != null) {
      return Image.asset(localAsset, fit: BoxFit.contain);
    }
    return Container(
      color: const Color(0xFFF1F5F9),
      child: const Icon(
        Icons.image_not_supported_outlined,
        size: 64,
        color: Colors.blueGrey,
      ),
    );
  }
}
