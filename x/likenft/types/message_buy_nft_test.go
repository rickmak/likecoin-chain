package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/likecoin/likecoin-chain/v3/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgBuyNFT_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgBuyNFT
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgBuyNFT{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgBuyNFT{
				Creator: sample.AccAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
