import 'dart:async';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/features/sales/presentation/pages/order_banner_page.dart';

// ─── Model ───────────────────────────────────────────────────────────────────

class BannerRequest {
  final int id;
  final String shopName;
  final String type;
  final String status;
  final double sizeLength;
  final double sizeWidth;
  final int qty;
  final String? notes;
  final String? designLink;
  final String? revisionNote;
  final double totalPrice;
  final String? installedAt;
  final String? createdAt;
  final String? salesName;
  final String? editorName;

  BannerRequest({
    required this.id,
    required this.shopName,
    required this.type,
    required this.status,
    required this.sizeLength,
    required this.sizeWidth,
    required this.qty,
    this.notes,
    this.designLink,
    this.revisionNote,
    required this.totalPrice,
    this.installedAt,
    this.createdAt,
    this.salesName,
    this.editorName,
  });

  factory BannerRequest.fromJson(Map<String, dynamic> json) {
    return BannerRequest(
      id: (json['id'] ?? 0) is int
          ? json['id'] ?? 0
          : int.tryParse(json['id'].toString()) ?? 0,
      shopName:
          json['shop_name'] ?? json['shop']?['name'] ?? 'Toko Tidak Diketahui',
      type: json['type'] ?? 'Spanduk',
      status: json['status'] ?? 'request',
      sizeLength: (json['size_length'] ?? 0).toDouble(),
      sizeWidth: (json['size_width'] ?? 0).toDouble(),
      qty: (json['qty'] ?? 1) is int
          ? json['qty'] ?? 1
          : int.tryParse(json['qty'].toString()) ?? 1,
      notes: json['notes'],
      designLink: json['design_link'],
      revisionNote: json['revision_note'],
      totalPrice: (json['total_price'] ?? 0).toDouble(),
      installedAt: json['installed_at'],
      createdAt: json['created_at'],
      salesName: json['sales_name'],
      editorName: json['editor_name'],
    );
  }
}

// ─── Status Config ────────────────────────────────────────────────────────────

class _StatusConfig {
  final String label;
  final Color color;
  final Color bgColor;
  final IconData icon;
  final int step; // 0-based index in the pipeline

  const _StatusConfig({
    required this.label,
    required this.color,
    required this.bgColor,
    required this.icon,
    required this.step,
  });
}

const Map<String, _StatusConfig> _statusMap = {
  'request': _StatusConfig(
    label: 'Request Dikirim',
    color: Color(0xFF3B82F6),
    bgColor: Color(0xFFEFF6FF),
    icon: Icons.send_rounded,
    step: 0,
  ),
  'validation': _StatusConfig(
    label: 'Sedang Didesain',
    color: Color(0xFF8B5CF6),
    bgColor: Color(0xFFF5F3FF),
    icon: Icons.palette_rounded,
    step: 1,
  ),
  'purchasing_queue': _StatusConfig(
    label: 'Antrian Pembelian',
    color: Color(0xFFF59E0B),
    bgColor: Color(0xFFFFFBEB),
    icon: Icons.shopping_cart_rounded,
    step: 2,
  ),
  'sent_to_vendor': _StatusConfig(
    label: 'Di Percetakan',
    color: Color(0xFFF97316),
    bgColor: Color(0xFFFFF7ED),
    icon: Icons.print_rounded,
    step: 3,
  ),
  'ready_for_operator': _StatusConfig(
    label: 'Siap Dipasang',
    color: Color(0xFF10B981),
    bgColor: Color(0xFFECFDF5),
    icon: Icons.local_shipping_rounded,
    step: 4,
  ),
  'installed': _StatusConfig(
    label: 'Terpasang',
    color: Color(0xFF059669),
    bgColor: Color(0xFFD1FAE5),
    icon: Icons.check_circle_rounded,
    step: 5,
  ),
  'terpasang': _StatusConfig(
    label: 'Terpasang',
    color: Color(0xFF059669),
    bgColor: Color(0xFFD1FAE5),
    icon: Icons.check_circle_rounded,
    step: 5,
  ),
  'revision': _StatusConfig(
    label: 'Perlu Revisi',
    color: Color(0xFFEF4444),
    bgColor: Color(0xFFFEF2F2),
    icon: Icons.edit_note_rounded,
    step: 1,
  ),
};

