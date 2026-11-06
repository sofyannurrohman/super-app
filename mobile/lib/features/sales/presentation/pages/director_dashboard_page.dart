import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:hris_app/injection.dart' as di;

class DirectorDashboardPage extends StatefulWidget {
  const DirectorDashboardPage({super.key});

  @override
  State<DirectorDashboardPage> createState() => _DirectorDashboardPageState();
}

class _DirectorDashboardPageState extends State<DirectorDashboardPage> {
  final _salesApi = di.sl<SalesApiService>();
  
  
  bool _isLoading = true;
  DateTime _selectedDate = DateTime.now();
  Map<String, dynamic>? _summaryData;
  
  String _selectedBranchId = 'all';


  @override
  void initState() {
    super.initState();
    _fetchSummary();
  }

  
  
  Future<void> _fetchSummary() async {
    setState(() => _isLoading = true);
    try {
      final dateStr = DateFormat('yyyy-MM-dd').format(_selectedDate);
      final data = await _salesApi.getDirectorSummary(date: dateStr);
      
      if (mounted) {
        setState(() {
          _summaryData = data;
          final availableBranchIds = (data['branches'] as List<dynamic>? ?? [])
              .map((b) => b['branch_id']?.toString())
              .whereType<String>()
              .toSet();
          if (_selectedBranchId != 'all' &&
              !availableBranchIds.contains(_selectedBranchId)) {
            _selectedBranchId = 'all';
          }
          _isLoading = false;
        });
      }
    } catch (e) {
      debugPrint('Error fetching director summary: $e');
      if (mounted) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal memuat data: $e')),
        );
      }
    }
  }

  Future<void> _fetchChartOnly() async {
    // No-op for production endpoint usage
    // Dropdown change directly triggers UI rebuild with existing _summaryData
  }




  Future<void> _selectDate(BuildContext context) async {
    final picked = await showDatePicker(
      context: context,
      initialDate: _selectedDate,
      firstDate: DateTime(2020),
      lastDate: DateTime.now(),
    );
    if (picked != null && picked != _selectedDate) {
      setState(() => _selectedDate = picked);
      _fetchSummary();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text(
          'DASHBOARD DIREKTUR',
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.w800,
            fontSize: 16,
            letterSpacing: 0.5,
          ),
        ),
        centerTitle: true,
        flexibleSpace: Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: [Color(0xFF059669), Color(0xFF10B981)], // Emerald for Director
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.calendar_month, color: Colors.white),
            onPressed: () => _selectDate(context),
            tooltip: 'Pilih Tanggal',
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _buildBody(),
    );
  }


  Widget _buildBody() {
    if (_summaryData == null) {
      return const Center(child: Text('Tidak ada data'));
    }

    final formatCurrency = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp',
      decimalDigits: 0,
    );

    final branches = _summaryData!['branches'] as List<dynamic>? ?? [];
    
    // Filter by branch
    num grandToday = 0;
    num grandYesterday = 0;
    num grandMonth = 0;
    
    if (_selectedBranchId == 'all') {
      grandToday = _summaryData!['grand_today'] as num? ?? 0;
      grandYesterday = _summaryData!['grand_yesterday'] as num? ?? 0;
      grandMonth = _summaryData!['grand_month'] as num? ?? 0;
    } else {
      final branchData = branches.cast<Map>().where((b) =>
          b['branch_id']?.toString() == _selectedBranchId).firstOrNull;
      if (branchData != null) {
        grandToday = branchData['today_omzet'] as num? ?? 0;
        grandYesterday = branchData['yesterday_omzet'] as num? ?? 0;
        grandMonth = branchData['month_omzet'] as num? ?? 0;
      }
    }

    // Branch Dropdown Items
    final branchItems = [
      const DropdownMenuItem(value: 'all', child: Text('Semua Cabang')),
      ...branches.map((b) => DropdownMenuItem(
        value: b['branch_id'].toString(),
        child: Text(b['branch_name']?.toString() ?? 'Unknown'),
      )),
    ];

    return RefreshIndicator(
      onRefresh: _fetchSummary,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Text(
            'Ringkasan bisnis dan performa omzet perusahaan',
            style: GoogleFonts.outfit(
              color: const Color(0xFF64748B),
              fontSize: 13,
            ),
          ),
          const SizedBox(height: 12),
          // Dropdown Cabang
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.grey.shade300),
            ),
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                value: _selectedBranchId,
                isExpanded: true,
                items: branchItems,
                onChanged: (val) {
                  if (val != null && val != _selectedBranchId) {
                    setState(() => _selectedBranchId = val);
                    _fetchChartOnly();
                  }
                },
              ),
            ),
          ),
          const SizedBox(height: 16),

          Text('INSIGHT & PERINGATAN', style: GoogleFonts.outfit(fontSize: 14, fontWeight: FontWeight.w800)),
          const SizedBox(height: 10),
          _buildPerformanceInsight(grandToday, grandYesterday, branches, formatCurrency),
          const SizedBox(height: 16),
          _buildTransactionDetails(formatCurrency),
          const SizedBox(height: 24),
          
          // Global/Filtered Summary
          Container(
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: [Color(0xFF2C3E50), Color(0xFF34495E)],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.1),
                  blurRadius: 10,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: Column(
              children: [
                Text(
                  _selectedBranchId == 'all' ? 'TOTAL OMZET GLOBAL' : 'OMZET CABANG TERPILIH',
                  style: GoogleFonts.outfit(
                    color: Colors.white.withOpacity(0.8),
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    letterSpacing: 1.5,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  formatCurrency.format(grandMonth),
                  style: GoogleFonts.outfit(
                    color: Colors.white,
                    fontSize: 32,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Text(
                  'Bulan Ini',
                  style: GoogleFonts.outfit(
                    color: Colors.white.withOpacity(0.7),
                    fontSize: 12,
                  ),
                ),
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 16),
                  child: Divider(color: Colors.white24, height: 1),
                ),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceAround,
                  children: [
                    Column(
                      children: [
                        Text(
                          'Hari Ini',
                          style: GoogleFonts.outfit(
                            color: Colors.white.withOpacity(0.8),
                            fontSize: 12,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          formatCurrency.format(grandToday),
                          style: GoogleFonts.outfit(
                            color: const Color(0xFF4ADE80),
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                    Container(width: 1, height: 30, color: Colors.white24),
                    Column(
                      children: [
                        Text(
                          'Kemarin',
                          style: GoogleFonts.outfit(
                            color: Colors.white.withOpacity(0.8),
                            fontSize: 12,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          formatCurrency.format(grandYesterday),
                          style: GoogleFonts.outfit(
                            color: Colors.white,
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),
          
          Text(
            _selectedBranchId == 'all' ? 'PERBANDINGAN OMZET (BULAN INI)' : 'OMZET HARI INI VS KEMARIN',
            style: GoogleFonts.outfit(
              color: Colors.black87,
              fontSize: 14,
              fontWeight: FontWeight.w800,
              letterSpacing: 1,
            ),
          ),
          const SizedBox(height: 16),
          _buildChart(formatCurrency),
          
          if (_selectedBranchId == 'all') ...[
            const SizedBox(height: 24),
            Text(
              'PERFORMA PER CABANG',
              style: GoogleFonts.outfit(
                color: Colors.black87,
                fontSize: 14,
                fontWeight: FontWeight.w800,
                letterSpacing: 1,
              ),
            ),
            const SizedBox(height: 12),

            if (branches.isEmpty)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(32.0),
                  child: Text('Tidak ada cabang'),
                ),
              ),

            ...branches.map((b) => _buildBranchCard(b, formatCurrency)),
          ]
        ],
      ),
    );
  }

  Widget _buildPerformanceInsight(
    num today,
    num yesterday,
    List<dynamic> branches,
    NumberFormat formatCurrency,
  ) {
    final difference = today - yesterday;
    final percentage = yesterday == 0 ? null : (difference / yesterday) * 100;
    final isUp = difference >= 0;
    String insight = isUp ? 'Omzet hari ini meningkat' : 'Omzet hari ini menurun';
    if (_selectedBranchId == 'all' && branches.isNotEmpty) {
      final sorted = [...branches]..sort((a, b) =>
          ((b['month_omzet'] as num?) ?? 0).compareTo((a['month_omzet'] as num?) ?? 0));
      final best = sorted.first['branch_name']?.toString() ?? '—';
      final lowest = sorted.last['branch_name']?.toString() ?? '—';
      insight += '. Terbaik: $best, perlu perhatian: $lowest.';
    }

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: isUp ? const Color(0xFFECFDF5) : const Color(0xFFFFF7ED),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: isUp ? const Color(0xFFA7F3D0) : const Color(0xFFFED7AA)),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(isUp ? Icons.trending_up_rounded : Icons.trending_down_rounded,
              color: isUp ? const Color(0xFF059669) : const Color(0xFFEA580C)),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              percentage == null
                  ? '$insight (${formatCurrency.format(difference.abs())})'
                  : '$insight ${percentage.abs().toStringAsFixed(1)}% dibanding kemarin.',
              style: GoogleFonts.outfit(fontSize: 12, fontWeight: FontWeight.w600),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTransactionDetails(NumberFormat formatCurrency) {
    final stats = (_summaryData!['transaction_stats'] as Map?)?.cast<String, dynamic>() ?? {};
    final items = [
      ('Total transaksi', stats['total'] ?? 0, Icons.receipt_long_rounded),
      ('Selesai', stats['completed'] ?? 0, Icons.check_circle_outline_rounded),
      ('Pending', stats['pending'] ?? 0, Icons.schedule_rounded),
      ('Retur / batal', stats['cancelled'] ?? 0, Icons.remove_circle_outline_rounded),
    ];
    final average = (stats['average'] as num?)?.toDouble() ?? 0;
    return Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Text('DETAIL TRANSAKSI BULAN INI', style: GoogleFonts.outfit(fontSize: 14, fontWeight: FontWeight.w800)),
      const SizedBox(height: 10),
      Wrap(spacing: 10, runSpacing: 10, children: items.map((item) => _statCard(item.$1, item.$2.toString(), item.$3)).toList()),
      const SizedBox(height: 10),
      _statCard('Rata-rata nilai transaksi', formatCurrency.format(average), Icons.payments_outlined),
    ]);
  }

  Widget _statCard(String label, String value, IconData icon) => Container(
    width: 160, padding: const EdgeInsets.all(12),
    decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
    child: Row(children: [Icon(icon, size: 20, color: const Color(0xFF059669)), const SizedBox(width: 8), Expanded(child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [Text(label, style: const TextStyle(fontSize: 10, color: Colors.grey)), Text(value, style: GoogleFonts.outfit(fontWeight: FontWeight.w700))]))]),
  );


  Widget _buildChart(NumberFormat formatCurrency) {
    if (_summaryData == null) return const SizedBox();
    
    final branches = _summaryData!['branches'] as List<dynamic>? ?? [];
    if (branches.isEmpty) {
      return Container(
        height: 200,
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
        ),
        child: const Center(child: Text('Belum ada data penjualan')),
      );
    }

    final isAllBranches = _selectedBranchId == 'all';
    double maxY = 0;
    final barGroups = <BarChartGroupData>[];
    final titles = <int, String>{};

    if (isAllBranches) {
      // Bandingkan omzet bulan ini antar cabang (Top 5 atau semua jika sedikit)
      for (int i = 0; i < branches.length; i++) {
        final b = branches[i];
        final val = (b['month_omzet'] as num?)?.toDouble() ?? 0.0;
        if (val > maxY) maxY = val;
        
        String bName = b['branch_name']?.toString() ?? 'Unknown';
        if (bName.length > 6) bName = bName.substring(0, 6);
        titles[i] = bName;

        barGroups.add(
          BarChartGroupData(
            x: i,
            barRods: [
              BarChartRodData(
                toY: val,
                color: const Color(0xFF059669),
                width: 16,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(4)),
              ),
            ],
          ),
        );
      }
    } else {
      // Bandingkan Hari Ini vs Kemarin untuk cabang spesifik
      final branchData = branches.cast<Map>().where((b) =>
          b['branch_id']?.toString() == _selectedBranchId).firstOrNull;
      
      if (branchData != null) {
        final today = (branchData['today_omzet'] as num?)?.toDouble() ?? 0.0;
        final yesterday = (branchData['yesterday_omzet'] as num?)?.toDouble() ?? 0.0;
        if (today > maxY) maxY = today;
        if (yesterday > maxY) maxY = yesterday;
        
        titles[0] = 'Kemarin';
        titles[1] = 'Hari Ini';
        
        barGroups.add(
          BarChartGroupData(
            x: 0,
            barRods: [
              BarChartRodData(
                toY: yesterday,
                color: const Color(0xFF64748B),
                width: 32,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(6)),
              ),
            ],
          ),
        );
        barGroups.add(
          BarChartGroupData(
            x: 1,
            barRods: [
              BarChartRodData(
                toY: today,
                color: const Color(0xFF059669),
                width: 32,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(6)),
              ),
            ],
          ),
        );
      }
    }

    if (maxY == 0) maxY = 1000000;
    
    return Container(
      height: 250,
      padding: const EdgeInsets.only(top: 24, bottom: 16, left: 16, right: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: BarChart(
        BarChartData(
          maxY: maxY * 1.2,
          barTouchData: BarTouchData(
            touchTooltipData: BarTouchTooltipData(
              getTooltipColor: (group) => Colors.blueGrey,
              getTooltipItem: (group, groupIndex, rod, rodIndex) {
                return BarTooltipItem(
                  '${titles[group.x]}\\n${formatCurrency.format(rod.toY)}',
                  const TextStyle(color: Colors.white, fontWeight: FontWeight.bold, fontSize: 10),
                );
              },
            ),
          ),
          gridData: const FlGridData(show: false),
          titlesData: FlTitlesData(
            leftTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
            rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
            topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
            bottomTitles: AxisTitles(
              sideTitles: SideTitles(
                showTitles: true,
                getTitlesWidget: (value, meta) {
                  final title = titles[value.toInt()] ?? '';
                  return Padding(
                    padding: const EdgeInsets.only(top: 8.0),
                    child: Text(title, style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w600)),
                  );
                },
              ),
            ),
          ),
          borderData: FlBorderData(show: false),
          barGroups: barGroups,
        ),
      ),
    );
  }


  Widget _buildBranchCard(Map<String, dynamic> branch, NumberFormat formatCurrency) {
    final branchName = branch['branch_name']?.toString() ?? 'Cabang Unknown';
    final today = branch['today_omzet'] as num? ?? 0;
    final yesterday = branch['yesterday_omzet'] as num? ?? 0;
    final month = branch['month_omzet'] as num? ?? 0;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: const BoxDecoration(
                color: Color(0xFFF1F5F9),
                border: Border(bottom: BorderSide(color: Color(0xFFE2E8F0))),
              ),
              child: Row(
                children: [
                  const Icon(Icons.storefront_rounded, color: Color(0xFF64748B), size: 20),
                  const SizedBox(width: 8),
                  Text(
                    branchName,
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w700,
                      color: const Color(0xFF334155),
                      fontSize: 14,
                    ),
                  ),
                ],
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  Expanded(
                    child: _buildMetricCol('Bulan Ini', formatCurrency.format(month), Colors.black87),
                  ),
                  Container(width: 1, height: 40, color: const Color(0xFFE2E8F0)),
                  Expanded(
                    child: _buildMetricCol('Hari Ini', formatCurrency.format(today), const Color(0xFF059669)),
                  ),
                  Container(width: 1, height: 40, color: const Color(0xFFE2E8F0)),
                  Expanded(
                    child: _buildMetricCol('Kemarin', formatCurrency.format(yesterday), const Color(0xFF64748B)),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMetricCol(String label, String value, Color valueColor) {
    return Column(
      children: [
        Text(
          label,
          style: GoogleFonts.outfit(
            color: const Color(0xFF94A3B8),
            fontSize: 10,
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          value,
          style: GoogleFonts.outfit(
            color: valueColor,
            fontSize: 13,
            fontWeight: FontWeight.bold,
          ),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
      ],
    );
  }
}
