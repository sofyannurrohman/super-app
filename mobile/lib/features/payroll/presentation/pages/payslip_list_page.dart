import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:flutter_staggered_animations/flutter_staggered_animations.dart';
import '../bloc/payroll_bloc.dart';
import '../bloc/payroll_event.dart';
import '../bloc/payroll_state.dart';
import '../../domain/entities/payslip.dart';
import 'payslip_detail_page.dart';
import 'payslip_request_page.dart';

class PayslipListPage extends StatefulWidget {
  const PayslipListPage({super.key});

  @override
  State<PayslipListPage> createState() => _PayslipListPageState();
}

class _PayslipListPageState extends State<PayslipListPage> {
  final currencyFormat = NumberFormat.currency(
    locale: 'id_ID',
    symbol: 'Rp ',
    decimalDigits: 0,
  );

  @override
  void initState() {
    super.initState();
    context.read<PayrollBloc>().add(FetchMyPayslipsRequested());
  }

  @override
  Widget build(BuildContext context) {
    const primaryBlue = Color(0xFF1B60F1);
    const bgGray = Color(0xFFF9FAFB);
    const textColor = Color(0xFF111827);
    const subtitleColor = Color(0xFF6B7280);

    return Scaffold(
      backgroundColor: bgGray,
      appBar: AppBar(
        title: Text(
          'Riwayat Slip Gaji',
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
          icon:
              const Icon(Icons.arrow_back_ios_new, color: textColor, size: 20),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocBuilder<PayrollBloc, PayrollState>(
        builder: (context, state) {
          if (state is PayrollLoading) {
            return const Center(
                child: CircularProgressIndicator(color: primaryBlue));
          }

          if (state is PayrollFailure) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline, color: Colors.red, size: 48),
                  const SizedBox(height: 16),
                  Text(
                    state.message,
                    textAlign: TextAlign.center,
                    style: GoogleFonts.inter(color: subtitleColor),
                  ),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () => context
                        .read<PayrollBloc>()
                        .add(FetchMyPayslipsRequested()),
                    style:
                        ElevatedButton.styleFrom(backgroundColor: primaryBlue),
                    child: const Text('Coba Lagi'),
                  ),
                ],
              ),
            );
          }

          if (state is PayrollLoaded) {
            if (state.payslips.isEmpty) {
              return Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.receipt_long_outlined,
                        color: Colors.grey[300], size: 80),
                    const SizedBox(height: 16),
                    Text(
                      'Belum ada riwayat slip gaji',
                      style: GoogleFonts.inter(
                        fontSize: 16,
                        color: subtitleColor,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              );
            }

            return RefreshIndicator(
              onRefresh: () async {
                context.read<PayrollBloc>().add(FetchMyPayslipsRequested());
              },
              color: primaryBlue,
              child: AnimationLimiter(
                child: ListView.builder(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                  itemCount: state.payslips.length,
                  itemBuilder: (context, index) {
                    final payslip = state.payslips[index];
                    return AnimationConfiguration.staggeredList(
                      position: index,
                      duration: const Duration(milliseconds: 500),
                      child: SlideAnimation(
                        verticalOffset: 50.0,
                        child: FadeInAnimation(
                          child: _buildPayslipCard(context, payslip,
                              primaryBlue, textColor, subtitleColor),
                        ),
                      ),
                    );
                  },
                ),
              ),
            );
          }

          return const SizedBox.shrink();
        },
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => const PayslipRequestPage(),
            ),
          );
        },
        backgroundColor: Colors.green,
        icon: const Icon(Icons.description_outlined),
        label: Text('Request Slip Gaji',
            style: GoogleFonts.inter(
                fontWeight: FontWeight.w600, color: Colors.white)),
      ),
    );
  }

  Widget _buildPayslipCard(
    BuildContext context,
    Payslip payslip,
    Color primaryBlue,
    Color textColor,
    Color subtitleColor,
  ) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () {
            Navigator.of(context).push(
              MaterialPageRoute(
                builder: (_) => PayslipDetailPage(payslip: payslip),
              ),
            );
          },
          borderRadius: BorderRadius.circular(16),
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: primaryBlue.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Icon(Icons.description_outlined,
                      color: primaryBlue, size: 24),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        payslip.period,
                        style: GoogleFonts.inter(
                          fontSize: 15,
                          fontWeight: FontWeight.bold,
                          color: textColor,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        'Tanggal Pembayaran: ${payslip.paymentDate}',
                        style: GoogleFonts.inter(
                          fontSize: 12,
                          color: subtitleColor,
                        ),
                      ),
                    ],
                  ),
                ),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text(
                      currencyFormat.format(payslip.takeHomePay),
                      style: GoogleFonts.inter(
                        fontSize: 15,
                        fontWeight: FontWeight.bold,
                        color: primaryBlue,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 8, vertical: 2),
                      decoration: BoxDecoration(
                        color: Colors.green.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Text(
                        'Lunas',
                        style: GoogleFonts.inter(
                          fontSize: 10,
                          fontWeight: FontWeight.w600,
                          color: Colors.green,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(width: 8),
                Icon(Icons.chevron_right, color: Colors.grey[400], size: 20),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
