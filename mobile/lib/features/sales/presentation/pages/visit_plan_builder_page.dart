import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/features/sales/data/models/visit_plan_model.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/features/sales/presentation/widgets/projection_bottom_sheet.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';

// ─── Enum untuk pilihan filter ─────────────────────────────────────────────
enum VisitFilter {
  semua,
  belumDikunjungi,
  belumTransaksi,
  transaksiLama,
  transaksiAktif
}

extension VisitFilterLabel on VisitFilter {
  String get label {
    switch (this) {
      case VisitFilter.semua:
        return 'Semua';
      case VisitFilter.belumDikunjungi:
        return 'Belum Dikunjungi';
      case VisitFilter.belumTransaksi:
        return 'Belum Transaksi';
      case VisitFilter.transaksiLama:
        return 'Tdk Transaksi ≥30 hari';
      case VisitFilter.transaksiAktif:
        return 'Aktif (< 30 hari)';
    }
  }

  IconData get icon {
    switch (this) {
      case VisitFilter.semua:
        return Icons.storefront_rounded;
      case VisitFilter.belumDikunjungi:
        return Icons.fiber_new_rounded;
      case VisitFilter.belumTransaksi:
        return Icons.warning_amber_rounded;
      case VisitFilter.transaksiLama:
        return Icons.trending_down_rounded;
      case VisitFilter.transaksiAktif:
        return Icons.check_circle_rounded;
    }
  }

  Color get color {
    switch (this) {
      case VisitFilter.semua:
        return const Color(0xFF64748B);
      case VisitFilter.belumDikunjungi:
        return const Color(0xFFEF4444);
      case VisitFilter.belumTransaksi:
        return const Color(0xFFF59E0B);
      case VisitFilter.transaksiLama:
        return const Color(0xFFF97316);
      case VisitFilter.transaksiAktif:
        return const Color(0xFF10B981);
    }
  }
}

class VisitPlanBuilderPage extends StatefulWidget {
  const VisitPlanBuilderPage({super.key});

  @override
  State<VisitPlanBuilderPage> createState() => _VisitPlanBuilderPageState();
}

class _VisitPlanBuilderPageState extends State<VisitPlanBuilderPage> {
  final _salesApi = SalesApiService(apiClient: di.sl<ApiClient>());

  // ─── State ─────────────────────────────────────────────────────────────
  bool _isLoading = true;
  bool _isManagement = false;
  bool _isSupervisor = false;
  String? _branchId;
  List<Map<String, dynamic>> _salesmenList = [];
  String? _selectedSalesmanId;
  String? _selectedSalesmanName;

  List<VisitPlanModel> _stores = []; // master list (sorted, tidak berubah)
  List<VisitPlanModel> _filteredStores = []; // hasil filter + search
  final Set<String> _selectedStoreIds = {};
  final Map<String, List<Map<String, dynamic>>> _storeProjections = {};
  DateTime _planDate = DateTime.now();

  // ─── Search & Filter ───────────────────────────────────────────────────
  final TextEditingController _searchController = TextEditingController();
  String _searchQuery = '';
  VisitFilter _activeFilter = VisitFilter.semua;
  bool _showFilterBar = false;

  @override
  void initState() {
    super.initState();
    _checkUserRoleAndFetchData();
    _searchController.addListener(_onSearchChanged);
  }

  Future<void> _checkUserRoleAndFetchData() async {
    final prefs = await SharedPreferences.getInstance();
    final cached = prefs.getString('cached_employee_profile');
    if (cached != null) {
      try {
        final decoded = jsonDecode(cached);
        final jobTitle = (decoded['job_position']?['title'] ?? decoded['job_title'] ?? '').toString().toLowerCase();
        final role = (decoded['role'] ?? '').toString().toLowerCase();
        final branchVal = decoded['branch_id'] ??
            decoded['BranchID'] ??
            decoded['branch']?['id'] ??
            decoded['Branch']?['ID'] ??
            decoded['employee']?['branch_id'];
        if (branchVal != null) {
          _branchId = branchVal.toString();
        }

        final isTopMgmt = role == 'direktur' || role == 'director' || role == 'owner' || 
                          role == 'superadmin' || role == 'manager' ||
                          jobTitle.contains('direktur') || jobTitle.contains('director') || 
                          jobTitle.contains('manager') || jobTitle.contains('owner');

        final isSpv = role == 'supervisor' || role == 'spv' ||
                      role == 'supervisor_sales' || role == 'supervisor_sales_mt' ||
                      jobTitle.contains('spv') || jobTitle.contains('supervisor');

        _isSupervisor = isSpv && !isTopMgmt;
        _isManagement = isTopMgmt || isSpv;
      } catch (_) {}
    }

    if (_isManagement) {
      final salesmen = await _salesApi.getSalesmenDirectory(
        branchId: _isSupervisor ? _branchId : null,
      );
      if (mounted) {
        setState(() {
          _salesmenList = salesmen;
          if (_isSupervisor) {
            _selectedSalesmanId = null;
            _selectedSalesmanName = null;
          } else {
            if (salesmen.isNotEmpty && _selectedSalesmanId == null) {
              _selectedSalesmanId = salesmen.first['id']?.toString();
              _selectedSalesmanName = salesmen.first['name']?.toString();
            }
          }
        });
      }
    }

    _fetchBuilderData();
  }

