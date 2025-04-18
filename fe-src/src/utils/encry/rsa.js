import JSEncrypt from 'jsencrypt';

// 加密
export function encrypt(data, publicKeyPem) {
  // 检查数据长度
  if (typeof data === 'string' && data.length > 117) {
    console.warn('RSA encryption data length exceeds 117 bytes limit. Consider using chunk encryption or other encryption methods.');
  }

  const encrypt = new JSEncrypt();
  const publicKey = publicKeyPem ?? `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCcbsc7X1y3xn7BvBL/bDCOqfng
ytBvn8mpvgZkOtEMcCLPmZu145BYn01OuZ7HQdb6tK7n7d5/y57avzZyJiAsVGR3
46FaU2AmvoNieoJ96K6GlnKHo8CgAyCwF3dVxp6TfIUHwGs4Z65m73XyXvrbKWW+
BInKK3XoG/qbdxdbpQIDAQAB
-----END PUBLIC KEY-----`;
  encrypt.setPublicKey(publicKey);
  const encryptedData = encrypt.encrypt(data);
  return encryptedData;
}

// 解密
export function decrypt(data, privateKeyPem) {
  const decryptor = new JSEncrypt();
  const privateKey = privateKeyPem ?? `
-----BEGIN RSA PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAJxuxztfXLfGfsG8Ev9sMI6p+eDK0G+fyam+BmQ60QxwIs+Zm7XjkFifTU65nsdB1vq0ruft3n/Lntq/NnImICxUZHfjoVpTYCa+g2J6gn3oroaWcoejwKADILAXd1XGnpN8hQfAazhnrmbvdfJe+tspZb4Eicordegb+pt3F1ulAgMBAAECgYAg7r1oxXG6isJCvPpu5XLvhd9CMNBiv4vv/T5ROYSrDqx1cgwy5Z6M2bSnvzIrFrRQgVtVHmG6G77spFas/1PES+evxGOV5AlXbyck2EwsRIKkIVOkUTAZwUDobF1z9eawDy54W1ko7uRIIDZIMJldSETSWfaKjBs5fwp5jxqb3QJBAOzGq3iVwYEiukyj50NcmKg63M2OEcO21urPTRrePd4zxJG4TrBapB3UT7Px9/InKkPtpdchiEvucdQfuGft3DMCQQCpIjFayOftXNi9YU8aQghYPZ6wiMT6LJOmlWCWjJTZW3bXFbBTqzDaQnYAQzuz9KC98g/Zq++D33TBF6SE2hDHAkEAwF7RZdFWPBL5BdeMx1/t75CTYLZynG5qwq/WV2QFJAkvRa1W0VVzTYD3mJ2Y8zb60eG9AcKOuBJsjQmQi2/nnQJALnycbiR8QqxbUioV0NTHcGF3ZXQiF9T6vDWgd6CqJNfT4Sgv779EzSipQEc6eKrLJ4oJuz1btrZLY+s4p9877wJBAMRM/E56TUPMedcOo7krWi/Rc4jfNWb0FFErNXJO6EEX+LmneUXF+zYqvGWjnC1SxqkYw7rCo+QwHu4lL5CEjMM=
-----END RSA PRIVATE KEY-----
`;
  decryptor.setPrivateKey(privateKey);
  const decryptedData = decryptor.decrypt(data);
  return decryptedData;
}


// 前端加密之分块加密（用于处理长文本）
export function webChunkEncrypt(data, publicKeyPem, chunkSize = 100) {
  if (typeof data !== 'string') {
    throw new Error('Data must be a string');
  }

  const chunks = [];
  for (let i = 0; i < data.length; i += chunkSize) {
    const chunk = data.slice(i, i + chunkSize);
    const encryptedChunk = encrypt(chunk, publicKeyPem);
    chunks.push(encryptedChunk);
  }

  return JSON.stringify(chunks); // 返回JSON字符串，方便传输
}

// 前端加密之分块解密
export function webChunkDecrypt(encryptedData, privateKeyPem) {
  try {
    const chunks = JSON.parse(encryptedData);
    if (!Array.isArray(chunks)) {
      throw new Error('Invalid encrypted data format');
    }

    return chunks
      .map(chunk => decrypt(chunk, privateKeyPem))
      .join('');
  } catch (error) {
    console.error('Chunk decrypt failed:', error);
    return null;
  }
}

/**
 * 后端返回数据的分块解密
 * @param {string} encryptedData - 后端返回的加密字符串
 * @param {string} [privateKeyPem] - RSA私钥
 * @returns {string} 解密后的字符串
 * 
 * @example
 * const encryptedData = "LlmPX2SGYaIonYjWfItAgiCG3fpJ75NtLdrTwgExeLZEg3mSdya9...";
 * const decrypted = chunkDecrypt(encryptedData);
 */
export function chunkDecrypt(encryptedData, privateKeyPem) {
  if (typeof encryptedData !== 'string') {
    throw new Error('Encrypted data must be a string');
  }

  try {
    // 每隔 172 个字符分割一次（因为RSA 1024位加密后的base64字符串长度固定为172）
    const chunkSize = 172;
    const chunks = [];

    for (let i = 0; i < encryptedData.length; i += chunkSize) {
      const chunk = encryptedData.slice(i, i + chunkSize);
      chunks.push(chunk);
    }

    const decryptedChunks = chunks.map(chunk => {
      if (!chunk) return '';
      const decrypted = decrypt(chunk, privateKeyPem);
      if (!decrypted) {
        console.warn('Decryption failed for chunk');
        return '';
      }
      return decrypted;
    });

    return decryptedChunks.filter(chunk => chunk).join('');
  } catch (error) {
    console.error('Chunk decrypt failed:', error);
    return '';
  }
}

// 使用示例：
/*
const encryptedData = "encrypted_chunk1.encrypted_chunk2.encrypted_chunk3";
const decrypted = chunkDecrypt(encryptedData);
*/

/**
 * 纯加密函数，用于生成与后端格式一致的加密字符串
 * @param {string} data - 需要加密的字符串
 * @param {string} [publicKeyPem] - RSA公钥
 * @returns {string} 加密后的字符串
 * 
 * @example
 * const data = "需要加密的内容";
 * const encrypted = chunkEncrypt(data);
 */
export function chunkEncrypt(data, publicKeyPem) {
  if (typeof data !== 'string') {
    throw new Error('Data must be a string');
  }

  try {
    // RSA 1024位密钥最大加密长度为117字节
    const chunkSize = 117;
    const chunks = [];

    // 分块加密
    for (let i = 0; i < data.length; i += chunkSize) {
      const chunk = data.slice(i, i + chunkSize);
      const encryptedChunk = encrypt(chunk, publicKeyPem);
      if (!encryptedChunk) {
        throw new Error('Encryption failed for chunk');
      }
      chunks.push(encryptedChunk);
    }

    // 直接拼接所有加密块（不使用分隔符）
    return chunks.join('');
  } catch (error) {
    console.error('Pure encrypt failed:', error);
    return '';
  }
}