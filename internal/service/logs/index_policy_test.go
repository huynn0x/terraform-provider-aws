// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflogs "github.com/hashicorp/terraform-provider-aws/internal/service/logs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLogsIndexPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	logGroupName := "/aws/testacc/index-policy-" + sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDocument := `{Fields:[\"eventName\"]}`
	resourceName := "aws_cloudwatch_log_index_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(logGroupName, policyDocument),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName),
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

func TestAccLogsIndexPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	logGroupName := "/aws/testacc/index-policy-" + sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyDocument := `{Fields:[\"eventName\"]}`
	resourceName := "aws_cloudwatch_log_index_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.CloudWatchEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LogsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIndexPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIndexPolicyConfig_basic(logGroupName, policyDocument),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIndexPolicyExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflogs.ResourceIndexPolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckIndexPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_cloudwatch_log_index_policy" {
				continue
			}

			_, err := tflogs.FindIndexPolicyByLogGroupName(ctx, conn, rs.Primary.Attributes[names.AttrLogGroupName])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("CloudWatch Logs Index Policy still exists: %s", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckIndexPolicyExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LogsClient(ctx)

		_, err := tflogs.FindIndexPolicyByLogGroupName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccIndexPolicyConfig_basic(logGroupName string, policyDocument string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_index_policy" "test" {
  log_group_name  = aws_cloudwatch_log_group.test.name
  policy_document = jsonencode(%[2]q)
}
`, logGroupName, policyDocument)
}