  @override
  void dispose() {
    _searchController.removeListener(_onSearchChanged);
    _searchController.dispose();
    super.dispose();
  }

  void _onSearchChanged() {
    setState(() {
      _searchQuery = _searchController.text.trim().toLowerCase();
      _applyFilterAndSearch();
    });
  }

  // ─── Data Fetching ──────────────────────────────────────────────────────
  Future<void> _fetchExistingPlansForDate() async {
    try {
      final dateStr = DateFormat('yyyy-MM-dd').format(_planDate);
      final plans = await _salesApi.getVisitPlans(date: dateStr);
      final Set<String> plannedStoreIds = {};
      for (var p in plans) {
        if (p['store_id'] != null) {
          plannedStoreIds.add(p['store_id'].toString());
        }
      }
      
      final prefs = await SharedPreferences.getInstance();
      final key = 'deleted_stores_$dateStr';
      final deletedList = prefs.getStringList(key) ?? [];
      final deletedSet = deletedList.toSet();
      
      plannedStoreIds.removeWhere((id) => deletedSet.contains(id));

      setState(() {
        _selectedStoreIds.clear();
        _selectedStoreIds.addAll(plannedStoreIds);
      });
    } catch (e) {
      debugPrint('Error fetching existing plans: $e');
    }
  }

