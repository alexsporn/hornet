package debug

import (
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/gohornet/hornet/pkg/dag"
	"github.com/gohornet/hornet/pkg/model/hornet"
	"github.com/gohornet/hornet/pkg/model/storage"
	"github.com/gohornet/hornet/pkg/model/utxo"
	"github.com/gohornet/hornet/pkg/restapi"
	v1 "github.com/gohornet/hornet/plugins/restapi/v1"
)

func debugOutputsIDs(c echo.Context) (*outputIDsResponse, error) {

	outputIDs := []string{}
	outputConsumerFunc := func(output *utxo.Output) bool {
		outputIDs = append(outputIDs, output.OutputID().ToHex())
		return true
	}

	err := deps.UTXO.ForEachOutput(outputConsumerFunc, utxo.ReadLockLedger(false))
	if err != nil {
		return nil, errors.WithMessagef(restapi.ErrInternalError, "reading unspent outputs failed, error: %s", err)
	}

	return &outputIDsResponse{
		OutputIDs: outputIDs,
	}, nil
}

func debugUnspentOutputsIDs(c echo.Context) (*outputIDsResponse, error) {

	outputIDs := []string{}
	outputConsumerFunc := func(output *utxo.Output) bool {
		outputIDs = append(outputIDs, output.OutputID().ToHex())
		return true
	}

	err := deps.UTXO.ForEachUnspentOutput(outputConsumerFunc, utxo.ReadLockLedger(false))
	if err != nil {
		return nil, errors.WithMessagef(restapi.ErrInternalError, "reading unspent outputs failed, error: %s", err)
	}

	return &outputIDsResponse{
		OutputIDs: outputIDs,
	}, nil
}

func debugSpentOutputsIDs(c echo.Context) (*outputIDsResponse, error) {

	outputIDs := []string{}

	spentConsumerFunc := func(spent *utxo.Spent) bool {
		outputIDs = append(outputIDs, spent.OutputID().ToHex())
		return true
	}

	err := deps.UTXO.ForEachSpentOutput(spentConsumerFunc, utxo.ReadLockLedger(false))
	if err != nil {
		return nil, errors.WithMessagef(restapi.ErrInternalError, "reading spent outputs failed, error: %s", err)
	}

	return &outputIDsResponse{
		OutputIDs: outputIDs,
	}, nil
}

func debugAddresses(c echo.Context) (*addressesResponse, error) {

	addressMap := map[string]*address{}

	outputConsumerFunc := func(output *utxo.Output) bool {
		if addr, exists := addressMap[output.Address().String()]; exists {
			// add balance to total balance
			addr.Balance += output.Amount()
			return true
		}

		addressMap[output.Address().String()] = &address{
			AddressType: output.Address().Type(),
			Address:     output.Address().String(),
			Balance:     output.Amount(),
		}

		return true
	}

	err := deps.UTXO.ForEachUnspentOutput(outputConsumerFunc, utxo.ReadLockLedger(false))
	if err != nil {
		return nil, errors.WithMessagef(restapi.ErrInternalError, "reading addresses failed, error: %s", err)
	}

	addresses := make([]*address, 0, len(addressMap))
	for _, addr := range addressMap {
		addresses = append(addresses, addr)
	}

	return &addressesResponse{
		Addresses: addresses,
	}, nil
}

func debugAddressesEd25519(c echo.Context) (*addressesResponse, error) {

	addressMap := map[string]*address{}

	outputConsumerFunc := func(output *utxo.Output) bool {
		// ToDo: allow ed25519 address type only

		if addr, exists := addressMap[output.Address().String()]; exists {
			// add balance to total balance
			addr.Balance += output.Amount()
			return true
		}

		addressMap[output.Address().String()] = &address{
			AddressType: output.Address().Type(),
			Address:     output.Address().String(),
			Balance:     output.Amount(),
		}

		return true
	}

	err := deps.UTXO.ForEachUnspentOutput(outputConsumerFunc, utxo.ReadLockLedger(false))
	if err != nil {
		return nil, errors.WithMessagef(restapi.ErrInternalError, "reading addresses failed, error: %s", err)
	}

	addresses := make([]*address, 0, len(addressMap))
	for _, addr := range addressMap {
		addresses = append(addresses, addr)
	}

	return &addressesResponse{
		Addresses: addresses,
	}, nil
}

func debugMilestoneDiff(c echo.Context) (*milestoneDiffResponse, error) {

	msIndex, err := v1.ParseMilestoneIndexParam(c)
	if err != nil {
		return nil, err
	}

	diff, err := deps.UTXO.GetMilestoneDiffWithoutLocking(msIndex)

	outputs := make([]*v1.OutputResponse, len(diff.Outputs))
	spents := make([]*v1.OutputResponse, len(diff.Spents))

	for i, output := range diff.Outputs {
		o, err := v1.NewOutputResponse(output, false)
		if err != nil {
			return nil, err
		}
		outputs[i] = o
	}

	for i, spent := range diff.Spents {
		o, err := v1.NewOutputResponse(spent.Output(), true)
		if err != nil {
			return nil, err
		}
		spents[i] = o
	}

	return &milestoneDiffResponse{
		MilestoneIndex: msIndex,
		Outputs:        outputs,
		Spents:         spents,
	}, nil
}

