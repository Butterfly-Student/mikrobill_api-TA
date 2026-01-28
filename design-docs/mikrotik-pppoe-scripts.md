# MikroTik PPP Profile Scripts

Copy dan paste script berikut ke **PPP Profile** yang digunakan (misalnya: `default` atau profile khusus).

Ganti `http://192.168.100.2:8080` dengan alamat IP dan port server API Anda.

## On Up Script

Script ini akan dijalankan ketika user terkoneksi. Ini akan memanggil API `/v1/callbacks/pppoe/up`.

```routeros
:local apiUrl "http://192.168.100.2:8080/v1/callbacks/pppoe/up"

:local user $user
:local callerId $"caller-id"
:local interfaceName $interface
:local localAddr $"local-address"
:local remoteAddr $"remote-address"
:local service $service

# Log debug (optional)
# :log info ("PPPoE UP: " . $user . " MAC:" . $callerId)

# Construct JSON payload manually
:local jsonData ("{\"name\":\"" . $user . \
               "\",\"caller_id\":\"" . $callerId . \
               "\",\"interface\":\"" . $interfaceName . \
               "\",\"local_address\":\"" . $localAddr . \
               "\",\"remote_address\":\"" . $remoteAddr . \
               "\",\"service\":\"" . $service . "\"}")

/tool fetch url=$apiUrl \
    http-method=post \
    http-header-field="Content-Type: application/json" \
    http-data=$jsonData \
    keep-result=no
```

## On Down Script

Script ini akan dijalankan ketika user disconnect. Ini akan memanggil API `/v1/callbacks/pppoe/down`.

```routeros
:local apiUrl "http://192.168.100.2:8080/v1/callbacks/pppoe/down"

:local user $user
:local callerId $"caller-id"
:local interfaceName $interface
:local localAddr $"local-address"
:local remoteAddr $"remote-address"
:local service $service

# Log debug (optional)
# :log info ("PPPoE DOWN: " . $user)

:local jsonData ("{\"name\":\"" . $user . \
               "\",\"caller_id\":\"" . $callerId . \
               "\",\"interface\":\"" . $interfaceName . \
               "\",\"local_address\":\"" . $localAddr . \
               "\",\"remote_address\":\"" . $remoteAddr . \
               "\",\"service\":\"" . $service . "\"}")

/tool fetch url=$apiUrl \
    http-method=post \
    http-header-field="Content-Type: application/json" \
    http-data=$jsonData \
    keep-result=no
```

## Cara Pasang di MikroTik

1. Buka WinBox / WebFig
2. Masuk ke menu **PPP** -> **Profiles**
3. Buka profile yang digunakan user (contoh: `default` atau `default-encryption`)
4. Pindah ke tab **Scripts**
5. Paste script di atas ke kolom **On Up** dan **On Down** (jangan lupa sesuaikan IP API)
6. Klik **OK**

Setelah terpasang, API Anda akan otomatis:
1. Menerima data user yang connect/disconnect
2. Mengupdate status di database
3. Mengupdate cache Redis
4. Mem-broadcast event ke WebSocket secara realtime
