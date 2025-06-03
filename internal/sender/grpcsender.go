package sender

import (
	"context"
	"time"

	"github.com/lenarlenar/go-my-metrics-service/internal/agent/flags"
	"github.com/lenarlenar/go-my-metrics-service/internal/log"
	"github.com/lenarlenar/go-my-metrics-service/internal/model"
	pb "github.com/lenarlenar/go-my-metrics-service/proto/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SendGRPC(flags flags.Flags, metrics map[string]model.Metrics) {
	conn, err := grpc.NewClient(flags.GRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.I().Fatalf("failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := pb.NewMetricsServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, m := range metrics {
		msg := &pb.Metric{
			Id:   m.ID,
			Type: m.MType,
		}
		if m.MType == "gauge" && m.Value != nil {
			msg.Value = *m.Value
		}
		if m.MType == "counter" && m.Delta != nil {
			msg.Delta = *m.Delta
		}
		grpcMetrics = append(grpcMetrics, msg)
	}

	_, err = client.UpdateMetricsBatch(ctx, &pb.MetricsBatch{Metrics: grpcMetrics})
	if err != nil {
		log.I().Warnf("failed to send gRPC batch: %v", err)
		return
	}

	log.I().Infof("successfully sent %d metrics via gRPC", len(grpcMetrics))
}
