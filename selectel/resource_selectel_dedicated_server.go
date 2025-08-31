package selectel

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-selectel/selectel/ddaas"
)

func resourceDedicatedServerV1() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDedicatedServerV1Create,
		ReadContext:   resourceDedicatedServerV1Read,
		UpdateContext: resourceDedicatedServerV1Update,
		DeleteContext: resourceDedicatedServerV1Delete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDedicatedServerV1ImportState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		Schema: resourceDedicatedServerV1Schema(),
	}
}

func resourceDedicatedServerV1Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	// Валидация проекта
	projectID := d.Get("project_id").(string)
	if err := validateProjectAccess(ctx, client, projectID); err != nil {
		return diag.FromErr(fmt.Errorf("project validation failed: %w", err))
	}

	// Получение параметров
	locationUUID := d.Get("location_uuid").(string)
	configurationUUID := d.Get("configuration_uuid").(string)
	tariffUUID := d.Get("tariff_uuid").(string)
	osImageUUID := d.Get("os_image_uuid").(string)

	// Валидация локации
	if err := validateLocation(ctx, client, locationUUID); err != nil {
		return diag.FromErr(fmt.Errorf("location validation failed: %w", err))
	}

	// Валидация конфигурации в локации
	if err := validateConfigurationInLocation(ctx, client, configurationUUID, locationUUID); err != nil {
		return diag.FromErr(fmt.Errorf("configuration validation failed: %w", err))
	}

	// Валидация тарифа для конфигурации
	if err := validateTariffForConfiguration(ctx, client, tariffUUID, configurationUUID); err != nil {
		return diag.FromErr(fmt.Errorf("tariff validation failed: %w", err))
	}

	// Валидация образа ОС
	if err := validateOSImageForConfiguration(ctx, client, osImageUUID, locationUUID, configurationUUID); err != nil {
		return diag.FromErr(fmt.Errorf("OS image validation failed: %w", err))
	}

	// Генерация имени сервера если не указано
	serverName := d.Get("name").(string)
	if serverName == "" {
		serverName = generateServerName()
	} else {
		// Проверка уникальности имени в проекте
		if err := validateServerNameUniqueness(ctx, client, projectID, serverName); err != nil {
			return diag.FromErr(fmt.Errorf("server name validation failed: %w", err))
		}
	}

	// Валидация приватной сети (если указана)
	if privateNetworkUUID := d.Get("private_network_uuid").(string); privateNetworkUUID != "" {
		if err := validatePrivateNetworkForConfiguration(ctx, client, privateNetworkUUID, configurationUUID, locationUUID); err != nil {
			return diag.FromErr(fmt.Errorf("private network validation failed: %w", err))
		}
	}

	// Валидация публичной сети (если указана)
	if publicNetworkUUID := d.Get("public_network_uuid").(string); publicNetworkUUID != "" {
		if err := validatePublicNetworkForLocation(ctx, client, publicNetworkUUID, locationUUID); err != nil {
			return diag.FromErr(fmt.Errorf("public network validation failed: %w", err))
		}
	}

	// Обработка параметров ОС
	osParams, diagErr := processOSParams(ctx, client, d, osImageUUID, configurationUUID)
	if diagErr != nil {
		return diagErr
	}

	dedicatedServerCreateOpts := ddaas.DedicatedServerCreateOpts{
		ProjectID:          projectID,
		LocationUUID:       locationUUID,
		ConfigurationUUID:  configurationUUID,
		TariffUUID:         tariffUUID,
		OSImageUUID:        osImageUUID,
		Name:               serverName,
		PublicNetworkUUID:  d.Get("public_network_uuid").(string),
		PrivateNetworkUUID: d.Get("private_network_uuid").(string),
		OsParams:           osParams,
	}

	log.Printf("[DEBUG] Creating dedicated server with options: %+v", dedicatedServerCreateOpts)

	server, err := client.CreateDedicatedServer(ctx, dedicatedServerCreateOpts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating dedicated server: %w", err))
	}

	d.SetId(server.UUID)
	d.Set("name", serverName)

	// Ожидание готовности сервера
	if err := client.WaitForServerStatus(ctx, server.UUID, ddaas.StatusActive, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(fmt.Errorf("server creation timeout: %w", err))
	}

	return resourceDedicatedServerV1Read(ctx, d, meta)
}

