// @expect: scriptgo-challenge
// @expect: true
// Certificate.exportChallenge reads the SPKAC challenge through OpenSSL's
// ASN1_STRING accessor, which OpenSSL 4.0 requires.
import { Certificate } from "node:crypto";
const spkac = "MIIBTDCBtjCBnzANBgkqhkiG9w0BAQEFAAOBjQAwgYkCgYEAza9vdBXOGdEazKuBRPsAC64f92xKOMg2OyLZYjgH7G50Z92F4rthZLovRYH+0DtQ4GcEge0igoGkkVKFqUVXW4fKtixiJYHtMjy2tECdy1MbD9roPVatRDI6xq30zyUXGkUF+kmIyAqaYrEITJ1+JTUmY9Q81uOOc4XdZmgKSRECAwEAARYSc2NyaXB0Z28tY2hhbGxlbmdlMA0GCSqGSIb3DQEBBAUAA4GBAIM+T89JfzUNnjHzyXEDzsE21fy052jQJgugIj/0OrUFPf6m6VxO8+ReizRJrv4n5tziPiQcsLq5JSqd2+D2aMxcrxAMDW2snANeMcwGTRjUifkVrpBX4PajCNjdl/sXny66fauBoKij6PnStXQwz2VSL53DHs9Yf3Jzv5Ggpb5e";
console.log(Certificate.exportChallenge(spkac).toString());
console.log(Certificate.verifySpkac(Buffer.from(spkac)));
