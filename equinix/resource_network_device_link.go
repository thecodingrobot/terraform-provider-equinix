package equinix

import (
	"context"
	"fmt"
	"time"

	"github.com/equinix/ne-go"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var networkDeviceLinkSchemaNames = map[string]string{
	"UUID":    "uuid",
	"Name":    "name",
	"Subnet":  "subnet",
	"Devices": "device",
	"Links":   "link",
	"Status":  "status",
}

var networkDeviceLinkDescriptions = map[string]string{
	"UUID":    "Device link unique identifier",
	"Name":    "Device link name",
	"Subnet":  "Device link subnet CIDR",
	"Devices": "Definition of one or more devices belonging to the device link",
	"Links":   "Definition of one or more, inter metro connections belonging to the device link",
	"Status":  "Device link provisioning status",
}

var networkDeviceLinkDeviceSchemaNames = map[string]string{
	"DeviceID":    "id",
	"ASN":         "asn",
	"InterfaceID": "interface_id",
	"Status":      "status",
	"IPAddress":   "ip_address",
}

var networkDeviceLinkDeviceDescriptions = map[string]string{
	"DeviceID":    "Device identifier",
	"ASN":         "Device ASN number",
	"InterfaceID": "Device network interface identifier to use for device link connection",
	"Status":      "Device link connection provisioning status",
	"IPAddress":   "Assigned IP address from device link subnet",
}

var networkDeviceLinkConnectionSchemaNames = map[string]string{
	"AccountNumber":        "account_number",
	"Throughput":           "throughput",
	"ThroughputUnit":       "throughput_unit",
	"SourceMetroCode":      "src_metro_code",
	"DestinationMetroCode": "dst_metro_code",
	"SourceZoneCode":       "src_zone_code",
	"DestinationZoneCode":  "dst_zone_code",
}

var networkDeviceLinkConnectionDescriptions = map[string]string{
	"AccountNumber":        "Billing account number to be used for connection charges",
	"Throughput":           "Connection throughput",
	"ThroughputUnit":       "Connection throughput unit",
	"SourceMetroCode":      "Connection source metro code",
	"DestinationMetroCode": "Connection destination metro code",
	"SourceZoneCode":       "Connection source zone code",
	"DestinationZoneCode":  "Connection destination zone code",
}

func resourceNetworkDeviceLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkDeviceLinkCreate,
		ReadContext:   resourceNetworkDeviceLinkRead,
		UpdateContext: resourceNetworkDeviceLinkUpdate,
		DeleteContext: resourceNetworkDeviceLinkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: createNetworkDeviceLinkResourceSchema(),
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
		Description: "Resource allows creation and management of Equinix Network Edge device links",
	}
}

func createNetworkDeviceLinkResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		networkDeviceLinkSchemaNames["UUID"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: networkDeviceLinkDescriptions["UUID"],
		},
		networkDeviceLinkSchemaNames["Status"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: networkDeviceLinkDescriptions["Status"],
		},
		networkDeviceLinkSchemaNames["Name"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringLenBetween(3, 50),
			Description:  networkDeviceLinkSchemaNames["Name"],
		},
		networkDeviceLinkSchemaNames["Subnet"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.IsCIDR,
			Description:  networkDeviceLinkSchemaNames["Subnet"],
		},
		networkDeviceLinkSchemaNames["Devices"]: {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 2,
			Elem: &schema.Resource{
				Schema: createNetworkDeviceLinkDeviceResourceSchema(),
			},
			Set:         networkDeviceLinkDeviceHash,
			Description: networkDeviceLinkSchemaNames["Device"],
		},
		networkDeviceLinkSchemaNames["Links"]: {
			Type:     schema.TypeSet,
			Optional: true,
			Elem: &schema.Resource{
				Schema: createNetworkDeviceLinkConnectionResourceSchema(),
			},
			Set:         networkDeviceLinkConnectionHash,
			Description: networkDeviceLinkSchemaNames["Links"],
		},
	}
}

func createNetworkDeviceLinkDeviceResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		networkDeviceLinkDeviceSchemaNames["DeviceID"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  networkDeviceLinkDeviceDescriptions["DeviceID"],
		},
		networkDeviceLinkDeviceSchemaNames["ASN"]: {
			Type:         schema.TypeInt,
			Required:     true,
			ValidateFunc: validation.IntAtLeast(1),
			Description:  networkDeviceLinkDeviceDescriptions["ASN"],
		},
		networkDeviceLinkDeviceSchemaNames["InterfaceID"]: {
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntAtLeast(1),
			Description:  networkDeviceLinkDeviceDescriptions["InterfaceID"],
		},
		networkDeviceLinkDeviceSchemaNames["Status"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: networkDeviceLinkDeviceDescriptions["Status"],
		},
		networkDeviceLinkDeviceSchemaNames["IPAddress"]: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: networkDeviceLinkDeviceDescriptions["IPAddress"],
		},
	}
}

func createNetworkDeviceLinkConnectionResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		networkDeviceLinkConnectionSchemaNames["AccountNumber"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  networkDeviceLinkConnectionDescriptions["AccountNumber"],
		},
		networkDeviceLinkConnectionSchemaNames["Throughput"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  networkDeviceLinkConnectionDescriptions["Throughput"],
		},
		networkDeviceLinkConnectionSchemaNames["ThroughputUnit"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice([]string{"Mbps", "Gbps"}, false),
			Description:  networkDeviceLinkConnectionDescriptions["ThroughputUnit"],
		},
		networkDeviceLinkConnectionSchemaNames["SourceMetroCode"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: stringIsMetroCode(),
			Description:  networkDeviceLinkConnectionDescriptions["SourceMetroCode"],
		},
		networkDeviceLinkConnectionSchemaNames["DestinationMetroCode"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: stringIsMetroCode(),
			Description:  networkDeviceLinkConnectionDescriptions["DestinationMetroCode"],
		},
		networkDeviceLinkConnectionSchemaNames["SourceZoneCode"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  networkDeviceLinkConnectionDescriptions["SourceZoneCode"],
		},
		networkDeviceLinkConnectionSchemaNames["DestinationZoneCode"]: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringIsNotEmpty,
			Description:  networkDeviceLinkConnectionDescriptions["DestinationZoneCode"],
		},
	}
}

func resourceNetworkDeviceLinkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	link := createNetworkDeviceLink(d)
	uuid, err := conf.ne.CreateDeviceLinkGroup(link)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(ne.StringValue(uuid))
	if _, err := createDeviceLinkStatusProvisioningWaitConfiguration(conf.ne.GetDeviceLinkGroup, d.Id(), 2*time.Second, d.Timeout(schema.TimeoutCreate)).WaitForStateContext(ctx); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Failed to wait for device link to become provisioned",
			Detail:        err.Error(),
			AttributePath: cty.GetAttrPath(networkDeviceLinkSchemaNames["Status"]),
		})
	}
	diags = append(diags, resourceNetworkDeviceLinkRead(ctx, d, m)...)
	return diags
}

func resourceNetworkDeviceLinkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	link, err := conf.ne.GetDeviceLinkGroup(d.Id())
	if err != nil {
		if isRestNotFoundError(err) {
			d.SetId("")
			return nil
		}
	}
	if err := updateNetworkDeviceLinkResource(link, d); err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceNetworkDeviceLinkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	changes := getResourceDataChangedKeys([]string{
		networkDeviceLinkSchemaNames["Name"], networkDeviceLinkSchemaNames["Subnet"],
		networkDeviceLinkSchemaNames["Devices"], networkDeviceLinkSchemaNames["Links"],
	}, d)
	updateReq := conf.ne.NewDeviceLinkGroupUpdateRequest(d.Id())
	for change, changeValue := range changes {
		switch change {
		case networkDeviceLinkSchemaNames["Name"]:
			updateReq.WithGroupName(changeValue.(string))
		case networkDeviceLinkSchemaNames["Subnet"]:
			updateReq.WithSubnet(changeValue.(string))
		case networkDeviceLinkSchemaNames["Devices"]:
			deviceList := expandNetworkDeviceLinkDevices(changeValue.(*schema.Set))
			updateReq.WithDevices(deviceList)
		case networkDeviceLinkSchemaNames["Links"]:
			connectionList := expandNetworkDeviceLinkConnections(changeValue.(*schema.Set))
			updateReq.WithLinks(connectionList)
		}
	}
	if err := updateReq.Execute(); err != nil {
		return diag.FromErr(err)
	}
	if _, err := createDeviceLinkStatusProvisioningWaitConfiguration(conf.ne.GetDeviceLinkGroup, d.Id(), 2*time.Second, d.Timeout(schema.TimeoutCreate)).WaitForStateContext(ctx); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Failed to wait for device link to become provisioned",
			Detail:        err.Error(),
			AttributePath: cty.GetAttrPath(networkDeviceLinkSchemaNames["Status"]),
		})
	}
	diags = append(diags, resourceNetworkDeviceLinkRead(ctx, d, m)...)
	return diags
}

func resourceNetworkDeviceLinkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	conf := m.(*Config)
	var diags diag.Diagnostics
	if err := conf.ne.DeleteDeviceLinkGroup(d.Id()); err != nil {
		if isRestNotFoundError(err) {
			return nil
		}
		return diag.FromErr(err)
	}
	if _, err := createDeviceLinkStatusDeleteWaitConfiguration(conf.ne.GetDeviceLinkGroup, d.Id(), 2*time.Second, d.Timeout(schema.TimeoutDelete)).WaitForStateContext(ctx); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       "Failed to wait for device link to become deprovisioned",
			Detail:        err.Error(),
			AttributePath: cty.GetAttrPath(networkDeviceLinkSchemaNames["Status"]),
		})
	}
	return diags
}

func createNetworkDeviceLink(d *schema.ResourceData) ne.DeviceLinkGroup {
	link := ne.DeviceLinkGroup{}
	if v, ok := d.GetOk(networkDeviceLinkSchemaNames["Name"]); ok {
		link.Name = ne.String(v.(string))
	}
	if v, ok := d.GetOk(networkDeviceLinkSchemaNames["Subnet"]); ok {
		link.Subnet = ne.String(v.(string))
	}
	if v, ok := d.GetOk(networkDeviceLinkSchemaNames["Devices"]); ok {
		link.Devices = expandNetworkDeviceLinkDevices(v.(*schema.Set))
	}
	if v, ok := d.GetOk(networkDeviceLinkSchemaNames["Links"]); ok {
		link.Links = expandNetworkDeviceLinkConnections(v.(*schema.Set))
	}
	return link
}

func updateNetworkDeviceLinkResource(link *ne.DeviceLinkGroup, d *schema.ResourceData) error {
	if err := d.Set(networkDeviceLinkSchemaNames["UUID"], link.UUID); err != nil {
		return fmt.Errorf("error setting UUID: %s", err)
	}
	if err := d.Set(networkDeviceLinkSchemaNames["Name"], link.Name); err != nil {
		return fmt.Errorf("error setting Name: %s", err)
	}
	if err := d.Set(networkDeviceLinkSchemaNames["Subnet"], link.Subnet); err != nil {
		return fmt.Errorf("error setting Subnet: %s", err)
	}
	if err := d.Set(networkDeviceLinkSchemaNames["Status"], link.Status); err != nil {
		return fmt.Errorf("error setting Status: %s", err)
	}
	if err := d.Set(networkDeviceLinkSchemaNames["Devices"], flattenNetworkDeviceLinkDevices(d.Get(networkDeviceLinkSchemaNames["Devices"]).(*schema.Set), link.Devices)); err != nil {
		return fmt.Errorf("error setting Devices: %s", err)
	}
	if err := d.Set(networkDeviceLinkSchemaNames["Links"], flattenNetworkDeviceLinkConnections(d.Get(networkDeviceLinkSchemaNames["Devices"]).(*schema.Set), link.Links)); err != nil {
		return fmt.Errorf("error setting Links: %s", err)
	}
	return nil
}

