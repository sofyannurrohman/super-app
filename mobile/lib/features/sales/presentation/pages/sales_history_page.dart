import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/features/sales/data/services/banner_order_api_service.dart';
import 'package:hris_app/injection.dart' as di;
import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:intl/intl.dart';
import './digital_receipt_page.dart';
import './edit_sales_order_page.dart';
import './edit_offline_transaction_page.dart';
import './tempo_payment_page.dart';
import 'package:drift/drift.dart' show Value;
import 'package:hris_app/core/database/database.dart';
import 'package:hris_app/features/sync/data/repositories/sync_repository.dart';
import 'package:hris_app/core/services/data_integrity_service.dart';

class SalesHistoryPage extends StatefulWidget {
  const SalesHistoryPage({super.key});

  @override
  State<SalesHistoryPage> createState() => _SalesHistoryPageState();
}

class _SalesHistoryPageState extends State<SalesHistoryPage> {
  final _apiService = SalesApiService(apiClient: di.sl<ApiClient>());
  bool _isLoading = true;
  List<dynamic> _historyTransactions = [];
  List<dynamic> _salesOrders = [];
  bool _isLoadingSO = false;
  final Set<String> _retryingIds = {};

  @override
  void initState() {
    super.initState();
    _fetchHistory();
    _fetchSalesOrders();
  }

