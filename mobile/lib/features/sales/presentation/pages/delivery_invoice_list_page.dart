import 'package:cached_network_image/cached_network_image.dart';
import 'dart:io' show File, Directory;
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter/material.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:intl/intl.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:url_launcher/url_launcher.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart' show networkImage, Printing;
import 'package:share_plus/share_plus.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import './tempo_payment_page.dart';

class DeliveryInvoiceListPage extends StatefulWidget {
  final String paymentMethod; // 'CASH', 'TRANSFER', 'TEMPO'
  final List<dynamic> tasks;

  const DeliveryInvoiceListPage({
    super.key,
    required this.paymentMethod,
    required this.tasks,
  });

  @override
  State<DeliveryInvoiceListPage> createState() =>
      _DeliveryInvoiceListPageState();
}

class _DeliveryInvoiceListPageState extends State<DeliveryInvoiceListPage> {
  late List<dynamic> _allInvoices;
  List<dynamic> _filteredInvoices = [];
  final TextEditingController _searchController = TextEditingController();
  String _statusFilter = 'ALL'; // 'ALL', 'PAID', 'UNPAID'
  bool _isGeneratingPdf = false;

  @override
  void initState() {
    super.initState();
    _loadInvoices();
    _searchController.addListener(_applyFilters);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  bool _isInvoicePaid(dynamic inv) {
    final orderData = inv['sales_transaction'] ?? inv['sales_order'] ?? {};
    final payStatus =
        (orderData['payment_status'] ??
                inv['payment_status'] ??
                inv['status'] ??
                'UNPAID')
            .toString()
            .toUpperCase();
    if (payStatus == 'PAID' || payStatus == 'SUCCESS' || payStatus == 'LUNAS') {
      return true;
    }

    // Check if collected amount equals total amount (fully paid cash)
    final paidAmount =
        (orderData['paid_amount'] ??
                orderData['payment_collected_amount'] ??
                0.0)
            .toDouble();
    final totalAmount = (orderData['total_amount'] ?? 0.0).toDouble();
    if (totalAmount > 0 && paidAmount >= totalAmount) {
      return true;
    }

    return false;
  }

  void _loadInvoices() {
    List<dynamic> matched = [];
    for (var task in widget.tasks) {
      final List items = task['items'] as List? ?? [];
      for (var item in items) {
        final orderData =
            item['sales_transaction'] ?? item['sales_order'] ?? {};

        // Get effective payment method
        String method = (item['payment_method'] ?? '').toString().toUpperCase();
        if (method.isEmpty || method == 'PENDING') {
          method = (orderData['payment_method'] ?? '').toString().toUpperCase();
        }

        bool isMatch = false;
        if (widget.paymentMethod == 'CASH') {
          isMatch = (method == 'CASH' || method == 'COD');
        } else if (widget.paymentMethod == 'TRANSFER') {
          isMatch =
              (method == 'QRIS' ||
              method == 'VA' ||
              method == 'TRANSFER' ||
              method == 'MIDTRANS_QRIS' ||
              method == 'MIDTRANS_VA');
        } else if (widget.paymentMethod == 'TEMPO') {
          isMatch = (method == 'TEMPO' || method == 'CREDIT');
        }

        if (isMatch) {
          final Map<String, dynamic> enrichedItem = Map<String, dynamic>.from(
            item,
          );
          enrichedItem['_delivery_order_no'] =
              task['delivery_order_no'] ?? 'SJ-PENDING';
          matched.add(enrichedItem);
        }
      }
    }
    setState(() {
      _allInvoices = matched;
      _filteredInvoices = matched;
    });
  }

  void _applyFilters() {
    final query = _searchController.text.toLowerCase();
    setState(() {
      _filteredInvoices = _allInvoices.where((inv) {
        final orderData = inv['sales_transaction'] ?? inv['sales_order'] ?? {};
        final storeName = (orderData['store']?['name'] ?? '')
            .toString()
            .toLowerCase();
        final invoiceNo =
            (orderData['invoice_no'] ?? orderData['so_number'] ?? '')
                .toString()
                .toLowerCase();
        final sjNo = (inv['_delivery_order_no'] ?? '').toString().toLowerCase();

        // Search filter
        final matchesSearch =
            storeName.contains(query) ||
            invoiceNo.contains(query) ||
            sjNo.contains(query);

        final isPaid = _isInvoicePaid(inv);

        bool matchesStatus = true;
        if (_statusFilter == 'PAID') {
          matchesStatus = isPaid;
        } else if (_statusFilter == 'UNPAID') {
          matchesStatus = !isPaid;
        }

        return matchesSearch && matchesStatus;
      }).toList();
    });
  }

  double _calculateTotalAmount() {
    double total = 0.0;
    for (var inv in _filteredInvoices) {
      final orderData = inv['sales_transaction'] ?? inv['sales_order'] ?? {};
      total += (orderData['total_amount'] ?? inv['payment_amount'] ?? 0.0)
          .toDouble();
    }
    return total;
  }

  Color _getThemeColor() {
    if (widget.paymentMethod == 'CASH') return Colors.green;
    if (widget.paymentMethod == 'TRANSFER') return AppColors.blueDelivery;
    return Colors.deepOrange;
  }

  String _getPageTitle() {
    if (widget.paymentMethod == 'CASH') return 'Faktur Kas COD (Cash)';
    if (widget.paymentMethod == 'TRANSFER') return 'Faktur Transfer & Digital';
    return 'Faktur Tagihan Tempo';
  }

  @override
  Widget build(BuildContext context) {
    final themeColor = _getThemeColor();
    final formatter = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    final totalAmount = _calculateTotalAmount();

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        elevation: 0,
        backgroundColor: const Color(0xFF0F172A),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_rounded, color: Colors.white),
          onPressed: () => Navigator.pop(context),
        ),
        title: Text(
          _getPageTitle(),
          style: GoogleFonts.outfit(
            fontWeight: FontWeight.w900,
            color: Colors.white,
            fontSize: 16,
          ),
        ),
        centerTitle: true,
      ),
      body: Column(
        children: [
          // Elegant Header Summary Card
          Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
            decoration: const BoxDecoration(
              color: Color(0xFF0F172A),
              borderRadius: BorderRadius.only(
                bottomLeft: Radius.circular(32),
                bottomRight: Radius.circular(32),
              ),
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
                          'TOTAL NOMINAL',
                          style: GoogleFonts.outfit(
                            color: Colors.white.withOpacity(0.6),
                            fontWeight: FontWeight.w800,
                            fontSize: 10,
                            letterSpacing: 1,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          formatter.format(totalAmount),
                          style: GoogleFonts.outfit(
                            color: Colors.white,
                            fontWeight: FontWeight.w900,
                            fontSize: 24,
                          ),
                        ),
                      ],
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 8,
                      ),
                      decoration: BoxDecoration(
                        color: themeColor.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: themeColor.withOpacity(0.4)),
                      ),
                      child: Text(
                        '${_filteredInvoices.length} Faktur',
                        style: GoogleFonts.outfit(
                          color: Colors.white,
                          fontWeight: FontWeight.w800,
                          fontSize: 12,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
              ],
            ),
          ),

          // Filters and Search Bar
          Padding(
            padding: const EdgeInsets.fromLTRB(20, 16, 20, 8),
            child: Column(
              children: [
                // Search Input
                Container(
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
                  child: TextField(
                    controller: _searchController,
                    decoration: InputDecoration(
                      hintText: 'Cari nama toko, nomor faktur/SJ...',
                      hintStyle: GoogleFonts.outfit(
                        fontSize: 13,
                        color: Colors.blueGrey.withOpacity(0.6),
                      ),
                      prefixIcon: Icon(
                        Icons.search_rounded,
                        color: Colors.blueGrey.withOpacity(0.6),
                      ),
                      border: InputBorder.none,
                      contentPadding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 14,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 12),

                // Status Filter Chips
                Row(
                  children: [
                    _buildFilterChip('ALL', 'Semua'),
                    const SizedBox(width: 8),
                    _buildFilterChip('PAID', 'Lunas'),
                    const SizedBox(width: 8),
                    _buildFilterChip('UNPAID', 'Belum Lunas'),
                  ],
                ),
              ],
            ),
          ),

          // Invoices List
          Expanded(
            child: _filteredInvoices.isEmpty
                ? _buildEmptyState()
                : ListView.builder(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 20,
                      vertical: 8,
                    ),
                    physics: const BouncingScrollPhysics(),
                    itemCount: _filteredInvoices.length,
                    itemBuilder: (context, index) {
                      return _buildInvoiceCard(
                        context,
                        _filteredInvoices[index],
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildFilterChip(String filterValue, String label) {
    final isSelected = _statusFilter == filterValue;
    final themeColor = _getThemeColor();

    return ChoiceChip(
      label: Text(
        label,
        style: GoogleFonts.outfit(
          fontSize: 12,
          fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
          color: isSelected ? Colors.white : const Color(0xFF64748B),
        ),
      ),
      selected: isSelected,
      onSelected: (selected) {
        if (selected) {
          setState(() {
            _statusFilter = filterValue;
            _applyFilters();
          });
        }
      },
      selectedColor: themeColor,
      backgroundColor: Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(12),
        side: BorderSide(
          color: isSelected ? themeColor : const Color(0xFFE2E8F0),
          width: 1,
        ),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
    );
  }

  Widget _buildInvoiceCard(BuildContext context, dynamic inv) {
    final orderData = inv['sales_transaction'] ?? inv['sales_order'] ?? {};
    final storeName = orderData['store']?['name'] ?? 'Toko Tanpa Nama';
    final invoiceNo =
        orderData['invoice_no'] ?? orderData['so_number'] ?? 'INV-PENDING';
    final sjNo = inv['_delivery_order_no'] ?? 'SJ-PENDING';
    final amount = (orderData['total_amount'] ?? inv['payment_amount'] ?? 0.0)
        .toDouble();
    final formatter = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );

    final isPaid = _isInvoicePaid(inv);

    Color statusColor = isPaid ? Colors.green : Colors.orange;
    String statusLabel = isPaid ? 'LUNAS' : 'BELUM LUNAS';

    // Build subtext based on payment type
    String paymentDetail = '';
    if (widget.paymentMethod == 'TRANSFER') {
      final bank = (orderData['midtrans_bank'] ?? 'MIDTRANS')
          .toString()
          .toUpperCase();
      final va =
          orderData['midtrans_va_number'] ??
          orderData['midtrans_bill_key'] ??
          '';
      paymentDetail = va.isNotEmpty ? '$bank VA: $va' : 'Transfer Digital';
    } else if (widget.paymentMethod == 'TEMPO') {
      if (orderData['payment_due_date'] != null) {
        final date = DateTime.tryParse(
          orderData['payment_due_date'].toString(),
        );
        if (date != null) {
          paymentDetail =
              'Jatuh Tempo: ${DateFormat('dd MMM yyyy').format(date)}';
        }
      } else {
        paymentDetail = 'Tempo / Kredit';
      }
    } else {
      paymentDetail = 'Tunai / Cash COD';
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.02),
            blurRadius: 10,
            offset: const Offset(0, 4),
          ),
        ],
        border: Border.all(color: const Color(0xFFF1F5F9)),
      ),
      child: InkWell(
        onTap: () => _showInvoiceDetailModal(context, inv),
        borderRadius: BorderRadius.circular(24),
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header of card (Faktur # & Status)
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          invoiceNo,
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w900,
                            fontSize: 14,
                            color: const Color(0xFF1E293B),
                          ),
                        ),
                        Text(
                          'Surat Jalan: #$sjNo',
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w600,
                            fontSize: 10,
                            color: Colors.blueGrey,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 4,
                    ),
                    decoration: BoxDecoration(
                      color: statusColor.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Text(
                      statusLabel,
                      style: GoogleFonts.outfit(
                        fontSize: 9,
                        fontWeight: FontWeight.w900,
                        color: statusColor,
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),

              // Middle of card (Store Name & Payment Details)
              Row(
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: _getThemeColor().withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      widget.paymentMethod == 'CASH'
                          ? Icons.payments_rounded
                          : widget.paymentMethod == 'TRANSFER'
                          ? Icons.account_balance_wallet_rounded
                          : Icons.calendar_month_rounded,
                      color: _getThemeColor(),
                      size: 20,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          storeName,
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w800,
                            fontSize: 15,
                            color: const Color(0xFF0F172A),
                          ),
                        ),
                        Text(
                          paymentDetail,
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w500,
                            fontSize: 11,
                            color: Colors.blueGrey,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),

              // Bottom of card (Total Amount & Quick Action Buttons)
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'NILAI FAKTUR',
                        style: GoogleFonts.outfit(
                          fontSize: 9,
                          fontWeight: FontWeight.w800,
                          color: Colors.blueGrey,
                          letterSpacing: 0.5,
                        ),
                      ),
                      Text(
                        formatter.format(amount),
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w900,
                          fontSize: 17,
                          color: _getThemeColor(),
                        ),
                      ),
                    ],
                  ),
                  Row(
                    children: [
                      // Print/Download Quick Action
                      IconButton(
                        icon: const Icon(
                          Icons.download_rounded,
                          color: Colors.blueAccent,
                        ),
                        style: IconButton.styleFrom(
                          backgroundColor: Colors.blueAccent.withOpacity(0.1),
                          padding: const EdgeInsets.all(8),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(10),
                          ),
                        ),
                        onPressed: () => _downloadInvoicePdf(context, inv),
                      ),
                      const SizedBox(width: 8),
                      // Share to WhatsApp Quick Action
                      IconButton(
                        icon: const Icon(
                          Icons.share_rounded,
                          color: Colors.green,
                        ),
                        style: IconButton.styleFrom(
                          backgroundColor: Colors.green.withOpacity(0.1),
                          padding: const EdgeInsets.all(8),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(10),
                          ),
                        ),
                        onPressed: () => _shareToWhatsApp(context, inv),
                      ),
                    ],
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.receipt_long_rounded,
            size: 80,
            color: Colors.grey.shade300,
          ),
          const SizedBox(height: 16),
          Text(
            'Tidak ada faktur ditemukan',
            style: GoogleFonts.outfit(
              fontSize: 16,
              fontWeight: FontWeight.w800,
              color: const Color(0xFF475569),
            ),
          ),
          const SizedBox(height: 4),
          Text(
            'Sesuaikan kata kunci pencarian atau filter status.',
            style: GoogleFonts.outfit(fontSize: 12, color: Colors.blueGrey),
          ),
        ],
      ),
    );
  }

  void _showTempoPaymentDialog(BuildContext context, dynamic inv) {
    final orderData = inv['sales_transaction'] ?? inv['sales_order'] ?? {};

    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => TempoPaymentPage(
          transactionData: inv,
          onSuccess: (method, amount, signatureUrl) async {
            Navigator.pop(context); // Close payment page
            Navigator.pop(context); // Close invoice modal

            ScaffoldMessenger.of(context).showSnackBar(
              const SnackBar(
                content: Text('Pembayaran berhasil dicatat!'),
                backgroundColor: Colors.green,
              ),
            );

            setState(() {
              orderData['payment_method'] = method;
              orderData['payment_status'] = 'PAID';
              orderData['status'] = 'PAID';
              if (signatureUrl != null) {
                orderData['signature_image_url'] = signatureUrl;
              }
            });

            // Launch WhatsApp automatically to share receipt
            await _shareToWhatsApp(context, inv);
          },
        ),
      ),
    );
  }

  void _showInvoiceDetailModal(BuildContext context, dynamic inv) {
    final orderData = inv['sales_transaction'] ?? inv['sales_order'] ?? {};
    final storeName = orderData['store']?['name'] ?? 'Toko Tanpa Nama';
    final storeAddress =
        orderData['store']?['address'] ?? 'Alamat tidak tersedia';
    final invoiceNo =
        orderData['invoice_no'] ?? orderData['so_number'] ?? 'INV-PENDING';
    final formatter = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    final amount = (orderData['total_amount'] ?? inv['payment_amount'] ?? 0.0)
        .toDouble();
    final items = orderData['items'] as List? ?? [];

    final isPaid = _isInvoicePaid(inv);

    final midtransQrisUrl = orderData['midtrans_qris_url'] ?? '';
    final midtransVaNumber =
        orderData['midtrans_va_number'] ?? orderData['midtrans_bill_key'] ?? '';
    final midtransBank = (orderData['midtrans_bank'] ?? 'Bank')
        .toString()
        .toUpperCase();
    final midtransBillerCode = orderData['midtrans_biller_code'];

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) {
        return Container(
          height: MediaQuery.of(context).size.height * 0.85,
          decoration: const BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.only(
              topLeft: Radius.circular(32),
              topRight: Radius.circular(32),
            ),
          ),
          child: Column(
            children: [
              // Drag Indicator & Modal Title
              Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.symmetric(vertical: 12),
                decoration: BoxDecoration(
                  color: Colors.grey.shade300,
                  borderRadius: BorderRadius.circular(10),
                ),
              ),
              Padding(
                padding: const EdgeInsets.symmetric(
                  horizontal: 20,
                  vertical: 8,
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Rincian Faktur Penjualan',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w900,
                        fontSize: 18,
                        color: const Color(0xFF0F172A),
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.close_rounded),
                      onPressed: () => Navigator.pop(context),
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),

              // Scrollable content
              Expanded(
                child: ListView(
                  padding: const EdgeInsets.all(20),
                  children: [
                    // Invoice Status Header Card
                    Container(
                      padding: const EdgeInsets.all(20),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(24),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'Nomor Faktur',
                                    style: GoogleFonts.outfit(
                                      fontSize: 10,
                                      color: Colors.blueGrey,
                                      fontWeight: FontWeight.w800,
                                    ),
                                  ),
                                  Text(
                                    invoiceNo,
                                    style: GoogleFonts.outfit(
                                      fontSize: 16,
                                      fontWeight: FontWeight.w900,
                                      color: const Color(0xFF1E293B),
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
                                  color: (isPaid ? Colors.green : Colors.orange)
                                      .withOpacity(0.12),
                                  borderRadius: BorderRadius.circular(10),
                                ),
                                child: Text(
                                  isPaid ? 'LUNAS' : 'BELUM LUNAS',
                                  style: GoogleFonts.outfit(
                                    fontSize: 10,
                                    fontWeight: FontWeight.w900,
                                    color: isPaid
                                        ? Colors.green
                                        : Colors.orange,
                                  ),
                                ),
                              ),
                            ],
                          ),
                          const Divider(height: 24, thickness: 1),
                          // Payment detail
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text(
                                'Metode Bayar:',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  color: Colors.blueGrey,
                                  fontWeight: FontWeight.w700,
                                ),
                              ),
                              Text(
                                widget.paymentMethod == 'CASH'
                                    ? 'CASH (COD)'
                                    : widget.paymentMethod == 'TRANSFER'
                                    ? 'TRANSFER DIGITAL / MIDTRANS'
                                    : 'TEMPO / KREDIT',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  color: const Color(0xFF1E293B),
                                  fontWeight: FontWeight.w900,
                                ),
                              ),
                            ],
                          ),
                          if (widget.paymentMethod == 'TEMPO' &&
                              orderData['payment_due_date'] != null) ...[
                            const SizedBox(height: 8),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'Jatuh Tempo:',
                                  style: GoogleFonts.outfit(
                                    fontSize: 12,
                                    color: Colors.red,
                                    fontWeight: FontWeight.w700,
                                  ),
                                ),
                                Text(
                                  DateFormat('dd MMMM yyyy').format(
                                    DateTime.parse(
                                      orderData['payment_due_date'],
                                    ).toLocal(),
                                  ),
                                  style: GoogleFonts.outfit(
                                    fontSize: 12,
                                    color: Colors.red,
                                    fontWeight: FontWeight.w900,
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ],
                      ),
                    ),
                    const SizedBox(height: 20),

                    // Midtrans section for Transfer method
                    if (widget.paymentMethod == 'TRANSFER') ...[
                      Text(
                        'DATA MIDTRANS INTEGRATION',
                        style: GoogleFonts.outfit(
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                          color: Colors.blueGrey,
                          letterSpacing: 1,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Container(
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: AppColors.blueDelivery.withOpacity(0.05),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(
                            color: AppColors.blueDelivery.withOpacity(0.1),
                          ),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            if (midtransVaNumber.isNotEmpty) ...[
                              Text(
                                'Virtual Account ($midtransBank)',
                                style: GoogleFonts.outfit(
                                  fontSize: 10,
                                  color: Colors.blueGrey,
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                              const SizedBox(height: 4),
                              Row(
                                mainAxisAlignment:
                                    MainAxisAlignment.spaceBetween,
                                children: [
                                  Text(
                                    midtransVaNumber,
                                    style: GoogleFonts.outfit(
                                      fontSize: 18,
                                      fontWeight: FontWeight.w900,
                                      color: const Color(0xFF0F172A),
                                    ),
                                  ),
                                  IconButton(
                                    icon: const Icon(
                                      Icons.copy_rounded,
                                      size: 16,
                                    ),
                                    onPressed: () {
                                      // Copy to clipboard
                                    },
                                  ),
                                ],
                              ),
                              if (midtransBillerCode != null) ...[
                                const SizedBox(height: 4),
                                Text(
                                  'Kode Biller: $midtransBillerCode',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    fontWeight: FontWeight.w700,
                                    color: Colors.blueGrey,
                                  ),
                                ),
                              ],
                            ] else if (midtransQrisUrl.isNotEmpty) ...[
                              Center(
                                child: Column(
                                  children: [
                                    midtransQrisUrl.startsWith('http')
                                        ? CachedNetworkImage(
                                            imageUrl:
                                                '${AppConstants.baseUrl}sales/qris-image?url=${Uri.encodeComponent(midtransQrisUrl)}',
                                            height: 160,
                                            width: 160,
                                            fit: BoxFit.contain,
                                            placeholder: (context, url) =>
                                                const Center(
                                                  child:
                                                      CircularProgressIndicator(),
                                                ),
                                            errorWidget:
                                                (context, url, error) =>
                                                    const Icon(Icons.error),
                                            memCacheWidth: 800,
                                          )
                                        : QrImageView(
                                            data: midtransQrisUrl,
                                            version: QrVersions.auto,
                                            size: 160.0,
                                          ),
                                    const SizedBox(height: 8),
                                    Text(
                                      'Barcode QRIS Midtrans',
                                      style: GoogleFonts.outfit(
                                        fontSize: 11,
                                        fontWeight: FontWeight.w800,
                                        color: AppColors.blueDelivery,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ] else ...[
                              Text(
                                'Tidak ada link Virtual Account atau QRIS aktif dari Midtrans.',
                                style: GoogleFonts.outfit(
                                  fontSize: 12,
                                  color: Colors.blueGrey,
                                  fontStyle: FontStyle.italic,
                                ),
                              ),
                            ],
                            const Divider(height: 20),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  'Status Transfer:',
                                  style: GoogleFonts.outfit(
                                    fontSize: 12,
                                    fontWeight: FontWeight.w800,
                                    color: Colors.blueGrey,
                                  ),
                                ),
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 10,
                                    vertical: 4,
                                  ),
                                  decoration: BoxDecoration(
                                    color:
                                        (isPaid ? Colors.green : Colors.orange)
                                            .withOpacity(0.12),
                                    borderRadius: BorderRadius.circular(8),
                                  ),
                                  child: Text(
                                    isPaid
                                        ? 'BERHASIL DITRANSFER'
                                        : 'MENUNGGU / PENDING',
                                    style: GoogleFonts.outfit(
                                      fontSize: 10,
                                      fontWeight: FontWeight.w900,
                                      color: isPaid
                                          ? Colors.green
                                          : Colors.orange,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                      const SizedBox(height: 20),
                    ],

                    // Store Info
                    Text(
                      'INFORMASI TOKO',
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        fontWeight: FontWeight.w900,
                        color: Colors.blueGrey,
                        letterSpacing: 1,
                      ),
                    ),
                    const SizedBox(height: 8),
                    ListTile(
                      contentPadding: EdgeInsets.zero,
                      leading: CircleAvatar(
                        backgroundColor: _getThemeColor().withOpacity(0.1),
                        child: Icon(
                          Icons.store_rounded,
                          color: _getThemeColor(),
                        ),
                      ),
                      title: Text(
                        storeName,
                        style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w900,
                          fontSize: 16,
                          color: const Color(0xFF1E293B),
                        ),
                      ),
                      subtitle: Text(
                        storeAddress,
                        style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: Colors.blueGrey,
                        ),
                      ),
                    ),
                    const SizedBox(height: 20),

                    // Products Table
                    Text(
                      'DAFTAR BARANG YANG DIKIRIM',
                      style: GoogleFonts.outfit(
                        fontSize: 11,
                        fontWeight: FontWeight.w900,
                        color: Colors.blueGrey,
                        letterSpacing: 1,
                      ),
                    ),
                    const SizedBox(height: 8),
                    if (items.isEmpty)
                      Text(
                        'Tidak ada produk tercatat di data lokal.',
                        style: GoogleFonts.outfit(
                          fontSize: 12,
                          fontStyle: FontStyle.italic,
                          color: Colors.blueGrey,
                        ),
                      )
                    else
                      Container(
                        decoration: BoxDecoration(
                          color: const Color(0xFFF8FAFC),
                          borderRadius: BorderRadius.circular(20),
                          border: Border.all(color: const Color(0xFFE2E8F0)),
                        ),
                        child: Column(
                          children: [
                            // Table Headers
                            Padding(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 16,
                                vertical: 12,
                              ),
                              child: Row(
                                children: [
                                  Expanded(
                                    flex: 3,
                                    child: Text(
                                      'Produk',
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w800,
                                        fontSize: 11,
                                        color: Colors.blueGrey,
                                      ),
                                    ),
                                  ),
                                  Expanded(
                                    flex: 1,
                                    child: Text(
                                      'Qty',
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w800,
                                        fontSize: 11,
                                        color: Colors.blueGrey,
                                      ),
                                      textAlign: TextAlign.center,
                                    ),
                                  ),
                                  Expanded(
                                    flex: 2,
                                    child: Text(
                                      'Harga',
                                      style: GoogleFonts.outfit(
                                        fontWeight: FontWeight.w800,
                                        fontSize: 11,
                                        color: Colors.blueGrey,
                                      ),
                                      textAlign: TextAlign.right,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const Divider(height: 1, thickness: 1),
                            // Table Items
                            ...items.map((it) {
                              final name = it['product']?['name'] ?? 'Barang';
                              final ordered = it['ordered_quantity'] ?? 0;
                              final actual = it['actual_quantity'] ?? ordered;
                              final price =
                                  (it['price_at_transaction'] ??
                                          it['price'] ??
                                          0.0)
                                      .toDouble();

                              return Padding(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 16,
                                  vertical: 10,
                                ),
                                child: Row(
                                  children: [
                                    Expanded(
                                      flex: 3,
                                      child: Text(
                                        name,
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: FontWeight.w700,
                                          color: const Color(0xFF1E293B),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 1,
                                      child: Text(
                                        '$actual',
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: FontWeight.w800,
                                          color: const Color(0xFF1E293B),
                                        ),
                                        textAlign: TextAlign.center,
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: Text(
                                        formatter.format(price),
                                        style: GoogleFonts.outfit(
                                          fontSize: 12,
                                          fontWeight: FontWeight.w700,
                                          color: const Color(0xFF1E293B),
                                        ),
                                        textAlign: TextAlign.right,
                                      ),
                                    ),
                                  ],
                                ),
                              );
                            }),
                          ],
                        ),
                      ),
                    const SizedBox(height: 20),

                    // Order Totals Summary
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.end,
                          children: [
                            Text(
                              'TOTAL FAKTUR',
                              style: GoogleFonts.outfit(
                                fontSize: 10,
                                fontWeight: FontWeight.w800,
                                color: Colors.blueGrey,
                              ),
                            ),
                            Text(
                              formatter.format(amount),
                              style: GoogleFonts.outfit(
                                fontSize: 20,
                                fontWeight: FontWeight.w900,
                                color: _getThemeColor(),
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                    const SizedBox(height: 32),

                    // Share Actions
                    if (orderData['payment_method'] == 'TEMPO' && !isPaid) ...[
                      SizedBox(
                        width: double.infinity,
                        height: 54,
                        child: ElevatedButton.icon(
                          icon: const Icon(
                            Icons.payments_rounded,
                            color: Colors.white,
                          ),
                          label: Text(
                            'TERIMA PEMBAYARAN',
                            style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w900,
                              color: Colors.white,
                            ),
                          ),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.blueAccent,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16),
                            ),
                          ),
                          onPressed: () =>
                              _showTempoPaymentDialog(context, inv),
                        ),
                      ),
                      const SizedBox(height: 12),
                    ],
                    Row(
                      children: [
                        Expanded(
                          child: ElevatedButton.icon(
                            icon: const Icon(
                              Icons.download_rounded,
                              color: Colors.white,
                            ),
                            label: Text(
                              'UNDUH PDF',
                              style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w900,
                                color: Colors.white,
                              ),
                            ),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: Colors.blueAccent,
                              minimumSize: const Size(double.infinity, 54),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(16),
                              ),
                            ),
                            onPressed: () => _downloadInvoicePdf(context, inv),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: ElevatedButton.icon(
                            icon: const Icon(
                              Icons.share_rounded,
                              color: Colors.white,
                            ),
                            label: Text(
                              'WHATSAPP',
                              style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w900,
                                color: Colors.white,
                              ),
                            ),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: const Color(0xFF25D366),
                              minimumSize: const Size(double.infinity, 54),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(16),
                              ),
                            ),
                            onPressed: () => _shareToWhatsApp(context, inv),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 24),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Future<void> _downloadInvoicePdf(BuildContext context, dynamic inv) async {
    setState(() => _isGeneratingPdf = true);
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Sedang membuat berkas PDF Faktur...'),
        duration: Duration(seconds: 1),
      ),
    );

    try {
      final orderData = inv['sales_transaction'] ?? inv['sales_order'] ?? {};
      final invoiceNo =
          orderData['invoice_no'] ?? orderData['so_number'] ?? 'INV-PENDING';

      final pdf = pw.Document();
      final currencyFormat = NumberFormat.currency(
        locale: 'id_ID',
        symbol: 'Rp ',
        decimalDigits: 0,
      );

      final storeName = orderData['store']?['name'] ?? 'Toko';
      final storeAddress = orderData['store']?['address'] ?? 'Alamat Toko';
      final dateStr = DateFormat('dd MMMM yyyy').format(DateTime.now());

      // Payment details
      String methodLabel = widget.paymentMethod;
      if (widget.paymentMethod == 'CASH') methodLabel = 'Tunai (Cash)';
      if (widget.paymentMethod == 'TRANSFER') methodLabel = 'Transfer Digital';
      if (widget.paymentMethod == 'TEMPO') methodLabel = 'Tempo / Kredit';

      final payAmount =
          (orderData['total_amount'] ?? inv['payment_amount'] ?? 0.0)
              .toDouble();

      final isPaid = _isInvoicePaid(inv);

      final companyName =
          orderData['employee']?['company']?['name'] ??
          orderData['company']?['name'] ??
          'Perusahaan Demo';
      final companyAddress =
          orderData['company']?['address'] ??
          'Kawasan Industri & Distribusi Sentra Prima';
      final companyTaxNumber = orderData['company']?['tax_number'] ?? '';

      final items = orderData['items'] as List? ?? [];
      final receivedBy = inv['received_by'] ?? '';

      // Extract signatures if available
      final String sigUrl =
          inv['signature_image_url'] ?? orderData['signature_image_url'] ?? '';
      List<String> sigUrls = sigUrl.split(',');
      String? courierSigUrl = sigUrls.isNotEmpty && sigUrls[0].isNotEmpty
          ? sigUrls[0]
          : null;
      String? customerSigUrl = sigUrls.length > 1 && sigUrls[1].isNotEmpty
          ? sigUrls[1]
          : null;

      pw.Widget? courierSigWidget;
      pw.Widget? customerSigWidget;

      // Draw network signatures in PDF if online
      if (courierSigUrl != null && courierSigUrl.startsWith('http')) {
        try {
          courierSigWidget = pw.Image(
            await networkImage(courierSigUrl),
            height: 35,
            width: 70,
          );
        } catch (e) {
          courierSigWidget = pw.Text(
            'Driver Delivery (Ttd digital)',
            style: pw.TextStyle(fontSize: 7, fontStyle: pw.FontStyle.italic),
          );
        }
      } else {
        courierSigWidget = pw.Text(
          'Driver Delivery',
          style: pw.TextStyle(fontSize: 7, fontStyle: pw.FontStyle.italic),
        );
      }

      if (customerSigUrl != null && customerSigUrl.startsWith('http')) {
        try {
          customerSigWidget = pw.Image(
            await networkImage(customerSigUrl),
            height: 35,
            width: 70,
          );
        } catch (e) {
          customerSigWidget = pw.Text(
            receivedBy.isNotEmpty ? receivedBy : 'Pelanggan',
            style: pw.TextStyle(fontSize: 7, fontStyle: pw.FontStyle.italic),
          );
        }
      } else {
        customerSigWidget = pw.Text(
          receivedBy.isNotEmpty ? receivedBy : 'Pelanggan',
          style: pw.TextStyle(fontSize: 7, fontStyle: pw.FontStyle.italic),
        );
      }

      pdf.addPage(
        pw.Page(
          pageFormat: PdfPageFormat.a4,
          margin: const pw.EdgeInsets.symmetric(vertical: 16, horizontal: 24),
          build: (pw.Context context) {
            return pw.Column(
              crossAxisAlignment: pw.CrossAxisAlignment.start,
              children: [
                // Header
                pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                  children: [
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.start,
                      children: [
                        pw.Text(
                          companyName.toUpperCase(),
                          style: pw.TextStyle(
                            fontSize: 9,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.Container(
                          width: 250,
                          child: pw.Text(
                            companyAddress,
                            style: const pw.TextStyle(fontSize: 7),
                          ),
                        ),
                        if (companyTaxNumber.isNotEmpty)
                          pw.Text(
                            'NPWP: $companyTaxNumber',
                            style: const pw.TextStyle(fontSize: 7),
                          ),
                      ],
                    ),
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.end,
                      children: [
                        pw.Text(
                          'FAKTUR PENJUALAN',
                          style: pw.TextStyle(
                            fontSize: 11,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.black,
                          ),
                        ),
                        pw.Text(
                          invoiceNo,
                          style: pw.TextStyle(
                            fontSize: 8,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.Text(
                          'Tanggal: $dateStr',
                          style: const pw.TextStyle(fontSize: 7),
                        ),
                      ],
                    ),
                  ],
                ),
                pw.Divider(
                  thickness: 1.0,
                  color: PdfColors.grey400,
                  height: 10,
                ),
                pw.SizedBox(height: 4),

                // Billing Details
                pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
                  crossAxisAlignment: pw.CrossAxisAlignment.start,
                  children: [
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.start,
                      children: [
                        pw.Text(
                          'DITAGIHKAN KEPADA:',
                          style: pw.TextStyle(
                            fontSize: 7,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.grey700,
                          ),
                        ),
                        pw.Text(
                          storeName,
                          style: pw.TextStyle(
                            fontSize: 8.5,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.Container(
                          width: 220,
                          child: pw.Text(
                            storeAddress,
                            style: const pw.TextStyle(fontSize: 7),
                            maxLines: 3,
                          ),
                        ),
                      ],
                    ),
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.end,
                      children: [
                        pw.Text(
                          'METODE PEMBAYARAN:',
                          style: pw.TextStyle(
                            fontSize: 7,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.grey700,
                          ),
                        ),
                        pw.Text(
                          methodLabel.toUpperCase(),
                          style: pw.TextStyle(
                            fontSize: 8.5,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        if (widget.paymentMethod == 'TEMPO' &&
                            orderData['payment_due_date'] != null)
                          pw.Text(
                            'Jatuh Tempo: ${DateFormat('dd MMM yyyy').format(DateTime.parse(orderData['payment_due_date']).toLocal())}',
                            style: pw.TextStyle(
                              fontSize: 7,
                              fontWeight: pw.FontWeight.bold,
                              color: PdfColors.black,
                            ),
                          ),
                        pw.Text(
                          'Status: ${widget.paymentMethod == 'TEMPO' ? 'BELUM LUNAS (TEMPO)' : (widget.paymentMethod == 'TRANSFER' && !isPaid ? 'PENDING' : 'LUNAS')}',
                          style: pw.TextStyle(
                            fontSize: 7.5,
                            fontWeight: pw.FontWeight.bold,
                            color: PdfColors.black,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                pw.SizedBox(height: 6),

                // Product Table Headers
                pw.Container(
                  color: PdfColors.grey200,
                  padding: const pw.EdgeInsets.symmetric(
                    vertical: 3,
                    horizontal: 6,
                  ),
                  child: pw.Row(
                    children: [
                      pw.Expanded(
                        flex: 3,
                        child: pw.Text(
                          'Nama Produk',
                          style: pw.TextStyle(
                            color: PdfColors.black,
                            fontWeight: pw.FontWeight.bold,
                            fontSize: 7,
                          ),
                        ),
                      ),
                      pw.Expanded(
                        flex: 1,
                        child: pw.Text(
                          'Qty',
                          style: pw.TextStyle(
                            color: PdfColors.black,
                            fontWeight: pw.FontWeight.bold,
                            fontSize: 7,
                          ),
                          textAlign: pw.TextAlign.center,
                        ),
                      ),
                      pw.Expanded(
                        flex: 2,
                        child: pw.Text(
                          'Harga',
                          style: pw.TextStyle(
                            color: PdfColors.black,
                            fontWeight: pw.FontWeight.bold,
                            fontSize: 7,
                          ),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                      pw.Expanded(
                        flex: 2,
                        child: pw.Text(
                          'Subtotal',
                          style: pw.TextStyle(
                            color: PdfColors.black,
                            fontWeight: pw.FontWeight.bold,
                            fontSize: 7,
                          ),
                          textAlign: pw.TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                ),

                // Product Table Items
                ...items.map((item) {
                  final prodName = item['product']?['name'] ?? 'Produk';
                  final orderedQty = item['ordered_quantity'] ?? 0;
                  final actualQty = item['actual_quantity'] ?? orderedQty;
                  final price =
                      (item['price_at_transaction'] ?? item['price'] ?? 0)
                          .toDouble();
                  final subtotal = actualQty * price;

                  return pw.Container(
                    padding: const pw.EdgeInsets.symmetric(
                      vertical: 2,
                      horizontal: 6,
                    ),
                    decoration: const pw.BoxDecoration(
                      border: pw.Border(
                        bottom: pw.BorderSide(color: PdfColors.grey300),
                      ),
                    ),
                    child: pw.Row(
                      children: [
                        pw.Expanded(
                          flex: 3,
                          child: pw.Text(
                            prodName,
                            style: const pw.TextStyle(fontSize: 7),
                          ),
                        ),
                        pw.Expanded(
                          flex: 1,
                          child: pw.Text(
                            '$actualQty',
                            style: const pw.TextStyle(fontSize: 7),
                            textAlign: pw.TextAlign.center,
                          ),
                        ),
                        pw.Expanded(
                          flex: 2,
                          child: pw.Text(
                            currencyFormat.format(price),
                            style: const pw.TextStyle(fontSize: 7),
                            textAlign: pw.TextAlign.right,
                          ),
                        ),
                        pw.Expanded(
                          flex: 2,
                          child: pw.Text(
                            currencyFormat.format(subtotal),
                            style: const pw.TextStyle(fontSize: 7),
                            textAlign: pw.TextAlign.right,
                          ),
                        ),
                      ],
                    ),
                  );
                }),

                pw.SizedBox(height: 6),

                // Totals
                pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.end,
                  children: [
                    pw.Column(
                      crossAxisAlignment: pw.CrossAxisAlignment.end,
                      children: [
                        pw.Row(
                          children: [
                            pw.Text(
                              'Total Bayar: ',
                              style: const pw.TextStyle(fontSize: 8),
                            ),
                            pw.Text(
                              currencyFormat.format(payAmount),
                              style: pw.TextStyle(
                                fontSize: 8,
                                fontWeight: pw.FontWeight.bold,
                                color: PdfColors.black,
                              ),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ],
                ),
                pw.SizedBox(height: 10),

                // Signatures
                pw.Row(
                  mainAxisAlignment: pw.MainAxisAlignment.spaceAround,
                  children: [
                    pw.Column(
                      children: [
                        pw.Text(
                          'Penerima / Toko',
                          style: pw.TextStyle(
                            fontSize: 8,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.SizedBox(height: 2),
                        (customerSigWidget ?? pw.SizedBox(height: 20)),
                        pw.SizedBox(height: 2),
                        pw.Text(
                          receivedBy.isNotEmpty ? receivedBy : 'Pelanggan',
                          style: const pw.TextStyle(fontSize: 7),
                        ),
                      ],
                    ),
                    pw.Column(
                      children: [
                        pw.Text(
                          'Kurir / Driver',
                          style: pw.TextStyle(
                            fontSize: 8,
                            fontWeight: pw.FontWeight.bold,
                          ),
                        ),
                        pw.SizedBox(height: 2),
                        (courierSigWidget ?? pw.SizedBox(height: 20)),
                        pw.SizedBox(height: 2),
                        pw.Text(
                          'Driver Delivery',
                          style: const pw.TextStyle(fontSize: 7),
                        ),
                      ],
                    ),
                  ],
                ),

                pw.Spacer(),
                pw.Center(
                  child: pw.Text(
                    'Terima kasih atas kerja sama dan kepercayaan Anda.',
                    style: pw.TextStyle(
                      fontSize: 7,
                      fontStyle: pw.FontStyle.italic,
                      color: PdfColors.grey600,
                    ),
                  ),
                ),
              ],
            );
          },
        ),
      );

      final pdfBytes = await pdf.save();
      if (kIsWeb) {
        final xFile = XFile.fromData(
          pdfBytes,
          mimeType: 'application/pdf',
          name: 'Faktur_$invoiceNo.pdf',
        );
        await Share.shareXFiles(
          [xFile],
          text:
              'Halo! Berikut terlampir salinan digital Faktur Penjualan resmi Anda ($invoiceNo). Terima kasih!',
        );
      } else {
        final tempDir = Directory.systemTemp;
        final file = File('${tempDir.path}/Faktur_$invoiceNo.pdf');
        await file.writeAsBytes(pdfBytes);

        final xFile = XFile(file.path, mimeType: 'application/pdf');
        await Share.shareXFiles(
          [xFile],
          text:
              'Halo! Berikut terlampir salinan digital Faktur Penjualan resmi Anda ($invoiceNo). Terima kasih!',
        );
      }
    } catch (e) {
      debugPrint('Gagal mengunduh PDF: $e');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Gagal membuat berkas PDF: $e'),
          backgroundColor: Colors.red,
        ),
      );
    } finally {
      setState(() => _isGeneratingPdf = false);
    }
  }

  Future<void> _shareToWhatsApp(BuildContext context, dynamic inv) async {
    final orderData = inv['sales_transaction'] ?? inv['sales_order'] ?? {};
    final storeName = orderData['store']?['name'] ?? 'Toko';
    final invoiceNo =
        orderData['invoice_no'] ?? orderData['so_number'] ?? 'INV-PENDING';
    final isPaid = _isInvoicePaid(inv);
    final amount = (orderData['total_amount'] ?? inv['payment_amount'] ?? 0.0)
        .toDouble();

    final currencyFormat = NumberFormat.currency(
      locale: 'id',
      symbol: 'Rp ',
      decimalDigits: 0,
    );
    String message =
        "Halo *$storeName*,\n\nBerikut rincian Faktur Penjualan resmi Anda:\n";
    message += "📄 *Nomor Faktur:* `$invoiceNo`\n";
    message += "💰 *Total Nominal:* ${currencyFormat.format(amount)}\n";
    message += "💳 *Metode Pembayaran:* ${widget.paymentMethod}\n";
    message +=
        "🟢 *Status Pembayaran:* ${isPaid ? 'LUNAS / BERHASIL' : 'BELUM LUNAS / PENDING'}\n";

    if (widget.paymentMethod == 'TRANSFER') {
      final bank = (orderData['midtrans_bank'] ?? 'Bank')
          .toString()
          .toUpperCase();
      final va =
          orderData['midtrans_va_number'] ??
          orderData['midtrans_bill_key'] ??
          '';
      if (va.isNotEmpty) {
        message += "🏦 *Virtual Account ($bank):* `$va`\n";
      }
    } else if (widget.paymentMethod == 'TEMPO' &&
        orderData['payment_due_date'] != null) {
      final dueDate = DateTime.tryParse(
        orderData['payment_due_date'].toString(),
      );
      if (dueDate != null) {
        message +=
            "📅 *Jatuh Tempo:* ${DateFormat('dd MMMM yyyy').format(dueDate)}\n";
      }
    }

    message += "\nTerima kasih atas kerja sama Anda!";

    final whatsappUrl = Uri.parse(
      "https://wa.me/?text=${Uri.encodeComponent(message)}",
    );
    if (await canLaunchUrl(whatsappUrl)) {
      await launchUrl(whatsappUrl, mode: LaunchMode.externalApplication);
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Aplikasi WhatsApp tidak dapat dibuka.'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }
}