func debugRequests(c echo.Context) (*requestsResponse, error) {

	queued, pending, processing := deps.RequestQueue.Requests()
	debugReqs := make([]*request, 0, len(queued)+len(pending)+len(processing))

	for _, req := range queued {
		debugReqs = append(debugReqs, &request{
			MessageID:        req.MessageID.ToHex(),
			Type:             "queued",
			MessageExists:    deps.Storage.ContainsMessage(req.MessageID),
			EnqueueTimestamp: req.EnqueueTime.Format(time.RFC3339),
			MilestoneIndex:   req.MilestoneIndex,
		})
	}

	for _, req := range pending {
		debugReqs = append(debugReqs, &request{
			MessageID:        req.MessageID.ToHex(),
			Type:             "pending",
			MessageExists:    deps.Storage.ContainsMessage(req.MessageID),
			EnqueueTimestamp: req.EnqueueTime.Format(time.RFC3339),
			MilestoneIndex:   req.MilestoneIndex,
		})
	}

	for _, req := range processing {
		debugReqs = append(debugReqs, &request{
			MessageID:        req.MessageID.ToHex(),
			Type:             "processing",
			MessageExists:    deps.Storage.ContainsMessage(req.MessageID),
			EnqueueTimestamp: req.EnqueueTime.Format(time.RFC3339),
			MilestoneIndex:   req.MilestoneIndex,
		})
	}

	return &requestsResponse{
		Requests: debugReqs,
	}, nil
}

func debugMessageCone(c echo.Context) (*messageConeResponse, error) {
	messageIDHex := strings.ToLower(c.Param(ParameterMessageID))

	messageID, err := hornet.MessageIDFromHex(messageIDHex)
	if err != nil {
		return nil, errors.WithMessagef(restapi.ErrInvalidParameter, "invalid message ID: %s, error: %s", messageIDHex, err)
	}

	cachedStartMsgMeta := deps.Storage.GetCachedMessageMetadataOrNil(messageID) // meta +1
	if cachedStartMsgMeta == nil {
		return nil, errors.WithMessagef(restapi.ErrInvalidParameter, "message not found: %s", messageIDHex)
	}
	defer cachedStartMsgMeta.Release(true)

	if !cachedStartMsgMeta.GetMetadata().IsSolid() {
		return nil, errors.WithMessagef(restapi.ErrInvalidParameter, "start message is not solid: %s", messageIDHex)
	}

	startMsgReferened, startMsgReferenedAt := cachedStartMsgMeta.GetMetadata().GetReferenced()

	entryPointIndex := deps.Storage.GetSnapshotInfo().EntryPointIndex
	entryPoints := []*entryPoint{}
	tanglePath := []*messageWithParents{}

	if err := dag.TraverseParentsOfMessage(deps.Storage, messageID,
		// traversal stops if no more messages pass the given condition
		// Caution: condition func is not in DFS order
		func(cachedMsgMeta *storage.CachedMetadata) (bool, error) { // meta +1
			defer cachedMsgMeta.Release(true) // meta -1

			if referenced, at := cachedMsgMeta.GetMetadata().GetReferenced(); referenced {
				if !startMsgReferened || (at < startMsgReferenedAt) {
					entryPoints = append(entryPoints, &entryPoint{MessageID: cachedMsgMeta.GetMetadata().GetMessageID().ToHex(), ReferencedByMilestone: at})
					return false, nil
				}
			}

			return true, nil
		},
		// consumer
		func(cachedMsgMeta *storage.CachedMetadata) error { // meta +1
			cachedMsgMeta.ConsumeMetadata(func(metadata *storage.MessageMetadata) { // meta -1
				tanglePath = append(tanglePath,
					&messageWithParents{
						MessageID: metadata.GetMessageID().ToHex(),
						Parents:   metadata.GetParents().ToHex(),
					},
				)
			})

			return nil
		},
		// called on missing parents
		// return error on missing parents
		nil,
		// called on solid entry points
		func(messageID hornet.MessageID) {
			entryPoints = append(entryPoints, &entryPoint{MessageID: messageID.ToHex(), ReferencedByMilestone: entryPointIndex})
		},
		false, nil); err != nil {
		return nil, errors.WithMessagef(restapi.ErrInternalError, "traverse parents failed, error: %s", err)
	}

	if len(entryPoints) == 0 {
		return nil, errors.WithMessagef(restapi.ErrInternalError, "no referenced parents found: %s", messageIDHex)
	}

	return &messageConeResponse{
		ConeElementsCount: len(tanglePath),
		EntryPointsCount:  len(entryPoints),
		Cone:              tanglePath,
		EntryPoints:       entryPoints,
	}, nil
}