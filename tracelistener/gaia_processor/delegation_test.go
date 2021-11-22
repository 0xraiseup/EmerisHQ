package gaia_processor

import (
	"testing"

	sdk_types "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	models "github.com/allinbits/demeris-backend-models/tracelistener"
	"github.com/allinbits/tracelistener/tracelistener"
	"github.com/allinbits/tracelistener/tracelistener/config"
)

func TestDelegationProcess(t *testing.T) {
	d := delegationsProcessor{}

	DataProcessor, err := New(zap.NewNop().Sugar(), &config.Config{})
	require.NoError(t, err)

	gp := DataProcessor.(*Processor)
	require.NotNil(t, gp)
	p.cdc = gp.cdc

	tests := []struct {
		name        string
		delegation  types.Delegation
		newMessage  tracelistener.TraceOperation
		wantErr     bool
		expectedLen int
	}{
		{
			"Delete operation of delegation - no error",
			types.Delegation{},
			tracelistener.TraceOperation{
				Operation: string(tracelistener.DeleteOp),
				Key:       []byte("QXRkbFY4cUQ2bzZKMnNoc2o5YWNwSSs5T3BkL2U1dVRxWklpN05LNWkzeTk="),
			},
			false,
			1,
		},
		{
			"Write new delegation - no error",
			types.Delegation{
				DelegatorAddress: "cosmos1xrnner9s783446yz3hhshpr5fpz6wzcwkvwv5j",
				ValidatorAddress: "cosmosvaloper19xawgvgn887e9gef5vkzkemwh33mtgwa6haa7s",
				Shares:           sdk_types.NewDec(100),
			},
			tracelistener.TraceOperation{
				Operation:   string(tracelistener.WriteOp),
				Key:         []byte("AtdlV8qD6o6J2shsj9acpI+9Opd/e5uTqZIi7NK5i3y9"),
				BlockHeight: 1,
				TxHash:      "A5CF62609D62ADDE56816681B6191F5F0252D2800FC2C312EB91D962AB7A97CB",
			},
			false,
			1,
		},
		{
			"Invalid addresses - error",
			types.Delegation{
				DelegatorAddress: "",
				ValidatorAddress: "",
				Shares:           sdk_types.NewDec(100),
			},
			tracelistener.TraceOperation{
				Operation:   string(tracelistener.WriteOp),
				Key:         []byte("AtdlV8qD6o6J2shsj9acpI+9Opd/e5uTqZIi7NK5i3y9"),
				BlockHeight: 1,
				TxHash:      "A5CF62609D62ADDE56816681B6191F5F0252D2800FC2C312EB91D962AB7A97CB",
			},
			true,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d.insertHeightCache = map[delegationCacheEntry]models.DelegationRow{}
			d.deleteHeightCache = map[delegationCacheEntry]models.DelegationRow{}
			d.l = zap.NewNop().Sugar()

			delValue, err := p.cdc.MarshalBinaryBare(&tt.delegation)
			require.NoError(t, err)

			tt.newMessage.Value = delValue

			err = d.Process(tt.newMessage)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.newMessage.Operation == tracelistener.DeleteOp.String() {
				require.Len(t, d.deleteHeightCache, tt.expectedLen)

				for k, _ := range d.deleteHeightCache {
					row := d.deleteHeightCache[delegationCacheEntry{delegator: k.delegator, validator: k.validator}]
					require.NotNil(t, row)

					return
				}
			} else {
				require.Len(t, d.insertHeightCache, tt.expectedLen)

				for k, _ := range d.insertHeightCache {
					row := d.insertHeightCache[delegationCacheEntry{delegator: k.delegator, validator: k.validator}]
					require.NotNil(t, row)

					amount := row.Amount
					require.Equal(t, tt.delegation.Shares.String(), amount)

					return
				}
			}
		})
	}
}
