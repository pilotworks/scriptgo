package llvm

import (
	"fmt"
	"strings"

	"github.com/pilotworks/scriptgo/internal/ir"
)

func (e *functionEmitter) emitCryptoIntrinsic(out *strings.Builder, instruction ir.Instruction) error {
	switch instruction.Callee {
	case "__crypto.randomUUID":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.randomUUID has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_random_uuid(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.hashDigest":
		if len(instruction.Args) < 2 || len(instruction.Args) > 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.hashDigest has invalid signature")
		}
		encodingArg := "null"
		if len(instruction.Args) == 3 {
			encodingArg = fmt.Sprintf("%%%s", instruction.Args[2])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hash_digest(ptr %%%s, ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], encodingArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.hashDigestBuffer":
		if len(instruction.Args) < 2 || len(instruction.Args) > 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.hashDigestBuffer has invalid signature")
		}
		encodingArg := "null"
		if len(instruction.Args) == 3 {
			encodingArg = fmt.Sprintf("%%%s", instruction.Args[2])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hash_digest_buffer(ptr %%%s, ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], encodingArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.randomBytes":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.randomBytes has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_random_bytes(double %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.randomInt":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("crypto.randomInt has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_random_int(double %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.randomFill":
		if len(instruction.Args) < 1 || len(instruction.Args) > 4 {
			return fmt.Errorf("crypto.randomFill has invalid signature")
		}
		offArg := "0.0"
		szArg := "0.0"
		if len(instruction.Args) >= 2 {
			offArg = fmt.Sprintf("%%%s", instruction.Args[1])
		}
		if len(instruction.Args) >= 3 {
			szArg = fmt.Sprintf("%%%s", instruction.Args[2])
		}
		callback := "null"
		if len(instruction.Args) == 4 {
			callback = fmt.Sprintf("%%%s", instruction.Args[3])
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_random_fill(ptr %%%s, double %s, double %s, ptr %s)\n", status, instruction.Args[0], offArg, szArg, callback)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		if instruction.Type == ir.TypeBuffer {
			fmt.Fprintf(out, "  %%%s = bitcast ptr %%%s to ptr\n", instruction.Result, instruction.Args[0])
		}
		return nil

	case "__crypto.timingSafeEqual":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("crypto.timingSafeEqual has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_timing_safe_equal(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil

	case "__crypto.hmacDigest":
		if len(instruction.Args) < 3 || len(instruction.Args) > 4 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.hmacDigest has invalid signature")
		}
		encodingArg := "null"
		if len(instruction.Args) == 4 {
			encodingArg = fmt.Sprintf("%%%s", instruction.Args[3])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hmac_digest(ptr %%%s, ptr %%%s, ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], encodingArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.hmacDigestBuffer":
		if len(instruction.Args) < 3 || len(instruction.Args) > 4 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.hmacDigestBuffer has invalid signature")
		}
		encodingArg := "null"
		if len(instruction.Args) == 4 {
			encodingArg = fmt.Sprintf("%%%s", instruction.Args[3])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hmac_digest_buffer(ptr %%%s, ptr %%%s, ptr %%%s, ptr %s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], encodingArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.pbkdf2Sync":
		if len(instruction.Args) < 4 || len(instruction.Args) > 5 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.pbkdf2Sync has invalid signature")
		}
		digestArg := "null"
		if len(instruction.Args) == 5 {
			digestArg = fmt.Sprintf("%%%s", instruction.Args[4])
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_pbkdf2_sync(ptr %%%s, ptr %%%s, double %%%s, double %%%s, ptr %s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], digestArg, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.hkdfSync":
		if len(instruction.Args) != 5 || instruction.Type != ir.TypeArrayBuffer {
			return fmt.Errorf("crypto.hkdfSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_hkdf_sync(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], instruction.Args[4], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.scryptSync":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.scryptSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_scrypt_sync(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.cipherCreate":
		if len(instruction.Args) != 4 || instruction.Type != ir.TypePointer {
			return fmt.Errorf("crypto.cipherCreate has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_cipher_create(ptr %%%s, ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.cipherUpdate":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.cipherUpdate has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_cipher_update(ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.cipherFinal":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.cipherFinal has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_cipher_final(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.cipherSetAAD":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("crypto.cipherSetAAD has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_cipher_set_aad(ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__crypto.cipherGetTag":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.cipherGetTag has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_cipher_get_tag(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.cipherSetTag":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("crypto.cipherSetTag has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_cipher_set_tag(ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__crypto.cipherSetAutoPadding":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("crypto.cipherSetAutoPadding has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_cipher_set_auto_padding(ptr %%%s, double %%%s)\n",
			status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__crypto.cipherDestroy":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("crypto.cipherDestroy has invalid signature")
		}
		fmt.Fprintf(out, "  call i32 @scriptgo_crypto_cipher_destroy(ptr %%%s)\n", instruction.Args[0])
		return nil

	case "__crypto.getCipherInfo":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.getCipherInfo has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_get_cipher_info(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.sign":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.sign has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_sign(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.verify":
		if len(instruction.Args) != 4 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("crypto.verify has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_verify(ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], instruction.Args[3], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil

	case "__crypto.publicEncrypt":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.publicEncrypt has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_public_encrypt(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.publicDecrypt":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.publicDecrypt has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_public_decrypt(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.privateEncrypt":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.privateEncrypt has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_private_encrypt(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.privateDecrypt":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.privateDecrypt has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_private_decrypt(ptr %%%s, ptr %%%s, double %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.generateKeyPairSync":
		if len(instruction.Args) != 3 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.generateKeyPairSync has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_generate_key_pair_sync(ptr %%%s, double %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.keyDetails":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.keyDetails has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_key_details(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.dhCreate":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypePointer {
			return fmt.Errorf("crypto.dhCreate has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_dh_create(ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.dhCreateGroup":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypePointer {
			return fmt.Errorf("crypto.dhCreateGroup has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_dh_create_group(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.dhGenerateKeys":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.dhGenerateKeys has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_dh_generate_keys(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.dhComputeSecret":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.dhComputeSecret has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_dh_compute_secret(ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.dhGetKey":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.dhGetKey has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_dh_get_key(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.dhSetKey":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("crypto.dhSetKey has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_dh_set_key(ptr %%%s, double %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__crypto.dhDestroy":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("crypto.dhDestroy has invalid signature")
		}
		fmt.Fprintf(out, "  call i32 @scriptgo_crypto_dh_destroy(ptr %%%s)\n", instruction.Args[0])
		return nil

	case "__crypto.ecdhCreate":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypePointer {
			return fmt.Errorf("crypto.ecdhCreate has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_ecdh_create(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.ecdhGenerateKeys":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.ecdhGenerateKeys has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_ecdh_generate_keys(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.ecdhComputeSecret":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.ecdhComputeSecret has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_ecdh_compute_secret(ptr %%%s, ptr %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.ecdhGetKey":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBuffer {
			return fmt.Errorf("crypto.ecdhGetKey has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_ecdh_get_key(ptr %%%s, double %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.ecdhSetKey":
		if len(instruction.Args) != 3 {
			return fmt.Errorf("crypto.ecdhSetKey has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_ecdh_set_key(ptr %%%s, double %%%s, ptr %%%s)\n",
			status, instruction.Args[0], instruction.Args[1], instruction.Args[2])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__crypto.ecdhDestroy":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("crypto.ecdhDestroy has invalid signature")
		}
		fmt.Fprintf(out, "  call i32 @scriptgo_crypto_ecdh_destroy(ptr %%%s)\n", instruction.Args[0])
		return nil

	case "__crypto.spkacVerify":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("crypto.spkacVerify has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_spkac_verify(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil

	case "__crypto.spkacExportChallenge":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.spkacExportChallenge has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_spkac_export_challenge(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.spkacExportPublicKey":
		if len(instruction.Args) != 1 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.spkacExportPublicKey has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_spkac_export_public_key(ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.x509CheckPrivateKey":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("crypto.x509CheckPrivateKey has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_x509_check_private_key(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil

	case "__crypto.x509Verify":
		if len(instruction.Args) != 2 || instruction.Type != ir.TypeBool {
			return fmt.Errorf("crypto.x509Verify has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_x509_verify(ptr %%%s, ptr %%%s, ptr %%%s)\n", status, instruction.Args[0], instruction.Args[1], slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s.f64 = load double, ptr %%%s\n", instruction.Result, slot)
		fmt.Fprintf(out, "  %%%s = fcmp one double %%%s.f64, 0.0\n", instruction.Result, instruction.Result)
		return nil

	case "__crypto.getFips":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeNumber {
			return fmt.Errorf("crypto.getFips has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca double\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_get_fips(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load double, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.setFips":
		if len(instruction.Args) != 1 {
			return fmt.Errorf("crypto.setFips has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_set_fips(double %%%s)\n", status, instruction.Args[0])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	case "__crypto.secureHeapUsed":
		if len(instruction.Args) != 0 || instruction.Type != ir.TypeString {
			return fmt.Errorf("crypto.secureHeapUsed has invalid signature")
		}
		slot := instruction.Result + ".slot"
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = alloca ptr\n", slot)
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_secure_heap_used(ptr %%%s)\n", status, slot)
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		fmt.Fprintf(out, "  %%%s = load ptr, ptr %%%s\n", instruction.Result, slot)
		return nil

	case "__crypto.setEngine":
		if len(instruction.Args) != 2 {
			return fmt.Errorf("crypto.setEngine has invalid signature")
		}
		status := fmt.Sprintf("runtime.status.%d", e.runtimeStatus)
		e.runtimeStatus++
		fmt.Fprintf(out, "  %%%s = call i32 @scriptgo_crypto_set_engine(ptr %%%s, double %%%s)\n", status, instruction.Args[0], instruction.Args[1])
		fmt.Fprintf(out, "  call void @scriptgo_runtime_abort_if_failed(i32 %%%s)\n", status)
		return nil

	default:
		return fmt.Errorf("unknown crypto intrinsic %q", instruction.Callee)
	}
}
