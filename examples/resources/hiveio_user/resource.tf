#Add users from Domain Admins on realm1 to be fabric admins
resource "hiveio_user" "users" {
  groupname = "Domain Admins"
  realm     = hiveio_realm.realm1.name
  role      = "admin"
}

#set user1 to be a realmAdmin
resource "hiveio_user" "user1" {
  username = "user1"
  realm    = hiveio_realm.realm1.name
  role     = "realmAdmin"
}

resource "hiveio_user" "user2" {
  username = "user2"
  realm    = hiveio_realm.realm1.name
  role     = "read-only"
}

