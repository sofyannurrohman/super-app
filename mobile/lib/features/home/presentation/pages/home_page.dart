import 'package:cached_network_image/cached_network_image.dart';
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/features/attendance/presentation/bloc/attendance_bloc.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_event.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_state.dart';
import 'package:hris_app/features/attendance/presentation/pages/face_verification_page.dart';
import 'package:hris_app/core/utils/dialog_utils.dart';
import 'package:hris_app/features/attendance/presentation/pages/attendance_page.dart';
import 'package:hris_app/features/attendance/presentation/pages/activity_history_page.dart';
import 'package:hris_app/features/leave/presentation/pages/leave_page.dart';
import 'package:hris_app/features/kpi/presentation/bloc/kpi_bloc.dart';
import 'package:hris_app/features/kpi/presentation/pages/kpi_page.dart';
import 'package:hris_app/features/attendance/domain/entities/attendance.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_event.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:hris_app/features/payroll/presentation/bloc/payroll_bloc.dart';
import 'package:hris_app/features/payroll/presentation/pages/payslip_list_page.dart';
import 'package:hris_app/features/overtime/presentation/bloc/overtime_bloc.dart';
import 'package:hris_app/features/overtime/presentation/pages/overtime_list_page.dart';
import 'package:hris_app/features/profile/presentation/pages/face_registration_page.dart';
import 'package:hris_app/features/profile/presentation/pages/profile_overview_page.dart';
import 'package:hris_app/features/directory/presentation/pages/directory_page.dart';
import 'package:hris_app/features/directory/presentation/bloc/directory_bloc.dart';
import 'package:hris_app/features/supervisor/presentation/pages/supervisor_dashboard_page.dart';
import 'package:hris_app/features/approval/presentation/bloc/approval_bloc.dart';
import 'package:hris_app/features/announcement/presentation/bloc/announcement_bloc.dart';
import 'package:hris_app/features/announcement/data/models/announcement_model.dart';
import 'package:hris_app/features/announcement/presentation/pages/announcement_detail_page.dart';
import 'package:hris_app/features/reimbursement/presentation/pages/reimbursement_list_page.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/sales/presentation/pages/sales_dashboard_page.dart';
import 'package:hris_app/features/home/presentation/widgets/warning_popup_dialog.dart';
import 'package:hris_app/features/sales/presentation/pages/director_dashboard_page.dart';
import 'package:hris_app/features/sales/presentation/pages/delivery_dashboard_page.dart';
import 'package:hris_app/features/sales/presentation/pages/operator_spanduk_dashboard_page.dart';
import 'package:hris_app/features/sales/presentation/pages/branding_campaign_video_page.dart';
import 'package:hris_app/features/inventory/presentation/pages/warehouse_dashboard_page.dart';
import 'package:hris_app/features/home/presentation/pages/hr_dashboard_page.dart';
import 'package:hris_app/features/finance/presentation/pages/cashier_dashboard_page.dart';

import 'package:hris_app/features/announcement/presentation/pages/notification_list_page.dart';
import 'package:hris_app/features/announcement/presentation/widgets/announcement_popup_dialog.dart';
import 'package:hris_app/features/notification/presentation/bloc/notification_bloc.dart';
import 'package:hris_app/features/epkwt/presentation/pages/epkwt_list_page.dart';
import 'package:hris_app/features/memo_request/presentation/pages/memo_request_page.dart';
import 'package:hris_app/features/purchase_request/presentation/pages/purchase_request_page.dart';
import 'package:hris_app/core/services/notification_service.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> with WidgetsBindingObserver {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addObserver(this);
    context.read<AttendanceBloc>().add(FetchHomeDataRequested());
    context.read<AnnouncementBloc>().add(GetPendingPopupsEvent());
    context.read<AnnouncementBloc>().add(FetchAnnouncementsRequested());
    di.sl<NotificationService>().requestPermissions();
    _checkLastDashboardRoute();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      context.read<AnnouncementBloc>().add(GetPendingPopupsEvent());
      context.read<AnnouncementBloc>().add(FetchAnnouncementsRequested());
    }
  }

  Future<void> _checkLastDashboardRoute() async {
    final prefs = await SharedPreferences.getInstance();
    final lastRoute = prefs.getString('last_dashboard_route');
    if (lastRoute != null && mounted) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) {
          Navigator.of(context).pushNamed(lastRoute).then((_) async {
            final p = await SharedPreferences.getInstance();
            await p.remove('last_dashboard_route');
          });
        }
      });
    }
  }

  void _showPendingPopups(List<dynamic> popups) async {
    for (var popup in popups) {
      await _showSinglePopup(popup);
    }
  }

  Future<void> _showSinglePopup(dynamic popup) async {
    if (!mounted) return;
    await AnnouncementPopupDialog.show(
      context: context,
      announcement: popup,
      isAlreadyConfirmed: false,
      onAcknowledge: () {
        if (popup.id != null) {
          context.read<AnnouncementBloc>().add(
            AcknowledgePopupEvent(popup.id!),
          );
        }
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<AnnouncementBloc, AnnouncementState>(
      listenWhen: (previous, current) =>
          previous.pendingPopups != current.pendingPopups &&
          current.pendingPopups.isNotEmpty,
      listener: (context, state) {
        if (state.pendingPopups.isNotEmpty) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (mounted) {
              _showPendingPopups(state.pendingPopups);
            }
          });
        }
      },
      child: const HomeView(),
    );
  }
}

class HomeView extends StatefulWidget {
  const HomeView({super.key});

  @override
  State<HomeView> createState() => _HomeViewState();
}

class _HomeViewState extends State<HomeView> {
  int _currentIndex = 0;
  bool _hasShownFaceRegisterWarning = false;

