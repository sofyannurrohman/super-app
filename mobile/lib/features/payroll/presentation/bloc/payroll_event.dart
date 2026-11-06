import 'package:equatable/equatable.dart';

abstract class PayrollEvent extends Equatable {
  const PayrollEvent();

  @override
  List<Object?> get props => [];
}

class FetchMyPayslipsRequested extends PayrollEvent {}
