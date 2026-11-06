import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/supervisor/presentation/bloc/supervisor_bloc.dart';

class SupervisorTransferPage extends StatefulWidget {
  final String branchId;
  const SupervisorTransferPage({Key? key, required this.branchId})
      : super(key: key);

  @override
  State<SupervisorTransferPage> createState() => _SupervisorTransferPageState();
}

class _SupervisorTransferPageState extends State<SupervisorTransferPage> {
  @override
  void initState() {
    super.initState();
    context.read<SupervisorBloc>().add(LoadPendingTransfers(widget.branchId));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text('Approval Barang Bawaan',
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
          if (state.isLoadingTransfers) {
            return const Center(
                child:
                    CircularProgressIndicator(color: AppColors.primaryGreen));
          }
          if (state.pendingTransfers.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(Icons.inventory_2_outlined,
                      size: 64, color: AppColors.textTertiary.withOpacity(0.5)),
                  const SizedBox(height: 16),
                  Text(
                    'Tidak ada permintaan pending',
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
            itemCount: state.pendingTransfers.length,
            itemBuilder: (context, index) {
              final t = state.pendingTransfers[index];
              return Container(
                margin: const EdgeInsets.only(bottom: 16),
                padding: const EdgeInsets.all(20),
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
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Container(
                          padding: const EdgeInsets.all(10),
                          decoration: BoxDecoration(
                            color: AppColors.warning.withOpacity(0.1),
                            shape: BoxShape.circle,
                          ),
                          child: const Icon(Icons.inventory_2_rounded,
                              color: AppColors.warning, size: 24),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                '${t.product?.name ?? 'Produk'}',
                                style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w700,
                                    fontSize: 16,
                                    color: AppColors.textPrimary),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                'Jumlah: ${t.quantity} ${t.product?.unit ?? 'pcs'}',
                                style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w600,
                                    fontSize: 14,
                                    color: AppColors.warning),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                    const Padding(
                      padding: EdgeInsets.symmetric(vertical: 16),
                      child: Divider(color: AppColors.grayBorder, height: 1),
                    ),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text('Sales',
                                style: GoogleFonts.outfit(
                                    fontSize: 12,
                                    color: AppColors.textSecondary)),
                            const SizedBox(height: 2),
                            Text(t.employeeName,
                                style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w600,
                                    fontSize: 14,
                                    color: AppColors.textPrimary)),
                          ],
                        ),
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.end,
                          children: [
                            Text('Stok Gudang',
                                style: GoogleFonts.outfit(
                                    fontSize: 12,
                                    color: AppColors.textSecondary)),
                            const SizedBox(height: 2),
                            Text('${t.product?.warehouseStock ?? 0}',
                                style: GoogleFonts.outfit(
                                    fontWeight: FontWeight.w600,
                                    fontSize: 14,
                                    color: AppColors.textPrimary)),
                          ],
                        ),
                      ],
                    ),
                    const SizedBox(height: 20),
                    Row(
                      children: [
                        Expanded(
                          child: OutlinedButton(
                            onPressed: () => context
                                .read<SupervisorBloc>()
                                .add(DoRejectTransfer(t.id)),
                            style: OutlinedButton.styleFrom(
                              foregroundColor: AppColors.error,
                              side: const BorderSide(color: AppColors.error),
                              padding: const EdgeInsets.symmetric(vertical: 14),
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
                                .add(DoApproveTransfer(t.id, t.quantity)),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: AppColors.primaryGreen,
                              foregroundColor: AppColors.surface,
                              elevation: 0,
                              padding: const EdgeInsets.symmetric(vertical: 14),
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
              );
            },
          );
        },
      ),
    );
  }
}
