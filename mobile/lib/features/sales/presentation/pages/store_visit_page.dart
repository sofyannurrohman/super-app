import 'package:flutter/material.dart';
import '../widgets/senior_button.dart';

class StoreVisitPage extends StatefulWidget {
  const StoreVisitPage({super.key});

  @override
  State<StoreVisitPage> createState() => _StoreVisitPageState();
}

class _StoreVisitPageState extends State<StoreVisitPage> {
  String detectedStoreName = "Loading location...";
  String storeCategory = ""; // 'TOKO_BARU' or 'TOKO_LAMA'
  bool isStoreDetected = false;

  @override
  void initState() {
    super.initState();
    _detectStoreByProximity();
  }

  Future<void> _detectStoreByProximity() async {
    // Simulated automatic store tagging based on GPS coordinates
    await Future.delayed(const Duration(seconds: 2));
    setState(() {
      detectedStoreName = "Toko Makmur Jaya (10m)";
      storeCategory = "TOKO_LAMA"; // Auto-tagged
      isStoreDetected = true;
    });
  }

  void _scanReceipt() {
    // 1. Open Camera
    // 2. Capture Image
    // 3. Simpan foto nota untuk admin
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Membuka kamera untuk foto nota...')),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Kunjungan Toko', style: TextStyle(fontSize: 24)),
        backgroundColor: Colors.blueAccent,
      ),
      body: Padding(
        padding: const EdgeInsets.all(24.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Text(
              'Lokasi Anda Saat Ini:',
              style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: isStoreDetected
                    ? Colors.green.shade100
                    : Colors.grey.shade200,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(
                    color: isStoreDetected ? Colors.green : Colors.grey,
                    width: 2),
              ),
              child: Column(
                children: [
                  Text(
                    detectedStoreName,
                    style: TextStyle(
                      fontSize: 28,
                      fontWeight: FontWeight.bold,
                      color: isStoreDetected
                          ? Colors.green.shade800
                          : Colors.black54,
                    ),
                    textAlign: TextAlign.center,
                  ),
                  if (isStoreDetected) ...[
                    const SizedBox(height: 10),
                    Chip(
                      label: Text(
                        storeCategory == 'TOKO_BARU'
                            ? 'Toko Baru (Ekspansi)'
                            : 'Toko Lama (Rutin)',
                        style:
                            const TextStyle(fontSize: 18, color: Colors.white),
                      ),
                      backgroundColor: storeCategory == 'TOKO_BARU'
                          ? Colors.orange
                          : Colors.blue,
                    ),
                  ]
                ],
              ),
            ),
            const Spacer(),
            SeniorButton(
              text: 'SCAN NOTA',
              icon: Icons.camera_alt,
              color: Colors.blue.shade700,
              onPressed: isStoreDetected ? _scanReceipt : () {},
            ),
            const SizedBox(height: 20),
            SeniorButton(
              text: 'SELESAI KUNJUNGAN',
              icon: Icons.check_circle,
              color: Colors.green.shade600,
              onPressed: () {
                // Submit data and go back
                Navigator.pop(context);
              },
            ),
          ],
        ),
      ),
    );
  }
}
