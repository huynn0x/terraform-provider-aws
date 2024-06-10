// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appfabric"
	"github.com/aws/aws-sdk-go-v2/service/appfabric/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfappfabric "github.com/hashicorp/terraform-provider-aws/internal/service/appfabric"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppFabricIngestion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ingestion appfabric.GetIngestionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_ingestion.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					resource.TestCheckResourceAttrSet(resourceName, "app"),
					resource.TestCheckResourceAttrSet(resourceName, "app_bundle_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "ingestion_type"),
					resource.TestCheckResourceAttrSet(resourceName, "tenant_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAppFabricIngestion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ingestion appfabric.GetIngestionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_appfabric_ingestion.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.AppFabricServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIngestionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIngestionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIngestionExists(ctx, resourceName, &ingestion),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfappfabric.ResourceIngestion, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIngestionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appfabric_ingestion" {
				continue
			}
			_, err := conn.GetIngestion(ctx, &appfabric.GetIngestionInput{
				AppBundleIdentifier: aws.String(rs.Primary.Attributes["app_bundle_identifier"]),
				IngestionIdentifier: aws.String(rs.Primary.Attributes[names.AttrARN]),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.AppFabric, create.ErrActionCheckingDestroyed, tfappfabric.ResNameIngestion, rs.Primary.ID, err)
			}
			return create.Error(names.AppFabric, create.ErrActionCheckingDestroyed, tfappfabric.ResNameIngestion, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckIngestionExists(ctx context.Context, name string, ingestion *appfabric.GetIngestionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameIngestion, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameIngestion, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)
		resp, err := conn.GetIngestion(ctx, &appfabric.GetIngestionInput{
			AppBundleIdentifier: aws.String(rs.Primary.Attributes["app_bundle_identifier"]),
			IngestionIdentifier: aws.String(rs.Primary.Attributes[names.AttrARN]),
		})

		if err != nil {
			return create.Error(names.AppFabric, create.ErrActionCheckingExistence, tfappfabric.ResNameIngestion, rs.Primary.ID, err)
		}

		*ingestion = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AppFabricClient(ctx)

	input := &appfabric.ListAppBundlesInput{}
	_, err := conn.ListAppBundles(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccIngestionConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appfabric_ingestion" "test" {
  app                   = "OKTA"
  app_bundle_identifier = aws_appfabric_app_bundle.arn
  tenant_id             = "test-tenant-id"
  ingestion_type        = "auditLog"
  tags = {
    Name = "AppFabricTesting"
  }
}
`)
}
