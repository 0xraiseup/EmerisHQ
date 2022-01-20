package processor

import (
	"bytes"

	models "github.com/allinbits/demeris-backend-models/tracelistener"

	host "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	"go.uber.org/zap"

	"github.com/allinbits/tracelistener/tracelistener"
	"github.com/allinbits/tracelistener/tracelistener/processor/datamarshaler"
)

type clientCacheEntry struct {
	chainID  string
	clientID string
}

type ibcClientsProcessor struct {
	l            *zap.SugaredLogger
	clientsCache map[clientCacheEntry]models.IBCClientStateRow
}

func (*ibcClientsProcessor) TableSchema() string {
	return createClientsTable
}

func (b *ibcClientsProcessor) ModuleName() string {
	return "ibc_clients"
}

func (b *ibcClientsProcessor) FlushCache() []tracelistener.WritebackOp {
	if len(b.clientsCache) == 0 {
		return nil
	}

	l := make([]models.DatabaseEntrier, 0, len(b.clientsCache))

	for _, c := range b.clientsCache {
		l = append(l, c)
	}

	b.clientsCache = map[clientCacheEntry]models.IBCClientStateRow{}

	return []tracelistener.WritebackOp{
		{
			DatabaseExec: insertClient,
			Data:         l,
		},
	}
}

func (b *ibcClientsProcessor) OwnsKey(key []byte) bool {
	return bytes.Contains(key, []byte(host.KeyClientState))
}

func (b *ibcClientsProcessor) Process(data tracelistener.TraceOperation) error {
	res, err := datamarshaler.NewDataMarshaler(b.l).IBCClients(data)
	if err != nil {
		return err
	}

	b.clientsCache[clientCacheEntry{
		chainID:  res.ChainID,
		clientID: res.ClientID,
	}] = res

	return nil
}
