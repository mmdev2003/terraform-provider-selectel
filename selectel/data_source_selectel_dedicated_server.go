package selectel

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/ddaas"
	"log"
)

// Data Source: Location
func dataSourceDedicatedServerLocationV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerLocationV1Read,
		Schema:      dataSourceDedicatedServerLocationV1Schema(),
	}
}

func dataSourceDedicatedServerLocationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading dedicated server locations")

	locations, err := client.Locations(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading locations: %w", err))
	}

	var targetLocation *ddaas.Location

	// Поиск по UUID или имени
	if uuid, ok := d.GetOk("uuid"); ok {
		for _, location := range locations {
			if location.UUID == uuid.(string) {
				targetLocation = &location
				break
			}
		}
	} else if name, ok := d.GetOk("name"); ok {
		for _, location := range locations {
			if location.Name == name.(string) {
				targetLocation = &location
				break
			}
		}
	} else {
		return diag.Errorf("either 'uuid' or 'name' must be specified")
	}

	if targetLocation == nil {
		return diag.Errorf("location not found")
	}

	d.SetId(targetLocation.UUID)
	d.Set("uuid", targetLocation.UUID)
	d.Set("name", targetLocation.Name)
	d.Set("location_id", targetLocation.LocationID)
	d.Set("description", targetLocation.Description)
	d.Set("enable", targetLocation.Enable)

	return nil
}

// Data Source: Configuration
func dataSourceDedicatedServerConfigurationV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerConfigurationV1Read,
		Schema:      dataSourceDedicatedServerConfigurationV1Schema(),
	}
}

func dataSourceDedicatedServerConfigurationV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading dedicated server configurations")

	locationUUID := d.Get("location_uuid").(string)
	configurations, err := client.Configurations(ctx, locationUUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading configurations: %w", err))
	}

	var targetConfiguration *ddaas.Configuration

	// Поиск по UUID или имени
	if uuid, ok := d.GetOk("uuid"); ok {
		for _, config := range configurations {
			if config.UUID == uuid.(string) {
				targetConfiguration = &config
				break
			}
		}
	} else if name, ok := d.GetOk("name"); ok {
		for _, config := range configurations {
			if config.Name == name.(string) {
				targetConfiguration = &config
				break
			}
		}
	} else {
		return diag.Errorf("either 'uuid' or 'name' must be specified")
	}

	if targetConfiguration == nil {
		return diag.Errorf("configuration not found")
	}

	d.SetId(targetConfiguration.UUID)
	d.Set("uuid", targetConfiguration.UUID)
	d.Set("name", targetConfiguration.Name)
	d.Set("tariff_line", targetConfiguration.TariffLine)
	d.Set("model", targetConfiguration.Model)
	d.Set("cpu", targetConfiguration.CPU)
	d.Set("ram", targetConfiguration.RAM)
	d.Set("storage", targetConfiguration.Storage)

	return nil
}

// Data Source: Tariff
func dataSourceDedicatedServerTariffV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerTariffV1Read,
		Schema:      dataSourceDedicatedServerTariffV1Schema(),
	}
}

func dataSourceDedicatedServerTariffV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading dedicated server tariffs")

	configUUID := d.Get("configuration_uuid").(string)
	tariffs, err := client.Tariffs(ctx, configUUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading tariffs: %w", err))
	}

	var targetTariff *ddaas.Tariff

	// Поиск по UUID или имени
	if uuid, ok := d.GetOk("uuid"); ok {
		for _, tariff := range tariffs {
			if tariff.UUID == uuid.(string) {
				targetTariff = &tariff
				break
			}
		}
	} else if name, ok := d.GetOk("name"); ok {
		for _, tariff := range tariffs {
			if tariff.Name == name.(string) {
				targetTariff = &tariff
				break
			}
		}
	} else {
		return diag.Errorf("either 'uuid' or 'name' must be specified")
	}

	if targetTariff == nil {
		return diag.Errorf("tariff not found")
	}

	d.SetId(targetTariff.UUID)
	d.Set("uuid", targetTariff.UUID)
	d.Set("name", targetTariff.Name)
	d.Set("period", targetTariff.Period)
	d.Set("price", targetTariff.Price)
	d.Set("currency", targetTariff.Currency)

	return nil
}

