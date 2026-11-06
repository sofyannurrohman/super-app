import 'package:drift/drift.dart';
import './connection/connection_stub.dart'
    if (dart.library.io) './connection/native_connection.dart'
    if (dart.library.html) './connection/web_connection.dart'
    if (dart.library.js_interop) './connection/web_connection.dart';

part 'database.g.dart';

class Products extends Table {
  TextColumn get id => text()();
  TextColumn get companyId => text()();
  TextColumn get name => text()();
  TextColumn get sku => text().nullable()();
  RealColumn get sellingPrice => real()();
  RealColumn get dbp => real().withDefault(const Constant(0.0))();
  RealColumn get wbp => real().withDefault(const Constant(0.0))();
  RealColumn get rbp => real().withDefault(const Constant(0.0))();
  RealColumn get costPrice => real().withDefault(const Constant(0.0))();
  TextColumn get unit => text().nullable()();
  TextColumn get category => text().nullable()();
  IntColumn get warehouseStock => integer().withDefault(const Constant(0))();
  IntColumn get pcsPerUnit => integer().withDefault(const Constant(1))();
  TextColumn get imageUrl => text().nullable()();

  @override
  Set<Column> get primaryKey => {id};
}

class Companies extends Table {
  TextColumn get id => text()();
  TextColumn get name => text()();
  TextColumn get code => text().nullable()();

  @override
  Set<Column> get primaryKey => {id};
}

class LocalCheckins extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get storeId => text()();
  TextColumn get storeName => text()();
  RealColumn get latitude => real()();
  RealColumn get longitude => real()();
  TextColumn get selfiePath => text()();
  DateTimeColumn get createdAt => dateTime()();
  TextColumn get syncStatus => text().withDefault(const Constant('pending'))();
}

class LocalTransactions extends Table {
  TextColumn get localId => text()(); // UUID v4
  TextColumn get receiptNo =>
      text().nullable()(); // Official invoice ID from backend
  TextColumn get companyId => text()();
  TextColumn get companyName => text()();
  TextColumn get storeId => text()();
  TextColumn get storeName => text()();
  RealColumn get totalAmount => real()();
  RealColumn get paidAmount => real().withDefault(const Constant(0.0))();
  TextColumn get selfiePath => text().nullable()();
  TextColumn get receiptPath => text().nullable()();
  TextColumn get signaturePath => text().nullable()();
  TextColumn get notes => text().nullable()();
  TextColumn get paymentMethod => text()
      .withDefault(const Constant('CASH'))(); // 'CASH', 'QRIS', 'TEMPO', 'VA'
  TextColumn get dpPaymentMethod => text().nullable()();
  TextColumn get paymentBank => text().nullable()(); // 'bca', 'bni', 'bri'
  TextColumn get midtransId => text().nullable()();
  TextColumn get midtransQrisUrl => text().nullable()();
  TextColumn get midtransVaNumber => text().nullable()();
  TextColumn get midtransBank => text().nullable()();
  TextColumn get midtransBillKey => text().nullable()();
  TextColumn get midtransBillerCode => text().nullable()();
  DateTimeColumn get createdAt => dateTime()();
  TextColumn get syncStatus => text().withDefault(const Constant('pending'))();
  DateTimeColumn get paymentDueDate => dateTime().nullable()();
  IntColumn get tenor => integer().withDefault(const Constant(1))();

  /// MT flag: true jika transaksi ini dibuat oleh Sales MT.
  /// Faktur dan omzetnya dikelola oleh pusat/pabrik.
  BoolColumn get isMt => boolean().withDefault(const Constant(false))();

  /// Distributor flag: true jika transaksi ini dibuat khusus untuk alur sales distributor
  BoolColumn get isDistributorOrder =>
      boolean().withDefault(const Constant(false))();

  /// UUID company pabrik/pusat yang mengelola omzet MT ini.
  TextColumn get factoryCompanyId => text().nullable()();
  TextColumn get companyBankAccountId => text().nullable()();
  TextColumn get visitId => text().nullable()();
  RealColumn get latitude => real().nullable()();
  RealColumn get longitude => real().nullable()();
  TextColumn get syncErrorMessage => text().nullable()();
  DateTimeColumn get lastSyncAttempt => dateTime().nullable()();

  @override
  Set<Column> get primaryKey => {localId};
}

class LocalTransactionItems extends Table {
  IntColumn get id => integer().autoIncrement()();
  TextColumn get transactionLocalId =>
      text().references(LocalTransactions, #localId)();
  TextColumn get productId => text()();
  TextColumn get productName => text()();
  IntColumn get quantity => integer()(); // Total pcs (base unit)
  IntColumn get orderedQuantity => integer().withDefault(const Constant(0))();
  TextColumn get unit => text().withDefault(const Constant('PCS'))();
  IntColumn get piecesPerUnit => integer().withDefault(const Constant(1))();
  RealColumn get price => real()();

  /// Diskon per item dalam persen (0–100). Default 0 = tidak ada diskon.
  RealColumn get discountPercent => real().withDefault(const Constant(0.0))();
}

class SalesStock extends Table {
  TextColumn get productId => text()();
  IntColumn get quantity => integer()();
  DateTimeColumn get updatedAt => dateTime()();

