import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:hris_app/core/network/api_client.dart';
import 'package:hris_app/core/theme/app_colors.dart';
import 'package:hris_app/injection.dart' as di;
import 'package:dio/dio.dart';

class RequestItem {
  final String productId;
  final String productName;
  final int cartonCount;
  final int piecesPerCarton;
  final int totalQuantity;
  final String packagingType;
  final bool isBonus;

  RequestItem({
    required this.productId,
    required this.productName,
    required this.cartonCount,
    required this.piecesPerCarton,
    required this.totalQuantity,
    required this.packagingType,
    this.isBonus = false,
  });
}

class SalesStockRequestPage extends StatefulWidget {
  final bool isReturn;
  const SalesStockRequestPage({super.key, this.isReturn = false});

  @override
  State<SalesStockRequestPage> createState() => _SalesStockRequestPageState();
}

class _SalesStockRequestPageState extends State<SalesStockRequestPage> {
  final ApiClient apiClient = di.sl<ApiClient>();
  bool _isLoading = false;
  List<dynamic> _products = [];
  List<dynamic> _bonusProducts = [];
  bool _isBonusMode = false;
  final List<RequestItem> _requestItems = [];

  String? _selectedProductId;
  int _cartonCount = 1;
  int _piecesPerCarton = 24;
  int _totalQuantity = 24;
  String _selectedPackagingType = 'KARTON';
  final TextEditingController _notesController = TextEditingController();
  final TextEditingController _cartonCountController =
      TextEditingController(text: '1');
  final TextEditingController _piecesPerCartonController =
      TextEditingController(text: '24');

  final bool _showQR = false;
  final String _requestId = '';
  String? _employeeId;

  @override
  void initState() {
    super.initState();
    _initData();
    _cartonCountController.addListener(_onCartonCountChanged);
    _piecesPerCartonController.addListener(_onPiecesPerCartonChanged);
  }

  Future<void> _initData() async {
    await _fetchProfile();
    await _fetchProducts();
  }

