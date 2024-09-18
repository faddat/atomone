package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/atomone-hub/atomone/x/gov/types"
	v1 "github.com/atomone-hub/atomone/x/gov/types/v1"
)

// GetGovernor returns the governor with the provided address
func (k Keeper) GetGovernor(ctx sdk.Context, addr types.GovernorAddress) (v1.Governor, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GovernorKey(addr))
	if bz == nil {
		return v1.Governor{}, false
	}

	var governor v1.Governor
	v1.MustMarshalGovernor(k.cdc, &governor)
	return governor, true
}

// SetGovernor sets the governor in the store
func (k Keeper) SetGovernor(ctx sdk.Context, governor v1.Governor) {
	store := ctx.KVStore(k.storeKey)
	bz := v1.MustMarshalGovernor(k.cdc, &governor)
	store.Set(types.GovernorKey(governor.GetAddress()), bz)
}

// GetAllGovernors returns all governors
func (k Keeper) GetAllGovernors(ctx sdk.Context) (governors v1.Governors) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.GovernorKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		governor := v1.MustUnmarshalGovernor(k.cdc, iterator.Value())
		governors = append(governors, &governor)
	}

	return governors
}

// GetAllActiveGovernors returns all active governors
func (k Keeper) GetAllActiveGovernors(ctx sdk.Context) (governors v1.Governors) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.GovernorKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		governor := v1.MustUnmarshalGovernor(k.cdc, iterator.Value())
		if governor.IsActive() {
			governors = append(governors, &governor)
		}
	}

	return governors
}

// IterateGovernors iterates over all governors and performs a callback function
func (k Keeper) IterateGovernors(ctx sdk.Context, cb func(index int64, governor v1.GovernorI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.GovernorKeyPrefix)
	defer iterator.Close()

	for i := int64(0); iterator.Valid(); iterator.Next() {
		governor := v1.MustUnmarshalGovernor(k.cdc, iterator.Value())
		if cb(i, governor) {
			break
		}
		i++
	}
}

// governor by power index
func (k Keeper) SetGovernorByPowerIndex(ctx sdk.Context, governor v1.Governor) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GovernorsByPowerKey(governor.GetAddress(), governor.GetVotingPower()), governor.GetAddress())
}

// governor by power index
func (k Keeper) DeleteGovernorByPowerIndex(ctx sdk.Context, governor v1.Governor) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GovernorsByPowerKey(governor.GetAddress(), governor.GetVotingPower()))
}

// UpdateGovernorByPowerIndex updates the governor in the governor by power index
func (k Keeper) UpdateGovernorByPowerIndex(ctx sdk.Context, governor v1.Governor) {
	oldGovernor, _ := k.GetGovernor(ctx, governor.GetAddress())
	k.DeleteGovernorByPowerIndex(ctx, oldGovernor)
	k.SetGovernorByPowerIndex(ctx, governor)
	k.SetGovernor(ctx, governor)
}

// IterateMaxGovernorsByGovernancePower iterates over the top params.MaxGovernors governors by governance power
func (k Keeper) IterateMaxGovernorsByGovernancePower(ctx sdk.Context, cb func(index int64, governor v1.GovernorI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	maxGovernors := k.GetParams(ctx).MaxGovernors
	var totGovernors uint64 = 0

	iterator := sdk.KVStoreReversePrefixIterator(store, types.GovernorsByPowerKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid() && totGovernors <= maxGovernors; iterator.Next() {
		// the value stored is the governor address
		governorAddr := types.GovernorAddress(iterator.Value())
		governor, _ := k.GetGovernor(ctx, governorAddr)
		if governor.IsActive() {
			if cb(int64(totGovernors), governor) {
				break
			}
			totGovernors++
		}
	}
}