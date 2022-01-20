package processor

import (
	"bytes"

	models "github.com/allinbits/demeris-backend-models/tracelistener"
	"github.com/allinbits/tracelistener/tracelistener"
	"github.com/allinbits/tracelistener/tracelistener/processor/datamarshaler"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"go.uber.org/zap"
)

type authCacheEntry struct {
	address   string
	accNumber uint64
}

type authProcessor struct {
	l           *zap.SugaredLogger
	heightCache map[authCacheEntry]models.AuthRow
}

func (*authProcessor) TableSchema() string {
	return createAuthTable
}

func (b *authProcessor) ModuleName() string {
	return "auth"
}

func (b *authProcessor) FlushCache() []tracelistener.WritebackOp {
	if len(b.heightCache) == 0 {
		return nil
	}

	l := make([]models.DatabaseEntrier, 0, len(b.heightCache))

	for _, v := range b.heightCache {
		l = append(l, v)
	}

	b.heightCache = map[authCacheEntry]models.AuthRow{}

	return []tracelistener.WritebackOp{
		{
			DatabaseExec: insertAuth,
			Data:         l,
		},
	}
}

func (b *authProcessor) OwnsKey(key []byte) bool {
	return bytes.HasPrefix(key, types.AddressStoreKeyPrefix)
}

func (b *authProcessor) Process(data tracelistener.TraceOperation) error {
	res, err := datamarshaler.NewDataMarshaler(b.l).Auth(data)
	if err != nil {
		return err
	}

	b.heightCache[authCacheEntry{
		address:   res.Address,
		accNumber: res.AccountNumber,
	}] = res

	return nil
}