  Future<void> _fetchHistory() async {
    setState(() => _isLoading = true);
    try {
      // 1. Fetch from Local Database (Pending & Synced)
      final db = di.sl<AppDatabase>();
      final localTrxs = await db.select(db.localTransactions).get();
      final allLocalItems = await db.select(db.localTransactionItems).get();

      // Convert local items to Map structure matching API
      final mappedLocal = localTrxs.map((t) {
        final items = allLocalItems
            .where((i) => i.transactionLocalId == t.localId)
            .map((i) => {
                  'product_name': i.productName,
                  'quantity': i.quantity,
                  'unit': i.unit,
                  'pieces_per_unit': i.piecesPerUnit,
                  'price': i.price,
                })
            .toList();

        return {
          'id': t.localId,
          'receipt_no': t.receiptNo ??
              'PENDING-${t.localId.substring(0, 8).toUpperCase()}',
          'store': {'name': t.storeName},
          'total_amount': t.totalAmount,
          'created_at': t.createdAt.toIso8601String(),
          'transaction_date': t.createdAt.toIso8601String(),
          'sync_status': t.syncStatus,
          'sync_error_message': t.syncErrorMessage,
          'last_sync_attempt': t.lastSyncAttempt?.toIso8601String(),
          'status': t.syncStatus == 'synced'
              ? (t.paymentMethod == 'CASH' ? 'LUNAS' : 'PENDING')
              : (t.paymentMethod == 'CASH'
                  ? 'LUNAS (OFFLINE)'
                  : 'PENDING (OFFLINE)'),
          'payment_method': t.paymentMethod,
          'dp_payment_method': t.dpPaymentMethod,
          'dp_amount': t.paidAmount,
          'notes': t.notes,
          'is_local': true,
          'local_id': t.localId,
          'midtrans_qris_url': t.midtransQrisUrl,
          'midtrans_va_number': t.midtransVaNumber,
          'midtrans_bank': t.midtransBank,
          'midtrans_bill_key': t.midtransBillKey,
          'midtrans_biller_code': t.midtransBillerCode,
          'signature_image_url': t.signaturePath,
          'receipt_path': t.receiptPath,
          'receipt_image_url': t.receiptPath,
          'payment_status': t.paymentMethod == 'SALES_ORDER'
              ? 'UNPAID'
              : (t.paymentMethod == 'TEMPO' ? 'TEMPO' : 'PAID'),
          'items': items,
        };
      }).toList();

      // 2. Fetch from API
      List<dynamic> remoteHistory = [];
      String? myEmployeeId;
      String? myFirstName;
      try {
        // Fetch current profile to get exact employee ID
        final profileResponse =
            await _apiService.apiClient.client.get('employees/profile');
        myEmployeeId = profileResponse.data['data']['id']?.toString();
        myFirstName = profileResponse.data['data']['first_name']?.toString().toLowerCase().trim();

        final rawHistory = await _apiService.getHistoryTransactions();

        // Strictly filter remote history to only include transactions belonging to the current employee
        if (myEmployeeId != null) {
          remoteHistory = rawHistory
              .where((t) => t['employee_id']?.toString() == myEmployeeId)
              .toList();
        } else {
          remoteHistory = rawHistory;
        }
      } catch (e) {
        debugPrint('API History fetch failed, using local only: $e');
      }

      // 3. Merge (Avoid duplicates if possible, but localId is different from API ID)
      // We'll use receipt_no for deduplication if synced.
      // IMPORTANT: Payment logs have IsPaymentLog=true. If a payment log exists, it should not be overwritten by the original invoice.
      // Fetch Banners
      List<dynamic> remoteBanners = [];
      try {
        final bannerApi = BannerOrderApiService(apiClient: di.sl<ApiClient>());
        final rawBanners = await bannerApi.getBanners();

        if (myEmployeeId != null) {
          remoteBanners = rawBanners.where((b) {
            final matchId = b['employee_id']?.toString() == myEmployeeId || b['sales_id']?.toString() == myEmployeeId;
            if (matchId) return true;
            if (myFirstName != null && myFirstName.isNotEmpty) {
              final sName = b['sales_name']?.toString().toLowerCase().trim() ?? '';
              if (sName.contains(myFirstName)) return true;
            }
            return false;
          }).toList();
        } else {
          remoteBanners = rawBanners;
        }

        // Baca data spanduk offline jika ada
        try {
          final prefs = await SharedPreferences.getInstance();
          final String? existingJson = prefs.getString('offline_banner_orders');
          if (existingJson != null) {
            final List<dynamic> offlineBanners = jsonDecode(existingJson);
            for (var b in offlineBanners) {
              b['id'] = b['id'] ?? 'OFFLINE-${DateTime.now().millisecondsSinceEpoch}';
              b['status'] = 'PENDING (OFFLINE)';
              b['shop_name'] = b['shop_name'] ?? b['store_name'] ?? 'Toko Tidak Diketahui';
              remoteBanners.add(b);
            }
          }
        } catch (e) {
          debugPrint('Gagal memuat offline banners: $e');
        }
      } catch (e) {
        debugPrint('API Banners fetch failed: $e');
      }

      final Map<String, dynamic> merged = {};

      for (var t in mappedLocal) {
        merged[t['receipt_no'].toString()] = t;
      }
      for (var t in remoteHistory) {
        final key = t['receipt_no'].toString();
        final remoteLocalId = t['local_id']?.toString();
        final modifiableT = Map<String, dynamic>.from(t as Map);

        String? matchedKey;
        if (merged.containsKey(key)) {
          matchedKey = key;
        } else if (remoteLocalId != null) {
          final match = merged.entries
              .where((e) => e.value['local_id'] == remoteLocalId)
              .firstOrNull;
          if (match != null) {
            matchedKey = match.key;
          }
        }

        if (matchedKey != null) {
          final existingIsPayment =
              merged[matchedKey]!['is_payment_log'] == true;
          final newIsPayment = modifiableT['is_payment_log'] == true;

          if (!existingIsPayment || newIsPayment) {
            final isLocalPending =
                merged[matchedKey]!['sync_status'] == 'pending';
            final localId = merged[matchedKey]!['local_id'];

            if (isLocalPending) {
              // Preserve local edits for items and amounts, but keep remote structure/status
              final localTotal = merged[matchedKey]!['total_amount'];
              final localPaid = merged[matchedKey]!['dp_amount'] ??
                  merged[matchedKey]!['paid_amount'];
              final localItems = merged[matchedKey]!['items'];

              merged[matchedKey] = modifiableT;
              merged[matchedKey]!['total_amount'] = localTotal;
              merged[matchedKey]!['paid_amount'] = localPaid;
              merged[matchedKey]!['items'] = localItems;

              if (localId != null) {
                merged[matchedKey]!['local_id'] = localId;
              }
            } else {
              merged[matchedKey] = modifiableT;
              if (localId != null) {
                merged[matchedKey]!['local_id'] = localId;
              }
            }
          }
        } else {
          merged[key] = modifiableT;
        }
      }
      for (var b in remoteBanners) {
        final key = 'BANNER-${b['id']}';

        double w = (b['size_width'] ?? 0).toDouble();
        double h = (b['size_length'] ?? 0).toDouble();

        // Normalize to meters if seemingly entered in cm (e.g. > 20)
        if (w > 20) w = w / 100;
        if (h > 20) h = h / 100;

        merged[key] = {
          'id': b['id'],
          'is_banner': true,
          'created_at': b['created_at'] ?? DateTime.now().toIso8601String(),
          'store': {'name': b['shop_name'] ?? 'Toko Tidak Diketahui'},
          'status': b['status']?.toString().toUpperCase() ?? 'REQUEST',
          'type': b['type'] ?? 'BANNER',
          'size_width': w,
          'size_length': h,
          'notes': b['notes'],
        };
      }

      final combined = merged.values.toList();
      
      // Hapus transaksi discarded yang tidak ada di server
      combined.removeWhere((t) => 
          t['sync_status'] == 'discarded' && t['is_local'] == true);

      // Hapus transaksi dummy (kunjungan tanpa transaksi sesungguhnya, misal Input Transaksi Nanti)
      combined.removeWhere((t) {
        if (t['is_banner'] == true) return false;
        final total = (t['total_amount'] as num?)?.toDouble() ?? 0.0;
        final items = t['items'] as List?;
        final isDummy = total == 0 && (items == null || items.isEmpty);
        return isDummy;
      });

      // 4. Sort by date descending
      combined.sort((a, b) {
        final dateA = DateTime.parse(a['created_at']);
        final dateB = DateTime.parse(b['created_at']);
        return dateB.compareTo(dateA);
      });

      setState(() {
        _historyTransactions = combined;
      });
    } catch (e) {
      debugPrint('Error fetching history: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _retrySyncTransaction(String localId) async {
    if (_retryingIds.contains(localId)) return;

    setState(() {
      _retryingIds.add(localId);
    });

    try {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Menyinkronkan transaksi ke server...'),
            duration: Duration(seconds: 2),
            behavior: SnackBarBehavior.floating,
          ),
        );
      }

      final db = di.sl<AppDatabase>();
      await (db.update(db.localTransactions)
            ..where((t) => t.localId.equals(localId)))
          .write(const LocalTransactionsCompanion(
        syncStatus: Value('pending'),
        syncErrorMessage: Value(null),
      ));

      // Refresh list to show pending/syncing state
      await _fetchHistory();

      // Trigger sync with force: true
      await di.sl<SyncRepository>().syncTransactions(force: true);

      // Refresh to show final state
      await _fetchHistory();
    } catch (e) {
      debugPrint('Error retrying transaction sync: $e');
    } finally {
      if (mounted) {
        setState(() {
          _retryingIds.remove(localId);
        });
      }
    }
  }

  // Fetch SO langsung dari backend (Sales Order endpoint mobile)
  Future<void> _fetchSalesOrders() async {
    if (mounted) setState(() => _isLoadingSO = true);
    try {
      final orders = await _apiService.getSalesOrders();
      if (mounted) setState(() => _salesOrders = orders);
    } catch (e) {
      debugPrint('Error fetching sales orders: $e');
    } finally {
      if (mounted) setState(() => _isLoadingSO = false);
    }
  }

  Future<void> _confirmDeleteSO(String soId, String soNumber) async {
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text('Hapus Pesanan', style: GoogleFonts.outfit(fontWeight: FontWeight.bold)),
        content: Text('Apakah Anda yakin ingin menghapus pesanan $soNumber? Tindakan ini tidak dapat dibatalkan.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Batal'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Hapus', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );

    if (confirm == true) {
      if (mounted) setState(() => _isLoadingSO = true);
      try {
        await _apiService.deleteSalesOrder(soId);
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Pesanan berhasil dihapus')),
          );
        }
        _fetchSalesOrders();
      } catch (e) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Gagal menghapus pesanan: $e')),
          );
          setState(() => _isLoadingSO = false);
        }
      }
    }
  }

  Future<void> _deletePendingServerTransaction(dynamic txn) async {
    final id = txn['id']?.toString();
    if (id == null || id.isEmpty) return;
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Hapus Transaksi Menunggu Validasi'),
        content: Text('Hapus transaksi ${txn['receipt_no'] ?? ''} dari server? Stok sales akan dikembalikan. Transaksi yang sudah dibayar atau divalidasi Admin Nota tidak dapat dihapus.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Kembali')),
          TextButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Hapus', style: TextStyle(color: Colors.red))),
        ],
      ),
    );
    if (confirmed != true) return;
    if (mounted) setState(() => _isLoading = true);
    try {
      await _apiService.deletePendingTransaction(id);
      final localId = txn['local_id']?.toString();
      if (localId != null && localId.isNotEmpty) {
        final db = di.sl<AppDatabase>();
        await db.transaction(() async {
          final local = await (db.select(db.localTransactions)
                ..where((t) => t.localId.equals(localId))).getSingleOrNull();
          if (local == null) return;
          final items = await (db.select(db.localTransactionItems)
                ..where((t) => t.transactionLocalId.equals(localId))).get();
          for (final item in items) {
            final stock = await (db.select(db.salesStock)
                  ..where((t) => t.productId.equals(item.productId))).getSingleOrNull();
            if (stock != null) {
              await (db.update(db.salesStock)
                    ..where((t) => t.productId.equals(item.productId)))
                  .write(SalesStockCompanion(
                    quantity: Value(stock.quantity + item.quantity),
                    updatedAt: Value(DateTime.now()),
                  ));
            }
          }
          if (local.notes != null && local.notes!.isNotEmpty) {
            try {
              final parsed = jsonDecode(local.notes!);
              if (parsed is Map) {
                for (final entry in [
                  (parsed['bonus_items'] as List?) ?? [],
                  (parsed['return_items'] as List?) ?? [],
                ].indexed) {
                  for (final row in entry.$2) {
                    final productId = row['product_id']?.toString();
                    final quantity = (row['quantity'] as num?)?.toInt() ?? 0;
                    if (productId == null || quantity <= 0) continue;
                    final stock = await (db.select(db.salesStock)
                          ..where((t) => t.productId.equals(productId))).getSingleOrNull();
                    if (stock == null) continue;
                    final updated = entry.$1 == 0
                        ? stock.quantity + quantity
                        : (stock.quantity - quantity).clamp(0, 999999);
                    await (db.update(db.salesStock)
                          ..where((t) => t.productId.equals(productId)))
                        .write(SalesStockCompanion(
                          quantity: Value(updated),
                          updatedAt: Value(DateTime.now()),
                        ));
                  }
                }
              }
            } catch (_) {}
          }
          await (db.delete(db.localTransactionItems)
                ..where((t) => t.transactionLocalId.equals(localId))).go();
          await (db.delete(db.localTransactions)
                ..where((t) => t.localId.equals(localId))).go();
        });
      }
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Transaksi berhasil dihapus dari server')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Gagal menghapus transaksi: $e')),
        );
      }
    } finally {
      if (mounted) await _fetchHistory();
    }
  }

  Future<void> _deleteOfflineTransaction(
      String localId, String storeName, double totalAmount) async {
    final currencyFormat =
        NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);
    final confirm = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.all(8),
              decoration: BoxDecoration(
                color: Colors.red.shade50,
                shape: BoxShape.circle,
              ),
              child: Icon(Icons.delete_forever_rounded,
                  color: Colors.red.shade700, size: 24),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Text('Hapus Transaksi Offline?',
                  style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w800, fontSize: 17)),
            ),
          ],
        ),
        content: Text(
          'Transaksi untuk $storeName sebesar ${currencyFormat.format(totalAmount)} belum tersinkron ke server.\n\n'
          'Jika dihapus, transaksi ini akan dibatalkan, stok barang akan dikembalikan ke stok fisik lokal Anda, dan total omzet hari ini akan diperbarui.',
          style:
              GoogleFonts.outfit(fontSize: 13, color: Colors.blueGrey.shade800),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: Text('Batal',
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w600, color: Colors.blueGrey)),
          ),
          FilledButton(
            style: FilledButton.styleFrom(backgroundColor: Colors.red.shade600),
            onPressed: () => Navigator.pop(ctx, true),
            child: Text('Hapus Transaksi',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    setState(() => _isLoading = true);
    try {
      final db = di.sl<AppDatabase>();
      await db.transaction(() async {
        // 1. Ambil transaksi & items
        final tx = await (db.select(db.localTransactions)
              ..where((t) => t.localId.equals(localId)))
            .getSingleOrNull();
        if (tx == null ||
            tx.syncStatus != 'pending' ||
            tx.lastSyncAttempt != null) {
          throw StateError('Transaksi sudah diproses sinkronisasi. Muat ulang riwayat sebelum menghapus.');
        }
        final items = await (db.select(db.localTransactionItems)
              ..where((t) => t.transactionLocalId.equals(localId)))
            .get();

        // 2. Kembalikan stok item reguler
        for (final item in items) {
          final stock = await (db.select(db.salesStock)
                ..where((t) => t.productId.equals(item.productId)))
              .getSingleOrNull();
          if (stock != null) {
            await (db.update(db.salesStock)
                  ..where((t) => t.productId.equals(item.productId)))
                .write(SalesStockCompanion(
              quantity: Value(stock.quantity + item.quantity),
              updatedAt: Value(DateTime.now()),
            ));
          }
        }

        // 3. Kembalikan stok bonus & rollback retur jika tersimpan di notes
        if (tx.notes != null && tx.notes!.isNotEmpty) {
          try {
            final parsed = jsonDecode(tx.notes!);
            if (parsed is Map) {
              if (parsed.containsKey('bonus_items')) {
                final bonuses = parsed['bonus_items'] as List<dynamic>?;
                if (bonuses != null) {
                  for (var b in bonuses) {
                    final pid = b['product_id'] as String?;
                    final qty = (b['quantity'] as num?)?.toInt() ?? 0;
                    if (pid != null && qty > 0) {
                      final stock = await (db.select(db.salesStock)
                            ..where((t) => t.productId.equals(pid)))
                          .getSingleOrNull();
                      if (stock != null) {
                        await (db.update(db.salesStock)
                              ..where((t) => t.productId.equals(pid)))
                            .write(SalesStockCompanion(
                          quantity: Value(stock.quantity + qty),
                          updatedAt: Value(DateTime.now()),
                        ));
                      }
                    }
                  }
                }
              }
              if (parsed.containsKey('return_items')) {
                final returns = parsed['return_items'] as List<dynamic>?;
                if (returns != null) {
                  for (var r in returns) {
                    final pid = r['product_id'] as String?;
                    final qty = (r['quantity'] as num?)?.toInt() ?? 0;
                    if (pid != null && qty > 0) {
                      final stock = await (db.select(db.salesStock)
                            ..where((t) => t.productId.equals(pid)))
                          .getSingleOrNull();
                      if (stock != null) {
                        final newQty = (stock.quantity - qty).clamp(0, 999999);
                        await (db.update(db.salesStock)
                              ..where((t) => t.productId.equals(pid)))
                            .write(SalesStockCompanion(
                          quantity: Value(newQty),
                          updatedAt: Value(DateTime.now()),
                        ));
                      }
                    }
                  }
                }
              }
            }
          } catch (_) {}
        }

        // 4. Hapus data dari DB lokal
        await (db.delete(db.localTransactionItems)
              ..where((t) => t.transactionLocalId.equals(localId)))
            .go();
        await (db.delete(db.localTransactions)
              ..where((t) => t.localId.equals(localId)))
            .go();
      });

      // 5. Bersihkan retry metadata
      await di.sl<DataIntegrityService>().resetRetry(localId);
      await di.sl<DataIntegrityService>().log(
            category: 'SALES',
            action: 'DELETE_OFFLINE_TRANSACTION',
            detail: 'localId=$localId dihapus oleh sales. Stok dikembalikan.',
          );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
                'Transaksi offline berhasil dihapus dan stok dikembalikan.',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Gagal menghapus transaksi: $e',
                style: GoogleFonts.outfit()),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        _fetchHistory();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final banners =
        _historyTransactions.where((t) => t['is_banner'] == true).toList();
    final belumLunas = _historyTransactions.where((t) {
      if (t['is_banner'] == true) return false;
      final paymentMethod = t['payment_method']?.toString().toUpperCase() ?? '';
      final paymentStatus = t['payment_status']?.toString().toUpperCase() ?? '';
      final status = t['status']?.toString().toUpperCase() ?? '';
      bool isActuallyPaid = paymentStatus == 'PAID' ||
          paymentStatus == 'LUNAS' ||
          status == 'LUNAS';
      final total = (t['total_amount'] as num?)?.toDouble() ?? 0.0;
      final paid = (t['paid_amount'] as num?)?.toDouble() ??
          (t['dp_amount'] as num?)?.toDouble() ??
          0.0;
      if (paymentMethod == 'TEMPO' && paid < total) isActuallyPaid = false;
      if (paymentMethod == 'SALES_ORDER' &&
          paymentStatus != 'PAID' &&
          paymentStatus != 'LUNAS') isActuallyPaid = false;
      return (paymentMethod == 'TEMPO' || paymentMethod == 'SALES_ORDER') &&
          !isActuallyPaid;
    }).toList();
    final lunas = _historyTransactions.where((t) {
      if (t['is_banner'] == true) return false;
      final paymentMethod = t['payment_method']?.toString().toUpperCase() ?? '';
      final paymentStatus = t['payment_status']?.toString().toUpperCase() ?? '';
      final status = t['status']?.toString().toUpperCase() ?? '';
      bool isActuallyPaid = paymentStatus == 'PAID' ||
          paymentStatus == 'LUNAS' ||
          status == 'LUNAS';
      final total = (t['total_amount'] as num?)?.toDouble() ?? 0.0;
      final paid = (t['paid_amount'] as num?)?.toDouble() ??
          (t['dp_amount'] as num?)?.toDouble() ??
          0.0;
      if (paymentMethod == 'TEMPO' && paid < total) isActuallyPaid = false;
      if (paymentMethod == 'SALES_ORDER' &&
          paymentStatus != 'PAID' &&
          paymentStatus != 'LUNAS') isActuallyPaid = false;
      return !((paymentMethod == 'TEMPO' || paymentMethod == 'SALES_ORDER') &&
          !isActuallyPaid);
    }).toList();

    return DefaultTabController(
      length: 4,
      child: Scaffold(
        backgroundColor: const Color(0xFFF1F5F9),
        appBar: AppBar(
          backgroundColor: Colors.white,
          elevation: 0,
          leading: IconButton(
            icon: const Icon(Icons.arrow_back_ios_rounded,
                color: Color(0xFF334155)),
            onPressed: () => Navigator.pop(context),
          ),
          title: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('LOG OPERASIONAL',
                  style: GoogleFonts.outfit(
                      fontSize: 10,
                      color: const Color(0xFF10B981),
                      fontWeight: FontWeight.w800,
                      letterSpacing: 1)),
              Text('Riwayat',
                  style: GoogleFonts.outfit(
                      fontSize: 17,
                      fontWeight: FontWeight.w800,
                      color: const Color(0xFF1E293B))),
            ],
          ),
          bottom: TabBar(
            isScrollable: true,
            labelColor: const Color(0xFF10B981),
            unselectedLabelColor: Colors.blueGrey,
            indicatorColor: const Color(0xFF10B981),
            labelStyle:
                GoogleFonts.outfit(fontWeight: FontWeight.w800, fontSize: 13),
            unselectedLabelStyle:
                GoogleFonts.outfit(fontWeight: FontWeight.w600, fontSize: 13),
            tabs: const [
              Tab(text: 'LUNAS'),
              Tab(text: 'BELUM LUNAS'),
              Tab(text: 'SPANDUK'),
              Tab(text: 'PESANAN SO'),
            ],
          ),
        ),
        body: _isLoading
            ? const Center(child: CircularProgressIndicator())
            : TabBarView(
                children: [
                  RefreshIndicator(
                    onRefresh: _fetchHistory,
                    child: lunas.isEmpty
                        ? _buildEmptyState('Belum ada riwayat transaksi lunas.')
                        : ListView.builder(
                            physics: const AlwaysScrollableScrollPhysics(),
                            padding: const EdgeInsets.all(20),
                            itemCount: lunas.length,
                            itemBuilder: (ctx, i) =>
                                _buildHistoryCard(lunas[i]),
                          ),
                  ),
                  RefreshIndicator(
                    onRefresh: _fetchHistory,
                    child: belumLunas.isEmpty
                        ? _buildEmptyState(
                            'Tidak ada tagihan yang belum lunas.')
                        : ListView.builder(
                            physics: const AlwaysScrollableScrollPhysics(),
                            padding: const EdgeInsets.all(20),
                            itemCount: belumLunas.length,
                            itemBuilder: (ctx, i) =>
                                _buildHistoryCard(belumLunas[i]),
                          ),
                  ),
                  RefreshIndicator(
                    onRefresh: _fetchHistory,
                    child: banners.isEmpty
                        ? _buildEmptyState(
                            'Belum ada riwayat pemesanan branding.')
                        : ListView.builder(
                            physics: const AlwaysScrollableScrollPhysics(),
                            padding: const EdgeInsets.all(20),
                            itemCount: banners.length,
                            itemBuilder: (ctx, i) =>
                                _buildHistoryCard(banners[i]),
                          ),
                  ),
                  // ── Tab SO SAYA ──
                  RefreshIndicator(
                    onRefresh: _fetchSalesOrders,
                    child: _isLoadingSO
                        ? const Center(child: CircularProgressIndicator())
                        : _salesOrders.isEmpty
                            ? _buildEmptyState(
                                'Belum ada Sales Order yang dibuat.')
                            : ListView.builder(
                                physics: const AlwaysScrollableScrollPhysics(),
                                padding: const EdgeInsets.all(20),
                                itemCount: _salesOrders.length,
                                itemBuilder: (ctx, i) =>
                                    _buildSOCard(_salesOrders[i]),
                              ),
                  ),
                ],
              ),
      ),
    );
  }

  Widget _buildHistoryCard(dynamic txn) {
    if (txn['is_banner'] == true) {
      return _buildBannerCard(txn);
    }

    final storeName = txn['store']?['name'] ?? 'Toko Tidak Diketahui';
    final amount =
        NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0)
            .format(txn['total_amount'] ?? 0);
    final dateObj = DateTime.parse(txn['created_at']).toLocal();
    final dateStr = DateFormat('dd MMM yyyy, HH:mm').format(dateObj);
    final status = (txn['status']?.toString() ?? '').toUpperCase();
    final paymentMethod = txn['payment_method']?.toString().toUpperCase() ?? '';
    final dpPaymentMethod =
        txn['dp_payment_method']?.toString().toUpperCase() ?? '';
    final isLocal = txn['is_local'] == true;

    final bool isVerified = status == 'VERIFIED' ||
        status == 'LUNAS' ||
        status == 'CONVERTED' ||
        status == 'DELIVERED';
        
    final bool isApprovedByAdmin = status == 'VERIFIED' ||
        status == 'CONVERTED' ||
        status == 'DELIVERED' ||
        status == 'REJECTED' ||
        status == 'CANCELLED';
    final bool isSO = paymentMethod == 'SALES_ORDER';
    final bool isPaymentLog = txn['is_payment_log'] == true;
    final bool isRejected = status == 'REJECTED' || status == 'CANCELLED';
    final bool isPending = status == 'PENDING' ||
        status == 'WAITING_WAREHOUSE' ||
        status == 'WAITING_STOCK' ||
        status == 'DRAFT';
    final paymentStatus =
        (txn['payment_status'] ?? '').toString().toUpperCase();

    final totalAmount = (txn['total_amount'] as num?)?.toDouble() ?? 0.0;
    final paidAmount = (txn['paid_amount'] as num?)?.toDouble() ??
        (txn['dp_amount'] as num?)?.toDouble() ??
        0.0;
    bool isActuallyPaid = paymentStatus == 'PAID' ||
        paymentStatus == 'LUNAS' ||
        status == 'LUNAS';
    if (paymentMethod == 'TEMPO' && paidAmount < totalAmount) {
      isActuallyPaid = false;
    }
    if (paymentMethod == 'SALES_ORDER' &&
        paymentStatus != 'PAID' &&
        paymentStatus != 'LUNAS') {
      isActuallyPaid = false;
    }
    final bool showAsUnpaid =
        (paymentMethod == 'TEMPO' || paymentMethod == 'SALES_ORDER') &&
            !isActuallyPaid;

    final displayMethod = dpPaymentMethod.isNotEmpty
        ? dpPaymentMethod
        : (paymentMethod.isNotEmpty && paymentMethod != 'TEMPO'
            ? paymentMethod
            : 'TEMPO');

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withValues(alpha: 0.04),
              blurRadius: 20,
              offset: const Offset(0, 8))
        ],
      ),
      clipBehavior: Clip.antiAlias,
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: () {
            Navigator.push(
              context,
              MaterialPageRoute(
                builder: (_) => DigitalReceiptPage(
                  transaction: txn,
                  localId: txn['local_id'],
                ),
              ),
            );
          },
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  width: 56,
                  height: 56,
                  decoration: BoxDecoration(
                    color: isPaymentLog
                        ? const Color(0xFF10B981).withValues(alpha: 0.1)
                        : (showAsUnpaid
                            ? Colors.orange.withValues(alpha: 0.1)
                            : (isVerified
                                ? Colors.green.withValues(alpha: 0.1)
                                : (isRejected
                                    ? Colors.red.withValues(alpha: 0.1)
                                    : Colors.orange.withValues(alpha: 0.1)))),
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    isPaymentLog
                        ? Icons.payments_rounded
                        : (showAsUnpaid
                            ? Icons.assignment_late_rounded
                            : (isVerified
                                ? Icons.check_circle_rounded
                                : (isRejected
                                    ? Icons.cancel_rounded
                                    : Icons.pending_actions_rounded))),
                    color: isPaymentLog
                        ? const Color(0xFF10B981)
                        : (showAsUnpaid
                            ? Colors.orange.shade700
                            : (isVerified
                                ? Colors.green
                                : (isRejected ? Colors.red : Colors.orange))),
                    size: 28,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(storeName,
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800,
                              fontSize: 16,
                              color: const Color(0xFF1E293B))),
                      const SizedBox(height: 4),
                      Text(amount,
                          style: GoogleFonts.outfit(
                              fontSize: 14,
                              color: isPaymentLog
                                  ? const Color(0xFF10B981)
                                  : (showAsUnpaid
                                      ? Colors.orange.shade700
                                      : (isVerified
                                          ? Colors.green
                                          : Colors.orange)),
                              fontWeight: FontWeight.w800)),
                      const SizedBox(height: 4),
                      Text(dateStr,
                          style: GoogleFonts.outfit(
                              fontSize: 12,
                              color: Colors.blueGrey.withValues(alpha: 0.8))),
                      if (txn['items'] != null &&
                          (txn['items'] as List).isNotEmpty) ...[
                        const SizedBox(height: 6),
                        Text(
                          (txn['items'] as List)
                              .map((i) =>
                                  '${i['quantity']}x ${i['product_name'] ?? (i['product'] as Map?)?['name'] ?? i['name']}')
                              .join(', '),
                          style: GoogleFonts.outfit(
                              fontSize: 12,
                              color: Colors.blueGrey,
                              fontStyle: FontStyle.italic),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                      if (isLocal &&
                          (txn['sync_status'] == 'failed' ||
                              (txn['sync_error_message'] != null &&
                                  txn['sync_error_message'].toString().isNotEmpty))) ...[
                        const SizedBox(height: 6),
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 8, vertical: 6),
                          decoration: BoxDecoration(
                            color: Colors.red.shade50,
                            borderRadius: BorderRadius.circular(6),
                            border: Border.all(color: Colors.red.shade200),
                          ),
                          child: Row(
                            children: [
                              const Icon(Icons.error_outline_rounded,
                                  size: 14, color: Colors.red),
                              const SizedBox(width: 6),
                              Expanded(
                                child: Text(
                                  txn['sync_error_message'] != null &&
                                          txn['sync_error_message'].toString().isNotEmpty
                                      ? 'Gagal: ${txn['sync_error_message']}'
                                      : 'Gagal sinkronisasi ke server',
                                  style: GoogleFonts.outfit(
                                    fontSize: 11,
                                    color: Colors.red.shade800,
                                    fontWeight: FontWeight.w600,
                                  ),
                                  maxLines: 2,
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                              if (txn['local_id'] != null) ...[
                                const SizedBox(width: 8),
                                InkWell(
                                  onTap: _retryingIds.contains(txn['local_id'])
                                      ? null
                                      : () => _retrySyncTransaction(txn['local_id'].toString()),
                                  borderRadius: BorderRadius.circular(4),
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(
                                        horizontal: 8, vertical: 4),
                                    decoration: BoxDecoration(
                                      color: Colors.red.shade600,
                                      borderRadius: BorderRadius.circular(4),
                                    ),
                                    child: _retryingIds.contains(txn['local_id'])
                                        ? const SizedBox(
                                            width: 12,
                                            height: 12,
                                            child: CircularProgressIndicator(
                                              strokeWidth: 2,
                                              color: Colors.white,
                                            ),
                                          )
                                        : Row(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              const Icon(Icons.refresh_rounded,
                                                  size: 12, color: Colors.white),
                                              const SizedBox(width: 3),
                                              Text(
                                                'Coba Lagi',
                                                style: GoogleFonts.outfit(
                                                  fontSize: 10,
                                                  fontWeight: FontWeight.bold,
                                                  color: Colors.white,
                                                ),
                                              ),
                                            ],
                                          ),
                                  ),
                                ),
                              ],
                            ],
                          ),
                        ),
                      ],
                      const SizedBox(height: 6),
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 10, vertical: 3),
                        decoration: BoxDecoration(
                          color: isPaymentLog
                              ? const Color(0xFF10B981).withValues(alpha: 0.1)
                              : (showAsUnpaid
                                  ? Colors.orange.withValues(alpha: 0.1)
                                  : (isVerified
                                      ? Colors.green.withValues(alpha: 0.1)
                                      : (isLocal
                                          ? (txn['sync_status'] == 'failed'
                                              ? Colors.red.withValues(alpha: 0.1)
                                              : (txn['sync_status'] == 'syncing'
                                                  ? Colors.blue.withValues(alpha: 0.1)
                                                  : (txn['sync_status'] == 'synced'
                                                      ? Colors.green.withValues(alpha: 0.1)
                                                      : Colors.orange.withValues(alpha: 0.1))))
                                          : (isRejected
                                              ? Colors.red.withValues(alpha: 0.1)
                                              : Colors.orange.withValues(alpha: 0.1))))),
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Text(
                          isLocal
                              ? (txn['sync_status'] == 'failed'
                                  ? 'GAGAL SYNC'
                                  : (txn['sync_status'] == 'syncing'
                                      ? 'MENYINKRONKAN...'
                                      : (txn['sync_status'] == 'synced'
                                          ? 'TERSINKRON'
                                          : 'MENUNGGU SINKRONISASI')))
                              : (isSO
                                  ? 'SO: $status (${isActuallyPaid ? 'LUNAS' : 'BELUM LUNAS'})'
                                  : (isPaymentLog
                                      ? 'PEMBAYARAN $displayMethod'
                                      : ((paymentMethod == 'TEMPO' &&
                                              !isActuallyPaid)
                                          ? 'BELUM LUNAS (TEMPO)'
                                          : (isVerified
                                              ? 'LUNAS (VERIFIED)'
                                              : 'NEED REVIEW')))),
                          style: GoogleFonts.outfit(
                            fontSize: 10,
                            fontWeight: FontWeight.w900,
                            color: isPaymentLog
                                ? const Color(0xFF10B981)
                                : (showAsUnpaid
                                    ? Colors.orange.shade800
                                    : (isVerified
                                        ? Colors.green
                                        : (isLocal
                                            ? (txn['sync_status'] == 'failed'
                                                ? Colors.red
                                                : (txn['sync_status'] == 'syncing'
                                                    ? Colors.blue
                                                    : (txn['sync_status'] == 'synced'
                                                        ? Colors.green
                                                        : Colors.orange)))
                                            : (isRejected
                                                ? Colors.red
                                                : Colors.orange)))),
                          ),
                        ),
                      ),
                      if (!isActuallyPaid && !isRejected && status != 'DELETED' && !isPaymentLog) ...[
                        const SizedBox(height: 8),
                        SizedBox(
                          width: double.infinity,
                          child: OutlinedButton.icon(
                            onPressed: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (_) => TempoPaymentPage(
                                    transactionData: txn,
                                    onSuccess: (method, amount, signatureUrl) {
                                      Navigator.pop(context);
                                      ScaffoldMessenger.of(context).showSnackBar(
                                        const SnackBar(
                                          content: Text('Pembayaran berhasil dicatat!'),
                                          backgroundColor: Colors.green,
                                        ),
                                      );
                                      _fetchHistory();
                                    },
                                  ),
                                ),
                              );
                            },
                            icon: Icon(
                              (paymentMethod == 'TRANSFER' || paymentMethod == 'VA')
                                  ? Icons.upload_file_rounded
                                  : Icons.payments_rounded,
                              size: 14,
                              color: Colors.indigo.shade700,
                            ),
                            label: Text(
                              (paymentMethod == 'TRANSFER' || paymentMethod == 'VA')
                                  ? 'Upload Bukti Transfer'
                                  : 'Terima Pembayaran',
                              style: GoogleFonts.outfit(
                                fontSize: 11,
                                fontWeight: FontWeight.w800,
                                color: Colors.indigo.shade700,
                              ),
                            ),
                            style: OutlinedButton.styleFrom(
                              side: BorderSide(color: Colors.indigo.shade200),
                              backgroundColor: Colors.indigo.shade50.withValues(alpha: 0.5),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(10),
                              ),
                              padding: const EdgeInsets.symmetric(vertical: 6, horizontal: 10),
                            ),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                if (!isApprovedByAdmin)
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      IconButton(
                        icon:
                            const Icon(Icons.edit_note_rounded, color: Colors.blue),
                        tooltip: 'Edit Transaksi',
                        onPressed: () async {
                          final result = await Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (_) => EditOfflineTransactionPage(
                                localId: txn['local_id'],
                                transaction: txn,
                              ),
                            ),
                          );
                          if (result == true) {
                            _fetchHistory();
                          }
                        },
                      ),
                      if (isLocal &&
                          txn['sync_status'] == 'pending' &&
                          txn['last_sync_attempt'] == null &&
                          txn['local_id'] != null)
                        IconButton(
                          icon: Icon(Icons.delete_outline_rounded,
                              color: Colors.red.shade600),
                          tooltip: 'Hapus Transaksi Offline',
                          onPressed: () => _deleteOfflineTransaction(
                            txn['local_id'].toString(),
                            storeName,
                            (txn['total_amount'] as num?)?.toDouble() ?? 0.0,
                          ),
                        ),
                      if (!isLocal &&
                          !isSO &&
                          !isPaymentLog &&
                          status == 'PENDING' &&
                          paymentStatus == 'UNPAID' &&
                          paidAmount == 0 &&
                          txn['verified_at'] == null)
                        IconButton(
                          icon: Icon(Icons.delete_outline_rounded,
                              color: Colors.red.shade600),
                          tooltip: 'Hapus Transaksi Menunggu Validasi',
                          onPressed: () => _deletePendingServerTransaction(txn),
                        ),
                    ],
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildEmptyState(String message) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.history_edu_rounded,
              size: 80, color: Colors.grey.shade300),
          const SizedBox(height: 24),
          Text('Belum Ada Riwayat',
              style: GoogleFonts.outfit(
                  fontSize: 20,
                  fontWeight: FontWeight.w800,
                  color: const Color(0xFF334155))),
          const SizedBox(height: 8),
          Text(message,
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(fontSize: 14, color: Colors.blueGrey)),
        ],
      ),
    );
  }

  /// Kartu untuk tab "PESANAN SO" — menampilkan SO dari backend dengan tombol EDIT jika DRAFT.
  Widget _buildSOCard(dynamic so) {
    final currency =
        NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);
    final storeName = so['store']?['name'] ?? 'Toko Tidak Diketahui';
    final soNumber = so['so_number'] ?? so['SONumber'] ?? 'SO';
    final status = (so['status'] ?? 'DRAFT').toString().toUpperCase();
    final totalAmount = (so['total_amount'] as num?)?.toDouble() ?? 0.0;
    final createdAt = so['created_at'] ?? so['order_date'];
    String dateStr = '-';
    if (createdAt != null) {
      try {
        dateStr = DateFormat('dd MMM yyyy, HH:mm')
            .format(DateTime.parse(createdAt.toString()).toLocal());
      } catch (_) {}
    }
    final itemCount = (so['items'] as List?)?.length ?? 0;

    final bool isDraft = status == 'DRAFT';
    final bool isConfirmed = status == 'CONFIRMED';
    final bool isDelivered = status == 'DELIVERED' || status == 'IN_DELIVERY';
    final bool isPaid = status == 'PAID';
    final bool isCancelled = status == 'CANCELLED' || status == 'REJECTED';

    Color statusColor;
    Color statusBg;
    IconData statusIcon;
    String statusLabel;

    if (isDraft) {
      statusColor = const Color(0xFFF59E0B);
      statusBg = const Color(0xFFFFFBEB);
      statusIcon = Icons.edit_note_rounded;
      statusLabel = 'DRAFT — Menunggu Verifikasi';
    } else if (isConfirmed) {
      statusColor = const Color(0xFF3B82F6);
      statusBg = const Color(0xFFEFF6FF);
      statusIcon = Icons.check_circle_outline_rounded;
      statusLabel = 'DIKONFIRMASI';
    } else if (isDelivered) {
      statusColor = const Color(0xFF8B5CF6);
      statusBg = const Color(0xFFF5F3FF);
      statusIcon = Icons.local_shipping_rounded;
      statusLabel = status;
    } else if (isPaid) {
      statusColor = const Color(0xFF10B981);
      statusBg = const Color(0xFFF0FDF4);
      statusIcon = Icons.payments_rounded;
      statusLabel = 'LUNAS';
    } else if (isCancelled) {
      statusColor = Colors.red;
      statusBg = const Color(0xFFFFF1F2);
      statusIcon = Icons.cancel_rounded;
      statusLabel = status;
    } else {
      statusColor = Colors.blueGrey;
      statusBg = const Color(0xFFF1F5F9);
      statusIcon = Icons.pending_actions_rounded;
      statusLabel = status;
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withValues(alpha: 0.04),
              blurRadius: 20,
              offset: const Offset(0, 8)),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 56,
              height: 56,
              decoration: BoxDecoration(
                  color: statusBg, shape: BoxShape.circle),
              child: Icon(statusIcon, color: statusColor, size: 28),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(storeName,
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.w800,
                          fontSize: 16,
                          color: const Color(0xFF1E293B))),
                  const SizedBox(height: 4),
                  Text(currency.format(totalAmount),
                      style: GoogleFonts.outfit(
                          fontSize: 14,
                          color: const Color(0xFF10B981),
                          fontWeight: FontWeight.w800)),
                  const SizedBox(height: 4),
                  Text(soNumber,
                      style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: Colors.blueGrey,
                          fontWeight: FontWeight.w600)),
                  const SizedBox(height: 4),
                  Text(dateStr,
                      style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: Colors.blueGrey.withValues(alpha: 0.8))),
                  const SizedBox(height: 6),
                  Text('$itemCount item',
                      style: GoogleFonts.outfit(
                          fontSize: 12,
                          color: Colors.blueGrey,
                          fontStyle: FontStyle.italic)),
                  const SizedBox(height: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 10, vertical: 3),
                    decoration: BoxDecoration(
                        color: statusBg,
                        borderRadius: BorderRadius.circular(20)),
                    child: Text(statusLabel,
                        style: GoogleFonts.outfit(
                            fontSize: 10,
                            fontWeight: FontWeight.w900,
                            color: statusColor)),
                  ),
                  if (isDraft) ...[
                    const SizedBox(height: 6),
                    Text(
                      'Menunggu verifikasi Admin.',
                      style: GoogleFonts.outfit(
                          fontSize: 11,
                          color: const Color(0xFF92400E),
                          fontStyle: FontStyle.italic),
                    ),
                  ],
                ],
              ),
            ),
            if (isDraft)
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  IconButton(
                    icon: const Icon(Icons.edit_note_rounded, color: Colors.blue),
                    onPressed: () async {
                      final result = await Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => EditSalesOrderPage(
                              salesOrder: Map<String, dynamic>.from(so as Map)),
                        ),
                      );
                      if (result == true) _fetchSalesOrders();
                    },
                  ),
                  IconButton(
                    icon: const Icon(Icons.delete_rounded, color: Colors.red),
                    onPressed: () => _confirmDeleteSO(so['id'] ?? so['ID'], soNumber),
                  ),
                ],
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildBannerCard(dynamic txn) {
    final storeName = txn['store']?['name'] ?? 'Toko Tidak Diketahui';
    final dateObj = DateTime.parse(txn['created_at']).toLocal();
    final dateStr = DateFormat('dd MMM yyyy, HH:mm').format(dateObj);
    final status = txn['status'] ?? '';
    final type = (txn['type'] ?? 'BANNER').toString().toUpperCase();
    final width = txn['size_width'] ?? 0;
    final length = txn['size_length'] ?? 0;
    final notes = txn['notes'] ?? '-';

    final bool isRequest = status == 'REQUEST';

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withValues(alpha: 0.04),
              blurRadius: 20,
              offset: const Offset(0, 8))
        ],
      ),
      clipBehavior: Clip.antiAlias,
      child: Material(
        color: Colors.transparent,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                width: 56,
                height: 56,
                decoration: BoxDecoration(
                  color: Colors.purple.withValues(alpha: 0.1),
                  shape: BoxShape.circle,
                ),
                child: const Icon(
                  Icons.branding_watermark_rounded,
                  color: Colors.purple,
                  size: 28,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(storeName,
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w800,
                            fontSize: 16,
                            color: const Color(0xFF1E293B))),
                    const SizedBox(height: 4),
                    Text('PESANAN $type (${width}x${length}m)',
                        style: GoogleFonts.outfit(
                            fontSize: 12,
                            color: Colors.purple,
                            fontWeight: FontWeight.w800)),
                    const SizedBox(height: 4),
                    Text(dateStr,
                        style: GoogleFonts.outfit(
                            fontSize: 12,
                            color: Colors.blueGrey.withValues(alpha: 0.8))),
                    const SizedBox(height: 6),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 10, vertical: 3),
                      decoration: BoxDecoration(
                        color: isRequest
                            ? Colors.orange.withValues(alpha: 0.1)
                            : Colors.purple.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Text(
                        status,
                        style: GoogleFonts.outfit(
                          fontSize: 10,
                          fontWeight: FontWeight.w900,
                          color: isRequest ? Colors.orange : Colors.purple,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              if (isRequest)
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    IconButton(
                      icon: const Icon(Icons.edit_rounded,
                          color: Colors.blueAccent),
                      onPressed: () => _editBanner(txn),
                    ),
                    IconButton(
                      icon: const Icon(Icons.delete_rounded,
                          color: Colors.redAccent),
                      onPressed: () => _deleteBanner(txn['id']),
                    ),
                  ],
                ),
            ],
          ),
        ),
      ),
    );
  }

  void _editBanner(dynamic txn) {
    showDialog(
      context: context,
      builder: (ctx) => _EditBannerDialog(
        txn: txn,
        onSuccess: _fetchHistory,
      ),
    );
  }

  void _deleteBanner(dynamic id) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        title: Row(
          children: [
            const Icon(Icons.warning_amber_rounded, color: Colors.red, size: 28),
            const SizedBox(width: 8),
            Text('Peringatan Hapus',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w800, fontSize: 18)),
          ],
        ),
        content: Text(
          'Apakah Anda yakin ingin menghapus pesanan spanduk ini? Tindakan ini tidak dapat dikembalikan.',
          style: GoogleFonts.outfit(fontSize: 14, color: Colors.blueGrey),
        ),
        actionsPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx), 
            child: Text('Batal', style: GoogleFonts.outfit(fontWeight: FontWeight.bold, color: Colors.blueGrey))
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.redAccent,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              elevation: 0,
            ),
            onPressed: () async {
              Navigator.pop(ctx);
              setState(() => _isLoading = true);
              try {
                final api = BannerOrderApiService(apiClient: di.sl<ApiClient>());
                await api.deleteBannerOrder(id);
                _fetchHistory();
              } catch (e) {
                setState(() => _isLoading = false);
                ScaffoldMessenger.of(context)
                    .showSnackBar(SnackBar(content: Text(e.toString())));
              }
            },
            child: Text('Ya, Hapus', style: GoogleFonts.outfit(fontWeight: FontWeight.bold, color: Colors.white)),
          ),
        ],
      ),
    );
  }
}

