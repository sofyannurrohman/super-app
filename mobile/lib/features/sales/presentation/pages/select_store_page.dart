import 'package:intl/intl.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/sales/data/models/store_model.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/presentation/pages/add_store_page.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_checkin_page.dart';
import 'package:hris_app/features/sales/presentation/pages/receipt_camera_page.dart';
import 'package:hris_app/features/sales/data/services/store_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/features/auth/domain/repositories/auth_repository.dart';

class SelectStorePage extends StatefulWidget {
  final bool isQuickOrder;
  const SelectStorePage({super.key, this.isQuickOrder = false});

  @override
  State<SelectStorePage> createState() => _SelectStorePageState();
}

class _SelectStorePageState extends State<SelectStorePage> {
  final _apiService = StoreApiService(apiClient: di.sl<ApiClient>(), prefs: di.sl<SharedPreferences>());
  bool _isLoading = true;
  List<StoreModel> _stores = [];
  String _jobPositionTitle = '';

  // Filter: 0 = Semua, 1 = Toko Baru, 2 = Toko Lama
  int _selectedFilter = 0;

  List<StoreModel> get _filteredStores {
    switch (_selectedFilter) {
      case 1:
        return _stores.where((s) => s.isNew).toList();
      case 2:
        return _stores.where((s) => !s.isNew).toList();
      default:
        return _stores;
    }
  }

  @override
  void initState() {
    super.initState();

    // Phase 1: Read from cache SYNCHRONOUSLY — no async, no spinner, instant!
    final cached = _apiService.getStoresCached();
    if (cached != null && cached.isNotEmpty) {
      _stores = cached;
      _isLoading = false; // Spinner never shows if cache exists
    }

    // Phase 2: Background refresh — only shows spinner if cache was empty
    _fetchStores();
    _fetchProfile();
  }

  Future<void> _fetchProfile() async {
    try {
      final authRepo = di.sl<AuthRepository>();
      final result = await authRepo.getProfile();
      result.fold(
        (_) => null,
        (profile) {
          if (mounted) {
            setState(() {
              _jobPositionTitle = profile['job_position']?['title'] ?? '';
            });
          }
        },
      );
    } catch (_) {}
  }

