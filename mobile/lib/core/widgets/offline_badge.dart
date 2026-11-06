import 'dart:async';
import 'package:flutter/material.dart';
import 'package:internet_connection_checker_plus/internet_connection_checker_plus.dart';
import 'package:hris_app/core/network/network_info.dart';

class OfflineBadge extends StatefulWidget {
  final Widget child;

  const OfflineBadge({super.key, required this.child});

  @override
  State<OfflineBadge> createState() => _OfflineBadgeState();
}

class _OfflineBadgeState extends State<OfflineBadge> {
  bool _isOffline = false;
  late StreamSubscription<InternetStatus> _connectivitySubscription;
  Timer? _debounceTimer;

  @override
  void initState() {
    super.initState();
    _checkConnectivity();
    _connectivitySubscription =
        NetworkInfo().onStatusChange.listen((status) {
      _updateOfflineStatus(status);
    });
  }

  Future<void> _checkConnectivity() async {
    final hasInternet = await NetworkInfo().hasInternetAccess;
    _updateOfflineStatus(
        hasInternet ? InternetStatus.connected : InternetStatus.disconnected);
  }

  void _updateOfflineStatus(InternetStatus status) {
    final isOffline = status == InternetStatus.disconnected;
    _debounceTimer?.cancel();
    if (isOffline) {
      // Debounce 2.5s so brief mobile network jitter doesn't falsely flash offline mode
      _debounceTimer = Timer(const Duration(milliseconds: 2500), () {
        if (mounted && !_isOffline) {
          setState(() => _isOffline = true);
        }
      });
    } else {
      if (_isOffline) {
        if (mounted) {
          setState(() => _isOffline = false);
        }
      }
    }
  }

  @override
  void dispose() {
    _debounceTimer?.cancel();
    _connectivitySubscription.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        widget.child,
        if (_isOffline)
          Positioned(
            top: 0,
            left: 0,
            right: 0,
            child: SafeArea(
              bottom: false,
              child: Material(
                color: Colors.transparent,
                child: Container(
                  padding:
                      const EdgeInsets.symmetric(vertical: 4, horizontal: 8),
                  color: Colors.yellow.shade700,
                  child: const Text(
                    'MODE OFFLINE',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      color: Colors.black,
                      fontWeight: FontWeight.bold,
                      fontSize: 12,
                    ),
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }
}
