package processor

import (
	"bytes"
	"sync"

	models "github.com/emerishq/demeris-backend-models/tracelistener"

	"go.uber.org/zap"

	"github.com/emerishq/tracelistener/tracelistener"
	"github.com/emerishq/tracelistener/tracelistener/processor/datamarshaler"
)

type clientCacheEntry struct {
	chainID  string
	clientID string
}

type ibcClientsProcessor struct {
	l            *zap.SugaredLogger
	clientsCache map[clientCacheEntry]models.IBCClientStateRow
	m            sync.Mutex
}

func (*ibcClientsProcessor) Migrations() []string {
	return []string{createClientsTable}
}

func (b *ibcClientsProcessor) ModuleName() string {
	return "ibc_clients"
}

func (b *ibcClientsProcessor) SDKModuleName() tracelistener.SDKModuleName {
	return tracelistener.IBC
}

func (b *ibcClientsProcessor) UpsertStatement() string {
	return insertClient
}

func (b *ibcClientsProcessor) InsertStatement() string {
	return upsertClient
}

func (b *ibcClientsProcessor) DeleteStatement() string {
	panic("ibc clients processor never deletes")
}

func (b *ibcClientsProcessor) FlushCache() []tracelistener.WritebackOp {
	b.m.Lock()
	defer b.m.Unlock()

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
			Type: tracelistener.Write,
			Data: l,
		},
	}
}

func (b *ibcClientsProcessor) OwnsKey(key []byte) bool {
	return bytes.Contains(key, []byte(datamarshaler.IBCClientsKey))
}

func (b *ibcClientsProcessor) Process(data tracelistener.TraceOperation) error {
	b.m.Lock()
	defer b.m.Unlock()

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
