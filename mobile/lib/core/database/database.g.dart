// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'database.dart';

// ignore_for_file: type=lint
class $ProductsTable extends Products with TableInfo<$ProductsTable, Product> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $ProductsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<String> id = GeneratedColumn<String>(
      'id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _companyIdMeta =
      const VerificationMeta('companyId');
  @override
  late final GeneratedColumn<String> companyId = GeneratedColumn<String>(
      'company_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _nameMeta = const VerificationMeta('name');
  @override
  late final GeneratedColumn<String> name = GeneratedColumn<String>(
      'name', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _skuMeta = const VerificationMeta('sku');
  @override
  late final GeneratedColumn<String> sku = GeneratedColumn<String>(
      'sku', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _sellingPriceMeta =
      const VerificationMeta('sellingPrice');
  @override
  late final GeneratedColumn<double> sellingPrice = GeneratedColumn<double>(
      'selling_price', aliasedName, false,
      type: DriftSqlType.double, requiredDuringInsert: true);
  static const VerificationMeta _dbpMeta = const VerificationMeta('dbp');
  @override
  late final GeneratedColumn<double> dbp = GeneratedColumn<double>(
      'dbp', aliasedName, false,
      type: DriftSqlType.double,
      requiredDuringInsert: false,
      defaultValue: const Constant(0.0));
  static const VerificationMeta _wbpMeta = const VerificationMeta('wbp');
  @override
  late final GeneratedColumn<double> wbp = GeneratedColumn<double>(
      'wbp', aliasedName, false,
      type: DriftSqlType.double,
      requiredDuringInsert: false,
      defaultValue: const Constant(0.0));
  static const VerificationMeta _rbpMeta = const VerificationMeta('rbp');
  @override
  late final GeneratedColumn<double> rbp = GeneratedColumn<double>(
      'rbp', aliasedName, false,
      type: DriftSqlType.double,
      requiredDuringInsert: false,
      defaultValue: const Constant(0.0));
  static const VerificationMeta _costPriceMeta =
      const VerificationMeta('costPrice');
  @override
  late final GeneratedColumn<double> costPrice = GeneratedColumn<double>(
      'cost_price', aliasedName, false,
      type: DriftSqlType.double,
      requiredDuringInsert: false,
      defaultValue: const Constant(0.0));
  static const VerificationMeta _unitMeta = const VerificationMeta('unit');
  @override
  late final GeneratedColumn<String> unit = GeneratedColumn<String>(
      'unit', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _categoryMeta =
      const VerificationMeta('category');
  @override
  late final GeneratedColumn<String> category = GeneratedColumn<String>(
      'category', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _warehouseStockMeta =
      const VerificationMeta('warehouseStock');
  @override
  late final GeneratedColumn<int> warehouseStock = GeneratedColumn<int>(
      'warehouse_stock', aliasedName, false,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultValue: const Constant(0));
  static const VerificationMeta _pcsPerUnitMeta =
      const VerificationMeta('pcsPerUnit');
  @override
  late final GeneratedColumn<int> pcsPerUnit = GeneratedColumn<int>(
      'pcs_per_unit', aliasedName, false,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultValue: const Constant(1));
  static const VerificationMeta _imageUrlMeta =
      const VerificationMeta('imageUrl');
  @override
  late final GeneratedColumn<String> imageUrl = GeneratedColumn<String>(
      'image_url', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  @override
  List<GeneratedColumn> get $columns => [
        id,
        companyId,
        name,
        sku,
        sellingPrice,
        dbp,
        wbp,
        rbp,
        costPrice,
        unit,
        category,
        warehouseStock,
        pcsPerUnit,
        imageUrl
      ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'products';
  @override
  VerificationContext validateIntegrity(Insertable<Product> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    } else if (isInserting) {
      context.missing(_idMeta);
    }
    if (data.containsKey('company_id')) {
      context.handle(_companyIdMeta,
          companyId.isAcceptableOrUnknown(data['company_id']!, _companyIdMeta));
    } else if (isInserting) {
      context.missing(_companyIdMeta);
    }
    if (data.containsKey('name')) {
      context.handle(
          _nameMeta, name.isAcceptableOrUnknown(data['name']!, _nameMeta));
    } else if (isInserting) {
      context.missing(_nameMeta);
    }
    if (data.containsKey('sku')) {
      context.handle(
          _skuMeta, sku.isAcceptableOrUnknown(data['sku']!, _skuMeta));
    }
    if (data.containsKey('selling_price')) {
      context.handle(
          _sellingPriceMeta,
          sellingPrice.isAcceptableOrUnknown(
              data['selling_price']!, _sellingPriceMeta));
    } else if (isInserting) {
      context.missing(_sellingPriceMeta);
    }
    if (data.containsKey('dbp')) {
      context.handle(
          _dbpMeta, dbp.isAcceptableOrUnknown(data['dbp']!, _dbpMeta));
    }
    if (data.containsKey('wbp')) {
      context.handle(
          _wbpMeta, wbp.isAcceptableOrUnknown(data['wbp']!, _wbpMeta));
    }
    if (data.containsKey('rbp')) {
      context.handle(
          _rbpMeta, rbp.isAcceptableOrUnknown(data['rbp']!, _rbpMeta));
    }
    if (data.containsKey('cost_price')) {
      context.handle(_costPriceMeta,
          costPrice.isAcceptableOrUnknown(data['cost_price']!, _costPriceMeta));
    }
    if (data.containsKey('unit')) {
      context.handle(
          _unitMeta, unit.isAcceptableOrUnknown(data['unit']!, _unitMeta));
    }
    if (data.containsKey('category')) {
      context.handle(_categoryMeta,
          category.isAcceptableOrUnknown(data['category']!, _categoryMeta));
    }
    if (data.containsKey('warehouse_stock')) {
      context.handle(
          _warehouseStockMeta,
          warehouseStock.isAcceptableOrUnknown(
              data['warehouse_stock']!, _warehouseStockMeta));
    }
    if (data.containsKey('pcs_per_unit')) {
      context.handle(
          _pcsPerUnitMeta,
          pcsPerUnit.isAcceptableOrUnknown(
              data['pcs_per_unit']!, _pcsPerUnitMeta));
    }
    if (data.containsKey('image_url')) {
      context.handle(_imageUrlMeta,
          imageUrl.isAcceptableOrUnknown(data['image_url']!, _imageUrlMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  Product map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return Product(
      id: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}id'])!,
      companyId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}company_id'])!,
      name: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}name'])!,
      sku: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}sku']),
      sellingPrice: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}selling_price'])!,
      dbp: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}dbp'])!,
      wbp: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}wbp'])!,
      rbp: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}rbp'])!,
      costPrice: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}cost_price'])!,
      unit: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}unit']),
      category: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}category']),
      warehouseStock: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}warehouse_stock'])!,
      pcsPerUnit: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}pcs_per_unit'])!,
      imageUrl: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}image_url']),
    );
  }

  @override
  $ProductsTable createAlias(String alias) {
    return $ProductsTable(attachedDatabase, alias);
  }
}

class Product extends DataClass implements Insertable<Product> {
  final String id;
  final String companyId;
  final String name;
  final String? sku;
  final double sellingPrice;
  final double dbp;
  final double wbp;
  final double rbp;
  final double costPrice;
  final String? unit;
  final String? category;
  final int warehouseStock;
  final int pcsPerUnit;
  final String? imageUrl;
  const Product(
      {required this.id,
      required this.companyId,
      required this.name,
      this.sku,
      required this.sellingPrice,
      required this.dbp,
      required this.wbp,
      required this.rbp,
      required this.costPrice,
      this.unit,
      this.category,
      required this.warehouseStock,
      required this.pcsPerUnit,
      this.imageUrl});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<String>(id);
    map['company_id'] = Variable<String>(companyId);
    map['name'] = Variable<String>(name);
    if (!nullToAbsent || sku != null) {
      map['sku'] = Variable<String>(sku);
    }
    map['selling_price'] = Variable<double>(sellingPrice);
    map['dbp'] = Variable<double>(dbp);
    map['wbp'] = Variable<double>(wbp);
    map['rbp'] = Variable<double>(rbp);
    map['cost_price'] = Variable<double>(costPrice);
    if (!nullToAbsent || unit != null) {
      map['unit'] = Variable<String>(unit);
    }
    if (!nullToAbsent || category != null) {
      map['category'] = Variable<String>(category);
    }
    map['warehouse_stock'] = Variable<int>(warehouseStock);
    map['pcs_per_unit'] = Variable<int>(pcsPerUnit);
    if (!nullToAbsent || imageUrl != null) {
      map['image_url'] = Variable<String>(imageUrl);
    }
    return map;
  }

  ProductsCompanion toCompanion(bool nullToAbsent) {
    return ProductsCompanion(
      id: Value(id),
      companyId: Value(companyId),
      name: Value(name),
      sku: sku == null && nullToAbsent ? const Value.absent() : Value(sku),
      sellingPrice: Value(sellingPrice),
      dbp: Value(dbp),
      wbp: Value(wbp),
      rbp: Value(rbp),
      costPrice: Value(costPrice),
      unit: unit == null && nullToAbsent ? const Value.absent() : Value(unit),
      category: category == null && nullToAbsent
          ? const Value.absent()
          : Value(category),
      warehouseStock: Value(warehouseStock),
      pcsPerUnit: Value(pcsPerUnit),
      imageUrl: imageUrl == null && nullToAbsent
          ? const Value.absent()
          : Value(imageUrl),
    );
  }

  factory Product.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return Product(
      id: serializer.fromJson<String>(json['id']),
      companyId: serializer.fromJson<String>(json['companyId']),
      name: serializer.fromJson<String>(json['name']),
      sku: serializer.fromJson<String?>(json['sku']),
      sellingPrice: serializer.fromJson<double>(json['sellingPrice']),
      dbp: serializer.fromJson<double>(json['dbp']),
      wbp: serializer.fromJson<double>(json['wbp']),
      rbp: serializer.fromJson<double>(json['rbp']),
      costPrice: serializer.fromJson<double>(json['costPrice']),
      unit: serializer.fromJson<String?>(json['unit']),
      category: serializer.fromJson<String?>(json['category']),
      warehouseStock: serializer.fromJson<int>(json['warehouseStock']),
      pcsPerUnit: serializer.fromJson<int>(json['pcsPerUnit']),
      imageUrl: serializer.fromJson<String?>(json['imageUrl']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<String>(id),
      'companyId': serializer.toJson<String>(companyId),
      'name': serializer.toJson<String>(name),
      'sku': serializer.toJson<String?>(sku),
      'sellingPrice': serializer.toJson<double>(sellingPrice),
      'dbp': serializer.toJson<double>(dbp),
      'wbp': serializer.toJson<double>(wbp),
      'rbp': serializer.toJson<double>(rbp),
      'costPrice': serializer.toJson<double>(costPrice),
      'unit': serializer.toJson<String?>(unit),
      'category': serializer.toJson<String?>(category),
      'warehouseStock': serializer.toJson<int>(warehouseStock),
      'pcsPerUnit': serializer.toJson<int>(pcsPerUnit),
      'imageUrl': serializer.toJson<String?>(imageUrl),
    };
  }

  Product copyWith(
          {String? id,
          String? companyId,
          String? name,
          Value<String?> sku = const Value.absent(),
          double? sellingPrice,
          double? dbp,
          double? wbp,
          double? rbp,
          double? costPrice,
          Value<String?> unit = const Value.absent(),
          Value<String?> category = const Value.absent(),
          int? warehouseStock,
          int? pcsPerUnit,
          Value<String?> imageUrl = const Value.absent()}) =>
      Product(
        id: id ?? this.id,
        companyId: companyId ?? this.companyId,
        name: name ?? this.name,
        sku: sku.present ? sku.value : this.sku,
        sellingPrice: sellingPrice ?? this.sellingPrice,
        dbp: dbp ?? this.dbp,
        wbp: wbp ?? this.wbp,
        rbp: rbp ?? this.rbp,
        costPrice: costPrice ?? this.costPrice,
        unit: unit.present ? unit.value : this.unit,
        category: category.present ? category.value : this.category,
        warehouseStock: warehouseStock ?? this.warehouseStock,
        pcsPerUnit: pcsPerUnit ?? this.pcsPerUnit,
        imageUrl: imageUrl.present ? imageUrl.value : this.imageUrl,
      );
  Product copyWithCompanion(ProductsCompanion data) {
    return Product(
      id: data.id.present ? data.id.value : this.id,
      companyId: data.companyId.present ? data.companyId.value : this.companyId,
      name: data.name.present ? data.name.value : this.name,
      sku: data.sku.present ? data.sku.value : this.sku,
      sellingPrice: data.sellingPrice.present
          ? data.sellingPrice.value
          : this.sellingPrice,
      dbp: data.dbp.present ? data.dbp.value : this.dbp,
      wbp: data.wbp.present ? data.wbp.value : this.wbp,
      rbp: data.rbp.present ? data.rbp.value : this.rbp,
      costPrice: data.costPrice.present ? data.costPrice.value : this.costPrice,
      unit: data.unit.present ? data.unit.value : this.unit,
      category: data.category.present ? data.category.value : this.category,
      warehouseStock: data.warehouseStock.present
          ? data.warehouseStock.value
          : this.warehouseStock,
      pcsPerUnit:
          data.pcsPerUnit.present ? data.pcsPerUnit.value : this.pcsPerUnit,
      imageUrl: data.imageUrl.present ? data.imageUrl.value : this.imageUrl,
    );
  }

  @override
  String toString() {
    return (StringBuffer('Product(')
          ..write('id: $id, ')
          ..write('companyId: $companyId, ')
          ..write('name: $name, ')
          ..write('sku: $sku, ')
          ..write('sellingPrice: $sellingPrice, ')
          ..write('dbp: $dbp, ')
          ..write('wbp: $wbp, ')
          ..write('rbp: $rbp, ')
          ..write('costPrice: $costPrice, ')
          ..write('unit: $unit, ')
          ..write('category: $category, ')
          ..write('warehouseStock: $warehouseStock, ')
          ..write('pcsPerUnit: $pcsPerUnit, ')
          ..write('imageUrl: $imageUrl')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
      id,
      companyId,
      name,
      sku,
      sellingPrice,
      dbp,
      wbp,
      rbp,
      costPrice,
      unit,
      category,
      warehouseStock,
      pcsPerUnit,
      imageUrl);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is Product &&
          other.id == this.id &&
          other.companyId == this.companyId &&
          other.name == this.name &&
          other.sku == this.sku &&
          other.sellingPrice == this.sellingPrice &&
          other.dbp == this.dbp &&
          other.wbp == this.wbp &&
          other.rbp == this.rbp &&
          other.costPrice == this.costPrice &&
          other.unit == this.unit &&
          other.category == this.category &&
          other.warehouseStock == this.warehouseStock &&
          other.pcsPerUnit == this.pcsPerUnit &&
          other.imageUrl == this.imageUrl);
}