func resourceDedicatedServerV1Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	serverUUID := d.Id()
	log.Printf("[DEBUG] Reading dedicated server %s", serverUUID)

	server, err := client.DedicatedServer(ctx, serverUUID)
	if err != nil {
		// Если сервер не найден, удаляем из состояния
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[WARN] Dedicated server %s not found, removing from state", serverUUID)
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf("error reading dedicated server %s: %w", serverUUID, err))
	}

	// Установка атрибутов
	d.Set("uuid", server.UUID)
	d.Set("name", server.Name)
	d.Set("status", string(server.Status))
	d.Set("project_id", server.ProjectID)
	d.Set("location_uuid", server.LocationUUID)
	d.Set("service_uuid", server.ServiceUUID)
	d.Set("configuration_uuid", server.ConfigurationUUID)
	d.Set("tariff_uuid", server.TariffUUID)
	d.Set("os_image_uuid", server.OSImageUUID)
	d.Set("created_at", server.CreatedAt.Format(time.RFC3339))
	d.Set("updated_at", server.UpdatedAt.Format(time.RFC3339))

	// Установка IP адресов
	if len(server.IPAddresses) > 0 {
		ipAddresses := make([]map[string]interface{}, len(server.IPAddresses))
		for i, ip := range server.IPAddresses {
			ipAddresses[i] = map[string]interface{}{
				"type":    ip.Type,
				"ip":      ip.IP,
				"netmask": ip.Netmask,
				"gateway": ip.Gateway,
			}
		}
		d.Set("ip_addresses", ipAddresses)
	}

	return nil
}

func resourceDedicatedServerV1Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	serverUUID := d.Id()

	// Обновление возможно только для os_image_uuid (переустановка ОС)
	if d.HasChange("os_image_uuid") {
		return reinstallServerOS(ctx, d, meta, client, serverUUID)
	}

	// Для остальных параметров просто читаем состояние
	return resourceDedicatedServerV1Read(ctx, d, meta)
}

func resourceDedicatedServerV1Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, diagErr := getDedicatedServerClient(d, meta)
	if diagErr != nil {
		return diagErr
	}

	serverUUID := d.Id()
	log.Printf("[DEBUG] Deleting dedicated server %s", serverUUID)

	// Удаление сервера
	err := client.DeleteDedicatedServer(ctx, serverUUID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.Printf("[WARN] Dedicated server %s not found during deletion", serverUUID)
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting dedicated server %s: %w", serverUUID, err))
	}

	// Ожидание удаления
	if err := waitForServerDeleted(ctx, client, serverUUID, d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(fmt.Errorf("server deletion timeout: %w", err))
	}

	return nil
}

func resourceDedicatedServerV1ImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	config := meta.(*Config)
	if config.ProjectID == "" {
		return nil, fmt.Errorf("SELECTEL_PROJECT_ID must be set for the resource import")
	}

	d.Set("project_id", config.ProjectID)
	return []*schema.ResourceData{d}, nil
}

// Функции валидации
func validateProjectAccess(ctx context.Context, client *ddaas.API, projectID string) error {
	// Проверяем доступ к проекту через получение списка серверов
	_, err := client.DedicatedServers(ctx, &ddaas.DedicatedServerQueryParams{
		ProjectID: projectID,
	})
	if err != nil {
		return fmt.Errorf("access denied to project %s: %w", projectID, err)
	}
	return nil
}

func validateLocation(ctx context.Context, client *ddaas.API, locationUUID string) error {
	_, err := client.Location(ctx, locationUUID)
	if err != nil {
		return fmt.Errorf("location %s not found or unavailable: %w", locationUUID, err)
	}
	return nil
}

func validateConfigurationInLocation(ctx context.Context, client *ddaas.API, configUUID, locationUUID string) error {
	configurations, err := client.Configurations(ctx, locationUUID)
	if err != nil {
		return fmt.Errorf("unable to get configurations for location %s: %w", locationUUID, err)
	}

	for _, config := range configurations {
		if config.UUID == configUUID {
			return nil
		}
	}

	return fmt.Errorf("configuration %s not available in location %s", configUUID, locationUUID)
}

func validateTariffForConfiguration(ctx context.Context, client *ddaas.API, tariffUUID, configUUID string) error {
	tariffs, err := client.Tariffs(ctx, configUUID)
	if err != nil {
		return fmt.Errorf("unable to get tariffs for configuration %s: %w", configUUID, err)
	}

	for _, tariff := range tariffs {
		if tariff.UUID == tariffUUID {
			return nil
		}
	}

	return fmt.Errorf("tariff %s not available for configuration %s", tariffUUID, configUUID)
}

