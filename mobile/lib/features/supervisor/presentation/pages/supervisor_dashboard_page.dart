import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_state.dart';
import 'package:hris_app/features/supervisor/presentation/pages/supervisor_transfer_page.dart';
import 'package:hris_app/features/supervisor/presentation/pages/supervisor_delivery_batch_page.dart';
import 'package:hris_app/features/supervisor/presentation/pages/supervisor_kpi_page.dart';
import 'package:hris_app/features/sales/presentation/pages/select_store_page.dart';
import 'package:hris_app/features/sales/presentation/pages/store_list_page.dart';
import 'package:hris_app/features/supervisor/presentation/bloc/supervisor_bloc.dart';
import 'package:intl/intl.dart';
import 'package:shared_preferences/shared_preferences.dart';

class SupervisorDashboardPage extends StatefulWidget {
  final String branchId;
  final String branchName;
  const SupervisorDashboardPage(
      {Key? key, required this.branchId, this.branchName = ''})
      : super(key: key);

  @override
  State<SupervisorDashboardPage> createState() =>
      _SupervisorDashboardPageState();
}

class _SupervisorDashboardPageState extends State<SupervisorDashboardPage> {
  @override
  void initState() {
    super.initState();
    final now = DateTime.now();
    context
        .read<SupervisorBloc>()
        .add(LoadSalesKPIs(now.month, now.year, widget.branchId));
    context.read<SupervisorBloc>().add(LoadPendingTransfers(widget.branchId));
    context
        .read<SupervisorBloc>()
        .add(LoadPendingDeliveryBatches(widget.branchId));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text('Supervisor Dashboard',
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w700, color: AppColors.surface)),
        backgroundColor: AppColors.primaryGreen,
        foregroundColor: AppColors.surface,
        elevation: 0,
        centerTitle: false,
        flexibleSpace: Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: AppColors.primaryGradient,
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
        ),
      ),
      body: SingleChildScrollView(
        physics: const BouncingScrollPhysics(),
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildHeaderTitle(),
            const SizedBox(height: 20),
            _buildSummaryCards(),
            const SizedBox(height: 28),
            _buildQuickAccess(),
            const SizedBox(height: 28),
            _buildSalesFieldSection(),
            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }

  Widget _buildHeaderTitle() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Overview',
          style: GoogleFonts.outfit(
            fontSize: 22,
            fontWeight: FontWeight.w800,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          'Ringkasan operasional dan performa cabang',
          style: GoogleFonts.outfit(
            fontSize: 14,
            fontWeight: FontWeight.w400,
            color: AppColors.textSecondary,
          ),
        ),
      ],
    );
  }

  Widget _buildSummaryCards() {
    return BlocBuilder<SupervisorBloc, SupervisorState>(
      builder: (context, state) {
        if (state.isLoadingKPIs && state.salesKPIs.isEmpty) {
          return const Center(
              child: Padding(
                  padding: EdgeInsets.all(40.0),
                  child: CircularProgressIndicator(
                      color: AppColors.primaryGreen)));
        }
        if (state.errorMessage != null && state.salesKPIs.isEmpty) {
          return Center(
              child: Padding(
            padding: const EdgeInsets.all(20.0),
            child: Text('Error: ${state.errorMessage}',
                style: GoogleFonts.outfit(color: AppColors.error)),
          ));
        }

        final formatter = NumberFormat.currency(
            locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);
        final profileState = context.read<ProfileBloc>().state;
        String profileBranchName = '';
        if (profileState is ProfileLoaded) {
          final p = profileState.profile;
          if (p['branch'] != null && p['branch']['name'] != null) {
            profileBranchName = p['branch']['name'];
          } else if (p['employee'] != null &&
              p['employee']['branch'] != null &&
              p['employee']['branch']['name'] != null) {
            profileBranchName = p['employee']['branch']['name'];
          } else if (p['Branch'] != null && p['Branch']['name'] != null) {
            profileBranchName = p['Branch']['name'];
          } else if (p['branch_name'] != null) {
            profileBranchName = p['branch_name'];
          }
        }

        // Calculate total target & achieved omzet
        double totalTarget = 0;
        double totalAchieved = 0;
        String kpiBranchName = '';
        for (var kpi in state.salesKPIs) {
          totalTarget += kpi.targetOmzet;
          totalAchieved += kpi.achievedOmzet;
          if (kpiBranchName.isEmpty &&
              kpi.branchName.isNotEmpty &&
              kpi.branchName != 'Cabang Utama') {
            kpiBranchName = kpi.branchName;
          }
        }

        final double progress = totalTarget > 0
            ? (totalAchieved / totalTarget).clamp(0.0, 1.0)
            : 0.0;
        final int pendingTransfers = state.pendingTransfers.length;
        final int pendingBatches = state.pendingBatches.length;

        String finalBranchName = kpiBranchName.isNotEmpty
            ? kpiBranchName
            : (widget.branchName.isNotEmpty
                ? widget.branchName
                : profileBranchName);

        final String omzetTitle = finalBranchName.isNotEmpty
            ? 'Pencapaian Omzet $finalBranchName'
            : (widget.branchId == 'ALL_BRANCHES' || widget.branchId.isEmpty
                ? 'Pencapaian Omzet Semua Cabang'
                : 'Pencapaian Omzet');

        return Column(
          children: [
            // Premium Omzet Card
            Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  colors: AppColors.primaryGradient,
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                ),
                borderRadius: BorderRadius.circular(24),
                boxShadow: [
                  BoxShadow(
                      color: AppColors.primaryGreen.withOpacity(0.3),
                      blurRadius: 20,
                      offset: const Offset(0, 8)),
                ],
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                            color: AppColors.surface.withOpacity(0.2),
                            shape: BoxShape.circle),
                        child: const Icon(Icons.show_chart_rounded,
                            color: AppColors.surface, size: 24),
                      ),
                      const SizedBox(width: 16),
                      Expanded(
                        child: Text(
                          omzetTitle,
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w600,
                              fontSize: 16,
                              color: AppColors.surface.withOpacity(0.9)),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text('Tercapai',
                              style: GoogleFonts.outfit(
                                  fontSize: 13,
                                  color: AppColors.surface.withOpacity(0.7))),
                          const SizedBox(height: 4),
                          Text(formatter.format(totalAchieved),
                              style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w800,
                                  fontSize: 22,
                                  color: AppColors.surface)),
                        ],
                      ),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          Text('Target',
                              style: GoogleFonts.outfit(
                                  fontSize: 13,
                                  color: AppColors.surface.withOpacity(0.7))),
                          const SizedBox(height: 4),
                          Text(formatter.format(totalTarget),
                              style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w600,
                                  fontSize: 16,
                                  color: AppColors.surface.withOpacity(0.9))),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: 20),
                  ClipRRect(
                    borderRadius: BorderRadius.circular(12),
                    child: LinearProgressIndicator(
                      value: progress,
                      minHeight: 8,
                      backgroundColor: AppColors.surface.withOpacity(0.2),
                      valueColor: const AlwaysStoppedAnimation<Color>(
                          AppColors.surface),
                    ),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                          '${(progress * 100).toStringAsFixed(1)}% dari target',
                          style: GoogleFonts.outfit(
                              fontSize: 13,
                              color: AppColors.surface.withOpacity(0.9),
                              fontWeight: FontWeight.w600)),
                      if (progress >= 1.0)
                        const Icon(Icons.check_circle_rounded,
                            color: AppColors.surface, size: 16)
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(height: 20),
            // Pending Approval Counters
            Row(
              children: [
                Expanded(
                  child: _buildCounterCard(
                    title: 'Pending\nBarang',
                    count: pendingTransfers,
                    icon: Icons.inventory_2_rounded,
                    color: AppColors.warning,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: _buildCounterCard(
                    title: 'Pending\nSurat Jalan',
                    count: pendingBatches,
                    icon: Icons.local_shipping_rounded,
                    color: AppColors.info,
                  ),
                ),
              ],
            ),
          ],
        );
      },
    );
  }

  Widget _buildCounterCard(
      {required String title,
      required int count,
      required IconData icon,
      required Color color}) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 20, horizontal: 16),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: AppColors.grayBorder, width: 1),
        boxShadow: [
          BoxShadow(
              color: color.withOpacity(0.05),
              blurRadius: 12,
              offset: const Offset(0, 6)),
        ],
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
                color: color.withOpacity(0.1),
                borderRadius: BorderRadius.circular(16)),
            child: Icon(icon, color: color, size: 26),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  count.toString(),
                  style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w800,
                      fontSize: 24,
                      color: AppColors.textPrimary,
                      height: 1.0),
                ),
                const SizedBox(height: 6),
                Text(
                  title,
                  style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w600,
                      fontSize: 13,
                      color: AppColors.textSecondary,
                      height: 1.2),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildQuickAccess() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Akses Cepat',
          style: GoogleFonts.outfit(
              fontSize: 18,
              fontWeight: FontWeight.w800,
              color: AppColors.textPrimary),
        ),
        const SizedBox(height: 16),
        GridView.count(
          crossAxisCount: 3,
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          mainAxisSpacing: 16,
          crossAxisSpacing: 16,
          childAspectRatio: 0.85,
          children: [
            _buildMenuButton(
              title: 'Approval\nBarang',
              icon: Icons.inventory_2_rounded,
              color: AppColors.warning,
              onTap: () => Navigator.of(context).push(MaterialPageRoute(
                builder: (_) => BlocProvider.value(
                  value: context.read<SupervisorBloc>(),
                  child: SupervisorTransferPage(branchId: widget.branchId),
                ),
              )),
            ),
            _buildMenuButton(
              title: 'Approval\nSurat Jalan',
              icon: Icons.local_shipping_rounded,
              color: AppColors.info,
              onTap: () => Navigator.of(context).push(MaterialPageRoute(
                builder: (_) => BlocProvider.value(
                  value: context.read<SupervisorBloc>(),
                  child: SupervisorDeliveryBatchPage(branchId: widget.branchId),
                ),
              )),
            ),
            _buildMenuButton(
              title: 'Performa\nKPI Sales',
              icon: Icons.analytics_rounded,
              color: AppColors.success,
              onTap: () => Navigator.of(context).push(MaterialPageRoute(
                builder: (_) => BlocProvider.value(
                  value: context.read<SupervisorBloc>(),
                  child: SupervisorKpiPage(branchId: widget.branchId),
                ),
              )),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildSalesFieldSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'Operasional Sales Lapangan',
              style: GoogleFonts.outfit(
                  fontSize: 18,
                  fontWeight: FontWeight.w800,
                  color: AppColors.textPrimary),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: AppColors.primaryGreen.withOpacity(0.1),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Text(
                'Toko Cabang',
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.w700,
                  color: AppColors.primaryGreen,
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.blue.shade50.withOpacity(0.7),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.blue.shade200, width: 1),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(Icons.info_outline_rounded, color: Colors.blue.shade800, size: 20),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  'Mode Operasional: Sales TO (Take Order). Kunjungan dan pesanan yang dibuat otomatis berjenis Sales Order (SO) dan omset diatribusikan ke Sales PIC pemilik toko.',
                  style: GoogleFonts.outfit(
                    fontSize: 12,
                    color: Colors.blue.shade900,
                    fontWeight: FontWeight.w500,
                    height: 1.3,
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 16),
        GridView.count(
          crossAxisCount: 3,
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          mainAxisSpacing: 16,
          crossAxisSpacing: 16,
          childAspectRatio: 0.85,
          children: [
            _buildMenuButton(
              title: 'Kunjungan\nSales TO',
              icon: Icons.location_on_rounded,
              color: AppColors.primaryGreen,
              onTap: () async {
                final prefs = await SharedPreferences.getInstance();
                await prefs.setString('sales_mode', 'TO');
                if (mounted) {
                  Navigator.of(context).push(MaterialPageRoute(
                    builder: (_) => const SelectStorePage(),
                  ));
                }
              },
            ),
            _buildMenuButton(
              title: 'Daftar Toko\nCabang',
              icon: Icons.storefront_rounded,
              color: Colors.indigo,
              onTap: () => Navigator.of(context).push(MaterialPageRoute(
                builder: (_) => const StoreListPage(),
              )),
            ),
            _buildMenuButton(
              title: 'Dashboard\nSales Mode',
              icon: Icons.dashboard_rounded,
              color: Colors.teal,
              onTap: () async {
                final prefs = await SharedPreferences.getInstance();
                await prefs.setString('sales_mode', 'TO');
                if (mounted) {
                  Navigator.of(context).pushNamed(
                    '/sales_dashboard',
                    arguments: {'forceTO': true},
                  );
                }
              },
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildMenuButton({
    required String title,
    required IconData icon,
    required Color color,
    required VoidCallback onTap,
  }) {
    return Material(
      color: AppColors.surface,
      borderRadius: BorderRadius.circular(20),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(20),
        splashColor: color.withOpacity(0.1),
        highlightColor: color.withOpacity(0.05),
        child: Container(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(20),
            border: Border.all(color: AppColors.grayBorder, width: 1),
            boxShadow: [
              BoxShadow(
                  color: AppColors.textTertiary.withOpacity(0.05),
                  blurRadius: 10,
                  offset: const Offset(0, 4)),
            ],
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(
                  color: color.withOpacity(0.1),
                  shape: BoxShape.circle,
                ),
                child: Icon(icon, color: color, size: 28),
              ),
              const SizedBox(height: 12),
              Text(
                title,
                textAlign: TextAlign.center,
                style: GoogleFonts.outfit(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textPrimary,
                  height: 1.2,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