class ProductsCompanion extends UpdateCompanion<Product> {
  final Value<String> id;
  final Value<String> companyId;
  final Value<String> name;
  final Value<String?> sku;
  final Value<double> sellingPrice;
  final Value<double> dbp;
  final Value<double> wbp;
  final Value<double> rbp;
  final Value<double> costPrice;
  final Value<String?> unit;
  final Value<String?> category;
  final Value<int> warehouseStock;
  final Value<int> pcsPerUnit;
  final Value<String?> imageUrl;
  final Value<int> rowid;
  const ProductsCompanion({
    this.id = const Value.absent(),
    this.companyId = const Value.absent(),
    this.name = const Value.absent(),
    this.sku = const Value.absent(),
    this.sellingPrice = const Value.absent(),
    this.dbp = const Value.absent(),
    this.wbp = const Value.absent(),
    this.rbp = const Value.absent(),
    this.costPrice = const Value.absent(),
    this.unit = const Value.absent(),
    this.category = const Value.absent(),
    this.warehouseStock = const Value.absent(),
    this.pcsPerUnit = const Value.absent(),
    this.imageUrl = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  ProductsCompanion.insert({
    required String id,
    required String companyId,
    required String name,
    this.sku = const Value.absent(),
    required double sellingPrice,
    this.dbp = const Value.absent(),
    this.wbp = const Value.absent(),
    this.rbp = const Value.absent(),
    this.costPrice = const Value.absent(),
    this.unit = const Value.absent(),
    this.category = const Value.absent(),
    this.warehouseStock = const Value.absent(),
    this.pcsPerUnit = const Value.absent(),
    this.imageUrl = const Value.absent(),
    this.rowid = const Value.absent(),
  })  : id = Value(id),
        companyId = Value(companyId),
        name = Value(name),
        sellingPrice = Value(sellingPrice);
  static Insertable<Product> custom({
    Expression<String>? id,
    Expression<String>? companyId,
    Expression<String>? name,
    Expression<String>? sku,
    Expression<double>? sellingPrice,
    Expression<double>? dbp,
    Expression<double>? wbp,
    Expression<double>? rbp,
    Expression<double>? costPrice,
    Expression<String>? unit,
    Expression<String>? category,
    Expression<int>? warehouseStock,
    Expression<int>? pcsPerUnit,
    Expression<String>? imageUrl,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (companyId != null) 'company_id': companyId,
      if (name != null) 'name': name,
      if (sku != null) 'sku': sku,
      if (sellingPrice != null) 'selling_price': sellingPrice,
      if (dbp != null) 'dbp': dbp,
      if (wbp != null) 'wbp': wbp,
      if (rbp != null) 'rbp': rbp,
      if (costPrice != null) 'cost_price': costPrice,
      if (unit != null) 'unit': unit,
      if (category != null) 'category': category,
      if (warehouseStock != null) 'warehouse_stock': warehouseStock,
      if (pcsPerUnit != null) 'pcs_per_unit': pcsPerUnit,
      if (imageUrl != null) 'image_url': imageUrl,
      if (rowid != null) 'rowid': rowid,
    });
  }

  ProductsCompanion copyWith(
      {Value<String>? id,
      Value<String>? companyId,
      Value<String>? name,
      Value<String?>? sku,
      Value<double>? sellingPrice,
      Value<double>? dbp,
      Value<double>? wbp,
      Value<double>? rbp,
      Value<double>? costPrice,
      Value<String?>? unit,
      Value<String?>? category,
      Value<int>? warehouseStock,
      Value<int>? pcsPerUnit,
      Value<String?>? imageUrl,
      Value<int>? rowid}) {
    return ProductsCompanion(
      id: id ?? this.id,
      companyId: companyId ?? this.companyId,
      name: name ?? this.name,
      sku: sku ?? this.sku,
      sellingPrice: sellingPrice ?? this.sellingPrice,
      dbp: dbp ?? this.dbp,
      wbp: wbp ?? this.wbp,
      rbp: rbp ?? this.rbp,
      costPrice: costPrice ?? this.costPrice,
      unit: unit ?? this.unit,
      category: category ?? this.category,
      warehouseStock: warehouseStock ?? this.warehouseStock,
      pcsPerUnit: pcsPerUnit ?? this.pcsPerUnit,
      imageUrl: imageUrl ?? this.imageUrl,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<String>(id.value);
    }
    if (companyId.present) {
      map['company_id'] = Variable<String>(companyId.value);
    }
    if (name.present) {
      map['name'] = Variable<String>(name.value);
    }
    if (sku.present) {
      map['sku'] = Variable<String>(sku.value);
    }
    if (sellingPrice.present) {
      map['selling_price'] = Variable<double>(sellingPrice.value);
    }
    if (dbp.present) {
      map['dbp'] = Variable<double>(dbp.value);
    }
    if (wbp.present) {
      map['wbp'] = Variable<double>(wbp.value);
    }
    if (rbp.present) {
      map['rbp'] = Variable<double>(rbp.value);
    }
    if (costPrice.present) {
      map['cost_price'] = Variable<double>(costPrice.value);
    }
    if (unit.present) {
      map['unit'] = Variable<String>(unit.value);
    }
    if (category.present) {
      map['category'] = Variable<String>(category.value);
    }
    if (warehouseStock.present) {
      map['warehouse_stock'] = Variable<int>(warehouseStock.value);
    }
    if (pcsPerUnit.present) {
      map['pcs_per_unit'] = Variable<int>(pcsPerUnit.value);
    }
    if (imageUrl.present) {
      map['image_url'] = Variable<String>(imageUrl.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('ProductsCompanion(')
          ..write('id: $id, ')
          ..write('companyId: $companyId, ')
          ..write('name: $name, ')
          ..write('sku: $sku, ')
          ..write('sellingPrice: $sellingPrice, ')
          ..write('dbp: $dbp, ')
          ..write('wbp: $wbp, ')
          ..write('rbp: $rbp, ')
          ..write('costPrice: $costPrice, ')
          ..write('unit: $unit, ')
          ..write('category: $category, ')
          ..write('warehouseStock: $warehouseStock, ')
          ..write('pcsPerUnit: $pcsPerUnit, ')
          ..write('imageUrl: $imageUrl, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $CompaniesTable extends Companies
    with TableInfo<$CompaniesTable, Company> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $CompaniesTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<String> id = GeneratedColumn<String>(
      'id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _nameMeta = const VerificationMeta('name');
  @override
  late final GeneratedColumn<String> name = GeneratedColumn<String>(
      'name', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _codeMeta = const VerificationMeta('code');
  @override
  late final GeneratedColumn<String> code = GeneratedColumn<String>(
      'code', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  @override
  List<GeneratedColumn> get $columns => [id, name, code];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'companies';
  @override
  VerificationContext validateIntegrity(Insertable<Company> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    } else if (isInserting) {
      context.missing(_idMeta);
    }
    if (data.containsKey('name')) {
      context.handle(
          _nameMeta, name.isAcceptableOrUnknown(data['name']!, _nameMeta));
    } else if (isInserting) {
      context.missing(_nameMeta);
    }
    if (data.containsKey('code')) {
      context.handle(
          _codeMeta, code.isAcceptableOrUnknown(data['code']!, _codeMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  Company map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return Company(
      id: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}id'])!,
      name: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}name'])!,
      code: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}code']),
    );
  }

  @override
  $CompaniesTable createAlias(String alias) {
    return $CompaniesTable(attachedDatabase, alias);
  }
}

class Company extends DataClass implements Insertable<Company> {
  final String id;
  final String name;
  final String? code;
  const Company({required this.id, required this.name, this.code});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<String>(id);
    map['name'] = Variable<String>(name);
    if (!nullToAbsent || code != null) {
      map['code'] = Variable<String>(code);
    }
    return map;
  }

  CompaniesCompanion toCompanion(bool nullToAbsent) {
    return CompaniesCompanion(
      id: Value(id),
      name: Value(name),
      code: code == null && nullToAbsent ? const Value.absent() : Value(code),
    );
  }

  factory Company.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return Company(
      id: serializer.fromJson<String>(json['id']),
      name: serializer.fromJson<String>(json['name']),
      code: serializer.fromJson<String?>(json['code']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<String>(id),
      'name': serializer.toJson<String>(name),
      'code': serializer.toJson<String?>(code),
    };
  }

  Company copyWith(
          {String? id,
          String? name,
          Value<String?> code = const Value.absent()}) =>
      Company(
        id: id ?? this.id,
        name: name ?? this.name,
        code: code.present ? code.value : this.code,
      );
  Company copyWithCompanion(CompaniesCompanion data) {
    return Company(
      id: data.id.present ? data.id.value : this.id,
      name: data.name.present ? data.name.value : this.name,
      code: data.code.present ? data.code.value : this.code,
    );
  }

  @override
  String toString() {
    return (StringBuffer('Company(')
          ..write('id: $id, ')
          ..write('name: $name, ')
          ..write('code: $code')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(id, name, code);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is Company &&
          other.id == this.id &&
          other.name == this.name &&
          other.code == this.code);
}

class CompaniesCompanion extends UpdateCompanion<Company> {
  final Value<String> id;
  final Value<String> name;
  final Value<String?> code;
  final Value<int> rowid;
  const CompaniesCompanion({
    this.id = const Value.absent(),
    this.name = const Value.absent(),
    this.code = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  CompaniesCompanion.insert({
    required String id,
    required String name,
    this.code = const Value.absent(),
    this.rowid = const Value.absent(),
  })  : id = Value(id),
        name = Value(name);
  static Insertable<Company> custom({
    Expression<String>? id,
    Expression<String>? name,
    Expression<String>? code,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (name != null) 'name': name,
      if (code != null) 'code': code,
      if (rowid != null) 'rowid': rowid,
    });
  }

  CompaniesCompanion copyWith(
      {Value<String>? id,
      Value<String>? name,
      Value<String?>? code,
      Value<int>? rowid}) {
    return CompaniesCompanion(
      id: id ?? this.id,
      name: name ?? this.name,
      code: code ?? this.code,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<String>(id.value);
    }
    if (name.present) {
      map['name'] = Variable<String>(name.value);
    }
    if (code.present) {
      map['code'] = Variable<String>(code.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('CompaniesCompanion(')
          ..write('id: $id, ')
          ..write('name: $name, ')
          ..write('code: $code, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $LocalCheckinsTable extends LocalCheckins
    with TableInfo<$LocalCheckinsTable, LocalCheckin> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $LocalCheckinsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<int> id = GeneratedColumn<int>(
      'id', aliasedName, false,
      hasAutoIncrement: true,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultConstraints:
          GeneratedColumn.constraintIsAlways('PRIMARY KEY AUTOINCREMENT'));
  static const VerificationMeta _storeIdMeta =
      const VerificationMeta('storeId');
  @override
  late final GeneratedColumn<String> storeId = GeneratedColumn<String>(
      'store_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _storeNameMeta =
      const VerificationMeta('storeName');
  @override
  late final GeneratedColumn<String> storeName = GeneratedColumn<String>(
      'store_name', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _latitudeMeta =
      const VerificationMeta('latitude');
  @override
  late final GeneratedColumn<double> latitude = GeneratedColumn<double>(
      'latitude', aliasedName, false,
      type: DriftSqlType.double, requiredDuringInsert: true);
  static const VerificationMeta _longitudeMeta =
      const VerificationMeta('longitude');
  @override
  late final GeneratedColumn<double> longitude = GeneratedColumn<double>(
      'longitude', aliasedName, false,
      type: DriftSqlType.double, requiredDuringInsert: true);
  static const VerificationMeta _selfiePathMeta =
      const VerificationMeta('selfiePath');
  @override
  late final GeneratedColumn<String> selfiePath = GeneratedColumn<String>(
      'selfie_path', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _createdAtMeta =
      const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
      'created_at', aliasedName, false,
      type: DriftSqlType.dateTime, requiredDuringInsert: true);
  static const VerificationMeta _syncStatusMeta =
      const VerificationMeta('syncStatus');
  @override
  late final GeneratedColumn<String> syncStatus = GeneratedColumn<String>(
      'sync_status', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('pending'));
  @override
  List<GeneratedColumn> get $columns => [
        id,
        storeId,
        storeName,
        latitude,
        longitude,
        selfiePath,
        createdAt,
        syncStatus
      ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'local_checkins';
  @override
  VerificationContext validateIntegrity(Insertable<LocalCheckin> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    }
    if (data.containsKey('store_id')) {
      context.handle(_storeIdMeta,
          storeId.isAcceptableOrUnknown(data['store_id']!, _storeIdMeta));
    } else if (isInserting) {
      context.missing(_storeIdMeta);
    }
    if (data.containsKey('store_name')) {
      context.handle(_storeNameMeta,
          storeName.isAcceptableOrUnknown(data['store_name']!, _storeNameMeta));
    } else if (isInserting) {
      context.missing(_storeNameMeta);
    }
    if (data.containsKey('latitude')) {
      context.handle(_latitudeMeta,
          latitude.isAcceptableOrUnknown(data['latitude']!, _latitudeMeta));
    } else if (isInserting) {
      context.missing(_latitudeMeta);
    }
    if (data.containsKey('longitude')) {
      context.handle(_longitudeMeta,
          longitude.isAcceptableOrUnknown(data['longitude']!, _longitudeMeta));
    } else if (isInserting) {
      context.missing(_longitudeMeta);
    }
    if (data.containsKey('selfie_path')) {
      context.handle(
          _selfiePathMeta,
          selfiePath.isAcceptableOrUnknown(
              data['selfie_path']!, _selfiePathMeta));
    } else if (isInserting) {
      context.missing(_selfiePathMeta);
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta,
          createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    if (data.containsKey('sync_status')) {
      context.handle(
          _syncStatusMeta,
          syncStatus.isAcceptableOrUnknown(
              data['sync_status']!, _syncStatusMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  LocalCheckin map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return LocalCheckin(
      id: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}id'])!,
      storeId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}store_id'])!,
      storeName: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}store_name'])!,
      latitude: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}latitude'])!,
      longitude: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}longitude'])!,
      selfiePath: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}selfie_path'])!,
      createdAt: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}created_at'])!,
      syncStatus: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}sync_status'])!,
    );
  }

  @override
  $LocalCheckinsTable createAlias(String alias) {
    return $LocalCheckinsTable(attachedDatabase, alias);
  }
}

class LocalCheckin extends DataClass implements Insertable<LocalCheckin> {
  final int id;
  final String storeId;
  final String storeName;
  final double latitude;
  final double longitude;
  final String selfiePath;
  final DateTime createdAt;
  final String syncStatus;
  const LocalCheckin(
      {required this.id,
      required this.storeId,
      required this.storeName,
      required this.latitude,
      required this.longitude,
      required this.selfiePath,
      required this.createdAt,
      required this.syncStatus});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<int>(id);
    map['store_id'] = Variable<String>(storeId);
    map['store_name'] = Variable<String>(storeName);
    map['latitude'] = Variable<double>(latitude);
    map['longitude'] = Variable<double>(longitude);
    map['selfie_path'] = Variable<String>(selfiePath);
    map['created_at'] = Variable<DateTime>(createdAt);
    map['sync_status'] = Variable<String>(syncStatus);
    return map;
  }

  LocalCheckinsCompanion toCompanion(bool nullToAbsent) {
    return LocalCheckinsCompanion(
      id: Value(id),
      storeId: Value(storeId),
      storeName: Value(storeName),
      latitude: Value(latitude),
      longitude: Value(longitude),
      selfiePath: Value(selfiePath),
      createdAt: Value(createdAt),
      syncStatus: Value(syncStatus),
    );
  }

  factory LocalCheckin.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return LocalCheckin(
      id: serializer.fromJson<int>(json['id']),
      storeId: serializer.fromJson<String>(json['storeId']),
      storeName: serializer.fromJson<String>(json['storeName']),
      latitude: serializer.fromJson<double>(json['latitude']),
      longitude: serializer.fromJson<double>(json['longitude']),
      selfiePath: serializer.fromJson<String>(json['selfiePath']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
      syncStatus: serializer.fromJson<String>(json['syncStatus']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<int>(id),
      'storeId': serializer.toJson<String>(storeId),
      'storeName': serializer.toJson<String>(storeName),
      'latitude': serializer.toJson<double>(latitude),
      'longitude': serializer.toJson<double>(longitude),
      'selfiePath': serializer.toJson<String>(selfiePath),
      'createdAt': serializer.toJson<DateTime>(createdAt),
      'syncStatus': serializer.toJson<String>(syncStatus),
    };
  }

  LocalCheckin copyWith(
          {int? id,
          String? storeId,
          String? storeName,
          double? latitude,
          double? longitude,
          String? selfiePath,
          DateTime? createdAt,
          String? syncStatus}) =>
      LocalCheckin(
        id: id ?? this.id,
        storeId: storeId ?? this.storeId,
        storeName: storeName ?? this.storeName,
        latitude: latitude ?? this.latitude,
        longitude: longitude ?? this.longitude,
        selfiePath: selfiePath ?? this.selfiePath,
        createdAt: createdAt ?? this.createdAt,
        syncStatus: syncStatus ?? this.syncStatus,
      );
  LocalCheckin copyWithCompanion(LocalCheckinsCompanion data) {
    return LocalCheckin(
      id: data.id.present ? data.id.value : this.id,
      storeId: data.storeId.present ? data.storeId.value : this.storeId,
      storeName: data.storeName.present ? data.storeName.value : this.storeName,
      latitude: data.latitude.present ? data.latitude.value : this.latitude,
      longitude: data.longitude.present ? data.longitude.value : this.longitude,
      selfiePath:
          data.selfiePath.present ? data.selfiePath.value : this.selfiePath,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      syncStatus:
          data.syncStatus.present ? data.syncStatus.value : this.syncStatus,
    );
  }

  @override
  String toString() {
    return (StringBuffer('LocalCheckin(')
          ..write('id: $id, ')
          ..write('storeId: $storeId, ')
          ..write('storeName: $storeName, ')
          ..write('latitude: $latitude, ')
          ..write('longitude: $longitude, ')
          ..write('selfiePath: $selfiePath, ')
          ..write('createdAt: $createdAt, ')
          ..write('syncStatus: $syncStatus')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(id, storeId, storeName, latitude, longitude,
      selfiePath, createdAt, syncStatus);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is LocalCheckin &&
          other.id == this.id &&
          other.storeId == this.storeId &&
          other.storeName == this.storeName &&
          other.latitude == this.latitude &&
          other.longitude == this.longitude &&
          other.selfiePath == this.selfiePath &&
          other.createdAt == this.createdAt &&
          other.syncStatus == this.syncStatus);
}

class LocalCheckinsCompanion extends UpdateCompanion<LocalCheckin> {
  final Value<int> id;
  final Value<String> storeId;
  final Value<String> storeName;
  final Value<double> latitude;
  final Value<double> longitude;
  final Value<String> selfiePath;
  final Value<DateTime> createdAt;
  final Value<String> syncStatus;
  const LocalCheckinsCompanion({
    this.id = const Value.absent(),
    this.storeId = const Value.absent(),
    this.storeName = const Value.absent(),
    this.latitude = const Value.absent(),
    this.longitude = const Value.absent(),
    this.selfiePath = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.syncStatus = const Value.absent(),
  });
  LocalCheckinsCompanion.insert({
    this.id = const Value.absent(),
    required String storeId,
    required String storeName,
    required double latitude,
    required double longitude,
    required String selfiePath,
    required DateTime createdAt,
    this.syncStatus = const Value.absent(),
  })  : storeId = Value(storeId),
        storeName = Value(storeName),
        latitude = Value(latitude),
        longitude = Value(longitude),
        selfiePath = Value(selfiePath),
        createdAt = Value(createdAt);
  static Insertable<LocalCheckin> custom({
    Expression<int>? id,
    Expression<String>? storeId,
    Expression<String>? storeName,
    Expression<double>? latitude,
    Expression<double>? longitude,
    Expression<String>? selfiePath,
    Expression<DateTime>? createdAt,
    Expression<String>? syncStatus,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (storeId != null) 'store_id': storeId,
      if (storeName != null) 'store_name': storeName,
      if (latitude != null) 'latitude': latitude,
      if (longitude != null) 'longitude': longitude,
      if (selfiePath != null) 'selfie_path': selfiePath,
      if (createdAt != null) 'created_at': createdAt,
      if (syncStatus != null) 'sync_status': syncStatus,
    });
  }

  LocalCheckinsCompanion copyWith(
      {Value<int>? id,
      Value<String>? storeId,
      Value<String>? storeName,
      Value<double>? latitude,
      Value<double>? longitude,
      Value<String>? selfiePath,
      Value<DateTime>? createdAt,
      Value<String>? syncStatus}) {
    return LocalCheckinsCompanion(
      id: id ?? this.id,
      storeId: storeId ?? this.storeId,
      storeName: storeName ?? this.storeName,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      selfiePath: selfiePath ?? this.selfiePath,
      createdAt: createdAt ?? this.createdAt,
      syncStatus: syncStatus ?? this.syncStatus,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<int>(id.value);
    }
    if (storeId.present) {
      map['store_id'] = Variable<String>(storeId.value);
    }
    if (storeName.present) {
      map['store_name'] = Variable<String>(storeName.value);
    }
    if (latitude.present) {
      map['latitude'] = Variable<double>(latitude.value);
    }
    if (longitude.present) {
      map['longitude'] = Variable<double>(longitude.value);
    }
    if (selfiePath.present) {
      map['selfie_path'] = Variable<String>(selfiePath.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (syncStatus.present) {
      map['sync_status'] = Variable<String>(syncStatus.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('LocalCheckinsCompanion(')
          ..write('id: $id, ')
          ..write('storeId: $storeId, ')
          ..write('storeName: $storeName, ')
          ..write('latitude: $latitude, ')
          ..write('longitude: $longitude, ')
          ..write('selfiePath: $selfiePath, ')
          ..write('createdAt: $createdAt, ')
          ..write('syncStatus: $syncStatus')
          ..write(')'))
        .toString();
  }
}

class $LocalTransactionsTable extends LocalTransactions
    with TableInfo<$LocalTransactionsTable, LocalTransaction> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $LocalTransactionsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _localIdMeta =
      const VerificationMeta('localId');
  @override
  late final GeneratedColumn<String> localId = GeneratedColumn<String>(
      'local_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _receiptNoMeta =
      const VerificationMeta('receiptNo');
  @override
  late final GeneratedColumn<String> receiptNo = GeneratedColumn<String>(
      'receipt_no', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _companyIdMeta =
      const VerificationMeta('companyId');
  @override
  late final GeneratedColumn<String> companyId = GeneratedColumn<String>(
      'company_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _companyNameMeta =
      const VerificationMeta('companyName');
  @override
  late final GeneratedColumn<String> companyName = GeneratedColumn<String>(
      'company_name', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _storeIdMeta =
      const VerificationMeta('storeId');
  @override
  late final GeneratedColumn<String> storeId = GeneratedColumn<String>(
      'store_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _storeNameMeta =
      const VerificationMeta('storeName');
  @override
  late final GeneratedColumn<String> storeName = GeneratedColumn<String>(
      'store_name', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _totalAmountMeta =
      const VerificationMeta('totalAmount');
  @override
  late final GeneratedColumn<double> totalAmount = GeneratedColumn<double>(
      'total_amount', aliasedName, false,
      type: DriftSqlType.double, requiredDuringInsert: true);
  static const VerificationMeta _paidAmountMeta =
      const VerificationMeta('paidAmount');
  @override
  late final GeneratedColumn<double> paidAmount = GeneratedColumn<double>(
      'paid_amount', aliasedName, false,
      type: DriftSqlType.double,
      requiredDuringInsert: false,
      defaultValue: const Constant(0.0));
  static const VerificationMeta _selfiePathMeta =
      const VerificationMeta('selfiePath');
  @override
  late final GeneratedColumn<String> selfiePath = GeneratedColumn<String>(
      'selfie_path', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _receiptPathMeta =
      const VerificationMeta('receiptPath');
  @override
  late final GeneratedColumn<String> receiptPath = GeneratedColumn<String>(
      'receipt_path', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _signaturePathMeta =
      const VerificationMeta('signaturePath');
  @override
  late final GeneratedColumn<String> signaturePath = GeneratedColumn<String>(
      'signature_path', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _notesMeta = const VerificationMeta('notes');
  @override
  late final GeneratedColumn<String> notes = GeneratedColumn<String>(
      'notes', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _paymentMethodMeta =
      const VerificationMeta('paymentMethod');
  @override
  late final GeneratedColumn<String> paymentMethod = GeneratedColumn<String>(
      'payment_method', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('CASH'));
  static const VerificationMeta _dpPaymentMethodMeta =
      const VerificationMeta('dpPaymentMethod');
  @override
  late final GeneratedColumn<String> dpPaymentMethod = GeneratedColumn<String>(
      'dp_payment_method', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _paymentBankMeta =
      const VerificationMeta('paymentBank');
  @override
  late final GeneratedColumn<String> paymentBank = GeneratedColumn<String>(
      'payment_bank', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _midtransIdMeta =
      const VerificationMeta('midtransId');
  @override
  late final GeneratedColumn<String> midtransId = GeneratedColumn<String>(
      'midtrans_id', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _midtransQrisUrlMeta =
      const VerificationMeta('midtransQrisUrl');
  @override
  late final GeneratedColumn<String> midtransQrisUrl = GeneratedColumn<String>(
      'midtrans_qris_url', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _midtransVaNumberMeta =
      const VerificationMeta('midtransVaNumber');
  @override
  late final GeneratedColumn<String> midtransVaNumber = GeneratedColumn<String>(
      'midtrans_va_number', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _midtransBankMeta =
      const VerificationMeta('midtransBank');
  @override
  late final GeneratedColumn<String> midtransBank = GeneratedColumn<String>(
      'midtrans_bank', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _midtransBillKeyMeta =
      const VerificationMeta('midtransBillKey');
  @override
  late final GeneratedColumn<String> midtransBillKey = GeneratedColumn<String>(
      'midtrans_bill_key', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _midtransBillerCodeMeta =
      const VerificationMeta('midtransBillerCode');
  @override
  late final GeneratedColumn<String> midtransBillerCode =
      GeneratedColumn<String>('midtrans_biller_code', aliasedName, true,
          type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _createdAtMeta =
      const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
      'created_at', aliasedName, false,
      type: DriftSqlType.dateTime, requiredDuringInsert: true);
  static const VerificationMeta _syncStatusMeta =
      const VerificationMeta('syncStatus');
  @override
  late final GeneratedColumn<String> syncStatus = GeneratedColumn<String>(
      'sync_status', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('pending'));
  static const VerificationMeta _paymentDueDateMeta =
      const VerificationMeta('paymentDueDate');
  @override
  late final GeneratedColumn<DateTime> paymentDueDate =
      GeneratedColumn<DateTime>('payment_due_date', aliasedName, true,
          type: DriftSqlType.dateTime, requiredDuringInsert: false);
  static const VerificationMeta _tenorMeta = const VerificationMeta('tenor');
  @override
  late final GeneratedColumn<int> tenor = GeneratedColumn<int>(
      'tenor', aliasedName, false,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultValue: const Constant(1));
  static const VerificationMeta _isMtMeta = const VerificationMeta('isMt');
  @override
  late final GeneratedColumn<bool> isMt = GeneratedColumn<bool>(
      'is_mt', aliasedName, false,
      type: DriftSqlType.bool,
      requiredDuringInsert: false,
      defaultConstraints:
          GeneratedColumn.constraintIsAlways('CHECK ("is_mt" IN (0, 1))'),
      defaultValue: const Constant(false));
  static const VerificationMeta _isDistributorOrderMeta =
      const VerificationMeta('isDistributorOrder');
  @override
  late final GeneratedColumn<bool> isDistributorOrder = GeneratedColumn<bool>(
      'is_distributor_order', aliasedName, false,
      type: DriftSqlType.bool,
      requiredDuringInsert: false,
      defaultConstraints: GeneratedColumn.constraintIsAlways(
          'CHECK ("is_distributor_order" IN (0, 1))'),
      defaultValue: const Constant(false));
  static const VerificationMeta _factoryCompanyIdMeta =
      const VerificationMeta('factoryCompanyId');
  @override
  late final GeneratedColumn<String> factoryCompanyId = GeneratedColumn<String>(
      'factory_company_id', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _companyBankAccountIdMeta =
      const VerificationMeta('companyBankAccountId');
  @override
  late final GeneratedColumn<String> companyBankAccountId =
      GeneratedColumn<String>('company_bank_account_id', aliasedName, true,
          type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _visitIdMeta =
      const VerificationMeta('visitId');
  @override
  late final GeneratedColumn<String> visitId = GeneratedColumn<String>(
      'visit_id', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _latitudeMeta =
      const VerificationMeta('latitude');
  @override
  late final GeneratedColumn<double> latitude = GeneratedColumn<double>(
      'latitude', aliasedName, true,
      type: DriftSqlType.double, requiredDuringInsert: false);
  static const VerificationMeta _longitudeMeta =
      const VerificationMeta('longitude');
  @override
  late final GeneratedColumn<double> longitude = GeneratedColumn<double>(
      'longitude', aliasedName, true,
      type: DriftSqlType.double, requiredDuringInsert: false);
  static const VerificationMeta _syncErrorMessageMeta =
      const VerificationMeta('syncErrorMessage');
  @override
  late final GeneratedColumn<String> syncErrorMessage = GeneratedColumn<String>(
      'sync_error_message', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _lastSyncAttemptMeta =
      const VerificationMeta('lastSyncAttempt');
  @override
  late final GeneratedColumn<DateTime> lastSyncAttempt =
      GeneratedColumn<DateTime>('last_sync_attempt', aliasedName, true,
          type: DriftSqlType.dateTime, requiredDuringInsert: false);
  @override
  List<GeneratedColumn> get $columns => [
        localId,
        receiptNo,
        companyId,
        companyName,
        storeId,
        storeName,
        totalAmount,
        paidAmount,
        selfiePath,
        receiptPath,
        signaturePath,
        notes,
        paymentMethod,
        dpPaymentMethod,
        paymentBank,
        midtransId,
        midtransQrisUrl,
        midtransVaNumber,
        midtransBank,
        midtransBillKey,
        midtransBillerCode,
        createdAt,
        syncStatus,
        paymentDueDate,
        tenor,
        isMt,
        isDistributorOrder,
        factoryCompanyId,
        companyBankAccountId,
        visitId,
        latitude,
        longitude,
        syncErrorMessage,
        lastSyncAttempt
      ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'local_transactions';
  @override
  VerificationContext validateIntegrity(Insertable<LocalTransaction> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('local_id')) {
      context.handle(_localIdMeta,
          localId.isAcceptableOrUnknown(data['local_id']!, _localIdMeta));
    } else if (isInserting) {
      context.missing(_localIdMeta);
    }
    if (data.containsKey('receipt_no')) {
      context.handle(_receiptNoMeta,
          receiptNo.isAcceptableOrUnknown(data['receipt_no']!, _receiptNoMeta));
    }
    if (data.containsKey('company_id')) {
      context.handle(_companyIdMeta,
          companyId.isAcceptableOrUnknown(data['company_id']!, _companyIdMeta));
    } else if (isInserting) {
      context.missing(_companyIdMeta);
    }
    if (data.containsKey('company_name')) {
      context.handle(
          _companyNameMeta,
          companyName.isAcceptableOrUnknown(
              data['company_name']!, _companyNameMeta));
    } else if (isInserting) {
      context.missing(_companyNameMeta);
    }
    if (data.containsKey('store_id')) {
      context.handle(_storeIdMeta,
          storeId.isAcceptableOrUnknown(data['store_id']!, _storeIdMeta));
    } else if (isInserting) {
      context.missing(_storeIdMeta);
    }
    if (data.containsKey('store_name')) {
      context.handle(_storeNameMeta,
          storeName.isAcceptableOrUnknown(data['store_name']!, _storeNameMeta));
    } else if (isInserting) {
      context.missing(_storeNameMeta);
    }
    if (data.containsKey('total_amount')) {
      context.handle(
          _totalAmountMeta,
          totalAmount.isAcceptableOrUnknown(
              data['total_amount']!, _totalAmountMeta));
    } else if (isInserting) {
      context.missing(_totalAmountMeta);
    }
    if (data.containsKey('paid_amount')) {
      context.handle(
          _paidAmountMeta,
          paidAmount.isAcceptableOrUnknown(
              data['paid_amount']!, _paidAmountMeta));
    }
    if (data.containsKey('selfie_path')) {
      context.handle(
          _selfiePathMeta,
          selfiePath.isAcceptableOrUnknown(
              data['selfie_path']!, _selfiePathMeta));
    }
    if (data.containsKey('receipt_path')) {
      context.handle(
          _receiptPathMeta,
          receiptPath.isAcceptableOrUnknown(
              data['receipt_path']!, _receiptPathMeta));
    }
    if (data.containsKey('signature_path')) {
      context.handle(
          _signaturePathMeta,
          signaturePath.isAcceptableOrUnknown(
              data['signature_path']!, _signaturePathMeta));
    }
    if (data.containsKey('notes')) {
      context.handle(
          _notesMeta, notes.isAcceptableOrUnknown(data['notes']!, _notesMeta));
    }
    if (data.containsKey('payment_method')) {
      context.handle(
          _paymentMethodMeta,
          paymentMethod.isAcceptableOrUnknown(
              data['payment_method']!, _paymentMethodMeta));
    }
    if (data.containsKey('dp_payment_method')) {
      context.handle(
          _dpPaymentMethodMeta,
          dpPaymentMethod.isAcceptableOrUnknown(
              data['dp_payment_method']!, _dpPaymentMethodMeta));
    }
    if (data.containsKey('payment_bank')) {
      context.handle(
          _paymentBankMeta,
          paymentBank.isAcceptableOrUnknown(
              data['payment_bank']!, _paymentBankMeta));
    }
    if (data.containsKey('midtrans_id')) {
      context.handle(
          _midtransIdMeta,
          midtransId.isAcceptableOrUnknown(
              data['midtrans_id']!, _midtransIdMeta));
    }
    if (data.containsKey('midtrans_qris_url')) {
      context.handle(
          _midtransQrisUrlMeta,
          midtransQrisUrl.isAcceptableOrUnknown(
              data['midtrans_qris_url']!, _midtransQrisUrlMeta));
    }
    if (data.containsKey('midtrans_va_number')) {
      context.handle(
          _midtransVaNumberMeta,
          midtransVaNumber.isAcceptableOrUnknown(
              data['midtrans_va_number']!, _midtransVaNumberMeta));
    }
    if (data.containsKey('midtrans_bank')) {
      context.handle(
          _midtransBankMeta,
          midtransBank.isAcceptableOrUnknown(
              data['midtrans_bank']!, _midtransBankMeta));
    }
    if (data.containsKey('midtrans_bill_key')) {
      context.handle(
          _midtransBillKeyMeta,
          midtransBillKey.isAcceptableOrUnknown(
              data['midtrans_bill_key']!, _midtransBillKeyMeta));
    }
    if (data.containsKey('midtrans_biller_code')) {
      context.handle(
          _midtransBillerCodeMeta,
          midtransBillerCode.isAcceptableOrUnknown(
              data['midtrans_biller_code']!, _midtransBillerCodeMeta));
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta,
          createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    if (data.containsKey('sync_status')) {
      context.handle(
          _syncStatusMeta,
          syncStatus.isAcceptableOrUnknown(
              data['sync_status']!, _syncStatusMeta));
    }
    if (data.containsKey('payment_due_date')) {
      context.handle(
          _paymentDueDateMeta,
          paymentDueDate.isAcceptableOrUnknown(
              data['payment_due_date']!, _paymentDueDateMeta));
    }
    if (data.containsKey('tenor')) {
      context.handle(
          _tenorMeta, tenor.isAcceptableOrUnknown(data['tenor']!, _tenorMeta));
    }
    if (data.containsKey('is_mt')) {
      context.handle(
          _isMtMeta, isMt.isAcceptableOrUnknown(data['is_mt']!, _isMtMeta));
    }
    if (data.containsKey('is_distributor_order')) {
      context.handle(
          _isDistributorOrderMeta,
          isDistributorOrder.isAcceptableOrUnknown(
              data['is_distributor_order']!, _isDistributorOrderMeta));
    }
    if (data.containsKey('factory_company_id')) {
      context.handle(
          _factoryCompanyIdMeta,
          factoryCompanyId.isAcceptableOrUnknown(
              data['factory_company_id']!, _factoryCompanyIdMeta));
    }
    if (data.containsKey('company_bank_account_id')) {
      context.handle(
          _companyBankAccountIdMeta,
          companyBankAccountId.isAcceptableOrUnknown(
              data['company_bank_account_id']!, _companyBankAccountIdMeta));
    }
    if (data.containsKey('visit_id')) {
      context.handle(_visitIdMeta,
          visitId.isAcceptableOrUnknown(data['visit_id']!, _visitIdMeta));
    }
    if (data.containsKey('latitude')) {
      context.handle(_latitudeMeta,
          latitude.isAcceptableOrUnknown(data['latitude']!, _latitudeMeta));
    }
    if (data.containsKey('longitude')) {
      context.handle(_longitudeMeta,
          longitude.isAcceptableOrUnknown(data['longitude']!, _longitudeMeta));
    }
    if (data.containsKey('sync_error_message')) {
      context.handle(
          _syncErrorMessageMeta,
          syncErrorMessage.isAcceptableOrUnknown(
              data['sync_error_message']!, _syncErrorMessageMeta));
    }
    if (data.containsKey('last_sync_attempt')) {
      context.handle(
          _lastSyncAttemptMeta,
          lastSyncAttempt.isAcceptableOrUnknown(
              data['last_sync_attempt']!, _lastSyncAttemptMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {localId};
  @override
  LocalTransaction map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return LocalTransaction(
      localId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}local_id'])!,
      receiptNo: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}receipt_no']),
      companyId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}company_id'])!,
      companyName: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}company_name'])!,
      storeId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}store_id'])!,
      storeName: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}store_name'])!,
      totalAmount: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}total_amount'])!,
      paidAmount: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}paid_amount'])!,
      selfiePath: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}selfie_path']),
      receiptPath: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}receipt_path']),
      signaturePath: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}signature_path']),
      notes: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}notes']),
      paymentMethod: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}payment_method'])!,
      dpPaymentMethod: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}dp_payment_method']),
      paymentBank: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}payment_bank']),
      midtransId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}midtrans_id']),
      midtransQrisUrl: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}midtrans_qris_url']),
      midtransVaNumber: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}midtrans_va_number']),
      midtransBank: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}midtrans_bank']),
      midtransBillKey: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}midtrans_bill_key']),
      midtransBillerCode: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}midtrans_biller_code']),
      createdAt: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}created_at'])!,
      syncStatus: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}sync_status'])!,
      paymentDueDate: attachedDatabase.typeMapping.read(
          DriftSqlType.dateTime, data['${effectivePrefix}payment_due_date']),
      tenor: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}tenor'])!,
      isMt: attachedDatabase.typeMapping
          .read(DriftSqlType.bool, data['${effectivePrefix}is_mt'])!,
      isDistributorOrder: attachedDatabase.typeMapping.read(
          DriftSqlType.bool, data['${effectivePrefix}is_distributor_order'])!,
      factoryCompanyId: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}factory_company_id']),
      companyBankAccountId: attachedDatabase.typeMapping.read(
          DriftSqlType.string,
          data['${effectivePrefix}company_bank_account_id']),
      visitId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}visit_id']),
      latitude: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}latitude']),
      longitude: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}longitude']),
      syncErrorMessage: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}sync_error_message']),
      lastSyncAttempt: attachedDatabase.typeMapping.read(
          DriftSqlType.dateTime, data['${effectivePrefix}last_sync_attempt']),
    );
  }

  @override
  $LocalTransactionsTable createAlias(String alias) {
    return $LocalTransactionsTable(attachedDatabase, alias);
  }
}

class LocalTransaction extends DataClass
    implements Insertable<LocalTransaction> {
  final String localId;
  final String? receiptNo;
  final String companyId;
  final String companyName;
  final String storeId;
  final String storeName;
  final double totalAmount;
  final double paidAmount;
  final String? selfiePath;
  final String? receiptPath;
  final String? signaturePath;
  final String? notes;
  final String paymentMethod;
  final String? dpPaymentMethod;
  final String? paymentBank;
  final String? midtransId;
  final String? midtransQrisUrl;
  final String? midtransVaNumber;
  final String? midtransBank;
  final String? midtransBillKey;
  final String? midtransBillerCode;
  final DateTime createdAt;
  final String syncStatus;
  final DateTime? paymentDueDate;
  final int tenor;

  /// MT flag: true jika transaksi ini dibuat oleh Sales MT.
  /// Faktur dan omzetnya dikelola oleh pusat/pabrik.
  final bool isMt;

  /// Distributor flag: true jika transaksi ini dibuat khusus untuk alur sales distributor
  final bool isDistributorOrder;

  /// UUID company pabrik/pusat yang mengelola omzet MT ini.
  final String? factoryCompanyId;
  final String? companyBankAccountId;
  final String? visitId;
  final double? latitude;
  final double? longitude;
  final String? syncErrorMessage;
  final DateTime? lastSyncAttempt;
  const LocalTransaction(
      {required this.localId,
      this.receiptNo,
      required this.companyId,
      required this.companyName,
      required this.storeId,
      required this.storeName,
      required this.totalAmount,
      required this.paidAmount,
      this.selfiePath,
      this.receiptPath,
      this.signaturePath,
      this.notes,
      required this.paymentMethod,
      this.dpPaymentMethod,
      this.paymentBank,
      this.midtransId,
      this.midtransQrisUrl,
      this.midtransVaNumber,
      this.midtransBank,
      this.midtransBillKey,
      this.midtransBillerCode,
      required this.createdAt,
      required this.syncStatus,
      this.paymentDueDate,
      required this.tenor,
      required this.isMt,
      required this.isDistributorOrder,
      this.factoryCompanyId,
      this.companyBankAccountId,
      this.visitId,
      this.latitude,
      this.longitude,
      this.syncErrorMessage,
      this.lastSyncAttempt});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['local_id'] = Variable<String>(localId);
    if (!nullToAbsent || receiptNo != null) {
      map['receipt_no'] = Variable<String>(receiptNo);
    }
    map['company_id'] = Variable<String>(companyId);
    map['company_name'] = Variable<String>(companyName);
    map['store_id'] = Variable<String>(storeId);
    map['store_name'] = Variable<String>(storeName);
    map['total_amount'] = Variable<double>(totalAmount);
    map['paid_amount'] = Variable<double>(paidAmount);
    if (!nullToAbsent || selfiePath != null) {
      map['selfie_path'] = Variable<String>(selfiePath);
    }
    if (!nullToAbsent || receiptPath != null) {
      map['receipt_path'] = Variable<String>(receiptPath);
    }
    if (!nullToAbsent || signaturePath != null) {
      map['signature_path'] = Variable<String>(signaturePath);
    }
    if (!nullToAbsent || notes != null) {
      map['notes'] = Variable<String>(notes);
    }
    map['payment_method'] = Variable<String>(paymentMethod);
    if (!nullToAbsent || dpPaymentMethod != null) {
      map['dp_payment_method'] = Variable<String>(dpPaymentMethod);
    }
    if (!nullToAbsent || paymentBank != null) {
      map['payment_bank'] = Variable<String>(paymentBank);
    }
    if (!nullToAbsent || midtransId != null) {
      map['midtrans_id'] = Variable<String>(midtransId);
    }
    if (!nullToAbsent || midtransQrisUrl != null) {
      map['midtrans_qris_url'] = Variable<String>(midtransQrisUrl);
    }
    if (!nullToAbsent || midtransVaNumber != null) {
      map['midtrans_va_number'] = Variable<String>(midtransVaNumber);
    }
    if (!nullToAbsent || midtransBank != null) {
      map['midtrans_bank'] = Variable<String>(midtransBank);
    }
    if (!nullToAbsent || midtransBillKey != null) {
      map['midtrans_bill_key'] = Variable<String>(midtransBillKey);
    }
    if (!nullToAbsent || midtransBillerCode != null) {
      map['midtrans_biller_code'] = Variable<String>(midtransBillerCode);
    }
    map['created_at'] = Variable<DateTime>(createdAt);
    map['sync_status'] = Variable<String>(syncStatus);
    if (!nullToAbsent || paymentDueDate != null) {
      map['payment_due_date'] = Variable<DateTime>(paymentDueDate);
    }
    map['tenor'] = Variable<int>(tenor);
    map['is_mt'] = Variable<bool>(isMt);
    map['is_distributor_order'] = Variable<bool>(isDistributorOrder);
    if (!nullToAbsent || factoryCompanyId != null) {
      map['factory_company_id'] = Variable<String>(factoryCompanyId);
    }
    if (!nullToAbsent || companyBankAccountId != null) {
      map['company_bank_account_id'] = Variable<String>(companyBankAccountId);
    }
    if (!nullToAbsent || visitId != null) {
      map['visit_id'] = Variable<String>(visitId);
    }
    if (!nullToAbsent || latitude != null) {
      map['latitude'] = Variable<double>(latitude);
    }
    if (!nullToAbsent || longitude != null) {
      map['longitude'] = Variable<double>(longitude);
    }
    if (!nullToAbsent || syncErrorMessage != null) {
      map['sync_error_message'] = Variable<String>(syncErrorMessage);
    }
    if (!nullToAbsent || lastSyncAttempt != null) {
      map['last_sync_attempt'] = Variable<DateTime>(lastSyncAttempt);
    }
    return map;
  }

  LocalTransactionsCompanion toCompanion(bool nullToAbsent) {
    return LocalTransactionsCompanion(
      localId: Value(localId),
      receiptNo: receiptNo == null && nullToAbsent
          ? const Value.absent()
          : Value(receiptNo),
      companyId: Value(companyId),
      companyName: Value(companyName),
      storeId: Value(storeId),
      storeName: Value(storeName),
      totalAmount: Value(totalAmount),
      paidAmount: Value(paidAmount),
      selfiePath: selfiePath == null && nullToAbsent
          ? const Value.absent()
          : Value(selfiePath),
      receiptPath: receiptPath == null && nullToAbsent
          ? const Value.absent()
          : Value(receiptPath),
      signaturePath: signaturePath == null && nullToAbsent
          ? const Value.absent()
          : Value(signaturePath),
      notes:
          notes == null && nullToAbsent ? const Value.absent() : Value(notes),
      paymentMethod: Value(paymentMethod),
      dpPaymentMethod: dpPaymentMethod == null && nullToAbsent
          ? const Value.absent()
          : Value(dpPaymentMethod),
      paymentBank: paymentBank == null && nullToAbsent
          ? const Value.absent()
          : Value(paymentBank),
      midtransId: midtransId == null && nullToAbsent
          ? const Value.absent()
          : Value(midtransId),
      midtransQrisUrl: midtransQrisUrl == null && nullToAbsent
          ? const Value.absent()
          : Value(midtransQrisUrl),
      midtransVaNumber: midtransVaNumber == null && nullToAbsent
          ? const Value.absent()
          : Value(midtransVaNumber),
      midtransBank: midtransBank == null && nullToAbsent
          ? const Value.absent()
          : Value(midtransBank),
      midtransBillKey: midtransBillKey == null && nullToAbsent
          ? const Value.absent()
          : Value(midtransBillKey),
      midtransBillerCode: midtransBillerCode == null && nullToAbsent
          ? const Value.absent()
          : Value(midtransBillerCode),
      createdAt: Value(createdAt),
      syncStatus: Value(syncStatus),
      paymentDueDate: paymentDueDate == null && nullToAbsent
          ? const Value.absent()
          : Value(paymentDueDate),
      tenor: Value(tenor),
      isMt: Value(isMt),
      isDistributorOrder: Value(isDistributorOrder),
      factoryCompanyId: factoryCompanyId == null && nullToAbsent
          ? const Value.absent()
          : Value(factoryCompanyId),
      companyBankAccountId: companyBankAccountId == null && nullToAbsent
          ? const Value.absent()
          : Value(companyBankAccountId),
      visitId: visitId == null && nullToAbsent
          ? const Value.absent()
          : Value(visitId),
      latitude: latitude == null && nullToAbsent
          ? const Value.absent()
          : Value(latitude),
      longitude: longitude == null && nullToAbsent
          ? const Value.absent()
          : Value(longitude),
      syncErrorMessage: syncErrorMessage == null && nullToAbsent
          ? const Value.absent()
          : Value(syncErrorMessage),
      lastSyncAttempt: lastSyncAttempt == null && nullToAbsent
          ? const Value.absent()
          : Value(lastSyncAttempt),
    );
  }

  factory LocalTransaction.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return LocalTransaction(
      localId: serializer.fromJson<String>(json['localId']),
      receiptNo: serializer.fromJson<String?>(json['receiptNo']),
      companyId: serializer.fromJson<String>(json['companyId']),
      companyName: serializer.fromJson<String>(json['companyName']),
      storeId: serializer.fromJson<String>(json['storeId']),
      storeName: serializer.fromJson<String>(json['storeName']),
      totalAmount: serializer.fromJson<double>(json['totalAmount']),
      paidAmount: serializer.fromJson<double>(json['paidAmount']),
      selfiePath: serializer.fromJson<String?>(json['selfiePath']),
      receiptPath: serializer.fromJson<String?>(json['receiptPath']),
      signaturePath: serializer.fromJson<String?>(json['signaturePath']),
      notes: serializer.fromJson<String?>(json['notes']),
      paymentMethod: serializer.fromJson<String>(json['paymentMethod']),
      dpPaymentMethod: serializer.fromJson<String?>(json['dpPaymentMethod']),
      paymentBank: serializer.fromJson<String?>(json['paymentBank']),
      midtransId: serializer.fromJson<String?>(json['midtransId']),
      midtransQrisUrl: serializer.fromJson<String?>(json['midtransQrisUrl']),
      midtransVaNumber: serializer.fromJson<String?>(json['midtransVaNumber']),
      midtransBank: serializer.fromJson<String?>(json['midtransBank']),
      midtransBillKey: serializer.fromJson<String?>(json['midtransBillKey']),
      midtransBillerCode:
          serializer.fromJson<String?>(json['midtransBillerCode']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
      syncStatus: serializer.fromJson<String>(json['syncStatus']),
      paymentDueDate: serializer.fromJson<DateTime?>(json['paymentDueDate']),
      tenor: serializer.fromJson<int>(json['tenor']),
      isMt: serializer.fromJson<bool>(json['isMt']),
      isDistributorOrder: serializer.fromJson<bool>(json['isDistributorOrder']),
      factoryCompanyId: serializer.fromJson<String?>(json['factoryCompanyId']),
      companyBankAccountId:
          serializer.fromJson<String?>(json['companyBankAccountId']),
      visitId: serializer.fromJson<String?>(json['visitId']),
      latitude: serializer.fromJson<double?>(json['latitude']),
      longitude: serializer.fromJson<double?>(json['longitude']),
      syncErrorMessage: serializer.fromJson<String?>(json['syncErrorMessage']),
      lastSyncAttempt: serializer.fromJson<DateTime?>(json['lastSyncAttempt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'localId': serializer.toJson<String>(localId),
      'receiptNo': serializer.toJson<String?>(receiptNo),
      'companyId': serializer.toJson<String>(companyId),
      'companyName': serializer.toJson<String>(companyName),
      'storeId': serializer.toJson<String>(storeId),
      'storeName': serializer.toJson<String>(storeName),
      'totalAmount': serializer.toJson<double>(totalAmount),
      'paidAmount': serializer.toJson<double>(paidAmount),
      'selfiePath': serializer.toJson<String?>(selfiePath),
      'receiptPath': serializer.toJson<String?>(receiptPath),
      'signaturePath': serializer.toJson<String?>(signaturePath),
      'notes': serializer.toJson<String?>(notes),
      'paymentMethod': serializer.toJson<String>(paymentMethod),
      'dpPaymentMethod': serializer.toJson<String?>(dpPaymentMethod),
      'paymentBank': serializer.toJson<String?>(paymentBank),
      'midtransId': serializer.toJson<String?>(midtransId),
      'midtransQrisUrl': serializer.toJson<String?>(midtransQrisUrl),
      'midtransVaNumber': serializer.toJson<String?>(midtransVaNumber),
      'midtransBank': serializer.toJson<String?>(midtransBank),
      'midtransBillKey': serializer.toJson<String?>(midtransBillKey),
      'midtransBillerCode': serializer.toJson<String?>(midtransBillerCode),
      'createdAt': serializer.toJson<DateTime>(createdAt),
      'syncStatus': serializer.toJson<String>(syncStatus),
      'paymentDueDate': serializer.toJson<DateTime?>(paymentDueDate),
      'tenor': serializer.toJson<int>(tenor),
      'isMt': serializer.toJson<bool>(isMt),
      'isDistributorOrder': serializer.toJson<bool>(isDistributorOrder),
      'factoryCompanyId': serializer.toJson<String?>(factoryCompanyId),
      'companyBankAccountId': serializer.toJson<String?>(companyBankAccountId),
      'visitId': serializer.toJson<String?>(visitId),
      'latitude': serializer.toJson<double?>(latitude),
      'longitude': serializer.toJson<double?>(longitude),
      'syncErrorMessage': serializer.toJson<String?>(syncErrorMessage),
      'lastSyncAttempt': serializer.toJson<DateTime?>(lastSyncAttempt),
    };
  }

  LocalTransaction copyWith(
          {String? localId,
          Value<String?> receiptNo = const Value.absent(),
          String? companyId,
          String? companyName,
          String? storeId,
          String? storeName,
          double? totalAmount,
          double? paidAmount,
          Value<String?> selfiePath = const Value.absent(),
          Value<String?> receiptPath = const Value.absent(),
          Value<String?> signaturePath = const Value.absent(),
          Value<String?> notes = const Value.absent(),
          String? paymentMethod,
          Value<String?> dpPaymentMethod = const Value.absent(),
          Value<String?> paymentBank = const Value.absent(),
          Value<String?> midtransId = const Value.absent(),
          Value<String?> midtransQrisUrl = const Value.absent(),
          Value<String?> midtransVaNumber = const Value.absent(),
          Value<String?> midtransBank = const Value.absent(),
          Value<String?> midtransBillKey = const Value.absent(),
          Value<String?> midtransBillerCode = const Value.absent(),
          DateTime? createdAt,
          String? syncStatus,
          Value<DateTime?> paymentDueDate = const Value.absent(),
          int? tenor,
          bool? isMt,
          bool? isDistributorOrder,
          Value<String?> factoryCompanyId = const Value.absent(),
          Value<String?> companyBankAccountId = const Value.absent(),
          Value<String?> visitId = const Value.absent(),
          Value<double?> latitude = const Value.absent(),
          Value<double?> longitude = const Value.absent(),
          Value<String?> syncErrorMessage = const Value.absent(),
          Value<DateTime?> lastSyncAttempt = const Value.absent()}) =>
      LocalTransaction(
        localId: localId ?? this.localId,
        receiptNo: receiptNo.present ? receiptNo.value : this.receiptNo,
        companyId: companyId ?? this.companyId,
        companyName: companyName ?? this.companyName,
        storeId: storeId ?? this.storeId,
        storeName: storeName ?? this.storeName,
        totalAmount: totalAmount ?? this.totalAmount,
        paidAmount: paidAmount ?? this.paidAmount,
        selfiePath: selfiePath.present ? selfiePath.value : this.selfiePath,
        receiptPath: receiptPath.present ? receiptPath.value : this.receiptPath,
        signaturePath:
            signaturePath.present ? signaturePath.value : this.signaturePath,
        notes: notes.present ? notes.value : this.notes,
        paymentMethod: paymentMethod ?? this.paymentMethod,
        dpPaymentMethod: dpPaymentMethod.present
            ? dpPaymentMethod.value
            : this.dpPaymentMethod,
        paymentBank: paymentBank.present ? paymentBank.value : this.paymentBank,
        midtransId: midtransId.present ? midtransId.value : this.midtransId,
        midtransQrisUrl: midtransQrisUrl.present
            ? midtransQrisUrl.value
            : this.midtransQrisUrl,
        midtransVaNumber: midtransVaNumber.present
            ? midtransVaNumber.value
            : this.midtransVaNumber,
        midtransBank:
            midtransBank.present ? midtransBank.value : this.midtransBank,
        midtransBillKey: midtransBillKey.present
            ? midtransBillKey.value
            : this.midtransBillKey,
        midtransBillerCode: midtransBillerCode.present
            ? midtransBillerCode.value
            : this.midtransBillerCode,
        createdAt: createdAt ?? this.createdAt,
        syncStatus: syncStatus ?? this.syncStatus,
        paymentDueDate:
            paymentDueDate.present ? paymentDueDate.value : this.paymentDueDate,
        tenor: tenor ?? this.tenor,
        isMt: isMt ?? this.isMt,
        isDistributorOrder: isDistributorOrder ?? this.isDistributorOrder,
        factoryCompanyId: factoryCompanyId.present
            ? factoryCompanyId.value
            : this.factoryCompanyId,
        companyBankAccountId: companyBankAccountId.present
            ? companyBankAccountId.value
            : this.companyBankAccountId,
        visitId: visitId.present ? visitId.value : this.visitId,
        latitude: latitude.present ? latitude.value : this.latitude,
        longitude: longitude.present ? longitude.value : this.longitude,
        syncErrorMessage: syncErrorMessage.present
            ? syncErrorMessage.value
            : this.syncErrorMessage,
        lastSyncAttempt: lastSyncAttempt.present
            ? lastSyncAttempt.value
            : this.lastSyncAttempt,
      );
  LocalTransaction copyWithCompanion(LocalTransactionsCompanion data) {
    return LocalTransaction(
      localId: data.localId.present ? data.localId.value : this.localId,
      receiptNo: data.receiptNo.present ? data.receiptNo.value : this.receiptNo,
      companyId: data.companyId.present ? data.companyId.value : this.companyId,
      companyName:
          data.companyName.present ? data.companyName.value : this.companyName,
      storeId: data.storeId.present ? data.storeId.value : this.storeId,
      storeName: data.storeName.present ? data.storeName.value : this.storeName,
      totalAmount:
          data.totalAmount.present ? data.totalAmount.value : this.totalAmount,
      paidAmount:
          data.paidAmount.present ? data.paidAmount.value : this.paidAmount,
      selfiePath:
          data.selfiePath.present ? data.selfiePath.value : this.selfiePath,
      receiptPath:
          data.receiptPath.present ? data.receiptPath.value : this.receiptPath,
      signaturePath: data.signaturePath.present
          ? data.signaturePath.value
          : this.signaturePath,
      notes: data.notes.present ? data.notes.value : this.notes,
      paymentMethod: data.paymentMethod.present
          ? data.paymentMethod.value
          : this.paymentMethod,
      dpPaymentMethod: data.dpPaymentMethod.present
          ? data.dpPaymentMethod.value
          : this.dpPaymentMethod,
      paymentBank:
          data.paymentBank.present ? data.paymentBank.value : this.paymentBank,
      midtransId:
          data.midtransId.present ? data.midtransId.value : this.midtransId,
      midtransQrisUrl: data.midtransQrisUrl.present
          ? data.midtransQrisUrl.value
          : this.midtransQrisUrl,
      midtransVaNumber: data.midtransVaNumber.present
          ? data.midtransVaNumber.value
          : this.midtransVaNumber,
      midtransBank: data.midtransBank.present
          ? data.midtransBank.value
          : this.midtransBank,
      midtransBillKey: data.midtransBillKey.present
          ? data.midtransBillKey.value
          : this.midtransBillKey,
      midtransBillerCode: data.midtransBillerCode.present
          ? data.midtransBillerCode.value
          : this.midtransBillerCode,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      syncStatus:
          data.syncStatus.present ? data.syncStatus.value : this.syncStatus,
      paymentDueDate: data.paymentDueDate.present
          ? data.paymentDueDate.value
          : this.paymentDueDate,
      tenor: data.tenor.present ? data.tenor.value : this.tenor,
      isMt: data.isMt.present ? data.isMt.value : this.isMt,
      isDistributorOrder: data.isDistributorOrder.present
          ? data.isDistributorOrder.value
          : this.isDistributorOrder,
      factoryCompanyId: data.factoryCompanyId.present
          ? data.factoryCompanyId.value
          : this.factoryCompanyId,
      companyBankAccountId: data.companyBankAccountId.present
          ? data.companyBankAccountId.value
          : this.companyBankAccountId,
      visitId: data.visitId.present ? data.visitId.value : this.visitId,
      latitude: data.latitude.present ? data.latitude.value : this.latitude,
      longitude: data.longitude.present ? data.longitude.value : this.longitude,
      syncErrorMessage: data.syncErrorMessage.present
          ? data.syncErrorMessage.value
          : this.syncErrorMessage,
      lastSyncAttempt: data.lastSyncAttempt.present
          ? data.lastSyncAttempt.value
          : this.lastSyncAttempt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('LocalTransaction(')
          ..write('localId: $localId, ')
          ..write('receiptNo: $receiptNo, ')
          ..write('companyId: $companyId, ')
          ..write('companyName: $companyName, ')
          ..write('storeId: $storeId, ')
          ..write('storeName: $storeName, ')
          ..write('totalAmount: $totalAmount, ')
          ..write('paidAmount: $paidAmount, ')
          ..write('selfiePath: $selfiePath, ')
          ..write('receiptPath: $receiptPath, ')
          ..write('signaturePath: $signaturePath, ')
          ..write('notes: $notes, ')
          ..write('paymentMethod: $paymentMethod, ')
          ..write('dpPaymentMethod: $dpPaymentMethod, ')
          ..write('paymentBank: $paymentBank, ')
          ..write('midtransId: $midtransId, ')
          ..write('midtransQrisUrl: $midtransQrisUrl, ')
          ..write('midtransVaNumber: $midtransVaNumber, ')
          ..write('midtransBank: $midtransBank, ')
          ..write('midtransBillKey: $midtransBillKey, ')
          ..write('midtransBillerCode: $midtransBillerCode, ')
          ..write('createdAt: $createdAt, ')
          ..write('syncStatus: $syncStatus, ')
          ..write('paymentDueDate: $paymentDueDate, ')
          ..write('tenor: $tenor, ')
          ..write('isMt: $isMt, ')
          ..write('isDistributorOrder: $isDistributorOrder, ')
          ..write('factoryCompanyId: $factoryCompanyId, ')
          ..write('companyBankAccountId: $companyBankAccountId, ')
          ..write('visitId: $visitId, ')
          ..write('latitude: $latitude, ')
          ..write('longitude: $longitude, ')
          ..write('syncErrorMessage: $syncErrorMessage, ')
          ..write('lastSyncAttempt: $lastSyncAttempt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hashAll([
        localId,
        receiptNo,
        companyId,
        companyName,
        storeId,
        storeName,
        totalAmount,
        paidAmount,
        selfiePath,
        receiptPath,
        signaturePath,
        notes,
        paymentMethod,
        dpPaymentMethod,
        paymentBank,
        midtransId,
        midtransQrisUrl,
        midtransVaNumber,
        midtransBank,
        midtransBillKey,
        midtransBillerCode,
        createdAt,
        syncStatus,
        paymentDueDate,
        tenor,
        isMt,
        isDistributorOrder,
        factoryCompanyId,
        companyBankAccountId,
        visitId,
        latitude,
        longitude,
        syncErrorMessage,
        lastSyncAttempt
      ]);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is LocalTransaction &&
          other.localId == this.localId &&
          other.receiptNo == this.receiptNo &&
          other.companyId == this.companyId &&
          other.companyName == this.companyName &&
          other.storeId == this.storeId &&
          other.storeName == this.storeName &&
          other.totalAmount == this.totalAmount &&
          other.paidAmount == this.paidAmount &&
          other.selfiePath == this.selfiePath &&
          other.receiptPath == this.receiptPath &&
          other.signaturePath == this.signaturePath &&
          other.notes == this.notes &&
          other.paymentMethod == this.paymentMethod &&
          other.dpPaymentMethod == this.dpPaymentMethod &&
          other.paymentBank == this.paymentBank &&
          other.midtransId == this.midtransId &&
          other.midtransQrisUrl == this.midtransQrisUrl &&
          other.midtransVaNumber == this.midtransVaNumber &&
          other.midtransBank == this.midtransBank &&
          other.midtransBillKey == this.midtransBillKey &&
          other.midtransBillerCode == this.midtransBillerCode &&
          other.createdAt == this.createdAt &&
          other.syncStatus == this.syncStatus &&
          other.paymentDueDate == this.paymentDueDate &&
          other.tenor == this.tenor &&
          other.isMt == this.isMt &&
          other.isDistributorOrder == this.isDistributorOrder &&
          other.factoryCompanyId == this.factoryCompanyId &&
          other.companyBankAccountId == this.companyBankAccountId &&
          other.visitId == this.visitId &&
          other.latitude == this.latitude &&
          other.longitude == this.longitude &&
          other.syncErrorMessage == this.syncErrorMessage &&
          other.lastSyncAttempt == this.lastSyncAttempt);
}

class LocalTransactionsCompanion extends UpdateCompanion<LocalTransaction> {
  final Value<String> localId;
  final Value<String?> receiptNo;
  final Value<String> companyId;
  final Value<String> companyName;
  final Value<String> storeId;
  final Value<String> storeName;
  final Value<double> totalAmount;
  final Value<double> paidAmount;
  final Value<String?> selfiePath;
  final Value<String?> receiptPath;
  final Value<String?> signaturePath;
  final Value<String?> notes;
  final Value<String> paymentMethod;
  final Value<String?> dpPaymentMethod;
  final Value<String?> paymentBank;
  final Value<String?> midtransId;
  final Value<String?> midtransQrisUrl;
  final Value<String?> midtransVaNumber;
  final Value<String?> midtransBank;
  final Value<String?> midtransBillKey;
  final Value<String?> midtransBillerCode;
  final Value<DateTime> createdAt;
  final Value<String> syncStatus;
  final Value<DateTime?> paymentDueDate;
  final Value<int> tenor;
  final Value<bool> isMt;
  final Value<bool> isDistributorOrder;
  final Value<String?> factoryCompanyId;
  final Value<String?> companyBankAccountId;
  final Value<String?> visitId;
  final Value<double?> latitude;
  final Value<double?> longitude;
  final Value<String?> syncErrorMessage;
  final Value<DateTime?> lastSyncAttempt;
  final Value<int> rowid;
  const LocalTransactionsCompanion({
    this.localId = const Value.absent(),
    this.receiptNo = const Value.absent(),
    this.companyId = const Value.absent(),
    this.companyName = const Value.absent(),
    this.storeId = const Value.absent(),
    this.storeName = const Value.absent(),
    this.totalAmount = const Value.absent(),
    this.paidAmount = const Value.absent(),
    this.selfiePath = const Value.absent(),
    this.receiptPath = const Value.absent(),
    this.signaturePath = const Value.absent(),
    this.notes = const Value.absent(),
    this.paymentMethod = const Value.absent(),
    this.dpPaymentMethod = const Value.absent(),
    this.paymentBank = const Value.absent(),
    this.midtransId = const Value.absent(),
    this.midtransQrisUrl = const Value.absent(),
    this.midtransVaNumber = const Value.absent(),
    this.midtransBank = const Value.absent(),
    this.midtransBillKey = const Value.absent(),
    this.midtransBillerCode = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.syncStatus = const Value.absent(),
    this.paymentDueDate = const Value.absent(),
    this.tenor = const Value.absent(),
    this.isMt = const Value.absent(),
    this.isDistributorOrder = const Value.absent(),
    this.factoryCompanyId = const Value.absent(),
    this.companyBankAccountId = const Value.absent(),
    this.visitId = const Value.absent(),
    this.latitude = const Value.absent(),
    this.longitude = const Value.absent(),
    this.syncErrorMessage = const Value.absent(),
    this.lastSyncAttempt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  LocalTransactionsCompanion.insert({
    required String localId,
    this.receiptNo = const Value.absent(),
    required String companyId,
    required String companyName,
    required String storeId,
    required String storeName,
    required double totalAmount,
    this.paidAmount = const Value.absent(),
    this.selfiePath = const Value.absent(),
    this.receiptPath = const Value.absent(),
    this.signaturePath = const Value.absent(),
    this.notes = const Value.absent(),
    this.paymentMethod = const Value.absent(),
    this.dpPaymentMethod = const Value.absent(),
    this.paymentBank = const Value.absent(),
    this.midtransId = const Value.absent(),
    this.midtransQrisUrl = const Value.absent(),
    this.midtransVaNumber = const Value.absent(),
    this.midtransBank = const Value.absent(),
    this.midtransBillKey = const Value.absent(),
    this.midtransBillerCode = const Value.absent(),
    required DateTime createdAt,
    this.syncStatus = const Value.absent(),
    this.paymentDueDate = const Value.absent(),
    this.tenor = const Value.absent(),
    this.isMt = const Value.absent(),
    this.isDistributorOrder = const Value.absent(),
    this.factoryCompanyId = const Value.absent(),
    this.companyBankAccountId = const Value.absent(),
    this.visitId = const Value.absent(),
    this.latitude = const Value.absent(),
    this.longitude = const Value.absent(),
    this.syncErrorMessage = const Value.absent(),
    this.lastSyncAttempt = const Value.absent(),
    this.rowid = const Value.absent(),
  })  : localId = Value(localId),
        companyId = Value(companyId),
        companyName = Value(companyName),
        storeId = Value(storeId),
        storeName = Value(storeName),
        totalAmount = Value(totalAmount),
        createdAt = Value(createdAt);
  static Insertable<LocalTransaction> custom({
    Expression<String>? localId,
    Expression<String>? receiptNo,
    Expression<String>? companyId,
    Expression<String>? companyName,
    Expression<String>? storeId,
    Expression<String>? storeName,
    Expression<double>? totalAmount,
    Expression<double>? paidAmount,
    Expression<String>? selfiePath,
    Expression<String>? receiptPath,
    Expression<String>? signaturePath,
    Expression<String>? notes,
    Expression<String>? paymentMethod,
    Expression<String>? dpPaymentMethod,
    Expression<String>? paymentBank,
    Expression<String>? midtransId,
    Expression<String>? midtransQrisUrl,
    Expression<String>? midtransVaNumber,
    Expression<String>? midtransBank,
    Expression<String>? midtransBillKey,
    Expression<String>? midtransBillerCode,
    Expression<DateTime>? createdAt,
    Expression<String>? syncStatus,
    Expression<DateTime>? paymentDueDate,
    Expression<int>? tenor,
    Expression<bool>? isMt,
    Expression<bool>? isDistributorOrder,
    Expression<String>? factoryCompanyId,
    Expression<String>? companyBankAccountId,
    Expression<String>? visitId,
    Expression<double>? latitude,
    Expression<double>? longitude,
    Expression<String>? syncErrorMessage,
    Expression<DateTime>? lastSyncAttempt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (localId != null) 'local_id': localId,
      if (receiptNo != null) 'receipt_no': receiptNo,
      if (companyId != null) 'company_id': companyId,
      if (companyName != null) 'company_name': companyName,
      if (storeId != null) 'store_id': storeId,
      if (storeName != null) 'store_name': storeName,
      if (totalAmount != null) 'total_amount': totalAmount,
      if (paidAmount != null) 'paid_amount': paidAmount,
      if (selfiePath != null) 'selfie_path': selfiePath,
      if (receiptPath != null) 'receipt_path': receiptPath,
      if (signaturePath != null) 'signature_path': signaturePath,
      if (notes != null) 'notes': notes,
      if (paymentMethod != null) 'payment_method': paymentMethod,
      if (dpPaymentMethod != null) 'dp_payment_method': dpPaymentMethod,
      if (paymentBank != null) 'payment_bank': paymentBank,
      if (midtransId != null) 'midtrans_id': midtransId,
      if (midtransQrisUrl != null) 'midtrans_qris_url': midtransQrisUrl,
      if (midtransVaNumber != null) 'midtrans_va_number': midtransVaNumber,
      if (midtransBank != null) 'midtrans_bank': midtransBank,
      if (midtransBillKey != null) 'midtrans_bill_key': midtransBillKey,
      if (midtransBillerCode != null)
        'midtrans_biller_code': midtransBillerCode,
      if (createdAt != null) 'created_at': createdAt,
      if (syncStatus != null) 'sync_status': syncStatus,
      if (paymentDueDate != null) 'payment_due_date': paymentDueDate,
      if (tenor != null) 'tenor': tenor,
      if (isMt != null) 'is_mt': isMt,
      if (isDistributorOrder != null)
        'is_distributor_order': isDistributorOrder,
      if (factoryCompanyId != null) 'factory_company_id': factoryCompanyId,
      if (companyBankAccountId != null)
        'company_bank_account_id': companyBankAccountId,
      if (visitId != null) 'visit_id': visitId,
      if (latitude != null) 'latitude': latitude,
      if (longitude != null) 'longitude': longitude,
      if (syncErrorMessage != null) 'sync_error_message': syncErrorMessage,
      if (lastSyncAttempt != null) 'last_sync_attempt': lastSyncAttempt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  LocalTransactionsCompanion copyWith(
      {Value<String>? localId,
      Value<String?>? receiptNo,
      Value<String>? companyId,
      Value<String>? companyName,
      Value<String>? storeId,
      Value<String>? storeName,
      Value<double>? totalAmount,
      Value<double>? paidAmount,
      Value<String?>? selfiePath,
      Value<String?>? receiptPath,
      Value<String?>? signaturePath,
      Value<String?>? notes,
      Value<String>? paymentMethod,
      Value<String?>? dpPaymentMethod,
      Value<String?>? paymentBank,
      Value<String?>? midtransId,
      Value<String?>? midtransQrisUrl,
      Value<String?>? midtransVaNumber,
      Value<String?>? midtransBank,
      Value<String?>? midtransBillKey,
      Value<String?>? midtransBillerCode,
      Value<DateTime>? createdAt,
      Value<String>? syncStatus,
      Value<DateTime?>? paymentDueDate,
      Value<int>? tenor,
      Value<bool>? isMt,
      Value<bool>? isDistributorOrder,
      Value<String?>? factoryCompanyId,
      Value<String?>? companyBankAccountId,
      Value<String?>? visitId,
      Value<double?>? latitude,
      Value<double?>? longitude,
      Value<String?>? syncErrorMessage,
      Value<DateTime?>? lastSyncAttempt,
      Value<int>? rowid}) {
    return LocalTransactionsCompanion(
      localId: localId ?? this.localId,
      receiptNo: receiptNo ?? this.receiptNo,
      companyId: companyId ?? this.companyId,
      companyName: companyName ?? this.companyName,
      storeId: storeId ?? this.storeId,
      storeName: storeName ?? this.storeName,
      totalAmount: totalAmount ?? this.totalAmount,
      paidAmount: paidAmount ?? this.paidAmount,
      selfiePath: selfiePath ?? this.selfiePath,
      receiptPath: receiptPath ?? this.receiptPath,
      signaturePath: signaturePath ?? this.signaturePath,
      notes: notes ?? this.notes,
      paymentMethod: paymentMethod ?? this.paymentMethod,
      dpPaymentMethod: dpPaymentMethod ?? this.dpPaymentMethod,
      paymentBank: paymentBank ?? this.paymentBank,
      midtransId: midtransId ?? this.midtransId,
      midtransQrisUrl: midtransQrisUrl ?? this.midtransQrisUrl,
      midtransVaNumber: midtransVaNumber ?? this.midtransVaNumber,
      midtransBank: midtransBank ?? this.midtransBank,
      midtransBillKey: midtransBillKey ?? this.midtransBillKey,
      midtransBillerCode: midtransBillerCode ?? this.midtransBillerCode,
      createdAt: createdAt ?? this.createdAt,
      syncStatus: syncStatus ?? this.syncStatus,
      paymentDueDate: paymentDueDate ?? this.paymentDueDate,
      tenor: tenor ?? this.tenor,
      isMt: isMt ?? this.isMt,
      isDistributorOrder: isDistributorOrder ?? this.isDistributorOrder,
      factoryCompanyId: factoryCompanyId ?? this.factoryCompanyId,
      companyBankAccountId: companyBankAccountId ?? this.companyBankAccountId,
      visitId: visitId ?? this.visitId,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      syncErrorMessage: syncErrorMessage ?? this.syncErrorMessage,
      lastSyncAttempt: lastSyncAttempt ?? this.lastSyncAttempt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (localId.present) {
      map['local_id'] = Variable<String>(localId.value);
    }
    if (receiptNo.present) {
      map['receipt_no'] = Variable<String>(receiptNo.value);
    }
    if (companyId.present) {
      map['company_id'] = Variable<String>(companyId.value);
    }
    if (companyName.present) {
      map['company_name'] = Variable<String>(companyName.value);
    }
    if (storeId.present) {
      map['store_id'] = Variable<String>(storeId.value);
    }
    if (storeName.present) {
      map['store_name'] = Variable<String>(storeName.value);
    }
    if (totalAmount.present) {
      map['total_amount'] = Variable<double>(totalAmount.value);
    }
    if (paidAmount.present) {
      map['paid_amount'] = Variable<double>(paidAmount.value);
    }
    if (selfiePath.present) {
      map['selfie_path'] = Variable<String>(selfiePath.value);
    }
    if (receiptPath.present) {
      map['receipt_path'] = Variable<String>(receiptPath.value);
    }
    if (signaturePath.present) {
      map['signature_path'] = Variable<String>(signaturePath.value);
    }
    if (notes.present) {
      map['notes'] = Variable<String>(notes.value);
    }
    if (paymentMethod.present) {
      map['payment_method'] = Variable<String>(paymentMethod.value);
    }
    if (dpPaymentMethod.present) {
      map['dp_payment_method'] = Variable<String>(dpPaymentMethod.value);
    }
    if (paymentBank.present) {
      map['payment_bank'] = Variable<String>(paymentBank.value);
    }
    if (midtransId.present) {
      map['midtrans_id'] = Variable<String>(midtransId.value);
    }
    if (midtransQrisUrl.present) {
      map['midtrans_qris_url'] = Variable<String>(midtransQrisUrl.value);
    }
    if (midtransVaNumber.present) {
      map['midtrans_va_number'] = Variable<String>(midtransVaNumber.value);
    }
    if (midtransBank.present) {
      map['midtrans_bank'] = Variable<String>(midtransBank.value);
    }
    if (midtransBillKey.present) {
      map['midtrans_bill_key'] = Variable<String>(midtransBillKey.value);
    }
    if (midtransBillerCode.present) {
      map['midtrans_biller_code'] = Variable<String>(midtransBillerCode.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (syncStatus.present) {
      map['sync_status'] = Variable<String>(syncStatus.value);
    }
    if (paymentDueDate.present) {
      map['payment_due_date'] = Variable<DateTime>(paymentDueDate.value);
    }
    if (tenor.present) {
      map['tenor'] = Variable<int>(tenor.value);
    }
    if (isMt.present) {
      map['is_mt'] = Variable<bool>(isMt.value);
    }
    if (isDistributorOrder.present) {
      map['is_distributor_order'] = Variable<bool>(isDistributorOrder.value);
    }
    if (factoryCompanyId.present) {
      map['factory_company_id'] = Variable<String>(factoryCompanyId.value);
    }
    if (companyBankAccountId.present) {
      map['company_bank_account_id'] =
          Variable<String>(companyBankAccountId.value);
    }
    if (visitId.present) {
      map['visit_id'] = Variable<String>(visitId.value);
    }
    if (latitude.present) {
      map['latitude'] = Variable<double>(latitude.value);
    }
    if (longitude.present) {
      map['longitude'] = Variable<double>(longitude.value);
    }
    if (syncErrorMessage.present) {
      map['sync_error_message'] = Variable<String>(syncErrorMessage.value);
    }
    if (lastSyncAttempt.present) {
      map['last_sync_attempt'] = Variable<DateTime>(lastSyncAttempt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('LocalTransactionsCompanion(')
          ..write('localId: $localId, ')
          ..write('receiptNo: $receiptNo, ')
          ..write('companyId: $companyId, ')
          ..write('companyName: $companyName, ')
          ..write('storeId: $storeId, ')
          ..write('storeName: $storeName, ')
          ..write('totalAmount: $totalAmount, ')
          ..write('paidAmount: $paidAmount, ')
          ..write('selfiePath: $selfiePath, ')
          ..write('receiptPath: $receiptPath, ')
          ..write('signaturePath: $signaturePath, ')
          ..write('notes: $notes, ')
          ..write('paymentMethod: $paymentMethod, ')
          ..write('dpPaymentMethod: $dpPaymentMethod, ')
          ..write('paymentBank: $paymentBank, ')
          ..write('midtransId: $midtransId, ')
          ..write('midtransQrisUrl: $midtransQrisUrl, ')
          ..write('midtransVaNumber: $midtransVaNumber, ')
          ..write('midtransBank: $midtransBank, ')
          ..write('midtransBillKey: $midtransBillKey, ')
          ..write('midtransBillerCode: $midtransBillerCode, ')
          ..write('createdAt: $createdAt, ')
          ..write('syncStatus: $syncStatus, ')
          ..write('paymentDueDate: $paymentDueDate, ')
          ..write('tenor: $tenor, ')
          ..write('isMt: $isMt, ')
          ..write('isDistributorOrder: $isDistributorOrder, ')
          ..write('factoryCompanyId: $factoryCompanyId, ')
          ..write('companyBankAccountId: $companyBankAccountId, ')
          ..write('visitId: $visitId, ')
          ..write('latitude: $latitude, ')
          ..write('longitude: $longitude, ')
          ..write('syncErrorMessage: $syncErrorMessage, ')
          ..write('lastSyncAttempt: $lastSyncAttempt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $LocalTransactionItemsTable extends LocalTransactionItems
    with TableInfo<$LocalTransactionItemsTable, LocalTransactionItem> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $LocalTransactionItemsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<int> id = GeneratedColumn<int>(
      'id', aliasedName, false,
      hasAutoIncrement: true,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultConstraints:
          GeneratedColumn.constraintIsAlways('PRIMARY KEY AUTOINCREMENT'));
  static const VerificationMeta _transactionLocalIdMeta =
      const VerificationMeta('transactionLocalId');
  @override
  late final GeneratedColumn<String> transactionLocalId =
      GeneratedColumn<String>('transaction_local_id', aliasedName, false,
          type: DriftSqlType.string,
          requiredDuringInsert: true,
          defaultConstraints: GeneratedColumn.constraintIsAlways(
              'REFERENCES local_transactions (local_id)'));
  static const VerificationMeta _productIdMeta =
      const VerificationMeta('productId');
  @override
  late final GeneratedColumn<String> productId = GeneratedColumn<String>(
      'product_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _productNameMeta =
      const VerificationMeta('productName');
  @override
  late final GeneratedColumn<String> productName = GeneratedColumn<String>(
      'product_name', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _quantityMeta =
      const VerificationMeta('quantity');
  @override
  late final GeneratedColumn<int> quantity = GeneratedColumn<int>(
      'quantity', aliasedName, false,
      type: DriftSqlType.int, requiredDuringInsert: true);
  static const VerificationMeta _orderedQuantityMeta =
      const VerificationMeta('orderedQuantity');
  @override
  late final GeneratedColumn<int> orderedQuantity = GeneratedColumn<int>(
      'ordered_quantity', aliasedName, false,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultValue: const Constant(0));
  static const VerificationMeta _unitMeta = const VerificationMeta('unit');
  @override
  late final GeneratedColumn<String> unit = GeneratedColumn<String>(
      'unit', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('PCS'));
  static const VerificationMeta _piecesPerUnitMeta =
      const VerificationMeta('piecesPerUnit');
  @override
  late final GeneratedColumn<int> piecesPerUnit = GeneratedColumn<int>(
      'pieces_per_unit', aliasedName, false,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultValue: const Constant(1));
  static const VerificationMeta _priceMeta = const VerificationMeta('price');
  @override
  late final GeneratedColumn<double> price = GeneratedColumn<double>(
      'price', aliasedName, false,
      type: DriftSqlType.double, requiredDuringInsert: true);
  static const VerificationMeta _discountPercentMeta =
      const VerificationMeta('discountPercent');
  @override
  late final GeneratedColumn<double> discountPercent = GeneratedColumn<double>(
      'discount_percent', aliasedName, false,
      type: DriftSqlType.double,
      requiredDuringInsert: false,
      defaultValue: const Constant(0.0));
  @override
  List<GeneratedColumn> get $columns => [
        id,
        transactionLocalId,
        productId,
        productName,
        quantity,
        orderedQuantity,
        unit,
        piecesPerUnit,
        price,
        discountPercent
      ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'local_transaction_items';
  @override
  VerificationContext validateIntegrity(
      Insertable<LocalTransactionItem> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    }
    if (data.containsKey('transaction_local_id')) {
      context.handle(
          _transactionLocalIdMeta,
          transactionLocalId.isAcceptableOrUnknown(
              data['transaction_local_id']!, _transactionLocalIdMeta));
    } else if (isInserting) {
      context.missing(_transactionLocalIdMeta);
    }
    if (data.containsKey('product_id')) {
      context.handle(_productIdMeta,
          productId.isAcceptableOrUnknown(data['product_id']!, _productIdMeta));
    } else if (isInserting) {
      context.missing(_productIdMeta);
    }
    if (data.containsKey('product_name')) {
      context.handle(
          _productNameMeta,
          productName.isAcceptableOrUnknown(
              data['product_name']!, _productNameMeta));
    } else if (isInserting) {
      context.missing(_productNameMeta);
    }
    if (data.containsKey('quantity')) {
      context.handle(_quantityMeta,
          quantity.isAcceptableOrUnknown(data['quantity']!, _quantityMeta));
    } else if (isInserting) {
      context.missing(_quantityMeta);
    }
    if (data.containsKey('ordered_quantity')) {
      context.handle(
          _orderedQuantityMeta,
          orderedQuantity.isAcceptableOrUnknown(
              data['ordered_quantity']!, _orderedQuantityMeta));
    }
    if (data.containsKey('unit')) {
      context.handle(
          _unitMeta, unit.isAcceptableOrUnknown(data['unit']!, _unitMeta));
    }
    if (data.containsKey('pieces_per_unit')) {
      context.handle(
          _piecesPerUnitMeta,
          piecesPerUnit.isAcceptableOrUnknown(
              data['pieces_per_unit']!, _piecesPerUnitMeta));
    }
    if (data.containsKey('price')) {
      context.handle(
          _priceMeta, price.isAcceptableOrUnknown(data['price']!, _priceMeta));
    } else if (isInserting) {
      context.missing(_priceMeta);
    }
    if (data.containsKey('discount_percent')) {
      context.handle(
          _discountPercentMeta,
          discountPercent.isAcceptableOrUnknown(
              data['discount_percent']!, _discountPercentMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  LocalTransactionItem map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return LocalTransactionItem(
      id: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}id'])!,
      transactionLocalId: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}transaction_local_id'])!,
      productId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}product_id'])!,
      productName: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}product_name'])!,
      quantity: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}quantity'])!,
      orderedQuantity: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}ordered_quantity'])!,
      unit: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}unit'])!,
      piecesPerUnit: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}pieces_per_unit'])!,
      price: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}price'])!,
      discountPercent: attachedDatabase.typeMapping.read(
          DriftSqlType.double, data['${effectivePrefix}discount_percent'])!,
    );
  }

  @override
  $LocalTransactionItemsTable createAlias(String alias) {
    return $LocalTransactionItemsTable(attachedDatabase, alias);
  }
}

class LocalTransactionItem extends DataClass
    implements Insertable<LocalTransactionItem> {
  final int id;
  final String transactionLocalId;
  final String productId;
  final String productName;
  final int quantity;
  final int orderedQuantity;
  final String unit;
  final int piecesPerUnit;
  final double price;

  /// Diskon per item dalam persen (0–100). Default 0 = tidak ada diskon.
  final double discountPercent;
  const LocalTransactionItem(
      {required this.id,
      required this.transactionLocalId,
      required this.productId,
      required this.productName,
      required this.quantity,
      required this.orderedQuantity,
      required this.unit,
      required this.piecesPerUnit,
      required this.price,
      required this.discountPercent});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<int>(id);
    map['transaction_local_id'] = Variable<String>(transactionLocalId);
    map['product_id'] = Variable<String>(productId);
    map['product_name'] = Variable<String>(productName);
    map['quantity'] = Variable<int>(quantity);
    map['ordered_quantity'] = Variable<int>(orderedQuantity);
    map['unit'] = Variable<String>(unit);
    map['pieces_per_unit'] = Variable<int>(piecesPerUnit);
    map['price'] = Variable<double>(price);
    map['discount_percent'] = Variable<double>(discountPercent);
    return map;
  }

  LocalTransactionItemsCompanion toCompanion(bool nullToAbsent) {
    return LocalTransactionItemsCompanion(
      id: Value(id),
      transactionLocalId: Value(transactionLocalId),
      productId: Value(productId),
      productName: Value(productName),
      quantity: Value(quantity),
      orderedQuantity: Value(orderedQuantity),
      unit: Value(unit),
      piecesPerUnit: Value(piecesPerUnit),
      price: Value(price),
      discountPercent: Value(discountPercent),
    );
  }

  factory LocalTransactionItem.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return LocalTransactionItem(
      id: serializer.fromJson<int>(json['id']),
      transactionLocalId:
          serializer.fromJson<String>(json['transactionLocalId']),
      productId: serializer.fromJson<String>(json['productId']),
      productName: serializer.fromJson<String>(json['productName']),
      quantity: serializer.fromJson<int>(json['quantity']),
      orderedQuantity: serializer.fromJson<int>(json['orderedQuantity']),
      unit: serializer.fromJson<String>(json['unit']),
      piecesPerUnit: serializer.fromJson<int>(json['piecesPerUnit']),
      price: serializer.fromJson<double>(json['price']),
      discountPercent: serializer.fromJson<double>(json['discountPercent']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<int>(id),
      'transactionLocalId': serializer.toJson<String>(transactionLocalId),
      'productId': serializer.toJson<String>(productId),
      'productName': serializer.toJson<String>(productName),
      'quantity': serializer.toJson<int>(quantity),
      'orderedQuantity': serializer.toJson<int>(orderedQuantity),
      'unit': serializer.toJson<String>(unit),
      'piecesPerUnit': serializer.toJson<int>(piecesPerUnit),
      'price': serializer.toJson<double>(price),
      'discountPercent': serializer.toJson<double>(discountPercent),
    };
  }

  LocalTransactionItem copyWith(
          {int? id,
          String? transactionLocalId,
          String? productId,
          String? productName,
          int? quantity,
          int? orderedQuantity,
          String? unit,
          int? piecesPerUnit,
          double? price,
          double? discountPercent}) =>
      LocalTransactionItem(
        id: id ?? this.id,
        transactionLocalId: transactionLocalId ?? this.transactionLocalId,
        productId: productId ?? this.productId,
        productName: productName ?? this.productName,
        quantity: quantity ?? this.quantity,
        orderedQuantity: orderedQuantity ?? this.orderedQuantity,
        unit: unit ?? this.unit,
        piecesPerUnit: piecesPerUnit ?? this.piecesPerUnit,
        price: price ?? this.price,
        discountPercent: discountPercent ?? this.discountPercent,
      );
  LocalTransactionItem copyWithCompanion(LocalTransactionItemsCompanion data) {
    return LocalTransactionItem(
      id: data.id.present ? data.id.value : this.id,
      transactionLocalId: data.transactionLocalId.present
          ? data.transactionLocalId.value
          : this.transactionLocalId,
      productId: data.productId.present ? data.productId.value : this.productId,
      productName:
          data.productName.present ? data.productName.value : this.productName,
      quantity: data.quantity.present ? data.quantity.value : this.quantity,
      orderedQuantity: data.orderedQuantity.present
          ? data.orderedQuantity.value
          : this.orderedQuantity,
      unit: data.unit.present ? data.unit.value : this.unit,
      piecesPerUnit: data.piecesPerUnit.present
          ? data.piecesPerUnit.value
          : this.piecesPerUnit,
      price: data.price.present ? data.price.value : this.price,
      discountPercent: data.discountPercent.present
          ? data.discountPercent.value
          : this.discountPercent,
    );
  }

  @override
  String toString() {
    return (StringBuffer('LocalTransactionItem(')
          ..write('id: $id, ')
          ..write('transactionLocalId: $transactionLocalId, ')
          ..write('productId: $productId, ')
          ..write('productName: $productName, ')
          ..write('quantity: $quantity, ')
          ..write('orderedQuantity: $orderedQuantity, ')
          ..write('unit: $unit, ')
          ..write('piecesPerUnit: $piecesPerUnit, ')
          ..write('price: $price, ')
          ..write('discountPercent: $discountPercent')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
      id,
      transactionLocalId,
      productId,
      productName,
      quantity,
      orderedQuantity,
      unit,
      piecesPerUnit,
      price,
      discountPercent);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is LocalTransactionItem &&
          other.id == this.id &&
          other.transactionLocalId == this.transactionLocalId &&
          other.productId == this.productId &&
          other.productName == this.productName &&
          other.quantity == this.quantity &&
          other.orderedQuantity == this.orderedQuantity &&
          other.unit == this.unit &&
          other.piecesPerUnit == this.piecesPerUnit &&
          other.price == this.price &&
          other.discountPercent == this.discountPercent);
}

class LocalTransactionItemsCompanion
    extends UpdateCompanion<LocalTransactionItem> {
  final Value<int> id;
  final Value<String> transactionLocalId;
  final Value<String> productId;
  final Value<String> productName;
  final Value<int> quantity;
  final Value<int> orderedQuantity;
  final Value<String> unit;
  final Value<int> piecesPerUnit;
  final Value<double> price;
  final Value<double> discountPercent;
  const LocalTransactionItemsCompanion({
    this.id = const Value.absent(),
    this.transactionLocalId = const Value.absent(),
    this.productId = const Value.absent(),
    this.productName = const Value.absent(),
    this.quantity = const Value.absent(),
    this.orderedQuantity = const Value.absent(),
    this.unit = const Value.absent(),
    this.piecesPerUnit = const Value.absent(),
    this.price = const Value.absent(),
    this.discountPercent = const Value.absent(),
  });
  LocalTransactionItemsCompanion.insert({
    this.id = const Value.absent(),
    required String transactionLocalId,
    required String productId,
    required String productName,
    required int quantity,
    this.orderedQuantity = const Value.absent(),
    this.unit = const Value.absent(),
    this.piecesPerUnit = const Value.absent(),
    required double price,
    this.discountPercent = const Value.absent(),
  })  : transactionLocalId = Value(transactionLocalId),
        productId = Value(productId),
        productName = Value(productName),
        quantity = Value(quantity),
        price = Value(price);
  static Insertable<LocalTransactionItem> custom({
    Expression<int>? id,
    Expression<String>? transactionLocalId,
    Expression<String>? productId,
    Expression<String>? productName,
    Expression<int>? quantity,
    Expression<int>? orderedQuantity,
    Expression<String>? unit,
    Expression<int>? piecesPerUnit,
    Expression<double>? price,
    Expression<double>? discountPercent,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (transactionLocalId != null)
        'transaction_local_id': transactionLocalId,
      if (productId != null) 'product_id': productId,
      if (productName != null) 'product_name': productName,
      if (quantity != null) 'quantity': quantity,
      if (orderedQuantity != null) 'ordered_quantity': orderedQuantity,
      if (unit != null) 'unit': unit,
      if (piecesPerUnit != null) 'pieces_per_unit': piecesPerUnit,
      if (price != null) 'price': price,
      if (discountPercent != null) 'discount_percent': discountPercent,
    });
  }

  LocalTransactionItemsCompanion copyWith(
      {Value<int>? id,
      Value<String>? transactionLocalId,
      Value<String>? productId,
      Value<String>? productName,
      Value<int>? quantity,
      Value<int>? orderedQuantity,
      Value<String>? unit,
      Value<int>? piecesPerUnit,
      Value<double>? price,
      Value<double>? discountPercent}) {
    return LocalTransactionItemsCompanion(
      id: id ?? this.id,
      transactionLocalId: transactionLocalId ?? this.transactionLocalId,
      productId: productId ?? this.productId,
      productName: productName ?? this.productName,
      quantity: quantity ?? this.quantity,
      orderedQuantity: orderedQuantity ?? this.orderedQuantity,
      unit: unit ?? this.unit,
      piecesPerUnit: piecesPerUnit ?? this.piecesPerUnit,
      price: price ?? this.price,
      discountPercent: discountPercent ?? this.discountPercent,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<int>(id.value);
    }
    if (transactionLocalId.present) {
      map['transaction_local_id'] = Variable<String>(transactionLocalId.value);
    }
    if (productId.present) {
      map['product_id'] = Variable<String>(productId.value);
    }
    if (productName.present) {
      map['product_name'] = Variable<String>(productName.value);
    }
    if (quantity.present) {
      map['quantity'] = Variable<int>(quantity.value);
    }
    if (orderedQuantity.present) {
      map['ordered_quantity'] = Variable<int>(orderedQuantity.value);
    }
    if (unit.present) {
      map['unit'] = Variable<String>(unit.value);
    }
    if (piecesPerUnit.present) {
      map['pieces_per_unit'] = Variable<int>(piecesPerUnit.value);
    }
    if (price.present) {
      map['price'] = Variable<double>(price.value);
    }
    if (discountPercent.present) {
      map['discount_percent'] = Variable<double>(discountPercent.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('LocalTransactionItemsCompanion(')
          ..write('id: $id, ')
          ..write('transactionLocalId: $transactionLocalId, ')
          ..write('productId: $productId, ')
          ..write('productName: $productName, ')
          ..write('quantity: $quantity, ')
          ..write('orderedQuantity: $orderedQuantity, ')
          ..write('unit: $unit, ')
          ..write('piecesPerUnit: $piecesPerUnit, ')
          ..write('price: $price, ')
          ..write('discountPercent: $discountPercent')
          ..write(')'))
        .toString();
  }
}

class $SalesStockTable extends SalesStock
    with TableInfo<$SalesStockTable, SalesStockData> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $SalesStockTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _productIdMeta =
      const VerificationMeta('productId');
  @override
  late final GeneratedColumn<String> productId = GeneratedColumn<String>(
      'product_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _quantityMeta =
      const VerificationMeta('quantity');
  @override
  late final GeneratedColumn<int> quantity = GeneratedColumn<int>(
      'quantity', aliasedName, false,
      type: DriftSqlType.int, requiredDuringInsert: true);
  static const VerificationMeta _updatedAtMeta =
      const VerificationMeta('updatedAt');
  @override
  late final GeneratedColumn<DateTime> updatedAt = GeneratedColumn<DateTime>(
      'updated_at', aliasedName, false,
      type: DriftSqlType.dateTime, requiredDuringInsert: true);
  @override
  List<GeneratedColumn> get $columns => [productId, quantity, updatedAt];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'sales_stock';
  @override
  VerificationContext validateIntegrity(Insertable<SalesStockData> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('product_id')) {
      context.handle(_productIdMeta,
          productId.isAcceptableOrUnknown(data['product_id']!, _productIdMeta));
    } else if (isInserting) {
      context.missing(_productIdMeta);
    }
    if (data.containsKey('quantity')) {
      context.handle(_quantityMeta,
          quantity.isAcceptableOrUnknown(data['quantity']!, _quantityMeta));
    } else if (isInserting) {
      context.missing(_quantityMeta);
    }
    if (data.containsKey('updated_at')) {
      context.handle(_updatedAtMeta,
          updatedAt.isAcceptableOrUnknown(data['updated_at']!, _updatedAtMeta));
    } else if (isInserting) {
      context.missing(_updatedAtMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {productId};
  @override
  SalesStockData map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return SalesStockData(
      productId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}product_id'])!,
      quantity: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}quantity'])!,
      updatedAt: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}updated_at'])!,
    );
  }

  @override
  $SalesStockTable createAlias(String alias) {
    return $SalesStockTable(attachedDatabase, alias);
  }
}

class SalesStockData extends DataClass implements Insertable<SalesStockData> {
  final String productId;
  final int quantity;
  final DateTime updatedAt;
  const SalesStockData(
      {required this.productId,
      required this.quantity,
      required this.updatedAt});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['product_id'] = Variable<String>(productId);
    map['quantity'] = Variable<int>(quantity);
    map['updated_at'] = Variable<DateTime>(updatedAt);
    return map;
  }

  SalesStockCompanion toCompanion(bool nullToAbsent) {
    return SalesStockCompanion(
      productId: Value(productId),
      quantity: Value(quantity),
      updatedAt: Value(updatedAt),
    );
  }

  factory SalesStockData.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return SalesStockData(
      productId: serializer.fromJson<String>(json['productId']),
      quantity: serializer.fromJson<int>(json['quantity']),
      updatedAt: serializer.fromJson<DateTime>(json['updatedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'productId': serializer.toJson<String>(productId),
      'quantity': serializer.toJson<int>(quantity),
      'updatedAt': serializer.toJson<DateTime>(updatedAt),
    };
  }

  SalesStockData copyWith(
          {String? productId, int? quantity, DateTime? updatedAt}) =>
      SalesStockData(
        productId: productId ?? this.productId,
        quantity: quantity ?? this.quantity,
        updatedAt: updatedAt ?? this.updatedAt,
      );
  SalesStockData copyWithCompanion(SalesStockCompanion data) {
    return SalesStockData(
      productId: data.productId.present ? data.productId.value : this.productId,
      quantity: data.quantity.present ? data.quantity.value : this.quantity,
      updatedAt: data.updatedAt.present ? data.updatedAt.value : this.updatedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('SalesStockData(')
          ..write('productId: $productId, ')
          ..write('quantity: $quantity, ')
          ..write('updatedAt: $updatedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(productId, quantity, updatedAt);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is SalesStockData &&
          other.productId == this.productId &&
          other.quantity == this.quantity &&
          other.updatedAt == this.updatedAt);
}

class SalesStockCompanion extends UpdateCompanion<SalesStockData> {
  final Value<String> productId;
  final Value<int> quantity;
  final Value<DateTime> updatedAt;
  final Value<int> rowid;
  const SalesStockCompanion({
    this.productId = const Value.absent(),
    this.quantity = const Value.absent(),
    this.updatedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  SalesStockCompanion.insert({
    required String productId,
    required int quantity,
    required DateTime updatedAt,
    this.rowid = const Value.absent(),
  })  : productId = Value(productId),
        quantity = Value(quantity),
        updatedAt = Value(updatedAt);
  static Insertable<SalesStockData> custom({
    Expression<String>? productId,
    Expression<int>? quantity,
    Expression<DateTime>? updatedAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (productId != null) 'product_id': productId,
      if (quantity != null) 'quantity': quantity,
      if (updatedAt != null) 'updated_at': updatedAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  SalesStockCompanion copyWith(
      {Value<String>? productId,
      Value<int>? quantity,
      Value<DateTime>? updatedAt,
      Value<int>? rowid}) {
    return SalesStockCompanion(
      productId: productId ?? this.productId,
      quantity: quantity ?? this.quantity,
      updatedAt: updatedAt ?? this.updatedAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (productId.present) {
      map['product_id'] = Variable<String>(productId.value);
    }
    if (quantity.present) {
      map['quantity'] = Variable<int>(quantity.value);
    }
    if (updatedAt.present) {
      map['updated_at'] = Variable<DateTime>(updatedAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('SalesStockCompanion(')
          ..write('productId: $productId, ')
          ..write('quantity: $quantity, ')
          ..write('updatedAt: $updatedAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $OfflineTasksTable extends OfflineTasks
    with TableInfo<$OfflineTasksTable, OfflineTask> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $OfflineTasksTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<String> id = GeneratedColumn<String>(
      'id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _typeMeta = const VerificationMeta('type');
  @override
  late final GeneratedColumn<String> type = GeneratedColumn<String>(
      'type', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _payloadMeta =
      const VerificationMeta('payload');
  @override
  late final GeneratedColumn<String> payload = GeneratedColumn<String>(
      'payload', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _filePathMeta =
      const VerificationMeta('filePath');
  @override
  late final GeneratedColumn<String> filePath = GeneratedColumn<String>(
      'file_path', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _filePath2Meta =
      const VerificationMeta('filePath2');
  @override
  late final GeneratedColumn<String> filePath2 = GeneratedColumn<String>(
      'file_path2', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _createdAtMeta =
      const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
      'created_at', aliasedName, false,
      type: DriftSqlType.dateTime, requiredDuringInsert: true);
  static const VerificationMeta _retryCountMeta =
      const VerificationMeta('retryCount');
  @override
  late final GeneratedColumn<int> retryCount = GeneratedColumn<int>(
      'retry_count', aliasedName, false,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultValue: const Constant(0));
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
      'status', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('pending'));
  @override
  List<GeneratedColumn> get $columns =>
      [id, type, payload, filePath, filePath2, createdAt, retryCount, status];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'offline_tasks';
  @override
  VerificationContext validateIntegrity(Insertable<OfflineTask> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    } else if (isInserting) {
      context.missing(_idMeta);
    }
    if (data.containsKey('type')) {
      context.handle(
          _typeMeta, type.isAcceptableOrUnknown(data['type']!, _typeMeta));
    } else if (isInserting) {
      context.missing(_typeMeta);
    }
    if (data.containsKey('payload')) {
      context.handle(_payloadMeta,
          payload.isAcceptableOrUnknown(data['payload']!, _payloadMeta));
    } else if (isInserting) {
      context.missing(_payloadMeta);
    }
    if (data.containsKey('file_path')) {
      context.handle(_filePathMeta,
          filePath.isAcceptableOrUnknown(data['file_path']!, _filePathMeta));
    }
    if (data.containsKey('file_path2')) {
      context.handle(_filePath2Meta,
          filePath2.isAcceptableOrUnknown(data['file_path2']!, _filePath2Meta));
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta,
          createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    if (data.containsKey('retry_count')) {
      context.handle(
          _retryCountMeta,
          retryCount.isAcceptableOrUnknown(
              data['retry_count']!, _retryCountMeta));
    }
    if (data.containsKey('status')) {
      context.handle(_statusMeta,
          status.isAcceptableOrUnknown(data['status']!, _statusMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  OfflineTask map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return OfflineTask(
      id: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}id'])!,
      type: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}type'])!,
      payload: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}payload'])!,
      filePath: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}file_path']),
      filePath2: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}file_path2']),
      createdAt: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}created_at'])!,
      retryCount: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}retry_count'])!,
      status: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}status'])!,
    );
  }

  @override
  $OfflineTasksTable createAlias(String alias) {
    return $OfflineTasksTable(attachedDatabase, alias);
  }
}

class OfflineTask extends DataClass implements Insertable<OfflineTask> {
  final String id;
  final String type;
  final String payload;
  final String? filePath;
  final String? filePath2;
  final DateTime createdAt;
  final int retryCount;
  final String status;
  const OfflineTask(
      {required this.id,
      required this.type,
      required this.payload,
      this.filePath,
      this.filePath2,
      required this.createdAt,
      required this.retryCount,
      required this.status});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<String>(id);
    map['type'] = Variable<String>(type);
    map['payload'] = Variable<String>(payload);
    if (!nullToAbsent || filePath != null) {
      map['file_path'] = Variable<String>(filePath);
    }
    if (!nullToAbsent || filePath2 != null) {
      map['file_path2'] = Variable<String>(filePath2);
    }
    map['created_at'] = Variable<DateTime>(createdAt);
    map['retry_count'] = Variable<int>(retryCount);
    map['status'] = Variable<String>(status);
    return map;
  }

  OfflineTasksCompanion toCompanion(bool nullToAbsent) {
    return OfflineTasksCompanion(
      id: Value(id),
      type: Value(type),
      payload: Value(payload),
      filePath: filePath == null && nullToAbsent
          ? const Value.absent()
          : Value(filePath),
      filePath2: filePath2 == null && nullToAbsent
          ? const Value.absent()
          : Value(filePath2),
      createdAt: Value(createdAt),
      retryCount: Value(retryCount),
      status: Value(status),
    );
  }

  factory OfflineTask.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return OfflineTask(
      id: serializer.fromJson<String>(json['id']),
      type: serializer.fromJson<String>(json['type']),
      payload: serializer.fromJson<String>(json['payload']),
      filePath: serializer.fromJson<String?>(json['filePath']),
      filePath2: serializer.fromJson<String?>(json['filePath2']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
      retryCount: serializer.fromJson<int>(json['retryCount']),
      status: serializer.fromJson<String>(json['status']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<String>(id),
      'type': serializer.toJson<String>(type),
      'payload': serializer.toJson<String>(payload),
      'filePath': serializer.toJson<String?>(filePath),
      'filePath2': serializer.toJson<String?>(filePath2),
      'createdAt': serializer.toJson<DateTime>(createdAt),
      'retryCount': serializer.toJson<int>(retryCount),
      'status': serializer.toJson<String>(status),
    };
  }

  OfflineTask copyWith(
          {String? id,
          String? type,
          String? payload,
          Value<String?> filePath = const Value.absent(),
          Value<String?> filePath2 = const Value.absent(),
          DateTime? createdAt,
          int? retryCount,
          String? status}) =>
      OfflineTask(
        id: id ?? this.id,
        type: type ?? this.type,
        payload: payload ?? this.payload,
        filePath: filePath.present ? filePath.value : this.filePath,
        filePath2: filePath2.present ? filePath2.value : this.filePath2,
        createdAt: createdAt ?? this.createdAt,
        retryCount: retryCount ?? this.retryCount,
        status: status ?? this.status,
      );
  OfflineTask copyWithCompanion(OfflineTasksCompanion data) {
    return OfflineTask(
      id: data.id.present ? data.id.value : this.id,
      type: data.type.present ? data.type.value : this.type,
      payload: data.payload.present ? data.payload.value : this.payload,
      filePath: data.filePath.present ? data.filePath.value : this.filePath,
      filePath2: data.filePath2.present ? data.filePath2.value : this.filePath2,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      retryCount:
          data.retryCount.present ? data.retryCount.value : this.retryCount,
      status: data.status.present ? data.status.value : this.status,
    );
  }

  @override
  String toString() {
    return (StringBuffer('OfflineTask(')
          ..write('id: $id, ')
          ..write('type: $type, ')
          ..write('payload: $payload, ')
          ..write('filePath: $filePath, ')
          ..write('filePath2: $filePath2, ')
          ..write('createdAt: $createdAt, ')
          ..write('retryCount: $retryCount, ')
          ..write('status: $status')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
      id, type, payload, filePath, filePath2, createdAt, retryCount, status);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is OfflineTask &&
          other.id == this.id &&
          other.type == this.type &&
          other.payload == this.payload &&
          other.filePath == this.filePath &&
          other.filePath2 == this.filePath2 &&
          other.createdAt == this.createdAt &&
          other.retryCount == this.retryCount &&
          other.status == this.status);
}

class OfflineTasksCompanion extends UpdateCompanion<OfflineTask> {
  final Value<String> id;
  final Value<String> type;
  final Value<String> payload;
  final Value<String?> filePath;
  final Value<String?> filePath2;
  final Value<DateTime> createdAt;
  final Value<int> retryCount;
  final Value<String> status;
  final Value<int> rowid;
  const OfflineTasksCompanion({
    this.id = const Value.absent(),
    this.type = const Value.absent(),
    this.payload = const Value.absent(),
    this.filePath = const Value.absent(),
    this.filePath2 = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.retryCount = const Value.absent(),
    this.status = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  OfflineTasksCompanion.insert({
    required String id,
    required String type,
    required String payload,
    this.filePath = const Value.absent(),
    this.filePath2 = const Value.absent(),
    required DateTime createdAt,
    this.retryCount = const Value.absent(),
    this.status = const Value.absent(),
    this.rowid = const Value.absent(),
  })  : id = Value(id),
        type = Value(type),
        payload = Value(payload),
        createdAt = Value(createdAt);
  static Insertable<OfflineTask> custom({
    Expression<String>? id,
    Expression<String>? type,
    Expression<String>? payload,
    Expression<String>? filePath,
    Expression<String>? filePath2,
    Expression<DateTime>? createdAt,
    Expression<int>? retryCount,
    Expression<String>? status,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (type != null) 'type': type,
      if (payload != null) 'payload': payload,
      if (filePath != null) 'file_path': filePath,
      if (filePath2 != null) 'file_path2': filePath2,
      if (createdAt != null) 'created_at': createdAt,
      if (retryCount != null) 'retry_count': retryCount,
      if (status != null) 'status': status,
      if (rowid != null) 'rowid': rowid,
    });
  }

  OfflineTasksCompanion copyWith(
      {Value<String>? id,
      Value<String>? type,
      Value<String>? payload,
      Value<String?>? filePath,
      Value<String?>? filePath2,
      Value<DateTime>? createdAt,
      Value<int>? retryCount,
      Value<String>? status,
      Value<int>? rowid}) {
    return OfflineTasksCompanion(
      id: id ?? this.id,
      type: type ?? this.type,
      payload: payload ?? this.payload,
      filePath: filePath ?? this.filePath,
      filePath2: filePath2 ?? this.filePath2,
      createdAt: createdAt ?? this.createdAt,
      retryCount: retryCount ?? this.retryCount,
      status: status ?? this.status,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<String>(id.value);
    }
    if (type.present) {
      map['type'] = Variable<String>(type.value);
    }
    if (payload.present) {
      map['payload'] = Variable<String>(payload.value);
    }
    if (filePath.present) {
      map['file_path'] = Variable<String>(filePath.value);
    }
    if (filePath2.present) {
      map['file_path2'] = Variable<String>(filePath2.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (retryCount.present) {
      map['retry_count'] = Variable<int>(retryCount.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('OfflineTasksCompanion(')
          ..write('id: $id, ')
          ..write('type: $type, ')
          ..write('payload: $payload, ')
          ..write('filePath: $filePath, ')
          ..write('filePath2: $filePath2, ')
          ..write('createdAt: $createdAt, ')
          ..write('retryCount: $retryCount, ')
          ..write('status: $status, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $StoresTable extends Stores with TableInfo<$StoresTable, Store> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $StoresTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<String> id = GeneratedColumn<String>(
      'id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _codeMeta = const VerificationMeta('code');
  @override
  late final GeneratedColumn<String> code = GeneratedColumn<String>(
      'code', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _nameMeta = const VerificationMeta('name');
  @override
  late final GeneratedColumn<String> name = GeneratedColumn<String>(
      'name', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _addressMeta =
      const VerificationMeta('address');
  @override
  late final GeneratedColumn<String> address = GeneratedColumn<String>(
      'address', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _latitudeMeta =
      const VerificationMeta('latitude');
  @override
  late final GeneratedColumn<double> latitude = GeneratedColumn<double>(
      'latitude', aliasedName, true,
      type: DriftSqlType.double, requiredDuringInsert: false);
  static const VerificationMeta _longitudeMeta =
      const VerificationMeta('longitude');
  @override
  late final GeneratedColumn<double> longitude = GeneratedColumn<double>(
      'longitude', aliasedName, true,
      type: DriftSqlType.double, requiredDuringInsert: false);
  static const VerificationMeta _ownerNameMeta =
      const VerificationMeta('ownerName');
  @override
  late final GeneratedColumn<String> ownerName = GeneratedColumn<String>(
      'owner_name', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _phoneNumberMeta =
      const VerificationMeta('phoneNumber');
  @override
  late final GeneratedColumn<String> phoneNumber = GeneratedColumn<String>(
      'phone_number', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _photoUrlMeta =
      const VerificationMeta('photoUrl');
  @override
  late final GeneratedColumn<String> photoUrl = GeneratedColumn<String>(
      'photo_url', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _localPhotoPathMeta =
      const VerificationMeta('localPhotoPath');
  @override
  late final GeneratedColumn<String> localPhotoPath = GeneratedColumn<String>(
      'local_photo_path', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _provinceMeta =
      const VerificationMeta('province');
  @override
  late final GeneratedColumn<String> province = GeneratedColumn<String>(
      'province', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _regencyMeta =
      const VerificationMeta('regency');
  @override
  late final GeneratedColumn<String> regency = GeneratedColumn<String>(
      'regency', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _districtMeta =
      const VerificationMeta('district');
  @override
  late final GeneratedColumn<String> district = GeneratedColumn<String>(
      'district', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _villageMeta =
      const VerificationMeta('village');
  @override
  late final GeneratedColumn<String> village = GeneratedColumn<String>(
      'village', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _visitFrequencyMeta =
      const VerificationMeta('visitFrequency');
  @override
  late final GeneratedColumn<String> visitFrequency = GeneratedColumn<String>(
      'visit_frequency', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _visitDaysMeta =
      const VerificationMeta('visitDays');
  @override
  late final GeneratedColumn<String> visitDays = GeneratedColumn<String>(
      'visit_days', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _assignedEmployeeIdMeta =
      const VerificationMeta('assignedEmployeeId');
  @override
  late final GeneratedColumn<String> assignedEmployeeId =
      GeneratedColumn<String>('assigned_employee_id', aliasedName, true,
          type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _companyIdMeta =
      const VerificationMeta('companyId');
  @override
  late final GeneratedColumn<String> companyId = GeneratedColumn<String>(
      'company_id', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _isOfflineMeta =
      const VerificationMeta('isOffline');
  @override
  late final GeneratedColumn<bool> isOffline = GeneratedColumn<bool>(
      'is_offline', aliasedName, false,
      type: DriftSqlType.bool,
      requiredDuringInsert: false,
      defaultConstraints:
          GeneratedColumn.constraintIsAlways('CHECK ("is_offline" IN (0, 1))'),
      defaultValue: const Constant(false));
  static const VerificationMeta _syncStatusMeta =
      const VerificationMeta('syncStatus');
  @override
  late final GeneratedColumn<String> syncStatus = GeneratedColumn<String>(
      'sync_status', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('synced'));
  static const VerificationMeta _syncErrorMessageMeta =
      const VerificationMeta('syncErrorMessage');
  @override
  late final GeneratedColumn<String> syncErrorMessage = GeneratedColumn<String>(
      'sync_error_message', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _createdAtMeta =
      const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
      'created_at', aliasedName, true,
      type: DriftSqlType.dateTime, requiredDuringInsert: false);
  static const VerificationMeta _accuracyMeta =
      const VerificationMeta('accuracy');
  @override
  late final GeneratedColumn<double> accuracy = GeneratedColumn<double>(
      'accuracy', aliasedName, true,
      type: DriftSqlType.double, requiredDuringInsert: false);
  @override
  List<GeneratedColumn> get $columns => [
        id,
        code,
        name,
        address,
        latitude,
        longitude,
        ownerName,
        phoneNumber,
        photoUrl,
        localPhotoPath,
        province,
        regency,
        district,
        village,
        visitFrequency,
        visitDays,
        assignedEmployeeId,
        companyId,
        isOffline,
        syncStatus,
        syncErrorMessage,
        createdAt,
        accuracy
      ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'stores';
  @override
  VerificationContext validateIntegrity(Insertable<Store> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    } else if (isInserting) {
      context.missing(_idMeta);
    }
    if (data.containsKey('code')) {
      context.handle(
          _codeMeta, code.isAcceptableOrUnknown(data['code']!, _codeMeta));
    }
    if (data.containsKey('name')) {
      context.handle(
          _nameMeta, name.isAcceptableOrUnknown(data['name']!, _nameMeta));
    } else if (isInserting) {
      context.missing(_nameMeta);
    }
    if (data.containsKey('address')) {
      context.handle(_addressMeta,
          address.isAcceptableOrUnknown(data['address']!, _addressMeta));
    }
    if (data.containsKey('latitude')) {
      context.handle(_latitudeMeta,
          latitude.isAcceptableOrUnknown(data['latitude']!, _latitudeMeta));
    }
    if (data.containsKey('longitude')) {
      context.handle(_longitudeMeta,
          longitude.isAcceptableOrUnknown(data['longitude']!, _longitudeMeta));
    }
    if (data.containsKey('owner_name')) {
      context.handle(_ownerNameMeta,
          ownerName.isAcceptableOrUnknown(data['owner_name']!, _ownerNameMeta));
    }
    if (data.containsKey('phone_number')) {
      context.handle(
          _phoneNumberMeta,
          phoneNumber.isAcceptableOrUnknown(
              data['phone_number']!, _phoneNumberMeta));
    }
    if (data.containsKey('photo_url')) {
      context.handle(_photoUrlMeta,
          photoUrl.isAcceptableOrUnknown(data['photo_url']!, _photoUrlMeta));
    }
    if (data.containsKey('local_photo_path')) {
      context.handle(
          _localPhotoPathMeta,
          localPhotoPath.isAcceptableOrUnknown(
              data['local_photo_path']!, _localPhotoPathMeta));
    }
    if (data.containsKey('province')) {
      context.handle(_provinceMeta,
          province.isAcceptableOrUnknown(data['province']!, _provinceMeta));
    }
    if (data.containsKey('regency')) {
      context.handle(_regencyMeta,
          regency.isAcceptableOrUnknown(data['regency']!, _regencyMeta));
    }
    if (data.containsKey('district')) {
      context.handle(_districtMeta,
          district.isAcceptableOrUnknown(data['district']!, _districtMeta));
    }
    if (data.containsKey('village')) {
      context.handle(_villageMeta,
          village.isAcceptableOrUnknown(data['village']!, _villageMeta));
    }
    if (data.containsKey('visit_frequency')) {
      context.handle(
          _visitFrequencyMeta,
          visitFrequency.isAcceptableOrUnknown(
              data['visit_frequency']!, _visitFrequencyMeta));
    }
    if (data.containsKey('visit_days')) {
      context.handle(_visitDaysMeta,
          visitDays.isAcceptableOrUnknown(data['visit_days']!, _visitDaysMeta));
    }
    if (data.containsKey('assigned_employee_id')) {
      context.handle(
          _assignedEmployeeIdMeta,
          assignedEmployeeId.isAcceptableOrUnknown(
              data['assigned_employee_id']!, _assignedEmployeeIdMeta));
    }
    if (data.containsKey('company_id')) {
      context.handle(_companyIdMeta,
          companyId.isAcceptableOrUnknown(data['company_id']!, _companyIdMeta));
    }
    if (data.containsKey('is_offline')) {
      context.handle(_isOfflineMeta,
          isOffline.isAcceptableOrUnknown(data['is_offline']!, _isOfflineMeta));
    }
    if (data.containsKey('sync_status')) {
      context.handle(
          _syncStatusMeta,
          syncStatus.isAcceptableOrUnknown(
              data['sync_status']!, _syncStatusMeta));
    }
    if (data.containsKey('sync_error_message')) {
      context.handle(
          _syncErrorMessageMeta,
          syncErrorMessage.isAcceptableOrUnknown(
              data['sync_error_message']!, _syncErrorMessageMeta));
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta,
          createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    }
    if (data.containsKey('accuracy')) {
      context.handle(_accuracyMeta,
          accuracy.isAcceptableOrUnknown(data['accuracy']!, _accuracyMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  Store map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return Store(
      id: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}id'])!,
      code: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}code']),
      name: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}name'])!,
      address: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}address']),
      latitude: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}latitude']),
      longitude: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}longitude']),
      ownerName: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}owner_name']),
      phoneNumber: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}phone_number']),
      photoUrl: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}photo_url']),
      localPhotoPath: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}local_photo_path']),
      province: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}province']),
      regency: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}regency']),
      district: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}district']),
      village: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}village']),
      visitFrequency: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}visit_frequency']),
      visitDays: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}visit_days']),
      assignedEmployeeId: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}assigned_employee_id']),
      companyId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}company_id']),
      isOffline: attachedDatabase.typeMapping
          .read(DriftSqlType.bool, data['${effectivePrefix}is_offline'])!,
      syncStatus: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}sync_status'])!,
      syncErrorMessage: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}sync_error_message']),
      createdAt: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}created_at']),
      accuracy: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}accuracy']),
    );
  }

  @override
  $StoresTable createAlias(String alias) {
    return $StoresTable(attachedDatabase, alias);
  }
}

class Store extends DataClass implements Insertable<Store> {
  final String id;
  final String? code;
  final String name;
  final String? address;
  final double? latitude;
  final double? longitude;
  final String? ownerName;
  final String? phoneNumber;
  final String? photoUrl;
  final String? localPhotoPath;
  final String? province;
  final String? regency;
  final String? district;
  final String? village;
  final String? visitFrequency;
  final String? visitDays;
  final String? assignedEmployeeId;
  final String? companyId;
  final bool isOffline;
  final String syncStatus;
  final String? syncErrorMessage;
  final DateTime? createdAt;
  final double? accuracy;
  const Store(
      {required this.id,
      this.code,
      required this.name,
      this.address,
      this.latitude,
      this.longitude,
      this.ownerName,
      this.phoneNumber,
      this.photoUrl,
      this.localPhotoPath,
      this.province,
      this.regency,
      this.district,
      this.village,
      this.visitFrequency,
      this.visitDays,
      this.assignedEmployeeId,
      this.companyId,
      required this.isOffline,
      required this.syncStatus,
      this.syncErrorMessage,
      this.createdAt,
      this.accuracy});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<String>(id);
    if (!nullToAbsent || code != null) {
      map['code'] = Variable<String>(code);
    }
    map['name'] = Variable<String>(name);
    if (!nullToAbsent || address != null) {
      map['address'] = Variable<String>(address);
    }
    if (!nullToAbsent || latitude != null) {
      map['latitude'] = Variable<double>(latitude);
    }
    if (!nullToAbsent || longitude != null) {
      map['longitude'] = Variable<double>(longitude);
    }
    if (!nullToAbsent || ownerName != null) {
      map['owner_name'] = Variable<String>(ownerName);
    }
    if (!nullToAbsent || phoneNumber != null) {
      map['phone_number'] = Variable<String>(phoneNumber);
    }
    if (!nullToAbsent || photoUrl != null) {
      map['photo_url'] = Variable<String>(photoUrl);
    }
    if (!nullToAbsent || localPhotoPath != null) {
      map['local_photo_path'] = Variable<String>(localPhotoPath);
    }
    if (!nullToAbsent || province != null) {
      map['province'] = Variable<String>(province);
    }
    if (!nullToAbsent || regency != null) {
      map['regency'] = Variable<String>(regency);
    }
    if (!nullToAbsent || district != null) {
      map['district'] = Variable<String>(district);
    }
    if (!nullToAbsent || village != null) {
      map['village'] = Variable<String>(village);
    }
    if (!nullToAbsent || visitFrequency != null) {
      map['visit_frequency'] = Variable<String>(visitFrequency);
    }
    if (!nullToAbsent || visitDays != null) {
      map['visit_days'] = Variable<String>(visitDays);
    }
    if (!nullToAbsent || assignedEmployeeId != null) {
      map['assigned_employee_id'] = Variable<String>(assignedEmployeeId);
    }
    if (!nullToAbsent || companyId != null) {
      map['company_id'] = Variable<String>(companyId);
    }
    map['is_offline'] = Variable<bool>(isOffline);
    map['sync_status'] = Variable<String>(syncStatus);
    if (!nullToAbsent || syncErrorMessage != null) {
      map['sync_error_message'] = Variable<String>(syncErrorMessage);
    }
    if (!nullToAbsent || createdAt != null) {
      map['created_at'] = Variable<DateTime>(createdAt);
    }
    if (!nullToAbsent || accuracy != null) {
      map['accuracy'] = Variable<double>(accuracy);
    }
    return map;
  }

  StoresCompanion toCompanion(bool nullToAbsent) {
    return StoresCompanion(
      id: Value(id),
      code: code == null && nullToAbsent ? const Value.absent() : Value(code),
      name: Value(name),
      address: address == null && nullToAbsent
          ? const Value.absent()
          : Value(address),
      latitude: latitude == null && nullToAbsent
          ? const Value.absent()
          : Value(latitude),
      longitude: longitude == null && nullToAbsent
          ? const Value.absent()
          : Value(longitude),
      ownerName: ownerName == null && nullToAbsent
          ? const Value.absent()
          : Value(ownerName),
      phoneNumber: phoneNumber == null && nullToAbsent
          ? const Value.absent()
          : Value(phoneNumber),
      photoUrl: photoUrl == null && nullToAbsent
          ? const Value.absent()
          : Value(photoUrl),
      localPhotoPath: localPhotoPath == null && nullToAbsent
          ? const Value.absent()
          : Value(localPhotoPath),
      province: province == null && nullToAbsent
          ? const Value.absent()
          : Value(province),
      regency: regency == null && nullToAbsent
          ? const Value.absent()
          : Value(regency),
      district: district == null && nullToAbsent
          ? const Value.absent()
          : Value(district),
      village: village == null && nullToAbsent
          ? const Value.absent()
          : Value(village),
      visitFrequency: visitFrequency == null && nullToAbsent
          ? const Value.absent()
          : Value(visitFrequency),
      visitDays: visitDays == null && nullToAbsent
          ? const Value.absent()
          : Value(visitDays),
      assignedEmployeeId: assignedEmployeeId == null && nullToAbsent
          ? const Value.absent()
          : Value(assignedEmployeeId),
      companyId: companyId == null && nullToAbsent
          ? const Value.absent()
          : Value(companyId),
      isOffline: Value(isOffline),
      syncStatus: Value(syncStatus),
      syncErrorMessage: syncErrorMessage == null && nullToAbsent
          ? const Value.absent()
          : Value(syncErrorMessage),
      createdAt: createdAt == null && nullToAbsent
          ? const Value.absent()
          : Value(createdAt),
      accuracy: accuracy == null && nullToAbsent
          ? const Value.absent()
          : Value(accuracy),
    );
  }

  factory Store.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return Store(
      id: serializer.fromJson<String>(json['id']),
      code: serializer.fromJson<String?>(json['code']),
      name: serializer.fromJson<String>(json['name']),
      address: serializer.fromJson<String?>(json['address']),
      latitude: serializer.fromJson<double?>(json['latitude']),
      longitude: serializer.fromJson<double?>(json['longitude']),
      ownerName: serializer.fromJson<String?>(json['ownerName']),
      phoneNumber: serializer.fromJson<String?>(json['phoneNumber']),
      photoUrl: serializer.fromJson<String?>(json['photoUrl']),
      localPhotoPath: serializer.fromJson<String?>(json['localPhotoPath']),
      province: serializer.fromJson<String?>(json['province']),
      regency: serializer.fromJson<String?>(json['regency']),
      district: serializer.fromJson<String?>(json['district']),
      village: serializer.fromJson<String?>(json['village']),
      visitFrequency: serializer.fromJson<String?>(json['visitFrequency']),
      visitDays: serializer.fromJson<String?>(json['visitDays']),
      assignedEmployeeId:
          serializer.fromJson<String?>(json['assignedEmployeeId']),
      companyId: serializer.fromJson<String?>(json['companyId']),
      isOffline: serializer.fromJson<bool>(json['isOffline']),
      syncStatus: serializer.fromJson<String>(json['syncStatus']),
      syncErrorMessage: serializer.fromJson<String?>(json['syncErrorMessage']),
      createdAt: serializer.fromJson<DateTime?>(json['createdAt']),
      accuracy: serializer.fromJson<double?>(json['accuracy']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<String>(id),
      'code': serializer.toJson<String?>(code),
      'name': serializer.toJson<String>(name),
      'address': serializer.toJson<String?>(address),
      'latitude': serializer.toJson<double?>(latitude),
      'longitude': serializer.toJson<double?>(longitude),
      'ownerName': serializer.toJson<String?>(ownerName),
      'phoneNumber': serializer.toJson<String?>(phoneNumber),
      'photoUrl': serializer.toJson<String?>(photoUrl),
      'localPhotoPath': serializer.toJson<String?>(localPhotoPath),
      'province': serializer.toJson<String?>(province),
      'regency': serializer.toJson<String?>(regency),
      'district': serializer.toJson<String?>(district),
      'village': serializer.toJson<String?>(village),
      'visitFrequency': serializer.toJson<String?>(visitFrequency),
      'visitDays': serializer.toJson<String?>(visitDays),
      'assignedEmployeeId': serializer.toJson<String?>(assignedEmployeeId),
      'companyId': serializer.toJson<String?>(companyId),
      'isOffline': serializer.toJson<bool>(isOffline),
      'syncStatus': serializer.toJson<String>(syncStatus),
      'syncErrorMessage': serializer.toJson<String?>(syncErrorMessage),
      'createdAt': serializer.toJson<DateTime?>(createdAt),
      'accuracy': serializer.toJson<double?>(accuracy),
    };
  }

  Store copyWith(
          {String? id,
          Value<String?> code = const Value.absent(),
          String? name,
          Value<String?> address = const Value.absent(),
          Value<double?> latitude = const Value.absent(),
          Value<double?> longitude = const Value.absent(),
          Value<String?> ownerName = const Value.absent(),
          Value<String?> phoneNumber = const Value.absent(),
          Value<String?> photoUrl = const Value.absent(),
          Value<String?> localPhotoPath = const Value.absent(),
          Value<String?> province = const Value.absent(),
          Value<String?> regency = const Value.absent(),
          Value<String?> district = const Value.absent(),
          Value<String?> village = const Value.absent(),
          Value<String?> visitFrequency = const Value.absent(),
          Value<String?> visitDays = const Value.absent(),
          Value<String?> assignedEmployeeId = const Value.absent(),
          Value<String?> companyId = const Value.absent(),
          bool? isOffline,
          String? syncStatus,
          Value<String?> syncErrorMessage = const Value.absent(),
          Value<DateTime?> createdAt = const Value.absent(),
          Value<double?> accuracy = const Value.absent()}) =>
      Store(
        id: id ?? this.id,
        code: code.present ? code.value : this.code,
        name: name ?? this.name,
        address: address.present ? address.value : this.address,
        latitude: latitude.present ? latitude.value : this.latitude,
        longitude: longitude.present ? longitude.value : this.longitude,
        ownerName: ownerName.present ? ownerName.value : this.ownerName,
        phoneNumber: phoneNumber.present ? phoneNumber.value : this.phoneNumber,
        photoUrl: photoUrl.present ? photoUrl.value : this.photoUrl,
        localPhotoPath:
            localPhotoPath.present ? localPhotoPath.value : this.localPhotoPath,
        province: province.present ? province.value : this.province,
        regency: regency.present ? regency.value : this.regency,
        district: district.present ? district.value : this.district,
        village: village.present ? village.value : this.village,
        visitFrequency:
            visitFrequency.present ? visitFrequency.value : this.visitFrequency,
        visitDays: visitDays.present ? visitDays.value : this.visitDays,
        assignedEmployeeId: assignedEmployeeId.present
            ? assignedEmployeeId.value
            : this.assignedEmployeeId,
        companyId: companyId.present ? companyId.value : this.companyId,
        isOffline: isOffline ?? this.isOffline,
        syncStatus: syncStatus ?? this.syncStatus,
        syncErrorMessage: syncErrorMessage.present
            ? syncErrorMessage.value
            : this.syncErrorMessage,
        createdAt: createdAt.present ? createdAt.value : this.createdAt,
        accuracy: accuracy.present ? accuracy.value : this.accuracy,
      );
  Store copyWithCompanion(StoresCompanion data) {
    return Store(
      id: data.id.present ? data.id.value : this.id,
      code: data.code.present ? data.code.value : this.code,
      name: data.name.present ? data.name.value : this.name,
      address: data.address.present ? data.address.value : this.address,
      latitude: data.latitude.present ? data.latitude.value : this.latitude,
      longitude: data.longitude.present ? data.longitude.value : this.longitude,
      ownerName: data.ownerName.present ? data.ownerName.value : this.ownerName,
      phoneNumber:
          data.phoneNumber.present ? data.phoneNumber.value : this.phoneNumber,
      photoUrl: data.photoUrl.present ? data.photoUrl.value : this.photoUrl,
      localPhotoPath: data.localPhotoPath.present
          ? data.localPhotoPath.value
          : this.localPhotoPath,
      province: data.province.present ? data.province.value : this.province,
      regency: data.regency.present ? data.regency.value : this.regency,
      district: data.district.present ? data.district.value : this.district,
      village: data.village.present ? data.village.value : this.village,
      visitFrequency: data.visitFrequency.present
          ? data.visitFrequency.value
          : this.visitFrequency,
      visitDays: data.visitDays.present ? data.visitDays.value : this.visitDays,
      assignedEmployeeId: data.assignedEmployeeId.present
          ? data.assignedEmployeeId.value
          : this.assignedEmployeeId,
      companyId: data.companyId.present ? data.companyId.value : this.companyId,
      isOffline: data.isOffline.present ? data.isOffline.value : this.isOffline,
      syncStatus:
          data.syncStatus.present ? data.syncStatus.value : this.syncStatus,
      syncErrorMessage: data.syncErrorMessage.present
          ? data.syncErrorMessage.value
          : this.syncErrorMessage,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      accuracy: data.accuracy.present ? data.accuracy.value : this.accuracy,
    );
  }

  @override
  String toString() {
    return (StringBuffer('Store(')
          ..write('id: $id, ')
          ..write('code: $code, ')
          ..write('name: $name, ')
          ..write('address: $address, ')
          ..write('latitude: $latitude, ')
          ..write('longitude: $longitude, ')
          ..write('ownerName: $ownerName, ')
          ..write('phoneNumber: $phoneNumber, ')
          ..write('photoUrl: $photoUrl, ')
          ..write('localPhotoPath: $localPhotoPath, ')
          ..write('province: $province, ')
          ..write('regency: $regency, ')
          ..write('district: $district, ')
          ..write('village: $village, ')
          ..write('visitFrequency: $visitFrequency, ')
          ..write('visitDays: $visitDays, ')
          ..write('assignedEmployeeId: $assignedEmployeeId, ')
          ..write('companyId: $companyId, ')
          ..write('isOffline: $isOffline, ')
          ..write('syncStatus: $syncStatus, ')
          ..write('syncErrorMessage: $syncErrorMessage, ')
          ..write('createdAt: $createdAt, ')
          ..write('accuracy: $accuracy')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hashAll([
        id,
        code,
        name,
        address,
        latitude,
        longitude,
        ownerName,
        phoneNumber,
        photoUrl,
        localPhotoPath,
        province,
        regency,
        district,
        village,
        visitFrequency,
        visitDays,
        assignedEmployeeId,
        companyId,
        isOffline,
        syncStatus,
        syncErrorMessage,
        createdAt,
        accuracy
      ]);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is Store &&
          other.id == this.id &&
          other.code == this.code &&
          other.name == this.name &&
          other.address == this.address &&
          other.latitude == this.latitude &&
          other.longitude == this.longitude &&
          other.ownerName == this.ownerName &&
          other.phoneNumber == this.phoneNumber &&
          other.photoUrl == this.photoUrl &&
          other.localPhotoPath == this.localPhotoPath &&
          other.province == this.province &&
          other.regency == this.regency &&
          other.district == this.district &&
          other.village == this.village &&
          other.visitFrequency == this.visitFrequency &&
          other.visitDays == this.visitDays &&
          other.assignedEmployeeId == this.assignedEmployeeId &&
          other.companyId == this.companyId &&
          other.isOffline == this.isOffline &&
          other.syncStatus == this.syncStatus &&
          other.syncErrorMessage == this.syncErrorMessage &&
          other.createdAt == this.createdAt &&
          other.accuracy == this.accuracy);
}

class StoresCompanion extends UpdateCompanion<Store> {
  final Value<String> id;
  final Value<String?> code;
  final Value<String> name;
  final Value<String?> address;
  final Value<double?> latitude;
  final Value<double?> longitude;
  final Value<String?> ownerName;
  final Value<String?> phoneNumber;
  final Value<String?> photoUrl;
  final Value<String?> localPhotoPath;
  final Value<String?> province;
  final Value<String?> regency;
  final Value<String?> district;
  final Value<String?> village;
  final Value<String?> visitFrequency;
  final Value<String?> visitDays;
  final Value<String?> assignedEmployeeId;
  final Value<String?> companyId;
  final Value<bool> isOffline;
  final Value<String> syncStatus;
  final Value<String?> syncErrorMessage;
  final Value<DateTime?> createdAt;
  final Value<double?> accuracy;
  final Value<int> rowid;
  const StoresCompanion({
    this.id = const Value.absent(),
    this.code = const Value.absent(),
    this.name = const Value.absent(),
    this.address = const Value.absent(),
    this.latitude = const Value.absent(),
    this.longitude = const Value.absent(),
    this.ownerName = const Value.absent(),
    this.phoneNumber = const Value.absent(),
    this.photoUrl = const Value.absent(),
    this.localPhotoPath = const Value.absent(),
    this.province = const Value.absent(),
    this.regency = const Value.absent(),
    this.district = const Value.absent(),
    this.village = const Value.absent(),
    this.visitFrequency = const Value.absent(),
    this.visitDays = const Value.absent(),
    this.assignedEmployeeId = const Value.absent(),
    this.companyId = const Value.absent(),
    this.isOffline = const Value.absent(),
    this.syncStatus = const Value.absent(),
    this.syncErrorMessage = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.accuracy = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  StoresCompanion.insert({
    required String id,
    this.code = const Value.absent(),
    required String name,
    this.address = const Value.absent(),
    this.latitude = const Value.absent(),
    this.longitude = const Value.absent(),
    this.ownerName = const Value.absent(),
    this.phoneNumber = const Value.absent(),
    this.photoUrl = const Value.absent(),
    this.localPhotoPath = const Value.absent(),
    this.province = const Value.absent(),
    this.regency = const Value.absent(),
    this.district = const Value.absent(),
    this.village = const Value.absent(),
    this.visitFrequency = const Value.absent(),
    this.visitDays = const Value.absent(),
    this.assignedEmployeeId = const Value.absent(),
    this.companyId = const Value.absent(),
    this.isOffline = const Value.absent(),
    this.syncStatus = const Value.absent(),
    this.syncErrorMessage = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.accuracy = const Value.absent(),
    this.rowid = const Value.absent(),
  })  : id = Value(id),
        name = Value(name);
  static Insertable<Store> custom({
    Expression<String>? id,
    Expression<String>? code,
    Expression<String>? name,
    Expression<String>? address,
    Expression<double>? latitude,
    Expression<double>? longitude,
    Expression<String>? ownerName,
    Expression<String>? phoneNumber,
    Expression<String>? photoUrl,
    Expression<String>? localPhotoPath,
    Expression<String>? province,
    Expression<String>? regency,
    Expression<String>? district,
    Expression<String>? village,
    Expression<String>? visitFrequency,
    Expression<String>? visitDays,
    Expression<String>? assignedEmployeeId,
    Expression<String>? companyId,
    Expression<bool>? isOffline,
    Expression<String>? syncStatus,
    Expression<String>? syncErrorMessage,
    Expression<DateTime>? createdAt,
    Expression<double>? accuracy,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (code != null) 'code': code,
      if (name != null) 'name': name,
      if (address != null) 'address': address,
      if (latitude != null) 'latitude': latitude,
      if (longitude != null) 'longitude': longitude,
      if (ownerName != null) 'owner_name': ownerName,
      if (phoneNumber != null) 'phone_number': phoneNumber,
      if (photoUrl != null) 'photo_url': photoUrl,
      if (localPhotoPath != null) 'local_photo_path': localPhotoPath,
      if (province != null) 'province': province,
      if (regency != null) 'regency': regency,
      if (district != null) 'district': district,
      if (village != null) 'village': village,
      if (visitFrequency != null) 'visit_frequency': visitFrequency,
      if (visitDays != null) 'visit_days': visitDays,
      if (assignedEmployeeId != null)
        'assigned_employee_id': assignedEmployeeId,
      if (companyId != null) 'company_id': companyId,
      if (isOffline != null) 'is_offline': isOffline,
      if (syncStatus != null) 'sync_status': syncStatus,
      if (syncErrorMessage != null) 'sync_error_message': syncErrorMessage,
      if (createdAt != null) 'created_at': createdAt,
      if (accuracy != null) 'accuracy': accuracy,
      if (rowid != null) 'rowid': rowid,
    });
  }

  StoresCompanion copyWith(
      {Value<String>? id,
      Value<String?>? code,
      Value<String>? name,
      Value<String?>? address,
      Value<double?>? latitude,
      Value<double?>? longitude,
      Value<String?>? ownerName,
      Value<String?>? phoneNumber,
      Value<String?>? photoUrl,
      Value<String?>? localPhotoPath,
      Value<String?>? province,
      Value<String?>? regency,
      Value<String?>? district,
      Value<String?>? village,
      Value<String?>? visitFrequency,
      Value<String?>? visitDays,
      Value<String?>? assignedEmployeeId,
      Value<String?>? companyId,
      Value<bool>? isOffline,
      Value<String>? syncStatus,
      Value<String?>? syncErrorMessage,
      Value<DateTime?>? createdAt,
      Value<double?>? accuracy,
      Value<int>? rowid}) {
    return StoresCompanion(
      id: id ?? this.id,
      code: code ?? this.code,
      name: name ?? this.name,
      address: address ?? this.address,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
      ownerName: ownerName ?? this.ownerName,
      phoneNumber: phoneNumber ?? this.phoneNumber,
      photoUrl: photoUrl ?? this.photoUrl,
      localPhotoPath: localPhotoPath ?? this.localPhotoPath,
      province: province ?? this.province,
      regency: regency ?? this.regency,
      district: district ?? this.district,
      village: village ?? this.village,
      visitFrequency: visitFrequency ?? this.visitFrequency,
      visitDays: visitDays ?? this.visitDays,
      assignedEmployeeId: assignedEmployeeId ?? this.assignedEmployeeId,
      companyId: companyId ?? this.companyId,
      isOffline: isOffline ?? this.isOffline,
      syncStatus: syncStatus ?? this.syncStatus,
      syncErrorMessage: syncErrorMessage ?? this.syncErrorMessage,
      createdAt: createdAt ?? this.createdAt,
      accuracy: accuracy ?? this.accuracy,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<String>(id.value);
    }
    if (code.present) {
      map['code'] = Variable<String>(code.value);
    }
    if (name.present) {
      map['name'] = Variable<String>(name.value);
    }
    if (address.present) {
      map['address'] = Variable<String>(address.value);
    }
    if (latitude.present) {
      map['latitude'] = Variable<double>(latitude.value);
    }
    if (longitude.present) {
      map['longitude'] = Variable<double>(longitude.value);
    }
    if (ownerName.present) {
      map['owner_name'] = Variable<String>(ownerName.value);
    }
    if (phoneNumber.present) {
      map['phone_number'] = Variable<String>(phoneNumber.value);
    }
    if (photoUrl.present) {
      map['photo_url'] = Variable<String>(photoUrl.value);
    }
    if (localPhotoPath.present) {
      map['local_photo_path'] = Variable<String>(localPhotoPath.value);
    }
    if (province.present) {
      map['province'] = Variable<String>(province.value);
    }
    if (regency.present) {
      map['regency'] = Variable<String>(regency.value);
    }
    if (district.present) {
      map['district'] = Variable<String>(district.value);
    }
    if (village.present) {
      map['village'] = Variable<String>(village.value);
    }
    if (visitFrequency.present) {
      map['visit_frequency'] = Variable<String>(visitFrequency.value);
    }
    if (visitDays.present) {
      map['visit_days'] = Variable<String>(visitDays.value);
    }
    if (assignedEmployeeId.present) {
      map['assigned_employee_id'] = Variable<String>(assignedEmployeeId.value);
    }
    if (companyId.present) {
      map['company_id'] = Variable<String>(companyId.value);
    }
    if (isOffline.present) {
      map['is_offline'] = Variable<bool>(isOffline.value);
    }
    if (syncStatus.present) {
      map['sync_status'] = Variable<String>(syncStatus.value);
    }
    if (syncErrorMessage.present) {
      map['sync_error_message'] = Variable<String>(syncErrorMessage.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (accuracy.present) {
      map['accuracy'] = Variable<double>(accuracy.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('StoresCompanion(')
          ..write('id: $id, ')
          ..write('code: $code, ')
          ..write('name: $name, ')
          ..write('address: $address, ')
          ..write('latitude: $latitude, ')
          ..write('longitude: $longitude, ')
          ..write('ownerName: $ownerName, ')
          ..write('phoneNumber: $phoneNumber, ')
          ..write('photoUrl: $photoUrl, ')
          ..write('localPhotoPath: $localPhotoPath, ')
          ..write('province: $province, ')
          ..write('regency: $regency, ')
          ..write('district: $district, ')
          ..write('village: $village, ')
          ..write('visitFrequency: $visitFrequency, ')
          ..write('visitDays: $visitDays, ')
          ..write('assignedEmployeeId: $assignedEmployeeId, ')
          ..write('companyId: $companyId, ')
          ..write('isOffline: $isOffline, ')
          ..write('syncStatus: $syncStatus, ')
          ..write('syncErrorMessage: $syncErrorMessage, ')
          ..write('createdAt: $createdAt, ')
          ..write('accuracy: $accuracy, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $VisitPlansTable extends VisitPlans
    with TableInfo<$VisitPlansTable, VisitPlan> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $VisitPlansTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<String> id = GeneratedColumn<String>(
      'id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _storeIdMeta =
      const VerificationMeta('storeId');
  @override
  late final GeneratedColumn<String> storeId = GeneratedColumn<String>(
      'store_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _storeNameMeta =
      const VerificationMeta('storeName');
  @override
  late final GeneratedColumn<String> storeName = GeneratedColumn<String>(
      'store_name', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _dateMeta = const VerificationMeta('date');
  @override
  late final GeneratedColumn<String> date = GeneratedColumn<String>(
      'date', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
      'status', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('pending'));
  @override
  List<GeneratedColumn> get $columns => [id, storeId, storeName, date, status];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'visit_plans';
  @override
  VerificationContext validateIntegrity(Insertable<VisitPlan> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    } else if (isInserting) {
      context.missing(_idMeta);
    }
    if (data.containsKey('store_id')) {
      context.handle(_storeIdMeta,
          storeId.isAcceptableOrUnknown(data['store_id']!, _storeIdMeta));
    } else if (isInserting) {
      context.missing(_storeIdMeta);
    }
    if (data.containsKey('store_name')) {
      context.handle(_storeNameMeta,
          storeName.isAcceptableOrUnknown(data['store_name']!, _storeNameMeta));
    } else if (isInserting) {
      context.missing(_storeNameMeta);
    }
    if (data.containsKey('date')) {
      context.handle(
          _dateMeta, date.isAcceptableOrUnknown(data['date']!, _dateMeta));
    } else if (isInserting) {
      context.missing(_dateMeta);
    }
    if (data.containsKey('status')) {
      context.handle(_statusMeta,
          status.isAcceptableOrUnknown(data['status']!, _statusMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  VisitPlan map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return VisitPlan(
      id: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}id'])!,
      storeId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}store_id'])!,
      storeName: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}store_name'])!,
      date: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}date'])!,
      status: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}status'])!,
    );
  }

  @override
  $VisitPlansTable createAlias(String alias) {
    return $VisitPlansTable(attachedDatabase, alias);
  }
}

class VisitPlan extends DataClass implements Insertable<VisitPlan> {
  final String id;
  final String storeId;
  final String storeName;
  final String date;
  final String status;
  const VisitPlan(
      {required this.id,
      required this.storeId,
      required this.storeName,
      required this.date,
      required this.status});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<String>(id);
    map['store_id'] = Variable<String>(storeId);
    map['store_name'] = Variable<String>(storeName);
    map['date'] = Variable<String>(date);
    map['status'] = Variable<String>(status);
    return map;
  }

  VisitPlansCompanion toCompanion(bool nullToAbsent) {
    return VisitPlansCompanion(
      id: Value(id),
      storeId: Value(storeId),
      storeName: Value(storeName),
      date: Value(date),
      status: Value(status),
    );
  }

  factory VisitPlan.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return VisitPlan(
      id: serializer.fromJson<String>(json['id']),
      storeId: serializer.fromJson<String>(json['storeId']),
      storeName: serializer.fromJson<String>(json['storeName']),
      date: serializer.fromJson<String>(json['date']),
      status: serializer.fromJson<String>(json['status']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<String>(id),
      'storeId': serializer.toJson<String>(storeId),
      'storeName': serializer.toJson<String>(storeName),
      'date': serializer.toJson<String>(date),
      'status': serializer.toJson<String>(status),
    };
  }

  VisitPlan copyWith(
          {String? id,
          String? storeId,
          String? storeName,
          String? date,
          String? status}) =>
      VisitPlan(
        id: id ?? this.id,
        storeId: storeId ?? this.storeId,
        storeName: storeName ?? this.storeName,
        date: date ?? this.date,
        status: status ?? this.status,
      );
  VisitPlan copyWithCompanion(VisitPlansCompanion data) {
    return VisitPlan(
      id: data.id.present ? data.id.value : this.id,
      storeId: data.storeId.present ? data.storeId.value : this.storeId,
      storeName: data.storeName.present ? data.storeName.value : this.storeName,
      date: data.date.present ? data.date.value : this.date,
      status: data.status.present ? data.status.value : this.status,
    );
  }

  @override
  String toString() {
    return (StringBuffer('VisitPlan(')
          ..write('id: $id, ')
          ..write('storeId: $storeId, ')
          ..write('storeName: $storeName, ')
          ..write('date: $date, ')
          ..write('status: $status')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(id, storeId, storeName, date, status);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is VisitPlan &&
          other.id == this.id &&
          other.storeId == this.storeId &&
          other.storeName == this.storeName &&
          other.date == this.date &&
          other.status == this.status);
}

class VisitPlansCompanion extends UpdateCompanion<VisitPlan> {
  final Value<String> id;
  final Value<String> storeId;
  final Value<String> storeName;
  final Value<String> date;
  final Value<String> status;
  final Value<int> rowid;
  const VisitPlansCompanion({
    this.id = const Value.absent(),
    this.storeId = const Value.absent(),
    this.storeName = const Value.absent(),
    this.date = const Value.absent(),
    this.status = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  VisitPlansCompanion.insert({
    required String id,
    required String storeId,
    required String storeName,
    required String date,
    this.status = const Value.absent(),
    this.rowid = const Value.absent(),
  })  : id = Value(id),
        storeId = Value(storeId),
        storeName = Value(storeName),
        date = Value(date);
  static Insertable<VisitPlan> custom({
    Expression<String>? id,
    Expression<String>? storeId,
    Expression<String>? storeName,
    Expression<String>? date,
    Expression<String>? status,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (storeId != null) 'store_id': storeId,
      if (storeName != null) 'store_name': storeName,
      if (date != null) 'date': date,
      if (status != null) 'status': status,
      if (rowid != null) 'rowid': rowid,
    });
  }

  VisitPlansCompanion copyWith(
      {Value<String>? id,
      Value<String>? storeId,
      Value<String>? storeName,
      Value<String>? date,
      Value<String>? status,
      Value<int>? rowid}) {
    return VisitPlansCompanion(
      id: id ?? this.id,
      storeId: storeId ?? this.storeId,
      storeName: storeName ?? this.storeName,
      date: date ?? this.date,
      status: status ?? this.status,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<String>(id.value);
    }
    if (storeId.present) {
      map['store_id'] = Variable<String>(storeId.value);
    }
    if (storeName.present) {
      map['store_name'] = Variable<String>(storeName.value);
    }
    if (date.present) {
      map['date'] = Variable<String>(date.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('VisitPlansCompanion(')
          ..write('id: $id, ')
          ..write('storeId: $storeId, ')
          ..write('storeName: $storeName, ')
          ..write('date: $date, ')
          ..write('status: $status, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $SalesDailyJourneysTable extends SalesDailyJourneys
    with TableInfo<$SalesDailyJourneysTable, SalesDailyJourney> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $SalesDailyJourneysTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _idMeta = const VerificationMeta('id');
  @override
  late final GeneratedColumn<String> id = GeneratedColumn<String>(
      'id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _userIdMeta = const VerificationMeta('userId');
  @override
  late final GeneratedColumn<String> userId = GeneratedColumn<String>(
      'user_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _dateMeta = const VerificationMeta('date');
  @override
  late final GeneratedColumn<String> date = GeneratedColumn<String>(
      'date', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _startKmMeta =
      const VerificationMeta('startKm');
  @override
  late final GeneratedColumn<int> startKm = GeneratedColumn<int>(
      'start_km', aliasedName, false,
      type: DriftSqlType.int, requiredDuringInsert: true);
  static const VerificationMeta _startPhotoUrlMeta =
      const VerificationMeta('startPhotoUrl');
  @override
  late final GeneratedColumn<String> startPhotoUrl = GeneratedColumn<String>(
      'start_photo_url', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _startLocationMeta =
      const VerificationMeta('startLocation');
  @override
  late final GeneratedColumn<String> startLocation = GeneratedColumn<String>(
      'start_location', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _startTimeMeta =
      const VerificationMeta('startTime');
  @override
  late final GeneratedColumn<DateTime> startTime = GeneratedColumn<DateTime>(
      'start_time', aliasedName, false,
      type: DriftSqlType.dateTime, requiredDuringInsert: true);
  static const VerificationMeta _endKmMeta = const VerificationMeta('endKm');
  @override
  late final GeneratedColumn<int> endKm = GeneratedColumn<int>(
      'end_km', aliasedName, true,
      type: DriftSqlType.int, requiredDuringInsert: false);
  static const VerificationMeta _endPhotoUrlMeta =
      const VerificationMeta('endPhotoUrl');
  @override
  late final GeneratedColumn<String> endPhotoUrl = GeneratedColumn<String>(
      'end_photo_url', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _endLocationMeta =
      const VerificationMeta('endLocation');
  @override
  late final GeneratedColumn<String> endLocation = GeneratedColumn<String>(
      'end_location', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _endTimeMeta =
      const VerificationMeta('endTime');
  @override
  late final GeneratedColumn<DateTime> endTime = GeneratedColumn<DateTime>(
      'end_time', aliasedName, true,
      type: DriftSqlType.dateTime, requiredDuringInsert: false);
  static const VerificationMeta _totalKmMeta =
      const VerificationMeta('totalKm');
  @override
  late final GeneratedColumn<int> totalKm = GeneratedColumn<int>(
      'total_km', aliasedName, true,
      type: DriftSqlType.int, requiredDuringInsert: false);
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
      'status', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('started'));
  static const VerificationMeta _isSyncedMeta =
      const VerificationMeta('isSynced');
  @override
  late final GeneratedColumn<bool> isSynced = GeneratedColumn<bool>(
      'is_synced', aliasedName, false,
      type: DriftSqlType.bool,
      requiredDuringInsert: false,
      defaultConstraints:
          GeneratedColumn.constraintIsAlways('CHECK ("is_synced" IN (0, 1))'),
      defaultValue: const Constant(false));
  @override
  List<GeneratedColumn> get $columns => [
        id,
        userId,
        date,
        startKm,
        startPhotoUrl,
        startLocation,
        startTime,
        endKm,
        endPhotoUrl,
        endLocation,
        endTime,
        totalKm,
        status,
        isSynced
      ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'sales_daily_journeys';
  @override
  VerificationContext validateIntegrity(Insertable<SalesDailyJourney> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('id')) {
      context.handle(_idMeta, id.isAcceptableOrUnknown(data['id']!, _idMeta));
    } else if (isInserting) {
      context.missing(_idMeta);
    }
    if (data.containsKey('user_id')) {
      context.handle(_userIdMeta,
          userId.isAcceptableOrUnknown(data['user_id']!, _userIdMeta));
    } else if (isInserting) {
      context.missing(_userIdMeta);
    }
    if (data.containsKey('date')) {
      context.handle(
          _dateMeta, date.isAcceptableOrUnknown(data['date']!, _dateMeta));
    } else if (isInserting) {
      context.missing(_dateMeta);
    }
    if (data.containsKey('start_km')) {
      context.handle(_startKmMeta,
          startKm.isAcceptableOrUnknown(data['start_km']!, _startKmMeta));
    } else if (isInserting) {
      context.missing(_startKmMeta);
    }
    if (data.containsKey('start_photo_url')) {
      context.handle(
          _startPhotoUrlMeta,
          startPhotoUrl.isAcceptableOrUnknown(
              data['start_photo_url']!, _startPhotoUrlMeta));
    } else if (isInserting) {
      context.missing(_startPhotoUrlMeta);
    }
    if (data.containsKey('start_location')) {
      context.handle(
          _startLocationMeta,
          startLocation.isAcceptableOrUnknown(
              data['start_location']!, _startLocationMeta));
    } else if (isInserting) {
      context.missing(_startLocationMeta);
    }
    if (data.containsKey('start_time')) {
      context.handle(_startTimeMeta,
          startTime.isAcceptableOrUnknown(data['start_time']!, _startTimeMeta));
    } else if (isInserting) {
      context.missing(_startTimeMeta);
    }
    if (data.containsKey('end_km')) {
      context.handle(
          _endKmMeta, endKm.isAcceptableOrUnknown(data['end_km']!, _endKmMeta));
    }
    if (data.containsKey('end_photo_url')) {
      context.handle(
          _endPhotoUrlMeta,
          endPhotoUrl.isAcceptableOrUnknown(
              data['end_photo_url']!, _endPhotoUrlMeta));
    }
    if (data.containsKey('end_location')) {
      context.handle(
          _endLocationMeta,
          endLocation.isAcceptableOrUnknown(
              data['end_location']!, _endLocationMeta));
    }
    if (data.containsKey('end_time')) {
      context.handle(_endTimeMeta,
          endTime.isAcceptableOrUnknown(data['end_time']!, _endTimeMeta));
    }
    if (data.containsKey('total_km')) {
      context.handle(_totalKmMeta,
          totalKm.isAcceptableOrUnknown(data['total_km']!, _totalKmMeta));
    }
    if (data.containsKey('status')) {
      context.handle(_statusMeta,
          status.isAcceptableOrUnknown(data['status']!, _statusMeta));
    }
    if (data.containsKey('is_synced')) {
      context.handle(_isSyncedMeta,
          isSynced.isAcceptableOrUnknown(data['is_synced']!, _isSyncedMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {id};
  @override
  SalesDailyJourney map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return SalesDailyJourney(
      id: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}id'])!,
      userId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}user_id'])!,
      date: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}date'])!,
      startKm: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}start_km'])!,
      startPhotoUrl: attachedDatabase.typeMapping.read(
          DriftSqlType.string, data['${effectivePrefix}start_photo_url'])!,
      startLocation: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}start_location'])!,
      startTime: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}start_time'])!,
      endKm: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}end_km']),
      endPhotoUrl: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}end_photo_url']),
      endLocation: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}end_location']),
      endTime: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}end_time']),
      totalKm: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}total_km']),
      status: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}status'])!,
      isSynced: attachedDatabase.typeMapping
          .read(DriftSqlType.bool, data['${effectivePrefix}is_synced'])!,
    );
  }

  @override
  $SalesDailyJourneysTable createAlias(String alias) {
    return $SalesDailyJourneysTable(attachedDatabase, alias);
  }
}

class SalesDailyJourney extends DataClass
    implements Insertable<SalesDailyJourney> {
  final String id;
  final String userId;
  final String date;
  final int startKm;
  final String startPhotoUrl;
  final String startLocation;
  final DateTime startTime;
  final int? endKm;
  final String? endPhotoUrl;
  final String? endLocation;
  final DateTime? endTime;
  final int? totalKm;
  final String status;
  final bool isSynced;
  const SalesDailyJourney(
      {required this.id,
      required this.userId,
      required this.date,
      required this.startKm,
      required this.startPhotoUrl,
      required this.startLocation,
      required this.startTime,
      this.endKm,
      this.endPhotoUrl,
      this.endLocation,
      this.endTime,
      this.totalKm,
      required this.status,
      required this.isSynced});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['id'] = Variable<String>(id);
    map['user_id'] = Variable<String>(userId);
    map['date'] = Variable<String>(date);
    map['start_km'] = Variable<int>(startKm);
    map['start_photo_url'] = Variable<String>(startPhotoUrl);
    map['start_location'] = Variable<String>(startLocation);
    map['start_time'] = Variable<DateTime>(startTime);
    if (!nullToAbsent || endKm != null) {
      map['end_km'] = Variable<int>(endKm);
    }
    if (!nullToAbsent || endPhotoUrl != null) {
      map['end_photo_url'] = Variable<String>(endPhotoUrl);
    }
    if (!nullToAbsent || endLocation != null) {
      map['end_location'] = Variable<String>(endLocation);
    }
    if (!nullToAbsent || endTime != null) {
      map['end_time'] = Variable<DateTime>(endTime);
    }
    if (!nullToAbsent || totalKm != null) {
      map['total_km'] = Variable<int>(totalKm);
    }
    map['status'] = Variable<String>(status);
    map['is_synced'] = Variable<bool>(isSynced);
    return map;
  }

  SalesDailyJourneysCompanion toCompanion(bool nullToAbsent) {
    return SalesDailyJourneysCompanion(
      id: Value(id),
      userId: Value(userId),
      date: Value(date),
      startKm: Value(startKm),
      startPhotoUrl: Value(startPhotoUrl),
      startLocation: Value(startLocation),
      startTime: Value(startTime),
      endKm:
          endKm == null && nullToAbsent ? const Value.absent() : Value(endKm),
      endPhotoUrl: endPhotoUrl == null && nullToAbsent
          ? const Value.absent()
          : Value(endPhotoUrl),
      endLocation: endLocation == null && nullToAbsent
          ? const Value.absent()
          : Value(endLocation),
      endTime: endTime == null && nullToAbsent
          ? const Value.absent()
          : Value(endTime),
      totalKm: totalKm == null && nullToAbsent
          ? const Value.absent()
          : Value(totalKm),
      status: Value(status),
      isSynced: Value(isSynced),
    );
  }

  factory SalesDailyJourney.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return SalesDailyJourney(
      id: serializer.fromJson<String>(json['id']),
      userId: serializer.fromJson<String>(json['userId']),
      date: serializer.fromJson<String>(json['date']),
      startKm: serializer.fromJson<int>(json['startKm']),
      startPhotoUrl: serializer.fromJson<String>(json['startPhotoUrl']),
      startLocation: serializer.fromJson<String>(json['startLocation']),
      startTime: serializer.fromJson<DateTime>(json['startTime']),
      endKm: serializer.fromJson<int?>(json['endKm']),
      endPhotoUrl: serializer.fromJson<String?>(json['endPhotoUrl']),
      endLocation: serializer.fromJson<String?>(json['endLocation']),
      endTime: serializer.fromJson<DateTime?>(json['endTime']),
      totalKm: serializer.fromJson<int?>(json['totalKm']),
      status: serializer.fromJson<String>(json['status']),
      isSynced: serializer.fromJson<bool>(json['isSynced']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'id': serializer.toJson<String>(id),
      'userId': serializer.toJson<String>(userId),
      'date': serializer.toJson<String>(date),
      'startKm': serializer.toJson<int>(startKm),
      'startPhotoUrl': serializer.toJson<String>(startPhotoUrl),
      'startLocation': serializer.toJson<String>(startLocation),
      'startTime': serializer.toJson<DateTime>(startTime),
      'endKm': serializer.toJson<int?>(endKm),
      'endPhotoUrl': serializer.toJson<String?>(endPhotoUrl),
      'endLocation': serializer.toJson<String?>(endLocation),
      'endTime': serializer.toJson<DateTime?>(endTime),
      'totalKm': serializer.toJson<int?>(totalKm),
      'status': serializer.toJson<String>(status),
      'isSynced': serializer.toJson<bool>(isSynced),
    };
  }

  SalesDailyJourney copyWith(
          {String? id,
          String? userId,
          String? date,
          int? startKm,
          String? startPhotoUrl,
          String? startLocation,
          DateTime? startTime,
          Value<int?> endKm = const Value.absent(),
          Value<String?> endPhotoUrl = const Value.absent(),
          Value<String?> endLocation = const Value.absent(),
          Value<DateTime?> endTime = const Value.absent(),
          Value<int?> totalKm = const Value.absent(),
          String? status,
          bool? isSynced}) =>
      SalesDailyJourney(
        id: id ?? this.id,
        userId: userId ?? this.userId,
        date: date ?? this.date,
        startKm: startKm ?? this.startKm,
        startPhotoUrl: startPhotoUrl ?? this.startPhotoUrl,
        startLocation: startLocation ?? this.startLocation,
        startTime: startTime ?? this.startTime,
        endKm: endKm.present ? endKm.value : this.endKm,
        endPhotoUrl: endPhotoUrl.present ? endPhotoUrl.value : this.endPhotoUrl,
        endLocation: endLocation.present ? endLocation.value : this.endLocation,
        endTime: endTime.present ? endTime.value : this.endTime,
        totalKm: totalKm.present ? totalKm.value : this.totalKm,
        status: status ?? this.status,
        isSynced: isSynced ?? this.isSynced,
      );
  SalesDailyJourney copyWithCompanion(SalesDailyJourneysCompanion data) {
    return SalesDailyJourney(
      id: data.id.present ? data.id.value : this.id,
      userId: data.userId.present ? data.userId.value : this.userId,
      date: data.date.present ? data.date.value : this.date,
      startKm: data.startKm.present ? data.startKm.value : this.startKm,
      startPhotoUrl: data.startPhotoUrl.present
          ? data.startPhotoUrl.value
          : this.startPhotoUrl,
      startLocation: data.startLocation.present
          ? data.startLocation.value
          : this.startLocation,
      startTime: data.startTime.present ? data.startTime.value : this.startTime,
      endKm: data.endKm.present ? data.endKm.value : this.endKm,
      endPhotoUrl:
          data.endPhotoUrl.present ? data.endPhotoUrl.value : this.endPhotoUrl,
      endLocation:
          data.endLocation.present ? data.endLocation.value : this.endLocation,
      endTime: data.endTime.present ? data.endTime.value : this.endTime,
      totalKm: data.totalKm.present ? data.totalKm.value : this.totalKm,
      status: data.status.present ? data.status.value : this.status,
      isSynced: data.isSynced.present ? data.isSynced.value : this.isSynced,
    );
  }

  @override
  String toString() {
    return (StringBuffer('SalesDailyJourney(')
          ..write('id: $id, ')
          ..write('userId: $userId, ')
          ..write('date: $date, ')
          ..write('startKm: $startKm, ')
          ..write('startPhotoUrl: $startPhotoUrl, ')
          ..write('startLocation: $startLocation, ')
          ..write('startTime: $startTime, ')
          ..write('endKm: $endKm, ')
          ..write('endPhotoUrl: $endPhotoUrl, ')
          ..write('endLocation: $endLocation, ')
          ..write('endTime: $endTime, ')
          ..write('totalKm: $totalKm, ')
          ..write('status: $status, ')
          ..write('isSynced: $isSynced')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
      id,
      userId,
      date,
      startKm,
      startPhotoUrl,
      startLocation,
      startTime,
      endKm,
      endPhotoUrl,
      endLocation,
      endTime,
      totalKm,
      status,
      isSynced);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is SalesDailyJourney &&
          other.id == this.id &&
          other.userId == this.userId &&
          other.date == this.date &&
          other.startKm == this.startKm &&
          other.startPhotoUrl == this.startPhotoUrl &&
          other.startLocation == this.startLocation &&
          other.startTime == this.startTime &&
          other.endKm == this.endKm &&
          other.endPhotoUrl == this.endPhotoUrl &&
          other.endLocation == this.endLocation &&
          other.endTime == this.endTime &&
          other.totalKm == this.totalKm &&
          other.status == this.status &&
          other.isSynced == this.isSynced);
}

class SalesDailyJourneysCompanion extends UpdateCompanion<SalesDailyJourney> {
  final Value<String> id;
  final Value<String> userId;
  final Value<String> date;
  final Value<int> startKm;
  final Value<String> startPhotoUrl;
  final Value<String> startLocation;
  final Value<DateTime> startTime;
  final Value<int?> endKm;
  final Value<String?> endPhotoUrl;
  final Value<String?> endLocation;
  final Value<DateTime?> endTime;
  final Value<int?> totalKm;
  final Value<String> status;
  final Value<bool> isSynced;
  final Value<int> rowid;
  const SalesDailyJourneysCompanion({
    this.id = const Value.absent(),
    this.userId = const Value.absent(),
    this.date = const Value.absent(),
    this.startKm = const Value.absent(),
    this.startPhotoUrl = const Value.absent(),
    this.startLocation = const Value.absent(),
    this.startTime = const Value.absent(),
    this.endKm = const Value.absent(),
    this.endPhotoUrl = const Value.absent(),
    this.endLocation = const Value.absent(),
    this.endTime = const Value.absent(),
    this.totalKm = const Value.absent(),
    this.status = const Value.absent(),
    this.isSynced = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  SalesDailyJourneysCompanion.insert({
    required String id,
    required String userId,
    required String date,
    required int startKm,
    required String startPhotoUrl,
    required String startLocation,
    required DateTime startTime,
    this.endKm = const Value.absent(),
    this.endPhotoUrl = const Value.absent(),
    this.endLocation = const Value.absent(),
    this.endTime = const Value.absent(),
    this.totalKm = const Value.absent(),
    this.status = const Value.absent(),
    this.isSynced = const Value.absent(),
    this.rowid = const Value.absent(),
  })  : id = Value(id),
        userId = Value(userId),
        date = Value(date),
        startKm = Value(startKm),
        startPhotoUrl = Value(startPhotoUrl),
        startLocation = Value(startLocation),
        startTime = Value(startTime);
  static Insertable<SalesDailyJourney> custom({
    Expression<String>? id,
    Expression<String>? userId,
    Expression<String>? date,
    Expression<int>? startKm,
    Expression<String>? startPhotoUrl,
    Expression<String>? startLocation,
    Expression<DateTime>? startTime,
    Expression<int>? endKm,
    Expression<String>? endPhotoUrl,
    Expression<String>? endLocation,
    Expression<DateTime>? endTime,
    Expression<int>? totalKm,
    Expression<String>? status,
    Expression<bool>? isSynced,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (id != null) 'id': id,
      if (userId != null) 'user_id': userId,
      if (date != null) 'date': date,
      if (startKm != null) 'start_km': startKm,
      if (startPhotoUrl != null) 'start_photo_url': startPhotoUrl,
      if (startLocation != null) 'start_location': startLocation,
      if (startTime != null) 'start_time': startTime,
      if (endKm != null) 'end_km': endKm,
      if (endPhotoUrl != null) 'end_photo_url': endPhotoUrl,
      if (endLocation != null) 'end_location': endLocation,
      if (endTime != null) 'end_time': endTime,
      if (totalKm != null) 'total_km': totalKm,
      if (status != null) 'status': status,
      if (isSynced != null) 'is_synced': isSynced,
      if (rowid != null) 'rowid': rowid,
    });
  }

  SalesDailyJourneysCompanion copyWith(
      {Value<String>? id,
      Value<String>? userId,
      Value<String>? date,
      Value<int>? startKm,
      Value<String>? startPhotoUrl,
      Value<String>? startLocation,
      Value<DateTime>? startTime,
      Value<int?>? endKm,
      Value<String?>? endPhotoUrl,
      Value<String?>? endLocation,
      Value<DateTime?>? endTime,
      Value<int?>? totalKm,
      Value<String>? status,
      Value<bool>? isSynced,
      Value<int>? rowid}) {
    return SalesDailyJourneysCompanion(
      id: id ?? this.id,
      userId: userId ?? this.userId,
      date: date ?? this.date,
      startKm: startKm ?? this.startKm,
      startPhotoUrl: startPhotoUrl ?? this.startPhotoUrl,
      startLocation: startLocation ?? this.startLocation,
      startTime: startTime ?? this.startTime,
      endKm: endKm ?? this.endKm,
      endPhotoUrl: endPhotoUrl ?? this.endPhotoUrl,
      endLocation: endLocation ?? this.endLocation,
      endTime: endTime ?? this.endTime,
      totalKm: totalKm ?? this.totalKm,
      status: status ?? this.status,
      isSynced: isSynced ?? this.isSynced,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (id.present) {
      map['id'] = Variable<String>(id.value);
    }
    if (userId.present) {
      map['user_id'] = Variable<String>(userId.value);
    }
    if (date.present) {
      map['date'] = Variable<String>(date.value);
    }
    if (startKm.present) {
      map['start_km'] = Variable<int>(startKm.value);
    }
    if (startPhotoUrl.present) {
      map['start_photo_url'] = Variable<String>(startPhotoUrl.value);
    }
    if (startLocation.present) {
      map['start_location'] = Variable<String>(startLocation.value);
    }
    if (startTime.present) {
      map['start_time'] = Variable<DateTime>(startTime.value);
    }
    if (endKm.present) {
      map['end_km'] = Variable<int>(endKm.value);
    }
    if (endPhotoUrl.present) {
      map['end_photo_url'] = Variable<String>(endPhotoUrl.value);
    }
    if (endLocation.present) {
      map['end_location'] = Variable<String>(endLocation.value);
    }
    if (endTime.present) {
      map['end_time'] = Variable<DateTime>(endTime.value);
    }
    if (totalKm.present) {
      map['total_km'] = Variable<int>(totalKm.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (isSynced.present) {
      map['is_synced'] = Variable<bool>(isSynced.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('SalesDailyJourneysCompanion(')
          ..write('id: $id, ')
          ..write('userId: $userId, ')
          ..write('date: $date, ')
          ..write('startKm: $startKm, ')
          ..write('startPhotoUrl: $startPhotoUrl, ')
          ..write('startLocation: $startLocation, ')
          ..write('startTime: $startTime, ')
          ..write('endKm: $endKm, ')
          ..write('endPhotoUrl: $endPhotoUrl, ')
          ..write('endLocation: $endLocation, ')
          ..write('endTime: $endTime, ')
          ..write('totalKm: $totalKm, ')
          ..write('status: $status, ')
          ..write('isSynced: $isSynced, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

abstract class _$AppDatabase extends GeneratedDatabase {
  _$AppDatabase(QueryExecutor e) : super(e);
  $AppDatabaseManager get managers => $AppDatabaseManager(this);
  late final $ProductsTable products = $ProductsTable(this);
  late final $CompaniesTable companies = $CompaniesTable(this);
  late final $LocalCheckinsTable localCheckins = $LocalCheckinsTable(this);
  late final $LocalTransactionsTable localTransactions =
      $LocalTransactionsTable(this);
  late final $LocalTransactionItemsTable localTransactionItems =
      $LocalTransactionItemsTable(this);
  late final $SalesStockTable salesStock = $SalesStockTable(this);
  late final $OfflineTasksTable offlineTasks = $OfflineTasksTable(this);
  late final $StoresTable stores = $StoresTable(this);
  late final $VisitPlansTable visitPlans = $VisitPlansTable(this);
  late final $SalesDailyJourneysTable salesDailyJourneys =
      $SalesDailyJourneysTable(this);
  @override
  Iterable<TableInfo<Table, Object?>> get allTables =>
      allSchemaEntities.whereType<TableInfo<Table, Object?>>();
  @override
  List<DatabaseSchemaEntity> get allSchemaEntities => [
        products,
        companies,
        localCheckins,
        localTransactions,
        localTransactionItems,
        salesStock,
        offlineTasks,
        stores,
        visitPlans,
        salesDailyJourneys
      ];
}

typedef $$ProductsTableCreateCompanionBuilder = ProductsCompanion Function({
  required String id,
  required String companyId,
  required String name,
  Value<String?> sku,
  required double sellingPrice,
  Value<double> dbp,
  Value<double> wbp,
  Value<double> rbp,
  Value<double> costPrice,
  Value<String?> unit,
  Value<String?> category,
  Value<int> warehouseStock,
  Value<int> pcsPerUnit,
  Value<String?> imageUrl,
  Value<int> rowid,
});
typedef $$ProductsTableUpdateCompanionBuilder = ProductsCompanion Function({
  Value<String> id,
  Value<String> companyId,
  Value<String> name,
  Value<String?> sku,
  Value<double> sellingPrice,
  Value<double> dbp,
  Value<double> wbp,
  Value<double> rbp,
  Value<double> costPrice,
  Value<String?> unit,
  Value<String?> category,
  Value<int> warehouseStock,
  Value<int> pcsPerUnit,
  Value<String?> imageUrl,
  Value<int> rowid,
});

class $$ProductsTableFilterComposer
    extends Composer<_$AppDatabase, $ProductsTable> {
  $$ProductsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get companyId => $composableBuilder(
      column: $table.companyId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get name => $composableBuilder(
      column: $table.name, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get sku => $composableBuilder(
      column: $table.sku, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get sellingPrice => $composableBuilder(
      column: $table.sellingPrice, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get dbp => $composableBuilder(
      column: $table.dbp, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get wbp => $composableBuilder(
      column: $table.wbp, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get rbp => $composableBuilder(
      column: $table.rbp, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get costPrice => $composableBuilder(
      column: $table.costPrice, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get unit => $composableBuilder(
      column: $table.unit, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get category => $composableBuilder(
      column: $table.category, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get warehouseStock => $composableBuilder(
      column: $table.warehouseStock,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get pcsPerUnit => $composableBuilder(
      column: $table.pcsPerUnit, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get imageUrl => $composableBuilder(
      column: $table.imageUrl, builder: (column) => ColumnFilters(column));
}

class $$ProductsTableOrderingComposer
    extends Composer<_$AppDatabase, $ProductsTable> {
  $$ProductsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get companyId => $composableBuilder(
      column: $table.companyId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get name => $composableBuilder(
      column: $table.name, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get sku => $composableBuilder(
      column: $table.sku, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get sellingPrice => $composableBuilder(
      column: $table.sellingPrice,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get dbp => $composableBuilder(
      column: $table.dbp, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get wbp => $composableBuilder(
      column: $table.wbp, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get rbp => $composableBuilder(
      column: $table.rbp, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get costPrice => $composableBuilder(
      column: $table.costPrice, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get unit => $composableBuilder(
      column: $table.unit, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get category => $composableBuilder(
      column: $table.category, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get warehouseStock => $composableBuilder(
      column: $table.warehouseStock,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get pcsPerUnit => $composableBuilder(
      column: $table.pcsPerUnit, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get imageUrl => $composableBuilder(
      column: $table.imageUrl, builder: (column) => ColumnOrderings(column));
}

class $$ProductsTableAnnotationComposer
    extends Composer<_$AppDatabase, $ProductsTable> {
  $$ProductsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get companyId =>
      $composableBuilder(column: $table.companyId, builder: (column) => column);

  GeneratedColumn<String> get name =>
      $composableBuilder(column: $table.name, builder: (column) => column);

  GeneratedColumn<String> get sku =>
      $composableBuilder(column: $table.sku, builder: (column) => column);

  GeneratedColumn<double> get sellingPrice => $composableBuilder(
      column: $table.sellingPrice, builder: (column) => column);

  GeneratedColumn<double> get dbp =>
      $composableBuilder(column: $table.dbp, builder: (column) => column);

  GeneratedColumn<double> get wbp =>
      $composableBuilder(column: $table.wbp, builder: (column) => column);

  GeneratedColumn<double> get rbp =>
      $composableBuilder(column: $table.rbp, builder: (column) => column);

  GeneratedColumn<double> get costPrice =>
      $composableBuilder(column: $table.costPrice, builder: (column) => column);

  GeneratedColumn<String> get unit =>
      $composableBuilder(column: $table.unit, builder: (column) => column);

  GeneratedColumn<String> get category =>
      $composableBuilder(column: $table.category, builder: (column) => column);

  GeneratedColumn<int> get warehouseStock => $composableBuilder(
      column: $table.warehouseStock, builder: (column) => column);

  GeneratedColumn<int> get pcsPerUnit => $composableBuilder(
      column: $table.pcsPerUnit, builder: (column) => column);

  GeneratedColumn<String> get imageUrl =>
      $composableBuilder(column: $table.imageUrl, builder: (column) => column);
}

class $$ProductsTableTableManager extends RootTableManager<
    _$AppDatabase,
    $ProductsTable,
    Product,
    $$ProductsTableFilterComposer,
    $$ProductsTableOrderingComposer,
    $$ProductsTableAnnotationComposer,
    $$ProductsTableCreateCompanionBuilder,
    $$ProductsTableUpdateCompanionBuilder,
    (Product, BaseReferences<_$AppDatabase, $ProductsTable, Product>),
    Product,
    PrefetchHooks Function()> {
  $$ProductsTableTableManager(_$AppDatabase db, $ProductsTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$ProductsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$ProductsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$ProductsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> id = const Value.absent(),
            Value<String> companyId = const Value.absent(),
            Value<String> name = const Value.absent(),
            Value<String?> sku = const Value.absent(),
            Value<double> sellingPrice = const Value.absent(),
            Value<double> dbp = const Value.absent(),
            Value<double> wbp = const Value.absent(),
            Value<double> rbp = const Value.absent(),
            Value<double> costPrice = const Value.absent(),
            Value<String?> unit = const Value.absent(),
            Value<String?> category = const Value.absent(),
            Value<int> warehouseStock = const Value.absent(),
            Value<int> pcsPerUnit = const Value.absent(),
            Value<String?> imageUrl = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              ProductsCompanion(
            id: id,
            companyId: companyId,
            name: name,
            sku: sku,
            sellingPrice: sellingPrice,
            dbp: dbp,
            wbp: wbp,
            rbp: rbp,
            costPrice: costPrice,
            unit: unit,
            category: category,
            warehouseStock: warehouseStock,
            pcsPerUnit: pcsPerUnit,
            imageUrl: imageUrl,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String id,
            required String companyId,
            required String name,
            Value<String?> sku = const Value.absent(),
            required double sellingPrice,
            Value<double> dbp = const Value.absent(),
            Value<double> wbp = const Value.absent(),
            Value<double> rbp = const Value.absent(),
            Value<double> costPrice = const Value.absent(),
            Value<String?> unit = const Value.absent(),
            Value<String?> category = const Value.absent(),
            Value<int> warehouseStock = const Value.absent(),
            Value<int> pcsPerUnit = const Value.absent(),
            Value<String?> imageUrl = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              ProductsCompanion.insert(
            id: id,
            companyId: companyId,
            name: name,
            sku: sku,
            sellingPrice: sellingPrice,
            dbp: dbp,
            wbp: wbp,
            rbp: rbp,
            costPrice: costPrice,
            unit: unit,
            category: category,
            warehouseStock: warehouseStock,
            pcsPerUnit: pcsPerUnit,
            imageUrl: imageUrl,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$ProductsTableProcessedTableManager = ProcessedTableManager<
    _$AppDatabase,
    $ProductsTable,
    Product,
    $$ProductsTableFilterComposer,
    $$ProductsTableOrderingComposer,
    $$ProductsTableAnnotationComposer,
    $$ProductsTableCreateCompanionBuilder,
    $$ProductsTableUpdateCompanionBuilder,
    (Product, BaseReferences<_$AppDatabase, $ProductsTable, Product>),
    Product,
    PrefetchHooks Function()>;
typedef $$CompaniesTableCreateCompanionBuilder = CompaniesCompanion Function({
  required String id,
  required String name,
  Value<String?> code,
  Value<int> rowid,
});
typedef $$CompaniesTableUpdateCompanionBuilder = CompaniesCompanion Function({
  Value<String> id,
  Value<String> name,
  Value<String?> code,
  Value<int> rowid,
});

class $$CompaniesTableFilterComposer
    extends Composer<_$AppDatabase, $CompaniesTable> {
  $$CompaniesTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get name => $composableBuilder(
      column: $table.name, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get code => $composableBuilder(
      column: $table.code, builder: (column) => ColumnFilters(column));
}

class $$CompaniesTableOrderingComposer
    extends Composer<_$AppDatabase, $CompaniesTable> {
  $$CompaniesTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get name => $composableBuilder(
      column: $table.name, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get code => $composableBuilder(
      column: $table.code, builder: (column) => ColumnOrderings(column));
}

class $$CompaniesTableAnnotationComposer
    extends Composer<_$AppDatabase, $CompaniesTable> {
  $$CompaniesTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get name =>
      $composableBuilder(column: $table.name, builder: (column) => column);

  GeneratedColumn<String> get code =>
      $composableBuilder(column: $table.code, builder: (column) => column);
}

class $$CompaniesTableTableManager extends RootTableManager<
    _$AppDatabase,
    $CompaniesTable,
    Company,
    $$CompaniesTableFilterComposer,
    $$CompaniesTableOrderingComposer,
    $$CompaniesTableAnnotationComposer,
    $$CompaniesTableCreateCompanionBuilder,
    $$CompaniesTableUpdateCompanionBuilder,
    (Company, BaseReferences<_$AppDatabase, $CompaniesTable, Company>),
    Company,
    PrefetchHooks Function()> {
  $$CompaniesTableTableManager(_$AppDatabase db, $CompaniesTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$CompaniesTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$CompaniesTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$CompaniesTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> id = const Value.absent(),
            Value<String> name = const Value.absent(),
            Value<String?> code = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              CompaniesCompanion(
            id: id,
            name: name,
            code: code,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String id,
            required String name,
            Value<String?> code = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              CompaniesCompanion.insert(
            id: id,
            name: name,
            code: code,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$CompaniesTableProcessedTableManager = ProcessedTableManager<
    _$AppDatabase,
    $CompaniesTable,
    Company,
    $$CompaniesTableFilterComposer,
    $$CompaniesTableOrderingComposer,
    $$CompaniesTableAnnotationComposer,
    $$CompaniesTableCreateCompanionBuilder,
    $$CompaniesTableUpdateCompanionBuilder,
    (Company, BaseReferences<_$AppDatabase, $CompaniesTable, Company>),
    Company,
    PrefetchHooks Function()>;
typedef $$LocalCheckinsTableCreateCompanionBuilder = LocalCheckinsCompanion
    Function({
  Value<int> id,
  required String storeId,
  required String storeName,
  required double latitude,
  required double longitude,
  required String selfiePath,
  required DateTime createdAt,
  Value<String> syncStatus,
});
typedef $$LocalCheckinsTableUpdateCompanionBuilder = LocalCheckinsCompanion
    Function({
  Value<int> id,
  Value<String> storeId,
  Value<String> storeName,
  Value<double> latitude,
  Value<double> longitude,
  Value<String> selfiePath,
  Value<DateTime> createdAt,
  Value<String> syncStatus,
});

class $$LocalCheckinsTableFilterComposer
    extends Composer<_$AppDatabase, $LocalCheckinsTable> {
  $$LocalCheckinsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<int> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get storeId => $composableBuilder(
      column: $table.storeId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get storeName => $composableBuilder(
      column: $table.storeName, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get latitude => $composableBuilder(
      column: $table.latitude, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get longitude => $composableBuilder(
      column: $table.longitude, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get selfiePath => $composableBuilder(
      column: $table.selfiePath, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get syncStatus => $composableBuilder(
      column: $table.syncStatus, builder: (column) => ColumnFilters(column));
}

class $$LocalCheckinsTableOrderingComposer
    extends Composer<_$AppDatabase, $LocalCheckinsTable> {
  $$LocalCheckinsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<int> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get storeId => $composableBuilder(
      column: $table.storeId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get storeName => $composableBuilder(
      column: $table.storeName, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get latitude => $composableBuilder(
      column: $table.latitude, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get longitude => $composableBuilder(
      column: $table.longitude, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get selfiePath => $composableBuilder(
      column: $table.selfiePath, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get syncStatus => $composableBuilder(
      column: $table.syncStatus, builder: (column) => ColumnOrderings(column));
}

class $$LocalCheckinsTableAnnotationComposer
    extends Composer<_$AppDatabase, $LocalCheckinsTable> {
  $$LocalCheckinsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<int> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get storeId =>
      $composableBuilder(column: $table.storeId, builder: (column) => column);

  GeneratedColumn<String> get storeName =>
      $composableBuilder(column: $table.storeName, builder: (column) => column);

  GeneratedColumn<double> get latitude =>
      $composableBuilder(column: $table.latitude, builder: (column) => column);

  GeneratedColumn<double> get longitude =>
      $composableBuilder(column: $table.longitude, builder: (column) => column);

  GeneratedColumn<String> get selfiePath => $composableBuilder(
      column: $table.selfiePath, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<String> get syncStatus => $composableBuilder(
      column: $table.syncStatus, builder: (column) => column);
}

class $$LocalCheckinsTableTableManager extends RootTableManager<
    _$AppDatabase,
    $LocalCheckinsTable,
    LocalCheckin,
    $$LocalCheckinsTableFilterComposer,
    $$LocalCheckinsTableOrderingComposer,
    $$LocalCheckinsTableAnnotationComposer,
    $$LocalCheckinsTableCreateCompanionBuilder,
    $$LocalCheckinsTableUpdateCompanionBuilder,
    (
      LocalCheckin,
      BaseReferences<_$AppDatabase, $LocalCheckinsTable, LocalCheckin>
    ),
    LocalCheckin,
    PrefetchHooks Function()> {
  $$LocalCheckinsTableTableManager(_$AppDatabase db, $LocalCheckinsTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$LocalCheckinsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$LocalCheckinsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$LocalCheckinsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<int> id = const Value.absent(),
            Value<String> storeId = const Value.absent(),
            Value<String> storeName = const Value.absent(),
            Value<double> latitude = const Value.absent(),
            Value<double> longitude = const Value.absent(),
            Value<String> selfiePath = const Value.absent(),
            Value<DateTime> createdAt = const Value.absent(),
            Value<String> syncStatus = const Value.absent(),
          }) =>
              LocalCheckinsCompanion(
            id: id,
            storeId: storeId,
            storeName: storeName,
            latitude: latitude,
            longitude: longitude,
            selfiePath: selfiePath,
            createdAt: createdAt,
            syncStatus: syncStatus,
          ),
          createCompanionCallback: ({
            Value<int> id = const Value.absent(),
            required String storeId,
            required String storeName,
            required double latitude,
            required double longitude,
            required String selfiePath,
            required DateTime createdAt,
            Value<String> syncStatus = const Value.absent(),
          }) =>
              LocalCheckinsCompanion.insert(
            id: id,
            storeId: storeId,
            storeName: storeName,
            latitude: latitude,
            longitude: longitude,
            selfiePath: selfiePath,
            createdAt: createdAt,
            syncStatus: syncStatus,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$LocalCheckinsTableProcessedTableManager = ProcessedTableManager<
    _$AppDatabase,
    $LocalCheckinsTable,
    LocalCheckin,
    $$LocalCheckinsTableFilterComposer,
    $$LocalCheckinsTableOrderingComposer,
    $$LocalCheckinsTableAnnotationComposer,
    $$LocalCheckinsTableCreateCompanionBuilder,
    $$LocalCheckinsTableUpdateCompanionBuilder,
    (
      LocalCheckin,
      BaseReferences<_$AppDatabase, $LocalCheckinsTable, LocalCheckin>
    ),
    LocalCheckin,
    PrefetchHooks Function()>;
typedef $$LocalTransactionsTableCreateCompanionBuilder
    = LocalTransactionsCompanion Function({
  required String localId,
  Value<String?> receiptNo,
  required String companyId,
  required String companyName,
  required String storeId,
  required String storeName,
  required double totalAmount,
  Value<double> paidAmount,
  Value<String?> selfiePath,
  Value<String?> receiptPath,
  Value<String?> signaturePath,
  Value<String?> notes,
  Value<String> paymentMethod,
  Value<String?> dpPaymentMethod,
  Value<String?> paymentBank,
  Value<String?> midtransId,
  Value<String?> midtransQrisUrl,
  Value<String?> midtransVaNumber,
  Value<String?> midtransBank,
  Value<String?> midtransBillKey,
  Value<String?> midtransBillerCode,
  required DateTime createdAt,
  Value<String> syncStatus,
  Value<DateTime?> paymentDueDate,
  Value<int> tenor,
  Value<bool> isMt,
  Value<bool> isDistributorOrder,
  Value<String?> factoryCompanyId,
  Value<String?> companyBankAccountId,
  Value<String?> visitId,
  Value<double?> latitude,
  Value<double?> longitude,
  Value<String?> syncErrorMessage,
  Value<DateTime?> lastSyncAttempt,
  Value<int> rowid,
});
typedef $$LocalTransactionsTableUpdateCompanionBuilder
    = LocalTransactionsCompanion Function({
  Value<String> localId,
  Value<String?> receiptNo,
  Value<String> companyId,
  Value<String> companyName,
  Value<String> storeId,
  Value<String> storeName,
  Value<double> totalAmount,
  Value<double> paidAmount,
  Value<String?> selfiePath,
  Value<String?> receiptPath,
  Value<String?> signaturePath,
  Value<String?> notes,
  Value<String> paymentMethod,
  Value<String?> dpPaymentMethod,
  Value<String?> paymentBank,
  Value<String?> midtransId,
  Value<String?> midtransQrisUrl,
  Value<String?> midtransVaNumber,
  Value<String?> midtransBank,
  Value<String?> midtransBillKey,
  Value<String?> midtransBillerCode,
  Value<DateTime> createdAt,
  Value<String> syncStatus,
  Value<DateTime?> paymentDueDate,
  Value<int> tenor,
  Value<bool> isMt,
  Value<bool> isDistributorOrder,
  Value<String?> factoryCompanyId,
  Value<String?> companyBankAccountId,
  Value<String?> visitId,
  Value<double?> latitude,
  Value<double?> longitude,
  Value<String?> syncErrorMessage,
  Value<DateTime?> lastSyncAttempt,
  Value<int> rowid,
});

final class $$LocalTransactionsTableReferences extends BaseReferences<
    _$AppDatabase, $LocalTransactionsTable, LocalTransaction> {
  $$LocalTransactionsTableReferences(
      super.$_db, super.$_table, super.$_typedResult);

  static MultiTypedResultKey<$LocalTransactionItemsTable,
      List<LocalTransactionItem>> _localTransactionItemsRefsTable(
          _$AppDatabase db) =>
      MultiTypedResultKey.fromTable(db.localTransactionItems,
          aliasName:
              'local_transactions__local_id__local_transaction_items__transaction_local_id');

  $$LocalTransactionItemsTableProcessedTableManager
      get localTransactionItemsRefs {
    final manager = $$LocalTransactionItemsTableTableManager(
            $_db, $_db.localTransactionItems)
        .filter((f) => f.transactionLocalId.localId
            .sqlEquals($_itemColumn<String>('local_id')!));

    final cache =
        $_typedResult.readTableOrNull(_localTransactionItemsRefsTable($_db));
    return ProcessedTableManager(
        manager.$state.copyWith(prefetchedData: cache));
  }
}

class $$LocalTransactionsTableFilterComposer
    extends Composer<_$AppDatabase, $LocalTransactionsTable> {
  $$LocalTransactionsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get localId => $composableBuilder(
      column: $table.localId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get receiptNo => $composableBuilder(
      column: $table.receiptNo, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get companyId => $composableBuilder(
      column: $table.companyId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get companyName => $composableBuilder(
      column: $table.companyName, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get storeId => $composableBuilder(
      column: $table.storeId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get storeName => $composableBuilder(
      column: $table.storeName, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get totalAmount => $composableBuilder(
      column: $table.totalAmount, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get paidAmount => $composableBuilder(
      column: $table.paidAmount, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get selfiePath => $composableBuilder(
      column: $table.selfiePath, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get receiptPath => $composableBuilder(
      column: $table.receiptPath, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get signaturePath => $composableBuilder(
      column: $table.signaturePath, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get notes => $composableBuilder(
      column: $table.notes, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get paymentMethod => $composableBuilder(
      column: $table.paymentMethod, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get dpPaymentMethod => $composableBuilder(
      column: $table.dpPaymentMethod,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get paymentBank => $composableBuilder(
      column: $table.paymentBank, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get midtransId => $composableBuilder(
      column: $table.midtransId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get midtransQrisUrl => $composableBuilder(
      column: $table.midtransQrisUrl,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get midtransVaNumber => $composableBuilder(
      column: $table.midtransVaNumber,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get midtransBank => $composableBuilder(
      column: $table.midtransBank, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get midtransBillKey => $composableBuilder(
      column: $table.midtransBillKey,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get midtransBillerCode => $composableBuilder(
      column: $table.midtransBillerCode,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get syncStatus => $composableBuilder(
      column: $table.syncStatus, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get paymentDueDate => $composableBuilder(
      column: $table.paymentDueDate,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get tenor => $composableBuilder(
      column: $table.tenor, builder: (column) => ColumnFilters(column));

  ColumnFilters<bool> get isMt => $composableBuilder(
      column: $table.isMt, builder: (column) => ColumnFilters(column));

  ColumnFilters<bool> get isDistributorOrder => $composableBuilder(
      column: $table.isDistributorOrder,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get factoryCompanyId => $composableBuilder(
      column: $table.factoryCompanyId,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get companyBankAccountId => $composableBuilder(
      column: $table.companyBankAccountId,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get visitId => $composableBuilder(
      column: $table.visitId, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get latitude => $composableBuilder(
      column: $table.latitude, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get longitude => $composableBuilder(
      column: $table.longitude, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get syncErrorMessage => $composableBuilder(
      column: $table.syncErrorMessage,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get lastSyncAttempt => $composableBuilder(
      column: $table.lastSyncAttempt,
      builder: (column) => ColumnFilters(column));

  Expression<bool> localTransactionItemsRefs(
      Expression<bool> Function($$LocalTransactionItemsTableFilterComposer f)
          f) {
    final $$LocalTransactionItemsTableFilterComposer composer =
        $composerBuilder(
            composer: this,
            getCurrentColumn: (t) => t.localId,
            referencedTable: $db.localTransactionItems,
            getReferencedColumn: (t) => t.transactionLocalId,
            builder: (joinBuilder,
                    {$addJoinBuilderToRootComposer,
                    $removeJoinBuilderFromRootComposer}) =>
                $$LocalTransactionItemsTableFilterComposer(
                  $db: $db,
                  $table: $db.localTransactionItems,
                  $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
                  joinBuilder: joinBuilder,
                  $removeJoinBuilderFromRootComposer:
                      $removeJoinBuilderFromRootComposer,
                ));
    return f(composer);
  }
}

class $$LocalTransactionsTableOrderingComposer
    extends Composer<_$AppDatabase, $LocalTransactionsTable> {
  $$LocalTransactionsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get localId => $composableBuilder(
      column: $table.localId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get receiptNo => $composableBuilder(
      column: $table.receiptNo, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get companyId => $composableBuilder(
      column: $table.companyId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get companyName => $composableBuilder(
      column: $table.companyName, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get storeId => $composableBuilder(
      column: $table.storeId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get storeName => $composableBuilder(
      column: $table.storeName, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get totalAmount => $composableBuilder(
      column: $table.totalAmount, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get paidAmount => $composableBuilder(
      column: $table.paidAmount, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get selfiePath => $composableBuilder(
      column: $table.selfiePath, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get receiptPath => $composableBuilder(
      column: $table.receiptPath, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get signaturePath => $composableBuilder(
      column: $table.signaturePath,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get notes => $composableBuilder(
      column: $table.notes, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get paymentMethod => $composableBuilder(
      column: $table.paymentMethod,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get dpPaymentMethod => $composableBuilder(
      column: $table.dpPaymentMethod,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get paymentBank => $composableBuilder(
      column: $table.paymentBank, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get midtransId => $composableBuilder(
      column: $table.midtransId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get midtransQrisUrl => $composableBuilder(
      column: $table.midtransQrisUrl,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get midtransVaNumber => $composableBuilder(
      column: $table.midtransVaNumber,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get midtransBank => $composableBuilder(
      column: $table.midtransBank,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get midtransBillKey => $composableBuilder(
      column: $table.midtransBillKey,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get midtransBillerCode => $composableBuilder(
      column: $table.midtransBillerCode,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get syncStatus => $composableBuilder(
      column: $table.syncStatus, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get paymentDueDate => $composableBuilder(
      column: $table.paymentDueDate,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get tenor => $composableBuilder(
      column: $table.tenor, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<bool> get isMt => $composableBuilder(
      column: $table.isMt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<bool> get isDistributorOrder => $composableBuilder(
      column: $table.isDistributorOrder,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get factoryCompanyId => $composableBuilder(
      column: $table.factoryCompanyId,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get companyBankAccountId => $composableBuilder(
      column: $table.companyBankAccountId,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get visitId => $composableBuilder(
      column: $table.visitId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get latitude => $composableBuilder(
      column: $table.latitude, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get longitude => $composableBuilder(
      column: $table.longitude, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get syncErrorMessage => $composableBuilder(
      column: $table.syncErrorMessage,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get lastSyncAttempt => $composableBuilder(
      column: $table.lastSyncAttempt,
      builder: (column) => ColumnOrderings(column));
}

class $$LocalTransactionsTableAnnotationComposer
    extends Composer<_$AppDatabase, $LocalTransactionsTable> {
  $$LocalTransactionsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get localId =>
      $composableBuilder(column: $table.localId, builder: (column) => column);

  GeneratedColumn<String> get receiptNo =>
      $composableBuilder(column: $table.receiptNo, builder: (column) => column);

  GeneratedColumn<String> get companyId =>
      $composableBuilder(column: $table.companyId, builder: (column) => column);

  GeneratedColumn<String> get companyName => $composableBuilder(
      column: $table.companyName, builder: (column) => column);

  GeneratedColumn<String> get storeId =>
      $composableBuilder(column: $table.storeId, builder: (column) => column);

  GeneratedColumn<String> get storeName =>
      $composableBuilder(column: $table.storeName, builder: (column) => column);

  GeneratedColumn<double> get totalAmount => $composableBuilder(
      column: $table.totalAmount, builder: (column) => column);

  GeneratedColumn<double> get paidAmount => $composableBuilder(
      column: $table.paidAmount, builder: (column) => column);

  GeneratedColumn<String> get selfiePath => $composableBuilder(
      column: $table.selfiePath, builder: (column) => column);

  GeneratedColumn<String> get receiptPath => $composableBuilder(
      column: $table.receiptPath, builder: (column) => column);

  GeneratedColumn<String> get signaturePath => $composableBuilder(
      column: $table.signaturePath, builder: (column) => column);

  GeneratedColumn<String> get notes =>
      $composableBuilder(column: $table.notes, builder: (column) => column);

  GeneratedColumn<String> get paymentMethod => $composableBuilder(
      column: $table.paymentMethod, builder: (column) => column);

  GeneratedColumn<String> get dpPaymentMethod => $composableBuilder(
      column: $table.dpPaymentMethod, builder: (column) => column);

  GeneratedColumn<String> get paymentBank => $composableBuilder(
      column: $table.paymentBank, builder: (column) => column);

  GeneratedColumn<String> get midtransId => $composableBuilder(
      column: $table.midtransId, builder: (column) => column);

  GeneratedColumn<String> get midtransQrisUrl => $composableBuilder(
      column: $table.midtransQrisUrl, builder: (column) => column);

  GeneratedColumn<String> get midtransVaNumber => $composableBuilder(
      column: $table.midtransVaNumber, builder: (column) => column);

  GeneratedColumn<String> get midtransBank => $composableBuilder(
      column: $table.midtransBank, builder: (column) => column);

  GeneratedColumn<String> get midtransBillKey => $composableBuilder(
      column: $table.midtransBillKey, builder: (column) => column);

  GeneratedColumn<String> get midtransBillerCode => $composableBuilder(
      column: $table.midtransBillerCode, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<String> get syncStatus => $composableBuilder(
      column: $table.syncStatus, builder: (column) => column);

  GeneratedColumn<DateTime> get paymentDueDate => $composableBuilder(
      column: $table.paymentDueDate, builder: (column) => column);

  GeneratedColumn<int> get tenor =>
      $composableBuilder(column: $table.tenor, builder: (column) => column);

  GeneratedColumn<bool> get isMt =>
      $composableBuilder(column: $table.isMt, builder: (column) => column);

  GeneratedColumn<bool> get isDistributorOrder => $composableBuilder(
      column: $table.isDistributorOrder, builder: (column) => column);

  GeneratedColumn<String> get factoryCompanyId => $composableBuilder(
      column: $table.factoryCompanyId, builder: (column) => column);

  GeneratedColumn<String> get companyBankAccountId => $composableBuilder(
      column: $table.companyBankAccountId, builder: (column) => column);

  GeneratedColumn<String> get visitId =>
      $composableBuilder(column: $table.visitId, builder: (column) => column);

  GeneratedColumn<double> get latitude =>
      $composableBuilder(column: $table.latitude, builder: (column) => column);

  GeneratedColumn<double> get longitude =>
      $composableBuilder(column: $table.longitude, builder: (column) => column);

  GeneratedColumn<String> get syncErrorMessage => $composableBuilder(
      column: $table.syncErrorMessage, builder: (column) => column);

  GeneratedColumn<DateTime> get lastSyncAttempt => $composableBuilder(
      column: $table.lastSyncAttempt, builder: (column) => column);

  Expression<T> localTransactionItemsRefs<T extends Object>(
      Expression<T> Function($$LocalTransactionItemsTableAnnotationComposer a)
          f) {
    final $$LocalTransactionItemsTableAnnotationComposer composer =
        $composerBuilder(
            composer: this,
            getCurrentColumn: (t) => t.localId,
            referencedTable: $db.localTransactionItems,
            getReferencedColumn: (t) => t.transactionLocalId,
            builder: (joinBuilder,
                    {$addJoinBuilderToRootComposer,
                    $removeJoinBuilderFromRootComposer}) =>
                $$LocalTransactionItemsTableAnnotationComposer(
                  $db: $db,
                  $table: $db.localTransactionItems,
                  $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
                  joinBuilder: joinBuilder,
                  $removeJoinBuilderFromRootComposer:
                      $removeJoinBuilderFromRootComposer,
                ));
    return f(composer);
  }
}

class $$LocalTransactionsTableTableManager extends RootTableManager<
    _$AppDatabase,
    $LocalTransactionsTable,
    LocalTransaction,
    $$LocalTransactionsTableFilterComposer,
    $$LocalTransactionsTableOrderingComposer,
    $$LocalTransactionsTableAnnotationComposer,
    $$LocalTransactionsTableCreateCompanionBuilder,
    $$LocalTransactionsTableUpdateCompanionBuilder,
    (LocalTransaction, $$LocalTransactionsTableReferences),
    LocalTransaction,
    PrefetchHooks Function({bool localTransactionItemsRefs})> {
  $$LocalTransactionsTableTableManager(
      _$AppDatabase db, $LocalTransactionsTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$LocalTransactionsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$LocalTransactionsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$LocalTransactionsTableAnnotationComposer(
                  $db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> localId = const Value.absent(),
            Value<String?> receiptNo = const Value.absent(),
            Value<String> companyId = const Value.absent(),
            Value<String> companyName = const Value.absent(),
            Value<String> storeId = const Value.absent(),
            Value<String> storeName = const Value.absent(),
            Value<double> totalAmount = const Value.absent(),
            Value<double> paidAmount = const Value.absent(),
            Value<String?> selfiePath = const Value.absent(),
            Value<String?> receiptPath = const Value.absent(),
            Value<String?> signaturePath = const Value.absent(),
            Value<String?> notes = const Value.absent(),
            Value<String> paymentMethod = const Value.absent(),
            Value<String?> dpPaymentMethod = const Value.absent(),
            Value<String?> paymentBank = const Value.absent(),
            Value<String?> midtransId = const Value.absent(),
            Value<String?> midtransQrisUrl = const Value.absent(),
            Value<String?> midtransVaNumber = const Value.absent(),
            Value<String?> midtransBank = const Value.absent(),
            Value<String?> midtransBillKey = const Value.absent(),
            Value<String?> midtransBillerCode = const Value.absent(),
            Value<DateTime> createdAt = const Value.absent(),
            Value<String> syncStatus = const Value.absent(),
            Value<DateTime?> paymentDueDate = const Value.absent(),
            Value<int> tenor = const Value.absent(),
            Value<bool> isMt = const Value.absent(),
            Value<bool> isDistributorOrder = const Value.absent(),
            Value<String?> factoryCompanyId = const Value.absent(),
            Value<String?> companyBankAccountId = const Value.absent(),
            Value<String?> visitId = const Value.absent(),
            Value<double?> latitude = const Value.absent(),
            Value<double?> longitude = const Value.absent(),
            Value<String?> syncErrorMessage = const Value.absent(),
            Value<DateTime?> lastSyncAttempt = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              LocalTransactionsCompanion(
            localId: localId,
            receiptNo: receiptNo,
            companyId: companyId,
            companyName: companyName,
            storeId: storeId,
            storeName: storeName,
            totalAmount: totalAmount,
            paidAmount: paidAmount,
            selfiePath: selfiePath,
            receiptPath: receiptPath,
            signaturePath: signaturePath,
            notes: notes,
            paymentMethod: paymentMethod,
            dpPaymentMethod: dpPaymentMethod,
            paymentBank: paymentBank,
            midtransId: midtransId,
            midtransQrisUrl: midtransQrisUrl,
            midtransVaNumber: midtransVaNumber,
            midtransBank: midtransBank,
            midtransBillKey: midtransBillKey,
            midtransBillerCode: midtransBillerCode,
            createdAt: createdAt,
            syncStatus: syncStatus,
            paymentDueDate: paymentDueDate,
            tenor: tenor,
            isMt: isMt,
            isDistributorOrder: isDistributorOrder,
            factoryCompanyId: factoryCompanyId,
            companyBankAccountId: companyBankAccountId,
            visitId: visitId,
            latitude: latitude,
            longitude: longitude,
            syncErrorMessage: syncErrorMessage,
            lastSyncAttempt: lastSyncAttempt,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String localId,
            Value<String?> receiptNo = const Value.absent(),
            required String companyId,
            required String companyName,
            required String storeId,
            required String storeName,
            required double totalAmount,
            Value<double> paidAmount = const Value.absent(),
            Value<String?> selfiePath = const Value.absent(),
            Value<String?> receiptPath = const Value.absent(),
            Value<String?> signaturePath = const Value.absent(),
            Value<String?> notes = const Value.absent(),
            Value<String> paymentMethod = const Value.absent(),
            Value<String?> dpPaymentMethod = const Value.absent(),
            Value<String?> paymentBank = const Value.absent(),
            Value<String?> midtransId = const Value.absent(),
            Value<String?> midtransQrisUrl = const Value.absent(),
            Value<String?> midtransVaNumber = const Value.absent(),
            Value<String?> midtransBank = const Value.absent(),
            Value<String?> midtransBillKey = const Value.absent(),
            Value<String?> midtransBillerCode = const Value.absent(),
            required DateTime createdAt,
            Value<String> syncStatus = const Value.absent(),
            Value<DateTime?> paymentDueDate = const Value.absent(),
            Value<int> tenor = const Value.absent(),
            Value<bool> isMt = const Value.absent(),
            Value<bool> isDistributorOrder = const Value.absent(),
            Value<String?> factoryCompanyId = const Value.absent(),
            Value<String?> companyBankAccountId = const Value.absent(),
            Value<String?> visitId = const Value.absent(),
            Value<double?> latitude = const Value.absent(),
            Value<double?> longitude = const Value.absent(),
            Value<String?> syncErrorMessage = const Value.absent(),
            Value<DateTime?> lastSyncAttempt = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              LocalTransactionsCompanion.insert(
            localId: localId,
            receiptNo: receiptNo,
            companyId: companyId,
            companyName: companyName,
            storeId: storeId,
            storeName: storeName,
            totalAmount: totalAmount,
            paidAmount: paidAmount,
            selfiePath: selfiePath,
            receiptPath: receiptPath,
            signaturePath: signaturePath,
            notes: notes,
            paymentMethod: paymentMethod,
            dpPaymentMethod: dpPaymentMethod,
            paymentBank: paymentBank,
            midtransId: midtransId,
            midtransQrisUrl: midtransQrisUrl,
            midtransVaNumber: midtransVaNumber,
            midtransBank: midtransBank,
            midtransBillKey: midtransBillKey,
            midtransBillerCode: midtransBillerCode,
            createdAt: createdAt,
            syncStatus: syncStatus,
            paymentDueDate: paymentDueDate,
            tenor: tenor,
            isMt: isMt,
            isDistributorOrder: isDistributorOrder,
            factoryCompanyId: factoryCompanyId,
            companyBankAccountId: companyBankAccountId,
            visitId: visitId,
            latitude: latitude,
            longitude: longitude,
            syncErrorMessage: syncErrorMessage,
            lastSyncAttempt: lastSyncAttempt,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (
                    e.readTable(table),
                    $$LocalTransactionsTableReferences(db, table, e)
                  ))
              .toList(),
          prefetchHooksCallback: ({localTransactionItemsRefs = false}) {
            return PrefetchHooks(
              db: db,
              explicitlyWatchedTables: [
                if (localTransactionItemsRefs) db.localTransactionItems
              ],
              addJoins: null,
              getPrefetchedDataCallback: (items) async {
                return [
                  if (localTransactionItemsRefs)
                    await $_getPrefetchedData<LocalTransaction,
                            $LocalTransactionsTable, LocalTransactionItem>(
                        currentTable: table,
                        referencedTable: $$LocalTransactionsTableReferences
                            ._localTransactionItemsRefsTable(db),
                        managerFromTypedResult: (p0) =>
                            $$LocalTransactionsTableReferences(db, table, p0)
                                .localTransactionItemsRefs,
                        referencedItemsForCurrentItem:
                            (item, referencedItems) => referencedItems.where(
                                (e) => e.transactionLocalId == item.localId),
                        typedResults: items)
                ];
              },
            );
          },
        ));
}

typedef $$LocalTransactionsTableProcessedTableManager = ProcessedTableManager<
    _$AppDatabase,
    $LocalTransactionsTable,
    LocalTransaction,
    $$LocalTransactionsTableFilterComposer,
    $$LocalTransactionsTableOrderingComposer,
    $$LocalTransactionsTableAnnotationComposer,
    $$LocalTransactionsTableCreateCompanionBuilder,
    $$LocalTransactionsTableUpdateCompanionBuilder,
    (LocalTransaction, $$LocalTransactionsTableReferences),
    LocalTransaction,
    PrefetchHooks Function({bool localTransactionItemsRefs})>;
typedef $$LocalTransactionItemsTableCreateCompanionBuilder
    = LocalTransactionItemsCompanion Function({
  Value<int> id,
  required String transactionLocalId,
  required String productId,
  required String productName,
  required int quantity,
  Value<int> orderedQuantity,
  Value<String> unit,
  Value<int> piecesPerUnit,
  required double price,
  Value<double> discountPercent,
});
typedef $$LocalTransactionItemsTableUpdateCompanionBuilder
    = LocalTransactionItemsCompanion Function({
  Value<int> id,
  Value<String> transactionLocalId,
  Value<String> productId,
  Value<String> productName,
  Value<int> quantity,
  Value<int> orderedQuantity,
  Value<String> unit,
  Value<int> piecesPerUnit,
  Value<double> price,
  Value<double> discountPercent,
});

final class $$LocalTransactionItemsTableReferences extends BaseReferences<
    _$AppDatabase, $LocalTransactionItemsTable, LocalTransactionItem> {
  $$LocalTransactionItemsTableReferences(
      super.$_db, super.$_table, super.$_typedResult);

  static $LocalTransactionsTable _transactionLocalIdTable(_$AppDatabase db) =>
      db.localTransactions.createAlias(
          'local_transaction_items__transaction_local_id__local_transactions__local_id');

  $$LocalTransactionsTableProcessedTableManager get transactionLocalId {
    final $_column = $_itemColumn<String>('transaction_local_id')!;

    final manager =
        $$LocalTransactionsTableTableManager($_db, $_db.localTransactions)
            .filter((f) => f.localId.sqlEquals($_column));
    final item = $_typedResult.readTableOrNull(_transactionLocalIdTable($_db));
    if (item == null) return manager;
    return ProcessedTableManager(
        manager.$state.copyWith(prefetchedData: [item]));
  }
}

class $$LocalTransactionItemsTableFilterComposer
    extends Composer<_$AppDatabase, $LocalTransactionItemsTable> {
  $$LocalTransactionItemsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<int> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get productId => $composableBuilder(
      column: $table.productId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get productName => $composableBuilder(
      column: $table.productName, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get quantity => $composableBuilder(
      column: $table.quantity, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get orderedQuantity => $composableBuilder(
      column: $table.orderedQuantity,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get unit => $composableBuilder(
      column: $table.unit, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get piecesPerUnit => $composableBuilder(
      column: $table.piecesPerUnit, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get price => $composableBuilder(
      column: $table.price, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get discountPercent => $composableBuilder(
      column: $table.discountPercent,
      builder: (column) => ColumnFilters(column));

  $$LocalTransactionsTableFilterComposer get transactionLocalId {
    final $$LocalTransactionsTableFilterComposer composer = $composerBuilder(
        composer: this,
        getCurrentColumn: (t) => t.transactionLocalId,
        referencedTable: $db.localTransactions,
        getReferencedColumn: (t) => t.localId,
        builder: (joinBuilder,
                {$addJoinBuilderToRootComposer,
                $removeJoinBuilderFromRootComposer}) =>
            $$LocalTransactionsTableFilterComposer(
              $db: $db,
              $table: $db.localTransactions,
              $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
              joinBuilder: joinBuilder,
              $removeJoinBuilderFromRootComposer:
                  $removeJoinBuilderFromRootComposer,
            ));
    return composer;
  }
}

class $$LocalTransactionItemsTableOrderingComposer
    extends Composer<_$AppDatabase, $LocalTransactionItemsTable> {
  $$LocalTransactionItemsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<int> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get productId => $composableBuilder(
      column: $table.productId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get productName => $composableBuilder(
      column: $table.productName, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get quantity => $composableBuilder(
      column: $table.quantity, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get orderedQuantity => $composableBuilder(
      column: $table.orderedQuantity,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get unit => $composableBuilder(
      column: $table.unit, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get piecesPerUnit => $composableBuilder(
      column: $table.piecesPerUnit,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get price => $composableBuilder(
      column: $table.price, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get discountPercent => $composableBuilder(
      column: $table.discountPercent,
      builder: (column) => ColumnOrderings(column));

  $$LocalTransactionsTableOrderingComposer get transactionLocalId {
    final $$LocalTransactionsTableOrderingComposer composer = $composerBuilder(
        composer: this,
        getCurrentColumn: (t) => t.transactionLocalId,
        referencedTable: $db.localTransactions,
        getReferencedColumn: (t) => t.localId,
        builder: (joinBuilder,
                {$addJoinBuilderToRootComposer,
                $removeJoinBuilderFromRootComposer}) =>
            $$LocalTransactionsTableOrderingComposer(
              $db: $db,
              $table: $db.localTransactions,
              $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
              joinBuilder: joinBuilder,
              $removeJoinBuilderFromRootComposer:
                  $removeJoinBuilderFromRootComposer,
            ));
    return composer;
  }
}

class $$LocalTransactionItemsTableAnnotationComposer
    extends Composer<_$AppDatabase, $LocalTransactionItemsTable> {
  $$LocalTransactionItemsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<int> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get productId =>
      $composableBuilder(column: $table.productId, builder: (column) => column);

  GeneratedColumn<String> get productName => $composableBuilder(
      column: $table.productName, builder: (column) => column);

  GeneratedColumn<int> get quantity =>
      $composableBuilder(column: $table.quantity, builder: (column) => column);

  GeneratedColumn<int> get orderedQuantity => $composableBuilder(
      column: $table.orderedQuantity, builder: (column) => column);

  GeneratedColumn<String> get unit =>
      $composableBuilder(column: $table.unit, builder: (column) => column);

  GeneratedColumn<int> get piecesPerUnit => $composableBuilder(
      column: $table.piecesPerUnit, builder: (column) => column);

  GeneratedColumn<double> get price =>
      $composableBuilder(column: $table.price, builder: (column) => column);

  GeneratedColumn<double> get discountPercent => $composableBuilder(
      column: $table.discountPercent, builder: (column) => column);

  $$LocalTransactionsTableAnnotationComposer get transactionLocalId {
    final $$LocalTransactionsTableAnnotationComposer composer =
        $composerBuilder(
            composer: this,
            getCurrentColumn: (t) => t.transactionLocalId,
            referencedTable: $db.localTransactions,
            getReferencedColumn: (t) => t.localId,
            builder: (joinBuilder,
                    {$addJoinBuilderToRootComposer,
                    $removeJoinBuilderFromRootComposer}) =>
                $$LocalTransactionsTableAnnotationComposer(
                  $db: $db,
                  $table: $db.localTransactions,
                  $addJoinBuilderToRootComposer: $addJoinBuilderToRootComposer,
                  joinBuilder: joinBuilder,
                  $removeJoinBuilderFromRootComposer:
                      $removeJoinBuilderFromRootComposer,
                ));
    return composer;
  }
}

class $$LocalTransactionItemsTableTableManager extends RootTableManager<
    _$AppDatabase,
    $LocalTransactionItemsTable,
    LocalTransactionItem,
    $$LocalTransactionItemsTableFilterComposer,
    $$LocalTransactionItemsTableOrderingComposer,
    $$LocalTransactionItemsTableAnnotationComposer,
    $$LocalTransactionItemsTableCreateCompanionBuilder,
    $$LocalTransactionItemsTableUpdateCompanionBuilder,
    (LocalTransactionItem, $$LocalTransactionItemsTableReferences),
    LocalTransactionItem,
    PrefetchHooks Function({bool transactionLocalId})> {
  $$LocalTransactionItemsTableTableManager(
      _$AppDatabase db, $LocalTransactionItemsTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$LocalTransactionItemsTableFilterComposer(
                  $db: db, $table: table),
          createOrderingComposer: () =>
              $$LocalTransactionItemsTableOrderingComposer(
                  $db: db, $table: table),
          createComputedFieldComposer: () =>
              $$LocalTransactionItemsTableAnnotationComposer(
                  $db: db, $table: table),
          updateCompanionCallback: ({
            Value<int> id = const Value.absent(),
            Value<String> transactionLocalId = const Value.absent(),
            Value<String> productId = const Value.absent(),
            Value<String> productName = const Value.absent(),
            Value<int> quantity = const Value.absent(),
            Value<int> orderedQuantity = const Value.absent(),
            Value<String> unit = const Value.absent(),
            Value<int> piecesPerUnit = const Value.absent(),
            Value<double> price = const Value.absent(),
            Value<double> discountPercent = const Value.absent(),
          }) =>
              LocalTransactionItemsCompanion(
            id: id,
            transactionLocalId: transactionLocalId,
            productId: productId,
            productName: productName,
            quantity: quantity,
            orderedQuantity: orderedQuantity,
            unit: unit,
            piecesPerUnit: piecesPerUnit,
            price: price,
            discountPercent: discountPercent,
          ),
          createCompanionCallback: ({
            Value<int> id = const Value.absent(),
            required String transactionLocalId,
            required String productId,
            required String productName,
            required int quantity,
            Value<int> orderedQuantity = const Value.absent(),
            Value<String> unit = const Value.absent(),
            Value<int> piecesPerUnit = const Value.absent(),
            required double price,
            Value<double> discountPercent = const Value.absent(),
          }) =>
              LocalTransactionItemsCompanion.insert(
            id: id,
            transactionLocalId: transactionLocalId,
            productId: productId,
            productName: productName,
            quantity: quantity,
            orderedQuantity: orderedQuantity,
            unit: unit,
            piecesPerUnit: piecesPerUnit,
            price: price,
            discountPercent: discountPercent,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (
                    e.readTable(table),
                    $$LocalTransactionItemsTableReferences(db, table, e)
                  ))
              .toList(),
          prefetchHooksCallback: ({transactionLocalId = false}) {
            return PrefetchHooks(
              db: db,
              explicitlyWatchedTables: [],
              addJoins: <
                  T extends TableManagerState<
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic,
                      dynamic>>(state) {
                if (transactionLocalId) {
                  state = state.withJoin(
                    currentTable: table,
                    currentColumn: table.transactionLocalId,
                    referencedTable: $$LocalTransactionItemsTableReferences
                        ._transactionLocalIdTable(db),
                    referencedColumn: $$LocalTransactionItemsTableReferences
                        ._transactionLocalIdTable(db)
                        .localId,
                  ) as T;
                }

                return state;
              },
              getPrefetchedDataCallback: (items) async {
                return [];
              },
            );
          },
        ));
}

typedef $$LocalTransactionItemsTableProcessedTableManager
    = ProcessedTableManager<
        _$AppDatabase,
        $LocalTransactionItemsTable,
        LocalTransactionItem,
        $$LocalTransactionItemsTableFilterComposer,
        $$LocalTransactionItemsTableOrderingComposer,
        $$LocalTransactionItemsTableAnnotationComposer,
        $$LocalTransactionItemsTableCreateCompanionBuilder,
        $$LocalTransactionItemsTableUpdateCompanionBuilder,
        (LocalTransactionItem, $$LocalTransactionItemsTableReferences),
        LocalTransactionItem,
        PrefetchHooks Function({bool transactionLocalId})>;
typedef $$SalesStockTableCreateCompanionBuilder = SalesStockCompanion Function({
  required String productId,
  required int quantity,
  required DateTime updatedAt,
  Value<int> rowid,
});
typedef $$SalesStockTableUpdateCompanionBuilder = SalesStockCompanion Function({
  Value<String> productId,
  Value<int> quantity,
  Value<DateTime> updatedAt,
  Value<int> rowid,
});

class $$SalesStockTableFilterComposer
    extends Composer<_$AppDatabase, $SalesStockTable> {
  $$SalesStockTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get productId => $composableBuilder(
      column: $table.productId, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get quantity => $composableBuilder(
      column: $table.quantity, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get updatedAt => $composableBuilder(
      column: $table.updatedAt, builder: (column) => ColumnFilters(column));
}

class $$SalesStockTableOrderingComposer
    extends Composer<_$AppDatabase, $SalesStockTable> {
  $$SalesStockTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get productId => $composableBuilder(
      column: $table.productId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get quantity => $composableBuilder(
      column: $table.quantity, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get updatedAt => $composableBuilder(
      column: $table.updatedAt, builder: (column) => ColumnOrderings(column));
}

class $$SalesStockTableAnnotationComposer
    extends Composer<_$AppDatabase, $SalesStockTable> {
  $$SalesStockTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get productId =>
      $composableBuilder(column: $table.productId, builder: (column) => column);

  GeneratedColumn<int> get quantity =>
      $composableBuilder(column: $table.quantity, builder: (column) => column);

  GeneratedColumn<DateTime> get updatedAt =>
      $composableBuilder(column: $table.updatedAt, builder: (column) => column);
}

class $$SalesStockTableTableManager extends RootTableManager<
    _$AppDatabase,
    $SalesStockTable,
    SalesStockData,
    $$SalesStockTableFilterComposer,
    $$SalesStockTableOrderingComposer,
    $$SalesStockTableAnnotationComposer,
    $$SalesStockTableCreateCompanionBuilder,
    $$SalesStockTableUpdateCompanionBuilder,
    (
      SalesStockData,
      BaseReferences<_$AppDatabase, $SalesStockTable, SalesStockData>
    ),
    SalesStockData,
    PrefetchHooks Function()> {
  $$SalesStockTableTableManager(_$AppDatabase db, $SalesStockTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$SalesStockTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$SalesStockTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$SalesStockTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> productId = const Value.absent(),
            Value<int> quantity = const Value.absent(),
            Value<DateTime> updatedAt = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              SalesStockCompanion(
            productId: productId,
            quantity: quantity,
            updatedAt: updatedAt,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String productId,
            required int quantity,
            required DateTime updatedAt,
            Value<int> rowid = const Value.absent(),
          }) =>
              SalesStockCompanion.insert(
            productId: productId,
            quantity: quantity,
            updatedAt: updatedAt,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$SalesStockTableProcessedTableManager = ProcessedTableManager<
    _$AppDatabase,
    $SalesStockTable,
    SalesStockData,
    $$SalesStockTableFilterComposer,
    $$SalesStockTableOrderingComposer,
    $$SalesStockTableAnnotationComposer,
    $$SalesStockTableCreateCompanionBuilder,
    $$SalesStockTableUpdateCompanionBuilder,
    (
      SalesStockData,
      BaseReferences<_$AppDatabase, $SalesStockTable, SalesStockData>
    ),
    SalesStockData,
    PrefetchHooks Function()>;
typedef $$OfflineTasksTableCreateCompanionBuilder = OfflineTasksCompanion
    Function({
  required String id,
  required String type,
  required String payload,
  Value<String?> filePath,
  Value<String?> filePath2,
  required DateTime createdAt,
  Value<int> retryCount,
  Value<String> status,
  Value<int> rowid,
});
typedef $$OfflineTasksTableUpdateCompanionBuilder = OfflineTasksCompanion
    Function({
  Value<String> id,
  Value<String> type,
  Value<String> payload,
  Value<String?> filePath,
  Value<String?> filePath2,
  Value<DateTime> createdAt,
  Value<int> retryCount,
  Value<String> status,
  Value<int> rowid,
});

class $$OfflineTasksTableFilterComposer
    extends Composer<_$AppDatabase, $OfflineTasksTable> {
  $$OfflineTasksTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get type => $composableBuilder(
      column: $table.type, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get payload => $composableBuilder(
      column: $table.payload, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get filePath => $composableBuilder(
      column: $table.filePath, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get filePath2 => $composableBuilder(
      column: $table.filePath2, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get retryCount => $composableBuilder(
      column: $table.retryCount, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnFilters(column));
}

class $$OfflineTasksTableOrderingComposer
    extends Composer<_$AppDatabase, $OfflineTasksTable> {
  $$OfflineTasksTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get type => $composableBuilder(
      column: $table.type, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get payload => $composableBuilder(
      column: $table.payload, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get filePath => $composableBuilder(
      column: $table.filePath, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get filePath2 => $composableBuilder(
      column: $table.filePath2, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get retryCount => $composableBuilder(
      column: $table.retryCount, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnOrderings(column));
}

class $$OfflineTasksTableAnnotationComposer
    extends Composer<_$AppDatabase, $OfflineTasksTable> {
  $$OfflineTasksTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get type =>
      $composableBuilder(column: $table.type, builder: (column) => column);

  GeneratedColumn<String> get payload =>
      $composableBuilder(column: $table.payload, builder: (column) => column);

  GeneratedColumn<String> get filePath =>
      $composableBuilder(column: $table.filePath, builder: (column) => column);

  GeneratedColumn<String> get filePath2 =>
      $composableBuilder(column: $table.filePath2, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<int> get retryCount => $composableBuilder(
      column: $table.retryCount, builder: (column) => column);

  GeneratedColumn<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => column);
}

class $$OfflineTasksTableTableManager extends RootTableManager<
    _$AppDatabase,
    $OfflineTasksTable,
    OfflineTask,
    $$OfflineTasksTableFilterComposer,
    $$OfflineTasksTableOrderingComposer,
    $$OfflineTasksTableAnnotationComposer,
    $$OfflineTasksTableCreateCompanionBuilder,
    $$OfflineTasksTableUpdateCompanionBuilder,
    (
      OfflineTask,
      BaseReferences<_$AppDatabase, $OfflineTasksTable, OfflineTask>
    ),
    OfflineTask,
    PrefetchHooks Function()> {
  $$OfflineTasksTableTableManager(_$AppDatabase db, $OfflineTasksTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$OfflineTasksTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$OfflineTasksTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$OfflineTasksTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> id = const Value.absent(),
            Value<String> type = const Value.absent(),
            Value<String> payload = const Value.absent(),
            Value<String?> filePath = const Value.absent(),
            Value<String?> filePath2 = const Value.absent(),
            Value<DateTime> createdAt = const Value.absent(),
            Value<int> retryCount = const Value.absent(),
            Value<String> status = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              OfflineTasksCompanion(
            id: id,
            type: type,
            payload: payload,
            filePath: filePath,
            filePath2: filePath2,
            createdAt: createdAt,
            retryCount: retryCount,
            status: status,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String id,
            required String type,
            required String payload,
            Value<String?> filePath = const Value.absent(),
            Value<String?> filePath2 = const Value.absent(),
            required DateTime createdAt,
            Value<int> retryCount = const Value.absent(),
            Value<String> status = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              OfflineTasksCompanion.insert(
            id: id,
            type: type,
            payload: payload,
            filePath: filePath,
            filePath2: filePath2,
            createdAt: createdAt,
            retryCount: retryCount,
            status: status,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$OfflineTasksTableProcessedTableManager = ProcessedTableManager<
    _$AppDatabase,
    $OfflineTasksTable,
    OfflineTask,
    $$OfflineTasksTableFilterComposer,
    $$OfflineTasksTableOrderingComposer,
    $$OfflineTasksTableAnnotationComposer,
    $$OfflineTasksTableCreateCompanionBuilder,
    $$OfflineTasksTableUpdateCompanionBuilder,
    (
      OfflineTask,
      BaseReferences<_$AppDatabase, $OfflineTasksTable, OfflineTask>
    ),
    OfflineTask,
    PrefetchHooks Function()>;
typedef $$StoresTableCreateCompanionBuilder = StoresCompanion Function({
  required String id,
  Value<String?> code,
  required String name,
  Value<String?> address,
  Value<double?> latitude,
  Value<double?> longitude,
  Value<String?> ownerName,
  Value<String?> phoneNumber,
  Value<String?> photoUrl,
  Value<String?> localPhotoPath,
  Value<String?> province,
  Value<String?> regency,
  Value<String?> district,
  Value<String?> village,
  Value<String?> visitFrequency,
  Value<String?> visitDays,
  Value<String?> assignedEmployeeId,
  Value<String?> companyId,
  Value<bool> isOffline,
  Value<String> syncStatus,
  Value<String?> syncErrorMessage,
  Value<DateTime?> createdAt,
  Value<double?> accuracy,
  Value<int> rowid,
});
typedef $$StoresTableUpdateCompanionBuilder = StoresCompanion Function({
  Value<String> id,
  Value<String?> code,
  Value<String> name,
  Value<String?> address,
  Value<double?> latitude,
  Value<double?> longitude,
  Value<String?> ownerName,
  Value<String?> phoneNumber,
  Value<String?> photoUrl,
  Value<String?> localPhotoPath,
  Value<String?> province,
  Value<String?> regency,
  Value<String?> district,
  Value<String?> village,
  Value<String?> visitFrequency,
  Value<String?> visitDays,
  Value<String?> assignedEmployeeId,
  Value<String?> companyId,
  Value<bool> isOffline,
  Value<String> syncStatus,
  Value<String?> syncErrorMessage,
  Value<DateTime?> createdAt,
  Value<double?> accuracy,
  Value<int> rowid,
});

class $$StoresTableFilterComposer
    extends Composer<_$AppDatabase, $StoresTable> {
  $$StoresTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get code => $composableBuilder(
      column: $table.code, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get name => $composableBuilder(
      column: $table.name, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get address => $composableBuilder(
      column: $table.address, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get latitude => $composableBuilder(
      column: $table.latitude, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get longitude => $composableBuilder(
      column: $table.longitude, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get ownerName => $composableBuilder(
      column: $table.ownerName, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get phoneNumber => $composableBuilder(
      column: $table.phoneNumber, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get photoUrl => $composableBuilder(
      column: $table.photoUrl, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get localPhotoPath => $composableBuilder(
      column: $table.localPhotoPath,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get province => $composableBuilder(
      column: $table.province, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get regency => $composableBuilder(
      column: $table.regency, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get district => $composableBuilder(
      column: $table.district, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get village => $composableBuilder(
      column: $table.village, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get visitFrequency => $composableBuilder(
      column: $table.visitFrequency,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get visitDays => $composableBuilder(
      column: $table.visitDays, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get assignedEmployeeId => $composableBuilder(
      column: $table.assignedEmployeeId,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get companyId => $composableBuilder(
      column: $table.companyId, builder: (column) => ColumnFilters(column));

  ColumnFilters<bool> get isOffline => $composableBuilder(
      column: $table.isOffline, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get syncStatus => $composableBuilder(
      column: $table.syncStatus, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get syncErrorMessage => $composableBuilder(
      column: $table.syncErrorMessage,
      builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get accuracy => $composableBuilder(
      column: $table.accuracy, builder: (column) => ColumnFilters(column));
}

class $$StoresTableOrderingComposer
    extends Composer<_$AppDatabase, $StoresTable> {
  $$StoresTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get code => $composableBuilder(
      column: $table.code, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get name => $composableBuilder(
      column: $table.name, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get address => $composableBuilder(
      column: $table.address, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get latitude => $composableBuilder(
      column: $table.latitude, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get longitude => $composableBuilder(
      column: $table.longitude, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get ownerName => $composableBuilder(
      column: $table.ownerName, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get phoneNumber => $composableBuilder(
      column: $table.phoneNumber, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get photoUrl => $composableBuilder(
      column: $table.photoUrl, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get localPhotoPath => $composableBuilder(
      column: $table.localPhotoPath,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get province => $composableBuilder(
      column: $table.province, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get regency => $composableBuilder(
      column: $table.regency, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get district => $composableBuilder(
      column: $table.district, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get village => $composableBuilder(
      column: $table.village, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get visitFrequency => $composableBuilder(
      column: $table.visitFrequency,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get visitDays => $composableBuilder(
      column: $table.visitDays, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get assignedEmployeeId => $composableBuilder(
      column: $table.assignedEmployeeId,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get companyId => $composableBuilder(
      column: $table.companyId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<bool> get isOffline => $composableBuilder(
      column: $table.isOffline, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get syncStatus => $composableBuilder(
      column: $table.syncStatus, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get syncErrorMessage => $composableBuilder(
      column: $table.syncErrorMessage,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get accuracy => $composableBuilder(
      column: $table.accuracy, builder: (column) => ColumnOrderings(column));
}

class $$StoresTableAnnotationComposer
    extends Composer<_$AppDatabase, $StoresTable> {
  $$StoresTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get code =>
      $composableBuilder(column: $table.code, builder: (column) => column);

  GeneratedColumn<String> get name =>
      $composableBuilder(column: $table.name, builder: (column) => column);

  GeneratedColumn<String> get address =>
      $composableBuilder(column: $table.address, builder: (column) => column);

  GeneratedColumn<double> get latitude =>
      $composableBuilder(column: $table.latitude, builder: (column) => column);

  GeneratedColumn<double> get longitude =>
      $composableBuilder(column: $table.longitude, builder: (column) => column);

  GeneratedColumn<String> get ownerName =>
      $composableBuilder(column: $table.ownerName, builder: (column) => column);

  GeneratedColumn<String> get phoneNumber => $composableBuilder(
      column: $table.phoneNumber, builder: (column) => column);

  GeneratedColumn<String> get photoUrl =>
      $composableBuilder(column: $table.photoUrl, builder: (column) => column);

  GeneratedColumn<String> get localPhotoPath => $composableBuilder(
      column: $table.localPhotoPath, builder: (column) => column);

  GeneratedColumn<String> get province =>
      $composableBuilder(column: $table.province, builder: (column) => column);

  GeneratedColumn<String> get regency =>
      $composableBuilder(column: $table.regency, builder: (column) => column);

  GeneratedColumn<String> get district =>
      $composableBuilder(column: $table.district, builder: (column) => column);

  GeneratedColumn<String> get village =>
      $composableBuilder(column: $table.village, builder: (column) => column);

  GeneratedColumn<String> get visitFrequency => $composableBuilder(
      column: $table.visitFrequency, builder: (column) => column);

  GeneratedColumn<String> get visitDays =>
      $composableBuilder(column: $table.visitDays, builder: (column) => column);

  GeneratedColumn<String> get assignedEmployeeId => $composableBuilder(
      column: $table.assignedEmployeeId, builder: (column) => column);

  GeneratedColumn<String> get companyId =>
      $composableBuilder(column: $table.companyId, builder: (column) => column);

  GeneratedColumn<bool> get isOffline =>
      $composableBuilder(column: $table.isOffline, builder: (column) => column);

  GeneratedColumn<String> get syncStatus => $composableBuilder(
      column: $table.syncStatus, builder: (column) => column);

  GeneratedColumn<String> get syncErrorMessage => $composableBuilder(
      column: $table.syncErrorMessage, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<double> get accuracy =>
      $composableBuilder(column: $table.accuracy, builder: (column) => column);
}

class $$StoresTableTableManager extends RootTableManager<
    _$AppDatabase,
    $StoresTable,
    Store,
    $$StoresTableFilterComposer,
    $$StoresTableOrderingComposer,
    $$StoresTableAnnotationComposer,
    $$StoresTableCreateCompanionBuilder,
    $$StoresTableUpdateCompanionBuilder,
    (Store, BaseReferences<_$AppDatabase, $StoresTable, Store>),
    Store,
    PrefetchHooks Function()> {
  $$StoresTableTableManager(_$AppDatabase db, $StoresTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$StoresTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$StoresTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$StoresTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> id = const Value.absent(),
            Value<String?> code = const Value.absent(),
            Value<String> name = const Value.absent(),
            Value<String?> address = const Value.absent(),
            Value<double?> latitude = const Value.absent(),
            Value<double?> longitude = const Value.absent(),
            Value<String?> ownerName = const Value.absent(),
            Value<String?> phoneNumber = const Value.absent(),
            Value<String?> photoUrl = const Value.absent(),
            Value<String?> localPhotoPath = const Value.absent(),
            Value<String?> province = const Value.absent(),
            Value<String?> regency = const Value.absent(),
            Value<String?> district = const Value.absent(),
            Value<String?> village = const Value.absent(),
            Value<String?> visitFrequency = const Value.absent(),
            Value<String?> visitDays = const Value.absent(),
            Value<String?> assignedEmployeeId = const Value.absent(),
            Value<String?> companyId = const Value.absent(),
            Value<bool> isOffline = const Value.absent(),
            Value<String> syncStatus = const Value.absent(),
            Value<String?> syncErrorMessage = const Value.absent(),
            Value<DateTime?> createdAt = const Value.absent(),
            Value<double?> accuracy = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              StoresCompanion(
            id: id,
            code: code,
            name: name,
            address: address,
            latitude: latitude,
            longitude: longitude,
            ownerName: ownerName,
            phoneNumber: phoneNumber,
            photoUrl: photoUrl,
            localPhotoPath: localPhotoPath,
            province: province,
            regency: regency,
            district: district,
            village: village,
            visitFrequency: visitFrequency,
            visitDays: visitDays,
            assignedEmployeeId: assignedEmployeeId,
            companyId: companyId,
            isOffline: isOffline,
            syncStatus: syncStatus,
            syncErrorMessage: syncErrorMessage,
            createdAt: createdAt,
            accuracy: accuracy,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String id,
            Value<String?> code = const Value.absent(),
            required String name,
            Value<String?> address = const Value.absent(),
            Value<double?> latitude = const Value.absent(),
            Value<double?> longitude = const Value.absent(),
            Value<String?> ownerName = const Value.absent(),
            Value<String?> phoneNumber = const Value.absent(),
            Value<String?> photoUrl = const Value.absent(),
            Value<String?> localPhotoPath = const Value.absent(),
            Value<String?> province = const Value.absent(),
            Value<String?> regency = const Value.absent(),
            Value<String?> district = const Value.absent(),
            Value<String?> village = const Value.absent(),
            Value<String?> visitFrequency = const Value.absent(),
            Value<String?> visitDays = const Value.absent(),
            Value<String?> assignedEmployeeId = const Value.absent(),
            Value<String?> companyId = const Value.absent(),
            Value<bool> isOffline = const Value.absent(),
            Value<String> syncStatus = const Value.absent(),
            Value<String?> syncErrorMessage = const Value.absent(),
            Value<DateTime?> createdAt = const Value.absent(),
            Value<double?> accuracy = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              StoresCompanion.insert(
            id: id,
            code: code,
            name: name,
            address: address,
            latitude: latitude,
            longitude: longitude,
            ownerName: ownerName,
            phoneNumber: phoneNumber,
            photoUrl: photoUrl,
            localPhotoPath: localPhotoPath,
            province: province,
            regency: regency,
            district: district,
            village: village,
            visitFrequency: visitFrequency,
            visitDays: visitDays,
            assignedEmployeeId: assignedEmployeeId,
            companyId: companyId,
            isOffline: isOffline,
            syncStatus: syncStatus,
            syncErrorMessage: syncErrorMessage,
            createdAt: createdAt,
            accuracy: accuracy,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$StoresTableProcessedTableManager = ProcessedTableManager<
    _$AppDatabase,
    $StoresTable,
    Store,
    $$StoresTableFilterComposer,
    $$StoresTableOrderingComposer,
    $$StoresTableAnnotationComposer,
    $$StoresTableCreateCompanionBuilder,
    $$StoresTableUpdateCompanionBuilder,
    (Store, BaseReferences<_$AppDatabase, $StoresTable, Store>),
    Store,
    PrefetchHooks Function()>;
typedef $$VisitPlansTableCreateCompanionBuilder = VisitPlansCompanion Function({
  required String id,
  required String storeId,
  required String storeName,
  required String date,
  Value<String> status,
  Value<int> rowid,
});
typedef $$VisitPlansTableUpdateCompanionBuilder = VisitPlansCompanion Function({
  Value<String> id,
  Value<String> storeId,
  Value<String> storeName,
  Value<String> date,
  Value<String> status,
  Value<int> rowid,
});

class $$VisitPlansTableFilterComposer
    extends Composer<_$AppDatabase, $VisitPlansTable> {
  $$VisitPlansTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get storeId => $composableBuilder(
      column: $table.storeId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get storeName => $composableBuilder(
      column: $table.storeName, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get date => $composableBuilder(
      column: $table.date, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnFilters(column));
}

class $$VisitPlansTableOrderingComposer
    extends Composer<_$AppDatabase, $VisitPlansTable> {
  $$VisitPlansTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get storeId => $composableBuilder(
      column: $table.storeId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get storeName => $composableBuilder(
      column: $table.storeName, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get date => $composableBuilder(
      column: $table.date, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnOrderings(column));
}

class $$VisitPlansTableAnnotationComposer
    extends Composer<_$AppDatabase, $VisitPlansTable> {
  $$VisitPlansTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get storeId =>
      $composableBuilder(column: $table.storeId, builder: (column) => column);

  GeneratedColumn<String> get storeName =>
      $composableBuilder(column: $table.storeName, builder: (column) => column);

  GeneratedColumn<String> get date =>
      $composableBuilder(column: $table.date, builder: (column) => column);

  GeneratedColumn<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => column);
}

class $$VisitPlansTableTableManager extends RootTableManager<
    _$AppDatabase,
    $VisitPlansTable,
    VisitPlan,
    $$VisitPlansTableFilterComposer,
    $$VisitPlansTableOrderingComposer,
    $$VisitPlansTableAnnotationComposer,
    $$VisitPlansTableCreateCompanionBuilder,
    $$VisitPlansTableUpdateCompanionBuilder,
    (VisitPlan, BaseReferences<_$AppDatabase, $VisitPlansTable, VisitPlan>),
    VisitPlan,
    PrefetchHooks Function()> {
  $$VisitPlansTableTableManager(_$AppDatabase db, $VisitPlansTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$VisitPlansTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$VisitPlansTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$VisitPlansTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> id = const Value.absent(),
            Value<String> storeId = const Value.absent(),
            Value<String> storeName = const Value.absent(),
            Value<String> date = const Value.absent(),
            Value<String> status = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              VisitPlansCompanion(
            id: id,
            storeId: storeId,
            storeName: storeName,
            date: date,
            status: status,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String id,
            required String storeId,
            required String storeName,
            required String date,
            Value<String> status = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              VisitPlansCompanion.insert(
            id: id,
            storeId: storeId,
            storeName: storeName,
            date: date,
            status: status,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$VisitPlansTableProcessedTableManager = ProcessedTableManager<
    _$AppDatabase,
    $VisitPlansTable,
    VisitPlan,
    $$VisitPlansTableFilterComposer,
    $$VisitPlansTableOrderingComposer,
    $$VisitPlansTableAnnotationComposer,
    $$VisitPlansTableCreateCompanionBuilder,
    $$VisitPlansTableUpdateCompanionBuilder,
    (VisitPlan, BaseReferences<_$AppDatabase, $VisitPlansTable, VisitPlan>),
    VisitPlan,
    PrefetchHooks Function()>;
typedef $$SalesDailyJourneysTableCreateCompanionBuilder
    = SalesDailyJourneysCompanion Function({
  required String id,
  required String userId,
  required String date,
  required int startKm,
  required String startPhotoUrl,
  required String startLocation,
  required DateTime startTime,
  Value<int?> endKm,
  Value<String?> endPhotoUrl,
  Value<String?> endLocation,
  Value<DateTime?> endTime,
  Value<int?> totalKm,
  Value<String> status,
  Value<bool> isSynced,
  Value<int> rowid,
});
typedef $$SalesDailyJourneysTableUpdateCompanionBuilder
    = SalesDailyJourneysCompanion Function({
  Value<String> id,
  Value<String> userId,
  Value<String> date,
  Value<int> startKm,
  Value<String> startPhotoUrl,
  Value<String> startLocation,
  Value<DateTime> startTime,
  Value<int?> endKm,
  Value<String?> endPhotoUrl,
  Value<String?> endLocation,
  Value<DateTime?> endTime,
  Value<int?> totalKm,
  Value<String> status,
  Value<bool> isSynced,
  Value<int> rowid,
});

class $$SalesDailyJourneysTableFilterComposer
    extends Composer<_$AppDatabase, $SalesDailyJourneysTable> {
  $$SalesDailyJourneysTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get userId => $composableBuilder(
      column: $table.userId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get date => $composableBuilder(
      column: $table.date, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get startKm => $composableBuilder(
      column: $table.startKm, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get startPhotoUrl => $composableBuilder(
      column: $table.startPhotoUrl, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get startLocation => $composableBuilder(
      column: $table.startLocation, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get startTime => $composableBuilder(
      column: $table.startTime, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get endKm => $composableBuilder(
      column: $table.endKm, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get endPhotoUrl => $composableBuilder(
      column: $table.endPhotoUrl, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get endLocation => $composableBuilder(
      column: $table.endLocation, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get endTime => $composableBuilder(
      column: $table.endTime, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get totalKm => $composableBuilder(
      column: $table.totalKm, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnFilters(column));

  ColumnFilters<bool> get isSynced => $composableBuilder(
      column: $table.isSynced, builder: (column) => ColumnFilters(column));
}

class $$SalesDailyJourneysTableOrderingComposer
    extends Composer<_$AppDatabase, $SalesDailyJourneysTable> {
  $$SalesDailyJourneysTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get id => $composableBuilder(
      column: $table.id, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get userId => $composableBuilder(
      column: $table.userId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get date => $composableBuilder(
      column: $table.date, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get startKm => $composableBuilder(
      column: $table.startKm, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get startPhotoUrl => $composableBuilder(
      column: $table.startPhotoUrl,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get startLocation => $composableBuilder(
      column: $table.startLocation,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get startTime => $composableBuilder(
      column: $table.startTime, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get endKm => $composableBuilder(
      column: $table.endKm, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get endPhotoUrl => $composableBuilder(
      column: $table.endPhotoUrl, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get endLocation => $composableBuilder(
      column: $table.endLocation, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get endTime => $composableBuilder(
      column: $table.endTime, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get totalKm => $composableBuilder(
      column: $table.totalKm, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<bool> get isSynced => $composableBuilder(
      column: $table.isSynced, builder: (column) => ColumnOrderings(column));
}

class $$SalesDailyJourneysTableAnnotationComposer
    extends Composer<_$AppDatabase, $SalesDailyJourneysTable> {
  $$SalesDailyJourneysTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get id =>
      $composableBuilder(column: $table.id, builder: (column) => column);

  GeneratedColumn<String> get userId =>
      $composableBuilder(column: $table.userId, builder: (column) => column);

  GeneratedColumn<String> get date =>
      $composableBuilder(column: $table.date, builder: (column) => column);

  GeneratedColumn<int> get startKm =>
      $composableBuilder(column: $table.startKm, builder: (column) => column);

  GeneratedColumn<String> get startPhotoUrl => $composableBuilder(
      column: $table.startPhotoUrl, builder: (column) => column);

  GeneratedColumn<String> get startLocation => $composableBuilder(
      column: $table.startLocation, builder: (column) => column);

  GeneratedColumn<DateTime> get startTime =>
      $composableBuilder(column: $table.startTime, builder: (column) => column);

  GeneratedColumn<int> get endKm =>
      $composableBuilder(column: $table.endKm, builder: (column) => column);

  GeneratedColumn<String> get endPhotoUrl => $composableBuilder(
      column: $table.endPhotoUrl, builder: (column) => column);

  GeneratedColumn<String> get endLocation => $composableBuilder(
      column: $table.endLocation, builder: (column) => column);

  GeneratedColumn<DateTime> get endTime =>
      $composableBuilder(column: $table.endTime, builder: (column) => column);

  GeneratedColumn<int> get totalKm =>
      $composableBuilder(column: $table.totalKm, builder: (column) => column);

  GeneratedColumn<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => column);

  GeneratedColumn<bool> get isSynced =>
      $composableBuilder(column: $table.isSynced, builder: (column) => column);
}

class $$SalesDailyJourneysTableTableManager extends RootTableManager<
    _$AppDatabase,
    $SalesDailyJourneysTable,
    SalesDailyJourney,
    $$SalesDailyJourneysTableFilterComposer,
    $$SalesDailyJourneysTableOrderingComposer,
    $$SalesDailyJourneysTableAnnotationComposer,
    $$SalesDailyJourneysTableCreateCompanionBuilder,
    $$SalesDailyJourneysTableUpdateCompanionBuilder,
    (
      SalesDailyJourney,
      BaseReferences<_$AppDatabase, $SalesDailyJourneysTable, SalesDailyJourney>
    ),
    SalesDailyJourney,
    PrefetchHooks Function()> {
  $$SalesDailyJourneysTableTableManager(
      _$AppDatabase db, $SalesDailyJourneysTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$SalesDailyJourneysTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$SalesDailyJourneysTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$SalesDailyJourneysTableAnnotationComposer(
                  $db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> id = const Value.absent(),
            Value<String> userId = const Value.absent(),
            Value<String> date = const Value.absent(),
            Value<int> startKm = const Value.absent(),
            Value<String> startPhotoUrl = const Value.absent(),
            Value<String> startLocation = const Value.absent(),
            Value<DateTime> startTime = const Value.absent(),
            Value<int?> endKm = const Value.absent(),
            Value<String?> endPhotoUrl = const Value.absent(),
            Value<String?> endLocation = const Value.absent(),
            Value<DateTime?> endTime = const Value.absent(),
            Value<int?> totalKm = const Value.absent(),
            Value<String> status = const Value.absent(),
            Value<bool> isSynced = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              SalesDailyJourneysCompanion(
            id: id,
            userId: userId,
            date: date,
            startKm: startKm,
            startPhotoUrl: startPhotoUrl,
            startLocation: startLocation,
            startTime: startTime,
            endKm: endKm,
            endPhotoUrl: endPhotoUrl,
            endLocation: endLocation,
            endTime: endTime,
            totalKm: totalKm,
            status: status,
            isSynced: isSynced,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String id,
            required String userId,
            required String date,
            required int startKm,
            required String startPhotoUrl,
            required String startLocation,
            required DateTime startTime,
            Value<int?> endKm = const Value.absent(),
            Value<String?> endPhotoUrl = const Value.absent(),
            Value<String?> endLocation = const Value.absent(),
            Value<DateTime?> endTime = const Value.absent(),
            Value<int?> totalKm = const Value.absent(),
            Value<String> status = const Value.absent(),
            Value<bool> isSynced = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              SalesDailyJourneysCompanion.insert(
            id: id,
            userId: userId,
            date: date,
            startKm: startKm,
            startPhotoUrl: startPhotoUrl,
            startLocation: startLocation,
            startTime: startTime,
            endKm: endKm,
            endPhotoUrl: endPhotoUrl,
            endLocation: endLocation,
            endTime: endTime,
            totalKm: totalKm,
            status: status,
            isSynced: isSynced,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$SalesDailyJourneysTableProcessedTableManager = ProcessedTableManager<
    _$AppDatabase,
    $SalesDailyJourneysTable,
    SalesDailyJourney,
    $$SalesDailyJourneysTableFilterComposer,
    $$SalesDailyJourneysTableOrderingComposer,
    $$SalesDailyJourneysTableAnnotationComposer,
    $$SalesDailyJourneysTableCreateCompanionBuilder,
    $$SalesDailyJourneysTableUpdateCompanionBuilder,
    (
      SalesDailyJourney,
      BaseReferences<_$AppDatabase, $SalesDailyJourneysTable, SalesDailyJourney>
    ),
    SalesDailyJourney,
    PrefetchHooks Function()>;

class $AppDatabaseManager {
  final _$AppDatabase _db;
  $AppDatabaseManager(this._db);
  $$ProductsTableTableManager get products =>
      $$ProductsTableTableManager(_db, _db.products);
  $$CompaniesTableTableManager get companies =>
      $$CompaniesTableTableManager(_db, _db.companies);
  $$LocalCheckinsTableTableManager get localCheckins =>
      $$LocalCheckinsTableTableManager(_db, _db.localCheckins);
  $$LocalTransactionsTableTableManager get localTransactions =>
      $$LocalTransactionsTableTableManager(_db, _db.localTransactions);
  $$LocalTransactionItemsTableTableManager get localTransactionItems =>
      $$LocalTransactionItemsTableTableManager(_db, _db.localTransactionItems);
  $$SalesStockTableTableManager get salesStock =>
      $$SalesStockTableTableManager(_db, _db.salesStock);
  $$OfflineTasksTableTableManager get offlineTasks =>
      $$OfflineTasksTableTableManager(_db, _db.offlineTasks);
  $$StoresTableTableManager get stores =>
      $$StoresTableTableManager(_db, _db.stores);
  $$VisitPlansTableTableManager get visitPlans =>
      $$VisitPlansTableTableManager(_db, _db.visitPlans);
  $$SalesDailyJourneysTableTableManager get salesDailyJourneys =>
      $$SalesDailyJourneysTableTableManager(_db, _db.salesDailyJourneys);
}
