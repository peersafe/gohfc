package cauthdsl

import (
	"fmt"

	"math"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/policies"
	"github.com/hyperledger/fabric/msp"
	cb "github.com/hyperledger/fabric/protos/common"
	mp "github.com/hyperledger/fabric/protos/msp"
)

type roleProvider struct {
	mspManager msp.MSPManager
}

// NewRolePolicyProvider provides a role policy generator for cauthdsl type policies
func NewRolePolicyProvider(mspManager msp.MSPManager) policies.Provider {
	return &roleProvider{
		mspManager: mspManager,
	}
}

func (p *roleProvider) NewPolicy(data []byte) (policies.Policy, proto.Message, error) {
	rolePolicy := &cb.RolePolicy{}
	if err := proto.Unmarshal(data, rolePolicy); err != nil {
		return nil, nil, fmt.Errorf("Error unmarshaling to RolePolicy: %s", err)
	}
	cauthdslLogger.Debugf("New RolePolicy: Role:%v, Percent:%d", rolePolicy.Role, rolePolicy.Percent)
	evaluator, err := roleEvaluator(rolePolicy, p.mspManager)
	if err != nil {
		return nil, nil, err
	}

	return &policy{
		evaluator:    evaluator,
		deserializer: p.mspManager,
	}, rolePolicy, nil
}

// returns siguatures of a policy need
func PolicyCalculate(mspManager msp.MSPManager, policy *cb.RolePolicy) (int, error) {
	if mspManager == nil {
		return 0, fmt.Errorf("mspManager is nil")
	}

	mspMap, _ := mspManager.GetMSPs()
	if len(mspMap) == 0 {
		return 0, fmt.Errorf("Get none msp from mspManager")
	}
	coreMembers := 0
	for mspid, v := range mspMap {
		cauthdslLogger.Debugf("MspMap: MspID:%v, MSPOrgRole:%v", mspid, v.GetMSPOrgRole())
		if v.GetMSPOrgRole() == mp.MSPOrgRole_CORE {
			coreMembers++
		}
	}

	if coreMembers == 0 {
		cauthdslLogger.Error("There are no core members in this channel")
		return 0, fmt.Errorf("There are no core members in this channel")
	}

	if policy.Percent <= 0 {
		return 0, nil
	}
	if policy.Percent > 100 {
		policy.Percent = 100
	}

	need := math.Ceil(float64(policy.Percent) / 100 * float64(coreMembers))
	return int(need), nil
}

// returns
func MSPSatisfyRolePolicy(signMsp msp.MSP, policy *cb.RolePolicy) bool {
	signMspRole := signMsp.GetMSPOrgRole()
	cauthdslLogger.Debugf("Signature MSP Role:%v, %d", signMspRole, signMspRole)
	switch policy.Role {
	case mp.MSPOrgRole_CORE:
		if signMspRole != mp.MSPOrgRole_CORE {
			cauthdslLogger.Debugf("Signature MSP Role %v is not satisfy CORE, continue", signMspRole)
			return false
		}
	case mp.MSPOrgRole_MAIN:
		if signMspRole != mp.MSPOrgRole_CORE && signMspRole != mp.MSPOrgRole_MAIN && signMspRole != mp.MSPOrgRole_ORDERER {
			cauthdslLogger.Debugf("Signature MSP Role %v is not satisfy MAIN, continue", signMspRole)
			return false
		}
	case mp.MSPOrgRole_NORMAL:
		if signMspRole != mp.MSPOrgRole_CORE && signMspRole != mp.MSPOrgRole_MAIN && signMspRole != mp.MSPOrgRole_NORMAL && signMspRole != mp.MSPOrgRole_ORDERER {
			cauthdslLogger.Debugf("Signature MSP Role %v is not satisfy NORMAL, continue", signMspRole)
			return false
		}
	case mp.MSPOrgRole_ORDERER:
		if signMspRole != mp.MSPOrgRole_CORE && signMspRole != mp.MSPOrgRole_ORDERER {
			cauthdslLogger.Debugf("Signature MSP Role %v is not satisfy ORDERER, continue", signMspRole)
			return false
		}
	default:
		cauthdslLogger.Errorf("Signature MSP Role %v is not valid", signMspRole)
		return false
	}
	return true
}

func roleEvaluator(policy *cb.RolePolicy, mspManager msp.MSPManager) (func([]IdentityAndSignature, []bool) bool, error) {
	mspMap, _ := mspManager.GetMSPs()
	if len(mspMap) == 0 {
		return nil, fmt.Errorf("Get none msp from mspManager")
	}

	return func(signedData []IdentityAndSignature, used []bool) bool {
		cauthdslLogger.Debug("==========Start evaluate role policy==========")
		cnt := 0
		need, err := PolicyCalculate(mspManager, policy)
		if err != nil {
			cauthdslLogger.Error(err)
			return false
		}

		cauthdslLogger.Debugf("Need msp number: %v, role: %v", need, policy.Role)
		for _, sd := range signedData {
			identity, err := sd.Identity()
			if err != nil {
				cauthdslLogger.Errorf("Principal deserialization failure (%s) for identity %x", err, sd.Identity)
				continue
			}
			mspid := identity.GetIdentifier().Mspid
			cauthdslLogger.Debugf("Signature MSP ID:%v", mspid)
			signMsp, ok := mspMap[mspid]
			if !ok {
				cauthdslLogger.Warningf("Cloud not find %v in mspMsp", mspid)
				continue
			}

			if !MSPSatisfyRolePolicy(signMsp, policy) {
				continue
			}

			err = sd.Verify()
			if err != nil {
				cauthdslLogger.Debugf("signature for identity %v is invalid: %s", mspid, err)
				continue
			}
			cnt++
			if cnt >= int(need) {
				return true
			}
		}
		return false
	}, nil
}
