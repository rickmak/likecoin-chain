package keeper_test

import (
	"testing"
	"time"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/golang/mock/gomock"
	"github.com/likecoin/likechain/backport/cosmos-sdk/v0.46.0-alpha2/x/nft"
	"github.com/likecoin/likechain/testutil/keeper"
	"github.com/likecoin/likechain/x/likenft/testutil"
	"github.com/likecoin/likechain/x/likenft/types"
	"github.com/stretchr/testify/require"
)

func TestUpdateMintableNFTNormal(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	iscnKeeper := testutil.NewMockIscnKeeper(ctrl)
	nftKeeper := testutil.NewMockNftKeeper(ctrl)
	msgServer, goCtx, keeper := setupMsgServer(t, keeper.LikenftDependedKeepers{
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		IscnKeeper:    iscnKeeper,
		NftKeeper:     nftKeeper,
	})
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Test Input
	ownerAddressBytes := []byte{0, 1, 0, 1, 0, 1, 0, 1}
	ownerAddress, _ := sdk.Bech32ifyAddressBytes("like", ownerAddressBytes)
	classId := "likenft1aabbccddeeff"
	mintableId := "mintable1"

	// Mock calls
	classData := types.ClassData{
		Metadata: types.JsonInput(`{"1234": "5678"}`),
		Parent: types.ClassParent{
			Type:    types.ClassParentType_ACCOUNT,
			Account: ownerAddress,
		},
		Config: types.ClassConfig{
			Burnable:  false,
			MaxSupply: uint64(5),
			BlindBoxConfig: &types.BlindBoxConfig{
				MintPeriods: []types.MintPeriod{
					{
						StartTime:        time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
						AllowedAddresses: []string{},
						MintPrice:        1000,
					},
				},
				RevealTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		ToBeRevealed: true,
	}
	classDataInAny, _ := cdctypes.NewAnyWithValue(&classData)
	nftKeeper.EXPECT().GetClass(gomock.Any(), classId).Return(nft.Class{
		Id:          classId,
		Name:        "Class Name",
		Symbol:      "ABC",
		Description: "Testing Class 123",
		Uri:         "ipfs://abcdef",
		UriHash:     "abcdef",
		Data:        classDataInAny,
	}, true).MinTimes(1)
	keeper.SetClassesByAccount(ctx, types.ClassesByAccount{
		Account:  ownerAddress,
		ClassIds: []string{classId},
	})

	nftKeeper.EXPECT().GetTotalSupply(gomock.Any(), classId).Return(uint64(0))

	// once at seeding
	nftKeeper.EXPECT().UpdateClass(gomock.Any(), gomock.Any()).Return(nil)
	keeper.SetMintableNFT(ctx, types.MintableNFT{
		ClassId: classId,
		Id:      mintableId,
	})

	// Run
	nftInput := types.NFTInput{
		Uri:      "ipfs://123456",
		UriHash:  "123456",
		Metadata: types.JsonInput(`{"aaaa": "bbbb"}`),
	}
	res, err := msgServer.UpdateMintableNFT(goCtx, &types.MsgUpdateMintableNFT{
		Creator: ownerAddress,
		ClassId: classId,
		Id:      mintableId,
		Input:   nftInput,
	})

	// Check output
	require.NoError(t, err)
	require.NotNil(t, res)

	updated, found := keeper.GetMintableNFT(ctx, classId, mintableId)
	require.True(t, found)
	require.Equal(t, types.MintableNFT{
		ClassId: classId,
		Id:      mintableId,
		Input:   nftInput,
	}, updated)

	ctrl.Finish()
}

func TestUpdateMintableNFTClassNotFound(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	iscnKeeper := testutil.NewMockIscnKeeper(ctrl)
	nftKeeper := testutil.NewMockNftKeeper(ctrl)
	msgServer, goCtx, keeper := setupMsgServer(t, keeper.LikenftDependedKeepers{
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		IscnKeeper:    iscnKeeper,
		NftKeeper:     nftKeeper,
	})
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Test Input
	ownerAddressBytes := []byte{0, 1, 0, 1, 0, 1, 0, 1}
	ownerAddress, _ := sdk.Bech32ifyAddressBytes("like", ownerAddressBytes)
	classId := "likenft1aabbccddeeff"
	mintableId := "mintable1"

	// Mock calls
	nftKeeper.EXPECT().GetClass(gomock.Any(), classId).Return(nft.Class{}, false).MinTimes(1)
	keeper.SetClassesByAccount(ctx, types.ClassesByAccount{
		Account:  ownerAddress,
		ClassIds: []string{classId},
	})

	// Run
	nftInput := types.NFTInput{
		Uri:      "ipfs://123456",
		UriHash:  "123456",
		Metadata: types.JsonInput(`{"aaaa": "bbbb"}`),
	}
	res, err := msgServer.UpdateMintableNFT(goCtx, &types.MsgUpdateMintableNFT{
		Creator: ownerAddress,
		ClassId: classId,
		Id:      mintableId,
		Input:   nftInput,
	})

	// Check output
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrNftClassNotFound.Error())
	require.Nil(t, res)

	ctrl.Finish()
}

func TestUpdateMintableNFTBadRelation(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	iscnKeeper := testutil.NewMockIscnKeeper(ctrl)
	nftKeeper := testutil.NewMockNftKeeper(ctrl)
	msgServer, goCtx, _ := setupMsgServer(t, keeper.LikenftDependedKeepers{
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		IscnKeeper:    iscnKeeper,
		NftKeeper:     nftKeeper,
	})

	// Test Input
	ownerAddressBytes := []byte{0, 1, 0, 1, 0, 1, 0, 1}
	ownerAddress, _ := sdk.Bech32ifyAddressBytes("like", ownerAddressBytes)
	classId := "likenft1aabbccddeeff"
	mintableId := "mintable1"

	// Mock calls
	classData := types.ClassData{
		Metadata: types.JsonInput(`{"1234": "5678"}`),
		Parent: types.ClassParent{
			Type:    types.ClassParentType_ACCOUNT,
			Account: ownerAddress,
		},
		Config: types.ClassConfig{
			Burnable:  false,
			MaxSupply: uint64(5),
			BlindBoxConfig: &types.BlindBoxConfig{
				MintPeriods: []types.MintPeriod{
					{
						StartTime:        time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
						AllowedAddresses: []string{},
						MintPrice:        1000,
					},
				},
				RevealTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		ToBeRevealed: true,
	}
	classDataInAny, _ := cdctypes.NewAnyWithValue(&classData)
	nftKeeper.EXPECT().GetClass(gomock.Any(), classId).Return(nft.Class{
		Id:          classId,
		Name:        "Class Name",
		Symbol:      "ABC",
		Description: "Testing Class 123",
		Uri:         "ipfs://abcdef",
		UriHash:     "abcdef",
		Data:        classDataInAny,
	}, true).MinTimes(1)

	// Run
	nftInput := types.NFTInput{
		Uri:      "ipfs://123456",
		UriHash:  "123456",
		Metadata: types.JsonInput(`{"aaaa": "bbbb"}`),
	}
	res, err := msgServer.UpdateMintableNFT(goCtx, &types.MsgUpdateMintableNFT{
		Creator: ownerAddress,
		ClassId: classId,
		Id:      mintableId,
		Input:   nftInput,
	})

	// Check output
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrNftClassNotRelatedToAnyAccount.Error())
	require.Nil(t, res)

	ctrl.Finish()
}

func TestUpdateMintableNFTAlreadyMinted(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	iscnKeeper := testutil.NewMockIscnKeeper(ctrl)
	nftKeeper := testutil.NewMockNftKeeper(ctrl)
	msgServer, goCtx, keeper := setupMsgServer(t, keeper.LikenftDependedKeepers{
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		IscnKeeper:    iscnKeeper,
		NftKeeper:     nftKeeper,
	})
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Test Input
	ownerAddressBytes := []byte{0, 1, 0, 1, 0, 1, 0, 1}
	ownerAddress, _ := sdk.Bech32ifyAddressBytes("like", ownerAddressBytes)
	classId := "likenft1aabbccddeeff"
	mintableId := "mintable1"

	// Mock calls
	classData := types.ClassData{
		Metadata: types.JsonInput(`{"1234": "5678"}`),
		Parent: types.ClassParent{
			Type:    types.ClassParentType_ACCOUNT,
			Account: ownerAddress,
		},
		Config: types.ClassConfig{
			Burnable:  false,
			MaxSupply: uint64(5),
			BlindBoxConfig: &types.BlindBoxConfig{
				MintPeriods: []types.MintPeriod{
					{
						StartTime:        time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
						AllowedAddresses: []string{},
						MintPrice:        1000,
					},
				},
				RevealTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		ToBeRevealed: true,
	}
	classDataInAny, _ := cdctypes.NewAnyWithValue(&classData)
	nftKeeper.EXPECT().GetClass(gomock.Any(), classId).Return(nft.Class{
		Id:          classId,
		Name:        "Class Name",
		Symbol:      "ABC",
		Description: "Testing Class 123",
		Uri:         "ipfs://abcdef",
		UriHash:     "abcdef",
		Data:        classDataInAny,
	}, true).MinTimes(1)
	keeper.SetClassesByAccount(ctx, types.ClassesByAccount{
		Account:  ownerAddress,
		ClassIds: []string{classId},
	})

	nftKeeper.EXPECT().GetTotalSupply(gomock.Any(), classId).Return(uint64(1))

	// Run
	nftInput := types.NFTInput{
		Uri:      "ipfs://123456",
		UriHash:  "123456",
		Metadata: types.JsonInput(`{"aaaa": "bbbb"}`),
	}
	res, err := msgServer.UpdateMintableNFT(goCtx, &types.MsgUpdateMintableNFT{
		Creator: ownerAddress,
		ClassId: classId,
		Id:      mintableId,
		Input:   nftInput,
	})

	// Check output
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrCannotUpdateClassWithMintedTokens.Error())
	require.Nil(t, res)

	ctrl.Finish()
}

func TestUpdateMintableNFTNotOwner(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	iscnKeeper := testutil.NewMockIscnKeeper(ctrl)
	nftKeeper := testutil.NewMockNftKeeper(ctrl)
	msgServer, goCtx, keeper := setupMsgServer(t, keeper.LikenftDependedKeepers{
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		IscnKeeper:    iscnKeeper,
		NftKeeper:     nftKeeper,
	})
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Test Input
	ownerAddressBytes := []byte{0, 1, 0, 1, 0, 1, 0, 1}
	ownerAddress, _ := sdk.Bech32ifyAddressBytes("like", ownerAddressBytes)
	classId := "likenft1aabbccddeeff"
	mintableId := "mintable1"

	// Mock calls
	classData := types.ClassData{
		Metadata: types.JsonInput(`{"1234": "5678"}`),
		Parent: types.ClassParent{
			Type:    types.ClassParentType_ACCOUNT,
			Account: ownerAddress,
		},
		Config: types.ClassConfig{
			Burnable:  false,
			MaxSupply: uint64(5),
			BlindBoxConfig: &types.BlindBoxConfig{
				MintPeriods: []types.MintPeriod{
					{
						StartTime:        time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
						AllowedAddresses: []string{},
						MintPrice:        1000,
					},
				},
				RevealTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		ToBeRevealed: true,
	}
	classDataInAny, _ := cdctypes.NewAnyWithValue(&classData)
	nftKeeper.EXPECT().GetClass(gomock.Any(), classId).Return(nft.Class{
		Id:          classId,
		Name:        "Class Name",
		Symbol:      "ABC",
		Description: "Testing Class 123",
		Uri:         "ipfs://abcdef",
		UriHash:     "abcdef",
		Data:        classDataInAny,
	}, true).MinTimes(1)
	keeper.SetClassesByAccount(ctx, types.ClassesByAccount{
		Account:  ownerAddress,
		ClassIds: []string{classId},
	})

	nftKeeper.EXPECT().GetTotalSupply(gomock.Any(), classId).Return(uint64(0))

	// Run
	notOwnerAddressBytes := []byte{1, 0, 1, 0, 1, 0, 1, 0}
	notOwnerAddress, _ := sdk.Bech32ifyAddressBytes("like", notOwnerAddressBytes)
	nftInput := types.NFTInput{
		Uri:      "ipfs://123456",
		UriHash:  "123456",
		Metadata: types.JsonInput(`{"aaaa": "bbbb"}`),
	}
	res, err := msgServer.UpdateMintableNFT(goCtx, &types.MsgUpdateMintableNFT{
		Creator: notOwnerAddress,
		ClassId: classId,
		Id:      mintableId,
		Input:   nftInput,
	})

	// Check output
	require.Error(t, err)
	require.Contains(t, err.Error(), sdkerrors.ErrUnauthorized.Error())
	require.Nil(t, res)

	ctrl.Finish()
}

func TestUpdateMintableNFTDoNotExist(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	iscnKeeper := testutil.NewMockIscnKeeper(ctrl)
	nftKeeper := testutil.NewMockNftKeeper(ctrl)
	msgServer, goCtx, keeper := setupMsgServer(t, keeper.LikenftDependedKeepers{
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
		IscnKeeper:    iscnKeeper,
		NftKeeper:     nftKeeper,
	})
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Test Input
	ownerAddressBytes := []byte{0, 1, 0, 1, 0, 1, 0, 1}
	ownerAddress, _ := sdk.Bech32ifyAddressBytes("like", ownerAddressBytes)
	classId := "likenft1aabbccddeeff"
	mintableId := "mintable1"

	// Mock calls
	classData := types.ClassData{
		Metadata: types.JsonInput(`{"1234": "5678"}`),
		Parent: types.ClassParent{
			Type:    types.ClassParentType_ACCOUNT,
			Account: ownerAddress,
		},
		Config: types.ClassConfig{
			Burnable:  false,
			MaxSupply: uint64(5),
			BlindBoxConfig: &types.BlindBoxConfig{
				MintPeriods: []types.MintPeriod{
					{
						StartTime:        time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
						AllowedAddresses: []string{},
						MintPrice:        1000,
					},
				},
				RevealTime: time.Date(2022, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		ToBeRevealed: true,
	}
	classDataInAny, _ := cdctypes.NewAnyWithValue(&classData)
	nftKeeper.EXPECT().GetClass(gomock.Any(), classId).Return(nft.Class{
		Id:          classId,
		Name:        "Class Name",
		Symbol:      "ABC",
		Description: "Testing Class 123",
		Uri:         "ipfs://abcdef",
		UriHash:     "abcdef",
		Data:        classDataInAny,
	}, true).MinTimes(1)
	keeper.SetClassesByAccount(ctx, types.ClassesByAccount{
		Account:  ownerAddress,
		ClassIds: []string{classId},
	})

	nftKeeper.EXPECT().GetTotalSupply(gomock.Any(), classId).Return(uint64(0))

	// Run
	nftInput := types.NFTInput{
		Uri:      "ipfs://123456",
		UriHash:  "123456",
		Metadata: types.JsonInput(`{"aaaa": "bbbb"}`),
	}
	res, err := msgServer.UpdateMintableNFT(goCtx, &types.MsgUpdateMintableNFT{
		Creator: ownerAddress,
		ClassId: classId,
		Id:      mintableId,
		Input:   nftInput,
	})

	// Check output
	require.Error(t, err)
	require.Contains(t, err.Error(), types.ErrMintableNftNotFound.Error())
	require.Nil(t, res)

	ctrl.Finish()
}
