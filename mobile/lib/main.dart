import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:hris_app/features/auth/presentation/bloc/auth_bloc.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_event.dart';
import 'package:hris_app/features/auth/presentation/bloc/auth_state.dart';
import 'package:hris_app/features/attendance/presentation/bloc/attendance_bloc.dart';
import 'package:hris_app/features/leave/presentation/bloc/leave_bloc.dart';
import 'package:hris_app/features/profile/presentation/bloc/profile_bloc.dart';
import 'package:hris_app/features/home/presentation/pages/home_page.dart';
import 'package:hris_app/features/auth/presentation/pages/login_page.dart';
import 'package:hris_app/features/auth/presentation/pages/splash_screen.dart';
import 'package:hris_app/core/theme/app_theme.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/services/notification_service.dart';
import 'package:hris_app/core/services/heartbeat_service.dart';
import 'package:hris_app/features/announcement/presentation/bloc/announcement_bloc.dart';
import 'package:hris_app/features/notification/presentation/bloc/notification_bloc.dart';
import 'package:hris_app/features/sync/presentation/bloc/sync_bloc.dart';
import 'package:hris_app/features/sales/presentation/pages/sales_dashboard_page.dart';
import 'package:hris_app/features/sales/presentation/pages/delivery_dashboard_page.dart';
import 'package:hris_app/features/sales/presentation/pages/operator_spanduk_dashboard_page.dart';
import 'package:hris_app/features/sales/presentation/pages/visit_schedule_page.dart';
import 'package:hris_app/features/sales/presentation/pages/director_dashboard_page.dart';
import 'package:hris_app/features/inventory/presentation/pages/warehouse_dashboard_page.dart';
import 'package:hris_app/features/home/presentation/pages/hr_dashboard_page.dart';
import 'package:hris_app/features/finance/presentation/pages/cashier_dashboard_page.dart';
import 'package:hris_app/features/supervisor/presentation/pages/supervisor_dashboard_page.dart';
import 'package:hris_app/features/supervisor/presentation/bloc/supervisor_bloc.dart';
import 'package:hris_app/features/approval/presentation/bloc/approval_bloc.dart';
import 'package:hris_app/core/utils/dialog_utils.dart';
import 'package:hris_app/core/services/data_integrity_service.dart';
import 'package:hris_app/core/services/app_update_service.dart';
import 'package:hris_app/core/services/live_location_tracking_service.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_bloc.dart';
import 'package:hris_app/core/widgets/offline_badge.dart';

