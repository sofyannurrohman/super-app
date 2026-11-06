import 'dart:math' show cos, sqrt, asin;
import 'dart:ui' as ui;
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:geolocator/geolocator.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:latlong2/latlong.dart';
import 'package:hris_app/features/sales/data/models/visit_plan_model.dart';
import 'package:url_launcher/url_launcher.dart';

class OptimalRouteMapPage extends StatefulWidget {
  final List<VisitPlanModel> plans;

  const OptimalRouteMapPage({super.key, required this.plans});

  @override
  State<OptimalRouteMapPage> createState() => _OptimalRouteMapPageState();
}

class _OptimalRouteMapPageState extends State<OptimalRouteMapPage> {
  bool _isLoading = true;
  String? _loadingMessage;
  LatLng? _currentLocation;
  List<VisitPlanModel> _orderedPlans = [];

  /// Full road-snapped polyline from OSRM
  List<LatLng> _roadPolyline = [];

  /// Straight-line fallback (used when OSRM fails)
  List<LatLng> _fallbackPolyline = [];

  bool _usingRoadRoute = false;
  final MapController _mapController = MapController();
  final _dio = Dio(BaseOptions(
    connectTimeout: const Duration(seconds: 30),
    receiveTimeout: const Duration(seconds: 30),
  ));

  @override
  void initState() {
    super.initState();
    _calculateRoute();
  }

  // ─── Nearest Neighbor TSP ────────────────────────────────────────────────

