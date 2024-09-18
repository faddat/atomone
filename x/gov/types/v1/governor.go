package v1

import (
	"bytes"
	"sort"
	"strings"
	time "time"

	"cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/atomone-hub/atomone/x/gov/types"
)

var (
	GovernorStatusUnspecified = GovernorStatus_name[int32(Unspecified)]
	GovernorStatusActive      = GovernorStatus_name[int32(Active)]
	GovernorStatusInactive    = GovernorStatus_name[int32(Inactive)]
)

var _ GovernorI = Governor{}

// NewGovernor constructs a new Governor
func NewGovernor(address string, description GovernorDescription, creationTime time.Time) (Governor, error) {
	return Governor{
		GovernorAddress:      address,
		Description:          description,
		Status:               Active,
		VotingPower:          sdk.ZeroDec(),
		LastStatusChangeTime: &creationTime,
	}, nil
}

// Governors is a collection of Governor
type Governors []*Governor

func (g Governors) String() (out string) {
	for _, gov := range g {
		out += gov.String() + "\n"
	}

	return strings.TrimSpace(out)
}

// Sort Governors sorts governor array in ascending governor address order
func (g Governors) Sort() {
	sort.Sort(g)
}

// Implements sort interface
func (g Governors) Len() int {
	return len(g)
}

// Implements sort interface
func (g Governors) Less(i, j int) bool {
	return bytes.Compare(g[i].GetAddress().Bytes(), g[j].GetAddress().Bytes()) == -1
}

// Implements sort interface
func (g Governors) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// GovernorsByVotingPower implements sort.Interface for []Governor based on
// the VotingPower and Address fields.
// The governors are sorted first by their voting power (descending). Secondary index - Address (ascending).
// Copied from tendermint/types/validator_set.go
type GovernorsByVotingPower []Governor

func (govz GovernorsByVotingPower) Len() int { return len(govz) }

func (govz GovernorsByVotingPower) Less(i, j int) bool {
	if govz[i].GetVotingPower().Equal(govz[j].GetVotingPower()) {
		addrI := govz[i].GetAddress()
		addrJ := govz[j].GetAddress()
		return bytes.Compare(addrI, addrJ) == -1
	}
	return govz[i].GetVotingPower().GT(govz[j].GetVotingPower())
}

func (govz GovernorsByVotingPower) Swap(i, j int) {
	govz[i], govz[j] = govz[j], govz[i]
}

func MustMarshalGovernor(cdc codec.BinaryCodec, governor *Governor) []byte {
	return cdc.MustMarshal(governor)
}

func MustUnmarshalGovernor(cdc codec.BinaryCodec, value []byte) Governor {
	governor, err := UnmarshalGovernor(cdc, value)
	if err != nil {
		panic(err)
	}

	return governor
}

// unmarshal a redelegation from a store value
func UnmarshalGovernor(cdc codec.BinaryCodec, value []byte) (g Governor, err error) {
	err = cdc.Unmarshal(value, &g)
	return g, err
}

// IsActive checks if the governor status equals Active
func (g Governor) IsActive() bool {
	return g.GetStatus() == Active
}

// IsInactive checks if the governor status equals Inactive
func (g Governor) IsInactive() bool {
	return g.GetStatus() == Inactive
}

func NewGovernorDescription(moniker, identity, website, securityContact, details string) GovernorDescription {
	return GovernorDescription{
		Moniker:         moniker,
		Identity:        identity,
		Website:         website,
		SecurityContact: securityContact,
		Details:         details,
	}
}

// UpdateDescription updates the fields of a given description. An error is
// returned if the resulting description contains an invalid length.
func (d GovernorDescription) UpdateDescription(d2 GovernorDescription) (GovernorDescription, error) {
	if d2.Moniker == stakingtypes.DoNotModifyDesc {
		d2.Moniker = d.Moniker
	}

	if d2.Identity == stakingtypes.DoNotModifyDesc {
		d2.Identity = d.Identity
	}

	if d2.Website == stakingtypes.DoNotModifyDesc {
		d2.Website = d.Website
	}

	if d2.SecurityContact == stakingtypes.DoNotModifyDesc {
		d2.SecurityContact = d.SecurityContact
	}

	if d2.Details == stakingtypes.DoNotModifyDesc {
		d2.Details = d.Details
	}

	return NewGovernorDescription(
		d2.Moniker,
		d2.Identity,
		d2.Website,
		d2.SecurityContact,
		d2.Details,
	).EnsureLength()
}

// EnsureLength ensures the length of a vovernor's description.
func (d GovernorDescription) EnsureLength() (GovernorDescription, error) {
	if len(d.Moniker) > stakingtypes.MaxMonikerLength {
		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid moniker length; got: %d, max: %d", len(d.Moniker), stakingtypes.MaxMonikerLength)
	}

	if len(d.Identity) > stakingtypes.MaxIdentityLength {
		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid identity length; got: %d, max: %d", len(d.Identity), stakingtypes.MaxIdentityLength)
	}

	if len(d.Website) > stakingtypes.MaxWebsiteLength {
		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid website length; got: %d, max: %d", len(d.Website), stakingtypes.MaxWebsiteLength)
	}

	if len(d.SecurityContact) > stakingtypes.MaxSecurityContactLength {
		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid security contact length; got: %d, max: %d", len(d.SecurityContact), stakingtypes.MaxSecurityContactLength)
	}

	if len(d.Details) > stakingtypes.MaxDetailsLength {
		return d, errors.Wrapf(sdkerrors.ErrInvalidRequest, "invalid details length; got: %d, max: %d", len(d.Details), stakingtypes.MaxDetailsLength)
	}

	return d, nil
}

func (s GovernorStatus) IsValid() bool {
	return s == Active || s == Inactive
}

// GovernorStatusFromString returns a GovernorStatus from a string. It returns an
// error if the string is invalid.
func GovernorStatusFromString(str string) (GovernorStatus, error) {
	switch str {
	case GovernorStatusActive:
		return Active, nil
	case GovernorStatusInactive:
		return Inactive, nil
	default:
		return Unspecified, types.ErrInvalidGovernorStatus.Wrapf("unrecognized governor status %s", str)
	}
}

// MinEqual defines a more minimum set of equality conditions when comparing two
// governors.
func (g *Governor) MinEqual(other *Governor) bool {
	return g.GovernorAddress == other.GovernorAddress &&
		g.Status == other.Status &&
		g.Description.Equal(other.Description) &&
		g.VotingPower.Equal(other.VotingPower)
}

// Equal checks if the receiver equals the parameter
func (g *Governor) Equal(v2 *Governor) bool {
	return g.MinEqual(v2)
}

func (g Governor) GetVotingPower() sdk.Dec             { return g.VotingPower }
func (g Governor) GetMoniker() string                  { return g.Description.Moniker }
func (g Governor) GetStatus() GovernorStatus           { return g.Status }
func (g Governor) GetDescription() GovernorDescription { return g.Description }
func (g Governor) GetLastStatusChangeTime() *time.Time { return g.LastStatusChangeTime }
func (g Governor) GetAddress() types.GovernorAddress {
	return types.MustGovernorAddressFromBech32(g.GovernorAddress)
}