func expandNetworkDeviceLinkDevices(devices *schema.Set) []ne.DeviceLinkGroupDevice {
	deviceList := devices.List()
	transformed := make([]ne.DeviceLinkGroupDevice, len(deviceList))
	for i := range deviceList {
		deviceMap := deviceList[i].(map[string]interface{})
		transformed[i] = ne.DeviceLinkGroupDevice{
			DeviceID:    ne.String(deviceMap[networkDeviceLinkDeviceSchemaNames["DeviceID"]].(string)),
			ASN:         ne.Int(deviceMap[networkDeviceLinkDeviceSchemaNames["ASN"]].(int)),
			InterfaceID: ne.Int(deviceMap[networkDeviceLinkDeviceSchemaNames["InterfaceID"]].(int)),
		}
	}
	return transformed
}

func flattenNetworkDeviceLinkDevices(currentDevices *schema.Set, devices []ne.DeviceLinkGroupDevice) interface{} {
	transformed := make([]interface{}, 0, len(devices))
	currentDevicesMap := schemaSetToMap(currentDevices)
	for i := range devices {
		transformedDevice := map[string]interface{}{
			networkDeviceLinkDeviceSchemaNames["DeviceID"]:    devices[i].DeviceID,
			networkDeviceLinkDeviceSchemaNames["InterfaceID"]: devices[i].InterfaceID,
			networkDeviceLinkDeviceSchemaNames["Status"]:      devices[i].Status,
			networkDeviceLinkDeviceSchemaNames["IPAddress"]:   devices[i].IPAddress,
		}
		if v, ok := currentDevicesMap[networkDeviceLinkDeviceHash(devices[i])]; ok {
			currentDeviceMap := v.(map[string]interface{})
			transformedDevice[networkDeviceLinkDeviceSchemaNames["ASN"]] = currentDeviceMap[networkDeviceLinkDeviceSchemaNames["ASN"]]
		}
		transformed = append(transformed, transformedDevice)
	}
	return transformed
}

func expandNetworkDeviceLinkConnections(connections *schema.Set) []ne.DeviceLinkGroupLink {
	connectionList := connections.List()
	transformed := make([]ne.DeviceLinkGroupLink, len(connectionList))
	for i := range connectionList {
		connectionMap := connectionList[i].(map[string]interface{})
		transformed[i] = ne.DeviceLinkGroupLink{
			AccountNumber:        ne.String(connectionMap[networkDeviceLinkConnectionSchemaNames["AccountNumber"]].(string)),
			Throughput:           ne.String(connectionMap[networkDeviceLinkConnectionSchemaNames["Throughput"]].(string)),
			ThroughputUnit:       ne.String(connectionMap[networkDeviceLinkConnectionSchemaNames["ThroughputUnit"]].(string)),
			SourceMetroCode:      ne.String(connectionMap[networkDeviceLinkConnectionSchemaNames["SourceMetroCode"]].(string)),
			DestinationMetroCode: ne.String(connectionMap[networkDeviceLinkConnectionSchemaNames["DestinationMetroCode"]].(string)),
			SourceZoneCode:       ne.String(connectionMap[networkDeviceLinkConnectionSchemaNames["SourceZoneCode"]].(string)),
			DestinationZoneCode:  ne.String(connectionMap[networkDeviceLinkConnectionSchemaNames["DestinationZoneCode"]].(string)),
		}
	}
	return transformed
}

