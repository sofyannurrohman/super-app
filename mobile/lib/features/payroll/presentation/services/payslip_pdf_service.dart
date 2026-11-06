import 'package:hris_app/features/payroll/domain/entities/payslip.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:intl/intl.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart';

class PayslipPdfService {
  static Future<void> generateAndPrint(Payslip payslip) async {
    final pdf = pw.Document();
    final currencyFormat = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    // Header info
    pdf.addPage(
      pw.Page(
        pageFormat: PdfPageFormat.a4,
        build: (pw.Context context) {
          return pw.Column(
            crossAxisAlignment: pw.CrossAxisAlignment.start,
            children: [
              // Company Header
              pw.Row(
                mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                children: [
                  pw.Column(
                    crossAxisAlignment: pw.CrossAxisAlignment.start,
                    children: [
                      pw.Text(
                        'Perusahaan Demo',
                        style: pw.TextStyle(
                          fontSize: 18,
                          fontWeight: pw.FontWeight.bold,
                        ),
                      ),
                      pw.Text(
                        'Dokumen contoh untuk penggunaan trial',
                        style: const pw.TextStyle(fontSize: 10),
                      ),
                    ],
                  ),
                  pw.Column(
                    crossAxisAlignment: pw.CrossAxisAlignment.end,
                    children: [
                      pw.Text(
                        'SLIP GAJI DIGITAL',
                        style: pw.TextStyle(
                          fontSize: 14,
                          fontWeight: pw.FontWeight.bold,
                          color: PdfColors.blue900,
                        ),
                      ),
                      pw.Text(
                        'Periode: ${payslip.period}',
                        style: const pw.TextStyle(fontSize: 10),
                      ),
                    ],
                  ),
                ],
              ),
              pw.Divider(thickness: 1, color: PdfColors.grey300, height: 24),

              // Employee Info
              pw.Row(
                crossAxisAlignment: pw.CrossAxisAlignment.start,
                children: [
                  pw.Expanded(
                    child: pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.start,
                      children: [
                        _pdfInfoField(
                          'Nama Karyawan',
                          'Sofyan Nurrohman',
                        ), // Mock name for now
                        _pdfInfoField('Jabatan', payslip.jobTitle),
                      ],
                    ),
                  ),
                  pw.Expanded(
                    child: pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.start,
                      children: [
                        _pdfInfoField(
                          'No. Referensi',
                          payslip.id.substring(0, 8).toUpperCase(),
                        ),
                        _pdfInfoField('Tanggal Bayar', payslip.paymentDate),
                      ],
                    ),
                  ),
                ],
              ),
              pw.SizedBox(height: 32),

              // Earnings Section
              pw.Text(
                'PENDAPATAN (EARNINGS)',
                style: pw.TextStyle(
                  fontSize: 11,
                  fontWeight: pw.FontWeight.bold,
                  color: PdfColors.green900,
                ),
              ),
              pw.Divider(thickness: 0.5, color: PdfColors.grey300),
              ...payslip.earnings.map(
                (item) =>
                    _pdfItemRow(item.name, currencyFormat.format(item.amount)),
              ),
              pw.SizedBox(height: 16),

              // Deductions Section
              pw.Text(
                'POTONGAN (DEDUCTIONS)',
                style: pw.TextStyle(
                  fontSize: 11,
                  fontWeight: pw.FontWeight.bold,
                  color: PdfColors.red900,
                ),
              ),
              pw.Divider(thickness: 0.5, color: PdfColors.grey300),
              ...payslip.deductions.map(
                (item) => _pdfItemRow(
                  item.name,
                  '(${currencyFormat.format(item.amount)})',
                ),
              ),
              pw.SizedBox(height: 32),

              // Summary Section
              pw.Container(
                padding: const pw.EdgeInsets.all(16),
                decoration: const pw.BoxDecoration(
                  color: PdfColors.grey100,
                  borderRadius: pw.BorderRadius.all(pw.Radius.circular(8)),
                ),
                child: pw.Column(
                  children: [
                    _pdfTotalRow(
                      'TOTAL PENDAPATAN',
                      currencyFormat.format(
                        payslip.basicSalary + payslip.totalAllowance,
                      ),
                    ),
                    _pdfTotalRow(
                      'TOTAL POTONGAN',
                      currencyFormat.format(payslip.totalDeduction),
                    ),
                    pw.Divider(
                      thickness: 1,
                      color: PdfColors.grey300,
                      height: 20,
                    ),
                    _pdfTotalRow(
                      'PENGHASILAN BERSIH (THP)',
                      currencyFormat.format(payslip.takeHomePay),
                      isFinal: true,
                    ),
                  ],
                ),
              ),

              pw.Spacer(),
              pw.Center(
                child: pw.Text(
                  'Dokumen ini sah dan diterbitkan secara elektronik oleh Karsa HR.',
                  style: const pw.TextStyle(
                    fontSize: 8,
                    color: PdfColors.grey600,
                  ),
                ),
              ),
            ],
          );
        },
      ),
    );

    // Preview and download
    await Printing.layoutPdf(
      onLayout: (PdfPageFormat format) async => pdf.save(),
      name: 'SlipGaji_${payslip.period.replaceAll(' ', '_')}.pdf',
    );
  }

  static pw.Widget _pdfInfoField(String label, String value) {
    return pw.Padding(
      padding: const pw.EdgeInsets.only(bottom: 6),
      child: pw.Column(
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        children: [
          pw.Text(
            label,
            style: const pw.TextStyle(fontSize: 8, color: PdfColors.grey600),
          ),
          pw.Text(
            value,
            style: pw.TextStyle(fontSize: 10, fontWeight: pw.FontWeight.bold),
          ),
        ],
      ),
    );
  }

  static pw.Widget _pdfItemRow(String label, String value) {
    return pw.Padding(
      padding: const pw.EdgeInsets.symmetric(vertical: 4),
      child: pw.Row(
        mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
        children: [
          pw.Text(label, style: const pw.TextStyle(fontSize: 10)),
          pw.Text(
            value,
            style: pw.TextStyle(fontSize: 10, fontWeight: pw.FontWeight.bold),
          ),
        ],
      ),
    );
  }

  static pw.Widget _pdfTotalRow(
    String label,
    String value, {
    bool isFinal = false,
  }) {
    return pw.Padding(
      padding: const pw.EdgeInsets.symmetric(vertical: 4),
      child: pw.Row(
        mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
        children: [
          pw.Text(
            label,
            style: pw.TextStyle(
              fontSize: 10,
              fontWeight: isFinal ? pw.FontWeight.bold : pw.FontWeight.normal,
            ),
          ),
          pw.Text(
            value,
            style: pw.TextStyle(
              fontSize: 12,
              fontWeight: pw.FontWeight.bold,
              color: isFinal ? PdfColors.blue900 : PdfColors.black,
            ),
          ),
        ],
      ),
    );
  }
}
