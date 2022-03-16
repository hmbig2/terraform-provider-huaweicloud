package dew

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	kps "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/kps/v3"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/kps/v3/model"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/logp"
)

func ResourceKeypair() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceKeypairCreate,
		UpdateContext: resourceKeypairUpdate,
		DeleteContext: resourceKeypairDelete,
		ReadContext:   resourceKeypairRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile("^[\\-_A-Za-z0-9]+$"),
						"The name can contain a maximum of 64 characters, including letters, digits, underscores (_),"+
							" and hyphens (-)."),
				),
			},

			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"ssh", "x509"}, false),
			},

			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"domain", "user"}, false),
			},

			"encryption_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"default", "kms"}, false),
			},

			"encryption_kms_key_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"public_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"key_file": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"is_managed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func resourceKeypairCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	region := c.GetRegion(d)
	client, err := config.NewKmsClient(c, region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating KMS v3 client: %s", err)
	}

	var importPublicKey *string
	if v, ok := d.GetOk("public_key"); ok {
		importPublicKey = utils.String(v.(string))
	}

	createOpts, err := buildCreateParams(d)
	if err != nil {
		return diag.FromErr(err)
	}

	logp.Printf("[DEBUG] Create KeyPair : %#v", createOpts)

	response, err := client.CreateKeypair(createOpts)
	if err != nil {
		return fmtp.DiagErrorf("Error creating KeyPair: %s", err)
	}

	d.SetId(*response.Keypair.Name)

	//update description
	updateErr := updateDesc(d, client)
	if updateErr != nil {
		return updateErr
	}

	//write private key to local. only when it is not import public_key and the key_file is not empty
	if fp := getKeyFilePath(d); importPublicKey != nil && fp != "" {
		if err = writeToPemFile(fp, *response.Keypair.PrivateKey); err != nil {
			return fmtp.DiagErrorf("Unable to generate private key: %s", err)
		}
		d.Set("key_file", fp)
	}

	return resourceKeypairRead(ctx, d, meta)
}

func resourceKeypairRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	region := c.GetRegion(d)
	client, err := config.NewKmsClient(c, region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating KMS v3 client: %s", err)
	}

	response, err := client.ListKeypairDetail(&model.ListKeypairDetailRequest{
		KeypairName: d.Id(),
	})
	if err != nil {
		return common.CheckDeletedError(d, err, "Error fetching keypair")
	}

	kType, err := parseEncodeValue(response.Keypair.Type.MarshalJSON())
	if err != nil {
		return fmtp.DiagErrorf("Can not parse the value of %q from response: %s", "type", err)
	}

	scope, err := parseEncodeValue(response.Keypair.Scope.MarshalJSON())
	if err != nil {
		return fmtp.DiagErrorf("Can not parse the value of %q from response: %s", "scope", err)
	}

	mErr := multierror.Append(nil,
		d.Set("region", region),
		d.Set("name", response.Keypair.Name),
		d.Set("type", kType),
		d.Set("scope", scope),
		d.Set("public_key", response.Keypair.PublicKey),
		d.Set("description", response.Keypair.Description),
		d.Set("created_at", utils.FormatTimeStampUTC(*response.Keypair.CreateTime/1000)),
		d.Set("fingerprint", response.Keypair.Fingerprint),
		d.Set("is_managed", response.Keypair.IsKeyProtection),
	)

	if err := mErr.ErrorOrNil(); err != nil {
		return fmtp.DiagErrorf("Error saving keypair fields: %s", err)
	}

	return nil
}

func resourceKeypairUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	region := c.GetRegion(d)
	client, err := config.NewKmsClient(c, region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating KMS v3 client: %s", err)
	}

	updateErr := updateDesc(d, client)
	if updateErr != nil {
		return updateErr
	}

	return resourceKeypairRead(ctx, d, meta)
}

func resourceKeypairDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*config.Config)
	region := c.GetRegion(d)
	client, err := config.NewKmsClient(c, region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating KMS v3 client: %s", err)
	}

	deleteOpts := &model.DeleteKeypairRequest{
		KeypairName: d.Id(),
	}

	_, err = client.DeleteKeypair(deleteOpts)
	if err != nil {
		return fmtp.DiagErrorf("Error deleting keypair: %s", err)
	}

	return nil
}

func buildCreateParams(d *schema.ResourceData) (*model.CreateKeypairRequest, error) {
	var importPublicKey *string
	if v, ok := d.GetOk("public_key"); ok {
		importPublicKey = utils.String(v.(string))
	}

	createOpts := &model.CreateKeypairRequest{
		Body: &model.CreateKeypairRequestBody{
			Keypair: &model.CreateKeypairAction{
				Name:      d.Get("name").(string),
				PublicKey: importPublicKey,
			},
		},
	}

	if v, ok := d.GetOk("type"); ok {
		var actionType model.CreateKeypairActionType
		err := actionType.UnmarshalJSON([]byte(v.(string)))
		if err != nil {
			return nil, fmtp.Errorf("Error parsing the argument %q: %s", "type", err)
		}
		createOpts.Body.Keypair.Type = &actionType
	}

	if v, ok := d.GetOk("scope"); ok {
		var actionScope model.CreateKeypairActionScope
		err := actionScope.UnmarshalJSON([]byte(v.(string)))
		if err != nil {
			return nil, fmtp.Errorf("Error parsing the argument %q: %s", "scope", err)
		}
		createOpts.Body.Keypair.Scope = &actionScope
	}

	if v, ok := d.GetOk("encryption_type"); ok {
		t := v.(string)
		var encryptionType model.EncryptionType
		err := encryptionType.UnmarshalJSON([]byte(t))
		if err != nil {
			return nil, fmtp.Errorf("Error parsing the argument %q: %s", "encryption_type", err)
		}

		keyProtection := model.KeyProtection{
			Encryption: &model.Encryption{
				Type: &encryptionType,
			},
		}

		//the kms key name is required when encryption_type="kms"
		if k, kmsExist := d.GetOk("encryption_kms_key_name"); t == "kms" && !kmsExist {
			return nil, fmtp.Errorf("encryption_kms_key_name is mandatory when the encryption_type is kms")
		} else {
			keyProtection.Encryption.KmsKeyName = utils.String(k.(string))
		}

		createOpts.Body.Keypair.KeyProtection = &keyProtection
	}

	return createOpts, nil
}

func parseEncodeValue(b []byte, err error) (*string, error) {
	if err != nil {
		return nil, err
	}

	var rst *string
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&rst)
	if err != nil {
		return nil, err
	}

	return rst, nil
}

func getKeyFilePath(d *schema.ResourceData) string {
	if path, ok := d.GetOk("key_file"); ok {
		return path.(string)
	}
	return ""
}

func writeToPemFile(path, privateKey string) error {
	var err error
	// If the private key exists, give it write permission for editing (-rw-------) for root user.
	if _, err = ioutil.ReadFile(path); err == nil {
		os.Chmod(path, 0600)
		defer os.Chmod(path, 0400) // read-only permission (-r--------).
	}
	if err = ioutil.WriteFile(path, []byte(privateKey), 0600); err != nil {
		return err
	}
	return nil
}

func updateDesc(d *schema.ResourceData, client *kps.KpsClient) diag.Diagnostics {
	if v, ok := d.GetOk("description"); ok {
		updateOpts := &model.UpdateKeypairDescriptionRequest{
			KeypairName: d.Id(),
			Body: &model.UpdateKeypairDescriptionRequestBody{
				Keypair: &model.UpdateKeypairDescriptionReq{
					Description: v.(string),
				},
			},
		}

		_, err := client.UpdateKeypairDescription(updateOpts)
		if err != nil {
			return fmtp.DiagErrorf("Error updating keypair: %s", err)
		}
	}

	return nil
}
