# fs Implementation Checklist

> **Category**: `CategoryNodeModule`  
> **Import Path**: `node:fs`  
> **Specification Reference**: [Node.js 22 LTS fs Documentation](https://nodejs.org/docs/latest-v22.x/api/fs.html)  
> **Type Definition Source**: [@types/node/fs.d.ts](https://github.com/DefinitelyTyped/DefinitelyTyped/tree/master/types/node)  
> **Gate Oracle**: Node.js 22 LTS test suite (test/parallel/test-fs-*.js)

---

## 1. Overview & Architectural Pipeline

Provide a concise technical summary:
- **Scope & Exposure**: Module-scoped symbols imported explicitly via `node:fs`.
- **Data & Memory Model**: Representation in IR (e.g., primitives, struct pointers, object shapes, buffer backing).
- **Lowering Pipeline**: Path from TypeScript AST → IR Instruction → Interpreter → LLVM runtime binding.

---

## 2. Parity Status Matrix

| API / Symbol / Property | TypeScript Signature | Lowering Target / Callee | Status | Corpus Test Path |
| :--- | :--- | :--- | :---: | :--- |
| `fs.appendFileSync(path, data[, options])` | `(...) => any` | `__fs.fs.appendFileSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/appendFileSync/` |
| `fs.copyFileSync(src, dest[, mode])` | `(...) => any` | `__fs.fs.copyFileSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/copyFileSync/` |
| `fs.existsSync(path)` | `(...) => any` | `__fs.fs.existsSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/existsSync/` |
| `fs.mkdirSync(path[, options])` | `(...) => any` | `__fs.fs.mkdirSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/mkdirSync/` |
| `fs.readFileSync(path[, options])` | `(...) => any` | `__fs.fs.readFileSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/readFileSync/` |
| `fs.readdirSync(path[, options])` | `(...) => any` | `__fs.fs.readdirSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/readdirSync/` |
| `fs.renameSync(oldPath, newPath)` | `(...) => any` | `__fs.fs.renameSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/renameSync/` |
| `fs.rmSync(path[, options])` | `(...) => any` | `__fs.fs.rmSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/rmSync/` |
| `fs.statSync(path[, options])` | `(...) => any` | `__fs.fs.statSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/statSync/` |
| `fs.unlinkSync(path)` | `(...) => any` | `__fs.fs.unlinkSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/unlinkSync/` |
| `fs.writeFileSync(file, data[, options])` | `(...) => any` | `__fs.fs.writeFileSync` | ✅ Done | `internal/compiler/testdata/corpus/api/fs/writeFileSync/` |
| `FileHandle` | `(...) => any` | `__fs.FileHandle` | 📋 Planned | - |
| `atime` | `any` | `__fs.atime` | 📋 Planned | - |
| `atimeMs` | `any` | `__fs.atimeMs` | 📋 Planned | - |
| `atimeNs` | `any` | `__fs.atimeNs` | 📋 Planned | - |
| `bavail` | `any` | `__fs.bavail` | 📋 Planned | - |
| `bfree` | `any` | `__fs.bfree` | 📋 Planned | - |
| `birthtime` | `any` | `__fs.birthtime` | 📋 Planned | - |
| `birthtimeMs` | `any` | `__fs.birthtimeMs` | 📋 Planned | - |
| `birthtimeNs` | `any` | `__fs.birthtimeNs` | 📋 Planned | - |
| `blksize` | `any` | `__fs.blksize` | 📋 Planned | - |
| `blocks` | `any` | `__fs.blocks` | 📋 Planned | - |
| `bsize` | `any` | `__fs.bsize` | 📋 Planned | - |
| `bytesRead` | `any` | `__fs.bytesRead` | 📋 Planned | - |
| `constants` | `any` | `__fs.constants` | 📋 Planned | - |
| `ctime` | `any` | `__fs.ctime` | 📋 Planned | - |
| `ctimeMs` | `any` | `__fs.ctimeMs` | 📋 Planned | - |
| `ctimeNs` | `any` | `__fs.ctimeNs` | 📋 Planned | - |
| `dev` | `any` | `__fs.dev` | 📋 Planned | - |
| `dir.close()` | `(...) => any` | `__fs.dir.close` | 📋 Planned | - |
| `dir.close(callback)` | `(...) => any` | `__fs.dir.close` | 📋 Planned | - |
| `dir.closeSync()` | `(...) => any` | `__fs.dir.closeSync` | 📋 Planned | - |
| `dir.read()` | `(...) => any` | `__fs.dir.read` | 📋 Planned | - |
| `dir.read(callback)` | `(...) => any` | `__fs.dir.read` | 📋 Planned | - |
| `dir.readSync()` | `(...) => any` | `__fs.dir.readSync` | 📋 Planned | - |
| `dir[Symbol.Dispose]()` | `(...) => any` | `__fs.dir[Symbol.Dispose]` | 📋 Planned | - |
| `dir[Symbol.asyncDispose]()` | `(...) => any` | `__fs.dir[Symbol.asyncDispose]` | 📋 Planned | - |
| `dir[Symbol.asyncIterator]()` | `(...) => any` | `__fs.dir[Symbol.asyncIterator]` | 📋 Planned | - |
| `dirent.isBlockDevice()` | `(...) => any` | `__fs.dirent.isBlockDevice` | 📋 Planned | - |
| `dirent.isCharacterDevice()` | `(...) => any` | `__fs.dirent.isCharacterDevice` | 📋 Planned | - |
| `dirent.isDirectory()` | `(...) => any` | `__fs.dirent.isDirectory` | 📋 Planned | - |
| `dirent.isFIFO()` | `(...) => any` | `__fs.dirent.isFIFO` | 📋 Planned | - |
| `dirent.isFile()` | `(...) => any` | `__fs.dirent.isFile` | 📋 Planned | - |
| `dirent.isSocket()` | `(...) => any` | `__fs.dirent.isSocket` | 📋 Planned | - |
| `dirent.isSymbolicLink()` | `(...) => any` | `__fs.dirent.isSymbolicLink` | 📋 Planned | - |
| `fd` | `any` | `__fs.fd` | 📋 Planned | - |
| `ffree` | `any` | `__fs.ffree` | 📋 Planned | - |
| `filehandle.appendFile(data[, options])` | `(...) => any` | `__fs.filehandle.appendFile` | 📋 Planned | - |
| `filehandle.chmod(mode)` | `(...) => any` | `__fs.filehandle.chmod` | 📋 Planned | - |
| `filehandle.chown(uid, gid)` | `(...) => any` | `__fs.filehandle.chown` | 📋 Planned | - |
| `filehandle.close()` | `(...) => any` | `__fs.filehandle.close` | 📋 Planned | - |
| `filehandle.createReadStream([options])` | `(...) => any` | `__fs.filehandle.createReadStream` | 📋 Planned | - |
| `filehandle.createWriteStream([options])` | `(...) => any` | `__fs.filehandle.createWriteStream` | 📋 Planned | - |
| `filehandle.datasync()` | `(...) => any` | `__fs.filehandle.datasync` | 📋 Planned | - |
| `filehandle.read([options])` | `(...) => any` | `__fs.filehandle.read` | 📋 Planned | - |
| `filehandle.read(buffer, offset, length, position)` | `(...) => any` | `__fs.filehandle.read` | 📋 Planned | - |
| `filehandle.read(buffer[, options])` | `(...) => any` | `__fs.filehandle.read` | 📋 Planned | - |
| `filehandle.readFile(options)` | `(...) => any` | `__fs.filehandle.readFile` | 📋 Planned | - |
| `filehandle.readLines([options])` | `(...) => any` | `__fs.filehandle.readLines` | 📋 Planned | - |
| `filehandle.readableWebStream([options])` | `(...) => any` | `__fs.filehandle.readableWebStream` | 📋 Planned | - |
| `filehandle.readv(buffers[, position])` | `(...) => any` | `__fs.filehandle.readv` | 📋 Planned | - |
| `filehandle.stat([options])` | `(...) => any` | `__fs.filehandle.stat` | 📋 Planned | - |
| `filehandle.sync()` | `(...) => any` | `__fs.filehandle.sync` | 📋 Planned | - |
| `filehandle.truncate(len)` | `(...) => any` | `__fs.filehandle.truncate` | 📋 Planned | - |
| `filehandle.utimes(atime, mtime)` | `(...) => any` | `__fs.filehandle.utimes` | 📋 Planned | - |
| `filehandle.write(buffer, offset[, length[, position]])` | `(...) => any` | `__fs.filehandle.write` | 📋 Planned | - |
| `filehandle.write(buffer[, options])` | `(...) => any` | `__fs.filehandle.write` | 📋 Planned | - |
| `filehandle.write(string[, position[, encoding]])` | `(...) => any` | `__fs.filehandle.write` | 📋 Planned | - |
| `filehandle.writeFile(data, options)` | `(...) => any` | `__fs.filehandle.writeFile` | 📋 Planned | - |
| `filehandle.writev(buffers[, position])` | `(...) => any` | `__fs.filehandle.writev` | 📋 Planned | - |
| `filehandle[Symbol.asyncDispose]()` | `(...) => any` | `__fs.filehandle[Symbol.asyncDispose]` | 📋 Planned | - |
| `files` | `any` | `__fs.files` | 📋 Planned | - |
| `fs.Dir` | `(...) => any` | `__fs.fs.Dir` | 📋 Planned | - |
| `fs.Dirent` | `(...) => any` | `__fs.fs.Dirent` | 📋 Planned | - |
| `fs.FSWatcher` | `(...) => any` | `__fs.fs.FSWatcher` | 📋 Planned | - |
| `fs.ReadStream` | `(...) => any` | `__fs.fs.ReadStream` | 📋 Planned | - |
| `fs.StatFs` | `(...) => any` | `__fs.fs.StatFs` | 📋 Planned | - |
| `fs.StatWatcher` | `(...) => any` | `__fs.fs.StatWatcher` | 📋 Planned | - |
| `fs.Stats` | `(...) => any` | `__fs.fs.Stats` | 📋 Planned | - |
| `fs.WriteStream` | `(...) => any` | `__fs.fs.WriteStream` | 📋 Planned | - |
| `fs.access(path[, mode], callback)` | `(...) => any` | `__fs.fs.access` | 📋 Planned | - |
| `fs.accessSync(path[, mode])` | `(...) => any` | `__fs.fs.accessSync` | 📋 Planned | - |
| `fs.appendFile(path, data[, options], callback)` | `(...) => any` | `__fs.fs.appendFile` | 📋 Planned | - |
| `fs.chmod(path, mode, callback)` | `(...) => any` | `__fs.fs.chmod` | 📋 Planned | - |
| `fs.chmodSync(path, mode)` | `(...) => any` | `__fs.fs.chmodSync` | 📋 Planned | - |
| `fs.chown(path, uid, gid, callback)` | `(...) => any` | `__fs.fs.chown` | 📋 Planned | - |
| `fs.chownSync(path, uid, gid)` | `(...) => any` | `__fs.fs.chownSync` | 📋 Planned | - |
| `fs.close(fd[, callback])` | `(...) => any` | `__fs.fs.close` | 📋 Planned | - |
| `fs.closeSync(fd)` | `(...) => any` | `__fs.fs.closeSync` | 📋 Planned | - |
| `fs.copyFile(src, dest[, mode], callback)` | `(...) => any` | `__fs.fs.copyFile` | 📋 Planned | - |
| `fs.cp(src, dest[, options], callback)` | `(...) => any` | `__fs.fs.cp` | 📋 Planned | - |
| `fs.cpSync(src, dest[, options])` | `(...) => any` | `__fs.fs.cpSync` | 📋 Planned | - |
| `fs.createReadStream(path[, options])` | `(...) => any` | `__fs.fs.createReadStream` | 📋 Planned | - |
| `fs.createWriteStream(path[, options])` | `(...) => any` | `__fs.fs.createWriteStream` | 📋 Planned | - |
| `fs.exists(path, callback)` | `(...) => any` | `__fs.fs.exists` | 📋 Planned | - |
| `fs.fchmod(fd, mode, callback)` | `(...) => any` | `__fs.fs.fchmod` | 📋 Planned | - |
| `fs.fchmodSync(fd, mode)` | `(...) => any` | `__fs.fs.fchmodSync` | 📋 Planned | - |
| `fs.fchown(fd, uid, gid, callback)` | `(...) => any` | `__fs.fs.fchown` | 📋 Planned | - |
| `fs.fchownSync(fd, uid, gid)` | `(...) => any` | `__fs.fs.fchownSync` | 📋 Planned | - |
| `fs.fdatasync(fd, callback)` | `(...) => any` | `__fs.fs.fdatasync` | 📋 Planned | - |
| `fs.fdatasyncSync(fd)` | `(...) => any` | `__fs.fs.fdatasyncSync` | 📋 Planned | - |
| `fs.fstat(fd[, options], callback)` | `(...) => any` | `__fs.fs.fstat` | 📋 Planned | - |
| `fs.fstatSync(fd[, options])` | `(...) => any` | `__fs.fs.fstatSync` | 📋 Planned | - |
| `fs.fsync(fd, callback)` | `(...) => any` | `__fs.fs.fsync` | 📋 Planned | - |
| `fs.fsyncSync(fd)` | `(...) => any` | `__fs.fs.fsyncSync` | 📋 Planned | - |
| `fs.ftruncate(fd[, len], callback)` | `(...) => any` | `__fs.fs.ftruncate` | 📋 Planned | - |
| `fs.ftruncateSync(fd[, len])` | `(...) => any` | `__fs.fs.ftruncateSync` | 📋 Planned | - |
| `fs.futimes(fd, atime, mtime, callback)` | `(...) => any` | `__fs.fs.futimes` | 📋 Planned | - |
| `fs.futimesSync(fd, atime, mtime)` | `(...) => any` | `__fs.fs.futimesSync` | 📋 Planned | - |
| `fs.glob(pattern[, options], callback)` | `(...) => any` | `__fs.fs.glob` | 📋 Planned | - |
| `fs.globSync(pattern[, options])` | `(...) => any` | `__fs.fs.globSync` | 📋 Planned | - |
| `fs.lchmod(path, mode, callback)` | `(...) => any` | `__fs.fs.lchmod` | 📋 Planned | - |
| `fs.lchmodSync(path, mode)` | `(...) => any` | `__fs.fs.lchmodSync` | 📋 Planned | - |
| `fs.lchown(path, uid, gid, callback)` | `(...) => any` | `__fs.fs.lchown` | 📋 Planned | - |
| `fs.lchownSync(path, uid, gid)` | `(...) => any` | `__fs.fs.lchownSync` | 📋 Planned | - |
| `fs.link(existingPath, newPath, callback)` | `(...) => any` | `__fs.fs.link` | 📋 Planned | - |
| `fs.linkSync(existingPath, newPath)` | `(...) => any` | `__fs.fs.linkSync` | 📋 Planned | - |
| `fs.lstat(path[, options], callback)` | `(...) => any` | `__fs.fs.lstat` | 📋 Planned | - |
| `fs.lstatSync(path[, options])` | `(...) => any` | `__fs.fs.lstatSync` | 📋 Planned | - |
| `fs.lutimes(path, atime, mtime, callback)` | `(...) => any` | `__fs.fs.lutimes` | 📋 Planned | - |
| `fs.lutimesSync(path, atime, mtime)` | `(...) => any` | `__fs.fs.lutimesSync` | 📋 Planned | - |
| `fs.mkdir(path[, options], callback)` | `(...) => any` | `__fs.fs.mkdir` | 📋 Planned | - |
| `fs.mkdtemp(prefix[, options], callback)` | `(...) => any` | `__fs.fs.mkdtemp` | 📋 Planned | - |
| `fs.mkdtempSync(prefix[, options])` | `(...) => any` | `__fs.fs.mkdtempSync` | 📋 Planned | - |
| `fs.open(path[, flags[, mode]], callback)` | `(...) => any` | `__fs.fs.open` | 📋 Planned | - |
| `fs.openAsBlob(path[, options])` | `(...) => any` | `__fs.fs.openAsBlob` | 📋 Planned | - |
| `fs.openSync(path[, flags[, mode]])` | `(...) => any` | `__fs.fs.openSync` | 📋 Planned | - |
| `fs.opendir(path[, options], callback)` | `(...) => any` | `__fs.fs.opendir` | 📋 Planned | - |
| `fs.opendirSync(path[, options])` | `(...) => any` | `__fs.fs.opendirSync` | 📋 Planned | - |
| `fs.read(fd, buffer, offset, length, position, callback)` | `(...) => any` | `__fs.fs.read` | 📋 Planned | - |
| `fs.read(fd, buffer[, options], callback)` | `(...) => any` | `__fs.fs.read` | 📋 Planned | - |
| `fs.read(fd[, options], callback)` | `(...) => any` | `__fs.fs.read` | 📋 Planned | - |
| `fs.readFile(path[, options], callback)` | `(...) => any` | `__fs.fs.readFile` | 📋 Planned | - |
| `fs.readSync(fd, buffer, offset, length[, position])` | `(...) => any` | `__fs.fs.readSync` | 📋 Planned | - |
| `fs.readSync(fd, buffer[, options])` | `(...) => any` | `__fs.fs.readSync` | 📋 Planned | - |
| `fs.readdir(path[, options], callback)` | `(...) => any` | `__fs.fs.readdir` | 📋 Planned | - |
| `fs.readlink(path[, options], callback)` | `(...) => any` | `__fs.fs.readlink` | 📋 Planned | - |
| `fs.readlinkSync(path[, options])` | `(...) => any` | `__fs.fs.readlinkSync` | 📋 Planned | - |
| `fs.readv(fd, buffers[, position], callback)` | `(...) => any` | `__fs.fs.readv` | 📋 Planned | - |
| `fs.readvSync(fd, buffers[, position])` | `(...) => any` | `__fs.fs.readvSync` | 📋 Planned | - |
| `fs.realpath(path[, options], callback)` | `(...) => any` | `__fs.fs.realpath` | 📋 Planned | - |
| `fs.realpath.native(path[, options], callback)` | `(...) => any` | `__fs.fs.realpath.native` | 📋 Planned | - |
| `fs.realpathSync(path[, options])` | `(...) => any` | `__fs.fs.realpathSync` | 📋 Planned | - |
| `fs.realpathSync.native(path[, options])` | `(...) => any` | `__fs.fs.realpathSync.native` | 📋 Planned | - |
| `fs.rename(oldPath, newPath, callback)` | `(...) => any` | `__fs.fs.rename` | 📋 Planned | - |
| `fs.rm(path[, options], callback)` | `(...) => any` | `__fs.fs.rm` | 📋 Planned | - |
| `fs.rmdir(path[, options], callback)` | `(...) => any` | `__fs.fs.rmdir` | 📋 Planned | - |
| `fs.rmdirSync(path[, options])` | `(...) => any` | `__fs.fs.rmdirSync` | 📋 Planned | - |
| `fs.stat(path[, options], callback)` | `(...) => any` | `__fs.fs.stat` | 📋 Planned | - |
| `fs.statfs(path[, options], callback)` | `(...) => any` | `__fs.fs.statfs` | 📋 Planned | - |
| `fs.statfsSync(path[, options])` | `(...) => any` | `__fs.fs.statfsSync` | 📋 Planned | - |
| `fs.symlink(target, path[, type], callback)` | `(...) => any` | `__fs.fs.symlink` | 📋 Planned | - |
| `fs.symlinkSync(target, path[, type])` | `(...) => any` | `__fs.fs.symlinkSync` | 📋 Planned | - |
| `fs.truncate(path[, len], callback)` | `(...) => any` | `__fs.fs.truncate` | 📋 Planned | - |
| `fs.truncateSync(path[, len])` | `(...) => any` | `__fs.fs.truncateSync` | 📋 Planned | - |
| `fs.unlink(path, callback)` | `(...) => any` | `__fs.fs.unlink` | 📋 Planned | - |
| `fs.unwatchFile(filename[, listener])` | `(...) => any` | `__fs.fs.unwatchFile` | 📋 Planned | - |
| `fs.utimes(path, atime, mtime, callback)` | `(...) => any` | `__fs.fs.utimes` | 📋 Planned | - |
| `fs.utimesSync(path, atime, mtime)` | `(...) => any` | `__fs.fs.utimesSync` | 📋 Planned | - |
| `fs.watch(filename[, options][, listener])` | `(...) => any` | `__fs.fs.watch` | 📋 Planned | - |
| `fs.watchFile(filename[, options], listener)` | `(...) => any` | `__fs.fs.watchFile` | 📋 Planned | - |
| `fs.write(fd, buffer, offset[, length[, position]], callback)` | `(...) => any` | `__fs.fs.write` | 📋 Planned | - |
| `fs.write(fd, buffer[, options], callback)` | `(...) => any` | `__fs.fs.write` | 📋 Planned | - |
| `fs.write(fd, string[, position[, encoding]], callback)` | `(...) => any` | `__fs.fs.write` | 📋 Planned | - |
| `fs.writeFile(file, data[, options], callback)` | `(...) => any` | `__fs.fs.writeFile` | 📋 Planned | - |
| `fs.writeSync(fd, buffer, offset[, length[, position]])` | `(...) => any` | `__fs.fs.writeSync` | 📋 Planned | - |
| `fs.writeSync(fd, buffer[, options])` | `(...) => any` | `__fs.fs.writeSync` | 📋 Planned | - |
| `fs.writeSync(fd, string[, position[, encoding]])` | `(...) => any` | `__fs.fs.writeSync` | 📋 Planned | - |
| `fs.writev(fd, buffers[, position], callback)` | `(...) => any` | `__fs.fs.writev` | 📋 Planned | - |
| `fs.writevSync(fd, buffers[, position])` | `(...) => any` | `__fs.fs.writevSync` | 📋 Planned | - |
| `fsPromises.access(path[, mode])` | `(...) => any` | `__fs.fsPromises.access` | 📋 Planned | - |
| `fsPromises.appendFile(path, data[, options])` | `(...) => any` | `__fs.fsPromises.appendFile` | 📋 Planned | - |
| `fsPromises.chmod(path, mode)` | `(...) => any` | `__fs.fsPromises.chmod` | 📋 Planned | - |
| `fsPromises.chown(path, uid, gid)` | `(...) => any` | `__fs.fsPromises.chown` | 📋 Planned | - |
| `fsPromises.copyFile(src, dest[, mode])` | `(...) => any` | `__fs.fsPromises.copyFile` | 📋 Planned | - |
| `fsPromises.cp(src, dest[, options])` | `(...) => any` | `__fs.fsPromises.cp` | 📋 Planned | - |
| `fsPromises.glob(pattern[, options])` | `(...) => any` | `__fs.fsPromises.glob` | 📋 Planned | - |
| `fsPromises.lchmod(path, mode)` | `(...) => any` | `__fs.fsPromises.lchmod` | 📋 Planned | - |
| `fsPromises.lchown(path, uid, gid)` | `(...) => any` | `__fs.fsPromises.lchown` | 📋 Planned | - |
| `fsPromises.link(existingPath, newPath)` | `(...) => any` | `__fs.fsPromises.link` | 📋 Planned | - |
| `fsPromises.lstat(path[, options])` | `(...) => any` | `__fs.fsPromises.lstat` | 📋 Planned | - |
| `fsPromises.lutimes(path, atime, mtime)` | `(...) => any` | `__fs.fsPromises.lutimes` | 📋 Planned | - |
| `fsPromises.mkdir(path[, options])` | `(...) => any` | `__fs.fsPromises.mkdir` | 📋 Planned | - |
| `fsPromises.mkdtemp(prefix[, options])` | `(...) => any` | `__fs.fsPromises.mkdtemp` | 📋 Planned | - |
| `fsPromises.open(path, flags[, mode])` | `(...) => any` | `__fs.fsPromises.open` | 📋 Planned | - |
| `fsPromises.opendir(path[, options])` | `(...) => any` | `__fs.fsPromises.opendir` | 📋 Planned | - |
| `fsPromises.readFile(path[, options])` | `(...) => any` | `__fs.fsPromises.readFile` | 📋 Planned | - |
| `fsPromises.readdir(path[, options])` | `(...) => any` | `__fs.fsPromises.readdir` | 📋 Planned | - |
| `fsPromises.readlink(path[, options])` | `(...) => any` | `__fs.fsPromises.readlink` | 📋 Planned | - |
| `fsPromises.realpath(path[, options])` | `(...) => any` | `__fs.fsPromises.realpath` | 📋 Planned | - |
| `fsPromises.rename(oldPath, newPath)` | `(...) => any` | `__fs.fsPromises.rename` | 📋 Planned | - |
| `fsPromises.rm(path[, options])` | `(...) => any` | `__fs.fsPromises.rm` | 📋 Planned | - |
| `fsPromises.rmdir(path[, options])` | `(...) => any` | `__fs.fsPromises.rmdir` | 📋 Planned | - |
| `fsPromises.stat(path[, options])` | `(...) => any` | `__fs.fsPromises.stat` | 📋 Planned | - |
| `fsPromises.statfs(path[, options])` | `(...) => any` | `__fs.fsPromises.statfs` | 📋 Planned | - |
| `fsPromises.symlink(target, path[, type])` | `(...) => any` | `__fs.fsPromises.symlink` | 📋 Planned | - |
| `fsPromises.truncate(path[, len])` | `(...) => any` | `__fs.fsPromises.truncate` | 📋 Planned | - |
| `fsPromises.unlink(path)` | `(...) => any` | `__fs.fsPromises.unlink` | 📋 Planned | - |
| `fsPromises.utimes(path, atime, mtime)` | `(...) => any` | `__fs.fsPromises.utimes` | 📋 Planned | - |
| `fsPromises.watch(filename[, options])` | `(...) => any` | `__fs.fsPromises.watch` | 📋 Planned | - |
| `fsPromises.writeFile(file, data[, options])` | `(...) => any` | `__fs.fsPromises.writeFile` | 📋 Planned | - |
| `gid` | `any` | `__fs.gid` | 📋 Planned | - |
| `ino` | `any` | `__fs.ino` | 📋 Planned | - |
| `mode` | `any` | `__fs.mode` | 📋 Planned | - |
| `mtime` | `any` | `__fs.mtime` | 📋 Planned | - |
| `mtimeMs` | `any` | `__fs.mtimeMs` | 📋 Planned | - |
| `mtimeNs` | `any` | `__fs.mtimeNs` | 📋 Planned | - |
| `name` | `any` | `__fs.name` | 📋 Planned | - |
| `nlink` | `any` | `__fs.nlink` | 📋 Planned | - |
| `parentPath` | `any` | `__fs.parentPath` | 📋 Planned | - |
| `path` | `any` | `__fs.path` | 📋 Planned | - |
| `path` {string}` | `any` | `__fs.path` {string}` | 📋 Planned | - |
| `pending` | `any` | `__fs.pending` | 📋 Planned | - |
| `rdev` | `any` | `__fs.rdev` | 📋 Planned | - |
| `size` | `any` | `__fs.size` | 📋 Planned | - |
| `stats.isBlockDevice()` | `(...) => any` | `__fs.stats.isBlockDevice` | 📋 Planned | - |
| `stats.isCharacterDevice()` | `(...) => any` | `__fs.stats.isCharacterDevice` | 📋 Planned | - |
| `stats.isDirectory()` | `(...) => any` | `__fs.stats.isDirectory` | 📋 Planned | - |
| `stats.isFIFO()` | `(...) => any` | `__fs.stats.isFIFO` | 📋 Planned | - |
| `stats.isFile()` | `(...) => any` | `__fs.stats.isFile` | 📋 Planned | - |
| `stats.isSocket()` | `(...) => any` | `__fs.stats.isSocket` | 📋 Planned | - |
| `stats.isSymbolicLink()` | `(...) => any` | `__fs.stats.isSymbolicLink` | 📋 Planned | - |
| `type` | `any` | `__fs.type` | 📋 Planned | - |
| `uid` | `any` | `__fs.uid` | 📋 Planned | - |
| `watcher.close()` | `(...) => any` | `__fs.watcher.close` | 📋 Planned | - |
| `watcher.ref()` | `(...) => any` | `__fs.watcher.ref` | 📋 Planned | - |
| `watcher.unref()` | `(...) => any` | `__fs.watcher.unref` | 📋 Planned | - |
| `writeStream.bytesWritten` | `any` | `__fs.writeStream.bytesWritten` | 📋 Planned | - |
| `writeStream.close([callback])` | `(...) => any` | `__fs.writeStream.close` | 📋 Planned | - |
| `writeStream.path` | `any` | `__fs.writeStream.path` | 📋 Planned | - |

---

## 3. Semantic Details & Edge Cases

### 3.1. Standard Behaviors
Describe expected standard semantics (e.g., IEEE-754 float precision, UTF-8 string encoding, implicit coercions, default argument handling).

### 3.2. Native Subset Restrictions
Document any constraints imposed by Ahead-Of-Time compilation (e.g., monomorphic type requirements, unsupported dynamic reflection).

### 3.3. Dual-Surface Mapping (if applicable)
Corpus test cases for `fs` are organized per API under `internal/compiler/testdata/corpus/fs/<api_name>/<test_case>/` and verify identical lowering semantics across surfaces.

---

## 4. Step-by-Step Implementation Recipe

When implementing or extending any symbol in this file, execute the following technical workflow:

- [ ] **Step 1: Frontend Type Contract**: Verify or register the ambient declaration in `internal/typescriptgo/stdlib/`.
- [ ] **Step 2: Lowering Registration**: Register the global value in `builtinGlobals` or intrinsic function in `builtinIntrinsics` within `internal/lowering/builtins.go`.
- [ ] **Step 3: IR Instruction Emission**: Lower the expression into standard IR instructions (`ir.OpCall`, `ir.OpObjectNew`, `ir.OpFieldSet`).
- [ ] **Step 4: Interpreter Handler**: Implement or bind reference execution in `internal/interpreter/`.
- [ ] **Step 5: LLVM / Runtime C ABI**: Declare the external C ABI or emit native LLVM IR in `internal/backend/llvm/` and `internal/runtime/`.
- [ ] **Step 6: Corpus Test Directory**: Create test subfolder under `internal/compiler/testdata/corpus/fs/<api_name>/<test_case>/` with `main.ts` and `run.expected`.
- [ ] **Step 7: Documentation Sync**: Re-run `go run ./scripts/gendocs/main.go` to auto-reflect `✅ Done` status in this checklist.

---

## 5. Known Gaps & Future Roadmap

- [ ] Unimplemented overloads or secondary options arguments.
- [ ] Future ECMAScript / Node.js proposals slated for upcoming milestones.
