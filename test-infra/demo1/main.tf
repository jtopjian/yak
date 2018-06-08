resource "openstack_compute_instance_v2" "memcached" {
  count       = 2
  name        = "${format("memcached-%02d", count.index+1)}"
  image_name  = "Ubuntu 16.04"
  flavor_name = "m1.small"
  key_pair    = "infra"

  network {
    name = "default"
  }
}
