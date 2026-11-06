import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/directory/data/models/directory_model.dart';
import 'package:hris_app/features/directory/presentation/bloc/directory_bloc.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:url_launcher/url_launcher.dart';

class DirectoryPage extends StatefulWidget {
  const DirectoryPage({super.key});

  @override
  State<DirectoryPage> createState() => _DirectoryPageState();
}

class _DirectoryPageState extends State<DirectoryPage> {
  final TextEditingController _searchController = TextEditingController();
  Timer? _debounce;
  String _selectedBranch = 'Semua Cabang';

  @override
  void initState() {
    super.initState();
    context.read<DirectoryBloc>().add(const FetchDirectoryRequested());
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _searchController.dispose();
    super.dispose();
  }

  void _callNumber(String phoneNumber) async {
    final Uri url = Uri.parse('tel:$phoneNumber');
    if (await canLaunchUrl(url)) {
      await launchUrl(url);
    }
  }

  void _whatsappNumber(String phoneNumber) async {
    String normalized = phoneNumber.replaceAll(RegExp(r'\D'), '');
    if (normalized.startsWith('0')) {
      normalized = '62${normalized.substring(1)}';
    }
    final Uri url = Uri.parse('https://wa.me/$normalized');
    if (await canLaunchUrl(url)) {
      await launchUrl(url, mode: LaunchMode.externalApplication);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.backgroundAlt,
      body: BlocBuilder<DirectoryBloc, DirectoryState>(
        builder: (context, state) {
          int totalCount = 0;
          List<EmployeeDirectory> allEmployees = [];
          if (state is DirectoryLoaded) {
            allEmployees = state.directory;
            totalCount = allEmployees.length;
          }

          // Apply client-side branch filter
          List<EmployeeDirectory> displayedEmployees = allEmployees;
          if (_selectedBranch != 'Semua Cabang') {
            displayedEmployees = displayedEmployees.where((e) {
              final bName = (e.branchName ?? '').trim();
              return bName.toLowerCase() == _selectedBranch.toLowerCase();
            }).toList();
          }

          return NestedScrollView(
            headerSliverBuilder: (context, innerBoxIsScrolled) {
              return [
                _buildSliverAppBar(totalCount),
                _buildSearchAndFilterHeader(allEmployees, displayedEmployees.length),
              ];
            },
            body: _buildBody(state, displayedEmployees),
          );
        },
      ),
    );
  }

