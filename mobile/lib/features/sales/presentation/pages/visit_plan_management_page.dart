import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_plan_builder_page.dart';
import 'package:shimmer/shimmer.dart';
import 'package:hris_app/features/sales/presentation/widgets/projection_bottom_sheet.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';

// ─── Constants ────────────────────────────────────────────────────────────────
const _kEmerald = Color(0xFF10B981);
const _kEmeraldDark = Color(0xFF059669);
const _kEmeraldLight = Color(0xFFF0FDF4);
const _kSlate900 = Color(0xFF0F172A);
const _kSlate700 = Color(0xFF334155);
const _kSlate500 = Color(0xFF64748B);
const _kSlate300 = Color(0xFF94A3B8);
const _kSlate100 = Color(0xFFF1F5F9);
const _kBg = Color(0xFFF8FAFC);

class VisitPlanManagementPage extends StatefulWidget {
  const VisitPlanManagementPage({super.key});

  @override
  State<VisitPlanManagementPage> createState() =>
      _VisitPlanManagementPageState();
}

class _VisitPlanManagementPageState extends State<VisitPlanManagementPage> {
  final _salesApi = SalesApiService(apiClient: di.sl<ApiClient>());
  bool _isLoading = true;
  DateTime _selectedDate = DateTime.now();
  List<dynamic> _visitPlans = [];

  // ─── Projection State ────────────────────────────────────────────────────
  double _totalProjectedOmzet = 0.0;
  double _dailyTargetOmzet = 0.0;
  double _monthlyTargetOmzet = 0.0;
  double _currentOmzet = 0.0;
  Map<String, dynamic>? _kpiData;

  final _currencyFmt = NumberFormat.currency(
    locale: 'id_ID',
    symbol: 'Rp ',
    decimalDigits: 0,
  );

  @override
  void initState() {
    super.initState();
    _fetchPlans();
  }

