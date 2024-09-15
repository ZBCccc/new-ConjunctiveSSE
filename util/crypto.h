#ifndef CRYPTO_H
#define CRYPTO_H

#include <openssl/evp.h>
#include <openssl/hmac.h>
#include <openssl/rand.h>

void hmac_sha256(const unsigned char *key, int key_len, const unsigned char *data, int data_len, unsigned char *output, int *output_len);


#endif