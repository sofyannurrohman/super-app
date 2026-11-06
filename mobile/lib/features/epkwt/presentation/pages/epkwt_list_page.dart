import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_bloc.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_event.dart';
import 'package:hris_app/features/epkwt/presentation/bloc/epkwt_state.dart';
import 'package:hris_app/features/epkwt/presentation/pages/epkwt_request_page.dart';
import 'package:hris_app/features/epkwt/presentation/pages/epkwt_detail_page.dart';

class EpkwtListPage extends StatefulWidget {
  const EpkwtListPage({super.key});

  @override
  State<EpkwtListPage> createState() => _EpkwtListPageState();
}

class _EpkwtListPageState extends State<EpkwtListPage> {
  @override
  void initState() {
    super.initState();
    context.read<EpkwtBloc>().add(FetchMyEpkwtsRequested());
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      appBar: AppBar(
        title: Text(
          'Elektronik PKWT (EPKWT)',
          style: GoogleFonts.plusJakartaSans(
            fontWeight: FontWeight.w900,
            fontSize: 16,
            color: Colors.white,
          ),
        ),
        backgroundColor: AppColors.primaryGreen,
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
      ),
      body: BlocConsumer<EpkwtBloc, EpkwtState>(
        listener: (context, state) {
          if (state is EpkwtFailure) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(state.message)),
            );
          }
          if (state is EpkwtSuccess) {
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(state.message)),
            );
            context.read<EpkwtBloc>().add(FetchMyEpkwtsRequested());
          }
        },
        builder: (context, state) {
          if (state is EpkwtLoading) {
            return const Center(
              child: CircularProgressIndicator(color: AppColors.primaryGreen),
            );
          }

          if (state is EpkwtListLoaded) {
            final list = state.epkwts;

            if (list.isEmpty) {
              return Center(
                child: Padding(
                  padding: const EdgeInsets.all(32.0),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(
                        Icons.description_outlined,
                        size: 80,
                        color: Colors.grey.shade400,
                      ),
                      const SizedBox(height: 16),
                      Text(
                        'Belum ada Kontrak Kerja',
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 16,
                          fontWeight: FontWeight.w800,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Silakan ajukan pembuatan kontrak kerja PKWT baru secara mandiri.',
                        textAlign: TextAlign.center,
                        style: GoogleFonts.plusJakartaSans(
                          fontSize: 13,
                          color: AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }

            return RefreshIndicator(
              onRefresh: () async {
                context.read<EpkwtBloc>().add(FetchMyEpkwtsRequested());
              },
              child: ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: list.length,
                itemBuilder: (context, index) {
                  final item = list[index];

                  Color statusColor = Colors.grey.shade600;
                  String statusText = item.status;
                  if (item.status == 'PENDING_HR') {
                    statusColor = Colors.amber.shade800;
                    statusText = 'Menunggu Review HR';
                  } else if (item.status == 'APPROVED_DRAFT') {
                    statusColor = Colors.blue.shade700;
                    statusText = 'Siap Ditandatangani';
                  } else if (item.status == 'ACTIVE') {
                    statusColor = Colors.green.shade700;
                    statusText = 'Aktif / Selesai';
                  } else if (item.status == 'REJECTED') {
                    statusColor = Colors.red.shade700;
                    statusText = 'Ditolak';
                  }

                  return Card(
                    margin: const EdgeInsets.only(bottom: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    color: Colors.white,
                    elevation: 2,
                    shadowColor: Colors.black.withOpacity(0.04),
                    child: InkWell(
                      onTap: () {
                        Navigator.of(context).push(
                          MaterialPageRoute(
                            builder: (_) => EpkwtDetailPage(request: item),
                          ),
                        );
                      },
                      borderRadius: BorderRadius.circular(16),
                      child: Padding(
                        padding: const EdgeInsets.all(16.0),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Expanded(
                                  child: Text(
                                    item.documentNumber ?? 'DRAF KONTRAK',
                                    style: GoogleFonts.plusJakartaSans(
                                      fontWeight: FontWeight.w900,
                                      fontSize: 14,
                                      color: AppColors.textPrimary,
                                    ),
                                  ),
                                ),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 10,
                                    vertical: 4,
                                  ),
                                  decoration: BoxDecoration(
                                    color: statusColor.withOpacity(0.1),
                                    borderRadius: BorderRadius.circular(20),
                                  ),
                                  child: Text(
                                    statusText,
                                    style: GoogleFonts.plusJakartaSans(
                                      color: statusColor,
                                      fontWeight: FontWeight.w800,
                                      fontSize: 10,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 12),
                            const Divider(height: 1),
                            const SizedBox(height: 12),
                            Row(
                              children: [
                                Icon(Icons.work_outline_rounded,
                                    size: 16, color: Colors.grey.shade500),
                                const SizedBox(width: 8),
                                Text(
                                  item.jobPositionTitle ??
                                      'Menunggu penetapan HR',
                                  style: GoogleFonts.plusJakartaSans(
                                    fontWeight: FontWeight.w700,
                                    fontSize: 12,
                                    color: AppColors.textSecondary,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 6),
                            Row(
                              children: [
                                Icon(Icons.calendar_today_rounded,
                                    size: 16, color: Colors.grey.shade500),
                                const SizedBox(width: 8),
                                Text(
                                  item.startDate != null
                                      ? '${DateFormat('dd MMM yyyy').format(item.startDate!)} s/d ${item.endDate != null ? DateFormat('dd MMM yyyy').format(item.endDate!) : 'Selesai'}'
                                      : 'Durasi kontrak belum diset',
                                  style: GoogleFonts.plusJakartaSans(
                                    fontWeight: FontWeight.w600,
                                    fontSize: 12,
                                    color: AppColors.textSecondary,
                                  ),
                                ),
                              ],
                            ),
                            if (item.status == 'APPROVED_DRAFT') ...[
                              const SizedBox(height: 16),
                              SizedBox(
                                width: double.infinity,
                                height: 40,
                                child: ElevatedButton.icon(
                                  onPressed: () {
                                    Navigator.of(context).push(
                                      MaterialPageRoute(
                                        builder: (_) => EpkwtDetailPage(
                                          request: item,
                                          openSignatureInitially: true,
                                        ),
                                      ),
                                    );
                                  },
                                  icon: const Icon(Icons.edit_document,
                                      size: 16, color: Colors.white),
                                  label: Text(
                                    'TANDATANGANI KONTRAK',
                                    style: GoogleFonts.plusJakartaSans(
                                      color: Colors.white,
                                      fontWeight: FontWeight.w800,
                                      fontSize: 11,
                                    ),
                                  ),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: AppColors.primaryGreen,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(10),
                                    ),
                                  ),
                                ),
                              )
                            ]
                          ],
                        ),
                      ),
                    ),
                  );
                },
              ),
            );
          }

          return const SizedBox.shrink();
        },
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          Navigator.of(context).push(
            MaterialPageRoute(builder: (_) => const EpkwtRequestPage()),
          );
        },
        backgroundColor: AppColors.primaryGreen,
        icon: const Icon(Icons.add_rounded, color: Colors.white),
        label: Text(
          'AJUKAN PKWT',
          style: GoogleFonts.plusJakartaSans(
            fontWeight: FontWeight.w800,
            color: Colors.white,
            fontSize: 12,
          ),
        ),
      ),
    );
  }
}