func validateOSImageForConfiguration(ctx context.Context, client *ddaas.API, osImageUUID, locationUUID, configUUID string) error {
	images, err := client.OSImages(ctx, locationUUID, configUUID)
	if err != nil {
		return fmt.Errorf("unable to get OS images for location %s and configuration %s: %w", locationUUID, configUUID, err)
	}

	for _, image := range images {
		if image.UUID == osImageUUID {
			return nil
		}
	}

	return fmt.Errorf("OS image %s not available for configuration %s in location %s", osImageUUID, configUUID, locationUUID)
}

func validateServerNameUniqueness(ctx context.Context, client *ddaas.API, projectID, name string) error {
	servers, err := client.DedicatedServers(ctx, &ddaas.DedicatedServerQueryParams{
		ProjectID: projectID,
		Name:      name,
	})
	if err != nil {
		return fmt.Errorf("unable to check server name uniqueness: %w", err)
	}

	if len(servers) > 0 {
		return fmt.Errorf("server with name '%s' already exists in project %s", name, projectID)
	}

	return nil
}

func validatePrivateNetworkForConfiguration(ctx context.Context, client *ddaas.API, networkUUID, configUUID, locationUUID string) error {
	network, err := client.Network(ctx, networkUUID)
	if err != nil {
		return fmt.Errorf("private network %s not found: %w", networkUUID, err)
	}

	if network.Type != "private" {
		return fmt.Errorf("network %s is not a private network", networkUUID)
	}

	if network.LocationUUID != locationUUID {
		return fmt.Errorf("private network %s is not available in location %s", networkUUID, locationUUID)
	}

	// Проверка поддержки приватной сети конфигурацией
	config, err := client.Configuration(ctx, configUUID)
	if err != nil {
		return fmt.Errorf("unable to get configuration %s: %w", configUUID, err)
	}

	// Некоторые конфигурации (например, Chipcore Line) не поддерживают приватные сети
	if strings.Contains(strings.ToLower(config.TariffLine), "chipcore") {
		return fmt.Errorf("configuration %s (%s) does not support private networks", configUUID, config.Name)
	}

	return nil
}

func validatePublicNetworkForLocation(ctx context.Context, client *ddaas.API, networkUUID, locationUUID string) error {
	network, err := client.Network(ctx, networkUUID)
	if err != nil {
		return fmt.Errorf("public network %s not found: %w", networkUUID, err)
	}

	if network.Type != "public" {
		return fmt.Errorf("network %s is not a public network", networkUUID)
	}

	if network.LocationUUID != locationUUID {
		return fmt.Errorf("public network %s is not available in location %s", networkUUID, locationUUID)
	}

	return nil
}

func generateServerName() string {
	return fmt.Sprintf("terraform-server-%d", time.Now().Unix())
}

func processOSParams(ctx context.Context, client *ddaas.API, d *schema.ResourceData, osImageUUID, configUUID string) (map[string]interface{}, diag.Diagnostics) {
	osParamsList := d.Get("os_params").([]interface{})
	if len(osParamsList) == 0 {
		// Если параметры не указаны, возвращаем пустую карту
		return map[string]interface{}{}, nil
	}

	osParams := osParamsList[0].(map[string]interface{})
	processedParams := make(map[string]interface{})

	// Получение информации об образе ОС для валидации
	image, err := client.OSImage(ctx, osImageUUID, "", "")
	if err != nil {
		// Если не можем получить образ, просто передаем параметры как есть
		log.Printf("[WARN] Unable to get OS image %s for validation: %v", osImageUUID, err)
	}

	// Обработка основных параметров
	for _, key := range []string{"login", "password", "user_data"} {
		if val, ok := osParams[key].(string); ok && val != "" {
			processedParams[key] = val
		}
	}

	// Обработка SSH ключа
	if sshKey, ok := osParams["ssh_key"].(string); ok && sshKey != "" {
		// Если указано имя ключа, получаем его значение из Selectel
		// TODO: реализовать получение SSH ключа по имени
		processedParams["ssh_key"] = sshKey
	}

	// Обработка soft RAID и партиций (только для Linux)
	if image != nil && strings.ToLower(image.Family) == "linux" {
		if softRaid, ok := osParams["soft_raid"].(string); ok && softRaid != "" {
			// Валидация soft RAID для конфигурации
			if err := validateSoftRaidForConfiguration(ctx, client, softRaid, configUUID); err != nil {
				return nil, diag.FromErr(fmt.Errorf("soft RAID validation failed: %w", err))
			}
			processedParams["soft_raid"] = softRaid

			// Валидация партиций с soft RAID
			if partitions, ok := osParams["partitions"].(string); ok && partitions != "" {
				if err := validatePartitionsWithRaid(partitions, softRaid); err != nil {
					return nil, diag.FromErr(fmt.Errorf("partitions validation failed: %w", err))
				}
				processedParams["partitions"] = partitions
			}
		} else {
			// Установка стандартной разбивки для Linux
			defaultPartitions, err := getDefaultPartitionsForConfiguration(ctx, client, configUUID)
			if err != nil {
				log.Printf("[WARN] Unable to get default partitions for configuration %s: %v", configUUID, err)
			} else {
				processedParams["partitions"] = defaultPartitions
			}
		}
	}

	return processedParams, nil
}

