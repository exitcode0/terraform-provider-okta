resource "okta_device_posture_check" "example" {
  name          = "Disk Encryption Check"
  platform      = "MACOS"
  query         = "SELECT 1 FROM disk_encryption WHERE encrypted = 1;"
  description   = "Verifies that disk encryption is enabled"
  variable_name = "diskEncryptionCheck"
  mapping_type  = "CHECKBOX"

  remediation_settings = {
    message = {
      default_i18n_key = null
      custom_text      = "Please enable FileVault disk encryption on your Mac."
    }
    link = {
      default_url = null
      custom_url  = "https://support.apple.com/en-us/HT204837"
    }
  }
}
