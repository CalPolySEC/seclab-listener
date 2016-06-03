Seclab Bot Protocol
===================

**Message Spec**

The client will establish a TLS1.2 connection with the server using python ssl with a custom SSL context which pins thewhitehat.club's certificate.

Accepted ciphers list in openssl cipher list format: AES256:AESCCM:AESGCM:CHACHA20:SUITEB128:SUITEB192

The client will send the following message:

```
  open (0xFF) | close (0x00) | keygen (0xAA)  (1 byte)   (sanity check, because "change state" is really the only required message, or to request a new key)
  timestamp                                   (8 bytes)  python int.to_bytes(time.time().__trunc__(), 8, 'little')
  HMAC                                        (32 bytes)
```

python hmac will be used with SHA-256.

The HMAC will be keyed with a pre-shared secret 256-bit key, thus message validation will prove the sender knew the key. No other threat model really matters.

The server will simply disconnect on any error, or return an "all good" response on success.

"All good" message:

```
  0xFF                        (1 byte) ALL GOOD!
```

New key response:

```
  0x55                        (1 byte)
  8-byte timestamp            (as above)
  HMAC                        (32 bytes)
```


A note on TLS
-------------

As few services should have access to the SSL private key as possible. nginx's
master process can be reasonably trusted with this key, we are already trusting
it to secure our web traffic.

This program doesn't need access to the key. Instead, we use TLS termination in
nginx and it will decrypt the traffic for us, then forward along an unencrypted
TCP stream to our service. (fun fact, nginx can do this even with generic,
non-HTTP traffic).
