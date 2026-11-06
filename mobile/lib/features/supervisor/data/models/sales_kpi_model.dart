class SalesKpiModel {
  final String id;
  final String employeeId;
  final String employeeName;
  final double targetOmzet;
  final double achievedOmzet;
  final int targetNewStores;
  final int achievedNewStores;
  final int targetDailyVisits;
  final int totalVisits;
  final int periodMonth;
  final int periodYear;
  final String jobTitle;
  final String branchName;

  SalesKpiModel({
    required this.id,
    required this.employeeId,
    required this.employeeName,
    required this.targetOmzet,
    required this.achievedOmzet,
    required this.targetNewStores,
    required this.achievedNewStores,
    required this.targetDailyVisits,
    required this.totalVisits,
    required this.periodMonth,
    required this.periodYear,
    this.jobTitle = '',
    this.branchName = '',
  });

  factory SalesKpiModel.fromJson(Map<String, dynamic> json) {
    return SalesKpiModel(
      id: json['ID'] ?? json['id'] ?? '',
      employeeId: json['EmployeeID'] ?? json['employee_id'] ?? '',
      employeeName: (json['Employee'] != null)
          ? '${json['Employee']['first_name'] ?? ''} ${json['Employee']['last_name'] ?? ''}'
              .trim()
          : (json['employee'] != null)
              ? '${json['employee']['first_name'] ?? ''} ${json['employee']['last_name'] ?? ''}'
                  .trim()
              : '',
      targetOmzet: _parseDouble(json['TargetOmzet'] ?? json['target_omzet']),
      achievedOmzet:
          _parseDouble(json['AchievedOmzet'] ?? json['achieved_omzet']),
      targetNewStores:
          _parseInt(json['TargetNewStores'] ?? json['target_new_stores']),
      achievedNewStores:
          _parseInt(json['AchievedNewStores'] ?? json['achieved_new_stores']),
      targetDailyVisits:
          _parseInt(json['target_daily_visits'] ?? json['TargetDailyVisits']),
      totalVisits: _parseInt(json['TotalVisits'] ?? json['total_visits']),
      periodMonth: _parseInt(json['period_month'] ?? json['PeriodMonth'] ?? 1),
      periodYear: _parseInt(json['period_year'] ?? json['PeriodYear'] ?? 2024),
      jobTitle: _parseJobTitle(json),
      branchName: _parseBranchName(json),
    );
  }

  static String _parseBranchName(Map<String, dynamic> json) {
    if (json['Branch'] != null && json['Branch']['name'] != null)
      return json['Branch']['name'];
    if (json['branch'] != null && json['branch']['name'] != null)
      return json['branch']['name'];
    if (json['Employee'] != null &&
        json['Employee']['Branch'] != null &&
        json['Employee']['Branch']['name'] != null)
      return json['Employee']['Branch']['name'];
    if (json['employee'] != null &&
        json['employee']['branch'] != null &&
        json['employee']['branch']['name'] != null)
      return json['employee']['branch']['name'];
    if (json['branch_name'] != null) return json['branch_name'];
    return 'Cabang Utama'; // fallback
  }

  static String _parseJobTitle(Map<String, dynamic> json) {
    bool checkMt(Map<String, dynamic>? emp) {
      if (emp == null) return false;
      return emp['is_mt'] == 1 ||
          emp['is_mt'] == true ||
          emp['IsMT'] == 1 ||
          emp['IsMT'] == true;
    }

    if (json['is_mt'] == 1 ||
        json['is_mt'] == true ||
        json['IsMT'] == 1 ||
        json['IsMT'] == true) {
      return 'MT';
    }
    if (checkMt(json['Employee'])) return 'MT';
    if (checkMt(json['employee'])) return 'MT';

    String extractString(dynamic value) {
      if (value == null) return '';
      if (value is String) return value;
      if (value is Map)
        return (value['title'] ?? value['name'] ?? value['job_type'] ?? '')
            .toString();
      return value.toString();
    }

    if (json['Employee'] != null) {
      if (json['Employee']['job_position'] != null)
        return extractString(json['Employee']['job_position']);
      if (json['Employee']['job_type'] != null)
        return extractString(json['Employee']['job_type']);
    }

    if (json['employee'] != null) {
      if (json['employee']['job_position'] != null)
        return extractString(json['employee']['job_position']);
      if (json['employee']['job_type'] != null)
        return extractString(json['employee']['job_type']);
    }

    if (json['job_position'] != null)
      return extractString(json['job_position']);
    if (json['job_type'] != null) return extractString(json['job_type']);

    return '';
  }

  static double _parseDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }

  static int _parseInt(dynamic value) {
    if (value == null) return 0;
    if (value is num) return value.toInt();
    if (value is String) return int.tryParse(value) ?? 0;
    return 0;
  }
}