  @override
  Set<Column> get primaryKey => {productId};
}

class OfflineTasks extends Table {
  TextColumn get id => text()(); // UUID as idempotency key & local ID
  TextColumn get type => text()(); // 'ATTENDANCE', 'VISIT_CHECKIN', 'ORDER', dsb
  TextColumn get payload => text()(); // JSON payload
  TextColumn get filePath => text().nullable()(); // Path to image/file (optional)
  TextColumn get filePath2 => text().nullable()(); // Path to second file (optional)
  DateTimeColumn get createdAt => dateTime()();
  IntColumn get retryCount => integer().withDefault(const Constant(0))();
  TextColumn get status => text().withDefault(const Constant('pending'))(); // pending, processing, failed

  @override
  Set<Column> get primaryKey => {id};
}

class Stores extends Table {
  TextColumn get id => text()();
  TextColumn get code => text().nullable()();
  TextColumn get name => text()();
  TextColumn get address => text().nullable()();
  RealColumn get latitude => real().nullable()();
  RealColumn get longitude => real().nullable()();
  TextColumn get ownerName => text().nullable()();
  TextColumn get phoneNumber => text().nullable()();
  TextColumn get photoUrl => text().nullable()();
  TextColumn get localPhotoPath => text().nullable()();
  TextColumn get province => text().nullable()();
  TextColumn get regency => text().nullable()();
  TextColumn get district => text().nullable()();
  TextColumn get village => text().nullable()();
  TextColumn get visitFrequency => text().nullable()();
  TextColumn get visitDays => text().nullable()();
  TextColumn get assignedEmployeeId => text().nullable()();
  TextColumn get companyId => text().nullable()();
  BoolColumn get isOffline => boolean().withDefault(const Constant(false))();
  TextColumn get syncStatus => text().withDefault(const Constant('synced'))(); // 'pending', 'syncing', 'synced', 'failed'
  TextColumn get syncErrorMessage => text().nullable()();
  DateTimeColumn get createdAt => dateTime().nullable()();
  RealColumn get accuracy => real().nullable()();

  @override
  Set<Column> get primaryKey => {id};
}

class VisitPlans extends Table {
  TextColumn get id => text()();
  TextColumn get storeId => text()();
  TextColumn get storeName => text()();
  TextColumn get date => text()(); // yyyy-MM-dd
  TextColumn get status => text().withDefault(const Constant('pending'))();
  
  @override
  Set<Column> get primaryKey => {id};
}

class SalesDailyJourneys extends Table {
  TextColumn get id => text()();
  TextColumn get userId => text()();
  TextColumn get date => text()(); // yyyy-MM-dd
  IntColumn get startKm => integer()();
  TextColumn get startPhotoUrl => text()();
  TextColumn get startLocation => text()();
  DateTimeColumn get startTime => dateTime()();
  IntColumn get endKm => integer().nullable()();
  TextColumn get endPhotoUrl => text().nullable()();
  TextColumn get endLocation => text().nullable()();
  DateTimeColumn get endTime => dateTime().nullable()();
  IntColumn get totalKm => integer().nullable()();
  TextColumn get status => text().withDefault(const Constant('started'))();
  BoolColumn get isSynced => boolean().withDefault(const Constant(false))();

  @override
  Set<Column> get primaryKey => {id};
}

@DriftDatabase(tables: [
  Products,
  Companies,
  LocalCheckins,
  LocalTransactions,
  LocalTransactionItems,
  SalesStock,
  OfflineTasks,
  Stores,
  VisitPlans,
  SalesDailyJourneys
])
class AppDatabase extends _$AppDatabase {
  AppDatabase() : super(constructDb());
  AppDatabase.forTesting(super.e);

  @override
  int get schemaVersion => 22;

