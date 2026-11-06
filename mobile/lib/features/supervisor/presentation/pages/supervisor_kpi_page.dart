import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/supervisor/presentation/bloc/supervisor_bloc.dart';
import 'package:intl/intl.dart';

class SupervisorKpiPage extends StatefulWidget {
  final String branchId;
  const SupervisorKpiPage({Key? key, required this.branchId}) : super(key: key);

  @override
  State<SupervisorKpiPage> createState() => _SupervisorKpiPageState();
}

class _SupervisorKpiPageState extends State<SupervisorKpiPage> {
  @override
  void initState() {
    super.initState();
    final now = DateTime.now();
    context
        .read<SupervisorBloc>()
        .add(LoadSalesKPIs(now.month, now.year, widget.branchId));
  }

  @override
  Widget build(BuildContext context) {
    final currencyFormat =
        NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);

    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text('Performa KPI Sales',
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w700,
                color: AppColors.surface,
                fontSize: 18)),
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
      body: BlocConsumer<SupervisorBloc, SupervisorState>(
        listener: (context, state) {
          if (state.errorMessage != null) {
            ScaffoldMessenger.of(context).showSnackBar(SnackBar(
              content: Text(state.errorMessage!, style: GoogleFonts.outfit()),
              backgroundColor: AppColors.error,
              behavior: SnackBarBehavior.floating,
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10)),
            ));
          }
        },
        builder: (context, state) {
          if (state.isLoadingKPIs) {
            return const Center(
                child:
                    CircularProgressIndicator(color: AppColors.primaryGreen));
          }

          return Column(
            children: [
              if (widget.branchId == 'ALL_BRANCHES' || widget.branchId.isEmpty)
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                  color: AppColors.warning.withOpacity(0.15),
                  width: double.infinity,
                  child: Row(
                    children: [
                      const Icon(Icons.info_outline_rounded,
                          color: AppColors.warning, size: 20),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          'Perhatian: Anda sedang melihat data seluruh cabang (Pusat). Pastikan profil Anda memiliki pengaturan cabang yang benar.',
                          style: GoogleFonts.outfit(
                              color: AppColors.warning,
                              fontSize: 13,
                              fontWeight: FontWeight.w600),
                        ),
                      ),
                    ],
                  ),
                ),
              if (state.salesKPIs.isEmpty)
                Expanded(
                  child: Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.analytics_outlined,
                            size: 64,
                            color: AppColors.textTertiary.withOpacity(0.5)),
                        const SizedBox(height: 16),
                        Text(
                          'Tidak ada data performa bulan ini',
                          style: GoogleFonts.outfit(
                              fontSize: 16,
                              color: AppColors.textSecondary,
                              fontWeight: FontWeight.w500),
                        ),
                      ],
                    ),
                  ),
                )
              else
                Expanded(
                  child: ListView.builder(
                    physics: const BouncingScrollPhysics(),
                    padding: const EdgeInsets.symmetric(
                        horizontal: 16, vertical: 20),
                    itemCount: state.salesKPIs.length,
                    itemBuilder: (context, index) {
                      final kpi = state.salesKPIs[index];
                      final percent = kpi.targetOmzet > 0
                          ? (kpi.achievedOmzet / kpi.targetOmzet) * 100
                          : 0.0;
                      final bool isTargetAchieved = percent >= 100;

                      final jobTitleDisplay =
                          kpi.jobTitle.isNotEmpty ? kpi.jobTitle : 'Sales';

                      return Container(
                        margin: const EdgeInsets.only(bottom: 16),
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          color: AppColors.surface,
                          borderRadius: BorderRadius.circular(20),
                          border:
                              Border.all(color: AppColors.grayBorder, width: 1),
                          boxShadow: [
                            BoxShadow(
                                color: AppColors.textTertiary.withOpacity(0.05),
                                blurRadius: 10,
                                offset: const Offset(0, 4)),
                          ],
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.center,
                              children: [
                                Container(
                                  padding: const EdgeInsets.all(10),
                                  decoration: BoxDecoration(
                                    color: AppColors.info.withOpacity(0.1),
                                    shape: BoxShape.circle,
                                  ),
                                  child: const Icon(Icons.person_rounded,
                                      color: AppColors.info, size: 24),
                                ),
                                const SizedBox(width: 16),
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Text(kpi.employeeName,
                                          style: GoogleFonts.outfit(
                                              fontWeight: FontWeight.w700,
                                              fontSize: 16,
                                              color: AppColors.textPrimary)),
                                      const SizedBox(height: 2),
                                      Container(
                                        padding: const EdgeInsets.symmetric(
                                            horizontal: 8, vertical: 2),
                                        decoration: BoxDecoration(
                                          color: AppColors.grayLight,
                                          borderRadius:
                                              BorderRadius.circular(8),
                                        ),
                                        child: Text(jobTitleDisplay,
                                            style: GoogleFonts.outfit(
                                                fontWeight: FontWeight.w600,
                                                fontSize: 11,
                                                color:
                                                    AppColors.textSecondary)),
                                      ),
                                    ],
                                  ),
                                ),
                                if (isTargetAchieved)
                                  Container(
                                    padding: const EdgeInsets.all(6),
                                    decoration: BoxDecoration(
                                        color:
                                            AppColors.success.withOpacity(0.1),
                                        shape: BoxShape.circle),
                                    child: const Icon(
                                        Icons.emoji_events_rounded,
                                        color: AppColors.success,
                                        size: 20),
                                  )
                              ],
                            ),
                            const Padding(
                              padding: EdgeInsets.symmetric(vertical: 16),
                              child: Divider(
                                  color: AppColors.grayBorder, height: 1),
                            ),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text('Progress Omzet',
                                    style: GoogleFonts.outfit(
                                        color: AppColors.textSecondary,
                                        fontSize: 13,
                                        fontWeight: FontWeight.w500)),
                                Text('${percent.toStringAsFixed(1)}%',
                                    style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w800,
                                        fontSize: 14,
                                        color: isTargetAchieved
                                            ? AppColors.success
                                            : AppColors.primaryBlue)),
                              ],
                            ),
                            const SizedBox(height: 12),
                            ClipRRect(
                              borderRadius: BorderRadius.circular(8),
                              child: LinearProgressIndicator(
                                value: (percent / 100).clamp(0.0, 1.0),
                                minHeight: 8,
                                backgroundColor: AppColors.grayLight,
                                valueColor: AlwaysStoppedAnimation<Color>(
                                    isTargetAchieved
                                        ? AppColors.success
                                        : AppColors.primaryBlue),
                              ),
                            ),
                            const SizedBox(height: 16),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text('Tercapai',
                                        style: GoogleFonts.outfit(
                                            fontSize: 12,
                                            color: AppColors.textSecondary)),
                                    const SizedBox(height: 2),
                                    Text(
                                        currencyFormat
                                            .format(kpi.achievedOmzet),
                                        style: GoogleFonts.outfit(
                                            fontWeight: FontWeight.w700,
                                            fontSize: 15,
                                            color: AppColors.textPrimary)),
                                  ],
                                ),
                                Column(
                                  crossAxisAlignment: CrossAxisAlignment.end,
                                  children: [
                                    Text('Target',
                                        style: GoogleFonts.outfit(
                                            fontSize: 12,
                                            color: AppColors.textSecondary)),
                                    const SizedBox(height: 2),
                                    Text(currencyFormat.format(kpi.targetOmzet),
                                        style: GoogleFonts.outfit(
                                            fontWeight: FontWeight.w600,
                                            fontSize: 15,
                                            color: AppColors.textSecondary)),
                                    const SizedBox(height: 2),
                                    Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        const Icon(Icons.storefront_rounded,
                                            size: 12,
                                            color: AppColors.textTertiary),
                                        const SizedBox(width: 4),
                                        Text(kpi.branchName,
                                            style: GoogleFonts.outfit(
                                                fontSize: 11,
                                                color: AppColors.textTertiary)),
                                      ],
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ],
                        ),
                      );
                    },
                  ),
                ),
            ],
          );
        },
      ),
    );
  }
}