  Future<void> _fetchBuilderData() async {
    final prefs = await SharedPreferences.getInstance();
    final cacheKey = 'cached_visit_plan_builder_${_isManagement ? (_selectedSalesmanId ?? _branchId ?? "all") : "me"}';

    List<VisitPlanModel> cachedStores = [];
    // Step 1: Instant render from cache
    final cachedStr = prefs.getString(cacheKey);
    if (cachedStr != null && _stores.isEmpty) {
      try {
        final List<dynamic> decoded = jsonDecode(cachedStr);
        cachedStores = decoded.map((e) => VisitPlanModel.fromJson(e)).toList();
        if (cachedStores.isNotEmpty && mounted) {
          setState(() {
            _stores = cachedStores;
            _applyFilterAndSearch();
            _isLoading = false; // Kill spinner immediately
          });
        }
      } catch (_) {}
    } else if (_stores.isEmpty) {
      setState(() => _isLoading = true);
    }

    try {
      final data = await _salesApi.getVisitPlanBuilder(
        salesmanId: _isManagement ? _selectedSalesmanId : null,
        branchId: _isSupervisor ? _branchId : null,
      );
      final fetchedStores = data.map((e) => VisitPlanModel.fromJson(e)).toList();

      // Merge cachedStores (which preserve lastVisitDate/lastTransactionDate) with fetchedStores (which include newly created offline stores)
      final Map<String, VisitPlanModel> mergedMap = {};
      for (final cs in cachedStores) {
        if (cs.storeId.isNotEmpty) {
          mergedMap[cs.storeId] = cs;
        }
      }
      for (final fs in fetchedStores) {
        if (fs.storeId.isNotEmpty) {
          final existing = mergedMap[fs.storeId];
          if (existing != null && fs.lastVisitDate == null && existing.lastVisitDate != null) {
            mergedMap[fs.storeId] = existing;
          } else {
            mergedMap[fs.storeId] = fs;
          }
        }
      }
      final stores = mergedMap.values.toList();

      // Sorting prioritas kunjungan
      stores.sort((a, b) {
        if (a.isPriority && !b.isPriority) return -1;
        if (!a.isPriority && b.isPriority) return 1;
        final aNeverVisited = a.lastVisitDate == null;
        final bNeverVisited = b.lastVisitDate == null;
        if (aNeverVisited && !bNeverVisited) return -1;
        if (!aNeverVisited && bNeverVisited) return 1;
        final aTrxDate = a.lastTransactionDate;
        final bTrxDate = b.lastTransactionDate;
        if (aTrxDate == null && bTrxDate == null) return 0;
        if (aTrxDate == null) return -1;
        if (bTrxDate == null) return 1;
        return aTrxDate.compareTo(bTrxDate);
      });

      await _fetchExistingPlansForDate();

      if (mounted) {
        setState(() {
          _stores = stores;
          _applyFilterAndSearch();
          _isLoading = false;
        });
      }

      if (data.isNotEmpty) {
        // Persist to local cache for instant future opens
        prefs.setString(cacheKey, jsonEncode(data)).ignore();
      }
    } catch (e) {
      debugPrint('Error fetching builder data: $e');
      if (mounted && _stores.isEmpty) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal memuat data toko: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  // ─── Filter + Search Logic ──────────────────────────────────────────────
  void _applyFilterAndSearch() {
    List<VisitPlanModel> result = List.from(_stores);

    // 1. Terapkan filter kategori
    switch (_activeFilter) {
      case VisitFilter.belumDikunjungi:
        result = result.where((s) => s.lastVisitDate == null).toList();
        break;
      case VisitFilter.belumTransaksi:
        result = result.where((s) => s.lastTransactionDate == null).toList();
        break;
      case VisitFilter.transaksiLama:
        result = result.where((s) {
          if (s.lastTransactionDate == null) return false;
          final days = DateTime.now().difference(s.lastTransactionDate!).inDays;
          return days >= 30;
        }).toList();
        break;
      case VisitFilter.transaksiAktif:
        result = result.where((s) {
          if (s.lastTransactionDate == null) return false;
          final days = DateTime.now().difference(s.lastTransactionDate!).inDays;
          return days < 30;
        }).toList();
        break;
      case VisitFilter.semua:
        break;
    }

    // 2. Terapkan pencarian teks (nama toko atau alamat)
    if (_searchQuery.isNotEmpty) {
      result = result.where((s) {
        return s.storeName.toLowerCase().contains(_searchQuery) ||
            s.address.toLowerCase().contains(_searchQuery);
      }).toList();
    }

    _filteredStores = result;
  }

  void _setFilter(VisitFilter filter) {
    setState(() {
      _activeFilter = filter;
      _applyFilterAndSearch();
    });
  }

  void _resetAll() {
    setState(() {
      _searchController.clear();
      _searchQuery = '';
      _activeFilter = VisitFilter.semua;
      _applyFilterAndSearch();
    });
  }

  // ─── Selection ──────────────────────────────────────────────────────────
  void _toggleSelection(String id) {
    setState(() {
      if (_selectedStoreIds.contains(id)) {
        _selectedStoreIds.remove(id);
        _storeProjections.remove(id);
      } else {
        _selectedStoreIds.add(id);
      }
    });
  }

  void _selectAllVisible() {
    setState(() {
      for (final s in _filteredStores) {
        _selectedStoreIds.add(s.storeId);
      }
    });
  }

  void _deselectAllVisible() {
    setState(() {
      for (final s in _filteredStores) {
        _selectedStoreIds.remove(s.storeId);
        _storeProjections.remove(s.storeId);
      }
    });
  }

  bool get _allVisibleSelected =>
      _filteredStores.isNotEmpty &&
      _filteredStores.every((s) => _selectedStoreIds.contains(s.storeId));

  Future<void> _selectDate() async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: _planDate,
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 90)),
    );
    if (picked != null && picked != _planDate) {
      setState(() {
        _planDate = picked;
        _isLoading = true;
      });
      await _fetchExistingPlansForDate();
      setState(() {
        _isLoading = false;
      });
    }
  }

  // ─── Save ───────────────────────────────────────────────────────────────
  Future<void> _savePlan() async {
    if (_selectedStoreIds.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Pilih minimal 1 toko untuk disimpan')),
      );
      return;
    }
    setState(() => _isLoading = true);
    try {
      List<Map<String, dynamic>> storesPayload = [];
      for (final id in _selectedStoreIds) {
        storesPayload.add({
          'store_id': id,
          if (_storeProjections.containsKey(id)) 'projections': _storeProjections[id],
        });
      }

      await _salesApi.saveManualVisitPlan(
          stores: storesPayload, planDate: _planDate);
      try {
        di.sl<SyncBloc>().add(const SyncDataRequested(force: true));
      } catch (_) {}
          
      final prefs = await SharedPreferences.getInstance();
      final dateStr = DateFormat('yyyy-MM-dd').format(_planDate);
      final key = 'deleted_stores_$dateStr';
      final deletedList = prefs.getStringList(key) ?? [];
      final deletedSet = deletedList.toSet();
      deletedSet.removeWhere((id) => _selectedStoreIds.contains(id));
      await prefs.setStringList(key, deletedSet.toList());
      
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Rencana kunjungan berhasil disimpan!'),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
              content: Text('Gagal menyimpan rencana: $e'),
              backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  // ─── Build ──────────────────────────────────────────────────────────────
  @override
  Widget build(BuildContext context) {
    final hasActiveFilter =
        _activeFilter != VisitFilter.semua || _searchQuery.isNotEmpty;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        toolbarHeight: 76,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded,
              color: Color(0xFF334155)),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
                _isManagement && _selectedSalesmanName != null
                    ? 'Toko Sales: $_selectedSalesmanName'
                    : (_isSupervisor ? 'Rencana Kunjungan Cabang' : 'Pilih Toko Kunjungan'),
                style: GoogleFonts.outfit(
                    fontSize: 17,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B))),
            const SizedBox(height: 6),
            InkWell(
              onTap: _selectDate,
              borderRadius: BorderRadius.circular(20),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                decoration: BoxDecoration(
                  color: const Color(0xFF10B981).withOpacity(0.12),
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: const Color(0xFF10B981).withOpacity(0.4)),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.calendar_month_rounded, size: 14, color: Color(0xFF10B981)),
                    const SizedBox(width: 6),
                    Text('Untuk Tanggal: ',
                        style: GoogleFonts.outfit(
                            fontSize: 12,
                            color: const Color(0xFF334155),
                            fontWeight: FontWeight.w600)),
                    Text('${_planDate.day}/${_planDate.month}/${_planDate.year}',
                        style: GoogleFonts.outfit(
                            fontSize: 13,
                            color: const Color(0xFF10B981),
                            fontWeight: FontWeight.w900)),
                    const SizedBox(width: 4),
                    const Icon(Icons.keyboard_arrow_down_rounded, size: 18, color: Color(0xFF10B981)),
                  ],
                ),
              ),
            ),
          ],
        ),
        actions: [
          // Tombol filter
          Stack(
            alignment: Alignment.topRight,
            children: [
              IconButton(
                onPressed: () =>
                    setState(() => _showFilterBar = !_showFilterBar),
                icon: Icon(
                  _showFilterBar
                      ? Icons.filter_list_off_rounded
                      : Icons.filter_list_rounded,
                  color: hasActiveFilter
                      ? const Color(0xFF10B981)
                      : const Color(0xFF334155),
                ),
                tooltip: 'Filter',
              ),
              if (hasActiveFilter)
                Positioned(
                  top: 8,
                  right: 8,
                  child: Container(
                    width: 8,
                    height: 8,
                    decoration: const BoxDecoration(
                      color: Color(0xFF10B981),
                      shape: BoxShape.circle,
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(width: 4),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : Column(
              children: [
                // ─── Salesman Selector for Management ──────────────────
                _buildSalesmanSelector(),

                // ─── Search Bar ───────────────────────────────────────
                Container(
                  color: Colors.white,
                  padding: const EdgeInsets.fromLTRB(16, 0, 16, 12),
                  child: TextField(
                    controller: _searchController,
                    style: GoogleFonts.outfit(
                        fontSize: 14, color: const Color(0xFF1E293B)),
                    decoration: InputDecoration(
                      hintText: 'Cari nama toko atau alamat...',
                      hintStyle: GoogleFonts.outfit(
                          fontSize: 13, color: Colors.blueGrey.shade300),
                      prefixIcon: const Icon(Icons.search_rounded,
                          color: Color(0xFF94A3B8), size: 20),
                      suffixIcon: _searchQuery.isNotEmpty
                          ? IconButton(
                              icon: const Icon(Icons.close_rounded,
                                  color: Color(0xFF94A3B8), size: 18),
                              onPressed: () {
                                _searchController.clear();
                              },
                            )
                          : null,
                      filled: true,
                      fillColor: const Color(0xFFF1F5F9),
                      contentPadding: const EdgeInsets.symmetric(
                          vertical: 0, horizontal: 16),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: BorderSide.none,
                      ),
                    ),
                  ),
                ),

                // ─── Filter Chips (collapsible) ───────────────────────
                AnimatedSize(
                  duration: const Duration(milliseconds: 220),
                  curve: Curves.easeInOut,
                  child: _showFilterBar
                      ? Container(
                          color: Colors.white,
                          padding: const EdgeInsets.fromLTRB(16, 0, 16, 14),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Divider(
                                  height: 1, color: Color(0xFFF1F5F9)),
                              const SizedBox(height: 10),
                              Text('Filter Toko',
                                  style: GoogleFonts.outfit(
                                      fontSize: 11,
                                      fontWeight: FontWeight.w700,
                                      color: Colors.blueGrey)),
                              const SizedBox(height: 8),
                              Wrap(
                                spacing: 8,
                                runSpacing: 6,
                                children: VisitFilter.values.map((f) {
                                  final isActive = _activeFilter == f;
                                  return GestureDetector(
                                    onTap: () => _setFilter(f),
                                    child: AnimatedContainer(
                                      duration:
                                          const Duration(milliseconds: 160),
                                      padding: const EdgeInsets.symmetric(
                                          horizontal: 12, vertical: 6),
                                      decoration: BoxDecoration(
                                        color: isActive
                                            ? f.color
                                            : f.color.withOpacity(0.08),
                                        borderRadius: BorderRadius.circular(20),
                                        border: Border.all(
                                          color: isActive
                                              ? f.color
                                              : f.color.withOpacity(0.3),
                                        ),
                                      ),
                                      child: Row(
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          Icon(f.icon,
                                              size: 12,
                                              color: isActive
                                                  ? Colors.white
                                                  : f.color),
                                          const SizedBox(width: 5),
                                          Text(
                                            f.label,
                                            style: GoogleFonts.outfit(
                                              fontSize: 11,
                                              fontWeight: FontWeight.w700,
                                              color: isActive
                                                  ? Colors.white
                                                  : f.color,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                  );
                                }).toList(),
                              ),
                            ],
                          ),
                        )
                      : const SizedBox.shrink(),
                ),

                // ─── Results Summary Bar ──────────────────────────────
                Container(
                  color: const Color(0xFFF8FAFC),
                  padding:
                      const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
                  child: Row(
                    children: [
                      Text(
                        '${_filteredStores.length} dari ${_stores.length} toko',
                        style: GoogleFonts.outfit(
                            fontSize: 12,
                            fontWeight: FontWeight.w700,
                            color: Colors.blueGrey),
                      ),
                      if (hasActiveFilter) ...[
                        const SizedBox(width: 6),
                        GestureDetector(
                          onTap: _resetAll,
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 8, vertical: 2),
                            decoration: BoxDecoration(
                              color: const Color(0xFFEF4444).withOpacity(0.1),
                              borderRadius: BorderRadius.circular(10),
                            ),
                            child: Text('Reset',
                                style: GoogleFonts.outfit(
                                    fontSize: 10,
                                    fontWeight: FontWeight.w700,
                                    color: const Color(0xFFEF4444))),
                          ),
                        ),
                      ],
                      const Spacer(),
                      // Pilih semua / batalkan yang tampil
                      GestureDetector(
                        onTap: _allVisibleSelected
                            ? _deselectAllVisible
                            : _selectAllVisible,
                        child: Text(
                          _allVisibleSelected
                              ? 'Batalkan Semua'
                              : 'Pilih Semua (${_filteredStores.length})',
                          style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                              color: const Color(0xFF10B981)),
                        ),
                      ),
                    ],
                  ),
                ),

                // ─── List ─────────────────────────────────────────────
                Expanded(
                  child: _filteredStores.isEmpty
                      ? _buildEmptyState(hasActiveFilter)
                      : ListView.builder(
                          padding: const EdgeInsets.fromLTRB(20, 4, 20, 100),
                          itemCount: _filteredStores.length,
                          itemBuilder: (ctx, i) {
                            final store = _filteredStores[i];
                            // Section header hanya tampil jika tidak ada filter aktif
                            final showHeaders =
                                _activeFilter == VisitFilter.semua &&
                                    _searchQuery.isEmpty;
                            final isFirstVisited = showHeaders &&
                                i > 0 &&
                                _filteredStores[i - 1].lastVisitDate == null &&
                                store.lastVisitDate != null;
                            final isFirstNeverVisited = showHeaders &&
                                i == 0 &&
                                store.lastVisitDate == null;

                            // Nomor urut mengacu pada posisi di master list
                            final masterIndex = _stores.indexOf(store);

                            return Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                if (isFirstNeverVisited)
                                  _buildSectionHeader(
                                    Icons.fiber_new_rounded,
                                    'Belum Pernah Dikunjungi',
                                    const Color(0xFFEF4444),
                                    'Prioritas tertinggi — toko ini belum pernah disambangi',
                                  ),
                                if (isFirstVisited)
                                  _buildSectionHeader(
                                    Icons.history_rounded,
                                    'Urutkan Berdasarkan Transaksi Terlama',
                                    const Color(0xFF2563EB),
                                    'Toko dengan transaksi paling lama muncul paling atas',
                                  ),
                                _buildStoreCard(store, masterIndex),
                              ],
                            );
                          },
                        ),
                ),
              ],
            ),
      floatingActionButton: _selectedStoreIds.isNotEmpty
          ? FloatingActionButton.extended(
              onPressed: _savePlan,
              backgroundColor: const Color(0xFF10B981),
              icon: const Icon(Icons.save_rounded, color: Colors.white),
              label: Text(
                'Simpan Rencana (${_selectedStoreIds.length})',
                style: GoogleFonts.outfit(
                    color: Colors.white, fontWeight: FontWeight.w900),
              ),
            )
          : null,
    );
  }

  // ─── Widgets ────────────────────────────────────────────────────────────

  Widget _buildSectionHeader(
    IconData icon,
    String title,
    Color color,
    String subtitle,
  ) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12, top: 4),
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: color.withOpacity(0.08),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: color.withOpacity(0.25)),
      ),
      child: Row(
        children: [
          Icon(icon, color: color, size: 18),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title,
                    style: GoogleFonts.outfit(
                        fontSize: 12,
                        fontWeight: FontWeight.w800,
                        color: color)),
                Text(subtitle,
                    style: GoogleFonts.outfit(
                        fontSize: 10,
                        color: color.withOpacity(0.75),
                        fontWeight: FontWeight.w500)),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStoreCard(VisitPlanModel store, int masterIndex) {
    final isSelected = _selectedStoreIds.contains(store.storeId);
    final neverVisited = store.lastVisitDate == null;
    final neverTransacted = store.lastTransactionDate == null;

    String lastVisitStr = neverVisited
        ? 'Belum Pernah'
        : () {
            final days = DateTime.now().difference(store.lastVisitDate!).inDays;
            return days == 0 ? 'Hari Ini' : '$days hari lalu';
          }();

    String lastTrxStr = neverTransacted
        ? 'Belum Ada'
        : () {
            final days =
                DateTime.now().difference(store.lastTransactionDate!).inDays;
            if (days == 0) return 'Hari Ini';
            if (days == 1) return 'Kemarin';
            return '$days hari lalu';
          }();

    Color priorityColor;
    String priorityLabel;
    IconData priorityIcon;
    if (neverVisited) {
      priorityColor = const Color(0xFFEF4444);
      priorityLabel = 'Belum Dikunjungi';
      priorityIcon = Icons.fiber_new_rounded;
    } else if (neverTransacted) {
      priorityColor = const Color(0xFFF59E0B);
      priorityLabel = 'Belum Ada Transaksi';
      priorityIcon = Icons.warning_amber_rounded;
    } else {
      final daysSinceTrx =
          DateTime.now().difference(store.lastTransactionDate!).inDays;
      if (daysSinceTrx >= 30) {
        priorityColor = const Color(0xFFF97316);
        priorityLabel = '$daysSinceTrx hari tdk ada transaksi';
        priorityIcon = Icons.trending_down_rounded;
      } else {
        priorityColor = const Color(0xFF10B981);
        priorityLabel = 'Transaksi $daysSinceTrx hari lalu';
        priorityIcon = Icons.check_circle_rounded;
      }
    }

    final itemsSold = store.lastTransactionItems ?? '';

    return Container(
      margin: const EdgeInsets.only(bottom: 14),
      decoration: BoxDecoration(
        color: isSelected
            ? const Color(0xFF10B981).withOpacity(0.05)
            : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: isSelected ? const Color(0xFF10B981) : Colors.transparent,
          width: 2,
        ),
        boxShadow: [
          if (!isSelected)
            BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 12,
              offset: const Offset(0, 4),
            )
        ],
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(20),
        onTap: () => _toggleSelection(store.storeId),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Nomor urut prioritas dari master list
                  Container(
                    width: 28,
                    height: 28,
                    margin: const EdgeInsets.only(right: 10, top: 8),
                    decoration: BoxDecoration(
                      color: priorityColor.withOpacity(0.12),
                      shape: BoxShape.circle,
                    ),
                    alignment: Alignment.center,
                    child: Text(
                      '${masterIndex + 1}',
                      style: GoogleFonts.outfit(
                          fontSize: 12,
                          fontWeight: FontWeight.w900,
                          color: priorityColor),
                    ),
                  ),
                  Container(
                    width: 44,
                    height: 44,
                    decoration: BoxDecoration(
                      color: const Color(0xFF10B981).withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.storefront_rounded,
                        color: Color(0xFF10B981)),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Highlight teks pencarian pada nama toko
                        _highlightText(store.storeName, _searchQuery),
                        Text(store.address,
                            style: GoogleFonts.outfit(
                                fontSize: 11, color: Colors.blueGrey),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis),
                        const SizedBox(height: 6),
                        Wrap(
                          spacing: 6,
                          runSpacing: 4,
                          children: [
                            Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 8, vertical: 3),
                              decoration: BoxDecoration(
                                color: priorityColor.withOpacity(0.1),
                                borderRadius: BorderRadius.circular(20),
                              ),
                              child: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  Icon(priorityIcon,
                                      size: 11, color: priorityColor),
                                  const SizedBox(width: 4),
                                  Text(priorityLabel,
                                      style: GoogleFonts.outfit(
                                          fontSize: 10,
                                          fontWeight: FontWeight.w700,
                                          color: priorityColor)),
                                ],
                              ),
                            ),
                            if (store.assignedEmployeeName != null &&
                                store.assignedEmployeeName!.isNotEmpty &&
                                store.assignedEmployeeName != 'Belum Ditugaskan')
                              Container(
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 8, vertical: 3),
                                decoration: BoxDecoration(
                                  color: const Color(0xFF6366F1).withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(20),
                                ),
                                child: Row(
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    const Icon(Icons.person_pin_rounded,
                                        size: 11, color: Color(0xFF6366F1)),
                                    const SizedBox(width: 4),
                                    Text('PIC: ${store.assignedEmployeeName}',
                                        style: GoogleFonts.outfit(
                                            fontSize: 10,
                                            fontWeight: FontWeight.w700,
                                            color: const Color(0xFF6366F1))),
                                  ],
                                ),
                              ),
                          ],
                        ),
                      ],
                    ),
                  ),
                  Checkbox(
                    value: isSelected,
                    onChanged: (val) => _toggleSelection(store.storeId),
                    activeColor: const Color(0xFF10B981),
                    shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(4)),
                  ),
                ],
              ),
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 12),
                child: Divider(height: 1, color: Color(0xFFF1F5F9)),
              ),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  _buildInfoBadge(
                    Icons.calendar_month_rounded,
                    'Kunjungan Terakhir',
                    lastVisitStr,
                    neverVisited ? const Color(0xFFEF4444) : Colors.blueGrey,
                  ),
                  _buildInfoBadge(
                    Icons.receipt_long_rounded,
                    'Transaksi Terakhir',
                    lastTrxStr,
                    neverTransacted
                        ? const Color(0xFFF59E0B)
                        : const Color(0xFF10B981),
                  ),
                ],
              ),
              if (itemsSold.isNotEmpty &&
                  itemsSold != 'Tidak ada data barang') ...[
                const SizedBox(height: 12),
                Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF8FAFC),
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(color: const Color(0xFFE2E8F0)),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text('Barang Terakhir Dijual:',
                          style: GoogleFonts.outfit(
                              fontSize: 10,
                              fontWeight: FontWeight.w700,
                              color: Colors.blueGrey)),
                      const SizedBox(height: 4),
                      Text(itemsSold,
                          style: GoogleFonts.outfit(
                              fontSize: 12,
                              fontWeight: FontWeight.w600,
                              color: const Color(0xFF334155)),
                          maxLines: 2),
                    ],
                  ),
                ),
              ],
              if (isSelected) ...[
                const SizedBox(height: 12),
                SizedBox(
                  width: double.infinity,
                  child: OutlinedButton.icon(
                    onPressed: () => _showProjectionDialog(store.storeId, store.storeName),
                    icon: const Icon(Icons.add_shopping_cart_rounded, size: 16),
                    label: Text(
                      _storeProjections.containsKey(store.storeId) 
                        ? 'Proyeksi Diatur (${_storeProjections[store.storeId]!.length} Produk)'
                        : 'Atur Proyeksi Omzet (Opsional)',
                      style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
                    ),
                    style: OutlinedButton.styleFrom(
                      foregroundColor: _storeProjections.containsKey(store.storeId) 
                        ? const Color(0xFF10B981) 
                        : const Color(0xFF64748B),
                      side: BorderSide(
                        color: _storeProjections.containsKey(store.storeId) 
                          ? const Color(0xFF10B981) 
                          : const Color(0xFFCBD5E1),
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                      ),
                    ),
                  ),
                ),
              ]
            ],
          ),
        ),
      ),
    );
  }

  /// Highlight teks yang cocok dengan query pencarian
  Widget _highlightText(String text, String query) {
    if (query.isEmpty) {
      return Text(text,
          style: GoogleFonts.outfit(
              fontWeight: FontWeight.w800,
              fontSize: 15,
              color: const Color(0xFF1E293B)));
    }

    final lowerText = text.toLowerCase();
    final lowerQuery = query.toLowerCase();
    final index = lowerText.indexOf(lowerQuery);

    if (index < 0) {
      return Text(text,
          style: GoogleFonts.outfit(
              fontWeight: FontWeight.w800,
              fontSize: 15,
              color: const Color(0xFF1E293B)));
    }

    return RichText(
      text: TextSpan(
        style: GoogleFonts.outfit(
            fontWeight: FontWeight.w800,
            fontSize: 15,
            color: const Color(0xFF1E293B)),
        children: [
          TextSpan(text: text.substring(0, index)),
          TextSpan(
            text: text.substring(index, index + query.length),
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w900,
              fontSize: 15,
              color: const Color(0xFF10B981),
              backgroundColor: const Color(0xFF10B981).withOpacity(0.12),
            ),
          ),
          TextSpan(text: text.substring(index + query.length)),
        ],
      ),
    );
  }

  Widget _buildInfoBadge(
      IconData icon, String label, String value, Color color) {
    return Row(
      children: [
        Icon(icon, size: 14, color: color),
        const SizedBox(width: 4),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(label,
                style: GoogleFonts.outfit(
                    fontSize: 9,
                    color: Colors.blueGrey,
                    fontWeight: FontWeight.w600)),
            Text(value,
                style: GoogleFonts.outfit(
                    fontSize: 12, fontWeight: FontWeight.w800, color: color)),
          ],
        ),
      ],
    );
  }

  void _showProjectionDialog(String storeId, String storeName) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (ctx) {
        return ProjectionBottomSheet(
          storeId: storeId,
          storeName: storeName,
          salesApi: _salesApi,
          initialProjections: _storeProjections[storeId] ?? [],
          onSave: (projections) {
            setState(() {
              if (projections.isEmpty) {
                _storeProjections.remove(storeId);
              } else {
                _storeProjections[storeId] = projections;
              }
            });
            Navigator.pop(ctx);
          },
        );
      },
    );
  }

  Widget _buildSalesmanSelector() {
    if (!_isManagement) return const SizedBox.shrink();
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.fromLTRB(16, 4, 16, 12),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          color: const Color(0xFFF8FAFC),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: const Color(0xFF10B981).withOpacity(0.35),
            width: 1.5,
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(_isSupervisor ? Icons.tune_rounded : Icons.people_alt_rounded,
                    size: 16, color: const Color(0xFF10B981)),
                const SizedBox(width: 8),
                Text(
                  _isSupervisor
                      ? 'FILTER TOKO SALESMAN CABANG (OPSIONAL)'
                      : 'PILIH SALESMAN PIC (PENDAMPINGAN)',
                  style: GoogleFonts.outfit(
                    fontSize: 11,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF10B981),
                    letterSpacing: 0.5,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 4),
            DropdownButtonHideUnderline(
              child: DropdownButton<String?>(
                isExpanded: true,
                value: _selectedSalesmanId,
                items: [
                  DropdownMenuItem<String?>(
                    value: null,
                    child: Text(
                      _isSupervisor
                          ? '⭐ Semua Toko Cabang Saya (Rute Bebas)'
                          : '⭐ Semua Salesman (Semua Toko)',
                      style: GoogleFonts.outfit(
                        fontSize: 13,
                        fontWeight: FontWeight.w700,
                        color: const Color(0xFF10B981),
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  ..._salesmenList.map((s) {
                    final id = s['id']?.toString() ?? '';
                    final name = s['name']?.toString() ?? 'Tanpa Nama';
                    final pos = s['position']?.toString() ??
                        s['job_title']?.toString() ??
                        'Sales';
                    return DropdownMenuItem<String?>(
                      value: id,
                      child: Text(
                        '$name ($pos)',
                        style: GoogleFonts.outfit(
                          fontSize: 13,
                          fontWeight: FontWeight.w600,
                          color: const Color(0xFF1E293B),
                        ),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    );
                  }),
                ],
                onChanged: (val) {
                  if (val != _selectedSalesmanId) {
                    final matched = val == null
                        ? null
                        : _salesmenList.firstWhere(
                            (e) => e['id']?.toString() == val,
                            orElse: () => {},
                          );
                    setState(() {
                      _selectedSalesmanId = val;
                      _selectedSalesmanName =
                          matched != null ? matched['name']?.toString() : null;
                      _isLoading = true;
                    });
                    _fetchBuilderData();
                  }
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState(bool hasFilter) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              hasFilter
                  ? Icons.search_off_rounded
                  : Icons.store_mall_directory_rounded,
              size: 80,
              color: Colors.grey.shade300,
            ),
            const SizedBox(height: 24),
            Text(
              hasFilter ? 'Tidak Ada Hasil' : 'Belum Ada Toko',
              style: GoogleFonts.outfit(
                  fontSize: 20,
                  fontWeight: FontWeight.w800,
                  color: const Color(0xFF334155)),
            ),
            const SizedBox(height: 8),
            Text(
              hasFilter
                  ? 'Tidak ada toko yang sesuai\ndengan filter atau pencarian Anda.'
                  : 'Tidak ada data toko yang bisa\nditambahkan ke rencana Anda.',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(fontSize: 14, color: Colors.blueGrey),
            ),
            if (hasFilter) ...[
              const SizedBox(height: 20),
              ElevatedButton.icon(
                onPressed: _resetAll,
                icon: const Icon(Icons.refresh_rounded, size: 16),
                label: Text('Reset Filter',
                    style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF10B981),
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12)),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
