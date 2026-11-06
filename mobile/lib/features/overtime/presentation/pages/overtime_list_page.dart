import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/features/overtime/presentation/bloc/overtime_bloc.dart';
import 'package:hris_app/features/overtime/presentation/bloc/overtime_event.dart';
import 'package:hris_app/features/overtime/presentation/bloc/overtime_state.dart';
import 'package:hris_app/features/overtime/presentation/pages/overtime_request_page.dart';
import 'package:intl/intl.dart';
import "package:hris_app/core/utils/dialog_utils.dart";
import 'package:hris_app/core/utils/snackbar_utils.dart';
import 'package:flutter_staggered_animations/flutter_staggered_animations.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import '../../domain/entities/overtime.dart';

class OvertimeListPage extends StatefulWidget {
  const OvertimeListPage({super.key});

  @override
  State<OvertimeListPage> createState() => _OvertimeListPageState();
}

class _OvertimeListPageState extends State<OvertimeListPage> {
  @override
  void initState() {
    super.initState();
    context.read<OvertimeBloc>().add(FetchMyOvertimesRequested());
  }

  @override
  Widget build(BuildContext context) {
    const primaryRed = AppColors.primaryRed;
    const bgGray = AppColors.backgroundAlt;
    const textColor = AppColors.textPrimary;
    const subtitleColor = AppColors.textSecondary;

    return Scaffold(
      backgroundColor: bgGray,
      appBar: AppBar(
        title: Text(
          'Pengajuan Lembur',
          style: GoogleFonts.plusJakartaSans(
              fontSize: 18, fontWeight: FontWeight.w800, color: textColor),
        ),
        centerTitle: true,
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new,
              color: AppColors.textPrimary, size: 20),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocConsumer<OvertimeBloc, OvertimeState>(
        listener: (context, state) {
          if (state is OvertimeSuccess) {
            SnackBarUtils.showSuccess(context, state.message);
            context.read<OvertimeBloc>().add(FetchMyOvertimesRequested());
          } else if (state is OvertimeFailure) {
            DialogUtils.showError(
                context: context, title: "Gagal", message: state.errorMessage);
          }
        },
        builder: (context, state) {
          if (state is OvertimeLoading) {
            return const Center(
                child: CircularProgressIndicator(color: primaryRed));
          }

          if (state is MyOvertimesLoaded) {
            final overtimes = state.overtimes;

            if (overtimes.isEmpty) {
              return Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.history_toggle_off,
                        color: Colors.grey[300], size: 80),
                    const SizedBox(height: 16),
                    Text(
                      'Belum ada pengajuan lembur',
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 16,
                          color: subtitleColor,
                          fontWeight: FontWeight.w700),
                    ),
                  ],
                ),
              );
            }

