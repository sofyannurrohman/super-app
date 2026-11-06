import 'package:flutter_test/flutter_test.dart';
import 'package:hris_app/core/network/network_info.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('NetworkInfo', () {
    test('singleton returns same instance', () {
      final instance1 = NetworkInfo();
      final instance2 = NetworkInfo();
      expect(instance1, same(instance2));
    });

    test('getNetworkType returns valid network string', () async {
      final networkInfo = NetworkInfo();
      final type = await networkInfo.getNetworkType();
      expect(type, isNotNull);
      expect(
        ['wifi', 'cellular', 'ethernet', 'none', 'other', 'unknown', 'web'],
        contains(type),
      );
    });

    test('onStatusChange stream is not null', () {
      final networkInfo = NetworkInfo();
      expect(networkInfo.onStatusChange, isNotNull);
    });
  });
}
