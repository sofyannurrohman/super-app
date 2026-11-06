import 'package:hris_app/features/kpi/data/models/kpi_model.dart';

abstract class KPIRepository {
  Future<List<KPIModel>> getKPIHistory();
  Future<KPIModel> getKPIByPeriod(String yearMonth);
}
