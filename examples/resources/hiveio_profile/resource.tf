
#Basic profile without a realm
resource "hiveio_profile" "default_profile" {
  name     = "default"
  timezone = "disabled"
}

#Profile with a realm, user volumes and broker options
resource "hiveio_profile" "test" {
  name     = "test"
  timezone = "disabled"
  ad_config {
    domain     = hiveio_realm.test.name
    user_group = "Domain Users"
  }
  user_volumes {
    repository = hiveio_storage_pool.uvs.id
    target     = "nocache"
    size       = 10
  }
  broker_options {
    allow_desktop_composition   = true
    audio_capture               = true
    disable_full_window_drag    = false
    disable_menu_anims          = false
    disable_printer             = false
    disable_themes              = false
    disable_wallpaper           = false
    hide_authentication_failure = true
    inject_password             = false
    credssp                     = true
    redirect_clipboard          = true
    redirect_disk               = true
    redirect_pnp                = true
    redirect_printer            = true
    redirect_smartcard          = false
    redirect_usb                = true
    smart_resize                = true
    fail_on_cert_mismatch       = false
    html5                       = true
  }
}