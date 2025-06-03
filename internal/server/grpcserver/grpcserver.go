package grpcserver

import (
	"context"
	"errors"

	"github.com/lenarlenar/go-my-metrics-service/internal/interfaces"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/lenarlenar/go-my-metrics-service/proto/metrics"
)

var (
	ErrNotFound = errors.New("метрика не найдена")
	ErrBadType  = errors.New("неизвестный тип метрики")
)

// Server реализует интерфейс pb.MetricsServiceServer
type Server struct {
	pb.UnimplementedMetricsServiceServer
	storage interfaces.Storage
}

// NewGRPCServer возвращает gRPC-сервер
func NewServer(storage interfaces.Storage) *Server {
	return &Server{storage: storage}
}

// Ping обрабатывает Ping-запрос
func (s *Server) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := s.storage.Ping(); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// GetMetric обрабатывает GET /value/:type/:name
func (s *Server) GetMetric(ctx context.Context, req *pb.MetricRequestByPath) (*pb.MetricResponse, error) {
	metric, ok := s.storage.GetMetrics()[req.Name]
	if !ok {
		return nil, ErrNotFound
	}
	return wrapMetric(&metric)
}

// GetMetricByJSON — POST /value/
func (s *Server) GetMetricByJSON(ctx context.Context, req *pb.Metric) (*pb.MetricResponse, error) {
	metric, ok := s.storage.GetMetrics()[req.Id]
	if !ok {
		return nil, ErrNotFound
	}
	return wrapMetric(&metric)
}

// UpdateMetric — POST /update/
func (s *Server) UpdateMetric(ctx context.Context, req *pb.Metric) (*pb.MetricResponse, error) {
	switch req.Type {
	case "gauge":
		s.storage.SetGauge(req.Id, req.Value)
	case "counter":
		s.storage.AddCounter(req.Id, req.Delta)
	default:
		return nil, ErrBadType
	}

	m, ok := s.storage.GetMetrics()[req.Id]
	if !ok {
		return nil, ErrNotFound
	}
	return wrapMetric(&m)
}

// UpdateMetricsBatch — POST /updates/
func (s *Server) UpdateMetricsBatch(ctx context.Context, req *pb.MetricsBatch) (*emptypb.Empty, error) {
	for _, m := range req.Metrics {
		switch m.Type {
		case "gauge":
			s.storage.SetGauge(m.Id, m.Value)
		case "counter":
			s.storage.AddCounter(m.Id, m.Delta)
		default:
			continue
		}
	}
	return &emptypb.Empty{}, nil
}

// Вспомогательная обёртка
func wrapMetric(m *model.Metrics) (*pb.MetricResponse, error) {
	if m == nil {
		return nil, ErrNotFound
	}
	resp := &pb.Metric{
		Id:    m.ID,
		Type:  m.MType,
		Value: 0,
		Delta: 0,
	}
	if m.MType == "gauge" && m.Value != nil {
		resp.Value = *m.Value
	}
	if m.MType == "counter" && m.Delta != nil {
		resp.Delta = *m.Delta
	}
	return &pb.MetricResponse{Metric: resp}, nil
}