class _EditBannerDialog extends StatefulWidget {
  final dynamic txn;
  final VoidCallback onSuccess;

  const _EditBannerDialog({
    required this.txn,
    required this.onSuccess,
  });

  @override
  State<_EditBannerDialog> createState() => _EditBannerDialogState();
}

class _EditBannerDialogState extends State<_EditBannerDialog> {
  final _formKey = GlobalKey<FormState>();
  late String _bannerType;
  String _sizeUnit = 'm'; // 'm' or 'cm'
  late TextEditingController _lengthCtrl;
  late TextEditingController _widthCtrl;
  late TextEditingController _qtyCtrl;
  late TextEditingController _notesCtrl;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    final rawType = widget.txn['type']?.toString() ?? 'Custom';
    _bannerType =
        (rawType.toLowerCase() == 'universal') ? 'Universal' : 'Custom';

    final rawLen = double.tryParse(widget.txn['size_length']?.toString() ?? '') ?? 0;
    final rawWid = double.tryParse(widget.txn['size_width']?.toString() ?? '') ?? 0;
    
    // In backend/database, dimensions are saved in cm (e.g. 300 x 100) or occasionally in m (< 15)
    // If raw value is > 15, it's stored in cm, so default to 'm' with converted value
    double initLenM = rawLen > 15 ? rawLen / 100 : rawLen;
    double initWidM = rawWid > 15 ? rawWid / 100 : rawWid;