func flattenNetworkDeviceLinkConnections(currentConnections *schema.Set, connections []ne.DeviceLinkGroupLink) interface{} {
	transformed := make([]interface{}, 0, len(connections))
	currentConnectionsMap := schemaSetToMap(currentConnections)
	for i := range connections {
		transformedConnection := map[string]interface{}{
			networkDeviceLinkConnectionSchemaNames["Throughput"]:           connections[i].Throughput,
			networkDeviceLinkConnectionSchemaNames["ThroughputUnit"]:       connections[i].ThroughputUnit,
			networkDeviceLinkConnectionSchemaNames["SourceMetroCode"]:      connections[i].SourceMetroCode,
			networkDeviceLinkConnectionSchemaNames["DestinationMetroCode"]: connections[i].DestinationMetroCode,
			networkDeviceLinkConnectionSchemaNames["SourceZoneCode"]:       connections[i].SourceZoneCode,
			networkDeviceLinkConnectionSchemaNames["DestinationZoneCode"]:  connections[i].DestinationZoneCode,
		}
		if v, ok := currentConnectionsMap[networkDeviceLinkConnectionHash(connections[i])]; ok {
			currentConnectionMap := v.(map[string]interface{})
			transformedConnection[networkDeviceLinkConnectionSchemaNames["AccountNumber"]] = currentConnectionMap[networkDeviceLinkConnectionSchemaNames["AccountNumber"]]
		}
		transformed = append(transformed, transformedConnection)
	}
	return transformed
}

type getDeviceLinkGroup func(uuid string) (*ne.DeviceLinkGroup, error)

func createDeviceLinkStatusProvisioningWaitConfiguration(fetchFunc getDeviceLinkGroup, id string, delay time.Duration, timeout time.Duration) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{
			ne.DeviceLinkGroupStatusProvisioning,
		},
		Target: []string{
			ne.DeviceLinkGroupStatusProvisioned,
		},
		Timeout:    timeout,
		Delay:      0,
		MinTimeout: delay,
		Refresh: func() (interface{}, string, error) {
			resp, err := fetchFunc(id)
			if err != nil {
				return nil, "", err
			}
			return resp, ne.StringValue(resp.Status), nil
		},
	}
}

func createDeviceLinkStatusDeleteWaitConfiguration(fetchFunc getDeviceLinkGroup, id string, delay time.Duration, timeout time.Duration) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{
			ne.DeviceLinkGroupStatusDeprovisioning,
		},
		Target: []string{
			ne.DeviceLinkGroupStatusDeprovisioned,
		},
		Timeout:    timeout,
		Delay:      0,
		MinTimeout: delay,
		Refresh: func() (interface{}, string, error) {
			resp, err := fetchFunc(id)
			if err != nil {
				if isRestNotFoundError(err) {
					return resp, ne.DeviceLinkGroupStatusDeprovisioned, nil
				}
				return nil, "", err
			}
			return resp, ne.StringValue(resp.Status), nil
		},
	}
}

func networkDeviceLinkDeviceKey(v interface{}) string {
	if v, ok := v.(ne.DeviceLinkGroupDevice); ok {
		return ne.StringValue(v.DeviceID)
	}
	if v, ok := v.(map[string]interface{}); ok {
		return fmt.Sprintf("%s", v[networkDeviceLinkDeviceSchemaNames["DeviceID"]])
	}
	return fmt.Sprintf("%v", v)
}

func networkDeviceLinkDeviceHash(v interface{}) int {
	return hashcodeString(networkDeviceLinkDeviceKey(v))
}

func networkDeviceLinkConnectionKey(v interface{}) string {
	if v, ok := v.(ne.DeviceLinkGroupLink); ok {
		return fmt.Sprintf("%s-%s-%s-%s",
			ne.StringValue(v.SourceMetroCode),
			ne.StringValue(v.SourceZoneCode),
			ne.StringValue(v.DestinationMetroCode),
			ne.StringValue(v.DestinationZoneCode))
	}
	if v, ok := v.(map[string]interface{}); ok {
		return fmt.Sprintf("%s-%s-%s-%s",
			v[networkDeviceLinkConnectionSchemaNames["SourceMetroCode"]],
			v[networkDeviceLinkConnectionSchemaNames["SourceZoneCode"]],
			v[networkDeviceLinkConnectionSchemaNames["DestinationMetroCode"]],
			v[networkDeviceLinkConnectionSchemaNames["DestinationZoneCode"]])
	}
	return fmt.Sprintf("%v", v)
}

func networkDeviceLinkConnectionHash(v interface{}) int {
	return hashcodeString(networkDeviceLinkConnectionKey(v))
}