  void _showFaceRegisterWarningDialog(BuildContext context) {
    showGeneralDialog(
      context: context,
      barrierDismissible: false,
      barrierLabel: '',
      barrierColor: Colors.black.withOpacity(0.6),
      transitionDuration: const Duration(milliseconds: 450),
      pageBuilder: (dialogContext, anim1, anim2) {
        return Center(
          child: Dialog(
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(28),
            ),
            elevation: 0,
            backgroundColor: Colors.transparent,
            child: Container(
              padding: const EdgeInsets.all(28),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(28),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.15),
                    blurRadius: 25,
                    offset: const Offset(0, 15),
                  ),
                ],
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      color: AppColors.warning.withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      Icons.face_retouching_off_rounded,
                      color: AppColors.warning,
                      size: 56,
                    ),
                  ),
                  const SizedBox(height: 24),
                  Text(
                    'Registrasi Wajah Diperlukan',
                    textAlign: TextAlign.center,
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 20,
                      fontWeight: FontWeight.w900,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    'Anda belum mendaftarkan wajah Anda di sistem. Silakan daftarkan wajah Anda terlebih dahulu sebelum memulai absensi untuk pertama kali.',
                    textAlign: TextAlign.center,
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 14,
                      color: AppColors.textSecondary,
                      height: 1.5,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  const SizedBox(height: 32),
                  Column(
                    children: [
                      SizedBox(
                        width: double.infinity,
                        height: 52,
                        child: Container(
                          decoration: BoxDecoration(
                            borderRadius: BorderRadius.circular(16),
                            gradient: const LinearGradient(
                              colors: AppColors.primaryGradient,
                              begin: Alignment.topLeft,
                              end: Alignment.bottomRight,
                            ),
                          ),
                          child: ElevatedButton(
                            onPressed: () {
                              Navigator.of(
                                dialogContext,
                                rootNavigator: true,
                              ).pop();
                              Navigator.of(context).push(
                                MaterialPageRoute(
                                  builder: (_) => const FaceRegistrationPage(),
                                ),
                              );
                            },
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.transparent,
                              shadowColor: Colors.transparent,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(16),
                              ),
                            ),
                            child: Text(
                              'DAFTARKAN SEKARANG',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 13,
                                fontWeight: FontWeight.w800,
                                color: Colors.white,
                                letterSpacing: 0.5,
                              ),
                            ),
                          ),
                        ),
                      ),
                      const SizedBox(height: 12),
                      TextButton(
                        onPressed: () {
                          Navigator.of(
                            dialogContext,
                            rootNavigator: true,
                          ).pop();
                        },
                        style: TextButton.styleFrom(
                          foregroundColor: AppColors.textTertiary,
                          padding: const EdgeInsets.symmetric(vertical: 12),
                        ),
                        child: Text(
                          'NANTI SAJA',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 13,
                            fontWeight: FontWeight.w700,
                            letterSpacing: 0.5,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        );
      },
      transitionBuilder: (context, anim1, anim2, child) {
        final curveValue = Curves.easeInOutBack.transform(anim1.value);
        return Transform.scale(
          scale: curveValue,
          child: Opacity(opacity: anim1.value, child: child),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final authState = context.watch<AuthBloc>().state;
    if (authState is Authenticated) {
      final profile = authState.userProfile;
      if (profile != null) {
        if (!_hasShownFaceRegisterWarning) {
          final isFaceRegistered =
              profile['is_face_registered'] ?? profile['IsFaceRegistered'];
          if (isFaceRegistered == false || isFaceRegistered == null) {
            _hasShownFaceRegisterWarning = true;
            WidgetsBinding.instance.addPostFrameCallback((_) {
              if (mounted) {
                _showFaceRegisterWarningDialog(context);
              }
            });
          }
        }

        // Check active warning from CS, SPV, or HR
        final activeWarning =
            profile['active_warning'] ?? profile['ActiveWarning'];
        if (activeWarning != null) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (mounted) {
              WarningPopupDialog.show(
                context: context,
                warningData: activeWarning,
              );
            }
          });
        }
      }
    }

    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      drawer: _buildDrawer(),
      body: _buildPage(_currentIndex),
      bottomNavigationBar: _buildBottomNav(),
      floatingActionButton: _currentIndex == 0
          ? FloatingActionButton.extended(
              onPressed: () => _launchWhatsAppCS(context),
              backgroundColor: const Color(0xFF25D366),
              elevation: 4,
              icon: const Icon(
                Icons.support_agent_rounded,
                color: Colors.white,
                size: 24,
              ),
              label: Text(
                'Layanan Bantuan',
                style: GoogleFonts.plusJakartaSans(
                  color: Colors.white,
                  fontWeight: FontWeight.w700,
                  fontSize: 12,
                ),
              ),
            )
          : null,
    );
  }

  Future<void> _launchWhatsAppCS(BuildContext context) async {
    const String phoneNumber = '6282322965764';
    const String groupUrl =
        'https://chat.whatsapp.com/LiOx9lICB1PHjtduvVbzM9?s=sw&p=a&mlu=4&ilr=4';

    final option = await showModalBottomSheet<String>(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (sheetContext) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(24, 20, 24, 16),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Pilih layanan bantuan',
                style: GoogleFonts.plusJakartaSans(
                  fontSize: 18,
                  fontWeight: FontWeight.w800,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Hubungi CS secara langsung atau bergabung ke grup bantuan.',
                style: GoogleFonts.plusJakartaSans(color: Colors.black54),
              ),
              const SizedBox(height: 16),
              ListTile(
                leading: const CircleAvatar(
                  backgroundColor: Color(0xFFE8F8EE),
                  child: Icon(Icons.chat_rounded, color: Color(0xFF25D366)),
                ),
                title: const Text('Chat Personal CS'),
                subtitle: const Text('Kirim pesan langsung ke CS'),
                onTap: () => Navigator.pop(sheetContext, 'personal'),
              ),
              ListTile(
                leading: const CircleAvatar(
                  backgroundColor: Color(0xFFE8F8EE),
                  child: Icon(Icons.groups_rounded, color: Color(0xFF25D366)),
                ),
                title: const Text('Gabung Grup WhatsApp'),
                subtitle: const Text('Bergabung ke grup bantuan Karsa HR'),
                onTap: () => Navigator.pop(sheetContext, 'group'),
              ),
            ],
          ),
        ),
      ),
    );
    if (option == null) return;

    final Uri waUrl = option == 'group'
        ? Uri.parse(groupUrl)
        : Uri.parse(
            'https://wa.me/$phoneNumber?text=${Uri.encodeComponent('Halo, saya butuh bantuan terkait aplikasi Karsa HR Super App.')}',
          );

    if (await canLaunchUrl(waUrl)) {
      await launchUrl(waUrl, mode: LaunchMode.externalApplication);
    } else if (context.mounted) {
      DialogUtils.showError(
        context: context,
        title: 'Gagal',
        message: 'Tidak dapat membuka WhatsApp.',
      );
    }
  }

  Widget _buildPage(int index) {
    switch (index) {
      case 0:
        return const DashboardTab();
      case 1:
        return const AttendancePage();
      case 2:
        return const LeavePage();
      case 3:
        return const ProfileOverviewPage();
      default:
        return const DashboardTab();
    }
  }

  Widget _buildBottomNav() {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.04),
            blurRadius: 20,
            offset: const Offset(0, -10),
          ),
        ],
      ),
      child: BottomNavigationBar(
        currentIndex: _currentIndex,
        onTap: (index) => setState(() => _currentIndex = index),
        type: BottomNavigationBarType.fixed,
        backgroundColor: Colors.white,
        selectedItemColor: AppColors.primaryGreen,
        unselectedItemColor: AppColors.textTertiary,
        selectedLabelStyle: const TextStyle(
          fontWeight: FontWeight.w800,
          fontSize: 11,
          letterSpacing: 0.5,
        ),
        unselectedLabelStyle: const TextStyle(
          fontWeight: FontWeight.w700,
          fontSize: 11,
        ),
        elevation: 0,
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.dashboard_rounded),
            label: 'Dashboard',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.qr_code_scanner_rounded),
            label: 'Presensi',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.event_available_rounded),
            label: 'Cuti',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.person_rounded),
            label: 'Profil',
          ),
        ],
      ),
    );
  }

  Widget _buildDrawer() {
    return Drawer(
      backgroundColor: Colors.white,
      child: BlocBuilder<AuthBloc, AuthState>(
        builder: (context, state) {
          Map<String, dynamic>? profile;
          if (state is Authenticated) {
            profile = state.userProfile;
          }
          final name = '${profile?['first_name'] ?? 'User'}'.toUpperCase();
          final job = profile?['job_position']?['title'] ?? 'DASHBOARD PANEL';
          final role =
              (profile?['role'] ??
                      profile?['Role'] ??
                      profile?['user']?['role'] ??
                      '')
                  .toString()
                  .toLowerCase();
          const campaignRoles = {
            'superadmin',
            'manager',
            'direktur',
            'admin_designer',
            'validator_desain',
            'validator_toko',
            'admin_purchasing',
          };

          return Column(
            children: [
              Container(
                width: double.infinity,
                padding: const EdgeInsets.fromLTRB(24, 60, 24, 32),
                decoration: const BoxDecoration(
                  gradient: LinearGradient(
                    colors: AppColors.primaryGradient,
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      width: 72,
                      height: 72,
                      decoration: const BoxDecoration(
                        color: Colors.white,
                        shape: BoxShape.circle,
                      ),
                      clipBehavior: Clip.antiAlias,
                      child:
                          (profile?['face_reference_url'] != null &&
                              profile!['face_reference_url']
                                  .toString()
                                  .isNotEmpty)
                          ? CachedNetworkImage(
                              imageUrl: AppConstants.buildFaceImageUrl(
                                profile['face_reference_url'].toString(),
                              ),
                              fit: BoxFit.cover,
                              placeholder: (context, url) => const Center(
                                child: CircularProgressIndicator(),
                              ),
                              errorWidget: (context, url, error) =>
                                  const Icon(Icons.error),
                              memCacheWidth: 800,
                            )
                          : const Icon(
                              Icons.person,
                              size: 40,
                              color: AppColors.primaryGreen,
                            ),
                    ),
                    const SizedBox(height: 16),
                    const Text(
                      'MENU UTAMA',
                      style: TextStyle(
                        color: Colors.white70,
                        fontWeight: FontWeight.w800,
                        fontSize: 10,
                        letterSpacing: 2,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      name,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.w900,
                        fontSize: 20,
                      ),
                    ),
                    Text(
                      job.toUpperCase(),
                      style: TextStyle(
                        color: Colors.white.withOpacity(0.8),
                        fontWeight: FontWeight.w700,
                        fontSize: 10,
                        letterSpacing: 1,
                      ),
                    ),
                  ],
                ),
              ),
              Expanded(
                child: ListView(
                  padding: EdgeInsets.zero,
                  children: [
                    if (campaignRoles.contains(role)) ...[
                      const Padding(
                        padding: EdgeInsets.fromLTRB(24, 24, 24, 8),
                        child: Text(
                          'BRANDING',
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w900,
                            color: AppColors.textTertiary,
                            letterSpacing: 1.5,
                          ),
                        ),
                      ),
                      _buildDrawerItem(
                        Icons.video_library_rounded,
                        'Video Kampanye',
                        () {
                          Navigator.pop(context);
                          Navigator.of(context).push(
                            MaterialPageRoute(
                              builder: (_) => const BrandingCampaignVideoPage(),
                            ),
                          );
                        },
                      ),
                    ],
                    const Padding(
                      padding: EdgeInsets.fromLTRB(24, 24, 24, 8),
                      child: Text(
                        'ADMINISTRASI',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                          color: AppColors.textTertiary,
                          letterSpacing: 1.5,
                        ),
                      ),
                    ),
                    _buildDrawerItem(
                      Icons.payments_rounded,
                      'Slip Gaji Digital',
                      () {
                        Navigator.pop(context);
                        Navigator.of(context).push(
                          MaterialPageRoute(
                            builder: (_) => BlocProvider(
                              create: (_) => di.sl<PayrollBloc>(),
                              child: const PayslipListPage(),
                            ),
                          ),
                        );
                      },
                    ),
                    _buildDrawerItem(Icons.timer_rounded, 'Lembur', () {
                      Navigator.pop(context);
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) => BlocProvider(
                            create: (_) => di.sl<OvertimeBloc>(),
                            child: const OvertimeListPage(),
                          ),
                        ),
                      );
                    }),
                    _buildDrawerItem(
                      Icons.receipt_long_rounded,
                      'Reimbursement',
                      () {
                        Navigator.pop(context);
                        Navigator.of(context).push(
                          MaterialPageRoute(
                            builder: (_) => const ReimbursementListPage(),
                          ),
                        );
                      },
                    ),
                    _buildDrawerItem(
                      Icons.face_retouching_natural_rounded,
                      'Registrasi Wajah',
                      () {
                        Navigator.pop(context);
                        Navigator.of(context).push(
                          MaterialPageRoute(
                            builder: (_) => const FaceRegistrationPage(),
                          ),
                        );
                      },
                    ),
                    _buildDrawerItem(
                      Icons.trending_up_rounded,
                      'KPI & Performa',
                      () {
                        Navigator.pop(context);
                        Navigator.of(context).push(
                          MaterialPageRoute(
                            builder: (_) => BlocProvider(
                              create: (_) => di.sl<KPIBloc>(),
                              child: const KPIPage(),
                            ),
                          ),
                        );
                      },
                    ),
                    _buildDrawerItem(
                      Icons.contact_phone_rounded,
                      'Direktori Karyawan',
                      () {
                        Navigator.pop(context);
                        Navigator.of(context).push(
                          MaterialPageRoute(
                            builder: (_) => BlocProvider(
                              create: (_) => di.sl<DirectoryBloc>(),
                              child: const DirectoryPage(),
                            ),
                          ),
                        );
                      },
                    ),
                    _buildDrawerItem(
                      Icons.assignment_turned_in_rounded,
                      'Elektronik PKWT (EPKWT)',
                      () {
                        Navigator.pop(context);
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (_) => const EpkwtListPage(),
                          ),
                        );
                      },
                    ),
                    _buildDrawerItem(Icons.edit_document, 'Request Memo', () {
                      Navigator.pop(context);
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => const MemoRequestPage(),
                        ),
                      );
                    }),
                    _buildDrawerItem(
                      Icons.assignment_rounded,
                      'Purchase Request',
                      () {
                        Navigator.pop(context);
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (_) => const PurchaseRequestPage(),
                          ),
                        );
                      },
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),
              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 20,
                ),
                child: InkWell(
                  onTap: () => _showLogoutConfirmation(context),
                  borderRadius: BorderRadius.circular(16),
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      vertical: 14,
                      horizontal: 16,
                    ),
                    decoration: BoxDecoration(
                      color: AppColors.error.withOpacity(0.08),
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: const Row(
                      children: [
                        Icon(Icons.logout_rounded, color: AppColors.error),
                        SizedBox(width: 16),
                        Text(
                          'Keluar Akun',
                          style: TextStyle(
                            color: AppColors.error,
                            fontWeight: FontWeight.w800,
                            fontSize: 14,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildDrawerItem(
    IconData icon,
    String title,
    VoidCallback onTap, {
    Color? color,
  }) {
    return ListTile(
      leading: Icon(icon, color: color ?? AppColors.textSecondary, size: 24),
      title: Text(
        title,
        style: TextStyle(
          color: color ?? AppColors.textPrimary,
          fontWeight: FontWeight.w700,
          fontSize: 14,
        ),
      ),
      onTap: onTap,
      contentPadding: const EdgeInsets.symmetric(horizontal: 24, vertical: 4),
    );
  }

  void _showLogoutConfirmation(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
        title: const Text(
          'Keluar Akun?',
          style: TextStyle(fontWeight: FontWeight.w900),
        ),
        content: const Text('Apakah Anda yakin ingin keluar dari aplikasi?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text(
              'BATAL',
              style: TextStyle(
                fontWeight: FontWeight.w700,
                color: AppColors.textTertiary,
              ),
            ),
          ),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(context);
              context.read<AuthBloc>().add(LogoutRequested());
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: AppColors.primaryGreen,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'KELUAR',
              style: TextStyle(
                fontWeight: FontWeight.w800,
                color: Colors.white,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class DashboardTab extends StatelessWidget {
  const DashboardTab({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<AttendanceBloc, AttendanceState>(
      builder: (context, state) {
        if (state is AttendanceLoading) {
          return const Center(
            child: CircularProgressIndicator(color: AppColors.primaryGreen),
          );
        }

        Map<String, dynamic>? profile;
        List<Attendance> recentActivity = [];
        Map<String, dynamic>? statistics;

        // Get profile from AuthBloc for stability
        final authState = context.read<AuthBloc>().state;
        if (authState is Authenticated) {
          profile = authState.userProfile;
        }

        // Handle offline state when no profile is cached
        if (profile == null) {
          return Center(
            child: Padding(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(
                    Icons.wifi_off_rounded,
                    size: 80,
                    color: AppColors.textTertiary,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'Tidak Ada Koneksi Internet',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 18,
                      fontWeight: FontWeight.w800,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Gagal memuat data profil. Pastikan Anda memiliki koneksi internet yang stabil untuk memuat data awal.',
                    textAlign: TextAlign.center,
                    style: GoogleFonts.plusJakartaSans(
                      color: AppColors.textSecondary,
                    ),
                  ),
                  const SizedBox(height: 24),
                  ElevatedButton.icon(
                    onPressed: () {
                      context.read<AttendanceBloc>().add(
                        FetchHomeDataRequested(),
                      );
                      context.read<AuthBloc>().add(CheckAuthStatusRequested());
                    },
                    icon: const Icon(
                      Icons.refresh_rounded,
                      color: Colors.white,
                    ),
                    label: Text(
                      'Coba Lagi',
                      style: GoogleFonts.plusJakartaSans(
                        fontWeight: FontWeight.w700,
                        color: Colors.white,
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppColors.primaryGreen,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 24,
                        vertical: 12,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          );
        }

        String? errorMessage;
        if (state is HomeDataLoaded) {
          // stats/history still come from AttendanceBloc
          recentActivity = state.recentActivity;
          statistics = state.statistics;
          // If AttendanceBloc also has profile, it can override
          profile = state.profile;
        } else if (state is AttendanceHistoryLoaded) {
          recentActivity = state.history;
          if (state.profile != null) profile = state.profile;
          if (state.statistics != null) statistics = state.statistics;
        } else if (state is HomeDataFailure) {
          errorMessage = state.message;
        }

        String displayName = '';
        if (profile != null) {
          final firstName = profile['first_name'] ?? profile['FirstName'] ?? '';
          final lastName = profile['last_name'] ?? profile['LastName'] ?? '';
          displayName = '$firstName $lastName'.trim();

          // Fallback if split first/last not found
          if (displayName.isEmpty) {
            displayName =
                profile['name'] ??
                profile['Name'] ??
                profile['full_name'] ??
                '';
          }
        }

        return RefreshIndicator(
          onRefresh: () async {
            context.read<AttendanceBloc>().add(FetchHomeDataRequested());
          },
          child: CustomScrollView(
            slivers: [
              _buildSliverAppBar(context, profile, displayName),
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.fromLTRB(24, 24, 24, 40),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // --- Banner Peringatan Karyawan Bermasalah ---
                      if (profile != null &&
                          (profile['active_warning'] != null ||
                              profile['ActiveWarning'] != null)) ...[
                        Builder(
                          builder: (context) {
                            final warningData =
                                profile!['active_warning'] ??
                                profile!['ActiveWarning'];
                            String displayMsg = warningData.toString();
                            String displayType = '-';
                            Color bgColor = Colors.red.shade50;
                            Color borderColor = Colors.red.shade200;
                            Color iconColor = Colors.red.shade700;
                            Color textColor = Colors.red.shade900;
                            String bannerTitle = 'Peringatan HRD';

                            if (warningData is Map) {
                              final type =
                                  (warningData['warning_type'] ??
                                          warningData['WarningType'] ??
                                          '-')
                                      .toString()
                                      .toLowerCase();
                              final desc =
                                  warningData['issue_description'] ??
                                  warningData['IssueDescription'] ??
                                  '';

                              switch (type) {
                                case 'cs_followup':
                                  displayType = 'Level 1: Tindak Lanjut CS';
                                  bgColor = Colors.blue.shade50;
                                  borderColor = Colors.blue.shade200;
                                  iconColor = Colors.blue.shade700;
                                  textColor = Colors.blue.shade900;
                                  bannerTitle = 'Informasi Tindak Lanjut';
                                  break;
                                case 'spv_followup':
                                  displayType = 'Level 2: Tindak Lanjut SPV';
                                  bgColor = Colors.orange.shade50;
                                  borderColor = Colors.orange.shade200;
                                  iconColor = Colors.orange.shade700;
                                  textColor = Colors.orange.shade900;
                                  bannerTitle = 'Informasi Tindak Lanjut';
                                  break;
                                case 'teguran':
                                  displayType = 'Teguran';
                                  break;
                                case 'sp1':
                                  displayType = 'SP 1';
                                  break;
                                case 'sp2':
                                  displayType = 'SP 2';
                                  break;
                                case 'sp3':
                                  displayType = 'SP 3';
                                  break;
                                default:
                                  displayType = type.toUpperCase();
                              }
                              displayMsg =
                                  "Jenis: $displayType\nCatatan: $desc";
                            }

                            return Container(
                              width: double.infinity,
                              margin: const EdgeInsets.only(bottom: 24),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 16,
                              ),
                              decoration: BoxDecoration(
                                color: bgColor,
                                borderRadius: BorderRadius.circular(16),
                                border: Border.all(color: borderColor),
                              ),
                              child: Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Icon(
                                    Icons.warning_amber_rounded,
                                    color: iconColor,
                                    size: 28,
                                  ),
                                  const SizedBox(width: 12),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          bannerTitle,
                                          style: GoogleFonts.plusJakartaSans(
                                            fontSize: 14,
                                            fontWeight: FontWeight.w800,
                                            color: textColor,
                                          ),
                                        ),
                                        const SizedBox(height: 4),
                                        Text(
                                          displayMsg,
                                          style: GoogleFonts.plusJakartaSans(
                                            fontSize: 12,
                                            fontWeight: FontWeight.w600,
                                            color: textColor.withOpacity(0.9),
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                              ),
                            );
                          },
                        ),
                      ],
                      // --- End Banner ---
                      if (errorMessage != null) ...[
                        Container(
                          width: double.infinity,
                          padding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 12,
                          ),
                          decoration: BoxDecoration(
                            color: AppColors.error.withOpacity(0.08),
                            borderRadius: BorderRadius.circular(16),
                            border: Border.all(
                              color: AppColors.error.withOpacity(0.2),
                            ),
                          ),
                          child: Row(
                            children: [
                              const Icon(
                                Icons.sync_problem_rounded,
                                color: AppColors.error,
                                size: 22,
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      'Gagal menyegarkan data terbaru',
                                      style: GoogleFonts.plusJakartaSans(
                                        fontSize: 12,
                                        fontWeight: FontWeight.w800,
                                        color: AppColors.error,
                                      ),
                                    ),
                                    const SizedBox(height: 2),
                                    Text(
                                      errorMessage.isNotEmpty
                                          ? 'Menampilkan data terakhir. ($errorMessage)'
                                          : 'Menampilkan data terakhir.',
                                      style: GoogleFonts.plusJakartaSans(
                                        fontSize: 10,
                                        fontWeight: FontWeight.w600,
                                        color: AppColors.textSecondary,
                                      ),
                                      maxLines: 2,
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ],
                                ),
                              ),
                              const SizedBox(width: 8),
                              ElevatedButton.icon(
                                onPressed: () {
                                  context.read<AttendanceBloc>().add(
                                    FetchHomeDataRequested(),
                                  );
                                },
                                icon: const Icon(
                                  Icons.refresh_rounded,
                                  size: 14,
                                  color: Colors.white,
                                ),
                                label: Text(
                                  'Coba Lagi',
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 10,
                                    fontWeight: FontWeight.w800,
                                    color: Colors.white,
                                  ),
                                ),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: AppColors.error,
                                  foregroundColor: Colors.white,
                                  elevation: 0,
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 12,
                                    vertical: 8,
                                  ),
                                  minimumSize: Size.zero,
                                  tapTargetSize:
                                      MaterialTapTargetSize.shrinkWrap,
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(10),
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 24),
                      ],
                      _buildStatistics(statistics),
                      const SizedBox(height: 32),
                      _buildSalesWorkingButton(
                        context,
                        state.attendanceStatus,
                        profile,
                      ),
                      _buildQuickActions(
                        context,
                        state.attendanceStatus,
                        recentActivity,
                      ),
                      const SizedBox(height: 40),
                      _buildAnnouncements(context),
                      const SizedBox(height: 40),
                      _buildRecentActivity(recentActivity),
                    ],
                  ),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildSalesWorkingButton(
    BuildContext context,
    AttendanceStatus status,
    Map<String, dynamic>? profile,
  ) {
    final jobTitle =
        (profile?['job_position']?['title'] ??
                profile?['JobPosition']?['Title'] ??
                '')
            .toString()
            .toLowerCase();
    final userMap = profile?['user'] ?? profile?['User'] ?? {};
    final userRole = (userMap['role'] ?? userMap['Role'] ?? '')
        .toString()
        .toLowerCase();
    final isDelivery =
        jobTitle.contains('pengiriman') ||
        jobTitle.contains('delivery') ||
        jobTitle.contains('kurir') ||
        jobTitle.contains('driver') ||
        jobTitle.contains('helper') ||
        jobTitle.contains('kernet');
    final isWarehouse =
        jobTitle.contains('gudang') ||
        jobTitle.contains('warehouse') ||
        jobTitle.contains('logistik');
    final isHR =
        userRole == 'hr_admin' ||
        jobTitle.contains('hr') ||
        jobTitle.contains('human resource') ||
        jobTitle.contains('personalia') ||
        jobTitle.contains('recruiter') ||
        ((userRole == 'manager' || jobTitle.contains('manager')) &&
            !jobTitle.contains('sales') &&
            !jobTitle.contains('marketing') &&
            !jobTitle.contains('gudang') &&
            !jobTitle.contains('warehouse') &&
            !jobTitle.contains('direktur') &&
            !jobTitle.contains('director'));
    final isDirector =
        !isHR &&
        (userRole == 'direktur' ||
            jobTitle.contains('direktur') ||
            jobTitle.contains('director') ||
            userRole == 'owner' ||
            userRole == 'superadmin');
    final isCashier =
        jobTitle.contains('kasir') ||
        jobTitle.contains('cashier') ||
        jobTitle.contains('finance') ||
        jobTitle.contains('keuangan');
    final isOperatorSpanduk =
        jobTitle.contains('operator') || jobTitle.contains('spanduk');
    final isSalesMT =
        jobTitle.contains('modern trade') ||
        jobTitle.contains('sales mt') ||
        (jobTitle.contains('sales') && jobTitle.contains(' mt')) ||
        jobTitle == 'mt' ||
        jobTitle.split(' ').contains('mt');
    final isSales =
        isSalesMT ||
        jobTitle.contains('sales') ||
        jobTitle.contains('marketing') ||
        jobTitle.contains('salesman') ||
        jobTitle.contains('salesperson') ||
        jobTitle.contains('motoris') ||
        jobTitle.contains('booster') ||
        jobTitle.contains('canvas') ||
        jobTitle.contains('task order') ||
        jobTitle.split(' ').contains('to');

    final isSupervisorSales =
        (jobTitle.contains('supervisor') || jobTitle.contains('spv')) &&
        (jobTitle.contains('sales') || jobTitle.contains('marketing'));
    final isManagerMarketing =
        jobTitle.contains('manager') &&
        (jobTitle.contains('sales') || jobTitle.contains('marketing'));

    // Only show operational/management working buttons to authorized roles
    if (!isWarehouse &&
        !isDelivery &&
        !isSales &&
        !isHR &&
        !isCashier &&
        !isOperatorSpanduk &&
        !isDirector) {
      return const SizedBox.shrink();
    }

    // Jika role adalah Sales MT murni, tampilkan HANYA tombol MT (violet)
    if (isSalesMT &&
        !isDelivery &&
        !isWarehouse &&
        !isHR &&
        !isCashier &&
        !isOperatorSpanduk &&
        !isDirector) {
      return _buildMtSalesButton(context, profile);
    }

    final themeColor = isOperatorSpanduk
        ? AppColors.primaryGreen
        : (isWarehouse
              ? AppColors.blueWarehouse
              : (isDelivery
                    ? AppColors.blueDelivery
                    : (isCashier
                          ? Colors.blueAccent
                          : (isHR
                                ? AppColors.blueHR
                                : (isDirector
                                      ? const Color(
                                          0xFF10B981,
                                        ) // Emerald for Director
                                      : AppColors.blueSales)))));

    final roleName = isOperatorSpanduk
        ? 'Operator Spanduk'
        : (isWarehouse
              ? 'Gudang'
              : (isDelivery
                    ? 'Pengiriman'
                    : (isCashier
                          ? 'Kasir'
                          : (isSupervisorSales
                                ? 'Supervisor Sales'
                                : (isHR
                                      ? 'HR'
                                      : (isDirector
                                            ? 'DIREKTUR'
                                            : 'Sales'))))));

    final btnText = roleName == 'DIREKTUR'
        ? 'DASHBOARD DIREKTUR'
        : 'MULAI BEKERJA SEBAGAI $roleName';

    String routeName;
    if (isOperatorSpanduk) {
      routeName = '/operator_spanduk_dashboard';
    } else if (isWarehouse) {
      routeName = '/warehouse_dashboard';
    } else if (isDelivery) {
      routeName = '/delivery_dashboard';
    } else if (isCashier) {
      routeName = '/cashier_dashboard';
    } else if (isSupervisorSales) {
      routeName = '/supervisor_dashboard';
    } else if (isHR) {
      routeName = '/hr_dashboard';
    } else if (isDirector) {
      routeName = '/director_dashboard';
    } else {
      routeName = '/sales_dashboard';
    }

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: double.infinity,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [themeColor, themeColor.withOpacity(0.85)],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(24),
            boxShadow: [
              BoxShadow(
                color: themeColor.withOpacity(0.3),
                blurRadius: 20,
                offset: const Offset(0, 10),
              ),
            ],
          ),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: () async {
                final prefs = await SharedPreferences.getInstance();
                await prefs.setString('last_dashboard_route', routeName);

                if (context.mounted) {
                  Object? arguments;
                  if (routeName == '/supervisor_dashboard') {
                    final branchId =
                        (profile?['branch_id'] ?? profile?['BranchID'])
                            ?.toString() ??
                        'ALL_BRANCHES';
                    final branchName =
                        (profile?['branch']?['name'] ??
                                profile?['Branch']?['Name'] ??
                                profile?['branch_name'] ??
                                profile?['BranchName'])
                            ?.toString() ??
                        '';
                    arguments = {
                      'branchId': branchId,
                      'branchName': branchName,
                    };
                  }
                  Navigator.of(
                    context,
                  ).pushNamed(routeName, arguments: arguments).then((_) async {
                    final p = await SharedPreferences.getInstance();
                    await p.remove('last_dashboard_route');
                  });
                }
              },
              borderRadius: BorderRadius.circular(24),
              child: Padding(
                padding: const EdgeInsets.symmetric(
                  vertical: 24,
                  horizontal: 24,
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.2),
                        shape: BoxShape.circle,
                      ),
                      child: Icon(
                        isWarehouse
                            ? Icons.inventory_2_rounded
                            : (isDelivery
                                  ? Icons.local_shipping_rounded
                                  : (isHR
                                        ? Icons.admin_panel_settings_rounded
                                        : (isDirector
                                              ? Icons.insights_rounded
                                              : (isCashier
                                                    ? Icons
                                                          .point_of_sale_rounded
                                                    : Icons
                                                          .storefront_rounded)))),
                        color: Colors.white,
                        size: 28,
                      ),
                    ),
                    const SizedBox(width: 20),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            isHR
                                ? 'MANAJEMEN'.toUpperCase()
                                : 'OPERASIONAL'.toUpperCase(),
                            style: GoogleFonts.outfit(
                              color: Colors.white.withOpacity(0.7),
                              fontWeight: FontWeight.w800,
                              fontSize: 10,
                              letterSpacing: 1.5,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            btnText.toUpperCase(),
                            style: GoogleFonts.outfit(
                              color: Colors.white,
                              fontWeight: FontWeight.w900,
                              fontSize: 16,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const Icon(
                      Icons.arrow_forward_ios_rounded,
                      color: Colors.white70,
                      size: 18,
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
        if (isSupervisorSales || isDirector || isManagerMarketing) ...[
          const SizedBox(height: 16),
          Container(
            width: double.infinity,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  AppColors.blueSales,
                  AppColors.blueSales.withOpacity(0.85),
                ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(24),
              boxShadow: [
                BoxShadow(
                  color: AppColors.blueSales.withOpacity(0.3),
                  blurRadius: 20,
                  offset: const Offset(0, 10),
                ),
              ],
            ),
            child: Material(
              color: Colors.transparent,
              child: InkWell(
                onTap: () async {
                  final prefs = await SharedPreferences.getInstance();
                  await prefs.setString(
                    'last_dashboard_route',
                    '/sales_dashboard',
                  );

                  if (context.mounted) {
                    Navigator.of(context)
                        .pushNamed(
                          '/sales_dashboard',
                          arguments: {'forceTO': true},
                        )
                        .then((_) async {
                          final p = await SharedPreferences.getInstance();
                          await p.remove('last_dashboard_route');
                        });
                  }
                },
                borderRadius: BorderRadius.circular(24),
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    vertical: 24,
                    horizontal: 24,
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: Colors.white.withOpacity(0.2),
                          shape: BoxShape.circle,
                        ),
                        child: const Icon(
                          Icons.storefront_rounded,
                          color: Colors.white,
                          size: 28,
                        ),
                      ),
                      const SizedBox(width: 20),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'OPERASIONAL LAPANGAN',
                              style: GoogleFonts.outfit(
                                color: Colors.white.withOpacity(0.7),
                                fontWeight: FontWeight.w800,
                                fontSize: 10,
                                letterSpacing: 1.5,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              'MASUK SEBAGAI SALES',
                              style: GoogleFonts.outfit(
                                color: Colors.white,
                                fontWeight: FontWeight.w900,
                                fontSize: 16,
                                letterSpacing: 0.5,
                              ),
                            ),
                          ],
                        ),
                      ),
                      const Icon(
                        Icons.arrow_forward_ios_rounded,
                        color: Colors.white70,
                        size: 18,
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ],
        const SizedBox(height: 32),
      ],
    );
  }

  /// Tombol khusus untuk Sales MT — warna violet, routing ke SalesDashboardPage(isMT: true)
  Widget _buildMtSalesButton(
    BuildContext context,
    Map<String, dynamic>? profile,
  ) {
    const Color mtColor = Color(0xFF10B981); // Emerald untuk MT
    const String routeName = '/mt_sales_dashboard';

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: double.infinity,
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: [Color(0xFF059669), Color(0xFF10B981)],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(24),
            boxShadow: [
              BoxShadow(
                color: mtColor.withOpacity(0.35),
                blurRadius: 20,
                offset: const Offset(0, 10),
              ),
            ],
          ),
          child: Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: () async {
                final prefs = await SharedPreferences.getInstance();
                await prefs.setString('last_dashboard_route', routeName);

                if (context.mounted) {
                  Navigator.of(context).pushNamed(routeName).then((_) async {
                    final p = await SharedPreferences.getInstance();
                    await p.remove('last_dashboard_route');
                  });
                }
              },
              borderRadius: BorderRadius.circular(24),
              child: Padding(
                padding: const EdgeInsets.symmetric(
                  vertical: 24,
                  horizontal: 24,
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.white.withOpacity(0.2),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.business_center_rounded,
                        color: Colors.white,
                        size: 28,
                      ),
                    ),
                    const SizedBox(width: 20),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'MODERN TRADE',
                            style: GoogleFonts.outfit(
                              color: Colors.white.withOpacity(0.7),
                              fontWeight: FontWeight.w800,
                              fontSize: 10,
                              letterSpacing: 1.5,
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            'MASUK SEBAGAI SALES MT',
                            style: GoogleFonts.outfit(
                              color: Colors.white,
                              fontWeight: FontWeight.w900,
                              fontSize: 16,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const Icon(
                      Icons.arrow_forward_ios_rounded,
                      color: Colors.white70,
                      size: 18,
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
        const SizedBox(height: 32),
      ],
    );
  }

  Widget _buildSliverAppBar(
    BuildContext context,
    Map<String, dynamic>? profile,
    String displayName,
  ) {
    return SliverAppBar(
      expandedHeight: 60,
      floating: false,
      pinned: true,
      backgroundColor: AppColors.primaryGreen,
      elevation: 0,
      iconTheme: const IconThemeData(color: Colors.white),
      flexibleSpace: FlexibleSpaceBar(
        background: Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: AppColors.primaryGradient,
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
        ),
      ),
      title: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            'Selamat Datang,',
            style: GoogleFonts.plusJakartaSans(
              color: Colors.white.withOpacity(0.8),
              fontWeight: FontWeight.w700,
              fontSize: 10,
              letterSpacing: 1,
            ),
          ),
          const SizedBox(height: 1),
          FittedBox(
            fit: BoxFit.scaleDown,
            alignment: Alignment.centerLeft,
            child: Text(
              (displayName.isNotEmpty ? displayName : 'User').toUpperCase(),
              style: GoogleFonts.plusJakartaSans(
                color: Colors.white,
                fontWeight: FontWeight.w900,
                fontSize: 13,
                letterSpacing: 0.5,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
      actions: [
        IconButton(
          padding: const EdgeInsets.all(6),
          constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
          visualDensity: VisualDensity.compact,
          tooltip: 'Segarkan Data',
          icon: const Icon(
            Icons.refresh_rounded,
            color: Colors.white,
            size: 20,
          ),
          onPressed: () {
            context.read<AttendanceBloc>().add(FetchHomeDataRequested());
          },
        ),
        // Tombol Akses Cepat Baca Ulang Pengumuman
        BlocBuilder<AnnouncementBloc, AnnouncementState>(
          builder: (context, state) {
            final hasAnnouncements = state.announcements.isNotEmpty;
            return Stack(
              alignment: Alignment.center,
              children: [
                IconButton(
                  padding: const EdgeInsets.all(6),
                  constraints: const BoxConstraints(
                    minWidth: 36,
                    minHeight: 36,
                  ),
                  visualDensity: VisualDensity.compact,
                  tooltip: 'Pengumuman Perusahaan',
                  icon: const Icon(
                    Icons.campaign_outlined,
                    color: Colors.white,
                    size: 22,
                  ),
                  onPressed: () => _showAnnouncementsBottomSheet(context),
                ),
                if (hasAnnouncements)
                  Positioned(
                    top: 8,
                    right: 8,
                    child: Container(
                      width: 7,
                      height: 7,
                      decoration: const BoxDecoration(
                        color: Colors.amberAccent,
                        shape: BoxShape.circle,
                      ),
                    ),
                  ),
              ],
            );
          },
        ),
        BlocBuilder<NotificationBloc, NotificationState>(
          builder: (context, state) {
            final hasNotifications =
                state is NotificationLoaded && state.notifications.isNotEmpty;
            return Stack(
              alignment: Alignment.center,
              children: [
                IconButton(
                  padding: const EdgeInsets.all(6),
                  constraints: const BoxConstraints(
                    minWidth: 36,
                    minHeight: 36,
                  ),
                  visualDensity: VisualDensity.compact,
                  tooltip: 'Notifikasi',
                  icon: const Icon(
                    Icons.notifications_none_rounded,
                    color: Colors.white,
                    size: 22,
                  ),
                  onPressed: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (context) => const NotificationListPage(),
                      ),
                    );
                  },
                ),
                if (hasNotifications)
                  Positioned(
                    top: 8,
                    right: 8,
                    child: Container(
                      width: 7,
                      height: 7,
                      decoration: const BoxDecoration(
                        color: Colors.yellow,
                        shape: BoxShape.circle,
                        border: Border.fromBorderSide(
                          BorderSide(color: AppColors.primaryGreen, width: 1),
                        ),
                      ),
                    ),
                  ),
              ],
            );
          },
        ),
        const SizedBox(width: 6),
      ],
    );
  }

  Widget _buildStatistics(Map<String, dynamic>? stats) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'RINGKASAN KEHADIRAN ANDA',
          style: GoogleFonts.plusJakartaSans(
            fontSize: 12,
            fontWeight: FontWeight.w800,
            color: AppColors.textTertiary,
            letterSpacing: 1.5,
          ),
        ),
        const SizedBox(height: 16),
        Container(
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: AppColors.primaryGradient,
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(32),
            boxShadow: [
              BoxShadow(
                color: AppColors.primaryGreen.withOpacity(0.3),
                blurRadius: 20,
                offset: const Offset(0, 10),
              ),
            ],
          ),
          child: Row(
            children: [
              Expanded(
                child: _buildRedStatItem(
                  'HADIR',
                  stats?['PresentCount']?.toString() ?? '0',
                  Icons.check_circle_rounded,
                ),
              ),
              Container(width: 1, height: 40, color: Colors.white24),
              Expanded(
                child: _buildRedStatItem(
                  'CUTI',
                  stats?['LeaveCount']?.toString() ?? '0',
                  Icons.event_available_rounded,
                ),
              ),
              Container(width: 1, height: 40, color: Colors.white24),
              Expanded(
                child: _buildRedStatItem(
                  'ALFA',
                  stats?['AlphaCount']?.toString() ?? '0',
                  Icons.error_rounded,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildRedStatItem(String label, String value, IconData icon) {
    return Column(
      children: [
        Icon(icon, color: Colors.white.withOpacity(0.8), size: 18),
        const SizedBox(height: 8),
        Text(
          value,
          style: GoogleFonts.plusJakartaSans(
            fontSize: 24,
            fontWeight: FontWeight.w900,
            color: Colors.white,
          ),
        ),
        Text(
          label,
          style: GoogleFonts.plusJakartaSans(
            fontSize: 9,
            fontWeight: FontWeight.w800,
            color: Colors.white70,
            letterSpacing: 1,
          ),
        ),
      ],
    );
  }

  void _showClockInModeDialog(
    BuildContext context, {
    bool isClockIn = true,
    Attendance? activeSession,
  }) {
    final TextEditingController destCtrl = TextEditingController(
      text: (activeSession != null && activeSession.isOutOfTown)
          ? activeSession.destination
          : '',
    );
    bool isOutOfTown = (activeSession != null && activeSession.isOutOfTown);

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setModalState) => Container(
          padding: EdgeInsets.only(
            left: 24,
            right: 24,
            top: 24,
            bottom: MediaQuery.of(ctx).viewInsets.bottom + 24,
          ),
          decoration: const BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Center(
                child: Container(
                  width: 40,
                  height: 4,
                  decoration: BoxDecoration(
                    color: Colors.grey.shade300,
                    borderRadius: BorderRadius.circular(2),
                  ),
                ),
              ),
              const SizedBox(height: 20),
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: isClockIn
                          ? AppColors.primaryGreen.withOpacity(0.12)
                          : AppColors.primaryRed.withOpacity(0.12),
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      isClockIn ? Icons.login_rounded : Icons.logout_rounded,
                      color: isClockIn
                          ? AppColors.primaryGreen
                          : AppColors.primaryRed,
                      size: 24,
                    ),
                  ),
                  const SizedBox(width: 14),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          isClockIn
                              ? 'PILIH MODE ABSEN MASUK'
                              : 'PILIH MODE ABSEN KELUAR',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 16,
                            fontWeight: FontWeight.w900,
                            color: AppColors.textPrimary,
                            letterSpacing: 0.5,
                          ),
                        ),
                        Text(
                          'Pilih jenis kehadiran hari ini',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 12,
                            color: AppColors.textSecondary,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              // Pilihan 1: Kantor / Reguler
              InkWell(
                onTap: () {
                  setModalState(() {
                    isOutOfTown = false;
                  });
                },
                borderRadius: BorderRadius.circular(16),
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: !isOutOfTown
                        ? AppColors.primaryRed.withOpacity(0.06)
                        : Colors.grey.shade50,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: !isOutOfTown
                          ? AppColors.primaryRed
                          : Colors.grey.shade200,
                      width: !isOutOfTown ? 2 : 1,
                    ),
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: !isOutOfTown
                              ? AppColors.primaryRed.withOpacity(0.15)
                              : Colors.grey.shade200,
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          Icons.business_rounded,
                          color: !isOutOfTown
                              ? AppColors.primaryRed
                              : AppColors.textTertiary,
                          size: 22,
                        ),
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Kantor / Reguler (WFO)',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 14,
                                fontWeight: FontWeight.w800,
                                color: !isOutOfTown
                                    ? AppColors.primaryRed
                                    : AppColors.textPrimary,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              'Presensi reguler di lokasi kantor cabang.',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 11,
                                color: AppColors.textSecondary,
                              ),
                            ),
                          ],
                        ),
                      ),
                      if (!isOutOfTown)
                        const Icon(
                          Icons.check_circle_rounded,
                          color: AppColors.primaryRed,
                          size: 22,
                        ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 12),
              // Pilihan 2: Dinas Luar Kota
              InkWell(
                onTap: () {
                  setModalState(() {
                    isOutOfTown = true;
                  });
                },
                borderRadius: BorderRadius.circular(16),
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: isOutOfTown
                        ? const Color(0xFF4F46E5).withOpacity(0.06)
                        : Colors.grey.shade50,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: isOutOfTown
                          ? const Color(0xFF4F46E5)
                          : Colors.grey.shade200,
                      width: isOutOfTown ? 2 : 1,
                    ),
                  ),
                  child: Row(
                    children: [
                      Container(
                        padding: const EdgeInsets.all(10),
                        decoration: BoxDecoration(
                          color: isOutOfTown
                              ? const Color(0xFF4F46E5).withOpacity(0.15)
                              : Colors.grey.shade200,
                          shape: BoxShape.circle,
                        ),
                        child: Icon(
                          Icons.flight_takeoff_rounded,
                          color: isOutOfTown
                              ? const Color(0xFF4F46E5)
                              : AppColors.textTertiary,
                          size: 22,
                        ),
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              'Dinas Luar Kota',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 14,
                                fontWeight: FontWeight.w800,
                                color: isOutOfTown
                                    ? const Color(0xFF4F46E5)
                                    : AppColors.textPrimary,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              'Bebas radius lokasi kantor untuk tugas luar.',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 11,
                                color: AppColors.textSecondary,
                              ),
                            ),
                          ],
                        ),
                      ),
                      if (isOutOfTown)
                        const Icon(
                          Icons.check_circle_rounded,
                          color: Color(0xFF4F46E5),
                          size: 22,
                        ),
                    ],
                  ),
                ),
              ),
              if (isOutOfTown) ...[
                const SizedBox(height: 16),
                Text(
                  'Kota / Tempat Tujuan Dinas *',
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 12,
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF4F46E5),
                  ),
                ),
                const SizedBox(height: 8),
                TextField(
                  controller: destCtrl,
                  decoration: InputDecoration(
                    hintText: 'Contoh: Surabaya - Kunjungan Proyek',
                    hintStyle: GoogleFonts.plusJakartaSans(
                      fontSize: 13,
                      color: AppColors.textTertiary,
                    ),
                    isDense: true,
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 14,
                      vertical: 12,
                    ),
                    filled: true,
                    fillColor: const Color(0xFF4F46E5).withOpacity(0.05),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                      borderSide: BorderSide.none,
                    ),
                  ),
                  style: GoogleFonts.plusJakartaSans(
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                height: 52,
                child: Container(
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(16),
                    gradient: LinearGradient(
                      colors: isOutOfTown
                          ? [const Color(0xFF4F46E5), const Color(0xFF6366F1)]
                          : AppColors.primaryGradient,
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                  ),
                  child: ElevatedButton(
                    onPressed: () {
                      final dest = destCtrl.text.trim();
                      if (isOutOfTown && dest.isEmpty) {
                        DialogUtils.showError(
                          context: ctx,
                          title: 'Tujuan Wajib Diisi',
                          message:
                              'Harap masukkan kota atau tempat tujuan dinas luar kota.',
                        );
                        return;
                      }
                      Navigator.of(ctx).pop();
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (_) => FaceVerificationPage(
                            isClockIn: isClockIn,
                            isOutOfTown: isOutOfTown,
                            destination: isOutOfTown ? dest : '',
                          ),
                        ),
                      );
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: Colors.transparent,
                      shadowColor: Colors.transparent,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(16),
                      ),
                    ),
                    child: Text(
                      'LANJUTKAN KE VERIFIKASI WAJAH',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 13,
                        fontWeight: FontWeight.w900,
                        color: Colors.white,
                        letterSpacing: 0.5,
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildQuickActions(
    BuildContext context,
    AttendanceStatus status,
    List<Attendance>? recentActivity,
  ) {
    Attendance? lastAttendance;
    if (recentActivity != null && recentActivity.isNotEmpty) {
      lastAttendance = recentActivity.first;
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'AKSES CEPAT',
          style: GoogleFonts.plusJakartaSans(
            fontSize: 12,
            fontWeight: FontWeight.w800,
            color: AppColors.textTertiary,
            letterSpacing: 1.5,
          ),
        ),
        const SizedBox(height: 20),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            _buildActionItem(
              Icons.qr_code_scanner_rounded,
              'Absen Masuk',
              AppColors.success,
              () {
                if (status == AttendanceStatus.clockedIn ||
                    status == AttendanceStatus.completed) {
                  DialogUtils.showError(
                    context: context,
                    title: 'Sudah Absen',
                    message: 'Sudah absen masuk hari ini',
                  );
                } else {
                  _showClockInModeDialog(context, isClockIn: true);
                }
              },
              isDisabled:
                  status == AttendanceStatus.clockedIn ||
                  status == AttendanceStatus.completed,
            ),
            _buildActionItem(
              Icons.logout_rounded,
              'Absen Keluar',
              AppColors.primaryGreen,
              () {
                if (status == AttendanceStatus.none) {
                  DialogUtils.showError(
                    context: context,
                    title: 'Belum Absen',
                    message: 'Anda belum melakukan absen masuk hari ini',
                  );
                } else if (status == AttendanceStatus.completed) {
                  DialogUtils.showError(
                    context: context,
                    title: 'Sudah Absen',
                    message: 'Anda sudah melakukan absen keluar hari ini',
                  );
                } else {
                  _showClockInModeDialog(
                    context,
                    isClockIn: false,
                    activeSession: lastAttendance,
                  );
                }
              },
              isDisabled:
                  status == AttendanceStatus.none ||
                  status == AttendanceStatus.completed,
            ),
            _buildActionItem(
              Icons.history_rounded,
              'Riwayat',
              AppColors.warning,
              () {
                Navigator.of(context)
                    .push(
                      MaterialPageRoute(
                        builder: (_) => const ActivityHistoryPage(),
                      ),
                    )
                    .then((_) {
                      context.read<AttendanceBloc>().add(
                        FetchHomeDataRequested(),
                      );
                    });
              },
            ),
            _buildActionItem(
              Icons.event_available_rounded,
              'Cuti',
              const Color(0xFF6366F1),
              () => Navigator.of(
                context,
              ).push(MaterialPageRoute(builder: (_) => const LeavePage())),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildActionItem(
    IconData icon,
    String label,
    Color color,
    VoidCallback onTap, {
    bool isDisabled = false,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20),
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(20),
            decoration: BoxDecoration(
              color: (isDisabled
                  ? Colors.grey.shade200
                  : color.withOpacity(0.12)),
              borderRadius: BorderRadius.circular(24),
            ),
            child: Icon(
              icon,
              color: (isDisabled ? Colors.grey.shade400 : color),
              size: 28,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            label,
            style: GoogleFonts.plusJakartaSans(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: (isDisabled
                  ? AppColors.textTertiary
                  : AppColors.textPrimary),
            ),
          ),
        ],
      ),
    );
  }

  void _showAnnouncementsBottomSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (bottomSheetContext) {
        return SizedBox(
          height: MediaQuery.of(context).size.height * 0.75,
          child: Container(
            decoration: const BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
            ),
            padding: const EdgeInsets.fromLTRB(20, 16, 20, 24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Center(
                  child: Container(
                    width: 40,
                    height: 4,
                    decoration: BoxDecoration(
                      color: Colors.grey.shade300,
                      borderRadius: BorderRadius.circular(2),
                    ),
                  ),
                ),
                const SizedBox(height: 16),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: AppColors.primaryGreen.withOpacity(0.1),
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(
                            Icons.campaign_rounded,
                            color: AppColors.primaryGreen,
                            size: 20,
                          ),
                        ),
                        const SizedBox(width: 10),
                        Text(
                          'PENGUMUMAN PERUSAHAAN',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 14,
                            fontWeight: FontWeight.w800,
                            color: AppColors.textPrimary,
                            letterSpacing: 0.5,
                          ),
                        ),
                      ],
                    ),
                    IconButton(
                      icon: const Icon(Icons.close_rounded, color: Colors.grey),
                      onPressed: () => Navigator.of(bottomSheetContext).pop(),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Expanded(
                  child: BlocBuilder<AnnouncementBloc, AnnouncementState>(
                    builder: (context, state) {
                      if (state.isLoading && state.announcements.isEmpty) {
                        return const Center(
                          child: CircularProgressIndicator(
                            color: AppColors.primaryGreen,
                          ),
                        );
                      }
                      if (state.announcements.isNotEmpty) {
                        return ListView.separated(
                          itemCount: state.announcements.length,
                          separatorBuilder: (_, __) =>
                              const SizedBox(height: 12),
                          itemBuilder: (context, index) {
                            final ann = state.announcements[index];
                            final isUrgent =
                                ann.category ==
                                    AnnouncementCategory.urgentCall ||
                                ann.category == AnnouncementCategory.alert;
                            return Container(
                              decoration: BoxDecoration(
                                color: Colors.grey.shade50,
                                borderRadius: BorderRadius.circular(18),
                                border: Border.all(
                                  color: isUrgent
                                      ? Colors.red.shade200
                                      : Colors.grey.shade200,
                                ),
                              ),
                              padding: const EdgeInsets.all(16),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Row(
                                    children: [
                                      if (isUrgent)
                                        Container(
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 8,
                                            vertical: 2,
                                          ),
                                          margin: const EdgeInsets.only(
                                            right: 8,
                                          ),
                                          decoration: BoxDecoration(
                                            color: Colors.red.shade100,
                                            borderRadius: BorderRadius.circular(
                                              8,
                                            ),
                                          ),
                                          child: Text(
                                            'PENTING',
                                            style: GoogleFonts.plusJakartaSans(
                                              fontSize: 9,
                                              fontWeight: FontWeight.w800,
                                              color: Colors.red.shade800,
                                            ),
                                          ),
                                        ),
                                      Expanded(
                                        child: Text(
                                          ann.title,
                                          style: GoogleFonts.plusJakartaSans(
                                            fontSize: 14,
                                            fontWeight: FontWeight.w800,
                                            color: AppColors.textPrimary,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 6),
                                  Text(
                                    ann.content,
                                    style: GoogleFonts.plusJakartaSans(
                                      fontSize: 12,
                                      height: 1.4,
                                      color: AppColors.textSecondary,
                                    ),
                                    maxLines: 2,
                                    overflow: TextOverflow.ellipsis,
                                  ),
                                  const SizedBox(height: 12),
                                  Row(
                                    mainAxisAlignment:
                                        MainAxisAlignment.spaceBetween,
                                    children: [
                                      Text(
                                        DateFormat(
                                          'dd MMM yyyy, HH:mm',
                                        ).format(ann.createdAt),
                                        style: GoogleFonts.plusJakartaSans(
                                          fontSize: 10,
                                          color: AppColors.textTertiary,
                                          fontWeight: FontWeight.w600,
                                        ),
                                      ),
                                      InkWell(
                                        onTap: () {
                                          Navigator.of(
                                            bottomSheetContext,
                                          ).pop();
                                          Navigator.of(context).push(
                                            MaterialPageRoute(
                                              builder: (_) =>
                                                  AnnouncementDetailPage(
                                                    announcement: ann,
                                                  ),
                                            ),
                                          );
                                        },
                                        borderRadius: BorderRadius.circular(12),
                                        child: Container(
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 12,
                                            vertical: 7,
                                          ),
                                          decoration: BoxDecoration(
                                            color: isUrgent
                                                ? Colors.red.shade600
                                                : AppColors.primaryGreen,
                                            borderRadius: BorderRadius.circular(
                                              12,
                                            ),
                                          ),
                                          child: Row(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              const Icon(
                                                Icons.menu_book_rounded,
                                                size: 13,
                                                color: Colors.white,
                                              ),
                                              const SizedBox(width: 5),
                                              Text(
                                                'BACA ULANG',
                                                style:
                                                    GoogleFonts.plusJakartaSans(
                                                      fontSize: 10,
                                                      fontWeight:
                                                          FontWeight.w800,
                                                      color: Colors.white,
                                                    ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                ],
                              ),
                            );
                          },
                        );
                      }
                      return Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.campaign_outlined,
                              size: 48,
                              color: Colors.grey.shade400,
                            ),
                            const SizedBox(height: 12),
                            Text(
                              'Belum Ada Pengumuman',
                              style: GoogleFonts.plusJakartaSans(
                                fontSize: 14,
                                fontWeight: FontWeight.w700,
                                color: AppColors.textSecondary,
                              ),
                            ),
                          ],
                        ),
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildAnnouncements(BuildContext context) {
    return BlocBuilder<AnnouncementBloc, AnnouncementState>(
      builder: (context, state) {
        return Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Text(
                      'KABAR & PENGUMUMAN',
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 12,
                        fontWeight: FontWeight.w800,
                        color: AppColors.textTertiary,
                        letterSpacing: 1.5,
                      ),
                    ),
                  ],
                ),
                TextButton(
                  onPressed: () => _showAnnouncementsBottomSheet(context),
                  child: Text(
                    'Lihat Semua',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 11,
                      fontWeight: FontWeight.w800,
                      color: AppColors.primaryGreen,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            if (state.isLoading && state.announcements.isEmpty)
              const Center(
                child: CircularProgressIndicator(color: AppColors.primaryGreen),
              )
            else if (state.announcements.isNotEmpty)
              SizedBox(
                height: 200,
                child: ListView.separated(
                  scrollDirection: Axis.horizontal,
                  itemCount: state.announcements.length,
                  separatorBuilder: (_, __) => const SizedBox(width: 14),
                  itemBuilder: (context, index) {
                    final ann = state.announcements[index];
                    final rawMedia =
                        (ann.imageUrl ??
                                (ann.mediaUrls.isNotEmpty
                                    ? ann.mediaUrls.first
                                    : ''))
                            .toLowerCase();
                    final isVideo =
                        ann.mediaType.toLowerCase() == 'video' ||
                        ann.mediaType.toLowerCase() == 'video_url' ||
                        rawMedia.endsWith('.mp4') ||
                        rawMedia.endsWith('.mov') ||
                        rawMedia.endsWith('.webm') ||
                        rawMedia.endsWith('.m4v') ||
                        rawMedia.contains('drive.google.com') ||
                        rawMedia.contains('/file/d/');

                    final hasImage =
                        !isVideo &&
                        ((ann.imageUrl != null && ann.imageUrl!.isNotEmpty) ||
                            ann.mediaUrls.isNotEmpty);
                    final displayImg = hasImage
                        ? (ann.imageUrl ?? ann.mediaUrls.first)
                        : null;

                    return GestureDetector(
                      onTap: () {
                        Navigator.of(context).push(
                          MaterialPageRoute(
                            builder: (_) =>
                                AnnouncementDetailPage(announcement: ann),
                          ),
                        );
                      },
                      child: Container(
                        width: 310,
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(24),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.04),
                              blurRadius: 16,
                              offset: const Offset(0, 6),
                            ),
                          ],
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            if (isVideo)
                              ClipRRect(
                                borderRadius: const BorderRadius.only(
                                  topLeft: Radius.circular(24),
                                  topRight: Radius.circular(24),
                                ),
                                child: Container(
                                  height: 85,
                                  width: double.infinity,
                                  decoration: BoxDecoration(
                                    gradient: LinearGradient(
                                      colors: [
                                        Colors.indigo.shade900,
                                        const Color(0xFF1E1B4B),
                                      ],
                                      begin: Alignment.topLeft,
                                      end: Alignment.bottomRight,
                                    ),
                                  ),
                                  child: Center(
                                    child: Container(
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 14,
                                        vertical: 6,
                                      ),
                                      decoration: BoxDecoration(
                                        color: Colors.white.withOpacity(0.15),
                                        borderRadius: BorderRadius.circular(20),
                                        border: Border.all(
                                          color: Colors.white24,
                                        ),
                                      ),
                                      child: Row(
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          const Icon(
                                            Icons.play_circle_fill_rounded,
                                            color: Colors.white,
                                            size: 18,
                                          ),
                                          const SizedBox(width: 6),
                                          Text(
                                            'Lampiran Video',
                                            style: GoogleFonts.plusJakartaSans(
                                              color: Colors.white,
                                              fontSize: 11,
                                              fontWeight: FontWeight.w700,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                  ),
                                ),
                              )
                            else if (hasImage && displayImg != null)
                              ClipRRect(
                                borderRadius: const BorderRadius.only(
                                  topLeft: Radius.circular(24),
                                  topRight: Radius.circular(24),
                                ),
                                child: CachedNetworkImage(
                                  imageUrl: displayImg,
                                  height: 85,
                                  width: double.infinity,
                                  fit: BoxFit.cover,
                                  placeholder: (_, __) => Container(
                                    height: 85,
                                    color: Colors.grey.shade100,
                                  ),
                                  errorWidget: (_, __, ___) => const SizedBox(),
                                ),
                              ),
                            Expanded(
                              child: Padding(
                                padding: const EdgeInsets.all(14),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      children: [
                                        if (!hasImage)
                                          Container(
                                            padding: const EdgeInsets.all(6),
                                            margin: const EdgeInsets.only(
                                              right: 8,
                                            ),
                                            decoration: BoxDecoration(
                                              color: AppColors.primaryGreen
                                                  .withOpacity(0.1),
                                              shape: BoxShape.circle,
                                            ),
                                            child: const Icon(
                                              Icons.campaign_rounded,
                                              color: AppColors.primaryGreen,
                                              size: 16,
                                            ),
                                          ),
                                        Expanded(
                                          child: Text(
                                            ann.title,
                                            style: GoogleFonts.plusJakartaSans(
                                              fontWeight: FontWeight.w800,
                                              fontSize: 13,
                                              color: AppColors.textPrimary,
                                            ),
                                            maxLines: 1,
                                            overflow: TextOverflow.ellipsis,
                                          ),
                                        ),
                                      ],
                                    ),
                                    const SizedBox(height: 4),
                                    Expanded(
                                      child: Text(
                                        ann.content,
                                        style: GoogleFonts.plusJakartaSans(
                                          fontSize: 11,
                                          color: AppColors.textSecondary,
                                          height: 1.4,
                                        ),
                                        maxLines: hasImage ? 1 : 2,
                                        overflow: TextOverflow.ellipsis,
                                      ),
                                    ),
                                    const SizedBox(height: 6),
                                    Row(
                                      mainAxisAlignment:
                                          MainAxisAlignment.spaceBetween,
                                      children: [
                                        Row(
                                          children: [
                                            Text(
                                              DateFormat(
                                                'dd MMM yyyy',
                                              ).format(ann.createdAt),
                                              style:
                                                  GoogleFonts.plusJakartaSans(
                                                    fontSize: 10,
                                                    color:
                                                        AppColors.textTertiary,
                                                    fontWeight: FontWeight.w700,
                                                  ),
                                            ),
                                            if (ann.likesCount > 0) ...[
                                              const SizedBox(width: 8),
                                              Icon(
                                                Icons.favorite_rounded,
                                                size: 11,
                                                color: Colors.red.shade400,
                                              ),
                                              const SizedBox(width: 3),
                                              Text(
                                                '${ann.likesCount}',
                                                style:
                                                    GoogleFonts.plusJakartaSans(
                                                      fontSize: 10,
                                                      color:
                                                          Colors.red.shade400,
                                                      fontWeight:
                                                          FontWeight.w800,
                                                    ),
                                              ),
                                            ],
                                          ],
                                        ),
                                        Row(
                                          children: [
                                            Text(
                                              'Buka',
                                              style:
                                                  GoogleFonts.plusJakartaSans(
                                                    fontSize: 10,
                                                    fontWeight: FontWeight.w800,
                                                    color:
                                                        AppColors.primaryGreen,
                                                  ),
                                            ),
                                            const SizedBox(width: 2),
                                            Icon(
                                              Icons.arrow_forward_rounded,
                                              size: 13,
                                              color: AppColors.primaryGreen
                                                  .withOpacity(0.8),
                                            ),
                                          ],
                                        ),
                                      ],
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
              )
            else
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(24),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(28),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.02),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Center(
                  child: Text(
                    'Belum ada pengumuman terbaru.',
                    style: GoogleFonts.plusJakartaSans(
                      fontSize: 12,
                      color: AppColors.textTertiary,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ),
          ],
        );
      },
    );
  }

  Widget _buildRecentActivity(List<Attendance> activity) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'AKTIVITAS TERBARU',
          style: GoogleFonts.plusJakartaSans(
            fontSize: 12,
            fontWeight: FontWeight.w800,
            color: AppColors.textTertiary,
            letterSpacing: 1.5,
          ),
        ),
        const SizedBox(height: 20),
        if (activity.isEmpty)
          _buildEmptyPlaceholder(
            Icons.history_rounded,
            'Belum ada riwayat absensi',
          )
        else
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: activity.length > 5 ? 5 : activity.length,
            itemBuilder: (context, index) {
              final att = activity[index];
              final isAlfa = att.status == 'ALFA';

              int hours = att.workDuration ~/ 60;
              int minutes = att.workDuration % 60;
              String durationStr = hours > 0
                  ? '${hours}j ${minutes}m'
                  : '${minutes}m';
              if (att.workDuration == 0 && att.checkOut == null)
                durationStr = '--';
              if (isAlfa) durationStr = '0m';

              final dayStr = DateFormat(
                'EEE',
                'id_ID',
              ).format(att.checkIn).toUpperCase();
              final dateStr = DateFormat('dd').format(att.checkIn);
              final monthStr = DateFormat(
                'MMM',
                'id_ID',
              ).format(att.checkIn).toUpperCase();
              final timeStr = isAlfa
                  ? '--:--'
                  : '${DateFormat('HH:mm').format(att.checkIn)} - ${att.checkOut != null ? DateFormat('HH:mm').format(att.checkOut!) : '--:--'}';

              return Container(
                margin: const EdgeInsets.only(bottom: 16),
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(28),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withOpacity(0.02),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 14,
                        vertical: 10,
                      ),
                      decoration: BoxDecoration(
                        color: AppColors.grayLight.withOpacity(0.5),
                        borderRadius: BorderRadius.circular(16),
                      ),
                      child: Column(
                        children: [
                          Text(
                            dayStr,
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 10,
                              fontWeight: FontWeight.w800,
                              color: AppColors.textTertiary,
                            ),
                          ),
                          Text(
                            dateStr,
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 20,
                              fontWeight: FontWeight.w900,
                              color: AppColors.textPrimary,
                            ),
                          ),
                          Text(
                            monthStr,
                            style: GoogleFonts.plusJakartaSans(
                              fontSize: 10,
                              fontWeight: FontWeight.w800,
                              color: AppColors.textTertiary,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 20),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 8,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color:
                                      (isAlfa
                                              ? AppColors.warning
                                              : AppColors.success)
                                          .withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: Text(
                                  isAlfa ? 'ALFA' : 'HADIR',
                                  style: GoogleFonts.plusJakartaSans(
                                    fontSize: 9,
                                    fontWeight: FontWeight.w800,
                                    color: isAlfa
                                        ? AppColors.warning
                                        : AppColors.success,
                                    letterSpacing: 0.5,
                                  ),
                                ),
                              ),
                            ],
                          ),
                          const SizedBox(height: 12),
                          Row(
                            children: [
                              const Icon(
                                Icons.access_time_filled_rounded,
                                size: 14,
                                color: AppColors.textTertiary,
                              ),
                              const SizedBox(width: 8),
                              Text(
                                timeStr,
                                style: GoogleFonts.plusJakartaSans(
                                  fontSize: 14,
                                  color: AppColors.textPrimary,
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                            ],
                          ),
                        ],
                      ),
                    ),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          durationStr,
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 16,
                            fontWeight: FontWeight.w900,
                            color: AppColors.textPrimary,
                          ),
                        ),
                        Text(
                          'DURASI',
                          style: GoogleFonts.plusJakartaSans(
                            fontSize: 9,
                            fontWeight: FontWeight.w800,
                            color: AppColors.textTertiary,
                            letterSpacing: 1,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              );
            },
          ),
      ],
    );
  }

  Widget _buildEmptyPlaceholder(IconData icon, String message) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(32),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(28),
        border: Border.all(color: AppColors.grayBorder, width: 1),
      ),
      child: Column(
        children: [
          Icon(icon, color: AppColors.textTertiary.withOpacity(0.2), size: 40),
          const SizedBox(height: 16),
          Text(
            message,
            style: GoogleFonts.plusJakartaSans(
              fontSize: 13,
              color: AppColors.textTertiary,
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}