    final lenText = initLenM <= 0 ? '' : (initLenM % 1 == 0 ? initLenM.toInt().toString() : initLenM.toString());
    final widText = initWidM <= 0 ? '' : (initWidM % 1 == 0 ? initWidM.toInt().toString() : initWidM.toString());
    final q = (widget.txn['qty'] ?? 1).toString();
    final n = (widget.txn['notes'] == '-' || widget.txn['notes'] == null)
        ? ''
        : widget.txn['notes'].toString();

    _lengthCtrl = TextEditingController(text: lenText);
    _widthCtrl = TextEditingController(text: widText);
    _qtyCtrl = TextEditingController(text: q);
    _notesCtrl = TextEditingController(text: n);

    _lengthCtrl.addListener(_onDimensionChanged);
    _widthCtrl.addListener(_onDimensionChanged);
  }

  void _onDimensionChanged() {
    if (mounted) setState(() {});
  }

  void _switchUnit(String newUnit) {
    if (_sizeUnit == newUnit) return;
    final lenStr = _lengthCtrl.text.trim().replaceAll(',', '.');
    final widStr = _widthCtrl.text.trim().replaceAll(',', '.');
    final lenVal = double.tryParse(lenStr);
    final widVal = double.tryParse(widStr);

    setState(() {
      _sizeUnit = newUnit;
      if (lenVal != null && lenVal > 0) {
        if (newUnit == 'cm') {
          final converted = lenVal * 100;
          _lengthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(1);
        } else {
          final converted = lenVal / 100;
          _lengthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(2);
        }
      }
      if (widVal != null && widVal > 0) {
        if (newUnit == 'cm') {
          final converted = widVal * 100;
          _widthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(1);
        } else {
          final converted = widVal / 100;
          _widthCtrl.text = converted % 1 == 0
              ? converted.toInt().toString()
              : converted.toStringAsFixed(2);
        }
      }
    });
  }

  void _applyPreset(double lengthM, double widthM) {
    setState(() {
      if (_sizeUnit == 'm') {
        _lengthCtrl.text = lengthM % 1 == 0
            ? lengthM.toInt().toString()
            : lengthM.toString();
        _widthCtrl.text = widthM % 1 == 0
            ? widthM.toInt().toString()
            : widthM.toString();
      } else {
        final lengthCm = lengthM * 100;
        final widthCm = widthM * 100;
        _lengthCtrl.text = lengthCm.toInt().toString();
        _widthCtrl.text = widthCm.toInt().toString();
      }
    });
  }

  @override
  void dispose() {
    _lengthCtrl.removeListener(_onDimensionChanged);
    _widthCtrl.removeListener(_onDimensionChanged);
    _lengthCtrl.dispose();
    _widthCtrl.dispose();
    _qtyCtrl.dispose();
    _notesCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _isLoading = true);

    try {
      final inputLen =
          double.parse(_lengthCtrl.text.trim().replaceAll(',', '.'));
      final inputWid =
          double.parse(_widthCtrl.text.trim().replaceAll(',', '.'));
      final lengthCm = _sizeUnit == 'm' ? (inputLen * 100) : inputLen;
      final widthCm = _sizeUnit == 'm' ? (inputWid * 100) : inputWid;
      final qty = int.tryParse(_qtyCtrl.text.trim()) ?? 1;

      final storeName =
          widget.txn['store']?['name'] ?? widget.txn['store_name'] ?? '';

      final payload = {
        'shop_name': storeName,
        'size_length': lengthCm,
        'size_width': widthCm,
        'type': _bannerType,
        'is_bihun': false,
        'qty': qty,
        'notes': _notesCtrl.text.trim(),
      };

      final api = BannerOrderApiService(apiClient: di.sl<ApiClient>());
      await api.updateBannerOrder(widget.txn['id'], payload);

      if (mounted) {
        Navigator.pop(context);
        widget.onSuccess();
        ScaffoldMessenger.of(context).showSnackBar(const SnackBar(
          content: Text('Order Spanduk berhasil diperbarui'),
          backgroundColor: Colors.green,
        ));
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString()), backgroundColor: Colors.red),
        );
      }
    }
  }

  Widget _buildDimensionHelper() {
    final lenStr = _lengthCtrl.text.trim().replaceAll(',', '.');
    final widStr = _widthCtrl.text.trim().replaceAll(',', '.');
    final lenVal = double.tryParse(lenStr);
    final widVal = double.tryParse(widStr);

    if (lenVal == null || lenVal <= 0 || widVal == null || widVal <= 0) {
      return const SizedBox.shrink();
    }

    final double lenCm = _sizeUnit == 'm' ? (lenVal * 100) : lenVal;
    final double widCm = _sizeUnit == 'm' ? (widVal * 100) : widVal;
    final double lenM = lenCm / 100;
    final double widM = widCm / 100;
    final double areaM2 = lenM * widM;

    final String alternateText = _sizeUnit == 'm'
        ? '≈ ${lenCm % 1 == 0 ? lenCm.toInt() : lenCm.toStringAsFixed(1)} cm × ${widCm % 1 == 0 ? widCm.toInt() : widCm.toStringAsFixed(1)} cm'
        : '≈ ${lenM.toStringAsFixed(2)} m × ${widM.toStringAsFixed(2)} m';

    return Container(
      margin: const EdgeInsets.only(top: 8),
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: AppColors.primaryGreen.withOpacity(0.08),
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: AppColors.primaryGreen.withOpacity(0.2)),
      ),
      child: Row(
        children: [
          const Icon(Icons.swap_horiz_rounded,
              color: AppColors.primaryGreen, size: 16),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              'Konversi: $alternateText  •  Luas: ${areaM2.toStringAsFixed(2)} m²',
              style: GoogleFonts.outfit(
                fontSize: 11.5,
                fontWeight: FontWeight.w700,
                color: const Color(0xFF0F766E),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildField(String label, String hint, TextEditingController ctrl,
      IconData icon,
      {bool isDecimal = false,
      bool isNumber = false,
      bool required = true,
      int maxLines = 1}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label,
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w700,
                color: const Color(0xFF475569),
                fontSize: 12)),
        const SizedBox(height: 6),
        TextFormField(
          controller: ctrl,
          keyboardType: isDecimal
              ? const TextInputType.numberWithOptions(decimal: true)
              : (isNumber ? TextInputType.number : TextInputType.text),
          maxLines: maxLines,
          style: GoogleFonts.outfit(
              fontSize: 14,
              fontWeight: FontWeight.w600,
              color: const Color(0xFF1E293B)),
          decoration: InputDecoration(
            prefixIcon: Icon(icon, color: const Color(0xFF64748B), size: 20),
            hintText: hint,
            hintStyle:
                GoogleFonts.outfit(color: Colors.grey.shade400, fontSize: 13),
            filled: true,
            fillColor: const Color(0xFFF8FAFC),
            border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
                borderSide: BorderSide(color: Colors.grey.shade200)),
            enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
                borderSide: BorderSide(color: Colors.grey.shade200)),
            focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(14),
                borderSide: const BorderSide(
                    color: AppColors.primaryGreen, width: 2)),
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
          ),
          validator: (v) {
            if (required && (v == null || v.trim().isEmpty)) {
              return 'Field ini wajib diisi';
            }
            if (isDecimal && v != null && v.trim().isNotEmpty) {
              final val = double.tryParse(v.trim().replaceAll(',', '.'));
              if (val == null || val <= 0) {
                return 'Angka harus > 0 (contoh: 3.0)';
              }
            }
            if (isNumber && v != null && v.trim().isNotEmpty) {
              final val = int.tryParse(v.trim());
              if (val == null || val <= 0) {
                return 'Jumlah harus > 0';
              }
            }
            return null;
          },
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final storeName =
        widget.txn['store']?['name'] ?? widget.txn['store_name'] ?? '';

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      elevation: 0,
      backgroundColor: Colors.white,
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 480),
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Header
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('Edit Request Spanduk',
                        style: GoogleFonts.outfit(
                            fontSize: 18,
                            fontWeight: FontWeight.w800,
                            color: const Color(0xFF1E293B))),
                    InkWell(
                      onTap: () => Navigator.pop(context),
                      child: Container(
                        padding: const EdgeInsets.all(6),
                        decoration: BoxDecoration(
                            color: Colors.grey.shade100, shape: BoxShape.circle),
                        child: const Icon(Icons.close_rounded,
                            size: 20, color: Colors.blueGrey),
                      ),
                    )
                  ],
                ),
                const SizedBox(height: 16),
                // Store badge
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF1F5F9),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    children: [
                      const Icon(Icons.storefront_rounded,
                          size: 20, color: Color(0xFF475569)),
                      const SizedBox(width: 10),
                      Expanded(
                        child: Text(
                          storeName.isNotEmpty ? storeName : 'Toko',
                          style: GoogleFonts.outfit(
                              fontSize: 14,
                              fontWeight: FontWeight.w700,
                              color: const Color(0xFF1E293B)),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 18),
                // Tipe Spanduk
                Text('Tipe Spanduk *',
                    style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w700,
                        color: const Color(0xFF475569),
                        fontSize: 12)),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: InkWell(
                        onTap: () => setState(() => _bannerType = 'Custom'),
                        borderRadius: BorderRadius.circular(14),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                              vertical: 12, horizontal: 12),
                          decoration: BoxDecoration(
                            color: _bannerType == 'Custom'
                                ? AppColors.primaryGreen.withOpacity(0.12)
                                : const Color(0xFFF8FAFC),
                            borderRadius: BorderRadius.circular(14),
                            border: Border.all(
                              color: _bannerType == 'Custom'
                                  ? AppColors.primaryGreen
                                  : Colors.grey.shade200,
                              width: _bannerType == 'Custom' ? 1.5 : 1,
                            ),
                          ),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(
                                _bannerType == 'Custom'
                                    ? Icons.radio_button_checked_rounded
                                    : Icons.radio_button_off_rounded,
                                size: 18,
                                color: _bannerType == 'Custom'
                                    ? AppColors.primaryGreen
                                    : Colors.grey.shade400,
                              ),
                              const SizedBox(width: 6),
                              Text(
                                'Custom',
                                style: GoogleFonts.outfit(
                                  fontSize: 13,
                                  fontWeight: FontWeight.w800,
                                  color: _bannerType == 'Custom'
                                      ? AppColors.primaryGreen
                                      : const Color(0xFF334155),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: InkWell(
                        onTap: () => setState(() => _bannerType = 'Universal'),
                        borderRadius: BorderRadius.circular(14),
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                              vertical: 12, horizontal: 12),
                          decoration: BoxDecoration(
                            color: _bannerType == 'Universal'
                                ? const Color(0xFF0284C7).withOpacity(0.12)
                                : const Color(0xFFF8FAFC),
                            borderRadius: BorderRadius.circular(14),
                            border: Border.all(
                              color: _bannerType == 'Universal'
                                  ? const Color(0xFF0284C7)
                                  : Colors.grey.shade200,
                              width: _bannerType == 'Universal' ? 1.5 : 1,
                            ),
                          ),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Icon(
                                _bannerType == 'Universal'
                                    ? Icons.radio_button_checked_rounded
                                    : Icons.radio_button_off_rounded,
                                size: 18,
                                color: _bannerType == 'Universal'
                                    ? const Color(0xFF0284C7)
                                    : Colors.grey.shade400,
                              ),
                              const SizedBox(width: 6),
                              Text(
                                'Universal',
                                style: GoogleFonts.outfit(
                                  fontSize: 13,
                                  fontWeight: FontWeight.w800,
                                  color: _bannerType == 'Universal'
                                      ? const Color(0xFF0284C7)
                                      : const Color(0xFF334155),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                // Header Dimensi & Unit Switcher
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    Text(
                      'Ukuran Dimensi',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w700,
                        color: const Color(0xFF475569),
                        fontSize: 12,
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.all(2),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF1F5F9),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.grey.shade300),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          GestureDetector(
                            onTap: () => _switchUnit('m'),
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 8, vertical: 4),
                              decoration: BoxDecoration(
                                color: _sizeUnit == 'm'
                                    ? AppColors.primaryGreen
                                    : Colors.transparent,
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(
                                'Meter (m)',
                                style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  fontWeight: _sizeUnit == 'm'
                                      ? FontWeight.w800
                                      : FontWeight.w600,
                                  color: _sizeUnit == 'm'
                                      ? Colors.white
                                      : const Color(0xFF64748B),
                                ),
                              ),
                            ),
                          ),
                          GestureDetector(
                            onTap: () => _switchUnit('cm'),
                            child: Container(
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 8, vertical: 4),
                              decoration: BoxDecoration(
                                color: _sizeUnit == 'cm'
                                    ? AppColors.primaryGreen
                                    : Colors.transparent,
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(
                                'Centimeter (cm)',
                                style: GoogleFonts.outfit(
                                  fontSize: 11,
                                  fontWeight: _sizeUnit == 'cm'
                                      ? FontWeight.w800
                                      : FontWeight.w600,
                                  color: _sizeUnit == 'cm'
                                      ? Colors.white
                                      : const Color(0xFF64748B),
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      {'label': '2 × 1 m', 'l': 2.0, 'w': 1.0},
                      {'label': '3 × 1 m', 'l': 3.0, 'w': 1.0},
                      {'label': '3 × 0.8 m', 'l': 3.0, 'w': 0.8},
                      {'label': '4 × 1 m', 'l': 4.0, 'w': 1.0},
                    ].map((p) {
                      return Padding(
                        padding: const EdgeInsets.only(right: 6),
                        child: InkWell(
                          borderRadius: BorderRadius.circular(8),
                          onTap: () => _applyPreset(
                              p['l'] as double, p['w'] as double),
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                                horizontal: 8, vertical: 4),
                            decoration: BoxDecoration(
                              color: Colors.white,
                              borderRadius: BorderRadius.circular(6),
                              border: Border.all(color: Colors.grey.shade300),
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                const Icon(Icons.straighten_rounded,
                                    size: 12, color: AppColors.primaryGreen),
                                const SizedBox(width: 4),
                                Text(
                                  p['label'] as String,
                                  style: GoogleFonts.outfit(
                                    fontSize: 10.5,
                                    fontWeight: FontWeight.w700,
                                    color: const Color(0xFF334155),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                ),
                const SizedBox(height: 10),
                // Dimensi Lebar & Tinggi
                Row(
                  children: [
                    Expanded(
                      child: _buildField(
                        'Panjang ($_sizeUnit) *',
                        _sizeUnit == 'm' ? 'Contoh: 3.0' : 'Contoh: 300',
                        _lengthCtrl,
                        Icons.width_normal_rounded,
                        isDecimal: true,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: _buildField(
                        'Tinggi ($_sizeUnit) *',
                        _sizeUnit == 'm' ? 'Contoh: 1.0' : 'Contoh: 100',
                        _widthCtrl,
                        Icons.height_rounded,
                        isDecimal: true,
                      ),
                    ),
                  ],
                ),
                _buildDimensionHelper(),
                const SizedBox(height: 16),
                // Qty
                _buildField('Banyaknya (Qty) *', '1', _qtyCtrl,
                    Icons.format_list_numbered_rounded,
                    isNumber: true),
                const SizedBox(height: 16),
                // Notes
                _buildField(
                    'Catatan Tambahan',
                    'Tulis pesan atau keterangan desain...',
                    _notesCtrl,
                    Icons.notes_rounded,
                    required: false,
                    maxLines: 3),
                const SizedBox(height: 24),
                // Submit Button
                SizedBox(
                  width: double.infinity,
                  height: 52,
                  child: ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppColors.primaryGreen,
                      shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(16)),
                      elevation: 0,
                    ),
                    onPressed: _isLoading ? null : _submit,
                    child: _isLoading
                        ? const SizedBox(
                            width: 24,
                            height: 24,
                            child: CircularProgressIndicator(
                                strokeWidth: 2.5, color: Colors.white),
                          )
                        : Text('SIMPAN PERUBAHAN',
                            style: GoogleFonts.outfit(
                                fontSize: 14,
                                fontWeight: FontWeight.w800,
                                color: Colors.white,
                                letterSpacing: 0.5)),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
