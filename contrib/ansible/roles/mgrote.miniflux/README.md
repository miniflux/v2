## mgrote.miniflux

### Beschreibung
Installiert miniflux.
Und konfiguriert es.
Es wird ein Nutzer(Linux, PosGreSQL und miniflux) angelegt: "admin:hallowelt"
### Funktioniert auf
- [x] Ubuntu (>=18.04)

### Variablen + Defaults
##### Linux Nutzer für miniflux
    miniflux_linux_user: miniflux
##### Datenbanknutzer für ...
    miniflux_db_user_name: miniflux_db_user
##### Datenbanknutzerpasswort für ...
    miniflux_db_user_password: qqqqqqqqqqqqq
##### Datenbank für ...
    miniflux_db: miniflux_db
##### Nutzername für den Administrator
    miniflux_admin_name: admin
##### Passwort für den Administrator
    miniflux_admin_passwort: hallowelt
##### Port auf dem Miniflux erreichbar sein soll
    miniflux_port: 8080