_StatusConfig _getStatus(String status) {
  return _statusMap[status.toLowerCase()] ??
      _StatusConfig(
        label: status,
        color: Colors.blueGrey,
        bgColor: const Color(0xFFF8FAFC),
        icon: Icons.help_outline_rounded,
        step: 0,
      );
}

// ─── Page ─────────────────────────────────────────────────────────────────────

class BrandingTrackerPage extends StatefulWidget {
  const BrandingTrackerPage({super.key});

  @override
  State<BrandingTrackerPage> createState() => _BrandingTrackerPageState();
}

class _BrandingTrackerPageState extends State<BrandingTrackerPage>
    with SingleTickerProviderStateMixin {
  final ApiClient _api = di.sl<ApiClient>();
  bool _isLoading = false;
  bool _hasError = false;
  String _errorMessage = '';
  List<BannerRequest> _banners = [];
  String _filterStatus = ''; // '' = all
  Timer? _refreshTimer;

  late TabController _tabController;
  final _tabs = const [
    {'label': 'Semua', 'status': ''},
    {'label': 'Request', 'status': 'request'},
    {'label': 'Desain', 'status': 'validation'},
    {'label': 'Cetak', 'status': 'sent_to_vendor'},
    {'label': 'Siap Pasang', 'status': 'ready_for_operator'},
    {'label': 'Terpasang', 'status': 'installed'},
  ];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: _tabs.length, vsync: this);
    _tabController.addListener(() {
      if (!_tabController.indexIsChanging) {
        setState(() => _filterStatus = _tabs[_tabController.index]['status']!);
      }
    });
    _fetchBanners();
    // Auto-refresh every 30 seconds
    _refreshTimer = Timer.periodic(
        const Duration(seconds: 30), (_) => _fetchBanners(silent: true));
  }

  @override
  void dispose() {
    _tabController.dispose();
    _refreshTimer?.cancel();
    super.dispose();
  }

  Future<void> _fetchBanners({bool silent = false}) async {
    if (!silent)
      setState(() {
        _isLoading = true;
        _hasError = false;
      });
    try {
      final response = await _api.client.get('branding/banners');
      final raw = response.data['data'] ?? [];
      final banners = (raw as List)
          .map((e) => BannerRequest.fromJson(e as Map<String, dynamic>))
          .toList();
      // Sort by newest first
      banners.sort((a, b) => (b.id).compareTo(a.id));
      if (mounted) setState(() => _banners = banners);
    } catch (e) {
      if (!silent && mounted) {
        setState(() {
          _hasError = true;
          _errorMessage = e.toString().replaceAll('Exception:', '').trim();
        });
      }
    } finally {
      if (!silent && mounted) setState(() => _isLoading = false);
    }
  }

  List<BannerRequest> get _filteredBanners {
    if (_filterStatus.isEmpty) return _banners;
    final norm = _filterStatus.toLowerCase();
    return _banners.where((b) {
      final s = b.status.toLowerCase();
      if (norm == 'installed') return s == 'installed' || s == 'terpasang';
      return s == norm;
    }).toList();
  }

  int _countByStatus(String status) {
    if (status.isEmpty) return _banners.length;
    if (status == 'installed') {
      return _banners.where((b) {
        final s = b.status.toLowerCase();
        return s == 'installed' || s == 'terpasang';
      }).length;
    }
    return _banners
        .where((b) => b.status.toLowerCase() == status.toLowerCase())
        .length;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      appBar: _buildAppBar(),
      body: Column(
        children: [
          _buildSummaryRow(),
          _buildTabBar(),
          Expanded(child: _buildBody()),
        ],
      ),
      floatingActionButton: _buildFAB(),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      backgroundColor: Colors.white,
      elevation: 0,
      leading: IconButton(
        icon:
            const Icon(Icons.arrow_back_ios_rounded, color: Color(0xFF334155)),
        onPressed: () => Navigator.pop(context),
      ),
      title: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('BRANDING',
              style: GoogleFonts.outfit(
                  fontSize: 10,
                  color: Colors.blueAccent,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1.2)),
          Text('Tracker Spanduk',
              style: GoogleFonts.outfit(
                  fontSize: 17,
                  fontWeight: FontWeight.w800,
                  color: const Color(0xFF1E293B))),
        ],
      ),
      actions: [
        IconButton(
          icon: const Icon(Icons.refresh_rounded, color: Color(0xFF334155)),
          onPressed: _fetchBanners,
          tooltip: 'Refresh',
        ),
        const SizedBox(width: 8),
      ],
    );
  }

  Widget _buildSummaryRow() {
    final installed = _countByStatus('installed');
    final inProgress = _banners.where((b) {
      final s = b.status.toLowerCase();
      return s != 'installed' && s != 'terpasang';
    }).length;
    final total = _banners.length;

    return Container(
      color: Colors.white,
      padding: const EdgeInsets.fromLTRB(20, 12, 20, 16),
      child: Row(
        children: [
          _buildSummaryCard(
              'Total', total, Colors.blueAccent, Icons.receipt_long_rounded),
          const SizedBox(width: 12),
          _buildSummaryCard('Proses', inProgress, const Color(0xFFF59E0B),
              Icons.pending_rounded),
          const SizedBox(width: 12),
          _buildSummaryCard('Terpasang', installed, const Color(0xFF10B981),
              Icons.check_circle_outline_rounded),
        ],
      ),
    );
  }

  Widget _buildSummaryCard(
      String label, int count, Color color, IconData icon) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 14),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: color.withValues(alpha: 0.2)),
        ),
        child: Row(
          children: [
            Icon(icon, color: color, size: 18),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('$count',
                      style: GoogleFonts.outfit(
                          fontSize: 20,
                          fontWeight: FontWeight.w900,
                          color: color,
                          height: 1)),
                  Text(label,
                      style: GoogleFonts.outfit(
                          fontSize: 11,
                          fontWeight: FontWeight.w600,
                          color: color.withValues(alpha: 0.7))),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTabBar() {
    return Container(
      color: Colors.white,
      child: TabBar(
        controller: _tabController,
        isScrollable: true,
        labelColor: Colors.blueAccent,
        unselectedLabelColor: Colors.blueGrey,
        labelStyle:
            GoogleFonts.outfit(fontWeight: FontWeight.w700, fontSize: 13),
        unselectedLabelStyle:
            GoogleFonts.outfit(fontWeight: FontWeight.w500, fontSize: 13),
        indicatorColor: Colors.blueAccent,
        indicatorWeight: 3,
        tabs: _tabs.map((t) {
          final count = _countByStatus(t['status']!);
          return Tab(
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(t['label']!),
                if (count > 0) ...[
                  const SizedBox(width: 6),
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: Colors.blueAccent.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text('$count',
                        style: GoogleFonts.outfit(
                            fontSize: 10,
                            fontWeight: FontWeight.w800,
                            color: Colors.blueAccent)),
                  ),
                ]
              ],
            ),
          );
        }).toList(),
      ),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(
          child: CircularProgressIndicator(color: Colors.blueAccent));
    }
    if (_hasError) {
      return _buildErrorState();
    }
    final items = _filteredBanners;
    if (items.isEmpty) {
      return _buildEmptyState();
    }
    return RefreshIndicator(
      onRefresh: _fetchBanners,
      color: Colors.blueAccent,
      child: ListView.builder(
        padding: const EdgeInsets.fromLTRB(16, 16, 16, 100),
        itemCount: items.length,
        itemBuilder: (_, i) => _BannerCard(banner: items[i]),
      ),
    );
  }

  Widget _buildErrorState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.wifi_off_rounded,
              size: 64, color: Colors.blueGrey.withValues(alpha: 0.4)),
          const SizedBox(height: 16),
          Text('Gagal memuat data',
              style: GoogleFonts.outfit(
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                  color: const Color(0xFF334155))),
          const SizedBox(height: 8),
          Text(_errorMessage,
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(fontSize: 12, color: Colors.blueGrey)),
          const SizedBox(height: 24),
          ElevatedButton.icon(
            onPressed: _fetchBanners,
            icon: const Icon(Icons.refresh_rounded),
            label: Text('Coba Lagi',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
            style: ElevatedButton.styleFrom(
                backgroundColor: Colors.blueAccent,
                foregroundColor: Colors.white,
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(14))),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(
                color: Colors.blueAccent.withValues(alpha: 0.08),
                shape: BoxShape.circle),
            child: const Icon(Icons.art_track_rounded,
                size: 56, color: Colors.blueAccent),
          ),
          const SizedBox(height: 20),
          Text('Belum Ada Request Spanduk',
              style: GoogleFonts.outfit(
                  fontSize: 16,
                  fontWeight: FontWeight.w800,
                  color: const Color(0xFF1E293B))),
          const SizedBox(height: 8),
          Text(
            _filterStatus.isEmpty
                ? 'Buat request spanduk pertama Anda\nuntuk outlet yang ingin dibranding.'
                : 'Tidak ada spanduk dengan status ini.',
            textAlign: TextAlign.center,
            style: GoogleFonts.outfit(fontSize: 13, color: Colors.blueGrey),
          ),
          if (_filterStatus.isEmpty) ...[
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: _goToCreateBanner,
              icon: const Icon(Icons.add_rounded),
              label: Text('Buat Request Baru',
                  style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
              style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.blueAccent,
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(14)),
                  padding:
                      const EdgeInsets.symmetric(horizontal: 24, vertical: 14)),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildFAB() {
    return FloatingActionButton.extended(
      onPressed: _goToCreateBanner,
      backgroundColor: Colors.blueAccent,
      elevation: 8,
      icon: const Icon(Icons.add_rounded, color: Colors.white),
      label: Text('Request Baru',
          style: GoogleFonts.outfit(
              fontWeight: FontWeight.w700, color: Colors.white)),
    );
  }

  void _goToCreateBanner() {
    Navigator.of(context)
        .push(MaterialPageRoute(builder: (_) => const OrderBannerPage()))
        .then((_) => _fetchBanners());
  }
}