  Widget _buildSliverAppBar(int count) {
    return SliverAppBar(
      expandedHeight: 180,
      pinned: true,
      backgroundColor: AppColors.primaryRed,
      elevation: 0,
      leading: IconButton(
        icon: const Icon(Icons.arrow_back_ios_new_rounded,
            color: Colors.white, size: 20),
        onPressed: () => Navigator.pop(context),
      ),
      title: Text('DIREKTORI KARYAWAN',
          style: GoogleFonts.plusJakartaSans(
              fontSize: 15,
              fontWeight: FontWeight.w800,
              letterSpacing: 1.2,
              color: Colors.white)),
      centerTitle: true,
      flexibleSpace: FlexibleSpaceBar(
        background: Container(
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              colors: AppColors.primaryGradient,
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const SizedBox(height: 40),
              Text(
                '$count',
                style: GoogleFonts.plusJakartaSans(
                  color: Colors.white,
                  fontSize: 48,
                  fontWeight: FontWeight.w900,
                ),
              ),
              Text(
                'TOTAL SELURUH KARYAWAN',
                style: GoogleFonts.plusJakartaSans(
                  color: Colors.white.withValues(alpha: 0.85),
                  fontSize: 10,
                  fontWeight: FontWeight.w800,
                  letterSpacing: 1.5,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSearchAndFilterHeader(List<EmployeeDirectory> allEmployees, int displayedCount) {
    // Extract unique branches
    final branchSet = <String>{'Semua Cabang'};
    for (final emp in allEmployees) {
      if (emp.branchName != null && emp.branchName!.trim().isNotEmpty) {
        branchSet.add(emp.branchName!.trim());
      }
    }
    final branchList = branchSet.toList();

    return SliverToBoxAdapter(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Search input bar
          Container(
            padding: const EdgeInsets.fromLTRB(20, 20, 20, 10),
            child: Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.04),
                    blurRadius: 10,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: TextField(
                controller: _searchController,
                onChanged: (val) {
                  if (_debounce?.isActive ?? false) _debounce!.cancel();
                  _debounce = Timer(const Duration(milliseconds: 400), () {
                    context
                        .read<DirectoryBloc>()
                        .add(FetchDirectoryRequested(query: val));
                  });
                },
                style: GoogleFonts.plusJakartaSans(
                    fontSize: 14, fontWeight: FontWeight.w600),
                decoration: InputDecoration(
                  hintText: 'Cari nama, jabatan, dept, cabang...',
                  hintStyle: GoogleFonts.plusJakartaSans(
                      color: AppColors.textTertiary,
                      fontSize: 13,
                      fontWeight: FontWeight.w500),
                  prefixIcon: const Icon(Icons.search_rounded,
                      color: AppColors.primaryRed, size: 22),
                  suffixIcon: _searchController.text.isNotEmpty
                      ? IconButton(
                          icon: const Icon(Icons.clear_rounded, size: 18),
                          onPressed: () {
                            _searchController.clear();
                            context
                                .read<DirectoryBloc>()
                                .add(const FetchDirectoryRequested());
                          },
                        )
                      : null,
                  border: InputBorder.none,
                  contentPadding: const EdgeInsets.symmetric(vertical: 14),
                ),
              ),
            ),
          ),

          // Horizontal Branch Filter Chips
          if (branchList.length > 1)
            Container(
              height: 42,
              margin: const EdgeInsets.only(bottom: 12),
              child: ListView.separated(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 20),
                itemCount: branchList.length,
                separatorBuilder: (_, __) => const SizedBox(width: 8),
                itemBuilder: (context, index) {
                  final branch = branchList[index];
                  final isSelected = _selectedBranch == branch;
                  return ChoiceChip(
                    label: Text(
                      branch,
                      style: GoogleFonts.plusJakartaSans(
                        fontSize: 12,
                        fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
                        color: isSelected ? Colors.white : AppColors.textSecondary,
                      ),
                    ),
                    selected: isSelected,
                    selectedColor: AppColors.primaryRed,
                    backgroundColor: Colors.white,
                    showCheckmark: false,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20),
                      side: BorderSide(
                        color: isSelected
                            ? AppColors.primaryRed
                            : const Color(0xFFE2E8F0),
                      ),
                    ),
                    onSelected: (selected) {
                      if (selected) {
                        setState(() {
                          _selectedBranch = branch;
                        });
                      }
                    },
                  );
                },
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildBody(DirectoryState state, List<EmployeeDirectory> employees) {
    if (state is DirectoryLoading) {
      return const Center(
          child: CircularProgressIndicator(color: AppColors.primaryRed));
    }

    if (state is DirectoryLoaded && employees.isEmpty) {
      return _buildEmptyState();
    }

    if (state is DirectoryFailure) {
      return Center(
          child: Text(state.message,
              style: GoogleFonts.plusJakartaSans(color: AppColors.error)));
    }

    return ListView.builder(
      padding: const EdgeInsets.fromLTRB(20, 0, 20, 40),
      itemCount: employees.length,
      itemBuilder: (context, index) => _buildProfessionalCard(employees[index]),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.person_search_rounded,
              size: 80, color: AppColors.textTertiary.withOpacity(0.3)),
          const SizedBox(height: 16),
          Text('Karyawan tidak ditemukan',
              style: GoogleFonts.plusJakartaSans(
                  fontSize: 14,
                  color: AppColors.textTertiary,
                  fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }

  Widget _buildProfessionalCard(EmployeeDirectory employee) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.03),
            blurRadius: 15,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(24),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              _buildAvatar(employee),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(employee.name,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: GoogleFonts.plusJakartaSans(
                            fontWeight: FontWeight.w900,
                            fontSize: 15,
                            color: AppColors.textPrimary)),
                    const SizedBox(height: 4),
                    _buildPositionBadge(employee.position),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        const Icon(Icons.business_rounded,
                            size: 14, color: AppColors.textTertiary),
                        const SizedBox(width: 6),
                        Expanded(
                          child: Text(
                            employee.branchName != null &&
                                    employee.branchName!.isNotEmpty
                                ? '${employee.department} • ${employee.branchName}'
                                : employee.department,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: GoogleFonts.plusJakartaSans(
                                color: AppColors.textTertiary,
                                fontSize: 11,
                                fontWeight: FontWeight.w700),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              _buildActionButtons(employee),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildAvatar(EmployeeDirectory employee) {
    return Container(
      padding: const EdgeInsets.all(3),
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        border:
            Border.all(color: AppColors.primaryRed.withOpacity(0.1), width: 2),
      ),
      child: CircleAvatar(
        radius: 32,
        backgroundColor: AppColors.grayLight,
        backgroundImage: (employee.profileUrl != null &&
                employee.profileUrl!.isNotEmpty)
            ? NetworkImage(employee.profileUrl!.startsWith('http')
                ? employee.profileUrl!
                : '${Uri.parse(AppConstants.baseUrl).origin}${employee.profileUrl!.startsWith('/') ? employee.profileUrl : '/${employee.profileUrl}'}')
            : null,
        child: (employee.profileUrl == null || employee.profileUrl!.isEmpty)
            ? const Icon(Icons.person_rounded,
                color: AppColors.textTertiary, size: 32)
            : null,
      ),
    );
  }

  Widget _buildPositionBadge(String position) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: BoxDecoration(
        color: AppColors.primaryRed.withOpacity(0.08),
        borderRadius: BorderRadius.circular(10),
      ),
      child: Text(
        position.toUpperCase(),
        style: GoogleFonts.plusJakartaSans(
          color: AppColors.primaryRed,
          fontSize: 9,
          fontWeight: FontWeight.w900,
          letterSpacing: 0.5,
        ),
      ),
    );
  }

  Widget _buildActionButtons(EmployeeDirectory employee) {
    return Column(
      children: [
        _buildCircularAction(
            Icons.phone_in_talk_rounded,
            const Color(0xFF3B82F6), // Professional Blue
            () => _callNumber(employee.phoneNumber)),
        const SizedBox(height: 12),
        _buildCircularAction(
            Icons.chat_bubble_rounded,
            const Color(0xFF10B981), // Success Green
            () => _whatsappNumber(employee.phoneNumber)),
      ],
    );
  }

  Widget _buildCircularAction(IconData icon, Color color, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        height: 40,
        width: 40,
        decoration: BoxDecoration(
          color: color.withOpacity(0.1),
          shape: BoxShape.circle,
        ),
        child: Icon(icon, color: color, size: 18),
      ),
    );
  }
}
