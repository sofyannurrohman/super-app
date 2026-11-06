import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/supervisor/presentation/bloc/supervisor_bloc.dart';
import 'package:intl/intl.dart';

class SupervisorDeliveryBatchPage extends StatefulWidget {
  final String branchId;
  const SupervisorDeliveryBatchPage({Key? key, required this.branchId})
      : super(key: key);

  @override
  State<SupervisorDeliveryBatchPage> createState() =>
      _SupervisorDeliveryBatchPageState();
}

class _SupervisorDeliveryBatchPageState
    extends State<SupervisorDeliveryBatchPage> {
  @override
  void initState() {
    super.initState();
    context
        .read<SupervisorBloc>()
        .add(LoadPendingDeliveryBatches(widget.branchId));
  }

  @override
  Widget build(BuildContext context) {
    final currencyFormat =
        NumberFormat.currency(locale: 'id_ID', symbol: 'Rp ', decimalDigits: 0);

    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text('Approval Surat Jalan',
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w700,
                color: AppColors.surface,
                fontSize: 18)),
        backgroundColor: AppColors.primaryGreen,
        foregroundColor: AppColors.surface,
        elevation: 0,
        centerTitle: false,
        flexibleSpace: Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: AppColors.primaryGradient,
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
        ),
      ),
      body: BlocConsumer<SupervisorBloc, SupervisorState>(
        listener: (context, state) {
          if (state.errorMessage != null) {
            ScaffoldMessenger.of(context).showSnackBar(SnackBar(
              content: Text(state.errorMessage!, style: GoogleFonts.outfit()),
              backgroundColor: AppColors.error,
              behavior: SnackBarBehavior.floating,
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10)),
            ));
          }
          if (state.successMessage != null) {
            ScaffoldMessenger.of(context).showSnackBar(SnackBar(
              content: Text(state.successMessage!, style: GoogleFonts.outfit()),
              backgroundColor: AppColors.success,
              behavior: SnackBarBehavior.floating,
              shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10)),
            ));
          }
        },
        builder: (context, state) {
          if (state.isLoadingBatches) {
            return const Center(
                child:
                    CircularProgressIndicator(color: AppColors.primaryGreen));
          }
          if (state.pendingBatches.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.local_shipping_outlined,
                      size: 64, color: AppColors.textTertiary.withOpacity(0.5)),
                  const SizedBox(height: 16),
                  Text(
                    'Tidak ada surat jalan pending',
                    style: GoogleFonts.outfit(
                        fontSize: 16,
                        color: AppColors.textSecondary,
                        fontWeight: FontWeight.w500),
                  ),
                ],
              ),
            );
          }
          return ListView.builder(
            physics: const BouncingScrollPhysics(),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 20),
            itemCount: state.pendingBatches.length,
            itemBuilder: (context, index) {
              final b = state.pendingBatches[index];
              return Container(
                margin: const EdgeInsets.only(bottom: 16),
                decoration: BoxDecoration(
                  color: AppColors.surface,
                  borderRadius: BorderRadius.circular(20),
                  border: Border.all(color: AppColors.grayBorder, width: 1),
                  boxShadow: [
                    BoxShadow(
                        color: AppColors.textTertiary.withOpacity(0.05),
                        blurRadius: 10,
                        offset: const Offset(0, 4)),
                  ],
                ),
                child: Theme(
                  data: Theme.of(context).copyWith(
                    dividerColor:
                        Colors.transparent, // Remove line on ExpansionTile
                  ),
                  child: ExpansionTile(
                    tilePadding:
                        const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
                    childrenPadding:
                        const EdgeInsets.only(left: 20, right: 20, bottom: 20),
                    leading: Container(
                      padding: const EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: AppColors.info.withOpacity(0.1),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(Icons.local_shipping_rounded,
                          color: AppColors.info, size: 24),
                    ),
                    title: Text(
                        b.deliveryOrderNo.isNotEmpty
                            ? b.deliveryOrderNo
                            : 'DRAFT-BATCH',
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w700,
                            fontSize: 16,
                            color: AppColors.textPrimary)),
                    subtitle: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const SizedBox(height: 4),
                        Text('Driver: ${b.driverName}',
                            style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w500,
                                fontSize: 13,
                                color: AppColors.textSecondary)),
                        const SizedBox(height: 4),
                        Text('Total: ${currencyFormat.format(b.totalAmount)}',
                            style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w700,
                                fontSize: 14,
                                color: AppColors.primaryGreen)),
                      ],
                    ),
                    children: [
                      const Padding(
                        padding: EdgeInsets.symmetric(vertical: 8),
                        child: Divider(color: AppColors.grayBorder, height: 1),
                      ),
                      ...b.items
                          .map((i) => Padding(
                                padding:
                                    const EdgeInsets.symmetric(vertical: 8),
                                child: Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Container(
                                      margin: const EdgeInsets.only(
                                          top: 4, right: 12),
                                      width: 8,
                                      height: 8,
                                      decoration: const BoxDecoration(
                                          color: AppColors.grayBorder,
                                          shape: BoxShape.circle),
                                    ),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          Text(i.storeName,
                                              style: GoogleFonts.outfit(
                                                  fontWeight: FontWeight.w600,
                                                  fontSize: 14,
                                                  color:
                                                      AppColors.textPrimary)),
                                          const SizedBox(height: 2),
                                          Text(i.receiptNo,
                                              style: GoogleFonts.outfit(
                                                  fontSize: 12,
                                                  color:
                                                      AppColors.textTertiary)),
                                        ],
                                      ),
                                    ),
                                    Text(currencyFormat.format(i.totalAmount),
                                        style: GoogleFonts.outfit(
                                            fontWeight: FontWeight.w700,
                                            fontSize: 14,
                                            color: AppColors.textPrimary)),
                                  ],
                                ),
                              ))
                          .toList(),
                      const SizedBox(height: 24),
                      Row(
                        children: [
                          Expanded(
                            child: OutlinedButton(
                              onPressed: () => context
                                  .read<SupervisorBloc>()
                                  .add(DoCancelDeliveryBatch(b.id)),
                              style: OutlinedButton.styleFrom(
                                foregroundColor: AppColors.error,
                                side: const BorderSide(color: AppColors.error),
                                padding:
                                    const EdgeInsets.symmetric(vertical: 14),
                                shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12)),
                              ),
                              child: Text('Tolak',
                                  style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w700)),
                            ),
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: ElevatedButton(
                              onPressed: () => context
                                  .read<SupervisorBloc>()
                                  .add(DoApproveDeliveryBatch(b.id)),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: AppColors.primaryGreen,
                                foregroundColor: AppColors.surface,
                                elevation: 0,
                                padding:
                                    const EdgeInsets.symmetric(vertical: 14),
                                shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12)),
                              ),
                              child: Text('Setujui',
                                  style: GoogleFonts.outfit(
                                      fontWeight: FontWeight.w700)),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              );
            },
          );
        },
      ),
    );
  }
}
