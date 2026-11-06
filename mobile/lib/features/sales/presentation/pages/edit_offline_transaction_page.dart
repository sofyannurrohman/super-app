import 'dart:convert';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/database/database.dart';
import 'package:drift/drift.dart' as drift;
import 'package:hris_app/injection.dart' as di;
import 'package:intl/intl.dart';
import 'package:hris_app/features/sales/data/services/sales_api_service.dart';
import 'package:hris_app/core/utils/constants.dart';
import 'package:hris_app/core/services/data_integrity_service.dart';

class EditOfflineTransactionPage extends StatefulWidget {
  final String? localId;
  final Map<String, dynamic> transaction;

  const EditOfflineTransactionPage(
      {super.key, required this.localId, required this.transaction});

  @override
  State<EditOfflineTransactionPage> createState() =>
      _EditOfflineTransactionPageState();
}

class _EditOfflineTransactionPageState
    extends State<EditOfflineTransactionPage> {
  final _db = di.sl<AppDatabase>();
  final _currency =
      NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);

  bool _isLoading = true;
  List<Map<String, dynamic>> _items = [];
  List<Map<String, dynamic>> _availableProducts = [];
  final _apiService = di.sl<SalesApiService>();

  double get _grandTotal {
    return _items.fold(0.0, (sum, item) {
      final qty = (item['quantity'] as num?)?.toInt() ?? 0;
      final price = (item['price'] as num?)?.toDouble() ?? 0.0;
      return sum + (qty * price);
    });
  }

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Widget _buildImage(String? rawUrl) {
    if (rawUrl == null || rawUrl.isEmpty) {
      return const Icon(Icons.inventory_2_outlined,
          size: 18, color: Colors.grey);
    }
    String imageUrl = rawUrl;
    if (!rawUrl.startsWith('http') && !rawUrl.startsWith('assets/')) {
      imageUrl = '${AppConstants.baseUrl.replaceAll('/api/v1/', '')}$rawUrl';
    }

    if (imageUrl.startsWith('http')) {
      return CachedNetworkImage(imageUrl: imageUrl, fit: BoxFit.cover, placeholder: (context, url) => const Center(child: CircularProgressIndicator()), errorWidget: (context, url, error) => const Icon(Icons.error), memCacheWidth: 800,);
    } else if (imageUrl.startsWith('assets/')) {
      return Image.asset(
        imageUrl,
        fit: BoxFit.cover,
        errorBuilder: (_, __, ___) => const Icon(Icons.inventory_2_outlined,
            size: 18, color: Colors.grey),
      );
    }
    return const Icon(Icons.inventory_2_outlined, size: 18, color: Colors.grey);
  }

  Future<void> _loadData() async {
    if (widget.localId != null) {
      final localItems = await (_db.select(_db.localTransactionItems)
            ..where((t) => t.transactionLocalId.equals(widget.localId!)))
          .get();

      setState(() {
        _items = localItems
            .map<Map<String, dynamic>>((i) => <String, dynamic>{
                  'id': i.id,
                  'product_id': i.productId,
                  'product_name': i.productName,
                  'quantity': i.quantity,
                  'price': i.price,
                  'unit': i.unit,
                  'is_new': false,
                })
            .toList();

        _isLoading = false;
      });
    } else {
      // Remote only transaction
      setState(() {
        _items = (widget.transaction['items'] as List?)
                ?.map<Map<String, dynamic>>((i) => <String, dynamic>{
                      'id': i['id'] ?? DateTime.now().microsecondsSinceEpoch,
                      'product_id': i['product_id'] ?? i['inventory_id'] ?? i['id'],
                      'product_name': i['product_name'] ?? i['name'] ?? 'Unknown',
                      'quantity': i['quantity'] ?? 0,
                      'price': (i['price'] ?? i['selling_price'] ?? 0 as num).toDouble(),
                      'unit': i['unit'] ?? 'PCS',
                      'is_new': false,
                    })
                .toList() ?? [];

        _isLoading = false;
      });
    }

    _fetchProducts();
  }

  Future<void> _fetchProducts() async {
    try {
      bool isInstantOrder = false;
      if (widget.localId != null) {
        final localTx = await (_db.select(_db.localTransactions)
              ..where((t) => t.localId.equals(widget.localId!)))
            .getSingleOrNull();
        isInstantOrder = localTx?.companyName == 'Instant Order';
      } else {
        isInstantOrder = widget.transaction['company_name'] == 'Instant Order' || 
                         widget.transaction['store']?['name'] == 'Instant Order';
      }

      if (isInstantOrder) {
        final res = await _apiService.apiClient.client.get('factory/products');
        if (res.statusCode == 200) {
          List? raw;
          if (res.data is Map && res.data.containsKey('data')) {
            raw = res.data['data'] as List?;
          } else if (res.data is List) {
            raw = res.data as List?;
          }
          if (raw != null && mounted) {
            final nonNullRaw = raw;
            setState(() {
              _availableProducts = nonNullRaw
                  .map((p) => {
                        'id': p['id']?.toString() ?? '',
                        'name': p['name']?.toString() ?? '',
                        'sku': p['sku']?.toString(),
                        'sellingPrice': ((p['selling_price'] ??
                                p['sellingPrice'] ??
                                p['SellingPrice'] ??
                                0) as num)
                            .toDouble(),
                        'stock':
                            9999, // Instant orders have no strict stock limit
                        'image_url': p['image_url']?.toString(),
                      })
                  .toList();
            });
            return;
          }
        }
      }

      final remoteStocks = await _apiService.getBranchWarehouseStock();
      debugPrint('[EditTx] remoteStocks count: ${remoteStocks.length}');
      if (remoteStocks.isNotEmpty) {
        debugPrint('[EditTx] sample stock entry: ${remoteStocks.first}');
      }
      if (mounted) {
        setState(() {
          final Map<String, Map<String, dynamic>> uniqueProducts = {};
          for (var s in remoteStocks) {
            final prod = s['product'];
            if (prod != null) {
              final pid = prod['id']?.toString() ?? '';
              if (!uniqueProducts.containsKey(pid)) {
                uniqueProducts[pid] = {
                  'id': pid,
                  'name': prod['name']?.toString() ?? '',
                  'sku': prod['sku']?.toString(),
                  'sellingPrice': (prod['selling_price'] ??
                          prod['price'] ??
                          0.0) is num
                      ? ((prod['selling_price'] ?? prod['price'] ?? 0.0) as num)
                          .toDouble()
                      : 0.0,
                  'stock': (s['quantity'] as num?)?.toInt() ?? 0,
                  'image_url': prod['image_url']?.toString(),
                };
              } else {
                uniqueProducts[pid]!['stock'] =
                    (uniqueProducts[pid]!['stock'] as int) +
                        ((s['quantity'] as num?)?.toInt() ?? 0);
              }
            }
          }
          _availableProducts = uniqueProducts.values.toList();
          debugPrint(
              '[EditTx] _availableProducts count: ${_availableProducts.length}');
        });
      }
    } catch (e) {
      debugPrint('[EditTx] Error fetching products from api: $e');
      // Fallback to local products
      final products = await _db.select(_db.products).get();
      if (mounted) {
        setState(() {
          _availableProducts = products
              .map((p) => {
                    'id': p.id,
                    'name': p.name,
                    'sku': p.sku,
                    'sellingPrice': p.sellingPrice,
                    'stock': 0,
                    'image_url': (p as dynamic).imageUrl?.toString(),
                  })
              .toList();
          debugPrint(
              '[EditTx] fallback local products count: ${_availableProducts.length}');
        });
      }
    }
  }

  Future<void> _showAddProductModal() async {
    debugPrint('[EditTx] FAB clicked — fetching products...');

    // Show loading snackbar while fetching
    if (_availableProducts.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Memuat daftar produk...', style: GoogleFonts.outfit()),
          duration: const Duration(seconds: 2),
          backgroundColor: Colors.indigo,
        ),
      );
      await _fetchProducts();
      if (!mounted) return;
      ScaffoldMessenger.of(context).hideCurrentSnackBar();
    }

    if (_availableProducts.isEmpty) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Produk tidak tersedia.')));
      return;
    }

    setState(() {
      final defaultProduct = _availableProducts.first;
      _items.add(<String, dynamic>{
        'id': null,
        'product_id': defaultProduct['id'],
        'product_name': defaultProduct['name'],
        'quantity': 1,
        'price': (defaultProduct['sellingPrice'] as num?)?.toDouble() ?? 0.0,
        'unit': defaultProduct['unit']?.toString() ?? 'PCS',
        'is_new': true,
      });
    });
  }

  Future<void> _showEditPriceDialog(int index) async {
    final item = _items[index];
    final TextEditingController priceCtrl =
        TextEditingController(text: item['price'].toString());

    await showDialog(
        context: context,
        builder: (ctx) {
          return AlertDialog(
            shape:
                RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
            title: Text('Edit Harga Produk',
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w800,
                    color: const Color(0xFF1E293B))),
            content: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(item['product_name'],
                    style: GoogleFonts.outfit(color: const Color(0xFF64748B))),
                const SizedBox(height: 16),
                TextField(
                  controller: priceCtrl,
                  keyboardType:
                      const TextInputType.numberWithOptions(decimal: true),
                  decoration: InputDecoration(
                    labelText: 'Harga Jual (Rp)',
                    labelStyle:
                        GoogleFonts.outfit(color: const Color(0xFF64748B)),
                    prefixText: 'Rp ',
                    border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12)),
                    focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide:
                            const BorderSide(color: Colors.indigo, width: 2)),
                  ),
                ),
              ],
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: Text('BATAL',
                    style: GoogleFonts.outfit(color: Colors.grey)),
              ),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: const Color(0xFF10B981),
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8)),
                ),
                onPressed: () {
                  final newPrice =
                      double.tryParse(priceCtrl.text) ?? item['price'];
                  setState(() {
                    _items[index]['price'] = newPrice;
                  });
                  Navigator.pop(ctx);
                },
                child: Text('SIMPAN',
                    style: GoogleFonts.outfit(
                        color: Colors.white, fontWeight: FontWeight.w700)),
              ),
            ],
          );
        });
  }

  Future<void> _saveChanges() async {
    setState(() => _isLoading = true);
    try {
      double totalAmount = 0;
      final isAlreadySynced = widget.transaction['is_local'] != true;

      for (var item in _items) {
        final int newQty = item['quantity'];
        final double price = item['price'];
        totalAmount += newQty * price;
      }

      if (isAlreadySynced) {
        try {
          final payload = {
            'total_amount': totalAmount,
            'items': _items.map((i) => {
              'product_id': i['product_id'],
              'quantity': i['quantity'],
              'price': i['price'],
              'unit': i['unit'],
              'pieces_per_unit': i['pieces_per_unit'] ?? 1,
            }).toList()
          };
          
          await _apiService.apiClient.client.put(
            'sales/transactions/${widget.transaction['id']}', 
            data: payload
          );
        } catch (e) {
           throw Exception('Gagal menyimpan ke server: $e');
        }
      }

      if (widget.localId != null) {
        await _db.transaction(() async {
          final originalItems = await (_db.select(_db.localTransactionItems)
                ..where((t) => t.transactionLocalId.equals(widget.localId!)))
              .get();

          // 1. Handle Deletions (Items in DB but not in _items)
          for (var orig in originalItems) {
            final stillExists =
                _items.any((i) => i['id'] == orig.id && i['is_new'] == false);
            if (!stillExists) {
              // Restore stock
              final stock = await (_db.select(_db.salesStock)
                    ..where((t) => t.productId.equals(orig.productId)))
                  .getSingleOrNull();
              if (stock != null) {
                await (_db.update(_db.salesStock)
                      ..where((t) => t.productId.equals(orig.productId)))
                    .write(
                  SalesStockCompanion(
                      quantity: drift.Value(stock.quantity + orig.quantity)),
                );
              }
              // Delete item
              await (_db.delete(_db.localTransactionItems)
                    ..where((t) => t.id.equals(orig.id)))
                  .go();
            }
          }

          // 2. Handle Updates and Insertions
          for (var item in _items) {
            final int newQty = item['quantity'];
            final double price = item['price'];

            if (item['is_new']) {
              // Deduct stock
              final stock = await (_db.select(_db.salesStock)
                    ..where((t) => t.productId.equals(item['product_id'])))
                  .getSingleOrNull();
              if (stock != null) {
                await (_db.update(_db.salesStock)
                      ..where((t) => t.productId.equals(item['product_id'])))
                    .write(
                  SalesStockCompanion(
                      quantity: drift.Value(stock.quantity - newQty)),
                );
              }
              // Insert new item
              await _db
                  .into(_db.localTransactionItems)
                  .insert(LocalTransactionItemsCompanion.insert(
                    transactionLocalId: widget.localId!,
                    productId: item['product_id'],
                    productName: item['product_name'],
                    quantity: newQty,
                    orderedQuantity: drift.Value(newQty),
                    price: price,
                    unit: drift.Value(item['unit']),
                  ));
            } else {
              // Update existing item
              final currentItem =
                  originalItems.firstWhere((orig) => orig.id == item['id']);
              
              if (currentItem.productId != item['product_id']) {
                // Jika produk diganti:
                // 1. Kembalikan stok produk lama
                final oldStock = await (_db.select(_db.salesStock)
                      ..where((t) => t.productId.equals(currentItem.productId)))
                    .getSingleOrNull();
                if (oldStock != null) {
                  await (_db.update(_db.salesStock)
                        ..where((t) => t.productId.equals(currentItem.productId)))
                      .write(SalesStockCompanion(
                          quantity: drift.Value(oldStock.quantity + currentItem.quantity)));
                }
                // 2. Kurangi stok produk baru
                final newStock = await (_db.select(_db.salesStock)
                      ..where((t) => t.productId.equals(item['product_id'])))
                    .getSingleOrNull();
                if (newStock != null) {
                  await (_db.update(_db.salesStock)
                        ..where((t) => t.productId.equals(item['product_id'])))
                      .write(SalesStockCompanion(
                          quantity: drift.Value(newStock.quantity - newQty)));
                }
              } else {
                // Jika produk sama, hitung selisih qty saja
                final int diff = newQty - currentItem.quantity;
                if (diff != 0) {
                  final stock = await (_db.select(_db.salesStock)
                        ..where((t) => t.productId.equals(item['product_id'])))
                      .getSingleOrNull();
                  if (stock != null) {
                    await (_db.update(_db.salesStock)
                          ..where((t) => t.productId.equals(item['product_id'])))
                        .write(
                      SalesStockCompanion(
                          quantity: drift.Value(stock.quantity - diff)),
                    );
                  }
                }
              }

              await (_db.update(_db.localTransactionItems)
                    ..where((t) => t.id.equals(item['id'])))
                  .write(
                LocalTransactionItemsCompanion(
                  productId: drift.Value(item['product_id']),
                  productName: drift.Value(item['product_name']),
                  price: drift.Value(price),
                  unit: drift.Value(item['unit']),
                  quantity: drift.Value(newQty),
                  orderedQuantity: drift.Value(newQty),
                ),
              );
            }
          }

          // 3. Update transaction total
          await (_db.update(_db.localTransactions)
                ..where((t) => t.localId.equals(widget.localId!)))
              .write(
            LocalTransactionsCompanion(
              totalAmount: drift.Value(totalAmount),
              paidAmount: drift.Value(totalAmount),
              syncStatus: isAlreadySynced 
                  ? const drift.Value('synced') 
                  : const drift.Value('pending'),
            ),
          );
        });
      } else {
        // Just update stock for purely remote transactions!
        await _db.transaction(() async {
          final originalItems = widget.transaction['items'] as List? ?? [];
          
          // 1. Handle Deletions (In original but not in _items)
          for (var orig in originalItems) {
            final String origProductId = orig['product_id']?.toString() ?? orig['inventory_id']?.toString() ?? '';
            final int origQty = (orig['quantity'] as num?)?.toInt() ?? 0;
            
            final stillExists = _items.any((i) => i['product_id']?.toString() == origProductId && i['is_new'] == false);
            if (!stillExists) {
              // Restore stock
              final stock = await (_db.select(_db.salesStock)
                    ..where((t) => t.productId.equals(origProductId)))
                  .getSingleOrNull();
              if (stock != null) {
                await (_db.update(_db.salesStock)
                      ..where((t) => t.productId.equals(origProductId)))
                    .write(
                  SalesStockCompanion(
                      quantity: drift.Value(stock.quantity + origQty)),
                );
              }
            }
          }
          
          // 2. Handle Updates and Insertions
          for (var item in _items) {
             final int newQty = item['quantity'];
             final String productId = item['product_id']?.toString() ?? '';
             
             if (item['is_new']) {
               // Deduct stock
               final stock = await (_db.select(_db.salesStock)
                     ..where((t) => t.productId.equals(productId)))
                   .getSingleOrNull();
               if (stock != null) {
                 await (_db.update(_db.salesStock)
                       ..where((t) => t.productId.equals(productId)))
                     .write(
                   SalesStockCompanion(
                       quantity: drift.Value(stock.quantity - newQty)),
                 );
               }
             } else {
               // Update existing item
               final orig = originalItems.firstWhere((o) => (o['product_id']?.toString() ?? o['inventory_id']?.toString()) == productId, orElse: () => null);
               if (orig != null) {
                 final int origQty = (orig['quantity'] as num?)?.toInt() ?? 0;
                 final int diff = newQty - origQty;
                 if (diff != 0) {
                   final stock = await (_db.select(_db.salesStock)
                         ..where((t) => t.productId.equals(productId)))
                       .getSingleOrNull();
                   if (stock != null) {
                     await (_db.update(_db.salesStock)
                           ..where((t) => t.productId.equals(productId)))
                         .write(
                       SalesStockCompanion(
                           quantity: drift.Value(stock.quantity - diff)),
                     );
                   }
                 }
               }
             }
          }
        });
      }

      if (mounted) {
        Navigator.pop(context, true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(
            content: Text('Gagal menyimpan: $e',
                style: GoogleFonts.outfit(color: Colors.white)),
            backgroundColor: Colors.red));
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _deleteTransaction() async {
    final currencyFormat =
        NumberFormat.currency(locale: 'id', symbol: 'Rp ', decimalDigits: 0);
    final storeName = widget.transaction['store']?['name'] ??
        widget.transaction['store_name'] ??
        'Toko';
    final totalAmount = _grandTotal;

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
      if (widget.localId != null &&
          widget.transaction['is_local'] == true &&
          widget.transaction['sync_status'] == 'pending' &&
          widget.transaction['last_sync_attempt'] == null) {
        await _db.transaction(() async {
          // 1. Ambil transaksi & items
          final tx = await (_db.select(_db.localTransactions)
                ..where((t) => t.localId.equals(widget.localId!)))
              .getSingleOrNull();
          if (tx == null ||
              tx.syncStatus != 'pending' ||
              tx.lastSyncAttempt != null) {
            throw StateError('Transaksi sudah diproses sinkronisasi. Muat ulang riwayat sebelum menghapus.');
          }
          final originalItems = await (_db.select(_db.localTransactionItems)
                ..where((t) => t.transactionLocalId.equals(widget.localId!)))
              .get();

          // 2. Kembalikan stok item reguler
          for (final orig in originalItems) {
            final stock = await (_db.select(_db.salesStock)
                  ..where((t) => t.productId.equals(orig.productId)))
                .getSingleOrNull();
            if (stock != null) {
              await (_db.update(_db.salesStock)
                    ..where((t) => t.productId.equals(orig.productId)))
                  .write(
                SalesStockCompanion(
                    quantity: drift.Value(stock.quantity + orig.quantity),
                    updatedAt: drift.Value(DateTime.now())),
              );
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
                        final stock = await (_db.select(_db.salesStock)
                              ..where((t) => t.productId.equals(pid)))
                            .getSingleOrNull();
                        if (stock != null) {
                          await (_db.update(_db.salesStock)
                                ..where((t) => t.productId.equals(pid)))
                              .write(SalesStockCompanion(
                            quantity: drift.Value(stock.quantity + qty),
                            updatedAt: drift.Value(DateTime.now()),
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
                        final stock = await (_db.select(_db.salesStock)
                              ..where((t) => t.productId.equals(pid)))
                            .getSingleOrNull();
                        if (stock != null) {
                          final newQty = (stock.quantity - qty).clamp(0, 999999);
                          await (_db.update(_db.salesStock)
                                ..where((t) => t.productId.equals(pid)))
                              .write(SalesStockCompanion(
                            quantity: drift.Value(newQty),
                            updatedAt: drift.Value(DateTime.now()),
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
          await (_db.delete(_db.localTransactionItems)
                ..where((t) => t.transactionLocalId.equals(widget.localId!)))
              .go();
          await (_db.delete(_db.localTransactions)
                ..where((t) => t.localId.equals(widget.localId!)))
              .go();
        });

        // 5. Bersihkan retry metadata
        await di.sl<DataIntegrityService>().resetRetry(widget.localId!);
        await di.sl<DataIntegrityService>().log(
              category: 'SALES',
              action: 'DELETE_OFFLINE_TRANSACTION',
              detail:
                  'localId=${widget.localId!} dihapus dari EditOfflineTransactionPage. Stok dikembalikan.',
            );
      } else {
        throw StateError('Hanya transaksi lokal yang belum disinkronkan yang dapat dihapus.');
      }

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
                'Transaksi offline berhasil dihapus dan stok dikembalikan.',
                style: GoogleFonts.outfit(fontWeight: FontWeight.w700)),
            backgroundColor: Colors.green,
          ),
        );
        Navigator.pop(context, true);
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
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final bool canDelete = widget.localId != null &&
        widget.transaction['is_local'] == true &&
        widget.transaction['sync_status'] == 'pending' &&
        widget.transaction['last_sync_attempt'] == null;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        title: Text('Edit Pesanan',
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w900, color: const Color(0xFF1E293B))),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF1E293B),
        elevation: 0,
        centerTitle: true,
        actions: [
          if (canDelete)
            IconButton(
              icon: Icon(Icons.delete_outline_rounded,
                  color: Colors.red.shade600),
              tooltip: 'Hapus Transaksi Offline',
              onPressed: _isLoading ? null : _deleteTransaction,
            ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _items.isEmpty
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.shopping_cart_outlined,
                          size: 64, color: Colors.grey.shade300),
                      const SizedBox(height: 16),
                      Text('Belum ada produk',
                          style: GoogleFonts.outfit(
                              fontSize: 18,
                              fontWeight: FontWeight.w700,
                              color: Colors.grey.shade500)),
                    ],
                  ),
                )
              : ListView.builder(
                  padding: const EdgeInsets.all(16),
                  itemCount: _items.length,
                  itemBuilder: (ctx, i) {
                    final item = _items[i];
                    return Container(
                      margin: const EdgeInsets.only(bottom: 12),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(20),
                        boxShadow: [
                          BoxShadow(
                              color: Colors.black.withValues(alpha: 0.02),
                              blurRadius: 10,
                              offset: const Offset(0, 4))
                        ],
                      ),
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Row(
                          children: [
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  _availableProducts.isEmpty
                                      ? Text(item['product_name'],
                                          style: GoogleFonts.outfit(
                                              fontWeight: FontWeight.w800,
                                              fontSize: 16,
                                              color: const Color(0xFF1E293B)))
                                      : DropdownButtonFormField<String>(
                                          value: _availableProducts.any((p) =>
                                                  p['id'].toString() ==
                                                  item['product_id'].toString())
                                              ? item['product_id'].toString()
                                              : null,
                                          hint: Text(
                                              item['product_name'] ??
                                                  'Pilih Produk',
                                              style: GoogleFonts.outfit(
                                                  fontWeight: FontWeight.w800,
                                                  fontSize: 16,
                                                  color:
                                                      const Color(0xFF1E293B))),
                                          decoration: InputDecoration(
                                            contentPadding:
                                                const EdgeInsets.symmetric(
                                                    horizontal: 12,
                                                    vertical: 8),
                                            border: OutlineInputBorder(
                                                borderRadius:
                                                    BorderRadius.circular(12),
                                                borderSide: BorderSide(
                                                    color:
                                                        Colors.grey.shade300)),
                                            enabledBorder: OutlineInputBorder(
                                                borderRadius:
                                                    BorderRadius.circular(12),
                                                borderSide: BorderSide(
                                                    color:
                                                        Colors.grey.shade300)),
                                            filled: true,
                                            fillColor: const Color(0xFFF8FAFC),
                                          ),
                                          isExpanded: true,
                                          icon: const Icon(
                                              Icons.arrow_drop_down_rounded,
                                              color: Color(0xFF10B981)),
                                          items: _availableProducts.map((p) {
                                            return DropdownMenuItem<String>(
                                              value: p['id'].toString(),
                                              child: Row(
                                                children: [
                                                  Container(
                                                    width: 32,
                                                    height: 32,
                                                    margin:
                                                        const EdgeInsets.only(
                                                            right: 12),
                                                    decoration: BoxDecoration(
                                                      color: Colors.white,
                                                      borderRadius:
                                                          BorderRadius.circular(
                                                              6),
                                                      border: Border.all(
                                                          color: Colors
                                                              .grey.shade200),
                                                    ),
                                                    child: ClipRRect(
                                                      borderRadius:
                                                          BorderRadius.circular(
                                                              6),
                                                      child: _buildImage(
                                                          p['image_url']
                                                              ?.toString()),
                                                    ),
                                                  ),
                                                  ConstrainedBox(
                                                    constraints: BoxConstraints(
                                                      maxWidth:
                                                          MediaQuery.of(context)
                                                                  .size
                                                                  .width -
                                                              130,
                                                    ),
                                                    child: Text(
                                                      p['name']?.toString() ??
                                                          '',
                                                      style: GoogleFonts.outfit(
                                                          fontWeight:
                                                              FontWeight.w700,
                                                          fontSize: 14,
                                                          color: const Color(
                                                              0xFF1E293B)),
                                                      overflow:
                                                          TextOverflow.ellipsis,
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            );
                                          }).toList(),
                                          onChanged: (newId) {
                                            if (newId != null) {
                                              final selectedProduct =
                                                  _availableProducts.firstWhere(
                                                      (p) =>
                                                          p['id'].toString() ==
                                                          newId);
                                              setState(() {
                                                item['product_id'] = newId;
                                                item['product_name'] =
                                                    selectedProduct['name'];
                                                item[
                                                    'price'] = (selectedProduct[
                                                                'sellingPrice']
                                                            as num?)
                                                        ?.toDouble() ??
                                                    0.0;
                                                item['unit'] =
                                                    selectedProduct['unit']
                                                            ?.toString() ??
                                                        'PCS';
                                                item['is_new'] =
                                                    true; // Mark as modified
                                              });
                                            }
                                          },
                                        ),
                                  const SizedBox(height: 6),
                                  InkWell(
                                    borderRadius: BorderRadius.circular(8),
                                    onTap: () {
                                      _showEditPriceDialog(i);
                                    },
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Text(_currency.format(item['price']),
                                            style: GoogleFonts.outfit(
                                                color: const Color(0xFF10B981),
                                                fontWeight: FontWeight.w800,
                                                fontSize: 14)),
                                        const SizedBox(width: 4),
                                        const Icon(Icons.edit_rounded,
                                            size: 14, color: Color(0xFF10B981)),
                                      ],
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            Container(
                              decoration: BoxDecoration(
                                color: const Color(0xFFF1F5F9),
                                borderRadius: BorderRadius.circular(16),
                              ),
                              child: Row(
                                children: [
                                  IconButton(
                                    icon: const Icon(Icons.remove,
                                        size: 20, color: Color(0xFF64748B)),
                                    onPressed: () {
                                      if (item['quantity'] > 1) {
                                        setState(() => item['quantity']--);
                                      } else {
                                        setState(() => _items.removeAt(i));
                                      }
                                    },
                                  ),
                                  SizedBox(
                                    width: 32,
                                    child: Text('${item['quantity']}',
                                        textAlign: TextAlign.center,
                                        style: GoogleFonts.outfit(
                                            fontSize: 16,
                                            fontWeight: FontWeight.w900,
                                            color: const Color(0xFF1E293B))),
                                  ),
                                  IconButton(
                                    icon: const Icon(Icons.add,
                                        size: 20, color: Color(0xFF64748B)),
                                    onPressed: () {
                                      setState(() => item['quantity']++);
                                    },
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                ),
      floatingActionButton: FloatingActionButton(
        onPressed: _showAddProductModal,
        backgroundColor: Colors.indigo,
        elevation: 4,
        child: const Icon(Icons.add_shopping_cart_rounded, color: Colors.white),
      ),
      bottomNavigationBar: Container(
        padding: const EdgeInsets.fromLTRB(20, 16, 20, 32),
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
                color: Colors.black.withValues(alpha: 0.05),
                blurRadius: 20,
                offset: const Offset(0, -5))
          ],
          borderRadius: const BorderRadius.vertical(top: Radius.circular(32)),
        ),
        child: SafeArea(
          child: Row(
            children: [
              Expanded(
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Total Pesanan',
                        style: GoogleFonts.outfit(
                            fontSize: 12,
                            color: Colors.blueGrey,
                            fontWeight: FontWeight.w600)),
                    Text(_currency.format(_grandTotal),
                        style: GoogleFonts.outfit(
                            fontSize: 20,
                            fontWeight: FontWeight.w900,
                            color: const Color(0xFF10B981))),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  minimumSize: const Size(120, 52),
                  backgroundColor: const Color(0xFF10B981),
                  padding:
                      const EdgeInsets.symmetric(vertical: 18, horizontal: 24),
                  shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16)),
                  elevation: 4,
                  shadowColor: const Color(0xFF10B981).withValues(alpha: 0.5),
                ),
                onPressed: _items.isEmpty ? null : _saveChanges,
                child: Text('SIMPAN',
                    style: GoogleFonts.outfit(
                        fontWeight: FontWeight.w900,
                        fontSize: 16,
                        color: Colors.white,
                        letterSpacing: 1)),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
