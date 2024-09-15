#include "crypto.h"

void hmac_sha256(const unsigned char *key, int key_len, const unsigned char *data, int data_len, unsigned char *output, int *output_len) {
    HMAC_CTX *ctx = HMAC_CTX_new();
    if (ctx == NULL) {
        return;
    }

    HMAC_Init_ex(ctx, key, key_len, EVP_sha256(), NULL);
    HMAC_Update(ctx, data, data_len);
    HMAC_Final(ctx, output, output_len);

    HMAC_CTX_free(ctx);
}