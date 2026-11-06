import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class AttendanceReportPage extends StatefulWidget {
  const AttendanceReportPage({super.key});

  @override
  State<AttendanceReportPage> createState() => _AttendanceReportPageState();
}

class _AttendanceReportPageState extends State<AttendanceReportPage> {
  int _bottomNavIndex = 1; // 1 = Reports

  @override
  Widget build(BuildContext context) {
    const primaryBlue = Color(0xFF1B60F1);
    const bgGray = Color(0xFFF9FAFB);
    const textColor = Color(0xFF111827);
    const subtitleColor = Color(0xFF6B7280);

    return Scaffold(
      backgroundColor: bgGray,
      appBar: AppBar(
        title: Column(
          children: [
            Text(
              'Perusahaan Demo',
              style: GoogleFonts.inter(
                fontSize: 10,
                fontWeight: FontWeight.bold,
                color: subtitleColor,
                letterSpacing: 1.0,
              ),
            ),
            const SizedBox(height: 2),
            Text(
              'Laporan Absensi & Lembur',
              style: GoogleFonts.inter(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: textColor,
              ),
            ),
          ],
        ),
        centerTitle: true,
        backgroundColor: Colors.white,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(
            Icons.arrow_back_ios_new,
            color: primaryBlue,
            size: 20,
          ),
          onPressed: () => Navigator.of(context).pop(),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.more_horiz, color: primaryBlue),
            onPressed: () {},
          ),
        ],
      ),
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(vertical: 16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Filters Row
              SingleChildScrollView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 8,
                      ),
                      decoration: BoxDecoration(
                        color: primaryBlue,
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Row(
                        children: [
                          const Icon(
                            Icons.filter_list,
                            color: Colors.white,
                            size: 16,
                          ),
                          const SizedBox(width: 8),
                          Text(
                            'All Depts',
                            style: GoogleFonts.inter(
                              color: Colors.white,
                              fontSize: 13,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          const SizedBox(width: 4),
                          const Icon(
                            Icons.keyboard_arrow_down,
                            color: Colors.white,
                            size: 16,
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(width: 12),
                    _buildFilterChip('IT Support'),
                    const SizedBox(width: 12),
                    _buildFilterChip('Marketing'),
                  ],
                ),
              ),
              const SizedBox(height: 20),

              // Calendar Strip Card
              Container(
                margin: const EdgeInsets.symmetric(horizontal: 16),
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
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
                        const Icon(Icons.chevron_left, color: subtitleColor),
                        Text(
                          'October 2026',
                          style: GoogleFonts.inter(
                            fontSize: 15,
                            fontWeight: FontWeight.bold,
                            color: textColor,
                          ),
                        ),
                        const Icon(Icons.chevron_right, color: subtitleColor),
                      ],
                    ),
                    const SizedBox(height: 16),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        _buildDayHeader('S'),
                        _buildDayHeader('M'),
                        _buildDayHeader('T'),
                        _buildDayHeader('W'),
                        _buildDayHeader('T'),
                        _buildDayHeader('F'),
                        _buildDayHeader('S'),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        _buildDateItem('24', false),
                        _buildDateItem('25', false),
                        _buildDateItem('26', false),
                        _buildDateItem('27', false),
                        _buildDateItem('28', false),
                        _buildDateItem('29', false),
                        _buildDateItem('30', false),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        _buildDateItem('1', true, isPale: true),
                        _buildDateItem('2', true),
                        _buildDateItem('3', true),
                        _buildDateItem('4', true),
                        _buildDateItem('5', true),
                        _buildDateItem('6', true),
                        _buildDateItem('7', true, isPale: true),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),

              // Summary Header
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Employee Summary (12)',
                      style: GoogleFonts.inter(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: textColor,
                      ),
                    ),
                    Text(
                      'Oct 1 - Oct 31',
                      style: GoogleFonts.inter(
                        fontSize: 12,
                        color: subtitleColor,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 12),

              // Employee Cards
              _buildEmployeeCard(
                name: 'Marcus Holloway',
                role: 'Senior Web Developer',
                avatarUrl:
                    'https://ui-avatars.com/api/?name=Marcus&background=random',
                present: '22 Days',
                absences: '0',
                overtime: '14.5h',
                isExpanded: false,
              ),

              _buildEmployeeCard(
                name: 'Sarah Chen',
                role: 'UX Designer',
                avatarUrl:
                    'https://ui-avatars.com/api/?name=Sarah&background=random',
                present: '19 Days',
                absences: '3',
                overtime: '8.0h',
                isExpanded: true,
              ),

              _buildEmployeeCard(
                name: 'Jason Vor',
                role: 'Marketing Operations',
                avatarUrl:
                    'https://ui-avatars.com/api/?name=Jason&background=random',
                present: '21 Days',
                absences: '1',
                overtime: '22.0h',
                isExpanded: false,
              ),
            ],
          ),
        ),
      ),
      bottomNavigationBar: Container(
        decoration: BoxDecoration(
          boxShadow: [
            BoxShadow(
              color: primaryBlue.withOpacity(0.15),
              spreadRadius: 0,
              blurRadius: 20,
              offset: const Offset(0, -5),
            ),
          ],
        ),
        child: ClipRRect(
          borderRadius: const BorderRadius.only(
            topLeft: Radius.circular(20),
            topRight: Radius.circular(20),
          ),
          child: BottomNavigationBar(
            currentIndex: _bottomNavIndex,
            onTap: (i) => setState(() => _bottomNavIndex = i),
            type: BottomNavigationBarType.fixed,
            selectedItemColor: primaryBlue,
            unselectedItemColor: subtitleColor,
            showUnselectedLabels: true,
            selectedLabelStyle: GoogleFonts.inter(
              fontSize: 10,
              fontWeight: FontWeight.bold,
            ),
            unselectedLabelStyle: GoogleFonts.inter(fontSize: 10),
            items: const [
              BottomNavigationBarItem(
                icon: Icon(Icons.home_outlined),
                label: 'Beranda',
              ),
              BottomNavigationBarItem(
                icon: Icon(Icons.description),
                label: 'Laporan',
              ),
              BottomNavigationBarItem(
                icon: Icon(Icons.people_outline),
                label: 'Karyawan',
              ),
              BottomNavigationBarItem(
                icon: Icon(Icons.settings_outlined),
                label: 'Pengaturan',
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFilterChip(String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: const Color(0xFFE5E7EB)),
      ),
      child: Text(
        label,
        style: GoogleFonts.inter(
          color: const Color(0xFF4B5563),
          fontSize: 13,
          fontWeight: FontWeight.w600,
        ),
      ),
    );
  }

  Widget _buildDayHeader(String day) {
    return SizedBox(
      width: 36,
      child: Center(
        child: Text(
          day,
          style: GoogleFonts.inter(
            fontSize: 12,
            fontWeight: FontWeight.bold,
            color: const Color(0xFF9CA3AF),
          ),
        ),
      ),
    );
  }

  Widget _buildDateItem(String date, bool isSelected, {bool isPale = false}) {
    const primaryBlue = Color(0xFF1B60F1);
    Color bgColor = Colors.transparent;
    Color textColor = const Color(0xFFE5E7EB);

    if (isSelected) {
      if (isPale) {
        bgColor = const Color(0xFFEBF2FF);
        textColor = primaryBlue;
      } else {
        bgColor = primaryBlue;
        textColor = Colors.white;
      }
    }

    return Container(
      width: 36,
      height: 36,
      decoration: BoxDecoration(color: bgColor, shape: BoxShape.circle),
      child: Center(
        child: Text(
          date,
          style: GoogleFonts.inter(
            fontSize: 14,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
            color: textColor,
          ),
        ),
      ),
    );
  }

  Widget _buildEmployeeCard({
    required String name,
    required String role,
    required String avatarUrl,
    required String present,
    required String absences,
    required String overtime,
    required bool isExpanded,
  }) {
    const primaryBlue = Color(0xFF1B60F1);

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFF3F4F6)),
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
          // Header Row
          Padding(
            padding: const EdgeInsets.all(16.0),
            child: Row(
              children: [
                CircleAvatar(
                  radius: 20,
                  backgroundImage: NetworkImage(avatarUrl),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        name,
                        style: GoogleFonts.inter(
                          fontSize: 15,
                          fontWeight: FontWeight.bold,
                          color: const Color(0xFF111827),
                        ),
                      ),
                      Text(
                        role,
                        style: GoogleFonts.inter(
                          fontSize: 12,
                          color: const Color(0xFF6B7280),
                        ),
                      ),
                    ],
                  ),
                ),
                Icon(
                  isExpanded ? Icons.keyboard_arrow_down : Icons.chevron_right,
                  color: const Color(0xFF9CA3AF),
                ),
              ],
            ),
          ),

          if (isExpanded) ...[
            Container(
              margin: const EdgeInsets.symmetric(horizontal: 16),
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: const Color(0xFFF8FAFC),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'RECENT DAILY LOGS',
                    style: GoogleFonts.inter(
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                      color: primaryBlue,
                      letterSpacing: 0.5,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Oct 06, 2026',
                        style: GoogleFonts.inter(
                          fontSize: 12,
                          color: const Color(0xFF4B5563),
                        ),
                      ),
                      Row(
                        children: [
                          Text(
                            '08:55 - 17:30',
                            style: GoogleFonts.inter(
                              fontSize: 12,
                              color: const Color(0xFF111827),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Text(
                            '+0.5h',
                            style: GoogleFonts.inter(
                              fontSize: 12,
                              fontWeight: FontWeight.bold,
                              color: primaryBlue,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        'Oct 05, 2026',
                        style: GoogleFonts.inter(
                          fontSize: 12,
                          color: const Color(0xFF4B5563),
                        ),
                      ),
                      Text(
                        'Leave (Medical)',
                        style: GoogleFonts.inter(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          color: const Color(0xFFEF4444),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
          ],

          const Divider(height: 1, color: Color(0xFFF3F4F6)),

          // Stats Row
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 12.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                _buildStatColumn('PRESENT', present, primaryBlue),
                Container(width: 1, height: 24, color: const Color(0xFFF3F4F6)),
                _buildStatColumn('ABSENCES', absences, const Color(0xFFEF4444)),
                Container(width: 1, height: 24, color: const Color(0xFFF3F4F6)),
                _buildStatColumn('OVERTIME', overtime, const Color(0xFF111827)),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatColumn(String label, String value, Color valueColor) {
    return Column(
      children: [
        Text(
          label,
          style: GoogleFonts.inter(
            fontSize: 10,
            fontWeight: FontWeight.bold,
            color: const Color(0xFF9CA3AF),
            letterSpacing: 0.5,
          ),
        ),
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
}