  @override
  void dispose() {
    _cartonCountController.dispose();
    _piecesPerCartonController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  void _onCartonCountChanged() {
    final val = int.tryParse(_cartonCountController.text) ?? 1;
    setState(() {
      _cartonCount = val;
      _syncTotalQuantity();
    });
  }

  void _onPiecesPerCartonChanged() {
    final val = int.tryParse(_piecesPerCartonController.text) ?? 24;
    setState(() {
      _piecesPerCarton = val;
      _syncTotalQuantity();
    });
  }

  void _syncTotalQuantity() {
    _totalQuantity = (_cartonCount ?? 1) * (_piecesPerCarton ?? 24);
  }

  Future<void> _fetchProfile() async {
    try {
      final response = await apiClient.client.get('employees/profile');
      if (mounted) {
        setState(() {
          _employeeId = response.data['data']['id'];
        });
      }
    } catch (e) {
      debugPrint('Error fetching profile for stock request: $e');
    }
  }

  Future<void> _fetchProducts() async {
    if (widget.isReturn && _employeeId == null) {
      await _fetchProfile();
      if (_employeeId == null) {
        _showError('Gagal mendapatkan identitas karyawan.');
        return;
      }
    }
    setState(() => _isLoading = true);
    try {
      // Fetch available products from warehouse stock or sales carried stock
      final response = widget.isReturn
          ? await apiClient.client.get('sales-transfers/stock/$_employeeId')
          : await apiClient.client.get('warehouse/stock');

      if (!widget.isReturn) {
        try {
          final bonusResponse =
              await apiClient.client.get('bonus-inventory/branch/ALL_BRANCHES');
          if (mounted) {
            if (bonusResponse.data is Map &&
                bonusResponse.data.containsKey('data')) {
              _bonusProducts =
                  (bonusResponse.data['data'] as List<dynamic>?) ?? [];
            } else {
              _bonusProducts = (bonusResponse.data as List<dynamic>?) ?? [];
            }
          }
        } catch (e) {
          debugPrint('Error fetching bonus inventory: $e');
        }
      }

      setState(() {
        if (response.data is Map && response.data.containsKey('data')) {
          _products = (response.data['data'] as List<dynamic>?) ?? [];
        } else {
          _products = (response.data as List<dynamic>?) ?? [];
        }
      });
    } catch (e) {
      _showError('Gagal mengambil daftar produk: $e');
    } finally {
      setState(() => _isLoading = false);
    }
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(msg), backgroundColor: Colors.red));
  }

  Future<void> _submitRequest() async {
    if (_employeeId == null) {
      _showError('Gagal mendapatkan identitas karyawan. Coba lagi nanti.');
      return;
    }

    // Auto-add currently selected product if it hasn't been added to the list yet
    if (_selectedProductId != null) {
      final currentList = _isBonusMode ? _bonusProducts : _products;
      final selectedItem = currentList.firstWhere(
        (item) {
          if (_isBonusMode) return item['id'] == _selectedProductId;
          return item['product']?['id'] == _selectedProductId;
        },
        orElse: () => null,
      );

      if (selectedItem != null) {
        final productName = _isBonusMode
            ? (selectedItem['name'] ?? 'Bonus')
            : (selectedItem['product']?['name'] ?? 'Produk');

        final alreadyAdded =
            _requestItems.any((item) => item.productId == _selectedProductId);
        if (!alreadyAdded) {
          if (widget.isReturn &&
              !_isBonusMode &&
              _totalQuantity > (selectedItem['quantity'] ?? 0)) {
            _showError(
                'Jumlah retur untuk $productName melebihi sisa stok Anda (${selectedItem['quantity']} PCS)');
            return;
          }
          _requestItems.add(RequestItem(
            productId: _selectedProductId!,
            productName: productName,
            cartonCount: _cartonCount,
            piecesPerCarton: _piecesPerCarton,
            totalQuantity: _totalQuantity,
            packagingType: _selectedPackagingType,
            isBonus: _isBonusMode,
          ));
        }
      }
    }

    if (_requestItems.isEmpty) {
      _showError('Tambahkan minimal satu produk ke daftar terlebih dahulu');
      return;
    }

    setState(() => _isLoading = true);
    try {
      final userNotes = _notesController.text.trim();

      // Submit each item in the list sequentially
      for (final item in _requestItems) {
        if (widget.isReturn && !item.isBonus) {
          final selectedItem = _products.firstWhere(
            (p) => p['product']?['id'] == item.productId,
            orElse: () => null,
          );
          if (selectedItem != null &&
              item.totalQuantity > (selectedItem['quantity'] ?? 0)) {
            _showError(
                'Jumlah retur untuk ${item.productName} melebihi sisa stok Anda (${selectedItem['quantity']} PCS)');
            return;
          }
        }

        final packagingLabel =
            item.packagingType == 'KARTON' ? 'Karton' : 'Krat';
        final actionWord = widget.isReturn ? 'Retur' : 'Ambil';
        final cartonNotes =
            '$actionWord: ${item.cartonCount} $packagingLabel @ ${item.piecesPerCarton} isi (Total: ${item.totalQuantity} PCS)';
        final finalNotes = userNotes.isEmpty
            ? cartonNotes
            : '$cartonNotes\nCatatan: $userNotes';

        if (item.isBonus) {
          await apiClient.client.post('sales-bonus-requests', data: {
            'employee_id': _employeeId,
            'bonus_inventory_id': item.productId,
            'quantity': item.totalQuantity,
            'notes': finalNotes,
          });
        } else {
          await apiClient.client.post('sales-transfers/request', data: {
            'employee_id': _employeeId,
            'product_id': item.productId,
            'quantity': item.totalQuantity,
            'unit': 'PCS',
            'type': widget.isReturn ? 'RETURN' : 'TRANSFER',
            'notes': finalNotes,
          });
        }
      }

      // Show success snackbar
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(widget.isReturn
              ? '✓ Permintaan retur ${_requestItems.length} produk berhasil dikirim!'
              : '✓ Permintaan ${_requestItems.length} produk berhasil dikirim!'),
          backgroundColor: Colors.green,
        ),
      );

      setState(() {
        _requestItems.clear();
        _selectedProductId = null;
      });

      // Pop back to dashboard after a delay
      Future.delayed(const Duration(seconds: 2), () {
        if (mounted) Navigator.pop(context);
      });
    } on DioException catch (e) {
      final msg = e.response?.data?['message'] ?? e.message ?? e.toString();
      _showError('Gagal mengirim permintaan: $msg');
    } catch (e) {
      _showError('Gagal mengirim permintaan: $e');
    } finally {
      setState(() => _isLoading = false);
    }
  }

  void _addAllRemainingStock() {
    if (!widget.isReturn) return;

    setState(() {
      _requestItems.clear(); // Clear existing to avoid duplicates

      for (final item in _products) {
        final product = item['product'];
        if (product == null) continue;

        final quantity = item['quantity'] ?? 0;
        if (quantity <= 0) continue; // Only return items with remaining stock

        final id = product['id'];
        final name = product['name'] ?? 'Produk';
        final pUnit = product['pcs_per_unit'] ?? 24;

        final carton = quantity ~/ pUnit;

        _requestItems.add(RequestItem(
          productId: id,
          productName: name,
          cartonCount: carton,
          piecesPerCarton: pUnit,
          totalQuantity: quantity,
          packagingType: 'KARTON',
        ));
      }
      _selectedProductId = null; // Clear selection
    });

    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content:
            Text('✓ Berhasil menambahkan semua sisa barang ke daftar retur'),
        backgroundColor: Colors.green,
        duration: Duration(seconds: 2),
      ),
    );
  }

  Widget _buildRequestItemsList() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              'Daftar Permintaan (${_requestItems.length})',
              style: GoogleFonts.outfit(
                fontWeight: FontWeight.w800,
                color: const Color(0xFF1E293B),
                fontSize: 14,
              ),
            ),
            if (_requestItems.isNotEmpty)
              TextButton.icon(
                onPressed: () {
                  setState(() {
                    _requestItems.clear();
                  });
                },
                icon: const Icon(Icons.clear_all_rounded,
                    size: 16, color: Colors.redAccent),
                label: Text(
                  'Hapus Semua',
                  style: GoogleFonts.outfit(
                    color: Colors.redAccent,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
          ],
        ),
        const SizedBox(height: 8),
        ListView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: _requestItems.length,
          itemBuilder: (context, index) {
            final item = _requestItems[index];
            return Container(
              margin: const EdgeInsets.only(bottom: 8),
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: const Color(0xFFE2E8F0)),
              ),
              child: Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: const Color(0xFF3B82F6).withOpacity(0.1),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      Icons.inventory_2_rounded,
                      color: Color(0xFF3B82F6),
                      size: 16,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          item.productName,
                          style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w800,
                            color: const Color(0xFF1E293B),
                            fontSize: 13,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          '${item.cartonCount} ${item.packagingType == 'KARTON' ? 'Karton' : 'Krat'} @ ${item.piecesPerCarton} PCS',
                          style: GoogleFonts.outfit(
                            color: Colors.blueGrey,
                            fontSize: 11,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Text(
                    '${item.totalQuantity} PCS',
                    style: GoogleFonts.outfit(
                      fontWeight: FontWeight.w900,
                      color: const Color(0xFF10B981),
                      fontSize: 14,
                    ),
                  ),
                  const SizedBox(width: 4),
                  IconButton(
                    icon: const Icon(
                      Icons.delete_outline_rounded,
                      color: Colors.redAccent,
                      size: 20,
                    ),
                    onPressed: () {
                      setState(() {
                        _requestItems.removeAt(index);
                      });
                    },
                  ),
                ],
              ),
            );
          },
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    final bool hasItems =
        _requestItems.isNotEmpty || _selectedProductId != null;

    return Scaffold(
      backgroundColor: const Color(0xFFF8FAFC),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: Text(
            widget.isReturn ? 'RETUR SISA BARANG' : 'AMBIL BARANG GUDANG',
            style: GoogleFonts.outfit(
                fontWeight: FontWeight.w900,
                color: const Color(0xFF1E293B),
                fontSize: 16)),
        leading: IconButton(
            icon: const Icon(Icons.arrow_back_ios_new_rounded,
                color: Color(0xFF1E293B)),
            onPressed: () => Navigator.pop(context)),
      ),
      body: _isLoading && _products.isEmpty
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildHeader(),
                  const SizedBox(height: 24),

                  if (widget.isReturn) ...[
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton.icon(
                        onPressed:
                            _products.isEmpty ? null : _addAllRemainingStock,
                        icon: const Icon(
                            Icons.playlist_add_check_circle_rounded,
                            color: Colors.white),
                        label: Text(
                          'RETUR SEMUA SISA BARANG',
                          style: GoogleFonts.outfit(
                              fontWeight: FontWeight.w800,
                              color: Colors.white,
                              letterSpacing: 1),
                        ),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.amber.shade800,
                          padding: const EdgeInsets.symmetric(vertical: 16),
                          shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16)),
                          elevation: 0,
                        ),
                      ),
                    ),
                    const SizedBox(height: 24),
                  ],

                  // Daftar Permintaan (Cart)
                  if (_requestItems.isNotEmpty) ...[
                    _buildRequestItemsList(),
                    const SizedBox(height: 24),
                  ],

                  // Section: Pilih / Tambah Produk
                  Text(
                    _requestItems.isEmpty
                        ? 'Pilih Produk'
                        : 'Tambah Produk Lain',
                    style: GoogleFonts.outfit(
                        fontWeight: FontWeight.bold, color: Colors.blueGrey),
                  ),
                  const SizedBox(height: 12),

                  // Toggle Mode
                  if (!widget.isReturn && _bonusProducts.isNotEmpty) ...[
                    Row(
                      children: [
                        Expanded(
                          child: GestureDetector(
                            onTap: () => setState(() {
                              _isBonusMode = false;
                              _selectedProductId = null;
                            }),
                            child: Container(
                              padding: const EdgeInsets.symmetric(vertical: 12),
                              decoration: BoxDecoration(
                                color: !_isBonusMode
                                    ? const Color(0xFF1E293B)
                                    : Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border:
                                    Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              alignment: Alignment.center,
                              child: Text('Produk Utama',
                                  style: GoogleFonts.outfit(
                                      color: !_isBonusMode
                                          ? Colors.white
                                          : Colors.blueGrey,
                                      fontWeight: FontWeight.bold)),
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: GestureDetector(
                            onTap: () => setState(() {
                              _isBonusMode = true;
                              _selectedProductId = null;
                            }),
                            child: Container(
                              padding: const EdgeInsets.symmetric(vertical: 12),
                              decoration: BoxDecoration(
                                color: _isBonusMode
                                    ? Colors.amber.shade800
                                    : Colors.white,
                                borderRadius: BorderRadius.circular(12),
                                border:
                                    Border.all(color: const Color(0xFFE2E8F0)),
                              ),
                              alignment: Alignment.center,
                              child: Text('Bonus / Tambahan',
                                  style: GoogleFonts.outfit(
                                      color: _isBonusMode
                                          ? Colors.white
                                          : Colors.blueGrey,
                                      fontWeight: FontWeight.bold)),
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                  ],

                  _buildProductDropdown(),
                  const SizedBox(height: 24),
                  _buildCartonCalculatorCard(),
                  const SizedBox(height: 24),

                  Text('Catatan Tambahan (Opsional)',
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.bold, color: Colors.blueGrey)),
                  const SizedBox(height: 12),
                  TextField(
                    controller: _notesController,
                    maxLines: 3,
                    decoration: InputDecoration(
                      hintText: 'Contoh: Untuk stok tambahan rute sore...',
                      fillColor: Colors.white,
                      filled: true,
                      border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(16),
                          borderSide: BorderSide.none),
                    ),
                  ),
                  const SizedBox(height: 40),
                  SizedBox(
                    width: double.infinity,
                    height: 60,
                    child: ElevatedButton(
                      onPressed:
                          (_isLoading || !hasItems) ? null : _submitRequest,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF1E293B),
                        disabledBackgroundColor: Colors.grey.shade300,
                        shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(16)),
                        elevation: 0,
                      ),
                      child: _isLoading
                          ? const CircularProgressIndicator(color: Colors.white)
                          : Text(
                              _requestItems.isEmpty
                                  ? 'KIRIM PERMINTAAN'
                                  : 'KIRIM PERMINTAAN (${_requestItems.length} PRODUK)',
                              style: GoogleFonts.outfit(
                                fontWeight: FontWeight.w900,
                                color: !hasItems
                                    ? Colors.grey.shade500
                                    : Colors.white,
                                letterSpacing: 1,
                              ),
                            ),
                    ),
                  ),
                ],
              ),
            ),
    );
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: widget.isReturn ? Colors.amber.shade50 : Colors.blue.shade50,
        borderRadius: BorderRadius.circular(24),
        border: Border.all(
            color:
                widget.isReturn ? Colors.amber.shade100 : Colors.blue.shade100),
      ),
      child: Row(
        children: [
          Icon(Icons.info_outline_rounded,
              color: widget.isReturn ? Colors.amber.shade800 : Colors.blue),
          const SizedBox(width: 16),
          Expanded(
            child: Text(
              widget.isReturn
                  ? 'Gunakan menu ini untuk mengembalikan sisa barang jualan yang belum laku ke gudang. Stok gudang akan terupdate otomatis setelah dikonfirmasi warehouse.'
                  : 'Gunakan menu ini untuk mengambil stok dari gudang. Warehouse akan melakukan approval setelah permintaan dikirim.',
              style: GoogleFonts.outfit(
                  fontSize: 13,
                  color: widget.isReturn
                      ? Colors.amber.shade900
                      : Colors.blue.shade900),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildProductDropdown() {
    final currentProducts = _isBonusMode ? _bonusProducts : _products;
    final seenIds = <String>{};

    final filteredProducts = currentProducts.where((item) {
      if (_isBonusMode) {
        final id = item['id']?.toString();
        if (id == null) return false;
        if (_requestItems.any((req) => req.productId == id)) return false;
        return seenIds.add(id);
      } else {
        final product = item['product'];
        if (product == null) return false;
        final id = product['id']?.toString();
        if (id == null) return false;
        if (widget.isReturn && (item['quantity'] ?? 0) <= 0) return false;
        if (_requestItems.any((req) => req.productId == id)) return false;
        return seenIds.add(id);
      }
    }).toList();

    if (_selectedProductId != null &&
        !filteredProducts.any((item) {
          if (_isBonusMode) return item['id']?.toString() == _selectedProductId;
          return item['product']?['id']?.toString() == _selectedProductId;
        })) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted) setState(() => _selectedProductId = null);
      });
    }

    if (filteredProducts.isEmpty && currentProducts.isNotEmpty) {
      return Container(
        width: double.infinity,
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.amber.shade50,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.amber.shade200),
        ),
        child: Row(
          children: [
            Icon(Icons.warning_amber_rounded,
                color: Colors.amber.shade800, size: 20),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                'Semua ${_isBonusMode ? 'bonus/tambahan' : 'produk'} tersedia telah ditambahkan ke dalam daftar permintaan.',
                style: GoogleFonts.outfit(
                    color: Colors.amber.shade900,
                    fontSize: 12,
                    fontWeight: FontWeight.w600),
              ),
            ),
          ],
        ),
      );
    }

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<String>(
          isExpanded: true,
          value: _selectedProductId,
          hint: Text(_isBonusMode
              ? 'Pilih barang bonus / tambahan'
              : (widget.isReturn
                  ? 'Pilih barang sisa jualan'
                  : 'Pilih barang yang ingin diambil')),
          items: filteredProducts.map((item) {
            if (_isBonusMode) {
              final id = item['id'];
              int totalAvail = 0;
              for (var p in currentProducts) {
                if (p['id'] == id) {
                  final int qty = p['quantity'] ?? 0;
                  final int reserved = widget.isReturn ? 0 : (p['reserved_quantity'] ?? 0);
                  totalAvail += (qty - reserved);
                }
              }
              return DropdownMenuItem<String>(
                value: id,
                child: Text('${item['name']} (Sisa Gudang: $totalAvail)'),
              );
            } else {
              final product = item['product'];
              final id = product['id'];
              final stockLabel = widget.isReturn ? 'Stok Bawaan' : 'Sisa Gudang';
              int totalAvail = 0;
              for (var p in currentProducts) {
                if (p['product'] != null && p['product']['id'] == id) {
                  final int qty = p['quantity'] ?? 0;
                  final int reserved = widget.isReturn ? 0 : (p['reserved_quantity'] ?? 0);
                  totalAvail += (qty - reserved);
                }
              }
              return DropdownMenuItem<String>(
                value: id,
                child: Text('${product['name']} ($stockLabel: $totalAvail)'),
              );
            }
          }).toList(),
          onChanged: (val) {
            setState(() {
              _selectedProductId = val;
              final selectedItem = currentProducts.firstWhere(
                (item) {
                  if (_isBonusMode) return item['id'] == val;
                  return item['product']?['id'] == val;
                },
                orElse: () => null,
              );
              if (selectedItem != null) {
                if (_isBonusMode) {
                  _piecesPerCartonController.text = '1';
                } else {
                  final product = selectedItem['product'];
                  final pUnit = product['pcs_per_unit'] ?? 24;
                  _piecesPerCartonController.text = pUnit.toString();
                }
              }
            });
          },
        ),
      ),
    );
  }

  Widget _buildCartonCalculatorCard() {
    if (_selectedProductId == null) {
      return Container(
        width: double.infinity,
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(24),
          border: Border.all(color: const Color(0xFFE2E8F0)),
        ),
        child: Column(
          children: [
            Icon(Icons.layers_clear_rounded,
                color: Colors.blueGrey.shade200, size: 40),
            const SizedBox(height: 12),
            Text(
              'Silakan pilih produk terlebih dahulu untuk mengatur jumlah pengambilan.',
              textAlign: TextAlign.center,
              style: GoogleFonts.outfit(
                color: Colors.blueGrey.shade400,
                fontSize: 12,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
      );
    }

    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withOpacity(0.02),
              blurRadius: 10,
              offset: const Offset(0, 4)),
        ],
        border: Border.all(color: const Color(0xFFE2E8F0)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                    color: const Color(0xFF10B981).withOpacity(0.1),
                    shape: BoxShape.circle),
                child: const Icon(Icons.inventory_2_rounded,
                    color: Color(0xFF10B981), size: 16),
              ),
              const SizedBox(width: 10),
              Text(
                'KALKULATOR KEMASAN (${_selectedPackagingType ?? 'KARTON'})',
                style: GoogleFonts.outfit(
                    fontWeight: FontWeight.w900,
                    color: const Color(0xFF10B981),
                    fontSize: 11,
                    letterSpacing: 1.2),
              ),
            ],
          ),
          const SizedBox(height: 20),
          Row(
            children: [
              Expanded(
                child: GestureDetector(
                  onTap: () {
                    setState(() {
                      _selectedPackagingType = 'KARTON';
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(vertical: 10),
                    decoration: BoxDecoration(
                      color: (_selectedPackagingType ?? 'KARTON') == 'KARTON'
                          ? const Color(0xFF10B981)
                          : const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    alignment: Alignment.center,
                    child: Text(
                      'KARTON',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.bold,
                        fontSize: 12,
                        color: (_selectedPackagingType ?? 'KARTON') == 'KARTON'
                            ? Colors.white
                            : const Color(0xFF64748B),
                      ),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: GestureDetector(
                  onTap: () {
                    setState(() {
                      _selectedPackagingType = 'KRAT';
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(vertical: 10),
                    decoration: BoxDecoration(
                      color: (_selectedPackagingType ?? 'KARTON') == 'KRAT'
                          ? const Color(0xFF10B981)
                          : const Color(0xFFF1F5F9),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    alignment: Alignment.center,
                    child: Text(
                      'KRAT',
                      style: GoogleFonts.outfit(
                        fontWeight: FontWeight.bold,
                        fontSize: 12,
                        color: (_selectedPackagingType ?? 'KARTON') == 'KRAT'
                            ? Colors.white
                            : const Color(0xFF64748B),
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 20),
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Jumlah ${(_selectedPackagingType ?? 'KARTON') == 'KARTON' ? 'Karton' : 'Krat'}',
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                          color: Colors.blueGrey),
                    ),
                    const SizedBox(height: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 4),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          IconButton(
                            icon: const Icon(Icons.remove_rounded,
                                color: Colors.redAccent, size: 18),
                            onPressed: () {
                              final current =
                                  int.tryParse(_cartonCountController.text) ??
                                      1;
                              if (current > 1) {
                                _cartonCountController.text =
                                    (current - 1).toString();
                              }
                            },
                          ),
                          Expanded(
                            child: TextField(
                              controller: _cartonCountController,
                              keyboardType: TextInputType.number,
                              textAlign: TextAlign.center,
                              style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w900, fontSize: 16),
                              decoration: const InputDecoration(
                                isDense: true,
                                contentPadding:
                                    EdgeInsets.symmetric(vertical: 8),
                                border: InputBorder.none,
                              ),
                            ),
                          ),
                          IconButton(
                            icon: const Icon(Icons.add_rounded,
                                color: Colors.green, size: 18),
                            onPressed: () {
                              final current =
                                  int.tryParse(_cartonCountController.text) ??
                                      0;
                              _cartonCountController.text =
                                  (current + 1).toString();
                            },
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Isi per ${(_selectedPackagingType ?? 'KARTON') == 'KARTON' ? 'Karton' : 'Krat'}',
                      style: GoogleFonts.outfit(
                          fontWeight: FontWeight.bold,
                          fontSize: 12,
                          color: Colors.blueGrey),
                    ),
                    const SizedBox(height: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 4),
                      decoration: BoxDecoration(
                        color: const Color(0xFFF8FAFC),
                        borderRadius: BorderRadius.circular(16),
                        border: Border.all(color: const Color(0xFFE2E8F0)),
                      ),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          IconButton(
                            icon: const Icon(Icons.remove_rounded,
                                color: Colors.redAccent, size: 18),
                            onPressed: () {
                              final current = int.tryParse(
                                      _piecesPerCartonController.text) ??
                                  24;
                              if (current > 1) {
                                _piecesPerCartonController.text =
                                    (current - 1).toString();
                              }
                            },
                          ),
                          Expanded(
                            child: TextField(
                              controller: _piecesPerCartonController,
                              keyboardType: TextInputType.number,
                              textAlign: TextAlign.center,
                              style: GoogleFonts.outfit(
                                  fontWeight: FontWeight.w900, fontSize: 16),
                              decoration: const InputDecoration(
                                isDense: true,
                                contentPadding:
                                    EdgeInsets.symmetric(vertical: 8),
                                border: InputBorder.none,
                              ),
                            ),
                          ),
                          IconButton(
                            icon: const Icon(Icons.add_rounded,
                                color: Colors.green, size: 18),
                            onPressed: () {
                              final current = int.tryParse(
                                      _piecesPerCartonController.text) ??
                                  0;
                              _piecesPerCartonController.text =
                                  (current + 1).toString();
                            },
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 20),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                colors: AppColors.accentGradient,
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              borderRadius: BorderRadius.circular(16),
              boxShadow: [
                BoxShadow(
                    color: AppColors.primaryGreenLight.withOpacity(0.2),
                    blurRadius: 10,
                    offset: const Offset(0, 4)),
              ],
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                        widget.isReturn
                            ? 'TOTAL RETUR BARANG'
                            : 'TOTAL AMBIL BARANG',
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w800,
                            color: Colors.white.withOpacity(0.8),
                            fontSize: 9,
                            letterSpacing: 1)),
                    const SizedBox(height: 2),
                    Text('${_totalQuantity ?? 24} PCS',
                        style: GoogleFonts.outfit(
                            fontWeight: FontWeight.w900,
                            color: Colors.white,
                            fontSize: 18)),
                  ],
                ),
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                      color: Colors.white.withOpacity(0.15),
                      shape: BoxShape.circle),
                  child: const Icon(Icons.calculate_rounded,
                      color: Colors.white, size: 22),
                )
              ],
            ),
          ),
          const SizedBox(height: 20),
          SizedBox(
            width: double.infinity,
            height: 52,
            child: ElevatedButton.icon(
              onPressed: () {
                if (_selectedProductId == null) return;

                final currentList = _isBonusMode ? _bonusProducts : _products;
                final selectedItem = currentList.firstWhere(
                  (item) {
                    if (_isBonusMode) return item['id'] == _selectedProductId;
                    return item['product']?['id'] == _selectedProductId;
                  },
                  orElse: () => null,
                );

                if (selectedItem == null) return;

                final productName = _isBonusMode
                    ? (selectedItem['name'] ?? 'Bonus')
                    : (selectedItem['product']?['name'] ?? 'Produk');

                if (widget.isReturn && !_isBonusMode) {
                  final avail = selectedItem['quantity'] ?? 0;
                  if (_totalQuantity > avail) {
                    _showError(
                        'Jumlah retur untuk $productName melebihi sisa stok Anda ($avail PCS)');
                    return;
                  }
                }

                final newItem = RequestItem(
                  productId: _selectedProductId!,
                  productName: productName,
                  cartonCount: _cartonCount,
                  piecesPerCarton: _piecesPerCarton,
                  totalQuantity: _totalQuantity,
                  packagingType: _selectedPackagingType,
                  isBonus: _isBonusMode,
                );

                setState(() {
                  _requestItems.add(newItem);
                  _selectedProductId = null; // Clear selection
                  _cartonCountController.text = '1'; // Reset count
                });

                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content:
                        Text('✓ Berhasil menambahkan $productName ke daftar'),
                    backgroundColor: Colors.blueGrey,
                    duration: const Duration(seconds: 1),
                  ),
                );
              },
              icon: const Icon(Icons.add_circle_outline_rounded,
                  color: Colors.white),
              label: Text(
                widget.isReturn
                    ? 'TAMBAH KE DAFTAR RETUR'
                    : 'TAMBAH KE DAFTAR PERMINTAAN',
                style: GoogleFonts.outfit(
                  fontWeight: FontWeight.w900,
                  color: Colors.white,
                  letterSpacing: 0.5,
                  fontSize: 12,
                ),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF1E293B),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(16)),
                elevation: 0,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
