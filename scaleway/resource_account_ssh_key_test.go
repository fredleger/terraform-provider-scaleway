package scaleway

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	account "github.com/scaleway/scaleway-sdk-go/api/account/v2alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func init() {
	resource.AddTestSweepers("scaleway_account_ssh_key", &resource.Sweeper{
		Name: "scaleway_account_ssh_key",
		F:    testSweepAccountSSHKey,
	})
}

func testSweepAccountSSHKey(region string) error {
	scwClient, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client in sweeper: %s", err)
	}
	accountAPI := account.NewAPI(scwClient)

	l.Debugf("sweeper: destroying the SSH keys")

	listSSHKeys, err := accountAPI.ListSSHKeys(&account.ListSSHKeysRequest{}, scw.WithAllPages())
	if err != nil {
		return fmt.Errorf("error listing SSH keys in sweeper: %s", err)
	}

	for _, sshKey := range listSSHKeys.SSHKeys {
		err := accountAPI.DeleteSSHKey(&account.DeleteSSHKeyRequest{
			SSHKeyID: sshKey.ID,
		})
		if err != nil {
			return fmt.Errorf("error deleting SSH key in sweeper: %s", err)
		}
	}

	return nil
}

func TestAccScalewayAccountSSHKey(t *testing.T) {
	name := newRandomName("ssh-key")
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEEYrzDOZmhItdKaDAEqJQ4ORS2GyBMtBozYsK5kiXXX opensource@scaleway.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayAccountSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `"
						public_key = "` + SSHKey + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists("scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `-updated"
						public_key = "` + SSHKey + `"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists("scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name+"-updated"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func TestAccScalewayAccountSSHKey_WithNewLine(t *testing.T) {
	name := newRandomName("ssh-key")
	SSHKey := "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIDjfkdWCwkYlVQMDUfiZlVrmjaGOfBYnmkucssae8Iup opensource@scaleway.com"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckScalewayAccountSSHKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: `
					resource "scaleway_account_ssh_key" "main" {
						name 	   = "` + name + `"
						public_key = "\n\n` + SSHKey + `\n\n"
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckScalewayAccountSSHKeyExists("scaleway_account_ssh_key.main"),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "name", name),
					resource.TestCheckResourceAttr("scaleway_account_ssh_key.main", "public_key", SSHKey),
				),
			},
		},
	})
}

func testAccCheckScalewayAccountSSHKeyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "scaleway_account_ssh_key" {
			continue
		}

		accountAPI := accountAPI(testAccProvider.Meta())

		_, err := accountAPI.GetSSHKey(&account.GetSSHKeyRequest{
			SSHKeyID: rs.Primary.ID,
		})

		// If no error resource still exist
		if err == nil {
			return fmt.Errorf("SSH key (%s) still exists", rs.Primary.ID)
		}

		// Unexpected api error we return it
		if !is404Error(err) {
			return err
		}
	}

	return nil
}

func testAccCheckScalewayAccountSSHKeyExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		accountAPI := accountAPI(testAccProvider.Meta())

		_, err := accountAPI.GetSSHKey(&account.GetSSHKeyRequest{
			SSHKeyID: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		return nil
	}
}