// ─── Banner Card ──────────────────────────────────────────────────────────────

class _BannerCard extends StatelessWidget {
  final BannerRequest banner;
  const _BannerCard({required this.banner});

  @override
  Widget build(BuildContext context) {
    final cfg = _getStatus(banner.status);
    final currencyFormat =
        NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);
    final size = '${banner.sizeLength} × ${banner.sizeWidth} m';
    final createdStr = _formatDate(banner.createdAt);

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withValues(alpha: 0.05),
              blurRadius: 20,
              offset: const Offset(0, 8))
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ── Header ──
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 18, 20, 14),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                      color: cfg.bgColor,
                      borderRadius: BorderRadius.circular(14)),
                  child: Icon(cfg.icon, color: cfg.color, size: 22),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(banner.shopName,
                          style: GoogleFonts.outfit(
                              fontSize: 15,
                              fontWeight: FontWeight.w800,
                              color: const Color(0xFF1E293B))),
                      const SizedBox(height: 4),
                      Row(
                        children: [
                          _chip(
                            banner.type,
                            banner.type.toLowerCase() == 'universal'
                                ? const Color(0xFF0284C7)
                                : const Color(0xFF6366F1),
                          ),
                          const SizedBox(width: 6),
                          _chip(size, Colors.blueGrey),
                          if (banner.qty > 1) ...[
                            const SizedBox(width: 6),
                            _chip('×${banner.qty}', Colors.blueGrey)
                          ],
                        ],
                      ),
                    ],
                  ),
                ),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 10, vertical: 5),
                      decoration: BoxDecoration(
                          color: cfg.bgColor,
                          borderRadius: BorderRadius.circular(20)),
                      child: Text(cfg.label,
                          style: GoogleFonts.outfit(
                              fontSize: 11,
                              fontWeight: FontWeight.w700,
                              color: cfg.color)),
                    ),
                    const SizedBox(height: 6),
                    Text('#${banner.id}',
                        style: GoogleFonts.outfit(
                            fontSize: 11,
                            color: Colors.blueGrey.withValues(alpha: 0.6))),
                  ],
                ),
              ],
            ),
          ),

          // ── Pipeline Progress ──
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            child: _PipelineIndicator(
                currentStep: cfg.step,
                isRevision: banner.status.toLowerCase() == 'revision'),
          ),
          const SizedBox(height: 14),

          // ── Revision Note (if any) ──
          if (banner.revisionNote != null && banner.revisionNote!.isNotEmpty)
            Padding(
              padding: const EdgeInsets.fromLTRB(20, 0, 20, 12),
              child: Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: const Color(0xFFFEF2F2),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: const Color(0xFFFCA5A5)),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Icon(Icons.edit_note_rounded,
                        color: Color(0xFFEF4444), size: 18),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Catatan Revisi',
                              style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w800,
                                  color: const Color(0xFFEF4444))),
                          const SizedBox(height: 2),
                          Text(banner.revisionNote!,
                              style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  color: const Color(0xFF7F1D1D))),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),

          // ── Footer ──
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 0, 20, 16),
            child: Row(
              children: [
                if (createdStr.isNotEmpty) ...[
                  const Icon(Icons.access_time_rounded,
                      size: 13, color: Color(0xFF94A3B8)),
                  const SizedBox(width: 4),
                  Text(createdStr,
                      style: GoogleFonts.outfit(
                          fontSize: 11, color: const Color(0xFF94A3B8))),
                ],
                const Spacer(),
                if (banner.totalPrice > 0)
                  Text(currencyFormat.format(banner.totalPrice),
                      style: GoogleFonts.outfit(
                          fontSize: 13,
                          fontWeight: FontWeight.w800,
                          color: const Color(0xFF059669))),
                if (banner.installedAt != null) ...[
                  const SizedBox(width: 8),
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                        color: const Color(0xFFD1FAE5),
                        borderRadius: BorderRadius.circular(8)),
                    child: Row(
                      children: [
                        const Icon(Icons.check_rounded,
                            size: 12, color: Color(0xFF059669)),
                        const SizedBox(width: 4),
                        Text(_formatDate(banner.installedAt),
                            style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w700,
                                color: const Color(0xFF059669))),
                      ],
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _chip(String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
          color: color.withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular(8)),
      child: Text(label,
          style: GoogleFonts.outfit(
              fontSize: 11, fontWeight: FontWeight.w600, color: color)),
    );
  }

  String _formatDate(String? raw) {
    if (raw == null || raw.isEmpty) return '';
    try {
      final dt = DateTime.parse(raw).toLocal();
      return DateFormat('d MMM yyyy').format(dt);
    } catch (_) {
      return raw.length > 10 ? raw.substring(0, 10) : raw;
    }
  }
}