func validateSoftRaidForConfiguration(ctx context.Context, client *ddaas.API, softRaid, configUUID string) error {
	config, err := client.Configuration(ctx, configUUID)
	if err != nil {
		return fmt.Errorf("unable to get configuration %s: %w", configUUID, err)
	}

	// Проверка поддержки RAID для конфигурации
	// Большинство конфигураций поддерживают RAID, но некоторые могут не поддерживать
	if strings.Contains(strings.ToLower(config.TariffLine), "entry") && softRaid != "no_raid" {
		return fmt.Errorf("configuration %s (%s) supports only no_raid", configUUID, config.Name)
	}

	return nil
}

func validatePartitionsWithRaid(partitions, raid string) error {
	// Базовая валидация совместимости партиций с RAID
	if raid == "raid0" && strings.Contains(partitions, "swap") {
		return fmt.Errorf("swap partition is not recommended with RAID0")
	}
	return nil
}

func getDefaultPartitionsForConfiguration(ctx context.Context, client *ddaas.API, configUUID string) (string, error) {
	// Получение стандартной разбивки для конфигурации
	config, err := client.Configuration(ctx, configUUID)
	if err != nil {
		return "", err
	}

	// Базовая разбивка в зависимости от конфигурации
	if strings.Contains(strings.ToLower(config.TariffLine), "entry") {
		return "/=100%", nil
	}
	return "/=50%,/var=25%,/home=25%", nil
}

func reinstallServerOS(ctx context.Context, d *schema.ResourceData, meta interface{}, client *ddaas.API, serverUUID string) diag.Diagnostics {
	newOSImageUUID := d.Get("os_image_uuid").(string)

	log.Printf("[DEBUG] Reinstalling OS on server %s with image %s", serverUUID, newOSImageUUID)

	// Валидация нового образа ОС
	configUUID := d.Get("configuration_uuid").(string)
	locationUUID := d.Get("location_uuid").(string)
	if err := validateOSImageForConfiguration(ctx, client, newOSImageUUID, locationUUID, configUUID); err != nil {
		return diag.FromErr(fmt.Errorf("new OS image validation failed: %w", err))
	}

	// Обработка новых параметров ОС
	osParams, diagErr := processOSParams(ctx, client, d, newOSImageUUID, configUUID)
	if diagErr != nil {
		return diagErr
	}

	// Вызов API для переустановки ОС
	updateOpts := ddaas.DedicatedServerUpdateOpts{
		OSImageUUID: newOSImageUUID,
		OsParams:    osParams,
	}

	_, err := client.UpdateDedicatedServer(ctx, serverUUID, updateOpts)
	if err != nil {
		return diag.FromErr(fmt.Errorf("OS reinstall failed: %w", err))
	}

	// Ожидание завершения переустановки
	if err := client.WaitForServerStatus(ctx, serverUUID, ddaas.StatusActive, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.FromErr(fmt.Errorf("OS reinstall timeout: %w", err))
	}

	return resourceDedicatedServerV1Read(ctx, d, meta)
}

func waitForServerDeleted(ctx context.Context, client *ddaas.API, serverUUID string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return fmt.Errorf("timeout waiting for server %s to be deleted", serverUUID)
		case <-ticker.C:
			_, err := client.DedicatedServer(ctx, serverUUID)
			if err != nil {
				if strings.Contains(err.Error(), "not found") {
					return nil // Сервер удален
				}
				return fmt.Errorf("error checking server status: %w", err)
			}
		}
	}
}
