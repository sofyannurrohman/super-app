import 'dart:convert';
import 'dart:ui' as ui;
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';

class SignatureCanvas extends StatefulWidget {
  final Function(String base64Image) onSave;

  const SignatureCanvas({super.key, required this.onSave});

  @override
  State<SignatureCanvas> createState() => _SignatureCanvasState();
}

class _SignatureCanvasState extends State<SignatureCanvas> {
  List<Offset?> points = [];

  Future<void> _saveSignature() async {
    if (points.isEmpty) {
      widget.onSave('');
      return;
    }

    try {
      final recorder = ui.PictureRecorder();
      const canvasSize = Size(400, 200);
      final canvas = Canvas(
        recorder,
        Rect.fromLTWH(0, 0, canvasSize.width, canvasSize.height),
      );

      // Draw white background
      final bgPaint = Paint()..color = Colors.white;
      canvas.drawRect(
        Rect.fromLTWH(0, 0, canvasSize.width, canvasSize.height),
        bgPaint,
      );

      // Draw signature paths
      final paint = Paint()
        ..color = Colors.black
        ..strokeCap = StrokeCap.round
        ..strokeWidth = 4.0;

      for (int i = 0; i < points.length - 1; i++) {
        if (points[i] != null && points[i + 1] != null) {
          canvas.drawLine(points[i]!, points[i + 1]!, paint);
        }
      }

      final picture = recorder.endRecording();
      final img = await picture.toImage(
        canvasSize.width.toInt(),
        canvasSize.height.toInt(),
      );
      final byteData = await img.toByteData(format: ui.ImageByteFormat.png);

      if (byteData != null) {
        final bytes = byteData.buffer.asUint8List();
        final base64String = base64Encode(bytes);
        widget.onSave(base64String);
      }
    } catch (e) {
      debugPrint('Gagal mengekspor tanda tangan: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          'Gambarkan Tanda Tangan Anda di Bawah ini:',
          style: GoogleFonts.plusJakartaSans(
            fontSize: 13,
            fontWeight: FontWeight.w700,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 12),
        Container(
          height: 200,
          width: double.infinity,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.grey.shade300),
          ),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(16),
            child: GestureDetector(
              behavior: HitTestBehavior.opaque,
              onVerticalDragStart: (_) {},
              onVerticalDragUpdate: (_) {},
              onHorizontalDragStart: (_) {},
              onHorizontalDragUpdate: (_) {},
              child: Listener(
                onPointerDown: (event) {
                  setState(() {
                    points.add(event.localPosition);
                  });
                },
                onPointerMove: (event) {
                  setState(() {
                    points.add(event.localPosition);
                  });
                },
                onPointerUp: (event) {
                  setState(() {
                    points.add(null);
                  });
                  _saveSignature();
                },
                child: SizedBox.expand(
                  child: CustomPaint(
                    painter: SignaturePainter(points: points),
                  ),
                ),
              ),
            ),
          ),
        ),
        const SizedBox(height: 16),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            TextButton.icon(
              onPressed: () {
                setState(() => points.clear());
                widget.onSave('');
              },
              icon: const Icon(Icons.clear_rounded, color: AppColors.error),
              label: Text(
                'BERSIHKAN',
                style: GoogleFonts.plusJakartaSans(
                  color: AppColors.error,
                  fontWeight: FontWeight.w800,
                  fontSize: 12,
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }
}

class SignaturePainter extends CustomPainter {
  final List<Offset?> points;

  SignaturePainter({required this.points});

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = Colors.black
      ..strokeCap = StrokeCap.round
      ..strokeWidth = 4.0;

    for (int i = 0; i < points.length - 1; i++) {
      if (points[i] != null && points[i + 1] != null) {
        canvas.drawLine(points[i]!, points[i + 1]!, paint);
      }
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => true;
}
