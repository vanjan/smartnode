package node

import (
	"math"
	"math/big"
	"time"

	"github.com/rocket-pool/rocketpool-go/dao/trustednode"
	"github.com/rocket-pool/rocketpool-go/node"
	"github.com/rocket-pool/rocketpool-go/rewards"
	"github.com/rocket-pool/rocketpool-go/tokens"
	"github.com/rocket-pool/rocketpool-go/utils/eth"
	"github.com/urfave/cli"
	"golang.org/x/sync/errgroup"

	"github.com/rocket-pool/smartnode/shared/services"
	"github.com/rocket-pool/smartnode/shared/types/api"
)


func getRewards(c *cli.Context) (*api.NodeRewardsResponse, error) {

    // Get services
    if err := services.RequireNodeWallet(c); err != nil { return nil, err }
    if err := services.RequireRocketStorage(c); err != nil { return nil, err }
    w, err := services.GetWallet(c)
    if err != nil { return nil, err }
    rp, err := services.GetRocketPool(c)
    if err != nil { return nil, err }

    // Response
    response := api.NodeRewardsResponse{}

    // Get node account
    nodeAccount, err := w.GetNodeAccount()
    if err != nil {
        return nil, err
    }

    var periodStart time.Time
    var rewardsInterval time.Duration
    var effectiveStake *big.Int
    var totalEffectiveStake *big.Int
    var totalRplSupply *big.Int
    var inflationInterval *big.Int
    var odaoSize uint64
    var nodeOperatorRewardsPercent float64
    var trustedNodeOperatorRewardsPercent float64

    // Sync
    var wg errgroup.Group

    // Get node trusted status
    wg.Go(func() error {
        trusted, err := trustednode.GetMemberExists(rp, nodeAccount.Address, nil)
        if err == nil {
            response.Trusted = trusted
        }
        return err
    })

    // Get cumulative rewards
    wg.Go(func() error {
        rewards, err := rewards.CalculateLifetimeNodeRewards(rp, nodeAccount.Address)
        if err == nil {
            response.CumulativeRewards = eth.WeiToEth(rewards)
        }
        return err
    })

    // Get cumulative ODAO rewards
    wg.Go(func() error {
        rewards, err := rewards.CalculateLifetimeTrustedNodeRewards(rp, nodeAccount.Address)
        if err == nil {
            response.CumulativeTrustedRewards = eth.WeiToEth(rewards)
        }
        return err
    })

    // Get the start of the rewards checkpoint
    wg.Go(func() error {
        periodStart, err = rewards.GetClaimIntervalTimeStart(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the rewards checkpoint interval
    wg.Go(func() error {
        rewardsInterval, err = rewards.GetClaimIntervalTime(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the node's effective stake
    wg.Go(func() error {
        effectiveStake, err = node.GetNodeEffectiveRPLStake(rp, nodeAccount.Address, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the total network effective stake
    wg.Go(func() error {
        totalEffectiveStake, err = node.GetTotalEffectiveRPLStake(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the total RPL supply
    wg.Go(func() error {
        totalRplSupply, err = tokens.GetRPLTotalSupply(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the RPL inflation interval
    wg.Go(func() error {
        inflationInterval, err = tokens.GetRPLInflationIntervalRate(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the ODAO member count
    wg.Go(func() error {
        odaoSize, err = trustednode.GetMemberCount(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the node operator rewards percent
    wg.Go(func() error {
        nodeOperatorRewardsPercent, err = rewards.GetNodeOperatorRewardsPercent(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Get the trusted node operator rewards percent
    wg.Go(func() error {
        trustedNodeOperatorRewardsPercent, err = rewards.GetTrustedNodeOperatorRewardsPercent(rp, nil)
        if err != nil {
            return err
        }
        return nil
    })

    // Wait for data
    if err := wg.Wait(); err != nil {
        return nil, err
    }

    // Set the time until the next rewards checkpoint
    response.TimeToCheckpoint = time.Now().Sub(periodStart.Add(rewardsInterval))

    // Calculate the estimated rewards
    rewardsIntervalDays := rewardsInterval.Seconds() / (60*60*24)
    inflationPerDay := eth.WeiToEth(inflationInterval)
    totalRplAtNextCheckpoint := (math.Pow(inflationPerDay, float64(rewardsIntervalDays)) - 1) * eth.WeiToEth(totalRplSupply)

    if totalEffectiveStake.Cmp(big.NewInt(0)) == 1 {
        response.EstimatedRewards = eth.WeiToEth(effectiveStake) / eth.WeiToEth(totalEffectiveStake) * totalRplAtNextCheckpoint * nodeOperatorRewardsPercent
    }

    if response.Trusted {
        response.EstimatedTrustedRewards = totalRplAtNextCheckpoint * trustedNodeOperatorRewardsPercent / float64(odaoSize)
    }

    // Return response
    return &response, nil

}
