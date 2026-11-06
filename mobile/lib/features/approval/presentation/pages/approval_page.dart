import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:hris_app/features/approval/presentation/bloc/approval_bloc.dart';
import 'package:hris_app/features/leave/domain/entities/leave.dart';
import 'package:hris_app/features/reimbursement/domain/entities/reimbursement.dart';
import 'package:intl/intl.dart';

class ApprovalPage extends StatefulWidget {
  const ApprovalPage({super.key});

  @override
  State<ApprovalPage> createState() => _ApprovalPageState();
}

class _ApprovalPageState extends State<ApprovalPage>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    context.read<ApprovalBloc>().add(FetchPendingApprovalsRequested());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text('APPROVAL PANEL',
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
        bottom: TabBar(
          controller: _tabController,
          labelColor: AppColors.primaryRed,
          unselectedLabelColor: AppColors.textTertiary,
          indicatorColor: AppColors.primaryRed,
          labelStyle: GoogleFonts.plusJakartaSans(
              fontWeight: FontWeight.w800, fontSize: 13),
          tabs: const [
            Tab(text: 'CUTI & IZIN'),
            Tab(text: 'REIMBURSE'),
          ],
        ),
      ),
      body: BlocListener<ApprovalBloc, ApprovalState>(
        listener: (context, state) {
          if (state is ApprovalActionSuccess) {
            SnackBarUtils.showSuccess(context, state.message);
          } else if (state is ApprovalFailure) {
            SnackBarUtils.showError(context, state.message);
          }
        },
        child: BlocBuilder<ApprovalBloc, ApprovalState>(
          builder: (context, state) {
            if (state is ApprovalLoading) {
              return const Center(child: CircularProgressIndicator());
            }

            if (state is ApprovalDataLoaded) {
              return TabBarView(
                controller: _tabController,
                children: [
                  _buildLeaveList(state.pendingLeaves),
                  _buildReimbursementList(state.pendingReimbursements),
                ],
              );
            }

            return const SizedBox();
          },
        ),
      ),
    );
  }

  Widget _buildLeaveList(List<Leave> leaves) {
    if (leaves.isEmpty)
      return _buildEmptyState('Tidak ada pengajuan cuti pending.');

    return ListView.builder(
      padding: const EdgeInsets.all(20),
      itemCount: leaves.length,
      itemBuilder: (context, index) => _buildLeaveCard(leaves[index]),
    );
  }

  Widget _buildReimbursementList(List<Reimbursement> reimbursements) {
    if (reimbursements.isEmpty)
      return _buildEmptyState('Tidak ada pengajuan reimburse pending.');

    return ListView.builder(
      padding: const EdgeInsets.all(20),
      itemCount: reimbursements.length,
      itemBuilder: (context, index) =>
          _buildReimbursementCard(reimbursements[index]),
    );
  }

  Widget _buildLeaveCard(Leave leave) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 10)
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text((leave.leaveTypeName ?? 'PENGAJUAN').toUpperCase(),
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 11,
                      fontWeight: FontWeight.w800,
                      color: AppColors.primaryRed)),
              Text(
                  DateFormat('dd MMM yyyy')
                      .format(leave.createdAt ?? DateTime.now()),
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 11, color: AppColors.textTertiary)),
            ],
          ),
          const SizedBox(height: 12),
          Text(leave.employeeName ?? 'Karyawan',
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 16, fontWeight: FontWeight.w800)),
          const SizedBox(height: 8),
          Text(leave.reason,
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 13, color: AppColors.textSecondary)),
          const SizedBox(height: 16),
          _buildDateInfo(leave.startDate, leave.endDate),
          const Divider(height: 32),
          _buildActionButtons(
            onApprove: () => context
                .read<ApprovalBloc>()
                .add(ApproveLeaveRequested(leave.id, 'approved')),
            onReject: () => _showRejectLeaveDialog(context, leave.id),
          ),
        ],
      ),
    );
  }

  void _showRejectLeaveDialog(BuildContext context, String leaveId) {
    final reasonController = TextEditingController();
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: const Text('Tolak Pengajuan Cuti',
            style: TextStyle(fontWeight: FontWeight.bold, color: AppColors.error)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Masukkan alasan penolakan untuk karyawan (wajib):',
              style: TextStyle(fontSize: 13, color: AppColors.textSecondary),
            ),
            const SizedBox(height: 12),
            TextField(
              controller: reasonController,
              maxLines: 3,
              decoration: InputDecoration(
                hintText: 'Contoh: Kuota tim pada tanggal tersebut sudah penuh...',
                hintStyle: TextStyle(
                    fontSize: 12,
                    color: AppColors.textTertiary.withOpacity(0.6)),
                border:
                    OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                contentPadding: const EdgeInsets.all(12),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              final reason = reasonController.text.trim();
              if (reason.isEmpty) {
                SnackBarUtils.showError(
                    dialogContext, 'Alasan penolakan wajib diisi.');
                return;
              }
              Navigator.pop(dialogContext);
              context.read<ApprovalBloc>().add(
                  ApproveLeaveRequested(leaveId, 'rejected', rejectReason: reason));
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.error,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10)),
            ),
            child: const Text('Tolak'),
          ),
        ],
      ),
    );
  }

  Widget _buildReimbursementCard(Reimbursement r) {
    final currencyFormat =
        NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 10)
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('REIMBURSEMENT',
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 11,
                      fontWeight: FontWeight.w800,
                      color: Colors.blue)),
              Text(DateFormat('dd MMM yyyy').format(r.createdAt),
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 11, color: AppColors.textTertiary)),
            ],
          ),
          const SizedBox(height: 12),
          Text(r.employeeName ?? 'Karyawan',
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 16, fontWeight: FontWeight.w800)),
          const SizedBox(height: 4),
          Text(r.title,
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary)),
          const SizedBox(height: 4),
          Text(r.description ?? '',
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 12, color: AppColors.textSecondary)),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
                color: AppColors.grayLight,
                borderRadius: BorderRadius.circular(12)),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('Total Nominal:',
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 12, fontWeight: FontWeight.bold)),
                Text(currencyFormat.format(r.amount),
                    style: GoogleFonts.plusJakartaSans(
                        fontSize: 14,
                        fontWeight: FontWeight.w900,
                        color: AppColors.primaryRed)),
              ],
            ),
          ),
          const Divider(height: 32),
          _buildActionButtons(
            onApprove: () => context
                .read<ApprovalBloc>()
                .add(ApproveReimbursementRequested(r.id, 'approved')),
            onReject: () => context
                .read<ApprovalBloc>()
                .add(ApproveReimbursementRequested(r.id, 'rejected')),
          ),
        ],
      ),
    );
  }

  Widget _buildDateInfo(DateTime start, DateTime end) {
    return Row(
      children: [
        const Icon(Icons.calendar_today_rounded,
            size: 14, color: AppColors.textTertiary),
        const SizedBox(width: 8),
        Text(
            '${DateFormat('dd MMM').format(start)} - ${DateFormat('dd MMM yyyy').format(end)}',
            style: GoogleFonts.plusJakartaSans(
                fontSize: 12,
                fontWeight: FontWeight.w700,
                color: AppColors.textSecondary)),
      ],
    );
  }

  Widget _buildActionButtons(
      {required VoidCallback onApprove, required VoidCallback onReject}) {
    return Row(
      children: [
        Expanded(
          child: OutlinedButton(
            onPressed: onReject,
            style: OutlinedButton.styleFrom(
              foregroundColor: AppColors.error,
              side: const BorderSide(color: AppColors.error),
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12)),
            ),
            child: const Text('REJECT'),
          ),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: ElevatedButton(
            onPressed: onApprove,
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.green,
              foregroundColor: Colors.white,
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12)),
            ),
            child: const Text('APPROVE'),
          ),
        ),
      ],
    );
  }

  Widget _buildEmptyState(String message) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.assignment_turned_in_rounded,
              size: 64, color: AppColors.grayBorder),
          const SizedBox(height: 16),
          Text(message,
              style: GoogleFonts.plusJakartaSans(
                  color: AppColors.textTertiary, fontWeight: FontWeight.w700)),
        ],
      ),
    );
  }
}