// Data Source: OS Image
func dataSourceDedicatedServerOSImageV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerOSImageV1Read,
		Schema:      dataSourceDedicatedServerOSImageV1Schema(),
	}
}

func dataSourceDedicatedServerOSImageV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading dedicated server OS images")

	locationUUID := d.Get("location_uuid").(string)
	serviceUUID := d.Get("service_uuid").(string)

	images, err := client.OSImages(ctx, locationUUID, serviceUUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading OS images: %w", err))
	}

	var targetImage *ddaas.OSImage

	// Поиск по UUID, имени или семейству
	if uuid, ok := d.GetOk("uuid"); ok {
		for _, image := range images {
			if image.UUID == uuid.(string) {
				targetImage = &image
				break
			}
		}
	} else if name, ok := d.GetOk("name"); ok {
		for _, image := range images {
			if image.Name == name.(string) {
				targetImage = &image
				break
			}
		}
	} else if family, ok := d.GetOk("family"); ok {
		// Найти первый образ указанного семейства
		for _, image := range images {
			if image.Family == family.(string) {
				targetImage = &image
				break
			}
		}
	} else {
		return diag.Errorf("one of 'uuid', 'name', or 'family' must be specified")
	}

	if targetImage == nil {
		return diag.Errorf("OS image not found")
	}

	d.SetId(targetImage.UUID)
	d.Set("uuid", targetImage.UUID)
	d.Set("name", targetImage.Name)
	d.Set("family", targetImage.Family)
	d.Set("version", targetImage.Version)
	d.Set("architecture", targetImage.Architecture)
	d.Set("parameters", targetImage.Parameters)

	return nil
}

// Data Source: Network
func dataSourceDedicatedServerNetworkV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerNetworkV1Read,
		Schema:      dataSourceDedicatedServerNetworkV1Schema(),
	}
}

func dataSourceDedicatedServerNetworkV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading dedicated server networks")

	locationUUID := d.Get("location_uuid").(string)
	networks, err := client.Networks(ctx, locationUUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading networks: %w", err))
	}

	var targetNetwork *ddaas.Network

	// Поиск по UUID, имени или типу
	if uuid, ok := d.GetOk("uuid"); ok {
		for _, network := range networks {
			if network.UUID == uuid.(string) {
				targetNetwork = &network
				break
			}
		}
	} else if name, ok := d.GetOk("name"); ok {
		for _, network := range networks {
			if network.Name == name.(string) {
				targetNetwork = &network
				break
			}
		}
	} else if networkType, ok := d.GetOk("type"); ok {
		// Найти первую сеть указанного типа
		for _, network := range networks {
			if network.Type == networkType.(string) {
				targetNetwork = &network
				break
			}
		}
	} else {
		return diag.Errorf("one of 'uuid', 'name', or 'type' must be specified")
	}

	if targetNetwork == nil {
		return diag.Errorf("network not found")
	}

	d.SetId(targetNetwork.UUID)
	d.Set("uuid", targetNetwork.UUID)
	d.Set("name", targetNetwork.Name)
	d.Set("type", targetNetwork.Type)
	d.Set("vlan", targetNetwork.VLAN)

	return nil
}

// Data Source: Multiple Locations
func dataSourceDedicatedServerLocationsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerLocationsV1Read,
		Schema: map[string]*schema.Schema{
			"locations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: dataSourceDedicatedServerLocationV1Schema(),
				},
			},
		},
	}
}

func dataSourceDedicatedServerLocationsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading all dedicated server locations")

	locations, err := client.Locations(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading locations: %w", err))
	}

	locationsList := make([]map[string]interface{}, len(locations))
	for i, location := range locations {
		locationsList[i] = map[string]interface{}{
			"uuid":        location.UUID,
			"name":        location.Name,
			"location_id": location.LocationID,
			"description": location.Description,
			"enable":      location.Enable,
		}
	}

	d.SetId("locations")
	d.Set("locations", locationsList)

	return nil
}

// Data Source: Multiple Configurations
func dataSourceDedicatedServerConfigurationsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerConfigurationsV1Read,
		Schema: map[string]*schema.Schema{
			"location_uuid": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by location UUID",
			},
			"configurations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: dataSourceDedicatedServerConfigurationV1Schema(),
				},
			},
		},
	}
}

func dataSourceDedicatedServerConfigurationsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading all dedicated server configurations")

	locationUUID := d.Get("location_uuid").(string)
	configurations, err := client.Configurations(ctx, locationUUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading configurations: %w", err))
	}

	configurationsList := make([]map[string]interface{}, len(configurations))
	for i, config := range configurations {
		configurationsList[i] = map[string]interface{}{
			"uuid":        config.UUID,
			"name":        config.Name,
			"tariff_line": config.TariffLine,
			"model":       config.Model,
			"cpu":         config.CPU,
			"ram":         config.RAM,
			"storage":     config.Storage,
		}
	}

	d.SetId("configurations")
	d.Set("configurations", configurationsList)

	return nil
}

// Data Source: Multiple Tariffs
func dataSourceDedicatedServerTariffsV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerTariffsV1Read,
		Schema: map[string]*schema.Schema{
			"configuration_uuid": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by configuration UUID",
			},
			"tariffs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: dataSourceDedicatedServerTariffV1Schema(),
				},
			},
		},
	}
}

func dataSourceDedicatedServerTariffsV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading all dedicated server tariffs")

	configUUID := d.Get("configuration_uuid").(string)
	tariffs, err := client.Tariffs(ctx, configUUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading tariffs: %w", err))
	}

	tariffsList := make([]map[string]interface{}, len(tariffs))
	for i, tariff := range tariffs {
		tariffsList[i] = map[string]interface{}{
			"uuid":     tariff.UUID,
			"name":     tariff.Name,
			"period":   tariff.Period,
			"price":    tariff.Price,
			"currency": tariff.Currency,
		}
	}

	d.SetId("tariffs")
	d.Set("tariffs", tariffsList)

	return nil
}

// Data Source: Multiple OS Images
func dataSourceDedicatedServerOSImagesV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerOSImagesV1Read,
		Schema: map[string]*schema.Schema{
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
				Description: "Filter by OS family",
			},
			"os_images": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: dataSourceDedicatedServerOSImageV1Schema(),
				},
			},
		},
	}
}

func dataSourceDedicatedServerOSImagesV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading all dedicated server OS images")

	locationUUID := d.Get("location_uuid").(string)
	serviceUUID := d.Get("service_uuid").(string)
	family := d.Get("family").(string)

	images, err := client.OSImages(ctx, locationUUID, serviceUUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading OS images: %w", err))
	}

	// Фильтрация по семейству ОС если указано
	var filteredImages []ddaas.OSImage
	if family != "" {
		for _, image := range images {
			if image.Family == family {
				filteredImages = append(filteredImages, image)
			}
		}
	} else {
		filteredImages = images
	}

	imagesList := make([]map[string]interface{}, len(filteredImages))
	for i, image := range filteredImages {
		imagesList[i] = map[string]interface{}{
			"uuid":         image.UUID,
			"name":         image.Name,
			"family":       image.Family,
			"version":      image.Version,
			"architecture": image.Architecture,
			"parameters":   image.Parameters,
		}
	}

	d.SetId("os_images")
	d.Set("os_images", imagesList)

	return nil
}

// Data Source: Multiple Networks
func dataSourceDedicatedServerNetworksV1() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDedicatedServerNetworksV1Read,
		Schema: map[string]*schema.Schema{
			"location_uuid": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by location UUID",
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by network type",
			},
			"networks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: dataSourceDedicatedServerNetworkV1Schema(),
				},
			},
		},
	}
}

func dataSourceDedicatedServerNetworksV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	log.Printf("[DEBUG] Reading all dedicated server networks")

	locationUUID := d.Get("location_uuid").(string)
	networkType := d.Get("type").(string)

	networks, err := client.Networks(ctx, locationUUID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading networks: %w", err))
	}

	// Фильтрация по типу сети если указано
	var filteredNetworks []ddaas.Network
	if networkType != "" {
		for _, network := range networks {
			if network.Type == networkType {
				filteredNetworks = append(filteredNetworks, network)
			}
		}
	} else {
		filteredNetworks = networks
	}

	networksList := make([]map[string]interface{}, len(filteredNetworks))
	for i, network := range filteredNetworks {
		networksList[i] = map[string]interface{}{
			"uuid": network.UUID,
			"name": network.Name,
			"type": network.Type,
			"vlan": network.VLAN,
		}
	}

	d.SetId("networks")
	d.Set("networks", networksList)

	return nil
}
