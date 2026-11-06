import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/reimbursement/presentation/bloc/reimbursement_bloc.dart';
import 'package:hris_app/features/reimbursement/presentation/bloc/reimbursement_event.dart';
import 'package:hris_app/features/reimbursement/presentation/bloc/reimbursement_state.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';
import 'add_reimbursement_page.dart';
import "package:hris_app/core/utils/dialog_utils.dart";
import 'package:hris_app/core/utils/snackbar_utils.dart';

class ReimbursementListPage extends StatefulWidget {
  const ReimbursementListPage({super.key});

  @override
  State<ReimbursementListPage> createState() => _ReimbursementListPageState();
}

class _ReimbursementListPageState extends State<ReimbursementListPage> {
  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => di.sl<ReimbursementBloc>()
        ..add(const FetchReimbursementHistoryRequested()),
      child: Scaffold(
        backgroundColor: AppColors.backgroundAlt,
        body: BlocListener<ReimbursementBloc, ReimbursementState>(
          listener: (context, state) {
            if (state is ReimbursementActionSuccess) {
              SnackBarUtils.showSuccess(context, state.message);
              context
                  .read<ReimbursementBloc>()
                  .add(const FetchReimbursementHistoryRequested());
            } else if (state is ReimbursementFailure) {
              DialogUtils.showError(
                  context: context, title: "Gagal", message: state.message);
            }
          },
          child: const ReimbursementListView(),
        ),
        floatingActionButton: Builder(
          builder: (context) => FloatingActionButton.extended(
            onPressed: () async {
              final result = await Navigator.push(
                context,
                MaterialPageRoute(builder: (_) => const AddReimbursementPage()),
              );
              if (result == true && context.mounted) {
                context
                    .read<ReimbursementBloc>()
                    .add(const FetchReimbursementHistoryRequested());
              }
            },
            backgroundColor: AppColors.primaryRed,
            icon: const Icon(Icons.add, color: Colors.white),
            label: Text('PENGAJUAN BARU',
                style: GoogleFonts.plusJakartaSans(
                    fontWeight: FontWeight.w800,
                    color: Colors.white,
                    letterSpacing: 1,
                    fontSize: 13)),
            shape:
                RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          ),
        ),
      ),
    );
  }
}

