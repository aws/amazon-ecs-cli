// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package regcreds

import (
	"encoding/json"

	kmsClient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/kms"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcredio"
)

const (
	rolePolicyVersion = "2012-10-17"
)

// PolicyDocument contains the statements that make up an IAM policy
type PolicyDocument struct {
	Version   string
	Statement []StatementEntry
}

// StatementEntry contains a set of actions and the resources they apply to
type StatementEntry struct {
	Effect   string
	Action   []string
	Resource []string
}

func generateSecretsPolicy(credEntries map[string]regcredio.CredsOutputEntry, kmsClient kmsClient.Client) (string, error) {
	policyStatements := make([]StatementEntry, 0, len(credEntries))

	for _, entry := range credEntries {
		keyARN := ""
		if entry.KMSKeyID != "" {
			validARN, err := kmsClient.GetValidKeyARN(entry.KMSKeyID)
			if err != nil {
				return "", err
			}
			keyARN = validARN
		}
		statement := generatePolicyStatement(entry.CredentialARN, keyARN)
		policyStatements = append(policyStatements, statement)
	}

	policyDoc := PolicyDocument{Version: rolePolicyVersion, Statement: policyStatements}
	policyBytes, err := json.Marshal(&policyDoc)
	if err != nil {
		return "", err
	}

	return string(policyBytes), nil
}

func generatePolicyStatement(credARN, kmsKeyARN string) StatementEntry {
	customDecryptActions := []string{"kms:Decrypt", "secretsmanager:GetSecretValue"}
	defaultActions := []string{"secretsmanager:GetSecretValue"}

	if kmsKeyARN != "" {
		return StatementEntry{
			Effect:   "Allow",
			Action:   customDecryptActions,
			Resource: []string{kmsKeyARN, credARN},
		}
	}
	// TODO: look for unspecified KMS Key on in-region secrets
	return StatementEntry{
		Effect:   "Allow",
		Action:   defaultActions,
		Resource: []string{credARN},
	}
}