// ─── Pipeline Indicator ───────────────────────────────────────────────────────

class _PipelineIndicator extends StatelessWidget {
  final int currentStep; // 0–5
  final bool isRevision;
  const _PipelineIndicator(
      {required this.currentStep, this.isRevision = false});

  static const _steps = [
    Icons.send_rounded,
    Icons.palette_rounded,
    Icons.shopping_cart_rounded,
    Icons.print_rounded,
    Icons.local_shipping_rounded,
    Icons.check_circle_rounded,
  ];

  @override
  Widget build(BuildContext context) {
    return Row(
      children: List.generate(_steps.length * 2 - 1, (index) {
        if (index.isOdd) {
          // Connector line
          final stepBefore = index ~/ 2;
          final isDone = stepBefore < currentStep;
          return Expanded(
            child: Container(
              height: 2,
              color: isDone ? Colors.blueAccent : const Color(0xFFE2E8F0),
            ),
          );
        }
        // Step circle
        final step = index ~/ 2;
        final isDone = step < currentStep;
        final isCurrent = step == currentStep;
        final isRevisionStep = isRevision && step == currentStep;

        Color circleColor;
        Color iconColor;
        if (isRevisionStep) {
          circleColor = const Color(0xFFFEF2F2);
          iconColor = const Color(0xFFEF4444);
        } else if (isDone) {
          circleColor = Colors.blueAccent;
          iconColor = Colors.white;
        } else if (isCurrent) {
          circleColor = Colors.blueAccent.withValues(alpha: 0.15);
          iconColor = Colors.blueAccent;
        } else {
          circleColor = const Color(0xFFF1F5F9);
          iconColor = const Color(0xFFCBD5E1);
        }

        return AnimatedContainer(
          duration: const Duration(milliseconds: 300),
          width: isCurrent ? 32 : 26,
          height: isCurrent ? 32 : 26,
          decoration: BoxDecoration(
            color: circleColor,
            shape: BoxShape.circle,
            border: isCurrent
                ? Border.all(color: Colors.blueAccent, width: 2)
                : null,
            boxShadow: isCurrent
                ? [
                    BoxShadow(
                        color: Colors.blueAccent.withValues(alpha: 0.3),
                        blurRadius: 8)
                  ]
                : null,
          ),
          child:
              Icon(_steps[step], size: isCurrent ? 16 : 13, color: iconColor),
        );
      }),
    );
  }
}
