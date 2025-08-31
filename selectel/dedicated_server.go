package selectel

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/ddaas"
)

func getDedicatedServerClient(d *schema.ResourceData, meta interface{}) (*ddaas.API, diag.Diagnostics) {
	config := meta.(*Config)
	projectID := d.Get("project_id").(string)

	selvpcClient, err := config.GetSelVPCClientWithProjectScope(projectID)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get project-scope selvpc client for ddaas: %w", err))
	}

	region := d.Get("region").(string)
	err = validateRegion(selvpcClient, DedicatedServer, region)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't validate region: %w", err))
	}

	endpoint, err := selvpcClient.Catalog.GetEndpoint(DedicatedServer, region)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't get endpoint to init ddaas client: %w", err))
	}

	client, err := ddaas.New(selvpcClient.GetXAuthToken(), endpoint.URL)
	if err != nil {
		return nil, diag.FromErr(fmt.Errorf("can't create ddaas client: %w", err))
	}
	return client, nil
}
