Seclab Bot Protocol
===================

**Message Spec**

The client will establish a TLS1.2 connection with the server using python ssl with a custom SSL context which pins thewhitehat.club's certificate.

The client will send the following message:

```
  open (0xFF) | close (0x00)  (1 byte)   (sanity check, because "change state" is really the only required message)
  timestamp                   (4 bytes)  python int.to_bytes(time.time().__trunc__(), 4, 'little')
  HMAC                        (32 bytes)
```

python hmac will be used with SHA-256.

The HMAC will be keyed with a pre-shared secret 256-bit key, thus message validation will prove the sender knew the key. No other threat model really matters.

The server will simply disconnect on any error, or return an "all good" response on success.

The server will send the following message:

TBD
