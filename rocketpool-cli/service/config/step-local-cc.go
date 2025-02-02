package config

import (
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/pbnjay/memory"
	"github.com/rocket-pool/smartnode/shared/services/config"
)

const localCcStepID string = "step-local-cc"

func createLocalCcStep(wiz *wizard, currentStep int, totalSteps int) *choiceWizardStep {

	// Get the list of clients
	badClients, badFallbackClients := wiz.md.Config.GetIncompatibleConsensusClients()

	// Create the button names and descriptions from the config
	clientNames := []string{"Random (Recommended)"}
	clientDescriptions := []string{"Select a client randomly to help promote the diversity of the Beacon Chain. We recommend you do this unless you have a strong reason to pick a specific client. To learn more about why client diversity is important, please visit https://clientdiversity.org for an explanation."}

	goodClients := []config.ParameterOption{}
	for _, client := range wiz.md.Config.ConsensusClient.Options {
		isGood := true
		for _, badClient := range badClients {
			if badClient.Value == client.Value {
				isGood = false
				break
			}
		}
		for _, badClient := range badFallbackClients {
			if badClient.Value == client.Value {
				isGood = false
				break
			}
		}
		if isGood {
			clientNames = append(clientNames, client.Name)
			clientDescriptions = append(clientDescriptions, getAugmentedCcDescription(client.Value.(config.ConsensusClient), client.Description))
			goodClients = append(goodClients, client)
		}
	}

	incompatibleClientWarning := ""
	if len(badClients) > 0 {
		badClientNames := []string{}
		for _, badClient := range badClients {
			badClientNames = append(badClientNames, badClient.Name)
		}
		incompatibleClientWarning = fmt.Sprintf("\n\n[orange]NOTE: The following clients are incompatible with your choice of Execution client: %s", strings.Join(badClientNames, ", "))
	} else if len(badFallbackClients) > 0 {
		badClientNames := []string{}
		for _, badClient := range badFallbackClients {
			badClientNames = append(badClientNames, badClient.Name)
		}
		incompatibleClientWarning = fmt.Sprintf("\n\n[orange]NOTE: The following clients are incompatible with your choice of fallback Execution client: %s", strings.Join(badClientNames, ", "))
	}

	helperText := fmt.Sprintf("Please select the Consensus client you would like to use.\n\nHighlight each one to see a brief description of it, or go to https://docs.rocketpool.net/guides/node/eth-clients.html#eth2-clients to learn more about them.%s", incompatibleClientWarning)

	show := func(modal *choiceModalLayout) {
		wiz.md.setPage(modal.page)
		modal.focus(0) // Catch-all for safety

		if wiz.md.isMigration || !wiz.md.isNew {
			var ccName string
			for _, option := range wiz.md.Config.ConsensusClient.Options {
				if option.Value == wiz.md.Config.ConsensusClient.Value {
					ccName = option.Name
					break
				}
			}
			for i, clientName := range clientNames {
				if ccName == clientName {
					modal.focus(i)
					break
				}
			}
		}
	}

	done := func(buttonIndex int, buttonLabel string) {
		if buttonIndex == 0 {
			wiz.md.pages.RemovePage(randomCcPrysmID)
			wiz.md.pages.RemovePage(randomCcID)
			selectRandomCC(goodClients, true, wiz, currentStep, totalSteps)
		} else {
			buttonLabel = strings.TrimSpace(buttonLabel)
			selectedClient := config.ConsensusClient_Unknown
			for _, client := range wiz.md.Config.ConsensusClient.Options {
				if client.Name == buttonLabel {
					selectedClient = client.Value.(config.ConsensusClient)
					break
				}
			}
			if selectedClient == config.ConsensusClient_Unknown {
				panic(fmt.Sprintf("Local CC selection buttons didn't match any known clients, buttonLabel = %s\n", buttonLabel))
			}
			wiz.md.Config.ConsensusClient.Value = selectedClient
			switch selectedClient {
			//case config.ConsensusClient_Prysm:
			//	wiz.consensusLocalPrysmWarning.show()
			case config.ConsensusClient_Teku:
				totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
				if runtime.GOARCH == "arm64" || totalMemoryGB < 15 {
					wiz.consensusLocalTekuWarning.show()
				} else {
					wiz.graffitiModal.show()
				}
			default:
				wiz.graffitiModal.show()
			}
		}
	}

	back := func() {
		wiz.consensusModeModal.show()
	}

	return newChoiceStep(
		wiz,
		currentStep,
		totalSteps,
		helperText,
		clientNames,
		clientDescriptions,
		100,
		"Consensus Client > Selection",
		DirectionalModalVertical,
		show,
		done,
		back,
		localCcStepID,
	)

}

// Get a random client compatible with the user's hardware and EC choices.
func selectRandomCC(goodOptions []config.ParameterOption, includeSupermajority bool, wiz *wizard, currentStep int, totalSteps int) {

	// Get system specs
	totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
	isLowPower := (totalMemoryGB < 15 || runtime.GOARCH == "arm64")

	// Filter out the clients based on system specs
	filteredClients := []config.ConsensusClient{}
	for _, clientOption := range goodOptions {
		client := clientOption.Value.(config.ConsensusClient)
		switch client {
		case config.ConsensusClient_Teku:
			if !isLowPower {
				filteredClients = append(filteredClients, client)
			}
		/*
			case config.ConsensusClient_Prysm:
				if includeSupermajority {
					filteredClients = append(filteredClients, client)
				}
		*/
		default:
			filteredClients = append(filteredClients, client)
		}
	}

	// Select a random client
	rand.Seed(time.Now().UnixNano())
	selectedClient := filteredClients[rand.Intn(len(filteredClients))]
	wiz.md.Config.ConsensusClient.Value = selectedClient

	// Show the selection page
	/*
		if selectedClient == config.ConsensusClient_Prysm {
			wiz.consensusLocalRandomPrysmModal = createRandomPrysmStep(wiz, currentStep, totalSteps, goodOptions)
			wiz.consensusLocalRandomPrysmModal.show()
		} else {
			wiz.consensusLocalRandomModal = createRandomStep(wiz, currentStep, totalSteps, goodOptions)
			wiz.consensusLocalRandomModal.show()
		}
	*/
	wiz.consensusLocalRandomModal = createRandomCCStep(wiz, currentStep, totalSteps, goodOptions)
	wiz.consensusLocalRandomModal.show()

}

// Get a more verbose client description, including warnings
func getAugmentedCcDescription(client config.ConsensusClient, originalDescription string) string {

	switch client {
	/*
		case config.ConsensusClient_Prysm:
			return fmt.Sprintf("%s\n\n[orange]NOTE: Prysm currently has a very high representation of the Beacon Chain. For the health of the network and the overall safety of your funds, please consider choosing a client with a lower representation. Please visit https://clientdiversity.org to learn more.", originalDescription)
	*/
	case config.ConsensusClient_Teku:
		totalMemoryGB := memory.TotalMemory() / 1024 / 1024 / 1024
		if runtime.GOARCH == "arm64" || totalMemoryGB < 15 {
			return fmt.Sprintf("%s\n\n[orange]WARNING: Teku is a resource-heavy client and will likely not perform well on your system given your CPU power or amount of available RAM. We recommend you pick a lighter client instead.", originalDescription)
		}
	}

	return originalDescription

}
