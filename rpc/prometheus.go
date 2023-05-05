package rpc

import (
	"context"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/exp/slices"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/protocol/swap"
)

var namespace = "swapdaemon"

// Metrics represents our prometheus metrics
type Metrics struct {
	peersCount            prometheus.GaugeFunc
	ongoingSwapsCount     prometheus.GaugeFunc
	pastSwapsSuccessCount prometheus.GaugeFunc
	pastSwapsRefundCount  prometheus.GaugeFunc
	pastSwapsAbortCount   prometheus.GaugeFunc
	offersCount           prometheus.GaugeFunc
	moneroBalance         prometheus.GaugeFunc
	ethereumBalance       prometheus.GaugeFunc
	averageSwapDuration   prometheus.GaugeFunc
}

func pastSwapsMetric(swapManager SwapManager, status swap.Status, statusLabel string) prometheus.GaugeFunc {
	return promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   namespace,
			Name:        "past_swaps_count",
			Help:        "The number of past swaps by status",
			ConstLabels: prometheus.Labels{"status": statusLabel},
		},
		func() float64 {
			pastIDs, err := swapManager.GetPastIDs()
			if err != nil {
				return -1
			}

			var count int

			for _, pastID := range pastIDs {
				pastSwap, err := swapManager.GetPastSwap(pastID)
				if err != nil {
					continue
				}

				if pastSwap.Status == status {
					count++
				}
			}

			return float64(count)
		},
	)
}

// SetupMetrics creates prometheus metrics and returns a new Metrics
func SetupMetrics(ctx context.Context, net Net, swapManager SwapManager, pb ProtocolBackend, maker XMRMaker) Metrics {
	return Metrics{
		peersCount: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "peers_count",
				Help:      "The number of connected peers",
			},
			func() float64 {
				peerIDs := []peer.ID{}

				for _, addr := range net.ConnectedPeers() {
					addrInfo, err := peer.AddrInfoFromString(addr)
					if err != nil {
						continue
					}

					if slices.Index(peerIDs, addrInfo.ID) == -1 {
						peerIDs = append(peerIDs, addrInfo.ID)
					}
				}

				return float64(len(peerIDs))
			},
		),

		ongoingSwapsCount: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "ongoing_swaps_count",
				Help:      "The number of ongoing swaps",
			},
			func() float64 {
				swaps, err := swapManager.GetOngoingSwaps()
				if err != nil {
					return float64(-1)
				}
				return float64(len(swaps))
			},
		),

		pastSwapsSuccessCount: pastSwapsMetric(swapManager, types.CompletedSuccess, "success"),
		pastSwapsRefundCount:  pastSwapsMetric(swapManager, types.CompletedRefund, "refund"),
		pastSwapsAbortCount:   pastSwapsMetric(swapManager, types.CompletedAbort, "abort"),

		offersCount: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "owned_offers_count",
				Help:      "The number of offers",
			},
			func() float64 {
				offers := maker.GetOffers()
				return float64(len(offers))
			},
		),

		moneroBalance: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "Balance",
				ConstLabels: prometheus.Labels{"coin": "xmr"},
			},
			func() float64 {
				_, balanceResp, err := maker.GetMoneroBalance()
				if err != nil {
					return float64(-1)
				}
				return float64(balanceResp.Balance)
			},
		),

		ethereumBalance: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace:   namespace,
				Name:        "balance",
				Help:        "Balance",
				ConstLabels: prometheus.Labels{"coin": "eth"},
			},
			func() float64 {
				balance, err := pb.ETHClient().Balance(ctx)
				if err != nil {
					return float64(-1)
				}
				fBalance, err := balance.Decimal().Float64()
				if err != nil {
					return float64(-1)
				}
				return fBalance
			},
		),

		averageSwapDuration: promauto.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "avg_swap_duration_seconds",
				Help:      "The average swap duration",
			},
			func() float64 {
				pastIDs, err := swapManager.GetPastIDs()
				if err != nil {
					return float64(-1)
				}

				var (
					sum   float64
					count int
				)

				for _, pastID := range pastIDs {
					pastSwap, err := swapManager.GetPastSwap(pastID)
					if err != nil {
						continue
					}

					sum += float64(pastSwap.EndTime.Unix() - pastSwap.StartTime.Unix())
					count++
				}

				return sum / float64(count)
			},
		),
	}
}
