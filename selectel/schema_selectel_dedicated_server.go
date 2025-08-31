package selectel

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDedicatedServerV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project_id": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Project ID",
		},
		"location_uuid": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Location UUID where server will be deployed",
		},
		"configuration_uuid": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Server configuration UUID",
		},
		"tariff_uuid": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Tariff plan UUID",
		},
		"os_image_uuid": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "OS image UUID (changing this will reinstall the OS)",
		},

		// Опциональные параметры
		"name": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			Description:  "Server name (auto-generated if not specified)",
			ValidateFunc: validation.StringLenBetween(1, 255),
		},
		"public_network_uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Public network UUID",
		},
		"private_network_uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			ForceNew:    true,
			Description: "Private network UUID (available only for supported configurations)",
		},

		// Дополнительные параметры для ОС
		"os_params": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Additional OS parameters",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"login": {
						Type:         schema.TypeString,
						Optional:     true,
						Description:  "Login for Windows/Linux OS",
						ValidateFunc: validation.StringLenBetween(1, 64),
					},
					"password": {
						Type:         schema.TypeString,
						Optional:     true,
						Sensitive:    true,
						Description:  "Password for Windows/Linux OS",
						ValidateFunc: validation.StringLenBetween(8, 128),
					},
					"soft_raid": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Soft RAID configuration (Linux only)",
						ValidateFunc: validation.StringInSlice([]string{
							"raid0", "raid1", "raid5", "raid10", "no_raid",
						}, false),
					},
					"partitions": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Custom partitions configuration (Linux only)",
					},
					"ssh_key": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "SSH key name or content (Linux only)",
					},
					"user_data": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "User data script (Linux only)",
					},
				},
			},
		},

		// Вычисляемые атрибуты
		"uuid": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Server UUID",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Server status",
		},
		"service_uuid": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Service UUID",
		},
		"ip_addresses": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Server IP addresses",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "IP address type (public/private)",
					},
					"ip": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "IP address",
					},
					"netmask": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Network mask",
					},
					"gateway": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Gateway IP",
					},
				},
			},
		},
		"created_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Server creation timestamp",
		},
		"updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Server last update timestamp",
		},
	}
}

// Data Sources Schemas

func dataSourceDedicatedServerLocationV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Location UUID",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Location name",
		},
		"location_id": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Location ID",
		},
		"description": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Location description",
		},
		"enable": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Location availability",
		},
	}
}

func dataSourceDedicatedServerConfigurationV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Configuration UUID",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Configuration name",
		},
		"location_uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filter by location UUID",
		},
		"tariff_line": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Tariff line",
		},
		"model": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Server model",
		},
		"cpu": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "CPU specifications",
		},
		"ram": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "RAM specifications",
		},
		"storage": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Storage specifications",
		},
	}
}

func dataSourceDedicatedServerTariffV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Tariff UUID",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Tariff name",
		},
		"configuration_uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filter by configuration UUID",
		},
		"period": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Billing period",
		},
		"price": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Price",
		},
		"currency": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Currency",
		},
	}
}

func dataSourceDedicatedServerOSImageV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "OS image UUID",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "OS image name",
		},
		"location_uuid": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Location UUID",
		},
		"service_uuid": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Service UUID (configuration UUID)",
		},
		"family": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "OS family (linux, windows)",
		},
		"version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "OS version",
		},
		"architecture": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Architecture (x86_64, i386)",
		},
		"parameters": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "OS parameters",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func dataSourceDedicatedServerNetworkV1Schema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Network UUID",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Network name",
		},
		"location_uuid": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Filter by location UUID",
		},
		"type": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Network type (public, private)",
			ValidateFunc: validation.StringInSlice([]string{
				"public", "private",
			}, false),
		},
		"vlan": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "VLAN ID",
		},
	}
}
