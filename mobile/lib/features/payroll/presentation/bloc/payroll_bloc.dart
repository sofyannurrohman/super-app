import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/usecases/get_payslip_history_usecase.dart';
import 'payroll_event.dart';
import 'payroll_state.dart';

class PayrollBloc extends Bloc<PayrollEvent, PayrollState> {
  final GetMyPayslipHistoryUseCase getMyPayslipHistoryUseCase;

  PayrollBloc({
    required this.getMyPayslipHistoryUseCase,
  }) : super(PayrollInitial()) {
    on<FetchMyPayslipsRequested>(_onFetchMyPayslipsRequested);
  }

  Future<void> _onFetchMyPayslipsRequested(
    FetchMyPayslipsRequested event,
    Emitter<PayrollState> emit,
  ) async {
    emit(PayrollLoading());
    final result = await getMyPayslipHistoryUseCase();
    result.fold(
      (failure) => emit(PayrollFailure(failure.message)),
      (payslips) => emit(PayrollLoaded(payslips)),
    );
  }
}
