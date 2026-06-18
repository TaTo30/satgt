import * as CryptoJS from "crypto-js"

function generateIV() {
    return CryptoJS.lib.WordArray.random(16);
}

function hashKey(key) {
    return CryptoJS.SHA256(key).toString(CryptoJS.enc.Hex);
}

function encrypt(originalString, key){
    const hashedKey = CryptoJS.enc.Hex.parse(hashKey(key));
    const iv = generateIV();

    const encrypted = CryptoJS.AES.encrypt(originalString, hashedKey, {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
    });


    const ivCiphertext = iv.concat(encrypted.ciphertext);
    return CryptoJS.enc.Base64.stringify(ivCiphertext);
}

function decrypt(encryptedBase64, key) {
    const hashedKey = CryptoJS.enc.Hex.parse(hashKey(key));

    // Decode the Base64 payload
    const ivCiphertext = CryptoJS.enc.Base64.parse(encryptedBase64);

    // Extract IV (first 16 bytes = 4 words)
    const iv = CryptoJS.lib.WordArray.create(
        ivCiphertext.words.slice(0, 4),
        16
    );

    // Extract ciphertext (remaining bytes)
    const ciphertext = CryptoJS.lib.WordArray.create(
        ivCiphertext.words.slice(4),
        ivCiphertext.sigBytes - 16
    );

    const decrypted = CryptoJS.AES.decrypt(
        { ciphertext: ciphertext },
        hashedKey,
        {
            iv: iv,
            mode: CryptoJS.mode.CBC,
            padding: CryptoJS.pad.Pkcs7
        }
    );

    return decrypted.toString(CryptoJS.enc.Utf8);
}


