package aws

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/waiter"
)

func resourceAwsConnectLexBotAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsConnectLexBotAssociationCreate,
		ReadContext:   resourceAwsConnectLexBotAssociationRead,
		UpdateContext: resourceAwsConnectLexBotAssociationRead,
		DeleteContext: resourceAwsConnectLexBotAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				instanceID, name, region, err := resourceAwsConnectLexBotAssociationParseID(d.Id())
				if err != nil {
					return nil, err
				}

				d.Set("bot_name", name)
				d.Set("instance_id", instanceID)
				d.Set("lex_region", name)
				d.SetId(fmt.Sprintf("%s:%s:%s", instanceID, name, region))

				return []*schema.ResourceData{d}, nil
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(waiter.ConnectLexBotAssociationCreateTimeout),
			Delete: schema.DefaultTimeout(waiter.ConnectLexBotAssociationDeleteTimeout),
		},
		Schema: map[string]*schema.Schema{
			"bot_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(2, 50),
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			//Documentation is wrong, this is required.
			"lex_region": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsConnectLexBotAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	botAssociation := &connect.LexBot{
		Name:      aws.String(d.Get("bot_name").(string)),
		LexRegion: aws.String(d.Get("lex_region").(string)),
	}
	input := &connect.AssociateLexBotInput{
		InstanceId: aws.String(d.Get("instance_id").(string)),
		LexBot:     botAssociation,
	}

	lbaId := fmt.Sprintf("%s:%s:%s", d.Get("instance_id").(string), d.Get("bot_name").(string), d.Get("lex_region").(string))

	log.Printf("[DEBUG] Creating Connect Lex Bot Association %s", input)

	_, err := conn.AssociateLexBotWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Lex Bot Association (%s): %s", lbaId, err))
	}

	d.SetId(lbaId)
	return resourceAwsConnectLexBotAssociationRead(ctx, d, meta)
}

func resourceAwsConnectLexBotAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn
	instanceID := d.Get("instance_id")
	name := d.Get("bot_name")

	lexBot, err := finder.LexBotAssociationByName(ctx, conn, instanceID.(string), name.(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding LexBot Association by name (%s): %w", name, err))
	}

	if lexBot == nil {
		return diag.FromErr(fmt.Errorf("error finding LexBot Association by name (%s): not found", name))
	}

	d.Set("bot_name", lexBot.Name)
	d.Set("instance_id", instanceID)
	d.Set("lex_region", lexBot.LexRegion)
	d.SetId(fmt.Sprintf("%s:%s:%s", instanceID, d.Get("bot_name").(string), d.Get("lex_region").(string)))

	return nil
}

func resourceAwsConnectLexBotAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	instanceID, name, region, err := resourceAwsConnectLexBotAssociationParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &connect.DisassociateLexBotInput{
		InstanceId: aws.String(instanceID),
		BotName:    aws.String(name),
		LexRegion:  aws.String(region),
	}

	log.Printf("[DEBUG] Deleting Connect Lex Bot Association %s", d.Id())
	_, dissErr := conn.DisassociateLexBot(input)

	if dissErr != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Lex Bot Association (%s): %s", instanceID, err))
	}
	return nil
}

func resourceAwsConnectLexBotAssociationParseID(id string) (string, string, string, error) {
	parts := strings.SplitN(id, ":", 3)

	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("unexpected format of ID (%s), expected instanceID:name:region", id)
	}

	return parts[0], parts[1], parts[2], nil
}