void main() {
  runZonedGuarded(
    () async {
      WidgetsFlutterBinding.ensureInitialized();

      // --- GLOBAL ERROR HANDLER ---
      FlutterError.onError = (FlutterErrorDetails details) {
        FlutterError.presentError(details);
      };

      ErrorWidget.builder = (FlutterErrorDetails details) {
        return MaterialApp(
          debugShowCheckedModeBanner: false,
          home: Scaffold(
            body: Center(
              child: Padding(
                padding: const EdgeInsets.all(24.0),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(
                      Icons.error_outline_rounded,
                      color: Colors.redAccent,
                      size: 60,
                    ),
                    const SizedBox(height: 16),
                    const Text(
                      'Terjadi Kesalahan',
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: Colors.black87,
                      ),
                    ),
                    const SizedBox(height: 8),
                    const Text(
                      'Gagal memuat tampilan layar ini.\nSilakan mulai ulang aplikasi atau periksa kembali koneksi internet Anda.',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.black54, fontSize: 13),
                    ),
                    const SizedBox(height: 24),
                    ElevatedButton.icon(
                      onPressed: () {
                        // Ini akan me-restart main logic di beberapa skenario
                        // Tapi karena stateless, biasanya pengguna cukup menekan back/kill app
                      },
                      icon: const Icon(Icons.info_outline, size: 18),
                      label: const Text('Silakan Muat Ulang Aplikasi'),
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      };

      PlatformDispatcher.instance.onError = (error, stack) {
        debugPrint('Unhandled Exception: $error');
        return true;
      };
      // -----------------------------

      await initializeDateFormatting('id_ID', null);
      await di.init();
      // Initialize notifications & notification channels first
      final notificationService = di.sl<NotificationService>();
      await notificationService.init();

      // Configure background location tracking service
      await LiveLocationTrackingService.configure();

      // Initialize heartbeat
      final heartbeatService = di.sl<HeartbeatService>();
      heartbeatService.init();

      if (kReleaseMode) {
        debugPrint = (String? message, {int? wrapWidth}) {};
      }

      // Jalankan integrity check awal saat app pertama kali dibuka
      // Ini mendeteksi jika ada data pending yang tertinggal dari sesi sebelumnya
      unawaited(
        DataIntegrityService().log(
          category: 'APP',
          action: 'APP_STARTED',
          detail: 'Aplikasi dimulai pada ${DateTime.now().toIso8601String()}',
        ),
      );

      runApp(const MyApp());
    },
    (error, stackTrace) {
      if (kDebugMode) {
        debugPrint('Caught error: $error');
        debugPrint(stackTrace.toString());
      }
    },
    zoneSpecification: ZoneSpecification(
      print: (Zone self, ZoneDelegate parent, Zone zone, String line) {
        if (!kReleaseMode) {
          parent.print(zone, line);
        }
      },
    ),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    final authBloc = di.sl<AuthBloc>()..add(CheckAuthStatusRequested());

    // Listen for global 401 Unauthorized events
    di.sl<ApiClient>().onUnauthorized.listen((msg) {
      authBloc.add(SessionExpired(message: msg));
    });

    return MultiBlocProvider(
      providers: [
        BlocProvider(create: (_) => authBloc),
        BlocProvider(create: (_) => di.sl<AttendanceBloc>()),
        BlocProvider(create: (_) => di.sl<LeaveBloc>()),
        BlocProvider(create: (_) => di.sl<ProfileBloc>()),
        BlocProvider(
          create: (_) =>
              di.sl<AnnouncementBloc>()..add(FetchAnnouncementsRequested()),
        ),
        BlocProvider(
          create: (_) =>
              di.sl<NotificationBloc>()..add(FetchNotificationsRequested()),
        ),
        BlocProvider(create: (_) => di.sl<SyncBloc>()),
        BlocProvider(create: (_) => di.sl<EpkwtBloc>()),
      ],
      child: MaterialApp(
        title: 'Karsa HR',
        debugShowCheckedModeBanner: false,
        theme: AppTheme.lightTheme,
        navigatorObservers: [visitScheduleRouteObserver],
        builder: (context, child) {
          return OfflineBadge(child: child ?? const SizedBox());
        },
        home: const AppUpdateWrapper(child: AuthWrapper()),
        routes: {
          '/director_dashboard': (context) => const DirectorDashboardPage(),
          '/sales_dashboard': (context) {
            final args = ModalRoute.of(context)?.settings.arguments;
            bool forceTO = false;
            if (args is Map<String, dynamic>) {
              forceTO = args['forceTO'] as bool? ?? false;
            }
            return SalesDashboardPage(forceTO: forceTO);
          },
          '/mt_sales_dashboard': (context) =>
              const SalesDashboardPage(isMT: true),
          '/delivery_dashboard': (context) => const DeliveryDashboardPage(),
          '/warehouse_dashboard': (context) => const WarehouseDashboardPage(),
          '/cashier_dashboard': (context) => const CashierDashboardPage(),
          '/operator_spanduk_dashboard': (context) =>
              const OperatorSpandukDashboardPage(),
          '/hr_dashboard': (context) => BlocProvider(
            create: (_) => di.sl<ApprovalBloc>(),
            child: const HrDashboardPage(),
          ),
          '/supervisor_dashboard': (context) {
            final args = ModalRoute.of(context)?.settings.arguments;
            String branchId = '0';
            String branchName = '';

            if (args is String) {
              branchId = args;
            } else if (args is Map<String, dynamic>) {
              branchId = args['branchId'] as String? ?? '0';
              branchName = args['branchName'] as String? ?? '';
            }

            return BlocProvider(
              create: (_) => di.sl<SupervisorBloc>(),
              child: SupervisorDashboardPage(
                branchId: branchId,
                branchName: branchName,
              ),
            );
          },
        },
      ),
    );
  }
}

class AuthWrapper extends StatelessWidget {
  const AuthWrapper({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocListener<AuthBloc, AuthState>(
      listenWhen: (previous, current) =>
          current is AuthError ||
          (previous is Authenticated && current is Unauthenticated),
      listener: (context, state) {
        if (state is AuthError) {
          DialogUtils.showError(
            context: context,
            title: 'Gagal',
            message: state.message,
          );
        }

        if (state is Unauthenticated) {
          // Clear all routes above the base AuthWrapper when user logs out
          Navigator.of(context).popUntil((route) => route.isFirst);
          if (state.sessionMessage != null &&
              state.sessionMessage!.isNotEmpty) {
            DialogUtils.showError(
              context: context,
              title: 'Sesi Berakhir',
              message: state.sessionMessage!,
            );
          }
        }
      },
      child: BlocBuilder<AuthBloc, AuthState>(
        builder: (context, state) {
          if (state is Authenticated ||
              state is ChangePasswordSuccess ||
              (state is AuthError && state.isAuthenticated)) {
            return const HomePage();
          }

          if (state is Unauthenticated ||
              (state is AuthError && !state.isAuthenticated) ||
              state is RegisterSuccess ||
              state is FaceRegistrationSuccess ||
              state is ForgotPasswordSuccess) {
            return const LoginPage(key: ValueKey('login_page'));
          }

          if (state is AuthInitial) {
            return const SplashScreen();
          }

          // Fallback for AuthLoading or other transient states
          return const Scaffold(
            body: Center(child: CircularProgressIndicator()),
          );
        },
      ),
    );
  }
}

/// Widget yang memeriksa pembaruan aplikasi dari Play Store
/// segera setelah frame pertama dirender.
///
/// Ditempatkan sebagai root dari aplikasi agar pengecekan terjadi
/// pada setiap kali aplikasi dibuka (cold start).
class AppUpdateWrapper extends StatefulWidget {
  final Widget child;

  const AppUpdateWrapper({super.key, required this.child});

  @override
  State<AppUpdateWrapper> createState() => _AppUpdateWrapperState();
}

class _AppUpdateWrapperState extends State<AppUpdateWrapper> {
  @override
  void initState() {
    super.initState();
    // Jalankan pengecekan update setelah frame pertama selesai dirender
    // agar context sudah valid dan ScaffoldMessenger tersedia.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        AppUpdateService.instance.checkAndUpdate(context);
      }
    });
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