            return RefreshIndicator(
              onRefresh: () async {
                context.read<OvertimeBloc>().add(FetchMyOvertimesRequested());
              },
              color: primaryRed,
              child: AnimationLimiter(
                child: ListView.builder(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 20, vertical: 24),
                  itemCount: overtimes.length,
                  itemBuilder: (context, index) {
                    return AnimationConfiguration.staggeredList(
                      position: index,
                      duration: const Duration(milliseconds: 500),
                      child: SlideAnimation(
                        verticalOffset: 50.0,
                        child: FadeInAnimation(
                          child: _buildOvertimeCard(context, overtimes[index],
                              primaryRed, textColor, subtitleColor),
                        ),
                      ),
                    );
                  },
                ),
              ),
            );
          }

          return const SizedBox.shrink();
        },
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          final overtimeBloc = context.read<OvertimeBloc>();
          Navigator.of(context)
              .push(
            MaterialPageRoute(
              builder: (context) => BlocProvider.value(
                value: overtimeBloc,
                child: const OvertimeRequestPage(),
              ),
            ),
          )
              .then((_) {
            if (mounted) {
              context.read<OvertimeBloc>().add(FetchMyOvertimesRequested());
            }
          });
        },
        backgroundColor: primaryRed,
        elevation: 8,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        icon: const Icon(Icons.add, color: Colors.white, size: 24),
        label: Text(
          'AJUKAN LEMBUR',
          style: GoogleFonts.plusJakartaSans(
            color: Colors.white,
            fontWeight: FontWeight.w800,
            fontSize: 13,
            letterSpacing: 0.5,
          ),
        ),
      ),
    );
  }

  Widget _buildOvertimeCard(
    BuildContext context,
    Overtime overtime,
    Color primaryRed,
    Color textColor,
    Color subtitleColor,
  ) {
    final statusColor = _getStatusColor(overtime.status);
    final isPending = overtime.status.toLowerCase() == 'pending';

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(color: AppColors.grayBorder),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 10,
              offset: const Offset(0, 4)),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      DateFormat('EEEE, dd MMM yyyy').format(overtime.date),
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 15,
                          fontWeight: FontWeight.w800,
                          color: textColor),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      '${DateFormat('HH:mm').format(overtime.startTime)} - ${DateFormat('HH:mm').format(overtime.endTime)} (${(overtime.durationMinutes / 60).toStringAsFixed(1)} Jam)',
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 12,
                          color: subtitleColor,
                          fontWeight: FontWeight.w600),
                    ),
                  ],
                ),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 10, vertical: 4),
                      decoration: BoxDecoration(
                        color: statusColor.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        overtime.status.toUpperCase(),
                        style: GoogleFonts.plusJakartaSans(
                            fontSize: 10,
                            fontWeight: FontWeight.w900,
                            color: statusColor,
                            letterSpacing: 0.5),
                      ),
                    ),
                    const SizedBox(height: 6),
                    Text(
                      overtime.type.replaceAll('_', ' ').toUpperCase(),
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 9,
                          fontWeight: FontWeight.w800,
                          color: AppColors.textTertiary,
                          letterSpacing: 0.5),
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 16),
            const Divider(height: 1),
            const SizedBox(height: 16),
            Text(
              overtime.reason,
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 14, color: textColor, fontWeight: FontWeight.w600),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
            if (isPending) ...[
              const SizedBox(height: 16),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  _buildMiniButton(
                    onPressed: () => _editOvertime(context, overtime),
                    icon: Icons.edit_rounded,
                    label: 'Edit',
                    color: AppColors.textSecondary,
                  ),
                  const SizedBox(width: 8),
                  _buildMiniButton(
                    onPressed: () => _deleteOvertime(context, overtime),
                    icon: Icons.delete_rounded,
                    label: 'Delete',
                    color: AppColors.error,
                  ),
                ],
              ),
            ],
            if (overtime.rejectReason != null &&
                overtime.rejectReason!.isNotEmpty) ...[
              const SizedBox(height: 12),
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: AppColors.error.withOpacity(0.05),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: AppColors.error.withOpacity(0.1)),
                ),
                child: Text(
                  'Alasan Penolakan: ${overtime.rejectReason}',
                  style: GoogleFonts.plusJakartaSans(
                      fontSize: 11,
                      color: AppColors.error,
                      fontWeight: FontWeight.w700,
                      fontStyle: FontStyle.italic),
                ),
              )
            ]
          ],
        ),
      ),
    );
  }

  Widget _buildMiniButton(
      {required VoidCallback onPressed,
      required IconData icon,
      required String label,
      required Color color}) {
    return InkWell(
      onTap: onPressed,
      borderRadius: BorderRadius.circular(8),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        decoration: BoxDecoration(
          color: color.withOpacity(0.08),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          children: [
            Icon(icon, size: 14, color: color),
            const SizedBox(width: 6),
            Text(label,
                style: TextStyle(
                    fontSize: 11, fontWeight: FontWeight.w800, color: color)),
          ],
        ),
      ),
    );
  }

  Color _getStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'pending':
        return Colors.orange;
      case 'approved':
        return const Color(0xFF10B981);
      case 'rejected':
        return AppColors.error;
      default:
        return AppColors.textTertiary;
    }
  }

  void _editOvertime(BuildContext context, Overtime overtime) {
    final overtimeBloc = context.read<OvertimeBloc>();
    Navigator.of(context)
        .push(
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: overtimeBloc,
          child: OvertimeRequestPage(overtimeToEdit: overtime),
        ),
      ),
    )
        .then((_) {
      if (mounted) {
        context.read<OvertimeBloc>().add(FetchMyOvertimesRequested());
      }
    });
  }

  void _deleteOvertime(BuildContext context, Overtime overtime) {
    final overtimeBloc = context.read<OvertimeBloc>();
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        title: Text('Hapus Pengajuan?',
            style: GoogleFonts.plusJakartaSans(fontWeight: FontWeight.w900)),
        content: Text('Apakah Anda yakin ingin menghapus pengajuan lembur ini?',
            style: GoogleFonts.plusJakartaSans(
                fontWeight: FontWeight.w600,
                fontSize: 13,
                color: AppColors.textSecondary)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('BATAL',
                style: TextStyle(
                    fontWeight: FontWeight.w800,
                    color: AppColors.textTertiary)),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(dialogContext);
              overtimeBloc.add(DeleteOvertimeRequested(overtime.id));
            },
            style: ElevatedButton.styleFrom(
                backgroundColor: AppColors.error,
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(10))),
            child: const Text('HAPUS',
                style: TextStyle(
                    fontWeight: FontWeight.w900, color: Colors.white)),
          ),
        ],
      ),
    );
  }
}
