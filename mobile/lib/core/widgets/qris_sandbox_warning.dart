import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:url_launcher/url_launcher.dart';

class QrisSandboxWarning extends StatelessWidget {
  final String qrisUrl;
  final EdgeInsetsGeometry? margin;

  const QrisSandboxWarning({
    super.key,
    required this.qrisUrl,
    this.margin,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin:
          margin ?? const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBEB), // Amber 50
        border: Border.all(color: const Color(0xFFFDE68A)), // Amber 200
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.warning_amber_rounded,
                  color: Color(0xFFD97706), size: 24), // Amber 600
              const SizedBox(width: 10),
              Text(
                'Mode Uji Coba (Sandbox)',
                style: GoogleFonts.outfit(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: const Color(0xFF92400E), // Amber 800
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(
            'QR Code ini diterbitkan oleh Midtrans Sandbox (lingkungan testing). QRIS Sandbox tidak dapat discan menggunakan aplikasi bank asli (seperti BRImo) atau e-wallet nyata dan akan menghasilkan gagal transaksi (kode error THIMP).\n\nUntuk melakukan testing pembayaran, silakan salin tautan QRIS di bawah dan gunakan Simulator Resmi Midtrans.',
            style: GoogleFonts.outfit(
              fontSize: 12,
              height: 1.5,
              color: const Color(0xFFB45309), // Amber 700
            ),
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: () async {
                final Uri url = Uri.parse(
                    'https://simulator.sandbox.midtrans.com/v2/qris/index');
                try {
                  await launchUrl(url, mode: LaunchMode.externalApplication);
                } catch (e) {
                  if (context.mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                          content: Text('Tidak dapat membuka simulator: $e')),
                    );
                  }
                }
              },
              icon: const Icon(Icons.open_in_new_rounded,
                  size: 16, color: Colors.white),
              label: Text(
                'Buka Simulator Midtrans',
                style: GoogleFonts.outfit(
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFFD97706), // Amber 600
                foregroundColor: Colors.white,
                elevation: 0,
                padding: const EdgeInsets.symmetric(vertical: 12),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12)),
              ),
            ),
          ),
        ],
      ),
    );
  }
}