  Future<void> _calculateRoute() async {
    setState(() {
      _isLoading = true;
      _loadingMessage = 'Mendapatkan lokasi GPS…';
      _usingRoadRoute = false;
      _roadPolyline = [];
      _fallbackPolyline = [];
    });

    try {
      // 1. GPS
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled)
        throw Exception('GPS tidak aktif. Mohon aktifkan GPS Anda.');

      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied)
          throw Exception('Izin lokasi ditolak');
      }
      if (permission == LocationPermission.deniedForever) {
        throw Exception(
            'Izin lokasi ditolak permanen. Buka pengaturan untuk mengaktifkan.');
      }

      Position position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
      );
      _currentLocation = LatLng(position.latitude, position.longitude);

      // 2. Separate completed vs remaining
      final plansWithCoords =
          widget.plans.where((p) => p.hasCoordinates).toList();
      final completed = plansWithCoords
          .where((p) => p.status == 'COMPLETED' || p.status == 'EXTRA')
          .toList();
      final remaining = plansWithCoords
          .where((p) => p.status != 'COMPLETED' && p.status != 'EXTRA')
          .toList();

      // 3. Nearest Neighbor on remaining stores
      List<VisitPlanModel> orderedRemaining = [];
      List<VisitPlanModel> unvisited = List.from(remaining);
      LatLng curPos = _currentLocation!;

      while (unvisited.isNotEmpty) {
        VisitPlanModel? nearest;
        double minDist = double.infinity;
        for (var p in unvisited) {
          final d = _haversine(curPos, LatLng(p.latitude!, p.longitude!));
          if (d < minDist) {
            minDist = d;
            nearest = p;
          }
        }
        if (nearest == null) break;
        orderedRemaining.add(nearest);
        curPos = LatLng(nearest.latitude!, nearest.longitude!);
        unvisited.remove(nearest);
      }

      // Straight-line fallback waypoints
      final waypoints = <LatLng>[_currentLocation!];
      for (var p in orderedRemaining) {
        waypoints.add(LatLng(p.latitude!, p.longitude!));
      }

      setState(() {
        _orderedPlans = [...orderedRemaining, ...completed];
        _fallbackPolyline = waypoints;
        _loadingMessage = 'Mengambil rute jalan dari OSRM…';
      });

      // 4. Fetch road-snapped route from OSRM
      if (waypoints.length >= 2) {
        await _fetchOsrmRoute(waypoints);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: Colors.red),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  // ─── OSRM Road Routing ───────────────────────────────────────────────────

  Future<void> _fetchOsrmRoute(List<LatLng> waypoints) async {
    try {
      // Request GeoJSON geometry — plain [lon, lat] arrays, no encoding/decoding needed.
      final coords =
          waypoints.map((p) => '${p.longitude},${p.latitude}').join(';');
      final url = 'https://router.project-osrm.org/route/v1/driving/$coords'
          '?overview=full&geometries=geojson&steps=false';

      final response = await _dio.get(url);

      if (response.statusCode == 200 &&
          response.data['code'] == 'Ok' &&
          response.data['routes'] != null &&
          (response.data['routes'] as List).isNotEmpty) {
        // geometry = { "type": "LineString", "coordinates": [[lon, lat], ...] }
        final geometry = response.data['routes'][0]['geometry'];
        final coordinates = geometry['coordinates'] as List;

        final points = coordinates
            .map<LatLng>((c) {
              final lon = (c[0] as num).toDouble();
              final lat = (c[1] as num).toDouble();
              return LatLng(lat, lon); // GeoJSON order is [lon, lat]
            })
            .where(
              (p) =>
                  p.latitude >= -90 &&
                  p.latitude <= 90 &&
                  p.longitude >= -180 &&
                  p.longitude <= 180,
            )
            .toList();

        if (points.isNotEmpty && mounted) {
          setState(() {
            _roadPolyline = points;
            _usingRoadRoute = true;
          });
        }
      }
    } catch (e) {
      // OSRM failed — silently fall back to straight lines
      debugPrint('OSRM route fetch failed: $e');
      if (mounted) {
        setState(() => _usingRoadRoute = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              '⚠️ Rute jalan tidak tersedia, menampilkan garis lurus.',
              style: GoogleFonts.outfit(),
            ),
            backgroundColor: Colors.orange.shade700,
            behavior: SnackBarBehavior.floating,
            duration: const Duration(seconds: 3),
          ),
        );
      }
    }
  }

  // ─── Helpers ─────────────────────────────────────────────────────────────

  double _haversine(LatLng p1, LatLng p2) {
    var p = 0.017453292519943295;
    var a = 0.5 -
        cos((p2.latitude - p1.latitude) * p) / 2 +
        cos(p1.latitude * p) *
            cos(p2.latitude * p) *
            (1 - cos((p2.longitude - p1.longitude) * p)) /
            2;
    return 12742 * asin(sqrt(a));
  }

  Future<void> _openGoogleMapsNav(double lat, double lng, String name) async {
    final uri = Uri.parse(
      'https://www.google.com/maps/dir/?api=1&destination=$lat,$lng&travelmode=driving',
    );
    if (await canLaunchUrl(uri)) {
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } else {
      final fallback = Uri.parse('geo:$lat,$lng?q=$lat,$lng($name)');
      await launchUrl(fallback, mode: LaunchMode.externalApplication);
    }
  }

  void _showStoreBottomSheet(VisitPlanModel plan, int sequenceIndex) {
    final isCompleted = plan.status == 'COMPLETED' || plan.status == 'EXTRA';

    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(28)),
      ),
      isScrollControlled: true,
      builder: (context) => Padding(
        padding: const EdgeInsets.fromLTRB(24, 20, 24, 36),
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
                  borderRadius: BorderRadius.circular(4),
                ),
              ),
            ),
            const SizedBox(height: 20),
            Row(
              children: [
                Container(
                  width: 48,
                  height: 48,
                  decoration: BoxDecoration(
                    color: isCompleted
                        ? Colors.green.withOpacity(0.12)
                        : Colors.blueAccent.withOpacity(0.12),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    isCompleted
                        ? Icons.check_circle_rounded
                        : Icons.store_rounded,
                    color: isCompleted ? Colors.green : Colors.blueAccent,
                    size: 26,
                  ),
                ),
                const SizedBox(width: 14),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        plan.storeName,
                        style: GoogleFonts.outfit(
                          fontSize: 18,
                          fontWeight: FontWeight.w900,
                          color: const Color(0xFF1E293B),
                        ),
                      ),
                      const SizedBox(height: 2),
                      Text(
                        plan.address,
                        style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: Colors.blueGrey,
                          fontWeight: FontWeight.w600,
                        ),
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            if (sequenceIndex >= 0)
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
                decoration: BoxDecoration(
                  color: Colors.blueAccent.withOpacity(0.08),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  'Urutan ke-${sequenceIndex + 1} dalam rute optimal',
                  style: GoogleFonts.outfit(
                    fontSize: 11,
                    fontWeight: FontWeight.w700,
                    color: Colors.blueAccent,
                  ),
                ),
              ),
            const SizedBox(height: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                color: isCompleted
                    ? Colors.green.withOpacity(0.08)
                    : Colors.blueAccent.withOpacity(0.08),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Text(
                isCompleted ? '✓ SUDAH DIKUNJUNGI' : 'BELUM DIKUNJUNGI',
                style: GoogleFonts.outfit(
                  fontSize: 11,
                  fontWeight: FontWeight.w800,
                  color: isCompleted ? Colors.green : Colors.blueAccent,
                  letterSpacing: 0.5,
                ),
              ),
            ),
            const SizedBox(height: 24),
            if (!isCompleted)
              ElevatedButton.icon(
                onPressed: () {
                  Navigator.pop(context);
                  _openGoogleMapsNav(
                      plan.latitude!, plan.longitude!, plan.storeName);
                },
                icon: const Icon(Icons.navigation_rounded, color: Colors.white),
                label: Text(
                  'NAVIGASI KE TOKO',
                  style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900, color: Colors.white),
                ),
                style: ElevatedButton.styleFrom(
                  minimumSize: const Size(double.infinity, 56),
                  backgroundColor: Colors.blueAccent,
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16)),
                  elevation: 0,
                ),
              )
            else
              OutlinedButton.icon(
                onPressed: () {
                  Navigator.pop(context);
                  _openGoogleMapsNav(
                      plan.latitude!, plan.longitude!, plan.storeName);
                },
                icon:
                    const Icon(Icons.near_me_rounded, color: Colors.blueAccent),
                label: Text(
                  'BUKA DI GOOGLE MAPS',
                  style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900, color: Colors.blueAccent),
                ),
                style: OutlinedButton.styleFrom(
                  minimumSize: const Size(double.infinity, 56),
                  side: const BorderSide(color: Colors.blueAccent, width: 1.5),
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16)),
                ),
              ),
          ],
        ),
      ),
    );
  }

  // ─── Build ───────────────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    final remaining = _orderedPlans
        .where((p) => p.status != 'COMPLETED' && p.status != 'EXTRA')
        .toList();
    final completed = _orderedPlans
        .where((p) => p.status == 'COMPLETED' || p.status == 'EXTRA')
        .toList();

    // Use road polyline if available, otherwise straight-line fallback
    final activePolyline = _usingRoadRoute ? _roadPolyline : _fallbackPolyline;

    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded,
              color: Color(0xFF334155)),
          onPressed: () => Navigator.pop(context),
        ),
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'RUTE OPTIMAL',
              style: GoogleFonts.outfit(
                fontSize: 10,
                color: Colors.blueAccent,
                fontWeight: FontWeight.w800,
                letterSpacing: 1.5,
              ),
            ),
            Text(
              'Peta Kunjungan Hari Ini',
              style: GoogleFonts.outfit(
                fontSize: 17,
                fontWeight: FontWeight.w800,
                color: const Color(0xFF1E293B),
              ),
            ),
          ],
        ),
        actions: [
          IconButton(
            onPressed: _calculateRoute,
            icon: const Icon(Icons.refresh_rounded, color: Colors.blueAccent),
            tooltip: 'Hitung ulang rute',
          ),
        ],
      ),
      body: _isLoading
          ? Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const CircularProgressIndicator(color: Colors.blueAccent),
                  const SizedBox(height: 20),
                  Text(
                    _loadingMessage ?? 'Memuat…',
                    style: GoogleFonts.outfit(
                      color: Colors.blueGrey,
                      fontWeight: FontWeight.w600,
                      fontSize: 14,
                    ),
                  ),
                ],
              ),
            )
          : _currentLocation == null
              ? Center(
                  child: Padding(
                    padding: const EdgeInsets.all(32),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.location_off_rounded,
                            size: 64, color: Colors.grey.shade300),
                        const SizedBox(height: 16),
                        Text(
                          'Gagal mendapatkan lokasi',
                          style: GoogleFonts.outfit(
                              fontSize: 18, fontWeight: FontWeight.w800),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Pastikan GPS aktif dan izin lokasi diberikan.',
                          textAlign: TextAlign.center,
                          style: GoogleFonts.outfit(color: Colors.blueGrey),
                        ),
                        const SizedBox(height: 24),
                        ElevatedButton.icon(
                          onPressed: _calculateRoute,
                          icon: const Icon(Icons.gps_fixed_rounded,
                              color: Colors.white),
                          label: Text(
                            'Coba Lagi',
                            style: GoogleFonts.outfit(
                                color: Colors.white,
                                fontWeight: FontWeight.w800),
                          ),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.blueAccent,
                            shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(14)),
                          ),
                        ),
                      ],
                    ),
                  ),
                )
              : Stack(
                  children: [
                    // ── MAP ──────────────────────────────────────────────
                    FlutterMap(
                      mapController: _mapController,
                      options: MapOptions(
                        initialCenter: _currentLocation!,
                        initialZoom: 13.0,
                      ),
                      children: [
                        TileLayer(
                          urlTemplate:
                              'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                          userAgentPackageName: 'com.example.karsahrtrial',
                        ),

                        // Road-following polyline
                        if (activePolyline.length >= 2)
                          PolylineLayer(
                            polylines: [
                              // Outer shadow stroke
                              Polyline(
                                points: activePolyline,
                                strokeWidth: 7.0,
                                color: Colors.blueAccent.withOpacity(0.25),
                              ),
                              // Main route line
                              Polyline(
                                points: activePolyline,
                                strokeWidth: 4.5,
                                color: Colors.blueAccent,
                                strokeCap: StrokeCap.round,
                                strokeJoin: StrokeJoin.round,
                              ),
                            ],
                          ),

                        MarkerLayer(
                          markers: [
                            // ── Current location
                            Marker(
                              point: _currentLocation!,
                              width: 48,
                              height: 48,
                              child: Container(
                                decoration: BoxDecoration(
                                  color: Colors.white,
                                  shape: BoxShape.circle,
                                  boxShadow: [
                                    BoxShadow(
                                      color:
                                          Colors.blueAccent.withOpacity(0.35),
                                      blurRadius: 14,
                                      spreadRadius: 2,
                                    ),
                                  ],
                                ),
                                child: const Icon(Icons.my_location_rounded,
                                    color: Colors.blueAccent, size: 28),
                              ),
                            ),

                            // ── Remaining (unvisited) — numbered blue pins
                            for (int i = 0; i < remaining.length; i++)
                              Marker(
                                point: LatLng(remaining[i].latitude!,
                                    remaining[i].longitude!),
                                width: 64,
                                height: 72,
                                child: GestureDetector(
                                  onTap: () =>
                                      _showStoreBottomSheet(remaining[i], i),
                                  child: Column(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Container(
                                        width: 36,
                                        height: 36,
                                        decoration: BoxDecoration(
                                          color: Colors.blueAccent,
                                          shape: BoxShape.circle,
                                          boxShadow: [
                                            BoxShadow(
                                              color: Colors.blueAccent
                                                  .withOpacity(0.45),
                                              blurRadius: 10,
                                              offset: const Offset(0, 4),
                                            ),
                                          ],
                                        ),
                                        child: Center(
                                          child: Text(
                                            '${i + 1}',
                                            style: GoogleFonts.outfit(
                                              color: Colors.white,
                                              fontWeight: FontWeight.w900,
                                              fontSize: 15,
                                            ),
                                          ),
                                        ),
                                      ),
                                      CustomPaint(
                                        size: const Size(12, 8),
                                        painter:
                                            _TrianglePainter(Colors.blueAccent),
                                      ),
                                    ],
                                  ),
                                ),
                              ),

                            // ── Completed — green check pins
                            for (var plan in completed)
                              Marker(
                                point: LatLng(plan.latitude!, plan.longitude!),
                                width: 44,
                                height: 52,
                                child: GestureDetector(
                                  onTap: () => _showStoreBottomSheet(plan, -1),
                                  child: Column(
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      Container(
                                        width: 32,
                                        height: 32,
                                        decoration: BoxDecoration(
                                          color: Colors.green,
                                          shape: BoxShape.circle,
                                          boxShadow: [
                                            BoxShadow(
                                              color:
                                                  Colors.green.withOpacity(0.4),
                                              blurRadius: 8,
                                              offset: const Offset(0, 3),
                                            ),
                                          ],
                                        ),
                                        child: const Icon(Icons.check_rounded,
                                            color: Colors.white, size: 18),
                                      ),
                                      CustomPaint(
                                        size: const Size(10, 7),
                                        painter: _TrianglePainter(Colors.green),
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                          ],
                        ),
                      ],
                    ),

                    // ── Road route badge (top left) ───────────────────────
                    Positioned(
                      top: 12,
                      left: 12,
                      child: AnimatedOpacity(
                        opacity: 1.0,
                        duration: const Duration(milliseconds: 300),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 10, vertical: 6),
                          decoration: BoxDecoration(
                            color: _usingRoadRoute
                                ? Colors.green.shade600
                                : Colors.orange.shade700,
                            borderRadius: BorderRadius.circular(20),
                            boxShadow: [
                              BoxShadow(
                                color: Colors.black.withOpacity(0.15),
                                blurRadius: 8,
                              ),
                            ],
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(
                                _usingRoadRoute
                                    ? Icons.directions_rounded
                                    : Icons.warning_amber_rounded,
                                color: Colors.white,
                                size: 14,
                              ),
                              const SizedBox(width: 5),
                              Text(
                                _usingRoadRoute
                                    ? 'Rute Mengikuti Jalan'
                                    : 'Garis Lurus (Offline)',
                                style: GoogleFonts.outfit(
                                  color: Colors.white,
                                  fontSize: 11,
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),

                    // ── My location button (top right) ───────────────────
                    Positioned(
                      top: 12,
                      right: 12,
                      child: Material(
                        color: Colors.white,
                        elevation: 4,
                        borderRadius: BorderRadius.circular(14),
                        child: InkWell(
                          borderRadius: BorderRadius.circular(14),
                          onTap: () {
                            if (_currentLocation != null) {
                              _mapController.move(_currentLocation!, 14.0);
                            }
                          },
                          child: const Padding(
                            padding: EdgeInsets.all(10),
                            child: Icon(Icons.my_location_rounded,
                                color: Colors.blueAccent, size: 24),
                          ),
                        ),
                      ),
                    ),

                    // ── Bottom summary card ───────────────────────────────
                    Positioned(
                      bottom: 16,
                      left: 16,
                      right: 16,
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 20, vertical: 14),
                        decoration: BoxDecoration(
                          color: Colors.white,
                          borderRadius: BorderRadius.circular(20),
                          boxShadow: [
                            BoxShadow(
                              color: Colors.black.withOpacity(0.10),
                              blurRadius: 20,
                              offset: const Offset(0, 8),
                            ),
                          ],
                        ),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceAround,
                          children: [
                            _legendItem(Colors.blueAccent,
                                '${remaining.length}', 'Belum\nDikunjungi'),
                            Container(
                                width: 1,
                                height: 36,
                                color: Colors.grey.shade200),
                            _legendItem(Colors.green, '${completed.length}',
                                'Sudah\nDikunjungi'),
                            Container(
                                width: 1,
                                height: 36,
                                color: Colors.grey.shade200),
                            _legendItem(const Color(0xFF475569),
                                '${_orderedPlans.length}', 'Total\nToko'),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
    );
  }

  Widget _legendItem(Color color, String count, String label) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(count,
            style: GoogleFonts.outfit(
                fontSize: 22, fontWeight: FontWeight.w900, color: color)),
        Text(label,
            textAlign: TextAlign.center,
            style: GoogleFonts.outfit(
                fontSize: 10,
                fontWeight: FontWeight.w700,
                color: Colors.blueGrey)),
      ],
    );
  }
}

/// Tiny triangle "pin tip" painter
class _TrianglePainter extends CustomPainter {
  final Color color;
  _TrianglePainter(this.color);

  @override
  void paint(ui.Canvas canvas, ui.Size size) {
    final paint = ui.Paint()..color = color;
    final path = ui.Path()
      ..moveTo(0, 0)
      ..lineTo(size.width, 0)
      ..lineTo(size.width / 2, size.height)
      ..close();
    canvas.drawPath(path, paint);
  }

  @override
  bool shouldRepaint(_TrianglePainter oldDelegate) =>
      oldDelegate.color != color;
}
