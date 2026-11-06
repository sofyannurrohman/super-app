import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/notification/presentation/bloc/notification_bloc.dart';
import 'package:hris_app/features/notification/domain/entities/app_notification.dart';
import 'package:hris_app/features/announcement/presentation/pages/announcement_detail_page.dart';
import 'package:hris_app/features/announcement/data/models/announcement_model.dart';
import 'package:hris_app/features/leave/presentation/pages/leave_page.dart';
import 'package:hris_app/features/overtime/presentation/pages/overtime_list_page.dart';
import 'package:hris_app/features/reimbursement/presentation/pages/reimbursement_list_page.dart';
import 'package:hris_app/features/overtime/presentation/bloc/overtime_bloc.dart';
import 'package:hris_app/features/reimbursement/presentation/bloc/reimbursement_bloc.dart';
import 'package:hris_app/injection.dart' as di;

class NotificationListPage extends StatelessWidget {
  const NotificationListPage({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text(
          'NOTIFIKASI',
          style: GoogleFonts.plusJakartaSans(
            fontSize: 14,
            fontWeight: FontWeight.w800,
            letterSpacing: 1.5,
            color: AppColors.textPrimary,
          ),
        ),
        centerTitle: true,
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new_rounded,
              color: AppColors.textPrimary, size: 20),
          onPressed: () => Navigator.of(context).pop(),
        ),
        actions: [
          IconButton(
            icon:
                const Icon(Icons.refresh_rounded, color: AppColors.textPrimary),
            onPressed: () => context
                .read<NotificationBloc>()
                .add(FetchNotificationsRequested()),
          ),
        ],
      ),
      body: BlocBuilder<NotificationBloc, NotificationState>(
        builder: (context, state) {
          if (state is NotificationLoading) {
            return const Center(
                child: CircularProgressIndicator(color: AppColors.primaryRed));
          } else if (state is NotificationLoaded) {
            if (state.notifications.isEmpty) {
              return _buildEmptyState();
            }
            return RefreshIndicator(
              onRefresh: () async {
                context
                    .read<NotificationBloc>()
                    .add(FetchNotificationsRequested());
              },
              child: ListView.separated(
                padding: const EdgeInsets.all(20),
                itemCount: state.notifications.length,
                separatorBuilder: (_, __) => const SizedBox(height: 12),
                itemBuilder: (context, index) {
                  final notification = state.notifications[index];
                  return _buildNotificationItem(context, notification);
                },
              ),
            );
          } else if (state is NotificationFailure) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text('Gagal memuat notifikasi',
                      style: GoogleFonts.plusJakartaSans(
                          fontWeight: FontWeight.w700)),
                  const SizedBox(height: 8),
                  Text(state.message,
                      style: GoogleFonts.plusJakartaSans(
                          fontSize: 12, color: AppColors.textTertiary),
                      textAlign: TextAlign.center),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => context
                        .read<NotificationBloc>()
                        .add(FetchNotificationsRequested()),
                    child: const Text('Coba Lagi'),
                  ),
                ],
              ),
            );
          }
          return const SizedBox();
        },
      ),
    );
  }

  Widget _buildNotificationItem(
      BuildContext context, AppNotification notification) {
    Color typeColor;
    IconData icon;
    String subtitle = notification.content;

    switch (notification.type) {
      case AppNotificationType.announcement:
        final ann = notification.originalData as Announcement;
        switch (ann.category) {
          case AnnouncementCategory.policy:
            typeColor = AppColors.primaryRed;
            icon = Icons.gavel_rounded;
            break;
          case AnnouncementCategory.birthday:
            typeColor = const Color(0xFFEC4899);
            icon = Icons.cake_rounded;
            break;
          case AnnouncementCategory.alert:
            typeColor = AppColors.warning;
            icon = Icons.priority_high_rounded;
            break;
          default:
            typeColor = const Color(0xFF6366F1);
            icon = Icons.info_rounded;
        }
        break;
      case AppNotificationType.leave:
        typeColor = const Color(0xFF8B5CF6);
        icon = Icons.event_available_rounded;
        break;
      case AppNotificationType.overtime:
        typeColor = const Color(0xFFF59E0B);
        icon = Icons.timer_rounded;
        break;
      case AppNotificationType.reimbursement:
        typeColor = const Color(0xFF10B981);
        icon = Icons.receipt_long_rounded;
        break;
    }

    return InkWell(
      onTap: () => _handleNotificationTap(context, notification),
      borderRadius: BorderRadius.circular(20),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(20),
          border: Border.all(color: AppColors.grayBorder.withOpacity(0.5)),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: typeColor.withOpacity(0.1),
                shape: BoxShape.circle,
              ),
              child: Icon(icon, color: typeColor, size: 20),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          notification.title,
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 14,
                            fontWeight: FontWeight.w800,
                            color: AppColors.textPrimary,
                          ),
                        ),
                      ),
                      if (notification.status != null)
                        _buildStatusBadge(notification.status!),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      color: AppColors.textSecondary,
                      height: 1.4,
                    ),
                    maxLines: 3,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 12),
                  Text(
                    DateFormat('dd MMM yyyy, HH:mm')
                        .format(notification.timestamp),
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 10,
                      fontWeight: FontWeight.w700,
                      color: AppColors.textTertiary,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusBadge(String status) {
    Color color;
    switch (status.toUpperCase()) {
      case 'APPROVED':
        color = AppColors.success;
        break;
      case 'REJECTED':
        color = AppColors.error;
        break;
      case 'PENDING':
        color = AppColors.warning;
        break;
      default:
        color = AppColors.textTertiary;
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        status.toUpperCase(),
        style: GoogleFonts.plusJakartaSans(
          fontSize: 9,
          fontWeight: FontWeight.w800,
          color: color,
        ),
      ),
    );
  }

  void _handleNotificationTap(
      BuildContext context, AppNotification notification) {
    switch (notification.type) {
      case AppNotificationType.announcement:
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) =>
                AnnouncementDetailPage(announcement: notification.originalData),
          ),
        );
        break;
      case AppNotificationType.leave:
        Navigator.of(context)
            .push(MaterialPageRoute(builder: (_) => const LeavePage()));
        break;
      case AppNotificationType.overtime:
        Navigator.of(context).push(MaterialPageRoute(
            builder: (_) => BlocProvider(
                create: (_) => di.sl<OvertimeBloc>(),
                child: const OvertimeListPage())));
        break;
      case AppNotificationType.reimbursement:
        Navigator.of(context).push(MaterialPageRoute(
            builder: (_) => BlocProvider(
                create: (_) => di.sl<ReimbursementBloc>(),
                child: const ReimbursementListPage())));
        break;
    }
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Container(
            padding: const EdgeInsets.all(32),
            decoration: BoxDecoration(
              color: Colors.white,
              shape: BoxShape.circle,
              border: Border.all(color: AppColors.grayBorder.withOpacity(0.5)),
            ),
            child: Icon(
              Icons.notifications_none_rounded,
              size: 64,
              color: AppColors.textTertiary.withOpacity(0.2),
            ),
          ),
          const SizedBox(height: 24),
          Text(
            'Belum Ada Notifikasi',
            style: GoogleFonts.plusJakartaSans(
              fontSize: 18,
              fontWeight: FontWeight.w800,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Kami akan memberi tahu Anda jika ada update pengajuan atau info terbaru.',
            style: GoogleFonts.plusJakartaSans(
              fontSize: 13,
              color: AppColors.textTertiary,
              fontWeight: FontWeight.w600,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }
}
