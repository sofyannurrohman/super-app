import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/kpi/presentation/bloc/kpi_bloc.dart';
import 'package:hris_app/features/kpi/data/models/kpi_model.dart';
import 'package:intl/intl.dart';

class KPIPage extends StatefulWidget {
  const KPIPage({super.key});

  @override
  State<KPIPage> createState() => _KPIPageState();
}

class _KPIPageState extends State<KPIPage> {
  String _selectedPeriod = DateFormat('yyyy-MM').format(DateTime.now());

  @override
  void initState() {
    super.initState();
    context.read<KPIBloc>().add(FetchKPIRequested(_selectedPeriod));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text('KPI & PERFORMA',
            style: GoogleFonts.plusJakartaSans(
                fontSize: 16, fontWeight: FontWeight.w800, letterSpacing: 1.2)),
        centerTitle: true,
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded,
              color: AppColors.textPrimary, size: 20),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: BlocBuilder<KPIBloc, KPIState>(
        builder: (context, state) {
          if (state is KPILoading) {
            return const Center(child: CircularProgressIndicator());
          }

          if (state is KPIError) {
            return Center(child: Text(state.message));
          }

          if (state is KPILoaded) {
            final kpi = state.currentKPI;
            return SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildPeriodSelector(),
                  const SizedBox(height: 24),
                  if (kpi.type == KPIType.sales)
                    _buildSalesKPI(kpi)
                  else
                    _buildInternalKPI(kpi),
                  const SizedBox(height: 24),
                  _buildApprovalStatus(kpi),
                  const SizedBox(height: 32),
                  _buildHistoryList(state.history),
                ],
              ),
            );
          }

          return const SizedBox();
        },
      ),
    );
  }

  Widget _buildPeriodSelector() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppColors.grayBorder),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text('Periode',
              style: GoogleFonts.plusJakartaSans(
                  fontWeight: FontWeight.bold, color: AppColors.textSecondary)),
          DropdownButton<String>(
            value: _selectedPeriod,
            underline: const SizedBox(),
            items: [
              DateFormat('yyyy-MM').format(DateTime.now()),
              DateFormat('yyyy-MM')
                  .format(DateTime.now().subtract(const Duration(days: 30))),
              DateFormat('yyyy-MM')
                  .format(DateTime.now().subtract(const Duration(days: 60))),
            ].map((String value) {
              return DropdownMenuItem<String>(
                value: value,
                child: Text(value,
                    style: GoogleFonts.plusJakartaSans(
                        fontWeight: FontWeight.w800)),
              );
            }).toList(),
            onChanged: (val) {
              if (val != null) {
                setState(() => _selectedPeriod = val);
                context.read<KPIBloc>().add(FetchKPIRequested(val));
              }
            },
          ),
        ],
      ),
    );
  }

  Widget _buildSalesKPI(KPIModel kpi) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 20,
              offset: const Offset(0, 10))
        ],
      ),
      child: Column(
        children: [
          Text('PENCAPAIAN TARGET',
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 12,
                  fontWeight: FontWeight.w800,
                  color: AppColors.textTertiary,
                  letterSpacing: 1.5)),
          const SizedBox(height: 24),
          Stack(
            alignment: Alignment.center,
            children: [
              SizedBox(
                width: 180,
                height: 180,
                child: CircularProgressIndicator(
                  value: kpi.achievementPercentage / 100,
                  strokeWidth: 15,
                  backgroundColor: AppColors.grayLight,
                  color: kpi.achievementPercentage >= 100
                      ? Colors.green
                      : AppColors.primaryRed,
                ),
              ),
              Column(
                children: [
                  Text('${kpi.achievementPercentage.toInt()}%',
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 42,
                          fontWeight: FontWeight.w900,
                          color: AppColors.textPrimary)),
                  Text('Achieved',
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textSecondary)),
                ],
              ),
            ],
          ),
          const SizedBox(height: 32),
          _buildInfoRow('Estimasi Bonus',
              'Rp ${NumberFormat.decimalPattern().format(kpi.achievementPercentage * 10000)}',
              isHighlight: true),
          const Divider(height: 32),
          ...kpi.items.map((item) => _buildDetailItem(
              item.label, '${item.score}%', item.weight ?? '')),
        ],
      ),
    );
  }

  Widget _buildInternalKPI(KPIModel kpi) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.04),
              blurRadius: 20,
              offset: const Offset(0, 10))
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('PERFORMA KERJA',
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textTertiary,
                      letterSpacing: 1.5)),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                decoration: BoxDecoration(
                    color: AppColors.primaryRed.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8)),
                child: Text('Score: ${kpi.finalScore}',
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 12,
                        fontWeight: FontWeight.w800,
                        color: AppColors.primaryRed)),
              ),
            ],
          ),
          const SizedBox(height: 24),
          ...kpi.items.map((item) =>
              _buildScoreItem(item.label, item.score, item.weight ?? '')),
        ],
      ),
    );
  }

  Widget _buildScoreItem(String label, double score, String weight) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(label,
                  style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary)),
              Text('$score/100',
                  style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.w900,
                      color: AppColors.textPrimary)),
            ],
          ),
          const SizedBox(height: 8),
          LinearProgressIndicator(
            value: score / 100,
            backgroundColor: AppColors.grayLight,
            color: AppColors.primaryRed,
            minHeight: 8,
            borderRadius: BorderRadius.circular(4),
          ),
          if (weight.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 4),
              child: Text('Weight: $weight',
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 10, color: AppColors.textTertiary)),
            ),
        ],
      ),
    );
  }

  Widget _buildInfoRow(String label, String value, {bool isHighlight = false}) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label,
            style: GoogleFonts.plusJakartaSans(
                fontWeight: FontWeight.bold, color: AppColors.textSecondary)),
        Text(value,
            style: GoogleFonts.plusJakartaSans(
                fontSize: 16,
                fontWeight: FontWeight.w900,
                color: isHighlight
                    ? AppColors.primaryRed
                    : AppColors.textPrimary)),
      ],
    );
  }

  Widget _buildDetailItem(String label, String value, String weight) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label,
                    style: GoogleFonts.plusJakartaSans(
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary)),
                Text('Weight: $weight',
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 10, color: AppColors.textTertiary)),
              ],
            ),
          ),
          Text(value,
              style: GoogleFonts.plusJakartaSans(
                  fontWeight: FontWeight.w900, color: AppColors.textPrimary)),
        ],
      ),
    );
  }

  Widget _buildApprovalStatus(KPIModel kpi) {
    final bool isApproved = kpi.status == KPIStatus.approved;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isApproved
            ? Colors.green.withOpacity(0.05)
            : Colors.orange.withOpacity(0.05),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(
            color: isApproved
                ? Colors.green.withOpacity(0.2)
                : Colors.orange.withOpacity(0.2)),
      ),
      child: Column(
        children: [
          Row(
            children: [
              Icon(
                  isApproved
                      ? Icons.verified_rounded
                      : Icons.pending_actions_rounded,
                  color: isApproved ? Colors.green : Colors.orange),
              const SizedBox(width: 12),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(isApproved ? 'KPI Telah Disetujui' : 'Menunggu Validasi',
                      style: GoogleFonts.plusJakartaSans(
                          fontWeight: FontWeight.w800,
                          color: isApproved
                              ? Colors.green[800]
                              : Colors.orange[800])),
                  Text(
                      isApproved
                          ? 'Divalidasi oleh: ${kpi.approvedBy}'
                          : 'Data akan divalidasi oleh Manager',
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 11,
                          color: isApproved
                              ? Colors.green[700]
                              : Colors.orange[700])),
                ],
              ),
            ],
          ),
          if (isApproved)
            Padding(
              padding: const EdgeInsets.only(top: 8, left: 36),
              child: Row(
                children: [
                  const Icon(Icons.calendar_today,
                      size: 12, color: Colors.grey),
                  const SizedBox(width: 4),
                  Text('Disetujui pada: ${kpi.approvedAt}',
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 10, color: Colors.grey)),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildHistoryList(List<KPIModel> history) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('RIWAYAT KPI',
            style: GoogleFonts.plusJakartaSans(
                fontSize: 12,
                fontWeight: FontWeight.w800,
                color: AppColors.textTertiary,
                letterSpacing: 1.5)),
        const SizedBox(height: 16),
        ...history.map((h) => Container(
              margin: const EdgeInsets.only(bottom: 12),
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: AppColors.grayBorder)),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(h.period,
                          style: GoogleFonts.plusJakartaSans(
                              fontWeight: FontWeight.w900,
                              color: AppColors.textPrimary)),
                      Text(
                          h.type == KPIType.sales
                              ? 'Sales Achievement'
                              : 'Internal Evaluation',
                          style: GoogleFonts.plusJakartaSans(
                              fontSize: 11, color: AppColors.textTertiary)),
                    ],
                  ),
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.end,
                    children: [
                      Text(
                          h.type == KPIType.sales
                              ? '${h.achievementPercentage}%'
                              : '${h.finalScore}',
                          style: GoogleFonts.plusJakartaSans(
                              fontWeight: FontWeight.w900,
                              color: AppColors.primaryRed)),
                      Text('Score Akhir',
                          style: GoogleFonts.plusJakartaSans(
                              fontSize: 10, color: AppColors.textTertiary)),
                    ],
                  ),
                ],
              ),
            )),
      ],
    );
  }
}