  @override
  MigrationStrategy get migration => MigrationStrategy(
        onCreate: (m) async {
          await m.createAll();
        },
        onUpgrade: (m, from, to) async {
          Future<void> addColumnSafe(dynamic table, dynamic column) async {
            try {
              await m.addColumn(table, column);
            } catch (e) {
              final msg = e.toString().toLowerCase();
              if (msg.contains('duplicate column name') ||
                  msg.contains('already exists')) {
                // ignore
              } else {
                rethrow;
              }
            }
          }

          if (from < 2) {
            // Add Midtrans columns
            await addColumnSafe(
                localTransactions, localTransactions.midtransId);
            await addColumnSafe(
                localTransactions, localTransactions.midtransQrisUrl);
            await addColumnSafe(
                localTransactions, localTransactions.midtransVaNumber);
            await addColumnSafe(
                localTransactions, localTransactions.midtransBank);
            await addColumnSafe(
                localTransactions, localTransactions.midtransBillKey);
            await addColumnSafe(
                localTransactions, localTransactions.midtransBillerCode);
          }
          if (from < 3) {
            await addColumnSafe(products, products.warehouseStock);
            await addColumnSafe(products, products.imageUrl);
          }
          if (from < 4) {
            await addColumnSafe(products, products.pcsPerUnit);
            await addColumnSafe(
                localTransactionItems, localTransactionItems.orderedQuantity);
            await addColumnSafe(
                localTransactionItems, localTransactionItems.unit);
            await addColumnSafe(
                localTransactionItems, localTransactionItems.piecesPerUnit);
          }
          if (from < 5) {
            await addColumnSafe(
                localTransactions, localTransactions.signaturePath);
          }
          if (from < 6) {
            // Add price column to transaction items for PDF rendering
            await addColumnSafe(
                localTransactionItems, localTransactionItems.price);
          }
          if (from < 7) {
            await addColumnSafe(
                localTransactions, localTransactions.paidAmount);
            await addColumnSafe(
                localTransactions, localTransactions.dpPaymentMethod);
          }
          if (from < 8) {
            await addColumnSafe(
                localTransactions, localTransactions.paymentDueDate);
            await addColumnSafe(localTransactions, localTransactions.tenor);
          }
          if (from < 9) {
            // MT flag columns for Sales Modern Trade
            await addColumnSafe(localTransactions, localTransactions.isMt);
            await addColumnSafe(
                localTransactions, localTransactions.factoryCompanyId);
          }
          if (from < 10) {
            await addColumnSafe(
                localTransactions, localTransactions.companyBankAccountId);
          }
          if (from < 11) {
            // Diskon per item (persen) untuk fitur diskon dinamis sales
            await addColumnSafe(
                localTransactionItems, localTransactionItems.discountPercent);
          }
          if (from < 12) {
            await addColumnSafe(localTransactions, localTransactions.visitId);
          }
          if (from < 13) {
            // Fix for users who were on v12 but missed companyBankAccountId due to the typo in previous migration
            await addColumnSafe(
                localTransactions, localTransactions.companyBankAccountId);
          }
          if (from < 14) {
            await m.createTable(offlineTasks);
          }
          if (from < 15) {
            await m.createTable(stores);
            await m.createTable(visitPlans);
          }
          if (from < 16) {
            // Fix for users who were on v13/v14 but missed companyBankAccountId due to a previous migration timing issue
            await addColumnSafe(
                localTransactions, localTransactions.companyBankAccountId);
          }
          if (from < 17) {
            await addColumnSafe(localTransactions, localTransactions.latitude);
            await addColumnSafe(localTransactions, localTransactions.longitude);
          }
          if (from < 18) {
            await m.createTable(salesDailyJourneys);
          }
          if (from < 19) {
            await addColumnSafe(
                localTransactions, localTransactions.syncErrorMessage);
            await addColumnSafe(
                localTransactions, localTransactions.lastSyncAttempt);
          }
          if (from < 20) {
            await addColumnSafe(stores, stores.ownerName);
            await addColumnSafe(stores, stores.phoneNumber);
            await addColumnSafe(stores, stores.photoUrl);
            await addColumnSafe(stores, stores.localPhotoPath);
            await addColumnSafe(stores, stores.province);
            await addColumnSafe(stores, stores.regency);
            await addColumnSafe(stores, stores.district);
            await addColumnSafe(stores, stores.village);
            await addColumnSafe(stores, stores.visitFrequency);
            await addColumnSafe(stores, stores.visitDays);
            await addColumnSafe(stores, stores.assignedEmployeeId);
            await addColumnSafe(stores, stores.companyId);
            await addColumnSafe(stores, stores.isOffline);
            await addColumnSafe(stores, stores.syncStatus);
            await addColumnSafe(stores, stores.syncErrorMessage);
            await addColumnSafe(stores, stores.createdAt);
            await addColumnSafe(stores, stores.accuracy);
          }
          if (from < 21) {
            await addColumnSafe(
                localTransactions, localTransactions.isDistributorOrder);
          }
          if (from < 22) {
            await addColumnSafe(products, products.dbp);
            await addColumnSafe(products, products.wbp);
            await addColumnSafe(products, products.rbp);
            await addColumnSafe(products, products.costPrice);
          }
        },
      );

  Future<void> clearAllData() async {
    await transaction(() async {
      // Delete in reverse dependency order to avoid foreign key constraint errors
      await delete(salesDailyJourneys).go();
      await delete(visitPlans).go();
      await delete(stores).go();
      await delete(offlineTasks).go();
      await delete(localTransactionItems).go();
      await delete(localTransactions).go();
      await delete(localCheckins).go();
      await delete(salesStock).go();
      await delete(products).go();
      await delete(companies).go();
    });
  }
}
