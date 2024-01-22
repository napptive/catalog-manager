/**
 * Copyright 2023 Napptive
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package utils

import (
	"context"
	"fmt"
	"github.com/napptive/catalog-manager/internal/pkg/entities"
	njwthelper "github.com/napptive/njwt/pkg/helper"
	njwt "github.com/napptive/njwt/pkg/njwt"
	"github.com/rs/xid"
	"google.golang.org/grpc/metadata"
	"os"
	"syreclabs.com/go/faker"
	"time"
)

// RunIntegrationTests checks whether integration tests should be executed.
func RunIntegrationTests(id string) bool {
	var runIntegration = os.Getenv("RUN_INTEGRATION_TEST")
	if runIntegration == "all" {
		return true
	}
	return runIntegration == id
}

func CreateTestApplicationMetadata() *entities.ApplicationMetadata {
	return &entities.ApplicationMetadata{
		APIVersion:  "core.napptive.com/v1alpha1",
		Kind:        "ApplicationMetadata",
		Name:        faker.App().String(),
		Version:     faker.App().Version(),
		Description: "faker.App().String()",
		Keywords:    []string{"tag1, tag2"},
		License:     "Apache License Version 2.0",
		Url:         faker.Internet().Url(),
		Doc:         faker.Internet().Url(),
		Requires: entities.ApplicationRequirement{
			Traits: []string{"trait2, trait2"},
			Scopes: []string{"scope2, scope2"},
			K8s: []entities.KubernetesEntities{{
				ApiVersion: "core.oam.dev/v1alpha1",
				Kind:       "CustomEntityKind",
				Name:       "name",
			}},
		},
		Logo: []entities.ApplicationLogo{
			{
				Src:  faker.Internet().Url(),
				Type: "image/png",
				Size: "120x120",
			},
			{
				Src:  faker.Internet().Url(),
				Type: "image/png",
				Size: "1024x256",
			},
		},
	}
}

var metadataExample = `
apiVersion: core.napptive.com/v1alpha1
kind: ApplicationMetadata
name: "NGINX server"
version: 1.20.0
description: NGINX Service Mesh
keywords:
  - "storage"
license: "Apache License Version 2.0"
url: "https://www.nginx.com/"
doc: "http://nginx.org/"
logo:
  - src: "https://my.domain/path/logo.png"
    type: "image/png"
    size: "120x120"
`

func CreateTestApplicationInfo() *entities.ApplicationInfo {
	return &entities.ApplicationInfo{
		Namespace:       faker.Name().FirstName(),
		ApplicationName: faker.App().Name(),
		Tag:             faker.App().Version(),
		Readme:          faker.Lorem().Paragraph(10),
		Metadata:        metadataExample,
		MetadataName:    faker.Name().FirstName(),
	}
}

var metadataWithoutLogoExample = `
apiVersion: core.napptive.com/v1alpha1
kind: ApplicationMetadata
name: "NGINX server"
version: 1.20.0
description: NGINX Service Mesh
keywords:
  - "storage"
license: "Apache License Version 2.0"
url: "https://www.nginx.com/"
doc: "http://nginx.org/"

`

func CreateTestApplicationInfoWithoutLogo() *entities.ApplicationInfo {
	return &entities.ApplicationInfo{
		Namespace:       faker.Name().FirstName(),
		ApplicationName: faker.App().Name(),
		Tag:             faker.App().Version(),
		Readme:          faker.Lorem().Paragraph(10),
		Metadata:        metadataWithoutLogoExample,
		MetadataName:    faker.Name().FirstName(),
	}
}

// CreateTestJWTAuthIncomingContext creates a test context with metadata as found
// after passing the interceptor. The jwt elements allow to test scenarios where
// no JWT has been provided in the request.
func CreateTestJWTAuthIncomingContext(username string, accountName string, accountAdmin bool, jwtKey string, jwt string) context.Context {

	userID := GetUserID()
	accountID := GetAccountID()
	environmentID := GetEnvironmentID()
	zoneID := GetZoneID()
	role := "Member"
	if accountAdmin {
		role = "Admin"
	}
	claim := njwt.NewAuthxClaim(userID, username, accountID, accountName, environmentID,
		accountAdmin, zoneID, "zone-test.napptive.dev",
		[]njwt.UserAccountClaim{njwt.UserAccountClaim{
			Id:   accountID,
			Name: accountName,
			Role: role,
		}})
	claimMap := claim.ToMap()
	claimMap[jwtKey] = jwt
	claimMap[njwthelper.JWTID] = GetTokenID()
	claimMap[njwthelper.JWTIssuedAt] = fmt.Sprint(time.Now().Add(time.Minute * (-30)).Unix())
	md := metadata.New(claimMap)
	parentCtx := context.Background()
	return metadata.NewIncomingContext(parentCtx, md)
}

// GetUserID generates a random UserId
func GetUserID() string {
	return xid.New().String()
}

// GetAccountID generates a random AccountId
func GetAccountID() string {
	return xid.New().String()
}

// GetEnvironmentID generates a random environmentId
func GetEnvironmentID() string {
	return xid.New().String()
}

// GetTokenID generates a random token ID
func GetTokenID() string {
	return xid.New().String()
}

// GetZoneID generates a random zone ID
func GetZoneID() string {
	return xid.New().String()
}