  // ─── Data Fetching ────────────────────────────────────────────────────────
  Future<void> _fetchPlans() async {
    if (!mounted) return;
    setState(() => _isLoading = true);
    try {
      final dateStr = DateFormat('yyyy-MM-dd').format(_selectedDate);

      // Fetch visit plans & KPI in parallel
      final results = await Future.wait([
        _salesApi.getVisitPlans(date: dateStr),
        _salesApi.getMyPerformance().catchError((_) => <String, dynamic>{}),
      ]);

      final plans = results[0] as List<dynamic>;
      final kpi = results[1] as Map<String, dynamic>;

      // Calculate total projected omzet from plans
      double totalProjected = 0.0;
      for (var p in plans) {
        totalProjected += (p['projected_omzet'] as num?)?.toDouble() ?? 0.0;
      }

      // Calculate daily target from monthly target ÷ 22 working days
      final items = kpi['items'] as List<dynamic>? ?? [];
      final monthlyTarget = items.isNotEmpty
          ? (items[0]['score'] as num?)?.toDouble() ?? 0.0
          : 0.0;
      final currentOmzet = items.length > 1
          ? (items[1]['score'] as num?)?.toDouble() ?? 0.0
          : 0.0;
      final dailyTarget = monthlyTarget > 0 ? monthlyTarget / 22 : 0.0;

      setState(() {
        _visitPlans = plans;
        _kpiData = kpi;
        _totalProjectedOmzet = totalProjected;
        _dailyTargetOmzet = dailyTarget;
        _monthlyTargetOmzet = monthlyTarget;
        _currentOmzet = currentOmzet;
      });
    } catch (e) {
      debugPrint('Error fetching visit plans: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal memuat rencana kunjungan: $e'),
            backgroundColor: Colors.red.shade700,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _selectDate() async {
    final DateTime? picked = await showDatePicker(
      context: context,
      initialDate: _selectedDate,
      firstDate: DateTime.now().subtract(const Duration(days: 30)),
      lastDate: DateTime.now().add(const Duration(days: 90)),
      builder: (context, child) {
        return Theme(
          data: Theme.of(context).copyWith(
            colorScheme: const ColorScheme.light(
              primary: _kEmerald,
              onPrimary: Colors.white,
              onSurface: _kSlate900,
            ),
          ),
          child: child!,
        );
      },
    );
    if (picked != null && picked != _selectedDate) {
      setState(() => _selectedDate = picked);
      _fetchPlans();
    }
  }

  Future<void> _deletePlan(String storeId) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: Colors.red.shade50,
                shape: BoxShape.circle,
              ),
              child: Icon(Icons.delete_outline_rounded,
                  color: Colors.red.shade400, size: 20),
            ),
            const SizedBox(width: 12),
            Text('Hapus Rencana',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w800)),
          ],
        ),
        content: Text(
          'Apakah Anda yakin ingin menghapus kunjungan ke toko ini dari rencana?',
          style: GoogleFonts.outfit(fontSize: 14, color: _kSlate500),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text('Batal',
                style: GoogleFonts.outfit(color: _kSlate500)),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red.shade500,
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10)),
            ),
            onPressed: () => Navigator.pop(ctx, true),
            child: Text('Hapus',
                style: GoogleFonts.outfit(
                    color: Colors.white, fontWeight: FontWeight.w700)),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    setState(() => _isLoading = true);
    try {
      final dateStr = DateFormat('yyyy-MM-dd').format(_selectedDate);
      await _salesApi.deleteManualVisitPlan(storeId, date: dateStr);
      try {
        di.sl<SyncBloc>().add(const SyncDataRequested(force: true));
      } catch (_) {}
      
      if (mounted) {
        setState(() {
          _visitPlans.removeWhere((plan) => plan['store_id'].toString() == storeId);
          
          _totalProjectedOmzet = 0.0;
          for (var p in _visitPlans) {
            _totalProjectedOmzet += (p['projected_omzet'] as num?)?.toDouble() ?? 0.0;
          }
          _isLoading = false;
        });

        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Rencana kunjungan berhasil dihapus'),
            backgroundColor: _kEmerald,
          ),
        );
      }
    } catch (e) {
      debugPrint('Error deleting plan: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal menghapus rencana: $e'),
            backgroundColor: Colors.red.shade700,
          ),
        );
      }
      setState(() => _isLoading = false);
    }
  }

  void _showProjectionDialog(String storeId, String storeName, List<dynamic> currentProjections) {
    final initial = currentProjections.map((e) => Map<String, dynamic>.from(e)).toList();

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
          initialProjections: initial,
          onSave: (newProjections) async {
            Navigator.pop(ctx);
            setState(() => _isLoading = true);
            try {
              await _salesApi.saveManualVisitPlan(
                planDate: _selectedDate,
                stores: [
                  {
                    'store_id': storeId,
                    'projections': newProjections,
                  }
                ],
              );
              try {
                di.sl<SyncBloc>().add(const SyncDataRequested(force: true));
              } catch (_) {}
              await _fetchPlans();
              if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Proyeksi berhasil disimpan'),
                    backgroundColor: _kEmerald,
                  ),
                );
              }
            } catch (e) {
              debugPrint('Error saving projections: $e');
              if (mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text('Gagal menyimpan proyeksi: $e'),
                    backgroundColor: Colors.red.shade700,
                  ),
                );
                setState(() => _isLoading = false);
              }
            }
          },
        );
      },
    );
  }

  // ─── Build ─────────────────────────────────────────────────────────────────
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _kBg,
      appBar: _buildAppBar(),
      body: Column(
        children: [
          // Date selector strip
          _buildDateSelector(),

          // Main content
          Expanded(
            child: _isLoading
                ? _buildSkeleton()
                : _visitPlans.isEmpty
                    ? _buildEmptyState()
                    : RefreshIndicator(
                        color: _kEmerald,
                        onRefresh: _fetchPlans,
                        child: CustomScrollView(
                          physics: const BouncingScrollPhysics(),
                          slivers: [
                            // Projection Header
                            SliverToBoxAdapter(
                              child: _buildProjectionHeader(),
                            ),
                            // Store list
                            SliverPadding(
                              padding: const EdgeInsets.fromLTRB(16, 8, 16, 100),
                              sliver: SliverList(
                                delegate: SliverChildBuilderDelegate(
                                  (context, index) =>
                                      _buildPlanCard(_visitPlans[index], index),
                                  childCount: _visitPlans.length,
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () async {
          await Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const VisitPlanBuilderPage()),
          );
          _fetchPlans();
        },
        backgroundColor: _kEmerald,
        icon: const Icon(Icons.add_rounded, color: Colors.white),
        label: Text(
          'Buat Rencana',
          style: GoogleFonts.outfit(
              fontWeight: FontWeight.w800, color: Colors.white),
        ),
      ),
    );
  }

  // ─── AppBar ───────────────────────────────────────────────────────────────
  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      backgroundColor: _kEmerald,
      elevation: 0,
      toolbarHeight: 70,
      leading: IconButton(
        icon: const Icon(Icons.arrow_back_ios_rounded, color: Colors.white),
        onPressed: () => Navigator.pop(context),
      ),
      title: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'KELOLA RENCANA',
            style: GoogleFonts.outfit(
              fontSize: 10,
              color: Colors.white70,
              fontWeight: FontWeight.w700,
              letterSpacing: 1.5,
            ),
          ),
          Text(
            'Daftar Kunjungan',
            style: GoogleFonts.outfit(
              fontSize: 18,
              fontWeight: FontWeight.w900,
              color: Colors.white,
            ),
          ),
        ],
      ),
      actions: [
        IconButton(
          icon: const Icon(Icons.refresh_rounded, color: Colors.white),
          tooltip: 'Refresh',
          onPressed: _fetchPlans,
        ),
        const SizedBox(width: 4),
      ],
    );
  }

  // ─── Date Selector ─────────────────────────────────────────────────────────
  Widget _buildDateSelector() {
    return Container(
      color: _kEmeraldDark,
      padding: const EdgeInsets.fromLTRB(16, 0, 16, 14),
      child: InkWell(
        onTap: _selectDate,
        borderRadius: BorderRadius.circular(14),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 11),
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(0.15),
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: Colors.white.withOpacity(0.3)),
          ),
          child: Row(
            children: [
              const Icon(Icons.calendar_month_rounded,
                  color: Colors.white, size: 20),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Menampilkan rencana untuk',
                      style: GoogleFonts.outfit(
                        fontSize: 10,
                        color: Colors.white70,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                    Text(
                      DateFormat('EEEE, dd MMMM yyyy', 'id_ID')
                          .format(_selectedDate),
                      style: GoogleFonts.outfit(
                        fontSize: 13,
                        fontWeight: FontWeight.w800,
                        color: Colors.white,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  'Ganti',
                  style: GoogleFonts.outfit(
                    fontSize: 12,
                    fontWeight: FontWeight.w700,
                    color: Colors.white,
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  // ─── Projection Header ─────────────────────────────────────────────────────
  Widget _buildProjectionHeader() {
    final completedCount =
        _visitPlans.where((p) => p['status'] == 'COMPLETED').length;
    final plannedCount = _visitPlans.length - completedCount;
    
    final currentPct = _monthlyTargetOmzet > 0 
        ? (_currentOmzet / _monthlyTargetOmzet).clamp(0.0, 1.0) 
        : 0.0;
    final projectedPct = _monthlyTargetOmzet > 0 
        ? (_totalProjectedOmzet / _monthlyTargetOmzet).clamp(0.0, 1.0 - currentPct) 
        : 0.0;
    final totalPct = ((currentPct + projectedPct) * 100).toStringAsFixed(1);

    return Container(
      margin: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [_kEmerald, _kEmeraldDark],
        ),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: _kEmerald.withOpacity(0.35),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Top row: label + total toko
            Row(
              children: [
                const Icon(Icons.trending_up_rounded,
                    color: Colors.white70, size: 14),
                const SizedBox(width: 6),
                Text(
                  'PROGRES OMZET BULAN INI',
                  style: GoogleFonts.outfit(
                    fontSize: 11,
                    color: Colors.white70,
                    fontWeight: FontWeight.w700,
                    letterSpacing: 1.2,
                  ),
                ),
                const Spacer(),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: Colors.white.withOpacity(0.15),
                    borderRadius: BorderRadius.circular(20),
                  ),
                  child: Text(
                    '${_visitPlans.length} Kunjungan',
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: Colors.white,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
              ],
            ),

            const SizedBox(height: 12),

            // Big projected value
            Text(
              _currencyFmt.format(_currentOmzet + _totalProjectedOmzet),
              style: GoogleFonts.outfit(
                fontSize: 30,
                fontWeight: FontWeight.w900,
                color: Colors.white,
                height: 1.1,
              ),
            ),
            const SizedBox(height: 2),
            Text(
              _monthlyTargetOmzet > 0
                  ? 'dari target bulanan ${_currencyFmt.format(_monthlyTargetOmzet)}'
                  : 'Target bulanan belum diatur',
              style: GoogleFonts.outfit(
                fontSize: 12,
                color: Colors.white70,
              ),
            ),

            const SizedBox(height: 16),

            // Progress bar
            ClipRRect(
              borderRadius: BorderRadius.circular(8),
              child: Container(
                height: 10,
                width: double.infinity,
                color: Colors.white.withOpacity(0.2),
                child: Row(
                  children: [
                    if (currentPct > 0)
                      Expanded(
                        flex: (currentPct * 1000).toInt(),
                        child: Container(color: Colors.white),
                      ),
                    if (projectedPct > 0)
                      Expanded(
                        flex: (projectedPct * 1000).toInt(),
                        child: Container(color: Colors.amberAccent),
                      ),
                    if (1.0 - currentPct - projectedPct > 0)
                      Expanded(
                        flex: ((1.0 - currentPct - projectedPct) * 1000).toInt(),
                        child: const SizedBox(),
                      ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Text(
                  '$totalPct% Tercapai',
                  style: GoogleFonts.outfit(
                    fontSize: 12,
                    color: Colors.white,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const Spacer(),
                if (_monthlyTargetOmzet > 0 && (currentPct + projectedPct) < 1.0)
                  Text(
                    'Kurang ${_currencyFmt.format(_monthlyTargetOmzet - (_currentOmzet + _totalProjectedOmzet))}',
                    style: GoogleFonts.outfit(
                      fontSize: 11,
                      color: Colors.white70,
                      fontWeight: FontWeight.w600,
                    ),
                  )
                else if (_monthlyTargetOmzet > 0)
                  Row(
                    children: [
                      const Icon(Icons.check_circle_rounded, size: 13, color: Colors.white),
                      const SizedBox(width: 4),
                      Text('Target Tercapai!',
                          style: GoogleFonts.outfit(
                              fontSize: 11,
                              color: Colors.white,
                              fontWeight: FontWeight.w800)),
                    ],
                  ),
              ],
            ),
            
            const SizedBox(height: 12),
            
            // Legend
            Row(
              children: [
                Container(
                  width: 8, height: 8,
                  decoration: const BoxDecoration(
                    color: Colors.white,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 4),
                Text('Saat ini: ${_currencyFmt.format(_currentOmzet)}', 
                  style: GoogleFonts.outfit(color: Colors.white, fontSize: 10)),
                const SizedBox(width: 12),
                Container(
                  width: 8, height: 8,
                  decoration: const BoxDecoration(
                    color: Colors.amberAccent,
                    shape: BoxShape.circle,
                  ),
                ),
                const SizedBox(width: 4),
                Text('+Proyeksi: ${_currencyFmt.format(_totalProjectedOmzet)}', 
                  style: GoogleFonts.outfit(color: Colors.white, fontSize: 10, fontWeight: FontWeight.w700)),
              ],
            ),

            const SizedBox(height: 16),
            const Divider(color: Colors.white24, height: 1),
            const SizedBox(height: 14),

            // Stat pills
            Row(
              children: [
                _buildStatPill(
                  Icons.check_circle_outline_rounded,
                  '$completedCount Selesai',
                  Colors.white,
                ),
                const SizedBox(width: 8),
                _buildStatPill(
                  Icons.schedule_rounded,
                  '$plannedCount Direncanakan',
                  Colors.white70,
                ),
                if (_totalProjectedOmzet > 0) ...[
                  const SizedBox(width: 8),
                  _buildStatPill(
                    Icons.show_chart_rounded,
                    'Avg ${_currencyFmt.format(_visitPlans.isEmpty ? 0 : _totalProjectedOmzet / _visitPlans.length)}/toko',
                    Colors.white70,
                  ),
                ],
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatPill(IconData icon, String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: Colors.white.withOpacity(0.15),
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: color),
          const SizedBox(width: 5),
          Text(
            label,
            style: GoogleFonts.outfit(
              fontSize: 11,
              fontWeight: FontWeight.w700,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  // ─── Plan Card ─────────────────────────────────────────────────────────────
  Widget _buildPlanCard(dynamic plan, int index) {
    final storeName =
        plan['store_name'] as String? ?? 'Toko Tidak Diketahui';
    final storeAddress = plan['store_address'] as String? ?? '-';
    final status = plan['status'] as String? ?? 'PLANNED';
    final storeId = plan['store_id']?.toString() ?? '';
    final isDone = status == 'COMPLETED' || status == 'EXTRA';

    final projections = plan['projections'] as List<dynamic>? ?? [];
    final projectedOmzet =
        (plan['projected_omzet'] as num?)?.toDouble() ?? 0.0;
    final lastTrxAmount =
        (plan['last_transaction_amount'] as num?)?.toDouble() ?? 0.0;
    final lastTrxItems = plan['last_transaction_items'] as String? ?? '';

    return Container(
      margin: const EdgeInsets.only(bottom: 14),
      decoration: BoxDecoration(
        color: isDone ? _kEmeraldLight : Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(
          color: isDone
              ? _kEmerald.withOpacity(0.3)
              : const Color(0xFFE2E8F0),
          width: 1.5,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 12,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ── Store Header Row ──────────────────────────────────────────
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 16, 12, 12),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Number badge
                Container(
                  width: 28,
                  height: 28,
                  margin: const EdgeInsets.only(right: 10, top: 6),
                  decoration: BoxDecoration(
                    color: _kEmerald.withOpacity(0.12),
                    shape: BoxShape.circle,
                  ),
                  alignment: Alignment.center,
                  child: Text(
                    '${index + 1}',
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w900,
                      color: _kEmeraldDark,
                    ),
                  ),
                ),
                // Store icon
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: isDone
                        ? _kEmerald.withOpacity(0.12)
                        : _kEmerald.withOpacity(0.08),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    Icons.storefront_rounded,
                    color: _kEmerald,
                    size: 24,
                  ),
                ),
                const SizedBox(width: 12),
                // Name & address
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        storeName,
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w800,
                          fontSize: 15,
                          color: _kSlate900,
                        ),
                      ),
                      const SizedBox(height: 3),
                      Row(
                        children: [
                          const Icon(Icons.location_on_rounded,
                              size: 12, color: _kSlate300),
                          const SizedBox(width: 3),
                          Expanded(
                            child: Text(
                              storeAddress,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: GoogleFonts.outfit(
                                  fontSize: 12, color: _kSlate500),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 8),
                // Status chip
                _buildStatusChip(status),
              ],
            ),
          ),

          // ── Divider ───────────────────────────────────────────────────
          if (projections.isNotEmpty || lastTrxAmount > 0)
            Divider(
              height: 1,
              color: isDone
                  ? _kEmerald.withOpacity(0.15)
                  : const Color(0xFFF1F5F9),
              indent: 16,
              endIndent: 16,
            ),

          // ── Projected Products ─────────────────────────────────────────
          if (projections.isNotEmpty)
            _buildProjectionSection(projections, projectedOmzet),

          // ── Last Transaction Reference ────────────────────────────────
          if (lastTrxAmount > 0)
            _buildLastTransactionRow(lastTrxAmount, lastTrxItems),

          // ── Action row ────────────────────────────────────────────────
          if (!isDone)
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 0, 12, 8),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton.icon(
                    onPressed: () => _showProjectionDialog(storeId, storeName, projections),
                    icon: const Icon(Icons.add_shopping_cart_rounded,
                        size: 16, color: _kEmerald),
                    label: Text(
                      'Atur Proyeksi',
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                        color: _kEmerald,
                      ),
                    ),
                    style: TextButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 12, vertical: 6),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                        side: BorderSide(
                            color: _kEmerald.withOpacity(0.3)),
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  TextButton.icon(
                    onPressed: () => _deletePlan(storeId),
                    icon: const Icon(Icons.delete_outline_rounded,
                        size: 16, color: Colors.redAccent),
                    label: Text(
                      'Hapus',
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                        color: Colors.redAccent,
                      ),
                    ),
                    style: TextButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 12, vertical: 6),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(10),
                        side: BorderSide(
                            color: Colors.redAccent.withOpacity(0.3)),
                      ),
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildStatusChip(String status) {
    final isDone = status == 'COMPLETED' || status == 'EXTRA';
    final isExtra = status == 'EXTRA';

    Color bg, fg;
    IconData icon;
    String label;

    if (isDone && isExtra) {
      bg = Colors.blue.shade50;
      fg = Colors.blue.shade700;
      icon = Icons.add_circle_outline_rounded;
      label = 'EXTRA';
    } else if (isDone) {
      bg = _kEmeraldLight;
      fg = _kEmeraldDark;
      icon = Icons.check_circle_outline_rounded;
      label = 'SELESAI';
    } else {
      bg = _kSlate100;
      fg = _kSlate500;
      icon = Icons.radio_button_unchecked_rounded;
      label = 'RENCANA';
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: fg),
          const SizedBox(width: 4),
          Text(
            label,
            style: GoogleFonts.outfit(
              fontSize: 10,
              fontWeight: FontWeight.w800,
              color: fg,
            ),
          ),
        ],
      ),
    );
  }

  // ── Projection section ─────────────────────────────────────────────────────
  Widget _buildProjectionSection(
      List<dynamic> projections, double projectedOmzet) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.inventory_2_rounded,
                  size: 13, color: _kSlate500),
              const SizedBox(width: 6),
              Text(
                'RENCANA PEMBELIAN',
                style: GoogleFonts.outfit(
                  fontSize: 10,
                  fontWeight: FontWeight.w800,
                  color: _kSlate500,
                  letterSpacing: 0.8,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),

          // Product rows
          ...projections.take(4).map((proj) {
            final productName =
                proj['product_name'] as String? ?? 'Produk';
            final qty = (proj['quantity'] as num?)?.toInt() ?? 0;
            final price = (proj['price'] as num?)?.toDouble() ?? 0.0;
            final subtotal = qty * price;

            return Padding(
              padding: const EdgeInsets.only(bottom: 5),
              child: Row(
                children: [
                  Container(
                    width: 6,
                    height: 6,
                    margin: const EdgeInsets.only(right: 8, top: 1),
                    decoration: const BoxDecoration(
                      color: _kEmerald,
                      shape: BoxShape.circle,
                    ),
                  ),
                  Expanded(
                    child: Text(
                      '$productName × $qty',
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        color: _kSlate700,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ),
                  Text(
                    _currencyFmt.format(subtotal),
                    style: GoogleFonts.outfit(
                      fontSize: 12,
                      fontWeight: FontWeight.w700,
                      color: _kSlate700,
                    ),
                  ),
                ],
              ),
            );
          }),

          if (projections.length > 4)
            Padding(
              padding: const EdgeInsets.only(bottom: 5),
              child: Text(
                '+${projections.length - 4} produk lainnya',
                style: GoogleFonts.outfit(
                    fontSize: 11,
                    color: _kSlate300,
                    fontStyle: FontStyle.italic),
              ),
            ),

          // Total projected
          if (projectedOmzet > 0) ...[
            const Padding(
              padding: EdgeInsets.symmetric(vertical: 6),
              child: Divider(color: Color(0xFFF1F5F9), height: 1),
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    const Icon(Icons.receipt_long_rounded,
                        size: 13, color: _kEmerald),
                    const SizedBox(width: 6),
                    Text(
                      'Proyeksi Total',
                      style: GoogleFonts.outfit(
                        fontSize: 12,
                        fontWeight: FontWeight.w700,
                        color: _kSlate700,
                      ),
                    ),
                  ],
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                  decoration: BoxDecoration(
                    color: _kEmerald.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Text(
                    _currencyFmt.format(projectedOmzet),
                    style: GoogleFonts.outfit(
                      fontSize: 13,
                      fontWeight: FontWeight.w900,
                      color: _kEmeraldDark,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
          ],
        ],
      ),
    );
  }

  // ── Last transaction reference ─────────────────────────────────────────────
  Widget _buildLastTransactionRow(double amount, String items) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 4, 16, 12),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 9),
        decoration: BoxDecoration(
          color: _kSlate100,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Icon(Icons.history_rounded, size: 14, color: _kSlate500),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Referensi Transaksi Terakhir',
                    style: GoogleFonts.outfit(
                      fontSize: 10,
                      fontWeight: FontWeight.w600,
                      color: _kSlate500,
                    ),
                  ),
                  const SizedBox(height: 2),
                  if (items.isNotEmpty)
                    Text(
                      items,
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        color: _kSlate700,
                        fontStyle: FontStyle.italic,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                ],
              ),
            ),
            const SizedBox(width: 8),
            Text(
              _currencyFmt.format(amount),
              style: GoogleFonts.outfit(
                fontSize: 12,
                fontWeight: FontWeight.w800,
                color: _kSlate700,
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ─── Empty State ───────────────────────────────────────────────────────────
  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(28),
            decoration: BoxDecoration(
              color: _kEmerald.withOpacity(0.07),
              shape: BoxShape.circle,
            ),
            child: const Icon(Icons.event_available_rounded,
                size: 64, color: _kEmerald),
          ),
          const SizedBox(height: 20),
          Text(
            'Belum Ada Rencana',
            style: GoogleFonts.outfit(
              fontSize: 20,
              fontWeight: FontWeight.w800,
              color: _kSlate700,
            ),
          ),
          const SizedBox(height: 8),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 40),
            child: Text(
              'Tidak ada rencana kunjungan pada tanggal ini.\nTekan tombol "Buat Rencana" untuk menambah.',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(
                fontSize: 13,
                color: _kSlate500,
                height: 1.5,
              ),
            ),
          ),
          const SizedBox(height: 24),
          ElevatedButton.icon(
            onPressed: () async {
              await Navigator.push(
                context,
                MaterialPageRoute(
                    builder: (_) => const VisitPlanBuilderPage()),
              );
              _fetchPlans();
            },
            icon: const Icon(Icons.add_rounded, size: 18),
            label: Text(
              'Buat Rencana Sekarang',
              style: GoogleFonts.outfit(fontWeight: FontWeight.w700),
            ),
            style: ElevatedButton.styleFrom(
              backgroundColor: _kEmerald,
              foregroundColor: Colors.white,
              padding:
                  const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(14)),
            ),
          ),
        ],
      ),
    );
  }

  // ─── Skeleton ──────────────────────────────────────────────────────────────
  Widget _buildSkeleton() {
    return Shimmer.fromColors(
      baseColor: Colors.grey.shade200,
      highlightColor: Colors.grey.shade50,
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            // Header skeleton
            Container(
              height: 190,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(24),
              ),
            ),
            const SizedBox(height: 16),
            // Card skeletons
            ...List.generate(
              4,
              (i) => Container(
                margin: const EdgeInsets.only(bottom: 14),
                height: 140,
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(20),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
