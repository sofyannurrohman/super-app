import 'package:hris_app/core/network/api_client.dart';
import 'package:get_it/get_it.dart';
import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:geolocator/geolocator.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:latlong2/latlong.dart';
import 'package:dio/dio.dart';

class MapPickerPage extends StatefulWidget {
  final double? initialLatitude;
  final double? initialLongitude;

  const MapPickerPage({super.key, this.initialLatitude, this.initialLongitude});

  @override
  State<MapPickerPage> createState() => _MapPickerPageState();
}

class _MapPickerPageState extends State<MapPickerPage> {
  final MapController _mapController = MapController();
  LatLng _selectedLatLng = const LatLng(
    -7.250445,
    112.768845,
  ); // Surabaya default fallback
  bool _isLoadingMap = false;
  bool _isLoadingAddress = false;
  String _resolvedAddress = '';
  Timer? _debounceTimer;
  final TextEditingController _searchController = TextEditingController();
  bool _hasValidLocation = false;
  bool _isDragging = false;

  @override
  void initState() {
    super.initState();
    if (widget.initialLatitude != null && widget.initialLongitude != null) {
      _hasValidLocation = true;
      _selectedLatLng = LatLng(
        widget.initialLatitude!,
        widget.initialLongitude!,
      );
      _reverseGeocode(_selectedLatLng);
    } else {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _centerOnCurrentLocation();
      });
    }
  }

  @override
  void dispose() {
    _debounceTimer?.cancel();
    _searchController.dispose();
    _mapController.dispose();
    super.dispose();
  }

  Future<void> _centerOnCurrentLocation() async {
    setState(() => _isLoadingMap = true);
    try {
      bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
      if (!serviceEnabled) {
        throw Exception('Layanan GPS Anda belum aktif.');
      }

      LocationPermission permission = await Geolocator.checkPermission();
      if (permission == LocationPermission.denied) {
        permission = await Geolocator.requestPermission();
        if (permission == LocationPermission.denied) {
          throw Exception('Izin lokasi ditolak oleh pengguna.');
        }
      }

      if (permission == LocationPermission.deniedForever) {
        throw Exception('Izin lokasi ditolak secara permanen di pengaturan.');
      }

      Position position = await Geolocator.getCurrentPosition(
        desiredAccuracy: LocationAccuracy.high,
      );

      final currentLatLng = LatLng(position.latitude, position.longitude);
      _mapController.move(currentLatLng, 16.0);
      setState(() {
        _selectedLatLng = currentLatLng;
        _hasValidLocation = true;
      });
      _reverseGeocode(currentLatLng);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(e.toString()),
            backgroundColor: Colors.redAccent,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoadingMap = false);
    }
  }

  Future<void> _reverseGeocode(LatLng latLng) async {
    setState(() => _isLoadingAddress = true);
    try {
      final response = await GetIt.I<ApiClient>().client.get(
        'https://nominatim.openstreetmap.org/reverse',
        queryParameters: {
          'lat': latLng.latitude,
          'lon': latLng.longitude,
          'format': 'json',
        },
        options: Options(headers: {'User-Agent': 'KarsaHRMobile/1.0'}),
      );

      if (response.statusCode == 200 && response.data != null) {
        final data = response.data;
        final String displayName =
            data['display_name'] ?? 'Alamat tidak ditemukan';
        setState(() {
          _resolvedAddress = displayName;
        });
      } else {
        setState(() {
          _resolvedAddress = '';
        });
      }
    } catch (e) {
      setState(() {
        _resolvedAddress = '';
      });
    } finally {
      setState(() => _isLoadingAddress = false);
    }
  }

  Future<void> _searchAddress() async {
    final query = _searchController.text.trim();
    if (query.isEmpty) return;

    FocusScope.of(context).unfocus();
    setState(() => _isLoadingMap = true);
    try {
      final response = await GetIt.I<ApiClient>().client.get(
        'https://nominatim.openstreetmap.org/search',
        queryParameters: {'q': query, 'format': 'json', 'limit': 1},
        options: Options(headers: {'User-Agent': 'KarsaHRMobile/1.0'}),
      );

      if (response.statusCode == 200 &&
          response.data != null &&
          (response.data as List).isNotEmpty) {
        final firstResult = (response.data as List).first;
        final double lat = double.parse(firstResult['lat']);
        final double lon = double.parse(firstResult['lon']);
        final String displayName = firstResult['display_name'] ?? query;

        final targetLatLng = LatLng(lat, lon);
        _mapController.move(targetLatLng, 16.0);
        setState(() {
          _selectedLatLng = targetLatLng;
          _resolvedAddress = displayName;
          _hasValidLocation = true;
        });
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text(
                'Lokasi tidak ditemukan. Coba ketik lebih spesifik.',
              ),
              backgroundColor: Colors.amber,
            ),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Pencarian gagal: $e'),
            backgroundColor: Colors.redAccent,
          ),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoadingMap = false);
    }
  }

  void _onMapPositionChanged(MapCamera camera, bool hasGesture) {
    if (hasGesture) {
      // Cancel previous timer
      _debounceTimer?.cancel();

      setState(() {
        _isDragging = true;
      });

      // Debounce the state update completely to prevent calling setState() on every frame of dragging,
      // which was causing the map rendering engine to stutter and feel laggy.
      // Set to 2 seconds as requested, giving sales plenty of time to position the map.
      _debounceTimer = Timer(const Duration(seconds: 2), () {
        if (mounted) {
          setState(() {
            _selectedLatLng = camera.center;
            _isLoadingAddress = true;
            _hasValidLocation = true;
            _isDragging = false;
          });
          _reverseGeocode(camera.center);
        }
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF1F5F9),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(
            Icons.arrow_back_ios_rounded,
            color: Color(0xFF334155),
          ),
          onPressed: () => Navigator.pop(context),
        ),
        title: Container(
          height: 44,
          decoration: BoxDecoration(
            color: const Color(0xFFF1F5F9),
            borderRadius: BorderRadius.circular(12),
          ),
          child: TextField(
            controller: _searchController,
            textInputAction: TextInputAction.search,
            onSubmitted: (_) => _searchAddress(),
            style: GoogleFonts.outfit(
              fontSize: 14,
              color: const Color(0xFF1E293B),
            ),
            decoration: InputDecoration(
              hintText: 'Cari kota, jalan, atau toko...',
              hintStyle: GoogleFonts.outfit(
                color: Colors.blueGrey.shade400,
                fontSize: 13,
              ),
              border: InputBorder.none,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 12,
              ),
              suffixIcon: IconButton(
                icon: const Icon(
                  Icons.search_rounded,
                  color: Colors.blueAccent,
                  size: 20,
                ),
                onPressed: _searchAddress,
              ),
            ),
          ),
        ),
      ),
      body: Stack(
        children: [
          FlutterMap(
            mapController: _mapController,
            options: MapOptions(
              initialCenter: _selectedLatLng,
              initialZoom: 16.0,
              onPositionChanged: _onMapPositionChanged,
            ),
            children: [
              TileLayer(
                urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                userAgentPackageName: 'com.example.karsahrtrial',
              ),
            ],
          ),

          // Locked Center Pin Icon
          IgnorePointer(
            child: Align(
              alignment: Alignment.center,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(
                    Icons.location_on_rounded,
                    color: Colors.redAccent,
                    size: 44,
                  ),
                  Container(
                    width: 8,
                    height: 8,
                    decoration: BoxDecoration(
                      color: Colors.black.withOpacity(0.3),
                      shape: BoxShape.circle,
                    ),
                  ),
                  const SizedBox(height: 36),
                ],
              ),
            ),
          ),

          // Loading Map Overlay
          if (_isLoadingMap)
            Container(
              color: Colors.black.withOpacity(0.1),
              child: const Center(
                child: CircularProgressIndicator(color: Colors.blueAccent),
              ),
            ),

          // Floating Current Location GPS Button
          Positioned(
            right: 16,
            bottom: 260, // Elegantly floating above the bottom details panel
            child: Container(
              decoration: BoxDecoration(
                boxShadow: [
                  BoxShadow(
                    color: Colors.blueAccent.withOpacity(0.2),
                    blurRadius: 12,
                    offset: const Offset(0, 6),
                  ),
                ],
              ),
              child: FloatingActionButton.extended(
                onPressed: _centerOnCurrentLocation,
                backgroundColor: Colors.white,
                elevation: 0,
                heroTag: 'gps_focus_button',
                icon: const Icon(
                  Icons.my_location_rounded,
                  color: Colors.blueAccent,
                  size: 18,
                ),
                label: Text(
                  'GPS SAYA',
                  style: GoogleFonts.outfit(
                    color: Colors.blueAccent,
                    fontWeight: FontWeight.w900,
                    fontSize: 12,
                    letterSpacing: 0.8,
                  ),
                ),
              ),
            ),
          ),

          // Details dynamic panel
          Positioned(
            left: 16,
            right: 16,
            bottom: 24,
            child: _buildBottomPanel(),
          ),
        ],
      ),
    );
  }

  Widget _buildBottomPanel() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.08),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'LOKASI TERPILIH',
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w900,
              fontSize: 10,
              color: Colors.blueAccent,
              letterSpacing: 1.2,
            ),
          ),
          const SizedBox(height: 6),
          Row(
            children: [
              const Icon(
                Icons.gps_fixed_rounded,
                color: Colors.blueGrey,
                size: 14,
              ),
              const SizedBox(width: 6),
              Text(
                'Lat: ${_selectedLatLng.latitude.toStringAsFixed(6)}, Lon: ${_selectedLatLng.longitude.toStringAsFixed(6)}',
                style: GoogleFonts.outfit(
                  fontSize: 12,
                  fontWeight: FontWeight.w700,
                  color: Colors.blueGrey,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          const Divider(height: 1, color: Color(0xFFF1F5F9)),
          const SizedBox(height: 12),
          Text(
            'ALAMAT LENGKAP',
            style: GoogleFonts.outfit(
              fontWeight: FontWeight.w800,
              fontSize: 10,
              color: Colors.blueGrey.shade400,
              letterSpacing: 1,
            ),
          ),
          const SizedBox(height: 6),
          SizedBox(
            height: 48,
            child: _isLoadingAddress
                ? Row(
                    children: [
                      const SizedBox(
                        width: 14,
                        height: 14,
                        child: CircularProgressIndicator(
                          strokeWidth: 2,
                          color: Colors.blueAccent,
                        ),
                      ),
                      const SizedBox(width: 10),
                      Text(
                        'Mendapatkan nama jalan...',
                        style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: Colors.blueGrey,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ],
                  )
                : SingleChildScrollView(
                    child: Text(
                      _resolvedAddress.isEmpty
                          ? 'Geser peta untuk mencari alamat.'
                          : _resolvedAddress,
                      style: GoogleFonts.outfit(
                        fontSize: 13,
                        color: const Color(0xFF334155),
                        fontWeight: FontWeight.w600,
                        height: 1.3,
                      ),
                    ),
                  ),
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            height: 52,
            child: ElevatedButton(
              onPressed:
                  (_isLoadingAddress || !_hasValidLocation || _isDragging)
                  ? null
                  : () {
                      Navigator.pop(context, {
                        'latitude': _selectedLatLng.latitude,
                        'longitude': _selectedLatLng.longitude,
                        'address': _resolvedAddress,
                      });
                    },
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.blueAccent,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                elevation: 0,
              ),
              child: Text(
                'GUNAKAN LOKASI INI',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  color: Colors.white,
                  letterSpacing: 0.5,
                  fontSize: 14,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
