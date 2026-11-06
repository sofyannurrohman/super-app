import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import '../../domain/entities/payslip.dart';
import '../services/payslip_pdf_service.dart';
import 'package:local_auth/local_auth.dart';

class PayslipDetailPage extends StatefulWidget {
  final Payslip payslip;

  const PayslipDetailPage({super.key, required this.payslip});

  @override
  State<PayslipDetailPage> createState() => _PayslipDetailPageState();
}

class _PayslipDetailPageState extends State<PayslipDetailPage> {
  bool _isAuthenticated = false;
  final LocalAuthentication auth = LocalAuthentication();

  @override
  void initState() {
    super.initState();
    _authenticate();
  }

  Future<void> _authenticate() async {
    try {
      final bool canAuthenticateWithBiometrics = await auth.canCheckBiometrics;
      final bool canAuthenticate =
          canAuthenticateWithBiometrics || await auth.isDeviceSupported();

      if (!canAuthenticate) {
        setState(() => _isAuthenticated = true); // Fallback if no biometrics
        return;
      }

      final bool didAuthenticate = await auth.authenticate(
        localizedReason:
            'Silakan verifikasi identitas Anda untuk melihat slip gaji.',
        options: const AuthenticationOptions(
          stickyAuth: true,
          biometricOnly: false,
        ),
      );

      if (didAuthenticate) {
        setState(() => _isAuthenticated = true);
      } else {
        if (mounted) Navigator.of(context).pop();
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isAuthenticated = true); // Fallback on error for now
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    if (!_isAuthenticated) {
      return const Scaffold(body: Center(child: CircularProgressIndicator()));
    }

    final payslip = widget.payslip;
    const primaryBlue = Color(0xFF1B60F1);
    const bgGray = Color(0xFFF9FAFB);
    const textColor = Color(0xFF111827);
    const subtitleColor = Color(0xFF6B7280);
    const earningColor = Color(0xFF10B981);
    const deductionColor = Color(0xFFEF4444);

    final currencyFormat = NumberFormat.currency(
      locale: 'id_ID',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    return Scaffold(
      backgroundColor: bgGray,
      appBar: AppBar(
        title: Text(
          'Detail Slip Gaji',
          style: GoogleFonts.inter(
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: textColor,
          ),
        ),
        centerTitle: true,
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(
            Icons.arrow_back_ios_new,
            color: textColor,
            size: 20,
          ),
          onPressed: () => Navigator.of(context).pop(),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.picture_as_pdf_rounded, color: primaryBlue),
            onPressed: () => PayslipPdfService.generateAndPrint(payslip),
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _buildHeaderCard(payslip, primaryBlue, textColor, subtitleColor),
            const SizedBox(height: 24),
            _buildSummaryCard(payslip, currencyFormat, primaryBlue),
            const SizedBox(height: 24),
            _buildSectionTitle('Pendapatan (Earnings)', earningColor),
            const SizedBox(height: 12),
            _buildItemsList(payslip.earnings, currencyFormat, textColor),
            const SizedBox(height: 24),
            _buildSectionTitle('Potongan (Deductions)', deductionColor),
            const SizedBox(height: 12),
            _buildItemsList(payslip.deductions, currencyFormat, textColor),
            const SizedBox(height: 32),
            _buildFooter(payslip, currencyFormat, textColor, subtitleColor),
          ],
        ),
      ),
    );
  }

  Widget _buildHeaderCard(
    Payslip payslip,
    Color primaryBlue,
    Color textColor,
    Color subtitleColor,
  ) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    payslip.period,
                    style: GoogleFonts.inter(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: textColor,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    payslip.jobTitle,
                    style: GoogleFonts.inter(
                      fontSize: 14,
                      color: subtitleColor,
                    ),
                  ),
                ],
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: Colors.green.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Text(
                  'Lunas',
                  style: GoogleFonts.inter(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: Colors.green,
                  ),
                ),
              ),
            ],
          ),
          const Divider(height: 32),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _buildInfoColumn(
                'ID Slip',
                payslip.id.substring(0, 8).toUpperCase(),
                subtitleColor,
                textColor,
              ),
              _buildInfoColumn(
                'Tanggal Bayar',
                payslip.paymentDate,
                subtitleColor,
                textColor,
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildInfoColumn(
    String label,
    String value,
    Color labelColor,
    Color valueColor,
  ) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: GoogleFonts.inter(fontSize: 12, color: labelColor)),
        const SizedBox(height: 4),
        Text(
          value,
          style: GoogleFonts.inter(
            fontSize: 14,
            fontWeight: FontWeight.bold,
            color: valueColor,
          ),
        ),
      ],
    );
  }

  Widget _buildSummaryCard(
    Payslip payslip,
    NumberFormat currencyFormat,
    Color primaryBlue,
  ) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF1B60F1), Color(0xFF104FD4)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: primaryBlue.withOpacity(0.3),
            blurRadius: 20,
            offset: const Offset(0, 10),
          ),
        ],
      ),
      child: Column(
        children: [
          Text(
            'Take Home Pay',
            style: GoogleFonts.inter(
              fontSize: 14,
              color: Colors.white.withOpacity(0.8),
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            currencyFormat.format(payslip.takeHomePay),
            style: GoogleFonts.inter(
              fontSize: 32,
              fontWeight: FontWeight.bold,
              color: Colors.white,
              letterSpacing: -0.5,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSectionTitle(String title, Color color) {
    return Row(
      children: [
        Container(
          width: 4,
          height: 16,
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(2),
          ),
        ),
        const SizedBox(width: 8),
        Text(
          title,
          style: GoogleFonts.inter(
            fontSize: 15,
            fontWeight: FontWeight.bold,
            color: const Color(0xFF111827),
          ),
        ),
      ],
    );
  }

  Widget _buildItemsList(
    List<PayslipItem> items,
    NumberFormat currencyFormat,
    Color textColor,
  ) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.01),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Column(
        children: items.map((item) {
          final isLast = items.indexOf(item) == items.length - 1;
          return Column(
            children: [
              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 14,
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      item.name,
                      style: GoogleFonts.inter(fontSize: 14, color: textColor),
                    ),
                    Text(
                      currencyFormat.format(item.amount),
                      style: GoogleFonts.inter(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: textColor,
                      ),
                    ),
                  ],
                ),
              ),
              if (!isLast) const Divider(height: 1, indent: 16, endIndent: 16),
            ],
          );
        }).toList(),
      ),
    );
  }

  Widget _buildFooter(
    Payslip payslip,
    NumberFormat currencyFormat,
    Color textColor,
    Color subtitleColor,
  ) {
    return Column(
      children: [
        _buildTotalRow(
          'Total Pendapatan',
          currencyFormat.format(payslip.totalAllowance + payslip.basicSalary),
          textColor,
        ),
        const SizedBox(height: 8),
        _buildTotalRow(
          'Total Potongan',
          currencyFormat.format(payslip.totalDeduction),
          Colors.redAccent,
        ),
        const SizedBox(height: 32),
        Center(
          child: Text(
            'Slip gaji ini dihasilkan secara otomatis oleh sistem Karsa HR.',
            textAlign: TextAlign.center,
            style: GoogleFonts.inter(
              fontSize: 12,
              color: subtitleColor,
              fontStyle: FontStyle.italic,
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildTotalRow(String label, String value, Color valueColor) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: GoogleFonts.inter(
            fontSize: 14,
            fontWeight: FontWeight.w500,
            color: const Color(0xFF6B7280),
          ),
        ),
        Text(
          value,
          style: GoogleFonts.inter(
            fontSize: 15,
            fontWeight: FontWeight.bold,
            color: valueColor,
          ),
        ),
      ],
    );
  }
}
