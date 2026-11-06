import 'package:equatable/equatable.dart';
import 'package:hris_app/features/kpi/data/models/kpi_model.dart';
import 'package:hris_app/features/kpi/domain/repositories/kpi_repository.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

// Events
abstract class KPIEvent extends Equatable {
  const KPIEvent();
  @override
  List<Object?> get props => [];
}

class FetchKPIRequested extends KPIEvent {
  final String period;
  const FetchKPIRequested(this.period);
  @override
  List<Object?> get props => [period];
}

class FetchKPIHistoryRequested extends KPIEvent {}

// States
abstract class KPIState extends Equatable {
  const KPIState();
  @override
  List<Object?> get props => [];
}

class KPIInitial extends KPIState {}

class KPILoading extends KPIState {}

class KPILoaded extends KPIState {
  final KPIModel currentKPI;
  final List<KPIModel> history;
  const KPILoaded({required this.currentKPI, required this.history});
  @override
  List<Object?> get props => [currentKPI, history];
}

class KPIError extends KPIState {
  final String message;
  const KPIError(this.message);
  @override
  List<Object?> get props => [message];
}

// BLoC
class KPIBloc extends Bloc<KPIEvent, KPIState> {
  final KPIRepository repository;

  KPIBloc({required this.repository}) : super(KPIInitial()) {
    on<FetchKPIRequested>((event, emit) async {
      emit(KPILoading());
      try {
        final current = await repository.getKPIByPeriod(event.period);
        final history = await repository.getKPIHistory();
        emit(KPILoaded(currentKPI: current, history: history));
      } catch (e) {
        emit(KPIError(e.toString()));
      }
    });

    on<FetchKPIHistoryRequested>((event, emit) async {
      emit(KPILoading());
      try {
        final history = await repository.getKPIHistory();
        if (history.isNotEmpty) {
          emit(KPILoaded(currentKPI: history.first, history: history));
        } else {
          emit(const KPIError('Tidak ada data KPI.'));
        }
      } catch (e) {
        emit(KPIError(e.toString()));
      }
    });
  }
}