  Future<void> _fetchStores() async {
    // If we already have data from cache, background-refresh silently without
    // touching _isLoading or rebuilding the list (avoids visible flicker).
    if (_stores.isNotEmpty) {
      _apiService.getStores(sortBy: 'last_visit_asc').then((fresh) {
        if (mounted && fresh.isNotEmpty) {
          setState(() => _stores = fresh);
        }
      }).ignore();
      return;
    }

    // Try reading from local Drift SQLite database before locking UI behind spinner
    try {
      final localStores = await _apiService.getStoresFromLocalDb();
      if (mounted && localStores.isNotEmpty) {
        setState(() {
          _stores = localStores;
          _isLoading = false;
        });
        // Silent background refresh
        _apiService.getStores(sortBy: 'last_visit_asc').then((fresh) {
          if (mounted && fresh.isNotEmpty) {
            setState(() => _stores = fresh);
          }
        }).ignore();
        return;
      }
    } catch (_) {}

    // Cache & Local DB were both empty (first clean install) — use paginated fetch with incremental display.
    setState(() => _isLoading = true);
    try {
      await _apiService.getStoresPaginated(
        sortBy: 'last_visit_asc',
        pageSize: 150,
        onPartialLoad: (partial) {
          if (mounted) {
            setState(() {
              _stores = List.from(partial);
              _isLoading = false; // Kill spinner as soon as first batch arrives
            });
          }
        },
      );
    } catch (e) {
      debugPrint('Error fetching stores: $e');
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
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
            Text(widget.isQuickOrder ? 'LANGKAH 1 DARI 3' : 'LANGKAH 1 DARI 4',
                style: GoogleFonts.outfit(
                    fontSize: 10,
                    color: Colors.blueAccent,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1)),
            Text(
                widget.isQuickOrder
                    ? 'Pilih Toko Pesanan'
                    : 'Pilih Toko Kunjungan',
                style: GoogleFonts.outfit(
                    fontSize: 17,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B))),
          ],
        ),
      ),
      body: Column(
        children: [
          // Progress indicator
          LinearProgressIndicator(
            value: widget.isQuickOrder ? 0.33 : 0.25,
            minHeight: 4,
            backgroundColor: Colors.grey.shade200,
            color: Colors.blueAccent,
          ),

          // Filter chips
          if (!_isLoading) _buildFilterBar(),

          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _filteredStores.isEmpty
                    ? _buildEmptyStateFiltered()
                    : Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Padding(
                            padding: const EdgeInsets.fromLTRB(20, 12, 20, 0),
                            child: Text(
                              '${_filteredStores.length} toko ditemukan',
                              style: GoogleFonts.outfit(
                                fontSize: 12,
                                color: Colors.blueGrey,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                          Expanded(
                            child: ListView.builder(
                              padding: const EdgeInsets.all(20),
                              itemCount: _filteredStores.length,
                              itemBuilder: (ctx, i) =>
                                  _buildStoreCard(_filteredStores[i]),
                            ),
                          ),
                        ],
                      ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () async {
          final newStore = await Navigator.push<StoreModel>(
            context,
            MaterialPageRoute(builder: (_) => const AddStorePage()),
          );
          if (newStore != null) {
            setState(() => _stores.add(newStore));
          }
        },
        backgroundColor: Colors.blueAccent,
        icon: const Icon(Icons.add_rounded, color: Colors.white),
        label: Text('Tambah Toko Baru',
            style: GoogleFonts.outfit(
                color: Colors.white, fontWeight: FontWeight.w800)),
      ),
    );
  }

  Widget _buildFilterBar() {
    final filters = [
      _FilterItem(
        label: 'Semua',
        count: _stores.length,
        color: Colors.blueAccent,
        bgColor: Colors.blueAccent.withOpacity(0.1),
      ),
      _FilterItem(
        label: 'Toko Baru',
        count: _stores.where((s) => s.isNew).length,
        color: Colors.orange,
        bgColor: Colors.orange.withOpacity(0.1),
      ),
      _FilterItem(
        label: 'Toko Lama',
        count: _stores.where((s) => !s.isNew).length,
        color: Colors.green,
        bgColor: Colors.green.withOpacity(0.1),
      ),
    ];

    return Container(
      color: Colors.white,
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 14),
      child: Row(
        children: List.generate(filters.length, (index) {
          final f = filters[index];
          final isSelected = _selectedFilter == index;
          return Expanded(
            child: GestureDetector(
              onTap: () => setState(() => _selectedFilter = index),
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                margin:
                    EdgeInsets.only(right: index < filters.length - 1 ? 8 : 0),
                padding: const EdgeInsets.symmetric(vertical: 10),
                decoration: BoxDecoration(
                  color: isSelected ? f.color : Colors.grey.shade100,
                  borderRadius: BorderRadius.circular(14),
                  boxShadow: isSelected
                      ? [
                          BoxShadow(
                              color: f.color.withOpacity(0.3),
                              blurRadius: 8,
                              offset: const Offset(0, 4))
                        ]
                      : [],
                ),
                child: Column(
                  children: [
                    Text(
                      '${f.count}',
                      style: GoogleFonts.outfit(
                        fontSize: 18,
                        fontWeight: FontWeight.w900,
                        color: isSelected ? Colors.white : f.color,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      f.label,
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        fontWeight: FontWeight.w700,
                        color: isSelected
                            ? Colors.white.withOpacity(0.9)
                            : Colors.blueGrey,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
        }),
      ),
    );
  }

  Widget _buildStoreCard(StoreModel store) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 20,
              offset: const Offset(0, 8))
        ],
      ),
      child: ListTile(
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
        leading: Container(
          width: 56,
          height: 56,
          decoration: BoxDecoration(
            color: store.isNew
                ? Colors.orange.withOpacity(0.1)
                : Colors.blueAccent.withOpacity(0.1),
            shape: BoxShape.circle,
          ),
          child: Icon(Icons.storefront_rounded,
              color: store.isNew ? Colors.orange : Colors.blueAccent, size: 28),
        ),
        title: Text(store.name,
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w800,
                fontSize: 16,
                color: const Color(0xFF1E293B))),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            Text(store.ownerName,
                style:
                    GoogleFonts.outfit(fontSize: 13, color: Colors.blueGrey)),
            const SizedBox(height: 2),
            Text(store.address,
                style: GoogleFonts.outfit(
                    fontSize: 12, color: Colors.blueGrey.withOpacity(0.8)),
                maxLines: 1,
                overflow: TextOverflow.ellipsis),
            Wrap(
              spacing: 6,
              runSpacing: 4,
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 3),
                  decoration: BoxDecoration(
                    color: store.isNew
                        ? Colors.orange.withOpacity(0.1)
                        : Colors.green.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Text(
                    store.isNew ? 'TOKO BARU' : 'TOKO LAMA',
                    style: GoogleFonts.outfit(
                        fontSize: 10,
                        fontWeight: FontWeight.w900,
                        color: store.isNew ? Colors.orange : Colors.green),
                  ),
                ),
                if (store.isOffline || store.syncStatus == 'pending')
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: Colors.amber.shade100,
                      borderRadius: BorderRadius.circular(20),
                      border: Border.all(color: Colors.amber.shade800, width: 0.8),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.cloud_off_rounded, size: 11, color: Colors.amber.shade900),
                        const SizedBox(width: 3),
                        Text(
                          'OFFLINE',
                          style: GoogleFonts.outfit(
                              fontSize: 9,
                              fontWeight: FontWeight.w900,
                              color: Colors.amber.shade900),
                        ),
                      ],
                    ),
                  ),
                if (store.assignedEmployeeName != null && store.assignedEmployeeName!.isNotEmpty)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 3),
                    decoration: BoxDecoration(
                      color: Colors.blue.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: Text(
                      'PIC: ${store.assignedEmployeeName}',
                      style: GoogleFonts.outfit(
                          fontSize: 10,
                          fontWeight: FontWeight.w700,
                          color: Colors.blue.shade800),
                    ),
                  ),
              ],
            ),
          ],
        ),
        trailing: Container(
          width: 44,
          height: 44,
          decoration: BoxDecoration(
              color: Colors.blueAccent,
              borderRadius: BorderRadius.circular(14)),
          child: const Icon(Icons.arrow_forward_ios_rounded,
              color: Colors.white, size: 16),
        ),
        onTap: () async {
          if (widget.isQuickOrder) {
            Navigator.push(
                context,
                MaterialPageRoute(
                    builder: (_) => ReceiptCameraPage(
                          store: store,
                          selfiePath: "",
                          jobPositionTitle: _jobPositionTitle,
                        )));
          } else {
            final apiService = di.sl<SalesApiService>();
            final journey = await apiService.getTodayJourney();
            final todayStr = DateFormat('yyyy-MM-dd').format(DateTime.now());
            final journeyDateStr = (journey?['date'] ?? '').toString();
            final isToday = journeyDateStr.isEmpty || journeyDateStr.startsWith(todayStr);

            if (journey == null || !isToday || (journey['status'] != 'started' && journey['status'] != 'ended' && journey['status'] != 'STARTED' && journey['status'] != 'ENDED')) {
              if (!mounted) return;
              await showDialog<void>(
                context: context,
                builder: (dialogContext) => AlertDialog(
                  icon: const Icon(Icons.speed_rounded, color: Colors.orange, size: 36),
                  title: const Text('Catat KM terlebih dahulu'),
                  content: const Text(
                    'Sebelum melakukan kunjungan, Anda harus mengisi catat KM Berangkat pada menu Mulai Bekerja.',
                  ),
                  actions: [
                    FilledButton(
                      onPressed: () => Navigator.pop(dialogContext),
                      child: const Text('Mengerti'),
                    ),
                  ],
                ),
              );
              return;
            }
            if (!mounted) return;
            Navigator.push(
                context,
                MaterialPageRoute(
                    builder: (_) => VisitCheckinPage(store: store)));
          }
        },
      ),
    );
  }

  Widget _buildEmptyStateFiltered() {
    final isFiltered = _selectedFilter != 0;
    final filterLabels = ['', 'Toko Baru', 'Toko Lama'];
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            isFiltered
                ? Icons.filter_list_off_rounded
                : Icons.storefront_outlined,
            size: 80,
            color: Colors.grey.shade300,
          ),
          const SizedBox(height: 24),
          Text(
            isFiltered
                ? 'Tidak Ada ${filterLabels[_selectedFilter]}'
                : 'Belum Ada Toko',
            style: GoogleFonts.outfit(
                fontSize: 20,
                fontWeight: FontWeight.w800,
                color: const Color(0xFF334155)),
          ),
          const SizedBox(height: 8),
          Text(
            isFiltered
                ? 'Tidak ada data untuk filter "${filterLabels[_selectedFilter]}"'
                : 'Tekan "Tambah Toko Baru" untuk\nmenambahkan toko kunjungan Anda.',
            textAlign: TextAlign.center,
            style: GoogleFonts.outfit(fontSize: 14, color: Colors.blueGrey),
          ),
          if (isFiltered) ...[
            const SizedBox(height: 16),
            TextButton.icon(
              onPressed: () => setState(() => _selectedFilter = 0),
              icon: const Icon(Icons.clear_rounded, size: 16),
              label: Text('Tampilkan Semua',
                  style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
            ),
          ],
        ],
      ),
    );
  }
}

class _FilterItem {
  final String label;
  final int count;
  final Color color;
  final Color bgColor;

  const _FilterItem({
    required this.label,
    required this.count,
    required this.color,
    required this.bgColor,
  });
}
