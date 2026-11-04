# TP-Link EasySmart UDP protocol — crypto reference

Reference notes on the RSA/RC4 scheme used by TP-Link's EasySmart discovery/management
UDP protocol (TL-SG105E, TL-SG108E, TL-SG116E and related switches). The production
RSA key is intentionally omitted. RSA test fixtures use a synthetic key and do not
interoperate with devices.

## Protocol flow

1. Client generates a random 8-character alphanumeric session key .
2. Client RSA-encrypts that session key with the switch's public key and sends it in a
   TLV of type **528**, using the *default* 256-byte RC4-like cipher
   to encrypt the UDP packet itself.
3. Switch decrypts the TLV 528 payload with the default rc4 key, recovering the 8-byte
   session key, and replies with errno **4098** on success.
4. From then on, traffic for that switch is encrypted with an
   **8-entry** RC4-like cipher seeded from the negotiated 8-byte session key,
   instead of the default 256-entry cipher.
5. Username/password login (TLV 512/514) happens after this, over the session cipher.

## RSA key (TLV 528 handshake)

The RSA implementation and TLV encoding are covered by unit tests using a synthetic
64-bit key. This intentionally small key is only suitable for tests and provides no
confidentiality.

### **528** encoded wire format (10 bytes)
1. Input bytes are reversed before being treated as a big-endian integer.
2. `c = m^e mod n` is computed with hand-rolled bignum arithmetic (16-bit limbs).
3. Output is 10 bytes: a 2-byte limb-count prefix, then 4 little-endian 16-bit limbs of `c`.

## RC4-like cipher (variable-size byte key, used before session key is established)

The product's default cipher key is intentionally omitted from this documentation.

## Hardware notes
- TL-SG105E / TL-SG108E (5/8 port): Realtek switch-on-chip with a built-in **8051**
  management core (classic MCS-51 opcodes confirmed by disassembling firmware bytes, 
  and separately in the SG108E firmware).
- TL-SG116E (16 port): bigger Realtek SoC (RTL8382-class "Longan") with a **MIPS32
  big-endian** core (confirmed by disassembling `lui`/`ori`/`lw`/`and`/`sw` sequences in
  its firmware), likely capable of running embedded Linux. OpenWrt has a `realtek`
  target for this SoC family, though flash/RAM on this switch tier (update image here is
  ~988KB) is likely too small for a full-featured image. (not investigated further)