class ReimbursementListView extends StatelessWidget {
  const ReimbursementListView({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<ReimbursementBloc, ReimbursementState>(
      builder: (context, state) {
        return NestedScrollView(
          headerSliverBuilder: (context, innerBoxIsScrolled) {
            return [
              _buildSliverAppBar(context, state),
            ];
          },
          body: _buildBody(context, state),
        );
      },
    );
  }

  Widget _buildSliverAppBar(BuildContext context, ReimbursementState state) {
    double totalApproved = 0;
    if (state is ReimbursementHistoryLoaded) {
      totalApproved = state.history
          .where((r) => r.status == 'APPROVED')
          .fold(0.0, (sum, r) => sum + r.amount);
    }

    final currencyFormat =
        NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);

    return SliverAppBar(
      expandedHeight: 200,
      pinned: true,
      backgroundColor: AppColors.primaryRed,
      elevation: 0,
      flexibleSpace: FlexibleSpaceBar(
        background: Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: AppColors.primaryGradient,
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const SizedBox(height: 40),
              Text(
                currencyFormat.format(totalApproved),
                style: GoogleFonts.plusJakartaSans(
                  color: Colors.white,
                  fontSize: 32,
                  fontWeight: FontWeight.w900,
                ),
              ),
              Text(
                'TOTAL REIMBURSEMENT DISETUJUI',
                style: GoogleFonts.plusJakartaSans(
                  color: Colors.white.withOpacity(0.8),
                  fontSize: 10,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1.5,
                ),
              ),
            ],
          ),
        ),
      ),
      title: Text('REIMBURSEMENT',
          style: GoogleFonts.plusJakartaSans(
              color: Colors.white,
              fontWeight: FontWeight.w800,
              fontSize: 16,
              letterSpacing: 1)),
      centerTitle: true,
      leading: IconButton(
        icon:
            const Icon(Icons.arrow_back_ios_new, color: Colors.white, size: 20),
        onPressed: () => Navigator.of(context).pop(),
      ),
    );
  }

  Widget _buildBody(BuildContext context, ReimbursementState state) {
    if (state is ReimbursementLoading) {
      return const Center(
          child: CircularProgressIndicator(color: AppColors.primaryRed));
    } else if (state is ReimbursementFailure) {
      return Center(child: Text(state.message));
    } else if (state is ReimbursementHistoryLoaded) {
      if (state.history.isEmpty) {
        return _buildEmptyState();
      }
      return RefreshIndicator(
        onRefresh: () async {
          context
              .read<ReimbursementBloc>()
              .add(const FetchReimbursementHistoryRequested());
        },
        child: ListView.builder(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
          itemCount: state.history.length,
          itemBuilder: (context, index) {
            final r = state.history[index];
            return _buildReimbursementCard(context, r);
          },
        ),
      );
    }
    return const SizedBox();
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.receipt_long_rounded,
              size: 64, color: AppColors.textTertiary.withOpacity(0.3)),
          const SizedBox(height: 16),
          Text('Belum ada riwayat reimbursement',
              style: GoogleFonts.plusJakartaSans(
                  fontWeight: FontWeight.w800, color: AppColors.textTertiary)),
        ],
      ),
    );
  }

  Widget _buildReimbursementCard(BuildContext context, dynamic r) {
    final currencyFormat =
        NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);

    Color statusColor;
    IconData statusIcon;
    bool canEditDelete = false;

    switch (r.status.toUpperCase()) {
      case 'APPROVED':
        statusColor = AppColors.success;
        statusIcon = Icons.check_circle_rounded;
        break;
      case 'REJECTED':
        statusColor = AppColors.error;
        statusIcon = Icons.cancel_rounded;
        break;
      case 'PENDING':
      default:
        statusColor = AppColors.warning;
        statusIcon = Icons.pending_rounded;
        canEditDelete = true;
        break;
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(20),
        child: Container(
          decoration: BoxDecoration(
            border: Border(left: BorderSide(color: statusColor, width: 6)),
          ),
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: Text(
                        r.title.toUpperCase(),
                        style: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w900,
                            fontSize: 14,
                            color: AppColors.textPrimary,
                            letterSpacing: 0.5),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                    if (canEditDelete)
                      _buildActionMenu(context, r)
                    else
                      Icon(statusIcon, color: statusColor, size: 20),
                  ],
                ),
                if (r.category == 'BBM' || r.startKm != null) ...[
                  const SizedBox(height: 6),
                  Container(
                    padding:
                        const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: Colors.amber.shade50,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.amber.shade200),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.local_gas_station_rounded,
                            size: 14, color: Colors.amber.shade800),
                        const SizedBox(width: 4),
                        Text(
                          'KM: ${r.startKm?.toInt() ?? '-'} → ${r.endKm?.toInt() ?? '-'} (${r.totalKm?.toInt() ?? ((r.endKm != null && r.startKm != null) ? (r.endKm! - r.startKm!).toInt() : '-')} KM)',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 10,
                            fontWeight: FontWeight.w700,
                            color: Colors.amber.shade900,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
                const SizedBox(height: 12),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          currencyFormat.format(r.amount),
                          style: GoogleFonts.plusJakartaSans(
                              fontWeight: FontWeight.w900,
                              fontSize: 18,
                              color: AppColors.primaryRed),
                        ),
                        const SizedBox(height: 4),
                        Row(
                          children: [
                            const Icon(Icons.calendar_today_rounded,
                                size: 12, color: AppColors.textTertiary),
                            const SizedBox(width: 4),
                            Text(
                              DateFormat('dd MMM yyyy').format(r.createdAt),
                              style: GoogleFonts.plusJakartaSans(
                                  fontSize: 11,
                                  fontWeight: FontWeight.w700,
                                  color: AppColors.textTertiary),
                            ),
                          ],
                        ),
                      ],
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 10, vertical: 6),
                      decoration: BoxDecoration(
                        color: statusColor.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Text(
                        r.status.toUpperCase(),
                        style: GoogleFonts.plusJakartaSans(
                            color: statusColor,
                            fontSize: 10,
                            fontWeight: FontWeight.w900,
                            letterSpacing: 0.5),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildActionMenu(BuildContext context, dynamic r) {
    return PopupMenuButton<String>(
      padding: EdgeInsets.zero,
      icon: const Icon(Icons.more_vert_rounded, color: AppColors.textTertiary),
      onSelected: (value) async {
        if (value == 'edit') {
          final result = await Navigator.push(
            context,
            MaterialPageRoute(
                builder: (_) => AddReimbursementPage(reimbursementToEdit: r)),
          );
          if (result == true && context.mounted) {
            context
                .read<ReimbursementBloc>()
                .add(const FetchReimbursementHistoryRequested());
          }
        } else if (value == 'delete') {
          _showDeleteConfirmation(context, r);
        }
      },
      itemBuilder: (context) => [
        PopupMenuItem(
          value: 'edit',
          child: Row(children: [
            const Icon(Icons.edit_rounded, size: 18),
            const SizedBox(width: 12),
            Text('Ubah',
                style: GoogleFonts.plusJakartaSans(
                    fontWeight: FontWeight.w600, fontSize: 13))
          ]),
        ),
        PopupMenuItem(
          value: 'delete',
          child: Row(children: [
            const Icon(Icons.delete_rounded, size: 18, color: AppColors.error),
            const SizedBox(width: 12),
            Text('Hapus',
                style: GoogleFonts.plusJakartaSans(
                    fontWeight: FontWeight.w600,
                    fontSize: 13,
                    color: AppColors.error))
          ]),
        ),
      ],
    );
  }

  void _showDeleteConfirmation(BuildContext context, dynamic r) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Text('Hapus Pengajuan',
            style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w800)),
        content: Text('Apakah Anda yakin ingin menghapus pengajuan ini?',
            style: GoogleFonts.plusJakartaSans()),
        actions: [
          TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: Text('BATAL',
                  style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.w800,
                      color: AppColors.textTertiary))),
          TextButton(
              onPressed: () => Navigator.pop(context, true),
              child: Text('HAPUS',
                  style: GoogleFonts.plusJakartaSans(
                      fontWeight: FontWeight.w800, color: AppColors.error))),
        ],
      ),
    );
    if (confirm == true && context.mounted) {
      context.read<ReimbursementBloc>().add(DeleteReimbursementRequested(r.id));
    }
  }
}
